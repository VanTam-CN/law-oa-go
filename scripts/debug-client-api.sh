#!/bin/bash

# 客户API调试脚本

echo "🔍 调试客户管理API..."
echo ""

API_BASE="http://localhost:8080/api"
TOKEN=""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查API连通性
check_api() {
    local url=$1
    local description=$2

    echo -n "检查 $description ... "

    response=$(curl -s -w "%{http_code}" "$url" -o /tmp/api_response.json)
    http_code="${response: -3}"

    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✅ 正常 ($http_code)${NC}"
        return 0
    elif [ "$http_code" = "401" ]; then
        echo -e "${YELLOW}⚠️ 需要认证 ($http_code)${NC}"
        return 1
    elif [ "$http_code" = "404" ]; then
        echo -e "${RED}❌ 不存在 ($http_code)${NC}"
        return 2
    else
        echo -e "${RED}❌ 错误 ($http_code)${NC}"
        if [ -f /tmp/api_response.json ]; then
            echo "响应内容:"
            cat /tmp/api_response.json | head -3
        fi
        return 3
    fi
}

# 格式化JSON显示
show_json() {
    local file=$1
    local title=$2

    if [ -f "$file" ]; then
        echo ""
        echo -e "${BLUE}$title:${NC}"
        echo "----------------------------------------"
        if command -v jq >/dev/null 2>&1; then
            cat "$file" | jq '.'
        else
            cat "$file" | python3 -m json.tool 2>/dev/null || cat "$file"
        fi
        echo "----------------------------------------"
    fi
}

# 1. 检查基础连通性
echo "📡 1. 检查API基础连通性"
echo "========================="

check_api "$API_BASE/health" "健康检查端点"
check_api "$API_BASE/" "API根路径"

echo ""

# 2. 检查认证端点（不需要认证）
echo "🔐 2. 检查认证相关端点"
echo "=========================="

check_api "$API_BASE/auth/login" "登录端点（GET）"

echo ""

# 3. 尝试登录获取token
echo "🎫 3. 尝试获取认证Token"
echo "======================="

echo "尝试使用默认用户登录..."
login_response=$(curl -s -X POST "$API_BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "admin@example.com",
        "password": "password123"
    }' -w "%{http_code}" -o /tmp/login_response.json)

login_code="${login_response: -3}"

if [ "$login_code" = "200" ]; then
    echo -e "${GREEN}✅ 登录成功${NC}"

    if command -v jq >/dev/null 2>&1; then
        TOKEN=$(cat /tmp/login_response.json | jq -r '.data.token // .token // empty')
    else
        # 简单的token提取（备用方案）
        TOKEN=$(grep -o '"token":"[^"]*"' /tmp/login_response.json | cut -d'"' -f4)
    fi

    if [ -n "$TOKEN" ]; then
        echo -e "${GREEN}✅ 获取到Token: ${TOKEN:0:20}...${NC}"
    else
        echo -e "${YELLOW}⚠️ 未能从响应中提取Token${NC}"
        show_json "/tmp/login_response.json" "登录响应"
    fi
else
    echo -e "${RED}❌ 登录失败 ($login_code)${NC}"
    show_json "/tmp/login_response.json" "登录响应"
fi

echo ""

# 4. 检查客户API端点
echo "👥 4. 检查客户管理API端点"
echo "==========================="

