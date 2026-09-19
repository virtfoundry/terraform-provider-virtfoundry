---
page_title: "virtfoundry_ssh_key Ephemeral Resource - virtfoundry"
subcategory: ""
description: |-
  Ephemeral VirtFoundry SSH key. The private key is never persisted in state.
---

# virtfoundry_ssh_key (Ephemeral Resource)

Ephemeral SSH key for provisioners or short-lived access. When `generate = true` the private key PEM is exposed only in the ephemeral result and the key is **deleted on Close**. For persistent keys register a `public_key` via the managed `virtfoundry_ssh_key` resource and generate the key outside Terraform (`tls_private_key`).

Requires Terraform >= 1.10 and provider `>= 0.3`.

## Example Usage

```hcl
# Ephemeral generated key — PEM never touches state
ephemeral "virtfoundry_ssh_key" "ephemeral" {
  name     = "ephemeral-deploy"
  generate = true
}

# Use elsewhere (e.g., provisioner):
# private_key = ephemeral.virtfoundry_ssh_key.ephemeral.private_key_pem

# Recommended persistent BYO (managed resource):
resource "tls_private_key" "deploy" {
  algorithm = "ED25519"
}
resource "virtfoundry_ssh_key" "deploy" {
  name       = "deploy"
  public_key = tls_private_key.deploy.public_key_openssh
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Key name within the tenant. |
| `public_key` | String | no | OpenSSH `authorized_keys` line. Omit when `generate = true`. |
| `generate` | Boolean | no | Generate Ed25519 key pair via API (ephemeral). |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | SSH key UUID. |
| `private_key_pem` | Generated private key PEM (sensitive, ephemeral only). |
| `fingerprint` | Key fingerprint. |
