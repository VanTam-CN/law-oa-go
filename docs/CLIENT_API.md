# 客户管理模块 API 文档

## 概述

客户管理模块提供完整的客户信息管理功能，包括客户的基本信息维护、数据验证、批量导入导出等功能。

## 功能特性

- ✅ 客户信息的增删改查
- ✅ 数据验证和业务规则检查
- ✅ 分页查询和条件筛选
- ✅ 客户搜索功能
- ✅ 批量导入（支持 CSV/Excel）
- ✅ 批量删除功能
- ✅ 客户统计信息
- ✅ 身份证号验证
- ✅ 重复数据检查

## API 接口

### 1. 获取客户列表

**请求**
```http
GET /api/v1/clients?page=1&size=10&keyword=张三&client_type=individual&status=active
Authorization: Bearer <token>
```

**查询参数**
- `page`: 页码（默认：1）
- `size`: 每页大小（默认：10，最大：100）
- `keyword`: 搜索关键词（可选，支持客户名称、手机号、邮箱、公司名称）
- `client_type`: 客户类型（可选：individual/company）
- `status`: 状态（可选：active/inactive）
- `sort_field`: 排序字段（可选：id/client_name/created_at）
- `sort_order`: 排序方式（可选：asc/desc）

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "clients": [
      {
        "id": 1,
        "client_name": "张三",
        "phone": "13800138001",
        "email": "zhangsan@example.com",
        "client_type": "individual",
        "company": "",
        "id_card": "110101199001011234",
        "address": "北京市朝阳区",
        "contact_person": "",
        "status": "active",
        "remark": "",
        "created_at": "2024-01-01 10:00:00",
        "updated_at": "2024-01-01 10:00:00",
        "case_count": 3
      }
    ],
    "total": 100,
    "page": 1,
    "size": 10
  }
}
```

### 2. 获取客户详情

**请求**
```http
GET /api/v1/clients/1
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "id": 1,
    "client_name": "张三",
    "phone": "13800138001",
    "email": "zhangsan@example.com",
    "client_type": "individual",
    "company": "",
    "id_card": "110101199001011234",
    "address": "北京市朝阳区",
    "contact_person": "",
    "status": "active",
    "remark": "",
    "created_at": "2024-01-01 10:00:00",
    "updated_at": "2024-01-01 10:00:00",
    "case_count": 3
  }
}
```

### 3. 创建客户

**请求**
```http
POST /api/v1/clients
Authorization: Bearer <token>
Content-Type: application/json

{
  "client_name": "李四",
  "phone": "13800138002",
  "email": "lisi@example.com",
  "client_type": "individual",
  "id_card": "110101199002021234",
  "address": "北京市海淀区",
  "contact_person": "",
  "remark": ""
}
```

**请求体参数**
- `client_name`: 客户名称（必填，2-100字符）
- `phone`: 手机号（必填，11位数字）
- `email`: 邮箱（必填，有效邮箱格式）
- `client_type`: 客户类型（必填：individual/company）
- `company`: 公司名称（企业客户必填，最多100字符）
- `id_card`: 身份证号（个人客户必填，18位）
- `address`: 地址（必填，最多255字符）
- `contact_person`: 联系人（可选，最多50字符）
- `remark`: 备注（可选，最多255字符）

**响应**
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": 2,
    "client_name": "李四",
    "phone": "13800138002",
    "email": "lisi@example.com"
  }
}
```

### 4. 更新客户

**请求**
```http
PUT /api/v1/clients/2
Authorization: Bearer <token>
Content-Type: application/json

{
  "client_name": "李四（更新）",
  "phone": "13800138002",
  "email": "lisi-updated@example.com",
  "client_type": "individual",
  "id_card": "110101199002021234",
  "address": "北京市海淀区更新地址",
  "contact_person": "",
  "status": "active",
  "remark": "更新备注"
}
```

**响应**
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": 2,
    "client_name": "李四（更新）",
    "phone": "13800138002",
    "email": "lisi-updated@example.com"
  }
}
```

### 5. 删除客户

**请求**
```http
DELETE /api/v1/clients/2
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "删除成功",
  "data": null
}
```

### 6. 获取客户统计信息

**请求**
```http
GET /api/v1/clients/stats
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total_clients": 150,
    "individual_count": 100,
    "company_count": 50,
    "active_count": 140,
    "inactive_count": 10,
    "monthly_new_clients": 25
  }
}
```



### 8. 批量导入客户 (待实现)

> **注意：** 此功能当前未实现。

**请求**
```http
POST /api/v1/clients/import
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: [文件数据]
```

**文件格式要求**
- 支持 CSV 和 Excel 文件
- 文件大小不超过 10MB
- CSV 格式示例：
```csv
客户名称,手机号,邮箱,客户类型,地址,公司名称,身份证号,联系人,备注
张三,13800138001,zhangsan@example.com,individual,北京市朝阳区,,110101199001011234,,
某公司,13800138003,company@example.com,company,北京市朝阳区,某科技有限公司,,王经理,
```

**响应**
```json
{
  "code": 200,
  "message": "导入成功",
  "data": {
    "success_count": 95,
    "duplicate_count": 3,
    "error_count": 2,
    "total_count": 100
  }
}
```

### 9. 批量删除客户

**请求**
```http
POST /api/v1/clients/batch-delete
Authorization: Bearer <token>
Content-Type: application/json

{
  "ids": [1, 2, 3]
}
```

**响应**
```json
{
  "code": 200,
  "message": "删除成功",
  "data": {
    "deleted_count": 3
  }
}
```

## 数据验证规则

### 客户名称
- 必填字段
- 长度：2-100 个字符

### 手机号
- 必填字段
- 格式：11位数字，以1开头
- 支持的手机号段：13x, 14x, 15x, 16x, 17x, 18x, 19x

### 邮箱
- 必填字段
- 格式：标准的邮箱格式（xxx@xxx.xxx）

### 客户类型
- 必填字段
- 值：individual（个人）或 company（企业）

### 身份证号
- 个人客户必填
- 格式：18位，包含校验码
- 支持自动校验身份证号有效性

### 地址
- 必填字段
- 长度：最多255个字符

### 公司名称
- 企业客户必填
- 长度：最多100个字符

## 错误响应

所有 API 错误响应都遵循统一格式：

```json
{
  "code": 400,
  "message": "错误描述",
  "data": null
}
```

### 常见错误码
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `409`: 数据冲突（如手机号重复）
- `429`: 请求过于频繁
- `500`: 服务器内部错误

## 业务规则

1. **数据唯一性**：手机号、邮箱、身份证号必须唯一
2. **删除限制**：有关联案件的客户无法删除
3. **导入限制**：重复数据会跳过，不会覆盖已有数据
4. **状态管理**：客户状态分为 active（活跃）和 inactive（停用）
5. **类型关联**：企业客户必须填写公司名称，个人客户必须填写身份证号

## 性能优化

1. **分页查询**：所有列表接口都支持分页
2. **索引优化**：关键字段都建立了数据库索引
3. **批量操作**：支持批量导入和删除，提高效率
4. **缓存策略**：统计信息可以考虑使用缓存
5. **搜索限制**：搜索结果限制前50条，提高响应速度

## 安全考虑

1. **认证授权**：所有接口都需要 JWT 认证
2. **数据验证**：严格的前后端数据验证
3. **敏感信息**：身份证号等敏感信息适当脱敏
4. **访问控制**：基于角色的权限控制
5. **操作日志**：重要操作记录日志