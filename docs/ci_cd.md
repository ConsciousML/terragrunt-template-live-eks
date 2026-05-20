# Continuous Integration and Deployment (CI/CD)

## Overview

The [CI](../.github/workflows/ci.yaml) and [CD](../.github/workflows/cd.yaml) workflows automate infrastructure validation, testing, and deployment on every pull request and merge.

## How It Works

### Continuous Integration (CI)

Runs automatically on every pull request to `main` and is composed of seven jobs:

#### 1. HCL Format Check
Validates that all Terragrunt (`.hcl`) files are properly formatted.

#### 2. Validate & Plan
Runs in parallel across **staging** and **prod** environments:
- Generates stack configurations
- Initializes Terragrunt with backend bootstrapping
- Validates Terraform/OpenTofu syntax
- Creates infrastructure plans showing what will change

For the **production environment**, the plan output is converted to HTML and uploaded as a downloadable artifact for review.

#### 3. Terratest Label Gate
After validate & plan, CI checks that the PR has exactly one of two labels before proceeding. **This is a hard gate — CI fails if neither label is present:**
- `run-terratest`: infrastructure tests will run
- `skip-terratest`: infrastructure tests will be skipped

#### 4. Terratest
See the [Terratest guide](../tests/README.md) for details on what is tested and how to extend it.

Runs only when the `run-terratest` label is present:
- Deploys the AWS infrastructure to the staging environment
- Runs Go-based validation tests
- Automatically destroys all test resources
- Tailscale credentials are provisioned for the duration of the test and revoked immediately after

#### 5. Comment on PR
Posts a comment with:
- Link to the production plan artifact
- Commit SHA that was tested
- Warning that merging will apply changes to production

### Continuous Deployment (CD)

Runs automatically when a PR is **merged to `main`** and deploys changes to the **production environment**.

**Important**: CD automatically applies to production. Always review the production plan from CI before merging. If you want to skip the deployment, add the `skip-cd` tag before merging the PR.

## Prerequisites
Perform the [quickstart](../../../README.md#getting-started) up to `Authenticate with AWS` (included).

Then, run the [bootstrap pipelines](../README.md#run-the-bootstrap-pipelines) once per repository.

## Using the CI/CD (Development Workflow)
1. Create a branch with your infrastructure changes:
   ```bash
   git checkout -b feature/update-instance-size
   ```

2. Make changes into the Terraform code, units, and stacks in the catalog repository by following its [development workflow](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/docs/development.md).

3. Next, update a stack configuration or add a new stack. For example:
```hcl
# live/dev/vpc_gce/terragrunt.stack.hcl
locals {
  version = "v0.0.3" # Change version, make sure to release on the catalog before
}

stack "vpc_ec2" {
  source = "github.com/ConsciousML/terragrunt-template-catalog-eks//stacks/vpc_ec2?ref=${local.version}"
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

6. Add a label — **required before CI can complete**:
   - `run-terratest`: deploys real infrastructure to staging, validates end-to-end, and cleans up automatically — use this before merging significant changes
   - `skip-terratest`: skips the staging deployment

7. Review and merge.

8. CD automatically deploys to production after merge completes.
