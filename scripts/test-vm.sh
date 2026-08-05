#!/usr/bin/env bash
# End-to-end test: apply + destroy a Cirros VM via Terraform.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE="$ROOT/examples/vm"
ENDPOINT="${VIRTFOUNDRY_ENDPOINT:-http://virtfoundry.homelab}"
USER="${VIRTFOUNDRY_USERNAME:-root}"
PASS="${VIRTFOUNDRY_PASSWORD:-virtfoundry}"
VM_NAME="${VM_NAME:-tf-test-$(date +%s)}"

echo "==> Build provider"
make -C "$ROOT" build

echo "==> Resolve tenant / catalog IDs from API"
TOKEN="$(curl -sf -X POST "$ENDPOINT/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"password\":\"$PASS\"}" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')"

TENANT_ID="$(curl -sf "$ENDPOINT/api/v1/tenants" -H "Authorization: Bearer $TOKEN" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["tenants"][0]["id"])')"

OFFERING_NAME="${OFFERING_NAME:-small}"
TEMPLATE_NAME="${TEMPLATE_NAME:-ubuntu-2204}"

SG_ID="$(curl -sf "$ENDPOINT/api/v1/security-groups" -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: $TENANT_ID" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin)["security_groups"][0]["id"])')"

echo "    tenant=$TENANT_ID template=$TEMPLATE_NAME offering=$OFFERING_NAME sg=$SG_ID vm=$VM_NAME"

export TF_CLI_CONFIG_FILE="$ROOT/examples/provider/.terraformrc"

cd "$EXAMPLE"
terraform init -input=false

echo "==> terraform apply"
terraform apply -auto-approve -input=false \
  -var="endpoint=$ENDPOINT" \
  -var="username=$USER" \
  -var="password=$PASS" \
  -var="tenant_id=$TENANT_ID" \
  -var="template_name=$TEMPLATE_NAME" \
  -var="service_offering_name=$OFFERING_NAME" \
  -var="security_group_id=$SG_ID" \
  -var="vm_name=$VM_NAME"

terraform output

echo "==> terraform destroy"
terraform destroy -auto-approve -input=false \
  -var="endpoint=$ENDPOINT" \
  -var="username=$USER" \
  -var="password=$PASS" \
  -var="tenant_id=$TENANT_ID" \
  -var="template_name=$TEMPLATE_NAME" \
  -var="service_offering_name=$OFFERING_NAME" \
  -var="security_group_id=$SG_ID" \
  -var="vm_name=$VM_NAME"

echo "==> OK"