if [ -n "$TOKEN" ]; then
    AUTH_HEADER="Authorization: Bearer $TOKEN"

    echo "使用认证Token检查客户API..."

    # 检查客户列表
    echo -n "检查客户列表API ... "
    clients_response=$(curl -s -w "%{http_code}" "$API_BASE/clients?page=1&page_size=5" \
        -H "$AUTH_HEADER" -o /tmp/clients_response.json)
    clients_code="${clients_response: -3}"

    if [ "$clients_code" = "200" ]; then
        echo -e "${GREEN}✅ 正常 ($clients_code)${NC}"
        show_json "/tmp/clients_response.json" "客户列表响应"

        # 分析数据结构
        if [ -f /tmp/clients_response.json ]; then
            echo ""
            echo -e "${BLUE}📊 数据结构分析:${NC}"
            echo "----------------------------------------"
            if command -v jq >/dev/null 2>&1; then
                echo "顶级键:"
                cat /tmp/clients_response.json | jq 'keys' 2>/dev/null || echo "无法解析JSON"

                echo ""
                echo "data字段类型:"
                cat /tmp/clients_response.json | jq -r '.data | type' 2>/dev/null || echo "无法确定data类型"

                echo ""
                echo "pagination信息:"
                cat /tmp/clients_response.json | jq '.pagination' 2>/dev/null || echo "无pagination字段"

                echo ""
                echo "data字段内容预览:"
                cat /tmp/clients_response.json | jq '.data[0:2]' 2>/dev/null || echo "无法显示data内容"
            fi
        fi
    else
        echo -e "${RED}❌ 错误 ($clients_code)${NC}"
        show_json "/tmp/clients_response.json" "客户列表错误响应"
    fi

    echo ""

    # 检查客户统计
    echo -n "检查客户统计API ... "
    stats_response=$(curl -s -w "%{http_code}" "$API_BASE/clients/stats" \
        -H "$AUTH_HEADER" -o /tmp/stats_response.json)
    stats_code="${stats_response: -3}"

    if [ "$stats_code" = "200" ]; then
        echo -e "${GREEN}✅ 正常 ($stats_code)${NC}"
        show_json "/tmp/stats_response.json" "客户统计响应"
    else
        echo -e "${RED}❌ 错误 ($stats_code)${NC}"
        show_json "/tmp/stats_response.json" "客户统计错误响应"
    fi

else
    echo -e "${YELLOW}⚠️ 跳过需要认证的API检查（无Token）${NC}"
fi

echo ""

# 5. 检查数据库状态
echo "🗄️ 5. 检查数据库状态"
echo "====================="

if [ -n "$TOKEN" ]; then
    echo -n "检查数据库连接状态 ... "
    db_response=$(curl -s -w "%{http_code}" "$API_BASE/health/database" \
        -H "$AUTH_HEADER" -o /tmp/db_response.json)
    db_code="${db_response: -3}"

    if [ "$db_code" = "200" ]; then
        echo -e "${GREEN}✅ 数据库连接正常${NC}"
        show_json "/tmp/db_response.json" "数据库状态"
    else
        echo -e "${RED}❌ 数据库连接异常 ($db_code)${NC}"
        show_json "/tmp/db_response.json" "数据库错误"
    fi
else
    echo -e "${YELLOW}⚠️ 跳过数据库检查（无Token）${NC}"
fi

echo ""

# 6. 建议和解决方案
echo "💡 6. 诊断建议"
echo "=============="

echo "如果客户管理界面看不到数据，可能的原因："
echo ""
echo "1. ${RED}数据库中没有客户数据${NC}"
echo "   - 解决方案：先创建一些测试客户数据"
echo ""
echo "2. ${RED}前端API调用参数不匹配${NC}"
echo "   - 检查前端发送的查询参数格式"
echo "   - 确认前端期望的响应数据结构"
echo ""
echo "3. ${RED}认证问题${NC}"
echo "   - 确认前端正确发送了认证Token"
echo "   - 检查Token是否有效"
echo ""
echo "4. ${RED}API响应格式不匹配${NC}"
echo "   - 后端返回新格式，前端可能需要适配"
echo "   - 检查前端是否正确解析pagination字段"
echo ""

echo "📋 调试完成！请查看上述响应数据来确定具体问题。"

# 清理临时文件
rm -f /tmp/api_response.json /tmp/login_response.json /tmp/clients_response.json /tmp/stats_response.json /tmp/db_response.json