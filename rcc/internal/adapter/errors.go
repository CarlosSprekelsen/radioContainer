// Package adapter defines IRadioAdapter interface from Architecture §5.
//
//   - Architecture §8.5: "Normalized error codes: INVALID_RANGE, BUSY, UNAVAILABLE, INTERNAL"
//   - Architecture §8.5.1: "Deterministic mapping with diagnostic preservation"
//
// PRE-INT-04: Deterministic Vendor Error Mapping (tables)
// This package provides table-driven error mapping to normalize vendor-specific
// error messages to standardized container error codes without heuristics.
package adapter

import (
	"errors"
	"fmt"
	"strings"
)

// Normalized container errors per Architecture §8.5
var (
	ErrInvalidRange = errors.New("INVALID_RANGE")
	ErrBusy         = errors.New("BUSY")
	ErrUnavailable  = errors.New("UNAVAILABLE")
	ErrInternal     = errors.New("INTERNAL")
)

// VendorMap defines the error token mapping for a specific vendor.
type VendorMap struct {
	Range       []string // Tokens that map to INVALID_RANGE
	Busy        []string // Tokens that map to BUSY
	Unavailable []string // Tokens that map to UNAVAILABLE
}

// VendorErrorMappings contains the deterministic error mapping tables for all vendors.
//
// README: Vendor Error Mapping Tables
// ===================================
//
// Current Silvus Tokens:
//   - Range: TX_POWER_OUT_OF_RANGE, FREQUENCY_OUT_OF_RANGE, INVALID_POWER_LEVEL,
//     INVALID_FREQUENCY, PARAMETER_OUT_OF_RANGE, VALUE_OUT_OF_BOUNDS, INVALID_PARAMETER
//   - Busy: RF_BUSY, TRANSMITTER_BUSY, RADIO_BUSY, OPERATION_IN_PROGRESS,
//     COMMAND_QUEUE_FULL, RATE_LIMITED
//   - Unavailable: NODE_UNAVAILABLE, RADIO_OFFLINE, REBOOTING, SOFT_BOOT_IN_PROGRESS,
//     SYSTEM_INITIALIZING, NOT_READY, OFFLINE
//
// How to Extend Safely:
// 1. Add new vendor entries to this map with specific token arrays
// 2. Test each token → exact normalized error mapping
// 3. Unknown tokens automatically map to INTERNAL
// 4. Use NormalizeVendorErrorWithVendor(vendorErr, payload, "vendorID") for specific vendors
// 5. Fallback to "generic" mapping for unknown vendors
//
// Note: If Silvus ICD is not available, use "generic" mapping and put Silvus behind a feature flag.
var VendorErrorMappings = map[string]VendorMap{
	"silvus": {
		Range: []string{
			"TX_POWER_OUT_OF_RANGE",
			"FREQUENCY_OUT_OF_RANGE",
			"INVALID_POWER_LEVEL",
			"INVALID_FREQUENCY",
			"PARAMETER_OUT_OF_RANGE",
			"VALUE_OUT_OF_BOUNDS",
			"INVALID_PARAMETER",
		},
		Busy: []string{
			"RF_BUSY",
			"TRANSMITTER_BUSY",
			"RADIO_BUSY",
			"OPERATION_IN_PROGRESS",
			"COMMAND_QUEUE_FULL",
			"RATE_LIMITED",
		},
		Unavailable: []string{
			"NODE_UNAVAILABLE",
			"RADIO_OFFLINE",
			"REBOOTING",
			"SOFT_BOOT_IN_PROGRESS",
			"SYSTEM_INITIALIZING",
			"NOT_READY",
			"OFFLINE",
		},
	},
	"generic": {
		Range: []string{
			"OUT_OF_RANGE",
			"INVALID_PARAMETER",
			"INVALID_RANGE",
			"BAD_VALUE",
			"RANGE_ERROR",
		},
		Busy: []string{
			"BUSY",
			"RETRY",
			"RATE_LIMIT",
			"TOO_MANY_REQUESTS",
			"BACKOFF",
		},
		Unavailable: []string{
			"UNAVAILABLE",
			"REBOOT",
			"SOFT_BOOT",
			"OFFLINE",
			"NOT_READY",
		},
	},
}

// VendorError wraps vendor error with diagnostic details per Architecture §8.5.1
type VendorError struct {
	Code     error       // Normalized container code
	Original error       // Vendor error
	Details  interface{} // Vendor payload (opaque)
}

func (e *VendorError) Error() string {
	return fmt.Sprintf("%v (vendor: %v)", e.Code, e.Original)
}

func (e *VendorError) Unwrap() error {
	return e.Code
}

// NormalizeVendorError maps vendor errors to Architecture §8.5 codes using table-driven matching.
func NormalizeVendorError(vendorErr error, vendorPayload interface{}) error {
	return NormalizeVendorErrorWithVendor(vendorErr, vendorPayload, "generic")
}

// NormalizeVendorErrorWithVendor maps vendor errors using specific vendor mapping tables.
func NormalizeVendorErrorWithVendor(vendorErr error, vendorPayload interface{}, vendorID string) error {
	if vendorErr == nil {
		return nil
	}

	msg := vendorErr.Error()
	code := mapVendorErrorToCode(msg, vendorID)

	return &VendorError{
		Code:     code,
		Original: vendorErr,
		Details:  vendorPayload,
	}
}

// mapVendorErrorToCode maps a vendor error message to normalized error code using table-driven matching.
func mapVendorErrorToCode(msg string, vendorID string) error {
	// Get vendor mapping, fallback to generic if vendor not found
	vendorMap, exists := VendorErrorMappings[vendorID]
	if !exists {
		vendorMap = VendorErrorMappings["generic"]
	}

	upperMsg := strings.ToUpper(msg)

	// Check for exact token matches in each category
	for _, token := range vendorMap.Range {
		if strings.Contains(upperMsg, strings.ToUpper(token)) {
			return ErrInvalidRange
		}
	}

	for _, token := range vendorMap.Busy {
		if strings.Contains(upperMsg, strings.ToUpper(token)) {
			return ErrBusy
		}
	}

	for _, token := range vendorMap.Unavailable {
		if strings.Contains(upperMsg, strings.ToUpper(token)) {
			return ErrUnavailable
		}
	}

	// Unknown token maps to INTERNAL
	return ErrInternal
}
