# VKS cluster example

Creates a VirtFoundry Kubernetes Service (VKS) cluster — a Kamaji-hosted control plane plus worker Instances — and reads its admin kubeconfig.

## What it creates

- `tls_private_key.workers` + `virtfoundry_ssh_key.workers` — SSH key injected into worker nodes
- `virtfoundry_vks_cluster.this` — cluster with `workers` and an optional `control_plane`
- `data.virtfoundry_vks_kubeconfig.this` — kubeconfig (sensitive)

## Prerequisites

- VirtFoundry control plane with the VKS operator installed
- A node image VM template whose Kubernetes version matches `kubernetes_version` (see `vks-image-factory/COMPATIBILITY.md`)
- An existing network and service offering
- API permissions `vks:write` (create/delete) and `vks:kubeconfig` (kubeconfig data source)

## Usage

```bash
terraform init
terraform apply \
  -var="endpoint=https://virtfoundry.example.com" \
  -var="password=..." \
  -var="tenant_id=<uuid>" \
  -var="template_ref=<node-template-name>" \
  -var="network_ref=<network-name>"

terraform output -raw kubeconfig > dev.kubeconfig
KUBECONFIG=dev.kubeconfig kubectl get nodes
```

## Notes

- `apply` returns once the API accepts the cluster. Track `phase` / `ready_workers`; re-run `terraform refresh` (or `apply`) to update them.
- The API has no in-place update: any change to the cluster arguments replaces it.
- The kubeconfig is a credential stored in state. Enable state encryption.

## Outputs

| Output | Description |
|--------|-------------|
| `phase` | Cluster phase |
| `control_plane_endpoint` | Kubernetes API endpoint |
| `ready_workers` | Workers ready |
| `kubeconfig` | Admin kubeconfig (sensitive) |
