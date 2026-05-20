locals {
  version = "v0.0.5"

  github_locals            = read_terragrunt_config(find_in_parent_folders("github.hcl")).locals
  github_username_catalog  = local.github_locals.github_username_catalog
  github_repo_name_catalog = local.github_locals.github_repo_name_catalog
}

stack "setup_dns" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//stacks/setup_dns?ref=${local.version}"
  path   = "setup_dns"
  values = {
    version = local.version
  }
}
