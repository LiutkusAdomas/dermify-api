# Azure DevOps — One-time Setup Guide

Run these steps **once** before the pipelines can execute.
Everything after this is fully automated on every push to `main`.

---

## 1. Provision Azure infrastructure

If you haven't run `deploy.ps1` yet, do that first (it creates the resource group,
ACR, PostgreSQL, Container Apps environment, Storage, and Static Web App):

```powershell
cd dermify-api/infra
.\deploy.ps1 -DbPassword "YourStr0ng!Pass" -JwtSecret "your-jwt-secret-here"
```

---

## 2. Create the Azure DevOps project

1. Go to https://dev.azure.com and sign in.
2. Create a new **Organization** (or use an existing one).
3. Create a new **Project** — e.g. `dermify`.

---

## 3. Connect your Git repositories

In **Project Settings → Repos** (or just push to Azure Repos), OR connect GitHub:

- **Azure Repos** (simplest): push both repos here.
- **GitHub**: in `Pipelines → New Pipeline → GitHub`, authorize the GitHub app.

---

## 4. Create the Azure service connection

`Project Settings → Service connections → New service connection`

| Field | Value |
|---|---|
| Type | Azure Resource Manager |
| Auth | Workload Identity Federation (recommended) or Service Principal |
| Scope | Subscription → Resource Group `rg-dermify` |
| Name | **`dermify-azure`** ← must match the YAML exactly |

---

## 5. Create the variable group

`Pipelines → Library → + Variable group`

Name: **`dermify-secrets`**

Add these variables (mark secrets with the 🔒 lock icon):

| Variable | How to get the value | Secret? |
|---|---|---|
| `ACR_PASSWORD` | `az acr credential show --name dermifycr --query "passwords[0].value" -o tsv` | 🔒 |
| `OVERRIDE_DATABASE_PASSWORD` | The password you used in `deploy.ps1 -DbPassword` | 🔒 |
| `OVERRIDE_AUTH_JWT_SECRET` | The secret you used in `deploy.ps1 -JwtSecret` | 🔒 |
| `SWA_DEPLOYMENT_TOKEN` | `az staticwebapp secrets list --name dermify-web --resource-group rg-dermify --query "properties.apiKey" -o tsv` | 🔒 |
| `VITE_API_DOMAIN` | `https://<container-app-fqdn>/api/v1` (see below) | No |

To get the Container App FQDN for `VITE_API_DOMAIN`:
```bash
az containerapp show \
  --name dermify-api \
  --resource-group rg-dermify \
  --query "properties.configuration.ingress.fqdn" -o tsv
```
Then prefix with `https://` and suffix with `/api/v1`.

**After a custom domain is configured**, update `VITE_API_DOMAIN` to
`https://api.yourdomain.com/api/v1`.

---

## 6. Create the pipelines

### API pipeline
`Pipelines → New Pipeline → [your repo] → Existing Azure Pipelines YAML`
→ select `azure-pipelines.yml` in the `dermify-api` repo root.

### Web pipeline
`Pipelines → New Pipeline → [your repo] → Existing Azure Pipelines YAML`
→ select `azure-pipelines.yml` in the `dermify-web` repo root.

When prompted to authorize the variable group `dermify-secrets`, click **Permit**.

---

## 7. What runs when

| Event | API pipeline | Web pipeline |
|---|---|---|
| PR to `main` | CI only (lint + test) | CI only (lint + type-check + vitest) |
| Push to `main` | CI → CD (build image, push ACR, deploy Container App) | CI → CD (Vite build, deploy SWA) |

---

## Cost reminder

| Resource | Approximate monthly cost |
|---|---|
| Azure Pipelines (free tier) | Free — 1 parallel job, 1 800 min/month |
| Azure Container Registry (Basic) | ~$5 |
| Container Apps (0.25 vCPU / 0.5 GB, min 0 replicas) | ~$0–10 (scales to zero) |
| PostgreSQL Flexible Server B1ms | ~$13 |
| Azure Static Web Apps (Free tier) | $0 |
| Storage Account (LRS, photos) | < $1 |
| **Total** | **~$20–30/month** |
