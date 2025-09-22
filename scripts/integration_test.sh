#!/bin/bash

echo "=== 法律办公自动化系统集成测试 ==="

# 1. 测试健康检查
echo -e "\n1. 测试健康检查..."
curl -s "http://localhost:8080/health" | jq .

# 2. 测试用户认证
echo -e "\n2. 测试用户认证..."
TOKEN_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"newuser@test.com","password":"simple123"}')

echo "Token响应: $TOKEN_RESPONSE"

# 提取token
TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.data.token')
echo "提取的Token: $TOKEN"

# 3. 测试搜索功能（需要认证）
echo -e "\n3. 测试搜索功能..."
if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
  curl -s "http://localhost:8080/api/v1/search?q=合同" \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo "Token无效，跳过搜索测试"
fi

# 4. 测试案例管理
echo -e "\n4. 测试案例管理..."
if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
  curl -s "http://localhost:8080/api/v1/cases" \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo "Token无效，跳过案例管理测试"
fi

# 5. 测试客户管理
echo -e "\n5. 测试客户管理..."
if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
  curl -s "http://localhost:8080/api/v1/clients" \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo "Token无效，跳过客户管理测试"
fi

# 6. 测试文档管理
echo -e "\n6. 测试文档管理..."
if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
  curl -s "http://localhost:8080/api/v1/documents" \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo "Token无效，跳过文档管理测试"
fi

# 7. 测试监控状态
echo -e "\n7. 测试监控状态..."
curl -s "http://localhost:8080/api/v1/monitor/status" | jq .

# 8. 测试缓存性能
echo -e "\n8. 测试缓存性能..."
curl -s "http://localhost:8080/performance/cache" | jq .

# 9. 测试搜索建议
echo -e "\n9. 测试搜索建议..."
if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
  curl -s "http://localhost:8080/api/v1/search/suggestions?q=合同" \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo "Token无效，跳过搜索建议测试"
fi

echo -e "\n=== 集成测试完成 ==="