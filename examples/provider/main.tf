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
  endpoint  = var.endpoint
  api_key   = var.api_key
  tenant_id = var.tenant_id
  insecure  = var.insecure
}
