---
page_title: "virtfoundry_service_offering Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry service offering (CPU/memory catalog entry).
---

# virtfoundry_service_offering (Resource)

Manages a platform-wide service offering. Requires **root** credentials (provider without `tenant_id`).

## Example Usage

```hcl
provider "virtfoundry" {
  alias    = "root"
  endpoint = var.endpoint
  username = var.username
  password = var.password
}

resource "virtfoundry_service_offering" "medium" {
  provider     = virtfoundry.root
  name         = "medium"
  display_name = "Medium (2 vCPU / 2 GiB)"
  cpu          = 2
  memory_mi    = 2048
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Unique slug. Forces replacement. |
| `display_name` | String | yes | Human-readable label. |
| `cpu` | Number | yes | vCPU count. |
| `memory_mi` | Number | yes | Memory in MiB. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Offering UUID. |
| `state` | `Active` or `Inactive` (soft-deleted). |

## Import

```shell
terraform import virtfoundry_service_offering.medium <offering_id>
```
