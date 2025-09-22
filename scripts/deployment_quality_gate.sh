#!/bin/bash

# Law OA Go 项目部署质量门禁检查脚本
# 在部署前检查各种质量指标，确保部署质量

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                 显示此帮助信息"
    echo "  -e, --env ENVIRONMENT      部署环境 (dev|staging|production)"
    echo "  -r, --report FILE          生成报告文件"
    echo "  -s, --strict               严格模式（任何门禁失败都停止）"
    echo "  --skip-coverage           跳过测试覆盖率检查"
    echo "  --skip-performance        跳过性能检查"
    echo "  --skip-security           跳过安全检查"
    echo "  --skip-fuzzing            跳过 Fuzzing 检查"
    echo "  --skip-pgo                跳过 PGO 检查"
    echo ""
    echo "环境质量标准:"
    echo "  - 开发环境(dev):"
    echo "    • 测试覆盖率 ≥ 60%"
    echo "    • Fuzzing crashes ≤ 10"
    echo "    • 内存分配 ≤ 2500 B/op"
    echo "    • 安全漏洞 ≤ 3 (中低危)"
    echo ""
    echo "  - 测试环境(staging):"
    echo "    • 测试覆盖率 ≥ 75%"
    echo "    • Fuzzing crashes ≤ 5"
    echo "    • 内存分配 ≤ 2000 B/op"
    echo "    • 安全漏洞 ≤ 1 (仅低危)"
    echo "    • PGO 性能提升 ≥ 3%"
    echo ""
    echo "  - 生产环境(production):"
    echo "    • 测试覆盖率 ≥ 85%"
    echo "    • Fuzzing crashes ≤ 2"
    echo "    • 内存分配 ≤ 1500 B/op"
    echo "    • 安全漏洞 = 0"
    echo "    • PGO 性能提升 ≥ 5%"
    echo ""
    echo "示例:"
    echo "  $0 -e staging -s                    # 测试环境严格模式"
    echo "  $0 -e production -r report.md      # 生产环境生成报告"
    echo "  $0 -e dev --skip-security          # 开发环境跳过安全检查"
}

# 检查测试覆盖率
check_test_coverage() {
    if [ "$SKIP_COVERAGE" = true ]; then
        log_warning "跳过测试覆盖率检查"
        add_report "⚠️" "测试覆盖率检查" "已跳过"
        return 0
    fi
    
    log_info "检查测试覆盖率..."
    
    local thresholds=($(get_quality_thresholds "$ENVIRONMENT"))
    local coverage_threshold=${thresholds[0]}
    
    # 检查覆盖率文件是否存在
    if [ ! -f "coverage.out" ]; then
        log_error "覆盖率文件不存在，请先运行测试"
        add_report "❌" "测试覆盖率检查" "覆盖率文件 coverage.out 不存在"
        return 1
    fi
    
    # 获取覆盖率
    local coverage=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
    local coverage_int=${coverage%.*}
    
    log_info "当前测试覆盖率: ${coverage}%"
    
    # 检查是否达到阈值
    if [ "$coverage_int" -ge "$coverage_threshold" ]; then
        log_success "测试覆盖率检查通过 (${coverage}% ≥ ${coverage_threshold}%)"
        add_report "✅" "测试覆盖率检查" "当前覆盖率: ${coverage}% (阈值: ${coverage_threshold}%)"
        return 0
    else
        log_error "测试覆盖率不足 (${coverage}% < ${coverage_threshold}%)"
        add_report "❌" "测试覆盖率检查" "当前覆盖率: ${coverage}% (阈值: ${coverage_threshold}%)"
        return 1
    fi
}

# 检查Fuzzing测试结果
check_fuzzing_results() {
    local environment=$1
    local max_crashes=""
    
    case $environment in
        dev)
            max_crashes=10
            ;;
        staging)
            max_crashes=5
            ;;
        production)
            max_crashes=2
            ;;
        *)
            max_crashes=10
            ;;
    esac
    
    log_info "检查Fuzzing测试结果 (要求: crashes < $max_crashes)..."
    
    # 检查Fuzzing结果目录
    local fuzzing_results="fuzzing_results"
    if [ ! -d "$fuzzing_results" ]; then
        log_warning "Fuzzing结果目录不存在: $fuzzing_results"
        echo "0"
        return 0
    fi
    
    # 统计crashes数量
    local crashes=$(find "$fuzzing_results" -name "*.crash" 2>/dev/null | wc -l)
    
    if [ "$crashes" -le "$max_crashes" ]; then
        log_success "Fuzzing crashes数量达标: $crashes (要求: <$max_crashes)"
        echo "$crashes"
        return 0
    else
        log_error "Fuzzing crashes数量不达标: $crashes (要求: <$max_crashes)"
        echo "$crashes"
        return 1
    fi
}

