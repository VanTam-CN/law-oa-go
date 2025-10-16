#!/bin/bash

# 测试冲突检测API脚本

BASE_URL="http://localhost:8080"

echo "=== 测试Law OA Go 冲突检测API ==="

# 1. 注册测试用户
echo "1. 注册测试用户..."
REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test123456!",
    "fullName": "测试用户",
    "role": "ADMIN"
  }')

echo "注册响应: ${REGISTER_RESPONSE}"

# 2. 登录获取token
echo -e "\n2. 用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test123456!"
  }')

echo "登录响应: ${LOGIN_RESPONSE}"

# 提取token
TOKEN=$(echo ${LOGIN_RESPONSE} | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "无法获取token，退出测试"
  exit 1
fi

echo "获取到Token: ${TOKEN:0:20}..."

# 3. 测试冲突规则API
echo -e "\n3. 测试冲突规则API..."
RULES_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/conflict/rules" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}")

echo "冲突规则响应: ${RULES_RESPONSE}"

# 4. 测试MCP标准API
echo -e "\n4. 测试MCP标准API..."
STANDARDS_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/conflict/standards" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}")

echo "MCP标准响应: ${STANDARDS_RESPONSE}"

# 5. 测试健康检查API
echo -e "\n5. 测试健康检查API..."
HEALTH_RESPONSE=$(curl -s -X GET "${BASE_URL}/api/v1/conflict/health" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}")

echo "健康检查响应: ${HEALTH_RESPONSE}"

echo -e "\n=== 测试完成 ==="