# Contributing

See [virtfoundry/core CONTRIBUTING](https://github.com/virtfoundry/core/blob/main/CONTRIBUTING.md) for org-wide conventions (Conventional Commits, branch workflow).

## Provider-specific

- Go module: `github.com/virtfoundry/terraform-provider-virtfoundry`
- Keep the provider a **thin client** over `/api/v1` — no direct KubeVirt or Kubernetes calls
- Update [TERRAFORM-PROVIDER-PLAN.md](https://github.com/virtfoundry/core/blob/main/docs/TERRAFORM-PROVIDER-PLAN.md) when adding resources
- Run `make lint` before opening a PR

## Sync with core API

When the core API adds or changes a resource, update the mapping table in the plan doc and implement the matching Terraform resource in this repo.
