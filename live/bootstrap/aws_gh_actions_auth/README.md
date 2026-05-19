# AWS GitHub Actions Auth

Sets up GitHub OIDC authentication with AWS and creates an IAM role for GitHub Actions. Also provisions deploy keys and GitHub secrets (`AWS_REGION`, `AWS_ROLE_ARN`, `DEPLOY_KEY_TG_CATALOG`, `DEPLOY_KEY_TG_LIVE`) so CI can authenticate to AWS and clone private repos. 

See the [catalog README](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/pipelines/bootstrap/aws_gh_actions_auth/README.md) for architecture details.

## Prerequisites

- Follow the [installation instructions](../../../README.md#installation)
- Same [prerequisites](../../../README.md#prerequisites) as in the main `README.md`

Set up `GITHUB_TOKEN` following the [environment variables guide](../../../docs/environment-variables.md).

## Deploy

Run once before CI/CD is operational:

```bash
source .env
cd live/bootstrap/aws_gh_actions_auth
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive
```
