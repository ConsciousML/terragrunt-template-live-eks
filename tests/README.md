# Test Terragrunt Stacks With Terratest

## Prerequisites 
Perform the [quickstart](../../../README.md#getting-started) up to `Authenticate with AWS` (included).

## What It Tests

`TestStack` deploys the [staging EKS stack](../live/staging/eks/terragrunt.stack.hcl) end-to-end and validates:

- The stack applies via `terragrunt apply --all`
- ArgoCD is reachable at the private Route53 domain (`/healthz` returns 200)
- ArgoCD login succeeds using the admin password stored in Secrets Manager (`/api/v1/session` returns a valid JWT token)
- The guestbook application is reachable at its public Route53 domain
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

Deploy and test the full stack (apply + assert + destroy):
```bash
go test -v -run TestStack ./tests/... -timeout 60m
```

Assert only against an already-deployed stack (no apply or destroy):
```bash
go test -v -run TestStackExists ./tests/... -timeout 10m
```

## Write a Test

Copy `tests/staging_stack_test.go` in the `tests` directory. Use the suffix `*_test.go`.

Next, change the stack directory to the path of the stack you want to test:
```go
stackDir := "../live/staging/eks"
```

Finally, write additional test steps. For example, you can perform health checks or make a request to an API to ensure your infrastructure was deployed properly.
