package tests

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

// updateKubeconfig points kubectl at the deployed cluster so app_of_apps_wait_test.go
// and printPodsAllNamespaces can reach it.
func updateKubeconfig(t *testing.T, allOutputs map[string]any, region string) {
	t.Helper()

	clusterName := unitOutput(t, allOutputs, "cluster", "cluster_name")

	out, err := exec.Command("aws", "eks", "update-kubeconfig", "--region", region, "--name", clusterName).CombinedOutput()
	require.NoError(t, err, "aws eks update-kubeconfig: %s", bytes.TrimSpace(out))
	t.Logf("[INFO] kubeconfig updated for cluster %s", clusterName)
}
