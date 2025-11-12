# RCC Web UI

A desktop-first, single-page Web UI for Radio Control Container (RCC) implementing OpenAPI v1 and Telemetry SSE v1 with CB-TIMING v0.3 conformance. Supports multiple radio types with abstract channel selection.

## 📁 Project Location
```
RadioControlContainer/web-ui/
```

## 🎯 Quick Start
```bash
cd RadioControlContainer/web-ui
./run-tests.sh
```

## Features

- **Radio Selection**: List and select active radio from available radios
- **Power Control**: Set power level (0-39 dBm) with real-time feedback
- **Channel Control**: Set channel by abstract index (1,2,3...) with frequency display for reference
- **Live Telemetry**: Real-time SSE stream with event resume and buffering
- **Error Handling**: Normalized error codes with CB-TIMING backoff policies
- **Audit Logging**: Structured logs with correlation ID passthrough

## Architecture Compliance

- **Gate A (OpenAPI parity)**: All endpoints match OpenAPI v1 specification
- **Gate B (Telemetry parity)**: SSE events match Telemetry SSE v1 specification  
- **Gate C (Timing policy)**: All timing values from CB-TIMING v0.3 config
- **Gate D (Indexing)**: Channel indices are 1-based throughout UI

## Quick Start

### Prerequisites

- Go 1.21+ 
- RCC container running on `http://localhost:8080`

### Installation

1. Clone or download the RCC Web UI files
2. Ensure `config.json` is present with RCC base URL
3. Run the server:

```bash
go run main.go
```

4. Open browser to `http://127.0.0.1:3000`

### Configuration

The `config.json` file contains timing parameters from CB-TIMING v0.3:

```json
{
  "rccBaseUrl": "http://localhost:8080",
  "timing": {
    "heartbeatIntervalSec": 15,
    "heartbeatTimeoutSec": 45,
    "probeNormalSec": 30,
    "probeRecoveringMinSec": 5,
    "probeRecoveringMaxSec": 15,
    "probeOfflineMinSec": 10,
    "probeOfflineMaxSec": 300,
    "cmdTimeoutsSec": {
      "setPower": 10,
      "setChannel": 30,
      "selectRadio": 5,
      "getState": 5
    },
    "retry": {
      "busyBaseMs": 1000,
      "unavailableBaseMs": 2000,
      "jitterMs": 200
    }
  }
}
```

## API Integration

### OpenAPI v1 Endpoints

- `GET /radios` - List available radios and active radio
- `POST /radios/select` - Select active radio
- `GET /radios/{id}/power` - Get current power level
- `POST /radios/{id}/power` - Set power level
- `GET /radios/{id}/channel` - Get current channel/frequency
- `POST /radios/{id}/channel` - Set channel by abstract index (1,2,3...)

### Response Envelopes

**Success Response:**
```json
{
  "result": "ok",
  "data": { ... },
  "correlationId": "uuid"
}
```

**Error Response:**
```json
{
  "result": "error", 
  "code": "INVALID_RANGE|BUSY|UNAVAILABLE|INTERNAL",
  "message": "Human readable error",
  "details": { ... },
  "correlationId": "uuid"
}
```

### Telemetry SSE v1 Events

- `ready` - System ready event
- `state` - Radio state change (online/recovering/offline)
- `powerChanged` - Power level change
- `channelChanged` - Channel/frequency change  
- `fault` - Fault condition
- `heartbeat` - Periodic heartbeat

## Testing

### Manual Testing with curl

```bash
# Test radio listing
curl -X GET http://localhost:3000/radios

# Test radio selection  
curl -X POST http://localhost:3000/radios/select \
  -H "Content-Type: application/json" \
  -d '{"id":"radio-01"}'

# Test power setting
curl -X POST http://localhost:3000/radios/radio-01/power \
  -H "Content-Type: application/json" \
  -d '{"powerDbm":30}'

# Test channel setting by abstract index (1,2,3...)
curl -X POST http://localhost:3000/radios/radio-01/channel \
  -H "Content-Type: application/json" \
  -d '{"channelIndex":6}'

# Test telemetry stream
curl -N http://localhost:3000/telemetry
```

