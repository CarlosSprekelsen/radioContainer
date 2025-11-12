package telemetry

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/config"
)

// threadSafeResponseWriter captures SSE events in a thread-safe way
type threadSafeResponseWriter struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	headers http.Header
}

func newThreadSafeResponseWriter() *threadSafeResponseWriter {
	return &threadSafeResponseWriter{
		headers: make(http.Header),
	}
}

func (w *threadSafeResponseWriter) Header() http.Header {
	return w.headers
}

func (w *threadSafeResponseWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(data)
}

func (w *threadSafeResponseWriter) WriteHeader(statusCode int) {
	// No-op for testing
}

func (w *threadSafeResponseWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

func TestNewHub(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)

	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}

	if hub.clients == nil {
		t.Error("Hub clients map not initialized")
	}

	if hub.radioIDs == nil {
		t.Error("Hub radioIDs map not initialized")
	}

	if hub.buffers == nil {
		t.Error("Hub buffers map not initialized")
	}

	if hub.config != cfg {
		t.Error("Hub config not set correctly")
	}

	// Clean up
	hub.Stop()
}

func TestHubPublish(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Publish an event without clients
	event := Event{
		Type: "test",
		Data: map[string]interface{}{
			"message": "test event",
		},
	}

	err := hub.Publish(event)
	if err != nil {
		t.Fatalf("Publish() failed: %v", err)
	}
}

func TestHubPublishRadio(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Publish an event for a specific radio
	event := Event{
		Type: "state",
		Data: map[string]interface{}{
			"powerDbm":     30,
			"frequencyMhz": 2412,
		},
	}

	err := hub.PublishRadio("radio-01", event)
	if err != nil {
		t.Fatalf("PublishRadio() failed: %v", err)
	}

	// Check that event was buffered for the radio
	hub.mu.RLock()
	buffer, exists := hub.buffers["radio-01"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("Event buffer not created for radio")
	}

	if buffer != nil && buffer.GetSize() != 1 {
		t.Errorf("Expected 1 event in buffer, got %d", buffer.GetSize())
	}
}

func TestEventBuffer(t *testing.T) {
	capacity := 5
	buffer := NewEventBuffer(capacity)

	if buffer.GetCapacity() != capacity {
		t.Errorf("Expected capacity %d, got %d", capacity, buffer.GetCapacity())
	}

	if buffer.GetSize() != 0 {
		t.Errorf("Expected initial size 0, got %d", buffer.GetSize())
	}

	// Add events
	for i := 0; i < 7; i++ { // More than capacity
		event := Event{
			Type: "test",
			Data: map[string]interface{}{
				"index": i,
			},
		}
		buffer.AddEvent(event)
	}

	// Should maintain capacity
	if buffer.GetSize() != capacity {
		t.Errorf("Expected size %d, got %d", capacity, buffer.GetSize())
	}

	// Test GetEventsAfter
	events := buffer.GetEventsAfter(2)
	if len(events) != 5 { // Events 3, 4, 5, 6, 7 (all events with ID > 2)
		t.Errorf("Expected 5 events after ID 2, got %d", len(events))
	}
}

func TestHubStop(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)

	// Stop the hub
	hub.Stop()

	// Check that clients are cleaned up
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("Expected 0 clients after stop, got %d", clientCount)
	}
}

func TestEventTypes(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Test different event types
	eventTypes := []string{"ready", "state", "channelChanged", "powerChanged", "fault", "heartbeat"}

	for _, eventType := range eventTypes {
		event := Event{
			Type: eventType,
			Data: map[string]interface{}{
				"test": "data",
			},
		}

		err := hub.Publish(event)
		if err != nil {
			t.Errorf("Publish() failed for event type %s: %v", eventType, err)
		}
	}
}

