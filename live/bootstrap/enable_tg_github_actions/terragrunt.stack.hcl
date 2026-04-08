locals {
  version = "main"

  github_repo_name         = "terragrunt-template-live-aws"
  github_repo_catalog_name = "terragrunt-template-catalog-aws"
}

stack "enable_tg_github_actions" {
  source = "github.com/ConsciousML/terragrunt-template-catalog-aws//stacks/enable_tg_github_actions?ref=${local.version}"
  path   = "github_actions_bootstrap"
  values = {
    version          = local.version
    github_username  = "ConsciousML"
    github_repo_name = local.github_repo_name
    github_token     = get_env("TF_VAR_github_token")
    iam_role_name    = "github-actions-tg-live-role"
    policy_arns = [
      "arn:aws:iam::aws:policy/AmazonEC2FullAccess",
      "arn:aws:iam::aws:policy/AmazonVPCFullAccess",
      "arn:aws:iam::aws:policy/IAMFullAccess",
      "arn:aws:iam::aws:policy/AmazonS3FullAccess",
      "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess"
    ]
    github_branch = "*"
    deploy_key_repositories = [
      local.github_repo_name,
      local.github_repo_catalog_name
    ]
    deploy_key_secret_names = [
      "DEPLOY_KEY_TG_LIVE",
      "DEPLOY_KEY_TG_CATALOG"
    ]
    deploy_key_title = "Terragrunt Live Deploy Key"

    # OIDC provider must be unique and is created in the catalog template
    create_oidc_provider = false
    oidc_url             = "https://token.actions.githubusercontent.com"
    oidc_client_id_list  = ["sts.amazonaws.com"]
    oidc_thumbprint_list = []
  }
}