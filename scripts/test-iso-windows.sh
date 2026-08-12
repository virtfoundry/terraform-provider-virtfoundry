#!/usr/bin/env bash
# E2E: Windows ISO template import (CDI) via Terraform; optional VM deploy.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE="$ROOT/examples/iso-windows"
ENDPOINT="${VIRTFOUNDRY_ENDPOINT:-http://virtfoundry.homelab}"
USER="${VIRTFOUNDRY_USERNAME:-root}"
PASS="${VIRTFOUNDRY_PASSWORD:-virtfoundry}"

SUFFIX="$(date +%s)"
TEMPLATE_NAME="${TEMPLATE_NAME:-tf-win-${SUFFIX}}"
VM_NAME="${VM_NAME:-tf-win-vm-${SUFFIX}}"
# import only by default — Windows ISO download can take 30–45 minutes
CREATE_ISO="${CREATE_ISO:-true}"
DEPLOY_VM="${DEPLOY_VM:-false}"
IMPORT_TIMEOUT_MIN="${IMPORT_TIMEOUT_MIN:-45}"

cleanup() {
  local code=$?
  if [[ $code -ne 0 && -d "$EXAMPLE" && -f "$EXAMPLE/terraform.tfstate" ]]; then
    echo "==> terraform destroy (cleanup)"
    (cd "$EXAMPLE" && terraform destroy -auto-approve -input=false \
      -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS" \
      -var="tenant_id=${TENANT_ID:-}" -var="create_iso_template=$CREATE_ISO" \
      -var="template_name=$TEMPLATE_NAME" -var="deploy_vm=$DEPLOY_VM" \
      -var="vm_name=$VM_NAME" -var="security_group_id=${SG_ID:-}") || true
  fi
}
trap cleanup EXIT

echo "==> Build provider"
make -C "$ROOT" build

echo "==> Resolve tenant + security group"
TOKEN="$(curl -sf -X POST "$ENDPOINT/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"password\":\"$PASS\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')"

TENANT_ID="$(curl -sf "$ENDPOINT/api/v1/tenants" -H "Authorization: Bearer $TOKEN" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["tenants"][0]["id"])')"

SG_ID="$(curl -sf "$ENDPOINT/api/v1/security-groups" -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT_ID" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["security_groups"][0]["id"])')"

echo "    tenant=$TENANT_ID template=$TEMPLATE_NAME create_iso=$CREATE_ISO deploy_vm=$DEPLOY_VM"

if [[ "$CREATE_ISO" == "false" ]]; then
  echo "==> Using platform template windows-server-2022 (skip CDI import)"
  READY="$(curl -sf "$ENDPOINT/api/v1/vm-templates" -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT_ID" \
    | python3 -c 'import sys,json; t=next(x for x in json.load(sys.stdin)["vm_templates"] if x["name"]=="windows-server-2022"); print(t.get("import_state") or "ready")')"
  echo "    platform import_state=$READY"
  if [[ "$READY" != "ready" && -n "$READY" ]]; then
    echo "FAIL: platform windows template not ready ($READY)"
    exit 1
  fi
else
  echo "==> Will import Windows ISO via CDI (timeout ${IMPORT_TIMEOUT_MIN}m — grab coffee)"
fi

export TF_CLI_CONFIG_FILE="$ROOT/examples/provider/.terraformrc"

cd "$EXAMPLE"
rm -f terraform.tfstate terraform.tfstate.backup
terraform init -input=false

echo "==> terraform apply"
terraform apply -auto-approve -input=false \
  -var="endpoint=$ENDPOINT" \
  -var="username=$USER" \
  -var="password=$PASS" \
  -var="tenant_id=$TENANT_ID" \
  -var="create_iso_template=$CREATE_ISO" \
  -var="template_name=$TEMPLATE_NAME" \
  -var="import_wait_timeout_minutes=$IMPORT_TIMEOUT_MIN" \
  -var="deploy_vm=$DEPLOY_VM" \
  -var="vm_name=$VM_NAME" \
  -var="security_group_id=$SG_ID"

if [[ "$CREATE_ISO" == "true" ]]; then
  STATE="$(terraform output -raw template_import_state 2>/dev/null || true)"
  echo "  ok template_import_state=$STATE"
  if [[ "$STATE" != "ready" ]]; then
    echo "FAIL: expected import_state=ready, got $STATE"
    exit 1
  fi
fi

terraform output

echo "==> terraform destroy"
terraform destroy -auto-approve -input=false \
  -var="endpoint=$ENDPOINT" \
  -var="username=$USER" \
  -var="password=$PASS" \
  -var="tenant_id=$TENANT_ID" \
  -var="create_iso_template=$CREATE_ISO" \
  -var="template_name=$TEMPLATE_NAME" \
  -var="import_wait_timeout_minutes=$IMPORT_TIMEOUT_MIN" \
  -var="deploy_vm=$DEPLOY_VM" \
  -var="vm_name=$VM_NAME" \
  -var="security_group_id=$SG_ID"

trap - EXIT
echo "==> OK — Windows ISO terraform e2e passed"
