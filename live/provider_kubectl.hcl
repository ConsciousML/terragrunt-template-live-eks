generate "provider_kubectl" {
  path      = "provider_kubectl.tf"
  if_exists = "overwrite"
  contents  = <<EOF
provider "kubectl" {
  host                   = data.aws_eks_cluster.this.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.this.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.this.token
  load_config_file       = false
}
EOF
}
