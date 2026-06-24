# Tailscale Bootstrap

Creates a Tailscale WIF credential and writes `TS_OAUTH_CLIENT_ID`, `TS_AUDIENCE`, and `TS_TAGS` to the live repo's GitHub secrets so CI can authenticate to Tailscale during Terratest runs.

See the [catalog README](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/pipelines/bootstrap/tailscale/README.md) for the full Tailscale setup flow.

## Prerequisites

Perform the [quickstart](../../../README.md#getting-started) up to `Authenticate with AWS` (included).

Create an account and login at [https://login.tailscale.com/admin/welcome](https://login.tailscale.com/admin/welcome).

Download and install the [Tailscale client](https://tailscale.com/download).

Set up `GITHUB_TOKEN`, `TAILSCALE_OAUTH_CLIENT_ID`, and `TAILSCALE_OAUTH_CLIENT_SECRET` following the [environment variables guide](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/docs/environment-variables.md).

## Deploy
Run once before running Terratest in CI:

```bash
source .env
cd live/bootstrap/tailscale
terragrunt stack run init
terragrunt run --all apply --backend-bootstrap --non-interactive --no-stack-generate
```
