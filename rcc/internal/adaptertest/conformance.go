// Package adaptertest provides vendor-agnostic conformance testing for radio adapters.
//
//   - Architecture §8.5: "Error normalization to INVALID_RANGE, BUSY, UNAVAILABLE, INTERNAL"
//
package adaptertest

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
)

// Capabilities defines the expected capabilities for conformance testing.
type Capabilities struct {
	MinPowerDbm      int
	MaxPowerDbm      int
	ValidFrequencies []float64
	Channels         []adapter.Channel
	ExpectedErrors   ErrorExpectations
}

// ErrorExpectations defines expected error mappings for conformance testing.
type ErrorExpectations struct {
	InvalidRangeKeywords []string
	BusyKeywords         []string
	UnavailableKeywords  []string
	InternalKeywords     []string
}

// ConformanceResult represents the result of a conformance test.
type ConformanceResult struct {
	TestName string
	Passed   bool
	Error    string
	Duration time.Duration
	Details  map[string]interface{}
}

// ConformanceReport represents the complete conformance test report.
type ConformanceReport struct {
	AdapterName   string
	TotalTests    int
	PassedTests   int
	FailedTests   int
	Results       []ConformanceResult
	OverallPassed bool
	Duration      time.Duration
}

// RunConformance runs the complete conformance test suite for an adapter.
func RunConformance(t *testing.T, newAdapter func() adapter.IRadioAdapter, caps Capabilities) {
	startTime := time.Now()

	// Initialize conformance report
	report := &ConformanceReport{
		AdapterName:   "Unknown Adapter",
		TotalTests:    0,
		PassedTests:   0,
		FailedTests:   0,
		Results:       []ConformanceResult{},
		OverallPassed: true,
	}

	// Run all conformance tests
	runGetStateTests(t, newAdapter, caps, report)
	runSetPowerTests(t, newAdapter, caps, report)
	runSetChannelTests(t, newAdapter, caps, report)
	runFailureMappingTests(t, newAdapter, caps, report)
	runIdempotencyTests(t, newAdapter, caps, report)
	runTimingTests(t, newAdapter, caps, report)

	report.Duration = time.Since(startTime)

	// Print conformance report
	printConformanceReport(t, report)

	// Fail test if any tests failed
	if !report.OverallPassed {
		t.Fatalf("Adapter conformance test failed: %d/%d tests passed", report.PassedTests, report.TotalTests)
	}
}

// runGetStateTests tests the GetState functionality.
func runGetStateTests(t *testing.T, newAdapter func() adapter.IRadioAdapter, caps Capabilities, report *ConformanceReport) {
	adapter := newAdapter()
	ctx := context.Background()

	// Test basic GetState functionality
	result := ConformanceResult{
		TestName: "GetState_Basic",
		Details:  make(map[string]interface{}),
	}
	start := time.Now()

	state, err := adapter.GetState(ctx)
	result.Duration = time.Since(start)

	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("GetState failed: %v", err)
	} else if state == nil {
		result.Passed = false
		result.Error = "GetState returned nil state"
	} else {
		result.Passed = true
		result.Details["powerDbm"] = state.PowerDbm
		result.Details["frequencyMhz"] = state.FrequencyMhz
	}

	report.addResult(result)
}

