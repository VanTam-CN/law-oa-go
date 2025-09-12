# 律师管理模块 API 文档

## 概述

律师管理模块提供完整的律师信息管理功能，包括律师的基本信息维护、数据验证、案件关联统计等功能。

## 功能特性

- ✅ 律师信息的增删改查
- ✅ 数据验证和业务规则检查
- ✅ 分页查询和条件筛选
- ✅ 律师搜索功能
- ✅ 批量删除功能
- ✅ 律师统计信息
- ✅ 执业证号验证
- ✅ 案件关联统计

## API 接口

### 1. 获取律师列表

**请求**
```http
GET /api/v1/lawyers?page=1&size=10&keyword=张三&department=诉讼部&position=合伙人&status=active
Authorization: Bearer <token>
```

**查询参数**
- `page`: 页码（默认：1）
- `size`: 每页大小（默认：10，最大：100）
- `keyword`: 搜索关键词（可选，支持律师姓名、手机号、邮箱、执业证号、专长领域）
- `department`: 部门（可选）
- `position`: 职位（可选）
- `status`: 状态（可选：active/inactive/suspended）
- `sort_field`: 排序字段（可选：id/lawyer_name/created_at）
- `sort_order`: 排序方式（可选：asc/desc）

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "lawyers": [
      {
        "id": 1,
        "lawyer_name": "张三",
        "phone": "13800138001",
        "email": "zhangsan@lawfirm.com",
        "license_no": "京2020123456",
        "position": "合伙人",
        "department": "诉讼部",
        "specialty": "民商法,公司法",
        "status": "active",
        "remark": "资深律师",
        "created_at": "2024-01-01 10:00:00",
        "updated_at": "2024-01-01 10:00:00",
        "case_count": 15,
        "assist_count": 8
      }
    ],
    "total": 50,
    "page": 1,
    "size": 10
  }
}
```

### 2. 获取律师详情

**请求**
```http
GET /api/v1/lawyers/1
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "id": 1,
    "lawyer_name": "张三",
    "phone": "13800138001",
    "email": "zhangsan@lawfirm.com",
    "license_no": "京2020123456",
    "position": "合伙人",
    "department": "诉讼部",
    "specialty": "民商法,公司法",
    "status": "active",
    "remark": "资深律师",
    "created_at": "2024-01-01 10:00:00",
    "updated_at": "2024-01-01 10:00:00",
    "case_count": 15,
    "assist_count": 8
  }
}
```

### 3. 创建律师

**请求**
```http
POST /api/v1/lawyers
Authorization: Bearer <token>
Content-Type: application/json

{
  "lawyer_name": "李四",
  "phone": "13800138002",
  "email": "lisi@lawfirm.com",
  "license_no": "京2021123457",
  "position": "高级律师",
  "department": "诉讼部",
  "specialty": "刑法,刑事诉讼法",
  "remark": "刑事辩护专家"
}
```

**请求体参数**
- `lawyer_name`: 律师姓名（必填，2-50字符）
- `phone`: 手机号（必填，11位数字）
- `email`: 邮箱（必填，有效邮箱格式）
- `license_no`: 执业证号（必填，10-50字符）
- `position`: 职位（必填，最多50字符）
- `department`: 部门（必填，最多100字符）
- `specialty`: 专长领域（必填，最多255字符）
- `remark`: 备注（可选，最多255字符）

**响应**
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": 2,
    "lawyer_name": "李四",
    "phone": "13800138002",
    "email": "lisi@lawfirm.com",
    "license_no": "京2021123457"
  }
}
```

### 4. 更新律师

**请求**
```http
PUT /api/v1/lawyers/2
Authorization: Bearer <token>
Content-Type: application/json

{
  "lawyer_name": "李四（更新）",
  "phone": "13800138002",
  "email": "lisi-updated@lawfirm.com",
  "license_no": "京2021123457",
  "position": "高级律师",
  "department": "诉讼部",
  "specialty": "刑法,刑事诉讼法,经济犯罪辩护",
  "status": "active",
  "remark": "刑事辩护专家，经济犯罪案件经验丰富"
}
```

**响应**
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": 2,
    "lawyer_name": "李四（更新）",
    "phone": "13800138002",
    "email": "lisi-updated@lawfirm.com",
    "license_no": "京2021123457"
  }
}
```

### 5. 删除律师

**请求**
```http
DELETE /api/v1/lawyers/2
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

