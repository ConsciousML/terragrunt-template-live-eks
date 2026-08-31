package tests

import (
	"bytes"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// reconnectTailscale cycles the Tailscale connection down and back up, flushing
// the DNS cache in between so that newly created private DNS records are resolved
// correctly in subsequent test assertions.
func reconnectTailscale(t *testing.T) {
	t.Helper()

	out, err := exec.Command("tailscale", "down").CombinedOutput()
	require.NoError(t, err, "[ERROR] tailscale down: %s", bytes.TrimSpace(out))
	t.Logf("[INFO] tailscale down: %s", bytes.TrimSpace(out))

	flushDNSCache(t)

	out, err = exec.Command("tailscale", "up").CombinedOutput()
	require.NoError(t, err, "[ERROR] tailscale up: %s", bytes.TrimSpace(out))
	t.Logf("[INFO] tailscale up: %s", bytes.TrimSpace(out))

	// let the backend settle before checking status.
	time.Sleep(2 * time.Second)

	out, err = exec.Command("tailscale", "status").CombinedOutput()
	require.NoError(t, err, "[ERROR] tailscale status: %s", bytes.TrimSpace(out))
	t.Logf("[INFO] tailscale status: %s", bytes.TrimSpace(out))
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
			// macOS doesn't reliably re-push DNS config on a plain down/up cycle.
			// Toggling accept-dns while still down forces the client to reapply it
			// once it comes back up.
			{"tailscale", "set", "--accept-dns=false"},
			{"tailscale", "set", "--accept-dns=true"},
		}
	case "linux":
		cmds = [][]string{
			{"resolvectl", "flush-caches"},
		}
	default:
		t.Logf("[INFO] flushDNSCache: unsupported platform %s, skipping", runtime.GOOS)
		return
	}
	for _, args := range cmds {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Logf("[ERROR] flushDNSCache: %v: %s (non-fatal)", args, bytes.TrimSpace(out))
		} else {
			t.Logf("[INFO] flushDNSCache: %v: ok", args)
		}
	}
}
