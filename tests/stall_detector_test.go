package tests

import "time"

// stallDetector reports whether a caller-supplied state string has stayed the same for
// at least `after`. waitForAppOfApps uses it to catch a poll loop that's making no
// progress (the composite sync and health state of every Application in the argocd
// namespace stops changing) instead of burning its full retry budget.
type stallDetector struct {
	after      time.Duration
	lastState  string
	lastChange time.Time
}

func newStallDetector(after time.Duration) *stallDetector {
	return &stallDetector{after: after, lastChange: time.Now()}
}

// Stalled records state and reports whether it has been unchanged for at least `after`.
// Each call feeds it the latest state string; the timer resets on every change.
func (s *stallDetector) Stalled(state string) bool {
	if state != s.lastState {
		s.lastState, s.lastChange = state, time.Now()
		return false
	}
	return time.Since(s.lastChange) >= s.after
}
