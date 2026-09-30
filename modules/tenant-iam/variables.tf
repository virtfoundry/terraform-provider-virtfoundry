variable "tenant_id" {
  type        = string
  description = "Target tenant UUID (from module.tenant or an existing tenant)."
}

variable "roles" {
  type = map(object({
    description = optional(string, "")
    permissions = list(string)
  }))
  default     = {}
  description = "Custom IAM roles keyed by role name."
}

variable "users" {
  type = map(object({
    username  = string
    password  = string
    email     = optional(string, "")
    role_name = optional(string)
    role_id   = optional(string)
    state     = optional(string, "active")
  }))
  default     = {}
  sensitive   = true
  description = "Tenant users keyed by a stable map key (not username). Marked sensitive so terraform plan does not print passwords."
}

variable "api_keys" {
  type = map(object({
    name            = string
    user_key        = optional(string)
    user_id         = optional(string)
    expires_in_days = optional(number, 90)
    scopes          = list(string)
  }))
  default     = {}
  description = "API keys keyed by map key. Assign to a user via user_key or user_id. scopes is required (non-empty); expires_in_days defaults to 90."
}
