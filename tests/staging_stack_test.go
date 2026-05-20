package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terragrunt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStack(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	region := os.Getenv("AWS_REGION")
	require.NotEmpty(t, region, "AWS_REGION must be set")

	// Tailscale CLI is needed to flush DNS cache after Terragrunt apply
	_, err := exec.LookPath("tailscale")
	require.NoError(t, err, "tailscale CLI not found in PATH — install it before running this test")

	stackDir := "../live/staging/eks"

	options := &terragrunt.Options{
		TerragruntDir:  stackDir,
		TerragruntArgs: []string{"--log-level", "error"},
	}

	defer terragrunt.DestroyAllContext(t, ctx, options)

	terragrunt.ApplyAllContext(t, ctx, options)

	// Use the Tailscale CLI to flush DNS cache
	reconnectTailscale(t)

	// Remove logger for stack output to avoid printing sensitive information
	silentOptions := &terragrunt.Options{
		TerragruntDir:  stackDir,
		TerragruntArgs: []string{"--log-level", "error"},
		Logger:         logger.Discard,
	}
	allOutputs := terragrunt.StackOutputAllContext(t, ctx, silentOptions)

	testArgoCDLogin(t, ctx, region, allOutputs)
}

func reconnectTailscale(t *testing.T) {
	t.Helper()

	out, err := exec.Command("sudo", "tailscale", "down").CombinedOutput()
	require.NoError(t, err, "tailscale down: %s", bytes.TrimSpace(out))
	t.Log("tailscale down")

	flushDNSCache(t)

	out, err = exec.Command("sudo", "tailscale", "up").CombinedOutput()
	require.NoError(t, err, "tailscale up: %s", bytes.TrimSpace(out))
	t.Log("tailscale up")
}

func flushDNSCache(t *testing.T) {
	t.Helper()
	var cmds [][]string
	switch runtime.GOOS {
	case "darwin":
		cmds = [][]string{
			{"dscacheutil", "-flushcache"},
		}
	case "linux":
		cmds = [][]string{
			{"sudo", "resolvectl", "flush-caches"},
		}
	default:
		t.Logf("flushDNSCache: unsupported platform %s, skipping", runtime.GOOS)
		return
	}
	for _, args := range cmds {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Logf("flushDNSCache: %v: %s (non-fatal)", args, bytes.TrimSpace(out))
		} else {
			t.Logf("flushDNSCache: %v: ok", args)
		}
	}
}

func testArgoCDLogin(t *testing.T, ctx context.Context, region string, allOutputs map[string]any) {
	t.Helper()

	privateZone, ok := allOutputs["route53_hosted_zone_private"].(map[string]any)
	require.True(t, ok, "stack output 'route53_hosted_zone_private' missing or wrong type")

	host, ok := privateZone["domain_name"].(string)
	require.True(t, ok, "output 'route53_hosted_zone_private.domain_name' missing or wrong type")

	// Retrieve the secret name from the argocd_password unit output instead of constructing it.
	argocdOut, ok := allOutputs["argocd_password"].(map[string]any)
	require.True(t, ok, "stack output 'argocd_password' missing or wrong type")

	secretName, ok := argocdOut["secret_name"].(string)
	require.True(t, ok, "output 'argocd_password.secret_name' missing or wrong type")

	t.Logf("ArgoCD host: %s | password secret name: %s", host, secretName)

	t.Log("retrieving ArgoCD password from Secrets Manager")
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	require.NoError(t, err, "failed to load AWS config for region %s", region)

	svc := secretsmanager.NewFromConfig(cfg)
	secret, err := svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	})
	require.NoError(t, err, "failed to retrieve secret %q from Secrets Manager", secretName)

	var secretData struct {
		Plaintext string `json:"plaintext"`
	}
	require.NoError(t, json.Unmarshal([]byte(*secret.SecretString), &secretData), "failed to unmarshal secret JSON for %q", secretName)
	require.NotEmpty(t, secretData.Plaintext, "plaintext field is empty in secret %q", secretName)
	t.Log("password retrieved successfully")

	t.Logf("polling https://%s/healthz until ArgoCD is ready", host)
	retry.DoWithRetry(t, "wait for ArgoCD to be ready", 20, 30*time.Second, func() (string, error) {
		resp, err := http.Get("https://" + host + "/healthz")
		if err != nil {
			return "", fmt.Errorf("GET https://%s/healthz: %w", host, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("GET https://%s/healthz: got %d, want 200", host, resp.StatusCode)
		}
		return "ready", nil
	})
	t.Log("ArgoCD is healthy")

	t.Logf("logging in to ArgoCD at https://%s/api/v1/session", host)
	body, err := json.Marshal(map[string]string{
		"username": "admin",
		"password": secretData.Plaintext,
	})
	require.NoError(t, err, "failed to marshal ArgoCD login request body")

	resp, err := http.Post("https://"+host+"/api/v1/session", "application/json", bytes.NewReader(body))
	require.NoError(t, err, "POST https://%s/api/v1/session failed", host)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "ArgoCD login returned unexpected status: got %d, want 200", resp.StatusCode)

	var session struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&session), "failed to decode ArgoCD session response")
	assert.NotEmpty(t, session.Token, "ArgoCD session token is empty — login may have succeeded but returned no token")
	t.Log("ArgoCD login succeeded and session token received")
}
