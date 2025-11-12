# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased] - 2024-12-19

### Added
- **Ports and Adapters Architecture**: Implemented clean architecture with explicit ports
  - `internal/command/ports.go`: OrchestratorPort and RadioManager interfaces
  - `internal/api/ports.go`: OrchestratorPort, TelemetryPort, and RadioReadPort interfaces
  - Compile-time assertions ensure concrete types implement ports
- **CI Enforcement**: Extended `scripts/anti-peek-check.sh` with architectural rules
  - No wrapper usage enforcement
  - API server port-only dependency enforcement  
  - Interface{} elimination enforcement
  - E2E test black-box enforcement

### Changed
- **Power Data Type**: Changed power dBm from `int` to `float64` throughout codebase
  - `IRadioAdapter.SetPower()` and `ReadPowerActual()` now use `float64`
  - `RadioState.PowerDbm` field changed to `float64`
  - All adapter implementations updated (SilvusMock, FakeAdapter, MockAdapter)
  - All test cases updated with float64 literals and format strings
- **API Server Architecture**: Refactored to use port interfaces instead of concrete types
  - `Server` struct fields now use `TelemetryPort`, `OrchestratorPort`, `RadioReadPort`
  - Removed direct dependencies on `*telemetry.Hub`, `*command.Orchestrator`, `*radio.Manager`
  - Constructor signatures updated to accept port interfaces
- **Test Harness**: Hardened to expose only `URL` and `Shutdown` methods
  - Removed internal component access from E2E tests
  - Enforced black-box testing principles

### Removed
- **Wrapper Elimination**: Removed all `radioManagerWrapper` instances
  - `cmd/rcc/main.go`: Direct `radio.Manager` usage
  - `test/harness/harness.go`: Direct `radio.Manager` usage  
  - `internal/command/orchestrator_bench_test.go`: Direct `radio.Manager` usage
- **Interface{} Usage**: Eliminated avoidable `interface{}` in non-JSON boundaries
  - `Orchestrator.SetChannelByIndex()` now uses typed `RadioManager` interface
  - `Orchestrator.SetPower()` now uses `float64` instead of `int`
  - All method signatures use concrete types or DTOs

### Security
- **CI Rules**: Added automated enforcement of architectural principles
  - Prevents regression to wrapper patterns
  - Enforces port-based dependencies
  - Maintains E2E test black-box requirements
  - Validates interface{} elimination

### Technical Details
- **Type Safety**: Improved compile-time type checking with explicit interfaces
- **Dependency Inversion**: High-level modules now depend on abstractions (ports)
- **Test Isolation**: E2E tests can no longer access internal implementation details
- **Maintainability**: Clear separation of concerns with explicit port boundaries

### Migration Notes
- **Breaking Change**: Power values must now be `float64` instead of `int`
- **API Compatibility**: External API remains unchanged (JSON still uses numbers)
- **Test Updates**: All test files updated to use `float64` literals and format strings
- **CI Integration**: New architectural rules will fail CI on violations

### Files Modified
- `internal/command/ports.go` (new)
- `internal/api/ports.go` (new)  
- `internal/command/orchestrator.go`
- `internal/api/server.go`
- `internal/adapter/adapter.go`
- `internal/adapter/silvusmock/silvusmock.go`
- `internal/adapter/fake/fake.go`
- `internal/adapter/adapter_test.go`
- `internal/adapter/fake/fake_test.go`
- `internal/adapter/silvusmock/silvusmock_test.go`
- `internal/adaptertest/conformance.go`
- `internal/api/routes.go`
- `internal/api/test_setup.go`
- `internal/command/orchestrator_test.go`
- `internal/command/silvus_band_plan_integration_test.go`
- `internal/command/orchestrator_bench_test.go`
- `cmd/rcc/main.go`
- `test/harness/harness.go`
- `scripts/anti-peek-check.sh`

### Architecture Compliance
- ✅ **ARCH-REMEDY-01**: Ports defined at consuming layers
- ✅ **ARCH-REMEDY-02**: Providers implement ports with compile-time assertions
- ✅ **ARCH-REMEDY-03**: API server uses ports only
- ✅ **ARCH-REMEDY-04**: Test harness hardened (black-box only)
- ✅ **ARCH-REMEDY-05**: Interface{} eliminated in non-JSON paths
- ✅ **ARCH-REMEDY-06**: Wrapper duplication removed
- ✅ **ARCH-RULES-07**: CI enforcement implemented
