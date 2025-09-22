#!/bin/bash

# Fuzzing Crash分析器
# 自动分析和分类Fuzzing测试发现的Crashers，生成修复建议

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

log_debug() {
    echo -e "${PURPLE}[DEBUG]${NC} $1"
}

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                 显示此帮助信息"
    echo "  -a, --analyze DIR          分析Crashers目录"
    echo "  -c, --crashers DIR         指定Crashers文件目录"
    echo "  -o, --output DIR           输出报告目录 (默认: crash_reports)"
    echo "  -r, --report FORMAT        报告格式 (markdown|json|html, 默认: markdown)"
    echo "  -t, --test PATH            运行指定的Fuzzing测试"
    echo "  -d, --debug                启用调试模式"
    echo "  -f, --fix-suggestions      生成修复建议"
    echo "  -j, --jira                 生成Jira工单模板"
    echo "  -s, --severity LEVEL       严重程度过滤 (critical|high|medium|low)"
    echo ""
    echo "示例:"
    echo "  $0 -a fuzzing-results/              # 分析Fuzzing结果"
    echo "  $0 -t ./internal/security/ -r json # 生成JSON报告"
    echo "  $0 -f -j                           # 生成修复建议和Jira模板"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装"
        exit 1
    fi
    
    # 检查jq (用于JSON处理)
    if ! command -v jq &> /dev/null; then
        log_warning "jq 未安装，JSON报告功能将受限"
    fi
    
    log_success "依赖检查通过"
}

# 分析Crashers目录
analyze_crashers_directory() {
    local crash_dir="$1"
    local output_dir="$2"
    local report_format="${3:-markdown}"
    
    log_info "分析Crashers目录: $crash_dir"
    
    if [ ! -d "$crash_dir" ]; then
        log_error "Crashers目录不存在: $crash_dir"
        return 1
    fi
    
    # 创建输出目录
    mkdir -p "$output_dir"
    
    # 查找所有crashers文件
    local crash_files=$(find "$crash_dir" -name "crashers" -type f 2>/dev/null || true)
    
    if [ -z "$crash_files" ]; then
        log_warning "未发现crashers文件"
        echo "✅ 未发现Fuzzing Crashers，所有测试通过！" > "$output_dir/no_crashers.txt"
        return 0
    fi
    
    log_info "发现 $(echo "$crash_files" | wc -l | tr -d ' ') 个crashers文件"
    
    # 分析每个crashers文件
    local total_crashes=0
    local critical_crashes=0
    local high_severity_crashes=0
    
    for crash_file in $crash_files; do
        local component=$(basename "$(dirname "$crash_file")")
        log_info "分析组件: $component"
        
        # 分析单个crashers文件
        local analysis_result=$(analyze_single_crasher "$crash_file" "$component" "$output_dir")
        local crash_count=$(echo "$analysis_result" | jq -r '.crash_count // 0' 2>/dev/null || echo "0")
        local critical_count=$(echo "$analysis_result" | jq -r '.critical_count // 0' 2>/dev/null || echo "0")
        local high_count=$(echo "$analysis_result" | jq -r '.high_count // 0' 2>/dev/null || echo "0")
        
        total_crashes=$((total_crashes + crash_count))
        critical_crashes=$((critical_crashes + critical_count))
        high_severity_crashes=$((high_severity_crashes + high_count))
        
        log_debug "组件 $component: $crash_count 个crashes, $critical_count 个critical, $high_count 个high"
    done
    
    # 生成汇总报告
    generate_summary_report "$output_dir" "$total_crashes" "$critical_crashes" "$high_severity_crashes" "$report_format"
    
    log_success "Crash分析完成，发现 $total_crashes 个crashes"
}

