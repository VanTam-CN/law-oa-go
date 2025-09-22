#!/bin/bash

# Fuzzing报告自动化生成器
# 自动生成、分析和分发Fuzzing测试报告

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
    echo "  -g, --generate TYPE        生成报告类型 (daily|weekly|monthly|ad-hoc)"
    echo "  -s, --source DIR           源数据目录 (默认: fuzzing_results)"
    echo "  -o, --output DIR           输出目录 (默认: reports)"
    echo "  -f, --format FORMAT        报告格式 (markdown|json|html|pdf, 默认: markdown)"
    echo "  -d, --distribute           分发报告到指定渠道"
    echo "  -c, --config FILE          配置文件路径 (默认: fuzzing.toml)"
    echo "  -t, --template FILE        报告模板文件"
    echo "  -e, --email EMAILS         邮件分发地址 (逗号分隔)"
    echo "  -w, --webhook URL          Webhook分发URL"
    echo "  -n, --notification         发送通知 (slack|discord|teams)"
    echo "  -a, --archive              归档报告"
    echo "  -r, --retention DAYS       保留天数 (默认: 30)"
    echo ""
    echo "示例:"
    echo "  $0 -g daily -f html                 # 生成每日HTML报告"
    echo "  $0 -g weekly -d -e dev@example.com # 生成并分发周报"
    echo "  $0 -g ad-hoc -s ./results -o ./out # 生成临时报告"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查必要的工具
    local required_tools=("jq" "curl")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "$tool 未安装"
            exit 1
        fi
    done
    
    # 检查可选工具
    local optional_tools=("pandoc" "wkhtmltopdf" "mail")
    for tool in "${optional_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_warning "$tool 未安装，相关功能将被禁用"
        fi
    done
    
    log_success "依赖检查通过"
}

# 加载配置文件
load_config() {
    local config_file="$1"
    
    if [ ! -f "$config_file" ]; then
        log_warning "配置文件不存在: $config_file，使用默认配置"
        return
    fi
    
    log_info "加载配置文件: $config_file"
    
    # 这里应该解析TOML配置文件
    # 简化实现，直接设置默认值
    REPORTING_ENABLED=true
    EMAIL_ENABLED=false
    WEBHOOK_ENABLED=false
    NOTIFICATION_ENABLED=false
    
    log_debug "配置加载完成"
}

# 收集Fuzzing数据
collect_fuzzing_data() {
    local source_dir="$1"
    local report_type="$2"
    
    log_info "收集Fuzzing数据: $source_dir (类型: $report_type)"
    
    if [ ! -d "$source_dir" ]; then
        log_error "源数据目录不存在: $source_dir"
        return 1
    fi
    
    local data_file=$(mktemp)
    
    # 根据报告类型收集数据
    case "$report_type" in
        "daily")
            collect_daily_data "$source_dir" "$data_file"
            ;;
        "weekly")
            collect_weekly_data "$source_dir" "$data_file"
            ;;
        "monthly")
            collect_monthly_data "$source_dir" "$data_file"
            ;;
        "ad-hoc")
            collect_adhoc_data "$source_dir" "$data_file"
            ;;
        *)
            log_error "未知的报告类型: $report_type"
            return 1
            ;;
    esac
    
    echo "$data_file"
}

