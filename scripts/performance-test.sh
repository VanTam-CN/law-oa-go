#!/bin/bash

# 性能测试脚本
# 运行系统性能基准测试并生成报告

set -e

echo "=== Law OA Go Performance Test Suite ==="
echo "Starting performance tests..."
echo "Backend entry: go run . or make build && ./bin/law-oa-go"

# 设置测试环境变量
export ENVIRONMENT=test
export DB_DATABASE=law_oa_test
export REDIS_DB=1

# 创建测试结果目录
RESULT_DIR="./test-results/performance"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
TEST_DIR="$RESULT_DIR/$TIMESTAMP"
mkdir -p "$TEST_DIR"

echo "Test results will be saved to: $TEST_DIR"

# 函数：运行基准测试
run_benchmark() {
    local test_name=$1
    local output_file="$TEST_DIR/${test_name}.txt"

    echo "Running $test_name benchmark..."

    # 运行基准测试并保存结果
    if go test -benchmem -run=^$ -bench=$test_name ./tests/performance/ > "$output_file" 2>&1; then
        echo "✓ $test_name benchmark completed"
    else
        echo "✗ $test_name benchmark failed"
        return 1
    fi

    # 生成性能报告
    if command -v benchstat >/dev/null 2>&1; then
        echo "Generating performance comparison with benchstat..."
        # 如果有之前的基准测试结果，进行比较
        latest_benchmark=$(find "$RESULT_DIR" -name "${test_name}.txt" -type f | sort -r | sed -n 2p)
        if [[ -n "$latest_benchmark" ]]; then
            benchstat "$latest_benchmark" "$output_file" > "$TEST_DIR/${test_name}_comparison.txt"
            echo "✓ Performance comparison saved"
        fi
    fi
}

# 函数：检查服务依赖
check_dependencies() {
    echo "Checking dependencies..."

    # 检查MySQL
    if ! command -v mysql >/dev/null 2>&1; then
        echo "⚠️  MySQL client not found. Install MySQL for full database testing."
    else
        echo "✓ MySQL client found"
    fi

    # 检查Redis
    if ! redis-cli ping >/dev/null 2>&1; then
        echo "⚠️  Redis server not running. Some cache tests may be skipped."
    else
        echo "✓ Redis server is running"
    fi

    # 检查Go工具
    if ! command -v go >/dev/null 2>&1; then
        echo "✗ Go not found. Please install Go."
        exit 1
    else
        echo "✓ Go found: $(go version)"
    fi

    # 检查必要的Go包
    echo "Installing/updating test dependencies..."
    go install github.com/itchyny/gojq/cmd/gojq@latest 2>/dev/null || true
    go install golang.org/x/perf/cmd/benchstat@latest 2>/dev/null || true
}

# 函数：运行负载测试
run_load_test() {
    echo "Running load tests..."

    # 使用Apache Bench (ab) 进行负载测试
    if command -v ab >/dev/null 2>&1; then
        echo "Running Apache Bench load test..."

        # 测试API端点
        endpoints=(
            "http://localhost:8080/api/dashboard/statistics"
            "http://localhost:8080/api/cases"
            "http://localhost:8080/api/clients"
        )

        for endpoint in "${endpoints[@]}"; do
            echo "Testing $endpoint..."
            ab -n 1000 -c 10 "$endpoint" > "$TEST_DIR/load_test_$(basename $endpoint).txt" 2>&1 || true
        done

        echo "✓ Load tests completed"
    else
        echo "⚠️  Apache Bench (ab) not found. Skipping load tests."
    fi

    # 使用hey进行HTTP负载测试（如果可用）
    if command -v hey >/dev/null 2>&1; then
        echo "Running hey load test..."
        hey -n 1000 -c 20 -m GET http://localhost:8080/api/dashboard/statistics > "$TEST_DIR/hey_load_test.txt" 2>&1 || true
        echo "✓ Hey load test completed"
    fi
}

# 函数：运行内存分析
run_memory_analysis() {
    echo "Running memory analysis..."

    # 使用Go的内存分析工具
    go test -memprofile="$TEST_DIR/mem.prof" -bench=BenchmarkAPIRequest ./tests/performance/ 2>/dev/null || true

    if [[ -f "$TEST_DIR/mem.prof" ]]; then
        echo "✓ Memory profile generated"

        # 生成内存报告
        if command -v go >/dev/null 2>&1; then
            go tool pprof -text "$TEST_DIR/mem.prof" > "$TEST_DIR/memory_report.txt" 2>&1 || true
            echo "✓ Memory report generated"
        fi
    fi
}

# 函数：运行CPU分析
run_cpu_analysis() {
    echo "Running CPU analysis..."

    # 使用Go的CPU分析工具
    go test -cpuprofile="$TEST_DIR/cpu.prof" -bench=BenchmarkAPIRequest ./tests/performance/ 2>/dev/null || true

    if [[ -f "$TEST_DIR/cpu.prof" ]]; then
        echo "✓ CPU profile generated"

        # 生成CPU报告
        if command -v go >/dev/null 2>&1; then
            go tool pprof -text "$TEST_DIR/cpu.prof" > "$TEST_DIR/cpu_report.txt" 2>&1 || true
            echo "✓ CPU report generated"
        fi
    fi
}