# 分析单个crashers文件
analyze_single_crasher() {
    local crash_file="$1"
    local component="$2"
    local output_dir="$3"
    
    local component_dir="$output_dir/$component"
    mkdir -p "$component_dir"
    
    # 复制crashers文件
    cp "$crash_file" "$component_dir/"
    
    # 分析crashes
    local crash_count=0
    local critical_count=0
    local high_count=0
    local medium_count=0
    local low_count=0
    
    # 临时存储分析结果
    local temp_analysis="$component_dir/analysis.json"
    echo '{"crashes": []}' > "$temp_analysis"
    
    # 逐行分析crashes
    while IFS= read -r crash_input; do
        if [ -n "$crash_input" ]; then
            crash_count=$((crash_count + 1))
            
            # 分析crash类型和严重程度
            local crash_analysis=$(analyze_crash_input "$crash_input" "$component")
            local crash_type=$(echo "$crash_analysis" | jq -r '.type // "unknown"' 2>/dev/null || echo "unknown")
            local severity=$(echo "$crash_analysis" | jq -r '.severity // "medium"' 2>/dev/null || echo "medium")
            local description=$(echo "$crash_analysis" | jq -r '.description // "Unknown crash"' 2>/dev/null || echo "Unknown crash")
            
            # 统计严重程度
            case "$severity" in
                critical) critical_count=$((critical_count + 1)) ;;
                high) high_count=$((high_count + 1)) ;;
                medium) medium_count=$((medium_count + 1)) ;;
                low) low_count=$((low_count + 1)) ;;
            esac
            
            # 保存crash详情
            local crash_data="$component_dir/crash_$crash_count.txt"
            echo "输入: $crash_input" > "$crash_data"
            echo "类型: $crash_type" >> "$crash_data"
            echo "严重程度: $severity" >> "$crash_data"
            echo "描述: $description" >> "$crash_data"
            echo "时间: $(date)" >> "$crash_data"
            
            # 更新JSON分析结果
            jq --arg input "$crash_input" \
               --arg type "$crash_type" \
               --arg severity "$severity" \
               --arg description "$description" \
               --arg number "$crash_count" \
               '.crashes += [{
                   "number": ($number | tonumber),
                   "input": $input,
                   "type": $type,
                   "severity": $severity,
                   "description": $description,
                   "component": "'"$component"'"
               }]' "$temp_analysis" > "${temp_analysis}.tmp" && mv "${temp_analysis}.tmp" "$temp_analysis"
        fi
    done < "$crash_file"
    
    # 生成组件报告
    generate_component_report "$component" "$component_dir" "$crash_count" "$critical_count" "$high_count" "$medium_count" "$low_count"
    
    # 返回分析结果
    cat << EOF
{
    "component": "$component",
    "crash_count": $crash_count,
    "critical_count": $critical_count,
    "high_count": $high_count,
    "medium_count": $medium_count,
    "low_count": $low_count
}
EOF
}

