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
  description = "Target tenant UUID"
}

variable "template_name" {
  type        = string
  default     = "ubuntu-2204"
  description = "VM template catalog name or UUID (e.g. ubuntu-2204, cirros, fedora-39)"
}

variable "service_offering_name" {
  type        = string
  default     = "small"
  description = "Service offering catalog name or UUID (e.g. small, medium, large)"
}

variable "security_group_id" {
  type        = string
  description = "Security group UUID for public IP access"
}

variable "vm_name" {
  type    = string
  default = "tf-demo-vm"
}
