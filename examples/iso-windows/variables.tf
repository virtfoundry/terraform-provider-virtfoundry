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
  type = string
}

variable "create_iso_template" {
  type        = bool
  default     = true
  description = "When false, use existing_template_name (e.g. platform windows-server-2022)."
}

variable "existing_template_name" {
  type        = string
  default     = "windows-server-2022"
  description = "Catalog template name when create_iso_template is false."
}

variable "template_name" {
  type    = string
  default = "tf-win-iso"
}

variable "template_display_name" {
  type    = string
  default = "Windows Server 2022 Eval (Terraform)"
}

variable "iso_url" {
  type        = string
  default     = "https://go.microsoft.com/fwlink/?linkid=2195280"
  description = "Windows Server 2022 eval ISO URL (CDI HTTP import)."
}

variable "iso_size_gi" {
  type    = number
  default = 8
}

variable "boot_disk_size_gi" {
  type    = number
  default = 32
}

variable "import_wait_timeout_minutes" {
  type    = number
  default = 45
}

variable "deploy_vm" {
  type        = bool
  default     = false
  description = "Deploy installer VM after template is ready (Windows setup is manual via console)."
}

variable "service_offering_name" {
  type    = string
  default = "windows-large"
}

variable "security_group_id" {
  type = string
}

variable "vm_name" {
  type    = string
  default = "tf-win-install"
}

variable "vm_display_name" {
  type    = string
  default = "Windows install VM"
}

variable "desired_state" {
  type    = string
  default = "running"
}
