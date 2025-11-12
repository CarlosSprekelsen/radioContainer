package adapter

import (
	"errors"
	"fmt"
	"testing"
)

func TestNormalizeVendorError(t *testing.T) {
	tests := []struct {
		name          string
		vendorErr     error
		vendorPayload interface{}
		expectedCode  error
		expectedMsg   string
	}{
		{
			name:          "nil error returns nil",
			vendorErr:     nil,
			vendorPayload: nil,
			expectedCode:  nil,
			expectedMsg:   "",
		},
		{
			name:          "unknown error maps to INTERNAL",
			vendorErr:     errors.New("UNKNOWN_ERROR"),
			vendorPayload: map[string]interface{}{"details": "test"},
			expectedCode:  ErrInternal,
			expectedMsg:   "INTERNAL (vendor: UNKNOWN_ERROR)",
		},
		{
			name:          "generic range error maps to INVALID_RANGE",
			vendorErr:     errors.New("OUT_OF_RANGE"),
			vendorPayload: nil,
			expectedCode:  ErrInvalidRange,
			expectedMsg:   "INVALID_RANGE (vendor: OUT_OF_RANGE)",
		},
		{
			name:          "generic busy error maps to BUSY",
			vendorErr:     errors.New("BUSY"),
			vendorPayload: nil,
			expectedCode:  ErrBusy,
			expectedMsg:   "BUSY (vendor: BUSY)",
		},
		{
			name:          "generic unavailable error maps to UNAVAILABLE",
			vendorErr:     errors.New("UNAVAILABLE"),
			vendorPayload: nil,
			expectedCode:  ErrUnavailable,
			expectedMsg:   "UNAVAILABLE (vendor: UNAVAILABLE)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVendorError(tt.vendorErr, tt.vendorPayload)

			if tt.expectedCode == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			vendorErr, ok := result.(*VendorError)
			if !ok {
				t.Fatalf("Expected VendorError, got %T", result)
			}

			if vendorErr.Code != tt.expectedCode {
				t.Errorf("Expected code %v, got %v", tt.expectedCode, vendorErr.Code)
			}

			if vendorErr.Error() != tt.expectedMsg {
				t.Errorf("Expected message %q, got %q", tt.expectedMsg, vendorErr.Error())
			}

			// Compare payloads by string representation to avoid map comparison issues
			expectedStr := ""
			if tt.vendorPayload != nil {
				expectedStr = fmt.Sprintf("%v", tt.vendorPayload)
			}
			actualStr := ""
			if vendorErr.Details != nil {
				actualStr = fmt.Sprintf("%v", vendorErr.Details)
			}
			if expectedStr != actualStr {
				t.Errorf("Expected payload %q, got %q", expectedStr, actualStr)
			}
		})
	}
}

