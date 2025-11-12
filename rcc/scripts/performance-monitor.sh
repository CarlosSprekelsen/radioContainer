#!/bin/bash

# Performance Monitoring Script
# Generates performance reports and tracks regressions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Performance Monitoring Dashboard${NC}"
echo "=================================="

# Create results directory
mkdir -p benchmark-results
mkdir -p performance-reports

# Function to run benchmarks and capture results
run_benchmarks() {
    local benchmark_type=$1
    local output_file="benchmark-results/${benchmark_type}-$(date +%Y%m%d-%H%M%S).txt"
    
    echo -e "${YELLOW}📊 Running ${benchmark_type} benchmarks...${NC}"
    
    if [ "$benchmark_type" = "fast" ]; then
        make bench-fast > "$output_file" 2>&1
    elif [ "$benchmark_type" = "slow" ]; then
        make bench-slow > "$output_file" 2>&1
    else
        make bench > "$output_file" 2>&1
    fi
    
    echo -e "${GREEN}✅ Benchmarks completed: $output_file${NC}"
    echo "$output_file"
}

# Function to extract key metrics
extract_metrics() {
    local file=$1
    local metrics_file="performance-reports/metrics-$(basename "$file" .txt).json"
    
    echo -e "${YELLOW}📈 Extracting performance metrics...${NC}"
    
    # Extract key performance metrics
    cat > "$metrics_file" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "benchmark_type": "$(basename "$file" | cut -d'-' -f1)",
  "metrics": {
    "command_operations": {
      "setPower": "$(grep "BenchmarkSetPower" "$file" | awk '{print $3}' | head -1 || echo 'N/A')",
      "setChannel": "$(grep "BenchmarkSetChannel" "$file" | awk '{print $3}' | head -1 || echo 'N/A')",
      "getState": "$(grep "BenchmarkGetState" "$file" | awk '{print $3}' | head -1 || echo 'N/A')"
    },
    "telemetry_operations": {
      "publish_no_subscribers": "$(grep "BenchmarkPublishWithoutSubscribers" "$file" | awk '{print $3}' | head -1 || echo 'N/A')",
      "publish_1_subscriber": "$(grep "Subscribers_1" "$file" | awk '{print $3}' | head -1 || echo 'N/A')",
      "publish_5_subscribers": "$(grep "Subscribers_5" "$file" | awk '{print $3}' | head -1 || echo 'N/A')",
      "publish_10_subscribers": "$(grep "Subscribers_10" "$file" | awk '{print $3}' | head -1 || echo 'N/A')"
    },
    "memory_usage": {
      "event_id_generation": "$(grep "BenchmarkEventIDGeneration" "$file" | awk '{print $5}' | head -1 || echo 'N/A')",
      "buffer_event": "$(grep "BenchmarkBufferEvent" "$file" | awk '{print $5}' | head -1 || echo 'N/A')",
      "hub_concurrent": "$(grep "BenchmarkHubConcurrent" "$file" | awk '{print $5}' | head -1 || echo 'N/A')"
    }
  }
}
EOF
    
    echo -e "${GREEN}✅ Metrics extracted: $metrics_file${NC}"
    echo "$metrics_file"
}

# Function to check for regressions
check_regressions() {
    local current_file=$1
    local baseline_file="performance-reports/baseline.json"
    
    echo -e "${YELLOW}🔍 Checking for performance regressions...${NC}"
    
    if [ ! -f "$baseline_file" ]; then
        echo -e "${YELLOW}⚠️  No baseline found. Creating baseline from current results.${NC}"
        cp "$current_file" "$baseline_file"
        echo -e "${GREEN}✅ Baseline created: $baseline_file${NC}"
        return 0
    fi
    
    # Simple regression check (20% threshold)
    local regression_found=false
    
    # Check command operations
    local current_setpower=$(grep "BenchmarkSetPower" "$current_file" | awk '{print $3}' | head -1 | sed 's/ns\/op//')
    local baseline_setpower=$(grep "BenchmarkSetPower" "$baseline_file" | awk '{print $3}' | head -1 | sed 's/ns\/op//')
    
    if [ -n "$current_setpower" ] && [ -n "$baseline_setpower" ]; then
        local regression_percent=$(( (current_setpower - baseline_setpower) * 100 / baseline_setpower ))
        if [ $regression_percent -gt 20 ]; then
            echo -e "${RED}❌ REGRESSION DETECTED: SetPower performance degraded by ${regression_percent}%${NC}"
            regression_found=true
        else
            echo -e "${GREEN}✅ SetPower performance: ${regression_percent}% change (within 20% threshold)${NC}"
        fi
    fi
    
    if [ "$regression_found" = true ]; then
        echo -e "${RED}❌ Performance regression detected!${NC}"
        return 1
    else
        echo -e "${GREEN}✅ No performance regressions detected${NC}"
        return 0
    fi
}

