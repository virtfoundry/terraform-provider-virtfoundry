# Full stack example

End-to-end infrastructure: VPC, private network, security group, SSH key, and VM.

## What it creates

| Resource | Purpose |
|----------|---------|
| `virtfoundry_vpc.main` | Tenant VPC with default network |
| `virtfoundry_network.private` | Application subnet |
| `virtfoundry_security_group.ssh` | Allow SSH (port 22) |
| `virtfoundry_ssh_key.admin` | Generated Ed25519 key pair |
| `virtfoundry_vm.app` | Application VM on private network |

Data sources resolve the service offering and VM template by name.

## Usage

```bash
terraform init
terraform apply \
  -var="endpoint=https://virtfoundry.example.com" \
  -var="username=admin" \
  -var="password=..." \
  -var="tenant_id=<uuid>" \
  -var="template_name=ubuntu-2204" \
  -var="service_offering_name=small"
```

## Outputs

| Output | Description |
|--------|-------------|
| `vpc_id` | VPC UUID |
| `network_id` | Private network UUID |
| `security_group_id` | Security group UUID |
| `ssh_key_id` | SSH key UUID |
| `vm_id` | VM UUID |
| `vm_ip` | VM IP address |
| `ssh_node_port` | NodePort when SSH is exposed |
| `ssh_exposed` | Whether SSH NodePort is active |

## Integration test

```bash
make test-integration-full
```
