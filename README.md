# Terraform Provider for VirtFoundry

[![Terraform Registry](https://img.shields.io/badge/registry-virtfoundry%2Fvirtfoundry-blue)](https://registry.terraform.io/providers/virtfoundry/virtfoundry/latest)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

Official [Terraform](https://www.terraform.io/) provider for [VirtFoundry](https://github.com/virtfoundry/core) — Kubernetes-native private cloud IaaS.

Manage tenants, networking, compute, storage, and IAM through the VirtFoundry REST API (`/api/v1`). VirtFoundry orchestrates **KubeVirt** VMs, **Multus** networks, and multi-tenant isolation on Kubernetes.

## Install

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_version = ">= 1.0"

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
```

Run `terraform init` to download the provider from the [Terraform Registry](https://registry.terraform.io/providers/virtfoundry/virtfoundry/latest).

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- A running VirtFoundry control plane — install with the [Helm chart](https://github.com/virtfoundry/helm-charts)

## Quick example

Deploy a VM with a public IP:

```hcl
data "virtfoundry_service_offerings" "catalog" {}
data "virtfoundry_vm_templates" "catalog" {}

resource "virtfoundry_security_group" "ssh" {
  name   = "allow-ssh"
  vpc_id = var.vpc_id

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
  template_id         = var.template_id
  service_offering_id = "small"
  public_ip           = true
  security_group_ids  = [virtfoundry_security_group.ssh.id]
  desired_state       = "running"
}
```

See [`examples/full-stack/`](examples/full-stack/) for VPC, network, security group, SSH key, and VM in one stack.

## Authentication

| Attribute | Description |
|-----------|-------------|
| `endpoint` | VirtFoundry API URL (required) |
| `api_key` | API key (`vfd_live_...`) — recommended for CI/CD |
| `username` / `password` | JWT login — suitable for development |
| `tenant_id` | Default tenant for tenant-scoped resources |
| `insecure` | Skip TLS verification (development only) |

Environment variables: `VIRTFOUNDRY_ENDPOINT`, `VIRTFOUNDRY_API_KEY`, `VIRTFOUNDRY_USERNAME`, `VIRTFOUNDRY_PASSWORD`, `VIRTFOUNDRY_TENANT_ID`, `VIRTFOUNDRY_INSECURE`.

## Resources

| Resource | Description |
|----------|-------------|
| [`virtfoundry_tenant`](docs/resources/tenant.md) | Platform tenant (root credentials) |
| [`virtfoundry_vpc`](docs/resources/vpc.md) | Tenant VPC |
| [`virtfoundry_network`](docs/resources/network.md) | Subnet in a VPC |
| [`virtfoundry_security_group`](docs/resources/security_group.md) | Firewall rules |
| [`virtfoundry_volume`](docs/resources/volume.md) | Block storage volume |
| [`virtfoundry_volume_snapshot`](docs/resources/volume_snapshot.md) | Volume snapshot |
| [`virtfoundry_vm_template`](docs/resources/vm_template.md) | VM template (container or ISO) |
| [`virtfoundry_vm`](docs/resources/vm.md) | Virtual machine |
| [`virtfoundry_vm_snapshot`](docs/resources/vm_snapshot.md) | VM snapshot |
| [`virtfoundry_ssh_key`](docs/resources/ssh_key.md) | SSH public key |
| [`virtfoundry_user`](docs/resources/user.md) | IAM user |
| [`virtfoundry_role`](docs/resources/role.md) | IAM role |
| [`virtfoundry_api_key`](docs/resources/api_key.md) | API key (secret shown once) |

## Data sources

| Data source | Description |
|-------------|-------------|
| [`virtfoundry_service_offerings`](docs/data-sources/service_offerings.md) | CPU/memory catalog |
| [`virtfoundry_vm_templates`](docs/data-sources/vm_templates.md) | VM templates |
| [`virtfoundry_vpcs`](docs/data-sources/vpcs.md) | VPC list |
| [`virtfoundry_networks`](docs/data-sources/networks.md) | Network list |
| [`virtfoundry_security_groups`](docs/data-sources/security_groups.md) | Security group list |
| [`virtfoundry_ssh_keys`](docs/data-sources/ssh_keys.md) | SSH key list |
| [`virtfoundry_users`](docs/data-sources/users.md) | IAM user list |
| [`virtfoundry_roles`](docs/data-sources/roles.md) | IAM role list |

## Examples

| Directory | Description |
|-----------|-------------|
| [`examples/provider/`](examples/provider/) | Minimal provider configuration |
| [`examples/vm/`](examples/vm/) | Single VM with public IP |
| [`examples/full-stack/`](examples/full-stack/) | VPC + network + SG + SSH key + VM |
| [`examples/tenant-with-iam/`](examples/tenant-with-iam/) | Tenant bootstrap with users and roles |

Reusable modules: [`modules/tenant/`](modules/tenant/), [`modules/tenant-iam/`](modules/tenant-iam/).

## Import

Tenant-scoped resources import as `<tenant_id>/<id>` or `<id>` when the provider `tenant_id` is set. VMs import as `<tenant_id>/<name>` or `<name>`.

```shell
terraform import virtfoundry_vm.web <tenant_id>/web-01
terraform import virtfoundry_vpc.main <vpc_id>
```

## Development

```bash
make build
make test
make install   # ~/.terraform.d/plugins for local dev_overrides
make test-integration       # quick VM-only apply/destroy
make test-integration-full  # full stack apply/destroy
```

Local Terraform with dev overrides:

```bash
export TF_CLI_CONFIG_FILE=examples/provider/.terraformrc
cd examples/vm && terraform init && terraform apply
```

## Documentation

- [Provider docs](docs/index.md) — full reference on the Terraform Registry
- [Installation guide](https://virtfoundry.github.io/helm-charts/docs/guide/installation/)
- [VirtFoundry core](https://github.com/virtfoundry/core)

## License

Apache License 2.0 — see [LICENSE](LICENSE).