# Function to generate performance report
generate_report() {
    local benchmark_file=$1
    local metrics_file=$2
    local report_file="performance-reports/performance-report-$(date +%Y%m%d-%H%M%S).md"
    
    echo -e "${YELLOW}📋 Generating performance report...${NC}"
    
    cat > "$report_file" << EOF
# Performance Report

**Generated:** $(date)
**Benchmark File:** $benchmark_file
**Metrics File:** $metrics_file

## Executive Summary

This report provides a comprehensive analysis of the Radio Control Container performance benchmarks.

## Key Metrics

### Command Operations
- **SetPower**: $(grep "BenchmarkSetPower" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A')
- **SetChannel**: $(grep "BenchmarkSetChannel" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A')
- **GetState**: $(grep "BenchmarkGetState" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A')

### Telemetry Operations
- **Publish (no subscribers)**: $(grep "BenchmarkPublishWithoutSubscribers" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A')
- **Publish (1 subscriber)**: $(grep "Subscribers_1" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A')
- **Publish (5 subscribers)**: $(grep "Subscribers_5" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A')
- **Publish (10 subscribers)**: $(grep "Subscribers_10" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A')

### Memory Usage
- **Event ID Generation**: $(grep "BenchmarkEventIDGeneration" "$benchmark_file" | awk '{print $5}' | head -1 || echo 'N/A')
- **Buffer Event**: $(grep "BenchmarkBufferEvent" "$benchmark_file" | awk '{print $5}' | head -1 || echo 'N/A')
- **Hub Concurrent**: $(grep "BenchmarkHubConcurrent" "$benchmark_file" | awk '{print $5}' | head -1 || echo 'N/A')

## Performance Targets

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| SetPower | < 20μs | $(grep "BenchmarkSetPower" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A') | $(grep "BenchmarkSetPower" "$benchmark_file" | awk '{print $3}' | head -1 | sed 's/ns\/op//' | awk '{if($1<20000) print "✅ PASS"; else print "❌ FAIL"}' || echo 'N/A') |
| SetChannel | < 20μs | $(grep "BenchmarkSetChannel" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A') | $(grep "BenchmarkSetChannel" "$benchmark_file" | awk '{print $3}' | head -1 | sed 's/ns\/op//' | awk '{if($1<20000) print "✅ PASS"; else print "❌ FAIL"}' || echo 'N/A') |
| Publish (no subs) | < 1μs | $(grep "BenchmarkPublishWithoutSubscribers" "$benchmark_file" | awk '{print $3}' | head -1 || echo 'N/A') | $(grep "BenchmarkPublishWithoutSubscribers" "$benchmark_file" | awk '{print $3}' | head -1 | sed 's/ns\/op//' | awk '{if($1<1000) print "✅ PASS"; else print "❌ FAIL"}' || echo 'N/A') |

## Recommendations

1. **Monitor Trends**: Track performance metrics over time
2. **Regression Detection**: Set up automated alerts for >20% degradation
3. **Optimization**: Focus on areas showing consistent performance issues
4. **Load Testing**: Integrate with Vegeta scenarios for end-to-end performance

## Next Steps

- [ ] Set up automated performance monitoring
- [ ] Create performance dashboards
- [ ] Implement regression alerts
- [ ] Integrate with CI/CD pipeline
EOF
    
    echo -e "${GREEN}✅ Performance report generated: $report_file${NC}"
    echo "$report_file"
}

# Main execution
main() {
    local benchmark_type=${1:-"fast"}
    
    echo -e "${BLUE}🎯 Running performance monitoring for: $benchmark_type${NC}"
    
    # Run benchmarks
    local benchmark_file=$(run_benchmarks "$benchmark_type")
    
    # Extract metrics
    local metrics_file=$(extract_metrics "$benchmark_file")
    
    # Check for regressions
    if check_regressions "$benchmark_file"; then
        echo -e "${GREEN}✅ Performance regression check passed${NC}"
    else
        echo -e "${RED}❌ Performance regression detected!${NC}"
        exit 1
    fi
    
    # Generate report
    local report_file=$(generate_report "$benchmark_file" "$metrics_file")
    
    echo -e "${GREEN}🎉 Performance monitoring completed successfully!${NC}"
    echo -e "${BLUE}📊 Results:${NC}"
    echo "   - Benchmark file: $benchmark_file"
    echo "   - Metrics file: $metrics_file"
    echo "   - Report file: $report_file"
}

# Run main function with arguments
main "$@"
