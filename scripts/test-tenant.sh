#!/usr/bin/env bash
# End-to-end: throwaway tenant create + DELETE.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE="$ROOT/examples/tenant"
ENDPOINT="${VIRTFOUNDRY_ENDPOINT:-http://virtfoundry.homelab}"
USER="${VIRTFOUNDRY_USERNAME:-root}"
PASS="${VIRTFOUNDRY_PASSWORD:-virtfoundry}"
unset VIRTFOUNDRY_TENANT_ID
SUFFIX="$(date +%s | tail -c 6)"
TENANT_SLUG="${TENANT_SLUG:-tf-e2e-${SUFFIX}}"
TENANT_NAME="${TENANT_NAME:-TF E2E ${SUFFIX}}"

cleanup() {
  local code=$?
  if [[ $code -ne 0 && -d "$EXAMPLE" && -f "$EXAMPLE/terraform.tfstate" ]]; then
    echo "==> terraform destroy (cleanup)"
    (cd "$EXAMPLE" && terraform destroy -auto-approve -input=false \
      -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS" \
      -var="tenant_name=$TENANT_NAME" -var="tenant_slug=$TENANT_SLUG") || true
  fi
}
trap cleanup EXIT

echo "==> Build provider"
make -C "$ROOT" build

TOKEN="$(curl -sf -X POST "$ENDPOINT/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"password\":\"$PASS\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')"

export TF_CLI_CONFIG_FILE="$ROOT/examples/provider/.terraformrc"
cd "$EXAMPLE"
rm -f terraform.tfstate terraform.tfstate.backup
terraform init -input=false

VARS=(
  -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS"
  -var="tenant_name=$TENANT_NAME" -var="tenant_slug=$TENANT_SLUG"
)

echo "==> terraform apply"
terraform apply -auto-approve -input=false "${VARS[@]}"

TENANT_ID="$(terraform output -raw tenant_id)"
[[ -n "$TENANT_ID" && "$TENANT_ID" != "null" ]] || { echo "FAIL: empty tenant_id"; exit 1; }
echo "  ok tenant_id=$TENANT_ID slug=$TENANT_SLUG"

echo "==> terraform destroy"
terraform destroy -auto-approve -input=false "${VARS[@]}"

FOUND="$(curl -sf "$ENDPOINT/api/v1/tenants" -H "Authorization: Bearer $TOKEN" \
  | python3 -c "import sys,json; ids=[t['id'] for t in (json.load(sys.stdin).get('tenants') or [])]; print('yes' if '$TENANT_ID' in ids else 'no')")"
if [[ "$FOUND" == "yes" ]]; then
  echo "FAIL: tenant $TENANT_ID still listed after destroy"
  exit 1
fi
echo "  ok tenant deleted from API"

trap - EXIT
echo "==> OK — tenant apply/destroy passed"
