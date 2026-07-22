# 律师事务所OA系统 API 文档

## 概述

本文档描述了律师事务所办公自动化系统的完整API接口，包括认证、案件管理、客户管理、律师管理、文件管理和冲突检测等功能模块。

### 基本信息

- **基础URL**: `http://localhost:8080/api`
- **API版本**: v1
- **数据格式**: JSON
- **字符编码**: UTF-8
- **认证方式**: Bearer Token

### 通用响应格式

所有API响应都遵循统一的格式：

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "version": "v1",
    "server": "law-oa-go",
    "environment": "development",
    "request_time": "2024-01-01T00:00:00Z",
    "duration_ms": 150
  },
  "request_id": "req_123456789",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 错误响应格式

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": "案件名称不能为空",
    "context": {
      "field_errors": {
        "title": "案件名称是必填字段"
      }
    },
    "suggestions": [
      "请检查输入数据的格式和内容",
      "确保所有必填字段都已填写"
    ]
  },
  "meta": { ... },
  "request_id": "req_123456789",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 认证模块

### 用户登录

**POST** `/auth/login`

登录获取访问令牌。

**请求参数:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "name": "张律师",
      "email": "user@example.com",
      "role": "lawyer"
    }
  }
}
```

### 用户注册

**POST** `/auth/register`

注册新用户账户。

**请求参数:**
```json
{
  "name": "张律师",
  "email": "lawyer@example.com",
  "password": "password123",
  "role": "lawyer"
}
```

### 刷新令牌

**POST** `/auth/refresh`

使用刷新令牌获取新的访问令牌。

**请求头:**
```
Authorization: Bearer <refresh_token>
```

### 用户登出

**POST** `/auth/logout`

登出当前用户，使令牌失效。

## 用户管理模块

### 获取用户资料

**GET** `/users/profile`

获取当前用户的详细信息。

**请求头:**
```
Authorization: Bearer <access_token>
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "张律师",
    "email": "lawyer@example.com",
    "role": "lawyer",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 更新用户资料

**PUT** `/users/profile`

更新当前用户的资料信息。

**请求参数:**
```json
{
  "name": "张律师",
  "email": "newemail@example.com",
  "phone": "+86 13800138000"
}
```

### 修改密码

**POST** `/users/change-password`

修改用户密码。

**请求参数:**
```json
{
  "current_password": "oldpassword123",
  "new_password": "newpassword123",
  "confirm_password": "newpassword123"
}
```

### 用户列表

**GET** `/users`

获取用户列表（需要管理员权限）。

**查询参数:**
- `page`: 页码（默认: 1）
- `page_size`: 每页数量（默认: 20）
- `role`: 按角色筛选
- `search`: 搜索关键词

**响应示例:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "张律师",
      "email": "lawyer@example.com",
      "role": "lawyer",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

## 案件管理模块

### 创建案件

**POST** `/cases`

创建新的法律案件。

**请求参数:**
```json
{
  "title": "商业合同纠纷案",
  "description": "涉及商业合同违约的案件",
  "client_id": 1,
  "lawyer_id": 1,
  "case_type": "commercial",
  "priority": "high",
  "status": "pending"
}
```

**字段说明:**
- `title`: 案件标题（必填，最大200字符）
- `description`: 案件描述（可选）
- `client_id`: 客户ID（必填）
- `lawyer_id`: 律师ID（必填）
- `case_type`: 案件类型（必填，可选值: civil, criminal, commercial, administrative）
- `priority`: 优先级（必填，可选值: low, medium, high, urgent）
- `status`: 状态（必填，可选值: pending, active, closed, suspended）

### 获取案件列表

**GET** `/cases`

获取案件列表，支持分页和筛选。

**查询参数:**
- `page`: 页码（默认: 1）
- `page_size`: 每页数量（默认: 20）
- `status`: 按状态筛选
- `case_type`: 按案件类型筛选
- `priority`: 按优先级筛选
- `client_id`: 按客户ID筛选
- `lawyer_id`: 按律师ID筛选
- `search`: 搜索关键词（支持标题、描述、ID）

