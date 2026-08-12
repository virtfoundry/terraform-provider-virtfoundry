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
  alias    = "root"
  endpoint = var.endpoint
  username = var.username
  password = var.password
}

provider "virtfoundry" {
  endpoint  = var.endpoint
  username  = var.username
  password  = var.password
  tenant_id = var.tenant_id
}

resource "virtfoundry_service_offering" "custom" {
  provider     = virtfoundry.root
  name         = var.offering_name
  display_name = var.offering_display_name
  cpu          = var.offering_cpu
  memory_mi    = var.offering_memory_mi
}

resource "virtfoundry_vm_template" "app" {
  name          = var.template_name
  display_name  = var.template_display_name
  image         = var.template_image
  source_type   = "container"
  os_type       = "linux"
  boot_disk_size_gi = 1
}

data "virtfoundry_service_offerings" "catalog" {}

data "virtfoundry_security_groups" "default" {}

locals {
  baseline_offering = one([
    for o in data.virtfoundry_service_offerings.catalog.offerings : o
    if o.name == var.baseline_offering_name
  ])
  security_group = one(data.virtfoundry_security_groups.default.security_groups)
  offering_id = var.use_custom_offering ? virtfoundry_service_offering.custom.id : local.baseline_offering.id
}

resource "virtfoundry_volume" "data" {
  name    = var.volume_name
  size_gi = var.volume_size_gi
}

resource "virtfoundry_vm" "app" {
  name                = var.vm_name
  display_name        = var.vm_display_name
  template_id         = virtfoundry_vm_template.app.id
  service_offering_id = local.offering_id
  public_ip           = true
  security_group_ids  = [local.security_group.id]
  desired_state       = var.desired_state
}

resource "virtfoundry_vm_volume_attachment" "data" {
  vm_name   = virtfoundry_vm.app.name
  volume_id = virtfoundry_volume.data.id
}

output "offering_id" {
  value = virtfoundry_service_offering.custom.id
}

output "template_id" {
  value = virtfoundry_vm_template.app.id
}

output "volume_id" {
  value = virtfoundry_volume.data.id
}

output "vm_id" {
  value = virtfoundry_vm.app.id
}

output "attachment_id" {
  value = virtfoundry_vm_volume_attachment.data.id
}

output "vm_cpu" {
  value = virtfoundry_vm.app.cpu
}

output "vm_memory_mi" {
  value = virtfoundry_vm.app.memory_mi
}
