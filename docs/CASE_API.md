# 案件管理模块 API 文档

## 概述

案件管理模块提供完整的案件信息管理功能，包括案件的基本信息维护、状态管理、利益冲突检查、案件编号生成等功能。

## 功能特性

- ✅ 案件信息的增删改查
- ✅ 数据验证和业务规则检查
- ✅ 分页查询和条件筛选
- ✅ 案件搜索功能
- ✅ 批量删除功能
- ✅ 案件统计信息
- ✅ 案件状态管理
- ✅ 利益冲突检查
- ✅ 案件编号自动生成
- ✅ 律师分配功能
- ✅ 案件风险分类管理

## API 接口

### 1. 获取案件列表

**请求**
```http
GET /api/v1/cases?page=1&size=10&keyword=合同纠纷&case_type=CIVIL&status=in_progress&client_id=1&lawyer_id=1&project_type=诉讼&billing_method=fixed&start_date=2024-01-01&end_date=2024-12-31&sort_field=created_at&sort_order=desc
Authorization: Bearer <token>
```

**查询参数**
- `page`: 页码（默认：1）
- `size`: 每页大小（默认：10，最大：100）
- `keyword`: 搜索关键词（可选，支持案件编号、案件名称、案由、委托人信息、对方当事人信息）
- `case_type`: 案件类型（可选）
- `status`: 状态（可选：pending/in_progress/completed/cancelled）
- `client_id`: 客户ID（可选）
- `lawyer_id`: 主办律师ID（可选）
- `project_type`: 项目类型（可选）
- `billing_method`: 计费方式（可选：fixed/hourly/risk/mixed）
- `start_date`: 开始日期（可选，格式：YYYY-MM-DD）
- `end_date`: 结束日期（可选，格式：YYYY-MM-DD）
- `sort_field`: 排序字段（可选：id/case_no/case_name/created_at/contract_amount）
- `sort_order`: 排序方式（可选：asc/desc）

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "cases": [
      {
        "id": 1,
        "case_no": "CV20240101001",
        "case_name": "张三诉李四合同纠纷案",
        "case_type": "CIVIL",
        "project_type": "诉讼",
        "cause_of_action": "合同纠纷",
        "principal_info": "张三，男，35岁，北京市朝阳区",
        "opponent_info": "李四，男，40岁，北京市海淀区",
        "client_id": 1,
        "lawyer_id": 1,
        "assisting_lawyer_id": 2,
        "contract_amount": 50000.00,
        "billing_method": "fixed",
        "status": "in_progress",
        "conflict_check_status": "passed",
        "is_major_risk": false,
        "is_mass_case": false,
        "is_sensitive_case": false,
        "project_code": "PROJ2024001",
        "start_date": "2024-01-01",
        "end_date": "2024-06-30",
        "contract_document": "/uploads/contracts/contract_001.pdf",
        "legal_letter_document": "/uploads/legal_letters/letter_001.pdf",
        "remark": "重要案件，需要密切关注",
        "created_at": "2024-01-01 10:00:00",
        "updated_at": "2024-01-01 10:00:00",
        "client": {
          "id": 1,
          "client_name": "张三",
          "phone": "13800138001",
          "email": "zhangsan@example.com"
        },
        "lawyer": {
          "id": 1,
          "lawyer_name": "王律师",
          "phone": "13800138002",
          "email": "wang@lawfirm.com"
        },
        "assisting_lawyer": {
          "id": 2,
          "lawyer_name": "李律师",
          "phone": "13800138003",
          "email": "li@lawfirm.com"
        }
      }
    ],
    "total": 50,
    "page": 1,
    "size": 10
  }
}
```

### 2. 获取案件详情

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
    "case_no": "CV20240101001",
    "case_name": "张三诉李四合同纠纷案",
    "case_type": "CIVIL",
    "project_type": "诉讼",
    "cause_of_action": "合同纠纷",
    "principal_info": "张三，男，35岁，北京市朝阳区",
    "opponent_info": "李四，男，40岁，北京市海淀区",
    "client_id": 1,
    "lawyer_id": 1,
    "assisting_lawyer_id": 2,
    "contract_amount": 50000.00,
    "billing_method": "fixed",
    "status": "in_progress",
    "conflict_check_status": "passed",
    "is_major_risk": false,
    "is_mass_case": false,
    "is_sensitive_case": false,
    "project_code": "PROJ2024001",
    "start_date": "2024-01-01",
    "end_date": "2024-06-30",
    "contract_document": "/uploads/contracts/contract_001.pdf",
    "legal_letter_document": "/uploads/legal_letters/letter_001.pdf",
    "remark": "重要案件，需要密切关注",
    "created_at": "2024-01-01 10:00:00",
    "updated_at": "2024-01-01 10:00:00",
    "client": {
      "id": 1,
      "client_name": "张三",
      "phone": "13800138001",
      "email": "zhangsan@example.com"
    },
    "lawyer": {
      "id": 1,
      "lawyer_name": "王律师",
      "phone": "13800138002",
      "email": "wang@lawfirm.com"
    },
    "assisting_lawyer": {
      "id": 2,
      "lawyer_name": "李律师",
      "phone": "13800138003",
      "email": "li@lawfirm.com"
    }
  }
}
```

