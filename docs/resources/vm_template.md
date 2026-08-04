---
page_title: "virtfoundry_vm_template Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry VM template (container disk or ISO).
---

# virtfoundry_vm_template (Resource)

Defines a reusable VM template from a container image or ISO volume.

## Example Usage

```hcl
resource "virtfoundry_vm_template" "ubuntu" {
  name         = "ubuntu-2204"
  display_name = "Ubuntu 22.04 LTS"
  image        = "quay.io/containerdisks/ubuntu:22.04"
  source_type  = "container"
  os_type      = "linux"
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Template slug. Forces replacement. |
| `image` | String | yes | Container image reference or ISO source. Forces replacement. |
| `display_name` | String | no | Human-readable name. |
| `description` | String | no | Template description. |
| `source_type` | String | no | `container` or `iso`. Forces replacement. |
| `os_type` | String | no | Operating system type. |
| `cloud_init_user_data` | String | no | Cloud-init user data. |
| `iso_volume_id` | String | no | ISO volume UUID (for `iso` templates). |
| `iso_size_gi` | Number | no | ISO size in GiB. |
| `boot_disk_size_gi` | Number | no | Boot disk size in GiB. |
| `storage_class` | String | no | Storage class for boot disk. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Template UUID. |
| `state` | Template state. |
| `hypervisor` | Hypervisor type. |

## Import

```shell
terraform import virtfoundry_vm_template.ubuntu <tenant_id>/<template_id>
```
