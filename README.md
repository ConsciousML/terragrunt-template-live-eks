# Terragrunt Template Live AWS

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub Release](https://img.shields.io/github/release/ConsciousML/terragrunt-template-live-eks.svg?style=flat)]()
[![CI](https://github.com/ConsciousML/terragrunt-template-live-eks/actions/workflows/ci.yaml/badge.svg)](https://github.com/ConsciousML/terragrunt-template-live-eks/actions/workflows/ci.yaml)
[![PR's Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com)

A prod-ready Terragrunt Template for deploying multi-environment [EKS](https://aws.amazon.com/eks/) clusters on AWS

## Catalog vs Live Infrastructure

This is a **live repository** for deploying infrastructure across multiple environments.

This IaC production toolkit follows [Gruntwork's official patterns](https://github.com/gruntwork-io/terragrunt-infrastructure-live-example) by using two template repositories:
- **Catalog repository**: Defines **what** can be deployed (reusable components: [modules, units, and stacks](https://github.com/ConsciousML/terragrunt-template-catalog-eks))
- **This repository** (live): Defines **where** and **how** catalog components are deployed in `dev`, `staging`, and `prod` environments with CI/CD


## What's Inside
- Multi-environment IaC support: build EKS cluster across `dev`, `staging`, and `prod`.
- [CI](.github/workflows/ci.yaml) (on PR): Runs `terragrunt plan` on each environment, uploads output plan to PR, deploys on the staging environment, runs some tests, and destroys.
- [CD](.github/workflows/cd.yaml) (on push `main`): Automatically deploys on `prod`
- [Bootstrap pipeline](live/bootstrap/enable_tg_github_actions/) to automatically authenticate GitHub Actions with AWS.

You're new to Terragrunt best practices? Read [Gruntwork's official production patterns](https://github.com/gruntwork-io/terragrunt-infrastructure-catalog-example) to get the foundations required to use this extended repository.

## Getting Started
Follow the getting started of the [EKS catalog repository](https://github.com/ConsciousML/terragrunt-template-catalog-eks) first, as you'll need to configure it before using this live repository. 

### Prerequisites
- AWS account with billing enabled
- GitHub account
- AWS IAM permissions to manage IAM roles, VPC resources, EKS resources, compute resources and S3 (see `policy_arns` in the [bootstrap stack](live/bootstrap/enable_tg_github_actions/terragrunt.stack.hcl) for a list of the specific IAM policies)

### Fork the Repositories (catalog and live)
1. Fork the catalog repository by following its [Fork the Repository section](https://github.com/ConsciousML/terragrunt-template-catalog-eks/#fork-the-repository).
2. Fork this repository (live) by clicking on `Use this template`

### Configuration
1. In `live/github.hcl`, modify:
```hcl
locals {
  github_username_catalog  = "YourUsernameWhereYourCatalogForkIs"
  github_username_live     = "YourUsernameWhereYourLiveForkIs"
  github_repo_name_catalog = "your-catalog-repo-name"
  github_repo_name_live    = "your-live-repo-name"
}
```
If you've forked both repositories, `github_username_catalog` and `github_username_live` should point to your username (`ConsciousML` for my own forks).
2. Change each `live/*/region.hcl` to match your desired AWS region.

### Installation

**Option 1: Use `mise` (recommended)**

First, `cd` at the root of this repository. 

Next, install mise:
```bash
curl https://mise.run | MISE_VERSION=v2026.4.0 sh
```

Then, install all the tools in the `mise.toml` file:
```bash
mise trust
mise install
```

Finally, run the following to automatically activate mise when starting a shell:
- For zsh: 
```bash
echo 'eval "$(~/.local/bin/mise activate zsh)"' >> ~/.zshrc && source ~/.zshrc
```
- For bash:
```bash
echo 'eval "$(~/.local/bin/mise activate bash)"' >> ~/.bashrc && source ~/.bashrc
```

For more information on how to use mise, read their [getting started guide](https://mise.jdx.dev/getting-started.html).

**Option 2: Install Tools Manually**
- [Terragrunt](https://terragrunt.gruntwork.io/docs/getting-started/install/)
- [OpenTofu](https://opentofu.org/docs/intro/install/) (or [Terraform](https://developer.hashicorp.com/terraform/install))
- [Go](https://go.dev/doc/install)
- [Python 3.13.1](https://www.python.org/downloads/)
- [GitHub CLI](https://github.com/cli/cli#installation)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)

See [mise.toml](./mise.toml) for specific versions.

### Authenticate with AWS
Authenticate to the AWS CLI:
```
aws configure
```

For more information, read the [AWS CLI authentication documentation](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).

### Run the Bootstrap Pipelines
Run the following once before using CI/CD:
- [AWS GitHub Actions Auth](live/bootstrap/aws_gh_actions_auth/README.md): authenticates GitHub Actions with AWS
- [Setup DNS](live/bootstrap/setup_dns/README.md): creates a [Route53 hosted zone](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/hosted-zones-working-with.html) per environment to sign TLS certificates with ACM
- [Tailscale](live/bootstrap/tailscale/README.md): creates [Tailscale](https://tailscale.com/) resources needed for CI and to access internal cluster tools (ArgoCD, etc.)

### Deploy a Staging EKS Cluster
Deploy a stack that creates a VPC and an EKS cluster.

Ensure `TAILSCALE_OAUTH_CLIENT_ID` and `TAILSCALE_OAUTH_CLIENT_SECRET` are set in your `.env` (see the [environment variables guide](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/docs/environment-variables.md)).

```bash
source .env
cd live/staging
terragrunt stack run init
terragrunt run --all apply --backend-bootstrap --non-interactive
```

Go into the AWS console and check that your resources have been created.

After around 15 min, your `staging` EKS cluster will be created.

Connect `kubectl` to your EKS cluster by creating a `kubeconfig` (replace `<region-code>` and `<cluster-name>`):
```bash
aws eks update-kubeconfig --region <region-code> --name <cluster-name>
```

Next, verify `kubectl` is connected:
```
kubectl get pods -n kube-system
```

You should see an output similar to:
```text
NAME                           READY   STATUS    RESTARTS   AGE
aws-node-59ld8                 2/2     Running   0          41m
coredns-845b86cddf-pg8hk       1/1     Running   0          40m
eks-pod-identity-agent-9pq6k   1/1     Running   0          41m
...
```

### Log in to ArgoCD

ArgoCD is only reachable with the Tailscale Client running. Make sure you have completed the [Tailscale prerequisites](live/bootstrap/tailscale/README.md#prerequisites) before proceeding.

The ArgoCD host is formed from `live/dns.hcl` as `<subdomain>.staging.<base_domain>` (replace `<subdomain>` and `<base_domain>` with the values from that file, e.g. `argocd.staging.axelmendoza.com`).

**Web UI**: Open `https://<subdomain>.staging.<base_domain>` in your browser and log in with username `admin`. Retrieve the password with:
```bash
aws secretsmanager get-secret-value \
  --secret-id staging-argocd-password \
  --query SecretString \
  --output text | jq -r .plaintext
```

**CLI**: Log in directly in one command:
```bash
argocd login <subdomain>.staging.<base_domain> \
  --username admin \
  --password $(aws secretsmanager get-secret-value \
    --secret-id staging-argocd-password \
    --query SecretString \
    --output text | jq -r .plaintext)
```

Finally, cleanup by destroying the infrastructure (cwd in `live/staging/`):

```bash
terragrunt run --all destroy --non-interactive
```

## Deployment Workflow

Follow the [Using the CI/CD section](docs/ci_cd.md#using-the-cicd) to deploy your infrastructure on prod.

## CI/CD Pipelines

### CI (Pull Requests)
- Validates HCL formatting
- Runs `terragrunt plan` on dev, staging, and prod in parallel
- Generates production plan artifact for review
- Runs infrastructure tests
- Upload prod plan output and comments on PR

### CD (Merge to main)
- Automatically deploys to the production environment (i.e `prod`).

See the [CI/CD workflow guide](docs/ci_cd.md) for detailed setup instructions and usage.

## Testing

This repository includes infrastructure testing with [Terratest](https://terratest.gruntwork.io/).

The test in `tests/staging_stack_test.go` deploys the staging environment, validates the infrastructure, and automatically destroys resources.

To extend testing, add Go code to validate your deployed infrastructure:
```go
// Example: Test that a EC2 instance is accessible
instanceIP := // ... get instance IP from AWS
resp, err := http.Get(fmt.Sprintf("http://%s", instanceIP))
require.NoError(t, err)
require.Equal(t, 200, resp.StatusCode)
```

Tests run automatically in CI when the `run-terratest` label is added to your PR. See the [CI/CD guide](docs/ci_cd.md) for details.

For more information on testing, read the [Terragrunt Catalog AWS documentation](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/tests/README.md).

## Related Documentation

- [Catalog Repository](https://github.com/ConsciousML/terragrunt-template-catalog-eks): Reusable IaC components
- [Bootstrap Setup](https://github.com/ConsciousML/terragrunt-template-catalog-eks/tree/main/bootstrap/README.md): Detailed GitHub Actions setup
- [Terragrunt Documentation](https://terragrunt.gruntwork.io/docs/): Official Terragrunt docs
- [Gruntwork Infrastructure Patterns](https://github.com/gruntwork-io/terragrunt-infrastructure-live-example): Reference architecture

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.