---
page_title: "virtfoundry_vpc Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry VPC. Creating a VPC also provisions a default network.
---

# virtfoundry_vpc (Resource)

Manages a tenant VPC. VirtFoundry automatically creates a default network when the VPC is provisioned.

## Example Usage

```hcl
resource "virtfoundry_vpc" "main" {
  name = "production"
  cidr = "10.0.0.0/16"
}

output "default_network_id" {
  value = virtfoundry_vpc.main.default_network_id
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | VPC name. Forces replacement. |
| `cidr` | String | yes | VPC CIDR block (e.g. `10.0.0.0/16`). Forces replacement. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | VPC UUID. |
| `namespace` | Kubernetes namespace. |
| `state` | VPC state. |
| `default_network_id` | UUID of the auto-created default network. |

## Import

```shell
terraform import virtfoundry_vpc.main <tenant_id>/<vpc_id>
terraform import virtfoundry_vpc.main <vpc_id>
```
