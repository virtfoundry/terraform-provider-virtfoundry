output "phase" {
  value = virtfoundry_vks_cluster.this.phase
}

output "control_plane_endpoint" {
  value = virtfoundry_vks_cluster.this.control_plane_endpoint
}

output "ready_workers" {
  value = virtfoundry_vks_cluster.this.ready_workers
}

output "kubeconfig" {
  value     = data.virtfoundry_vks_kubeconfig.this.kubeconfig
  sensitive = true
}
