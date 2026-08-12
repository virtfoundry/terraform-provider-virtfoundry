---
page_title: "virtfoundry_vm_template Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry VM template (container disk or ISO).
---

# virtfoundry_vm_template (Resource)

Defines a reusable VM template from a container image or ISO URL (CDI HTTP import).

## Container disk example

```hcl
resource "virtfoundry_vm_template" "ubuntu" {
  name         = "ubuntu-2204"
  display_name = "Ubuntu 22.04 LTS"
  image        = "quay.io/containerdisks/ubuntu:22.04"
  source_type  = "container"
  os_type      = "linux"
}
```

## Windows ISO example

Terraform blocks until CDI import finishes (`import_state = ready`) or fails:

```hcl
resource "virtfoundry_vm_template" "windows" {
  name              = "windows-server-2022"
  display_name      = "Windows Server 2022 Eval"
  image             = "https://go.microsoft.com/fwlink/?linkid=2195280"
  source_type       = "iso"
  os_type           = "windows"
  iso_size_gi       = 8
  boot_disk_size_gi = 32
  wait_for_import   = true
  import_wait_timeout_minutes = 45
}

resource "virtfoundry_vm" "installer" {
  name                = "win-install"
  template_id         = virtfoundry_vm_template.windows.name
  service_offering_id = "windows-large"
  public_ip           = true
  security_group_ids  = [var.security_group_id]
  desired_state       = "running"
}
```

Platform catalog already includes `windows-server-2022` on fresh installs — use `template_id = "windows-server-2022"` on the VM when import is already done.

See [`examples/iso-windows/`](../../examples/iso-windows/) and [VirtFoundry VM templates](https://github.com/virtfoundry/core/blob/main/docs/VM-TEMPLATES.md).

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Template slug. Forces replacement. |
| `image` | String | yes | Container image reference or ISO HTTP(S) URL. Forces replacement. |
| `display_name` | String | no | Human-readable name. |
| `description` | String | no | Template description. |
| `source_type` | String | no | `container` or `iso`. Forces replacement. |
| `os_type` | String | no | Operating system type (`linux`, `windows`). |
| `cloud_init_user_data` | String | no | Cloud-init user data (container disks). |
| `iso_volume_id` | String | no | Existing ISO volume UUID instead of URL import. |
| `iso_size_gi` | Number | no | ISO DataVolume size in GiB (default 8). |
| `boot_disk_size_gi` | Number | no | Blank boot disk at VM deploy (default 32 for Windows). |
| `storage_class` | String | no | StorageClass for CDI DataVolumes. |
| `wait_for_import` | Boolean | no | Wait for ISO CDI import on create (default `true`). |
| `import_wait_timeout_minutes` | Number | no | ISO wait timeout in minutes (default `45`). |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Template UUID. |
| `state` | `Active` / `Inactive`. |
| `import_state` | ISO only: `importing`, `ready`, or `failed`. |
| `hypervisor` | Hypervisor type. |

## Import

```shell
terraform import virtfoundry_vm_template.ubuntu <tenant_id>/<template_id>
```