# 检查PGO性能提升
check_pgo_performance() {
    local environment=$1
    local required_improvement=5
    
    # 只有生产环境需要PGO性能检查
    if [ "$environment" != "production" ]; then
        log_info "跳过PGO性能检查 (非生产环境)"
        echo "0"
        return 0
    fi
    
    log_info "检查PGO性能提升 (要求: >${required_improvement}%)..."
    
    # 检查PGO基准测试结果
    local baseline_results="pgo_results/baseline.json"
    local pgo_results="pgo_results/pgo.json"
    
    if [ ! -f "$baseline_results" ] || [ ! -f "$pgo_results" ]; then
        log_warning "PGO性能测试结果不存在，跳过检查"
        echo "0"
        return 0
    fi
    
    # 读取基准测试结果
    local baseline_rps=$(jq -r '.requests_per_second // 0' "$baseline_results" 2>/dev/null || echo "0")
    local pgo_rps=$(jq -r '.requests_per_second // 0' "$pgo_results" 2>/dev/null || echo "0")
    
    if [ "$baseline_rps" = "0" ] || [ "$pgo_rps" = "0" ]; then
        log_warning "PGO性能测试数据无效，跳过检查"
        echo "0"
        return 0
    fi
    
    # 计算性能提升
    local improvement=$(echo "scale=2; ($pgo_rps - $baseline_rps) * 100 / $baseline_rps" | bc -l 2>/dev/null || echo "0")
    
    if (( $(echo "$improvement >= $required_improvement" | bc -l) )); then
        log_success "PGO性能提升达标: ${improvement}% (要求: >${required_improvement}%)"
        echo "$improvement"
        return 0
    else
        log_error "PGO性能提升不达标: ${improvement}% (要求: >${required_improvement}%)"
        echo "$improvement"
        return 1
    fi
}

# 检查安全扫描结果
check_security_scan() {
    local environment=$1
    
    log_info "检查安全扫描结果..."
    
    # 检查安全扫描报告
    local security_report="security_scan_report.json"
    if [ ! -f "$security_report" ]; then
        log_warning "安全扫描报告不存在: $security_report"
        echo "0"
        return 0
    fi
    
    # 检查高危漏洞
    local critical_vulnerabilities=$(jq -r '.critical // 0' "$security_report" 2>/dev/null || echo "0")
    local high_vulnerabilities=$(jq -r '.high // 0' "$security_report" 2>/dev/null || echo "0")
    
    if [ "$environment" = "production" ] && [ "$critical_vulnerabilities" -gt 0 ]; then
        log_error "生产环境发现高危漏洞: $critical_vulnerabilities 个"
        echo "$critical_vulnerabilities"
        return 1
    fi
    
    if [ "$high_vulnerabilities" -gt 5 ]; then
        log_warning "发现较多高危漏洞: $high_vulnerabilities 个"
    fi
    
    log_success "安全扫描通过"
    echo "$((critical_vulnerabilities + high_vulnerabilities))"
    return 0
}

# 检查构建质量
check_build_quality() {
    local environment=$1
    
    log_info "检查构建质量..."
    
    # 检查构建是否成功
    if [ ! -f "law-oa-go" ] && [ ! -f "law-oa-go.exe" ]; then
        log_error "构建产物不存在"
        return 1
    fi
    
    # 检查二进制文件大小
    local binary_size=$(stat -c%s "law-oa-go" 2>/dev/null || stat -c%s "law-oa-go.exe" 2>/dev/null || echo "0")
    local max_size=52428800  # 50MB
    
    if [ "$binary_size" -gt "$max_size" ]; then
        log_warning "二进制文件过大: $binary_size bytes (建议: <$max_size bytes)"
    fi
    
    log_success "构建质量检查通过"
    echo "$binary_size"
    return 0
}

