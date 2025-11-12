# Route Parity Analysis

## Contract Version
- **SPEC_VERSION**: v1.0.0
- **COMMIT_HASH**: abc123def456
- **LAST_UPDATED**: 2025-10-03T09:56:00Z

## Route Parity Table

| Route | Method | Expected Status (from contract) | Actual Status (probe) | PASS/FAIL | Notes |
|-------|--------|--------------------------------|----------------------|-----------|-------|
| `/api/v1/radios` | GET | 200 | 200 | PASS | Returns radio list |
| `/api/v1/radios/select` | POST | 200 | 400 | FAIL | Radio selection failing |
| `/api/v1/radios/silvus-001/power` | GET | 200 | 200 | PASS | Power readback works |
| `/api/v1/radios/silvus-001/power` | POST | 200 | 200 | PASS | Power setting works |
| `/api/v1/radios/silvus-001/channel` | GET | 200 | 200 | PASS | Channel readback works |
| `/api/v1/radios/silvus-001/channel` | POST | 200 | 200 | PASS | Channel setting works |
| `/api/v1/radios/invalid-radio-id` | GET | 404 | 200 | FAIL | Should return 404 for invalid radio |
| `/api/v1/radios/non-existent/power` | POST | 404 | 400 | FAIL | Should return 404, getting 400 |
| `/api/v1/radios/silvus-001/power` | POST | 400 | 200 | FAIL | Invalid power should return 400 |
| `/api/v1/telemetry` | GET | 200 | 200 | PASS | SSE endpoint accessible |

## Summary
- **Total Routes Tested**: 10
- **PASS**: 6 routes
- **FAIL**: 4 routes
- **Success Rate**: 60%

## Key Issues Identified
1. **Radio Selection**: POST `/api/v1/radios/select` returns 400 instead of 200
2. **Invalid Radio Handling**: GET `/api/v1/radios/invalid-radio-id` returns 200 instead of 404
3. **Error Status Mapping**: Invalid requests return 200 instead of expected error codes
4. **Missing Error Response Structure**: Error responses lack required `error` field
