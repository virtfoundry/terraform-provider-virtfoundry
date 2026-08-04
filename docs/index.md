---
page_title: "VirtFoundry Provider"
subcategory: ""
description: |-
  Terraform provider for VirtFoundry — Kubernetes-native private cloud IaaS.
---

# VirtFoundry Provider

Manage [VirtFoundry](https://github.com/virtfoundry/core) infrastructure with Terraform: tenants, networking, compute, storage, and IAM — via the REST API (`/api/v1`).

VirtFoundry orchestrates **KubeVirt** VMs, **Multus** networks, and tenant isolation on Kubernetes. Install the control plane with the [Helm chart](https://github.com/virtfoundry/helm-charts).

## Quick start

```hcl
terraform {
  required_providers {
    virtfoundry = {
      source  = "virtfoundry/virtfoundry"
      version = "~> 0.2"
    }
  }
}

provider "virtfoundry" {
  endpoint  = "https://virtfoundry.example.com"
  api_key   = var.virtfoundry_api_key
  tenant_id = var.tenant_id
}

resource "virtfoundry_vm" "web" {
  name                = "web-01"
  template_id         = var.template_id
  service_offering_id = "small"
  public_ip           = true
  security_group_ids  = [virtfoundry_security_group.ssh.id]
  desired_state       = "running"
}
```

## Authentication

| Attribute | Description |
|-----------|-------------|
| `api_key` | API key (`vfd_live_...`) — **recommended** for automation |
| `username` / `password` | JWT login — suitable for development |
| `tenant_id` | Default tenant for tenant-scoped resources |
| `insecure` | Skip TLS verification (development only) |

Environment variables: `VIRTFOUNDRY_ENDPOINT`, `VIRTFOUNDRY_API_KEY`, `VIRTFOUNDRY_USERNAME`, `VIRTFOUNDRY_PASSWORD`, `VIRTFOUNDRY_TENANT_ID`, `VIRTFOUNDRY_INSECURE`.

## Resources

| Resource | Description |
|----------|-------------|
| [virtfoundry_tenant](resources/tenant.md) | Platform tenant (root credentials) |
| [virtfoundry_vpc](resources/vpc.md) | Tenant VPC |
| [virtfoundry_network](resources/network.md) | Subnet / network in a VPC |
| [virtfoundry_security_group](resources/security_group.md) | Security group + rules |
| [virtfoundry_volume](resources/volume.md) | Block volume |
| [virtfoundry_volume_snapshot](resources/volume_snapshot.md) | Volume snapshot |
| [virtfoundry_vm_template](resources/vm_template.md) | VM template (container or ISO) |
| [virtfoundry_vm](resources/vm.md) | Virtual machine |
| [virtfoundry_vm_snapshot](resources/vm_snapshot.md) | VM snapshot |
| [virtfoundry_ssh_key](resources/ssh_key.md) | SSH public key |
| [virtfoundry_user](resources/user.md) | IAM user |
| [virtfoundry_role](resources/role.md) | IAM role |
| [virtfoundry_api_key](resources/api_key.md) | API key (secret shown once) |

## Data sources

| Data source | Description |
|-------------|-------------|
| [virtfoundry_service_offerings](data-sources/service_offerings.md) | CPU/memory catalog |
| [virtfoundry_vm_templates](data-sources/vm_templates.md) | VM templates |
| [virtfoundry_vpcs](data-sources/vpcs.md) | VPC list |
| [virtfoundry_networks](data-sources/networks.md) | Networks |
| [virtfoundry_security_groups](data-sources/security_groups.md) | Security groups |
| [virtfoundry_ssh_keys](data-sources/ssh_keys.md) | SSH keys |
| [virtfoundry_users](data-sources/users.md) | IAM users |
| [virtfoundry_roles](data-sources/roles.md) | IAM roles |

## Examples

See the [examples/](https://github.com/virtfoundry/terraform-provider-virtfoundry/tree/main/examples) directory:

- **provider** — minimal provider configuration
- **vm** — single VM with public IP
- **full-stack** — VPC, network, security group, SSH key, and VM
- **tenant-with-iam** — tenant bootstrap with users and roles

## Links

- [Installation guide](https://virtfoundry.github.io/helm-charts/docs/guide/installation/)
- [VirtFoundry core](https://github.com/virtfoundry/core)
- [Report an issue](https://github.com/virtfoundry/terraform-provider-virtfoundry/issues)
