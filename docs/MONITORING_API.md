# 监控管理 API 文档

## 1. 系统监控

### 1.1 获取系统指标
- **URL**: `/api/v1/monitoring/metrics`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取当前系统性能指标

#### 响应示例
```json
{
  "success": true,
  "data": {
    "timestamp": "2024-01-01T10:00:00Z",
    "cpu_usage": 45.2,
    "cpu_count": 8,
    "cpu_freq": 2400.0,
    "memory_total": 17179869184,
    "memory_used": 8589934592,
    "memory_usage": 50.0,
    "disk_total": 536870912000,
    "disk_used": 268435456000,
    "disk_usage": 50.0,
    "disk_io_read": 1024000,
    "disk_io_write": 512000,
    "net_bytes_sent": 1048576,
    "net_bytes_recv": 2097152,
    "goroutines": 25,
    "gc_gc": 100,
    "gc_pause": 5000000,
    "request_count": 1000,
    "error_count": 10,
    "avg_response_time": 0.125,
    "db_connections": 5,
    "db_slow_queries": 2,
    "redis_used_memory": 16777216,
    "redis_connections": 3,
    "redis_keyspace_hits": 500,
    "redis_keyspace_misses": 50
  }
}
```

### 1.2 获取历史指标
- **URL**: `/api/v1/monitoring/metrics/history`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取指定时间范围内的系统历史指标

#### 查询参数
- `start` (string): 开始时间 (RFC3339格式)
- `end` (string): 结束时间 (RFC3339格式)

#### 响应示例
```json
{
  "success": true,
  "data": [
    {
      "timestamp": "2024-01-01T09:00:00Z",
      "cpu_usage": 40.2,
      "memory_usage": 48.5,
      "disk_usage": 49.8
    },
    {
      "timestamp": "2024-01-01T09:30:00Z",
      "cpu_usage": 42.1,
      "memory_usage": 49.2,
      "disk_usage": 49.9
    }
  ]
}
```

### 1.3 获取仪表板统计
- **URL**: `/api/v1/monitoring/dashboard`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取系统运行状态统计信息

#### 响应示例
```json
{
  "success": true,
  "data": {
    "system_metrics": {
      "cpu_usage": 45.2,
      "memory_usage": 50.0,
      "disk_usage": 50.0,
      "goroutines": 25,
      "request_count": 1000,
      "error_count": 10
    },
    "alerts": [
      {
        "id": "alert_123",
        "rule_name": "CPU使用率过高",
        "level": "warning",
        "message": "CPU使用率: 85.2 > 80.0",
        "value": 85.2,
        "threshold": 80.0,
        "started_at": "2024-01-01T09:45:00Z"
      }
    ],
    "log_stats": {
      "total_count": 10000,
      "level_counts": {
        "0": 5000,
        "1": 3000,
        "2": 1500,
        "3": 500
      },
      "hourly_stats": {
        "2024-01-01 09:00:00": 1000,
        "2024-01-01 10:00:00": 1200
      }
    },
    "uptime": "24h0m0s",
    "version": "1.0.0",
    "environment": "development"
  }
}
```

## 2. 告警管理

### 2.1 获取活跃告警
- **URL**: `/api/v1/monitoring/alerts`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取当前活跃的告警信息

#### 响应示例
```json
{
  "success": true,
  "data": [
    {
      "id": "alert_123",
      "rule_id": "cpu_high",
      "rule_name": "CPU使用率过高",
      "level": "warning",
      "message": "CPU使用率: 85.2 > 80.0",
      "value": 85.2,
      "threshold": 80.0,
      "status": "active",
      "started_at": "2024-01-01T09:45:00Z"
    }
  ]
}
```

### 2.2 获取告警历史
- **URL**: `/api/v1/monitoring/alerts/history`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取告警历史记录

#### 查询参数
- `limit` (int): 限制数量，默认100

#### 响应示例
```json
{
  "success": true,
  "data": [
    {
      "id": "alert_122",
      "rule_id": "memory_high",
      "rule_name": "内存使用率过高",
      "level": "warning",
      "message": "内存使用率: 88.5 > 85.0",
      "value": 88.5,
      "threshold": 85.0,
      "status": "resolved",
      "started_at": "2024-01-01T08:30:00Z",
      "resolved_at": "2024-01-01T09:00:00Z"
    }
  ]
}
```

