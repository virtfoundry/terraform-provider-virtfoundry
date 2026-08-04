---
page_title: "virtfoundry_users Data Source - virtfoundry"
subcategory: ""
description: |-
  Lists IAM users in a tenant.
---

# virtfoundry_users (Data Source)

Lists tenant IAM users.

## Example Usage

```hcl
data "virtfoundry_users" "all" {}

output "usernames" {
  value = [for u in data.virtfoundry_users.all.users : u.username]
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `users` | List of objects with `id`, `username`, `email`, `role`, `state`. |
