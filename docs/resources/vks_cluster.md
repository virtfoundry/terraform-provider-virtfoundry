---
page_title: "virtfoundry_vks_cluster Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry Kubernetes Service (VKS) cluster: a Kamaji control plane plus worker Instances.
---

# virtfoundry_vks_cluster (Resource)

Creates a VKS cluster through `/api/v1/vks/clusters`. The API has no in-place update, so changing any argument **replaces** the cluster. Create returns once the API accepts the cluster; `phase` and `ready_workers` reflect progress on refresh.

## Example Usage

```hcl
resource "virtfoundry_vks_cluster" "dev" {
  name               = "dev"
  kubernetes_version = "v1.31.4"

  workers = {
    count        = 2
    template_ref = "node-v1-31-4"
    offering_ref = "medium"
    network_ref  = "lan"
    ssh_key_refs = [virtfoundry_ssh_key.workers.name]
  }

  control_plane = {
    service_type = "LoadBalancer"
  }
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Cluster name (DNS label) within the tenant. |
| `kubernetes_version` | String | yes | Kubernetes version (for example `v1.31.4`); must match a published node image. |
| `workers` | Object | yes | Worker pool (see below). |
| `control_plane` | Object | no | Control plane exposure; omitted fields use platform defaults. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

### `workers`

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `count` | Number | yes | Number of workers (at least 1). |
| `template_ref` | String | yes | Name of the node image VM template. |
| `offering_ref` | String | yes | Service offering name. |
| `network_ref` | String | yes | Network name. |
| `ssh_key_refs` | Set of String | no | SSH key names injected into workers. |

### `control_plane`

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `service_type` | String | no | Service type fronting the API server (`LoadBalancer`, `NodePort`, `ClusterIP`). |
| `address` | String | no | Advertised API server address. |
| `port` | Number | no | API server port. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Cluster identifier (the cluster name). |
| `phase` | Cluster phase (for example `Provisioning`, `Ready`). |
| `control_plane_endpoint` | Kubernetes API endpoint once known. |
| `ready_workers` | Workers that have joined and are ready. |
| `namespace` | Management-cluster namespace hosting the cluster. |

## Import

```shell
terraform import virtfoundry_vks_cluster.dev <tenant_id>/<cluster_name>
```

See also the [`virtfoundry_vks_kubeconfig`](../data-sources/vks_kubeconfig.md) data source.
