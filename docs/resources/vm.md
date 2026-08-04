---
page_title: "virtfoundry_vm Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry virtual machine (KubeVirt).
---

# virtfoundry_vm (Resource)

Manages a virtual machine in a VirtFoundry tenant. VMs are backed by KubeVirt and support public or private networking, security groups, and cloud-init SSH keys.

## Example Usage

```hcl
data "virtfoundry_service_offerings" "catalog" {}
data "virtfoundry_vm_templates" "catalog" {}

resource "virtfoundry_security_group" "ssh" {
  name   = "allow-ssh"
  vpc_id = virtfoundry_vpc.main.id

  rule {
    direction = "ingress"
    protocol  = "tcp"
    port_from = 22
    port_to   = 22
    cidr      = "0.0.0.0/0"
  }
}

resource "virtfoundry_vm" "web" {
  name                = "web-01"
  display_name        = "Web server"
  template_id         = var.template_id
  service_offering_id = "small"
  public_ip           = true
  security_group_ids  = [virtfoundry_security_group.ssh.id]
  desired_state       = "running"
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | VM name (slug) within the tenant namespace. Forces replacement. |
| `display_name` | String | no | Human-readable name. |
| `template_id` | String | no | VM template UUID. Forces replacement. |
| `service_offering_id` | String | no | Service offering UUID or name (e.g. `small`). Forces replacement. |
| `public_ip` | Boolean | no | Attach shared public network (requires `security_group_ids`). |
| `network_ids` | List(String) | no | Private network UUIDs. Default VPC subnet is used when omitted. |
| `security_group_ids` | List(String) | no | Security group UUIDs. Required when `public_ip = true`. |
| `ssh_key_id` | String | no | SSH key UUID for cloud-init. |
| `data_volume_id` | String | no | Extra data volume UUID. |
| `expose_ssh` | Boolean | no | Expose SSH via NodePort on the cluster. |
| `desired_state` | String | no | `running` or `stopped`. Default: API default. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | VM UUID. |
| `state` | Current power state (`running`, `stopped`, …). |
| `ip` | Primary IP address. |
| `cpu` | vCPU count. |
| `memory_mi` | Memory in MiB. |
| `ssh_node_port` | NodePort when `expose_ssh = true`. |
| `ssh_exposed` | Whether SSH NodePort exposure is active. |

## Import

```shell
terraform import virtfoundry_vm.web <tenant_id>/<name>
terraform import virtfoundry_vm.web <name>   # when provider tenant_id is set
```
