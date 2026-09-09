---
page_title: "virtfoundry_service_offering Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a platform service offering. Requires root credentials.
---

# virtfoundry_service_offering (Resource)

Creates a CPU/memory catalog entry. Requires **root** credentials (no `tenant_id` on the provider). Changing `cpu` / `memory_mi` on the offering does **not** resize existing VMs — patch `virtfoundry_vm.service_offering_id` instead.

## Example Usage

```hcl
provider "virtfoundry" {
  endpoint = "https://virtfoundry.example.com"
  username = var.root_username
  password = var.root_password
}

resource "virtfoundry_service_offering" "xl" {
  name          = "xl"
  display_name  = "XL"
  cpu           = 4
  memory_mi     = 8192
  dedicated_cpu = true
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Offering slug. Forces replacement. |
| `cpu` | Number | yes | vCPU count. |
| `memory_mi` | Number | yes | Memory in MiB. |
| `display_name` | String | no | Human-readable name. Defaults to `name`. |
| `dedicated_cpu` | Boolean | no | Guaranteed CPU (request equals limit). |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Offering UUID. |
| `state` | Catalog state (`Active`). |

## Import

```shell
terraform import virtfoundry_service_offering.xl <offering_id>
```
