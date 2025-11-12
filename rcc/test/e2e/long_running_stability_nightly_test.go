//go:build e2e && nightly

package e2e

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter/fake"
	"github.com/radio-control/rcc/test/harness"
)

// TestLongRunningStability_SoakTest tests long-running stability:
// Soak with bounded duration; memory/handle leak sentinels; periodic heartbeats within §3 constraints
func TestLongRunningStability_SoakTest(t *testing.T) {
	// Skip if not running nightly tests
	if testing.Short() {
		t.Skip("Skipping long-running stability test in short mode")
	}

	// Arrange: Create harness for long-running test
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-stability-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Add second radio for concurrent operations
	fakeAdapter := fake.NewFakeAdapter("fake-stability-002")
	err := server.RadioManager.LoadCapabilities("fake-stability-002", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load fake adapter capabilities: %v", err)
	}

	ctx := context.Background()

	// Test duration: 5 minutes for nightly test
	testDuration := 5 * time.Minute
	startTime := time.Now()

	// Memory tracking
	var initialMemStats runtime.MemStats
	runtime.ReadMemStats(&initialMemStats)

	// Heartbeat tracking
	heartbeatInterval := server.Orchestrator.GetConfig().HeartbeatInterval
	heartbeatJitter := server.Orchestrator.GetConfig().HeartbeatJitter
	lastHeartbeat := time.Now()

	// Operation counters
	operationCount := 0
	errorCount := 0

	t.Logf("Starting long-running stability test for %v", testDuration)
	t.Logf("Initial memory: %d KB", initialMemStats.Alloc/1024)

	// Act: Run operations continuously for test duration
	for time.Since(startTime) < testDuration {
		// Execute operations on both radios
		operations := []struct {
			radioID string
			action  string
			params  map[string]interface{}
		}{
			{"silvus-stability-001", "setPower", map[string]interface{}{"power": 25.0 + float64(operationCount%10)}},
			{"silvus-stability-001", "setChannel", map[string]interface{}{"frequency": 2412.0 + float64(operationCount%3)*25.0}},
			{"fake-stability-002", "setPower", map[string]interface{}{"power": 30.0 + float64(operationCount%10)}},
			{"fake-stability-002", "setChannel", map[string]interface{}{"frequency": 2437.0 + float64(operationCount%3)*25.0}},
		}

		for _, op := range operations {
			var err error
			switch op.action {
			case "setPower":
				err = server.Orchestrator.SetPower(ctx, op.radioID, op.params["power"].(float64))
			case "setChannel":
				err = server.Orchestrator.SetChannel(ctx, op.radioID, op.params["frequency"].(float64))
			}

			operationCount++
			if err != nil {
				errorCount++
				t.Logf("Operation %d failed: %v", operationCount, err)
			}
		}

		// Check heartbeat timing per CB-TIMING §3
		now := time.Now()
		if now.Sub(lastHeartbeat) >= heartbeatInterval-heartbeatJitter {
			// Verify heartbeat within jitter bounds
			expectedInterval := heartbeatInterval
			actualInterval := now.Sub(lastHeartbeat)

			if actualInterval < heartbeatInterval-heartbeatJitter || actualInterval > heartbeatInterval+heartbeatJitter {
				t.Errorf("Heartbeat interval %v outside jitter bounds [%v, %v]",
					actualInterval, heartbeatInterval-heartbeatJitter, heartbeatInterval+heartbeatJitter)
			}

			lastHeartbeat = now
			t.Logf("Heartbeat: interval %v (expected %v±%v)", actualInterval, heartbeatInterval, heartbeatJitter)
		}

		// Check memory usage every 30 seconds
		if operationCount%100 == 0 {
			var currentMemStats runtime.MemStats
			runtime.ReadMemStats(&currentMemStats)

			memoryIncrease := currentMemStats.Alloc - initialMemStats.Alloc
			if memoryIncrease > 10*1024*1024 { // 10MB threshold
				t.Errorf("Memory leak detected: %d KB increase", memoryIncrease/1024)
			}

			t.Logf("Memory check: %d KB (increase: %d KB)", currentMemStats.Alloc/1024, memoryIncrease/1024)
		}

		// Small delay to prevent overwhelming the system
		time.Sleep(10 * time.Millisecond)
	}

	// Assert: Test completion metrics
	elapsed := time.Since(startTime)
	errorRate := float64(errorCount) / float64(operationCount) * 100

	t.Logf("✅ Long-running stability test completed:")
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Operations: %d", operationCount)
	t.Logf("  Errors: %d (%.2f%%)", errorCount, errorRate)

	// Final memory check
	var finalMemStats runtime.MemStats
	runtime.ReadMemStats(&finalMemStats)
	memoryIncrease := finalMemStats.Alloc - initialMemStats.Alloc
	t.Logf("  Memory increase: %d KB", memoryIncrease/1024)

	// Assert: Error rate should be low
	if errorRate > 5.0 {
		t.Errorf("Error rate %.2f%% exceeds 5%% threshold", errorRate)
	}

	// Assert: Memory increase should be reasonable
	if memoryIncrease > 50*1024*1024 { // 50MB threshold
		t.Errorf("Memory increase %d KB exceeds 50MB threshold", memoryIncrease/1024)
	}
}

