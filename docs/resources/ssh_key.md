---
page_title: "virtfoundry_ssh_key Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry tenant SSH key. Register an existing public key or generate a new Ed25519 pair.
---

# virtfoundry_ssh_key (Resource)

Register an existing SSH public key or generate a new Ed25519 key pair via the API. Attach keys to VMs with `ssh_key_id`.

> **Security:** Prefer `public_key` (BYO) with `tls_private_key` generated outside Terraform. `generate = true` returns `private_key_pem` only at `Create` and it is **not persisted after refresh** (`show -json` → `null`); still, the PEM briefly lives in state. Enable state encryption. See provider README Security.

## Example Usage

```hcl
# Recommended: bring your own public key (private key never touches state)
resource "tls_private_key" "deploy" {
  algorithm = "ED25519"
}
resource "virtfoundry_ssh_key" "deploy" {
  name       = "deploy"
  public_key = tls_private_key.deploy.public_key_openssh
}

# Or generate via API (PEM briefly in state, nulled after refresh — prefer BYO above)
resource "virtfoundry_ssh_key" "admin" {
  name     = "admin"
  generate = true
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Key name within the tenant. |
| `public_key` | String | no | OpenSSH `authorized_keys` line. Omit when `generate = true`. |
| `generate` | Boolean | no | Generate a new key pair via the API. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | SSH key UUID. |
| `private_key_pem` | Generated private key PEM (sensitive; only when `generate = true`, not persisted after refresh). Prefer BYO `public_key`. |
| `fingerprint` | Key fingerprint. |

## Import

```shell
terraform import virtfoundry_ssh_key.admin <tenant_id>/<key_id>
```
