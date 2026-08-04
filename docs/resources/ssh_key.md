---
page_title: "virtfoundry_ssh_key Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry tenant SSH key. Register an existing public key or generate a new Ed25519 pair.
---

# virtfoundry_ssh_key (Resource)

Register an existing SSH public key or generate a new Ed25519 key pair via the API. Attach keys to VMs with `ssh_key_id`.

## Example Usage

```hcl
# Generate a new key pair
resource "virtfoundry_ssh_key" "admin" {
  name     = "admin"
  generate = true
}

# Or register an existing public key
resource "virtfoundry_ssh_key" "deploy" {
  name       = "deploy"
  public_key = file("~/.ssh/id_ed25519.pub")
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
| `private_key_pem` | Generated private key PEM (sensitive; only when `generate = true`). |
| `fingerprint` | Key fingerprint. |

## Import

```shell
terraform import virtfoundry_ssh_key.admin <tenant_id>/<key_id>
```
