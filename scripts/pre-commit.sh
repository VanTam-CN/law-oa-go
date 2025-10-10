#!/bin/bash

# Git pre-commit 钩子
# 在提交代码前自动运行代码质量检查

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# 获取暂存的文件
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(go|ts|tsx|js|jsx)$')

if [ -z "$STAGED_FILES" ]; then
    log_info "没有需要检查的代码文件，跳过pre-commit检查"
    exit 0
fi

log_info "开始pre-commit代码检查..."

# 检查Go文件
GO_FILES=$(echo "$STAGED_FILES" | grep -E '\.go$' || true)
if [ -n "$GO_FILES" ]; then
    log_info "检查Go文件..."

    # 格式化检查
    log_info "运行go fmt..."
    echo "$GO_FILES" | xargs go fmt
    if [ $? -ne 0 ]; then
        log_error "Go代码格式化失败"
        exit 1
    fi

    # 重新添加格式化后的文件
    echo "$GO_FILES" | xargs git add

    # 静态检查（快速模式）
    log_info "运行快速golangci-lint检查..."
    if ! golangci-lint run --fast --config=.golangci.yml $(dirname "$GO_FILES" | sort -u | sed 's/^/\.\//'); then
        log_error "Go代码静态检查发现问题，请修复后再提交"
        exit 1
    fi

    log_success "Go文件检查完成"
fi

# 检查前端文件（Bootstrap版本）
FRONTEND_FILES=$(echo "$STAGED_FILES" | grep -E 'frontend/.*\.(ts|tsx|js|jsx)$' || true)
if [ -n "$FRONTEND_FILES" ]; then
    log_info "检查前端文件（Bootstrap版本）..."

    cd frontend

    # ESLint检查
    if [ -f ".eslintrc.js" ]; then
        log_info "运行ESLint检查..."
        if ! npm run lint --silent; then
            log_error "前端代码ESLint检查发现问题，请修复后再提交"
            exit 1
        fi
    fi

    # TypeScript类型检查
    if [ -f "tsconfig.json" ]; then
        log_info "运行TypeScript类型检查..."
        if ! npm run type-check --silent; then
            log_error "TypeScript类型检查失败，请修复后再提交"
            exit 1
        fi
    fi

    cd ..
    log_success "前端文件（Bootstrap版本）检查完成"
fi

# 检查前端文件（Ant Design版本）
FRONTEND_VUE_FILES=$(echo "$STAGED_FILES" | grep -E 'frontend-vue/.*\.(ts|tsx|js|jsx)$' || true)
if [ -n "$FRONTEND_VUE_FILES" ]; then
    log_info "检查前端文件（Ant Design版本）..."

    cd frontend-vue

    # ESLint检查
    if [ -f ".eslintrc.js" ]; then
        log_info "运行ESLint检查..."
        if ! npm run lint --silent; then
            log_error "前端代码ESLint检查发现问题，请修复后再提交"
            exit 1
        fi
    fi

    # TypeScript类型检查
    if [ -f "tsconfig.json" ]; then
        log_info "运行TypeScript类型检查..."
        if ! npm run type-check --silent; then
            log_error "TypeScript类型检查失败，请修复后再提交"
            exit 1
        fi
    fi

    cd ..
    log_success "前端文件（Ant Design版本）检查完成"
fi

# 检查大文件
log_info "检查大文件..."
LARGE_FILES=$(git diff --cached --name-only | xargs -r ls -la | awk '$5 > 1048576 {print $9}')
if [ -n "$LARGE_FILES" ]; then
    log_warning "检测到大文件: $LARGE_FILES"
    log_warning "请确认这些文件需要提交"
fi

# 检查敏感信息
log_info "检查敏感信息..."
if git diff --cached | grep -q -i "password\|secret\|key\|token"; then
    log_warning "检测到可能的敏感信息，请确认"
fi

log_success "pre-commit检查完成，可以提交"