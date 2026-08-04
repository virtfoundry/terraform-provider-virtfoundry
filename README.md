# Terraform Provider for VirtFoundry

Official [Terraform](https://www.terraform.io/) provider for [VirtFoundry](https://github.com/virtfoundry/core) — private cloud IaaS on Kubernetes.

## Requirements

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- A running VirtFoundry control plane ([helm-charts](https://github.com/virtfoundry/helm-charts))

## Usage

```hcl
terraform {
  required_providers {
    virtfoundry = {
      source  = "virtfoundry/virtfoundry"
      version = "~> 0.1"
    }
  }
}

provider "virtfoundry" {
  endpoint = "https://virtfoundry.example.com"
  api_key  = var.virtfoundry_api_key
  # username = "root"
  # password = var.root_password
  tenant_id = "tenant-abc123"
}
```

### Environment variables

| Variable | Description |
|----------|-------------|
| `VIRTFOUNDRY_ENDPOINT` | API base URL |
| `VIRTFOUNDRY_API_KEY` | API key (`vfd_live_...`) |
| `VIRTFOUNDRY_USERNAME` | Username (JWT login) |
| `VIRTFOUNDRY_PASSWORD` | Password (JWT login) |
| `VIRTFOUNDRY_TENANT_ID` | Default tenant ID |
| `VIRTFOUNDRY_INSECURE` | Set to `true` to skip TLS verification (dev only) |

## Development

```bash
make build
make test
make install   # installs to ~/.terraform.d/plugins/... for local terraform init
```

Local Terraform example:

```bash
cd examples/provider
terraform init
terraform plan
```

## Roadmap

See [TERRAFORM-PROVIDER-PLAN.md](https://github.com/virtfoundry/core/blob/main/docs/TERRAFORM-PROVIDER-PLAN.md) in the core repo.

**Phase 0 (current):** provider block, authentication, health check.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
