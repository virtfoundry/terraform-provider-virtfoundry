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

resource "virtfoundry_service_offering" "custom" {
  name          = var.offering_name
  display_name  = var.offering_display_name
  cpu           = 1
  memory_mi     = 512
  dedicated_cpu = true
}

data "virtfoundry_service_offerings" "catalog" {}

output "offering_id" {
  value = virtfoundry_service_offering.custom.id
}

output "offering_dedicated_cpu" {
  value = virtfoundry_service_offering.custom.dedicated_cpu
}