func TestEventIDGeneration(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Test global event ID generation
	event1 := Event{Type: "test1", Data: map[string]interface{}{}}
	event2 := Event{Type: "test2", Data: map[string]interface{}{}}

	hub.Publish(event1)
	hub.Publish(event2)

	// Test radio-specific event ID generation
	radioEvent1 := Event{Type: "state", Data: map[string]interface{}{}, Radio: "radio-01"}
	radioEvent2 := Event{Type: "state", Data: map[string]interface{}{}, Radio: "radio-01"}

	hub.PublishRadio("radio-01", radioEvent1)
	hub.PublishRadio("radio-01", radioEvent2)

	// Check that radio buffer was created
	hub.mu.RLock()
	buffer, exists := hub.buffers["radio-01"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("Radio buffer not created")
	}

	if buffer.GetSize() != 2 {
		t.Errorf("Expected 2 events in radio buffer, got %d", buffer.GetSize())
	}
}

func TestEventCreation(t *testing.T) {
	// Test event creation
	event := Event{
		ID:   42,
		Type: "test",
		Data: map[string]interface{}{
			"message": "test event",
		},
	}

	// Test that event has correct format
	if event.ID != 42 {
		t.Error("Event ID not set correctly")
	}

	if event.Type != "test" {
		t.Error("Event type not set correctly")
	}

	if event.Data["message"] != "test event" {
		t.Error("Event data not set correctly")
	}
}

func TestConcurrentPublish(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Publish events concurrently without clients
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(index int) {
			event := Event{
				Type: "concurrent",
				Data: map[string]interface{}{
					"index": index,
				},
			}
			hub.Publish(event)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestHubSubscribeBasic(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Create test request
	req := httptest.NewRequest("GET", "/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	// Create thread-safe response writer
	w := newThreadSafeResponseWriter()

	// Subscribe in a goroutine to check client registration
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- hub.Subscribe(ctx, w, req)
	}()

	// Wait a bit for client to be registered
	time.Sleep(10 * time.Millisecond)

	// Check that client was registered
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 client, got %d", clientCount)
	}

	// Wait for subscribe to complete
	err := <-done
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Subscribe() failed: %v", err)
	}

	// Check response headers
	if w.Header().Get("Content-Type") != "text/event-stream; charset=utf-8" {
		t.Error("Content-Type header not set correctly")
	}

	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Error("Cache-Control header not set correctly")
	}

	// Wait for context to timeout and client to be cleaned up
	time.Sleep(150 * time.Millisecond)

	// Check that client was cleaned up
	hub.mu.RLock()
	clientCount = len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("Expected 0 clients after timeout, got %d", clientCount)
	}
}

