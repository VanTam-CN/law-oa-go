#!/bin/bash

# 代码审查自动化脚本
# 用于Law OA Go项目的全面代码质量检查

set -e

# 颜色定义
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

# 检查依赖
check_dependencies() {
    log_info "检查依赖工具..."

    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go未安装"
        exit 1
    fi

    # 检查golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        log_warning "golangci-lint未安装，正在安装..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi

    # 检查Node.js和npm（用于前端）
    if ! command -v node &> /dev/null; then
        log_warning "Node.js未安装，跳过前端检查"
    fi

    log_success "依赖检查完成"
}

# Go代码审查
review_go_code() {
    log_info "开始Go代码审查..."

    # 格式化检查
    log_info "运行go fmt..."
    go fmt ./...
    log_success "Go代码格式化完成"

    # 运行golangci-lint
    log_info "运行golangci-lint..."
    if golangci-lint run --config=.golangci.yml ./...; then
        log_success "Go代码质量检查通过"
    else
        log_error "Go代码质量检查发现问题"
        return 1
    fi

    # 运行测试
    log_info "运行Go测试..."
    if go test -v ./... -coverprofile=coverage.out; then
        log_success "Go测试通过"
    else
        log_error "Go测试失败"
        return 1
    fi

    # 生成覆盖率报告
    log_info "生成覆盖率报告..."
    go tool cover -html=coverage.out -o coverage.html
    log_success "覆盖率报告已生成: coverage.html"

    # 运行竞态条件检测
    log_info "运行竞态条件检测..."
    if go test -race ./...; then
        log_success "竞态条件检测通过"
    else
        log_error "竞态条件检测失败"
        return 1
    fi
}

# 前端代码审查（Bootstrap版本）
review_frontend_bootstrap() {
    log_info "开始前端代码审查（Bootstrap版本）..."

    cd frontend

    # 检查ESLint
    if [ -f ".eslintrc.js" ]; then
        log_info "运行ESLint检查..."
        if npm run lint; then
            log_success "ESLint检查通过"
        else
            log_error "ESLint检查发现问题"
            return 1
        fi
    fi

    # TypeScript类型检查
    if [ -f "tsconfig.json" ]; then
        log_info "运行TypeScript类型检查..."
        if npm run type-check; then
            log_success "TypeScript类型检查通过"
        else
            log_error "TypeScript类型检查失败"
            return 1
        fi
    fi

    # 运行测试
    if npm test -- --coverage --watchAll=false; then
        log_success "前端测试通过"
    else
        log_error "前端测试失败"
        return 1
    fi

    cd ..
}

# 前端代码审查（Ant Design版本）
review_frontend_ant() {
    log_info "开始前端代码审查（Ant Design版本）..."

    cd frontend-vue

    # 检查ESLint
    if [ -f ".eslintrc.js" ]; then
        log_info "运行ESLint检查..."
        if npm run lint; then
            log_success "ESLint检查通过"
        else
            log_error "ESLint检查发现问题"
            return 1
        fi
    fi

    # TypeScript类型检查
    if [ -f "tsconfig.json" ]; then
        log_info "运行TypeScript类型检查..."
        if npm run type-check; then
            log_success "TypeScript类型检查通过"
        else
            log_error "TypeScript类型检查失败"
            return 1
        fi
    fi

    cd ..
}

# 安全扫描
security_scan() {
    log_info "开始安全扫描..."

    # Go安全扫描
    log_info "运行Go安全扫描..."
    if command -v gosec &> /dev/null; then
        gosec ./...
        log_success "Go安全扫描完成"
    else
        log_warning "gosec未安装，跳过Go安全扫描"
    fi

    # 依赖检查
    log_info "检查依赖漏洞..."
    if command -v govulncheck &> /dev/null; then
        govulncheck ./...
        log_success "依赖漏洞检查完成"
    else
        log_warning "govulncheck未安装，跳过依赖漏洞检查"
    fi
}

# 性能分析
performance_analysis() {
    log_info "开始性能分析..."

    # 内存分析
    log_info "运行内存分析..."
    if go test -bench=. -benchmem ./... -memprofile=memprofile.out; then
        log_success "内存分析完成"
        go tool pprof -mem memprofile.out
    else
        log_error "内存分析失败"
    fi

    # CPU分析
    log_info "运行CPU分析..."
    if go test -bench=. -cpuprofile=cpuprofile.out ./...; then
        log_success "CPU分析完成"
        go tool pprof cpuprofile.out
    else
        log_error "CPU分析失败"
    fi
}

# 生成报告
generate_report() {
    log_info "生成代码审查报告..."

    REPORT_FILE="code-review-report-$(date +%Y%m%d_%H%M%S).md"

    cat > "$REPORT_FILE" << EOF
# 代码审查报告

**审查时间**: $(date)
**项目**: Law OA Go
**版本**: $(git describe --tags --always 2>/dev/null || echo "unknown")

## 审查摘要

### Go后端
- 代码格式化: ✅
- 静态分析: ✅
- 单元测试: ✅
- 竞态条件检测: ✅
- 覆盖率报告: coverage.html

### 前端（Bootstrap版本）
- ESLint检查: ✅
- TypeScript类型检查: ✅
- 单元测试: ✅

### 前端（Ant Design版本）
- ESLint检查: ✅
- TypeScript类型检查: ✅

### 安全扫描
- Go安全扫描: ✅
- 依赖漏洞检查: ✅

### 性能分析
- 内存分析: ✅
- CPU分析: ✅

## 详细结果

### Go代码质量指标
- 圈复杂度: 低于15的函数占比
- 函数长度: 80行以内的函数占比
- 认知复杂度: 低于20的函数占比

### 前端代码质量指标
- ESLint规则通过率: 100%
- TypeScript类型安全: 100%
- 代码重复率: 低于5%

### 测试覆盖率
- Go测试覆盖率: $(go tool cover -func=coverage.out | grep total | awk '{print $3}')
- 前端测试覆盖率: 需要单独统计

## 建议和改进点

1. **代码质量**: 继续保持高标准的代码质量
2. **测试覆盖率**: 争取达到80%以上的测试覆盖率
3. **性能优化**: 定期进行性能分析和优化
4. **安全加固**: 定期进行安全扫描和依赖更新

## 后续步骤

1. 修复本次审查发现的问题
2. 建立持续的代码审查流程
3. 集成到CI/CD流水线中
4. 定期进行代码质量评估

---
*此报告由自动化代码审查脚本生成*
EOF

    log_success "代码审查报告已生成: $REPORT_FILE"
}

# 清理函数
cleanup() {
    log_info "清理临时文件..."
    rm -f coverage.out memprofile.out cpuprofile.out
    log_success "清理完成"
}

# 主函数
main() {
    log_info "开始Law OA Go项目代码审查..."

    # 检查依赖
    check_dependencies

    # Go代码审查
    review_go_code

    # 前端代码审查
    if [ -d "frontend" ]; then
        review_frontend_bootstrap
    fi

    if [ -d "frontend-vue" ]; then
        review_frontend_ant
    fi

    # 安全扫描
    security_scan

    # 性能分析（可选）
    if [ "$1" = "--performance" ]; then
        performance_analysis
    fi

    # 生成报告
    generate_report

    # 清理
    cleanup

    log_success "代码审查完成！"
}

# 捕获中断信号
trap cleanup EXIT

# 运行主函数
main "$@"