# 分析单个crash输入
analyze_crash_input() {
    local input="$1"
    local component="$2"
    
    local crash_type="unknown"
    local severity="medium"
    local description="Unknown crash"
    
    # 根据输入特征和组件类型进行分析
    case "$component" in
        *security*|*jwt*)
            if [[ "$input" == *"panic"* ]]; then
                crash_type="panic"
                severity="high"
                description="JWT验证导致panic"
            elif [[ "$input" == *"index out of range"* ]]; then
                crash_type="index_out_of_range"
                severity="medium"
                description="JWT验证索引越界"
            elif [[ "$input" == *"nil pointer"* ]]; then
                crash_type="nil_pointer"
                severity="high"
                description="JWT验证空指针解引用"
            elif [[ "$input" == *"invalid memory"* ]]; then
                crash_type="memory_error"
                severity="critical"
                description="JWT验证内存访问错误"
            elif [[ "$input" == *"JWT"* ]] || [[ "$input" == *"token"* ]]; then
                crash_type="jwt_validation"
                severity="medium"
                description="JWT令牌验证错误"
            else
                crash_type="security_validation"
                severity="medium"
                description="安全验证错误"
            fi
            ;;
            
        *validator*|*validation*)
            if [[ "$input" == *"JSON"* ]] || [[ "$input" == *"{"* ]]; then
                crash_type="json_parsing"
                severity="medium"
                description="JSON解析错误"
            elif [[ "$input" == *"panic"* ]]; then
                crash_type="panic"
                severity="high"
                description="验证器panic"
            elif [[ "$input" == *"index out of range"* ]]; then
                crash_type="index_out_of_range"
                severity="medium"
                description="验证器索引越界"
            elif [[ "$input" == *"XSS"* ]] || [[ "$input" == *"<script"* ]]; then
                crash_type="xss_vulnerability"
                severity="high"
                description="XSS漏洞检测"
            elif [[ "$input" == *"SQL"* ]] || [[ "$input" == *"'"* ]]; then
                crash_type="sql_injection"
                severity="high"
                description="SQL注入检测"
            else
                crash_type="input_validation"
                severity="medium"
                description="输入验证错误"
            fi
            ;;
            
        *database*|*query*)
            if [[ "$input" == *"SQL"* ]] || [[ "$input" == *"'"* ]]; then
                crash_type="sql_injection"
                severity="critical"
                description="SQL注入漏洞"
            elif [[ "$input" == *"panic"* ]]; then
                crash_type="panic"
                severity="high"
                description="数据库操作panic"
            elif [[ "$input" == *"syntax error"* ]]; then
                crash_type="query_syntax"
                severity="medium"
                description="查询语法错误"
            else
                crash_type="database_error"
                severity="medium"
                description="数据库操作错误"
            fi
            ;;
            
        *cache*|*memory*)
            if [[ "$input" == *"panic"* ]]; then
                crash_type="panic"
                severity="high"
                description="缓存操作panic"
            elif [[ "$input" == *"memory"* ]] || [[ "$input" == *"allocation"* ]]; then
                crash_type="memory_error"
                severity="critical"
                description="内存管理错误"
            elif [[ "$input" == *"concurrent"* ]]; then
                crash_type="race_condition"
                severity="high"
                description="并发访问竞态条件"
            else
                crash_type="cache_error"
                severity="medium"
                description="缓存操作错误"
            fi
            ;;
            
        *concurrency*|*worker*)
            if [[ "$input" == *"panic"* ]]; then
                crash_type="panic"
                severity="high"
                description="并发操作panic"
            elif [[ "$input" == *"deadlock"* ]] || [[ "$input" == *"lock"* ]]; then
                crash_type="deadlock"
                severity="critical"
                description="死锁错误"
            elif [[ "$input" == *"race"* ]] || [[ "$input" == *"concurrent"* ]]; then
                crash_type="race_condition"
                severity="high"
                description="竞态条件"
            elif [[ "$input" == *"timeout"* ]]; then
                crash_type="timeout"
                severity="medium"
                description="超时错误"
            else
                crash_type="concurrency_error"
                severity="medium"
                description="并发操作错误"
            fi
            ;;
            
        *)
            # 通用分析
            if [[ "$input" == *"panic"* ]]; then
                crash_type="panic"
                severity="high"
                description="程序panic"
            elif [[ "$input" == *"index out of range"* ]]; then
                crash_type="index_out_of_range"
                severity="medium"
                description="索引越界"
            elif [[ "$input" == *"nil pointer"* ]]; then
                crash_type="nil_pointer"
                severity="high"
                description="空指针解引用"
            elif [[ "$input" == *"invalid memory"* ]]; then
                crash_type="memory_error"
                severity="critical"
                description="内存访问错误"
            elif [[ "$input" == *"segmentation fault"* ]]; then
                crash_type="segmentation_fault"
                severity="critical"
                description="段错误"
            else
                crash_type="unknown"
                severity="medium"
                description="未知错误"
            fi
            ;;
    esac
    
    # 根据输入长度调整严重程度
    local input_length=${#input}
    if [ "$input_length" -gt 10000 ]; then
        # 超长输入可能导致拒绝服务
        if [ "$severity" != "critical" ]; then
            severity="high"
            fi
        description="$description (超长输入: $input_length 字符)"
    fi
    
    cat << EOF
{
    "type": "$crash_type",
    "severity": "$severity",
    "description": "$description",
    "input_length": $input_length
}
EOF
}

