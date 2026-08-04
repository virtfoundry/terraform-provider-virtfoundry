---
page_title: "virtfoundry_role Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry tenant IAM role and permissions.
---

# virtfoundry_role (Resource)

Defines a custom IAM role with fine-grained permissions for tenant users.

## Example Usage

```hcl
resource "virtfoundry_role" "vm_viewer" {
  name        = "vm-viewer"
  description = "Read-only access to VMs and networking"
  permissions = [
    "vms:read",
    "networks:read",
    "vpcs:read",
  ]
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Role name. Forces replacement. |
| `description` | String | no | Role description. |
| `permissions` | List(String) | no | Permission strings (e.g. `vms:read`, `vms:*`). |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Role UUID. |
| `is_system` | Whether this is a built-in system role. |

## Import

```shell
terraform import virtfoundry_role.vm_viewer <tenant_id>/<role_id>
```
