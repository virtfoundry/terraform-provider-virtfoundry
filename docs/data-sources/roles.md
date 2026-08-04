---
page_title: "virtfoundry_roles Data Source - virtfoundry"
subcategory: ""
description: |-
  Lists IAM roles in a tenant.
---

# virtfoundry_roles (Data Source)

Lists tenant IAM roles and their permissions.

## Example Usage

```hcl
data "virtfoundry_roles" "all" {}

output "custom_roles" {
  value = [
    for r in data.virtfoundry_roles.all.roles : r.name
    if !r.is_system
  ]
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `roles` | List of objects with `id`, `name`, `description`, `permissions`, `is_system`. |
