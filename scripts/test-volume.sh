#!/usr/bin/env bash
# End-to-end: Cirros VM + volume + attachment (apply, clean plan, destroy deletes PVC).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE="$ROOT/examples/volume"
ENDPOINT="${VIRTFOUNDRY_ENDPOINT:-http://virtfoundry.homelab}"
USER="${VIRTFOUNDRY_USERNAME:-root}"
PASS="${VIRTFOUNDRY_PASSWORD:-virtfoundry}"
SUFFIX="$(date +%s)"
VM_NAME="${VM_NAME:-tf-vol-vm-${SUFFIX}}"
VOLUME_NAME="${VOLUME_NAME:-tf-vol-${SUFFIX}}"

cleanup() {
  local code=$?
  if [[ $code -ne 0 && -d "$EXAMPLE" && -f "$EXAMPLE/terraform.tfstate" ]]; then
    echo "==> terraform destroy (cleanup)"
    (cd "$EXAMPLE" && terraform destroy -auto-approve -input=false \
      -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS" \
      -var="tenant_id=${TENANT_ID:-}" -var="template_id=${TEMPLATE_ID:-}" \
      -var="service_offering_id=${OFFERING_ID:-}" -var="security_group_id=${SG_ID:-}" \
      -var="vm_name=$VM_NAME" -var="volume_name=$VOLUME_NAME") || true
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

OFFERING_ID="$(curl -sf "$ENDPOINT/api/v1/service-offerings" -H "Authorization: Bearer $TOKEN" \
  | python3 -c 'import sys,json; print(next(o["id"] for o in json.load(sys.stdin)["service_offerings"] if o["name"]=="small"))')"

TEMPLATE_ID="$(curl -sf "$ENDPOINT/api/v1/vm-templates" -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT_ID" \
  | python3 -c 'import sys,json; print(next(t["id"] for t in json.load(sys.stdin)["vm_templates"] if t["name"]=="cirros"))')"

SG_ID="$(curl -sf "$ENDPOINT/api/v1/security-groups" -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT_ID" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["security_groups"][0]["id"])')"

echo "    tenant=$TENANT_ID vm=$VM_NAME volume=$VOLUME_NAME"

export TF_CLI_CONFIG_FILE="$ROOT/examples/provider/.terraformrc"
cd "$EXAMPLE"
rm -f terraform.tfstate terraform.tfstate.backup
terraform init -input=false

VARS=(
  -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS"
  -var="tenant_id=$TENANT_ID" -var="template_id=$TEMPLATE_ID"
  -var="service_offering_id=$OFFERING_ID" -var="security_group_id=$SG_ID"
  -var="vm_name=$VM_NAME" -var="volume_name=$VOLUME_NAME"
)

echo "==> terraform apply"
terraform apply -auto-approve -input=false "${VARS[@]}"

VOLUME_ID="$(terraform output -raw volume_id)"
[[ -n "$VOLUME_ID" && "$VOLUME_ID" != "null" ]] || { echo "FAIL: empty volume_id"; exit 1; }
echo "  ok volume_id=$VOLUME_ID"

echo "==> terraform plan (expect no changes)"
if ! terraform plan -detailed-exitcode -input=false "${VARS[@]}" >/tmp/tf-volume-plan.txt; then
  echo "FAIL: plan wants changes after apply"
  cat /tmp/tf-volume-plan.txt
  exit 1
fi
echo "  ok plan is clean"

echo "==> terraform destroy"
terraform destroy -auto-approve -input=false "${VARS[@]}"

FOUND="$(curl -sf "$ENDPOINT/api/v1/volumes" -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT_ID" \
  | python3 -c "import sys,json; ids=[v['id'] for v in (json.load(sys.stdin).get('volumes') or [])]; print('yes' if '$VOLUME_ID' in ids else 'no')")"
if [[ "$FOUND" == "yes" ]]; then
  echo "FAIL: volume $VOLUME_ID still listed after destroy"
  exit 1
fi
echo "  ok volume deleted from API"

trap - EXIT
echo "==> OK — volume apply/destroy passed"