// runSetPowerTests tests the SetPower functionality with various power ranges.
func runSetPowerTests(t *testing.T, newAdapter func() adapter.IRadioAdapter, caps Capabilities, report *ConformanceReport) {
	adapter := newAdapter()
	ctx := context.Background()

	// Test valid power ranges
	validPowers := []float64{float64(caps.MinPowerDbm), float64(caps.MaxPowerDbm), float64((caps.MinPowerDbm + caps.MaxPowerDbm) / 2)}

	for _, power := range validPowers {
		result := ConformanceResult{
			TestName: fmt.Sprintf("SetPower_Valid_%f", power),
			Details:  make(map[string]interface{}),
		}
		start := time.Now()

		err := adapter.SetPower(ctx, power)
		result.Duration = time.Since(start)

		if err != nil {
			result.Passed = false
			result.Error = fmt.Sprintf("SetPower(%f) failed: %v", power, err)
		} else {
			result.Passed = true
			result.Details["power"] = power
		}

		report.addResult(result)
	}

	// Test invalid power ranges
	invalidPowers := []float64{float64(caps.MinPowerDbm - 1), float64(caps.MaxPowerDbm + 1), -1.0, 100.0}

	for _, power := range invalidPowers {
		result := ConformanceResult{
			TestName: fmt.Sprintf("SetPower_Invalid_%f", power),
			Details:  make(map[string]interface{}),
		}
		start := time.Now()

		err := adapter.SetPower(ctx, power)
		result.Duration = time.Since(start)

		// Should return INVALID_RANGE error
		if err == nil {
			result.Passed = false
			result.Error = fmt.Sprintf("SetPower(%f) should have failed but succeeded", power)
		} else if !isInvalidRangeError(err) {
			result.Passed = false
			result.Error = fmt.Sprintf("SetPower(%f) should return INVALID_RANGE, got: %v", power, err)
		} else {
			result.Passed = true
			result.Details["expectedError"] = "INVALID_RANGE"
			result.Details["actualError"] = err.Error()
		}

		report.addResult(result)
	}
}

// runSetChannelTests tests the SetChannel functionality with frequency mapping.
func runSetChannelTests(t *testing.T, newAdapter func() adapter.IRadioAdapter, caps Capabilities, report *ConformanceReport) {
	adapter := newAdapter()
	ctx := context.Background()

	// Test valid frequencies
	for _, freq := range caps.ValidFrequencies {
		result := ConformanceResult{
			TestName: fmt.Sprintf("SetChannel_Valid_%.1f", freq),
			Details:  make(map[string]interface{}),
		}
		start := time.Now()

		err := adapter.SetFrequency(ctx, freq)
		result.Duration = time.Since(start)

		if err != nil {
			result.Passed = false
			result.Error = fmt.Sprintf("SetFrequency(%.1f) failed: %v", freq, err)
		} else {
			result.Passed = true
			result.Details["frequency"] = freq
		}

		report.addResult(result)
	}

	// Test invalid frequencies
	invalidFrequencies := []float64{0, -100, 100000}

	for _, freq := range invalidFrequencies {
		result := ConformanceResult{
			TestName: fmt.Sprintf("SetChannel_Invalid_%.1f", freq),
			Details:  make(map[string]interface{}),
		}
		start := time.Now()

		err := adapter.SetFrequency(ctx, freq)
		result.Duration = time.Since(start)

		// Should return INVALID_RANGE error
		if err == nil {
			result.Passed = false
			result.Error = fmt.Sprintf("SetFrequency(%.1f) should have failed but succeeded", freq)
		} else if !isInvalidRangeError(err) {
			result.Passed = false
			result.Error = fmt.Sprintf("SetFrequency(%.1f) should return INVALID_RANGE, got: %v", freq, err)
		} else {
			result.Passed = true
			result.Details["expectedError"] = "INVALID_RANGE"
			result.Details["actualError"] = err.Error()
		}

		report.addResult(result)
	}
}

// runFailureMappingTests tests error mapping to normalized error codes.
func runFailureMappingTests(t *testing.T, newAdapter func() adapter.IRadioAdapter, caps Capabilities, report *ConformanceReport) {
	adapter := newAdapter()
	ctx := context.Background()

	// Test error mapping (this would require a mock adapter that can simulate different error conditions)
	// For now, we'll test that the adapter properly handles context cancellation
	result := ConformanceResult{
		TestName: "FailureMapping_ContextCancellation",
		Details:  make(map[string]interface{}),
	}
	start := time.Now()

	// Create a cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	_, err := adapter.GetState(cancelledCtx)
	result.Duration = time.Since(start)

	// Should handle context cancellation gracefully
	if err == nil {
		result.Passed = false
		result.Error = "GetState with cancelled context should have failed"
	} else {
		result.Passed = true
		result.Details["error"] = err.Error()
	}

	report.addResult(result)
}

