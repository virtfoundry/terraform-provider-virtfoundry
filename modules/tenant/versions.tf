terraform {
  required_version = ">= 1.0"

  required_providers {
    virtfoundry = {
      source                = "virtfoundry/virtfoundry"
      configuration_aliases = [virtfoundry]
    }
  }
}
