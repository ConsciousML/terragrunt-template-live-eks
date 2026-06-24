locals {
  dns         = read_terragrunt_config(find_in_parent_folders("dns.hcl")).locals
  environment = read_terragrunt_config(find_in_parent_folders("environment.hcl")).locals.environment

  domain_env       = "${local.environment}.${local.dns.base_domain}"
  domain_argocd    = "${local.dns.subdomain_argocd}.${local.domain_env}"
  domain_guestbook = "${local.dns.subdomain_guestbook}.${local.domain_env}"
}
