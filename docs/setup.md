# Live Repository Setup

In this guide, you'll fork and set up the [live repository](https://github.com/ConsciousML/terragrunt-template-live-eks), deploy the EKS stack in the [`staging` environment](/docs/iac/#environments), and promote the stack to `prod`.

## Prerequisites
Live has the same [prerequisites](/docs/quickstart/prerequisites/) as the catalog. If you've performed the [quickstart](/docs/quickstart/), you should be good to go.

## Fork the Live Repository
The live repository deploys the [catalog stacks](/docs/architecture/) in the `staging` and `prod` environments.
Like the catalog, it's meant to be forked and extended.

### Private Fork
For creating a private repository, go to the [live home page](https://github.com/ConsciousML/terragrunt-template-live-eks) and:
1. click on `Use this template` in the top-right corner.
2. select `Create a new repository`
3. choose a repository name
4. under `Configuration`, click to the drop down next to `Choose visibility` and click on `Private`
5. click on `Create repository`

### Public Fork
For creating a public repository, go to the [live home page](https://github.com/ConsciousML/terragrunt-template-live-eks) and:
1. click on the `Fork` button
2. change the repository name if needed
3. click on `Create fork`

## Install the CLI Tools
Clone your fork and `cd` at the root of the repository.

Live uses [mise-en-place](https://mise.jdx.dev/) like the catalog, with a slightly different tool set pinned in [`mise.toml`](../mise.toml) and [`mise.local.toml`](../mise.local.toml). Follow the [CLI tools installation steps](/docs/quickstart/installation/#install-the-cli-tools) again, this time from your live fork.

## Live Configuration
Just like the catalog, live uses `.hcl` files to configure your pipelines, but under the [`live/`](../live/) directory. It points to the catalog's [units](/docs/iac/#units) 