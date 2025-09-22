# 系统监控指南

## 概述

本指南描述法律事务所自动化系统的监控体系，包括应用性能监控、健康检查、日志管理和告警机制。系统采用多层次的监控策略，确保服务稳定性和可用性。

## 监控架构

### 核心组件

1. **Prometheus** - 指标收集和存储
2. **Gin Prometheus 中间件** - HTTP 请求监控
3. **健康检查系统** - 服务状态监控
4. **日志系统** - 结构化日志记录
5. **告警机制** - 实时问题通知

### 监控层次

```
┌─────────────────────────────────────────────────────────────┐
│                    用户层监控                                │
├─────────────────────────────────────────────────────────────┤
│                    应用层监控                                │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │  API 监控   │ │  业务监控   │ │  错误监控   │            │
│  └─────────────┘ └─────────────┘ └─────────────┘            │
├─────────────────────────────────────────────────────────────┤
│                    基础设施监控                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │  数据库监控 │ │  缓存监控   │ │  系统监控   │            │
│  └─────────────┘ └─────────────┘ └─────────────┘            │
└─────────────────────────────────────────────────────────────┘
```

## 监控端点

### 1. Prometheus 指标端点

**端点**: `/metrics`

**用途**: 暴露应用指标给 Prometheus

**访问示例**:
```bash
curl http://localhost:8080/metrics
```

**主要指标**:
- `http_requests_total` - HTTP 请求总数
- `http_request_duration_seconds` - 请求响应时间
- `http_requests_in_progress` - 当前处理中的请求数
- `process_cpu_seconds_total` - CPU 使用时间
- `process_memory_bytes` - 内存使用量

### 2. 健康检查端点

#### 基础健康检查

**端点**: `/health`

**用途**: 快速服务可用性检查

```bash
curl http://localhost:8080/health
```

**响应格式**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

#### 详细健康检查

**端点**: `/api/v1/health/detailed`

**用途**: 详细的系统组件状态检查

```bash
curl http://localhost:8080/api/v1/health/detailed
```

**响应格式**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h 15m 30s",
  "components": {
    "database": {
      "status": "healthy",
      "response_time": "5ms",
      "last_check": "2024-01-15T10:29:55Z"
    },
    "cache": {
      "status": "healthy",
      "response_time": "1ms",
      "last_check": "2024-01-15T10:29:58Z"
    },
    "storage": {
      "status": "healthy",
      "available_space": "85GB",
      "last_check": "2024-01-15T10:29:56Z"
    }
  }
}
```

#### 存活探测

**端点**: `/health/live`

**用途**: Kubernetes Liveness Probe

```bash
curl http://localhost:8080/health/live
```

#### 就绪探测

**端点**: `/health/ready`

**用途**: Kubernetes Readiness Probe

```bash
curl http://localhost:8080/health/ready
```

### 3. 监控仪表板端点

#### 监控状态

**端点**: `/api/v1/monitor/status`

**用途**: 获取监控服务状态

```bash
curl http://localhost:8080/api/v1/monitor/status
```

**响应格式**:
```json
{
  "success": true,
  "data": {
    "service_status": "running",
    "start_time": "2024-01-15T08:00:00Z",
    "uptime": "2h 30m",
    "metrics_collected": 1250,
    "alerts_active": 0
  }
}
```

#### 性能统计

**端点**: `/api/v1/monitor/performance`

**用途**: 获取性能统计数据

```bash
curl http://localhost:8080/api/v1/monitor/performance
```

#### 监控仪表板

**端点**: `/api/v1/monitor/dashboard`

**用途**: 获取完整的监控仪表板数据

```bash
curl http://localhost:8080/api/v1/monitor/dashboard
```

### 4. 告警管理端点

#### 获取告警

**端点**: `/api/v1/monitor/alerts`

**用途**: 获取当前活动告警

```bash
curl http://localhost:8080/api/v1/monitor/alerts
```

**响应格式**:
```json
{
  "success": true,
  "data": [
    {
      "id": "alert_001",
      "type": "high_cpu_usage",
      "severity": "warning",
      "message": "CPU usage exceeded 80%",
      "timestamp": "2024-01-15T10:25:00Z",
      "resolved": false
    }
  ]
}
```

#### 解决告警

**端点**: `/api/v1/monitor/alerts/:id/resolve`

**方法**: `POST`

**用途**: 手动解决告警

```bash
curl -X POST http://localhost:8080/api/v1/monitor/alerts/alert_001/resolve
```

## 监控配置

### 环境配置

#### 开发环境

```go
// 开发环境监控配置
monitorConfig := metrics.MonitorConfig{
    EnableRealTimeAlerts: false,
    MetricsCollectionInterval: 10 * time.Second,
    HealthCheckInterval: 15 * time.Second,
    FailureThreshold: 2,
}
```

#### 生产环境

```go
// 生产环境监控配置
monitorConfig := metrics.MonitorConfig{
    EnableRealTimeAlerts: true,
    MetricsCollectionInterval: 30 * time.Second,
    HealthCheckInterval: 30 * time.Second,
    FailureThreshold: 3,
    EnableExternalAPICheck: true,
}
```

### 健康检查配置

```go
healthConfig := &health.HealthConfig{
    CheckInterval:            30 * time.Second,
    FailureThreshold:          3,
    SuccessThreshold:          2,
    Timeout:                  5 * time.Second,
    
    // 组件特定配置
    DatabaseTimeout:         3 * time.Second,
    CacheTimeout:           2 * time.Second,
    StorageTimeout:         5 * time.Second,
    ConcurrencyTimeout:      3 * time.Second,
    ExternalAPITimeout:      10 * time.Second,
    
    // 路径配置
    StoragePath:            "./storage",
    ExternalServiceURL:      "https://api.external-service.com/health",
    
    // 特性开关
    EnableExternalAPICheck:  true,
    EnableDetailedLogging:   true,
}
```

## 监控指标详解

### 应用指标

#### HTTP 请求指标

```yaml
# 请求总数
http_requests_total{method="GET", endpoint="/api/v1/cases", status="200"}

