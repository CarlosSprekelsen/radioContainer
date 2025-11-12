//
//
package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/radio-control/rcc/internal/auth"
)

// RegisterRoutes registers all OpenAPI v1 endpoints.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// API v1 base path
	apiV1 := "/api/v1"

	// Health endpoint (no auth required)
	mux.HandleFunc(apiV1+"/health", s.handleHealth)

	// If no auth middleware, register routes without protection
	if s.authMiddleware == nil {
		// Capabilities endpoint
		mux.HandleFunc(apiV1+"/capabilities", s.handleCapabilities)

		// Radios endpoints
		mux.HandleFunc(apiV1+"/radios", s.handleRadios)
		mux.HandleFunc(apiV1+"/radios/select", s.handleSelectRadio)

		// Radio-specific endpoints (power, channel, individual radio)
		mux.HandleFunc(apiV1+"/radios/", s.handleRadioEndpoints)

		// Telemetry endpoint
		mux.HandleFunc(apiV1+"/telemetry", s.handleTelemetry)
		return
	}

	// Register routes with authentication and authorization
	// Capabilities endpoint (viewer access)
	mux.HandleFunc(apiV1+"/capabilities", s.authMiddleware.RequireAuth(s.authMiddleware.RequireScope(auth.ScopeRead)(s.handleCapabilities)))

	// Radios endpoints (viewer access)
	mux.HandleFunc(apiV1+"/radios", s.authMiddleware.RequireAuth(s.authMiddleware.RequireScope(auth.ScopeRead)(s.handleRadios)))

	// Select radio endpoint (controller access)
	mux.HandleFunc(apiV1+"/radios/select", s.authMiddleware.RequireAuth(s.authMiddleware.RequireScope(auth.ScopeControl)(s.handleSelectRadio)))

	// Radio-specific endpoints (power, channel, individual radio)
	mux.HandleFunc(apiV1+"/radios/", s.handleRadioEndpoints)

	// Telemetry endpoint (viewer access)
	mux.HandleFunc(apiV1+"/telemetry", s.authMiddleware.RequireAuth(s.authMiddleware.RequireScope(auth.ScopeTelemetry)(s.handleTelemetry)))
}

// handleCapabilities handles GET /capabilities
func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// Return capabilities
	capabilities := map[string]interface{}{
		"telemetry": []string{"sse"},
		"commands":  []string{"http-json"},
		"version":   "1.0.0",
	}

	WriteSuccess(w, capabilities)
}

// handleRadios handles GET /radios
func (s *Server) handleRadios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// Fetch radios from RadioManager
	if s.radioManager == nil {
		WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE",
			"Radio manager not available", nil)
		return
	}

	list := s.radioManager.List()
	WriteSuccess(w, list)
}

