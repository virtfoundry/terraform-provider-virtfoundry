---
page_title: "virtfoundry_volume Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry block storage volume.
---

# virtfoundry_volume (Resource)

Manages a block storage volume backed by a Kubernetes PVC. Destroy deletes the volume via the API (fails with HTTP 409 if still attached to a VM).

## Example Usage

```hcl
resource "virtfoundry_volume" "data" {
  name    = "app-data"
  size_gi = 20
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Volume name. Forces replacement. |
| `size_gi` | Number | yes | Size in GiB. Forces replacement. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Volume UUID. |
| `namespace` | Kubernetes namespace. |
| `pvc_name` | Bound PVC name. |
| `state` | Volume state. |
| `vm_id` | Attached VM UUID, if any. |

## Import

```shell
terraform import virtfoundry_volume.data <tenant_id>/<volume_id>
```
