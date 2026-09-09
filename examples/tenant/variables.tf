variable "endpoint" {
  type    = string
  default = "http://virtfoundry.homelab"
}

variable "username" {
  type    = string
  default = "root"
}

variable "password" {
  type      = string
  sensitive = true
  default   = "virtfoundry"
}

variable "tenant_name" {
  type = string
}

variable "tenant_slug" {
  type = string
}

variable "admin_password" {
  type      = string
  sensitive = true
  default   = "virtfoundry"
}
