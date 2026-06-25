package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terragrunt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStack deploys the staging EKS stack, verifies ArgoCD is reachable and
// login succeeds, and asserts the guestbook application returns HTTP 200.
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

	argocdHost     := unitOutput(t, allOutputs, "domain_name_argocd", "value")
	secretName     := unitOutput(t, allOutputs, "argocd_password", "secret_name")
	argocdPassword := fetchAWSSecret(t, ctx, region, secretName)
	guestbookHost  := unitOutput(t, allOutputs, "domain_name_guestbook", "value")

	testArgoCDLogin(t, argocdHost, argocdPassword)
	testGuestbook(t, guestbookHost)
}

// testArgoCDLogin asserts that ArgoCD is healthy and that a login request with
// the given credentials returns a valid session token.
func testArgoCDLogin(t *testing.T, host string, password string) {
	t.Helper()

	t.Logf("polling https://%s/healthz until ArgoCD is ready", host)
	assertReachable(t, "wait for ArgoCD to be ready", "https://"+host+"/healthz", 20, 30*time.Second)
	t.Log("ArgoCD is healthy")

	t.Logf("logging in to ArgoCD at https://%s/api/v1/session", host)
	body, err := json.Marshal(map[string]string{
		"username": "admin",
		"password": password,
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

// testGuestbook asserts that the guestbook application is reachable and returns HTTP 200.
func testGuestbook(t *testing.T, host string) {
	t.Helper()

	t.Logf("polling https://%s until guestbook is ready", host)
	assertReachable(t, "wait for guestbook to be ready", "https://"+host, 20, 30*time.Second)
	t.Log("guestbook is reachable")
}
