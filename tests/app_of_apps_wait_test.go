package tests

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// appOfAppsName is the ArgoCD Application resource created by the argocd_app_of_apps
// catalog unit (values.name in terragrunt.stack.hcl). ArgoCD has a built-in health check
// for the Application kind. This parent's health reflects the aggregate health of every
// child Application it renders, so there's no need to enumerate them individually.
const appOfAppsName = "app-of-apps"

const (
	appOfAppsRetries    = 60
	appOfAppsSleep      = 30 * time.Second
	appOfAppsStallAfter = 10 * time.Minute

	// kubectlTimeout bounds each individual kubectl call so a stuck API server can't
	// hang the poll loop past a single attempt.
	kubectlTimeout = 30 * time.Second
)

// stallDetector reports whether a state string has stayed the same for at least `after`.
type stallDetector struct {
	after      time.Duration
	lastState  string
	lastChange time.Time
}

func newStallDetector(after time.Duration) *stallDetector {
	return &stallDetector{after: after, lastChange: time.Now()}
}

// Stalled records state and reports whether it has been unchanged for at least `after`.
func (s *stallDetector) Stalled(state string) bool {
	if state != s.lastState {
		s.lastState, s.lastChange = state, time.Now()
		return false
	}
	return time.Since(s.lastChange) >= s.after
}

type argoAppStatus struct {
	Status struct {
		Sync struct {
			Status string `json:"status"`
		} `json:"sync"`
		Health struct {
			Status string `json:"status"`
		} `json:"health"`
	} `json:"status"`
}

// waitForAppOfApps polls the app-of-apps Application until ArgoCD reports it Synced
// and Healthy, meaning every child application it renders is also Synced and Healthy.
// If the sync/health tuple doesn't change for appOfAppsStallAfter, it's considered stuck
// and the test fails early instead of burning the full retry budget.
func waitForAppOfApps(t *testing.T) {
	t.Helper()

	stall := newStallDetector(appOfAppsStallAfter)

	for attempt := 1; attempt <= appOfAppsRetries+1; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), kubectlTimeout)
		out, err := exec.CommandContext(ctx, "kubectl", "get", "application", appOfAppsName, "-n", "argocd", "-o", "json").Output()
		cancel()
		if err != nil {
			t.Logf("[ERROR] kubectl get application %s attempt %d/%d failed: %v", appOfAppsName, attempt, appOfAppsRetries+1, err)
		} else {
			var app argoAppStatus
			require.NoError(t, json.Unmarshal(out, &app), "failed to unmarshal application %s json", appOfAppsName)

			sync, health := app.Status.Sync.Status, app.Status.Health.Status
			t.Logf("[INFO] app-of-apps attempt %d/%d: sync=%s health=%s", attempt, appOfAppsRetries+1, sync, health)

			if sync == "Synced" && health == "Healthy" {
				t.Log("[INFO] app-of-apps is Synced and Healthy")
				return
			}

			if stall.Stalled(sync + "/" + health) {
				printPodsAllNamespaces(t)
				t.Fatalf("[ERROR] app-of-apps stuck at sync=%s health=%s for %s", sync, health, appOfAppsStallAfter)
			}
		}

		if attempt > appOfAppsRetries {
			break
		}
		time.Sleep(appOfAppsSleep)
	}

	printPodsAllNamespaces(t)
	t.Fatalf("[ERROR] app-of-apps did not become Synced and Healthy after %d retries", appOfAppsRetries)
}

// printPodsAllNamespaces dumps kubectl get pod -A for debugging when the app-of-apps
// wait stalls or times out.
func printPodsAllNamespaces(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), kubectlTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "kubectl", "get", "pod", "-A").CombinedOutput()
	if err != nil {
		t.Logf("[ERROR] kubectl get pod -A failed: %v: %s", err, out)
		return
	}
	t.Logf("[INFO] kubectl get pod -A:\n%s", out)
}
