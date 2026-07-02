---
name: Bug report
about: Report a problem with the provider
title: ""
labels: bug
assignees: ""
---

## Description

A clear and concise description of what the problem is.

## Terraform/OpenTofu configuration

```hcl
# Paste the relevant provider block and resource(s) here.
# Redact any secrets (username, password, tokens, etc.).
```

## Steps to reproduce

1.
2.
3.

## Expected behavior

What you expected to happen.

## Actual behavior

What actually happened. Include the exact error message or diagnostic
text if there is one — the specific wording often points to a different
root cause.

## Environment

- Terraform or OpenTofu version (`terraform version` / `tofu version`):
- `terraform-provider-uptimekuma` version:
- Uptime Kuma version:
- Uptime Kuma deployment (self-hosted, Docker, PikaPods, other):
- Reverse proxy in front of Uptime Kuma, if any (Traefik, nginx,
  Cloudflare, none, ...), and whether it forwards WebSocket upgrades:

## Debug logs

Re-run with debug logging enabled and paste the relevant output below:

```bash
TF_LOG=DEBUG SOCKETIO_LOG_LEVEL=DEBUG terraform plan
```

```text
# Paste the relevant log output here.
```

## Additional context

Add any other context about the problem here.
