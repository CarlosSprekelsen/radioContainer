//
//
package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/radio-control/rcc/internal/config"
)

// Event represents a telemetry event with SSE formatting.
type Event struct {
	ID    int64                  `json:"id,omitempty"`
	Type  string                 `json:"type"`
	Data  map[string]interface{} `json:"data"`
	Radio string                 `json:"radio,omitempty"`
}

// Client represents an SSE client connection.
type Client struct {
	ID      string
	Writer  http.ResponseWriter
	Request *http.Request
	Context context.Context
	Cancel  context.CancelFunc
	LastID  int64
	Radio   string
	Events  chan Event
	once    sync.Once
	mu      sync.Mutex // Protect Writer access
}

// Hub manages SSE telemetry distribution with per-radio buffering.
//
// LOCK ORDERING (if multiple locks are ever used):
// 1. h.mu (Hub's RWMutex) - protects clients, radioIDs, buffers maps
// 2. EventBuffer.mu (per-buffer mutex) - protects individual buffer state
// 3. Client.once (sync.Once) - ensures single channel close
//
// Current implementation uses only h.mu for Hub-level synchronization.
// EventBuffer has its own mutex for internal synchronization.
// Client channels use sync.Once for thread-safe closing.
type Hub struct {
	mu       sync.RWMutex
	clients  map[string]*Client
	radioIDs map[string]*int64 // Monotonic event IDs per radio (atomic counters)

	// Per-radio event buffers
	buffers map[string]*EventBuffer

	// Configuration
	config *config.TimingConfig

	// Heartbeat ticker
	heartbeatTicker *time.Ticker
	stopHeartbeat   chan bool

	// Synchronization for shutdown
	done chan struct{}
	wg   sync.WaitGroup
}

// EventBuffer maintains a circular buffer of events for a specific radio.
type EventBuffer struct {
	mu       sync.RWMutex
	events   []Event
	capacity int
	nextID   int64
	created  time.Time
}

// NewHub creates a new telemetry hub with the specified configuration.
func NewHub(timingConfig *config.TimingConfig) *Hub {
	hub := &Hub{
		clients:  make(map[string]*Client),
		radioIDs: make(map[string]*int64),
		buffers:  make(map[string]*EventBuffer),
		config:   timingConfig,
		done:     make(chan struct{}),
	}

	return hub
}

// Subscribe handles SSE client subscription with Last-Event-ID resume support.
func (h *Hub) Subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create client context
	clientCtx, cancel := context.WithCancel(ctx)

	// Generate client ID
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())

	// Parse Last-Event-ID header for resume
	lastEventID := int64(0)
	if lastIDStr := r.Header.Get("Last-Event-ID"); lastIDStr != "" {
		if id, err := strconv.ParseInt(lastIDStr, 10, 64); err == nil {
			lastEventID = id
		}
	}

	// Extract radio ID from query parameter
	radioID := r.URL.Query().Get("radio")

	// Create client
	client := &Client{
		ID:      clientID,
		Writer:  w,
		Request: r,
		Context: clientCtx,
		Cancel:  cancel,
		LastID:  lastEventID,
		Radio:   radioID,
		Events:  make(chan Event, 100), // Buffer for client events
	}

	// Register client
	h.mu.Lock()
	h.clients[clientID] = client
	h.mu.Unlock()

	// Send initial ready event
	if err := h.sendReadyEvent(client); err != nil {
		h.unregisterClient(clientID)
		return fmt.Errorf("failed to send ready event: %w", err)
	}

	// Replay buffered events if Last-Event-ID provided
	if lastEventID > 0 {
		if err := h.replayEvents(client, lastEventID); err != nil {
			h.unregisterClient(clientID)
			return fmt.Errorf("failed to replay events: %w", err)
		}
	}

	// Start heartbeat if this is the first client
	h.mu.Lock()
	if len(h.clients) == 1 && h.heartbeatTicker == nil {
		h.startHeartbeat()
	}
	h.mu.Unlock()

	// Handle client events (blocks until client disconnects)
	h.handleClient(client)

	return nil
}

// Publish publishes an event to all connected clients.
func (h *Hub) Publish(event Event) error {
	// Assign event ID if not set (needs write lock)
	if event.ID == 0 {
		event.ID = h.getNextEventID(event.Radio)
	}

	// Buffer the event (needs write lock)
	if event.Radio != "" {
		h.bufferEvent(event)
	}

	// Send to all clients (needs read lock)
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	// Send to all clients without holding the lock
	for _, client := range clients {
		select {
		case <-client.Context.Done():
			// Client context cancelled, skip this client - PRIORITY
			continue
		case <-h.done:
			// Hub is shutting down, don't send
			return nil
		case client.Events <- event:
		case <-time.After(100 * time.Millisecond):
			// Drop event if client is slow to prevent blocking
		}
	}

	return nil
}

