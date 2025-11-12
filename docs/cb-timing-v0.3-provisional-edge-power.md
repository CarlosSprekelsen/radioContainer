# CB-TIMING v0.3 (Provisional – Edge Power)

> **Document ID**: CB-TIMING-v0.3  
> **Version**: 0.3 (Provisional)  
> **Date**: 2025-01-15  
> **Classification**: Internal  
> **Scope**: Single source of truth for all timing, cadence, and buffer parameters in the Radio Control Container system.

---

## 1. Document Control

- **Revision History**

  | Version | Date       | Author            | Changes                                                                                         |
  | ------- | ---------- | ----------------- | ----------------------------------------------------------------------------------------------- |
  | 0.3     | 2025-01-15 | System Architect  | Initial baseline for edge power optimization; all numeric values centralized here. |

- **Approval Signatures**

  | Role               | Name  | Signature | Date |
  | ------------------ | ----- | --------- | ---- |
  | System Architect   | [TBD] |           |      |
  | IV&V Lead          | [TBD] |           |      |
  | Product Owner      | [TBD] |           |      |

---

## 2. Purpose & Scope

This document defines the **single source of truth** for all timing, cadence, buffer, and tolerance parameters used across the Radio Control Container documentation suite. All other documents must reference this baseline by name and version.

**Referenced by:**
- Architecture Document (§8.3, §8.3a)
- OpenAPI v1 Specification (§2.2, §3.9)
- Telemetry SSE v1 Specification (§1.3, §4)
- ICD Logical Interface (§10)

---

## 3. Telemetry & Heartbeat Parameters

### 3.1 Heartbeat Configuration
- **Heartbeat Interval**: 15 seconds (idle cap)
- **Heartbeat Jitter**: ±2 seconds
- **Heartbeat Timeout**: 45 seconds (3x interval)

### 3.2 State Update Cadence
- **Change-driven**: Immediate on state change
- **Background tick**: 1 Hz maximum when idle
- **Burst protection**: Maximum 10 events/second per radio

---

## 4. Health Probe Parameters

### 4.1 Probe States & Cadences

| State | Initial Interval | Backoff Factor | Max Interval | Jitter Window |
|-------|------------------|----------------|-------------|---------------|
| **Normal** | 30 seconds | 1.0 (no backoff) | 30 seconds | ±5 seconds |
| **Recovering** | 5 seconds | 1.5 | 15 seconds | ±2 seconds |
| **Offline** | 10 seconds | 2.0 | 300 seconds | ±10 seconds |

### 4.2 Probe Budget Limits
- **Maximum probes per radio**: 100 per hour
- **Probe timeout**: 5 seconds
- **Concurrent probe limit**: 3 radios maximum

---

## 5. Command Timeout Classes

| Command Class | Timeout | Retry Count | Backoff Base |
|---------------|---------|-------------|--------------|
| **setPower** | 10 seconds | 3 | 1 second |
| **setChannel** | 30 seconds | 2 | 2 seconds |
| **selectRadio** | 5 seconds | 3 | 500ms |
| **getState** | 5 seconds | 2 | 1 second |

---

## 6. Event Replay & Buffering

### 6.1 Event Buffer Configuration
- **Buffer size per radio**: 50 events
- **Buffer retention**: 1 hour
- **Event ID scheme**: Monotonic per radio, starting from 1
- **Resume support**: Last-Event-ID header with up to 50 events replay

### 6.2 Recovering Blackout Period
- **Channel change blackout**: 30 seconds
- **Power change blackout**: 5 seconds
- **Radio reset blackout**: 60 seconds

---

## 7. Frequency Validation Parameters

### 7.1 Tolerance Settings
- **Frequency tolerance**: ±0.1 MHz
- **Channel index validation**: Strict (must be 1-based)
- **Range validation**: Must be within supported_frequency_profiles

### 7.2 Channel Mapping
- **Index base**: 1-based (not 0-based)
- **Frequency precedence**: When both frequencyMhz and channelIndex provided, frequency takes precedence
- **Refresh trigger**: On capability change or region config update

---

## 8. Backoff & Retry Policies

### 8.1 Standard Backoff
- **Base delay**: 500ms
- **Jitter range**: ±250ms
- **Max retries**: 5
- **Exponential factor**: 2.0
- **Max delay**: 30 seconds

### 8.2 Error-Specific Policies

| Error Code | Initial Delay | Max Retries | Backoff Factor |
|------------|---------------|-------------|----------------|
| **BUSY** | 1 second | 3 | 1.5 |
| **UNAVAILABLE** | 2 seconds | 5 | 2.0 |
| **INTERNAL** | 500ms | 3 | 2.0 |

---

## 9. Power Management Parameters

