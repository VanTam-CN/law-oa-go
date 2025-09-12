# 日志管理 API 文档

## 1. 日志查询

### 1.1 查询日志
- **URL**: `/api/v1/logs`
- **方法**: `POST`
- **认证**: 需要
- **描述**: 根据条件查询系统日志

#### 请求体示例
```json
{
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-01T23:59:59Z",
  "levels": [1, 2, 3],
  "services": ["law-oa", "monitoring"],
  "users": [1, 2],
  "usernames": ["admin", "user1"],
  "request_ids": ["req_123", "req_456"],
  "ip_addresses": ["192.168.1.1", "192.168.1.2"],
  "methods": ["GET", "POST"],
  "paths": ["/api/v1/users", "/api/v1/cases"],
  "status_codes": [200, 404, 500],
  "keywords": ["error", "warning"],
  "tags": ["auth", "database"],
  "limit": 100,
  "offset": 0,
  "sort_by": "timestamp",
  "sort_order": "desc"
}
```

#### 响应示例
```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": 1,
        "timestamp": "2024-01-01T10:00:00Z",
        "level": 1,
        "level_name": "INFO",
        "message": "用户登录成功",
        "service": "law-oa",
        "function": "law-oa/internal/handlers/user_handler.(*UserHandler).Login",
        "file": "/app/internal/handlers/user_handler.go",
        "line": 45,
        "user_id": 1,
        "username": "admin",
        "request_id": "req_123",
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0",
        "method": "POST",
        "path": "/api/v1/auth/login",
        "status_code": 200,
        "duration": 0.125,
        "error": "",
        "stacktrace": "",
        "tags": ["auth"],
        "metadata": {
          "login_type": "password"
        },
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "size": 100
  }
}
```

### 1.2 搜索日志
- **URL**: `/api/v1/logs/search`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 根据关键词搜索日志

#### 查询参数
- `query` (string, 必需): 搜索关键词
- `level` (string, 可选): 日志级别
- `service` (string, 可选): 服务名称
- `start_time` (string, 可选): 开始时间 (RFC3339格式)
- `end_time` (string, 可选): 结束时间 (RFC3339格式)
- `limit` (int, 可选): 限制数量，默认100

#### 响应示例
```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": 1,
        "timestamp": "2024-01-01T10:00:00Z",
        "level": 3,
        "level_name": "ERROR",
        "message": "数据库连接失败",
        "service": "law-oa",
        "username": "system",
        "request_id": "req_456",
        "error": "connection refused"
      }
    ],
    "total": 1,
    "query": "数据库连接失败",
    "filters": {
      "keywords": ["数据库", "连接", "失败"],
      "level": 3
    },
    "timestamp": "2024-01-01T10:00:00Z"
  }
}
```

### 1.3 获取日志详情
- **URL**: `/api/v1/logs/{id}`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 根据日志ID获取详细信息

