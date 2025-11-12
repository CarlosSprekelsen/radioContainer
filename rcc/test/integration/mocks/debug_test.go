//go:build integration

package mocks

import (
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
)

func TestDebugTypes(t *testing.T) {
	// Try to create types
	var ch adapter.Channel
	ch.Index = 1
	ch.FrequencyMhz = 2412.0
	
	var caps adapter.RadioCapabilities
	caps.Channels = []adapter.Channel{ch}
	
	// This should work
	_ = ch
	_ = caps
	
	t.Logf("Types work: Channel=%+v, Capabilities=%+v", ch, caps)
}