### 9.1 Edge Power Optimization
- **CPU idle target**: <5%
- **CPU burst limit**: <20% (10 commands/min + 1Hz telemetry)
- **Probe budget**: Maximum 100 probes per radio per hour
- **Event-first policy**: Change-driven telemetry only

### 9.2 Duty Cycle Limits
- **Normal state**: 1 probe per 30 seconds
- **Recovering state**: 1 probe per 5-15 seconds (backoff)
- **Offline state**: 1 probe per 10-300 seconds (exponential backoff)

---

## 10. Network & Transport Parameters

### 10.1 Connection Management
- **SSE connection timeout**: 60 seconds
- **HTTP request timeout**: 30 seconds
- **Connection pool size**: 10 per radio
- **Keep-alive interval**: 30 seconds

### 10.2 Rate Limiting
- **Commands per minute**: 60 per radio
- **Telemetry events per second**: 10 per radio
- **Concurrent SSE clients**: 5 per radio

### 10.3 Log Rotation Parameters
- **log_max_file_mb**: 10 MB (maximum size per log file)
- **log_keep_files**: 5 files (number of rotated files to keep)
- **log_rotation_trigger**: Size-based (rotate when max size reached)
- **audit_log_retention_days**: 30 days (security event log retention)
- **telemetry_buffer_hours**: 1 hour (real-time telemetry retention)

---

## 11. Validation Rules

### 11.1 Parameter Validation
- All timing values must be positive integers (seconds) or positive floats (intervals)
- Backoff factors must be ≥ 1.0
- Jitter must be ≤ 50% of base interval
- Max intervals must be ≥ initial intervals

### 11.2 Consistency Checks
- Probe intervals must align with heartbeat intervals
- Command timeouts must be ≥ probe timeouts
- Buffer sizes must accommodate expected event rates
- Backoff policies must prevent resource exhaustion

---

## 12. Tuning Guidelines

### 12.1 Performance Tuning
- **High-latency networks**: Increase timeouts by 2x
- **Low-power mode**: Increase intervals by 3x
- **High-reliability mode**: Decrease intervals by 0.5x

### 12.2 Monitoring Thresholds
- **Probe success rate**: >95%
- **Command success rate**: >98%
- **Event delivery rate**: >99%
- **Buffer utilization**: <80%

---

## 13. Compliance Requirements

### 13.1 Documentation References
All documents referencing timing parameters must:
- Reference this document by name and version
- Not hardcode numeric values
- Use parameter names exactly as defined here
- Update references when this document changes

### 13.2 Implementation Requirements
- All timing parameters must be configurable at runtime
- Default values must match this baseline
- Parameter validation must enforce ranges defined here
- Logging must include parameter values for debugging

---

## 14. Change Management

### 14.1 Version Control
- **Major version**: Breaking changes to parameter structure
- **Minor version**: New parameters or expanded ranges
- **Patch version**: Bug fixes or clarifications

### 14.2 Impact Assessment
Changes to this document require:
- Impact analysis on all referencing documents
- Consistency matrix update
- Implementation testing
- Documentation update coordination

---

## 15. Annex A - Parameter Summary Table

| Parameter Category | Count | Key Parameters |
|-------------------|-------|----------------|
| **Telemetry** | 4 | Heartbeat interval, state cadence, burst protection |
| **Health Probes** | 12 | 3 states × 4 parameters each |
| **Command Timeouts** | 12 | 4 command classes × 3 parameters each |
| **Event Buffering** | 4 | Buffer size, retention, ID scheme, resume |
| **Frequency Validation** | 3 | Tolerance, indexing, precedence |
| **Backoff Policies** | 8 | Standard + error-specific policies |
| **Power Management** | 6 | CPU targets, duty cycles, probe budgets |
| **Network** | 6 | Timeouts, pools, rate limits |
| **Log Rotation** | 5 | File size, retention, rotation triggers |

**Total Parameters**: 60 timing, cadence, and logging parameters

---

## 16. Annex B - Cross-Reference Matrix

| Document | Section | Parameters Referenced |
|----------|---------|----------------------|
| **Architecture** | §8.3 | Heartbeat, probe cadences, command timeouts |
| **Architecture** | §8.3a | Power management, duty cycles |
| **OpenAPI** | §2.2 | Error backoff policies |
| **OpenAPI** | §3.9 | Telemetry buffering, resume |
| **Telemetry SSE** | §1.3 | Event buffering, resume |
| **Telemetry SSE** | §4 | Heartbeat, cadence |
| **ICD** | §10 | Command timeouts, validation |

---

> **Document Status**: Provisional v0.3  
> **Next Review**: 2025-02-15  
> **Stakeholders**: System Architect, IV&V Lead, Product Owner
