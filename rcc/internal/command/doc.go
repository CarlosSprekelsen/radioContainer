// Package command implements the command orchestrator for the Radio Control Container.
//
// The orchestrator validates requests, resolves channel mappings via ConfigStore,
// calls adapter methods, emits events to TelemetryHub, and writes audit logs.
//
// Architecture References:
//   - Architecture §8.5: Error code normalization
//   - CB-TIMING §5: Command timeout constraints
package command
