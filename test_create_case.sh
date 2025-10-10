#!/bin/bash

# 生成token
TOKEN=$(go run test_case_api.go 2>/dev/null | grep 'Generated Token:' | cut -d' ' -f3)

echo "使用Token: ${TOKEN:0:50}..."
echo ""

# 测试创建案件
echo "=== 测试创建新案件 ==="
curl -X POST "http://localhost:8080/api/cases" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试新建案件功能",
    "description": "通过API测试创建的案件，验证新建案件功能是否正常",
    "client_id": 1,
    "lawyer_id": 2,
    "case_type": "civil",
    "priority": "medium"
  }' | jq '.'

echo ""
echo "=== 验证案件创建成功 ==="
# 获取案件列表验证
curl -X GET "http://localhost:8080/api/cases?pageNum=1&pageSize=5" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | jq '.data[] | {id, title, description}'
