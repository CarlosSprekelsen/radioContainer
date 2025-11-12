# Integration Tests

## Scope and Rules

Integration tests validate multi-component interactions without HTTP endpoints. These tests bridge the gap between unit tests (isolated, mocked) and E2E tests (HTTP black-box).

## Build Tag
Use `//go:build integration` for all integration tests.

## Test Categories

### Orchestrator Integration (`test/integration/orchestrator/`)
- Command → Radio → Adapter flows
- Error propagation and normalization
- Audit log generation
- Telemetry event emission

### Telemetry Integration (`test/integration/telemetry/`)
- Hub → SSE → Client flows
- Event buffering and sequencing
- Heartbeat timing and jitter
- Connection lifecycle management

### Auth Integration (`test/integration/auth/`)
- Auth → API → Orchestrator flows
- Token validation and expiration
- Permission enforcement
- Session management

## Rules

1. **No HTTP**: Use direct component wiring, not HTTP clients
2. **Real Components**: Use actual implementations, not mocks
3. **Config-Driven**: Use test fixtures for consistent inputs
4. **Timing**: Use config values, not literals
5. **Error Mapping**: Validate normalized error codes per Architecture §8.5
6. **Audit Logs**: Verify structured audit entries per Architecture §8.6

## Coverage Target
70%+ of cross-package interaction paths

## Running Tests
```bash
# Run all integration tests
make test-integration-team

# Run specific category
go test ./test/integration/orchestrator/... -tags=integration -count=1

# With coverage
make integration-cover
```
