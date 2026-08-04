# Single VM example

Deploys one virtual machine with a public IP and an existing security group.

## What it creates

- `virtfoundry_vm.test` — VM with public IP, attached security group, running state

## Prerequisites

- VirtFoundry control plane running
- Existing security group ID (for public IP access)
- VM template ID and service offering ID

## Usage

```bash
terraform init
terraform apply \
  -var="endpoint=https://virtfoundry.example.com" \
  -var="username=admin" \
  -var="password=..." \
  -var="tenant_id=<uuid>" \
  -var="template_id=<uuid>" \
  -var="service_offering_id=small" \
  -var="security_group_id=<uuid>"
```

## Outputs

| Output | Description |
|--------|-------------|
| `vm_id` | VM UUID |
| `vm_ip` | Primary IP address |
| `vm_state` | Power state |
