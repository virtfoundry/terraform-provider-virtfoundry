terraform {
  required_version = ">= 1.0"

  required_providers {
    virtfoundry = {
      source  = "virtfoundry/virtfoundry"
      version = "~> 0.2"
    }
  }
}

# Root provider — no tenant_id (required for module.tenant)
provider "virtfoundry" {
  alias    = "root"
  endpoint = var.endpoint
  username = var.root_username
  password = var.root_password
}

# Tenant-scoped provider for IAM after tenant exists
provider "virtfoundry" {
  alias     = "tenant"
  endpoint  = var.endpoint
  username  = var.root_username
  password  = var.root_password
  tenant_id = local.tenant_id
}

locals {
  tenant_id = var.create_tenant ? module.tenant[0].id : var.tenant_id
}

module "tenant" {
  count  = var.create_tenant ? 1 : 0
  source = "../../modules/tenant"

  providers = {
    virtfoundry = virtfoundry.root
  }

  name           = var.tenant_name
  slug           = var.tenant_slug
  admin_password = var.tenant_admin_password
}

module "iam" {
  source = "../../modules/tenant-iam"

  providers = {
    virtfoundry = virtfoundry.tenant
  }

  tenant_id = local.tenant_id

  roles = {
    operator = {
      description = "Deploy and manage infrastructure"
      permissions = []
    }
    vm_viewer = {
      permissions = ["vms:read", "networks:read", "vpcs:read"]
    }
  }

  users = {
    operator = {
      username  = var.operator_username
      password  = var.operator_password
      role_name = "operator"
    }
  }
}

output "tenant_id" {
  value = local.tenant_id
}

output "operator_user_id" {
  value = module.iam.user_ids["operator"]
}

output "operator_role_id" {
  value = module.iam.role_ids["operator"]
}
