# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `virtfoundry_service_offering` resource (root-only CRUD)
- `virtfoundry_vm_volume_attachment` resource (hot-plug / detach)
- `examples/phases/` stack and `scripts/test-phases.sh` integration test

### Changed

- `virtfoundry_vm` `template_id` and `service_offering_id` accept catalog **names** (e.g. `ubuntu-2204`, `small`) in addition to UUIDs
- Document built-in template catalog in `docs/resources/vm.md` and `docs/data-sources/vm_templates.md`
- `virtfoundry_vm_template` waits for ISO CDI import (`import_state`, `wait_for_import`)
- `examples/iso-windows/` and `scripts/test-iso-windows.sh` for Windows ISO flow
- `virtfoundry_volume` destroy calls `DELETE /volumes/{id}` (was state-only)
- `virtfoundry_vm` supports in-place `service_offering_id` updates (stops VM before resize)
- Fix `virtfoundry_security_groups` data source schema/model mismatch

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