### 3. 创建案件

**请求**
```http
POST /api/v1/cases
Authorization: Bearer <token>
Content-Type: application/json

{
  "case_name": "张三诉李四合同纠纷案",
  "case_type": "CIVIL",
  "project_type": "诉讼",
  "cause_of_action": "合同纠纷",
  "principal_info": "张三，男，35岁，北京市朝阳区",
  "opponent_info": "李四，男，40岁，北京市海淀区",
  "client_id": 1,
  "lawyer_id": 1,
  "assisting_lawyer_id": 2,
  "contract_amount": 50000.00,
  "billing_method": "fixed",
  "status": "pending",
  "project_code": "PROJ2024001",
  "start_date": "2024-01-01",
  "end_date": "2024-06-30",
  "remark": "重要案件，需要密切关注"
}
```

**请求体参数**
- `case_name`: 案件名称（必填，2-200字符）
- `case_type`: 案件类型（必填，最多50字符）
- `project_type`: 项目类型（必填，最多50字符）
- `cause_of_action`: 案由（必填，最多200字符）
- `principal_info`: 委托人信息（必填，最多500字符）
- `opponent_info`: 对方当事人信息（必填，最多500字符）
- `client_id`: 客户ID（必填）
- `lawyer_id`: 主办律师ID（必填）
- `assisting_lawyer_id`: 协办律师ID（可选）
- `contract_amount`: 合同金额（可选，非负数）
- `billing_method`: 计费方式（可选：fixed/hourly/risk/mixed）
- `status`: 状态（可选：pending/in_progress/completed/cancelled）
- `project_code`: 项目代码（可选，最多50字符）
- `start_date`: 开始日期（可选，格式：YYYY-MM-DD）
- `end_date`: 结束日期（可选，格式：YYYY-MM-DD）
- `remark`: 备注（可选，最多255字符）

**响应**
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": 1,
    "case_no": "CV20240101001",
    "case_name": "张三诉李四合同纠纷案",
    "client_id": 1,
    "lawyer_id": 1
  }
}
```

### 4. 更新案件

**请求**
```http
PUT /api/v1/cases/1
Authorization: Bearer <token>
Content-Type: application/json

{
  "case_name": "张三诉李四合同纠纷案（更新）",
  "case_type": "CIVIL",
  "project_type": "诉讼",
  "cause_of_action": "合同纠纷",
  "principal_info": "张三，男，35岁，北京市朝阳区",
  "opponent_info": "李四，男，40岁，北京市海淀区",
  "client_id": 1,
  "lawyer_id": 1,
  "assisting_lawyer_id": 2,
  "contract_amount": 60000.00,
  "billing_method": "fixed",
  "status": "in_progress",
  "project_code": "PROJ2024001",
  "start_date": "2024-01-01",
  "end_date": "2024-06-30",
  "remark": "重要案件，需要密切关注，合同金额调整"
}
```

**响应**
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": 1,
    "case_no": "CV20240101001",
    "case_name": "张三诉李四合同纠纷案（更新）",
    "client_id": 1,
    "lawyer_id": 1
  }
}
```

