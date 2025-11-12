package api

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeOrchestrator is a test-only fake orchestrator for unit tests
type fakeOrchestrator struct {
	setChannelError        error
	setChannelByIndexError error

	// Record calls for assertions
	lastSetChannelRadioID        string
	lastSetChannelFrequency      float64
	lastSetChannelByIndexRadioID string
	lastSetChannelByIndexIndex   int
}

func (f *fakeOrchestrator) SetChannel(ctx context.Context, radioID string, frequencyMhz float64) error {
	f.lastSetChannelRadioID = radioID
	f.lastSetChannelFrequency = frequencyMhz
	return f.setChannelError
}

func (f *fakeOrchestrator) SetChannelByIndex(ctx context.Context, radioID string, channelIndex int, radioManager interface{}) error {
	f.lastSetChannelByIndexRadioID = radioID
	f.lastSetChannelByIndexIndex = channelIndex
	return f.setChannelByIndexError
}

// TestHandleSetChannel_Unit tests SetChannel handler error mapping
func TestHandleSetChannel_Unit(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedResult string
		expectedCode   string
	}{
		{
			name:           "no_parameters",
			requestBody:    `{}`,
			expectedStatus: 400,
			expectedResult: "error",
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "invalid_json",
			requestBody:    `{"channelIndex": 6,}`,
			expectedStatus: 400,
			expectedResult: "error",
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "empty_body",
			requestBody:    ``,
			expectedStatus: 400,
			expectedResult: "error",
			expectedCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create server with nil orchestrator to test error path
			server := &Server{
				orchestrator: nil,
			}

			// Create request
			req := httptest.NewRequest("POST", "/api/v1/radios/test-radio-1/channel", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call handler
			server.handleSetChannel(w, req, "test-radio-1")

			// Assert status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Parse and assert response envelope
			var response Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Result != tt.expectedResult {
				t.Errorf("Expected result '%s', got '%s'", tt.expectedResult, response.Result)
			}

			if tt.expectedCode != "" && response.Code != tt.expectedCode {
				t.Errorf("Expected code '%s', got '%s'", tt.expectedCode, response.Code)
			}

			// Assert envelope fields exist
			if response.CorrelationID == "" {
				t.Error("Expected non-empty correlationId")
			}

			// Assert Content-Type
			if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
				t.Errorf("Expected Content-Type 'application/json; charset=utf-8', got '%s'", w.Header().Get("Content-Type"))
			}
		})
	}
}
