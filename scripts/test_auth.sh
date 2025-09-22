#!/bin/bash

# 创建测试用户
echo "创建测试用户..."
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "email": "test@example.com",
    "real_name": "测试用户"
  }'

# 登录获取token
echo -e "\n登录获取token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }')

echo "登录响应: $TOKEN_RESPONSE"

# 提取token
TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"access_token":"[^"]*' | sed 's/"access_token":"//')
echo "获取到的Token: $TOKEN"