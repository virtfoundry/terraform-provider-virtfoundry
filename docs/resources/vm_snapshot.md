---
page_title: "virtfoundry_vm_snapshot Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a KubeVirt VM snapshot.
---

# virtfoundry_vm_snapshot (Resource)

Creates a point-in-time snapshot of a running or stopped VM using KubeVirt VirtualMachineSnapshot.

## Example Usage

```hcl
resource "virtfoundry_vm_snapshot" "pre_upgrade" {
  vm_name = virtfoundry_vm.app.name
  name    = "pre-upgrade-2026-08-04"
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `vm_name` | String | yes | Target VM name. Forces replacement. |
| `name` | String | yes | Snapshot name. Forces replacement. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Snapshot UUID. |
| `namespace` | Kubernetes namespace. |
| `snapshot_uid` | KubeVirt snapshot UID. |
| `phase` | Snapshot phase. |
| `vm_id` | Source VM UUID. |

## Import

```shell
terraform import virtfoundry_vm_snapshot.pre_upgrade <tenant_id>/<snapshot_id>
```
