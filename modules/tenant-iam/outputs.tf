output "role_ids" {
  description = "Map of role name to role UUID."
  value       = { for k, r in virtfoundry_role.this : k => r.id }
}

output "user_ids" {
  description = "Map of user map-key to user UUID."
  value       = { for k, u in virtfoundry_user.this : k => u.id }
}

output "usernames" {
  description = "Map of user map-key to username."
  value       = { for k, u in virtfoundry_user.this : k => u.username }
}

output "api_key_ids" {
  description = "Map of api_key map-key to key UUID."
  value       = { for k, key in virtfoundry_api_key.this : k => key.id }
}

output "api_key_secrets" {
  description = "Map of api_key map-key to secret (only available at creation)."
  value       = { for k, key in virtfoundry_api_key.this : k => key.secret }
  sensitive   = true
}