// runIdempotencyTests tests that setting the same power multiple times is idempotent.
func runIdempotencyTests(t *testing.T, newAdapter func() adapter.IRadioAdapter, caps Capabilities, report *ConformanceReport) {
	adapter := newAdapter()
	ctx := context.Background()

	// Test idempotency with a valid power level
	testPower := float64((caps.MinPowerDbm + caps.MaxPowerDbm) / 2)

	result := ConformanceResult{
		TestName: "Idempotency_SetSamePower",
		Details:  make(map[string]interface{}),
	}
	start := time.Now()

	// Set power twice
	err1 := adapter.SetPower(ctx, testPower)
	err2 := adapter.SetPower(ctx, testPower)
	result.Duration = time.Since(start)

	if err1 != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("First SetPower(%f) failed: %v", testPower, err1)
	} else if err2 != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Second SetPower(%f) failed: %v", testPower, err2)
	} else {
		result.Passed = true
		result.Details["power"] = testPower
		result.Details["firstCall"] = "success"
		result.Details["secondCall"] = "success"
	}

	report.addResult(result)
}

// runTimingTests tests that timeouts are honored and no sleeps are used in adaptertest.
func runTimingTests(t *testing.T, newAdapter func() adapter.IRadioAdapter, caps Capabilities, report *ConformanceReport) {
	adapter := newAdapter()

	// Test that operations complete within reasonable time (no sleeps in adaptertest)
	result := ConformanceResult{
		TestName: "Timing_NoSleeps",
		Details:  make(map[string]interface{}),
	}
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := adapter.GetState(ctx)
	result.Duration = time.Since(start)

	// Should complete quickly (no sleeps)
	if result.Duration > 50*time.Millisecond {
		result.Passed = false
		result.Error = fmt.Sprintf("Operation took too long: %v (no sleeps allowed in adaptertest)", result.Duration)
	} else if err != nil && !strings.Contains(err.Error(), "context deadline exceeded") {
		result.Passed = false
		result.Error = fmt.Sprintf("Unexpected error: %v", err)
	} else {
		result.Passed = true
		result.Details["duration"] = result.Duration.String()
	}

	report.addResult(result)
}

// Helper functions

func isInvalidRangeError(err error) bool {
	return strings.Contains(err.Error(), "INVALID_RANGE")
}

func (r *ConformanceReport) addResult(result ConformanceResult) {
	r.TotalTests++
	if result.Passed {
		r.PassedTests++
	} else {
		r.FailedTests++
		r.OverallPassed = false
	}
	r.Results = append(r.Results, result)
}

func printConformanceReport(t *testing.T, report *ConformanceReport) {
	t.Logf("\n%s", strings.Repeat("=", 80))
	t.Logf("ADAPTER CONFORMANCE REPORT")
	t.Logf("%s", strings.Repeat("=", 80))
	t.Logf("Adapter: %s", report.AdapterName)
	t.Logf("Total Tests: %d", report.TotalTests)
	t.Logf("Passed: %d", report.PassedTests)
	t.Logf("Failed: %d", report.FailedTests)
	t.Logf("Overall: %s", map[bool]string{true: "PASS", false: "FAIL"}[report.OverallPassed])
	t.Logf("Duration: %v", report.Duration)
	t.Logf("%s", strings.Repeat("-", 80))

	// Print detailed results table
	t.Logf("%-30s %-8s %-12s %-s", "TEST NAME", "RESULT", "DURATION", "DETAILS")
	t.Logf("%s", strings.Repeat("-", 80))

	for _, result := range report.Results {
		status := "PASS"
		if !result.Passed {
			status = "FAIL"
		}

		details := ""
		if result.Error != "" {
			details = result.Error
		} else if len(result.Details) > 0 {
			// Format details as key=value pairs
			var detailParts []string
			for k, v := range result.Details {
				detailParts = append(detailParts, fmt.Sprintf("%s=%v", k, v))
			}
			details = strings.Join(detailParts, ", ")
		}

		t.Logf("%-30s %-8s %-12s %-s",
			result.TestName,
			status,
			result.Duration.String(),
			details)
	}

	t.Logf("%s", strings.Repeat("=", 80))
}
