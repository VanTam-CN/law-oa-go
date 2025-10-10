#!/bin/bash

# 静态分析配置验证脚本
# Law OA Go 项目 - 验证所有静态分析工具配置

set -e

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 颜色配置
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 验证结果
VALIDATION_PASSED=true
TOTAL_CHECKS=0
PASSED_CHECKS=0

# 日志函数
log() {
    echo -e "${BLUE}[VALIDATION]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
}

log_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
    VALIDATION_PASSED=false
}

check_item() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
}

# 验证 golangci-lint 配置
validate_golangci_config() {
    log "验证 golangci-lint 配置..."
    
    check_item
    if [ -f "$PROJECT_ROOT/.golangci.yml" ]; then
        log_success ".golangci.yml 配置文件存在"
        
        # 验证配置语法
        if command -v golangci-lint >/dev/null 2>&1; then
            check_item
            if golangci-lint config path >/dev/null 2>&1; then
                log_success "golangci-lint 配置语法正确"
            else
                log_error "golangci-lint 配置语法错误"
            fi
            
            # 检查关键配置项
            check_item
            if grep -q "disable-all: true" "$PROJECT_ROOT/.golangci.yml"; then
                log_success "已启用 disable-all 模式"
            else
                log_warning "建议启用 disable-all 模式以精确控制检查器"
            fi
            
            check_item
            if grep -q "errcheck" "$PROJECT_ROOT/.golangci.yml"; then
                log_success "已启用 errcheck 检查器"
            else
                log_error "缺少 errcheck 检查器配置"
            fi
            
            check_item
            if grep -q "gosec" "$PROJECT_ROOT/.golangci.yml"; then
                log_success "已启用 gosec 安全检查器"
            else
                log_error "缺少 gosec 安全检查器配置"
            fi
            
            check_item
            if grep -q "staticcheck" "$PROJECT_ROOT/.golangci.yml"; then
                log_success "已启用 staticcheck 检查器"
            else
                log_warning "建议启用 staticcheck 检查器"
            fi
            
        else
            log_warning "golangci-lint 未安装，跳过语法验证"
        fi
    else
        log_error ".golangci.yml 配置文件不存在"
    fi
}

# 验证 ESLint 配置
validate_eslint_config() {
    log "验证 ESLint 配置..."
    
    # Bootstrap 前端配置
    check_item
    if [ -f "$PROJECT_ROOT/frontend/.eslintrc.js" ]; then
        log_success "Bootstrap 前端 ESLint 配置存在"
        
        # 检查关键配置
        check_item
        if grep -q "@typescript-eslint/recommended" "$PROJECT_ROOT/frontend/.eslintrc.js"; then
            log_success "Bootstrap 前端已启用 TypeScript 推荐规则"
        else
            log_error "Bootstrap 前端缺少 TypeScript 推荐规则"
        fi
        
        check_item
        if grep -q "react-hooks/recommended" "$PROJECT_ROOT/frontend/.eslintrc.js"; then
            log_success "Bootstrap 前端已启用 React Hooks 规则"
        else
            log_error "Bootstrap 前端缺少 React Hooks 规则"
        fi
    else
        log_error "Bootstrap 前端 ESLint 配置不存在"
    fi
    
    # Ant Design 前端配置
    check_item
    if [ -f "$PROJECT_ROOT/frontend-vue/.eslintrc.js" ]; then
        log_success "Ant Design 前端 ESLint 配置存在"
        
        # 检查关键配置
        check_item
        if grep -q "@typescript-eslint/strict" "$PROJECT_ROOT/frontend-vue/.eslintrc.js"; then
            log_success "Ant Design 前端已启用 TypeScript 严格模式"
        else
            log_warning "Ant Design 前端建议启用 TypeScript 严格模式"
        fi
        
        check_item
        if grep -q "security/recommended" "$PROJECT_ROOT/frontend-vue/.eslintrc.js"; then
            log_success "Ant Design 前端已启用安全规则"
        else
            log_warning "Ant Design 前端建议启用安全规则"
        fi
        
        check_item
        if grep -q "sonarjs/recommended" "$PROJECT_ROOT/frontend-vue/.eslintrc.js"; then
            log_success "Ant Design 前端已启用 SonarJS 规则"
        else
            log_warning "Ant Design 前端建议启用 SonarJS 规则"
        fi
    else
        log_error "Ant Design 前端 ESLint 配置不存在"
    fi
}

