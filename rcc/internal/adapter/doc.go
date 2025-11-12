// Package adapter defines the radio adapter interface for the Radio Control Container.
//
// Radio adapters implement vendor-specific protocols to communicate with radio hardware.
// The IRadioAdapter interface provides a stable API contract that all adapters must implement.
//
// Architecture References:
//   - Architecture §8.5: Error code normalization
//   - ICD §4: Adapter interface specifications
package adapter
