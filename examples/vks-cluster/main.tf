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

resource "tls_private_key" "workers" {
  algorithm = "ED25519"
}

resource "virtfoundry_ssh_key" "workers" {
  name       = "${var.cluster_name}-workers"
  public_key = tls_private_key.workers.public_key_openssh
}

resource "virtfoundry_vks_cluster" "this" {
  name               = var.cluster_name
  kubernetes_version = var.kubernetes_version

  workers = {
    count        = var.worker_count
    template_ref = var.template_ref
    offering_ref = var.offering_ref
    network_ref  = var.network_ref
    ssh_key_refs = [virtfoundry_ssh_key.workers.name]
  }

  # Optional: omit to use platform defaults.
  control_plane = {
    service_type = "LoadBalancer"
  }
}

# Requires the vks:kubeconfig permission. The kubeconfig is sensitive and lands in state.
data "virtfoundry_vks_kubeconfig" "this" {
  name       = virtfoundry_vks_cluster.this.name
  depends_on = [virtfoundry_vks_cluster.this]
}