// PublishRadio publishes an event for a specific radio.
func (h *Hub) PublishRadio(radioID string, event Event) error {
	event.Radio = radioID
	return h.Publish(event)
}

// sendReadyEvent sends the initial ready event to a client.
func (h *Hub) sendReadyEvent(client *Client) error {
	readyEvent := Event{
		ID:   h.getNextEventID(client.Radio),
		Type: "ready",
		Data: map[string]interface{}{
			"snapshot": map[string]interface{}{
				"activeRadioId": "",              // TODO: Get from radio manager
				"radios":        []interface{}{}, // TODO: Get from radio manager
			},
		},
	}

	return h.sendEventToClient(client, readyEvent)
}

// replayEvents replays buffered events for a client based on Last-Event-ID.
func (h *Hub) replayEvents(client *Client, lastEventID int64) error {
	h.mu.RLock()
	buffer, exists := h.buffers[client.Radio]
	h.mu.RUnlock()

	if !exists {
		return nil // No buffer for this radio
	}

	// Get events after the last event ID
	events := buffer.GetEventsAfter(lastEventID)

	// Send replayed events
	for _, event := range events {
		if err := h.sendEventToClient(client, event); err != nil {
			return err
		}
	}

	return nil
}

// sendEventToClient sends a single event to a client via SSE.
func (h *Hub) sendEventToClient(client *Client, event Event) error {
	// Protect Writer access with mutex to prevent race conditions
	client.mu.Lock()
	defer client.mu.Unlock()

	// Format as SSE
	if event.ID > 0 {
		if _, err := fmt.Fprintf(client.Writer, "id: %d\n", event.ID); err != nil {
			return fmt.Errorf("failed to write event ID: %w", err)
		}
	}
	if _, err := fmt.Fprintf(client.Writer, "event: %s\n", event.Type); err != nil {
		return fmt.Errorf("failed to write event type: %w", err)
	}

	// Serialize data as JSON
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	if _, err := fmt.Fprintf(client.Writer, "data: %s\n\n", string(data)); err != nil {
		return fmt.Errorf("failed to write event data: %w", err)
	}

	// Flush the response immediately
	if flusher, ok := client.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// handleClient manages a client connection and event delivery.
func (h *Hub) handleClient(client *Client) {
	defer func() {
		// Close the client's event channel when the handler exits
		// Use sync.Once to ensure the channel is only closed once
		client.once.Do(func() {
			close(client.Events)
		})
		h.unregisterClient(client.ID)
	}()

	for {
		// ✅ CHECK CONTEXT FIRST - before select
		select {
		case <-client.Context.Done():
			return // Immediate exit
		default:
		}

		// Then normal select with short timeout
		timeout := time.NewTimer(100 * time.Millisecond) // ✅ Shorter!
		select {
		case <-client.Context.Done():
			timeout.Stop()
			return
		case <-timeout.C:
			// Loop continues, rechecks context
			continue
		case event, ok := <-client.Events:
			timeout.Stop()
			if !ok {
				return
			}
			if err := h.sendEventToClient(client, event); err != nil {
				return
			}
		}
	}
}

// unregisterClient removes a client from the hub.
func (h *Hub) unregisterClient(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, exists := h.clients[clientID]; exists {
		client.Cancel()
		// Don't close the channel here to avoid race with heartbeat
		// The channel will be closed when the client goroutine exits
		delete(h.clients, clientID)

		// Stop heartbeat if no clients remain
		if len(h.clients) == 0 && h.heartbeatTicker != nil {
			h.heartbeatTicker.Stop()
			h.heartbeatTicker = nil
			if h.stopHeartbeat != nil {
				close(h.stopHeartbeat)
				h.stopHeartbeat = nil
			}
		}
	}
}

// getNextEventID returns the next monotonic event ID for a radio.
func (h *Hub) getNextEventID(radioID string) int64 {
	if radioID == "" {
		radioID = "global"
	}

	// Try to get existing counter with read lock
	h.mu.RLock()
	counter, exists := h.radioIDs[radioID]
	h.mu.RUnlock()

	if exists {
		// Use atomic operation for fast path
		return atomic.AddInt64(counter, 1)
	}

	// Create new counter with write lock
	h.mu.Lock()
	// Double-check pattern: another goroutine might have created it
	counter, exists = h.radioIDs[radioID]
	if !exists {
		var initial int64 = 0
		counter = &initial
		h.radioIDs[radioID] = counter
	}
	h.mu.Unlock()

	// Use atomic operation
	return atomic.AddInt64(counter, 1)
}

// bufferEvent adds an event to the per-radio buffer.
//
// SAFETY ASSUMPTION: EventBuffer references are never removed from h.buffers map.
// This allows safe access to the buffer reference after releasing h.mu, since
// the EventBuffer.AddEvent() method has its own internal synchronization.
func (h *Hub) bufferEvent(event Event) {
	if event.Radio == "" {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	buffer, exists := h.buffers[event.Radio]
	if !exists {
		buffer = NewEventBuffer(h.config.EventBufferSize)
		h.buffers[event.Radio] = buffer
	}

	buffer.AddEvent(event)
}

// startHeartbeat starts the heartbeat ticker.
func (h *Hub) startHeartbeat() {
	// Caller must hold h.mu and verify h.heartbeatTicker == nil

	interval := h.config.HeartbeatInterval
	jitter := h.config.HeartbeatJitter

	// Add jitter to prevent thundering herd
	actualInterval := interval + time.Duration(float64(jitter)*0.5)

	h.heartbeatTicker = time.NewTicker(actualInterval)
	h.stopHeartbeat = make(chan bool)

	// Store references to avoid race conditions
	ticker := h.heartbeatTicker
	stopChan := h.stopHeartbeat

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		defer func() {
			// Use mutex to safely access heartbeat ticker
			h.mu.Lock()
			if h.heartbeatTicker != nil {
				h.heartbeatTicker.Stop()
			}
			h.mu.Unlock()
		}()

		for {
			select {
			case <-ticker.C:
				h.sendHeartbeat()
			case <-stopChan:
				return
			case <-h.done:
				return
			}
		}
	}()
}

// sendHeartbeat sends a heartbeat event to all clients.
func (h *Hub) sendHeartbeat() {
	heartbeatEvent := Event{
		Type: "heartbeat",
		Data: map[string]interface{}{
			"ts": time.Now().UTC().Format(time.RFC3339),
		},
	}

	h.Publish(heartbeatEvent)
}

// Stop stops the telemetry hub and cleans up resources.
func (h *Hub) Stop() {
	// Signal shutdown first
	close(h.done)

	// Force cancel all client contexts immediately
	h.mu.Lock()
	for _, client := range h.clients {
		client.Cancel()
	}
	h.mu.Unlock()

	// Stop heartbeat ticker
	h.mu.Lock()
	if h.heartbeatTicker != nil {
		h.heartbeatTicker.Stop()
		h.heartbeatTicker = nil
	}
	if h.stopHeartbeat != nil {
		close(h.stopHeartbeat)
		h.stopHeartbeat = nil
	}
	h.mu.Unlock()

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
	case <-time.After(5 * time.Second):
		// Force cleanup after timeout - goroutines may be stuck
	}

	// Close all client connections
	h.mu.Lock()
	for _, client := range h.clients {
		client.Cancel()
		// Use sync.Once to ensure the channel is only closed once
		client.once.Do(func() {
			close(client.Events)
		})
	}
	h.clients = make(map[string]*Client)
	h.mu.Unlock()
}

// NewEventBuffer creates a new event buffer with the specified capacity.
func NewEventBuffer(capacity int) *EventBuffer {
	return &EventBuffer{
		events:   make([]Event, 0, capacity),
		capacity: capacity,
		nextID:   1,
		created:  time.Now(),
	}
}

// AddEvent adds an event to the buffer.
func (b *EventBuffer) AddEvent(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Assign ID if not set
	if event.ID == 0 {
		event.ID = b.nextID
		b.nextID++
	}

	// Add to buffer
	b.events = append(b.events, event)

	// Maintain capacity
	if len(b.events) > b.capacity {
		b.events = b.events[1:]
	}
}

// GetEventsAfter returns events after the specified ID.
func (b *EventBuffer) GetEventsAfter(lastID int64) []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result []Event
	for _, event := range b.events {
		if event.ID > lastID {
			result = append(result, event)
		}
	}

	return result
}

// GetCapacity returns the buffer capacity.
func (b *EventBuffer) GetCapacity() int {
	return b.capacity
}

// GetSize returns the current buffer size.
func (b *EventBuffer) GetSize() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.events)
}
