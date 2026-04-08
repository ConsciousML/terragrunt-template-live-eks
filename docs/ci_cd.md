# Continuous Integration and Deployment (CI/CD)

## Overview

The [CI](../.github/workflows/ci.yaml) and [CD](../.github/workflows/cd.yaml) workflows automate infrastructure validation, testing, and deployment on every pull request and merge.

They ensure that:
- Infrastructure changes are validated before reaching production
- All environments (dev, staging, prod) are tested
- Production deployments are reviewed
- Infrastructure can be deployed and destroyed correctly

## How It Works

### Continuous Integration (CI)

Runs automatically on every pull request to `main` and is composed of four jobs:

#### 1. HCL Format Check
Validates that all Terragrunt (`.hcl`) files are properly formatted.

#### 2. Validate & Plan
Runs in parallel across **dev**, **staging**, and **prod** environments:
- Generates stack configurations
- Initializes Terragrunt with backend bootstrapping
- Validates Terraform/OpenTofu syntax
- Creates infrastructure plans showing what will change

For the **production environment**, the plan output is converted to HTML and uploaded as a downloadable artifact for review.

#### 3. Terratest
Runs infrastructure tests when the `run-terratest` label is added to the PR:
- Deploys the AWS infrastructure to the staging environment
- Runs Go-based validation tests
- Automatically destroys all test resources

If you want to skip the testing phase, add the `skip-terratest` label to the PR.

#### 4. Comment on PR
Posts a comment with:
- Link to the production plan artifact
- Commit SHA that was tested
- Warning that merging will apply changes to production

### Continuous Deployment (CD)

Runs automatically when a PR is **merged to `main`** and deploys changes to the **production environment**.

**Important**: CD automatically applies to production. Always review the production plan from CI before merging. If you want to skip the deployment, add the `skip-cd` tag before merging the PR.

## Setup

### Initial Setup
Follow the [bootstrap guide](https://github.com/ConsciousML/terragrunt-template-catalog-aws/tree/main/bootstrap/enable_tg_github_actions) once to:
- Configure GitHub Actions authentication with AWS
- Create a IAM Role with the required policies
- Set up deploy keys for private repository access
- Add GitHub secrets to be retrieved by GitHub Actions workflows.

## Using the CI/CD
1. Create a branch with your infrastructure changes:
   ```bash
   git checkout -b feature/update-instance-size
   ```

2. Make changes into the Terraform code, units, and stacks in the catalog repository by following its [development workflow](https://github.com/ConsciousML/terragrunt-template-catalog-aws/blob/main/docs/development.md).

3. Next, update a stack configuration or add a new stack. For example:
```hcl
# live/dev/vpc_gce/terragrunt.stack.hcl
locals {
  version = "v0.0.3" # Change version, make sure to release on the catalog before
}

stack "vpc_ec2" {
  source = "github.com/ConsciousML/terragrunt-template-catalog-aws//stacks/vpc_ec2?ref=${local.version}"
  path   = "infrastructure"

  values = {
    ec2_instance_type  = "t3.micro" 
    # all other parameters ...
  }
}
```

3. Push your code.

4. Open a pull request. The CI pipeline runs automatically.

5. Review the CI results:
   - Check that all validation passes. If not, implement a fix, and go back to 3.
   - Download and review the **production plan artifact** (linked in PR comment)
   - The plan shows exactly what will be created, modified, or destroyed

6. Add the `run-terratest` label:
   - Tests deploy real infrastructure to staging
   - Validates functionality end-to-end
   - Automatically cleans up resources
   - Use this before merging significant changes

7. Optionally, add the `skip-terratest` if you do not want to deploy the infrastructure on staging.

7. Review and merge.

8. CD automatically deploys to production after merge completes
