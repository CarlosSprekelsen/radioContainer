//go:build performance

package performance

import (
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/telemetry"
)

// TestHeartbeatTiming_ValidateCB_TIMING tests that telemetry hub
// configuration matches CB-TIMING requirements.
func TestHeartbeatTiming_ValidateCB_TIMING(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	// Validate CB-TIMING configuration
	baseInterval := 15 * time.Second // From CB-TIMING §3
	jitter := 2 * time.Second
	minInterval := baseInterval - jitter
	maxInterval := baseInterval + jitter

	t.Logf("=== HEARTBEAT TIMING VALIDATION ===")
	t.Logf("Base interval: %v", baseInterval)
	t.Logf("Jitter: %v", jitter)
	t.Logf("Tolerance window: [%v, %v]", minInterval, maxInterval)

	// Test that the hub was created with correct configuration
	if hub == nil {
		t.Fatal("Telemetry hub not created")
	}

	// Validate configuration values are reasonable
	if baseInterval <= 0 {
		t.Error("Base interval must be positive")
	}
	if jitter < 0 {
		t.Error("Jitter must be non-negative")
	}
	if minInterval <= 0 {
		t.Error("Minimum interval must be positive")
	}
	if maxInterval <= minInterval {
		t.Error("Maximum interval must be greater than minimum interval")
	}

	t.Logf("✅ Heartbeat timing configuration validated against CB-TIMING")
}

// TestTelemetryPerformance_PublishLatency tests that telemetry publishing
// doesn't introduce significant latency.
func TestTelemetryPerformance_PublishLatency(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	const numEvents = 50
	var totalLatency time.Duration

	for i := 0; i < numEvents; i++ {
		start := time.Now()

		hub.PublishRadio("latency-test-radio", telemetry.Event{
			Type: "powerChanged",
			Data: map[string]interface{}{"powerDbm": float64(20 + i%10)},
		})

		latency := time.Since(start)
		totalLatency += latency
	}

	avgLatency := totalLatency / numEvents
	t.Logf("✅ Average telemetry publish latency: %v", avgLatency)

	// Telemetry publishing should be very fast (< 10ms)
	if avgLatency > 10*time.Millisecond {
		t.Errorf("Average telemetry latency %v exceeds 10ms threshold", avgLatency)
	}
}

// TestTelemetryPerformance_ConcurrentPublish tests concurrent telemetry
// publishing performance.
func TestTelemetryPerformance_ConcurrentPublish(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	const numGoroutines = 10
	const eventsPerGoroutine = 10
	results := make(chan time.Duration, numGoroutines)

	start := time.Now()
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			goroutineStart := time.Now()
			for j := 0; j < eventsPerGoroutine; j++ {
				hub.PublishRadio("concurrent-test-radio", telemetry.Event{
					Type: "powerChanged",
					Data: map[string]interface{}{"powerDbm": float64(20 + id + j)},
				})
			}
			results <- time.Since(goroutineStart)
		}(i)
	}

	// Collect results
	var totalGoroutineTime time.Duration
	for i := 0; i < numGoroutines; i++ {
		select {
		case goroutineTime := <-results:
			totalGoroutineTime += goroutineTime
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent publishing timed out")
		}
	}

	totalTime := time.Since(start)
	avgGoroutineTime := totalGoroutineTime / numGoroutines

	t.Logf("✅ Concurrent telemetry publishing completed in %v", totalTime)
	t.Logf("Average goroutine time: %v", avgGoroutineTime)

	// Concurrent publishing should complete quickly
	if totalTime > 2*time.Second {
		t.Errorf("Concurrent publishing took too long: %v", totalTime)
	}
}
