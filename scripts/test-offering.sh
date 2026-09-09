#!/usr/bin/env bash
# End-to-end: root service offering with dedicated_cpu (apply + destroy).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLE="$ROOT/examples/offering"
ENDPOINT="${VIRTFOUNDRY_ENDPOINT:-http://virtfoundry.homelab}"
USER="${VIRTFOUNDRY_USERNAME:-root}"
PASS="${VIRTFOUNDRY_PASSWORD:-virtfoundry}"
unset VIRTFOUNDRY_TENANT_ID
SUFFIX="$(date +%s | tail -c 6)"
OFFERING_NAME="${OFFERING_NAME:-tf-e2e-off-${SUFFIX}}"

cleanup() {
  local code=$?
  if [[ $code -ne 0 && -d "$EXAMPLE" && -f "$EXAMPLE/terraform.tfstate" ]]; then
    echo "==> terraform destroy (cleanup)"
    (cd "$EXAMPLE" && terraform destroy -auto-approve -input=false \
      -var="endpoint=$ENDPOINT" -var="username=$USER" -var="password=$PASS" \
      -var="offering_name=$OFFERING_NAME") || true
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
  -var="offering_name=$OFFERING_NAME"
)

echo "==> terraform apply"
terraform apply -auto-approve -input=false "${VARS[@]}"

DED="$(terraform output -raw offering_dedicated_cpu)"
[[ "$DED" == "true" ]] || { echo "FAIL: dedicated_cpu=$DED"; exit 1; }
OFFERING_ID="$(terraform output -raw offering_id)"
echo "  ok offering_id=$OFFERING_ID dedicated_cpu=true"

IN_CATALOG="$(curl -sf "$ENDPOINT/api/v1/service-offerings" -H "Authorization: Bearer $TOKEN" \
  | python3 -c "import sys,json
offs=json.load(sys.stdin).get('service_offerings') or []
hit=next((o for o in offs if o['id']=='$OFFERING_ID'), None)
print('yes' if hit and hit.get('dedicated_cpu') else 'no')")"
[[ "$IN_CATALOG" == "yes" ]] || { echo "FAIL: offering missing from catalog or dedicated_cpu false"; exit 1; }
echo "  ok catalog lists offering with dedicated_cpu"

echo "==> terraform destroy"
terraform destroy -auto-approve -input=false "${VARS[@]}"

GONE="$(curl -sf "$ENDPOINT/api/v1/service-offerings" -H "Authorization: Bearer $TOKEN" \
  | python3 -c "import sys,json; ids=[o['id'] for o in (json.load(sys.stdin).get('service_offerings') or [])]; print('yes' if '$OFFERING_ID' in ids else 'no')")"
if [[ "$GONE" == "yes" ]]; then
  echo "FAIL: offering $OFFERING_ID still listed after destroy"
  exit 1
fi
echo "  ok offering deleted from API"

trap - EXIT
echo "==> OK — offering apply/destroy passed"
