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
  endpoint  = var.endpoint
  username  = var.username
  password  = var.password
  tenant_id = var.tenant_id
}

resource "virtfoundry_vm" "test" {
  name                = var.vm_name
  display_name        = "Terraform volume test VM"
  template_id         = var.template_id
  service_offering_id = var.service_offering_id
  public_ip           = true
  security_group_ids  = [var.security_group_id]
  desired_state       = "running"
}

resource "virtfoundry_volume" "data" {
  name    = var.volume_name
  size_gi = 1
}

resource "virtfoundry_volume_attachment" "data" {
  vm_name   = virtfoundry_vm.test.name
  volume_id = virtfoundry_volume.data.id
}

output "vm_id" {
  value = virtfoundry_vm.test.id
}

output "volume_id" {
  value = virtfoundry_volume.data.id
}

output "attachment_id" {
  value = virtfoundry_volume_attachment.data.id
}