# 生成质量报告
generate_quality_report() {
    local environment=$1
    local report_file=$2
    local coverage=$3
    local crashes=$4
    local pgo_improvement=$5
    local security_issues=$6
    local build_size=$7
    
    log_info "生成质量报告: $report_file"
    
    cat > "$report_file" << EOF
# 部署质量门禁报告

## 基本信息
- **检查时间**: $(date)
- **部署环境**: $environment
- **检查工具**: 部署质量门禁检查器 v1.0

## 质量检查结果

### 测试覆盖率
- **实际值**: ${coverage}%
- **状态**: $([ "$coverage" != "ERROR" ] && echo "✅ 通过" || echo "❌ 失败")

### Fuzzing测试
- **Crashes数量**: $crashes
- **状态**: $([ "$crashes" != "ERROR" ] && echo "✅ 通过" || echo "❌ 失败")

### PGO性能提升
- **性能提升**: ${pgo_improvement}%
- **状态**: $([ "$pgo_improvement" != "ERROR" ] && echo "✅ 通过" || echo "❌ 失败")

### 安全扫描
- **安全问题数**: $security_issues
- **状态**: $([ "$security_issues" != "ERROR" ] && echo "✅ 通过" || echo "❌ 失败")

### 构建质量
- **二进制大小**: $build_size bytes
- **状态**: $([ "$build_size" != "ERROR" ] && echo "✅ 通过" || echo "❌ 失败")

## 总体评估
- **部署建议**: $([ "$coverage" != "ERROR" ] && [ "$crashes" != "ERROR" ] && [ "$pgo_improvement" != "ERROR" ] && [ "$security_issues" != "ERROR" ] && [ "$build_size" != "ERROR" ] && echo "🟢 允许部署" || echo "🔴 阻止部署")

## 详细信息
EOF
    
    # 添加Fuzzing crash详情
    if [ -d "fuzzing_results" ]; then
        echo "" >> "$report_file"
        echo "### Fuzzing Crashes详情" >> "$report_file"
        find "fuzzing_results" -name "*.crash" 2>/dev/null | head -10 | while read crash_file; do
            echo "- $(basename "$crash_file")" >> "$report_file"
        done
    fi
    
    # 添加安全漏洞详情
    if [ -f "security_scan_report.json" ]; then
        echo "" >> "$report_file"
        echo "### 安全漏洞详情" >> "$report_file"
        jq -r '.vulnerabilities[]? | "- \(.type): \(.severity)"' "security_scan_report.json" 2>/dev/null >> "$report_file"
    fi
    
    log_success "质量报告已生成: $report_file"
}

# 初始化变量
ENVIRONMENT=""
REPORT_FILE=""
STRICT_MODE=false
SKIP_COVERAGE=false
SKIP_PERFORMANCE=false
SKIP_SECURITY=false
SKIP_FUZZING=false
SKIP_PGO=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -r|--report)
            REPORT_FILE="$2"
            shift 2
            ;;
        -s|--strict)
            STRICT_MODE=true
            shift
            ;;
        --skip-coverage)
            SKIP_COVERAGE=true
            shift
            ;;
        --skip-performance)
            SKIP_PERFORMANCE=true
            shift
            ;;
        --skip-security)
            SKIP_SECURITY=true
            shift
            ;;
        --skip-fuzzing)
            SKIP_FUZZING=true
            shift
            ;;
        --skip-pgo)
            SKIP_PGO=true
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_usage
            exit 1
            ;;
    esac
done

# 验证环境参数
if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|production)$ ]]; then
    log_error "无效的环境参数: $ENVIRONMENT"
    echo "支持的环境: dev, staging, production"
    exit 1
fi

# 初始化报告
init_report() {
    local report_file=$1
    
    cat > "$report_file" << EOF
# Law OA Go 部署质量门禁报告

**检查时间:** $(date '+%Y-%m-%d %H:%M:%S')  
**部署环境:** $ENVIRONMENT  
**严格模式:** $STRICT_MODE  

## 质量标准

| 检查项 | 开发环境 | 测试环境 | 生产环境 |
|--------|----------|----------|----------|
| 测试覆盖率 | ≥ 60% | ≥ 75% | ≥ 85% |
| Fuzzing Crashes | ≤ 10 | ≤ 5 | ≤ 2 |
| 内存分配 | ≤ 2500 B/op | ≤ 2000 B/op | ≤ 1500 B/op |
| 安全漏洞 | ≤ 3 (中低危) | ≤ 1 (低危) | 0 |
| PGO 性能提升 | 不要求 | ≥ 3% | ≥ 5% |

## 检查结果

EOF
}

