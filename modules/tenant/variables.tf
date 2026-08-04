variable "name" {
  type        = string
  description = "Tenant display name."
}

variable "slug" {
  type        = string
  default     = null
  description = "URL-safe slug. Derived from name when null."
}

variable "admin_password" {
  type        = string
  default     = null
  sensitive   = true
  description = "Initial password for the tenant admin user."
}
