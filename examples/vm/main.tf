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
  username  = var.username
  password  = var.password
  tenant_id = var.tenant_id
}

resource "virtfoundry_vm" "test" {
  name                = var.vm_name
  display_name        = "Terraform test VM"
  template_id         = var.template_name
  service_offering_id = var.service_offering_name
  public_ip           = true
  security_group_ids  = [var.security_group_id]
  desired_state       = "running"
}

output "vm_id" {
  value = virtfoundry_vm.test.id
}

output "vm_ip" {
  value = virtfoundry_vm.test.ip
}

output "vm_state" {
  value = virtfoundry_vm.test.state
}
