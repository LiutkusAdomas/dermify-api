# Dermify Compliance and Data Encryption Roadmap

This document outlines the steps needed to make Dermify compliant with healthcare
data regulations when moving from demonstration to production with real patient data.

---

## Current State (Demo)

- PostgreSQL Flexible Server: encrypted at rest (Azure service-managed keys, AES-256)
- Azure Files / Blob Storage: encrypted at rest (AES-256, service-managed keys)
- TLS 1.2+ enforced on all Azure services
- PostgreSQL connections use `sslmode=require`
- HTTPS enforced on Static Web Apps and Container Apps

## Step 1: Encryption Upgrades

### 1a. Customer-Managed Keys (CMK) via Azure Key Vault

Switch from service-managed keys to customer-managed keys for full control:

```bash
# Create Key Vault
az keyvault create \
  --name dermify-kv \
  --resource-group rg-dermify \
  --location northeurope \
  --enable-purge-protection

# Create encryption key
az keyvault key create \
  --vault-name dermify-kv \
  --name dermify-db-key \
  --kty RSA --size 2048

# Enable CMK on PostgreSQL
az postgres flexible-server update \
  --resource-group rg-dermify \
  --name dermify-db \
  --key <key-identifier> \
  --identity <managed-identity-id>
```

### 1b. Application-Level Field Encryption

For PII fields (patient names, addresses, phone numbers, medical notes), add
AES-256-GCM column-level encryption in the Go backend:

- Store encryption keys in Azure Key Vault
- Add `github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys` dependency
- Create an encryption service that wraps/unwraps data keys using Key Vault
- Encrypt sensitive fields before database writes, decrypt after reads
- Fields to encrypt: patient first/last name, phone, email, address, medical notes

### 1c. Photo Storage Migration

Migrate from Azure Files to Azure Blob Storage:

- Add `github.com/Azure/azure-sdk-for-go/sdk/storage/azblob` dependency
- Create a `BlobStorageAdapter` implementing the existing storage interface
- Enable per-blob encryption with CMK
- Add CDN for photo delivery (optional)

## Step 2: GDPR Compliance (EU)

Deploy in EU region (already using `northeurope`).

### Required Implementation:

1. **Right to Erasure (Article 17)**: Add `DELETE /api/v1/patients/{id}/erase` endpoint
   that permanently deletes all patient data, sessions, photos, and audit entries.

2. **Data Portability (Article 20)**: Add `GET /api/v1/patients/{id}/export` endpoint
   returning all patient data in a machine-readable format (JSON/CSV).

3. **Consent Management**: Already have `consent` table. Add explicit consent tracking
   for data processing purposes with timestamps and versioned consent texts.

4. **Data Processing Agreement**: Azure provides a DPA as part of Microsoft Product Terms.
   No action needed from the application side.

5. **Data Breach Notification**: Implement alerting on unauthorized access patterns.
   Consider Azure Defender for PostgreSQL.

## Step 3: HIPAA Compliance (US patients)

### Administrative:

1. Sign a **Business Associate Agreement (BAA)** with Microsoft
   (available on qualifying Azure plans).

2. Document security policies and procedures.

3. Conduct regular risk assessments.

### Technical:

1. **Audit Logging**: Already have `audit_trail` table and triggers.
   Enable PostgreSQL audit logging:
   ```bash
   az postgres flexible-server parameter set \
     --resource-group rg-dermify \
     --server-name dermify-db \
     --name pgaudit.log \
     --value "read, write, ddl"
   ```

2. **Access Controls**: Already have role-based access (admin, doctor, receptionist, viewer).
   Add session timeout enforcement and failed login lockout.

3. **Azure Defender**: Enable threat detection:
   ```bash
   az postgres flexible-server advanced-threat-protection-setting update \
     --resource-group rg-dermify \
     --server-name dermify-db \
     --state Enabled
   ```

4. **Backup Encryption**: Azure PostgreSQL backups are encrypted by default.

## Step 4: Other Jurisdictions

| Region | Regulation | Key Requirements |
|--------|-----------|-----------------|
| EU | GDPR | Data residency, right to erasure, consent, DPA |
| US | HIPAA | BAA, encryption, audit trails, access controls |
| UK | UK GDPR | Similar to EU GDPR post-Brexit |
| Canada | PIPEDA | Consent, accountability, limiting use |
| Australia | Privacy Act | APP compliance, data breach notification |
| Brazil | LGPD | Similar to GDPR, data protection officer |

For global deployment, the EU region + GDPR compliance covers the strictest
requirements. HIPAA adds US-specific obligations for healthcare data.
