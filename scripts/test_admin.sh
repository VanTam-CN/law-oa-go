#!/bin/bash

# 测试API端点
BASE_URL="http://localhost:8080"

echo "=== 创建新用户并测试系统功能 ==="

# 1. 创建新用户
echo -e "\n1. 创建新用户..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "管理员用户",
    "email": "admin@lawoa.com",
    "password": "Admin123456",
    "role": "admin"
  }')
echo "$REGISTER_RESPONSE" | jq .

# 2. 用户登录
echo -e "\n2. 用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@lawoa.com",
    "password": "Admin123456"
  }')
echo "$LOGIN_RESPONSE" | jq .

# 提取token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // empty')
echo "获取的token: $TOKEN"

# 3. 测试各种功能
if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
  echo -e "\n3. 测试搜索功能..."
  curl -s "$BASE_URL/api/v1/search?query=知识产权" \
    -H "Authorization: Bearer $TOKEN" | jq .
    
  echo -e "\n4. 测试搜索建议..."
  curl -s "$BASE_URL/api/v1/search/suggestions?query=知识产权&limit=5" \
    -H "Authorization: Bearer $TOKEN" | jq .
    
  echo -e "\n5. 测试用户资料..."
  curl -s "$BASE_URL/api/v1/users/profile" \
    -H "Authorization: Bearer $TOKEN" | jq .
    
  echo -e "\n6. 测试客户端列表..."
  curl -s "$BASE_URL/api/v1/clients" \
    -H "Authorization: Bearer $TOKEN" | jq .
    
  echo -e "\n7. 测试案例列表..."
  curl -s "$BASE_URL/api/v1/cases" \
    -H "Authorization: Bearer $TOKEN" | jq .
    
  echo -e "\n8. 测试文档列表..."
  curl -s "$BASE_URL/api/v1/documents" \
    -H "Authorization: Bearer $TOKEN" | jq .
    
  echo -e "\n9. 测试性能监控..."
  curl -s "$BASE_URL/api/v1/monitor/performance" \
    -H "Authorization: Bearer $TOKEN" | jq .
    
  echo -e "\n10. 测试监控仪表板..."
  curl -s "$BASE_URL/api/v1/monitor/dashboard" \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo -e "\n3. 跳过功能测试（无法获取认证token）"
fi

echo -e "\n=== 测试完成 ==="