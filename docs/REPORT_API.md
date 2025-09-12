# 报表统计模块 API 文档

## 概述

报表统计模块提供了律师事务所管理系统的完整数据分析和可视化功能，包括仪表板统计、趋势分析、工作负载统计和综合报表等功能。

## 功能特性

- 📊 仪表板数据统计
- 📈 趋势分析（案件、收入）
- 👥 律师工作负载分析
- 📋 案件类型分布统计
- 🎯 客户分析数据
- 📄 综合报表生成
- 🔄 数据导出功能

## API 接口

### 1. 获取仪表板统计数据

**接口地址**: `GET /api/v1/reports/dashboard`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |
| lawyer_id | number | 否 | 律师ID过滤 |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "total_cases": 156,
    "active_cases": 45,
    "completed_cases": 98,
    "pending_cases": 13,
    "cases_this_month": 12,
    "cases_growth_rate": 15.2,
    "total_clients": 89,
    "active_clients": 67,
    "new_clients_month": 5,
    "clients_growth_rate": 8.5,
    "total_lawyers": 12,
    "active_lawyers": 10,
    "busy_lawyers": 3,
    "total_documents": 1245,
    "documents_this_month": 89,
    "storage_used": 2147483648,
    "storage_total": 5368709120,
    "total_revenue": 1580000.00,
    "revenue_this_month": 125000.00,
    "revenue_growth_rate": 12.8,
    "total_conflict_checks": 245,
    "conflict_checks_this_month": 18,
    "high_risk_conflicts": 3,
    "last_updated": "2024-09-11T10:30:00Z"
  }
}
```

### 2. 获取案件趋势数据

**接口地址**: `GET /api/v1/reports/case-trend`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |
| group_by | string | 否 | 分组方式 (day/week/month/year) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "period": "2024-01",
      "total_cases": 15,
      "completed_cases": 12,
      "active_cases": 3,
      "growth_rate": 8.5
    },
    {
      "period": "2024-02",
      "total_cases": 18,
      "completed_cases": 14,
      "active_cases": 4,
      "growth_rate": 12.2
    }
  ]
}
```

### 3. 获取收入趋势数据

**接口地址**: `GET /api/v1/reports/revenue-trend`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |
| group_by | string | 否 | 分组方式 (day/week/month/year) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "period": "2024-01",
      "total_revenue": 150000.00,
      "case_count": 15,
      "avg_case_value": 10000.00,
      "growth_rate": 10.5
    },
    {
      "period": "2024-02",
      "total_revenue": 180000.00,
      "case_count": 18,
      "avg_case_value": 10000.00,
      "growth_rate": 15.2
    }
  ]
}
```

### 4. 获取律师工作量统计

**接口地址**: `GET /api/v1/reports/lawyer-workload`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "lawyer_id": 1,
      "lawyer_name": "张律师",
      "total_cases": 25,
      "active_cases": 8,
      "completed_cases": 17,
      "total_revenue": 250000.00,
      "avg_case_value": 10000.00,
      "workload_score": 75.5,
      "status": "busy",
      "efficiency_rate": 85.2
    },
    {
      "lawyer_id": 2,
      "lawyer_name": "李律师",
      "total_cases": 18,
      "active_cases": 3,
      "completed_cases": 15,
      "total_revenue": 180000.00,
      "avg_case_value": 10000.00,
      "workload_score": 45.2,
      "status": "normal",
      "efficiency_rate": 92.1
    }
  ]
}
```

### 5. 获取案件类型分布

**接口地址**: `GET /api/v1/reports/case-type-distribution`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "case_type": "合同纠纷",
      "count": 45,
      "percentage": 28.8,
      "total_revenue": 450000.00,
      "avg_case_value": 10000.00
    },
    {
      "case_type": "劳动争议",
      "count": 32,
      "percentage": 20.5,
      "total_revenue": 320000.00,
      "avg_case_value": 10000.00
    },
    {
      "case_type": "婚姻家庭",
      "count": 28,
      "percentage": 17.9,
      "total_revenue": 280000.00,
      "avg_case_value": 10000.00
    }
  ]
}
```

### 6. 获取客户分析数据

**接口地址**: `GET /api/v1/reports/client-analysis`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "client_id": 1,
      "client_name": "ABC公司",
      "client_type": "company",
      "total_cases": 8,
      "active_cases": 2,
      "total_amount": 80000.00,
      "avg_case_value": 10000.00,
      "first_case_date": "2023-06-15T10:30:00Z",
      "last_case_date": "2024-09-10T14:20:00Z",
      "retention_rate": 85.5
    },
    {
      "client_id": 2,
      "client_name": "张三",
      "client_type": "individual",
      "total_cases": 3,
      "active_cases": 1,
      "total_amount": 30000.00,
      "avg_case_value": 10000.00,
      "first_case_date": "2024-01-20T09:15:00Z",
      "last_case_date": "2024-08-25T11:30:00Z",
      "retention_rate": 75.2
    }
  ]
}
```

