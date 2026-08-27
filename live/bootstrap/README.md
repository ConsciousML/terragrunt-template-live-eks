# Bootstrap Pipelines

Run each of these pipelines **once** before deploying any stack:

- **[AWS GitHub Actions Auth](aws_gh_actions_auth/README.md)**: authenticates GitHub Actions with AWS via OIDC
- **[Setup DNS](setup_dns/README.md)**: creates a public Route53 hosted zone per environment for ACM certificate validation
- **[Slack](slack/README.md)**: registers the Slack bot token as a GitHub Actions secret and creates each environment's Slack channels, so CI-deployed Alertmanager instances can send notifications to Slack
- **[Tailscale](tailscale/README.md)**: sets up the OAuth client for the Tailscale Kubernetes operator

`AWS Billing Alerts` is account-scoped, not repo-scoped. It's only instantiated once in the catalog repo, see [its README](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/pipelines/bootstrap/aws_billing_alerts/README.md).
