# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-09-19

### Security

- **BREAKING:** `virtfoundry_user.password` and `virtfoundry_tenant.admin_password` are now `WriteOnly` (TF >=1.11, framework v1.16+) — never written to state/plan. See `README.md` Security — State Encryption.
- `virtfoundry_api_key.secret` and `virtfoundry_ssh_key.private_key_pem` are write-once (managed) — set only on `Create`, nulled on `Read`/`Refresh` so `terraform show -json` after refresh is `null`. Prefer BYO `public_key`/`tls_private_key` over `generate = true`.
- Add ephemeral resources `virtfoundry_api_key` and `virtfoundry_ssh_key` (TF >=1.10) — secrets only in memory, auto-deleted on `Close`.
- Examples: `full-stack` now uses `tls_private_key` BYO, removed default `http`/`virtfoundry` passwords, bumped `required_version` to `>=1.11` and provider `~>0.3`.

### Changed

- Documentation: `README.md` `Security — State Encryption` section + state encryption example, `docs/resources` security warnings, `docs/ephemeral-resources` for ephemeral keys.

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
