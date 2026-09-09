---
page_title: "virtfoundry_volume_attachment Resource - virtfoundry"
subcategory: ""
description: |-
  Attaches a data volume to a VM (hot-plug).
---

# virtfoundry_volume_attachment (Resource)

Hot-plugs a volume onto a VM. Destroy detaches the volume. Do not attach the same volume via `virtfoundry_vm.data_volume_id`.

## Example Usage

```hcl
resource "virtfoundry_volume" "data" {
  name    = "app-data"
  size_gi = 20
}

resource "virtfoundry_volume_attachment" "data" {
  vm_name   = virtfoundry_vm.web.name
  volume_id = virtfoundry_volume.data.id
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `vm_name` | String | yes | VM name (slug). Forces replacement. |
| `volume_id` | String | yes | Volume UUID. Forces replacement. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Equals `volume_id`. |

## Import

```shell
terraform import virtfoundry_volume_attachment.data <tenant_id>/<volume_id>
```