**响应示例:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "商业合同纠纷案",
      "description": "涉及商业合同违约的案件",
      "case_type": "commercial",
      "status": "active",
      "priority": "high",
      "client_id": 1,
      "lawyer_id": 1,
      "client_name": "测试公司",
      "client_company": "测试公司",
      "lawyer_name": "张律师",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": { ... }
}
```

### 获取案件详情

**GET** `/cases/{id}`

获取指定案件的详细信息。

**路径参数:**
- `id`: 案件ID

**响应示例:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "商业合同纠纷案",
    "description": "涉及商业合同违约的案件",
    "case_type": "commercial",
    "status": "active",
    "priority": "high",
    "client_id": 1,
    "lawyer_id": 1,
    "client": {
      "id": 1,
      "name": "测试公司",
      "company": "测试公司",
      "email": "client@example.com"
    },
    "lawyer": {
      "id": 1,
      "name": "张律师",
      "email": "lawyer@example.com"
    },
    "document_count": 5,
    "last_activity_at": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 更新案件

**PUT** `/cases/{id}`

更新指定案件的信息。

**路径参数:**
- `id`: 案件ID

**请求参数:**
```json
{
  "title": "更新后的案件标题",
  "description": "更新后的案件描述",
  "case_type": "civil",
  "priority": "medium",
  "status": "active"
}
```

### 删除案件

**DELETE** `/cases/{id}`

删除指定的案件。

**路径参数:**
- `id`: 案件ID

### 获取案件统计

**GET** `/cases/stats`

获取案件相关的统计数据。

**响应示例:**
```json
{
  "success": true,
  "data": {
    "total_cases": 150,
    "active_cases": 45,
    "pending_cases": 12,
    "closed_cases": 93,
    "cases_by_type": {
      "civil": 60,
      "criminal": 30,
      "commercial": 45,
      "administrative": 15
    },
    "cases_by_priority": {
      "low": 20,
      "medium": 80,
      "high": 40,
      "urgent": 10
    }
  }
}
```

## 客户管理模块

### 创建客户

**POST** `/clients`

创建新的客户。

**请求参数:**
```json
{
  "name": "张三",
  "company": "测试公司",
  "email": "client@example.com",
  "phone": "+86 13800138000",
  "address": "北京市朝阳区",
  "contact_person": "李四",
  "description": "测试客户描述"
}
```

### 获取客户列表

**GET** `/clients`

获取客户列表。

**查询参数:**
- `page`: 页码（默认: 1）
- `page_size`: 每页数量（默认: 20）
- `search`: 搜索关键词

### 获取客户详情

**GET** `/clients/{id}`

获取指定客户的详细信息。

### 更新客户

**PUT** `/clients/{id}`

更新指定客户的信息。

### 删除客户

**DELETE** `/clients/{id}`

删除指定的客户。

### 获取客户统计

**GET** `/clients/stats`

获取客户相关的统计数据。

## 律师管理模块

### 获取律师列表

**GET** `/lawfirm/lawyers`

获取律师列表。

**查询参数:**
- `page`: 页码（默认: 1）
- `page_size`: 每页数量（默认: 20）
- `specialization`: 按专业领域筛选
- `search`: 搜索关键词

**响应示例:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "张律师",
      "email": "lawyer@example.com",
      "phone": "+86 13800138000",
      "specialization": "商法",
      "experience_years": 10,
      "status": "active",
      "case_count": 25
    }
  ],
  "pagination": { ... }
}
```

### 获取律师详情

**GET** `/lawfirm/lawyers/{id}`

获取指定律师的详细信息。

### 获取律师统计

**GET** `/lawfirm/lawyers/stats`

获取律师相关的统计数据。

## 冲突检测模块

### 执行冲突检测

**POST** `/conflict/check`

同步执行冲突检测。仅建议用于诊断；律师端默认使用下述异步任务接口。

### 创建异步冲突检测任务

**POST** `/conflict/tasks`

请求体与 `/conflict/check` 相同。接口立即返回 `taskId`、`status` 和
`recommendedPollingInterval`。任务状态为 `QUEUED`、`RUNNING`、`COMPLETED` 或 `FAILED`。

**GET** `/conflict/tasks/{taskId}` 获取进度。

**GET** `/conflict/tasks/{taskId}/result` 获取冻结结果。普通律师只能读取本人任务；合规、
风控、主任、合伙人和管理员可以按管理权限读取。

