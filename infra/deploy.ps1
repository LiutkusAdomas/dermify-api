<#
.SYNOPSIS
    Provisions and deploys the Dermify application to Azure.

.DESCRIPTION
    This script provisions all required Azure resources and deploys both the
    backend API (Container Apps) and frontend SPA (Static Web Apps).

    Prerequisites:
      - Azure CLI (az) installed and on PATH
      - Docker installed and running
      - Logged in to Azure (az login)

.PARAMETER Location
    Azure region for all resources. Default: northeurope

.PARAMETER ResourceGroup
    Name of the Azure resource group. Default: rg-dermify

.PARAMETER DbPassword
    Password for the PostgreSQL admin user. Required on first run.

.PARAMETER JwtSecret
    JWT signing secret for the API. Required on first run.

.PARAMETER CustomDomain
    Custom root domain (e.g. dermify.com). Used by the "domain" step.

.PARAMETER Step
    Run a specific step only: "infra", "backend", "frontend", "domain", or "all". Default: all

.EXAMPLE
    .\deploy.ps1 -DbPassword "MyStr0ng!Pass" -JwtSecret "super-secret-key-change-me"
    .\deploy.ps1 -Step backend
    .\deploy.ps1 -Step domain -CustomDomain dermify.com
#>

param(
    [string]$Location       = "northeurope",
    [string]$ResourceGroup  = "rg-dermify",
    [string]$DbPassword     = "",
    [string]$JwtSecret      = "",
    [string]$CustomDomain   = "",
    [string]$Step           = "all"
)

$ErrorActionPreference = "Stop"

# ---------- resource names ----------
$acrName         = "dermifycr"
$dbServerName    = "dermify-db"
$dbName          = "dermify"
$dbAdminUser     = "dermifyadmin"
$storageAccount  = "dermifystorage"
$fileShareName   = "photos"
$envName         = "dermify-env"
$apiAppName      = "dermify-api"
$swaName         = "dermify-web"
$storageMountName = "photostorage"
$imageName       = "$acrName.azurecr.io/dermify-api:latest"

# ---------- paths ----------
$repoRoot    = Split-Path -Parent $PSScriptRoot          # dermify-api
$webRoot     = Join-Path (Split-Path -Parent $repoRoot) "dermify-web"

# ============================================================
# Helper
# ============================================================
function Write-Step([string]$msg) {
    Write-Host "`n==> $msg" -ForegroundColor Cyan
}