### Fake vs Real Adapter Testing

1. **Fake Adapter**: Set RCC container to use fake adapter for testing
2. **Real Adapter**: Set RCC container to use real radio hardware
3. Toggle between adapters in RCC container configuration
4. UI automatically adapts to available radios and capabilities

## IV&V Acceptance Checklist

### Radio Management
- [ ] `GET /radios` populates radio list with active radio highlighted
- [ ] `POST /radios/select` updates active radio selection
- [ ] Radio status indicator shows online/recovering/offline states

### Power Control  
- [ ] `GET /radios/{id}/power` displays current power level
- [ ] `POST /radios/{id}/power` sets power with validation
- [ ] Power slider reflects current value and updates on changes
- [ ] Power changes trigger `powerChanged` telemetry events

### Channel Control
- [ ] `GET /radios/{id}/channel` displays current frequency and index
- [ ] `POST /radios/{id}/channel` accepts abstract channel index (1,2,3...)
- [ ] Channel selection uses abstract numbers, frequency shown for reference only
- [ ] Channel indices displayed as 1-based throughout UI
- [ ] Channel changes trigger `channelChanged` telemetry events

### Telemetry Integration
- [ ] SSE connection establishes and receives `ready` event
- [ ] `state` events update radio status indicator
- [ ] `powerChanged` and `channelChanged` events update UI
- [ ] `fault` events display error toasts
- [ ] `heartbeat` events logged to telemetry display
- [ ] Resume with `Last-Event-ID` works after reconnection
- [ ] Telemetry log shows last 50 events with timestamps

### Error Handling
- [ ] `INVALID_RANGE` errors display exact message from API
- [ ] `BUSY` errors trigger retry with CB-TIMING backoff
- [ ] `UNAVAILABLE` errors trigger retry with CB-TIMING backoff  
- [ ] `INTERNAL` errors display generic error message
- [ ] No polling loops - only event-driven updates

### Timing & Backoff
- [ ] All timing values loaded from CB-TIMING config
- [ ] No hardcoded timeouts or delays in code
- [ ] Retry backoff includes jitter per CB-TIMING
- [ ] Command timeouts enforced per CB-TIMING values

### Audit & Logging
- [ ] All API calls logged with correlation ID
- [ ] Audit entries include timestamp, actor, radioId, action, result, latency
- [ ] Client audit logs sent to server `/audit` endpoint
- [ ] Console logging for debugging and monitoring

## Troubleshooting

### Common Issues

**"Failed to connect to RCC"**
- Verify RCC container is running on `http://localhost:8080`
- Check `config.json` has correct `rccBaseUrl`
- Ensure no firewall blocking localhost connections

**"No radios available"**
- RCC container may not have radios configured
- Check RCC container logs for radio discovery issues
- Verify adapter configuration (Fake vs Real)

**"Telemetry connection error"**
- SSE endpoint may not be available on RCC
- Check RCC container supports `/telemetry` endpoint
- Verify CORS settings if accessing from different origin

**"Power/Channel setting fails"**
- Radio may be offline or busy
- Check radio capabilities and power limits
- Verify channel frequency is within radio's supported range

### Debug Mode

Enable debug logging by opening browser console. All API calls, telemetry events, and audit logs are displayed with full details.

### Log Files

- Server logs: Console output from `main.go`
- Audit logs: `audit.log` file (if enabled)
- Client logs: Browser console

## Security Notes

- UI binds to localhost only (`127.0.0.1:3000`)
- No authentication required for this environment
- All secrets excluded from audit logs
- CORS handled via reverse proxy to avoid browser restrictions

## Performance

- 95th percentile command RTT < 100ms (local link)
- SSE reconnect within 5 seconds
- Event-driven updates only (no polling)
- Telemetry log limited to 50 entries for memory efficiency

## Browser Compatibility

- Modern browsers with EventSource support
- Keyboard navigation support
- Responsive design for desktop-first usage
- Accessibility features for screen readers