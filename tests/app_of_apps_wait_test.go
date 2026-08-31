package tests

import (
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
	appOfAppsRetries = 60
	appOfAppsSleep   = 30 * time.Second
)

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
func waitForAppOfApps(t *testing.T) {
	t.Helper()

	for attempt := 1; attempt <= appOfAppsRetries+1; attempt++ {
		out, err := exec.Command("kubectl", "get", "application", appOfAppsName, "-n", "argocd", "-o", "json").Output()
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
		}

		if attempt > appOfAppsRetries {
			break
		}
		time.Sleep(appOfAppsSleep)
	}

	t.Fatalf("[ERROR] app-of-apps did not become Synced and Healthy after %d retries", appOfAppsRetries)
}
