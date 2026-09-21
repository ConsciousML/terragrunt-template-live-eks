{/* This doc is aggregated into the EKS Forge documentation site: https://eks-forge.readthedocs.io/latest/. It is not meant to be read directly in this repository. */}
### Slack
This pipeline adds your Slack bot token to your live fork's GitHub secrets and creates the `staging` and `prod` alert channels.

From the root of your live fork, run the `gh_secret` stack once:
```bash
source .env
cd live/bootstrap/slack/gh_secret
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive --no-stack-generate
```

Then, from the root of your live fork, run the following for each environment (replacing `<environment>` with `staging` and then `prod`):
```bash
source .env
cd live/bootstrap/slack/channels/<environment>/stack
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive --no-stack-generate
```

Finally, join the `staging-` and `prod-` prefixed channels, as you did for the `dev-` ones in the [Slack quickstart](/docs/quickstart/bootstrap/slack/) (see [`live/bootstrap/slack/channels.hcl`](channels.hcl) for the base names).
