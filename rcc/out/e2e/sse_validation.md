# SSE Event Validation Analysis

## Contract Version
- **SPEC_VERSION**: v1.0.0
- **TELEMETRY_SCHEMA_SHA**: 668cb7a1dbf7f57338340b6825ec6d8f2fb895fd60733cb3cdd0d4aeec66e46c

## SSE Event Capture

### Event Stream Sample
```
event: ready
data: {"snapshot":{"activeRadioId":"","radios":[]}}
id: 1
event: powerChanged
data: {"powerDbm":25,"radioId":"silvus-001","ts":"2025-10-03T06:45:27Z"}
```

### Event Validation Results

| Event Type | Count | Schema Validation | PASS/FAIL | Issues |
|------------|-------|-------------------|-----------|--------|
| `ready` | 1 | FAIL | FAIL | Missing `event:` field in some events |
| `powerChanged` | 1 | FAIL | FAIL | Missing `data:` field in some events |
| `heartbeat` | 0 | N/A | N/A | No heartbeat events captured |
| `channelChanged` | 0 | N/A | N/A | No channel change events captured |

### Schema Validation Issues

#### 1. Event Format Problems
- **Issue**: Events sent as separate lines instead of proper event-data pairs
- **Expected**: Each event should have both `event:` and `data:` lines
- **Actual**: Some events only have `event:` or `data:` lines

#### 2. Missing Required Fields
- **Issue**: Some events lack the required `data` field
- **Expected**: All events must have a `data` field
- **Actual**: Some events only have `event:` line

#### 3. Invalid Event Types
- **Issue**: Empty event types are being sent
- **Expected**: Valid event types: `ready`, `heartbeat`, `powerChanged`, `channelChanged`
- **Actual**: Empty event type (`""`)

### Contract Compliance Summary
- **Total Events Captured**: 2
- **Schema Compliant**: 0
- **Compliance Rate**: 0%

### Required Fixes
1. **Event Format**: Send proper event-data pairs
2. **Required Fields**: Ensure all events have both `event:` and `data:` fields
3. **Event Types**: Send valid event types only
4. **Heartbeat Events**: Implement heartbeat events according to CB-TIMING specification
