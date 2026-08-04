# Tenant with IAM example

Bootstraps a new tenant (optional) and configures IAM roles and users using reusable modules.

## What it creates

- `module.tenant` — new tenant via root credentials (when `create_tenant = true`)
- `module.iam` — custom roles and users inside the tenant

Uses two provider aliases:

- `virtfoundry.root` — platform admin (no `tenant_id`)
- `virtfoundry.tenant` — tenant-scoped operations

## Usage

```bash
terraform init
terraform apply \
  -var="endpoint=https://virtfoundry.example.com" \
  -var="root_username=root" \
  -var="root_password=..." \
  -var="create_tenant=true" \
  -var="tenant_name=Acme Corp" \
  -var="tenant_slug=acme" \
  -var="operator_username=operator" \
  -var="operator_password=..."
```

To manage IAM in an existing tenant, set `create_tenant = false` and pass `tenant_id`.

## Outputs

| Output | Description |
|--------|-------------|
| `tenant_id` | Tenant UUID |
| `operator_user_id` | Operator user UUID |
| `operator_role_id` | Operator role UUID |

## Related modules

- [`modules/tenant`](../../modules/tenant/) — tenant creation
- [`modules/tenant-iam`](../../modules/tenant-iam/) — roles and users
