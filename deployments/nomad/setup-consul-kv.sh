#!/bin/bash
# Setup Consul KV for railzway-auth
# This script populates Consul KV store with configuration from .env file

set -e

ENV_FILE="${1:-.env}"

if [ ! -f "$ENV_FILE" ]; then
  echo "❌ Error: $ENV_FILE not found"
  echo "Usage: $0 [path/to/.env]"
  exit 1
fi

echo "📦 Populating Consul KV from $ENV_FILE..."
echo ""

# Source the .env file
set -a
source "$ENV_FILE"
set +a

# Database
consul kv put railzway-auth/db_host "$DB_HOST"
consul kv put railzway-auth/db_port "$DB_PORT"
consul kv put railzway-auth/db_name "$DB_NAME"
consul kv put railzway-auth/db_user "$DB_USER"
consul kv put railzway-auth/db_password "$DB_PASSWORD"
consul kv put railzway-auth/db_ssl_mode "$DB_SSL_MODE"

# Redis
consul kv put railzway-auth/redis_addr "${REDIS_ADDR:-127.0.0.1:6379}"
consul kv put railzway-auth/redis_password "${REDIS_PASSWORD:-}"
consul kv put railzway-auth/redis_db "${REDIS_DB:-0}"

# Bootstrap Admin
consul kv put railzway-auth/admin_email "$ADMIN_EMAIL"
consul kv put railzway-auth/admin_password "$ADMIN_PASSWORD"
consul kv put railzway-auth/default_org "${DEFAULT_ORG:-1000}"

# Tokens
consul kv put railzway-auth/access_token_ttl "${ACCESS_TOKEN_TTL:-1h}"
consul kv put railzway-auth/refresh_token_ttl "${REFRESH_TOKEN_TTL:-720h}"
consul kv put railzway-auth/refresh_token_bytes "${REFRESH_TOKEN_BYTES:-32}"

# Rate Limiting
consul kv put railzway-auth/rate_limit_rpm "${RATE_LIMIT_RPM:-600}"

# CORS
consul kv put railzway-auth/cors_allowed_origins "${CORS_ALLOWED_ORIGINS:-*}"
consul kv put railzway-auth/cors_allowed_methods "${CORS_ALLOWED_METHODS:-GET,POST,OPTIONS}"
consul kv put railzway-auth/cors_allowed_headers "${CORS_ALLOWED_HEADERS:-Authorization,Content-Type}"
consul kv put railzway-auth/cors_allow_credentials "${CORS_ALLOW_CREDENTIALS:-false}"

# Observability
consul kv put railzway-auth/otlp_endpoint "${OTLP_ENDPOINT:-}"
consul kv put railzway-auth/otlp_insecure "${OTLP_INSECURE:-true}"

echo ""
echo "✅ Consul KV populated successfully!"
echo ""
echo "📊 Verify with:"
echo "  consul kv get -recurse railzway-auth/"
