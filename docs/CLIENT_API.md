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
- ✅ 自然人/企业通用强身份类型与加密存储
- ✅ 独立主联系人记录，电话和邮箱加密落库
- ✅ 重复数据检查

客户邮箱为客户主体的公共邮箱，不是主联系人邮箱。该字段可选：未填写时数据库保存为 `NULL`，多个无邮箱客户可以并存；非空邮箱仍要求唯一。`created_by` 由服务端根据当前登录账号写入，客户端不得传入或覆盖。创建律师在首个案件建立前可以继续查看和维护该客户，但任何后续隔离墙仍优先于创建人权限。

主联系人使用独立的 `/clients/:id/primary-contact` 接口。联系人电话和邮箱只以密文落库；只有通过该客户对象权限检查的账号才能读取，修改者还必须具有客户维护权限。旧 `clients.contact_person/contact_phone` 仅作为历史兼容显示，不再由新页面写入，也不得把脱敏后的旧电话回存为真实号码。

## API 接口

### 1. 获取客户列表

**请求**
```http
GET /api/v1/clients?page=1&page_size=10&search=张三&type=个人&status=active
Authorization: Bearer <token>
```

**查询参数**
- `page`: 页码（默认：1）
- `page_size`: 每页大小（默认：20，最大：100）
- `search`: 搜索关键词（可选；和 `name` 共用搜索入口，优先使用 `search`）
- `name`: 客户名称搜索（可选）
- `type`: 客户类型（可选：`个人`/`企业`）
- `status`: 状态（可选：active/inactive）
- `company`: 公司名称过滤（可选）

**响应**
```json
{
  "success": true,
  "data": {
    "clients": [
      {
        "id": 1,
        "version": 1,
        "name": "张三",
        "type": "个人",
        "phone": "138****8001",
        "email": "zhangsan@example.com",
        "company": "",
        "industry": "",
        "identity_type": "ID_CARD",
        "identity_number": "110***1234",
        "identity_status": "已登记（受保护）",
        "id_card": "110***1234",
        "aliases": "",
        "address": "北京市朝阳区",
        "contact_person": "",
        "contact_phone": "",
        "source": "",
        "status": "active",
        "notes": "",
        "created_at": "2024-01-01T10:00:00+08:00",
        "updated_at": "2024-01-01T10:00:00+08:00",
        "completeness": {
          "score": 100,
          "missing_fields": [],
          "ready_for_conflict_check": true
        },
        "related_parties": [],
        "historical_matters": []
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 10,
      "total": 100,
      "total_pages": 10,
      "has_next": true,
      "has_prev": false
    }
  },
  "meta": {
    "timestamp": "2026-08-24T12:00:00+08:00",
    "version": "v2.4.0",
    "server": "law-oa-go",
    "environment": "development"
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
  "success": true,
  "data": {
    "id": 1,
    "name": "张三",
    "type": "个人",
    "phone": "138****8001",
    "email": "zhangsan@example.com",
    "company": "",
    "identity_type": "ID_CARD",
    "identity_number": "110***1234",
    "identity_status": "已登记（受保护）",
    "address": "北京市朝阳区",
    "contact_person": "",
    "status": "active",
    "notes": "",
    "created_at": "2024-01-01 10:00:00",
    "updated_at": "2024-01-01 10:00:00",
    "version": 1
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
  "name": "李四",
  "type": "个人",
  "phone": "13800138002",
  "email": "lisi@example.com",
  "identity_type": "ID_CARD",
  "identity_number": "110101199002021234",
  "aliases": "",
  "address": "北京市海淀区",
  "contact_person": "",
  "notes": ""
}
```

**请求体参数**
- `name`: 客户名称（必填，1-100字符）
- `phone`: 联系电话（可选，最多20字符）
- `email`: 邮箱（可选，有效邮箱格式）
- `type`: 客户类型（必填：`个人`/`企业`）
- `company`: 公司名称（企业客户必填，最多100字符）
- `identity_type`: 身份类型；个人允许 `ID_CARD/PASSPORT/OTHER`，企业允许 `SOCIAL_CREDIT_CODE/BUSINESS_LICENSE/ORGANIZATION_CODE/OTHER`
- `identity_number`: 身份原文，仅在写入请求中传输；服务端加密后不会原样返回
- `aliases`: 别名/曾用名，多个值使用逗号或分号分隔
- `address`: 地址（可选，最多255字符）
- `contact_person`: 旧版兼容字段；新接入不得使用，请改用主联系人接口
- `notes`: 备注（可选，最多1000字符）

**响应**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "李四",
    "phone": "138****8002",
    "email": "lisi@example.com",
    "identity_type": "ID_CARD",
    "identity_number": "110***1234",
    "identity_status": "已登记（受保护）"
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
  "version": 1,
  "name": "李四（更新）",
  "phone": "13800138002",
  "email": "lisi-updated@example.com",
  "type": "个人",
  "identity_type": "ID_CARD",
  "identity_number": "110101199002021234",
  "address": "北京市海淀区更新地址",
  "contact_person": "",
  "status": "active",
  "notes": "更新备注"
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "李四（更新）",
    "phone": "138****8002",
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
  "success": true,
  "data": {
    "total": 150,
    "active_clients": 140,
    "inactive_clients": 10,
    "monthly_new": 25,
    "type_stats": null,
    "status_stats": null,
    "source_stats": null
  },
  "meta": {
    "timestamp": "2026-08-24T12:00:00+08:00",
    "version": "v2.4.0",
    "server": "law-oa-go",
    "environment": "development"
  }
}
```

### 7. 获取主联系人

```http
GET /api/v1/clients/2/primary-contact
Authorization: Bearer <token>
```

无独立联系人记录时 `data` 为 `null`。有记录时仅向有权查看该客户的账号返回：

```json
{
  "success": true,
  "data": {
    "id": 8,
    "client_id": 2,
    "name": "王示例",
    "position": "法务负责人",
    "phone": "021-55550000",
    "email": "legal@example.test",
    "is_primary": true,
    "version": 2
  }
}
```

### 8. 新建或更新主联系人

```http
PUT /api/v1/clients/2/primary-contact
Authorization: Bearer <token>
Content-Type: application/json

{
  "version": 2,
  "name": "王示例",
  "position": "总法律顾问",
  "phone": "021-55550001",
  "email": "gc@example.test"
}
```

首次创建时 `version` 传 `0` 或省略；更新时必须传当前联系人版本。版本过期返回 `409 CLIENT_CONTACT_VERSION_CONFLICT`，客户端应刷新后让用户确认，不得自动覆盖。保存联系人不会修改客户公共邮箱、客户电话或客户备注。



### 9. 批量导入客户 (待实现)

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