// handleSelectRadio handles POST /radios/select
func (s *Server) handleSelectRadio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only POST method is allowed", nil)
		return
	}

	// Parse request (strict JSON)
	var req struct {
		RadioID string `json:"radioId"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON or unknown fields", nil)
		return
	}
	// Trailing data check
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Trailing data after JSON object", nil)
		return
	}

	// Ensure services available
	if s.radioManager == nil || s.orchestrator == nil {
		WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
		return
	}

	// Call orchestrator to confirm selection (ping adapter/state)
	if err := s.orchestrator.SelectRadio(r.Context(), req.RadioID); err != nil {
		status, body := ToAPIError(err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write(body)
		return
	}

	WriteSuccess(w, map[string]string{"activeRadioId": req.RadioID})
}

// handleRadioEndpoints handles all radio-specific endpoints.
// Routes to appropriate handler based on path.
func (s *Server) handleRadioEndpoints(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Extract radio ID and determine endpoint type
	radioID := s.extractRadioID(path)
	if radioID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Radio ID is required", nil)
		return
	}

	// Apply authentication and authorization based on endpoint type
	if s.authMiddleware != nil {
		// Route based on path suffix with appropriate auth
		if strings.HasSuffix(path, "/power") {
			if r.Method == http.MethodGet {
				// GET power requires read scope
				s.authMiddleware.RequireAuth(s.authMiddleware.RequireScope(auth.ScopeRead)(s.handleRadioPower))(w, r)
			} else if r.Method == http.MethodPost {
				// POST power requires control scope
				s.authMiddleware.RequireAuth(s.authMiddleware.RequireScope(auth.ScopeControl)(s.handleRadioPower))(w, r)
			} else {
				s.handleRadioPower(w, r)
			}
		} else if strings.HasSuffix(path, "/channel") {
			if r.Method == http.MethodGet {
				// GET channel requires read scope
				s.authMiddleware.RequireAuth(s.authMiddleware.RequireScope(auth.ScopeRead)(s.handleRadioChannel))(w, r)
			} else if r.Method == http.MethodPost {
				// POST channel requires control scope
				s.authMiddleware.RequireAuth(s.authMiddleware.RequireScope(auth.ScopeControl)(s.handleRadioChannel))(w, r)
			} else {
				s.handleRadioChannel(w, r)
			}
		} else {
			// Individual radio endpoint requires read scope
			s.authMiddleware.RequireAuth(s.authMiddleware.RequireScope(auth.ScopeRead)(s.handleRadioByID))(w, r)
		}
	} else {
		// No auth middleware, route directly
		if strings.HasSuffix(path, "/power") {
			s.handleRadioPower(w, r)
		} else if strings.HasSuffix(path, "/channel") {
			s.handleRadioChannel(w, r)
		} else {
			// Default to individual radio endpoint
			s.handleRadioByID(w, r)
		}
	}
}

// handleRadioByID handles GET /radios/{id}
func (s *Server) handleRadioByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// Extract radio ID from path
	radioID := s.extractRadioID(r.URL.Path)
	if radioID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Radio ID is required", nil)
		return
	}

	if s.radioManager == nil {
		WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE",
			"Radio manager not available", nil)
		return
	}

	radio, err := s.radioManager.GetRadio(radioID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Radio not found", nil)
		return
	}

	WriteSuccess(w, radio)
}

// handleRadioPower handles GET/POST /radios/{id}/power
func (s *Server) handleRadioPower(w http.ResponseWriter, r *http.Request) {
	// Extract radio ID from path
	radioID := s.extractRadioID(r.URL.Path)
	if radioID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Radio ID is required", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetPower(w, r, radioID)
	case http.MethodPost:
		s.handleSetPower(w, r, radioID)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET and POST methods are allowed", nil)
	}
}

// handleGetPower handles GET /radios/{id}/power
func (s *Server) handleGetPower(w http.ResponseWriter, r *http.Request, radioID string) {
	if s.orchestrator == nil {
		WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
		return
	}
	state, err := s.orchestrator.GetState(r.Context(), radioID)
	if err != nil {
		status, body := ToAPIError(err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write(body)
		return
	}
	WriteSuccess(w, map[string]interface{}{"powerDbm": state.PowerDbm})
}

// handleSetPower handles POST /radios/{id}/power
func (s *Server) handleSetPower(w http.ResponseWriter, r *http.Request, radioID string) {
	// Parse request body (strict JSON)
	var request struct {
		PowerDbm float64 `json:"powerDbm"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST",
			"Malformed JSON or unknown fields", nil)
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Trailing data after JSON object", nil)
		return
	}

	if s.orchestrator == nil {
		WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
		return
	}
	if err := s.orchestrator.SetPower(r.Context(), radioID, request.PowerDbm); err != nil {
		status, body := ToAPIError(err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write(body)
		return
	}
	WriteSuccess(w, map[string]interface{}{"powerDbm": request.PowerDbm})
}

// handleRadioChannel handles GET/POST /radios/{id}/channel
func (s *Server) handleRadioChannel(w http.ResponseWriter, r *http.Request) {
	// Extract radio ID from path
	radioID := s.extractRadioID(r.URL.Path)
	if radioID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_RANGE",
			"Radio ID is required", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetChannel(w, r, radioID)
	case http.MethodPost:
		s.handleSetChannel(w, r, radioID)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET and POST methods are allowed", nil)
	}
}

