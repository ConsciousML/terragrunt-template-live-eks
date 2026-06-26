# Terragrunt Template Live AWS

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub Release](https://img.shields.io/github/release/ConsciousML/terragrunt-template-live-eks.svg?style=flat)]()
[![CI](https://github.com/ConsciousML/terragrunt-template-live-eks/actions/workflows/ci.yaml/badge.svg)](https://github.com/ConsciousML/terragrunt-template-live-eks/actions/workflows/ci.yaml)
[![PR's Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com)

A prod-ready live Terragrunt repository for deploying [EKS](https://aws.amazon.com/eks/) clusters across `staging` and `prod` with automated CI/CD.

The [EKS Cluster Stack](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/units/eks/README.md) supports:
- GitOps via ArgoCD and the App of Apps pattern
- Public traffic routing via ALB and Gateway API
- Automated DNS and TLS termination
- VPN access via Tailscale
- Node autoscaling via Karpenter

## Catalog vs Live Infrastructure

This toolkit uses two template repositories:
- **Catalog repository**: Defines **what** can be deployed (reusable components: [modules, units, and stacks](https://github.com/ConsciousML/terragrunt-template-catalog-eks))
- **This repository** (live): Defines **where** and **how** catalog components are deployed in `staging` and `prod` environments with CI/CD

You're new to Terragrunt best practices? Read [Gruntwork's official production patterns](https://github.com/gruntwork-io/terragrunt-infrastructure-live-example) to get the foundations required to use this toolkit.

## What's Inside
- Multi-environment IaC support: build EKS cluster across `staging`, and `prod`.
- [CI](.github/workflows/ci.yaml) (on PR): Runs `terragrunt plan` on each environment, uploads output plan to PR, deploys on the staging environment, runs tests, and destroys.
- [CD](.github/workflows/cd.yaml) (on push `main`): Automatically deploys on `prod`
- [Bootstrap pipelines](live/bootstrap/): one-time setup required before deploying the EKS stack.

## Getting Started

Follow the getting started of the [EKS catalog repository](https://github.com/ConsciousML/terragrunt-template-catalog-eks) first, as you'll need to configure it before using this live repository. 

### Prerequisites
- AWS account with billing enabled
- GitHub account
- `AdministratorAccess` AWS IAM Policy

### Fork the Repository
1. Click on the `Use this template` > `Create a new repository` button.
2. Under `Repository Name`, choose a name for your repository.

### Configuration
1. In `live/github.hcl`, modify (by replacing `<YourGitHubUsername*>`, `<your-catalog-repo-name>`, and `<your-live-repo-name>`):
```hcl
locals {
  github_username_catalog  = "<YourUsernameWhereYourCatalogForkIs>"
  github_username_live     = "<YourUsernameWhereYourLiveForkIs>"
  github_repo_name_catalog = "<your-catalog-repo-name>"
  github_repo_name_live    = "<your-live-repo-name>"
}
```
If you've forked both repositories, `github_username_catalog` and `github_username_live` should point to your username (`ConsciousML` for my own forks).

`<your-live-repo-name>` should be the same name you chose in the previous section and `<your-catalog-repo-name>` should be the name you chose in the `### Fork the Repository` section of the [catalog repository](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/README.md#fork-the-repository). 

2. Change each `live/*/region.hcl` to match your desired AWS region.

3. Set `TAILSCALE_OAUTH_CLIENT_ID` and `TAILSCALE_OAUTH_CLIENT_SECRET` in your `.env` (see the [environment variables guide](docs/environment-variables.md))

4. Karpenter's NodePool is capped at 10 vCPUs by default and provisions `spot` instances. Raise `spec.limits.cpu` or switch `karpenter.sh/capacity-type` to `on-demand` in the [staging](live/staging/eks/terragrunt.stack.hcl) and [prod](live/prod/eks/terragrunt.stack.hcl) EKS stacks for production stability.

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
- [Setup DNS](live/bootstrap/setup_dns/README.md): creates one public [Route53 hosted zone](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/hosted-zones-working-with.html) per environment, shared by all apps, delegated once at your registrar
- [Tailscale](live/bootstrap/tailscale/README.md): creates [Tailscale](https://tailscale.com/) resources needed for CI and to access internal cluster tools (ArgoCD, etc.)

Also run the following once per AWS account:
```bash
aws iam create-service-linked-role --aws-service-name spot.amazonaws.com || true
```
This creates the EC2 Spot service-linked role required for Karpenter to provision spot instances.

### Deploy a Staging EKS Cluster
Deploy the [EKS Cluster Stack](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/units/eks/README.md):

```bash
source .env
cd live/staging/eks
terragrunt stack run init
terragrunt run --all apply --backend-bootstrap --non-interactive --no-stack-generate
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

The ArgoCD host is `argocd.private.staging.<base_domain>` (replace `<base_domain>` with the value from `live/dns.hcl`, e.g. `argocd.private.staging.axelmendoza.com`).

**Web UI**: Open `https://argocd.private.staging.<base_domain>` in your browser and log in with username `admin`. Retrieve the password with:
```bash
aws secretsmanager get-secret-value \
  --secret-id staging-argocd-password \
  --query SecretString \
  --output text | jq -r .plaintext
```

**CLI**: Log in directly in one command:
```bash
argocd login argocd.private.staging.<base_domain> \
  --username admin \
  --password $(aws secretsmanager get-secret-value \
    --secret-id staging-argocd-password \
    --query SecretString \
    --output text | jq -r .plaintext)
```

### Access the Guestbook App

Open `https://guestbook.public.staging.<base_domain>` in your browser. No login required.

Apps are deployed using the [App of Apps](https://github.com/ConsciousML/argocd-app-of-apps-template) pattern: a single ArgoCD Application bootstraps all child apps from that repository.

### Destroy the Infrastructure

Cleanup by destroying the infrastructure (cwd in `live/staging/eks`):

```bash
terragrunt run --all destroy --non-interactive --no-stack-generate
```

## Extend this Repository
Follow the [development workflow guide](docs/ci_cd.md#using-the-cicd-development-workflow).

## CI/CD Pipelines

### CI (Pull Requests)
- Validates HCL formatting
- Runs `terragrunt plan` on `staging` and `prod` in parallel
- Runs infrastructure tests with Terratest
- Comments on PR with production plan artifact

### CD (Merge to main)
- Automatically deploys to the production environment (i.e `prod`).

See the [CI/CD workflow guide](docs/ci_cd.md) for detailed setup instructions and usage.

## Testing

See the [Terratest guide](tests/README.md) for running and writing infrastructure tests.

### Pre-commit Setup (recommended)
We use a more efficient framework than [pre-commit](https://github.com/pre-commit/pre-commit) called [prek](https://github.com/j178/prek).

Wire hooks automatically into git automatically:
```bash
prek install
```

Run hooks on demand:
```bash
prek run
```

## Create a New Environment

See the [new environment guide](docs/new-environment.md) for the full sequence of steps.

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.