# Windows ISO template example

Registers a Windows Server 2022 eval ISO via CDI and optionally deploys an installer VM.

## Default flow (import only)

```bash
terraform apply \
  -var="tenant_id=<uuid>" \
  -var="security_group_id=<uuid>"
```

Terraform blocks on `virtfoundry_vm_template` until `import_state = ready` (up to 45 minutes).

## Use platform catalog (skip import)

If `windows-server-2022` is already imported on the cluster:

```bash
terraform apply \
  -var="tenant_id=<uuid>" \
  -var="create_iso_template=false" \
  -var="deploy_vm=true" \
  -var="security_group_id=<uuid>"
```

## E2E script

```bash
./scripts/test-iso-windows.sh
# or skip CDI: CREATE_ISO=false DEPLOY_VM=true ./scripts/test-iso-windows.sh
```

Log: `/tmp/tf-iso-windows-e2e.log`
