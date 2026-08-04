# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-08-04

### Added

- Full resource coverage: tenant, VPC, network, security group, volume, snapshots, VM template, VM, SSH key, IAM (user, role, API key)
- Data sources: service offerings, VM templates, VPCs, networks, security groups, SSH keys, users, roles
- Registry documentation (`docs/`) with examples for every resource and data source
- Example stacks: `provider`, `vm`, `full-stack`, `tenant-with-iam`
- Reusable modules: `tenant`, `tenant-iam`
- Integration test scripts for VM and full-stack workflows

### Changed

- Professional README and Registry overview with quick-start examples
- Example provider configurations updated to `~> 0.2`

## [0.1.0] - 2026-08-04

### Added

- Initial Terraform Registry release
- `virtfoundry_vm` resource

[0.2.0]: https://github.com/virtfoundry/terraform-provider-virtfoundry/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/virtfoundry/terraform-provider-virtfoundry/releases/tag/v0.1.0
