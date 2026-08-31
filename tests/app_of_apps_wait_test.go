package tests

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// appOfAppsName is the ArgoCD Application from the argocd_app_of_apps catalog unit
// (values.name in terragrunt.stack.hcl). Its health aggregates every child Application it
// renders, so there's no need to enumerate them individually.
const appOfAppsName = "app-of-apps"

const (
	appOfAppsRetries    = 60
	appOfAppsSleep      = 30 * time.Second
	appOfAppsStallAfter = 10 * time.Minute

	// kubectlTimeout bounds each individual kubectl call so a stuck API server can't
	// hang the poll loop past a single attempt.
	kubectlTimeout = 30 * time.Second
)

// waitForAppOfApps polls the app-of-apps Application until ArgoCD reports it Synced
// and Healthy, meaning every child application it renders is also Synced and Healthy.
// Stall detection watches every Application in the argocd namespace, not just the
// app-of-apps parent. The parent's own sync and health state can sit at OutOfSync and
// Progressing for a while even as its children deploy normally one by one, so only
// when no application's state changes for appOfAppsStallAfter is it considered stuck,
// and the test fails early instead of burning the full retry budget.
func waitForAppOfApps(t *testing.T) {
	t.Helper()

	stall := newStallDetector(appOfAppsStallAfter)

	for attempt := 1; attempt <= appOfAppsRetries+1; attempt++ {
		list, err := listApps(t)
		if err != nil {
			t.Logf("[ERROR] kubectl get application attempt %d/%d failed: %v", attempt, appOfAppsRetries+1, err)
		} else {
			app, found := list.find(appOfAppsName)
			require.True(t, found, "[ERROR] application %s not found in argocd namespace", appOfAppsName)

			sync, health := app.Status.Sync.Status, app.Status.Health.Status
			t.Logf("[INFO] app-of-apps attempt %d/%d: sync=%s health=%s", attempt, appOfAppsRetries+1, sync, health)

			if sync == "Synced" && health == "Healthy" {
				t.Log("[INFO] app-of-apps is Synced and Healthy")
				return
			}

			// check if stale: no app's state changed in appOfAppsStallAfter
			if stall.Stalled(list.state()) {
				printPodsAllNamespaces(t)
				t.Fatalf("[ERROR] app-of-apps stuck at sync=%s health=%s: no child application changed state for %s", sync, health, appOfAppsStallAfter)
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
