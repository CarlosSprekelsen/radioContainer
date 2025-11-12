package harness

import (
	"testing"
)

func TestHarness(t *testing.T) {
	opts := DefaultOptions()
	server := NewServer(t, opts)
	defer server.Shutdown()

	// Verify the server is running
	if server.URL == "" {
		t.Fatal("Server URL should not be empty")
	}

	// Verify components are wired
	if server.RadioManager == nil {
		t.Fatal("RadioManager should not be nil")
	}
	if server.Orchestrator == nil {
		t.Fatal("Orchestrator should not be nil")
	}
	if server.TelemetryHub == nil {
		t.Fatal("TelemetryHub should not be nil")
	}
	if server.AuditLogger == nil {
		t.Fatal("AuditLogger should not be nil")
	}
	if server.SilvusAdapter == nil {
		t.Fatal("SilvusAdapter should not be nil")
	}

	t.Logf("Harness created successfully with URL: %s", server.URL)
}