### 5. 删除案件

**请求**
```http
DELETE /api/v1/cases/1
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

### 6. 获取案件统计信息

**请求**
```http
GET /api/v1/cases/stats
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total_cases": 150,
    "pending_count": 30,
    "in_progress_count": 80,
    "completed_count": 35,
    "cancelled_count": 5,
    "case_type_stats": {
      "CIVIL": 80,
      "CRIMINAL": 40,
      "ADMINISTRATIVE": 20,
      "ECONOMIC": 10
    },
    "project_type_stats": {
      "诉讼": 100,
      "非诉": 30,
      "仲裁": 15,
      "执行": 5
    },
    "total_contract_amount": 8500000.00,
    "monthly_new_cases": 12,
    "major_risk_count": 8,
    "mass_case_count": 3,
    "sensitive_case_count": 5
  }
}
```

### 7. 搜索案件

**请求**
```http
GET /api/v1/cases/search?keyword=合同纠纷
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "搜索成功",
  "data": {
    "cases": [
      {
        "id": 1,
        "case_no": "CV20240101001",
        "case_name": "张三诉李四合同纠纷案",
        "case_type": "CIVIL",
        "project_type": "诉讼",
        "cause_of_action": "合同纠纷",
        "status": "in_progress",
        "client_id": 1,
        "lawyer_id": 1,
        "created_at": "2024-01-01 10:00:00",
        "client": {
          "id": 1,
          "client_name": "张三",
          "phone": "13800138001"
        },
        "lawyer": {
          "id": 1,
          "lawyer_name": "王律师",
          "phone": "13800138002"
        }
      }
    ],
    "total": 1
  }
}
```

### 8. 获取案件下拉列表（用于选择框）

**请求**
```http
GET /api/v1/cases/select-options
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
      "case_no": "CV20240101001",
      "case_name": "张三诉李四合同纠纷案",
      "client_id": 1,
      "lawyer_id": 1,
      "status": "in_progress",
      "label": "CV20240101001 - 张三诉李四合同纠纷案 (张三)"
    },
    {
      "id": 2,
      "case_no": "CR20240102001",
      "case_name": "王五盗窃案",
      "client_id": 2,
      "lawyer_id": 2,
      "status": "pending",
      "label": "CR20240102001 - 王五盗窃案 (王五)"
    }
  ]
}
```

### 9. 批量删除案件

**请求**
```http
POST /api/v1/cases/batch-delete
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

### 10. 更新案件状态

**请求**
```http
PUT /api/v1/cases/1/status
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "completed",
  "remark": "案件已成功结案"
}
```

**响应**
```json
{
  "code": 200,
  "message": "状态更新成功",
  "data": {
    "id": 1,
    "status": "completed",
    "updated_at": "2024-06-30 15:00:00"
  }
}
```

### 11. 检查案件利益冲突

**请求**
```http
GET /api/v1/cases/1/conflict-check
Authorization: Bearer <token>
```

**响应**
```json
{
  "code": 200,
  "message": "检查完成",
  "data": {
    "case_id": 1,
    "conflict_check_status": "passed",
    "conflicts": [],
    "checked_at": "2024-01-02 10:00:00",
    "checker_id": 1,
    "checker_name": "系统管理员"
  }
}
```

### 12. 分配律师

**请求**
```http
POST /api/v1/cases/1/assign-lawyer
Authorization: Bearer <token>
Content-Type: application/json

{
  "lawyer_id": 2,
  "role": "assisting",
  "remark": "需要协办律师协助处理"
}
```