# 收集每日数据
collect_daily_data() {
    local source_dir="$1"
    local data_file="$2"
    
    log_info "收集每日Fuzzing数据..."
    
    # 查找最近24小时的测试结果
    local recent_results=$(find "$source_dir" -name "*.json" -mtime -1 -type f 2>/dev/null || true)
    
    # 分析数据
    local total_tests=0
    local total_crashes=0
    local critical_crashes=0
    local high_crashes=0
    local component_stats=$(mktemp)
    
    for result_file in $recent_results; do
        total_tests=$((total_tests + 1))
        
        # 解析JSON结果文件
        local component=$(jq -r '.component // "unknown"' "$result_file" 2>/dev/null || echo "unknown")
        local crashes=$(jq -r '.crashes_found // 0' "$result_file" 2>/dev/null || echo "0")
        local critical=$(jq -r '.critical_crashes // 0' "$result_file" 2>/dev/null || echo "0")
        local high=$(jq -r '.high_crashes // 0' "$result_file" 2>/dev/null || echo "0")
        
        total_crashes=$((total_crashes + crashes))
        critical_crashes=$((critical_crashes + critical))
        high_crashes=$((high_crashes + high))
        
        # 统计组件数据
        echo "$component $crashes $critical $high" >> "$component_stats"
    done
    
    # 生成汇总数据
    cat > "$data_file" << EOF
{
    "report_type": "daily",
    "generated_at": "$(date -Iseconds)",
    "time_range": {
        "start": "$(date -Iseconds -d '1 day ago')",
        "end": "$(date -Iseconds)"
    },
    "summary": {
        "total_tests": $total_tests,
        "total_crashes": $total_crashes,
        "critical_crashes": $critical_crashes,
        "high_crashes": $high_crashes,
        "crash_rate": $(echo "scale=2; $total_crashes * 100 / $total_tests" | bc -l 2>/dev/null || echo "0")
    },
    "components": [
EOF
    
    # 添加组件统计
    local first=true
    sort "$component_stats" | uniq | while read component crashes critical high; do
        if [ "$first" = true ]; then
            first=false
        else
            echo "," >> "$data_file"
        fi
        cat >> "$data_file" << EOF
        {
            "name": "$component",
            "crashes": $crashes,
            "critical": $critical,
            "high": $high
        }
EOF
    done
    
    cat >> "$data_file" << EOF
    ]
}
EOF
    
    rm -f "$component_stats"
    log_success "每日数据收集完成"
}

# 收集周报数据
collect_weekly_data() {
    local source_dir="$1"
    local data_file="$2"
    
    log_info "收集周报Fuzzing数据..."
    
    # 查找最近7天的测试结果
    local recent_results=$(find "$source_dir" -name "*.json" -mtime -7 -type f 2>/dev/null || true)
    
    # 生成趋势数据
    local trend_data=$(mktemp)
    
    for i in {6..0}; do
        local date=$(date -Iseconds -d "$i days ago")
        local day_results=$(find "$source_dir" -name "*.json" -mtime $i -type f 2>/dev/null || true)
        
        local day_tests=$(echo "$day_results" | wc -l)
        local day_crashes=0
        
        for result_file in $day_results; do
            local crashes=$(jq -r '.crashes_found // 0' "$result_file" 2>/dev/null || echo "0")
            day_crashes=$((day_crashes + crashes))
        done
        
        echo "$date $day_tests $day_crashes" >> "$trend_data"
    done
    
    cat > "$data_file" << EOF
{
    "report_type": "weekly",
    "generated_at": "$(date -Iseconds)",
    "time_range": {
        "start": "$(date -Iseconds -d '7 days ago')",
        "end": "$(date -Iseconds)"
    },
    "trends": [
EOF
    
    local first=true
    while read date tests crashes; do
        if [ "$first" = true ]; then
            first=false
        else
            echo "," >> "$data_file"
        fi
        cat >> "$data_file" << EOF
        {
            "date": "$date",
            "tests": $tests,
            "crashes": $crashes
        }
EOF
    done < "$trend_data"
    
    cat >> "$data_file" << EOF
    ]
}
EOF
    
    rm -f "$trend_data"
    log_success "周报数据收集完成"
}