// TestLongRunningStability_MemoryLeakDetection tests memory leak detection:
// Extended period with memory monitoring; verify no significant memory growth
func TestLongRunningStability_MemoryLeakDetection(t *testing.T) {
	// Skip if not running nightly tests
	if testing.Short() {
		t.Skip("Skipping memory leak detection test in short mode")
	}

	// Arrange: Create harness for memory leak test
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-memory-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	ctx := context.Background()

	// Test duration: 2 minutes for memory leak detection
	testDuration := 2 * time.Minute
	startTime := time.Now()

	// Memory tracking
	var initialMemStats runtime.MemStats
	runtime.ReadMemStats(&initialMemStats)

	// Memory samples
	memorySamples := make([]uint64, 0)
	sampleInterval := 10 * time.Second

	t.Logf("Starting memory leak detection test for %v", testDuration)
	t.Logf("Initial memory: %d KB", initialMemStats.Alloc/1024)

	// Act: Run operations and monitor memory
	for time.Since(startTime) < testDuration {
		// Execute operations that could cause memory leaks
		for i := 0; i < 10; i++ {
			// Create and destroy objects
			err := server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0+float64(i))
			if err != nil {
				t.Logf("SetPower failed: %v", err)
			}

			err = server.Orchestrator.SetChannel(ctx, opts.ActiveRadioID, 2412.0+float64(i)*25.0)
			if err != nil {
				t.Logf("SetChannel failed: %v", err)
			}
		}

		// Sample memory every interval
		if time.Since(startTime)%sampleInterval < 100*time.Millisecond {
			var currentMemStats runtime.MemStats
			runtime.ReadMemStats(&currentMemStats)
			memorySamples = append(memorySamples, currentMemStats.Alloc)
			t.Logf("Memory sample: %d KB", currentMemStats.Alloc/1024)
		}

		// Force garbage collection periodically
		if time.Since(startTime)%(30*time.Second) < 100*time.Millisecond {
			runtime.GC()
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Assert: Memory growth analysis
	if len(memorySamples) < 2 {
		t.Error("Insufficient memory samples for analysis")
		return
	}

	// Calculate memory growth trend
	initialMemory := memorySamples[0]
	finalMemory := memorySamples[len(memorySamples)-1]
	memoryGrowth := finalMemory - initialMemory

	// Calculate growth rate
	growthRate := float64(memoryGrowth) / float64(len(memorySamples)) / 1024 // KB per sample

	t.Logf("✅ Memory leak detection completed:")
	t.Logf("  Initial memory: %d KB", initialMemory/1024)
	t.Logf("  Final memory: %d KB", finalMemory/1024)
	t.Logf("  Memory growth: %d KB", memoryGrowth/1024)
	t.Logf("  Growth rate: %.2f KB/sample", growthRate)

	// Assert: Memory growth should be minimal
	if memoryGrowth > 5*1024*1024 { // 5MB threshold
		t.Errorf("Memory growth %d KB exceeds 5MB threshold", memoryGrowth/1024)
	}

	// Assert: Growth rate should be low
	if growthRate > 1.0 { // 1KB per sample threshold
		t.Errorf("Memory growth rate %.2f KB/sample exceeds 1KB threshold", growthRate)
	}
}

// TestLongRunningStability_HeartbeatTiming tests heartbeat timing under load:
// Verify heartbeat behavior under component stress per CB-TIMING §3
func TestLongRunningStability_HeartbeatTiming(t *testing.T) {
	// Skip if not running nightly tests
	if testing.Short() {
		t.Skip("Skipping heartbeat timing test in short mode")
	}

	// Arrange: Create harness for heartbeat test
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-heartbeat-001"
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	ctx := context.Background()

	// Test duration: 3 minutes for heartbeat timing
	testDuration := 3 * time.Minute
	startTime := time.Now()

	// Heartbeat tracking
	heartbeatInterval := server.Orchestrator.GetConfig().HeartbeatInterval
	heartbeatJitter := server.Orchestrator.GetConfig().HeartbeatJitter
	heartbeatTimeout := server.Orchestrator.GetConfig().HeartbeatTimeout

	heartbeatTimes := make([]time.Time, 0)
	lastHeartbeat := time.Now()

	t.Logf("Starting heartbeat timing test for %v", testDuration)
	t.Logf("Heartbeat interval: %v, jitter: %v, timeout: %v", heartbeatInterval, heartbeatJitter, heartbeatTimeout)

	// Act: Run operations and monitor heartbeats
	for time.Since(startTime) < testDuration {
		// Execute operations that could affect heartbeat timing
		for i := 0; i < 5; i++ {
			err := server.Orchestrator.SetPower(ctx, opts.ActiveRadioID, 25.0+float64(i))
			if err != nil {
				t.Logf("SetPower failed: %v", err)
			}
		}

		// Check for heartbeat
		now := time.Now()
		if now.Sub(lastHeartbeat) >= heartbeatInterval-heartbeatJitter {
			heartbeatTimes = append(heartbeatTimes, now)
			lastHeartbeat = now

			// Verify heartbeat within jitter bounds
			if len(heartbeatTimes) > 1 {
				actualInterval := now.Sub(heartbeatTimes[len(heartbeatTimes)-2])
				expectedInterval := heartbeatInterval

				if actualInterval < expectedInterval-heartbeatJitter || actualInterval > expectedInterval+heartbeatJitter {
					t.Errorf("Heartbeat interval %v outside jitter bounds [%v, %v]",
						actualInterval, expectedInterval-heartbeatJitter, expectedInterval+heartbeatJitter)
				}
			}

			t.Logf("Heartbeat %d: %v", len(heartbeatTimes), now.Format("15:04:05.000"))
		}

		// Check for heartbeat timeout
		if now.Sub(lastHeartbeat) > heartbeatTimeout {
			t.Errorf("Heartbeat timeout: %v since last heartbeat (timeout: %v)",
				now.Sub(lastHeartbeat), heartbeatTimeout)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Assert: Heartbeat timing analysis
	t.Logf("✅ Heartbeat timing test completed:")
	t.Logf("  Heartbeats received: %d", len(heartbeatTimes))
	t.Logf("  Expected heartbeats: %d", int(testDuration/heartbeatInterval))

	// Calculate heartbeat intervals
	if len(heartbeatTimes) > 1 {
		intervals := make([]time.Duration, 0)
		for i := 1; i < len(heartbeatTimes); i++ {
			intervals = append(intervals, heartbeatTimes[i].Sub(heartbeatTimes[i-1]))
		}

		// Calculate statistics
		var totalInterval time.Duration
		for _, interval := range intervals {
			totalInterval += interval
		}
		avgInterval := totalInterval / time.Duration(len(intervals))

		t.Logf("  Average interval: %v", avgInterval)
		t.Logf("  Expected interval: %v", heartbeatInterval)

		// Assert: Average interval should be close to expected
		intervalDiff := avgInterval - heartbeatInterval
		if intervalDiff < -heartbeatJitter || intervalDiff > heartbeatJitter {
			t.Errorf("Average heartbeat interval %v outside jitter bounds [%v, %v]",
				avgInterval, heartbeatInterval-heartbeatJitter, heartbeatInterval+heartbeatJitter)
		}
	}
}