# ============================================================
# STEP 1 - Provision infrastructure
# ============================================================
function Deploy-Infra {
    Write-Step "Registering required Azure resource providers"
    foreach ($provider in @(
        "Microsoft.ContainerRegistry",
        "Microsoft.App",
        "Microsoft.OperationalInsights",
        "Microsoft.DBforPostgreSQL",
        "Microsoft.Storage",
        "Microsoft.Web"
    )) {
        Write-Host "  Registering $provider ..."
        az provider register --namespace $provider --wait --output none
    }

    Write-Step "Creating resource group '$ResourceGroup' in '$Location'"
    az group create --name $ResourceGroup --location $Location --output none

    # --- Container Registry ---
    Write-Step "Creating Azure Container Registry '$acrName'"
    az acr create `
        --resource-group $ResourceGroup `
        --name $acrName `
        --sku Basic `
        --admin-enabled true `
        --output none

    # --- PostgreSQL Flexible Server ---
    Write-Step "Creating PostgreSQL Flexible Server '$dbServerName' (B1ms free-tier eligible)"
    if ([string]::IsNullOrEmpty($DbPassword)) {
        Write-Error "DbPassword is required for initial provisioning. Use -DbPassword parameter."
        return
    }

    az postgres flexible-server create `
        --resource-group $ResourceGroup `
        --name $dbServerName `
        --location $Location `
        --tier Burstable `
        --sku-name Standard_B1ms `
        --storage-size 32 `
        --version 15 `
        --admin-user $dbAdminUser `
        --admin-password $DbPassword `
        --public-access 0.0.0.0 `
        --yes `
        --output none

    az postgres flexible-server db create `
        --resource-group $ResourceGroup `
        --server-name $dbServerName `
        --database-name $dbName `
        --output none

    # --- Storage Account + File Share ---
    Write-Step "Creating Storage Account '$storageAccount' + file share '$fileShareName'"
    az storage account create `
        --name $storageAccount `
        --resource-group $ResourceGroup `
        --location $Location `
        --sku Standard_LRS `
        --output none

    az storage share-rm create `
        --storage-account $storageAccount `
        --name $fileShareName `
        --quota 5 `
        --output none

    # --- Container Apps Environment ---
    Write-Step "Creating Container Apps environment '$envName'"
    az containerapp env create `
        --name $envName `
        --resource-group $ResourceGroup `
        --location $Location `
        --output none

    $storageKey = (az storage account keys list `
        --account-name $storageAccount `
        --resource-group $ResourceGroup `
        --query "[0].value" -o tsv)

    az containerapp env storage set `
        --name $envName `
        --resource-group $ResourceGroup `
        --storage-name $storageMountName `
        --azure-file-account-name $storageAccount `
        --azure-file-account-key $storageKey `
        --azure-file-share-name $fileShareName `
        --access-mode ReadWrite `
        --output none

    # --- Static Web App ---
    # Static Web Apps only supports a limited set of regions; use westeurope as the nearest to northeurope
    $swaLocation = "westeurope"
    Write-Step "Creating Azure Static Web App '$swaName' (in '$swaLocation')"
    az staticwebapp create `
        --name $swaName `
        --resource-group $ResourceGroup `
        --location $swaLocation `
        --sku Free `
        --output none

    Write-Host "`nInfrastructure provisioning complete." -ForegroundColor Green
}

# ============================================================
# STEP 2 - Build & deploy backend
# ============================================================
function Deploy-Backend {
    if ([string]::IsNullOrEmpty($DbPassword) -or [string]::IsNullOrEmpty($JwtSecret)) {
        Write-Error "Both -DbPassword and -JwtSecret are required when deploying the backend."
        return
    }

    Write-Step "Logging in to ACR '$acrName'"
    az acr login --name $acrName

    Write-Step "Building Docker image"
    Push-Location $repoRoot
    docker build -t $imageName .
    Pop-Location

    Write-Step "Pushing image to ACR"
    docker push $imageName

    $acrPassword = (az acr credential show --name $acrName --query "passwords[0].value" -o tsv)
    $dbHost = "$dbServerName.postgres.database.azure.com"

    Write-Step "Deploying Container App '$apiAppName'"
    az containerapp create `
        --name $apiAppName `
        --resource-group $ResourceGroup `
        --environment $envName `
        --image $imageName `
        --registry-server "$acrName.azurecr.io" `
        --registry-username $acrName `
        --registry-password $acrPassword `
        --target-port 8080 `
        --ingress external `
        --min-replicas 0 `
        --max-replicas 1 `
        --cpu 0.25 --memory 0.5Gi `
        --env-vars `
            "OVERRIDE_ENVIRONMENT=production" `
            "OVERRIDE_DATABASE_HOST=$dbHost" `
            "OVERRIDE_DATABASE_PORT=5432" `
            "OVERRIDE_DATABASE_USER=$dbAdminUser" `
            "OVERRIDE_DATABASE_PASSWORD=$DbPassword" `
            "OVERRIDE_DATABASE_DBNAME=$dbName" `
            "OVERRIDE_DATABASE_SSLMODE=require" `
            "OVERRIDE_AUTH_JWT_SECRET=$JwtSecret" `
            "OVERRIDE_STORAGE_BASE_PATH=/tmp/photos" `
        --output none

    Write-Host "NOTE: Photo storage uses ephemeral /tmp/photos. For persistent storage," -ForegroundColor Yellow
    Write-Host "      add an Azure Files volume mount via the Azure Portal or YAML update." -ForegroundColor Yellow

    $apiFqdn = (az containerapp show `
        --name $apiAppName `
        --resource-group $ResourceGroup `
        --query "properties.configuration.ingress.fqdn" -o tsv)

    Write-Host "`nBackend deployed at: https://$apiFqdn" -ForegroundColor Green
    Write-Host "Update CORS origins if you add a custom domain." -ForegroundColor Yellow
}

