locals {
  # Sets the reference of the source code to:
  version_catalog = "v0.0.2"
  version_cluster = "21.15.1"
  version_vpc     = "6.6.0"

  github_locals            = read_terragrunt_config(find_in_parent_folders("github.hcl")).locals
  github_username_catalog  = local.github_locals.github_username_catalog
  github_repo_name_catalog = local.github_locals.github_repo_name_catalog

  vpc_cidr        = read_terragrunt_config(find_in_parent_folders("network.hcl")).locals.vpc_cidr
  private_subnets = [cidrsubnet(local.vpc_cidr, 8, 1), cidrsubnet(local.vpc_cidr, 8, 2)]
  public_subnets  = [cidrsubnet(local.vpc_cidr, 8, 3), cidrsubnet(local.vpc_cidr, 8, 4)]
}

unit "vpc" {
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/vpc_eks?ref=${local.version_catalog}"
  path   = "vpc_eks"

  values = {
    create_vpc = true
    version    = local.version_vpc

    name = "vpc-eks"

    cidr = local.vpc_cidr

    # For production, use at least 3 subnets
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
  source = "github.com/${local.github_username_catalog}/${local.github_repo_name_catalog}//units/eks_cluster?ref=${local.version_catalog}"
  path   = "eks_cluster"

  values = {
    version = local.version_cluster

    name = "eks-cluster"

    kubernetes_version = "1.35"

    endpoint_public_access = true

    # Adds the current caller identity as an administrator via cluster access entry
    enable_cluster_creator_admin_permissions = true

    control_plane_scaling_config = {
      tier = "standard"
    }

    # More info:
    # https://docs.aws.amazon.com/eks/latest/userguide/workloads-add-ons-available-eks.html
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
        # Starting on 1.30, AL2023 is the default AMI type for EKS managed node groups
        ami_type = "AL2023_x86_64_STANDARD"

        # Use cheapest config for testing purposes
        instance_types = ["t3.medium"]

        min_size     = 2
        max_size     = 10
        desired_size = 2
      }
    }

    # Disable EKS Auto mode
    compute_config = {
      enabled = false
    }
  }
}