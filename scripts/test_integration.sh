#!/bin/bash

# 获取认证token
echo "获取认证token..."
TOKEN_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"newuser@test.com","password":"simple123"}')

# 提取token
TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.data.token')
echo "Token: $TOKEN"

# 测试搜索功能
echo -e "\n=== 测试搜索功能 ==="
curl -s "http://localhost:8080/api/v1/search?q=合同" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 测试搜索建议功能
echo -e "\n=== 测试搜索建议功能 ==="
curl -s "http://localhost:8080/api/v1/search/suggestions?q=合同" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 测试缓存性能
echo -e "\n=== 测试缓存性能 ==="
curl -s "http://localhost:8080/performance/cache" | jq .

# 测试健康检查
echo -e "\n=== 测试健康检查 ==="
curl -s "http://localhost:8080/api/v1/health" | jq .

# 测试监控系统
echo -e "\n=== 测试监控系统 ==="
curl -s "http://localhost:8080/api/v1/monitor/status" | jq .

# 测试案例管理
echo -e "\n=== 测试案例管理 ==="
curl -s "http://localhost:8080/api/v1/cases" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 测试客户管理
echo -e "\n=== 测试客户管理 ==="
curl -s "http://localhost:8080/api/v1/clients" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 测试文档管理
echo -e "\n=== 测试文档管理 ==="
curl -s "http://localhost:8080/api/v1/documents" \
  -H "Authorization: Bearer $TOKEN" | jq .