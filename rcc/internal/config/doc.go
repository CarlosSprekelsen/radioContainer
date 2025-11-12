// Package config implements the configuration store for the Radio Control Container.
//
// The config store manages channel maps per radio/band, power limits, and supports
// hot-reload with signature verification for runtime configuration updates.
//
// Architecture References:
//   - CB-TIMING §3-6: Timing configuration constraints
//   - Architecture §8.4: Configuration management patterns
package config
