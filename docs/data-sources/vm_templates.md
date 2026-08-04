---
page_title: "virtfoundry_vm_templates Data Source - virtfoundry"
subcategory: ""
description: |-
  Lists VM templates available in a tenant.
---

# virtfoundry_vm_templates (Data Source)

Returns VM templates registered in the tenant.

## Example Usage

```hcl
data "virtfoundry_vm_templates" "catalog" {}

locals {
  ubuntu = one([
    for t in data.virtfoundry_vm_templates.catalog.templates : t
    if t.name == "ubuntu-2204"
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
| `templates` | List of objects with `id`, `name`, `display_name`, `source_type`, `state`. |