### 2.3 获取告警规则
- **URL**: `/api/v1/monitoring/alerts/rules`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取系统告警规则配置

#### 响应示例
```json
{
  "success": true,
  "data": [
    {
      "id": "cpu_high",
      "name": "CPU使用率过高",
      "metric": "cpu_usage",
      "operator": ">",
      "threshold": 80.0,
      "duration": 300,
      "level": "warning",
      "enabled": true
    },
    {
      "id": "memory_high",
      "name": "内存使用率过高",
      "metric": "memory_usage",
      "operator": ">",
      "threshold": 85.0,
      "duration": 300,
      "level": "warning",
      "enabled": true
    }
  ]
}
```

### 2.4 更新告警规则
- **URL**: `/api/v1/monitoring/alerts/rules/{id}`
- **方法**: `PUT`
- **认证**: 需要
- **描述**: 更新系统告警规则配置

#### 请求体示例
```json
{
  "id": "cpu_high",
  "name": "CPU使用率过高",
  "metric": "cpu_usage",
  "operator": ">",
  "threshold": 85.0,
  "duration": 300,
  "level": "warning",
  "enabled": true
}
```

## 3. 系统健康

### 3.1 获取系统健康状态
- **URL**: `/api/v1/monitoring/health`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取系统各组件健康状态

#### 响应示例
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "components": {
      "database": {
        "status": "healthy",
        "message": "数据库连接正常"
      },
      "redis": {
        "status": "healthy",
        "message": "Redis连接正常"
      },
      "elasticsearch": {
        "status": "healthy",
        "message": "Elasticsearch连接正常"
      },
      "storage": {
        "status": "healthy",
        "message": "存储空间充足"
      }
    },
    "timestamp": "2024-01-01T10:00:00Z"
  }
}
```

### 3.2 获取系统信息
- **URL**: `/api/v1/monitoring/info`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取系统基本信息和运行状态

#### 响应示例
```json
{
  "success": true,
  "data": {
    "version": "1.0.0",
    "build_time": "2024-01-01T00:00:00Z",
    "git_commit": "abc123",
    "go_version": "go1.21.0",
    "start_time": "2024-01-01T09:00:00Z",
    "uptime": "1h0m0s",
    "environment": "development",
    "system_metrics": {
      "cpu_usage": 45.2,
      "memory_usage": 50.0,
      "disk_usage": 50.0,
      "goroutines": 25,
      "request_count": 1000,
      "error_count": 10
    }
  }
}
```

## 4. 健康检查端点

### 4.1 健康检查
- **URL**: `/health`
- **方法**: `GET`
- **认证**: 不需要
- **描述**: 系统健康检查端点，用于负载均衡器和服务发现

#### 响应示例
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T10:00:00Z",
  "version": "1.0.0"
}
```

### 4.2 就绪检查
- **URL**: `/ready`
- **方法**: `GET`
- **认证**: 不需要
- **描述**: 系统就绪检查端点，用于Kubernetes就绪探针

#### 响应示例
```json
{
  "ready": true,
  "message": "系统就绪",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 4.3 存活检查 (待实现)

> **注意：** 此功能当前未实现。
- **URL**: `/live`
- **方法**: `GET`
- **认证**: 不需要
- **描述**: 系统存活检查端点，用于Kubernetes存活探针

#### 响应示例
```json
{
  "alive": true,
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 4.4 Prometheus指标
- **URL**: `/metrics`
- **方法**: `GET`
- **认证**: 不需要
- **描述**: 暴露Prometheus格式的监控指标

#### 响应格式
Prometheus格式的纯文本指标数据。

## 5. 错误响应

所有API返回统一的错误响应格式：

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "错误描述",
    "details": {}
  }
}
```

常见错误码：
- `UNAUTHORIZED`: 未认证
- `FORBIDDEN`: 权限不足
- `NOT_FOUND`: 资源不存在
- `BAD_REQUEST`: 请求参数错误
- `INTERNAL_ERROR`: 服务器内部错误