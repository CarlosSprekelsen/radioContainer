# E2E Test Drift Ledger

## DRIFT: BUSY scenario triggering
**Date**: 2025-01-27  
**Spec**: No black-box mechanism to induce vendor BUSY (no admin/fault endpoint)  
**Action**: Skip BUSY sub-test; propose test-only startup fault profile (build tag `e2e`, env `RCC_FAULT_PROFILE=silvus:busy`)  
**Status**: OPEN - Test skipped with clear reason in `TestE2E_AdapterBusy`

## DRIFT: API Contract Violations (Phase 3 Triage)
**Date**: 2025-01-27  
**Issues Found**:
- GET /api/v1/radios returns 200 but missing expected fields (id, type)
- POST /api/v1/radios/select returns 400 instead of expected 200
- Error responses missing 'error' field, wrong status codes (500 vs 400/404)
- Invalid JSON handling returns 200 instead of 400
- Radio not found returns 200 instead of 404
- Telemetry SSE missing powerChanged events

**Root Cause**: API handlers not conforming to OpenAPI spec
**Action**: Phase 4 - Fix handlers to match contract
**Status**: OPEN