// Package command defines ports (interfaces) for orchestrator operations.
package command

import (
	"context"
	"errors"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/radio"
)

// OrchestratorPort defines the minimal interface the API needs from the orchestrator.
type OrchestratorPort interface {
	SelectRadio(ctx context.Context, radioID string) error
	GetState(ctx context.Context, radioID string) (*adapter.RadioState, error)
	SetPower(ctx context.Context, radioID string, powerDbm float64) error
	SetChannel(ctx context.Context, radioID string, frequencyMhz float64) error
	SetChannelByIndex(ctx context.Context, radioID string, channelIndex int, radioManager RadioManager) error
}

// RadioManager interface for channel index resolution
type RadioManager interface {
	GetRadio(radioID string) (*radio.Radio, error)
	SetActive(radioID string) error
}

// ErrNotFound indicates a requested radio was not found.
var ErrNotFound = errors.New("NOT_FOUND")

// ErrInvalidParameter indicates a required parameter is missing or structurally invalid.
var ErrInvalidParameter = errors.New("BAD_REQUEST")
