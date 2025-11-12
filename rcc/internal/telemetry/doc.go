// Package telemetry implements the telemetry hub for the Radio Control Container.
//
// The telemetry hub fans out events to all SSE clients and buffers the last N events
// per client for reconnection support using Last-Event-ID headers.
//
// Architecture References:
//   - Telemetry SSE §2: Event streaming protocol
//   - CB-TIMING §6: Event buffering constraints
package telemetry
