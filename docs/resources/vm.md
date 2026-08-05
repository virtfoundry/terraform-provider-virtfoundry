---
page_title: "virtfoundry_vm Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry virtual machine (KubeVirt).
---

# virtfoundry_vm (Resource)

Manages a virtual machine in a VirtFoundry tenant. VMs are backed by KubeVirt and support public or private networking, security groups, and cloud-init SSH keys.

## Example Usage

Use **catalog names** for templates and offerings (no data source required):

```hcl
resource "virtfoundry_vm" "web" {
  name                = "web-01"
  display_name        = "Web server"
  template_id         = "ubuntu-2204"
  service_offering_id = "small"
  public_ip           = true
  security_group_ids  = [var.security_group_id]
  desired_state       = "running"
}
```

Or resolve dynamically with the data source:

```hcl
data "virtfoundry_vm_templates" "catalog" {}

locals {
  ubuntu = one([
    for t in data.virtfoundry_vm_templates.catalog.templates : t
    if t.name == "ubuntu-2204"
  ])
}

resource "virtfoundry_vm" "web" {
  name        = "web-01"
  template_id = local.ubuntu.id
  # ...
}
```

## Built-in template catalog

VirtFoundry seeds OS templates at install. Tenants see **platform** templates plus **tenant** defaults. Use the `name` column in `template_id`:

| Name | Scope | Description |
|------|-------|-------------|
| `cirros` | platform | Cirros demo (minimal) |
| `ubuntu-2204` | platform | Ubuntu 22.04 container disk |
| `windows-server-2022` | platform | Windows Server 2022 (ISO import) |
| `fedora-39` | tenant | Fedora 39 (seeded per tenant) |

List all available templates:

```hcl
data "virtfoundry_vm_templates" "catalog" {}
# terraform console → data.virtfoundry_vm_templates.catalog.templates
```

Custom images (Ubuntu 24.04, private registry, cloud-init baked in) → [`virtfoundry_vm_template`](vm_template.md).

See also [VirtFoundry VM templates](https://github.com/virtfoundry/core/blob/main/docs/VM-TEMPLATES.md) for image URLs and ISO flow.

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | VM name (slug) within the tenant namespace. Forces replacement. |
| `display_name` | String | no | Human-readable name. |
| `template_id` | String | no | VM template **UUID or catalog name** (e.g. `ubuntu-2204`, `fedora-39`). Forces replacement. |
| `service_offering_id` | String | no | Service offering **UUID or name** (e.g. `small`). Can be changed in-place when the VM is stopped. |
| `public_ip` | Boolean | no | Attach shared public network (requires `security_group_ids`). |
| `network_ids` | List(String) | no | Private network UUIDs. Default VPC subnet is used when omitted. |
| `security_group_ids` | List(String) | no | Security group UUIDs. Required when `public_ip = true`. |
| `ssh_key_id` | String | no | SSH key UUID for cloud-init. |
| `data_volume_id` | String | no | Extra data volume UUID. |
| `expose_ssh` | Boolean | no | Expose SSH via NodePort on the cluster. |
| `desired_state` | String | no | `running` or `stopped`. Default: API default. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | VM UUID. |
| `state` | Current power state (`running`, `stopped`, …). |
| `ip` | Primary IP address. |
| `cpu` | vCPU count. |
| `memory_mi` | Memory in MiB. |
| `ssh_node_port` | NodePort when `expose_ssh = true`. |
| `ssh_exposed` | Whether SSH NodePort exposure is active. |

## Import

```shell
terraform import virtfoundry_vm.web <tenant_id>/<name>
terraform import virtfoundry_vm.web <name>   # when provider tenant_id is set
```
