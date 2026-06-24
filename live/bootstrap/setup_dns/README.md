# DNS Bootstrap

Creates a public Route 53 hosted zone per environment.

See the [catalog README](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/pipelines/bootstrap/setup_dns/README.md) for the full DNS flow and zone purpose.

## Structure

```
live/bootstrap/setup_dns/
  staging/
    environment.hcl        ← environment = "staging"
    stack/
      terragrunt.stack.hcl
  prod/
    environment.hcl        ← environment = "prod"
    stack/
      terragrunt.stack.hcl
```

## Prerequisites

Perform the [quickstart](../../../README.md#getting-started) up to `Authenticate with AWS` (included).

## Deploy

Repeat the following for each environment (replacing `<environment>` by `staging` and then by `prod`):

```bash
source .env
cd live/bootstrap/setup_dns/<environment>/stack
terragrunt stack run init
terragrunt run --all apply --backend-bootstrap --non-interactive
```


Retrieve the 4 nameservers from the output:

```bash
terragrunt stack output --json setup_dns.route53_hosted_zone.name_servers
```

### Delegate the subdomain

Repeat the following for each environment.

In your domain registrar, add 4 NS records for the environment zone using the nameservers from the output above.

| Type | Host | Value |
|------|------|-------|
| NS | `<environment>` | `ns-123.awsdns-12.com` |
| NS | `<environment>` | `ns-456.awsdns-34.net` |
| NS | `<environment>` | `ns-789.awsdns-56.org` |
| NS | `<environment>` | `ns-012.awsdns-78.co.uk` |

Replace `<environment>` with your actual environment (`staging` or `prod`).

### Verify propagation

```bash
dig NS staging.axelmendoza.com
```

Delegation is working when 4 AWS nameservers appear in the `ANSWER SECTION`. With delegation in place, deploy the EKS stack — ACM validation, private zone creation, and ExternalDNS are all handled automatically.
