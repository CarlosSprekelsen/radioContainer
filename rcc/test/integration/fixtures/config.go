//go:build integration

package fixtures

import (
	"time"

	"github.com/radio-control/rcc/internal/config"
)

// TestTimingConfig returns a fast, deterministic timing config for integration tests.
// This isolates tests from CB-TIMING document changes and provides fast test execution.
func TestTimingConfig() *config.TimingConfig {
	return &config.TimingConfig{
		// Fast timeouts for integration tests
		HeartbeatInterval: 100 * time.Millisecond,
		HeartbeatJitter:   10 * time.Millisecond,
		HeartbeatTimeout:  500 * time.Millisecond,

		// Fast probe cadences
		ProbeNormalInterval:    50 * time.Millisecond,
		ProbeRecoveringInitial: 25 * time.Millisecond,
		ProbeRecoveringBackoff: 1.5,
		ProbeRecoveringMax:     100 * time.Millisecond,
		ProbeOfflineInitial:    100 * time.Millisecond,
		ProbeOfflineBackoff:    2.0,
		ProbeOfflineMax:        200 * time.Millisecond,

		// Fast command timeouts
		CommandTimeoutSetPower:    50 * time.Millisecond,
		CommandTimeoutSetChannel:  50 * time.Millisecond,
		CommandTimeoutSelectRadio: 50 * time.Millisecond,
		CommandTimeoutGetState:    50 * time.Millisecond,

		// Small event buffer for tests
		EventBufferSize:      10,
		EventBufferRetention: 1 * time.Second,
	}
}
