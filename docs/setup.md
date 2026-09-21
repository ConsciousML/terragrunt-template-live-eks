{/* This doc is aggregated into the EKS Forge documentation site: https://eks-forge.readthedocs.io/latest/. It is not meant to be read directly in this repository. */}
# Live Repository Setup

In this tutorial, you'll fork the [live repository](https://github.com/ConsciousML/terragrunt-template-live-eks), install its CLI tools, and point it at your catalog fork.

## Prerequisites
Complete the [quickstart](/docs/quickstart/) first. Live reuses its [prerequisites](/docs/quickstart/prerequisites/), bootstrap resources, and catalog fork.

## Fork the Live Repository
The live repository deploys the [catalog stacks](/docs/architecture/) in the `staging` and `prod` environments.
Like the catalog, it's meant to be forked and extended.

Follow the same steps as when you [forked the catalog](/docs/quickstart/installation/#fork-the-eks-forge-catalog). First, [create an empty repository](https://github.com/new) on GitHub, private or public. Leave the README, `.gitignore`, and license options unset.

Then, set your GitHub owner (user or organization) and the name of the repository you created, by replacing the `<...>`:
```bash
export GITHUB_OWNER=<your-github-owner>
export LIVE_REPO_NAME=<your-live-repo-name>
```

Clone the live repository and push it to your repository:
```bash
git clone https://github.com/ConsciousML/terragrunt-template-live-eks.git $LIVE_REPO_NAME
cd $LIVE_REPO_NAME
git remote set-url origin git@github.com:$GITHUB_OWNER/$LIVE_REPO_NAME.git
git push origin main
git push origin --tags
```

## Install the CLI Tools
Stay at the root of your live repository.

Live uses [mise-en-place](https://mise.jdx.dev/) like the catalog, with a slightly different tool set pinned in [`mise.toml`](../mise.toml) and [`mise.local.toml`](../mise.local.toml). Follow the [CLI tools installation steps](/docs/quickstart/installation/#install-the-cli-tools) again, this time from your live fork.

## Live Configuration
Live reads `.hcl` configuration files under [`live/`](../live/), like the catalog's `pipelines/`. They point to the catalog's [units](/docs/iac/#units), so `staging` and `prod` use the same components you deployed in [`dev`](/docs/iac/#dev).

Edit [`live/github.hcl`](../live/github.hcl) by replacing the `<...>`:
```hcl
locals {
  github_owner_catalog         = "<your-github-username-or-org-name-where-your-catalog-fork-is>"
  github_owner_live            = "<your-github-username-or-org-name-where-your-live-fork-is>"
  github_repo_name_catalog     = "<your-catalog-repo-name>"
  github_repo_name_live        = "<your-live-repo-name>"
}
```
This makes your live fork use the modules and units of your catalog fork instead of the original catalog repository.

Each [environment](/docs/iac/#environments) sets its AWS region in its own `region.hcl` file:
- [`live/staging/region.hcl`](../live/staging/region.hcl)
- [`live/prod/region.hcl`](../live/prod/region.hcl)
- [`live/bootstrap/region.hcl`](../live/bootstrap/region.hcl)

Like when you [configured the catalog](/docs/quickstart/configuration/#catalog-configuration), set `region` and `azs` in the `staging` and `prod` files:
```hcl
locals {
  region = "us-east-1"
  azs    = ["us-east-1a", "us-east-1b", "us-east-1c"]
}
```
The `bootstrap` file has no `azs`, so you only need to set `region`.

You can also leave them as is if you plan to use `us-east-1`.

## Bootstrap
Run the bootstrap pipelines once from your [live fork](/docs/deployment/live-repository-setup/#fork-the-live-repository), as you did from your [catalog fork](/docs/quickstart/installation/#fork-the-eks-forge-catalog) in the [quickstart bootstrap](/docs/quickstart/bootstrap/).

They read the same environment variables as the catalog. Copy your catalog `.env` to the root of your live fork:
```bash
cp <path-to-your-catalog-fork>/.env .env
```

:::warning
These pipelines need to be performed only once per live fork before deploying to [`staging`](/docs/deployment/deploy-to-staging/) and [`prod`](/docs/deployment/promote-to-production/).
:::

### Setup DNS
This pipeline creates a public hosted zone for `staging` and for `prod`.

In [`live/dns.hcl`](../live/dns.hcl), set `base_domain` to the domain you used in the catalog:
```hcl
locals {
  base_domain = "yourdomain.com"
  ...
}
```

:::info
Each subdirectory provisions its own hosted zone, giving that environment its own subdomain (e.g. `staging.yourdomain.com`):
```
setup_dns/
├── staging/
└── prod/
```
:::

Next, run the following for each environment (replacing `<environment>` with `staging` and then `prod`), from the root of your live fork:
```bash
cd live/bootstrap/setup_dns/<environment>/stack
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive --no-stack-generate
terragrunt stack output --json setup_dns.route53_hosted_zone.name_servers
```

You should see something similar to:
```txt
{
  "setup_dns": {
    "route53_hosted_zone": {
      "name_servers": [
        "<nameserver_1>",
        "<nameserver_2>",
        "<nameserver_3>",
        "<nameserver_4>",
      ]
    }
  }
}
```

For each environment, add 4 NS records for the environment subdomain in your domain registrar, using the nameservers from the output:

| Type | Host | Value |
|------|------|-------|
| NS | `<environment>` | `ns-123.awsdns-12.com` |
| NS | `<environment>` | `ns-456.awsdns-34.net` |
| NS | `<environment>` | `ns-789.awsdns-56.org` |
| NS | `<environment>` | `ns-012.awsdns-78.co.uk` |

Verify that the NS records are propagated for each environment:
```bash
dig NS <environment>.yourdomain.com
```

Delegation is working when 4 AWS nameservers appear in the `ANSWER SECTION`.

For more information, read the [Setup DNS quickstart](/docs/quickstart/bootstrap/setup_dns/).

### AWS GitHub Actions Authentication
This pipeline lets your live fork's CI/CD authenticate with AWS. It reuses the OIDC provider created from your catalog fork.

From the root of your live fork, run:
```bash
source .env
cd live/bootstrap/aws_gh_actions_auth
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive --no-stack-generate
```

For more information, read the [AWS GitHub Actions Authentication quickstart](/docs/quickstart/bootstrap/aws_gh_actions_auth/).

### Tailscale
This pipeline lets your live fork's CI authenticate to Tailscale. It reuses the tailnet ACL applied from your catalog fork.

From the root of your live fork, run:
```bash
source .env
cd live/bootstrap/tailscale
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive --no-stack-generate
```

For more information, read the [Tailscale quickstart](/docs/quickstart/bootstrap/tailscale/).

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

Finally, join the `staging-` and `prod-` prefixed channels, as you did for the `dev-` ones in the [Slack quickstart](/docs/quickstart/bootstrap/slack/) (see [`live/bootstrap/slack/channels.hcl`](../live/bootstrap/slack/channels.hcl) for the base names).
