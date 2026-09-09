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

variable "offering_name" {
  type = string
}

variable "offering_display_name" {
  type    = string
  default = "Terraform E2E offering"
}
