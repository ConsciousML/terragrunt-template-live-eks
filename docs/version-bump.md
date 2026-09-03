# Version Bump Workflow

How to bump the catalog version used by the live stacks and align the live stack files against the catalog pipeline at the new tag.

Do each section below in order, top to bottom. Finish one section's changes before starting the next, don't jump ahead. While you implement, update the docs.

## 1. Align `mise.toml`

Diff the catalog's `mise.toml` against live's `mise.toml`. For every tool present in both files, match the catalog's pinned version exactly, since CI and local runs must use the same tool versions the catalog pipeline was built and tested against. A tool that only exists in the catalog's file is module/dev tooling live doesn't need.

## 2. Diff Environment Variables

Diff the catalog's `.env.example` against live's `.env.example`. A new env var here usually pairs with a new `get_env(...)` call added in the bootstrap or stack-file alignment steps below, if that pipeline or unit is adopted, so revisit this diff after those steps if a new entry's purpose isn't clear yet. Cross-check any new entry against the CI/CD secrets documented in [`ci-cd.md`](ci-cd.md) and update both if needed.

Restate each new entry's comment through live's convention (a URL anchor into [`environment-variables.md`](environment-variables.md)) rather than copying the catalog's local-README-reference comment verbatim.

## 3. Bump `version_catalog` in Bootstrap Stack files

Bump the `version` in each bootstrap pipelines in `live/bootstrap/*`

## 4. Diff the Bootstrap Pipelines

Ask the user whether bootstrap should be checked this bump, it isn't applied by CI/CD, so a missed change won't break a build, but it drifts silently, and not every bump needs it.

If yes, diff `pipelines/bootstrap/` against `live/bootstrap/`. Look for a new top-level stack in the catalog with no live counterpart, an existing bootstrap stack restructured, and renamed units inside an existing stack. Decide per stack whether live should adopt it.

Never miss a bootstrap stack that's new in the catalog since the last bump. For each one, check the catalog's own README for that stack before deciding: some bootstrap pipelines are explicitly dev-only or CI-only and state so, and shouldn't be adopted into live.

After adopting any changes, run `terragrunt plan` against each affected live bootstrap stack to confirm whether an apply is actually needed. Don't assume the structural diff alone tells you the live state has drifted.

## 5. Check Shared HCL Files

Check shared HCL files for structural changes, including renames of shared locals, not just additions, since a rename ripples into every stack and bootstrap file reading that file's locals. See the "Catalog Equivalents" section of [`configuration-files.md`](configuration-files.md) for the full catalog-to-live file mapping to diff.


## 6. Bump `version_catalog` in Stack Files

Bump `version_catalog` in `locals` at the top of both:
- `live/staging/eks/stack/terragrunt.stack.hcl`
- `live/prod/eks/stack/terragrunt.stack.hcl`

## 7. Align the Stack Files

Align each live stack file with `pipelines/dev/eks/stack/terragrunt.stack.hcl` at the new tag. Match its units, its `values`, and its `locals` (chart versions and other pinned versions).

The required drift is whatever the dev pipeline's own inline comments mark as dev-only. Keep the live-side equivalent for those instead of adopting the dev value. Re-read these comments every bump, since which settings are dev-only can change between tags.

## 8. Verify the Docs

Align live's own docs with the catalog's equivalents at the new tag. Carry over content changes, but restate them through the required prod, staging, and dev drifts from the stack file alignment step rather than copying catalog prose verbatim, since live's docs describe live's stacks, not the dev pipeline.
