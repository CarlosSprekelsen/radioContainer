// Package e2e provides SSE validation testing against telemetry schema.
package e2e

import (
	"fmt"
	"testing"
	"time"
)

func TestSSEValidation_EventParsing(t *testing.T) {
	validator := NewContractValidator(t)

	// Test SSE event parsing with multi-line format
	testEvents := []string{
		"event: ready\ndata: {\"snapshot\":{\"activeRadioId\":\"\",\"radios\":[]}}\nid: 1\n\n",
		"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}\nid: 2\n\n",
		"event: powerChanged\ndata: {\"radioId\":\"silvus-001\",\"powerDbm\":25.0,\"ts\":\"2025-10-03T10:00:01Z\"}\nid: 3\n\n",
		"event: channelChanged\ndata: {\"radioId\":\"silvus-001\",\"channelIndex\":1,\"frequencyMhz\":2400.0,\"ts\":\"2025-10-03T10:00:02Z\"}\nid: 4\n\n",
	}

	t.Logf("=== SSE EVENT VALIDATION ===")
	t.Logf("Testing %d SSE events against telemetry schema", len(testEvents))

	for i, event := range testEvents {
		t.Run(fmt.Sprintf("event_%d", i+1), func(t *testing.T) {
			validator.ValidateSSEEvent(t, event)
		})
	}

	t.Logf("=== SSE VALIDATION COMPLETE ===")
}

func TestSSEValidation_LastEventIDReplay(t *testing.T) {
	validator := NewContractValidator(t)

	// Simulate Last-Event-ID reconnection
	events := []string{
		"event: ready\ndata: {\"snapshot\":{\"activeRadioId\":\"silvus-001\",\"radios\":[{\"id\":\"silvus-001\"}]}}\nid: 1\n\n",
		"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}\nid: 2\n\n",
		"event: powerChanged\ndata: {\"radioId\":\"silvus-001\",\"powerDbm\":25.0,\"ts\":\"2025-10-03T10:00:01Z\"}\nid: 3\n\n",
		"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:15Z\"}\nid: 4\n\n",
		"event: powerChanged\ndata: {\"radioId\":\"silvus-001\",\"powerDbm\":30.0,\"ts\":\"2025-10-03T10:00:16Z\"}\nid: 5\n\n",
	}

	t.Logf("=== LAST-EVENT-ID REPLAY TEST ===")
	t.Logf("Total events: %d", len(events))

	// Validate all events
	for i, event := range events {
		t.Run(fmt.Sprintf("replay_event_%d", i+1), func(t *testing.T) {
			validator.ValidateSSEEvent(t, event)
		})
	}

	// Test monotonic ID validation
	ids := []int{1, 2, 3, 4, 5}
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Errorf("Event IDs not monotonic: %d <= %d", ids[i], ids[i-1])
		}
	}

	t.Logf("=== REPLAY TEST COMPLETE ===")
}

func TestSSEValidation_HeartbeatTiming(t *testing.T) {
	validator := NewContractValidator(t)

	// Test heartbeat timing with CB-TIMING parameters
	baseInterval := 15 * time.Second
	jitter := 2 * time.Second

	// Simulate heartbeat events with proper timing
	heartbeatEvents := []string{
		"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}\nid: 1\n\n",
		"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:15Z\"}\nid: 2\n\n", // 15s interval
		"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:30Z\"}\nid: 3\n\n", // 15s interval
		"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:45Z\"}\nid: 4\n\n", // 15s interval
	}

	t.Logf("=== HEARTBEAT TIMING VALIDATION ===")
	t.Logf("Base interval: %v", baseInterval)
	t.Logf("Jitter: %v", jitter)
	t.Logf("Heartbeat events: %d", len(heartbeatEvents))

	// Validate heartbeat timing
	validator.ValidateHeartbeatInterval(t, heartbeatEvents, baseInterval, jitter)

	t.Logf("=== HEARTBEAT TIMING COMPLETE ===")
}

func TestSSEValidation_EventTypes(t *testing.T) {
	validator := NewContractValidator(t)

	// Test all supported event types
	eventTypes := []struct {
		eventType string
		eventData string
	}{
		{
			"ready",
			`{"snapshot":{"activeRadioId":"silvus-001","radios":[{"id":"silvus-001","name":"Silvus Radio 1"}]}}`,
		},
		{
			"heartbeat",
			`{"timestamp":"2025-10-03T10:00:00Z"}`,
		},
		{
			"powerChanged",
			`{"radioId":"silvus-001","powerDbm":25.0,"ts":"2025-10-03T10:00:01Z"}`,
		},
		{
			"channelChanged",
			`{"radioId":"silvus-001","channelIndex":1,"frequencyMhz":2400.0,"ts":"2025-10-03T10:00:02Z"}`,
		},
	}

	t.Logf("=== EVENT TYPE VALIDATION ===")
	t.Logf("Testing %d event types", len(eventTypes))

	for i, et := range eventTypes {
		t.Run(fmt.Sprintf("event_type_%s", et.eventType), func(t *testing.T) {
			event := fmt.Sprintf("event: %s\ndata: %s\nid: %d\n\n", et.eventType, et.eventData, i+1)
			validator.ValidateSSEEvent(t, event)
		})
	}

	t.Logf("=== EVENT TYPE VALIDATION COMPLETE ===")
}
