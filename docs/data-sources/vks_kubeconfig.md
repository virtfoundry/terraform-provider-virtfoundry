---
page_title: "virtfoundry_vks_kubeconfig Data Source - virtfoundry"
subcategory: ""
description: |-
  Reads the admin kubeconfig of a VKS cluster.
---

# virtfoundry_vks_kubeconfig (Data Source)

Reads the admin kubeconfig of a VKS cluster. Requires the `vks:kubeconfig` permission.

> **Security:** the kubeconfig is a credential. It is marked `sensitive` but is still stored in state; enable state encryption.

## Example Usage

```hcl
data "virtfoundry_vks_kubeconfig" "dev" {
  name = virtfoundry_vks_cluster.dev.name
}

output "kubeconfig" {
  value     = data.virtfoundry_vks_kubeconfig.dev.kubeconfig
  sensitive = true
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | VKS cluster name. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `kubeconfig` | Kubeconfig YAML (sensitive). |
