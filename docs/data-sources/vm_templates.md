---
page_title: "virtfoundry_vm_templates Data Source - virtfoundry"
subcategory: ""
description: |-
  Lists VM templates available in a tenant.
---

# virtfoundry_vm_templates (Data Source)

Returns **platform** templates (shared catalog) plus **tenant** templates for the configured tenant.

## Built-in catalog (by name)

These names work directly in `virtfoundry_vm.template_id` without this data source:

| Name | Scope | Image |
|------|-------|-------|
| `cirros` | platform | `quay.io/kubevirt/cirros-container-disk-demo` |
| `ubuntu-2204` | platform | `quay.io/containerdisks/ubuntu:22.04` |
| `windows-server-2022` | platform | ISO import (Windows) |
| `fedora-39` | tenant | `quay.io/kubevirt/fedora-container-disk-demo` |

Platform templates are seeded at VirtFoundry install; `fedora-39` is added when a tenant is bootstrapped.

More images: [quay.io/containerdisks](https://quay.io/organization/containerdisks) — register custom ones with [`virtfoundry_vm_template`](../resources/vm_template.md).

## Example Usage

```hcl
data "virtfoundry_vm_templates" "catalog" {}

output "template_names" {
  value = [for t in data.virtfoundry_vm_templates.catalog.templates : t.name]
}

locals {
  ubuntu = one([
    for t in data.virtfoundry_vm_templates.catalog.templates : t
    if t.name == "ubuntu-2204"
  ])
}
```

Shortcut without data source:

```hcl
resource "virtfoundry_vm" "app" {
  template_id = "ubuntu-2204"
  # ...
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `templates` | List of objects with `id`, `name`, `display_name`, `image`, `source_type`, `os_type`, `state`. |
