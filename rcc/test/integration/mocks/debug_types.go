//go:build integration

package mocks

import (
	"github.com/radio-control/rcc/internal/adapter"
)

// DebugTypes tests if adapter types can be used
func DebugTypes() {
	// Try to create a channel
	var ch adapter.Channel
	ch.Index = 1
	ch.FrequencyMhz = 2412.0

	// Try to create capabilities
	var caps adapter.RadioCapabilities
	caps.Channels = []adapter.Channel{ch}

	// This should compile if types are accessible
	_ = caps
}
