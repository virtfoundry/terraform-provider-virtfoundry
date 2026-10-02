variable "endpoint" {
  type    = string
  default = "https://virtfoundry.example.com"
}

variable "username" {
  type    = string
  default = "root"
}

variable "password" {
  type      = string
  sensitive = true
  # No default — provide via TF_VAR_password or a var file
}

variable "tenant_id" {
  type        = string
  description = "VirtFoundry tenant UUID"
}

variable "cluster_name" {
  type        = string
  description = "VKS cluster name (DNS label)"
  default     = "dev"
}

variable "kubernetes_version" {
  type        = string
  description = "Kubernetes version; must match a published node image (see vks-image-factory COMPATIBILITY)"
  default     = "v1.31.4"
}

variable "worker_count" {
  type    = number
  default = 2
}

variable "template_ref" {
  type        = string
  description = "Name of the node image VM template matching kubernetes_version"
}

variable "offering_ref" {
  type        = string
  description = "Service offering name for workers"
  default     = "medium"
}

variable "network_ref" {
  type        = string
  description = "Network name the workers attach to"
}
