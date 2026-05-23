import os

# ── API Gateway ───────────────────────────────────────────────────────────────
# Configure no arquivo .env (nunca commitar)
# API_BASE_URL=https://xxxxxxxxxx.execute-api.us-east-1.amazonaws.com/prod
API_BASE_URL = os.getenv("API_BASE_URL", "")

if not API_BASE_URL:
    raise EnvironmentError(
        "API_BASE_URL não configurada. "
        "Crie um arquivo .env com API_BASE_URL=<url do API Gateway>"
    )

# ── Endpoints ─────────────────────────────────────────────────────────────────
ENDPOINT_DEPLOY  = f"{API_BASE_URL}/documents/deploy"
ENDPOINT_STATUS  = f"{API_BASE_URL}/documents/deploy/status"

# ── Timeouts ──────────────────────────────────────────────────────────────────
TIMEOUT_UPLOAD  = 120
TIMEOUT_STATUS  = 10

# ── Polling ───────────────────────────────────────────────────────────────────
POLLING_INTERVAL = 3