terraform {
  required_version = ">= 1.11"

  required_providers {
    virtfoundry = {
      source                = "virtfoundry/virtfoundry"
      configuration_aliases = [virtfoundry]
    }
  }
}
