package tests

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terragrunt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// endpointRetries and endpointSleep bound how long each endpoint check waits for its tool to
// become reachable. waitForAppOfApps already confirms every app Synced and Healthy before this
// runs, so this budget only needs to cover DNS propagation and ALB/target-group lag behind that
// health check, not app startup.
const (
	endpointRetries = 30
	endpointSleep   = 10 * time.Second
)

// endpointCheck describes one stack tool to verify after deploy. Entries with no secretUnit are
// checked for plain reachability via validate; entries with a secretUnit fetch the password from
// Secrets Manager and hand it to login instead. Add a new tool here to cover it going forward.
type endpointCheck struct {
	name       string                                           // domain_name_<name> stack output
	path       string                                           // path to poll, only used when secretUnit is empty
	validate   func(status int, body string) bool               // reachability check, only used when secretUnit is empty
	secretUnit string                                           // stack unit exposing the password secret name, empty means no auth
	secretKey  string                                           // output key on secretUnit holding the secret name
	login      func(t *testing.T, host string, password string) // auth check, only used when secretUnit is set
}

// statusOK is a validate func for endpoints where reaching the page at all is proof enough.
func statusOK(status int, _ string) bool {
	return status == http.StatusOK
}

var endpointChecks = []endpointCheck{
	{name: "argocd", secretUnit: "argocd_password", secretKey: "secret_name", login: testArgoCDLogin},
	{name: "podinfo", path: "/", validate: statusOK},
	{name: "grafana", secretUnit: "grafana_password", secretKey: "secret_name", login: testGrafanaLogin},
	{name: "prometheus", path: "/-/healthy", validate: func(status int, body string) bool {
		return status == http.StatusOK && strings.Contains(body, "Healthy")
	}},
	{name: "alertmanager", path: "/api/v2/status", validate: func(status int, body string) bool {
		return status == http.StatusOK && strings.Contains(body, "cluster")
	}},
	{name: "hubble", path: "/", validate: statusOK},
	{name: "goldilocks", path: "/", validate: statusOK},
}

// TestStack deploys the staging EKS stack and validates every tool in endpointChecks.
func TestStack(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	region := os.Getenv("AWS_REGION")
	require.NotEmpty(t, region, "AWS_REGION must be set")

	// Tailscale CLI is needed to flush DNS cache after Terragrunt apply
	_, err := exec.LookPath("tailscale")
	require.NoError(t, err, "tailscale CLI not found in PATH — install it before running this test")

	// Disconnect before apply to avoid racing the in-cluster connector's split-DNS route (see
	// ci.yaml). Reconnected once the stack, and the connector, are up (reconnectTailscale below).
	out, err := exec.Command("tailscale", "down").CombinedOutput()
	require.NoError(t, err, "tailscale down: %s", strings.TrimSpace(string(out)))

	stackDir := "../live/staging/eks/stack"

	options := &terragrunt.Options{
		TerragruntDir:  stackDir,
		TerragruntArgs: []string{"--log-level", "error"},
	}

	defer terragrunt.DestroyAllContext(t, ctx, options)

	terragrunt.ApplyAllContext(t, ctx, options)

	// Fetched once and reused below: terragrunt output --all is expensive to re-run.
	silentOptions := &terragrunt.Options{
		TerragruntDir:  stackDir,
		TerragruntArgs: []string{"--log-level", "error"},
		Logger:         logger.Discard,
	}
	allOutputs := terragrunt.StackOutputAllContext(t, ctx, silentOptions)

	updateKubeconfig(t, allOutputs, region)

	waitForAppOfApps(t)

	reconnectTailscale(t)

	assertStack(t, ctx, allOutputs, region)
}

// TestStackExists runs only the assertion phase against an already-deployed
// staging stack. Use this when the infrastructure is already up and you want to
// iterate on the Go logic without triggering an apply or destroy.
//
// Usage:
//
//	go test -v -run TestStackAssertions -timeout 10m
func TestStackExists(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("skipped in CI — run locally against an already-deployed stack")
	}

	t.Parallel()

	ctx := t.Context()

	stackDir := "../live/staging/eks/stack"

	region := os.Getenv("AWS_REGION")
	require.NotEmpty(t, region, "AWS_REGION must be set")

	silentOptions := &terragrunt.Options{
		TerragruntDir:  stackDir,
		TerragruntArgs: []string{"--log-level", "error"},
		Logger:         logger.Discard,
	}
	allOutputs := terragrunt.StackOutputAllContext(t, ctx, silentOptions)

	assertStack(t, ctx, allOutputs, region)
}

