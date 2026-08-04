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
  api_key  = var.api_key
}

variable "endpoint" {
  type        = string
  description = "VirtFoundry API URL"
}

variable "api_key" {
  type        = string
  sensitive   = true
  description = "VirtFoundry API key (vfd_live_...)"
}
