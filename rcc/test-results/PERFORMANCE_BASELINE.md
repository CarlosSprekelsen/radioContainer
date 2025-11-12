# Performance Baseline Report

## Fast Benchmarks (1s benchtime, 30s timeout)

### Command Package
```
BenchmarkSetPower-4                   	   85626	     13816 ns/op	    1248 B/op	      15 allocs/op
BenchmarkSetPowerWithoutTelemetry-4   	  105225	     12024 ns/op	     752 B/op	       9 allocs/op
BenchmarkSetChannel-4                 	   67756	     16993 ns/op	    1247 B/op	      15 allocs/op
BenchmarkGetState-4                   	  129026	     11520 ns/op	     768 B/op	      10 allocs/op
BenchmarkOrchestratorConcurrent-4     	  120807	     10867 ns/op	     768 B/op	      10 allocs/op
```

### Telemetry Package
```
BenchmarkPublishWithSubscribers/Subscribers_1-4         	  280387	      3947 ns/op	    1215 B/op	      18 allocs/op
BenchmarkPublishWithSubscribers/Subscribers_5-4         	  118196	      8842 ns/op	    4437 B/op	      73 allocs/op
BenchmarkPublishWithSubscribers/Subscribers_10-4        	   59382	     18194 ns/op	    8411 B/op	     143 allocs/op
BenchmarkPublishWithoutSubscribers-4                    	 1719868	       898.7 ns/op	     439 B/op	       3 allocs/op
BenchmarkSubscribe-4                                    	      10	 100488002 ns/op	   15221 B/op	      75 allocs/op
BenchmarkEventIDGeneration-4                            	 3717672	       273.8 ns/op	       8 B/op	       1 allocs/op
BenchmarkBufferEvent-4                                  	 1000000	      1092 ns/op	     447 B/op	       4 allocs/op
BenchmarkHubConcurrent-4                                	19519808	       486.2 ns/op	     431 B/op	       2 allocs/op
BenchmarkHeartbeat-4                                    	 1000000	      1206 ns/op	     376 B/op	       4 allocs/op
```

## Slow Benchmarks (10s benchtime, 5m timeout)

### Command Package (Deep Profiling)
```
BenchmarkSetPower-4                   	  846730	     14742 ns/op	    1247 B/op	      15 allocs/op
BenchmarkSetPowerWithoutTelemetry-4   	 1000000	     11291 ns/op	     752 B/op	       9 allocs/op
BenchmarkSetChannel-4                 	  768312	     15740 ns/op	    1247 B/op	      15 allocs/op
BenchmarkGetState-4                   	  990370	     10655 ns/op	     768 B/op	      10 allocs/op
BenchmarkOrchestratorConcurrent-4     	 1000000	     13837 ns/op	    1247 B/op	      15 allocs/op
```

### Telemetry Package (Deep Profiling)
```
BenchmarkPublishWithManySubscribers/Subscribers_50-4         	  142720	     98621 ns/op	   39131 B/op	     703 allocs/op
BenchmarkPublishWithManySubscribers/Subscribers_100-4        	   67656	    222408 ns/op	   78519 B/op	    1403 allocs/op
BenchmarkPublishWithManySubscribers/Subscribers_500-4        	   11696	   1203908 ns/op	  418008 B/op	    6990 allocs/op
BenchmarkPublishStressTest-4                                 	 9457534	      1599 ns/op	     447 B/op	       4 allocs/op
BenchmarkConcurrentSubscribers-4                             	     476	  25164247 ns/op	   14371 B/op	      66 allocs/op
BenchmarkMemoryIntensive-4                                   	 4081322	      3019 ns/op	    2370 B/op	      10 allocs/op
```

## Performance Targets

### Fast Benchmarks (CI/CD)
- **Execution Time**: < 30 seconds
- **Purpose**: Regression detection
- **Frequency**: Every commit

### Slow Benchmarks (Deep Profiling)
- **Execution Time**: < 5 minutes
- **Purpose**: Performance optimization
- **Frequency**: Nightly/weekly

## Key Metrics

### Command Operations
- **SetPower**: ~12-15μs (with telemetry), ~11-12μs (without)
- **SetChannel**: ~15-17μs
- **GetState**: ~10-11μs
- **Concurrent**: ~10-14μs

### Telemetry Operations
- **Publish (no subscribers)**: ~0.9μs
- **Publish (1 subscriber)**: ~3.9μs
- **Publish (5 subscribers)**: ~8.8μs
- **Publish (10 subscribers)**: ~18.2μs
- **Publish (50 subscribers)**: ~98.6μs
- **Publish (100 subscribers)**: ~222.4μs
- **Publish (500 subscribers)**: ~1.2ms

### Memory Usage
- **Event ID Generation**: 8 bytes, 1 allocation
- **Buffer Event**: 447 bytes, 4 allocations
- **Hub Concurrent**: 431 bytes, 2 allocations
- **Heartbeat**: 376 bytes, 4 allocations

## Regression Detection

### Fast Benchmarks
- Run on every commit
- Fail if performance degrades > 20%
- Target: < 30s execution time

### Slow Benchmarks
- Run nightly/weekly
- Detailed profiling for optimization
- Target: < 5m execution time

## Usage

```bash
# Fast benchmarks (CI/CD)
make bench-fast

# Slow benchmarks (profiling)
make bench-slow

# All benchmarks
make bench
```

## Notes

- Fast benchmarks use reduced subscriber counts (1, 5, 10)
- Slow benchmarks use higher subscriber counts (50, 100, 500)
- All benchmarks have timeout protection to prevent hangs
- Memory allocation patterns are tracked for optimization
- Concurrent operations are tested for race conditions