### 7. 获取律师绩效统计

**接口地址**: `GET /api/v1/reports/lawyer-performance`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| lawyer_id | number | 是 | 律师ID |
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "lawyer_id": 1,
    "lawyer_name": "张律师",
    "total_cases": 25,
    "active_cases": 8,
    "completed_cases": 17,
    "total_revenue": 250000.00,
    "avg_case_value": 10000.00,
    "workload_score": 75.5,
    "status": "busy"
  }
}
```

### 8. 获取案件统计报表

**接口地址**: `GET /api/v1/reports/cases`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "summary": {
      "total_cases": 156,
      "active_cases": 45,
      "completed_cases": 98,
      "pending_cases": 13,
      "cases_this_month": 12,
      "growth_rate": 15.2
    },
    "trend_data": [
      {
        "period": "2024-01",
        "total_cases": 15,
        "completed_cases": 12,
        "active_cases": 3,
        "growth_rate": 8.5
      }
    ],
    "type_distribution": [
      {
        "case_type": "合同纠纷",
        "count": 45,
        "percentage": 28.8,
        "total_revenue": 450000.00,
        "avg_case_value": 10000.00
      }
    ],
    "last_updated": "2024-09-11T10:30:00Z"
  }
}
```

### 9. 获取财务统计报表

**接口地址**: `GET /api/v1/reports/financial`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "summary": {
      "total_revenue": 1580000.00,
      "revenue_this_month": 125000.00,
      "revenue_growth_rate": 12.8,
      "total_cases": 156
    },
    "revenue_trend": [
      {
        "period": "2024-01",
        "total_revenue": 150000.00,
        "case_count": 15,
        "avg_case_value": 10000.00,
        "growth_rate": 10.5
      }
    ],
    "revenue_by_type": [
      {
        "case_type": "合同纠纷",
        "count": 45,
        "percentage": 28.8,
        "total_revenue": 450000.00,
        "avg_case_value": 10000.00
      }
    ],
    "last_updated": "2024-09-11T10:30:00Z"
  }
}
```

### 10. 获取客户统计报表

**接口地址**: `GET /api/v1/reports/clients`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "summary": {
      "total_clients": 89,
      "active_clients": 67,
      "new_clients_this_month": 5,
      "clients_growth_rate": 8.5
    },
    "client_analysis": [
      {
        "client_id": 1,
        "client_name": "ABC公司",
        "client_type": "company",
        "total_cases": 8,
        "active_cases": 2,
        "total_amount": 80000.00,
        "avg_case_value": 10000.00,
        "retention_rate": 85.5
      }
    ],
    "client_type_dist": {
      "company": {
        "count": 45,
        "total_amount": 900000.00,
        "avg_case_count": 20.0
      },
      "individual": {
        "count": 44,
        "total_amount": 680000.00,
        "avg_case_count": 15.5
      }
    },
    "last_updated": "2024-09-11T10:30:00Z"
  }
}
```

### 11. 获取律师统计报表

