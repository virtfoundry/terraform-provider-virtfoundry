# Single VM example

Deploys one Linux virtual machine with a public IP, an existing security group, and an SSH key (required for guest login).

## What it creates

- `tls_private_key.vm` + `virtfoundry_ssh_key.vm` — Ed25519 key registered with VirtFoundry
- `virtfoundry_vm.test` — VM with public IP, `ssh_key_id`, security group, running state

## Prerequisites

- VirtFoundry control plane running
- Existing security group ID (for public IP access)
- VM template ID and service offering ID
- Linux templates require `ssh_key_id` (this example generates one). There is no default guest password.

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
  -var="security_group_id=<uuid>" \
  -var="vm_name=tf-test-01"
```

## Outputs

| Output | Description |
|--------|-------------|
| `vm_id` | VM UUID |
| `vm_ip` | Primary IP address |
| `vm_state` | Power state |
| `ssh_key_id` | Registered SSH key UUID |