# 收集月报数据
collect_monthly_data() {
    local source_dir="$1"
    local data_file="$2"
    
    log_info "收集月报Fuzzing数据..."
    
    # 查找最近30天的测试结果
    local recent_results=$(find "$source_dir" -name "*.json" -mtime -30 -type f 2>/dev/null || true)
    
    # 统计数据
    local total_tests=$(echo "$recent_results" | wc -l)
    local total_crashes=0
    local critical_crashes=0
    local high_crashes=0
    
    for result_file in $recent_results; do
        local crashes=$(jq -r '.crashes_found // 0' "$result_file" 2>/dev/null || echo "0")
        local critical=$(jq -r '.critical_crashes // 0' "$result_file" 2>/dev/null || echo "0")
        local high=$(jq -r '.high_crashes // 0' "$result_file" 2>/dev/null || echo "0")
        
        total_crashes=$((total_crashes + crashes))
        critical_crashes=$((critical_crashes + critical))
        high_crashes=$((high_crashes + high))
    done
    
    cat > "$data_file" << EOF
{
    "report_type": "monthly",
    "generated_at": "$(date -Iseconds)",
    "time_range": {
        "start": "$(date -Iseconds -d '30 days ago')",
        "end": "$(date -Iseconds)"
    },
    "summary": {
        "total_tests": $total_tests,
        "total_crashes": $total_crashes,
        "critical_crashes": $critical_crashes,
        "high_crashes": $high_crashes,
        "average_daily_tests": $(echo "scale=1; $total_tests / 30" | bc -l 2>/dev/null || echo "0"),
        "average_daily_crashes": $(echo "scale=1; $total_crashes / 30" | bc -l 2>/dev/null || echo "0")
    }
}
EOF
    
    log_success "月报数据收集完成"
}

# 收集临时数据
collect_adhoc_data() {
    local source_dir="$1"
    local data_file="$2"
    
    log_info "收集临时Fuzzing数据..."
    
    # 收集所有可用的测试结果
    local all_results=$(find "$source_dir" -name "*.json" -type f 2>/dev/null || true)
    
    # 生成汇总统计
    local component_stats=$(mktemp)
    local total_tests=0
    local total_crashes=0
    
    for result_file in $all_results; do
        total_tests=$((total_tests + 1))
        
        local component=$(jq -r '.component // "unknown"' "$result_file" 2>/dev/null || echo "unknown")
        local crashes=$(jq -r '.crashes_found // 0' "$result_file" 2>/dev/null || echo "0")
        local critical=$(jq -r '.critical_crashes // 0' "$result_file" 2>/dev/null || echo "0")
        local high=$(jq -r '.high_crashes // 0' "$result_file" 2>/dev/null || echo "0")
        
        total_crashes=$((total_crashes + crashes))
        
        echo "$component $crashes $critical $high" >> "$component_stats"
    done
    
    cat > "$data_file" << EOF
{
    "report_type": "ad-hoc",
    "generated_at": "$(date -Iseconds)",
    "summary": {
        "total_tests": $total_tests,
        "total_crashes": $total_crashes
    },
    "components": [
EOF
    
    local first=true
    sort "$component_stats" | uniq | while read component crashes critical high; do
        if [ "$first" = true ]; then
            first=false
        else
            echo "," >> "$data_file"
        fi
        cat >> "$data_file" << EOF
        {
            "name": "$component",
            "crashes": $crashes,
            "critical": $critical,
            "high": $high
        }
EOF
    done
    
    cat >> "$data_file" << EOF
    ]
}
EOF
    
    rm -f "$component_stats"
    log_success "临时数据收集完成"
}

# 生成报告
generate_report() {
    local data_file="$1"
    local output_dir="$2"
    local format="$3"
    local template_file="$4"
    local report_type="$5"
    
    log_info "生成报告: 格式=$format, 输出目录=$output_dir"
    
    mkdir -p "$output_dir"
    
    local report_name="fuzzing_report_$(date +%Y%m%d_%H%M%S)"
    local output_file="$output_dir/${report_name}.${format}"
    
    case "$format" in
        "json")
            generate_json_report "$data_file" "$output_file"
            ;;
        "markdown")
            generate_markdown_report "$data_file" "$output_file" "$report_type"
            ;;
        "html")
            generate_html_report "$data_file" "$output_file" "$template_file"
            ;;
        "pdf")
            generate_pdf_report "$data_file" "$output_file" "$template_file"
            ;;
        *)
            log_error "不支持的报告格式: $format"
            return 1
            ;;
    esac
    
    log_success "报告生成完成: $output_file"
    echo "$output_file"
}

# 生成JSON报告
generate_json_report() {
    local data_file="$1"
    local output_file="$2"
    
    cp "$data_file" "$output_file"
}

