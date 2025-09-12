# Law OA Go API 文档

## 概述

Law OA Go 提供完整的 RESTful API，支持案件管理、客户管理、律师管理、利益冲突检查等核心功能。

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: Bearer Token (JWT)
- **数据格式**: JSON

## 认证

### 登录

**请求**
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

**响应**
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@lawfirm.com",
      "real_name": "系统管理员"
    }
  }
}
```

### 注册

**请求**
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "newuser",
  "password": "password",
  "email": "newuser@example.com",
  "real_name": "新用户"
}
```

**响应**
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 2,
    "username": "newuser"
  }
}
```

## 用户管理

### 获取用户列表

**请求**
```http
GET /api/v1/users?page=1&size=10&keyword=admin
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "users": [
      {
        "id": 1,
        "username": "admin",
        "email": "admin@lawfirm.com",
        "real_name": "系统管理员",
        "status": "active",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "size": 10
  }
}
```

### 获取用户详情

**请求**
```http
GET /api/v1/users/1
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@lawfirm.com",
    "real_name": "系统管理员",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 创建用户

**请求**
```http
POST /api/v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newuser",
  "password": "password",
  "email": "newuser@example.com",
  "real_name": "新用户",
  "status": "active"
}
```

**响应**
```json
{
  "code": 200,
  "message": "创建成功",
  "data": {
    "id": 2,
    "username": "newuser"
  }
}
```

### 更新用户

**请求**
```http
PUT /api/v1/users/2
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "updated@example.com",
  "real_name": "更新用户",
  "status": "active"
}
```

**响应**
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": 2,
    "email": "updated@example.com",
    "real_name": "更新用户"
  }
}
```

### 删除用户

**请求**
```http
DELETE /api/v1/users/2
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

## 案件管理

### 获取案件列表

**请求**
```http
GET /api/v1/cases?page=1&size=10&case_name=合同&case_type=CIVIL&status=in_progress
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "cases": [
      {
        "id": 1,
        "case_no": "CASE2024001",
        "case_name": "张三诉李四合同纠纷",
        "case_type": "CIVIL",
        "project_type": "CASE",
        "status": "in_progress",
        "principal_info": "张三，联系电话：13900139001",
        "opponent_info": "李四，联系电话：13900139002",
        "cause_of_action": "合同纠纷",
        "description": "合同纠纷案件",
        "contract_amount": 50000.00,
        "billing_method": "FIXED",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "size": 10
  }
}
```

### 获取案件详情

**请求**
```http
GET /api/v1/cases/1
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "id": 1,
    "case_no": "CASE2024001",
    "case_name": "张三诉李四合同纠纷",
    "case_type": "CIVIL",
    "project_type": "CASE",
    "client_id": 1,
    "lawyer_id": 1,
    "assisting_lawyer_id": 2,
    "status": "in_progress",
    "principal_info": "张三，联系电话：13900139001，地址：北京市朝阳区",
    "opponent_info": "李四，联系电话：13900139002，地址：北京市海淀区",
    "cause_of_action": "合同纠纷",
    "description": "合同纠纷案件",
    "contract_amount": 50000.00,
    "billing_method": "FIXED",
    "conflict_check_status": "passed",
    "is_major_risk": false,
    "is_mass_case": false,
    "is_sensitive_case": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 创建案件

**请求**
```http
POST /api/v1/cases
Authorization: Bearer <token>
Content-Type: application/json

{
  "case_no": "CASE2024002",
  "case_name": "某科技公司商标侵权案",
  "case_type": "COMMERCIAL",
  "project_type": "CASE",
  "client_id": 1,
  "lawyer_id": 1,
  "assisting_lawyer_id": 2,
  "principal_info": "某科技有限公司，联系人：王经理，电话：010-12345678",
  "opponent_info": "某侵权公司",
  "cause_of_action": "商标侵权",
  "description": "商标侵权案件",
  "contract_amount": 100000.00,
  "billing_method": "RISK",
  "conflict_check_status": "pending",
  "is_major_risk": true,
  "is_mass_case": false,
  "is_sensitive_case": false
}
```

**响应**
```json
{
  "code": 200,
  "message": "创建成功",
  "data": {
    "id": 2,
    "case_no": "CASE2024002",
    "case_name": "某科技公司商标侵权案"
  }
}
```

### 更新案件

**请求**
```http
PUT /api/v1/cases/2
Authorization: Bearer <token>
Content-Type: application/json

{
  "case_name": "某科技公司商标侵权案（更新）",
  "status": "in_progress",
  "contract_amount": 120000.00,
  "conflict_check_status": "passed"
}
```

**响应**
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": 2,
    "case_name": "某科技公司商标侵权案（更新）",
    "status": "in_progress"
  }
}
```

### 删除案件

**请求**
```http
DELETE /api/v1/cases/2
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

## 统计信息

### 获取仪表板统计

**请求**
```http
GET /api/v1/stats/dashboard
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total_cases": 150,
    "active_cases": 45,
    "total_clients": 120,
    "total_lawyers": 8,
    "cases_by_type": {
      "CIVIL": 80,
      "COMMERCIAL": 45,
      "CRIMINAL": 20,
      "ADMINISTRATIVE": 5
    },
    "cases_by_status": {
      "pending": 30,
      "in_progress": 45,
      "completed": 60,
      "closed": 15
    },
    "recent_cases": [
      {
        "id": 1,
        "case_no": "CASE2024001",
        "case_name": "张三诉李四合同纠纷",
        "status": "in_progress",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

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
- `429`: 请求过于频繁
- `500`: 服务器内部错误

## 数据模型

### 用户 (User)

```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@lawfirm.com",
  "real_name": "系统管理员",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 案件 (Case)

```json
{
  "id": 1,
  "case_no": "CASE2024001",
  "case_name": "案件名称",
  "case_type": "CIVIL",
  "project_type": "CASE",
  "client_id": 1,
  "lawyer_id": 1,
  "assisting_lawyer_id": 2,
  "status": "in_progress",
  "principal_info": "委托人信息",
  "opponent_info": "对方当事人信息",
  "cause_of_action": "案由",
  "description": "案件描述",
  "contract_amount": 50000.00,
  "billing_method": "FIXED",
  "conflict_check_status": "passed",
  "is_major_risk": false,
  "is_mass_case": false,
  "is_sensitive_case": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## 请求限制

- API 请求频率限制：每分钟 100 次
- 文件上传大小限制：50MB
- JWT 令牌有效期：24 小时

## 认证说明

所有需要认证的 API 都需要在请求头中包含 JWT 令牌：

```http
Authorization: Bearer <your-jwt-token>
```

令牌可以通过登录接口获取，有效期 24 小时。