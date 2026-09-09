terraform {
  required_version = ">= 1.0"

  required_providers {
    virtfoundry = {
      source  = "virtfoundry/virtfoundry"
      version = "~> 0.3"
    }
  }
}

provider "virtfoundry" {
  endpoint = var.endpoint
  username = var.username
  password = var.password
}

resource "virtfoundry_tenant" "test" {
  name           = var.tenant_name
  slug           = var.tenant_slug
  admin_password = var.admin_password
}

output "tenant_id" {
  value = virtfoundry_tenant.test.id
}

output "tenant_slug" {
  value = virtfoundry_tenant.test.slug
}
