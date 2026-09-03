locals {
  dns         = read_terragrunt_config(find_in_parent_folders("dns.hcl")).locals
  environment = read_terragrunt_config(find_in_parent_folders("environment.hcl")).locals.environment

  domain_env         = "${local.environment}.${local.dns.base_domain}"
  domain_env_private = "private.${local.domain_env}"
  domain_env_public  = "public.${local.domain_env}"

  domain_private_argocd       = "${local.dns.subdomain_argocd}.${local.domain_env_private}"
  domain_public_podinfo       = "${local.dns.subdomain_podinfo}.${local.domain_env_public}"
  domain_private_prometheus   = "${local.dns.subdomain_prometheus}.${local.domain_env_private}"
  domain_private_alertmanager = "${local.dns.subdomain_alertmanager}.${local.domain_env_private}"
  domain_private_grafana      = "${local.dns.subdomain_grafana}.${local.domain_env_private}"
  domain_private_goldilocks   = "${local.dns.subdomain_goldilocks}.${local.domain_env_private}"
  domain_private_hubble       = "${local.dns.subdomain_hubble}.${local.domain_env_private}"
}