**请求参数:**
```json
{
  "clientId": "1",
  "clientName": "测试客户",
  "caseName": "测试案件",
  "caseType": "commercial",
  "clientType": "COMPANY",
  "otherParties": ["对方公司"],
  "searchYears": 5,
  "includeCorporateRelations": true,
  "searchDepth": "deep",
  "userId": "1",
  "requestTime": "2024-01-01T00:00:00Z"
}
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "checkId": "CC_1_1704067200",
    "hasConflict": false,
    "conflictCases": [],
    "checkStatistics": {
      "totalCasesChecked": 100,
      "clientHistoryCases": 5,
      "relatedPartiesChecked": 2,
      "corporateRelationsChecked": 10,
      "timeRange": "5年",
      "searchScope": "deep",
      "startTime": "2024-01-01T00:00:00Z",
      "endTime": "2024-01-01T00:01:00Z"
    },
    "riskAssessment": {
      "overallRisk": "LOW",
      "riskScore": 15,
      "riskReason": "未发现明显的利益冲突风险",
      "requiresApproval": false,
      "riskFactors": [],
      "mitigation": ["建议在案件进行过程中持续监控潜在冲突"]
    },
    "decision": {
      "status": "CLEAR",
      "recommendation": "未发现可识别的冲突线索，可继续进入人工确认环节。",
      "requiresManualReview": false,
      "evidenceCount": 0,
      "coverageNotice": "检索范围以本所已录入并获授权的数据为限。"
    },
    "normalizedSubjects": [],
    "recommendations": [
      "未发现明显的利益冲突",
      "建议在案件进行过程中持续监控",
      "如发现新的相关方，请及时进行补充检查"
    ],
    "checkTime": "2024-01-01 12:00:00",
    "duration": 1200
  }
}
```

`decision.status` 是接案门禁的权威状态：

- `CLEAR`：未发现可识别线索；
- `REVIEW_REQUIRED`：候选、关系或文本线索待人工复核；
- `BLOCKED`：确认直接冲突，暂停接案；
- `WAIVER_PENDING`：豁免正在独立复核，仍暂停接案；
- `WAIVED`：豁免已批准，可按批准条件继续。

`hasConflict=true` 仅表示系统已经确认直接冲突；名称候选和关系线索使用
`decision.status=REVIEW_REQUIRED`，不得伪装成已确认冲突。

### 人工复核

**POST** `/conflict/tasks/{taskId}/review`

```json
{
  "decision": "confirmed_conflict",
  "notes": "已核对客户主档案、历史事项和主体标识。"
}
```

复核记录不可变。可选结论为 `no_conflict`、`confirmed_conflict`、`false_positive`、
`insufficient_information` 和 `waiver_requested`。

**GET** `/conflict/tasks/{taskId}/review` 获取最新复核记录。

### 豁免评估

**POST** `/conflict/tasks/{taskId}/waiver`

```json
{
  "rationale": "客户已充分知情，拟通过独立团队和信息隔离控制剩余风险。",
  "waiver_type": "INFORMED_CONSENT",
  "waiver_category": "CLIENT_CONSENT",
  "proposed_conditions": ["建立信息隔离墙", "限制敏感资料访问", "定期合规复核"],
  "duration_days": 180
}
```

系统自动选择与申请人不同的复核人。申请人不得自批。创建后冲突状态变为
`WAIVER_PENDING`。

**GET** `/conflict/tasks/{taskId}/waiver` 获取该任务当前豁免申请。

**POST** `/waivers/{waiverId}/decision` 由指定复核人或管理角色作出决定；批准后冲突状态
反写为 `WAIVED`，拒绝后反写为 `BLOCKED`。

### 获取检查历史

**GET** `/conflict/history/{clientId}`

获取指定客户的冲突检查历史记录。

**路径参数:**
- `clientId`: 客户ID

**查询参数:**
- `limit`: 返回记录数量限制（默认: 10）

### 获取检查详情

**GET** `/conflict/details/{checkId}`

获取指定冲突检查的详细信息。

**路径参数:**
- `checkId`: 检查ID

### 获取冲突规则

**GET** `/conflict/rules`

获取所有的利益冲突检测规则。

### 获取MCP标准