# 生成Markdown报告
generate_markdown_report() {
    local data_file="$1"
    local output_file="$2"
    local report_type="$3"
    
    cat > "$output_file" << EOF
# Fuzzing测试报告

## 基本信息
- **报告类型**: $report_type
- **生成时间**: $(date)
- **生成工具**: Fuzzing报告自动化生成器

EOF

    # 添加汇总信息
    local total_tests=$(jq -r '.summary.total_tests // 0' "$data_file")
    local total_crashes=$(jq -r '.summary.total_crashes // 0' "$data_file")
    local critical_crashes=$(jq -r '.summary.critical_crashes // 0' "$data_file")
    local high_crashes=$(jq -r '.summary.high_crashes // 0' "$data_file")

    cat >> "$output_file" << EOF
## 执行摘要

### 测试统计
- **总测试次数**: $total_tests
- **发现Crashes**: $total_crashes
- **严重Crashes**: $critical_crashes
- **高优先级Crashes**: $high_crashes
- **Crash率**: $(echo "scale=2; $total_crashes * 100 / $total_tests" | bc -l 2>/dev/null || echo "0")%

### 风险评估
$(if [ "$critical_crashes" -gt 0 ]; then
    echo "- 🔴 **高风险**: 发现 $critical_crashes 个严重Crashes，需要立即修复"
elif [ "$high_crashes" -gt 0 ]; then
    echo "- 🟡 **中等风险**: 发现 $high_crashes 个高优先级Crashes，需要优先修复"
else
    echo "- 🟢 **低风险**: 未发现严重或高优先级Crashes"
fi)

EOF

    # 添加组件详情
    if jq -e '.components' "$data_file" > /dev/null 2>&1; then
        cat >> "$output_file" << EOF
## 组件详情

| 组件 | Crashes | 严重 | 高优先级 | 风险等级 |
|------|---------|------|----------|----------|
EOF

        jq -r '.components[] | "\(.name) \(.crashes) \(.critical) \(.high)"' "$data_file" | while read component crashes critical high; do
            local risk_level="🟢 低"
            if [ "$critical" -gt 0 ]; then
                risk_level="🔴 高"
            elif [ "$high" -gt 0 ]; then
                risk_level="🟡 中"
            fi
            echo "| $component | $crashes | $critical | $high | $risk_level |" >> "$output_file"
        done
    fi

    # 添加趋势分析（如果可用）
    if jq -e '.trends' "$data_file" > /dev/null 2>&1; then
        cat >> "$output_file" << EOF

## 趋势分析

$(jq -r '.trends[] | "- \(.date): \(.tests) 次测试, \(.crashes) 个Crashes"' "$data_file")
EOF
    fi

    # 添加建议
    cat >> "$output_file" << EOF

## 建议和后续步骤

### 立即行动
$(if [ "$critical_crashes" -gt 0 ]; then
    echo "1. **立即修复所有严重Crashes**"
    echo "2. **重新运行Fuzzing测试验证修复效果**"
else
    echo "1. **继续监控现有Crashes的修复进度**"
fi)

### 本周计划
$(if [ "$high_crashes" -gt 0 ]; then
    echo "1. **修复高优先级Crashes**"
    echo "2. **代码审查和测试**"
fi)

### 长期改进
1. **完善输入验证**
2. **增强错误处理**
3. **改进代码质量**
4. **定期Fuzzing测试**

---
*此报告由Fuzzing报告自动化生成器生成*
EOF
}

# 生成HTML报告
generate_html_report() {
    local data_file="$1"
    local output_file="$2"
    local template_file="$3"
    
    if [ -n "$template_file" ] && [ -f "$template_file" ]; then
        log_info "使用自定义模板: $template_file"
        # 这里应该使用模板引擎生成报告
        # 简化实现，生成基础HTML
        generate_basic_html_report "$data_file" "$output_file"
    else
        generate_basic_html_report "$data_file" "$output_file"
    fi
}

