# 070 - Customer Onboarding Documentation

**Depends on:** [065_TenantProvisioning](065_TenantProvisioning_done.md)

## Description
Documentation guides for customers to configure their Identity Provider (IdP) for OIDC integration with EASI.

## Azure Entra (Azure AD)

### Setup Steps
1. Go to Azure Portal → Microsoft Entra ID → App registrations
2. Click "New registration"
3. Enter application name (e.g., "EASI")
4. Select "Accounts in this organizational directory only"
5. Click "Register"

### Configure Redirect URI
1. Go to Authentication → Add a platform → Web
2. Set redirect URI: `https://easi.example.com/api/v1/auth/callback`
3. Click "Configure"

### Create Client Secret
1. Go to Certificates & secrets → Client secrets
2. Click "New client secret"
3. Set description and expiry
4. **Copy the secret value immediately** (only shown once)

### Gather Configuration
Provide these values to EASI platform admin:
- **Discovery URL**: `https://login.microsoftonline.com/{tenant-id}/v2.0/.well-known/openid-configuration`
  - Find tenant ID in Overview page
- **Client ID**: Application (client) ID from Overview page
- **Client Secret**: The secret value copied earlier

### Required Permissions
No additional API permissions needed. Required scopes:
- `openid` - Required for OIDC
- `email` - User's email address
- `profile` - User's name
- `offline_access` - **Required for refresh tokens** (enables long-lived sessions)

---

## Okta

### Setup Steps
1. Go to Okta Admin Console → Applications → Applications
2. Click "Create App Integration"
3. Select "OIDC - OpenID Connect"
4. Select "Web Application"
5. Click "Next"

### Configure Application
1. App integration name: "EASI"
2. Sign-in redirect URIs: `https://easi.example.com/auth/callback`
3. Sign-out redirect URIs: (leave empty)
4. Controlled access: Select appropriate assignment
5. Click "Save"

### Configure Scopes
Ensure the following scopes are enabled:
- `openid`, `email`, `profile` - Standard OIDC scopes
- `offline_access` - Required for refresh tokens

### Gather Configuration
Provide these values to EASI platform admin:
- **Discovery URL**: `https://{your-domain}.okta.com/.well-known/openid-configuration`
- **Client ID**: From General tab → Client Credentials
- **Client Secret**: From General tab → Client Credentials

---

## Google Workspace

### Setup Steps
1. Go to Google Cloud Console → APIs & Services → Credentials
2. Click "Create Credentials" → "OAuth client ID"
3. If prompted, configure OAuth consent screen first:
   - User type: Internal (for Google Workspace)
   - App name, support email, developer contact
   - Scopes: `openid`, `email`, `profile`
   - Note: Google uses `access_type=offline` parameter instead of `offline_access` scope
4. Application type: Web application
5. Name: "EASI"

### Configure Redirect URI
1. Under "Authorized redirect URIs", add:
   `https://easi.example.com/auth/callback`
2. Click "Create"

### Gather Configuration
Provide these values to EASI platform admin:
- **Discovery URL**: `https://accounts.google.com/.well-known/openid-configuration`
- **Client ID**: Shown after creation
- **Client Secret**: Shown after creation (download JSON for backup)

### Notes
- Google Workspace with Internal user type restricts login to organization members
- External user type requires Google verification for production

---

## Generic OIDC Provider

### Requirements
Your OIDC provider must support:
- OpenID Connect Discovery (`.well-known/openid-configuration`)
- Authorization Code flow
- ID tokens with `email` claim

### Setup Steps
1. Create a new OIDC client/application in your IdP
2. Configure as "Web Application" or "Confidential Client"
3. Set redirect URI: `https://easi.example.com/auth/callback`
4. Enable scopes: `openid`, `email`, `profile`, `offline_access`
   - `offline_access` is required for refresh tokens (long-lived sessions)

### Gather Configuration
Provide these values to EASI platform admin:
- **Discovery URL**: Your IdP's `.well-known/openid-configuration` URL
- **Client ID**: From your IdP's application settings
- **Client Secret**: From your IdP's application settings

### Troubleshooting
| Issue | Solution |
|-------|----------|
| Discovery URL not working | Ensure URL ends with `.well-known/openid-configuration` |
| Login fails with invalid_client | Verify client ID and secret are correct |
| Missing email in ID token | Ensure `email` scope is enabled and user has email |

---

## Security Recommendations

1. **Client Secret Rotation**: Rotate client secrets periodically (e.g., annually)
2. **Conditional Access**: Consider adding conditional access policies in your IdP
3. **MFA**: Enable multi-factor authentication for users in your IdP
4. **Session Duration**: EASI access tokens last 8 hours, sessions extend up to 7 days via refresh tokens; IdP session policies may differ

## Checklist
- [x] Customer onboarding guide (Azure Entra)
- [x] Customer onboarding guide (Okta)
- [x] Customer onboarding guide (Google Workspace)
- [x] Generic OIDC setup guide
- [x] User sign-off
