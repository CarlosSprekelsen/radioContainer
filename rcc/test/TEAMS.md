# Test Team Assignments

## Team 1: Unit Tests (Fast)
- **Packages**: `internal/auth`, `internal/command`, `internal/config`, `internal/adapter`
- **Coverage Target**: 85%+
- **Run Command**: `make test-unit-team`
- **Focus**: Isolated component testing with mocks
- **Time Target**: < 2 minutes

### Responsibilities
- Unit test coverage for critical packages
- Mock-based testing for isolated components
- Fast feedback loop for development
- Package-specific coverage validation

### Coverage Targets
- `auth`: 85%+ (currently 83.7%)
- `command`: 85%+ (currently 79.5%)
- `config`: 80%+ (currently 62.4%)
- `adapter/fake`: 80%+ (currently 59.6%)

---

## Team 2: Integration Tests (Medium)
- **Packages**: `test/integration/*`
- **Coverage Target**: 70%+
- **Run Command**: `make test-integration-team`
- **Focus**: Multi-component testing without HTTP
- **Time Target**: < 5 minutes

### Responsibilities
- Cross-component interaction testing
- Real component wiring (no mocks)
- Error propagation validation
- Audit log verification
- Telemetry event flow testing

### Test Categories
- **Orchestrator**: Command → Radio → Adapter flows
- **Telemetry**: Hub → SSE → Client flows  
- **Auth**: Auth → API → Orchestrator flows

### Coverage Targets
- Orchestrator flows: 70%+
- Telemetry flows: 70%+
- Auth flows: 70%+

---

## Team 3: E2E Tests (Slow)
- **Packages**: `test/e2e/*`
- **Coverage Target**: 100% contract validation
- **Run Command**: `make test-e2e-team`
- **Focus**: HTTP black-box testing
- **Time Target**: < 10 minutes

### Responsibilities
- API contract validation
- HTTP endpoint testing
- End-to-end user scenarios
- Anti-peek enforcement (no internal access)
- Golden test validation

### Contract Validation
- All API endpoints covered
- Error code mapping validated
- Response schema validation
- HTTP status code verification

---

## Parallel Execution

### Local Development
```bash
# Run all teams in parallel
make test-parallel

# Run individual teams
make test-unit-team      # Team 1
make test-integration-team  # Team 2  
make test-e2e-team       # Team 3
```

### CI Pipeline
- **Unit Tests**: Fast feedback (< 2 min)
- **Integration Tests**: Medium feedback (< 5 min)
- **E2E Tests**: Comprehensive validation (< 10 min)
- **Gate Job**: Final validation after all teams pass

## Quality Gates

### Coverage Requirements
- **Overall**: ≥ 80%
- **Critical Packages**: ≥ 85%
- **Integration**: ≥ 70%
- **E2E**: 100% contract coverage

### Anti-Peek Enforcement
- E2E tests must not access `internal/` packages
- Only HTTP client and JSON validation allowed
- Violations fail CI pipeline

### Flaky Test Management
- Tag flaky tests with `//go:build flaky`
- Run separately: `make test-quarantine`
- Don't block main pipeline

## Test Data Management

### Shared Fixtures
- **Location**: `test/fixtures/`
- **Usage**: All teams use same test data
- **Benefits**: Consistent, comparable results

### Fixture Categories
- **Radios**: `StandardSilvusRadio()`, `MultiRadioSetup()`
- **Channels**: `WiFi24GHzChannels()`, `UHFChannels()`
- **Errors**: `BusyError()`, `RangeError()`, `InternalError()`
- **Telemetry**: `HeartbeatSequence()`, `PowerChangeSequence()`
- **Scenarios**: `HappyPath()`, `ErrorRecovery()`, `Concurrent()`

## Success Metrics

| Metric | Current | Target | Team Owner |
|--------|---------|--------|------------|
| Unit coverage (critical) | 79-84% | 85%+ | Team 1 |
| Integration coverage | 0% | 70%+ | Team 2 |
| E2E contract validation | Partial | 100% | Team 3 |
| CI pipeline time | ~15min serial | ~5min parallel | DevOps |
| Test flakiness rate | Unknown | <2% | All teams |
| Coverage delta tracked | No | Yes (CI) | DevOps |

## Troubleshooting

### Team 1 Issues
- Low coverage: Add unit tests for uncovered paths
- Slow tests: Optimize test setup, use mocks effectively
- Flaky tests: Isolate timing dependencies

### Team 2 Issues  
- Missing integration coverage: Add cross-component tests
- Slow tests: Optimize component wiring
- Complex setup: Use fixtures for consistent setup

### Team 3 Issues
- Internal access violations: Refactor to HTTP-only
- Contract failures: Update tests to match API docs
- Slow tests: Optimize test scenarios, use parallel execution