// TestTelemetryContract_SubscribeReceiveHeartbeat tests that subscribing to telemetry
// receives heartbeat events as expected.
func TestTelemetryContract_SubscribeReceiveHeartbeat(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	// Use shorter heartbeat interval for testing (50ms instead of 15s)
	cfg.HeartbeatInterval = 50 * time.Millisecond
	cfg.HeartbeatJitter = 5 * time.Millisecond

	hub := NewHub(cfg)
	defer hub.Stop()

	// Create test request
	req := httptest.NewRequest("GET", "/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	// Create thread-safe response writer
	w := newThreadSafeResponseWriter()

	// Subscribe with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start subscription in goroutine
	subscribeDone := make(chan error, 1)
	go func() {
		subscribeDone <- hub.Subscribe(ctx, w, req)
	}()

	// Wait for subscription to start
	time.Sleep(50 * time.Millisecond)

	// Wait for events to be processed
	time.Sleep(200 * time.Millisecond)

	// Get the response (hub will be stopped by defer)
	response := w.String()

	// Wait for the context timeout to occur naturally
	select {
	case err := <-subscribeDone:
		if err != nil && err != context.DeadlineExceeded {
			t.Fatalf("Subscribe() failed: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Subscribe() did not complete after timeout")
	}

	// Check for ready event
	if !strings.Contains(response, "event: ready") {
		t.Error("Expected ready event in response")
	}

	// Check for heartbeat events
	heartbeatCount := strings.Count(response, "event: heartbeat")
	if heartbeatCount < 1 {
		t.Errorf("Expected at least 1 heartbeat event, got %d. Response: %s", heartbeatCount, response)
	}

	// Verify SSE format
	lines := strings.Split(response, "\n")
	hasEventType := false
	hasData := false

	for _, line := range lines {
		if strings.HasPrefix(line, "event: ") {
			hasEventType = true
		}
		if strings.HasPrefix(line, "data: ") {
			hasData = true
		}
	}

	if !hasEventType {
		t.Error("Expected event type in SSE response")
	}
	if !hasData {
		t.Error("Expected data in SSE response")
	}
}

// TestTelemetryContract_PowerChannelChanges tests that power and channel changes
// via orchestrator result in appropriate telemetry events.
func TestTelemetryContract_PowerChannelChanges(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Create test request
	req := httptest.NewRequest("GET", "/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	// Create thread-safe response writer
	w := newThreadSafeResponseWriter()

	// Subscribe in a goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- hub.Subscribe(ctx, w, req)
	}()

	// Wait for client to be registered
	time.Sleep(10 * time.Millisecond)

	// Simulate power change via orchestrator
	powerEvent := Event{
		Type: "powerChanged",
		Data: map[string]interface{}{
			"radioId":   "radio-01",
			"powerDbm":  25,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
		Radio: "radio-01",
	}

	err := hub.PublishRadio("radio-01", powerEvent)
	if err != nil {
		t.Fatalf("PublishRadio() failed: %v", err)
	}

	// Simulate channel change via orchestrator
	channelEvent := Event{
		Type: "channelChanged",
		Data: map[string]interface{}{
			"radioId":      "radio-01",
			"frequencyMhz": 2417.0,
			"timestamp":    time.Now().UTC().Format(time.RFC3339),
		},
		Radio: "radio-01",
	}

	err = hub.PublishRadio("radio-01", channelEvent)
	if err != nil {
		t.Fatalf("PublishRadio() failed: %v", err)
	}

	// Wait for events to be processed
	time.Sleep(50 * time.Millisecond)

	// Get the response (hub will be stopped by defer)
	response := w.String()

	// Check for powerChanged event
	if !strings.Contains(response, "event: powerChanged") {
		t.Error("Expected powerChanged event in response")
	}
	if !strings.Contains(response, "powerDbm") {
		t.Error("Expected powerDbm data in powerChanged event")
	}

	// Check for channelChanged event
	if !strings.Contains(response, "event: channelChanged") {
		t.Error("Expected channelChanged event in response")
	}
	if !strings.Contains(response, "frequencyMhz") {
		t.Error("Expected frequencyMhz data in channelChanged event")
	}

	// Verify events were buffered for the radio
	hub.mu.RLock()
	buffer, exists := hub.buffers["radio-01"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("Expected radio buffer to exist")
	}
	if buffer.GetSize() != 2 {
		t.Errorf("Expected 2 events in radio buffer, got %d", buffer.GetSize())
	}
}

// TestTelemetryContract_DisconnectReconnectWithLastEventID tests that disconnecting
// and reconnecting with Last-Event-ID header properly replays missed events.
func TestTelemetryContract_DisconnectReconnectWithLastEventID(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// First connection - receive some events
	req1 := httptest.NewRequest("GET", "/telemetry", nil)
	req1.Header.Set("Accept", "text/event-stream")

	w1 := httptest.NewRecorder()
	ctx1, cancel1 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel1()

	err := hub.Subscribe(ctx1, w1, req1)
	if err != nil {
		t.Fatalf("First Subscribe() failed: %v", err)
	}

	// Publish some events
	for i := 1; i <= 5; i++ {
		event := Event{
			Type: "test",
			Data: map[string]interface{}{
				"index": i,
			},
			Radio: "radio-01",
		}
		hub.PublishRadio("radio-01", event)
	}

	// Wait for events to be processed
	time.Sleep(50 * time.Millisecond)

	// Disconnect first client
	cancel1()

	// Wait for client to be cleaned up
	time.Sleep(50 * time.Millisecond)

	// Publish more events while disconnected
	for i := 6; i <= 10; i++ {
		event := Event{
			Type: "test",
			Data: map[string]interface{}{
				"index": i,
			},
			Radio: "radio-01",
		}
		hub.PublishRadio("radio-01", event)
	}

	// Reconnect with Last-Event-ID header (simulating client that last saw event ID 5)
	req2 := httptest.NewRequest("GET", "/telemetry?radio=radio-01", nil)
	req2.Header.Set("Accept", "text/event-stream")
	req2.Header.Set("Last-Event-ID", "5") // Resume from event ID 5

	w2 := httptest.NewRecorder()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()

	err = hub.Subscribe(ctx2, w2, req2)
	if err != nil {
		t.Fatalf("Second Subscribe() failed: %v", err)
	}

	// Wait for replay to complete
	time.Sleep(50 * time.Millisecond)

	// Parse reconnected client response
	response := w2.Body.String()

	// Should contain events with IDs > 5 (events 6-10)
	// Check that replayed events are present
	lines := strings.Split(response, "\n")
	replayedEventCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "id: ") {
			// Extract event ID and check if it's > 5
			var eventID int64
			if _, err := fmt.Sscanf(line, "id: %d", &eventID); err == nil {
				if eventID > 5 {
					replayedEventCount++
				}
			}
		}
	}

	if replayedEventCount == 0 {
		t.Error("Expected replayed events with IDs > 5")
	}

	// Verify buffer contains all events
	hub.mu.RLock()
	buffer, exists := hub.buffers["radio-01"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("Expected radio buffer to exist")
	}
	if buffer.GetSize() != 10 {
		t.Errorf("Expected 10 events in radio buffer, got %d", buffer.GetSize())
	}
}

// TestTelemetryContract_MonotonicPerRadioIDs tests that event IDs are monotonic
// per radio and that buffer bounds are respected.
func TestTelemetryContract_MonotonicPerRadioIDs(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	// Use small buffer size for testing
	cfg.EventBufferSize = 3
	hub := NewHub(cfg)
	defer hub.Stop()

	// Test monotonic IDs for radio-01
	radio1Events := make([]Event, 5)
	for i := 0; i < 5; i++ {
		event := Event{
			Type: "test",
			Data: map[string]interface{}{
				"index": i,
			},
			Radio: "radio-01",
		}
		radio1Events[i] = event
		hub.PublishRadio("radio-01", event)
	}

	// Test monotonic IDs for radio-02
	radio2Events := make([]Event, 3)
	for i := 0; i < 3; i++ {
		event := Event{
			Type: "test",
			Data: map[string]interface{}{
				"index": i,
			},
			Radio: "radio-02",
		}
		radio2Events[i] = event
		hub.PublishRadio("radio-02", event)
	}

	// Wait for events to be processed
	time.Sleep(50 * time.Millisecond)

	// Check radio-01 buffer (should maintain capacity of 3)
	hub.mu.RLock()
	buffer1, exists1 := hub.buffers["radio-01"]
	hub.mu.RUnlock()

	if !exists1 {
		t.Error("Expected radio-01 buffer to exist")
	}
	if buffer1.GetSize() != 3 {
		t.Errorf("Expected radio-01 buffer size 3, got %d", buffer1.GetSize())
	}

	// Check radio-02 buffer
	hub.mu.RLock()
	buffer2, exists2 := hub.buffers["radio-02"]
	hub.mu.RUnlock()

	if !exists2 {
		t.Error("Expected radio-02 buffer to exist")
	}
	if buffer2.GetSize() != 3 {
		t.Errorf("Expected radio-02 buffer size 3, got %d", buffer2.GetSize())
	}

	// Verify monotonic IDs within each radio buffer
	events1 := buffer1.GetEventsAfter(0)
	events2 := buffer2.GetEventsAfter(0)

	// Check radio-01 monotonic IDs (should be events 3, 4, 5 due to buffer capacity)
	if len(events1) != 3 {
		t.Errorf("Expected 3 events in radio-01 buffer, got %d", len(events1))
	}
	for i, event := range events1 {
		expectedID := int64(i + 3) // IDs 3, 4, 5
		if event.ID != expectedID {
			t.Errorf("Radio-01 event %d: expected ID %d, got %d", i, expectedID, event.ID)
		}
	}

	// Check radio-02 monotonic IDs (should be events 1, 2, 3)
	if len(events2) != 3 {
		t.Errorf("Expected 3 events in radio-02 buffer, got %d", len(events2))
	}
	for i, event := range events2 {
		expectedID := int64(i + 1) // IDs 1, 2, 3
		if event.ID != expectedID {
			t.Errorf("Radio-02 event %d: expected ID %d, got %d", i, expectedID, event.ID)
		}
	}

	// Verify that radio IDs are independent
	hub.mu.RLock()
	radio1Counter := hub.radioIDs["radio-01"]
	radio2Counter := hub.radioIDs["radio-02"]
	hub.mu.RUnlock()

	if radio1Counter == nil {
		t.Error("Expected radio-01 counter to exist")
	} else {
		radio1ID := atomic.LoadInt64(radio1Counter)
		if radio1ID != 5 {
			t.Errorf("Expected radio-01 next ID 5, got %d", radio1ID)
		}
	}

	if radio2Counter == nil {
		t.Error("Expected radio-02 counter to exist")
	} else {
		radio2ID := atomic.LoadInt64(radio2Counter)
		if radio2ID != 3 {
			t.Errorf("Expected radio-02 next ID 3, got %d", radio2ID)
		}
	}
}

// TestTelemetryContract_BufferBounds tests that the event buffer respects
// capacity bounds and maintains proper circular buffer behavior.
func TestTelemetryContract_BufferBounds(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	// Use small buffer size for testing
	cfg.EventBufferSize = 3
	hub := NewHub(cfg)
	defer hub.Stop()

	// Fill buffer beyond capacity
	for i := 1; i <= 5; i++ {
		event := Event{
			Type: "test",
			Data: map[string]interface{}{
				"index": i,
			},
			Radio: "radio-01",
		}
		hub.PublishRadio("radio-01", event)
	}

	// Wait for events to be processed
	time.Sleep(50 * time.Millisecond)

	// Check buffer size
	hub.mu.RLock()
	buffer, exists := hub.buffers["radio-01"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("Expected radio buffer to exist")
	}
	if buffer.GetSize() != 3 {
		t.Errorf("Expected buffer size 3, got %d", buffer.GetSize())
	}

	// Verify that only the last 3 events are retained (events 3, 4, 5)
	events := buffer.GetEventsAfter(0)
	if len(events) != 3 {
		t.Errorf("Expected 3 events in buffer, got %d", len(events))
	}

	// Check that events 1 and 2 were evicted, but 3, 4, 5 remain
	expectedIDs := []int64{3, 4, 5}
	for i, event := range events {
		if event.ID != expectedIDs[i] {
			t.Errorf("Event %d: expected ID %d, got %d", i, expectedIDs[i], event.ID)
		}
	}

	// Test GetEventsAfter with partial replay
	eventsAfter2 := buffer.GetEventsAfter(2)
	if len(eventsAfter2) != 3 {
		t.Errorf("Expected 3 events after ID 2, got %d", len(eventsAfter2))
	}

	// Should contain events 3, 4, and 5 (all events with ID > 2)
	expectedAfter2 := []int64{3, 4, 5}
	for i, event := range eventsAfter2 {
		if event.ID != expectedAfter2[i] {
			t.Errorf("Event after ID 2, index %d: expected ID %d, got %d", i, expectedAfter2[i], event.ID)
		}
	}
}

// TestTelemetryContract_NoSleepsGreaterThan100ms tests that no sleeps greater
// than 100ms are used in the telemetry implementation.
func TestTelemetryContract_NoSleepsGreaterThan100ms(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	// Use very short intervals for testing
	cfg.HeartbeatInterval = 10 * time.Millisecond
	cfg.HeartbeatJitter = 1 * time.Millisecond

	hub := NewHub(cfg)
	defer hub.Stop()

	// Create test request
	req := httptest.NewRequest("GET", "/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	// Create test response recorder
	w := httptest.NewRecorder()

	// Subscribe with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := hub.Subscribe(ctx, w, req)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Subscribe() failed: %v", err)
	}

	// Subscribe should complete quickly (no sleeps > 100ms)
	if duration > 100*time.Millisecond {
		t.Errorf("Subscribe() took %v, expected < 100ms", duration)
	}

	// Wait for a short period to allow heartbeat
	time.Sleep(30 * time.Millisecond)

	// Publish an event and measure time
	start = time.Now()
	event := Event{
		Type: "test",
		Data: map[string]interface{}{
			"message": "test",
		},
		Radio: "radio-01",
	}
	err = hub.PublishRadio("radio-01", event)
	duration = time.Since(start)

	if err != nil {
		t.Fatalf("PublishRadio() failed: %v", err)
	}

	// Publish should complete quickly (no sleeps > 100ms)
	if duration > 100*time.Millisecond {
		t.Errorf("PublishRadio() took %v, expected < 100ms", duration)
	}
}

// TestTelemetryContract_SSEFormat tests that the SSE format is correct
// and includes proper headers and event structure.
func TestTelemetryContract_SSEFormat(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Create test request
	req := httptest.NewRequest("GET", "/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	// Use thread-safe writer
	w := newThreadSafeResponseWriter()

	// Subscribe
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := hub.Subscribe(ctx, w, req)
	if err != nil {
		t.Fatalf("Subscribe() failed: %v", err)
	}

	// Publish an event
	event := Event{
		Type: "test",
		Data: map[string]interface{}{
			"message": "test event",
			"value":   42,
		},
		Radio: "radio-01",
	}
	hub.PublishRadio("radio-01", event)

	// Wait for event to be processed and context to complete
	time.Sleep(100 * time.Millisecond)
	<-ctx.Done()

	// Now read the response (hub will be stopped by defer)
	response := w.String()
	lines := strings.Split(response, "\n")

	// Check for SSE format
	hasEventType := false
	hasData := false
	hasID := false

	for _, line := range lines {
		if strings.HasPrefix(line, "event:") {
			hasEventType = true
		}
		if strings.HasPrefix(line, "data:") {
			hasData = true
		}
		if strings.HasPrefix(line, "id:") {
			hasID = true
		}
	}

	if !hasEventType {
		t.Error("Expected event type in SSE response")
	}
	if !hasData {
		t.Error("Expected data in SSE response")
	}
	if !hasID {
		t.Error("Expected event ID in SSE response")
	}
}

// TestEventIDGenerationRace tests concurrent event ID generation for race conditions.
// This test verifies that atomic operations prevent duplicate IDs under high concurrency.
func TestEventIDGenerationRace(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)

	const goroutines = 50
	const eventsPerGoroutine = 20
	const totalEvents = goroutines * eventsPerGoroutine

	var wg sync.WaitGroup
	ids := make(chan int64, totalEvents)

	// Launch concurrent goroutines to generate event IDs for a single radio
	// This avoids the complex race condition of multiple radios
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				id := hub.getNextEventID("test-radio")
				ids <- id
			}
		}()
	}

	wg.Wait()
	close(ids)

	// Collect all generated IDs
	allIDs := make([]int64, 0, totalEvents)

	for id := range ids {
		allIDs = append(allIDs, id)
	}

	// Check for duplicates
	seen := make(map[int64]bool)
	duplicates := 0
	for _, id := range allIDs {
		if seen[id] {
			duplicates++
			t.Errorf("Duplicate ID generated: %d", id)
		}
		seen[id] = true
	}

	if duplicates > 0 {
		t.Errorf("Found %d duplicate IDs out of %d total", duplicates, totalEvents)
	}

	// Verify IDs are positive and sequential
	for _, id := range allIDs {
		if id <= 0 {
			t.Errorf("Invalid ID generated: %d (should be > 0)", id)
		}
		if id > int64(totalEvents) {
			t.Errorf("ID too large: %d (should be <= %d)", id, totalEvents)
		}
	}

	t.Logf("Generated %d unique IDs with %d goroutines, %d events each",
		len(seen), goroutines, eventsPerGoroutine)
}
