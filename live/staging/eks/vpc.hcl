locals {
  environment = read_terragrunt_config(find_in_parent_folders("environment.hcl")).locals.environment

  vpc_name      = "vpc-eks"
  vpc_full_name = "${local.vpc_name}-${local.environment}"
}
