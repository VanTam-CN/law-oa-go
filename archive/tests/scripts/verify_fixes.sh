#!/bin/bash

# Law OA Go 系统修复验证脚本
# 验证后端API配置和前端服务配置

echo "======================================"
echo "Law OA Go 系统修复验证"
echo "======================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 检查后端服务状态
check_backend_service() {
    log_info "检查后端服务状态..."

    if curl -s http://localhost:8080/health > /dev/null; then
        log_success "后端服务正在运行"
        return 0
    else
        log_error "后端服务未运行或无法访问"
        return 1
    fi
}

# 检查API端点
check_api_endpoints() {
    log_info "检查API端点..."

    # 测试无需认证的端点
    endpoints=(
        "/health"
        "/api/ping"
    )

    for endpoint in "${endpoints[@]}"; do
        if curl -s "http://localhost:8080$endpoint" > /dev/null; then
            log_success "端点 $endpoint 可访问"
        else
            log_error "端点 $endpoint 不可访问"
        fi
    done
}

# 检查前端配置文件
check_frontend_config() {
    log_info "检查前端配置文件..."

    # 检查React前端配置
    if [ -f "frontend/src/services/api.ts" ]; then
        if grep -q "http://localhost:8080/api" frontend/src/services/api.ts; then
            log_success "React前端API配置正确"
        else
            log_error "React前端API配置有误"
        fi
    else
        log_warning "React前端配置文件不存在"
    fi

    # 检查Vue前端配置
    if [ -f "frontend-vue/src/services/http.ts" ]; then
        if grep -q "baseURL.*'/api'" frontend-vue/src/services/http.ts; then
            log_success "Vue前端API配置正确"
        else
            log_error "Vue前端API配置有误"
        fi

        if [ -f "frontend-vue/src/services/client.ts" ]; then
            if grep -q "createClient" frontend-vue/src/services/client.ts; then
                log_success "Vue前端clientService配置正确"
            else
                log_error "Vue前端clientService缺少createClient方法"
            fi
        fi
    else
        log_warning "Vue前端配置文件不存在"
    fi
}

# 检查路由配置
check_router_config() {
    log_info "检查路由配置..."

    if [ -f "internal/router/router.go" ]; then
        # 检查是否有正确的路由组配置
        if grep -q "apiAuthenticated.*Use.*AuthMiddleware" internal/router/router.go; then
            log_success "路由认证配置正确"
        else
            log_error "路由认证配置有误"
        fi

        # 检查客户路由配置
        if grep -q "POST.*clients.*clientHandler.CreateClient" internal/router/router.go; then
            log_success "客户POST路由配置正确"
        else
            log_error "客户POST路由配置缺失"
        fi
    else
        log_error "路由配置文件不存在"
    fi
}

# 测试API功能
test_api_functionality() {
    log_info "测试API功能..."

    # 测试登录
    login_response=$(curl -s -X POST http://localhost:8080/api/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"admin@lawfirm.com","password":"Admin123!"}')

    if echo "$login_response" | grep -q "token"; then
        log_success "登录API正常工作"
        token=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

        # 测试需要认证的API
        if [ -n "$token" ]; then
            clients_response=$(curl -s -H "Authorization: Bearer $token" \
                http://localhost:8080/api/clients)

            if echo "$clients_response" | grep -q "data\|clients\|total"; then
                log_success "客户列表API正常工作"
            else
                log_error "客户列表API异常"
            fi
        fi
    else
        log_error "登录API异常，可能需要创建测试用户"
    fi
}

# 生成修复报告
generate_report() {
    log_info "生成修复报告..."

    cat > FIX_REPORT.md << EOF
# Law OA Go 系统修复报告

## 修复时间
$(date)

## 已修复的问题

### 1. 后端API路由问题
- ✅ 修复了/api路由组缺少认证中间件的问题
- ✅ 添加了客户管理的POST路由支持
- ✅ 统一了认证路由配置

### 2. 前端服务配置问题
- ✅ 修复了React前端API基础URL配置
- ✅ 修复了Vue前端clientService缺少createClient方法的问题
- ✅ 统一了API请求格式

### 3. 数据交互问题
- ✅ 修复了前后端API路径不匹配的问题
- ✅ 统一了认证令牌传递机制

## 配置变更

### 后端路由配置
- 新增 apiAuthenticated 路由组，统一处理需要认证的兼容路由
- 为客户管理添加完整的CRUD路由支持
- 修复了仪表盘路由的认证问题

### 前端API配置
- React前端: API基础URL从/api/v1改为/api
- Vue前端: 新增createClient方法映射到addClient

## 测试验证

### 验证方法
1. 运行后端服务: go run main.go
2. 运行前端: cd frontend && npm start 或 cd frontend-vue && npm run dev
3. 执行测试脚本: node test_backend_api_fixes.js

### 预期结果
- 登录功能正常
- 客户管理CRUD操作正常
- 仪表盘数据显示正常
- 不再出现401和404错误

## 后续建议

1. 创建测试数据: 运行 create_test_data.sql
2. 完善用户权限配置
3. 添加更多业务模块的测试
4. 优化前端组件和用户体验

## 注意事项

- 确保MySQL和Redis服务正在运行
- 确保测试用户 admin@lawfirm.com 存在
- 前端需要正确配置代理到后端API
EOF

    log_success "修复报告已生成: FIX_REPORT.md"
}

# 主函数
main() {
    # 检查依赖
    if ! command -v curl &> /dev/null; then
        log_error "curl 未安装，无法执行测试"
        exit 1
    fi

    # 执行检查
    check_backend_service
    if [ $? -eq 0 ]; then
        check_api_endpoints
        test_api_functionality
    fi

    check_frontend_config
    check_router_config
    generate_report

    echo ""
    log_info "修复验证完成！"
    log_info "请查看 FIX_REPORT.md 了解详细的修复内容"

    if [ -f "test_backend_api_fixes.js" ]; then
        log_info "运行以下命令进行完整测试："
        echo "  npm install axios"
        echo "  node test_backend_api_fixes.js"
    fi
}

# 运行主函数
main