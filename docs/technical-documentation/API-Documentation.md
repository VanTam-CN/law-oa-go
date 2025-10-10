# Law OA Go API Documentation

**版本**: v2.1.0
**更新日期**: 2025-09-30
**API版本**: v1

---

## 📋 概述

Law OA Go 是律师事务所办公自动化系统的后端API，提供用户管理、案件管理、客户管理、权限控制等核心功能。

### 🌐 API 基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **内容类型**: `application/json`
- **字符编码**: UTF-8

### 🔧 技术栈

- **框架**: Go + Gin
- **数据库**: MySQL 8.0
- **ORM**: GORM
- **认证**: JWT
- **日志**: Zap Logger
- **监控**: Health Checks

---

## 🔐 认证机制

### JWT Token 认证

所有需要认证的API都需要在请求头中包含有效的JWT Token：

```http
Authorization: Bearer <your-jwt-token>
```

### 获取Token

```http
POST /api/v1/auth/login
Content-Type: application/json

{
    "username": "your-username",
    "password": "your-password"
}
```

**响应示例**:
```json
{
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "user": {
            "id": 1,
            "username": "admin",
            "email": "admin@example.com",
            "role": "admin"
        },
        "expires_in": 86400
    },
    "error": null
}
```

---

## 📊 通用响应格式

### 成功响应

```json
{
    "data": {
        // 响应数据
    },
    "error": null,
    "timestamp": "2025-09-30T12:00:00Z",
    "status": "success"
}
```

### 错误响应

```json
{
    "data": null,
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "请求参数验证失败",
        "details": {
            "field": "email",
            "reason": "邮箱格式不正确"
        }
    },
    "timestamp": "2025-09-30T12:00:00Z",
    "status": "error"
}
```

### HTTP状态码

| 状态码 | 说明 | 示例场景 |
|--------|------|----------|
| 200 | 请求成功 | 获取数据成功 |
| 201 | 创建成功 | 用户注册成功 |
| 400 | 请求参数错误 | 必填字段缺失 |
| 401 | 未授权 | Token无效或过期 |
| 403 | 禁止访问 | 权限不足 |
| 404 | 资源不存在 | 用户不存在 |
| 409 | 资源冲突 | 用户名已存在 |
| 500 | 服务器内部错误 | 数据库连接失败 |

---

## 👤 用户管理 API

### 用户注册

```http
POST /api/v1/auth/register
Content-Type: application/json

{
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "password123",
    "full_name": "New User",
    "phone": "13800138000"
}
```

**响应示例**:
```json
{
    "data": {
        "id": 2,
        "username": "newuser",
        "email": "newuser@example.com",
        "full_name": "New User",
        "phone": "13800138000",
        "role": "user",
        "created_at": "2025-09-30T12:00:00Z"
    },
    "error": null
}
```

### 获取用户信息

```http
GET /api/v1/users/profile
Authorization: Bearer <token>
```

### 更新用户信息

```http
PUT /api/v1/users/profile
Authorization: Bearer <token>
Content-Type: application/json

{
    "full_name": "Updated Name",
    "phone": "13900139000",
    "bio": "更新个人简介"
}
```

### 修改密码

```http
PUT /api/v1/users/password
Authorization: Bearer <token>
Content-Type: application/json

{
    "current_password": "oldpassword",
    "new_password": "newpassword123"
}
```

---

## 🏢 客户管理 API

### 获取客户列表

```http
GET /api/v1/clients?page=1&limit=10&search=张三
Authorization: Bearer <token>
```

**查询参数**:
- `page`: 页码（默认1）
- `limit`: 每页数量（默认10，最大100）
- `search`: 搜索关键词
- `client_type`: 客户类型（individual/company）

**响应示例**:
```json
{
    "data": {
        "clients": [
            {
                "id": 1,
                "name": "张三",
                "client_type": "individual",
                "contact_person": "张三",
                "phone": "13800138000",
                "email": "zhangsan@example.com",
                "address": "北京市朝阳区",
                "created_at": "2025-09-30T10:00:00Z"
            }
        ],
        "pagination": {
            "current_page": 1,
            "total_pages": 5,
            "total_count": 50,
            "per_page": 10
        }
    },
    "error": null
}
```

### 创建客户

```http
POST /api/v1/clients
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "李四公司",
    "client_type": "company",
    "contact_person": "李四",
    "phone": "13900139000",
    "email": "lisi@company.com",
    "address": "上海市浦东新区",
    "business_license": "91310000MA1FL0000X",
    "tax_number": "91310000MA1FL0000X"
}
```

### 更新客户信息

```http
PUT /api/v1/clients/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
    "contact_person": "王五",
    "phone": "13700137000"
}
```

### 删除客户

```http
DELETE /api/v1/clients/{id}
Authorization: Bearer <token>
```

---

## 📋 案件管理 API

### 获取案件列表

```http
GET /api/v1/cases?page=1&limit=10&status=active&client_id=1
Authorization: Bearer <token>
```

**查询参数**:
- `page`: 页码
- `limit`: 每页数量
- `status`: 案件状态（pending/active/completed/archived）
- `client_id`: 客户ID
- `lawyer_id`: 律师ID
- `case_type`: 案件类型
- `start_date`: 开始日期
- `end_date`: 结束日期

### 创建案件