# 验证 SonarQube 配置
validate_sonarqube_config() {
    log "验证 SonarQube 配置..."
    
    check_item
    if [ -f "$PROJECT_ROOT/sonar-project.properties" ]; then
        log_success "SonarQube 配置文件存在"
        
        # 检查关键配置项
        check_item
        if grep -q "sonar.projectKey=" "$PROJECT_ROOT/sonar-project.properties"; then
            log_success "已配置项目键"
        else
            log_error "缺少项目键配置"
        fi
        
        check_item
        if grep -q "sonar.sources=" "$PROJECT_ROOT/sonar-project.properties"; then
            log_success "已配置源代码路径"
        else
            log_error "缺少源代码路径配置"
        fi
        
        check_item
        if grep -q "sonar.exclusions=" "$PROJECT_ROOT/sonar-project.properties"; then
            log_success "已配置排除路径"
        else
            log_warning "建议配置排除路径"
        fi
        
        check_item
        if grep -q "sonar.go.coverage.reportPaths=" "$PROJECT_ROOT/sonar-project.properties"; then
            log_success "已配置 Go 覆盖率报告路径"
        else
            log_warning "建议配置 Go 覆盖率报告路径"
        fi
        
        check_item
        if grep -q "sonar.javascript.lcov.reportPaths=" "$PROJECT_ROOT/sonar-project.properties"; then
            log_success "已配置 JavaScript 覆盖率报告路径"
        else
            log_warning "建议配置 JavaScript 覆盖率报告路径"
        fi
        
    else
        log_error "SonarQube 配置文件不存在"
    fi
}

# 验证脚本文件
validate_scripts() {
    log "验证静态分析脚本..."
    
    local scripts=(
        "scripts/generate-static-analysis-report.sh"
        "scripts/ci-static-analysis.sh"
        "scripts/code-review-setup.sh"
        "scripts/validate-environment.sh"
    )
    
    for script in "${scripts[@]}"; do
        check_item
        if [ -f "$PROJECT_ROOT/$script" ]; then
            if [ -x "$PROJECT_ROOT/$script" ]; then
                log_success "$script 存在且可执行"
            else
                log_warning "$script 存在但不可执行"
                chmod +x "$PROJECT_ROOT/$script"
                log_success "已设置 $script 为可执行"
            fi
        else
            log_error "$script 不存在"
        fi
    done
}

# 验证 GitHub Actions 工作流
validate_github_workflows() {
    log "验证 GitHub Actions 工作流..."
    
    local workflows=(
        ".github/workflows/ci-cd.yml"
        ".github/workflows/security.yml"
        ".github/workflows/fuzzing.yml"
        ".github/workflows/pgo.yml"
    )
    
    for workflow in "${workflows[@]}"; do
        check_item
        if [ -f "$PROJECT_ROOT/$workflow" ]; then
            log_success "$workflow 存在"
            
            # 检查是否包含静态分析步骤
            if [ "$workflow" = ".github/workflows/ci-cd.yml" ]; then
                check_item
                if grep -q "integrated-static-analysis" "$PROJECT_ROOT/$workflow"; then
                    log_success "CI/CD 工作流包含集成静态分析"
                else
                    log_warning "CI/CD 工作流建议包含集成静态分析"
                fi
            fi
        else
            log_error "$workflow 不存在"
        fi
    done
}

# 验证工具可用性
validate_tool_availability() {
    log "验证静态分析工具可用性..."
    
    local go_tools=(
        "golangci-lint"
        "gosec"
        "govulncheck"
        "staticcheck"
    )
    
    local node_tools=(
        "eslint"
        "tsc"
        "prettier"
    )
    
    local system_tools=(
        "jq"
        "curl"
        "wget"
    )
    
    # 检查 Go 工具
    for tool in "${go_tools[@]}"; do
        check_item
        if command -v "$tool" >/dev/null 2>&1; then
            local version=$($tool --version 2>&1 | head -1 || echo "unknown")
            log_success "$tool 可用 ($version)"
        else
            log_warning "$tool 不可用，建议安装"
        fi
    done
    
    # 检查 Node.js 工具
    for tool in "${node_tools[@]}"; do
        check_item
        if command -v "$tool" >/dev/null 2>&1; then
            local version=$($tool --version 2>&1 | head -1 || echo "unknown")
            log_success "$tool 可用 ($version)"
        else
            log_warning "$tool 不可用，建议安装"
        fi
    done
    
    # 检查系统工具
    for tool in "${system_tools[@]}"; do
        check_item
        if command -v "$tool" >/dev/null 2>&1; then
            log_success "$tool 可用"
        else
            log_warning "$tool 不可用，建议安装"
        fi
    done
}

# 验证目录结构
validate_directory_structure() {
    log "验证目录结构..."
    
    local required_dirs=(
        "scripts"
        "docs/code-standards"
        ".github/workflows"
        "reports"
        "frontend"
        "frontend-vue"
        "internal"
    )
    
    for dir in "${required_dirs[@]}"; do
        check_item
        if [ -d "$PROJECT_ROOT/$dir" ]; then
            log_success "目录 $dir 存在"
        else
            if [ "$dir" = "reports" ]; then
                log_warning "目录 $dir 不存在，将在运行时创建"
            else
                log_error "目录 $dir 不存在"
            fi
        fi
    done
}

