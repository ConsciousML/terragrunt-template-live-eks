# DNS Bootstrap

Creates a public Route 53 hosted zone per environment. See the [catalog README](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/pipelines/bootstrap/setup_dns/README.md) for the full DNS flow, zone purpose, and NS delegation steps.

## Structure

```
live/bootstrap/setup_dns/
  staging/
    environment.hcl        ← environment = "staging"
    terragrunt.stack.hcl
  prod/
    environment.hcl        ← environment = "prod"
    terragrunt.stack.hcl
```

`live/dns.hcl` (at the live root) sets `base_domain` and `subdomain`, producing hosted zone names like `argocd.staging.axelmendoza.com`.

## Deploy

Run once per environment before deploying the EKS stack:

```bash
cd live/bootstrap/setup_dns/<env>
terragrunt stack generate
terragrunt run --all apply --backend-bootstrap --non-interactive
```

Retrieve the 4 nameservers from the output:

```bash
terragrunt stack output --json setup_dns.route53_hosted_zone.name_servers
```

### Delegate the subdomain

In your domain registrar, add 4 NS records for the subdomain using the nameservers above.

| Type | Host | Value |
|------|------|-------|
| NS | `argocd.staging` | `ns-123.awsdns-12.com` |
| NS | `argocd.staging` | `ns-456.awsdns-34.net` |
| NS | `argocd.staging` | `ns-789.awsdns-56.org` |
| NS | `argocd.staging` | `ns-012.awsdns-78.co.uk` |

Replace `argocd.staging` with your actual `{subdomain}.{environment}` and each value with the nameservers from the output.

### Verify propagation

```bash
dig NS argocd.staging.axelmendoza.com
```

Delegation is working when 4 AWS nameservers appear in the `ANSWER SECTION`. With delegation in place, deploy the EKS stack — ACM validation, private zone creation, and ExternalDNS are all handled automatically.
