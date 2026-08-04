#!/usr/bin/env bash
# End-to-end test: VPC + network + security group + SSH key + VM (apply + destroy).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE="$ROOT/examples/full-stack"
ENDPOINT="${VIRTFOUNDRY_ENDPOINT:-http://virtfoundry.homelab}"
USER="${VIRTFOUNDRY_USERNAME:-root}"
PASS="${VIRTFOUNDRY_PASSWORD:-virtfoundry}"

SUFFIX="$(date +%s)"
VPC_NAME="${VPC_NAME:-tf-vpc-${SUFFIX}}"
NET_NAME="${NET_NAME:-tf-net-${SUFFIX}}"
SG_NAME="${SG_NAME:-tf-sg-${SUFFIX}}"
SSH_NAME="${SSH_NAME:-tf-ssh-${SUFFIX}}"
VM_NAME="${VM_NAME:-tf-vm-${SUFFIX}}"
EXPOSE_SSH="${EXPOSE_SSH:-false}"
# Spread tenants across 10.42–10.241 to reduce CIDR collisions between runs.
VPC_OCTET=$((42 + SUFFIX % 200))
VPC_CIDR="${VPC_CIDR:-10.${VPC_OCTET}.0.0/16}"

cleanup() {
  local code=$?
  if [[ $code -ne 0 && -n "${TOKEN:-}" && -n "${TENANT_ID:-}" ]]; then
    echo "==> ERROR (exit $code) — cleanup orphaned VM if any"
    curl -sf -X POST "$ENDPOINT/api/v1/vms/delete" \
      -H "Authorization: Bearer $TOKEN" \
      -H "X-Tenant-ID: $TENANT_ID" \
      -H 'Content-Type: application/json' \
      -d "{\"name\":\"$VM_NAME\"}" >/dev/null 2>&1 || true
  fi
  if [[ $code -ne 0 && -d "$EXAMPLE" && -f "$EXAMPLE/terraform.tfstate" ]]; then
    echo "==> terraform destroy (cleanup)"
    (cd "$EXAMPLE" && terraform destroy -auto-approve -input=false \
      -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS" \
      -var="tenant_id=${TENANT_ID:-}" -var="vpc_name=$VPC_NAME" -var="vpc_cidr=$VPC_CIDR" \
      -var="network_name=$NET_NAME" -var="security_group_name=$SG_NAME" \
      -var="ssh_key_name=$SSH_NAME" -var="vm_name=$VM_NAME" -var="expose_ssh=$EXPOSE_SSH") || true
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

echo "    tenant=$TENANT_ID"
echo "    vpc=$VPC_NAME cidr=$VPC_CIDR net=$NET_NAME sg=$SG_NAME ssh=$SSH_NAME vm=$VM_NAME"

export TF_CLI_CONFIG_FILE="$ROOT/examples/provider/.terraformrc"

cd "$EXAMPLE"
rm -f terraform.tfstate terraform.tfstate.backup
terraform init -input=false

echo "==> terraform apply (full stack)"
terraform apply -auto-approve -input=false \
  -var="endpoint=$ENDPOINT" \
  -var="username=$USER" \
  -var="password=$PASS" \
  -var="tenant_id=$TENANT_ID" \
  -var="vpc_name=$VPC_NAME" \
  -var="vpc_cidr=$VPC_CIDR" \
  -var="network_name=$NET_NAME" \
  -var="security_group_name=$SG_NAME" \
  -var="ssh_key_name=$SSH_NAME" \
  -var="vm_name=$VM_NAME" \
  -var="expose_ssh=$EXPOSE_SSH"

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

echo "==> Validate outputs (network + identity)"
assert_output vpc_id
assert_output network_id
assert_output security_group_id
assert_output ssh_key_id
assert_output vm_id

VM_IP="$(terraform output -raw vm_ip 2>/dev/null || true)"
if [[ -z "$VM_IP" || "$VM_IP" == "null" ]]; then
  echo "  warn vm_ip not assigned yet (common on private networks / Starting state)"
  echo "==> terraform refresh (wait for VM IP, up to 3m)"
  deadline=$((SECONDS + 180))
  while [[ $SECONDS -lt $deadline ]]; do
    terraform refresh -input=false \
      -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS" \
      -var="tenant_id=$TENANT_ID" -var="vpc_name=$VPC_NAME" -var="vpc_cidr=$VPC_CIDR" \
      -var="network_name=$NET_NAME" -var="security_group_name=$SG_NAME" \
      -var="ssh_key_name=$SSH_NAME" -var="vm_name=$VM_NAME" -var="expose_ssh=$EXPOSE_SSH" >/dev/null
    VM_IP="$(terraform output -raw vm_ip 2>/dev/null || true)"
    if [[ -n "$VM_IP" && "$VM_IP" != "null" ]]; then
      echo "  ok vm_ip=$VM_IP"
      break
    fi
    sleep 10
  done
  if [[ -z "$VM_IP" || "$VM_IP" == "null" ]]; then
    echo "  warn vm_ip still empty — continuing (network resources validated)"
  fi
else
  echo "  ok vm_ip=$VM_IP"
fi

SSH_EXPOSED="$(terraform output -raw ssh_exposed 2>/dev/null || echo false)"
SSH_PORT="$(terraform output -raw ssh_node_port 2>/dev/null || echo 0)"
echo "  info ssh_exposed=$SSH_EXPOSED ssh_node_port=$SSH_PORT"

echo "==> Verify state (expected resources)"
terraform state list | tee /tmp/tf-full-stack-state.txt
for resource in \
  'virtfoundry_vpc.main' \
  'virtfoundry_network.private' \
  'virtfoundry_security_group.ssh' \
  'virtfoundry_ssh_key.admin' \
  'virtfoundry_vm.app' \
  'data.virtfoundry_service_offerings.catalog' \
  'data.virtfoundry_vm_templates.catalog'; do
  if ! grep -qx "$resource" /tmp/tf-full-stack-state.txt; then
    echo "FAIL: missing $resource in state"
    exit 1
  fi
  echo "  ok state contains $resource"
done

echo "==> terraform plan (expect no changes)"
if ! terraform plan -detailed-exitcode -input=false \
  -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS" \
  -var="tenant_id=$TENANT_ID" -var="vpc_name=$VPC_NAME" -var="vpc_cidr=$VPC_CIDR" \
  -var="network_name=$NET_NAME" -var="security_group_name=$SG_NAME" \
  -var="ssh_key_name=$SSH_NAME" -var="vm_name=$VM_NAME" >/tmp/tf-plan.txt; then
  echo "FAIL: plan wants changes after apply"
  cat /tmp/tf-plan.txt
  exit 1
fi
echo "  ok plan is clean"

terraform output

echo "==> terraform destroy"
terraform destroy -auto-approve -input=false \
  -var="endpoint=$ENDPOINT" \
  -var="username=$USER" \
  -var="password=$PASS" \
  -var="tenant_id=$TENANT_ID" \
  -var="vpc_name=$VPC_NAME" \
  -var="vpc_cidr=$VPC_CIDR" \
  -var="network_name=$NET_NAME" \
  -var="security_group_name=$SG_NAME" \
  -var="ssh_key_name=$SSH_NAME" \
  -var="vm_name=$VM_NAME" \
  -var="expose_ssh=$EXPOSE_SSH"

trap - EXIT
echo "==> OK — full stack apply/destroy passed"
