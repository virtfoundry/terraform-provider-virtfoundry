---
page_title: "virtfoundry_tenant Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry tenant. Requires root credentials.
---

# virtfoundry_tenant (Resource)

Creates a new tenant on the VirtFoundry platform. Requires **root** API credentials (no `tenant_id` on the provider). Tenants cannot be updated or deleted via the API — destroy removes Terraform state only.

## Example Usage

```hcl
provider "virtfoundry" {
  endpoint = "https://virtfoundry.example.com"
  username = var.root_username
  password = var.root_password
}

resource "virtfoundry_tenant" "acme" {
  name           = "Acme Corp"
  slug           = "acme"
  admin_password = var.tenant_admin_password
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Display name for the tenant. |
| `slug` | String | no | URL-safe slug; derived from `name` when omitted. |
| `admin_password` | String | no | Initial tenant admin password (sensitive). |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Tenant UUID. |
| `namespace` | Kubernetes namespace for tenant workloads. |
| `state` | Tenant lifecycle state. |

## Import

```shell
terraform import virtfoundry_tenant.acme <tenant_id>
```
