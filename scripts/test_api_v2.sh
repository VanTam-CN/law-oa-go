#!/bin/bash

# 测试API端点
BASE_URL="http://localhost:8080"

echo "=== 重新测试法律办公自动化系统 API ==="

# 1. 测试用户注册（使用更强的密码）
echo -e "\n1. 测试用户注册（使用强密码）..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试用户",
    "email": "test@example.com",
    "password": "Test123456",
    "role": "user"
  }')
echo "$REGISTER_RESPONSE" | jq .

# 2. 测试用户登录
echo -e "\n2. 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123456"
  }')
echo "$LOGIN_RESPONSE" | jq .

# 提取token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // empty')
echo "获取的token: $TOKEN"

# 3. 测试搜索功能（带认证）
if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
  echo -e "\n3. 测试搜索功能..."
  curl -s "$BASE_URL/api/v1/search?query=合同" \
    -H "Authorization: Bearer $TOKEN" | jq .
    
  echo -e "\n4. 测试搜索建议功能..."
  curl -s "$BASE_URL/api/v1/search/suggestions?query=合同&limit=5" \
    -H "Authorization: Bearer $TOKEN" | jq .
    
  echo -e "\n5. 测试用户资料..."
  curl -s "$BASE_URL/api/v1/users/profile" \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo -e "\n3. 跳过搜索功能测试（无法获取认证token）"
fi

echo -e "\n=== 测试完成 ==="