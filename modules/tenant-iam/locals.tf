locals {
  # Preset permission bundles aligned with VirtFoundry core IAM defaults.
  permission_presets = {
    admin = [
      "users:read", "users:write",
      "vpcs:read", "vpcs:write",
      "networks:read", "networks:write",
      "security_groups:read", "security_groups:write",
      "volumes:read", "volumes:write",
      "vms:read", "vms:write", "vms:console",
      "ssh_keys:read", "ssh_keys:write",
    ]
    operator = [
      "vpcs:read", "vpcs:write",
      "networks:read", "networks:write",
      "security_groups:read", "security_groups:write",
      "volumes:read", "volumes:write",
      "vms:read", "vms:write", "vms:console",
      "ssh_keys:read", "ssh_keys:write",
    ]
    viewer = [
      "vpcs:read", "networks:read", "security_groups:read",
      "volumes:read", "vms:read", "ssh_keys:read",
    ]
  }

  resolved_roles = {
    for name, cfg in var.roles : name => merge(cfg, {
      permissions = length(cfg.permissions) > 0 ? cfg.permissions : lookup(local.permission_presets, name, cfg.permissions)
    })
  }

  user_role_ids = {
    for key, u in var.users : key => coalesce(
      u.role_id,
      try(virtfoundry_role.this[u.role_name].id, null),
    )
  }
}
