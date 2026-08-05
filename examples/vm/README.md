# Single VM example

Deploys one virtual machine with a public IP and an existing security group.

## What it creates

- `virtfoundry_vm.test` — VM with public IP, attached security group, running state

## Prerequisites

- VirtFoundry control plane running
- Existing security group ID (for public IP access)

Templates and offerings use **catalog names** — no UUID lookup required:

| Variable | Default | Catalog |
|----------|---------|---------|
| `template_name` | `ubuntu-2204` | platform template |
| `service_offering_name` | `small` | 1 vCPU / 1 GiB |

Other built-in template names: `cirros`, `fedora-39`, `windows-server-2022`.  
List all: `data "virtfoundry_vm_templates" "catalog" {}`

## Usage

```bash
terraform init
terraform apply \
  -var="endpoint=https://virtfoundry.example.com" \
  -var="username=admin" \
  -var="password=..." \
  -var="tenant_id=<uuid>" \
  -var="template_name=ubuntu-2204" \
  -var="service_offering_name=small" \
  -var="security_group_id=<uuid>"
```

## Outputs

| Output | Description |
|--------|-------------|
| `vm_id` | VM UUID |
| `vm_ip` | Primary IP address |
| `vm_state` | Power state |