# 函数：生成性能报告
generate_performance_report() {
    echo "Generating performance report..."

    cat > "$TEST_DIR/performance_report.md" << EOF
# Law OA Go Performance Test Report

**Test Date:** $(date)
**Test Environment:** $ENVIRONMENT

## Test Summary

This report contains the results of various performance tests conducted on the Law OA Go application.

## Test Categories

### 1. API Performance Benchmarks

Files:
- \`api_benchmark.txt\` - Overall API performance results
- \`api_benchmark_comparison.txt\` - Comparison with previous results (if available)

### 2. Database Performance Tests

Files:
- \`database_benchmark.txt\` - Database query performance results
- \`database_benchmark_comparison.txt\` - Database performance comparison

### 3. Cache Performance Tests

Files:
- \`cache_benchmark.txt\` - Redis cache performance results
- \`cache_benchmark_comparison.txt\` - Cache performance comparison

### 4. Middleware Performance Tests

Files:
- \`middleware_benchmark.txt\` - Middleware performance results
- \`middleware_benchmark_comparison.txt\` - Middleware performance comparison

### 5. Concurrent Request Tests

Files:
- \`concurrent_benchmark.txt\` - Concurrent request performance results
- \`concurrent_benchmark_comparison.txt\` - Concurrent performance comparison

### 6. Load Tests

Files:
- \`load_test_*.txt\` - Apache Bench load test results
- \`hey_load_test.txt\` - Hey load test results

### 7. Memory and CPU Analysis

Files:
- \`memory_report.txt\` - Memory usage analysis
- \`cpu_report.txt\` - CPU usage analysis
- \`mem.prof\` - Memory profile (use with \`go tool pprof\`)
- \`cpu.prof\` - CPU profile (use with \`go tool pprof\`)

## Key Metrics to Monitor

- **Response Time**: API endpoint response times should be under 200ms for 95th percentile
- **Throughput**: System should handle at least 1000 requests per second
- **Memory Usage**: Memory usage should remain stable under load
- **Database Connections**: Connection pool usage should not exceed 80%
- **Cache Hit Ratio**: Cache hit ratio should be above 80% for cached endpoints

## Recommendations

Based on the test results, here are some performance optimization recommendations:

1. **Database Optimization**
   - Monitor slow queries and add appropriate indexes
   - Consider connection pool tuning based on usage patterns
   - Implement query result caching for frequently accessed data

2. **Cache Strategy**
   - Implement multi-level caching (application and database level)
   - Use cache warming for frequently accessed data
   - Monitor cache hit ratios and adjust TTL values

3. **Concurrency Management**
   - Optimize goroutine usage and avoid goroutine leaks
   - Use worker pools for CPU-intensive tasks
   - Implement proper rate limiting to prevent overload

4. **Memory Management**
   - Monitor memory usage patterns and identify potential leaks
   - Optimize data structures to reduce memory footprint
   - Implement memory pooling for frequently allocated objects

## How to Use This Report

1. Review benchmark results to identify performance bottlenecks
2. Compare with previous results to track performance trends
3. Use profiling results to optimize code
4. Monitor key metrics in production environment

---

**Generated by:** Performance Test Script
**Next Test Date:** $(date -d "+1 week" +%Y-%m-%d)
EOF

    echo "✓ Performance report generated: $TEST_DIR/performance_report.md"
}

# 主测试流程
main() {
    # 检查依赖
    check_dependencies

    # 确保服务正在运行（如果需要）
    if ! curl -s http://localhost:8080/health >/dev/null 2>&1; then
        echo "⚠️  Application not running on localhost:8080"
        echo "Please start the application with: go run . or make build && ./bin/law-oa-go"
        echo "Running tests with mock services..."
    else
        echo "✓ Application is running"
    fi

    # 运行基准测试
    echo "Starting benchmark tests..."
    run_benchmark "BenchmarkAPIRequest"
    run_benchmark "BenchmarkDatabaseQueries"
    run_benchmark "BenchmarkRedisOperations"
    run_benchmark "BenchmarkCacheMiddleware"
    run_benchmark "BenchmarkMiddlewarePerformance"
    run_benchmark "BenchmarkConcurrentRequests"

    # 运行分析和负载测试
    run_memory_analysis
    run_cpu_analysis
    run_load_test

    # 生成报告
    generate_performance_report

    echo ""
    echo "=== Performance Tests Completed ==="
    echo "Results saved in: $TEST_DIR"
    echo "View the main report: $TEST_DIR/performance_report.md"
    echo ""
    echo "Next steps:"
    echo "1. Review the performance report"
    echo "2. Analyze any performance regressions"
    echo "3. Implement optimizations based on findings"
    echo "4. Schedule regular performance tests"

    # 创建最新结果的符号链接
    if [[ -L "$RESULT_DIR/latest" ]]; then
        rm "$RESULT_DIR/latest"
    fi
    ln -s "$TIMESTAMP" "$RESULT_DIR/latest"
    echo "Latest results linked to: $RESULT_DIR/latest"
}

# 运行主函数
main "$@"