# 请求持续时间
http_request_duration_seconds_bucket{method="POST", endpoint="/api/v1/cases", le="0.1"}

# 当前进行中的请求
http_requests_in_progress{method="GET", endpoint="/api/v1/cases"}
```

#### 业务指标

```yaml
# 用户活跃度
user_active_sessions_total

# 案件处理统计
cases_processed_total{status="completed"}
cases_processing_duration_seconds_bucket

# 客户管理统计
clients_created_total
clients_updated_total
```

### 系统指标

#### 数据库指标

```yaml
# 数据库连接池
database_connections_active
database_connections_idle
database_connections_max

# 查询性能
database_query_duration_seconds_bucket{query="SELECT"}
database_query_errors_total{query="INSERT"}
```

#### 缓存指标

```yaml
# 缓存命中率
cache_hits_total
cache_misses_total
cache_hit_ratio

# 缓存性能
cache_operations_duration_seconds_bucket{operation="GET"}
cache_operations_duration_seconds_bucket{operation="SET"}
```

### 告警规则

#### 关键告警阈值

```yaml
# CPU 使用率告警
- alert: HighCPUUsage
  expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "高 CPU 使用率"
    description: "实例 {{ $labels.instance }} CPU 使用率超过 80%"

# 内存使用率告警
- alert: HighMemoryUsage
  expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 85
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "高内存使用率"
    description: "实例 {{ $labels.instance }} 内存使用率超过 85%"

# 数据库连接告警
- alert: DatabaseConnectionHigh
  expr: database_connections_active / database_connections_max > 0.8
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "数据库连接池使用率过高"
    description: "数据库连接池使用率超过 80%"

# HTTP 错误率告警
- alert: HighHTTPErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "HTTP 5xx 错误率过高"
    description: "过去5分钟内 HTTP 5xx 错误率超过 5%"
```

## 监控部署

### Prometheus 配置

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'law-office-api'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
    scrape_timeout: 10s

  - job_name: 'law-office-health'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics/health'
    scrape_interval: 30s
```

### Grafana 仪表板

推荐的关键仪表板:

1. **应用概览仪表板**
   - HTTP 请求量和响应时间
   - 错误率和状态码分布
   - 系统资源使用情况

2. **数据库监控仪表板**
   - 连接池状态
   - 查询性能
   - 慢查询统计

3. **业务指标仪表板**
   - 用户活跃度
   - 案件处理统计
   - 客户管理统计

### 日志监控

