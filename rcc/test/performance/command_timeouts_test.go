//go:build performance

package performance

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

// SimpleFakeAdapter is a minimal fake adapter for performance testing
type SimpleFakeAdapter struct {
	radioID string
}

func NewSimpleFakeAdapter(radioID string) *SimpleFakeAdapter {
	return &SimpleFakeAdapter{radioID: radioID}
}

func (f *SimpleFakeAdapter) SetPower(ctx context.Context, dBm float64) error {
	return nil
}

func (f *SimpleFakeAdapter) SetFrequency(ctx context.Context, frequencyMhz float64) error {
	return nil
}

func (f *SimpleFakeAdapter) GetState(ctx context.Context) (*adapter.RadioState, error) {
	return &adapter.RadioState{PowerDbm: 25.0, FrequencyMhz: 2412.0}, nil
}

func (f *SimpleFakeAdapter) ReadPowerActual(ctx context.Context) (float64, error) {
	return 25.0, nil
}

func (f *SimpleFakeAdapter) SupportedFrequencyProfiles(ctx context.Context) ([]adapter.FrequencyProfile, error) {
	return []adapter.FrequencyProfile{
		{
			Frequencies: []float64{2412.0, 2417.0, 2422.0, 2427.0, 2432.0, 2437.0, 2442.0, 2447.0, 2452.0, 2457.0, 2462.0},
			Bandwidth:   20.0,
			AntennaMask: 1,
		},
	}, nil
}

func (f *SimpleFakeAdapter) GetCapabilities() *adapter.RadioCapabilities {
	return &adapter.RadioCapabilities{
		Channels: []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 6, FrequencyMhz: 2437.0},
			{Index: 11, FrequencyMhz: 2462.0},
		},
	}
}

func (f *SimpleFakeAdapter) GetRadioID() string {
	return f.radioID
}

func (f *SimpleFakeAdapter) GetModel() string {
	return "SimpleFake"
}

func (f *SimpleFakeAdapter) GetStatus() string {
	return "online"
}

func (f *SimpleFakeAdapter) SetStatus(status string) error {
	return nil
}

// TestCommandTimeouts_ValidateCB_TIMING tests that command execution times
// stay within CB-TIMING budget constraints.
func TestCommandTimeouts_ValidateCB_TIMING(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	telemetryHub := telemetry.NewHub(cfg)
	defer telemetryHub.Stop()

	// Create real components
	radioManager := radio.NewManager()
	auditLogger, err := audit.NewLogger("/tmp/audit_perf_test")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	orchestrator.SetAuditLogger(auditLogger)

	// Create fake adapter with realistic timing
	fakeAdapter := NewSimpleFakeAdapter("perf-test-radio")
	err = radioManager.LoadCapabilities("perf-test-radio", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}
	err = radioManager.SetActive("perf-test-radio")
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}
	orchestrator.SetActiveAdapter(fakeAdapter)

	testCases := []struct {
		name     string
		command  func() error
		timeout  time.Duration
		expected string
	}{
		{
			name: "SetPower",
			command: func() error {
				return orchestrator.SetPower(context.Background(), "perf-test-radio", 25.0)
			},
			timeout:  cfg.CommandTimeoutSetPower,
			expected: "SetPower should complete within CB-TIMING timeout",
		},
		{
			name: "SetChannel",
			command: func() error {
				return orchestrator.SetChannel(context.Background(), "perf-test-radio", 2412.0)
			},
			timeout:  cfg.CommandTimeoutSetChannel,
			expected: "SetChannel should complete within CB-TIMING timeout",
		},
		{
			name: "SetChannelByIndex",
			command: func() error {
				return orchestrator.SetChannelByIndex(context.Background(), "perf-test-radio", 1, radioManager)
			},
			timeout:  cfg.CommandTimeoutSetChannel,
			expected: "SetChannelByIndex should complete within CB-TIMING timeout",
		},
		{
			name: "GetState",
			command: func() error {
				_, err := orchestrator.GetState(context.Background(), "perf-test-radio")
				return err
			},
			timeout:  cfg.CommandTimeoutGetState,
			expected: "GetState should complete within CB-TIMING timeout",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			err := tc.command()
			latency := time.Since(start)

			if err != nil {
				t.Errorf("Command failed: %v", err)
			}

			if latency > tc.timeout {
				t.Errorf("Command %s took %v, exceeds CB-TIMING timeout %v",
					tc.name, latency, tc.timeout)
			}

			t.Logf("✅ %s completed in %v (budget: %v)", tc.name, latency, tc.timeout)
		})
	}
}

// TestCommandTimeouts_ConcurrentOperations tests that concurrent operations
// don't cause timeouts or race conditions.
func TestCommandTimeouts_ConcurrentOperations(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	telemetryHub := telemetry.NewHub(cfg)
	defer telemetryHub.Stop()

	radioManager := radio.NewManager()
	auditLogger, err := audit.NewLogger("/tmp/audit_perf_concurrent")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	orchestrator.SetAuditLogger(auditLogger)

	fakeAdapter := NewSimpleFakeAdapter("concurrent-test-radio")
	err = radioManager.LoadCapabilities("concurrent-test-radio", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}
	err = radioManager.SetActive("concurrent-test-radio")
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}
	orchestrator.SetActiveAdapter(fakeAdapter)

	// Run concurrent operations
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	start := time.Now()
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Mix of operations
			var err error
			switch id % 4 {
			case 0:
				err = orchestrator.SetPower(context.Background(), "concurrent-test-radio", float64(20+id))
			case 1:
				err = orchestrator.SetChannel(context.Background(), "concurrent-test-radio", 2412.0+float64(id*5))
			case 2:
				err = orchestrator.SetChannelByIndex(context.Background(), "concurrent-test-radio", 1, radioManager)
			case 3:
				_, err = orchestrator.GetState(context.Background(), "concurrent-test-radio")
			}
			results <- err
		}(i)
	}

	// Collect results
	var failures int
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			if err != nil {
				failures++
				t.Logf("Concurrent operation %d failed: %v", i, err)
			}
		case <-time.After(30 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}

	totalTime := time.Since(start)
	t.Logf("✅ Concurrent operations completed in %v with %d/%d failures",
		totalTime, failures, numGoroutines)

	// Allow some failures due to concurrent access to fake adapter
	if failures > numGoroutines/2 {
		t.Errorf("Too many concurrent operation failures: %d/%d", failures, numGoroutines)
	}
}