```http
POST /api/v1/cases
Authorization: Bearer <token>
Content-Type: application/json

{
    "case_number": "LAW2025-001",
    "title": "合同纠纷案",
    "case_type": "contract_dispute",
    "client_id": 1,
    "description": "关于销售合同的纠纷案件",
    "estimated_value": 100000.00,
    "fee_structure": "hourly",
    "hourly_rate": 500.00
}
```

### 获取案件详情

```http
GET /api/v1/cases/{id}
Authorization: Bearer <token>
```

### 更新案件状态

```http
PUT /api/v1/cases/{id}/status
Authorization: Bearer <token>
Content-Type: application/json

{
    "status": "completed",
    "completion_note": "案件已顺利结案",
    "final_amount": 95000.00
}
```

---

## 👨‍💼 律师管理 API

### 获取律师列表

```http
GET /api/v1/lawyers?page=1&limit=10&specialization=contract
Authorization: Bearer <token>
```

### 创建律师档案

```http
POST /api/v1/lawyers
Authorization: Bearer <token>
Content-Type: application/json

{
    "name": "王律师",
    "bar_number": "京司律证字第12345号",
    "specialization": ["contract", "corporate"],
    "experience_years": 8,
    "education": "中国政法大学",
    "phone": "13600136000",
    "email": "lawyer@firm.com",
    "bio": "专业从事合同法和公司法业务"
}
```

---

## 📈 统计报表 API

### 获取仪表板数据

```http
GET /api/v1/dashboard/stats
Authorization: Bearer <token>
```

**响应示例**:
```json
{
    "data": {
        "total_cases": 156,
        "active_cases": 23,
        "total_clients": 89,
        "total_revenue": 2345678.90,
        "monthly_stats": [
            {
                "month": "2025-09",
                "new_cases": 12,
                "completed_cases": 8,
                "revenue": 123456.78
            }
        ],
        "case_type_distribution": [
            {
                "type": "contract_dispute",
                "count": 45,
                "percentage": 28.8
            }
        ]
    },
    "error": null
}
```

### 财务报表

```http
GET /api/v1/reports/finance?start_date=2025-01-01&end_date=2025-09-30
Authorization: Bearer <token>
```

---

## 🔍 搜索 API

### 全文搜索

```http
GET /api/v1/search?q=张三&type=clients,cases
Authorization: Bearer <token>
```

**查询参数**:
- `q`: 搜索关键词
- `type`: 搜索类型（逗号分隔）
- `page`: 页码
- `limit`: 每页数量

---

## 📄 文件管理 API

### 上传文件

```http
POST /api/v1/files
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary-data>
category: contract
case_id: 1
description: 合同扫描件
```

### 下载文件

```http
GET /api/v1/files/{id}/download
Authorization: Bearer <token>
```

---

## ⚙️ 系统配置 API

### 获取系统配置

```http
GET /api/v1/system/config
Authorization: Bearer <token>
```

### 更新系统配置

```http
PUT /api/v1/system/config
Authorization: Bearer <token>
Content-Type: application/json

{
    "firm_name": "北京某某律师事务所",
    "firm_address": "北京市朝阳区某某街道",
    "contact_phone": "010-12345678",
    "business_hours": "9:00-18:00"
}
```

---

## 🚨 错误代码说明

| 错误代码 | 说明 | HTTP状态码 |
|----------|------|------------|
| `VALIDATION_ERROR` | 请求参数验证失败 | 400 |
| `UNAUTHORIZED` | 未授权访问 | 401 |
| `FORBIDDEN` | 权限不足 | 403 |
| `NOT_FOUND` | 资源不存在 | 404 |
| `CONFLICT` | 资源冲突 | 409 |
| `INTERNAL_ERROR` | 服务器内部错误 | 500 |
| `DATABASE_ERROR` | 数据库操作错误 | 500 |
| `TOKEN_EXPIRED` | Token已过期 | 401 |
| `TOKEN_INVALID` | Token无效 | 401 |

---

## 🔄 API 版本控制

### 版本策略

- 使用语义化版本控制（Semantic Versioning）
- URL路径中包含版本号：`/api/v1/`, `/api/v2/`
- 向后兼容的更改不增加主版本号
- 破坏性更改会增加主版本号

### 版本生命周期

- **v1**: 当前稳定版本
- **v2**: 开发中版本
- 旧版本维护6个月的向后兼容性

---

## 📝 开发工具

### Postman Collection

提供完整的Postman集合文件，包含所有API的示例请求。

### Swagger/OpenAPI

访问 `/swagger/index.html` 查看交互式API文档。

### SDK支持

- Go SDK: `github.com/law-oa/go-sdk`
- JavaScript SDK: `@law-oa/js-sdk`

---

## 🛡️ 安全最佳实践

### 请求验证

1. **输入验证**: 所有输入参数都经过严格验证
2. **SQL注入防护**: 使用参数化查询
3. **XSS防护**: 输出内容转义
4. **CSRF防护**: 使用CSRF Token

### 认证安全

1. **密码加密**: 使用bcrypt加密存储
2. **Token安全**: JWT Token有效期控制
3. **会话管理**: 支持Token刷新机制

### 数据保护

1. **敏感数据**: 不在日志中记录敏感信息
2. **传输加密**: HTTPS加密传输
3. **访问控制**: 基于角色的权限控制

---

## 📞 技术支持

如有API相关问题，请联系：

- **技术团队**: dev-team@law-oa.com
- **文档反馈**: docs@law-oa.com
- **Bug报告**: GitHub Issues

---

**文档版本**: v2.1.0
**最后更新**: 2025-09-30
**下次审查**: 2025-10-30