# 添加报告内容
add_report() {
    local status=$1
    local title=$2
    local content=$3
    
    if [ -n "$REPORT_FILE" ]; then
        echo "### $status: $title" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "$content" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    fi
}

# 获取环境质量标准
get_quality_thresholds() {
    local env=$1
    
    case $env in
        dev)
            echo "60 10 2500 3 0"
            ;;
        staging)
            echo "75 5 2000 1 3"
            ;;
        production)
            echo "85 2 1500 0 5"
            ;;
        *)
            log_error "未知环境: $env"
            exit 1
            ;;
    esac
}

# 生成总结报告
generate_summary() {
    local total_checks=$1
    local passed_checks=$2
    local failed_checks=$3
    local skipped_checks=$4
    
    if [ -n "$REPORT_FILE" ]; then
        cat >> "$REPORT_FILE" << EOF

## 总结

- **总检查项:** $total_checks
- **通过项:** $passed_checks
- **失败项:** $failed_checks
- **跳过项:** $skipped_checks
- **通过率:** $(( passed_checks * 100 / total_checks ))%

## 建议

EOF
        
        if [ "$failed_checks" -gt 0 ]; then
            cat >> "$REPORT_FILE" << EOF
发现 $failed_checks 个检查项未通过，建议：

1. 修复失败的检查项
2. 重新运行测试
3. 考虑使用 \`--skip-\` 选项跳过非关键检查
4. 在开发环境中验证修复效果

EOF
        else
            cat >> "$REPORT_FILE" << EOF
所有检查项均通过，可以继续部署流程。

EOF
        fi
        
        cat >> "$REPORT_FILE" << EOF
---
*报告生成时间: $(date '+%Y-%m-%d %H:%M:%S')*  
*部署环境: $ENVIRONMENT*  
*严格模式: $STRICT_MODE*

EOF
    fi
}

# 主函数
main() {
    log_info "开始 $ENVIRONMENT 环境部署质量门禁检查..."
    
    # 初始化报告
    if [ -n "$REPORT_FILE" ]; then
        init_report "$REPORT_FILE"
    fi
    
    # 定义检查函数列表
    local check_functions=(
        "check_test_coverage"
        "check_fuzzing_results"
        "check_pgo_performance"
        "check_security_scan"
        "check_build_quality"
    )
    
    local total_checks=${#check_functions[@]}
    local passed_checks=0
    local failed_checks=0
    local skipped_checks=0
    
    # 执行所有检查
    for check_func in "${check_functions[@]}"; do
        log_info "执行检查: $check_func"
        
        if $check_func "$ENVIRONMENT"; then
            ((passed_checks++))
        else
            ((failed_checks++))
            
            # 严格模式下立即退出
            if [ "$STRICT_MODE" = true ]; then
                log_error "严格模式下检查失败，停止部署流程"
                generate_summary $total_checks $passed_checks $failed_checks $skipped_checks
                exit 1
            fi
        fi
    done
    
    # 生成总结报告
    generate_summary $total_checks $passed_checks $failed_checks $skipped_checks
    
    # 输出最终结果
    if [ "$failed_checks" -eq 0 ]; then
        log_success "所有质量门禁检查通过！"
        log_info "检查结果: $passed_checks/$total_checks 通过，$skipped_checks 跳过"
        
        if [ -n "$REPORT_FILE" ]; then
            log_success "质量门禁报告已生成: $REPORT_FILE"
        fi
        
        exit 0
    else
        log_error "质量门禁检查失败！"
        log_info "检查结果: $passed_checks 通过, $failed_checks 失败, $skipped_checks 跳过"
        
        if [ -n "$REPORT_FILE" ]; then
            log_success "质量门禁报告已生成: $REPORT_FILE"
        fi
        
        exit 1
    fi
}

# 运行主函数
main "$@"