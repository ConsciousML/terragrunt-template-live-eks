locals {
  version_catalog            = "v0.0.9"
  version_vpc                = "6.6.0"
  version_cluster            = "21.15.1"
  version_aws_lbc            = "3.2.1"
  version_argocd             = "9.5.0"
  version_external_dns       = "1.20.0"
  version_eso                = "2.4.1"
  version_tailscale_operator = "1.96.5"

  github_locals            = read_terragrunt_config(find_in_parent_folders("github.hcl")).locals
  github_username_catalog  = local.github_locals.github_username_catalog
  github_repo_name_catalog = local.github_locals.github_repo_name_catalog

  environment = read_terragrunt_config(find_in_parent_folders("environment.hcl")).locals.environment
  vpc_cidrs   = read_terragrunt_config(find_in_parent_folders("network.hcl")).locals.vpc_cidrs
  vpc_cidr    = local.vpc_cidrs[local.environment]

  private_subnets = [cidrsubnet(local.vpc_cidr, 8, 1), cidrsubnet(local.vpc_cidr, 8, 2)]
  public_subnets  = [cidrsubnet(local.vpc_cidr, 8, 3), cidrsubnet(local.vpc_cidr, 8, 4)]
}

unit "route53_hosted_zone_public" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/route53/hosted_zone_public?ref=${local.version_catalog}"
  path   = "eks/route53/hosted_zone_public"

  values = {
    version = local.version_catalog
    comment = "Managed by Terraform"
    create  = false
  }
}

unit "vpc" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/vpc?ref=${local.version_catalog}"
  path   = "vpc"

  values = {
    create_vpc = true
    version    = local.version_vpc

    name = "vpc-eks"

    private_subnets = local.private_subnets
    public_subnets  = local.public_subnets

    enable_nat_gateway     = true
    single_nat_gateway     = false
    one_nat_gateway_per_az = true

    enable_dns_hostnames = true
    enable_dns_support   = true

    public_subnet_tags = {
      "kubernetes.io/role/elb" = 1
    }

    private_subnet_tags = {
      "kubernetes.io/role/internal-elb" = 1
    }
  }
}

unit "cluster" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/cluster?ref=${local.version_catalog}"
  path   = "eks/cluster"

  values = {
    version = local.version_cluster

    kubernetes_version = "1.35"

    endpoint_public_access = true

    enable_cluster_creator_admin_permissions = true

    control_plane_scaling_config = {
      tier = "standard"
    }

    addons = {
      coredns = {}
      eks-pod-identity-agent = {
        before_compute = true
      }
      kube-proxy = {}
      vpc-cni = {
        before_compute = true
      }
    }

    eks_managed_node_groups = {
      example = {
        ami_type            = "AL2023_x86_64_STANDARD"
        ami_release_version = "1.35.6-20260618"

        instance_types = ["t3.medium"]

        min_size     = 2
        max_size     = 10
        desired_size = 2
      }
    }

    compute_config = {
      enabled = false
    }

    access_entries = {}
  }
}

unit "iam_policy_aws_lbc" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/aws_load_balancer_controller/iam_policy_url?ref=${local.version_catalog}"
  path   = "eks/addons/aws_load_balancer_controller/iam_policy_url"

  values = {
    version = local.version_catalog
    url     = "https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v${local.version_aws_lbc}/docs/install/iam_policy.json"
  }
}

unit "iam_role_aws_lbc" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/aws_load_balancer_controller/iam_role?ref=${local.version_catalog}"
  path   = "eks/addons/aws_load_balancer_controller/iam_role"

  values = {
    version = local.version_catalog
    tags    = {}
  }
}

unit "aws_load_balancer_controller" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/aws_load_balancer_controller/helm?ref=${local.version_catalog}"
  path   = "eks/addons/aws_load_balancer_controller/helm"

  values = {
    version                     = local.version_catalog
    helm_chart_version          = local.version_aws_lbc
    enableServiceMutatorWebhook = false
  }
}

