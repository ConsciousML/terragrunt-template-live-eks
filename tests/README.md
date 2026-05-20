# Test Terragrunt Stacks With Terratest

## Prerequisites 
Perform the [quickstart](../../../README.md#getting-started) up to `Authenticate with AWS` (included).

## What It Tests
`TestStack` deploys the [staging EKS stack](../live/staging/eks/terragrunt.stack.hcl) end-to-end and validates:

- The stack applies via `terragrunt apply --all`
- ArgoCD is reachable at the private Route53 domain (`/healthz` returns 200)
- ArgoCD login succeeds using the admin password stored in Secrets Manager (`/api/v1/session` returns a valid JWT token)
- Destroys the stack automatically (even if it fails)

## Run Terratest

Setup the go module (the module is already initialized — run these if you are adding new dependencies):
```bash
go get github.com/gruntwork-io/terratest@v1.0.0
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/secretsmanager
go mod tidy
```

Follow the [environment variables guide](https://github.com/ConsciousML/terragrunt-template-catalog-eks/blob/main/docs/environment-variables.md), then run:

```bash
source .env
```

Run the test:
```bash
go test -v ./tests/... -timeout 60m
```

## Write a Test

Copy `tests/staging_stack_test.go` in the `tests` directory. Use the suffix `*_test.go`.

Next, change the stack directory to the path of the stack you want to test:
```go
stackDir := "../live/staging/eks"
```

Finally, write additional test steps. For example, you can perform health checks or make a request to an API to ensure your infrastructure was deployed properly.