// pollUntilReady polls url until validate passes, sleeping sleepBetweenRetries between attempts
// up to retries times.
func pollUntilReady(t *testing.T, url string, retries int, sleepBetweenRetries time.Duration, validate func(int, string) bool) {
	t.Helper()

	for attempt := 1; attempt <= retries+1; attempt++ {
		status, body, err := http_helper.HttpGetE(t, url, nil)
		if err == nil && validate(status, body) {
			return
		}

		if err != nil {
			t.Logf("poll %s attempt %d/%d failed: %v", url, attempt, retries+1, err)
		} else {
			t.Logf("poll %s attempt %d/%d failed: unexpected status %d", url, attempt, retries+1, status)
		}

		if attempt > retries {
			break
		}

		time.Sleep(sleepBetweenRetries)
	}

	t.Fatalf("%s did not become ready after %d retries", url, retries)
}

// assertStack runs every check in endpointChecks against allOutputs.
func assertStack(t *testing.T, ctx context.Context, allOutputs map[string]any, region string) {
	t.Helper()

	for _, ep := range endpointChecks {
		host := unitOutput(t, allOutputs, "domain_name_"+ep.name, "value")

		if ep.secretUnit == "" {
			url := "https://" + host + ep.path
			t.Logf("polling %s until %s is ready", url, ep.name)
			pollUntilReady(t, url, endpointRetries, endpointSleep, ep.validate)
			t.Logf("%s is healthy", ep.name)
			continue
		}

		secretName := unitOutput(t, allOutputs, ep.secretUnit, ep.secretKey)
		password := fetchAWSSecret(t, ctx, region, secretName)
		ep.login(t, host, password)
	}

	testNoUnexpectedNetworkDrops(t)
}

// testNoUnexpectedNetworkDrops asserts Cilium hasn't dropped any traffic beyond the
// "Unsupported L3 protocol DROPPED (ICMPv6 RouterSolicitation)" noise this IPv4-only cluster
// always produces regardless of policy (see docs/network-policies.md in the catalog repo).
// hubble observe without -f reads the buffered flows and exits, no follow/timeout needed.
func testNoUnexpectedNetworkDrops(t *testing.T) {
	t.Helper()

	_, err := exec.LookPath("hubble")
	require.NoError(t, err, "hubble CLI not found in PATH — install it before running this test")

	// Only stdout carries flow lines, hubble logs warnings (e.g. CLI/relay version mismatch) to stderr.
	out, err := exec.Command("hubble", "observe", "--verdict", "DROPPED", "-P").Output()
	require.NoError(t, err, "hubble observe failed")

	var unexpected []string
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" && !strings.Contains(line, "Unsupported L3 protocol DROPPED") {
			unexpected = append(unexpected, line)
		}
	}
	assert.Empty(t, unexpected, "unexpected dropped flows:\n%s", strings.Join(unexpected, "\n"))
}

// testArgoCDLogin asserts that ArgoCD is healthy and that a login request with
// the given credentials returns a valid session token.
func testArgoCDLogin(t *testing.T, host string, password string) {
	t.Helper()

	t.Logf("polling https://%s/healthz until ArgoCD is ready", host)
	pollUntilReady(t, "https://"+host+"/healthz", endpointRetries, endpointSleep, statusOK)
	t.Log("ArgoCD is healthy")

	t.Logf("logging in to ArgoCD at https://%s/api/v1/session", host)
	var session struct {
		Token string `json:"token"`
	}
	postLoginJSON(t, "https://"+host+"/api/v1/session", map[string]string{
		"username": "admin",
		"password": password,
	}, &session)
	assert.NotEmpty(t, session.Token, "ArgoCD session token is empty — login may have succeeded but returned no token")
	t.Log("ArgoCD login succeeded and session token received")
}

// testGrafanaLogin asserts that Grafana is healthy and that an admin login request with
// the given credentials succeeds.
func testGrafanaLogin(t *testing.T, host string, password string) {
	t.Helper()

	t.Logf("polling https://%s/api/health until Grafana is ready", host)
	pollUntilReady(t, "https://"+host+"/api/health", endpointRetries, endpointSleep, statusOK)
	t.Log("Grafana is healthy")

	t.Logf("logging in to Grafana at https://%s/login", host)
	var loginResponse struct {
		Message string `json:"message"`
	}
	postLoginJSON(t, "https://"+host+"/login", map[string]string{
		"user":     "admin",
		"password": password,
	}, &loginResponse)
	assert.NotEmpty(t, loginResponse.Message, "Grafana login response message is empty — login may have failed")
	t.Log("Grafana login succeeded")
}
