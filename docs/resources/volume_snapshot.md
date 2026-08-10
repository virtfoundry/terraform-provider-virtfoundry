---
page_title: "virtfoundry_volume_snapshot Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry volume snapshot.
---

# virtfoundry_volume_snapshot (Resource)

Creates a snapshot of a block volume via the VirtFoundry API (`VolumeSnapshot` / `snapshot.storage.k8s.io`).

**Cluster prerequisites:** CSI external-snapshotter CRDs + snapshot-controller, a snapshot-capable StorageClass (e.g. Longhorn, Ceph RBD), and a `VolumeSnapshotClass`. This does **not** work with `local-path`. For guest-level snapshots without CSI, use [`virtfoundry_vm_snapshot`](vm_snapshot.md) instead.

**Note:** the API has no delete endpoint yet — `terraform destroy` removes state only.

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
