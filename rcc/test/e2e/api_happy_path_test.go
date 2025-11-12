// Package e2e provides end-to-end tests for the Radio Control Container API.
// This file implements black-box testing using only HTTP/SSE and contract validation.
package e2e

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/test/harness"
)

func TestE2E_HappyPath(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	// Create test harness with seeded state
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	opts.CorrelationID = "test-001"

	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: Route table and seeded IDs
	t.Logf("=== TEST EVIDENCE ===")
	t.Logf("Server URL: %s", server.URL)
	t.Logf("===================")

	// 1) List radios - should return seeded radio
	resp := httpGetWithStatus(t, server.URL+"/api/v1/radios")
	validator.ValidateHTTPResponse(t, resp, 200)

	body := httpGetJSON(t, server.URL+"/api/v1/radios")
	mustHave(t, body, "result", "ok")

	// Check response structure: data.activeRadioId and data.items
	data := body["data"]
	if data == nil {
		t.Fatal("Expected 'data' field in response")
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be a map, got %T", data)
	}

	// Check activeRadioId
	mustHave(t, dataMap, "activeRadioId", "silvus-001")

	// Check items list
	items, ok := dataMap["items"].([]interface{})
	if !ok || len(items) == 0 {
		t.Fatal("Expected items to be a non-empty list")
	}

	// Check first radio
	radio := items[0].(map[string]interface{})
	mustHave(t, radio, "id", "silvus-001")
	mustHave(t, radio, "model", "Unknown-Radio")

	// 2) Select radio (should already be active)
	httpPostJSON200(t, server.URL+"/api/v1/radios/select", map[string]any{"radioId": "silvus-001"})

	// 3) Set power
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 10.0})

	// 4) Set channel (by index ⇒ frequency mapping)
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/channel", map[string]any{"channelIndex": 6})

	// 5) Read-back checks
	gotP := httpGetJSON(t, server.URL+"/api/v1/radios/silvus-001/power")
	mustHaveNumber(t, gotP, "data.powerDbm", 10.0)

	gotC := httpGetJSON(t, server.URL+"/api/v1/radios/silvus-001/channel")
	mustHaveNumber(t, gotC, "data.frequencyMhz", 2437.0)

	// Audit logs are server-side only per Architecture §8.6; no E2E access

	t.Log("✅ Happy path working correctly")
}

func TestE2E_TelemetryIntegration(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	// Create test harness
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: Seeded state
	t.Logf("=== TEST EVIDENCE ===")
	t.Logf("Server URL: %s", server.URL)
	t.Logf("===================")

	// Subscribe to telemetry using HTTP SSE endpoint
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	// Start telemetry subscription via HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to connect to telemetry: %v", err)
	}
	defer resp.Body.Close()

	// Wait for subscription to start
	time.Sleep(100 * time.Millisecond)

	// Trigger power change
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 25.0})

	// Wait for events
	time.Sleep(200 * time.Millisecond)

	// Read SSE events with a timeout
	eventsChan := make(chan string, 100)
	buf := make([]byte, 1024)
	readDone := make(chan struct{})

	go func() {
		defer close(readDone)
		for {
			n, err := resp.Body.Read(buf)
			if err != nil {
				if err != io.EOF {
					t.Logf("SSE read error: %v", err)
				}
				return
			}
			if n > 0 {
				eventsChan <- string(buf[:n])
			}
		}
	}()

	// Collect events for test duration
	timeout := time.After(1 * time.Second)
	var events []string
	collecting:
	for {
		select {
		case event := <-eventsChan:
			events = append(events, event)
		case <-readDone:
			// Connection closed
			break collecting
		case <-timeout:
			// Timeout - collect events we have so far
			break collecting
		}
	}

	response := strings.Join(events, "")

	// Evidence: SSE events
	t.Logf("=== SSE EVIDENCE ===")
	t.Logf("Received %d events", len(events))
	for i, event := range events {
		t.Logf("Event %d: %s", i+1, strings.TrimSpace(event))
		// Validate each event against contract
		validator.ValidateSSEEvent(t, event)
	}
	t.Logf("===================")

	// Verify telemetry events
	if !strings.Contains(response, "event: ready") {
		t.Error("Expected ready event in telemetry")
	}

	if !strings.Contains(response, "powerChanged") {
		t.Error("Expected powerChanged event in telemetry")
	}

	t.Log("✅ Telemetry integration working correctly")
}