### 6. 获取律师统计信息

**请求**
```http
GET /api/v1/lawyers/stats
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total_lawyers": 50,
    "active_count": 45,
    "inactive_count": 3,
    "suspended_count": 2,
    "department_stats": {
      "诉讼部": 20,
      "非诉部": 15,
      "知识产权部": 10,
      "金融部": 5
    },
    "position_stats": {
      "合伙人": 8,
      "高级律师": 15,
      "律师": 20,
      "实习律师": 7
    },
    "monthly_new_lawyers": 3
  }
}
```

### 7. 搜索律师

**请求**
```http
GET /api/v1/lawyers/search?keyword=张三
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "搜索成功",
  "data": {
    "lawyers": [
      {
        "id": 1,
        "lawyer_name": "张三",
        "phone": "13800138001",
        "email": "zhangsan@lawfirm.com",
        "license_no": "京2020123456",
        "position": "合伙人",
        "department": "诉讼部",
        "specialty": "民商法,公司法",
        "status": "active",
        "remark": "资深律师",
        "created_at": "2024-01-01 10:00:00",
        "updated_at": "2024-01-01 10:00:00",
        "case_count": 15,
        "assist_count": 8
      }
    ],
    "total": 1
  }
}
```

### 8. 获取律师下拉列表（用于选择框）

**请求**
```http
GET /api/v1/lawyers/select-options
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "lawyer_name": "张三",
      "phone": "13800138001",
      "license_no": "京2020123456",
      "position": "合伙人",
      "department": "诉讼部",
      "label": "张三 (合伙人) - 诉讼部"
    },
    {
      "id": 2,
      "lawyer_name": "李四",
      "phone": "13800138002",
      "license_no": "京2021123457",
      "position": "高级律师",
      "department": "诉讼部",
      "label": "李四 (高级律师) - 诉讼部"
    }
  ]
}
```

### 9. 批量删除律师

**请求**
```http
POST /api/v1/lawyers/batch-delete
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

### 律师姓名
- 必填字段
- 长度：2-50 个字符

### 手机号
- 必填字段
- 格式：11位数字，以1开头
- 支持的手机号段：13x, 14x, 15x, 16x, 17x, 18x, 19x

### 邮箱
- 必填字段
- 格式：标准的邮箱格式（xxx@xxx.xxx）

### 执业证号
- 必填字段
- 长度：10-50 个字符
- 格式：省份缩写 + 字母 + 数字（如：京2020123456）

### 职位
- 必填字段
- 长度：最多50个字符

### 部门
- 必填字段
- 长度：最多100个字符

### 专长领域
- 必填字段
- 长度：最多255个字符
- 支持多个专长，用逗号、分号分隔

### 状态
- 可选字段
- 值：active（活跃）、inactive（停用）、suspended（暂停）

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
- `409`: 数据冲突（如执业证号重复）
- `429`: 请求过于频繁
- `500`: 服务器内部错误

## 业务规则

1. **数据唯一性**：执业证号、手机号、邮箱必须唯一
2. **删除限制**：有关联案件的律师无法删除
3. **批量操作**：批量删除最多支持100个律师
4. **状态管理**：律师状态分为 active（活跃）、inactive（停用）、suspended（暂停）
5. **案件关联**：统计律师主办和协办的案件数量

## 性能优化

1. **分页查询**：所有列表接口都支持分页
2. **索引优化**：关键字段都建立了数据库索引
3. **批量操作**：支持批量删除，提高效率
4. **缓存策略**：统计信息可以考虑使用缓存
5. **搜索限制**：搜索结果限制前50条，提高响应速度

## 安全考虑

1. **认证授权**：所有接口都需要 JWT 认证
2. **数据验证**：严格的前后端数据验证
3. **敏感信息**：执业证号等敏感信息适当保护
4. **访问控制**：基于角色的权限控制
5. **操作日志**：重要操作记录日志

## 相关数据模型

### 律师 (Lawyer)
- id: 律师ID
- lawyer_name: 律师姓名
- phone: 手机号
- email: 邮箱
- license_no: 执业证号
- position: 职位
- department: 部门
- specialty: 专长领域
- status: 状态
- remark: 备注
- created_at: 创建时间
- updated_at: 更新时间

### 关联字段
- case_count: 主办案件数量
- assist_count: 协办案件数量