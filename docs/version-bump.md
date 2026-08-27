# Version Bump Workflow

How to bump the catalog version used by the live stacks and align the live stack files against the catalog pipeline at the new tag.

## Bump `version_catalog`

Bump `version_catalog` in `locals` at the top of both:
- `live/staging/eks/stack/terragrunt.stack.hcl`
- `live/prod/eks/stack/terragrunt.stack.hcl`

## Align the Stack Files

Align each live stack file with `pipelines/dev/eks/stack/terragrunt.stack.hcl` at the new tag. Match its units, its `values`, and its `locals` (chart versions and other pinned versions).

The required drift is whatever the dev pipeline's own inline comments mark as dev-only, for example `min_size = 1`, `access_entries = {}`, or disabled log types. Keep the live-side equivalent for those instead of adopting the dev value. Re-read these comments every bump, since which settings are dev-only can change between tags.

## Check Shared HCL Files

Check shared HCL files for structural changes. These rarely change. See the "Catalog Equivalents" section of [`configuration-files.md`](configuration-files.md) for the full catalog-to-live file mapping to diff.

## Diff the Bootstrap Pipelines

Diff `pipelines/bootstrap/` against `live/bootstrap/`. Bootstrap isn't applied by CI/CD, so a missed change won't break a build, but it drifts silently.

Look for a new top-level stack in the catalog with no live counterpart, an existing bootstrap stack restructured (for example `tailscale` splitting into `acl` and `wif`), and renamed units inside an existing stack. Decide per stack whether live should adopt it.

## Diff Environment Variables

Diff the catalog's `.env.example` against live's `.env.example`. A new env var here usually pairs with a new `get_env(...)` call in a stack or bootstrap file from the previous two steps. Cross-check any new entry against the CI/CD secrets documented in [`ci-cd.md`](ci-cd.md) and update both if needed.

## Align the Docs

Align live's own docs with the catalog's equivalents at the new tag. Carry over content changes, but restate them through the required prod, staging, and dev drifts from the stack file alignment step rather than copying catalog prose verbatim, since live's docs describe live's stacks, not the dev pipeline.
