#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${OPUSDNS_API_KEY:-}" ]]; then
  cat >&2 <<'EOF'
OPUSDNS_API_KEY is required.

Example:
  OPUSDNS_API_KEY="opk_..." ./scripts/integration-test.sh

The tests run against the sandbox (https://sandbox.opusdns.com) unless
OPUSDNS_API_ENDPOINT says otherwise, so use a sandbox API key.

To include the DNS write lifecycle, also set OPUSDNS_INTEGRATION_ZONE to a unique disposable zone name.
EOF
  exit 2
fi

api_key="${OPUSDNS_API_KEY}"
if [[ "${api_key}" == "opk_..." || "${api_key}" == *"your_api_key"* || "${api_key}" == *"your_preview_key"* ]]; then
  echo "OPUSDNS_API_KEY still looks like a placeholder. Set it to a real API key for ${OPUSDNS_API_ENDPOINT:-the sandbox}." >&2
  exit 2
fi
if [[ "${api_key}" == Bearer\ * ]]; then
  echo "OPUSDNS_API_KEY must be the raw API key value, not a Bearer authorization header." >&2
  exit 2
fi
if [[ "${api_key}" == \"* || "${api_key}" == *\" || "${api_key}" == \'* || "${api_key}" == *\' ]]; then
  echo "OPUSDNS_API_KEY appears to include quote characters. Remove the literal quotes from the value." >&2
  exit 2
fi

# Default to the sandbox, like docker-compose.yml. The library itself defaults
# to production, so the endpoint is always exported explicitly.
endpoint="${OPUSDNS_API_ENDPOINT:-https://sandbox.opusdns.com}"
if [[ "${endpoint}" != http://* && "${endpoint}" != https://* ]]; then
  endpoint="https://${endpoint}"
fi
export OPUSDNS_API_ENDPOINT="${endpoint}"

host="${endpoint#*://}"
host="${host%%/*}"
host="${host%%:*}"
if [[ "${host}" == "api.opusdns.com" && "${OPUSDNS_INTEGRATION_ALLOW_PRODUCTION:-}" != "1" ]]; then
  echo "Refusing to run integration tests against production (${endpoint})." >&2
  echo "Use the sandbox, or set OPUSDNS_INTEGRATION_ALLOW_PRODUCTION=1 if you really mean production." >&2
  exit 2
fi

if [[ -n "${OPUSDNS_INTEGRATION_ZONE:-}" ]]; then
  echo "Running OpusDNS real API integration tests against ${endpoint}."
  echo "Write lifecycle enabled for disposable zone: ${OPUSDNS_INTEGRATION_ZONE}"
else
  echo "Running OpusDNS real API read-only integration tests against ${endpoint}."
  echo "Set OPUSDNS_INTEGRATION_ZONE to a unique disposable zone name to include create/update/delete DNS checks."
fi
echo "API key shape: prefix='${api_key:0:4}' length=${#api_key} characters"

go test -tags=integration -count=1 -run '^TestRealAPI' -v ./opusdns "$@"