// handleGetChannel handles GET /radios/{id}/channel
func (s *Server) handleGetChannel(w http.ResponseWriter, r *http.Request, radioID string) {
	if s.orchestrator == nil {
		WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
		return
	}
	state, err := s.orchestrator.GetState(r.Context(), radioID)
	if err != nil {
		status, body := ToAPIError(err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write(body)
		return
	}
	// channelIndex may be null if not in derived set; we return frequency
	WriteSuccess(w, map[string]interface{}{"frequencyMhz": state.FrequencyMhz, "channelIndex": nil})
}

// handleSetChannel handles POST /radios/{id}/channel
func (s *Server) handleSetChannel(w http.ResponseWriter, r *http.Request, radioID string) {
	// Parse request body (strict JSON)
	var request struct {
		ChannelIndex *int     `json:"channelIndex,omitempty"`
		FrequencyMhz *float64 `json:"frequencyMhz,omitempty"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST",
			"Malformed JSON or unknown fields", nil)
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Trailing data after JSON object", nil)
		return
	}

	// Validate that at least one parameter is provided (structural)
	if request.ChannelIndex == nil && request.FrequencyMhz == nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST",
			"Either channelIndex or frequencyMhz must be provided", nil)
		return
	}

	if s.orchestrator == nil {
		WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "Service not available", nil)
		return
	}

	// Frequency wins if both provided
	if request.FrequencyMhz != nil {
		if err := s.orchestrator.SetChannel(r.Context(), radioID, *request.FrequencyMhz); err != nil {
			status, body := ToAPIError(err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(status)
			_, _ = w.Write(body)
			return
		}
		WriteSuccess(w, map[string]interface{}{"frequencyMhz": *request.FrequencyMhz, "channelIndex": request.ChannelIndex})
		return
	}

	// If only index provided, use SetChannelByIndex method
	if request.ChannelIndex != nil {
		if err := s.orchestrator.SetChannelByIndex(r.Context(), radioID, *request.ChannelIndex, s.radioManager); err != nil {
			status, body := ToAPIError(err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(status)
			_, _ = w.Write(body)
			return
		}
		WriteSuccess(w, map[string]interface{}{"frequencyMhz": nil, "channelIndex": *request.ChannelIndex})
		return
	}
}

// handleTelemetry handles GET /telemetry (SSE)
func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// Wire to Telemetry Hub Subscribe
	if s.telemetryHub == nil {
		WriteError(w, http.StatusServiceUnavailable, "UNAVAILABLE",
			"Telemetry service not available", nil)
		return
	}

	// Subscribe to telemetry stream
	ctx := r.Context()
	if err := s.telemetryHub.Subscribe(ctx, w, r); err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL",
			"Failed to subscribe to telemetry stream", nil)
		return
	}
}

// handleHealth handles GET /health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED",
			"Only GET method is allowed", nil)
		return
	}

	// Calculate uptime
	uptime := 0.0
	if !s.startTime.IsZero() {
		uptime = time.Since(s.startTime).Seconds()
	}

	// Check subsystem health
	subsystems := s.checkSubsystemHealth()

	// Determine overall health status
	overallStatus := "ok"
	if !subsystems["telemetry"] || !subsystems["orchestrator"] || !subsystems["radioManager"] {
		overallStatus = "degraded"
	}

	health := map[string]interface{}{
		"status":     overallStatus,
		"uptimeSec":  uptime,
		"version":    "1.0.0",
		"subsystems": subsystems,
	}

	// Return appropriate HTTP status based on health
	if overallStatus == "ok" {
		WriteSuccess(w, health)
	} else {
		// Return 503 Service Unavailable for degraded health
		// Pass health data as details so it's available in the error response
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_DEGRADED",
			"One or more subsystems are unavailable", health)
	}
}

// checkSubsystemHealth checks the health of all subsystems.
func (s *Server) checkSubsystemHealth() map[string]bool {
	subsystems := make(map[string]bool)

	// Check telemetry hub
	subsystems["telemetry"] = s.telemetryHub != nil

	// Check orchestrator
	subsystems["orchestrator"] = s.orchestrator != nil

	// Check radio manager
	subsystems["radioManager"] = s.radioManager != nil

	// Check auth middleware (optional, so always true if not required)
	subsystems["auth"] = true // Auth is optional, so always considered healthy

	return subsystems
}

// extractRadioID extracts the radio ID from a URL path.
// Handles paths like /api/v1/radios/{id}/power, /api/v1/radios/{id}/channel, etc.
func (s *Server) extractRadioID(path string) string {
	// Remove /api/v1/radios/ prefix
	prefix := "/api/v1/radios/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	// Get the part after the prefix
	remaining := path[len(prefix):]

	// Split by '/' to get the radio ID (first part)
	parts := strings.Split(remaining, "/")
	if len(parts) == 0 {
		return ""
	}

	radioID := parts[0]
	if radioID == "" {
		return ""
	}

	return radioID
}

// parseRadioIDFromPath is a helper to parse radio ID from various path patterns.
func parseRadioIDFromPath(path string) string {
	// Handle different path patterns:
	// /api/v1/radios/{id}
	// /api/v1/radios/{id}/power
	// /api/v1/radios/{id}/channel

	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[1] != "api" || parts[2] != "v1" || parts[3] != "radios" {
		return ""
	}

	if len(parts) < 5 {
		return ""
	}

	return parts[4]
}
