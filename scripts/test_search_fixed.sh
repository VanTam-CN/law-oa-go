#!/bin/bash

# 获取认证 token
echo "获取认证 token..."
response=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"newuser@test.com","password":"simple123"}')

echo "登录响应: $response"

# 提取 token
token=$(echo $response | jq -r '.data.token // empty')

if [ -z "$token" ] || [ "$token" == "null" ]; then
    echo "无法获取 token，尝试创建新用户..."
    
    # 尝试注册用户
    register_response=$(curl -s -X POST "http://localhost:8080/api/v1/auth/register" \
      -H "Content-Type: application/json" \
      -d '{"name":"测试用户","email":"testuser@example.com","password":"simple123","role":"admin","phone":"13800138000"}')
    
    echo "注册响应: $register_response"
    
    # 重新尝试登录
    response=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
      -H "Content-Type: application/json" \
      -d '{"email":"testuser@example.com","password":"simple123"}')
    
    echo "重新登录响应: $response"
    token=$(echo $response | jq -r '.data.token // empty')
fi

if [ -n "$token" ] && [ "$token" != "null" ]; then
    echo "获取到 token: $token"
    
    # 测试搜索功能
    echo -e "\n测试搜索功能..."
    search_response=$(curl -s "http://localhost:8080/api/v1/search?q=合同" \
      -H "Authorization: Bearer $token")
    
    echo "搜索响应: $search_response"
    
    # 测试搜索建议功能
    echo -e "\n测试搜索建议功能..."
    suggest_response=$(curl -s "http://localhost:8080/api/v1/search/suggestions?q=合同" \
      -H "Authorization: Bearer $token")
    
    echo "搜索建议响应: $suggest_response"
    
else
    echo "认证失败，无法获取 token"
fi