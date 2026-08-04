---
page_title: "virtfoundry_volume_snapshot Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry volume snapshot.
---

# virtfoundry_volume_snapshot (Resource)

Creates a snapshot of a block volume. **Note:** the API has no delete endpoint yet — `terraform destroy` removes state only.

## Example Usage

```hcl
resource "virtfoundry_volume" "data" {
  name    = "app-data"
  size_gi = 20
}

resource "virtfoundry_volume_snapshot" "backup" {
  volume_id = virtfoundry_volume.data.id
  name      = "backup-2026-08-04"
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `volume_id` | String | yes | Source volume UUID. Forces replacement. |
| `name` | String | yes | Snapshot name. Forces replacement. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Snapshot UUID. |
| `namespace` | Kubernetes namespace. |
| `snapshot_uid` | Underlying snapshot UID. |
| `state` | Snapshot state. |

## Import

```shell
terraform import virtfoundry_volume_snapshot.backup <tenant_id>/<snapshot_id>
```
