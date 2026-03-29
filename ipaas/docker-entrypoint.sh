#!/bin/sh
set -e

echo "ICP ipaas Console - Initializing runtime configuration..."

THUNDER_URL="${VITE_THUNDER_URL:-}"
THUNDER_APP_ID="${VITE_THUNDER_APP_ID:-}"
THUNDER_CLIENT_ID="${VITE_THUNDER_CLIENT_ID:-}"
THUNDER_CLIENT_SECRET="${VITE_THUNDER_CLIENT_SECRET:-}"
THUNDER_REDIRECT_URI="${VITE_THUNDER_REDIRECT_URI:-}"
THUNDER_SCOPES="${VITE_THUNDER_SCOPES:-openid profile email}"
THUNDER_AUTHENTICATOR="${VITE_THUNDER_AUTHENTICATOR:-BasicAuthenticator}"
THUNDER_AFTER_SIGN_IN_URL="${VITE_THUNDER_AFTER_SIGN_IN_URL:-}"
THUNDER_AFTER_SIGN_OUT_URL="${VITE_THUNDER_AFTER_SIGN_OUT_URL:-}"

cat > /usr/share/nginx/html/env-config.js <<EOF_INNER || echo "env-config.js is read-only (provided by platform injection), skipping write"
// Runtime environment configuration
// Generated at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")

window._env_ = {
  // ICP BFF API base URL
  ICP_API_BASE_URL: "${ICP_API_BASE_URL:-/ipaas-service}",
  ICP_ORG_NUMERIC_ID: "${ICP_ORG_NUMERIC_ID:-}",

  // Thunder/Asgardeo Authentication
  VITE_THUNDER_URL: "${THUNDER_URL}",
  VITE_THUNDER_APP_ID: "${THUNDER_APP_ID}",
  VITE_THUNDER_CLIENT_ID: "${THUNDER_CLIENT_ID}",
  VITE_THUNDER_CLIENT_SECRET: "${THUNDER_CLIENT_SECRET}",
  VITE_THUNDER_REDIRECT_URI: "${THUNDER_REDIRECT_URI}",
  VITE_THUNDER_SCOPES: "${THUNDER_SCOPES}",
  VITE_THUNDER_AUTHENTICATOR: "${THUNDER_AUTHENTICATOR}",
  VITE_THUNDER_AFTER_SIGN_IN_URL: "${THUNDER_AFTER_SIGN_IN_URL}",
  VITE_THUNDER_AFTER_SIGN_OUT_URL: "${THUNDER_AFTER_SIGN_OUT_URL}",

  // Development
  VITE_DEV_BYPASS_AUTH: "${VITE_DEV_BYPASS_AUTH:-}",
};

console.log('[ICP ipaas Console] Runtime configuration loaded:', {
  icpApiBaseUrl: window._env_.ICP_API_BASE_URL,
  thunderUrl: window._env_.VITE_THUNDER_URL,
  clientId: window._env_.VITE_THUNDER_CLIENT_ID,
});
EOF_INNER

echo "Runtime configuration generated at /usr/share/nginx/html/env-config.js"

echo "Configuration Summary:"
echo "   ICP API Base URL: ${ICP_API_BASE_URL:-/ipaas-service}"
echo "   Thunder URL: ${THUNDER_URL:-[NOT SET]}"
echo "   Thunder Client ID: ${THUNDER_CLIENT_ID:-[NOT SET]}"

echo "Starting nginx on port 3000..."
exec "$@"
