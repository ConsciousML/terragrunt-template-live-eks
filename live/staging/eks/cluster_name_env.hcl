locals {
  environment_hcl = find_in_parent_folders("environment.hcl")
  environment     = read_terragrunt_config(local.environment_hcl).locals.environment

  cluster_name_hcl  = find_in_parent_folders("cluster_name.hcl")
  cluster_name      = read_terragrunt_config(local.cluster_name_hcl).locals.cluster_name
  cluster_name_full = "${local.environment}-${local.cluster_name}"
}
