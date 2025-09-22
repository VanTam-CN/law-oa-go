#!/bin/bash

# Fuzzing测试运行脚本
# 用于运行Go 1.23的模糊测试，发现潜在的安全漏洞和稳定性问题

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help          显示此帮助信息"
    echo "  -a, --all           运行所有Fuzzing测试"
    echo "  -s, --security      运行安全相关Fuzzing测试"
    echo "  -v, --validators    运行验证器Fuzzing测试"
    echo "  -r, --repositories  运行数据库Fuzzing测试"
    echo "  -c, --cache         运行缓存Fuzzing测试"
    echo "  -n, --concurrency   运行并发Fuzzing测试"
    echo "  -t, --time DURATION 设置Fuzzing运行时间 (默认: 30s)"
    echo "  -f, --fuzzers COUNT  并行Fuzzing进程数 (默认: CPU核心数)"
    echo "  -o, --output DIR     输出目录 (默认: fuzzing-results)"
    echo ""
    echo "示例:"
    echo "  $0 -a                    # 运行所有Fuzzing测试"
    echo "  $0 -s -t 60s            # 运行安全测试60秒"
    echo "  $0 -v -f 4              # 用4个进程运行验证器测试"
    echo "  $0 -r -o custom-output  # 运行数据库测试到自定义目录"
}

# 清理函数
cleanup() {
    log_info "清理临时文件..."
    # 清理fuzzing缓存
    go clean -fuzzcache
    log_success "清理完成"
}

# 运行特定的Fuzzing测试
run_fuzz_test() {
    local test_pattern=$1
    local output_dir=$2
    local fuzz_time=$3
    local fuzzers=$4
    
    log_info "运行Fuzzing测试: $test_pattern"
    log_info "输出目录: $output_dir"
    log_info "运行时间: $fuzz_time"
    log_info "并行进程数: $fuzzers"
    
    # 创建输出目录
    mkdir -p "$output_dir"
    
    # 运行Fuzzing测试
    if [ "$fuzzers" -eq 1 ]; then
        # 单进程运行
        go test -fuzz="$test_pattern" -fuzztime="$fuzz_time" -v ./... 2>&1 | tee "$output_dir/fuzz-$test_pattern.log"
    else
        # 多进程运行
        for ((i=1; i<=fuzzers; i++)); do
            log_info "启动Fuzzing进程 $i/$fuzzers"
            go test -fuzz="$test_pattern" -fuzztime="$fuzz_time" -v ./... > "$output_dir/fuzz-$test_pattern-worker-$i.log" 2>&1 &
        done
        
        # 等待所有进程完成
        wait
    fi
    
    # 检查是否有crashers
    local crasher_count=$(find "$output_dir" -name "crashers*" 2>/dev/null | wc -l)
    if [ "$crasher_count" -gt 0 ]; then
        log_error "发现 $crasher_count 个crasher文件！"
        find "$output_dir" -name "crashers*" -exec echo "发现crasher文件: {}" \;
    else
        log_success "未发现crasher文件"
    fi
    
    # 检查是否有suppressions
    local suppression_count=$(find "$output_dir" -name "suppressions*" 2>/dev/null | wc -l)
    if [ "$suppression_count" -gt 0 ]; then
        log_warning "发现 $suppression_count 个suppression文件"
    fi
}

# 生成Fuzzing报告
generate_report() {
    local output_dir=$1
    local report_file="$output_dir/fuzzing-report.md"
    
    log_info "生成Fuzzing测试报告..."
    
    {
        echo "# Fuzzing测试报告"
        echo ""
        echo "## 测试信息"
        echo "- 测试时间: $(date)"
        echo "- Go版本: $(go version)"
        echo "- 输出目录: $output_dir"
        echo ""
        
        echo "## 测试结果"
        echo ""
        
        # 统计各个测试的结果
        for pattern in security validators repositories cache concurrency; do
            if [ -f "$output_dir/fuzz-$pattern.log" ]; then
                echo "### $pattern 模块"
                echo ""
                
                # 统计执行次数
                local executions=$(grep -c "fuzzing" "$output_dir/fuzz-$pattern.log" 2>/dev/null || echo "0")
                echo "- 执行次数: $executions"
                
                # 检查crashers
                local crashers=$(find "$output_dir" -name "*crashers*" 2>/dev/null | wc -l)
                if [ "$crashers" -gt 0 ]; then
                    echo "- 发现crashers: $crashers ❌"
                else
                    echo "- 发现crashers: 0 ✅"
                fi
                
                # 检查suppressions
                local suppressions=$(find "$output_dir" -name "*suppressions*" 2>/dev/null | wc -l)
                if [ "$suppressions" -gt 0 ]; then
                    echo "- 发现suppressions: $suppressions ⚠️"
                else
                    echo "- 发现suppressions: 0 ✅"
                fi
                
                echo ""
            fi
        done
        
        echo "## 发现的问题"
        echo ""
        
        # 列出所有crasher文件
        find "$output_dir" -name "*crashers*" 2>/dev/null | while read -r crasher; do
            echo "### $(basename "$crasher")"
            echo ""
            echo "文件路径: $crasher"
            echo ""
            echo "\`\`\`"
            cat "$crasher" 2>/dev/null | head -20
            echo "..."
            echo "\`\`\`"
            echo ""
        done
        
        echo "## 建议"
        echo ""
        echo "1. **定期运行Fuzzing测试**: 在CI/CD中集成Fuzzing测试，定期运行"
        echo "2. **监控crashers**: 及时分析和修复发现的crashers"
        echo "3. **增加种子语料库**: 根据实际使用场景增加更多的种子输入"
        echo "4. **更新测试用例**: 根据Fuzzing结果增加常规测试用例"
        echo "5. **性能监控**: 监控Fuzzing测试的执行性能和资源使用"
        
    } > "$report_file"
    
    log_success "Fuzzing报告已生成: $report_file"
}

