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

variable "vpc_name" {
  type    = string
  default = "tf-demo-vpc"
}

variable "vpc_cidr" {
  type    = string
  default = "10.42.0.0/16"
}

variable "network_name" {
  type    = string
  default = "tf-demo-net"
}

variable "network_prefix" {
  type    = number
  default = 24
}

variable "security_group_name" {
  type    = string
  default = "tf-demo-ssh"
}

variable "ssh_key_name" {
  type    = string
  default = "tf-demo-key"
}

variable "template_name" {
  type    = string
  default = "cirros"
}

variable "service_offering_name" {
  type    = string
  default = "small"
}

variable "vm_name" {
  type    = string
  default = "tf-demo-vm"
}

variable "vm_display_name" {
  type    = string
  default = "Terraform full-stack demo"
}

variable "expose_ssh" {
  type        = bool
  default     = false
  description = "Expose SSH via NodePort (optional; requires API permission)"
}
