//go:build integration

package fixtures

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/radio-control/rcc/internal/telemetry"
)

// EventData represents the data portion of a telemetry event.
type EventData struct {
	RadioID       string                 `json:"radioId,omitempty"`
	PowerDbm      float64                `json:"powerDbm,omitempty"`
	FrequencyMhz  float64                `json:"frequencyMhz,omitempty"`
	ChannelIndex  int                    `json:"channelIndex,omitempty"`
	Code          string                 `json:"code,omitempty"`
	Message       string                 `json:"message,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	Timestamp     string                 `json:"ts,omitempty"`
	Snapshot      map[string]interface{} `json:"snapshot,omitempty"`
	Radios        []interface{}          `json:"radios,omitempty"`
	ActiveRadioID string                 `json:"activeRadioId,omitempty"`
}

// CollectedEvent represents a telemetry event collected during testing.
type CollectedEvent struct {
	ID       int64      `json:"id"`
	Type     string     `json:"type"`
	Radio    string     `json:"radio,omitempty"`
	Data     EventData  `json:"data"`
	Received time.Time  `json:"received"`
}

// EventCollector collects telemetry events in-process for testing.
type EventCollector struct {
	mu     sync.RWMutex
	events []CollectedEvent
	ch     chan CollectedEvent
	closed bool
}

// NewEventCollector creates a new event collector.
func NewEventCollector(bufferSize int) *EventCollector {
	return &EventCollector{
		events: make([]CollectedEvent, 0),
		ch:     make(chan CollectedEvent, bufferSize),
	}
}

// Handler returns a function that can be used to subscribe to telemetry events.
func (c *EventCollector) Handler() func(event telemetry.Event) {
	return func(event telemetry.Event) {
		c.mu.Lock()
		defer c.mu.Unlock()
		
		if c.closed {
			return
		}

		collected := CollectedEvent{
			ID:       event.ID,
			Type:     event.Type,
			Radio:    event.Radio,
			Received: time.Now(),
		}

		// Parse event data
		if event.Data != nil {
			// Convert map[string]interface{} to EventData
			if dataBytes, err := json.Marshal(event.Data); err == nil {
				json.Unmarshal(dataBytes, &collected.Data)
			}
		}

		c.events = append(c.events, collected)
		
		select {
		case c.ch <- collected:
		default:
			// Channel full, drop event
		}
	}
}

// Events returns all collected events.
func (c *EventCollector) Events() []CollectedEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make([]CollectedEvent, len(c.events))
	copy(result, c.events)
	return result
}

// Channel returns the event channel for real-time event consumption.
func (c *EventCollector) Channel() <-chan CollectedEvent {
	return c.ch
}

// Clear removes all collected events.
func (c *EventCollector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = c.events[:0]
}

// Close stops collecting new events.
func (c *EventCollector) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	close(c.ch)
}

// WaitForEvent waits for an event of the specified type within the timeout.
func (c *EventCollector) WaitForEvent(eventType string, timeout time.Duration) *CollectedEvent {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case event := <-c.ch:
			if event.Type == eventType {
				return &event
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// WaitForEvents waits for the specified number of events within the timeout.
func (c *EventCollector) WaitForEvents(count int, timeout time.Duration) []CollectedEvent {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var events []CollectedEvent
	for len(events) < count {
		select {
		case event := <-c.ch:
			events = append(events, event)
		case <-ctx.Done():
			return events
		}
	}
	return events
}