# 生成基础HTML报告
generate_basic_html_report() {
    local data_file="$1"
    local output_file="$2"
    
    # 从JSON数据提取信息
    local total_tests=$(jq -r '.summary.total_tests // 0' "$data_file")
    local total_crashes=$(jq -r '.summary.total_crashes // 0' "$data_file")
    local critical_crashes=$(jq -r '.summary.critical_crashes // 0' "$data_file")
    local high_crashes=$(jq -r '.summary.high_crashes // 0' "$data_file")
    local report_type=$(jq -r '.report_type // "unknown"' "$data_file")
    
    cat > "$output_file" << EOF
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Fuzzing测试报告 - $report_type</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 5px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .card { background-color: #ffffff; border: 1px solid #dee2e6; border-radius: 5px; padding: 15px; }
        .card h3 { margin-top: 0; color: #495057; }
        .danger { color: #dc3545; }
        .warning { color: #ffc107; }
        .success { color: #28a745; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #dee2e6; padding: 8px; text-align: left; }
        th { background-color: #f8f9fa; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Fuzzing测试报告</h1>
        <p><strong>报告类型:</strong> $report_type</p>
        <p><strong>生成时间:</strong> $(date)</p>
    </div>
    
    <div class="summary">
        <div class="card">
            <h3>总测试次数</h3>
            <p style="font-size: 24px; font-weight: bold;">$total_tests</p>
        </div>
        <div class="card">
            <h3>发现Crashes</h3>
            <p style="font-size: 24px; font-weight: bold;" class="danger">$total_crashes</p>
        </div>
        <div class="card">
            <h3>严重Crashes</h3>
            <p style="font-size: 24px; font-weight: bold;" class="danger">$critical_crashes</p>
        </div>
        <div class="card">
            <h3>高优先级Crashes</h3>
            <p style="font-size: 24px; font-weight: bold;" class="warning">$high_crashes</p>
        </div>
    </div>
    
    <div class="card">
        <h3>风险评估</h3>
        <p>
EOF

    if [ "$critical_crashes" -gt 0 ]; then
        echo "            <span class=\"danger\">🔴 高风险: 发现 $critical_crashes 个严重Crashes，需要立即修复</span>" >> "$output_file"
    elif [ "$high_crashes" -gt 0 ]; then
        echo "            <span class=\"warning\">🟡 中等风险: 发现 $high_crashes 个高优先级Crashes，需要优先修复</span>" >> "$output_file"
    else
        echo "            <span class=\"success\">🟢 低风险: 未发现严重或高优先级Crashes</span>" >> "$output_file"
    fi

    cat >> "$output_file" << EOF
        </p>
    </div>
    
EOF

    # 添加组件表格
    if jq -e '.components' "$data_file" > /dev/null 2>&1; then
        cat >> "$output_file" << EOF
    <div class="card">
        <h3>组件详情</h3>
        <table>
            <tr>
                <th>组件</th>
                <th>Crashes</th>
                <th>严重</th>
                <th>高优先级</th>
                <th>风险等级</th>
            </tr>
EOF

        jq -r '.components[] | "\(.name) \(.crashes) \(.critical) \(.high)"' "$data_file" | while read component crashes critical high; do
            local risk_level="🟢 低"
            local risk_class="success"
            if [ "$critical" -gt 0 ]; then
                risk_level="🔴 高"
                risk_class="danger"
            elif [ "$high" -gt 0 ]; then
                risk_level="🟡 中"
                risk_class="warning"
            fi
            echo "            <tr>" >> "$output_file"
            echo "                <td>$component</td>" >> "$output_file"
            echo "                <td>$crashes</td>" >> "$output_file"
            echo "                <td>$critical</td>" >> "$output_file"
            echo "                <td>$high</td>" >> "$output_file"
            echo "                <td class=\"$risk_class\">$risk_level</td>" >> "$output_file"
            echo "            </tr>" >> "$output_file"
        done

        cat >> "$output_file" << EOF
        </table>
    </div>
    
EOF
    fi

    cat >> "$output_file" << EOF
    <div class="card">
        <h3>建议和后续步骤</h3>
        <ul>
EOF

    if [ "$critical_crashes" -gt 0 ]; then
        echo "            <li>立即修复所有严重Crashes</li>" >> "$output_file"
        echo "            <li>重新运行Fuzzing测试验证修复效果</li>" >> "$output_file"
    else
        echo "            <li>继续监控现有Crashes的修复进度</li>" >> "$output_file"
    fi

    if [ "$high_crashes" -gt 0 ]; then
        echo "            <li>修复高优先级Crashes</li>" >> "$output_file"
        echo "            <li>代码审查和测试</li>" >> "$output_file"
    fi

    cat >> "$output_file" << EOF
            <li>完善输入验证</li>
            <li>增强错误处理</li>
            <li>改进代码质量</li>
            <li>定期Fuzzing测试</li>
        </ul>
    </div>
    
    <footer style="margin-top: 40px; text-align: center; color: #6c757d;">
        <p>此报告由Fuzzing报告自动化生成器生成</p>
    </footer>
</body>
</html>
EOF
}

# 生成PDF报告
generate_pdf_report() {
    local data_file="$1"
    local output_file="$2"
    local template_file="$3"
    
    log_info "生成PDF报告..."
    
    # 检查wkhtmltopdf是否可用
    if ! command -v wkhtmltopdf &> /dev/null; then
        log_error "wkhtmltopdf 未安装，无法生成PDF报告"
        return 1
    fi
    
    # 先生成HTML报告
    local html_file="${output_file%.pdf}.html"
    generate_html_report "$data_file" "$html_file" "$template_file"
    
    # 转换为PDF
    wkhtmltopdf "$html_file" "$output_file"
    
    # 清理临时文件
    rm -f "$html_file"
    
    log_success "PDF报告生成完成: $output_file"
}

# 分发报告
distribute_report() {
    local report_file="$1"
    local emails="$2"
    local webhook_url="$3"
    local notification_type="$4"
    
    log_info "分发报告: $report_file"
    
    # 邮件分发
    if [ -n "$emails" ] && [ "$EMAIL_ENABLED" = true ]; then
        distribute_via_email "$report_file" "$emails"
    fi
    
    # Webhook分发
    if [ -n "$webhook_url" ] && [ "$WEBHOOK_ENABLED" = true ]; then
        distribute_via_webhook "$report_file" "$webhook_url"
    fi
    
    # 通知分发
    if [ -n "$notification_type" ] && [ "$NOTIFICATION_ENABLED" = true ]; then
        distribute_via_notification "$report_file" "$notification_type"
    fi
    
    log_success "报告分发完成"
}

# 邮件分发
distribute_via_email() {
    local report_file="$1"
    local emails="$2"
    
    log_info "通过邮件分发报告..."
    
    # 检查mail命令是否可用
    if ! command -v mail &> /dev/null; then
        log_warning "mail 命令不可用，跳过邮件分发"
        return
    fi
    
    local subject="Fuzzing测试报告 - $(date +%Y-%m-%d)"
    local body="请查看附件中的Fuzzing测试报告。"
    
    # 发送邮件
    echo "$body" | mail -s "$subject" -a "$report_file" "$emails"
    
    log_success "邮件分发完成: $emails"
}

# Webhook分发
distribute_via_webhook() {
    local report_file="$1"
    local webhook_url="$2"
    
    log_info "通过Webhook分发报告..."
    
    # 读取报告内容
    local report_content=$(cat "$report_file")
    
    # 发送Webhook
    curl -X POST "$webhook_url" \
         -H "Content-Type: application/json" \
         -d "{
             \"text\": \"Fuzzing测试报告已生成\",
             \"report_type\": \"$(basename "$report_file")\",
             \"generated_at\": \"$(date -Iseconds)\",
             \"file_size\": $(stat -c%s "$report_file"),
             \"attachment\": \"$(echo "$report_content" | base64 -w 0)\"
         }"
    
    log_success "Webhook分发完成: $webhook_url"
}

# 通知分发
distribute_via_notification() {
    local report_file="$1"
    local notification_type="$2"
    
    log_info "通过$notification_type分发报告..."
    
    local message="Fuzzing测试报告已生成: $(basename "$report_file)"
    
    case "$notification_type" in
        "slack")
            # 发送Slack通知
            curl -X POST "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
                 -H "Content-Type: application/json" \
                 -d "{\"text\": \"$message\"}"
            ;;
        "discord")
            # 发送Discord通知
            curl -X POST "https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK" \
                 -H "Content-Type: application/json" \
                 -d "{\"content\": \"$message\"}"
            ;;
        "teams")
            # 发送Teams通知
            curl -X POST "https://outlook.office.com/webhook/YOUR/TEAMS/WEBHOOK" \
                 -H "Content-Type: application/json" \
                 -d "{\"text\": \"$message\"}"
            ;;
        *)
            log_warning "不支持的通知类型: $notification_type"
            ;;
    esac
    
    log_success "$notification_type 通知分发完成"
}

