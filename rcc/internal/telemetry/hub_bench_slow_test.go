//go:build slowbench

// Package telemetry provides slow performance benchmarks for deep profiling.
package telemetry

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/config"
)

func BenchmarkPublishWithManySubscribers(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Test with many subscribers for deep profiling
	subscriberCounts := []int{50, 100, 500}

	for _, count := range subscriberCounts {
		b.Run(fmt.Sprintf("Subscribers_%d", count), func(b *testing.B) {
			// Add timeout to prevent hangs
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

func BenchmarkPublishStressTest(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Add timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()

	// Run b.N iterations of Publish under stress
	for i := 0; i < b.N; i++ {
		// Check for timeout
		select {
		case <-ctx.Done():
			b.Fatal("Benchmark timed out - deadlock suspected")
		default:
		}

		event := Event{
			ID:    int64(i),
			Radio: fmt.Sprintf("radio-%d", i%100), // Cycle through 100 radios
			Type:  "powerChanged",
			Data:  map[string]interface{}{"powerDbm": 10.0 + float64(i%10)},
		}

		err := hub.Publish(event)
		if err != nil {
			b.Fatalf("Publish failed: %v", err)
		}
	}
}

func BenchmarkConcurrentSubscribers(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Add timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()

	// Run b.N iterations with concurrent subscribers
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
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
	})
}

func BenchmarkMemoryIntensive(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	// Use larger buffer for memory testing
	cfg.EventBufferSize = 1000
	cfg.EventBufferRetention = 10 * time.Minute

	hub := NewHub(cfg)
	defer hub.Stop()

	// Add timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()

	// Run b.N iterations with memory-intensive operations
	for i := 0; i < b.N; i++ {
		// Check for timeout
		select {
		case <-ctx.Done():
			b.Fatal("Benchmark timed out - deadlock suspected")
		default:
		}

		// Create large event with more data
		event := Event{
			ID:    int64(i),
			Radio: "silvus-001",
			Type:  "powerChanged",
			Data: map[string]interface{}{
				"powerDbm":   10.0 + float64(i%10),
				"frequency":  2412.0 + float64(i%100),
				"timestamp":  time.Now().Unix(),
				"metadata":   fmt.Sprintf("large_metadata_string_%d", i),
				"additional": make([]interface{}, 100), // Large data payload
			},
		}

		err := hub.Publish(event)
		if err != nil {
			b.Fatalf("Publish failed: %v", err)
		}
	}
}
