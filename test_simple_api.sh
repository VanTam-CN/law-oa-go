#!/bin/bash

# 简单的API测试，不依赖用户注册

BASE_URL="http://localhost:8080"

echo "=== 测试Law OA Go 简单API ==="

# 尝试登录一个已存在的管理员用户
echo "1. 尝试登录管理员用户..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

echo "登录响应: ${LOGIN_RESPONSE}"

# 如果成功，提取token并测试冲突规则API
if echo "${LOGIN_RESPONSE}" | grep -q "token"; then
  TOKEN=$(echo ${LOGIN_RESPONSE} | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
  echo "获取到Token: ${TOKEN:0:20}..."

  echo -e "\n2. 测试冲突规则API..."
  RULES_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/conflict/rules" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}")

  echo "冲突规则响应: ${RULES_RESPONSE}"

  echo -e "\n3. 测试MCP标准API..."
  STANDARDS_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/conflict/standards" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}")

  echo "MCP标准响应: ${STANDARDS_RESPONSE}"

  echo -e "\n4. 测试健康检查API..."
  HEALTH_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/conflict/health" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}")

  echo "健康检查响应: ${HEALTH_RESPONSE}"
else
  echo "无法获取管理员token，可能需要先创建管理员用户"
fi

echo -e "\n=== 测试完成 ==="