# 主函数
main() {
    # 默认参数
    FUZZ_TIME="30s"
    FUZZERS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
    OUTPUT_DIR="fuzzing-results"
    RUN_ALL=false
    RUN_SECURITY=false
    RUN_VALIDATORS=false
    RUN_REPOSITORIES=false
    RUN_CACHE=false
    RUN_CONCURRENCY=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -a|--all)
                RUN_ALL=true
                shift
                ;;
            -s|--security)
                RUN_SECURITY=true
                shift
                ;;
            -v|--validators)
                RUN_VALIDATORS=true
                shift
                ;;
            -r|--repositories)
                RUN_REPOSITORIES=true
                shift
                ;;
            -c|--cache)
                RUN_CACHE=true
                shift
                ;;
            -n|--concurrency)
                RUN_CONCURRENCY=true
                shift
                ;;
            -t|--time)
                FUZZ_TIME="$2"
                shift 2
                ;;
            -f|--fuzzers)
                FUZZERS="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            *)
                log_error "未知参数: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # 如果没有指定任何测试，默认运行全部
    if [ "$RUN_ALL" = false ] && [ "$RUN_SECURITY" = false ] && [ "$RUN_VALIDATORS" = false ] && \
       [ "$RUN_REPOSITORIES" = false ] && [ "$RUN_CACHE" = false ] && [ "$RUN_CONCURRENCY" = false ]; then
        RUN_ALL=true
    fi
    
    log_info "开始Fuzzing测试..."
    log_info "Go版本: $(go version)"
    log_info "Fuzzing时间: $FUZZ_TIME"
    log_info "并行进程数: $FUZZERS"
    log_info "输出目录: $OUTPUT_DIR"
    
    # 创建输出目录
    mkdir -p "$OUTPUT_DIR"
    
    # 设置退出时清理
    trap cleanup EXIT
    
    # 运行测试
    if [ "$RUN_ALL" = true ] || [ "$RUN_SECURITY" = true ]; then
        run_fuzz_test "Fuzz_JWTKeyManager_ValidateToken" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
    fi
    
    if [ "$RUN_ALL" = true ] || [ "$RUN_VALIDATORS" = true ]; then
        run_fuzz_test "Fuzz_CaseValidator_Validate" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_ClientValidator_Validate" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_LawyerValidator_Validate" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
    fi
    
    if [ "$RUN_ALL" = true ] || [ "$RUN_REPOSITORIES" = true ]; then
        run_fuzz_test "Fuzz_QueryBuilder_Where" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_QueryBuilder_Like" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_QueryBuilder_Order" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
    fi
    
    if [ "$RUN_ALL" = true ] || [ "$RUN_CACHE" = true ]; then
        run_fuzz_test "Fuzz_CacheService_SetAndGet" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_CacheService_SetWithExpiration" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_CacheService_ConcurrentAccess" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_LayeredCache_Get" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
    fi
    
    if [ "$RUN_ALL" = true ] || [ "$RUN_CONCURRENCY" = true ]; then
        run_fuzz_test "Fuzz_WorkerPool_SubmitTask" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_CircuitBreaker_Execute" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_RateLimiter_Allow" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
        run_fuzz_test "Fuzz_ConcurrentService_SubmitTask" "$OUTPUT_DIR" "$FUZZ_TIME" "$FUZZERS"
    fi
    
    # 生成报告
    generate_report "$OUTPUT_DIR"
    
    log_success "Fuzzing测试完成！"
    log_info "结果保存在: $OUTPUT_DIR"
    log_info "查看报告: cat $OUTPUT_DIR/fuzzing-report.md"
}

# 运行主函数
main "$@"