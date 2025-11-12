//go:build !slowbench

// Package telemetry provides performance benchmarks for the telemetry hub.
package telemetry

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/config"
)

func BenchmarkPublishWithSubscribers(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Test with different numbers of subscribers (reduced for fast execution)
	subscriberCounts := []int{1, 5, 10}

	for _, count := range subscriberCounts {
		b.Run(fmt.Sprintf("Subscribers_%d", count), func(b *testing.B) {
			// Add timeout to prevent 11-minute hangs
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Create subscribers
			for i := 0; i < count; i++ {
				req := httptest.NewRequest("GET", "/telemetry", nil)
				req.Header.Set("Accept", "text/event-stream")
				w := httptest.NewRecorder()

				// Run Subscribe in goroutine to avoid blocking
				go func() {
					hub.Subscribe(ctx, w, req)
				}()

				// Give Subscribe time to register the client
				time.Sleep(10 * time.Millisecond)
			}

			b.ResetTimer()

			// Run b.N iterations of Publish with timeout protection
			for i := 0; i < b.N; i++ {
				// Check for timeout
				select {
				case <-ctx.Done():
					b.Fatal("Benchmark timed out - deadlock suspected")
				default:
				}

				event := Event{
					ID:    int64(i),
					Radio: "silvus-001",
					Type:  "powerChanged",
					Data:  map[string]interface{}{"powerDbm": 10.0 + float64(i%10)},
				}

				err := hub.Publish(event)
				if err != nil {
					b.Fatalf("Publish failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkPublishWithoutSubscribers(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Add timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b.ResetTimer()

	// Run b.N iterations of Publish without subscribers
	for i := 0; i < b.N; i++ {
		// Check for timeout
		select {
		case <-ctx.Done():
			b.Fatal("Benchmark timed out - deadlock suspected")
		default:
		}

		event := Event{
			ID:    int64(i),
			Radio: "silvus-001",
			Type:  "powerChanged",
			Data:  map[string]interface{}{"powerDbm": 10.0 + float64(i%10)},
		}

		err := hub.Publish(event)
		if err != nil {
			b.Fatalf("Publish failed: %v", err)
		}
	}
}

func BenchmarkSubscribe(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Add timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b.ResetTimer()

	// Run b.N iterations of Subscribe
	for i := 0; i < b.N; i++ {
		// Check for timeout
		select {
		case <-ctx.Done():
			b.Fatal("Benchmark timed out - deadlock suspected")
		default:
		}

		req := httptest.NewRequest("GET", "/telemetry", nil)
		req.Header.Set("Accept", "text/event-stream")

		subscribeCtx, subscribeCancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer subscribeCancel()

		w := httptest.NewRecorder()
		hub.Subscribe(subscribeCtx, w, req)
	}
}

func BenchmarkEventIDGeneration(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Add timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b.ResetTimer()

	// Run b.N iterations of getNextEventID
	for i := 0; i < b.N; i++ {
		// Check for timeout
		select {
		case <-ctx.Done():
			b.Fatal("Benchmark timed out - deadlock suspected")
		default:
		}

		radioID := fmt.Sprintf("radio-%d", i%10) // Cycle through 10 radios
		hub.getNextEventID(radioID)
	}
}

func BenchmarkBufferEvent(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Add timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b.ResetTimer()

	// Run b.N iterations of bufferEvent
	for i := 0; i < b.N; i++ {
		// Check for timeout
		select {
		case <-ctx.Done():
			b.Fatal("Benchmark timed out - deadlock suspected")
		default:
		}

		event := Event{
			ID:    int64(i),
			Radio: fmt.Sprintf("radio-%d", i%10),
			Type:  "powerChanged",
			Data:  map[string]interface{}{"powerDbm": 10.0 + float64(i%10)},
		}

		hub.bufferEvent(event)
	}
}

func BenchmarkHubConcurrent(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Add timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b.ResetTimer()

	// Run b.N iterations with concurrent operations
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Check for timeout
			select {
			case <-ctx.Done():
				b.Fatal("Benchmark timed out - deadlock suspected")
			default:
			}

			// Mix of operations
			switch b.N % 3 {
			case 0:
				// Publish event
				event := Event{
					ID:    int64(b.N),
					Radio: "silvus-001",
					Type:  "powerChanged",
					Data:  map[string]interface{}{"powerDbm": 10.0},
				}
				hub.Publish(event)
			case 1:
				// Generate event ID
				hub.getNextEventID("silvus-001")
			case 2:
				// Buffer event
				event := Event{
					ID:    int64(b.N),
					Radio: "silvus-001",
					Type:  "powerChanged",
					Data:  map[string]interface{}{"powerDbm": 10.0},
				}
				hub.bufferEvent(event)
			}
		}
	})
}

func BenchmarkHeartbeat(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	// Use shorter heartbeat interval for testing
	cfg.HeartbeatInterval = 10 * time.Millisecond
	cfg.HeartbeatJitter = 1 * time.Millisecond

	hub := NewHub(cfg)
	defer hub.Stop()

	// Start heartbeat
	hub.startHeartbeat()

	// Add timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b.ResetTimer()

	// Run b.N iterations of sendHeartbeat
	for i := 0; i < b.N; i++ {
		// Check for timeout
		select {
		case <-ctx.Done():
			b.Fatal("Benchmark timed out - deadlock suspected")
		default:
		}

		hub.sendHeartbeat()
	}
}