# 验证配置文件语法
validate_config_syntax() {
    log "验证配置文件语法..."
    
    # 验证 YAML 文件
    local yaml_files=(
        ".golangci.yml"
        ".github/workflows/ci-cd.yml"
        ".github/workflows/security.yml"
    )
    
    for yaml_file in "${yaml_files[@]}"; do
        check_item
        if [ -f "$PROJECT_ROOT/$yaml_file" ]; then
            # 简单的 YAML 语法检查
            if python3 -c "import yaml; yaml.safe_load(open('$PROJECT_ROOT/$yaml_file'))" 2>/dev/null; then
                log_success "$yaml_file YAML 语法正确"
            else
                log_error "$yaml_file YAML 语法错误"
            fi
        fi
    done
    
    # 验证 JavaScript 文件
    local js_files=(
        "frontend/.eslintrc.js"
        "frontend-vue/.eslintrc.js"
    )
    
    for js_file in "${js_files[@]}"; do
        check_item
        if [ -f "$PROJECT_ROOT/$js_file" ]; then
            # 简单的 JavaScript 语法检查
            if node -c "$PROJECT_ROOT/$js_file" 2>/dev/null; then
                log_success "$js_file JavaScript 语法正确"
            else
                log_error "$js_file JavaScript 语法错误"
            fi
        fi
    done
}

# 生成验证报告
generate_validation_report() {
    log "生成验证报告..."
    
    local report_file="$PROJECT_ROOT/reports/config-validation-$(date +%Y%m%d_%H%M%S).md"
    mkdir -p "$PROJECT_ROOT/reports"
    
    cat > "$report_file" << EOF
# 静态分析配置验证报告

**验证时间**: $(date)  
**项目**: Law OA Go v2.1.0  
**验证结果**: $([ "$VALIDATION_PASSED" = true ] && echo "✅ 通过" || echo "❌ 失败")  

## 📊 验证统计

- **总检查项**: $TOTAL_CHECKS
- **通过检查**: $PASSED_CHECKS
- **失败检查**: $((TOTAL_CHECKS - PASSED_CHECKS))
- **通过率**: $(( PASSED_CHECKS * 100 / TOTAL_CHECKS ))%

## 🔧 配置文件状态

| 配置文件 | 状态 | 说明 |
|----------|------|------|
| .golangci.yml | $([ -f "$PROJECT_ROOT/.golangci.yml" ] && echo "✅ 存在" || echo "❌ 缺失") | Go 静态分析配置 |
| frontend/.eslintrc.js | $([ -f "$PROJECT_ROOT/frontend/.eslintrc.js" ] && echo "✅ 存在" || echo "❌ 缺失") | Bootstrap 前端配置 |
| frontend-vue/.eslintrc.js | $([ -f "$PROJECT_ROOT/frontend-vue/.eslintrc.js" ] && echo "✅ 存在" || echo "❌ 缺失") | Ant Design 前端配置 |
| sonar-project.properties | $([ -f "$PROJECT_ROOT/sonar-project.properties" ] && echo "✅ 存在" || echo "❌ 缺失") | SonarQube 配置 |

## 🛠️ 工具可用性

| 工具 | 状态 | 版本 |
|------|------|------|
| golangci-lint | $(command -v golangci-lint >/dev/null 2>&1 && echo "✅ 可用" || echo "❌ 不可用") | $(command -v golangci-lint >/dev/null 2>&1 && golangci-lint --version 2>/dev/null | head -1 || echo "N/A") |
| gosec | $(command -v gosec >/dev/null 2>&1 && echo "✅ 可用" || echo "❌ 不可用") | $(command -v gosec >/dev/null 2>&1 && gosec -version 2>/dev/null || echo "N/A") |
| eslint | $(command -v eslint >/dev/null 2>&1 && echo "✅ 可用" || echo "❌ 不可用") | $(command -v eslint >/dev/null 2>&1 && eslint --version 2>/dev/null || echo "N/A") |

## 🎯 建议和后续行动

$(if [ "$VALIDATION_PASSED" = true ]; then
    echo "✅ 所有关键配置验证通过，静态分析环境已就绪。"
else
    echo "❌ 发现配置问题，请根据上述检查结果进行修复："
    echo ""
    echo "1. 安装缺失的工具"
    echo "2. 修复配置文件语法错误"
    echo "3. 补充缺失的配置项"
    echo "4. 重新运行验证脚本"
fi)

---
*报告由静态分析配置验证系统自动生成*
EOF
    
    log_success "验证报告生成完成: $report_file"
}

# 主函数
main() {
    log "开始静态分析配置验证..."
    
    cd "$PROJECT_ROOT"
    
    validate_golangci_config
    validate_eslint_config
    validate_sonarqube_config
    validate_scripts
    validate_github_workflows
    validate_tool_availability
    validate_directory_structure
    validate_config_syntax
    generate_validation_report
    
    echo ""
    log "验证完成！"
    log "总检查项: $TOTAL_CHECKS"
    log "通过检查: $PASSED_CHECKS"
    log "失败检查: $((TOTAL_CHECKS - PASSED_CHECKS))"
    log "通过率: $(( PASSED_CHECKS * 100 / TOTAL_CHECKS ))%"
    
    if [ "$VALIDATION_PASSED" = true ]; then
        log_success "静态分析配置验证通过！"
        exit 0
    else
        log_error "静态分析配置验证失败，请修复上述问题。"
        exit 1
    fi
}

# 执行主函数
main "$@"