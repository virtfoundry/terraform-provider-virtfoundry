#!/usr/bin/env bash
# End-to-end: apply Cirros VM, in-place resize small → medium, destroy.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE="$ROOT/examples/vm"
ENDPOINT="${VIRTFOUNDRY_ENDPOINT:-http://virtfoundry.homelab}"
USER="${VIRTFOUNDRY_USERNAME:-root}"
PASS="${VIRTFOUNDRY_PASSWORD:-virtfoundry}"
VM_NAME="${VM_NAME:-tf-test-$(date +%s)}"

cleanup() {
  local code=$?
  if [[ $code -ne 0 && -d "$EXAMPLE" && -f "$EXAMPLE/terraform.tfstate" ]]; then
    echo "==> terraform destroy (cleanup)"
    (cd "$EXAMPLE" && terraform destroy -auto-approve -input=false \
      -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS" \
      -var="tenant_id=${TENANT_ID:-}" -var="template_id=${TEMPLATE_ID:-}" \
      -var="service_offering_id=${SMALL_ID:-${OFFERING_ID:-}}" \
      -var="security_group_id=${SG_ID:-}" -var="vm_name=$VM_NAME" \
      -var="dedicated_cpu=false") || true
  fi
}
trap cleanup EXIT

echo "==> Build provider"
make -C "$ROOT" build

echo "==> Resolve tenant / catalog IDs from API"
TOKEN="$(curl -sf -X POST "$ENDPOINT/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"password\":\"$PASS\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')"

TENANT_ID="$(curl -sf "$ENDPOINT/api/v1/tenants" -H "Authorization: Bearer $TOKEN" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["tenants"][0]["id"])')"

eval "$(curl -sf "$ENDPOINT/api/v1/service-offerings" -H "Authorization: Bearer $TOKEN" \
  | python3 -c '
import sys,json
offs=json.load(sys.stdin)["service_offerings"]
by={o["name"]: o["id"] for o in offs}
print("SMALL_ID="+by["small"])
print("MEDIUM_ID="+by["medium"])
')"

TEMPLATE_ID="$(curl -sf "$ENDPOINT/api/v1/vm-templates" -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT_ID" \
  | python3 -c 'import sys,json; print(next(t["id"] for t in json.load(sys.stdin)["vm_templates"] if t["name"]=="cirros"))')"

SG_ID="$(curl -sf "$ENDPOINT/api/v1/security-groups" -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT_ID" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["security_groups"][0]["id"])')"

echo "    tenant=$TENANT_ID template=$TEMPLATE_ID small=$SMALL_ID medium=$MEDIUM_ID sg=$SG_ID vm=$VM_NAME"

export TF_CLI_CONFIG_FILE="$ROOT/examples/provider/.terraformrc"
cd "$EXAMPLE"
rm -f terraform.tfstate terraform.tfstate.backup
terraform init -input=false

base_vars() {
  local offering=$1
  echo -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS" \
    -var="tenant_id=$TENANT_ID" -var="template_id=$TEMPLATE_ID" \
    -var="service_offering_id=$offering" -var="security_group_id=$SG_ID" \
    -var="vm_name=$VM_NAME" -var="dedicated_cpu=false"
}

echo "==> terraform apply (small)"
# shellcheck disable=SC2046
terraform apply -auto-approve -input=false $(base_vars "$SMALL_ID")

terraform output
VM_ID="$(terraform output -raw vm_id)"

echo "==> terraform plan (expect no changes)"
# shellcheck disable=SC2046
if ! terraform plan -detailed-exitcode -input=false $(base_vars "$SMALL_ID") >/tmp/tf-vm-plan.txt; then
  echo "FAIL: plan wants changes after apply"
  cat /tmp/tf-vm-plan.txt
  exit 1
fi
echo "  ok plan is clean"

echo "==> terraform apply (resize to medium, in-place)"
# shellcheck disable=SC2046
terraform apply -auto-approve -input=false $(base_vars "$MEDIUM_ID") | tee /tmp/tf-vm-resize.txt
if grep -E 'must be replaced|forces replacement|# .* will be replaced' /tmp/tf-vm-resize.txt; then
  echo "FAIL: resize replaced the VM"
  exit 1
fi
VM_ID2="$(terraform output -raw vm_id)"
if [[ "$VM_ID" != "$VM_ID2" ]]; then
  echo "FAIL: VM id changed ($VM_ID -> $VM_ID2)"
  exit 1
fi
echo "  ok in-place resize vm_id=$VM_ID"

echo "==> terraform destroy"
# shellcheck disable=SC2046
terraform destroy -auto-approve -input=false $(base_vars "$MEDIUM_ID")

trap - EXIT
echo "==> OK"
