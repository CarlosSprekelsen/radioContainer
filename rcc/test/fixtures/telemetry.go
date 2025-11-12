package fixtures

import (
	"math/rand"
	"time"

	"github.com/radio-control/rcc/internal/telemetry"
)

// TelemetryEvent represents a standardized telemetry event for testing
type TelemetryEvent struct {
	ID        string
	Type      string
	RadioID   string
	Timestamp time.Time
	Data      map[string]interface{}
}

// HeartbeatSequence returns a sequence of heartbeat events for testing
func HeartbeatSequence() []TelemetryEvent {
	events := make([]TelemetryEvent, 5)
	baseTime := time.Now()

	for i := 0; i < 5; i++ {
		events[i] = TelemetryEvent{
			ID:        "heartbeat-" + string(rune('0'+i)),
			Type:      "heartbeat",
			RadioID:   "silvus-001",
			Timestamp: baseTime.Add(time.Duration(i) * 15 * time.Second),
			Data: map[string]interface{}{
				"status": "online",
				"rssi":   -65 + i,
				"snr":    25 + i,
			},
		}
	}

	return events
}

// PowerChangeSequence returns a sequence of power change events for testing
func PowerChangeSequence() []TelemetryEvent {
	events := make([]TelemetryEvent, 3)
	baseTime := time.Now()

	powerLevels := []int{3, 5, 7}

	for i, power := range powerLevels {
		events[i] = TelemetryEvent{
			ID:        "power-change-" + string(rune('0'+i)),
			Type:      "power_change",
			RadioID:   "silvus-001",
			Timestamp: baseTime.Add(time.Duration(i) * 2 * time.Second),
			Data: map[string]interface{}{
				"old_power": power - 1,
				"new_power": power,
				"channel":   6,
			},
		}
	}

	return events
}

// ChannelChangeSequence returns a sequence of channel change events for testing
func ChannelChangeSequence() []TelemetryEvent {
	events := make([]TelemetryEvent, 3)
	baseTime := time.Now()

	channels := []int{1, 6, 11}

	for i, channel := range channels {
		events[i] = TelemetryEvent{
			ID:        "channel-change-" + string(rune('0'+i)),
			Type:      "channel_change",
			RadioID:   "silvus-001",
			Timestamp: baseTime.Add(time.Duration(i) * 5 * time.Second),
			Data: map[string]interface{}{
				"old_channel": channel - 1,
				"new_channel": channel,
				"frequency":   2412.0 + float64(channel-1)*5.0,
			},
		}
	}

	return events
}

// ErrorEventSequence returns a sequence of error events for testing
func ErrorEventSequence() []TelemetryEvent {
	events := make([]TelemetryEvent, 3)
	baseTime := time.Now()

	errorTypes := []string{"BUSY", "INVALID_RANGE", "INTERNAL"}

	for i, errorType := range errorTypes {
		events[i] = TelemetryEvent{
			ID:        "error-" + string(rune('0'+i)),
			Type:      "error",
			RadioID:   "silvus-001",
			Timestamp: baseTime.Add(time.Duration(i) * 10 * time.Second),
			Data: map[string]interface{}{
				"error_code":  errorType,
				"operation":   "setPower",
				"retry_count": i,
			},
		}
	}

	return events
}

// GenerateEventSequence generates a sequence of events with the specified count
func GenerateEventSequence(count int) []TelemetryEvent {
	events := make([]TelemetryEvent, count)
	baseTime := time.Now()

	// Use deterministic seed for consistent test results
	rand.Seed(42)

	for i := 0; i < count; i++ {
		eventTypes := []string{"heartbeat", "power_change", "channel_change", "error"}
		eventType := eventTypes[rand.Intn(len(eventTypes))]

		events[i] = TelemetryEvent{
			ID:        "event-" + string(rune('0'+i%10)),
			Type:      eventType,
			RadioID:   "silvus-001",
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			Data: map[string]interface{}{
				"sequence": i,
				"random":   rand.Float64(),
			},
		}
	}

	return events
}

// MixedEventSequence returns a mixed sequence of different event types for testing
func MixedEventSequence() []TelemetryEvent {
	var events []TelemetryEvent

	// Add heartbeat events
	events = append(events, HeartbeatSequence()...)

	// Add power change events
	events = append(events, PowerChangeSequence()...)

	// Add channel change events
	events = append(events, ChannelChangeSequence()...)

	// Add error events
	events = append(events, ErrorEventSequence()...)

	return events
}

// HighFrequencyEventSequence returns a high-frequency event sequence for testing
func HighFrequencyEventSequence() []TelemetryEvent {
	events := make([]TelemetryEvent, 100)
	baseTime := time.Now()

	for i := 0; i < 100; i++ {
		events[i] = TelemetryEvent{
			ID:        "hf-event-" + string(rune('0'+i%10)),
			Type:      "heartbeat",
			RadioID:   "silvus-001",
			Timestamp: baseTime.Add(time.Duration(i) * 100 * time.Millisecond), // 10 events per second
			Data: map[string]interface{}{
				"sequence":  i,
				"frequency": "high",
			},
		}
	}

	return events
}

// HeartbeatSequenceEvents returns a sequence of heartbeat events for testing using telemetry.Event
func HeartbeatSequenceEvents() []telemetry.Event {
	events := make([]telemetry.Event, 5)

	for i := 0; i < 5; i++ {
		events[i] = telemetry.Event{
			ID:   int64(i + 1),
			Type: "heartbeat",
			Data: map[string]interface{}{
				"status": "online",
				"rssi":   -65 + i,
				"snr":    25 + i,
			},
			Radio: "silvus-001",
		}
	}

	return events
}

// PowerChangeSequenceEvents returns a sequence of power change events for testing using telemetry.Event
func PowerChangeSequenceEvents() []telemetry.Event {
	events := make([]telemetry.Event, 3)

	powerLevels := []int{3, 5, 7}

	for i, power := range powerLevels {
		events[i] = telemetry.Event{
			ID:   int64(i + 1),
			Type: "power_change",
			Data: map[string]interface{}{
				"old_power": power - 1,
				"new_power": power,
				"channel":   6,
			},
			Radio: "silvus-001",
		}
	}

	return events
}

// GenerateEventSequenceEvents generates a sequence of events with the specified count using telemetry.Event
func GenerateEventSequenceEvents(count int) []telemetry.Event {
	events := make([]telemetry.Event, count)

	// Use deterministic seed for consistent test results
	rand.Seed(42)

	for i := 0; i < count; i++ {
		eventTypes := []string{"heartbeat", "power_change", "channel_change", "error"}
		eventType := eventTypes[rand.Intn(len(eventTypes))]

		events[i] = telemetry.Event{
			ID:   int64(i + 1),
			Type: eventType,
			Data: map[string]interface{}{
				"sequence": i,
				"random":   rand.Float64(),
			},
			Radio: "silvus-001",
		}
	}

	return events
}
