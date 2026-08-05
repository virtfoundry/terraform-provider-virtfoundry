#!/usr/bin/env bash
# E2E: phases 1–3 resources (volume, attachment, service offering, template, VM resize).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE="$ROOT/examples/phases"
ENDPOINT="${VIRTFOUNDRY_ENDPOINT:-http://virtfoundry.homelab}"
USER="${VIRTFOUNDRY_USERNAME:-root}"
PASS="${VIRTFOUNDRY_PASSWORD:-virtfoundry}"

SUFFIX="$(date +%s)"
OFFERING_NAME="${OFFERING_NAME:-tf-off-${SUFFIX}}"
TEMPLATE_NAME="${TEMPLATE_NAME:-tf-tpl-${SUFFIX}}"
VOLUME_NAME="${VOLUME_NAME:-tf-vol-${SUFFIX}}"
VM_NAME="${VM_NAME:-tf-vm-${SUFFIX}}"

common_vars=(
  -var="endpoint=$ENDPOINT"
  -var="username=$USER"
  -var="password=$PASS"
  -var="offering_name=$OFFERING_NAME"
  -var="template_name=$TEMPLATE_NAME"
  -var="volume_name=$VOLUME_NAME"
  -var="vm_name=$VM_NAME"
)

cleanup() {
  local code=$?
  if [[ $code -ne 0 && -d "$EXAMPLE" && -f "$EXAMPLE/terraform.tfstate" ]]; then
    echo "==> terraform destroy (cleanup)"
    (cd "$EXAMPLE" && terraform destroy -auto-approve -input=false \
      "${common_vars[@]}" -var="tenant_id=${TENANT_ID:-}" -var="use_custom_offering=true" \
      -var="desired_state=stopped") || true
  fi
}
trap cleanup EXIT

echo "==> Build provider"
make -C "$ROOT" build

echo "==> Resolve tenant from API"
TOKEN="$(curl -sf -X POST "$ENDPOINT/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"password\":\"$PASS\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')"

TENANT_ID="$(curl -sf "$ENDPOINT/api/v1/tenants" -H "Authorization: Bearer $TOKEN" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["tenants"][0]["id"])')"

echo "    tenant=$TENANT_ID offering=$OFFERING_NAME template=$TEMPLATE_NAME vm=$VM_NAME"

export TF_CLI_CONFIG_FILE="$ROOT/examples/provider/.terraformrc"

cd "$EXAMPLE"
rm -f terraform.tfstate terraform.tfstate.backup
terraform init -input=false

echo "==> terraform apply (baseline offering + volume attach)"
terraform apply -auto-approve -input=false \
  "${common_vars[@]}" \
  -var="tenant_id=$TENANT_ID" \
  -var="use_custom_offering=false" \
  -var="desired_state=running"

assert_output() {
  local name=$1
  local val
  val="$(terraform output -raw "$name" 2>/dev/null || true)"
  if [[ -z "$val" || "$val" == "null" ]]; then
    echo "FAIL: output '$name' is empty"
    terraform output
    exit 1
  fi
  echo "  ok $name=$val"
}

echo "==> Validate outputs"
assert_output offering_id
assert_output template_id
assert_output volume_id
assert_output vm_id
assert_output attachment_id

echo "==> Verify state (phases resources)"
terraform state list | tee /tmp/tf-phases-state.txt
for resource in \
  'virtfoundry_service_offering.custom' \
  'virtfoundry_vm_template.app' \
  'virtfoundry_volume.data' \
  'virtfoundry_vm.app' \
  'virtfoundry_vm_volume_attachment.data'; do
  if ! grep -qx "$resource" /tmp/tf-phases-state.txt; then
    echo "FAIL: missing $resource in state"
    exit 1
  fi
  echo "  ok state contains $resource"
done

echo "==> Stop VM and resize to custom offering"
terraform apply -auto-approve -input=false \
  "${common_vars[@]}" \
  -var="tenant_id=$TENANT_ID" \
  -var="use_custom_offering=true" \
  -var="desired_state=stopped"

VM_CPU="$(terraform output -raw vm_cpu)"
VM_MEM="$(terraform output -raw vm_memory_mi)"
echo "  info vm_cpu=$VM_CPU vm_memory_mi=$VM_MEM"
if [[ "$VM_CPU" != "2" ]]; then
  echo "FAIL: expected cpu=2 after resize, got $VM_CPU"
  exit 1
fi
echo "  ok resize applied"

echo "==> terraform plan (expect no changes)"
if ! terraform plan -detailed-exitcode -input=false \
  "${common_vars[@]}" \
  -var="tenant_id=$TENANT_ID" \
  -var="use_custom_offering=true" \
  -var="desired_state=stopped" >/tmp/tf-phases-plan.txt; then
  echo "FAIL: plan wants changes after resize"
  cat /tmp/tf-phases-plan.txt
  exit 1
fi
echo "  ok plan is clean"

echo "==> terraform destroy"
terraform destroy -auto-approve -input=false \
  "${common_vars[@]}" \
  -var="tenant_id=$TENANT_ID" \
  -var="use_custom_offering=true" \
  -var="desired_state=stopped"

trap - EXIT
echo "==> OK — phases 1–3 terraform e2e passed"
