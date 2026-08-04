---
page_title: "virtfoundry_network Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry tenant network within a VPC.
---

# virtfoundry_network (Resource)

Manages a subnet/network inside a VPC. Omit `cidr` to auto-allocate from the VPC address space.

## Example Usage

```hcl
resource "virtfoundry_vpc" "main" {
  name = "production"
  cidr = "10.0.0.0/16"
}

resource "virtfoundry_network" "private" {
  name   = "app-subnet"
  vpc_id = virtfoundry_vpc.main.id
  prefix = 24
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Network name. Forces replacement. |
| `vpc_id` | String | yes | Parent VPC UUID. Forces replacement. |
| `cidr` | String | no | Explicit CIDR; auto-allocated when omitted. |
| `prefix` | Number | no | Subnet prefix for auto-allocation (default `24`). |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | Network UUID. |
| `gateway` | Subnet gateway IP. |
| `network_type` | Network type. |
| `state` | Network state. |

## Import

```shell
terraform import virtfoundry_network.private <tenant_id>/<network_id>
```
