---
page_title: "virtfoundry_service_offerings Data Source - virtfoundry"
subcategory: ""
description: |-
  Lists VirtFoundry service offerings (CPU/memory catalog).
---

# virtfoundry_service_offerings (Data Source)

Returns the platform service offering catalog for VM sizing.

## Example Usage

```hcl
data "virtfoundry_service_offerings" "catalog" {}

output "offerings" {
  value = data.virtfoundry_service_offerings.catalog.offerings
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `offerings` | List of objects with `id`, `name`, `cpu`, `memory_mi`. |