#### 结构化日志格式

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "correlation_id": "req_123456",
  "event": "user_login",
  "user_id": "user_789",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "duration_ms": 125,
  "status": "success"
}
```

#### 日志收集配置

```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/law-office-app/*.log
  json.keys_under_root: true
  json.add_error_key: true

output.elasticsearch:
  hosts: ["localhost:9200"]
  index: "law-office-logs-%{+yyyy.MM.dd}"
```

## 监控最佳实践

### 1. 监控策略

- **分层监控**: 从基础设施到应用层的多层次监控
- **关键指标优先**: 优先监控对业务影响最大的指标
- **告警分级**: 根据严重程度设置不同告警级别
- **避免告警疲劳**: 合理设置告警阈值和频率

### 2. 性能基准

#### 响应时间基准

```yaml
API响应时间基准:
  P50:  < 100ms
  P90:  < 300ms
  P95:  < 500ms
  P99:  < 1000ms
```

#### 可用性基准

```yaml
可用性目标:
  整体可用性: 99.9%
  关键API:    99.95%
  非关键API:  99.5%
```

### 3. 监控维护

- **定期审查**: 每月审查监控配置和告警规则
- **性能优化**: 根据监控数据优化系统性能
- **文档更新**: 及时更新监控文档和配置

### 4. 故障响应

- **告警响应**: 建立告警响应流程和责任人制度
- **故障处理**: 制定标准化的故障处理流程
- **事后分析**: 对重大故障进行根因分析和总结

## 监控故障排除

### 常见问题

#### 1. Prometheus 无法抓取指标

**检查项**:
- 确认 `/metrics` 端点可访问
- 检查网络连接和防火墙设置
- 验证 Prometheus 配置文件语法

**解决方案**:
```bash
# 测试端点可访问性
curl http://localhost:8080/metrics

# 检查 Prometheus 配置
promtool check config prometheus.yml
```

#### 2. 健康检查失败

**检查项**:
- 数据库连接状态
- 缓存服务状态
- 存储空间可用性
- 外部依赖服务状态

**解决方案**:
```bash
# 检查详细健康状态
curl http://localhost:8080/api/v1/health/detailed

# 检查组件健康状态
curl http://localhost:8080/api/v1/health/dependencies
```

#### 3. 告警误报

**检查项**:
- 告警阈值设置是否合理
- 指标采集频率是否适当
- 是否存在异常数据点

**解决方案**:
- 调整告警阈值和持续时间
- 增加数据过滤和平滑处理
- 设置告警抑制规则

### 性能优化

#### 1. 指标收集优化

```go
// 优化指标收集配置
monitorConfig := metrics.MonitorConfig{
    MetricsCollectionInterval: 30 * time.Second,  // 合理的收集间隔
    EnableDetailedMetrics:    false,             // 关闭详细指标以减少开销
    BatchSize:               1000,              // 批量处理大小
}
```

#### 2. 健康检查优化

```go
// 优化健康检查配置
healthConfig := &health.HealthConfig{
    CheckInterval:     30 * time.Second,  // 适当延长检查间隔
    Timeout:          5 * time.Second,   // 合理的超时时间
    ParallelChecks:    true,              // 并行执行检查
    CacheResults:      true,              // 缓存检查结果
    CacheTTL:         10 * time.Second,  // 缓存有效期
}
```

## 监控扩展

### 自定义指标

#### 添加业务指标

```go
// 在业务代码中添加自定义指标
var (
    casesCreated = promauto.NewCounter(prometheus.CounterOpts{
        Name: "cases_created_total",
        Help: "Total number of cases created",
    })
    
    caseProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "case_processing_duration_seconds",
        Help: "Time spent processing cases",
        Buckets: prometheus.DefBuckets,
    })
)

// 在业务逻辑中使用
func CreateCase(case *Case) error {
    start := time.Now()
    defer func() {
        caseProcessingDuration.Observe(time.Since(start).Seconds())
    }()
    
    // 业务逻辑...
    casesCreated.Inc()
    
    return nil
}
```

### 第三方集成

#### DataDog 集成

```go
// DataDog 监控集成
import "github.com/DataDog/datadog-go/statsd"

func initDataDog() (*statsd.Client, error) {
    client, err := statsd.New("localhost:8125")
    if err != nil {
        return nil, err
    }
    
    // 设置全局标签
    client.Tags = []string{"service:law-office-api", "env:production"}
    
    return client, nil
}
```

#### New Relic 集成

```go
// New Relic 监控集成
import "github.com/newrelic/go-agent"

func initNewRelic() (newrelic.Application, error) {
    config := newrelic.NewConfig("Law Office API", os.Getenv("NEW_RELIC_LICENSE_KEY"))
    config.Enabled = true
    config.Labels = map[string]string{
        "environment": "production",
        "service":     "law-office-api",
    }
    
    return newrelic.NewApplication(config)
}
```

## 监控安全

### 安全措施

1. **认证和授权**
   - 监控端点需要适当的访问控制
   - 敏感指标需要特殊权限

2. **数据加密**
   - 监控数据传输加密
   - 存储数据加密保护

3. **访问控制**
   - 基于角色的访问控制
   - IP 地址白名单

### 监控端点安全

```go
// 添加认证中间件到监控端点
monitorGroup := app.Group("/api/v1/monitor")
monitorGroup.Use(middleware.AdminAuthMiddleware())
{
    monitorGroup.GET("/status", monitorStatusHandler)
    monitorGroup.GET("/performance", monitorPerformanceHandler)
    monitorGroup.GET("/dashboard", monitorDashboardHandler)
    monitorGroup.GET("/alerts", monitorAlertsHandler)
}
```

## 总结

本监控系统提供了全面的覆盖，从基础设施到应用层的多层次监控。通过合理配置和定期维护，可以确保系统的稳定运行和快速问题发现。

关键要点：
- 建立完整的监控指标体系
- 配置合理的告警规则
- 定期审查和优化监控配置
- 建立标准化的故障响应流程
- 保障监控系统的安全性

通过实施本监控指南，可以显著提高系统的可观测性和运维效率。