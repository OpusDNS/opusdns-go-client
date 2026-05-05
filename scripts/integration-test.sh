#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${OPUSDNS_API_KEY:-}" ]]; then
  cat >&2 <<'EOF'
OPUSDNS_API_KEY is required.

Example:
  OPUSDNS_API_KEY="opk_..." ./scripts/integration-test.sh

To include the DNS write lifecycle, also set OPUSDNS_INTEGRATION_ZONE to a unique disposable zone name.
EOF
  exit 2
fi

api_key="${OPUSDNS_API_KEY}"
if [[ "${api_key}" == "opk_..." || "${api_key}" == *"your_api_key"* || "${api_key}" == *"your_preview_key"* ]]; then
  echo "OPUSDNS_API_KEY still looks like a placeholder. Set it to a real preview1 API key value." >&2
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

endpoint="${OPUSDNS_API_ENDPOINT:-https://api.opusdns.com}"
if [[ "${endpoint}" != http://* && "${endpoint}" != https://* ]]; then
  endpoint="https://${endpoint}"
fi
export OPUSDNS_API_ENDPOINT="${endpoint}"

if [[ -n "${OPUSDNS_INTEGRATION_ZONE:-}" ]]; then
  echo "Running OpusDNS real API integration tests against ${endpoint}."
  echo "Write lifecycle enabled for disposable zone: ${OPUSDNS_INTEGRATION_ZONE}"
else
  echo "Running OpusDNS real API read-only integration tests against ${endpoint}."
  echo "Set OPUSDNS_INTEGRATION_ZONE to a unique disposable zone name to include create/update/delete DNS checks."
fi
echo "API key shape: prefix='${api_key:0:4}' length=${#api_key} characters"

go test -tags=integration -count=1 -run '^TestRealAPI' -v ./opusdns "$@"