**响应**
```json
{
  "code": 200,
  "message": "分配成功",
  "data": {
    "case_id": 1,
    "lawyer_id": 2,
    "role": "assisting",
    "assigned_at": "2024-01-02 10:00:00"
  }
}
```

### 13. 生成案件编号

**请求**
```http
POST /api/v1/cases/generate-case-no
Authorization: Bearer <token>
Content-Type: application/json

{
  "case_type": "CIVIL"
}
```

**响应**
```json
{
  "code": 200,
  "message": "生成成功",
  "data": {
    "case_no": "CV20240102001",
    "case_type": "CIVIL",
    "generated_at": "2024-01-02 10:00:00"
  }
}
```

## 数据验证规则

### 案件名称
- 必填字段
- 长度：2-200 个字符

### 案件类型
- 必填字段
- 长度：最多50个字符
- 常见类型：CIVIL（民事）、CRIMINAL（刑事）、ADMINISTRATIVE（行政）、ECONOMIC（经济）

### 项目类型
- 必填字段
- 长度：最多50个字符
- 常见类型：诉讼、非诉、仲裁、执行

### 委托人信息
- 必填字段
- 长度：最多500个字符

### 案由
- 必填字段
- 长度：最多200个字符

### 对方当事人信息
- 必填字段
- 长度：最多500个字符

### 合同金额
- 可选字段
- 类型：非负数

### 计费方式
- 可选字段
- 值：fixed（固定费用）、hourly（按时计费）、risk（风险代理）、mixed（混合计费）

### 状态
- 可选字段
- 值：pending（待处理）、in_progress（进行中）、completed（已完成）、cancelled（已取消）

## 业务规则

1. **案件编号唯一性**：案件编号必须唯一，自动生成格式为：类型缩写(2位) + 日期(8位) + 序号(3位)
2. **删除限制**：进行中的案件无法删除
3. **批量操作**：批量删除最多支持100个案件
4. **状态管理**：案件状态变更遵循业务流程规则
5. **利益冲突检查**：新建案件必须进行利益冲突检查
6. **律师分配**：每个案件必须有主办律师，可以有协办律师
7. **风险分类**：支持重大风险、群体性案件、敏感案件的标识
8. **文档管理**：支持合同文档和律师函文档的路径管理

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
- `409`: 数据冲突（如案件编号重复）
- `429`: 请求过于频繁
- `500`: 服务器内部错误

## 性能优化

1. **分页查询**：所有列表接口都支持分页
2. **索引优化**：关键字段都建立了数据库索引
3. **批量操作**：支持批量删除，提高效率
4. **缓存策略**：统计信息可以考虑使用缓存
5. **搜索限制**：搜索结果限制前50条，提高响应速度

## 安全考虑

1. **认证授权**：所有接口都需要 JWT 认证
2. **数据验证**：严格的前后端数据验证
3. **敏感信息**：案件相关敏感信息适当保护
4. **访问控制**：基于角色的权限控制
5. **操作日志**：重要操作记录日志

## 相关数据模型

### 案件 (Case)
- id: 案件ID
- case_no: 案件编号
- case_name: 案件名称
- case_type: 案件类型
- project_type: 项目类型
- cause_of_action: 案由
- principal_info: 委托人信息
- opponent_info: 对方当事人信息
- client_id: 客户ID
- lawyer_id: 主办律师ID
- assisting_lawyer_id: 协办律师ID
- contract_amount: 合同金额
- billing_method: 计费方式
- status: 状态
- conflict_check_status: 利益冲突检查状态
- is_major_risk: 是否重大风险
- is_mass_case: 是否群体性案件
- is_sensitive_case: 是否敏感案件
- project_code: 项目代码
- start_date: 开始日期
- end_date: 结束日期
- contract_document: 合同文档路径
- legal_letter_document: 律师函文档路径
- remark: 备注
- created_at: 创建时间
- updated_at: 更新时间

### 关联字段
- client: 客户信息
- lawyer: 主办律师信息
- assisting_lawyer: 协办律师信息