# 归档报告
archive_report() {
    local report_file="$1"
    local retention_days="$2"
    
    log_info "归档报告..."
    
    local archive_dir="reports/archive/$(date +%Y/%m)"
    mkdir -p "$archive_dir"
    
    local archive_file="$archive_dir/$(basename "$report_file")"
    mv "$report_file" "$archive_file"
    
    # 清理旧报告
    if [ -n "$retention_days" ]; then
        find reports/archive -name "*.pdf" -o -name "*.html" -o -name "*.md" -o -name "*.json" \
             -mtime +"$retention_days" -delete 2>/dev/null || true
    fi
    
    log_success "报告归档完成: $archive_file"
}

# 主函数
main() {
    # 默认参数
    ACTION="generate"
    REPORT_TYPE="ad-hoc"
    SOURCE_DIR="fuzzing_results"
    OUTPUT_DIR="reports"
    FORMAT="markdown"
    CONFIG_FILE="fuzzing.toml"
    DISTRIBUTE=false
    EMAILS=""
    WEBHOOK_URL=""
    NOTIFICATION_TYPE=""
    ARCHIVE=false
    RETENTION_DAYS="30"
    TEMPLATE_FILE=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -g|--generate)
                REPORT_TYPE="$2"
                shift 2
                ;;
            -s|--source)
                SOURCE_DIR="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -f|--format)
                FORMAT="$2"
                shift 2
                ;;
            -d|--distribute)
                DISTRIBUTE=true
                shift
                ;;
            -c|--config)
                CONFIG_FILE="$2"
                shift 2
                ;;
            -t|--template)
                TEMPLATE_FILE="$2"
                shift 2
                ;;
            -e|--email)
                EMAILS="$2"
                shift 2
                ;;
            -w|--webhook)
                WEBHOOK_URL="$2"
                shift 2
                ;;
            -n|--notification)
                NOTIFICATION_TYPE="$2"
                shift 2
                ;;
            -a|--archive)
                ARCHIVE=true
                shift
                ;;
            -r|--retention)
                RETENTION_DAYS="$2"
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
    
    # 加载配置
    load_config "$CONFIG_FILE"
    
    # 收集Fuzzing数据
    log_info "开始收集Fuzzing数据..."
    local data_file
    data_file=$(collect_fuzzing_data "$SOURCE_DIR" "$REPORT_TYPE")
    
    if [ $? -ne 0 ]; then
        log_error "数据收集失败"
        exit 1
    fi
    
    # 生成报告
    log_info "开始生成报告..."
    local report_file
    report_file=$(generate_report "$data_file" "$OUTPUT_DIR" "$FORMAT" "$TEMPLATE_FILE" "$REPORT_TYPE")
    
    if [ $? -ne 0 ]; then
        log_error "报告生成失败"
        exit 1
    fi
    
    # 分发报告
    if [ "$DISTRIBUTE" = true ]; then
        log_info "开始分发报告..."
        distribute_report "$report_file" "$EMAILS" "$WEBHOOK_URL" "$NOTIFICATION_TYPE"
    fi
    
    # 归档报告
    if [ "$ARCHIVE" = true ]; then
        log_info "开始归档报告..."
        archive_report "$report_file" "$RETENTION_DAYS"
    fi
    
    # 清理临时文件
    rm -f "$data_file"
    
    log_success "Fuzzing报告自动化完成"
    echo "生成的报告: $report_file"
}

# 运行主函数
main "$@"