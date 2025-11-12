// Package api defines ports (interfaces) for API server dependencies.
package api

import (
	"context"
	"net/http"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

// OrchestratorPort defines the minimal interface the API needs from the orchestrator.
type OrchestratorPort interface {
	SelectRadio(ctx context.Context, radioID string) error
	GetState(ctx context.Context, radioID string) (*adapter.RadioState, error)
	SetPower(ctx context.Context, radioID string, powerDbm float64) error
	SetChannel(ctx context.Context, radioID string, frequencyMhz float64) error
	SetChannelByIndex(ctx context.Context, radioID string, channelIndex int, radioManager command.RadioManager) error
}

// TelemetryPort defines the minimal interface the API needs from the telemetry hub.
type TelemetryPort interface {
	Subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

// RadioReadPort defines the minimal interface for radio read operations.
type RadioReadPort interface {
	GetRadio(radioID string) (*radio.Radio, error)
	List() *radio.RadioList
	SetActive(radioID string) error
}

// Compile-time assertions for port conformance
var _ OrchestratorPort = (*command.Orchestrator)(nil)
var _ TelemetryPort = (*telemetry.Hub)(nil)
var _ RadioReadPort = (*radio.Manager)(nil)
