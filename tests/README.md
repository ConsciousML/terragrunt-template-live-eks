# Test Terragrunt Stacks With Terratest

## Prerequisites 
Perform the [quickstart](../README.md#getting-started) up to `Authenticate with AWS` (included).

## What It Tests

`TestStack` deploys the [staging EKS stack](../live/staging/eks/terragrunt.stack.hcl) end-to-end and validates:

- The stack applies via `terragrunt apply --all`
- Kubeconfig points at the deployed cluster
- ArgoCD's app-of-apps is Synced and Healthy before checking any endpoint, failing fast if it
  stalls for too long
- Each tool's Route53 domain is reachable and, where the tool has an admin password in Secrets Manager, login succeeds:
  - ArgoCD (login)
  - Grafana (login)
  - podinfo
  - Prometheus
  - Alertmanager
  - Hubble
  - Goldilocks
- Destroys the stack automatically (even if it fails)

`TestStackExists` runs the same assertions against an already-deployed stack, skipping apply and destroy. Use it to iterate on test logic without re-deploying infrastructure.

## Run Terratest

Setup the go module (the module is already initialized — run these if you are adding new dependencies):
```bash
go get github.com/gruntwork-io/terratest@v1.0.0
go get github.com/aws/aws-sdk-go-v2@v1.41.7
go get github.com/aws/aws-sdk-go-v2/config@v1.32.17
go get github.com/aws/aws-sdk-go-v2/service/secretsmanager@v1.41.7
go mod tidy
```

Follow the [environment variables guide](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/docs/environment-variables.md), then run:

```bash
source .env
```

Deploy and test the full stack (apply + assert + destroy). A run can take a while end to end (see
the wait budgets in `app_of_apps_wait_test.go` and `staging_stack_test.go`), so pipe through `tee`
to keep a log file for debugging a stall or failure:
```bash
go test -v -run TestStack ./tests/... -timeout 120m 2>&1 | tee /tmp/test.log
```

Test against an already-deployed stack (no apply or destroy):
```bash
go test -v -run TestStackExists ./tests/... -timeout 10m
```

## Write a Test

Copy `tests/staging_stack_test.go` in the `tests` directory. Use the suffix `*_test.go`.

Next, change the stack directory to the path of the stack you want to test:
```go
stackDir := "../live/staging/eks/test"
```

Finally, write additional test steps. For example, you can perform health checks or make a request to an API to ensure your infrastructure was deployed properly.