# 生成组件报告
generate_component_report() {
    local component="$1"
    local component_dir="$2"
    local crash_count="$3"
    local critical_count="$4"
    local high_count="$5"
    local medium_count="$6"
    local low_count="$7"
    
    local report_file="$component_dir/report.md"
    
    cat > "$report_file" << EOF
# $component Fuzzing Crash分析报告

## 基本信息
- **分析时间**: $(date)
- **组件名称**: $component
- **总Crash数量**: $crash_count
- **严重程度分布**:
  - Critical: $critical_count
  - High: $high_count
  - Medium: $medium_count
  - Low: $low_count

## 严重程度分析

### Critical (严重)
- **数量**: $critical_count
- **影响**: 可能导致系统崩溃或安全漏洞
- **建议**: 立即修复

### High (高)
- **数量**: $high_count
- **影响**: 可能导致功能异常或安全隐患
- **建议**: 优先修复

### Medium (中等)
- **数量**: $medium_count
- **影响**: 可能导致用户体验问题
- **建议**: 计划修复

### Low (低)
- **数量**: $low_count
- **影响**: 影响较小
- **建议**: 视情况修复

## 详细Crash列表

EOF

    # 添加详细crash信息
    for crash_file in "$component_dir"/crash_*.txt; do
        if [ -f "$crash_file" ]; then
            local crash_num=$(basename "$crash_file" .txt | sed 's/crash_//')
            echo "### Crash $crash_num" >> "$report_file"
            echo "" >> "$report_file"
            echo "\`\`\`" >> "$report_file"
            cat "$crash_file" >> "$report_file"
            echo "\`\`\`" >> "$report_file"
            echo "" >> "$report_file"
        fi
    done
    
    # 添加修复建议
    cat >> "$report_file" << EOF

## 修复建议

### 立即修复 (Critical)
1. **输入验证**: 添加严格的输入长度和格式验证
2. **错误处理**: 实现完善的错误处理机制
3. **内存管理**: 检查指针和内存访问
4. **安全检查**: 增强安全相关的边界检查

### 优先修复 (High)
1. **边界检查**: 添加数组索引和边界检查
2. **异常处理**: 改进异常情况处理逻辑
3. **输入过滤**: 增强输入数据的过滤和清理
4. **并发安全**: 改进并发操作的安全性

### 计划修复 (Medium)
1. **用户体验**: 优化错误提示和用户体验
2. **日志记录**: 改进错误日志记录
3. **监控告警**: 添加监控和告警机制

## 后续步骤

1. **立即修复**: Critical和High级别的crashes
2. **测试验证**: 修复后重新运行Fuzzing测试
3. **监控部署**: 在生产环境监控修复效果
4. **预防措施**: 改进开发流程，防止类似问题

EOF
}

# 生成汇总报告
generate_summary_report() {
    local output_dir="$1"
    local total_crashes="$2"
    local critical_crashes="$3"
    local high_severity_crashes="$4"
    local report_format="$5"
    
    local summary_file="$output_dir/summary.$report_format"
    
    case "$report_format" in
        "json")
            generate_json_summary "$summary_file" "$total_crashes" "$critical_crashes" "$high_severity_crashes"
            ;;
        "html")
            generate_html_summary "$summary_file" "$total_crashes" "$critical_crashes" "$high_severity_crashes"
            ;;
        "markdown"|*)
            generate_markdown_summary "$summary_file" "$total_crashes" "$critical_crashes" "$high_severity_crashes"
            ;;
    esac
    
    log_success "汇总报告已生成: $summary_file"
}

# 生成Markdown汇总报告
generate_markdown_summary() {
    local report_file="$1"
    local total_crashes="$2"
    local critical_crashes="$3"
    local high_severity_crashes="$4"
    
    cat > "$report_file" << EOF
# Fuzzing Crash分析汇总报告

## 执行摘要
- **分析时间**: $(date)
- **总Crash数量**: $total_crashes
- **严重Crashes**: $critical_crashes
- **高优先级Crashes**: $high_severity_crashes

## 严重程度分布

| 严重程度 | 数量 | 百分比 | 紧急程度 |
|---------|------|--------|----------|
| Critical | $critical_crashes | $((critical_crashes * 100 / total_crashes))% | 🔴 立即修复 |
| High | $high_severity_crashes | $((high_severity_crashes * 100 / total_crashes))% | 🟡 优先修复 |
| Medium | $((total_crashes - critical_crashes - high_severity_crashes)) | $(((total_crashes - critical_crashes - high_severity_crashes) * 100 / total_crashes))% | 🟢 计划修复 |

## 风险评估

### 高风险组件
$(find "$output_dir" -name "report.md" | head -5 | while read report; do
    local component=$(basename "$(dirname "$report")")
    local critical=$(grep "Critical:" "$report" | awk '{print $2}' || echo "0")
    if [ "$critical" -gt 0 ]; then
        echo "- $component (Critical: $critical)"
    fi
done)

### 建议优先级

1. **立即处理**: Critical级别的crashes
2. **本周修复**: High级别的crashes
3. **下个Sprint**: Medium级别的crashes

## 行动计划

### 第1天: Critical修复
- 修复所有Critical级别的crashes
- 部署到测试环境验证
- 运行Fuzzing测试确认修复

### 第2-3天: High修复
- 修复High级别的crashes
- 代码审查和测试
- 集成测试验证

### 第1周: 部署监控
- 部署修复到生产环境
- 添加监控和告警
- 定期运行Fuzzing测试

## 组件详细报告

$(find "$output_dir" -name "report.md" | while read report; do
    local component=$(basename "$(dirname "$report")")
    echo "- [$component]($(dirname "$report")/report.md)"
done)

EOF
}

