// Package command provides performance benchmarks for the orchestrator.
package command

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adapter/silvusmock"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

func BenchmarkSetPower(b *testing.B) {
	// Setup orchestrator with mock
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	// Create temporary directory for audit logs
	tempDir := b.TempDir()
	aud, err := audit.NewLogger(tempDir)
	if err != nil {
		b.Fatalf("Failed to create audit logger: %v", err)
	}
	defer func() { _ = aud.Close() }()
	rm := radio.NewManager()

	// Register SilvusMock
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	err = rm.LoadCapabilities("silvus-001", silvus, 5*time.Second)
	if err != nil {
		b.Fatalf("Failed to load capabilities: %v", err)
	}

	err = rm.SetActive("silvus-001")
	if err != nil {
		b.Fatalf("Failed to set active radio: %v", err)
	}

	orch := NewOrchestrator(hub, cfg)
	orch.SetRadioManager(rm)
	orch.SetAuditLogger(aud)
	orch.SetActiveAdapter(silvus)

	b.ResetTimer()

	// Run b.N iterations of SetPower
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		err := orch.SetPower(ctx, "silvus-001", float64(10+i%10))
		if err != nil {
			b.Fatalf("SetPower failed: %v", err)
		}
	}
}

func BenchmarkSetPowerWithoutTelemetry(b *testing.B) {
	// Setup orchestrator without telemetry
	cfg := config.LoadCBTimingBaseline()
	// Create temporary directory for audit logs
	tempDir := b.TempDir()
	aud, err := audit.NewLogger(tempDir)
	if err != nil {
		b.Fatalf("Failed to create audit logger: %v", err)
	}
	defer func() { _ = aud.Close() }()
	rm := radio.NewManager()

	// Register SilvusMock
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	err = rm.LoadCapabilities("silvus-001", silvus, 5*time.Second)
	if err != nil {
		b.Fatalf("Failed to load capabilities: %v", err)
	}

	err = rm.SetActive("silvus-001")
	if err != nil {
		b.Fatalf("Failed to set active radio: %v", err)
	}

	// Create orchestrator without telemetry hub
	orch := NewOrchestrator(nil, cfg)
	orch.SetRadioManager(rm)
	orch.SetAuditLogger(aud)
	orch.SetActiveAdapter(silvus)

	b.ResetTimer()

	// Run b.N iterations of SetPower
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		err := orch.SetPower(ctx, "silvus-001", float64(10+i%10))
		if err != nil {
			b.Fatalf("SetPower failed: %v", err)
		}
	}
}

func BenchmarkSetChannel(b *testing.B) {
	// Setup orchestrator with mock
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	// Create temporary directory for audit logs
	tempDir := b.TempDir()
	aud, err := audit.NewLogger(tempDir)
	if err != nil {
		b.Fatalf("Failed to create audit logger: %v", err)
	}
	defer func() { _ = aud.Close() }()
	rm := radio.NewManager()

	ctx := context.Background()

	// Register SilvusMock
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	err = rm.LoadCapabilities("silvus-001", silvus, 5*time.Second)
	if err != nil {
		b.Fatalf("Failed to load capabilities: %v", err)
	}

	err = rm.SetActive("silvus-001")
	if err != nil {
		b.Fatalf("Failed to set active radio: %v", err)
	}

	orch := NewOrchestrator(hub, cfg)
	orch.SetRadioManager(rm)
	orch.SetAuditLogger(aud)
	orch.SetActiveAdapter(silvus)

	b.ResetTimer()

	// Run b.N iterations of SetChannel
	for i := 0; i < b.N; i++ {
		channel := 2412.0 + float64(i%3)*5.0 // Cycle through frequencies 2412, 2417, 2422
		err := orch.SetChannel(ctx, "silvus-001", channel)
		if err != nil {
			b.Fatalf("SetChannel failed: %v", err)
		}
	}
}

func BenchmarkGetState(b *testing.B) {
	// Setup orchestrator with mock
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	// Create temporary directory for audit logs
	tempDir := b.TempDir()
	aud, err := audit.NewLogger(tempDir)
	if err != nil {
		b.Fatalf("Failed to create audit logger: %v", err)
	}
	defer func() { _ = aud.Close() }()
	rm := radio.NewManager()

	ctx := context.Background()

	// Register SilvusMock
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	err = rm.LoadCapabilities("silvus-001", silvus, 5*time.Second)
	if err != nil {
		b.Fatalf("Failed to load capabilities: %v", err)
	}

	err = rm.SetActive("silvus-001")
	if err != nil {
		b.Fatalf("Failed to set active radio: %v", err)
	}

	orch := NewOrchestrator(hub, cfg)
	orch.SetRadioManager(rm)
	orch.SetAuditLogger(aud)
	orch.SetActiveAdapter(silvus)

	b.ResetTimer()

	// Run b.N iterations of GetState
	for i := 0; i < b.N; i++ {
		_, err := orch.GetState(ctx, "silvus-001")
		if err != nil {
			b.Fatalf("GetState failed: %v", err)
		}
	}
}

func BenchmarkOrchestratorConcurrent(b *testing.B) {
	// Setup orchestrator with mock
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	// Create temporary directory for audit logs
	tempDir := b.TempDir()
	aud, err := audit.NewLogger(tempDir)
	if err != nil {
		b.Fatalf("Failed to create audit logger: %v", err)
	}
	defer func() { _ = aud.Close() }()
	rm := radio.NewManager()

	ctx := context.Background()

	// Register SilvusMock
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	err = rm.LoadCapabilities("silvus-001", silvus, 5*time.Second)
	if err != nil {
		b.Fatalf("Failed to load capabilities: %v", err)
	}

	err = rm.SetActive("silvus-001")
	if err != nil {
		b.Fatalf("Failed to set active radio: %v", err)
	}

	orch := NewOrchestrator(hub, cfg)
	orch.SetRadioManager(rm)
	orch.SetAuditLogger(aud)
	orch.SetActiveAdapter(silvus)

	b.ResetTimer()

	// Run b.N iterations with concurrent operations
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of operations
			switch b.N % 4 {
			case 0:
				_ = orch.SetPower(ctx, "silvus-001", 10)
			case 1:
				_ = orch.SetChannel(ctx, "silvus-001", 2412.0)
			case 2:
				_, _ = orch.GetState(ctx, "silvus-001")
			case 3:
				// GetPower method doesn't exist, use GetState instead
				_, _ = orch.GetState(ctx, "silvus-001")
			}
		}
	})
}
