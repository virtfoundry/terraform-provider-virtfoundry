terraform {
  required_version = ">= 1.0"

  required_providers {
    virtfoundry = {
      source  = "virtfoundry/virtfoundry"
      version = "0.1.0"
    }
  }
}

provider "virtfoundry" {
  endpoint = var.endpoint
  username = var.username
  password = var.password
}

variable "endpoint" {
  type        = string
  description = "VirtFoundry API URL"
  default     = "http://virtfoundry.homelab"
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
