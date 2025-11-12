package adapter

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// TestAdapterErrorEnvelopesGolden tests adapter error normalization against golden files
func TestAdapterErrorEnvelopesGolden(t *testing.T) {
	tests := []struct {
		name       string
		vendorErr  error
		vendorID   string
		payload    interface{}
		goldenFile string
	}{
		{
			name:       "silvus_power_out_of_range",
			vendorErr:  fmt.Errorf("TX_POWER_OUT_OF_RANGE: power level 50 is outside valid range [0, 39]"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"requestedPower": 50, "validRange": []int{0, 39}},
			goldenFile: "silvus_power_out_of_range.json",
		},
		{
			name:       "silvus_frequency_out_of_range",
			vendorErr:  fmt.Errorf("FREQUENCY_OUT_OF_RANGE: frequency 100000.0 is outside valid range [100, 6000]"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"requestedFreq": 100000.0, "validRange": []float64{100, 6000}},
			goldenFile: "silvus_frequency_out_of_range.json",
		},
		{
			name:       "silvus_rf_busy",
			vendorErr:  fmt.Errorf("RF_BUSY: transmitter is currently busy with another operation"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"currentOperation": "frequency_change", "estimatedCompletion": "30s"},
			goldenFile: "silvus_rf_busy.json",
		},
		{
			name:       "silvus_radio_offline",
			vendorErr:  fmt.Errorf("RADIO_OFFLINE: radio device is not responding"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"lastSeen": "2024-01-01T12:00:00Z", "connectionStatus": "disconnected"},
			goldenFile: "silvus_radio_offline.json",
		},
		{
			name:       "silvus_operation_in_progress",
			vendorErr:  fmt.Errorf("OPERATION_IN_PROGRESS: command queue is full, please retry later"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"queueSize": 10, "maxQueueSize": 10, "retryAfter": "5s"},
			goldenFile: "silvus_operation_in_progress.json",
		},
		{
			name:       "silvus_invalid_parameter",
			vendorErr:  fmt.Errorf("INVALID_PARAMETER: channel index must be positive integer"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"parameter": "channelIndex", "value": -1, "expectedType": "positive integer"},
			goldenFile: "silvus_invalid_parameter.json",
		},
		{
			name:       "silvus_node_unavailable",
			vendorErr:  fmt.Errorf("NODE_UNAVAILABLE: radio node is not responding to commands"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"nodeId": "node-123", "lastHeartbeat": "2024-01-01T11:55:00Z"},
			goldenFile: "silvus_node_unavailable.json",
		},
		{
			name:       "silvus_rebooting",
			vendorErr:  fmt.Errorf("REBOOTING: radio is currently rebooting, please wait"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"rebootStartTime": "2024-01-01T12:00:00Z", "estimatedCompletion": "2m"},
			goldenFile: "silvus_rebooting.json",
		},
		{
			name:       "generic_out_of_range",
			vendorErr:  fmt.Errorf("OUT_OF_RANGE: parameter value is outside allowed range"),
			vendorID:   "generic",
			payload:    map[string]interface{}{"parameter": "power", "value": 100, "range": []int{0, 50}},
			goldenFile: "generic_out_of_range.json",
		},
		{
			name:       "generic_busy",
			vendorErr:  fmt.Errorf("BUSY: system is currently busy"),
			vendorID:   "generic",
			payload:    map[string]interface{}{"busyReason": "maintenance", "retryAfter": "10s"},
			goldenFile: "generic_busy.json",
		},
		{
			name:       "generic_unavailable",
			vendorErr:  fmt.Errorf("UNAVAILABLE: service is temporarily unavailable"),
			vendorID:   "generic",
			payload:    map[string]interface{}{"reason": "system maintenance", "estimatedRestore": "30m"},
			goldenFile: "generic_unavailable.json",
		},
		{
			name:       "unknown_vendor_error",
			vendorErr:  fmt.Errorf("UNKNOWN_ERROR: some unknown vendor error occurred"),
			vendorID:   "unknown",
			payload:    map[string]interface{}{"errorCode": "UNKNOWN_123", "message": "unexpected error"},
			goldenFile: "unknown_vendor_error.json",
		},
		{
			name:       "case_insensitive_matching",
			vendorErr:  fmt.Errorf("tx_power_out_of_range: power level 50 is invalid"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"requestedPower": 50},
			goldenFile: "case_insensitive_matching.json",
		},
		{
			name:       "mixed_case_matching",
			vendorErr:  fmt.Errorf("Rf_Busy: transmitter is busy"),
			vendorID:   "silvus",
			payload:    map[string]interface{}{"currentOperation": "frequency_change"},
			goldenFile: "mixed_case_matching.json",
		},
		{
			name:       "nil_error",
			vendorErr:  nil,
			vendorID:   "silvus",
			payload:    nil,
			goldenFile: "nil_error.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Normalize the vendor error
			normalizedErr := NormalizeVendorErrorWithVendor(tt.vendorErr, tt.payload, tt.vendorID)

			// Create error envelope
			envelope := createErrorEnvelope(t, normalizedErr, tt.vendorErr, tt.payload, tt.vendorID)

			// Marshal to JSON
			jsonData, err := json.MarshalIndent(envelope, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal error envelope: %v", err)
			}

			goldenPath := filepath.Join("testdata", "adapter", tt.goldenFile)

			if *update {
				// Update golden file
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("Failed to create testdata directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, jsonData, 0644); err != nil {
					t.Fatalf("Failed to write golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", goldenPath)
				return
			}

			// Read golden file
			golden, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
			}

			// Compare responses
			if string(jsonData) != string(golden) {
				t.Errorf("Error envelope doesn't match golden file %s\nExpected:\n%s\nGot:\n%s",
					goldenPath, string(golden), string(jsonData))
			}
		})
	}
}

