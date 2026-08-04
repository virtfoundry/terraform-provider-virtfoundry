---
page_title: "virtfoundry_networks Data Source - virtfoundry"
subcategory: ""
description: |-
  Lists networks in a tenant.
---

# virtfoundry_networks (Data Source)

Lists tenant networks and subnets.

## Example Usage

```hcl
data "virtfoundry_networks" "all" {}

output "network_cidrs" {
  value = { for n in data.virtfoundry_networks.all.networks : n.name => n.cidr }
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `networks` | List of objects with `id`, `name`, `cidr`, `vpc_id`, `state`. |