**接口地址**: `GET /api/v1/reports/lawyers`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "summary": {
      "total_lawyers": 12,
      "active_lawyers": 10,
      "busy_lawyers": 3
    },
    "workload_stats": [
      {
        "lawyer_id": 1,
        "lawyer_name": "张律师",
        "total_cases": 25,
        "active_cases": 8,
        "completed_cases": 17,
        "total_revenue": 250000.00,
        "avg_case_value": 10000.00,
        "workload_score": 75.5,
        "status": "busy"
      }
    ],
    "workload_dist": {
      "idle": 3,
      "normal": 4,
      "busy": 3,
      "overloaded": 0
    },
    "last_updated": "2024-09-11T10:30:00Z"
  }
}
```

### 12. 获取综合统计报表

**接口地址**: `GET /api/v1/reports/comprehensive`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "dashboard": {
      "cases": {
        "total": 156,
        "active": 45,
        "completed": 98,
        "pending": 13,
        "this_month": 12,
        "growth_rate": 15.2
      },
      "clients": {
        "total": 89,
        "active": 67,
        "new_this_month": 5,
        "growth_rate": 8.5
      },
      "lawyers": {
        "total": 12,
        "active": 10,
        "busy": 3
      },
      "documents": {
        "total": 1245,
        "this_month": 89,
        "storage_used": 2147483648,
        "storage_total": 5368709120
      },
      "revenue": {
        "total": 1580000.00,
        "this_month": 125000.00,
        "growth_rate": 12.8
      },
      "conflict_checks": {
        "total": 245,
        "this_month": 18,
        "high_risk": 3
      }
    },
    "trends": {
      "cases": [
        {
          "period": "2024-01",
          "total_cases": 15,
          "completed_cases": 12,
          "active_cases": 3,
          "growth_rate": 8.5
        }
      ],
      "revenue": [
        {
          "period": "2024-01",
          "total_revenue": 150000.00,
          "case_count": 15,
          "avg_case_value": 10000.00,
          "growth_rate": 10.5
        }
      ]
    },
    "analysis": {
      "workload_stats": [
        {
          "lawyer_id": 1,
          "lawyer_name": "张律师",
          "total_cases": 25,
          "active_cases": 8,
          "completed_cases": 17,
          "total_revenue": 250000.00,
          "avg_case_value": 10000.00,
          "workload_score": 75.5,
          "status": "busy"
        }
      ],
      "case_type_dist": [
        {
          "case_type": "合同纠纷",
          "count": 45,
          "percentage": 28.8,
          "total_revenue": 450000.00,
          "avg_case_value": 10000.00
        }
      ],
      "client_analysis": [
        {
          "client_id": 1,
          "client_name": "ABC公司",
          "client_type": "company",
          "total_cases": 8,
          "active_cases": 2,
          "total_amount": 80000.00,
          "avg_case_value": 10000.00,
          "retention_rate": 85.5
        }
      ]
    },
    "last_updated": "2024-09-11T10:30:00Z"
  }
}
```

### 13. 导出报表数据

**接口地址**: `GET /api/v1/reports/export`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type | string | 否 | 报表类型 (cases/financial/clients/lawyers/comprehensive) |
| format | string | 否 | 导出格式 (json/csv) |
| start_date | string | 否 | 开始日期 (YYYY-MM-DD) |
| end_date | string | 否 | 结束日期 (YYYY-MM-DD) |

**响应示例**:
```json
{
  "code": 200,
  "message": "导出成功",
  "data": {
    "report_type": "comprehensive",
    "format": "json",
    "generated_at": "2024-09-11T10:30:00Z",
    "data": {
      "summary": {
        "total_cases": 156,
        "active_cases": 45,
        "completed_cases": 98,
        "pending_cases": 13
      }
    }
  }
}
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 无权访问 |
| 500 | 服务器内部错误 |

## 数据统计逻辑

### 1. 仪表板统计
- 统计指定时间范围内的核心业务指标
- 支持按律师维度过滤数据
- 自动计算增长率和百分比

### 2. 趋势分析
- 支持按日、周、月、年分组统计
- 计算环比增长率
- 支持多维度对比分析

### 3. 工作负载统计
- 基于案件数量和收入计算工作量得分
- 根据工作量自动分类律师状态
- 计算工作效率指标

### 4. 客户分析
- 按客户类型进行分组统计
- 计算客户留存率
- 分析客户价值贡献

## 使用示例

### 获取仪表板数据
```bash
curl -X GET "http://localhost:8080/api/v1/reports/dashboard?start_date=2024-01-01&end_date=2024-12-31" \
  -H "Authorization: Bearer your-token"
```

### 获取案件趋势
```bash
curl -X GET "http://localhost:8080/api/v1/reports/case-trend?group_by=month" \
  -H "Authorization: Bearer your-token"
```

### 导出综合报表
```bash
curl -X GET "http://localhost:8080/api/v1/reports/export?type=comprehensive&format=json" \
  -H "Authorization: Bearer your-token"
```

## 注意事项

1. 所有报表接口都需要JWT认证
2. 日期格式必须为YYYY-MM-DD
3. 默认统计最近30天的数据
4. 分组方式默认为month
5. 导出CSV格式功能正在开发中