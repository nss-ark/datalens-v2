# 24. DigiLocker Integration

## Overview

DigiLocker is India's national digital document wallet, governed by the Ministry of Electronics and Information Technology (MeITY). DataLens integrates with DigiLocker to support:

- **Identity verification** for Data Principals during consent and DPR flows
- **Age verification** for parental consent requirements (DPDPA Section 9)
- **Consent artifact storage** — push consent receipts into the user's DigiLocker

> [!NOTE]
> DigiLocker integration is an **optional** identity verification channel. Existing OTP-based verification remains the default.

---

## Integration Architecture

```
  Data Principal           DataLens                DigiLocker
  ──────────────          ────────                ──────────
       │                      │                        │
       │  1. Click "Verify    │                        │
       │     via DigiLocker"  │                        │
       │─────────────────────►│                        │
       │                      │  2. Redirect to        │
       │                      │     DigiLocker OAuth   │
       │                      │───────────────────────►│
       │                      │                        │
       │  3. User authorizes  │                        │
       │◄─────────────────────┼────────────────────────│
       │                      │  4. Receive auth code  │
       │                      │◄───────────────────────│
       │                      │  5. Exchange for token │
       │                      │───────────────────────►│
       │                      │  6. Get user details   │
       │                      │───────────────────────►│
       │                      │  7. Verify identity    │
       │                      │◄───────────────────────│
       │  8. Verified ✓       │                        │
       │◄─────────────────────│                        │
```

---

## OAuth 2.0 + PKCE Flow

DigiLocker uses OAuth 2.0 with PKCE for secure authorization.

### Step 1: Get Authorization Code

```http
GET https://api.digitallocker.gov.in/public/oauth2/1/authorize
  ?response_type=code
  &client_id={client_id}
  &redirect_uri={redirect_uri}
  &state={random_state}
  &code_challenge={sha256(code_verifier)}
  &code_challenge_method=S256
  &consent_valid_till={iso_date}
  &req_doctype=ADHAR
  &purpose=identity_verification
```

### Step 2: Exchange Code for Access Token

```http
POST https://api.digitallocker.gov.in/public/oauth2/1/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code={authorization_code}
&client_id={client_id}
&client_secret={client_secret}
&redirect_uri={redirect_uri}
&code_verifier={code_verifier}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJ...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "dGhpcyBpcyBh..."
}
```

### Step 3: Refresh Token

```http
POST https://api.digitallocker.gov.in/public/oauth2/1/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token
&refresh_token={refresh_token}
&client_id={client_id}
&client_secret={client_secret}
```

### Token Revocation

```http
POST https://api.digitallocker.gov.in/public/oauth2/1/revoke
Content-Type: application/x-www-form-urlencoded

token={access_token}
&client_id={client_id}
&client_secret={client_secret}
```

---

## Identity Verification

### Get User Details

Used to verify the identity of a Data Principal during consent or DPR flows.

```http
GET https://api.digitallocker.gov.in/public/oauth2/2/user
Authorization: Bearer {access_token}

Response:
{
  "digilockerid": "user@digilocker",
  "name": "Rajesh Kumar",
  "dob": "15/08/1990",
  "gender": "M",
  "eaadhaar": "XXXX-XXXX-1234",
  "mobile": "9876543210"
}
```

### Use Cases

| Use Case | DigiLocker API | Purpose |
|----------|---------------|---------|
| **Identity verification** | Get User Details | Confirm Data Principal identity for DPR processing |
| **Age verification** | Get User Details (`dob` field) | Determine if user is a minor (< 18) for guardian consent (DPDPA § 9) |
| **Document pull** | Get Issued Documents / Get File from URI | Retrieve KYC documents for identity confirmation |

---

## Consent Artifact Push

DataLens can push consent receipts into the user's DigiLocker account using the Push URI API:

```http
POST https://api.digitallocker.gov.in/public/oauth2/3/uri/pushuri
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "uri": "https://portal.datalens.io/consent-receipt/{receipt_id}",
  "doctype": "CONRC",
  "name": "Consent Receipt - {organization_name}",
  "date": "2026-02-10",
  "desc": "Consent granted for marketing, analytics purposes"
}
```

This allows Data Principals to maintain their consent artifacts alongside their other digital documents.

---

## Error Handling

| Error Code | Meaning | DataLens Action |
|------------|---------|-----------------|
| `invalid_grant` | Authorization code expired/invalid | Restart OAuth flow |
| `invalid_token` | Access token expired | Attempt refresh, if fails restart |
| `insufficient_scope` | Missing required permissions | Request with correct scope |
| `server_error` | DigiLocker service unavailable | Fall back to OTP verification |
| `rate_limit_exceeded` | Too many API calls | Back off and retry with exponential delay |

### Fallback Strategy

If DigiLocker is unavailable, DataLens automatically falls back to:
1. **Email OTP verification** (default)
2. **Phone OTP verification** (if phone available)

---

## Security

| Aspect | Implementation |
|--------|----------------|
| **Transport** | TLS 1.2+ for all DigiLocker API calls |
| **Authentication** | OAuth 2.0 + PKCE (no client secret in browser) |
| **Request signing** | HMAC-SHA-256 for API request integrity |
| **Token storage** | Encrypted at rest, short-lived (1 hour access tokens) |
| **Credentials** | Client ID/Secret stored in Vault, never in code |
| **Audit** | Every DigiLocker API call logged in audit trail |

---

## Configuration

```yaml
digilocker:
  enabled: true
  client_id: "${DIGILOCKER_CLIENT_ID}"
  client_secret: "${DIGILOCKER_CLIENT_SECRET}"
  redirect_uri: "https://api.datalens.complyark.com/v1/integrations/digilocker/callback"
  base_url: "https://api.digitallocker.gov.in/public/oauth2"
  scopes:
    - "user_details"
    - "issued_documents"
    - "push_uri"
  fallback_on_error: true
  timeout_ms: 5000
  retry_count: 3
```

---

## Related Documents

- [08_Consent_Management.md](./08_Consent_Management.md) — Guardian consent flow using DigiLocker age verification
- [10_API_Reference.md](./10_API_Reference.md) — DigiLocker integration API endpoints
- [12_Security_Compliance.md](./12_Security_Compliance.md) — Security requirements for external integrations
