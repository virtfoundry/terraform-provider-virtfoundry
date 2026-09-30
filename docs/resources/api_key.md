---
page_title: "virtfoundry_api_key Resource - virtfoundry"
subcategory: ""
description: |-
  Manages a VirtFoundry API key. The secret is only available at creation time.
---

# virtfoundry_api_key (Resource)

Creates an API key for programmatic access. The full secret (`vfd_live_...`) is returned **once** at creation — store it securely.

> **Security:** `secret` is sensitive and only set at `Create`. It **briefly lives in state until the first `terraform refresh`/`apply -refresh-only`**, then is nulled (`terraform show -json` → `null`). For zero-state use `ephemeral "virtfoundry_api_key"` (TF >=1.10). Treat state as sensitive and enable [state encryption](https://developer.hashicorp.com/terraform/language/state/encryption) (TF >=1.11). See provider README Security section.

## Defaults (expiry and scopes)

| Argument | Default / required | Notes |
|----------|-------------------|--------|
| `expires_in_days` | **90** | Provider default TTL. Without a TTL, the VirtFoundry API treats keys as non-expiring. |
| `scopes` | **required** | Must be an explicit non-empty list. Historically, omitting scopes on the API meant all caller permissions (`["*"]`); the provider no longer allows that omission. |

## Example Usage

```hcl
resource "virtfoundry_api_key" "ci" {
  name            = "github-actions"
  expires_in_days = 90
  scopes          = ["vms:read", "vms:write"]
}

output "api_key_secret" {
  value     = virtfoundry_api_key.ci.secret
  sensitive = true
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | yes | Key name. Forces replacement. |
| `user_id` | String | no | Owner user UUID. Defaults to the authenticated user. |
| `expires_in_days` | Number | no | Expiration in days (default **90**). Forces replacement. |
| `scopes` | List(String) | yes | Permission scopes. Forces replacement. Empty / omitted historically = all caller perms; provider requires an explicit list. |
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `id` | API key UUID. |
| `prefix` | Key prefix for identification. |
| `secret` | Full API key secret (sensitive; only at create, not persisted after refresh). |

## Import

```shell
terraform import virtfoundry_api_key.ci <tenant_id>/<key_id>
```

> **Note:** Imported keys do not expose the secret. Rotate the key if the secret was lost.
