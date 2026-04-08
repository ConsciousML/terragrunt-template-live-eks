locals {
  # Sets the reference of the source code to:
  version_catalog = "v0.0.2"
  version_cluster = "21.15.1"
  version_vpc     = "6.6.0"

  github_repo_catalog_name = "terragrunt-template-catalog-eks"
}

unit "vpc" {
  source = "github.com/ConsciousML/${local.github_repo_catalog_name}//units/vpc_eks?ref=${local.version_catalog}"
  path   = "vpc_eks"

  values = {
    create_vpc = true
    version    = local.version_vpc

    name = "vpc-eks"

    cidr = "10.0.0.0/16"

    # For production, use at least 3 subnets
    private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
    public_subnets  = ["10.0.3.0/24", "10.0.4.0/24"]

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
  source = "github.com/ConsciousML/${local.github_repo_catalog_name}//units/eks_cluster?ref=${local.version_catalog}"
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