#### 响应示例
```json
{
  "success": true,
  "data": {
    "id": 1,
    "timestamp": "2024-01-01T10:00:00Z",
    "level": 3,
    "level_name": "ERROR",
    "message": "数据库连接失败",
    "service": "law-oa",
    "function": "law-oa/internal/db.(*Database).Connect",
    "file": "/app/internal/db/db.go",
    "line": 78,
    "user_id": null,
    "username": "system",
    "request_id": "req_456",
    "ip_address": "127.0.0.1",
    "user_agent": "",
    "method": "",
    "path": "",
    "status_code": 0,
    "duration": 0,
    "error": "connection refused",
    "stacktrace": "goroutine 1 [running]:\nmain.main()\n\t/app/main.go:45 +0x123",
    "tags": ["database", "error"],
    "metadata": {
      "db_host": "localhost",
      "db_port": 3306,
      "db_name": "law_oa"
    },
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

## 2. 日志统计

### 2.1 获取日志统计
- **URL**: `/api/v1/logs/stats`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取指定时间范围内的日志统计信息

#### 查询参数
- `start_time` (string, 可选): 开始时间 (RFC3339格式)
- `end_time` (string, 可选): 结束时间 (RFC3339格式)

#### 响应示例
```json
{
  "success": true,
  "data": {
    "total_count": 10000,
    "level_counts": {
      "0": 2000,
      "1": 3000,
      "2": 2500,
      "3": 1500,
      "4": 1000
    },
    "service_counts": {
      "law-oa": 5000,
      "monitoring": 3000,
      "logging": 2000
    },
    "user_counts": {
      "admin": 4000,
      "user1": 3000,
      "user2": 3000
    },
    "ip_counts": {
      "192.168.1.1": 2000,
      "192.168.1.2": 1500,
      "127.0.0.1": 1000
    },
    "method_counts": {
      "GET": 6000,
      "POST": 3000,
      "PUT": 1000
    },
    "error_counts": {
      "connection refused": 500,
      "timeout": 300,
      "not found": 200
    },
    "hourly_stats": {
      "2024-01-01 09:00:00": 500,
      "2024-01-01 10:00:00": 600,
      "2024-01-01 11:00:00": 400
    },
    "daily_stats": {
      "2024-01-01": 5000,
      "2024-01-02": 3000,
      "2024-01-03": 2000
    }
  }
}
```

### 2.2 获取日志趋势
- **URL**: `/api/v1/logs/trends`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取日志数量变化趋势

#### 查询参数
- `period` (string, 可选): 时间周期 (hourly, daily, weekly)，默认hourly
- `start_time` (string, 可选): 开始时间 (RFC3339格式)
- `end_time` (string, 可选): 结束时间 (RFC3339格式)

#### 响应示例
```json
{
  "success": true,
  "data": {
    "period": "hourly",
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-01-01T23:59:59Z",
    "data": {
      "2024-01-01 00:00:00": 100,
      "2024-01-01 01:00:00": 120,
      "2024-01-01 02:00:00": 80,
      "2024-01-01 03:00:00": 90,
      "2024-01-01 04:00:00": 110
    },
    "stats": {
      "total_count": 10000,
      "level_counts": {
        "0": 2000,
        "1": 3000,
        "2": 2500,
        "3": 1500,
        "4": 1000
      }
    }
  }
}
```

### 2.3 获取错误统计
- **URL**: `/api/v1/logs/errors/top`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取最常见的错误信息统计

#### 查询参数
- `start_time` (string, 可选): 开始时间 (RFC3339格式)
- `end_time` (string, 可选): 结束时间 (RFC3339格式)
- `limit` (int, 可选): 限制数量，默认10

#### 响应示例
```json
{
  "success": true,
  "data": {
    "start_time": "2024-01-01T00:00:00Z",
    "end_time": "2024-01-01T23:59:59Z",
    "errors": [
      {
        "error": "connection refused",
        "count": 500
      },
      {
        "error": "timeout",
        "count": 300
      },
      {
        "error": "not found",
        "count": 200
      }
    ],
    "total": 3
  }
}
```

## 3. 日志管理

### 3.1 导出日志
- **URL**: `/api/v1/logs/export`
- **方法**: `POST`
- **认证**: 需要
- **描述**: 导出指定条件的日志数据

#### 请求体示例
```json
{
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-01T23:59:59Z",
  "levels": [1, 2, 3],
  "services": ["law-oa"],
  "keywords": ["error"],
  "limit": 1000
}
```

#### 查询参数
- `format` (string, 可选): 导出格式 (json, csv)，默认json

### 3.2 清理日志
- **URL**: `/api/v1/logs/clear`
- **方法**: `POST`
- **认证**: 需要
- **描述**: 清理指定时间之前的日志数据

#### 请求体示例
```json
{
  "before_time": "2023-12-31T23:59:59Z",
  "level": 1,
  "service": "law-oa"
}
```

#### 响应示例
```json
{
  "success": true,
  "data": {
    "deleted_count": 1000,
    "before_time": "2023-12-31T23:59:59Z"
  }
}
```

## 4. 配置管理

### 4.1 获取日志配置
- **URL**: `/api/v1/logs/config`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取系统日志配置信息

#### 响应示例
```json
{
  "success": true,
  "data": {
    "level": 1,
    "format": "json",
    "output": "both",
    "file_path": "logs/app.log",
    "max_size": 100,
    "max_backups": 10,
    "max_age": 30,
    "compress": true,
    "enable_console": true,
    "enable_file": true,
    "enable_db": true,
    "enable_es": false
  }
}
```

### 4.2 更新日志配置
- **URL**: `/api/v1/logs/config`
- **方法**: `PUT`
- **认证**: 需要
- **描述**: 更新系统日志配置

#### 请求体示例
```json
{
  "level": 0,
  "format": "json",
  "output": "both",
  "file_path": "logs/app.log",
  "max_size": 200,
  "max_backups": 20,
  "max_age": 60,
  "compress": true,
  "enable_console": true,
  "enable_file": true,
  "enable_db": true,
  "enable_es": true
}
```

## 5. 辅助接口

### 5.1 获取日志级别
- **URL**: `/api/v1/logs/levels`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取系统支持的日志级别

#### 响应示例
```json
{
  "success": true,
  "data": [
    {
      "value": 0,
      "name": "DEBUG",
      "color": "gray"
    },
    {
      "value": 1,
      "name": "INFO",
      "color": "blue"
    },
    {
      "value": 2,
      "name": "WARN",
      "color": "orange"
    },
    {
      "value": 3,
      "name": "ERROR",
      "color": "red"
    },
    {
      "value": 4,
      "name": "FATAL",
      "color": "darkred"
    }
  ]
}
```

### 5.2 获取服务列表
- **URL**: `/api/v1/logs/services`
- **方法**: `GET`
- **认证**: 需要
- **描述**: 获取系统中所有服务的名称

#### 响应示例
```json
{
  "success": true,
  "data": [
    "law-oa",
    "monitoring",
    "logging",
    "database",
    "redis",
    "elasticsearch"
  ]
}
```

## 6. 错误响应

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

## 7. 日志级别说明

- **DEBUG (0)**: 调试信息，用于开发调试
- **INFO (1)**: 一般信息，记录系统运行状态
- **WARN (2)**: 警告信息，表示可能出现问题
- **ERROR (3)**: 错误信息，表示发生了错误
- **FATAL (4)**: 致命错误，表示系统无法继续运行

## 8. 时间格式

所有时间参数使用RFC3339格式：
- `2024-01-01T10:00:00Z`
- `2024-01-01T10:00:00+08:00`