# Slack Bootstrap

Registers the Slack bot token as a GitHub Actions secret, and creates each environment's Slack channels that CI-deployed Alertmanager instances post to.

See the [catalog README](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/pipelines/bootstrap/slack/README.md) for the full Slack app setup flow.

## Structure

```
live/bootstrap/slack/
  gh_secret/
    terragrunt.stack.hcl
  channels.hcl
  channels/
    staging/
      environment.hcl        ← environment = "staging"
      stack/
        terragrunt.stack.hcl
    prod/
      environment.hcl        ← environment = "prod"
      stack/
        terragrunt.stack.hcl
```

## Prerequisites

Perform the [quickstart](../../../README.md#getting-started) up to `Authenticate with AWS` (included).

Create a Slack app following the catalog README linked above, then set up `GITHUB_TOKEN` and `SLACK_BOT_TOKEN` following the [environment variables guide](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/docs/environment-variables.md).

## Deploy

Run `gh_secret` once, it's environment-independent:

```bash
source .env
cd live/bootstrap/slack/gh_secret
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive --no-stack-generate
```

Repeat the following for each environment (replacing `<environment>` by `staging` and then by `prod`):

```bash
source .env
cd live/bootstrap/slack/channels/<environment>/stack
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive --no-stack-generate
```

The bot is a member of each channel it creates by default, but you aren't. Join them from the Slack client following the catalog README's instructions.

## Module Details

See the [`units/slack`](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/units/slack/README.md) group README for what each unit provisions and how they compose.
