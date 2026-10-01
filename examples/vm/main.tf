terraform {
  required_version = ">= 1.11"

  required_providers {
    virtfoundry = {
      source  = "virtfoundry/virtfoundry"
      version = "~> 0.3"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "virtfoundry" {
  endpoint  = var.endpoint
  username  = var.username
  password  = var.password
  tenant_id = var.tenant_id
}

# Linux VMs require ssh_key_id (or cloud_init_password via the API). Prefer an SSH key.
resource "tls_private_key" "vm" {
  algorithm = "ED25519"
}

resource "virtfoundry_ssh_key" "vm" {
  name       = "${var.vm_name}-key"
  public_key = tls_private_key.vm.public_key_openssh
}

resource "virtfoundry_vm" "test" {
  name                = var.vm_name
  display_name        = "Terraform test VM"
  template_id         = var.template_id
  service_offering_id = var.service_offering_id
  public_ip           = true
  security_group_ids  = [var.security_group_id]
  ssh_key_id          = virtfoundry_ssh_key.vm.id
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

output "ssh_key_id" {
  value = virtfoundry_ssh_key.vm.id
}
