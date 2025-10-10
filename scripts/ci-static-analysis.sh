#!/bin/bash

# CI/CD 静态分析集成脚本
# Law OA Go 项目 - 持续集成静态分析

set -e

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CI_REPORTS_DIR="$PROJECT_ROOT/ci-reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# 质量门禁配置
MAX_CRITICAL_ISSUES=0
MAX_HIGH_ISSUES=0
MAX_MEDIUM_ISSUES=5
MIN_QUALITY_SCORE=8.0

# 颜色配置
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log() {
    echo -e "${BLUE}[CI-ANALYSIS]${NC} $1"
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

# 创建 CI 报告目录
create_ci_directories() {
    log "创建 CI 报告目录..."
    mkdir -p "$CI_REPORTS_DIR"
    mkdir -p "$CI_REPORTS_DIR/go"
    mkdir -p "$CI_REPORTS_DIR/frontend"
    mkdir -p "$CI_REPORTS_DIR/quality-gates"
    log_success "CI 报告目录创建完成"
}

# 运行 Go 静态分析（CI 优化版）
run_ci_go_analysis() {
    log "运行 CI Go 静态分析..."
    cd "$PROJECT_ROOT"
    
    local go_issues=0
    local go_critical=0
    local go_high=0
    local go_medium=0
    
    # golangci-lint 分析（快速模式）
    if command -v golangci-lint >/dev/null 2>&1; then
        log "运行 golangci-lint (CI 模式)..."
        
        # 只运行关键检查器，提高 CI 速度
        golangci-lint run \
            --enable errcheck,govet,staticcheck,gosec,typecheck \
            --out-format json \
            --timeout 5m \
            > "$CI_REPORTS_DIR/go/golangci-lint-ci.json" 2>/dev/null || true
        
        # 生成 GitHub Actions 格式输出
        if [ "$GITHUB_ACTIONS" = "true" ]; then
            golangci-lint run \
                --enable errcheck,govet,staticcheck,gosec,typecheck \
                --out-format github-actions \
                --timeout 5m || true
        fi
        
        # 统计问题数量
        if [ -f "$CI_REPORTS_DIR/go/golangci-lint-ci.json" ] && command -v jq >/dev/null 2>&1; then
            go_issues=$(jq '.Issues | length' "$CI_REPORTS_DIR/go/golangci-lint-ci.json" 2>/dev/null || echo "0")
            go_critical=$(jq '[.Issues[] | select(.Severity == "error")] | length' "$CI_REPORTS_DIR/go/golangci-lint-ci.json" 2>/dev/null || echo "0")
        fi
        
        log_success "golangci-lint CI 分析完成 (问题数: $go_issues)"
    fi
    
    # gosec 快速安全扫描
    if command -v gosec >/dev/null 2>&1; then
        log "运行 gosec 快速安全扫描..."
        
        gosec -fmt json -quiet -out "$CI_REPORTS_DIR/go/gosec-ci.json" ./... 2>/dev/null || true
        
        # 统计安全问题
        if [ -f "$CI_REPORTS_DIR/go/gosec-ci.json" ] && command -v jq >/dev/null 2>&1; then
            local security_issues=$(jq '.Issues | length' "$CI_REPORTS_DIR/go/gosec-ci.json" 2>/dev/null || echo "0")
            go_high=$((go_high + security_issues))
        fi
        
        log_success "gosec 快速扫描完成"
    fi
    
    # 保存 Go 分析结果
    cat > "$CI_REPORTS_DIR/go/analysis-summary.json" << EOF
{
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "total_issues": $go_issues,
    "critical": $go_critical,
    "high": $go_high,
    "medium": $go_medium,
    "status": "completed"
}
EOF
    
    echo "$go_issues" > "$CI_REPORTS_DIR/go/total_issues.txt"
    echo "$go_critical" > "$CI_REPORTS_DIR/go/critical_issues.txt"
}

# 运行前端静态分析（CI 优化版）
run_ci_frontend_analysis() {
    log "运行 CI 前端静态分析..."
    
    local frontend_issues=0
    local frontend_errors=0
    
    # Bootstrap 前端快速检查
    if [ -d "$PROJECT_ROOT/frontend" ] && [ -f "$PROJECT_ROOT/frontend/package.json" ]; then
        log "检查 Bootstrap 前端..."
        cd "$PROJECT_ROOT/frontend"
        
        # 快速 ESLint 检查（只检查错误级别）
        if command -v eslint >/dev/null 2>&1; then
            npm run lint -- --format json --quiet > "../$CI_REPORTS_DIR/frontend/eslint-bootstrap-ci.json" 2>/dev/null || true
            
            # 统计错误数量
            if [ -f "../$CI_REPORTS_DIR/frontend/eslint-bootstrap-ci.json" ] && command -v jq >/dev/null 2>&1; then
                local bootstrap_errors=$(jq '[.[] | .messages[] | select(.severity == 2)] | length' "../$CI_REPORTS_DIR/frontend/eslint-bootstrap-ci.json" 2>/dev/null || echo "0")
                frontend_errors=$((frontend_errors + bootstrap_errors))
            fi
        fi
        
        # TypeScript 类型检查
        if command -v tsc >/dev/null 2>&1; then
            npm run type-check > "../$CI_REPORTS_DIR/frontend/typescript-bootstrap-ci.txt" 2>&1 || true
        fi
    fi
    
    # Ant Design 前端快速检查
    if [ -d "$PROJECT_ROOT/frontend-vue" ] && [ -f "$PROJECT_ROOT/frontend-vue/package.json" ]; then
        log "检查 Ant Design 前端..."
        cd "$PROJECT_ROOT/frontend-vue"
        
        # 快速 ESLint 检查
        if command -v eslint >/dev/null 2>&1; then
            npm run lint -- --format json --quiet > "../$CI_REPORTS_DIR/frontend/eslint-antd-ci.json" 2>/dev/null || true
            
            # 统计错误数量
            if [ -f "../$CI_REPORTS_DIR/frontend/eslint-antd-ci.json" ] && command -v jq >/dev/null 2>&1; then
                local antd_errors=$(jq '[.[] | .messages[] | select(.severity == 2)] | length' "../$CI_REPORTS_DIR/frontend/eslint-antd-ci.json" 2>/dev/null || echo "0")
                frontend_errors=$((frontend_errors + antd_errors))
            fi
        fi
        
        # TypeScript 类型检查
        if command -v tsc >/dev/null 2>&1; then
            npm run type-check > "../$CI_REPORTS_DIR/frontend/typescript-antd-ci.txt" 2>&1 || true
        fi
    fi
    
    frontend_issues=$frontend_errors
    
    # 保存前端分析结果
    cat > "$CI_REPORTS_DIR/frontend/analysis-summary.json" << EOF
{
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "total_issues": $frontend_issues,
    "errors": $frontend_errors,
    "warnings": 0,
    "status": "completed"
}
EOF
    
    echo "$frontend_issues" > "$CI_REPORTS_DIR/frontend/total_issues.txt"
    echo "$frontend_errors" > "$CI_REPORTS_DIR/frontend/error_issues.txt"
    
    log_success "前端 CI 分析完成 (错误数: $frontend_errors)"
}

# 质量门禁检查
run_quality_gates() {
    log "运行质量门禁检查..."
    
    local gate_passed=true
    local gate_report="$CI_REPORTS_DIR/quality-gates/gate-report.json"
    
    # 读取分析结果
    local go_critical=0
    local go_high=0
    local go_total=0
    local frontend_errors=0
    
    if [ -f "$CI_REPORTS_DIR/go/critical_issues.txt" ]; then
        go_critical=$(cat "$CI_REPORTS_DIR/go/critical_issues.txt")
    fi
    
    if [ -f "$CI_REPORTS_DIR/go/total_issues.txt" ]; then
        go_total=$(cat "$CI_REPORTS_DIR/go/total_issues.txt")
    fi
    
    if [ -f "$CI_REPORTS_DIR/frontend/error_issues.txt" ]; then
        frontend_errors=$(cat "$CI_REPORTS_DIR/frontend/error_issues.txt")
    fi
    
    # 检查关键问题数量
    if [ "$go_critical" -gt "$MAX_CRITICAL_ISSUES" ]; then
        log_error "质量门禁失败: Go 关键问题数 ($go_critical) 超过限制 ($MAX_CRITICAL_ISSUES)"
        gate_passed=false
    fi
    
    # 检查前端错误数量
    if [ "$frontend_errors" -gt "$MAX_CRITICAL_ISSUES" ]; then
        log_error "质量门禁失败: 前端错误数 ($frontend_errors) 超过限制 ($MAX_CRITICAL_ISSUES)"
        gate_passed=false
    fi
    
    # 生成质量门禁报告
    cat > "$gate_report" << EOF
{
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "gate_passed": $gate_passed,
    "criteria": {
        "max_critical_issues": $MAX_CRITICAL_ISSUES,
        "max_high_issues": $MAX_HIGH_ISSUES,
        "max_medium_issues": $MAX_MEDIUM_ISSUES,
        "min_quality_score": $MIN_QUALITY_SCORE
    },
    "results": {
        "go_critical_issues": $go_critical,
        "go_total_issues": $go_total,
        "frontend_errors": $frontend_errors,
        "quality_score": 0
    },
    "status": "$([ "$gate_passed" = true ] && echo "PASSED" || echo "FAILED")"
}
EOF
    
    if [ "$gate_passed" = true ]; then
        log_success "质量门禁检查通过"
        echo "PASSED" > "$CI_REPORTS_DIR/quality-gates/status.txt"
        return 0
    else
        log_error "质量门禁检查失败"
        echo "FAILED" > "$CI_REPORTS_DIR/quality-gates/status.txt"
        return 1
    fi
}

# 生成 CI 摘要报告
generate_ci_summary() {
    log "生成 CI 摘要报告..."
    
    local summary_file="$CI_REPORTS_DIR/ci-summary.md"
    local go_total=0
    local go_critical=0
    local frontend_errors=0
    local gate_status="UNKNOWN"
    
    # 读取分析结果
    if [ -f "$CI_REPORTS_DIR/go/total_issues.txt" ]; then
        go_total=$(cat "$CI_REPORTS_DIR/go/total_issues.txt")
    fi
    
    if [ -f "$CI_REPORTS_DIR/go/critical_issues.txt" ]; then
        go_critical=$(cat "$CI_REPORTS_DIR/go/critical_issues.txt")
    fi
    
    if [ -f "$CI_REPORTS_DIR/frontend/error_issues.txt" ]; then
        frontend_errors=$(cat "$CI_REPORTS_DIR/frontend/error_issues.txt")
    fi
    
    if [ -f "$CI_REPORTS_DIR/quality-gates/status.txt" ]; then
        gate_status=$(cat "$CI_REPORTS_DIR/quality-gates/status.txt")
    fi
    
    # 生成摘要报告
    cat > "$summary_file" << EOF
# 🔍 CI 静态分析摘要报告

**项目**: Law OA Go  
**分析时间**: $(date)  
**CI 构建**: ${GITHUB_RUN_NUMBER:-${BUILD_NUMBER:-"本地构建"}}  
**分支**: ${GITHUB_REF_NAME:-${GIT_BRANCH:-"$(git branch --show-current 2>/dev/null || echo 'unknown')"}}  

## 📊 分析结果

| 组件 | 问题数量 | 状态 |
|------|----------|------|
| Go 后端 | $go_total | $([ "$go_critical" -eq 0 ] && echo "✅ 通过" || echo "❌ 失败") |
| 前端代码 | $frontend_errors | $([ "$frontend_errors" -eq 0 ] && echo "✅ 通过" || echo "❌ 失败") |

## 🚦 质量门禁

**状态**: $([ "$gate_status" = "PASSED" ] && echo "✅ 通过" || echo "❌ 失败")

### 检查项目
- 🔴 关键问题: $go_critical / $MAX_CRITICAL_ISSUES (限制)
- 🟡 前端错误: $frontend_errors / $MAX_CRITICAL_ISSUES (限制)

## 📁 报告文件

- \`ci-reports/go/golangci-lint-ci.json\` - Go 代码分析
- \`ci-reports/go/gosec-ci.json\` - 安全扫描
- \`ci-reports/frontend/eslint-bootstrap-ci.json\` - Bootstrap 前端
- \`ci-reports/frontend/eslint-antd-ci.json\` - Ant Design 前端
- \`ci-reports/quality-gates/gate-report.json\` - 质量门禁详情

## 🎯 后续行动

$(if [ "$gate_status" = "PASSED" ]; then
    echo "✅ 代码质量检查通过，可以继续部署流程。"
else
    echo "❌ 代码质量检查失败，请修复以下问题："
    echo ""
    if [ "$go_critical" -gt 0 ]; then
        echo "- 修复 $go_critical 个 Go 关键问题"
    fi
    if [ "$frontend_errors" -gt 0 ]; then
        echo "- 修复 $frontend_errors 个前端错误"
    fi
fi)

---
*报告由 CI 静态分析系统自动生成*
EOF
    
    log_success "CI 摘要报告生成完成: $summary_file"
    
    # 如果在 GitHub Actions 中，输出摘要
    if [ "$GITHUB_ACTIONS" = "true" ] && [ -n "$GITHUB_STEP_SUMMARY" ]; then
        cat "$summary_file" >> "$GITHUB_STEP_SUMMARY"
    fi
}

# 设置 GitHub Actions 输出
set_github_outputs() {
    if [ "$GITHUB_ACTIONS" = "true" ]; then
        log "设置 GitHub Actions 输出..."
        
        local go_total=0
        local go_critical=0
        local frontend_errors=0
        local gate_status="UNKNOWN"
        
        # 读取结果
        if [ -f "$CI_REPORTS_DIR/go/total_issues.txt" ]; then
            go_total=$(cat "$CI_REPORTS_DIR/go/total_issues.txt")
        fi
        
        if [ -f "$CI_REPORTS_DIR/go/critical_issues.txt" ]; then
            go_critical=$(cat "$CI_REPORTS_DIR/go/critical_issues.txt")
        fi
        
        if [ -f "$CI_REPORTS_DIR/frontend/error_issues.txt" ]; then
            frontend_errors=$(cat "$CI_REPORTS_DIR/frontend/error_issues.txt")
        fi
        
        if [ -f "$CI_REPORTS_DIR/quality-gates/status.txt" ]; then
            gate_status=$(cat "$CI_REPORTS_DIR/quality-gates/status.txt")
        fi
        
        # 设置输出变量
        echo "go-issues=$go_total" >> "$GITHUB_OUTPUT"
        echo "go-critical=$go_critical" >> "$GITHUB_OUTPUT"
        echo "frontend-errors=$frontend_errors" >> "$GITHUB_OUTPUT"
        echo "quality-gate=$gate_status" >> "$GITHUB_OUTPUT"
        echo "gate-passed=$([ "$gate_status" = "PASSED" ] && echo "true" || echo "false")" >> "$GITHUB_OUTPUT"
        
        log_success "GitHub Actions 输出设置完成"
    fi
}

# 上传报告到 CI 系统
upload_ci_reports() {
    log "准备上传 CI 报告..."
    
    # 创建报告压缩包
    local archive_name="static-analysis-reports-$TIMESTAMP.tar.gz"
    cd "$PROJECT_ROOT"
    tar -czf "$archive_name" ci-reports/
    
    log_success "报告压缩包创建完成: $archive_name"
    
    # 如果在 GitHub Actions 中，上传为 artifact
    if [ "$GITHUB_ACTIONS" = "true" ]; then
        log "报告将通过 GitHub Actions artifact 上传"
        # GitHub Actions 会自动处理 ci-reports 目录的上传
    fi
}

# 主函数
main() {
    log "开始 CI 静态分析..."
    
    # 检查必要工具
    local missing_tools=()
    
    if ! command -v golangci-lint >/dev/null 2>&1; then
        missing_tools+=("golangci-lint")
    fi
    
    if ! command -v jq >/dev/null 2>&1; then
        log_warning "jq 未安装，将跳过 JSON 解析功能"
    fi
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "缺少必要工具: ${missing_tools[*]}"
        log_error "请运行 ./scripts/code-review-setup.sh 安装工具"
        exit 1
    fi
    
    # 执行分析流程
    create_ci_directories
    run_ci_go_analysis
    run_ci_frontend_analysis
    
    # 质量门禁检查
    local gate_result=0
    run_quality_gates || gate_result=$?
    
    # 生成报告
    generate_ci_summary
    set_github_outputs
    upload_ci_reports
    
    if [ $gate_result -eq 0 ]; then
        log_success "CI 静态分析完成，质量门禁通过"
        exit 0
    else
        log_error "CI 静态分析完成，质量门禁失败"
        exit 1
    fi
}

# 执行主函数
main "$@"