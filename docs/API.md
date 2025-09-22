# Law OA Go API 文档

## 🚀 概述

Law OA Go 提供现代化的 RESTful API，基于 Go 1.23+ 标准构建，采用统一响应格式、类型安全处理和完善的中间件体系。

### 核心特性
- ✅ **统一响应格式**: 所有API使用一致的响应结构
- ✅ **类型安全**: 基于 Go generics 的类型安全处理
- ✅ **现代中间件**: CORS、限流、JWT认证、超时控制
- ✅ **完善错误处理**: 标准化错误响应和建议信息
- ✅ **请求追踪**: 每个请求都有唯一ID用于调试
- ✅ **分页支持**: 统一的分页参数和响应格式

### 基础信息
- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: Bearer Token (JWT)
- **数据格式**: JSON
- **API版本**: v1
- **Go版本**: 1.23+

### 📋 实现状态
- ✅ **认证系统**: 登录、注册、JWT令牌管理
- ✅ **用户管理**: 完整的CRUD操作和权限控制
- ✅ **客户管理**: 客户信息和统计功能
- ✅ **案件管理**: 案件CRUD、律师分配、状态管理
- ⚠️ **文档管理**: 基础框架已建立
- ❌ **利益冲突检查**: 待实现
- ❌ **报告生成**: 待实现
- ❌ **邮件通知**: 待实现
- ❌ **财务管理**: 待实现

---

## 📊 统一响应格式

所有API响应都遵循统一格式，确保客户端处理的一致性：

