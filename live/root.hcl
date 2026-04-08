locals {
  # Load AWS region variable
  region_vars = read_terragrunt_config(find_in_parent_folders("region.hcl"))

  # Load environment related variables (dev, staging, prod, ...)
  environment_vars = read_terragrunt_config(find_in_parent_folders("environment.hcl"))

  aws_region  = local.region_vars.locals.region
  environment = local.environment_vars.locals.environment
}

# Configure AWS backend for storing Terraform state files
remote_state {
  backend = "s3"
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
  config = {
    region = "${local.aws_region}"

    # The bucket name is suffixed using the env name (i.e `dev`, `staging`, ect.)
    # This allows to completely isolate states between environments
    bucket = "tofu-state-${get_aws_account_id()}-${local.environment}"

    # The state file path within the bucket, based on module's relative path to ensure each module has its own isolated state
    key            = "${path_relative_to_include()}/tofu.tfstate"
    dynamodb_table = "terragrunt_lock_table"
  }
}

generate "provider" {
  path      = "providers.tf"
  if_exists = "overwrite"
  contents  = <<EOF
provider "aws" {
  region = "${local.aws_region}"
}
EOF
}

generate "versions" {
  path      = "versions.tf"
  if_exists = "overwrite"
  contents  = <<EOF
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }

  required_version = ">= 1.9.1"
}
EOF
}

catalog {
  urls = [
    "https://github.com/ConsciousML/terragrunt-template-catalog-aws"
  ]
}

# Pass key variables to child configurations
inputs = merge(
  local.region_vars.locals,
  local.environment_vars.locals
)