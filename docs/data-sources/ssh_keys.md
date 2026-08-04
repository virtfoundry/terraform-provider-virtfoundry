---
page_title: "virtfoundry_ssh_keys Data Source - virtfoundry"
subcategory: ""
description: |-
  Lists SSH keys in a tenant.
---

# virtfoundry_ssh_keys (Data Source)

Lists SSH public keys registered in the tenant.

## Example Usage

```hcl
data "virtfoundry_ssh_keys" "all" {}

output "key_ids" {
  value = [for k in data.virtfoundry_ssh_keys.all.ssh_keys : k.id]
}
```

## Argument Reference

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tenant_id` | String | no | Tenant UUID. Defaults to provider `tenant_id`. |

## Attribute Reference

| Name | Description |
|------|-------------|
| `ssh_keys` | List of objects with `id`, `name`, `fingerprint`. |
