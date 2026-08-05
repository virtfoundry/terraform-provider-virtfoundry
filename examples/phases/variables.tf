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

variable "offering_name" {
  type    = string
  default = "tf-phase-offering"
}

variable "offering_display_name" {
  type    = string
  default = "Terraform phases test offering"
}

variable "offering_cpu" {
  type    = number
  default = 2
}

variable "offering_memory_mi" {
  type    = number
  default = 2048
}

variable "baseline_offering_name" {
  type    = string
  default = "small"
}

variable "use_custom_offering" {
  type        = bool
  default     = false
  description = "When true, VM uses the custom service offering (for resize test after stop)."
}

variable "template_name" {
  type    = string
  default = "tf-phase-template"
}

variable "template_display_name" {
  type    = string
  default = "Terraform phases test template"
}

variable "template_image" {
  type    = string
  default = "quay.io/kubevirt/cirros-container-disk-demo:latest"
}

variable "volume_name" {
  type    = string
  default = "tf-phase-vol"
}

variable "volume_size_gi" {
  type    = number
  default = 1
}

variable "vm_name" {
  type    = string
  default = "tf-phase-vm"
}

variable "vm_display_name" {
  type    = string
  default = "Terraform phases VM"
}

variable "desired_state" {
  type    = string
  default = "running"
}
