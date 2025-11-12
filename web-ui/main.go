package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Config represents the application configuration
type Config struct {
	RCCBaseURL string `json:"rccBaseUrl"`
	Timing     struct {
		HeartbeatIntervalSec  int `json:"heartbeatIntervalSec"`
		HeartbeatTimeoutSec   int `json:"heartbeatTimeoutSec"`
		ProbeNormalSec        int `json:"probeNormalSec"`
		ProbeRecoveringMinSec int `json:"probeRecoveringMinSec"`
		ProbeRecoveringMaxSec int `json:"probeRecoveringMaxSec"`
		ProbeOfflineMinSec    int `json:"probeOfflineMinSec"`
		ProbeOfflineMaxSec    int `json:"probeOfflineMaxSec"`
		CmdTimeoutsSec        struct {
			SetPower    int `json:"setPower"`
			SetChannel  int `json:"setChannel"`
			SelectRadio int `json:"selectRadio"`
			GetState    int `json:"getState"`
		} `json:"cmdTimeoutsSec"`
		Retry struct {
			BusyBaseMs        int `json:"busyBaseMs"`
			UnavailableBaseMs int `json:"unavailableBaseMs"`
			JitterMs          int `json:"jitterMs"`
		} `json:"retry"`
	} `json:"timing"`
}

// AuditEntry represents a structured audit log entry
type AuditEntry struct {
	Timestamp     time.Time `json:"timestamp"`
	Actor         string    `json:"actor"`
	RadioID       string    `json:"radioId"`
	Action        string    `json:"action"`
	Result        string    `json:"result"`
	LatencyMS     int64     `json:"latencyMs"`
	CorrelationID string    `json:"correlationId"`
}

var config Config

func loadConfig() error {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("failed to read config.json: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config.json: %w", err)
	}

	return nil
}

func logAudit(entry AuditEntry) {
	// Log to console
	log.Printf("AUDIT: %+v", entry)

	// Log to file via /audit endpoint
	jsonData, _ := json.Marshal(entry)
	resp, err := http.Post("http://localhost:8080/audit", "application/json", strings.NewReader(string(jsonData)))
	if err == nil {
		resp.Body.Close()
	}
}

func reverseProxy(w http.ResponseWriter, r *http.Request) {
	// Build target URL
	targetURL := config.RCCBaseURL + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Create request to RCC
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Handle SSE with Last-Event-ID injection
	if strings.HasPrefix(r.URL.Path, "/telemetry") {
		lastEventID := r.URL.Query().Get("lastEventId")
		if lastEventID != "" {
			req.Header.Set("Last-Event-ID", lastEventID)
		}
	}

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to connect to RCC", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status
	w.WriteHeader(resp.StatusCode)

	// Copy body
	io.Copy(w, resp.Body)
}

func handleAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var entry AuditEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Append to audit log file
	file, err := os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Failed to open audit log", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	jsonData, _ := json.Marshal(entry)
	file.Write(append(jsonData, '\n'))

	w.WriteHeader(http.StatusOK)
}

func main() {
	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Serve config.json
	http.HandleFunc("/config.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "config.json")
	})

	// API reverse proxy routes
	http.HandleFunc("/radios", reverseProxy)
	http.HandleFunc("/radios/", reverseProxy)
	http.HandleFunc("/telemetry", reverseProxy)

	// Audit endpoint
	http.HandleFunc("/audit", handleAudit)

	// Start server
	log.Println("RCC Web UI server starting on http://0.0.0.0:3000")
	log.Printf("Proxying to RCC at %s", config.RCCBaseURL)

	if err := http.ListenAndServe("0.0.0.0:3000", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
