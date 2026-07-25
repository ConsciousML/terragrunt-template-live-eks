# Configuration Files

```
live/
├── root.hcl                    S3 backend + AWS provider, inherited by all units
├── github.hcl                  GitHub repo names for module sources
├── dns.hcl                     Base domain + per-app subdomains (ArgoCD, guestbook)
├── network.hcl                 Per-environment VPC CIDR blocks
├── cluster_name.hcl            EKS cluster name stem
├── provider_k8s_base.hcl       EKS cluster data sources, skips units before the cluster exists
├── provider_helm.hcl           Helm provider, sourced from the EKS cluster output
├── provider_kubernetes.hcl     Kubernetes provider, sourced from the EKS cluster output
├── provider_kubectl.hcl        kubectl provider, sourced from the EKS cluster output
└── {env}/                      staging/, prod/
    ├── environment.hcl         Environment name, used for resource naming and state isolation
    ├── region.hcl              AWS region + availability zones
    └── eks/
        ├── cluster_name_env.hcl    Full cluster name: {environment}-{cluster_name}
        ├── domains.hcl             Per-app domains, derived from dns.hcl + environment.hcl
        └── stack/
            └── terragrunt.stack.hcl    Composes all catalog units, pins version_catalog
```

## Catalog Equivalents

Each live file maps to a catalog counterpart under `pipelines/`:

```
pipelines/
├── root.hcl                 → live/root.hcl
├── github.hcl               → live/github.hcl
├── dns.hcl                  → live/dns.hcl
├── network.hcl              → live/network.hcl
├── region.hcl                → live/{env}/region.hcl
├── version.hcl               (dev-only, no live counterpart, see below)
└── dev/
    ├── environment.hcl      → live/{env}/environment.hcl
    ├── cluster_name.hcl     → live/cluster_name.hcl
    ├── provider_helm.hcl    → live/provider_helm.hcl
    ├── provider_k8s_base.hcl → live/provider_k8s_base.hcl
    └── eks/
        ├── cluster_name_env.hcl → live/{env}/eks/cluster_name_env.hcl
        ├── domains.hcl          → live/{env}/eks/domains.hcl
        └── stack/
            └── terragrunt.stack.hcl → live/{env}/eks/stack/terragrunt.stack.hcl
```

Two exceptions:
- `pipelines/version.hcl` resolves `?ref=` from the current git branch, dev-only. Live pins `?ref=${local.version_catalog}` directly in each `stack/terragrunt.stack.hcl` instead, see the [development workflow](ci_cd.md#using-the-cicd-development-workflow) for how that tag is bumped.
- `live/provider_kubernetes.hcl` and `live/provider_kubectl.hcl` have no catalog `pipelines/dev/` counterpart.

When bumping `version_catalog`, diff each live file against its catalog counterpart for structural changes, see step 3 of the version bump workflow in `system.xml`.
