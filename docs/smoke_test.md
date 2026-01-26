# Smoke Test: Login → OAuth Token → Valora OSS

This smoke test uses `curl` to validate the end-to-end flow from Valora Cloud Auth to Valora OSS.

## Prereqs

- Valora Cloud Auth running (local or hosted).
- Valora OSS redirect URI registered in the OAuth client.
- `client_id` and `client_secret` are configured for the OAuth client.

## Environment

```bash
export CLOUD_BASE_URL="https://usevalora.cloud"
export OSS_REDIRECT_URI="https://org.usevalora.net/login/usevalora_cloud"
export CLIENT_ID="your-client-id"
export CLIENT_SECRET="your-client-secret"
export EMAIL="test@valora.dev"
export PASSWORD="password123"
```

## Step 1 — Start OAuth Authorize (Cloud)

```bash
AUTH_URL="${CLOUD_BASE_URL}/oauth/authorize?response_type=code&client_id=${CLIENT_ID}&redirect_uri=${OSS_REDIRECT_URI}&scope=openid%20email%20profile&state=smoketest123"
curl -sS -D /tmp/valora_auth.headers -o /tmp/valora_auth.body "${AUTH_URL}"
```

Expected:
- `302` redirect
- `Location` header pointing to `/login?state=...`

Extract the authorize state:

```bash
AUTHORIZE_STATE=$(python - <<'PY'
import re
with open("/tmp/valora_auth.headers", "r") as f:
    for line in f:
        if line.lower().startswith("location:"):
            m = re.search(r"state=([^&\\s]+)", line)
            if m:
                print(m.group(1))
                break
PY
)
echo "${AUTHORIZE_STATE}"
```

## Step 2 — Password Login (Cloud)

```bash
curl -sS -D /tmp/valora_auth.login.headers \
  -c /tmp/valora_auth.cookies \
  -o /tmp/valora_auth.login.json \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\",\"state\":\"${AUTHORIZE_STATE}\"}" \
  "${CLOUD_BASE_URL}/auth/password/login"
```

Expected:
- `200 OK`
- `Set-Cookie` for the Cloud session
- Response includes `authorize_url`

Extract the authorize URL (using jq):

```bash
AUTHORIZE_URL=$(jq -r .authorize_url /tmp/valora_auth.login.json)
echo "${AUTHORIZE_URL}"
```

## Step 3 — Authorize (OAuth)

```bash
curl -sS -i -b /tmp/valora_auth.cookies "${CLOUD_BASE_URL}${AUTHORIZE_URL}"
```

Expected:
- `302` redirect
- `Location` header with `redirect_uri?code=AUTH_CODE&state=smoketest123`

Extract the code:

```bash
AUTH_CODE=$(curl -sS -i -b /tmp/valora_auth.cookies "${CLOUD_BASE_URL}${AUTHORIZE_URL}" | python - <<'PY'
import sys, urllib.parse
for line in sys.stdin:
    if line.lower().startswith("location:"):
        loc = line.split(":", 1)[1].strip()
        qs = urllib.parse.parse_qs(urllib.parse.urlparse(loc).query)
        print(qs.get("code", [""])[0])
        break
PY
)
echo "${AUTH_CODE}"
```

Or copy the `code` manually from the `Location` header.

## Step 4 — Token Exchange (OSS Side)

```bash
export AUTH_CODE="paste-auth-code"
curl -sS \
  -u "${CLIENT_ID}:${CLIENT_SECRET}" \
  -d "grant_type=authorization_code" \
  -d "code=${AUTH_CODE}" \
  -d "redirect_uri=${OSS_REDIRECT_URI}" \
  "${CLOUD_BASE_URL}/oauth/token"
```

Expected:
- `200 OK`
- `access_token` returned
- `token_type=Bearer`
- `expires_in` present

## Step 5 — Userinfo

```bash
export ACCESS_TOKEN="paste-access-token"
curl -sS \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  "${CLOUD_BASE_URL}/oauth/userinfo"
```

Expected JSON (fields may include additional data):

```json
{
  "sub": "<stable_external_id>",
  "email": "test@valora.dev",
  "org_id": "<org_id>",
  "provider": "usevalora_cloud"
}
```

## Step 6 — OSS User Bootstrap (Assertion)

On the Valora OSS side, verify:
- User is upserted by `provider=usevalora_cloud` + `external_id=sub`.
- User is created without password.
- Org membership is assigned.
- OSS local session is created.

## Pass Criteria

- No public create-user API was called.
- No session/cookie was shared across domains.
- Identity came only from OAuth exchange.
- `org_id` is present end-to-end.
