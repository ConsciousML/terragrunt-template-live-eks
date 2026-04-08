# GitHub Actions AWS Authentication

## Overview

Authenticates GitHub Actions with AWS to deploy infrastructure with Terragrunt in your workflows.

The bootstrap pipeline has been develop in the [catalog repository](https://github.com/ConsciousML/terragrunt-template-catalog-aws). We will simply call the stack in the current repository (live).

For a detailed explanation of what this bootstrap stack does and its architecture, see the [catalog bootstrap documentation](https://github.com/ConsciousML/terragrunt-template-catalog-aws/tree/main/bootstrap/README.md).

## Prerequisites
- Same prerequisites as the [root README.md](../README.md#prerequisites)
- Follow the [installation section](../README.md#installation)

## Configuration

Open `live/bootstrap/enable_tg_github_actions/terragrunt.stack.hcl` and update it according to the [configuration documentation](https://github.com/ConsciousML/terragrunt-template-catalog-aws/blob/main/bootstrap/README.md#configuration)

Update the `live/bootstrap/region.hcl` to match your desired AWS region.

## Running the Bootstrap

1. Authenticate GitHub CLI:
   ```bash
   gh auth login --scopes "repo,admin:repo_hook"
   export TF_VAR_github_token="$(gh auth token)"
   ```

2. Deploy the bootstrap stack:
   ```bash
   cd live/bootstrap/enable_tg_github_actions
   terragrunt stack generate
   terragrunt stack run apply --backend-bootstrap --non-interactive
   ```

4. Verify setup:
   ```bash
   gh secret list
   ```

   You should see:
   ```bash
   AWS_REGION             about a minute ago
   AWS_ROLE_ARN           about a minute ago
   DEPLOY_KEY_TG_CATALOG  about a minute ago
   DEPLOY_KEY_TG_LIVE     about a minute ago
   ```

## Next Steps

- Review the [CI/CD workflow guide](ci_cd.md) to understand how to use the automated pipelines
- Create a test PR to verify CI/CD works correctly
