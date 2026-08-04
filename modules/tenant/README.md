# tenant

Creates a VirtFoundry tenant. **Root credentials only** — do not set `tenant_id` on the provider.

## Usage

```hcl
provider "virtfoundry" {
  alias    = "root"
  endpoint = var.endpoint
  username = var.root_username
  password = var.root_password
  # no tenant_id
}

module "acme" {
  source = "../../modules/tenant"

  providers = {
    virtfoundry = virtfoundry.root
  }

  name           = "Acme Corp"
  slug           = "acme"
  admin_password = var.tenant_admin_password
}
```

Pair with [`../tenant-iam`](../tenant-iam) for users and roles inside the tenant.
