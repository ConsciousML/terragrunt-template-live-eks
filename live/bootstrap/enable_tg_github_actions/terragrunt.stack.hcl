locals {
  version = "v0.0.2"

  github_repo_name         = "terragrunt-template-live-eks"
  github_repo_catalog_name = "terragrunt-template-catalog-eks"

  github_token = get_env("TF_VAR_github_token")
}

stack "enable_tg_github_actions" {
  source = "github.com/ConsciousML/${local.github_repo_catalog_name}//stacks/enable_tg_github_actions?ref=${local.version}"
  path   = "github_actions_bootstrap"
  values = {
    version          = local.version
    github_username  = "ConsciousML"
    github_repo_name = local.github_repo_name
    github_token     = local.github_token
    iam_role_name    = "gh-tg-live-eks-role"
    policy_arns = [
      "arn:aws:iam::aws:policy/AmazonEC2FullAccess",
      "arn:aws:iam::aws:policy/AmazonVPCFullAccess",
      "arn:aws:iam::aws:policy/IAMFullAccess",
      "arn:aws:iam::aws:policy/AmazonS3FullAccess",
      "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess",
      "arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess",
      "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess",
      "arn:aws:iam::aws:policy/AWSKeyManagementServicePowerUser"
    ]
    inline_policies = [
      {
        name = "EKSFullAccessLive"
        policy = jsonencode({
          Version = "2012-10-17"
          Statement = [
            {
              Sid    = "EKSFullAccessLive"
              Effect = "Allow"
              Action = [
                "eks:CreateCluster",
                "eks:DeleteCluster",
                "eks:DescribeCluster",
                "eks:ListClusters",
                "eks:UpdateClusterConfig",
                "eks:UpdateClusterVersion",
                "eks:DescribeUpdate",
                "eks:TagResource",
                "eks:UntagResource",
                "eks:ListTagsForResource",
                "eks:CreateFargateProfile",
                "eks:DeleteFargateProfile",
                "eks:DescribeFargateProfile",
                "eks:ListFargateProfiles",
                "eks:CreateNodegroup",
                "eks:DeleteNodegroup",
                "eks:DescribeNodegroup",
                "eks:ListNodegroups",
                "eks:UpdateNodegroupConfig",
                "eks:UpdateNodegroupVersion",
                "eks:CreateAddon",
                "eks:DeleteAddon",
                "eks:DescribeAddon",
                "eks:DescribeAddonVersions",
                "eks:ListAddons",
                "eks:UpdateAddon",
                "eks:CreateAccessEntry",
                "eks:DeleteAccessEntry",
                "eks:DescribeAccessEntry",
                "eks:ListAccessEntries",
                "eks:AssociateAccessPolicy",
                "eks:DisassociateAccessPolicy",
                "eks:ListAssociatedAccessPolicies",
                "eks:AssociateIdentityProviderConfig",
                "eks:DisassociateIdentityProviderConfig",
                "eks:DescribeIdentityProviderConfig",
                "eks:ListIdentityProviderConfigs",
              ]
              Resource = "*"
            }
          ]
        })
      }
    ]
    github_branch        = "*"
    oidc_url             = "https://token.actions.githubusercontent.com"
    oidc_client_id_list  = ["sts.amazonaws.com"]
    oidc_thumbprint_list = []

    # OIDC provider must be unique and is created in the catalog template
    create_oidc_provider = false

    deploy_key_repositories = [
      local.github_repo_name,
      local.github_repo_catalog_name
    ]
    deploy_key_secret_names = [
      "DEPLOY_KEY_TG_LIVE",
      "DEPLOY_KEY_TG_CATALOG"
    ]
    deploy_key_title = "Terragrunt Live EKS Deploy Key"
  }
}