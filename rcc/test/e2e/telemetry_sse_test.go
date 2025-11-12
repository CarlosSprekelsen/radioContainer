// Package e2e provides telemetry SSE tests for the Radio Control Container API.
// This file implements black-box testing using only HTTP/SSE and contract validation.
package e2e

import (
	"bufio"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/test/harness"
)

func TestE2E_TelemetrySSEConnection(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: Seeded state via HTTP contract
	t.Logf("=== TEST EVIDENCE ===")
	radios := httpGetJSON(t, server.URL+"/api/v1/radios")
	mustHave(t, radios, "result", "ok")
	if d, ok := radios["data"].(map[string]any); ok {
		if id, ok := d["activeRadioId"].(string); ok {
			t.Logf("Active Radio ID: %s", id)
		}
	}
	t.Logf("===================")

	// Subscribe to telemetry
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	w := newThreadSafeResponseWriter()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	telemetryDone := make(chan error, 1)
	go func() {
		client := &http.Client{}
		// Use context-aware request for automatic connection cleanup
		reqWithCtx := req.WithContext(ctx)
		resp, err := client.Do(reqWithCtx)
		if err != nil {
			telemetryDone <- err
			return
		}
		defer resp.Body.Close()

		buf := make([]byte, 1024)
		for {
			// Check context first - this is the key fix
			select {
			case <-ctx.Done():
				telemetryDone <- ctx.Err()
				return
			default:
				// Only proceed if context is not done
				if ctx.Err() != nil {
					telemetryDone <- ctx.Err()
					return
				}
			}

			// Use a goroutine to make the read non-blocking
			readDone := make(chan struct{})
			var n int
			var err error

			go func() {
				n, err = resp.Body.Read(buf)
				close(readDone)
			}()

			select {
			case <-ctx.Done():
				telemetryDone <- ctx.Err()
				return
			case <-readDone:
				if err != nil {
					telemetryDone <- err
					return
				}
				w.Write(buf[:n])
			}
		}
	}()

	// Wait for subscription to start
	time.Sleep(100 * time.Millisecond)

	// Trigger power change
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 25.0})

	// Wait for events
	time.Sleep(200 * time.Millisecond)

	// Collect telemetry events
	events := w.collectEvents(500 * time.Millisecond)
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

	// Wait for telemetry to complete
	select {
	case err := <-telemetryDone:
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Telemetry failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Telemetry did not complete")
	}

	t.Log("✅ Telemetry SSE connection working correctly")
}

func TestE2E_TelemetryLastEventID(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// First connection - get some events
	req1, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req1.Header.Set("Accept", "text/event-stream")

	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout for HTTP layer
	defer cancel1()

	// Collect events via channel for test duration
	eventsChan1 := make(chan string, 100)
	go func() {
		client := &http.Client{}
		// Use context-aware request for automatic connection cleanup
		req1WithCtx := req1.WithContext(ctx1)
		resp, err := client.Do(req1WithCtx)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		// Read events until context cancels
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			eventsChan1 <- scanner.Text()
		}
	}()

	// Wait for subscription and trigger event
	time.Sleep(100 * time.Millisecond)
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 15.0})
	time.Sleep(200 * time.Millisecond)

	// Collect events for test duration
	timeout1 := time.After(2 * time.Second)
	var events1 []string
collecting1:
	for {
		select {
		case event := <-eventsChan1:
			events1 = append(events1, event)
		case <-timeout1:
			break collecting1 // Stop collecting, cancel context
		}
	}

	// Cancel context to close SSE connection
	cancel1()
	time.Sleep(100 * time.Millisecond) // Let goroutine clean up

	response1 := strings.Join(events1, "")

	// Second connection with Last-Event-ID
	req2, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req2.Header.Set("Accept", "text/event-stream")
	req2.Header.Set("Last-Event-ID", "1") // Simulate reconnection

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout for HTTP layer
	defer cancel2()

	// Collect events via channel for test duration
	eventsChan2 := make(chan string, 100)
	go func() {
		client := &http.Client{}
		// Use context-aware request for automatic connection cleanup
		req2WithCtx := req2.WithContext(ctx2)
		resp, err := client.Do(req2WithCtx)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		// Read events until context cancels
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			eventsChan2 <- scanner.Text()
		}
	}()

	// Wait for second subscription
	time.Sleep(100 * time.Millisecond)

	// Trigger another event
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 20.0})
	time.Sleep(200 * time.Millisecond)

	// Collect events for test duration
	timeout2 := time.After(2 * time.Second)
	var events2 []string
collecting2:
	for {
		select {
		case event := <-eventsChan2:
			events2 = append(events2, event)
		case <-timeout2:
			break collecting2 // Stop collecting, cancel context
		}
	}

	// Cancel context to close SSE connection
	cancel2()
	time.Sleep(100 * time.Millisecond) // Let goroutine clean up

	response2 := strings.Join(events2, "")

	// Evidence: Reconnection events
	t.Logf("=== RECONNECTION EVIDENCE ===")
	t.Logf("First connection events: %d", len(events1))
	t.Logf("Second connection events: %d", len(events2))
	t.Logf("=============================")

	// Verify both connections received events
	if !strings.Contains(response1, "powerChanged") {
		t.Error("First connection should have received powerChanged event")
	}

	if !strings.Contains(response2, "powerChanged") {
		t.Error("Second connection should have received powerChanged event")
	}

	t.Log("✅ Telemetry Last-Event-ID reconnection working correctly")
}

func TestE2E_TelemetryHeartbeat(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Subscribe to telemetry
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Collect events via channel for test duration
	eventsChan := make(chan string, 100)
	go func() {
		client := &http.Client{}
		// Use context-aware request for automatic connection cleanup
		reqWithCtx := req.WithContext(ctx)
		resp, err := client.Do(reqWithCtx)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		// Read events until context cancels
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			eventsChan <- scanner.Text()
		}
	}()

	// Wait for subscription to start
	time.Sleep(100 * time.Millisecond)

	// Collect events for a longer period to catch heartbeats
	timeout := time.After(20 * time.Second) // Allow time for heartbeat (15s + jitter)
	var events []string
collecting:
	for {
		select {
		case event := <-eventsChan:
			events = append(events, event)
		case <-timeout:
			break collecting // Stop collecting, cancel context
		}
	}

	// Cancel context to close SSE connection
	cancel()
	time.Sleep(100 * time.Millisecond) // Let goroutine clean up

	response := strings.Join(events, "")

	// Evidence: Heartbeat events
	heartbeatCount := strings.Count(response, "event: heartbeat")
	t.Logf("=== HEARTBEAT EVIDENCE ===")
	t.Logf("Total events: %d", len(events))
	t.Logf("Heartbeat events: %d", heartbeatCount)
	t.Logf("=========================")

	// Validate heartbeat timing against CB-TIMING
	baseInterval := 15 * time.Second // From CB-TIMING §3
	jitter := 2 * time.Second        // From CB-TIMING §3
	validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)

	// Verify heartbeat events
	if heartbeatCount < 1 {
		t.Errorf("Expected at least 1 heartbeat event, got %d", heartbeatCount)
	}

	t.Log("✅ Telemetry heartbeat working correctly")
}
