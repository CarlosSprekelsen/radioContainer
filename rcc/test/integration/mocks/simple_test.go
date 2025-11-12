//go:build integration

package mocks

import (
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
)

func TestSimpleTypes(t *testing.T) {
	// Try to create a channel
	ch := adapter.Channel{
		Index:        1,
		FrequencyMhz: 2412.0,
	}
	
	// Try to create capabilities
	caps := adapter.RadioCapabilities{
		Channels: []adapter.Channel{ch},
	}
	
	t.Logf("Channel: %+v", ch)
	t.Logf("Capabilities: %+v", caps)
}


