---
page_title: "virtfoundry_security_groups Data Source - virtfoundry"
subcategory: ""
description: |-
  Lists security groups in a tenant.
---

# virtfoundry_security_groups (Data Source)

Lists security groups and their metadata.

## Example Usage

```hcl
data "virtfoundry_security_groups" "all" {}

output "default_sg" {
  value = one([
    for sg in data.virtfoundry_security_groups.all.security_groups : sg
    if sg.name == "default"
  ])
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `security_groups` | List of objects with `id`, `name`, `description`, `vpc_id`. |