// ErrorEnvelope represents the structure of a normalized error response
type ErrorEnvelope struct {
	Code        string      `json:"code"`
	Message     string      `json:"message"`
	VendorID    string      `json:"vendorId,omitempty"`
	OriginalErr string      `json:"originalError,omitempty"`
	Payload     interface{} `json:"payload,omitempty"`
	Timestamp   string      `json:"timestamp"`
}

// createErrorEnvelope creates a standardized error envelope for golden testing
func createErrorEnvelope(t *testing.T, normalizedErr, originalErr error, payload interface{}, vendorID string) ErrorEnvelope {
	envelope := ErrorEnvelope{
		VendorID:  vendorID,
		Payload:   payload,
		Timestamp: "2024-01-01T12:00:00Z", // Fixed timestamp for stable comparison
	}

	if normalizedErr == nil {
		envelope.Code = "SUCCESS"
		envelope.Message = "No error"
		return envelope
	}

	// Extract normalized error code
	if vendorErr, ok := normalizedErr.(*VendorError); ok {
		switch {
		case vendorErr.Code == ErrInvalidRange:
			envelope.Code = "INVALID_RANGE"
			envelope.Message = "Parameter value is outside the allowed range"
		case vendorErr.Code == ErrBusy:
			envelope.Code = "BUSY"
			envelope.Message = "Service is busy, please retry with backoff"
		case vendorErr.Code == ErrUnavailable:
			envelope.Code = "UNAVAILABLE"
			envelope.Message = "Service is temporarily unavailable"
		case vendorErr.Code == ErrInternal:
			envelope.Code = "INTERNAL"
			envelope.Message = "Internal server error"
		default:
			envelope.Code = "UNKNOWN"
			envelope.Message = "Unknown error"
		}
	} else {
		envelope.Code = "UNKNOWN"
		envelope.Message = "Unknown error"
	}

	// Add original error if available
	if originalErr != nil {
		envelope.OriginalErr = originalErr.Error()
	}

	return envelope
}

// TestVendorErrorMappingsGolden tests the vendor error mapping tables against golden files
func TestVendorErrorMappingsGolden(t *testing.T) {
	// Test each vendor mapping
	vendors := []string{"silvus", "generic"}

	for _, vendor := range vendors {
		t.Run(vendor, func(t *testing.T) {
			// Get vendor mapping
			mapping, exists := VendorErrorMappings[vendor]
			if !exists {
				t.Fatalf("Vendor mapping for %s not found", vendor)
			}

			// Create mapping structure for golden comparison
			mappingData := map[string]interface{}{
				"vendor":      vendor,
				"range":       mapping.Range,
				"busy":        mapping.Busy,
				"unavailable": mapping.Unavailable,
			}

			// Marshal to JSON
			jsonData, err := json.MarshalIndent(mappingData, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal vendor mapping: %v", err)
			}

			goldenPath := filepath.Join("testdata", "adapter", fmt.Sprintf("vendor_mapping_%s.json", vendor))

			if *update {
				// Update golden file
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("Failed to create testdata directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, jsonData, 0644); err != nil {
					t.Fatalf("Failed to write golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", goldenPath)
				return
			}

			// Read golden file
			golden, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
			}

			// Compare responses
			if string(jsonData) != string(golden) {
				t.Errorf("Vendor mapping doesn't match golden file %s\nExpected:\n%s\nGot:\n%s",
					goldenPath, string(golden), string(jsonData))
			}
		})
	}
}
