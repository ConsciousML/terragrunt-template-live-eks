locals {
  version = "v0.0.5"

  github_locals            = read_terragrunt_config(find_in_parent_folders("github.hcl")).locals
  github_username_live     = local.github_locals.github_username_live
  github_repo_name_live    = local.github_locals.github_repo_name_live
  github_username_catalog  = local.github_locals.github_username_catalog
  github_repo_name_catalog = local.github_locals.github_repo_name_catalog

  ci_tag = "tag:ci"
}

stack "tailscale_wif" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//stacks/tailscale_wif?ref=${local.version}"
  path   = "tailscale_wif"

  values = {
    version          = local.version
    github_username  = local.github_username_live
    github_repo_name = local.github_repo_name_live
    github_token     = get_env("GITHUB_TOKEN")
    issuer           = "https://token.actions.githubusercontent.com"
    scopes           = ["all"]
    ci_tag           = local.ci_tag
  }
}