func TestNormalizeVendorErrorWithVendor(t *testing.T) {
	tests := []struct {
		name          string
		vendorErr     error
		vendorPayload interface{}
		vendorID      string
		expectedCode  error
		expectedMsg   string
	}{
		{
			name:          "silvus range error maps to INVALID_RANGE",
			vendorErr:     errors.New("TX_POWER_OUT_OF_RANGE"),
			vendorPayload: map[string]interface{}{"power": 50.0},
			vendorID:      "silvus",
			expectedCode:  ErrInvalidRange,
			expectedMsg:   "INVALID_RANGE (vendor: TX_POWER_OUT_OF_RANGE)",
		},
		{
			name:          "silvus busy error maps to BUSY",
			vendorErr:     errors.New("RF_BUSY"),
			vendorPayload: nil,
			vendorID:      "silvus",
			expectedCode:  ErrBusy,
			expectedMsg:   "BUSY (vendor: RF_BUSY)",
		},
		{
			name:          "silvus unavailable error maps to UNAVAILABLE",
			vendorErr:     errors.New("RADIO_OFFLINE"),
			vendorPayload: nil,
			vendorID:      "silvus",
			expectedCode:  ErrUnavailable,
			expectedMsg:   "UNAVAILABLE (vendor: RADIO_OFFLINE)",
		},
		{
			name:          "unknown vendor falls back to generic mapping",
			vendorErr:     errors.New("OUT_OF_RANGE"),
			vendorPayload: nil,
			vendorID:      "unknown_vendor",
			expectedCode:  ErrInvalidRange,
			expectedMsg:   "INVALID_RANGE (vendor: OUT_OF_RANGE)",
		},
		{
			name:          "silvus unknown error maps to INTERNAL",
			vendorErr:     errors.New("SILVUS_UNKNOWN_ERROR"),
			vendorPayload: nil,
			vendorID:      "silvus",
			expectedCode:  ErrInternal,
			expectedMsg:   "INTERNAL (vendor: SILVUS_UNKNOWN_ERROR)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVendorErrorWithVendor(tt.vendorErr, tt.vendorPayload, tt.vendorID)

			vendorErr, ok := result.(*VendorError)
			if !ok {
				t.Fatalf("Expected VendorError, got %T", result)
			}

			if vendorErr.Code != tt.expectedCode {
				t.Errorf("Expected code %v, got %v", tt.expectedCode, vendorErr.Code)
			}

			if vendorErr.Error() != tt.expectedMsg {
				t.Errorf("Expected message %q, got %q", tt.expectedMsg, vendorErr.Error())
			}

			// Compare payloads by string representation to avoid map comparison issues
			expectedStr := ""
			if tt.vendorPayload != nil {
				expectedStr = fmt.Sprintf("%v", tt.vendorPayload)
			}
			actualStr := ""
			if vendorErr.Details != nil {
				actualStr = fmt.Sprintf("%v", vendorErr.Details)
			}
			if expectedStr != actualStr {
				t.Errorf("Expected payload %q, got %q", expectedStr, actualStr)
			}
		})
	}
}

func TestVendorErrorUnwrap(t *testing.T) {
	originalErr := errors.New("ORIGINAL_ERROR")
	vendorErr := &VendorError{
		Code:     ErrInvalidRange,
		Original: originalErr,
		Details:  map[string]interface{}{"test": true},
	}

	unwrapped := vendorErr.Unwrap()
	if unwrapped != ErrInvalidRange {
		t.Errorf("Expected unwrapped error %v, got %v", ErrInvalidRange, unwrapped)
	}
}

func TestVendorErrorMappings(t *testing.T) {
	// Test that all vendor mappings are properly configured
	expectedVendors := []string{"silvus", "generic"}
	for _, vendor := range expectedVendors {
		if _, exists := VendorErrorMappings[vendor]; !exists {
			t.Errorf("Expected vendor mapping for %s to exist", vendor)
		}
	}

	// Test that silvus mapping has expected tokens
	silvusMap := VendorErrorMappings["silvus"]
	expectedRangeTokens := []string{"TX_POWER_OUT_OF_RANGE", "FREQUENCY_OUT_OF_RANGE"}
	for _, token := range expectedRangeTokens {
		found := false
		for _, rangeToken := range silvusMap.Range {
			if rangeToken == token {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected silvus range token %s not found", token)
		}
	}

	// Test that generic mapping has expected tokens
	genericMap := VendorErrorMappings["generic"]
	expectedGenericTokens := []string{"OUT_OF_RANGE", "BUSY", "UNAVAILABLE"}
	for _, token := range expectedGenericTokens {
		found := false
		for _, rangeToken := range genericMap.Range {
			if rangeToken == token {
				found = true
				break
			}
		}
		for _, busyToken := range genericMap.Busy {
			if busyToken == token {
				found = true
				break
			}
		}
		for _, unavailableToken := range genericMap.Unavailable {
			if unavailableToken == token {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected generic token %s not found", token)
		}
	}
}
