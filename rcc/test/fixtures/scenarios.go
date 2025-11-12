package fixtures

import (
	"time"

	"github.com/radio-control/rcc/internal/config"
)

// TestScenario represents a complete test scenario with setup and expected outcomes
type TestScenario struct {
	Name        string
	Description string
	Setup       ScenarioSetup
	Actions     []ScenarioAction
	Expected    ScenarioExpected
}

type ScenarioSetup struct {
	Radios   []RadioProfile
	Channels []ChannelProfile
	Config   *config.TimingConfig
	Duration time.Duration
}

type ScenarioAction struct {
	Type    string
	RadioID string
	Params  map[string]interface{}
	Delay   time.Duration
}

type ScenarioExpected struct {
	SuccessCount  int
	ErrorCount    int
	EventCount    int
	AuditLogCount int
	MaxLatency    time.Duration
}

// HappyPath returns a happy path scenario for testing
func HappyPath(radio RadioProfile, channels []ChannelProfile) TestScenario {
	return TestScenario{
		Name:        "Happy Path",
		Description: "Standard successful operations flow",
		Setup: ScenarioSetup{
			Radios:   []RadioProfile{radio},
			Channels: channels,
			Config:   LoadTestConfig(),
			Duration: 30 * time.Second,
		},
		Actions: []ScenarioAction{
			{Type: "setChannel", RadioID: radio.ID, Params: map[string]interface{}{"channel": 6}},
			{Type: "setPower", RadioID: radio.ID, Params: map[string]interface{}{"power": 5}},
			{Type: "getState", RadioID: radio.ID, Params: map[string]interface{}{}},
		},
		Expected: ScenarioExpected{
			SuccessCount:  3,
			ErrorCount:    0,
			EventCount:    3,
			AuditLogCount: 3,
			MaxLatency:    5 * time.Second,
		},
	}
}

// ErrorRecovery returns an error recovery scenario for testing
func ErrorRecovery(radio RadioProfile, channels []ChannelProfile) TestScenario {
	return TestScenario{
		Name:        "Error Recovery",
		Description: "Operations with errors and recovery",
		Setup: ScenarioSetup{
			Radios:   []RadioProfile{radio},
			Channels: channels,
			Config:   LoadTestConfig(),
			Duration: 60 * time.Second,
		},
		Actions: []ScenarioAction{
			{Type: "setChannel", RadioID: radio.ID, Params: map[string]interface{}{"channel": 999}}, // Invalid
			{Type: "setChannel", RadioID: radio.ID, Params: map[string]interface{}{"channel": 6}},   // Valid
			{Type: "setPower", RadioID: radio.ID, Params: map[string]interface{}{"power": 15}},      // Invalid
			{Type: "setPower", RadioID: radio.ID, Params: map[string]interface{}{"power": 5}},       // Valid
		},
		Expected: ScenarioExpected{
			SuccessCount:  2,
			ErrorCount:    2,
			EventCount:    4,
			AuditLogCount: 4,
			MaxLatency:    10 * time.Second,
		},
	}
}

// Concurrent returns a concurrent operations scenario for testing
func Concurrent(radios []RadioProfile, channels []ChannelProfile) TestScenario {
	var actions []ScenarioAction

	// Generate concurrent actions for multiple radios
	for i, radio := range radios {
		actions = append(actions, []ScenarioAction{
			{Type: "setChannel", RadioID: radio.ID, Params: map[string]interface{}{"channel": channels[i%len(channels)].Index}},
			{Type: "setPower", RadioID: radio.ID, Params: map[string]interface{}{"power": 5 + i}},
			{Type: "getState", RadioID: radio.ID, Params: map[string]interface{}{}},
		}...)
	}

	return TestScenario{
		Name:        "Concurrent Operations",
		Description: "Multiple radios operating concurrently",
		Setup: ScenarioSetup{
			Radios:   radios,
			Channels: channels,
			Config:   LoadTestConfig(),
			Duration: 45 * time.Second,
		},
		Actions: actions,
		Expected: ScenarioExpected{
			SuccessCount:  len(actions),
			ErrorCount:    0,
			EventCount:    len(actions),
			AuditLogCount: len(actions),
			MaxLatency:    15 * time.Second,
		},
	}
}

// LoadTest returns a load testing scenario for testing
func LoadTest(radios []RadioProfile, channels []ChannelProfile) TestScenario {
	var actions []ScenarioAction

	// Generate high-frequency actions
	for i := 0; i < 100; i++ {
		radio := radios[i%len(radios)]
		channel := channels[i%len(channels)]

		actions = append(actions, ScenarioAction{
			Type:    "setChannel",
			RadioID: radio.ID,
			Params:  map[string]interface{}{"channel": channel.Index},
			Delay:   time.Duration(i%10) * time.Millisecond,
		})
	}

	return TestScenario{
		Name:        "Load Test",
		Description: "High-frequency operations for load testing",
		Setup: ScenarioSetup{
			Radios:   radios,
			Channels: channels,
			Config:   LoadTestConfig(),
			Duration: 2 * time.Minute,
		},
		Actions: actions,
		Expected: ScenarioExpected{
			SuccessCount:  90, // Allow 10% error rate
			ErrorCount:    10,
			EventCount:    100,
			AuditLogCount: 100,
			MaxLatency:    1 * time.Second,
		},
	}
}

// StressTest returns a stress testing scenario for testing
func StressTest(radios []RadioProfile, channels []ChannelProfile) TestScenario {
	var actions []ScenarioAction

	// Generate stress test actions with rapid succession
	for i := 0; i < 50; i++ {
		radio := radios[i%len(radios)]

		actions = append(actions, ScenarioAction{
			Type:    "setPower",
			RadioID: radio.ID,
			Params:  map[string]interface{}{"power": (i % 10) + 1},
			Delay:   time.Duration(i%5) * time.Millisecond,
		})
	}

	return TestScenario{
		Name:        "Stress Test",
		Description: "Rapid succession operations for stress testing",
		Setup: ScenarioSetup{
			Radios:   radios,
			Channels: channels,
			Config:   LoadTestConfig(),
			Duration: 1 * time.Minute,
		},
		Actions: actions,
		Expected: ScenarioExpected{
			SuccessCount:  40, // Allow 20% error rate under stress
			ErrorCount:    10,
			EventCount:    50,
			AuditLogCount: 50,
			MaxLatency:    2 * time.Second,
		},
	}
}

// FailureScenario returns a failure scenario for testing
func FailureScenario(radio RadioProfile, channels []ChannelProfile) TestScenario {
	return TestScenario{
		Name:        "Failure Scenario",
		Description: "Operations that are expected to fail",
		Setup: ScenarioSetup{
			Radios:   []RadioProfile{radio},
			Channels: channels,
			Config:   LoadTestConfig(),
			Duration: 30 * time.Second,
		},
		Actions: []ScenarioAction{
			{Type: "setChannel", RadioID: "nonexistent-radio", Params: map[string]interface{}{"channel": 6}},
			{Type: "setPower", RadioID: radio.ID, Params: map[string]interface{}{"power": 999}},
			{Type: "setChannel", RadioID: radio.ID, Params: map[string]interface{}{"channel": -1}},
		},
		Expected: ScenarioExpected{
			SuccessCount:  0,
			ErrorCount:    3,
			EventCount:    3,
			AuditLogCount: 3,
			MaxLatency:    5 * time.Second,
		},
	}
}
