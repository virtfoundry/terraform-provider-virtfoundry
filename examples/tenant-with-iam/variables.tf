variable "endpoint" {
  type    = string
  default = "https://virtfoundry.example.com"
}

variable "root_username" {
  type    = string
  default = "root"
}

variable "root_password" {
  type      = string
  sensitive = true
  # No default — must be provided via TF_VAR_root_password or var file
}

variable "create_tenant" {
  type        = bool
  default     = false
  description = "When true, creates a new tenant via modules/tenant (root only). When false, uses tenant_id."
}

variable "tenant_id" {
  type        = string
  default     = ""
  description = "Existing tenant UUID when create_tenant is false."
}

variable "tenant_name" {
  type    = string
  default = "tf-demo"
}

variable "tenant_slug" {
  type    = string
  default = null
}

variable "tenant_admin_password" {
  type      = string
  sensitive = true
  default   = null
}

variable "operator_username" {
  type    = string
  default = "tf-operator"
}

variable "operator_password" {
  type      = string
  sensitive = true
}
