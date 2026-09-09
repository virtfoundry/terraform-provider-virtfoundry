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

variable "tenant_id" {
  type        = string
  description = "VirtFoundry tenant UUID"
}

variable "template_id" {
  type = string
}

variable "service_offering_id" {
  type = string
}

variable "security_group_id" {
  type = string
}

variable "vm_name" {
  type = string
}

variable "volume_name" {
  type = string
}
