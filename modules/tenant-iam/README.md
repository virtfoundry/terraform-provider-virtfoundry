# tenant-iam

Manages **tenant-scoped IAM**: custom roles, users, and API keys. Does not create tenants — use [`../tenant`](../tenant) with root credentials first.

## Usage

```hcl
provider "virtfoundry" {
  alias     = "tenant"
  endpoint  = var.endpoint
  api_key   = var.root_api_key
  tenant_id = module.acme.id
}

module "acme_iam" {
  source = "../../modules/tenant-iam"

  providers = {
    virtfoundry = virtfoundry.tenant
  }

  tenant_id = module.acme.id

  roles = {
    operator = {
      description = "Deploy VMs and manage networking"
      permissions = [] # empty uses built-in operator preset when key matches
    }
    custom_viewer = {
      permissions = ["vms:read", "networks:read"]
    }
  }

  users = {
    alice = {
      username  = "alice"
      password  = var.alice_password
      role_name = "operator"
    }
  }
}
```

## Permission presets

When `roles.<name>.permissions` is empty, the module tries presets for keys `admin`, `operator`, and `viewer` (aligned with VirtFoundry core defaults). Otherwise set `permissions` explicitly.