### 成功响应
```json
{
  "success": true,
  "data": {
    // 响应数据
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 分页响应
```json
{
  "success": true,
  "data": [
    // 数据数组
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 错误响应
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": "Email format invalid",
    "suggestions": [
      "Check email format"
    ]
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

---

## 🔐 认证系统

### 登录
**POST** `/api/v1/auth/login`

**请求示例**
```json
{
  "email": "admin@lawfirm.com",
  "password": "password123"
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-12-31T23:59:59Z",
    "user": {
      "id": 1,
      "name": "系统管理员",
      "email": "admin@lawfirm.com",
      "role": "admin",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

**错误响应**
```json
{
  "success": false,
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "Invalid credentials",
    "details": "Email or password is incorrect",
    "suggestions": [
      "Check your email and password",
      "Ensure account is active"
    ]
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 注册
**POST** `/api/v1/auth/register`

**请求示例**
```json
{
  "name": "新用户",
  "email": "newuser@example.com",
  "password": "password123",
  "phone": "13800138000",
  "role": "user"
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-12-31T23:59:59Z",
    "user": {
      "id": 2,
      "name": "新用户",
      "email": "newuser@example.com",
      "role": "user",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 刷新令牌
**POST** `/api/v1/auth/refresh`

**请求示例**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-12-31T23:59:59Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

---

## 👤 用户管理

### 获取用户列表
**GET** `/api/v1/admin/users`

**查询参数**
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20, 最大: 100)
- `search`: 搜索关键词
- `role`: 角色过滤
- `status`: 状态过滤

**请求示例**
```
GET /api/v1/admin/users?page=1&page_size=10&search=admin&role=admin
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "系统管理员",
      "email": "admin@lawfirm.com",
      "role": "admin",
      "status": "active",
      "phone": "13800138000",
      "avatar": "https://example.com/avatar.jpg",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 1,
    "total_pages": 1
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 获取用户详情
**GET** `/api/v1/admin/users/{id}`

**请求示例**
```
GET /api/v1/admin/users/1
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "系统管理员",
    "email": "admin@lawfirm.com",
    "role": "admin",
    "status": "active",
    "phone": "13800138000",
    "avatar": "https://example.com/avatar.jpg",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 创建用户
**POST** `/api/v1/admin/users`

**请求示例**
```json
{
  "name": "新用户",
  "email": "newuser@example.com",
  "password": "password123",
  "phone": "13800138000",
  "role": "user",
  "status": "active"
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "新用户",
    "email": "newuser@example.com",
    "role": "user",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 更新用户
**PUT** `/api/v1/admin/users/{id}`

**请求示例**
```json
{
  "name": "更新用户",
  "email": "updated@example.com",
  "phone": "13800138001",
  "status": "active"
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "更新用户",
    "email": "updated@example.com",
    "phone": "13800138001",
    "status": "active",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 删除用户
**DELETE** `/api/v1/admin/users/{id}`

**请求示例**
```
DELETE /api/v1/admin/users/2
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "message": "User deleted successfully"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

---

## 👥 客户管理

### 获取客户列表
**GET** `/api/v1/clients`

**查询参数**
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20, 最大: 100)
- `search`: 搜索关键词
- `status`: 状态过滤

**请求示例**
```
GET /api/v1/clients?page=1&page_size=10&search=张&status=active
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com",
      "phone": "13900139000",
      "address": "北京市朝阳区",
      "company": "某科技有限公司",
      "notes": "重要客户",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 1,
    "total_pages": 1
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 获取客户详情
**GET** `/api/v1/clients/{id}`

**请求示例**
```
GET /api/v1/clients/1
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13900139000",
    "address": "北京市朝阳区",
    "company": "某科技有限公司",
    "notes": "重要客户",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 创建客户
**POST** `/api/v1/clients`

**请求示例**
```json
{
  "name": "李四",
  "email": "lisi@example.com",
  "phone": "13900139001",
  "address": "北京市海淀区",
  "company": "某科技公司",
  "notes": "新客户"
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "李四",
    "email": "lisi@example.com",
    "phone": "13900139001",
    "address": "北京市海淀区",
    "company": "某科技公司",
    "notes": "新客户",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 更新客户
**PUT** `/api/v1/clients/{id}`

**请求示例**
```json
{
  "name": "李四（更新）",
  "email": "lisi-updated@example.com",
  "phone": "13900139002",
  "address": "北京市西城区",
  "notes": "重要客户"
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "name": "李四（更新）",
    "email": "lisi-updated@example.com",
    "phone": "13900139002",
    "address": "北京市西城区",
    "company": "某科技公司",
    "notes": "重要客户",
    "status": "active",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 删除客户
**DELETE** `/api/v1/clients/{id}`

**请求示例**
```
DELETE /api/v1/clients/2
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "message": "Client deleted successfully"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 获取客户统计
**GET** `/api/v1/clients/stats`

**请求示例**
```
GET /api/v1/clients/stats
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "total": 50,
    "active": 45,
    "inactive": 5,
    "by_status": {
      "active": 45,
      "inactive": 5
    },
    "created_this_month": 8,
    "updated_this_month": 12
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

---

## ⚖️ 案件管理

### 获取案件列表
**GET** `/api/v1/cases`

**查询参数**
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20, 最大: 100)
- `search`: 搜索关键词
- `case_type`: 案件类型过滤
- `status`: 状态过滤
- `priority`: 优先级过滤

**请求示例**
```
GET /api/v1/cases?page=1&page_size=10&search=合同&case_type=civil&status=pending
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "张三诉李四合同纠纷",
      "description": "合同纠纷案件描述",
      "client_id": 1,
      "lawyer_id": 1,
      "case_type": "civil",
      "priority": "medium",
      "status": "pending",
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": null,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "client": {
        "id": 1,
        "name": "张三"
      },
      "lawyer": {
        "id": 1,
        "name": "李律师"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 1,
    "total_pages": 1
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 获取案件详情
**GET** `/api/v1/cases/{id}`

**请求示例**
```
GET /api/v1/cases/1
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "张三诉李四合同纠纷",
    "description": "合同纠纷案件描述",
    "client_id": 1,
    "lawyer_id": 1,
    "case_type": "civil",
    "priority": "medium",
    "status": "pending",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": null,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "client": {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com",
      "phone": "13900139000"
    },
    "lawyer": {
      "id": 1,
      "name": "李律师",
      "email": "lawyer@example.com"
    }
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 创建案件
**POST** `/api/v1/cases`

**请求示例**
```json
{
  "title": "某科技公司商标侵权案",
  "description": "商标侵权案件描述",
  "client_id": 1,
  "lawyer_id": 1,
  "case_type": "commercial",
  "priority": "high",
  "status": "pending",
  "start_date": "2024-01-01T00:00:00Z"
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "title": "某科技公司商标侵权案",
    "description": "商标侵权案件描述",
    "client_id": 1,
    "lawyer_id": 1,
    "case_type": "commercial",
    "priority": "high",
    "status": "pending",
    "start_date": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 更新案件
**PUT** `/api/v1/cases/{id}`

**请求示例**
```json
{
  "title": "某科技公司商标侵权案（更新）",
  "description": "更新后的案件描述",
  "priority": "urgent",
  "status": "active"
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 2,
    "title": "某科技公司商标侵权案（更新）",
    "description": "更新后的案件描述",
    "client_id": 1,
    "lawyer_id": 1,
    "case_type": "commercial",
    "priority": "urgent",
    "status": "active",
    "start_date": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 删除案件
**DELETE** `/api/v1/cases/{id}`

**请求示例**
```
DELETE /api/v1/cases/2
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "message": "Case deleted successfully"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 分配律师
**POST** `/api/v1/cases/{id}/assign`

**请求示例**
```json
{
  "lawyer_id": 2
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "message": "Lawyer assigned successfully"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 更新案件状态
**POST** `/api/v1/cases/{id}/status`

**请求示例**
```json
{
  "status": "active"
}
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "message": "Case status updated successfully"
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### 获取案件统计
**GET** `/api/v1/cases/stats`

**请求示例**
```
GET /api/v1/cases/stats
Authorization: Bearer <token>
```

**成功响应**
```json
{
  "success": true,
  "data": {
    "total": 100,
    "pending": 20,
    "active": 50,
    "closed": 30,
    "by_type": {
      "civil": 40,
      "commercial": 35,
      "criminal": 15,
      "administrative": 10
    },
    "by_priority": {
      "low": 20,
      "medium": 50,
      "high": 25,
      "urgent": 5
    },
    "by_status": {
      "pending": 20,
      "active": 50,
      "closed": 30
    }
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

---

## 📋 错误处理

### 标准错误码

| 错误码 | HTTP状态 | 说明 |
|--------|----------|------|
| `VALIDATION_ERROR` | 400 | 请求参数验证失败 |
| `AUTHENTICATION_ERROR` | 401 | 认证失败 |
| `AUTHORIZATION_ERROR` | 403 | 权限不足 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `CONFLICT` | 409 | 资源冲突 |
| `RATE_LIMIT_ERROR` | 429 | 请求频率超限 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |

### 错误响应示例

#### 验证错误
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": "Email format is invalid",
    "suggestions": [
      "Check email format",
      "Ensure email is not empty"
    ]
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

#### 认证错误
```json
{
  "success": false,
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "Authentication failed",
    "details": "Invalid token or token expired",
    "suggestions": [
      "Check your token",
      "Re-authenticate if needed"
    ]
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

#### 权限错误
```json
{
  "success": false,
  "error": {
    "code": "AUTHORIZATION_ERROR",
    "message": "Insufficient permissions",
    "details": "Admin role required for this action",
    "suggestions": [
      "Contact your administrator",
      "Request appropriate permissions"
    ]
  },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

---

## 📊 数据模型

### 用户模型
```json
{
  "id": 1,
  "name": "系统管理员",
  "email": "admin@lawfirm.com",
  "role": "admin",
  "phone": "13800138000",
  "avatar": "https://example.com/avatar.jpg",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 客户模型
```json
{
  "id": 1,
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13900139000",
  "address": "北京市朝阳区",
  "company": "某科技有限公司",
  "notes": "重要客户",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 案件模型
```json
{
  "id": 1,
  "title": "张三诉李四合同纠纷",
  "description": "合同纠纷案件描述",
  "client_id": 1,
  "lawyer_id": 1,
  "case_type": "civil",
  "priority": "medium",
  "status": "pending",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": null,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "client": {
    "id": 1,
    "name": "张三",
    "email": "zhangsan@example.com"
  },
  "lawyer": {
    "id": 1,
    "name": "李律师",
    "email": "lawyer@example.com"
  }
}
```

---

## 🔧 API限制和安全

### 请求限制
- **频率限制**: 每分钟100次请求
- **文件上传**: 单文件最大50MB
- **JWT有效期**: 24小时
- **请求超时**: 30秒

### 安全特性
- **JWT认证**: 基于角色的访问控制
- **CORS保护**: 配置允许的跨域访问
- **请求验证**: 输入参数严格验证
- **错误处理**: 不暴露敏感信息
- **请求追踪**: 每个请求唯一ID用于调试

### 中间件链
1. **CORS** - 跨域资源共享
2. **Rate Limit** - 请求频率限制
3. **Request ID** - 请求ID生成
4. **Timeout** - 请求超时控制
5. **Authentication** - JWT认证
6. **Authorization** - 权限验证
7. **Validation** - 参数验证
8. **Handler** - 业务处理

---

## 🚀 快速开始

### 1. 获取认证令牌
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@lawfirm.com",
    "password": "password123"
  }'
```

### 2. 使用令牌访问API
```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. 创建客户
```bash
curl -X POST http://localhost:8080/api/v1/clients \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13900139000",
    "address": "北京市朝阳区"
  }'
```

### 4. 创建案件
```bash
curl -X POST http://localhost:8080/api/v1/cases \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "合同纠纷案",
    "description": "合同纠纷案件",
    "client_id": 1,
    "case_type": "civil",
    "priority": "medium"
  }'
```

---

## 📝 OpenAPI规范

完整的OpenAPI 3.0规范文件位于 `docs/openapi.yaml`，包含所有API接口的详细定义，可以使用Swagger UI或其他OpenAPI工具进行查看和测试。

### 使用Swagger UI
1. 启动服务器：`go run cmd/server/main.go`
2. 访问：`http://localhost:8080/swagger/index.html`
3. 在线测试所有API接口

---

## 🔧 开发指南

### 代码结构
```
internal/
├── api/           # API响应格式和基础处理器
├── handlers/      # 业务处理器
├── middleware/    # 中间件
├── services/      # 业务逻辑
├── repositories/  # 数据访问层
└── errors/        # 错误处理
```

### 开发新API
1. 在 `internal/services/` 中实现业务逻辑
2. 在 `internal/handlers/` 中创建处理器
3. 在 `internal/router/` 中添加路由
4. 更新API文档和OpenAPI规范

### 测试
```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test ./internal/handlers

# 生成测试覆盖率报告
go test -cover ./...
```

---

## 📞 支持

如有问题或建议，请通过以下方式联系：

- **Issues**: GitHub Issues
- **Email**: support@lawfirm.com
- **文档**: 查看完整文档

---

**最后更新**: 2024-01-01  
**API版本**: v1.0.0  
**Go版本**: 1.23+