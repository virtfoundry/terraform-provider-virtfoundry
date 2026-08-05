---
page_title: "virtfoundry_vm_volume_attachment Resource - virtfoundry"
subcategory: ""
description: |-
  Attaches a block volume to a VM (hot-plug).
---

# virtfoundry_vm_volume_attachment (Resource)

Attaches an existing volume to a VM. Destroy detaches the volume.

## Example Usage

```hcl
resource "virtfoundry_volume" "data" {
  name    = "app-data"
  size_gi = 10
}

resource "virtfoundry_vm" "app" {
  name                = "app"
  template_id         = var.template_id
  service_offering_id = var.offering_id
  public_ip           = true
  security_group_ids  = [var.security_group_id]
}

resource "virtfoundry_vm_volume_attachment" "data" {
  vm_name   = virtfoundry_vm.app.name
  volume_id = virtfoundry_volume.data.id
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `vm_name` | String | yes | Target VM name. Forces replacement. |
| `volume_id` | String | yes | Volume UUID. Forces replacement. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Attachment id `<vm_name>/<volume_id>`. |

## Import

```shell
terraform import virtfoundry_vm_volume_attachment.data <vm_name>/<volume_id>
```