unit "argocd_password" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/argocd/aws_password_secret?ref=${local.version_catalog}"
  path   = "eks/addons/argocd/aws_password_secret"

  values = {
    version                 = local.version_catalog
    length                  = 16
    recovery_window_in_days = 0
    tags                    = {}
  }
}

unit "argocd" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/argocd/helm?ref=${local.version_catalog}"
  path   = "eks/addons/argocd/helm"

  values = {
    version            = local.version_catalog
    helm_chart_version = local.version_argocd
    helm_values = {
      configs = {
        params = {
          "server.insecure" = true
        }
      }
      server = {
        ingress = {
          enabled          = true
          controller       = "aws"
          ingressClassName = "alb"
          annotations = {
            "alb.ingress.kubernetes.io/scheme"           = "internal"
            "alb.ingress.kubernetes.io/target-type"      = "ip"
            "alb.ingress.kubernetes.io/backend-protocol" = "HTTP"
            "alb.ingress.kubernetes.io/listen-ports"     = "[{\"HTTP\":80}, {\"HTTPS\":443}]"
            "alb.ingress.kubernetes.io/ssl-redirect"     = "443"
            "external-dns.alpha.kubernetes.io/scope"     = "private"
          }
          aws = {
            serviceType            = "ClusterIP"
            backendProtocolVersion = "GRPC"
          }
        }
      }
    }
  }
}

unit "route53_hosted_zone_private" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/route53/hosted_zone_private?ref=${local.version_catalog}"
  path   = "eks/route53/hosted_zone_private"

  values = {
    version = local.version_catalog
    comment = "Managed by Terraform"
  }
}

unit "gateway_api_crds" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/kubectl_manifest_from_url?ref=${local.version_catalog}"
  path   = "eks/addons/gateway_api/crds"

  values = {
    version = local.version_catalog
    url     = "https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.5.0/standard-install.yaml"
  }
}

unit "aws_lbc_gateway_api_crds" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/kubectl_manifest_from_url?ref=${local.version_catalog}"
  path   = "eks/addons/aws_load_balancer_controller/gateway_api_crds"

  values = {
    version = local.version_catalog
    url     = "https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v${local.version_aws_lbc}/config/crd/gateway/gateway-crds.yaml"
  }
}

unit "gateway_api_namespace" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/gateway_api/namespace?ref=${local.version_catalog}"
  path   = "eks/addons/gateway_api/namespace"

  values = {
    version = local.version_catalog
  }
}

unit "gateway_api_gateway_class" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/gateway_api/gateway_class?ref=${local.version_catalog}"
  path   = "eks/addons/gateway_api/gateway_class"

  values = {
    version = local.version_catalog
  }
}

unit "gateway_api_target_group_configuration_public" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/gateway_api/target_group_configuration/public?ref=${local.version_catalog}"
  path   = "eks/addons/gateway_api/target_group_configuration/public"

  values = {
    version = local.version_catalog
  }
}

unit "gateway_api_load_balancer_configuration_public" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/gateway_api/load_balancer_configuration/public?ref=${local.version_catalog}"
  path   = "eks/addons/gateway_api/load_balancer_configuration/public"

  values = {
    version = local.version_catalog
  }
}

unit "gateway_api_gateway_public" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/gateway_api/gateway/public?ref=${local.version_catalog}"
  path   = "eks/addons/gateway_api/gateway/public"

  values = {
    version = local.version_catalog
  }
}

unit "iam_role_external_dns_private" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/external_dns/private/iam_role?ref=${local.version_catalog}"
  path   = "eks/addons/external_dns/private/iam_role"

  values = {
    version = local.version_catalog
    tags    = {}
  }
}

