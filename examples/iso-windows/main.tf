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

resource "virtfoundry_vm_template" "windows" {
  count = var.create_iso_template ? 1 : 0

  name              = var.template_name
  display_name      = var.template_display_name
  image             = var.iso_url
  source_type       = "iso"
  os_type           = "windows"
  iso_size_gi       = var.iso_size_gi
  boot_disk_size_gi = var.boot_disk_size_gi
  wait_for_import   = true
  import_wait_timeout_minutes = var.import_wait_timeout_minutes
}

locals {
  template_ref = var.create_iso_template ? virtfoundry_vm_template.windows[0].name : var.existing_template_name
}

resource "virtfoundry_vm" "installer" {
  count = var.deploy_vm ? 1 : 0

  name                = var.vm_name
  display_name        = var.vm_display_name
  template_id         = local.template_ref
  service_offering_id = var.service_offering_name
  public_ip           = true
  security_group_ids  = [var.security_group_id]
  desired_state       = var.desired_state
}

output "template_id" {
  value = var.create_iso_template ? virtfoundry_vm_template.windows[0].id : null
}

output "template_import_state" {
  value = var.create_iso_template ? virtfoundry_vm_template.windows[0].import_state : null
}

output "vm_id" {
  value = var.deploy_vm ? virtfoundry_vm.installer[0].id : null
}

output "template_name_used" {
  value = local.template_ref
}
