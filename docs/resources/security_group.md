---
page_title: "virtfoundry_security_group Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry security group with ingress/egress rules.
---

# virtfoundry_security_group (Resource)

Manages firewall rules for VMs. Required when attaching a public IP.

## Example Usage

```hcl
resource "virtfoundry_security_group" "web" {
  name        = "allow-http-ssh"
  description = "HTTP and SSH from the internet"
  vpc_id      = virtfoundry_vpc.main.id

  rule {
    direction = "ingress"
    protocol  = "tcp"
    port_from = 22
    port_to   = 22
    cidr      = "0.0.0.0/0"
  }

  rule {
    direction = "ingress"
    protocol  = "tcp"
    port_from = 80
    port_to   = 80
    cidr      = "0.0.0.0/0"
  }
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Security group name. |
| `description` | String | no | Human-readable description. |
| `vpc_id` | String | no | VPC UUID. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

### `rule` block

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `direction` | String | yes | `ingress` or `egress`. |
| `protocol` | String | yes | `tcp`, `udp`, or `icmp`. |
| `port_from` | Number | no | Start port. |
| `port_to` | Number | no | End port. |
| `cidr` | String | yes | Source or destination CIDR. |

## Import

```shell
terraform import virtfoundry_security_group.web <tenant_id>/<security_group_id>
```
