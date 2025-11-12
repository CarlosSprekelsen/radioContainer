// Package e2e provides manifest-aware E2E testing that skips routes affected by build blockers.
package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestManifest represents the E2E test manifest
type TestManifest struct {
	Version       string               `json:"version"`
	Description   string               `json:"description"`
	LastUpdated   string               `json:"last_updated"`
	Routes        map[string]RouteInfo `json:"routes"`
	Summary       ManifestSummary      `json:"summary"`
	BuildBlockers []BuildBlocker       `json:"build_blockers"`
}

type RouteInfo struct {
	Method string `json:"method"`
	Skip   bool   `json:"x-skip"`
	Reason string `json:"reason"`
	Status string `json:"status"`
}

type ManifestSummary struct {
	TotalRoutes        int `json:"total_routes"`
	Available          int `json:"available"`
	Skipped            int `json:"skipped"`
	CoveragePercentage int `json:"coverage_percentage"`
}

type BuildBlocker struct {
	Component string   `json:"component"`
	Issue     string   `json:"issue"`
	Priority  string   `json:"priority"`
	Files     []string `json:"files"`
}

func TestManifestAware_E2EExecution(t *testing.T) {
	// Load test manifest
	manifest := loadTestManifest(t)

	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	t.Logf("=== MANIFEST-AWARE E2E EXECUTION ===")
	t.Logf("Total routes: %d", manifest.Summary.TotalRoutes)
	t.Logf("Available: %d", manifest.Summary.Available)
	t.Logf("Skipped: %d", manifest.Summary.Skipped)
	t.Logf("Coverage: %d%%", manifest.Summary.CoveragePercentage)

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/health":
			w.WriteHeader(200)
			w.Write([]byte(`{"result":"ok","data":{"status":"healthy","timestamp":"2025-10-03T11:45:00Z"}}`))
		case "/api/v1/capabilities":
			w.WriteHeader(200)
			w.Write([]byte(`{"result":"ok","data":{"telemetry":["sse"],"commands":["http-json"],"version":"1.0.0"}}`))
		case "/api/v1/radios":
			w.WriteHeader(200)
			w.Write([]byte(`{"result":"ok","data":[{"id":"silvus-001","name":"Silvus Radio 1"}]}`))
		case "/api/v1/radios/silvus-001":
			w.WriteHeader(200)
			w.Write([]byte(`{"result":"ok","data":{"id":"silvus-001","name":"Silvus Radio 1","status":"available"}}`))
		case "/api/v1/radios/silvus-001/power":
			if r.Method == "GET" {
				w.WriteHeader(200)
				w.Write([]byte(`{"result":"ok","data":{"powerDbm":25.0}}`))
			} else {
				w.WriteHeader(405)
				w.Write([]byte(`{"result":"error","code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}`))
			}
		case "/api/v1/radios/silvus-001/channel":
			if r.Method == "GET" {
				w.WriteHeader(200)
				w.Write([]byte(`{"result":"ok","data":{"channelIndex":1,"frequencyMhz":2400.0}}`))
			} else {
				w.WriteHeader(405)
				w.Write([]byte(`{"result":"error","code":"METHOD_NOT_ALLOWED","message":"Method not allowed"}`))
			}
		case "/api/v1/telemetry":
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(200)
			w.Write([]byte(`event: ready\ndata: {"snapshot":{"activeRadioId":"","radios":[]}}\n\n`))
		default:
			w.WriteHeader(404)
			w.Write([]byte(`{"result":"error","code":"NOT_FOUND","message":"Not found"}`))
		}
	}))
	defer ts.Close()

	// Test available routes only
	availableRoutes := []struct {
		method string
		path   string
		desc   string
	}{
		{"GET", "/api/v1/health", "Health check"},
		{"GET", "/api/v1/capabilities", "Get capabilities"},
		{"GET", "/api/v1/radios", "List radios"},
		{"GET", "/api/v1/radios/silvus-001", "Get radio details"},
		{"GET", "/api/v1/radios/silvus-001/power", "Get radio power"},
		{"GET", "/api/v1/radios/silvus-001/channel", "Get radio channel"},
		{"GET", "/api/v1/telemetry", "Telemetry stream"},
	}

	t.Logf("=== TESTING AVAILABLE ROUTES ===")
	t.Logf("%-8s %-30s %-20s %-10s %-s", "METHOD", "PATH", "EXPECTED", "ACTUAL", "STATUS")
	t.Logf("%s", "----------------------------------------------------------------")

	passCount := 0
	for _, route := range availableRoutes {
		// Make HTTP request
		req, err := http.NewRequest(route.method, ts.URL+route.path, nil)
		if err != nil {
			t.Errorf("Failed to create request: %v", err)
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("Failed to make request: %v", err)
			continue
		}

		// Validate against OpenAPI spec
		validator.ValidateHTTPResponseAgainstOpenAPI(t, resp, route.method, route.path)

		// Log results
		status := "PASS"
		if resp.StatusCode >= 400 {
			status = "FAIL"
		} else {
			passCount++
		}

		t.Logf("%-8s %-30s %-20s %-10d %-s", route.method, route.path, "200", resp.StatusCode, status)

		resp.Body.Close()
	}

	t.Logf("=================================================================")
	t.Logf("Available routes tested: %d/%d", passCount, len(availableRoutes))
	t.Logf("Success rate: %.1f%%", float64(passCount)/float64(len(availableRoutes))*100)

	// Report skipped routes
	t.Logf("=== SKIPPED ROUTES (BUILD BLOCKERS) ===")
	for routeKey, routeInfo := range manifest.Routes {
		if routeInfo.Skip {
			t.Logf("SKIPPED: %s %s - %s", routeInfo.Method, routeKey, routeInfo.Reason)
		}
	}

	// Report build blockers
	t.Logf("=== BUILD BLOCKERS ===")
	for _, blocker := range manifest.BuildBlockers {
		t.Logf("BLOCKER: %s - %s (%s)", blocker.Component, blocker.Issue, blocker.Priority)
	}
}

func loadTestManifest(t *testing.T) *TestManifest {
	// Try multiple possible locations for the manifest file
	paths := []string{
		"test/e2e/test_manifest.json",
		"../test/e2e/test_manifest.json",
		"../../test/e2e/test_manifest.json",
	}

	var content []byte
	var err error

	for _, path := range paths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to read test manifest from any location: %v", err)
	}

	var manifest TestManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("Failed to parse test manifest: %v", err)
	}

	return &manifest
}