# 生成JSON汇总报告
generate_json_summary() {
    local report_file="$1"
    local total_crashes="$2"
    local critical_crashes="$3"
    local high_severity_crashes="$4"
    
    cat > "$report_file" << EOF
{
    "analysis_time": "$(date)",
    "total_crashes": $total_crashes,
    "severity_distribution": {
        "critical": $critical_crashes,
        "high": $high_severity_crashes,
        "medium": $((total_crashes - critical_crashes - high_severity_crashes)),
        "low": 0
    },
    "risk_assessment": {
        "critical_percentage": $((critical_crashes * 100 / total_crashes)),
        "high_percentage": $((high_severity_crashes * 100 / total_crashes)),
        "overall_risk": "$([ "$critical_crashes" -gt 0 ] && echo "high" || [ "$high_severity_crashes" -gt 0 ] && echo "medium" || echo "low")"
    },
    "action_plan": {
        "immediate": "修复所有Critical级别的crashes",
        "this_week": "修复High级别的crashes",
        "next_sprint": "修复Medium级别的crashes",
        "monitoring": "部署监控和定期Fuzzing测试"
    },
    "components": [
$(find "$output_dir" -name "analysis.json" | while read analysis_file; do
    local component=$(basename "$(dirname "$analysis_file")")
    cat "$analysis_file"
done | jq -s '.'
    ]
}
EOF
}

# 生成Jira工单模板
generate_jira_tickets() {
    local output_dir="$1"
    
    log_info "生成Jira工单模板..."
    
    local jira_dir="$output_dir/jira_tickets"
    mkdir -p "$jira_dir"
    
    find "$output_dir" -name "analysis.json" | while read analysis_file; do
        local component=$(basename "$(dirname "$analysis_file")")
        local component_jira_dir="$jira_dir/$component"
        mkdir -p "$component_jira_dir"
        
        # 为每个critical和high的crash生成Jira工单
        jq -r '.crashes[] | select(.severity == "critical" or .severity == "high") | {
            severity: .severity,
            type: .type,
            description: .description,
            input: .input,
            number: .number
        }' "$analysis_file" | while read -r crash; do
            local number=$(echo "$crash" | jq -r '.number')
            local severity=$(echo "$crash" | jq -r '.severity')
            local type=$(echo "$crash" | jq -r '.type')
            local description=$(echo "$crash" | jq -r '.description')
            local input=$(echo "$crash" | jq -r '.input')
            
            local ticket_file="$component_jira_dir/CRASH_${number}_${severity}.json"
            
            cat > "$ticket_file" << EOF
{
    "fields": {
        "project": {
            "key": "SEC"
        },
        "summary": "Fuzzing Crash: $component - $description",
        "description": {
            "version": 1,
            "type": "doc",
            "content": [
                {
                    "type": "paragraph",
                    "content": [
                        {
                            "type": "text",
                            "text": "Fuzzing测试发现的安全漏洞"
                        }
                    ]
                },
                {
                    "type": "paragraph",
                    "content": [
                        {
                            "type": "text",
                            "text": "h2. 组件信息\\n"
                        }
                    ]
                },
                {
                    "type": "paragraph",
                    "content": [
                        {
                            "type": "text",
                            "text": "- 组件: $component\\n- 类型: $type\\n- 严重程度: $severity\\n- 发现时间: $(date)"
                        }
                    ]
                },
                {
                    "type": "paragraph",
                    "content": [
                        {
                            "type": "text",
                            "text": "h2. 问题描述\\n"
                        }
                    ]
                },
                {
                    "type": "paragraph",
                    "content": [
                        {
                            "type": "text",
                            "text": "$description"
                        }
                    ]
                },
                {
                    "type": "paragraph",
                    "content": [
                        {
                            "type": "text",
                            "text": "h2. 触发输入\\n"
                        }
                    ]
                },
                {
                    "type": "codeBlock",
                    "attrs": {
                        "language": "text"
                    },
                    "content": [
                        {
                            "type": "text",
                            "text": "$input"
                        }
                    ]
                },
                {
                    "type": "paragraph",
                    "content": [
                        {
                            "type": "text",
                            "text": "h2. 影响评估\\n"
                        }
                    ]
                },
                {
                    "type": "paragraph",
                    "content": [
                        {
                            "type": "text",
                            "text": "这个crash可能导致系统不稳定或安全漏洞，需要立即修复。"
                        }
                    ]
                }
            ]
        },
        "issuetype": {
            "name": "Bug"
        },
        "priority": {
            "name": "$([ "$severity" == "critical" ] && echo "Highest" || echo "High")"
        },
        "labels": [
            "fuzzing",
            "security",
            "$component",
            "$type"
        ]
    }
}
EOF
        done
    done
    
    log_success "Jira工单模板已生成: $jira_dir"
}

