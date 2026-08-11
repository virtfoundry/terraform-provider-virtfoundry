# Security Policy

Please follow the security process documented in [virtfoundry/core SECURITY.md](https://github.com/virtfoundry/core/blob/main/SECURITY.md).

**Do not open public GitHub issues for security vulnerabilities.**

Provider-specific notes:

- Prefer environment variables or a secrets manager for API keys (`vfd_live_...`)
- Never commit live credentials in examples or CI
