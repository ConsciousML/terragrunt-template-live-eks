package tests

import (
	"context"
	"encoding/json"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

type argoAppStatus struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Status struct {
		Sync struct {
			Status string `json:"status"`
		} `json:"sync"`
		Health struct {
			Status string `json:"status"`
		} `json:"health"`
	} `json:"status"`
}

type argoAppList struct {
	Items []argoAppStatus `json:"items"`
}

// listApps fetches every Application in the argocd namespace. 
func listApps(t *testing.T) (argoAppList, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), kubectlTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "kubectl", "get", "application", "-n", "argocd", "-o", "json").Output()
	if err != nil {
		return argoAppList{}, err
	}

	var list argoAppList
	if err := json.Unmarshal(out, &list); err != nil {
		return argoAppList{}, err
	}
	return list, nil
}

// state returns a composite sync and health string keyed by app name, used for stall
// detection so a child app still making progress (e.g. mid-deploy) resets the stall
// timer even while the app-of-apps parent itself reports a stale sync and health state.
func (l argoAppList) state() string {
	items := append([]argoAppStatus(nil), l.Items...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Metadata.Name < items[j].Metadata.Name
	})

	var state strings.Builder
	for _, app := range items {
		state.WriteString(app.Metadata.Name)
		state.WriteByte('=')
		state.WriteString(app.Status.Sync.Status)
		state.WriteByte('/')
		state.WriteString(app.Status.Health.Status)
		state.WriteByte(';')
	}
	return state.String()
}

// find returns the app with the given name, if present.
func (l argoAppList) find(name string) (argoAppStatus, bool) {
	for _, app := range l.Items {
		if app.Metadata.Name == name {
			return app, true
		}
	}
	return argoAppStatus{}, false
}
