# Performance Testing

This directory contains performance testing scripts and scenarios for the Radio Control Container.

## Microbenchmarks

Run Go microbenchmarks:

```bash
# Run all benchmarks
go test -bench=. -benchmem ./internal/command ./internal/telemetry ./internal/adapter

# Run specific benchmark
go test -bench=BenchmarkSetPower -benchmem ./internal/command

# Run with race detection
go test -bench=. -race ./internal/command
```

## Load Testing

### Vegeta (HTTP Load Testing)

```bash
# Install vegeta
go install github.com/tsenart/vegeta@latest

# Run load tests
bash test/perf/vegeta_scenarios.sh
```


## Performance Targets

- **Control Endpoints**: p95 < 100ms (mock operations)
- **Telemetry**: p95 < 50ms (event publishing)
- **Concurrent Operations**: No deadlocks or race conditions
- **Memory**: Stable allocation patterns under load

## Benchmark Results

Expected performance characteristics:

- `BenchmarkSetPower`: < 1ms per operation
- `BenchmarkPublishWithSubscribers`: < 100μs per event (1 subscriber), < 1ms (100 subscribers)
- `BenchmarkEventIDGeneration`: < 100ns per ID
- `BenchmarkSubscribe`: < 1ms per subscription

## Load Testing Scenarios

1. **List Radios**: GET /api/v1/radios (100 req/s for 30s)
2. **Set Power**: POST /api/v1/radios/{id}/power (50 req/s for 30s)
3. **Set Channel**: POST /api/v1/radios/{id}/channel (25 req/s for 30s)
4. **Mixed Workload**: Combination of all operations (100 req/s for 60s)
5. **High Load**: 200 req/s for 30s