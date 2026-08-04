# Provider configuration example

Minimal Terraform configuration showing how to declare the VirtFoundry provider.

## Usage

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
```

Copy `terraform.tfvars.example` (if present) or pass variables via environment:

```bash
export VIRTFOUNDRY_ENDPOINT=https://virtfoundry.example.com
export VIRTFOUNDRY_API_KEY=vfd_live_...
export VIRTFOUNDRY_TENANT_ID=<tenant-uuid>
```

## Local development

When testing a locally built provider binary:

```bash
export TF_CLI_CONFIG_FILE=examples/provider/.terraformrc
terraform init
```
