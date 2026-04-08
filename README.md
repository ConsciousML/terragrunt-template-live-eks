# Terragrunt Template Live AWS

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub Release](https://img.shields.io/github/release/ConsciousML/terragrunt-template-live-aws.svg?style=flat)]()
[![CI](https://github.com/ConsciousML/terragrunt-template-live-aws/actions/workflows/ci.yaml/badge.svg)](https://github.com/ConsciousML/terragrunt-template-live-aws/actions/workflows/ci.yaml)
[![PR's Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat)](http://makeapullrequest.com)

A prod-ready Terragrunt Template for deploying multi-environment IaC on AWS

## Catalog vs Live Infrastructure

This is a **live repository** for deploying infrastructure across multiple environments.

This IaC production toolkit follows [Gruntwork's official patterns](https://github.com/gruntwork-io/terragrunt-infrastructure-live-example) by using two template repositories:
- **Catalog repository**: Defines **what** can be deployed (reusable components: [modules, units, and stacks](https://github.com/ConsciousML/terragrunt-template-catalog-eks))
- **This repository** (live): Defines **where** and **how** catalog components are deployed in `dev`, `staging`, and `prod` environments with CI/CD


## What's Inside

- Multi-environment IaC support
- [CI](.github/workflows/ci.yaml) (on PR): Runs `terragrunt plan` on each environment, uploads output plan to PR, deploys on the staging environment, runs some tests, and destroys.
- [CD](.github/workflows/cd.yaml) (on push `main`): Automatically deploys on `prod`
- [Bootstrap pipeline](live/bootstrap/enable_tg_github_actions/) to automatically authenticate GitHub Actions with AWS.

You're new to Terragrunt best practices? Read [Gruntwork's official production patterns](https://github.com/gruntwork-io/terragrunt-infrastructure-catalog-example) to get the foundations required to use this extended repository.

## Getting Started
**Protip**: Follow the getting started of the [EKS catalog repository](https://github.com/ConsciousML/terragrunt-template-catalog-eks) first, as you'll need to configure it before using this live repository. 

### Prerequisites
- AWS account with billing enabled
- GitHub account
- AWS IAM permissions to manage IAM roles, VPC resources, compute resources and S3 (see `policy_arns` in the [bootstrap stack](live/bootstrap/enable_tg_github_actions/terragrunt.stack.hcl) for a list of the specific IAM policies)

### Fork the Repositories (catalog and live)
First, fork the catalog repository by following its [Fork the Repository section](https://github.com/ConsciousML/terragrunt-template-catalog-eks/#fork-the-repository).

Next, you'll need to also fork this repository (live) and make a few changes:
1. Click on `Use this template` to create your own repository
2. Use your IDE of choice to replace every occurrence of `github.com/ConsciousML/terragrunt-template-catalog-eks` by your GitHub repo URL following the same format

**Warning**: If you skip step 2, the TG source links will still point to the original repository (on `github.com/ConsciousML/`).

### Installation

**Option 1: Use mise (recommended)**

First, `cd` at the root of this repository. 

Next, install mise:
```bash
curl https://mise.run | sh
```

Then, install all the tools in the `mise.toml` file:
```bash
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

Next, for each directory in `live/` (i.e `bootstrap/`, `dev/`, etc.), change `region.hcl` to match your desired AWS region.

### Deploy Dev Infrastructure Manually
Deploy the `dev` that creates a VPC and deploys an EC2 instance in a subnet:
```bash
cd live/dev
terragrunt stack generate
terragrunt stack run apply --backend-bootstrap --non-interactive
```

- Go into the AWS console and check that your resources have been created
- Then cleanup by destroying the infrastructure (cwd in `live/dev/`):

```bash
terragrunt stack generate
terragrunt stack run destroy --non-interactive
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
