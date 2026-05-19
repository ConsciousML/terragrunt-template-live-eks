# Tailscale Bootstrap

Creates a Tailscale WIF credential and writes `TS_OAUTH_CLIENT_ID`, `TS_AUDIENCE`, and `TS_TAGS` to the live repo's GitHub secrets so CI can authenticate to Tailscale during Terratest runs. See the [catalog README](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/pipelines/bootstrap/tailscale/README.md) for the full Tailscale setup flow.

## Prerequisites
- Follow the [installation instructions](../../../README.md#installation)
- Same [prerequisites](../../../README.md#prerequisites) as in the main `README.md`

Create an account and login at [https://login.tailscale.com/admin/welcome](https://login.tailscale.com/admin/welcome).

Download and install the [Tailscale client](https://tailscale.com/download).

Set up `TAILSCALE_OAUTH_CLIENT_ID` and `TAILSCALE_OAUTH_CLIENT_SECRET` following the [environment variables guide](../../../docs/environment-variables.md#tailscale_oauth_client_id-and-tailscale_oauth_client_secret).

## Deploy
Run once before running Terratest in CI:

```bash
source .env
cd live/bootstrap/tailscale
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive
```