# ============================================================
# STEP 3 - Build & deploy frontend
# ============================================================
function Deploy-Frontend {
    $apiFqdn = (az containerapp show `
        --name $apiAppName `
        --resource-group $ResourceGroup `
        --query "properties.configuration.ingress.fqdn" -o tsv)

    if ([string]::IsNullOrEmpty($apiFqdn)) {
        Write-Error "Could not determine API FQDN. Deploy the backend first."
        return
    }

    $apiUrl = "https://$apiFqdn/api/v1"

    Write-Step "Writing .env.production with API URL: $apiUrl"
    Set-Content -Path (Join-Path $webRoot ".env.production") -Value "VITE_API_DOMAIN=$apiUrl"

    Write-Step "Installing dependencies and building SPA"
    Push-Location $webRoot
    npm ci
    npm run build
    Pop-Location

    $deployToken = (az staticwebapp secrets list `
        --name $swaName `
        --resource-group $ResourceGroup `
        --query "properties.apiKey" -o tsv)

    Write-Step "Deploying to Azure Static Web Apps"
    Push-Location $webRoot
    npx --yes @azure/static-web-apps-cli deploy ./dist `
        --deployment-token $deployToken `
        --env production
    Pop-Location

    $swaHostname = (az staticwebapp show `
        --name $swaName `
        --resource-group $ResourceGroup `
        --query "defaultHostname" -o tsv)

    Write-Host "`nFrontend deployed at: https://$swaHostname" -ForegroundColor Green

    Write-Step "Updating CORS to allow frontend origin"
    az containerapp update `
        --name $apiAppName `
        --resource-group $ResourceGroup `
        --set-env-vars "OVERRIDE_CORS_ALLOWED_ORIGINS=https://$swaHostname" `
        --output none

    Write-Host "CORS updated to allow https://$swaHostname" -ForegroundColor Green
}

# ============================================================
# STEP 4 - Custom domain + SSL
# ============================================================
function Deploy-Domain {
    if ([string]::IsNullOrEmpty($CustomDomain)) {
        Write-Error "CustomDomain is required. Use -CustomDomain parameter (e.g. dermify.com)."
        return
    }

    $apiSubdomain = "api.$CustomDomain"

    # --- Static Web Apps custom domain ---
    Write-Step "Configuring custom domain '$CustomDomain' on Static Web App"

    $swaHostname = (az staticwebapp show `
        --name $swaName `
        --resource-group $ResourceGroup `
        --query "defaultHostname" -o tsv)

    Write-Host "Before proceeding, add these DNS records at your domain registrar:" -ForegroundColor Yellow
    Write-Host "  1. CNAME  www.$CustomDomain  ->  $swaHostname" -ForegroundColor Yellow
    Write-Host "  2. For apex ($CustomDomain), use an ALIAS/ANAME record -> $swaHostname" -ForegroundColor Yellow
    Write-Host "     (or use a www redirect if your registrar doesn't support ALIAS)" -ForegroundColor Yellow
    Write-Host ""

    $confirm = Read-Host "Have you created the DNS records? (y/n)"
    if ($confirm -ne "y") {
        Write-Host "Skipping domain configuration. Re-run with -Step domain when DNS is ready."
        return
    }

    az staticwebapp hostname set `
        --name $swaName `
        --resource-group $ResourceGroup `
        --hostname "www.$CustomDomain" `
        --output none

    Write-Host "Custom domain 'www.$CustomDomain' configured on Static Web App (auto-SSL)." -ForegroundColor Green

    # --- Container Apps custom domain ---
    Write-Step "Configuring custom domain '$apiSubdomain' on Container App"

    $apiFqdn = (az containerapp show `
        --name $apiAppName `
        --resource-group $ResourceGroup `
        --query "properties.configuration.ingress.fqdn" -o tsv)

    Write-Host "Add this DNS record at your domain registrar:" -ForegroundColor Yellow
    Write-Host "  CNAME  $apiSubdomain  ->  $apiFqdn" -ForegroundColor Yellow
    Write-Host ""

    $confirm2 = Read-Host "Have you created the CNAME record for $apiSubdomain? (y/n)"
    if ($confirm2 -ne "y") {
        Write-Host "Skipping API domain configuration. Re-run with -Step domain when DNS is ready."
        return
    }

    az containerapp hostname add `
        --name $apiAppName `
        --resource-group $ResourceGroup `
        --hostname $apiSubdomain `
        --output none

    az containerapp hostname bind `
        --name $apiAppName `
        --resource-group $ResourceGroup `
        --hostname $apiSubdomain `
        --environment $envName `
        --validation-method CNAME `
        --output none

    Write-Host "Custom domain '$apiSubdomain' configured on Container App (managed SSL)." -ForegroundColor Green

    # --- Update CORS and frontend env ---
    Write-Step "Updating CORS for custom domains"
    $origins = "https://$CustomDomain,https://www.$CustomDomain"
    az containerapp update `
        --name $apiAppName `
        --resource-group $ResourceGroup `
        --set-env-vars "OVERRIDE_CORS_ALLOWED_ORIGINS=$origins" `
                       "OVERRIDE_SMTP_FRONTEND_URL=https://$CustomDomain" `
        --output none

    Write-Host "CORS updated for $origins" -ForegroundColor Green

    Write-Step "Updating frontend .env.production for custom API domain"
    Set-Content -Path (Join-Path $webRoot ".env.production") -Value "VITE_API_DOMAIN=https://$apiSubdomain/api/v1"
    Write-Host "Run  .\deploy.ps1 -Step frontend  to rebuild and redeploy the frontend." -ForegroundColor Yellow
}

# ============================================================
# Dispatch
# ============================================================
switch ($Step) {
    "infra"    { Deploy-Infra }
    "backend"  { Deploy-Backend }
    "frontend" { Deploy-Frontend }
    "domain"   { Deploy-Domain }
    "all"      {
        Deploy-Infra
        Deploy-Backend
        Deploy-Frontend
        Write-Host "`n--- Custom domain step skipped during 'all'. Run separately: ---" -ForegroundColor Yellow
        Write-Host "  .\deploy.ps1 -Step domain -CustomDomain yourdomain.com" -ForegroundColor Yellow
    }
    default {
        Write-Error "Unknown step '$Step'. Use: infra, backend, frontend, domain, or all."
    }
}
