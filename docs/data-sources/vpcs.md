---
page_title: "virtfoundry_vpcs Data Source - virtfoundry"
subcategory: ""
description: |-
  Lists VPCs in a tenant.
---

# virtfoundry_vpcs (Data Source)

Lists all VPCs in the configured tenant.

## Example Usage

```hcl
data "virtfoundry_vpcs" "all" {}

output "vpc_names" {
  value = [for v in data.virtfoundry_vpcs.all.vpcs : v.name]
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `vpcs` | List of objects with `id`, `name`, `cidr`, `state`. |