unit "external_dns_private" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/external_dns/private/helm?ref=${local.version_catalog}"
  path   = "eks/addons/external_dns/private/helm"

  values = {
    version            = local.version_catalog
    helm_chart_version = local.version_external_dns
    helm_values = {
      sources = ["service", "ingress", "gateway-httproute"]
      provider = {
        name = "aws"
      }
      registry         = "txt"
      policy           = "sync"
      logLevel         = "info"
      annotationFilter = "external-dns.alpha.kubernetes.io/scope=private"
      extraArgs = {
        "aws-zone-type" = "private"
      }
    }
  }
}

unit "iam_role_external_dns_public" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/external_dns/public/iam_role?ref=${local.version_catalog}"
  path   = "eks/addons/external_dns/public/iam_role"

  values = {
    version = local.version_catalog
    tags    = {}
  }
}

unit "external_dns_public" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/external_dns/public/helm?ref=${local.version_catalog}"
  path   = "eks/addons/external_dns/public/helm"

  values = {
    version            = local.version_catalog
    helm_chart_version = local.version_external_dns
    helm_values = {
      sources = ["service", "ingress", "gateway-httproute"]
      provider = {
        name = "aws"
      }
      registry         = "txt"
      policy           = "sync"
      logLevel         = "info"
      annotationFilter = "external-dns.alpha.kubernetes.io/scope=public"
      extraArgs = {
        "aws-zone-type" = "public"
      }
    }
  }
}

unit "acm_certificate" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/route53/acm_certificate?ref=${local.version_catalog}"
  path   = "eks/route53/acm_certificate"

  values = {
    version = local.version_catalog
  }
}

unit "iam_role_eso" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/external_secrets_operator/iam_role?ref=${local.version_catalog}"
  path   = "eks/addons/external_secrets_operator/iam_role"

  values = {
    version = local.version_catalog
    tags    = {}
  }
}

unit "external_secrets_operator" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/external_secrets_operator/helm?ref=${local.version_catalog}"
  path   = "eks/addons/external_secrets_operator/helm"

  values = {
    version            = local.version_catalog
    helm_chart_version = local.version_eso
    helm_values        = {}
  }
}

unit "argocd_aws_secret_store" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/argocd/aws_secret_store?ref=${local.version_catalog}"
  path   = "eks/addons/argocd/aws_secret_store"

  values = {
    version = local.version_catalog
  }
}

unit "argocd_aws_external_secret" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/argocd/aws_external_secret?ref=${local.version_catalog}"
  path   = "eks/addons/argocd/aws_external_secret"

  values = {
    version = local.version_catalog
  }
}

unit "tailscale_oauth_client_tailscale_operator" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/tailscale/oauth_client_tailscale_operator?ref=${local.version_catalog}"
  path   = "eks/addons/tailscale/oauth_client_tailscale_operator"

  values = {
    version = local.version_catalog
  }
}

unit "tailscale_operator" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/tailscale/operator?ref=${local.version_catalog}"
  path   = "eks/addons/tailscale/operator"

  values = {
    version            = local.version_catalog
    helm_chart_version = local.version_tailscale_operator
  }
}

unit "tailscale_connector" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/tailscale/connector?ref=${local.version_catalog}"
  path   = "eks/addons/tailscale/connector"

  values = {
    version = local.version_catalog
  }
}

unit "tailscale_split_dns" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/tailscale/split_dns?ref=${local.version_catalog}"
  path   = "eks/addons/tailscale/split_dns"

  values = {
    version = local.version_catalog
  }
}

unit "argocd_app_of_apps" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks/addons/argocd/app_of_apps?ref=${local.version_catalog}"
  path   = "eks/addons/argocd/app_of_apps"

  values = {
    version               = local.version_catalog
    name                  = "app-of-apps"
    namespace             = "argocd"
    path                  = "apps"
    target_revision       = "main"
    project               = "default"
    destination_namespace = "argocd"
    destination_server    = "https://kubernetes.default.svc"
    finalizers            = ["resources-finalizer.argocd.argoproj.io"]
    sync_options          = ["CreateNamespace=true"]
    prune                 = true
  }
}
