locals {
  version = "v0.1.1"
}

stack "vpc_ec2" {
  source = "github.com/ConsciousML/terragrunt-template-catalog-aws//stacks/vpc_ec2?ref=${local.version}"
  path   = "infrastructure"
  values = {
    version            = local.version
    cidr_block_vpc     = "10.0.0.0/16"
    cidr_block_subnet  = "10.0.1.0/24"
    enable_dns_support = "false"
    zone               = "eu-west-3a"
    ec2_ami            = "ami-0ef9bcd5dfb57b968"
    ec2_instance_type  = "t3.micro"
  }
}