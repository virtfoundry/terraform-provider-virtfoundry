output "id" {
  description = "Tenant UUID."
  value       = virtfoundry_tenant.this.id
}

output "name" {
  description = "Tenant display name."
  value       = virtfoundry_tenant.this.name
}

output "slug" {
  description = "Tenant slug."
  value       = virtfoundry_tenant.this.slug
}

output "namespace" {
  description = "Kubernetes namespace for tenant workloads."
  value       = virtfoundry_tenant.this.namespace
}