# 运行指定的Fuzzing测试
run_fuzzing_test() {
    local test_path="$1"
    local output_dir="${2:-fuzzing_results}"
    
    log_info "运行Fuzzing测试: $test_path"
    
    if [ ! -d "$test_path" ]; then
        log_error "测试路径不存在: $test_path"
        return 1
    fi
    
    mkdir -p "$output_dir"
    
    # 运行Fuzzing测试
    local test_name=$(basename "$test_path")
    local result_dir="$output_dir/$test_name"
    mkdir -p "$result_dir"
    
    log_info "开始Fuzzing测试，输出目录: $result_dir"
    
    # 运行测试
    cd "$test_path"
    go test -fuzz=. -fuzztime=60s -fuzzminimizetime=10s -parallel=4 \
        > "$result_dir/fuzzing.log" 2>&1
    
    cd - > /dev/null
    
    # 检查是否有crashers
    if [ -f "$test_path/crashers" ]; then
        log_warning "发现crashers，复制到结果目录"
        cp "$test_path/crashers" "$result_dir/"
    fi
    
    log_success "Fuzzing测试完成: $test_name"
}

# 主函数
main() {
    # 默认参数
    ACTION=""
    CRASH_DIR=""
    OUTPUT_DIR="crash_reports"
    REPORT_FORMAT="markdown"
    TEST_PATH=""
    DEBUG=false
    FIX_SUGGESTIONS=false
    JIRA_TICKETS=false
    SEVERITY_FILTER=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -a|--analyze)
                ACTION="analyze"
                CRASH_DIR="$2"
                shift 2
                ;;
            -c|--crashers)
                CRASH_DIR="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -r|--report)
                REPORT_FORMAT="$2"
                shift 2
                ;;
            -t|--test)
                ACTION="test"
                TEST_PATH="$2"
                shift 2
                ;;
            -d|--debug)
                DEBUG=true
                shift
                ;;
            -f|--fix-suggestions)
                FIX_SUGGESTIONS=true
                shift
                ;;
            -j|--jira)
                JIRA_TICKETS=true
                shift
                ;;
            -s|--severity)
                SEVERITY_FILTER="$2"
                shift 2
                ;;
            *)
                log_error "未知参数: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # 检查依赖
    check_dependencies
    
    # 设置调试模式
    if [ "$DEBUG" = true ]; then
        set -x
    fi
    
    # 执行相应操作
    case $ACTION in
        analyze)
            if [ -z "$CRASH_DIR" ]; then
                log_error "请指定Crashers目录: -a|--analyze DIR"
                exit 1
            fi
            analyze_crashers_directory "$CRASH_DIR" "$OUTPUT_DIR" "$REPORT_FORMAT"
            
            # 生成修复建议
            if [ "$FIX_SUGGESTIONS" = true ]; then
                log_info "生成修复建议..."
                # 这里可以调用修复建议生成逻辑
            fi
            
            # 生成Jira工单
            if [ "$JIRA_TICKETS" = true ]; then
                generate_jira_tickets "$OUTPUT_DIR"
            fi
            ;;
        test)
            if [ -z "$TEST_PATH" ]; then
                log_error "请指定测试路径: -t|--test PATH"
                exit 1
            fi
            run_fuzzing_test "$TEST_PATH" "$OUTPUT_DIR"
            ;;
        *)
            log_error "请指定操作类型"
            show_usage
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"