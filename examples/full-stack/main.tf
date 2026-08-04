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

data "virtfoundry_service_offerings" "catalog" {}

data "virtfoundry_vm_templates" "catalog" {}

locals {
  offering = one([
    for o in data.virtfoundry_service_offerings.catalog.offerings : o
    if o.name == var.service_offering_name
  ])
  template = one([
    for t in data.virtfoundry_vm_templates.catalog.templates : t
    if t.name == var.template_name
  ])
}

resource "virtfoundry_vpc" "main" {
  name = var.vpc_name
  cidr = var.vpc_cidr
}

resource "virtfoundry_network" "private" {
  name   = var.network_name
  vpc_id = virtfoundry_vpc.main.id
  prefix = var.network_prefix
}

resource "virtfoundry_security_group" "ssh" {
  name        = var.security_group_name
  description = "Allow SSH from anywhere (demo)"
  vpc_id      = virtfoundry_vpc.main.id

  rule {
    direction = "ingress"
    protocol  = "tcp"
    port_from = 22
    port_to   = 22
    cidr      = "0.0.0.0/0"
  }
}

resource "virtfoundry_ssh_key" "admin" {
  name     = var.ssh_key_name
  generate = true
}

resource "virtfoundry_vm" "app" {
  name                = var.vm_name
  display_name        = var.vm_display_name
  template_id         = local.template.id
  service_offering_id = local.offering.id
  network_ids         = [virtfoundry_network.private.id]
  security_group_ids  = [virtfoundry_security_group.ssh.id]
  ssh_key_id          = virtfoundry_ssh_key.admin.id
  expose_ssh          = var.expose_ssh
  desired_state       = "running"
}

output "vpc_id" {
  value = virtfoundry_vpc.main.id
}

output "network_id" {
  value = virtfoundry_network.private.id
}

output "security_group_id" {
  value = virtfoundry_security_group.ssh.id
}

output "ssh_key_id" {
  value = virtfoundry_ssh_key.admin.id
}

output "vm_id" {
  value = virtfoundry_vm.app.id
}

output "vm_ip" {
  value = virtfoundry_vm.app.ip
}

output "ssh_node_port" {
  value = virtfoundry_vm.app.ssh_node_port
}

output "ssh_exposed" {
  value = virtfoundry_vm.app.ssh_exposed
}
