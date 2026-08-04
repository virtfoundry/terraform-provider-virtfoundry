variable "endpoint" {
  type        = string
  description = "VirtFoundry API endpoint URL."
  default     = "https://virtfoundry.example.com"
}

variable "api_key" {
  type        = string
  description = "API key for authentication."
  sensitive   = true
  default     = null
}

variable "tenant_id" {
  type        = string
  description = "Default tenant UUID."
  default     = null
}

variable "insecure" {
  type        = bool
  description = "Skip TLS verification (development only)."
  default     = false
}
