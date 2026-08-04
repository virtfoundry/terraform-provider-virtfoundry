---
page_title: "virtfoundry_user Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry tenant IAM user.
---

# virtfoundry_user (Resource)

Manages a tenant IAM user. Use a tenant-scoped provider or the [`tenant-iam` module](https://github.com/virtfoundry/terraform-provider-virtfoundry/tree/main/modules/tenant-iam).

## Example Usage

```hcl
resource "virtfoundry_role" "operator" {
  name        = "operator"
  description = "Deploy and manage infrastructure"
  permissions = ["vms:*", "networks:*", "vpcs:*"]
}

resource "virtfoundry_user" "alice" {
  username  = "alice"
  password  = var.alice_password
  email     = "alice@example.com"
  role_name = virtfoundry_role.operator.name
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `username` | String | yes | Login username. Forces replacement. |
| `password` | String | yes | Initial password (sensitive). Forces replacement. |
| `email` | String | no | Email address. |
| `role_id` | String | no | Role UUID. |
| `role_name` | String | no | Role name when `role_id` is omitted. |
| `state` | String | no | User state. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | User UUID. |
| `role` | Assigned role name. |

## Import

```shell
terraform import virtfoundry_user.alice <tenant_id>/<user_id>
```