**GET** `/conflict/standards`

获取最新的MCP利益冲突检测标准。

### 健康检查

**GET** `/conflict/health`

检查冲突检测服务的运行状态。

**响应示例:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "service": "conflict-detection",
    "timestamp": "2024-01-01T00:00:00Z",
    "checks": {
      "service_initialized": true,
      "mcp_standards": {
        "status": "ok"
      },
      "conflict_rules": {
        "status": "ok"
      }
    },
    "duration": 150
  }
}
```

## 文件管理模块

### 上传文件

**POST** `/documents`

上传文件到系统。

**请求参数:**
- `file`: 文件（multipart/form-data）
- `case_id`: 关联的案件ID（可选）
- `description`: 文件描述（可选）

### 获取文件列表

**GET** `/files`

获取文件列表。

**查询参数:**
- `page`: 页码（默认: 1）
- `page_size`: 每页数量（默认: 20）
- `case_id`: 按案件ID筛选
- `search`: 搜索关键词

### 获取文件详情

**GET** `/documents/{id}`

获取指定文件的详细信息。

### 下载文件

**GET** `/documents/{id}/download`

下载指定的文件。

### 删除文件

**DELETE** `/documents/{id}`

删除指定的文件。

## 仪表盘模块

### 获取统计信息

**GET** `/dashboard/statistics`

获取仪表盘的统计信息。

**响应示例:**
```json
{
  "success": true,
  "data": {
    "total_cases": 150,
    "active_cases": 45,
    "total_clients": 80,
    "total_lawyers": 12,
    "recent_activities": [
      {
        "id": 1,
        "type": "case_created",
        "description": "创建新案件",
        "timestamp": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 获取待办事项

**GET** `/dashboard/todos`

获取当前用户的待办事项。

### 获取活动日志

**GET** `/dashboard/activities`

获取活动日志列表。

## 错误代码说明

| 错误代码 | HTTP状态码 | 说明 |
|---------|-----------|------|
| `VALIDATION_ERROR` | 400 | 请求参数验证失败 |
| `UNAUTHORIZED` | 401 | 未授权访问 |
| `FORBIDDEN` | 403 | 权限不足 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `CONFLICT` | 409 | 资源冲突 |
| `RATE_LIMIT_EXCEEDED` | 429 | 请求频率超限 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |
| `SERVICE_UNAVAILABLE` | 503 | 服务不可用 |

## 限流说明

为了保护系统稳定性，API实施了限流策略：

- **普通用户**: 每分钟100次请求
- **认证相关**: 每分钟5次请求
- **文件上传**: 每分钟10次请求

超出限制时将返回 `429` 状态码。

## 缓存策略

系统对以下请求实施缓存策略：

- **案件列表**: 缓存5分钟
- **统计信息**: 缓存10分钟
- **用户信息**: 缓存30分钟

缓存键基于请求URL和用户ID生成。

## SDK和工具

### JavaScript SDK

```javascript
// 安装
npm install law-oa-sdk

// 使用
import { LawOA } from 'law-oa-sdk';

const client = new LawOA({
  baseURL: 'http://localhost:8080/api',
  token: 'your-access-token'
});

// 创建案件
const caseData = await client.cases.create({
  title: '测试案件',
  client_id: 1,
  lawyer_id: 1,
  case_type: 'civil'
});
```

### Python SDK

```python
# 安装
pip install lawoa-sdk

# 使用
from lawoa import LawOAClient

client = LawOAClient(
    base_url='http://localhost:8080/api',
    token='your-access-token'
)

# 创建案件
case_data = client.cases.create(
    title='测试案件',
    client_id=1,
    lawyer_id=1,
    case_type='civil'
)
```

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 实现基础的案件管理功能
- 添加冲突检测模块
- 完善用户认证系统

### v1.1.0 (2024-01-15)
- 优化案件筛选功能
- 修复案件创建422错误
- 增强错误处理机制

### v1.2.0 (2024-02-01)
- 添加性能缓存策略
- 优化查询性能
- 完善API文档

## 支持

如有任何问题或建议，请联系技术支持团队：

- 邮箱: support@law-oa.com
- 电话: 400-888-8888
- 在线文档: https://docs.law-oa.com
- GitHub: https://github.com/law-oa/api
