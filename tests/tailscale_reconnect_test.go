package tests

import (
	"bytes"
	"os/exec"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// reconnectTailscale cycles the Tailscale connection down and back up, flushing
// the DNS cache in between so that newly created private DNS records are resolved
// correctly in subsequent test assertions.
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

// flushDNSCache clears the OS DNS cache on macOS and Linux. Failures are logged
// but non-fatal since some environments flush automatically.
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
