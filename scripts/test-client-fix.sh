#!/bin/bash

# 客户管理界面修复测试脚本

echo "🔧 客户管理界面修复测试"
echo "===================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. 检查后端服务是否运行
echo -n "1. 检查后端服务状态 ... "
if curl -s http://localhost:8080/health >/dev/null; then
    echo -e "${GREEN}✅ 后端服务运行正常${NC}"
else
    echo -e "${RED}❌ 后端服务未运行，请先启动后端服务${NC}"
    echo "   启动命令: cd /Users/mac/Desktop/FT/law-oa-go && go run cmd/server/main.go"
    exit 1
fi

# 2. 检查前端服务是否运行
echo -n "2. 检查前端服务状态 ... "
if curl -s http://localhost:3003 >/dev/null; then
    echo -e "${GREEN}✅ 前端服务运行正常${NC}"
else
    echo -e "${RED}❌ 前端服务未运行${NC}"
    echo "   请启动前端开发服务器: cd frontend && npm start"
    exit 1
fi

# 3. 尝试登录获取token
echo "3. 尝试获取认证Token..."
echo "========================"

LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{
        "email": "admin@example.com",
        "password": "password123"
    }')

if echo "$LOGIN_RESPONSE" | grep -q "success.*true"; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$TOKEN" ]; then
        echo -e "${GREEN}✅ 登录成功，获取到Token${NC}"
        AUTH_HEADER="Authorization: Bearer $TOKEN"
    else
        echo -e "${YELLOW}⚠️ 登录成功但未能提取Token${NC}"
        AUTH_HEADER=""
    fi
else
    echo -e "${YELLOW}⚠️ 使用默认登录凭据失败，尝试其他用户${NC}"
    AUTH_HEADER=""
fi

# 4. 测试客户API
echo ""
echo "4. 测试客户管理API..."
echo "=================="

if [ -n "$AUTH_HEADER" ]; then
    echo -e "${BLUE}使用认证Token测试API...${NC}"

    # 测试客户列表API
    echo -n "测试客户列表API ... "
    CLIENTS_RESPONSE=$(curl -s "http://localhost:8080/api/clients?page=1&page_size=5" \
        -H "$AUTH_HEADER")

    if echo "$CLIENTS_RESPONSE" | grep -q "success.*true"; then
        echo -e "${GREEN}✅ 客户列表API正常${NC}"

        # 检查数据格式
        if echo "$CLIENTS_RESPONSE" | grep -q '"data".*\['; then
            echo -e "${GREEN}✅ 数据格式正确，包含data字段${NC}"
        else
            echo -e "${YELLOW}⚠️ 数据格式可能有问题${NC}"
        fi

        if echo "$CLIENTS_RESPONSE" | grep -q '"pagination"'; then
            echo -e "${GREEN}✅ 包含分页信息${NC}"
        else
            echo -e "${YELLOW}⚠️ 缺少分页信息${NC}"
        fi

        # 显示响应预览
        echo "响应预览:"
        echo "$CLIENTS_RESPONSE" | head -10
        echo "..."

    else
        echo -e "${RED}❌ 客户列表API调用失败${NC}"
        echo "响应内容:"
        echo "$CLIENTS_RESPONSE"
    fi

    # 测试客户统计API
    echo ""
    echo -n "测试客户统计API ... "
    STATS_RESPONSE=$(curl -s "http://localhost:8080/api/clients/stats" \
        -H "$AUTH_HEADER")

    if echo "$STATS_RESPONSE" | grep -q "success.*true"; then
        echo -e "${GREEN}✅ 客户统计API正常${NC}"
    else
        echo -e "${RED}❌ 客户统计API调用失败${NC}"
        echo "响应内容:"
        echo "$STATS_RESPONSE"
    fi

else
    echo -e "${YELLOW}⚠️ 无认证Token，尝试无认证测试...${NC}"

    # 测试无认证访问
    echo -n "测试无认证客户列表访问 ... "
    NO_AUTH_RESPONSE=$(curl -s "http://localhost:8080/api/clients?page=1&page_size=5")

    if echo "$NO_AUTH_RESPONSE" | grep -q "Unauthorized\|未授权"; then
        echo -e "${GREEN}✅ API正确要求认证${NC}"
    else
        echo -e "${YELLOW}⚠️ API可能不需要认证或有其他问题${NC}"
        echo "响应内容:"
        echo "$NO_AUTH_RESPONSE" | head -5
    fi
fi

# 5. 数据库检查建议
echo ""
echo "5. 数据库检查建议..."
echo "=================="

echo "如果客户管理界面仍然显示空白，请检查："
echo ""
echo "a) 数据库中是否有客户数据："
echo "   运行: cd /Users/mac/Desktop/FT/law-oa-go && go run scripts/seed-clients.go"
echo ""
echo "b) 检查前端控制台日志："
echo "   - 打开浏览器开发者工具 (F12)"
echo "   - 查看Console和Network标签页"
echo "   - 刷新客户管理页面，查看API调用和响应"
echo ""
echo "c) 验证API响应格式："
echo "   运行: ./scripts/debug-client-api.sh"
echo ""

# 6. 前端修复说明
echo "6. 已应用的前端修复..."
echo "=================="
echo ""
echo "✅ 增强了API响应数据的处理逻辑"
echo "✅ 添加了多种数据格式的兼容性支持"
echo "✅ 增加了调试日志输出"
echo "✅ 修复了客户详情显示中的空值处理"
echo ""
echo "修复内容："
echo "- 支持新的统一API格式: {success: true, data: [...], pagination: {...}}"
echo "- 兼容旧格式: {data: [...], total: ...}"
echo "- 添加了调试日志，便于排查问题"
echo ""

echo "🎯 测试完成！"
echo ""
echo "如果问题仍然存在，请："
echo "1. 运行 go run scripts/seed-clients.go 添加测试数据"
echo "2. 检查浏览器控制台的详细错误信息"
echo "3. 使用 ./scripts/debug-client-api.sh 进行深度调试"