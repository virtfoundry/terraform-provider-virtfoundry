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
  type        = string
  description = "VM template UUID (e.g. cirros)"
}

variable "service_offering_id" {
  type        = string
  description = "Service offering UUID (e.g. small)"
}

variable "security_group_id" {
  type        = string
  description = "Default security group UUID for public IP access"
}

variable "vm_name" {
  type        = string
  description = "Unique VM slug for this test run"
}
