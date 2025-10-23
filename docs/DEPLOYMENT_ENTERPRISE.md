# Law OA Go 企业级运维指南

<div align="center">

![Monitoring](https://img.shields.io/badge/Monitoring-Prometheus-orange?style=for-the-badge&logo=prometheus&logoColor=white)
![Logging](https://img.shields.io/badge/Logging-ELK-blue?style=for-the-badge&logo=elasticsearch&logoColor=white)
![Security](https://img.shields.io/badge/Security-RBAC-green?style=for-the-badge&logo=kubernetes&logoColor=white)
![Performance](https://img.shields.io/badge/Performance-Optimized-red?style=for-the-badge&logo=apachetomcat&logoColor=white)

**企业级运维最佳实践**

[监控配置](#-监控配置) • [安全配置](#-安全配置) • [性能优化](#-性能优化) • [备份策略](#-备份策略) • [故障排除](#-故障排除) • [CI/CD集成](#cicd集成)

</div>

---

## 📋 目录

- [5. 监控配置](#5-监控配置)
- [6. 安全配置](#-6-安全配置)
- [7. 性能优化](#-7-性能优化)
- [8. 备份策略](#-8-备份策略)
- [9. 故障排除](#-9-故障排除)
- [10. CI/CD集成](#-10-cicd集成)

---

## 5. 监控配置

### 5.1 Prometheus监控

#### ServiceMonitor配置
```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: law-oa-metrics
  namespace: law-oa
  labels:
    app: law-oa-app
spec:
  selector:
    matchLabels:
      app: law-oa-service
  endpoints:
  - port: metrics
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
```

#### Prometheus配置
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "law-oa-alerts.yml"

scrape_configs:
  - job_name: 'law-oa-app'
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
            - law-oa
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_name]
        action: keep
        regex: law-oa-service
      - source_labels: [__meta_kubernetes_endpoint_port_name]
        action: keep
        regex: metrics

  - job_name: 'mysql'
    static_configs:
      - targets: ['mysql-exporter:9104']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  - job_name: 'elasticsearch'
    static_configs:
      - targets: ['elasticsearch-exporter:9114']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

#### 告警规则
```yaml
# law-oa-alerts.yml
groups:
- name: law-oa.rules
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} errors per second"

  - alert: HighResponseTime
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High response time detected"
      description: "95th percentile response time is {{ $value }} seconds"

  - alert: HighCPUUsage
    expr: rate(container_cpu_usage_seconds_total[5m]) * 100 > 80
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: "High CPU usage detected"
      description: "CPU usage is {{ $value }}%"

  - alert: HighMemoryUsage
    expr: container_memory_usage_bytes / container_spec_memory_limit_bytes * 100 > 90
    for: 10m
    labels:
      severity: critical
    annotations:
      summary: "High memory usage detected"
      description: "Memory usage is {{ $value }}%"

  - alert: DatabaseConnectionsHigh
    expr: mysql_global_status_threads_connected / mysql_global_variables_max_connections * 100 > 80
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Database connections high"
      description: "Database connection usage is {{ $value }}%"

  - alert: RedisMemoryHigh
    expr: redis_memory_used_bytes / redis_memory_max_bytes * 100 > 90
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Redis memory usage high"
      description: "Redis memory usage is {{ $value }}%"
```

### 5.2 Grafana仪表板

#### 应用性能仪表板
```json
{
  "dashboard": {
    "title": "Law OA Go - Application Performance",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total[5m])) by (method)",
            "legendFormat": "{{method}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m]))",
            "legendFormat": "Error Rate"
          }
        ]
      }
    ]
  }
}
```

### 5.3 日志收集

#### Fluentd配置
```yaml
# fluentd-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
  namespace: law-oa
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/*law-oa*.log
      pos_file /var/log/fluentd-law-oa.log.pos
      tag kubernetes.*
      format json
      time_format %Y-%m-%dT%H:%M:%S.%NZ
    </source>

    <filter kubernetes.**>
      @type kubernetes_metadata
    </filter>

    <filter kubernetes.**>
      @type record_transformer
      <record>
        environment #{ENV['ENVIRONMENT'] || 'production'}
        service_name law-oa-go
      </record>
    </filter>

    <match kubernetes.**>
      @type elasticsearch
      host elasticsearch
      port 9200
      index_name law-oa-logs
      type_name _doc
      include_tag_key true
      tag_key @log_name
      <buffer>
        @type file
        path /var/log/fluentd-buffers/kubernetes.system.buffer
        flush_mode interval
        retry_type exponential_backoff
        flush_thread_count 2
        flush_interval 5s
        retry_forever
        retry_max_interval 30
        chunk_limit_size 2M
        queue_limit_length 8
        overflow_action block
      </buffer>
    </match>
```

#### 日志格式化配置
```go
// internal/logging/logger.go
package logging

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// StructuredLogger 结构化日志记录器
type StructuredLogger struct {
    *zap.Logger
}

// NewStructuredLogger 创建新的结构化日志记录器
func NewStructuredLogger(level string) *StructuredLogger {
    config := zap.NewProductionConfig()

    switch level {
    case "debug":
        config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
    case "info":
        config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
    case "warn":
        config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
    case "error":
        config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
    default:
        config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
    }

    config.OutputPaths = []string{"stdout"}
    config.ErrorOutputPaths = []string{"stderr"}

    logger, _ := config.Build()
    return &StructuredLogger{logger}
}

// WithContext 添加上下文信息
func (l *StructuredLogger) WithContext(fields ...zap.Field) *StructuredLogger {
    return &StructuredLogger{l.Logger.With(fields...)}
}

// LogRequest 记录HTTP请求
func (l *StructuredLogger) LogRequest(method, path string, statusCode int, duration float64, userID string) {
    l.Info("HTTP Request",
        zap.String("method", method),
        zap.String("path", path),
        zap.Int("status_code", statusCode),
        zap.Float64("duration_ms", duration),
        zap.String("user_id", userID),
        zap.String("request_id", generateRequestID()),
    )
}

// LogError 记录错误
func (l *StructuredLogger) LogError(err error, message string, fields ...zap.Field) {
    allFields := append([]zap.Field{
        zap.Error(err),
        zap.String("error_type", "application"),
    }, fields...)

    l.Error(message, allFields...)
}
```

### 5.4 分布式追踪

#### Jaeger配置
```yaml
# jaeger-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jaeger
  namespace: law-oa
spec:
  replicas: 1
  selector:
    matchLabels:
      app: jaeger
  template:
    metadata:
      labels:
        app: jaeger
    spec:
      containers:
      - name: jaeger
        image: jaegertracing/all-in-one:1.47
        ports:
        - containerPort: 16686
          name: ui
        - containerPort: 14268
          name: collector
        env:
        - name: SPAN_STORAGE_TYPE
          value: elasticsearch
        - name: ES_SERVER_URLS
          value: http://elasticsearch:9200
        - name: ES_INDEX_PREFIX
          value: jaeger
        - name: COLLECTOR_ZIPKIN_HTTP_PORT
          value: "9411"
```

#### OpenTelemetry集成
```go
// internal/tracing/tracer.go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// InitTracer 初始化分布式追踪
func InitTracer(serviceName, jaegerEndpoint string) (*trace.TracerProvider, error) {
    // 创建 Jaeger exporter
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
    if err != nil {
        return nil, err
    }

    // 创建资源
    res, err := resource.New(context.Background(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        ),
    )
    if err != nil {
        return nil, err
    }

    // 创建 TracerProvider
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(res),
    )

    // 设置全局 TracerProvider
    otel.SetTracerProvider(tp)

    return tp, nil
}

// StartSpan 开始一个新的 span
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
    return otel.Tracer("law-oa-go").Start(ctx, name)
}
```

---

## 6. 安全配置

### 6.1 网络安全

#### 网络策略
```yaml
# networkpolicy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: law-oa-netpol
  namespace: law-oa
spec:
  podSelector:
    matchLabels:
      app: law-oa-app
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: law-oa
    ports:
    - protocol: TCP
      port: 3306
    - protocol: TCP
      port: 6379
    - protocol: TCP
      port: 9200
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 443
```

#### 服务网格（可选）
```yaml
# istio-gateway.yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: law-oa-gateway
  namespace: law-oa
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: law-oa-tls
    hosts:
    - api.lawoa.com
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: law-oa-vs
  namespace: law-oa
spec:
  hosts:
  - api.lawoa.com
  gateways:
  - law-oa-gateway
  http:
  - match:
    - uri:
        prefix: /
    route:
    - destination:
        host: law-oa-service
        port:
          number: 8080
    fault:
      delay:
        percentage:
          value: 0.1
        fixedDelay: 5s
```

### 6.2 Pod安全

#### PodSecurityPolicy
```yaml
# podsecuritypolicy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: law-oa-psp
  namespace: law-oa
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true
```

#### 安全上下文
```yaml
# security-context.yaml
apiVersion: v1
kind: Pod
metadata:
  name: law-oa-app-secure
  namespace: law-oa
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 65534
    runAsGroup: 65534
    fsGroup: 65534
  containers:
  - name: law-oa-app
    image: law-registry.com/law-oa-go:2.1.0
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
      runAsNonRoot: true
      runAsUser: 65534
    volumeMounts:
    - name: tmp
      mountPath: /tmp
    - name: config
      mountPath: /app/config
      readOnly: true
  volumes:
  - name: tmp
    emptyDir: {}
  - name: config
    configMap:
      name: law-oa-config
```

### 6.3 RBAC配置

#### ServiceAccount和权限
```yaml
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: law-oa-sa
  namespace: law-oa
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: law-oa-role
  namespace: law-oa
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods", "services", "endpoints"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: law-oa-binding
  namespace: law-oa
subjects:
- kind: ServiceAccount
  name: law-oa-sa
  namespace: law-oa
roleRef:
  kind: Role
  name: law-oa-role
  apiGroup: rbac.authorization.k8s.io
```

### 6.4 密钥管理

#### External Secrets Operator
```yaml
# externalsecret.yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: law-oa-secret-store
  namespace: law-oa
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-west-2
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: law-oa-secrets
  namespace: law-oa
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: law-oa-secret-store
    kind: SecretStore
  target:
    name: law-oa-secret
    creationPolicy: Owner
  data:
  - secretKey: db-password
    remoteRef:
      key: law-oa/database-password
  - secretKey: jwt-secret
    remoteRef:
      key: law-oa/jwt-secret
```

---

## 7. 性能优化

### 7.1 数据库优化

#### MySQL配置优化
```yaml
# mysql-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mysql-config
  namespace: law-oa
data:
  my.cnf: |
    [mysqld]
    # 性能优化
    innodb_buffer_pool_size = 2G
    innodb_log_file_size = 256M
    innodb_log_buffer_size = 16M
    innodb_flush_log_at_trx_commit = 2
    innodb_flush_method = O_DIRECT
    innodb_io_capacity = 2000
    innodb_io_capacity_max = 4000

    # 连接配置
    max_connections = 500
    connect_timeout = 10
    wait_timeout = 600
    interactive_timeout = 600

    # 查询缓存
    query_cache_type = 1
    query_cache_size = 128M
    query_cache_limit = 2M

    # 慢查询日志
    slow_query_log = 1
    slow_query_log_file = /var/log/mysql/slow.log
    long_query_time = 2

    # 字符集
    character-set-server = utf8mb4
    collation-server = utf8mb4_unicode_ci

    # 二进制日志
    log_bin = /var/log/mysql/mysql-bin.log
    expire_logs_days = 7
    max_binlog_size = 100M

    # 性能模式
    performance_schema = ON
```

#### 连接池配置
```go
// internal/database/connection.go
package database

import (
    "time"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
    Host            string
    Port            int
    User            string
    Password        string
    Database        string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

// NewConnection 创建数据库连接
func NewConnection(config *DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.User, config.Password, config.Host, config.Port, config.Database)

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
        NowFunc: func() time.Time {
            return time.Now().Local()
        },
    })
    if err != nil {
        return nil, err
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // 连接池配置
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

    return db, nil
}
```

### 7.2 缓存优化

#### Redis配置
```yaml
# redis-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  namespace: law-oa
data:
  redis.conf: |
    # 内存优化
    maxmemory 2gb
    maxmemory-policy allkeys-lru

    # 持久化配置
    save 900 1
    save 300 10
    save 60 10000

    # 网络配置
    tcp-keepalive 300
    timeout 0

    # 安全配置
    requirepass your-redis-password
    rename-command FLUSHDB ""
    rename-command FLUSHALL ""

    # 慢日志
    slowlog-log-slower-than 10000
    slowlog-max-len 128

    # 客户端配置
    maxclients 10000

    # AOF配置
    appendonly yes
    appendfsync everysec
    no-appendfsync-on-rewrite no
    auto-aof-rewrite-percentage 100
    auto-aof-rewrite-min-size 64mb
```

#### 缓存策略实现
```go
// internal/cache/redis.go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/go-redis/redis/v8"
)

// RedisCache Redis缓存实现
type RedisCache struct {
    client *redis.Client
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(addr, password string, db int) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
        PoolSize: 50,
        MinIdleConns: 5,
        PoolTimeout: 30 * time.Second,
        IdleTimeout: 300 * time.Second,
        IdleCheckFrequency: 60 * time.Second,
    })

    return &RedisCache{client: rdb}
}

// Set 设置缓存
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    return c.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    data, err := c.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return ErrNotFound
        }
        return err
    }

    return json.Unmarshal([]byte(data), dest)
}

// Delete 删除缓存
func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

// SetNX 设置缓存（仅当key不存在时）
func (c *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
    data, err := json.Marshal(value)
    if err != nil {
        return false, err
    }

    return c.client.SetNX(ctx, key, data, expiration).Result()
}
```

### 7.3 应用优化

#### Go运行时配置
```yaml
# deployment-optimized.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: law-oa-app-optimized
  namespace: law-oa
spec:
  template:
    spec:
      containers:
      - name: law-oa-app
        image: law-registry.com/law-oa-go:2.1.0
        env:
        - name: GOMAXPROCS
          value: "4"
        - name: GOGC
          value: "100"
        - name: GOMEMLIMIT
          value: "2GiB"
        - name: GODEBUG
          value: "gctrace=1"
        resources:
          requests:
            cpu: 1000m
            memory: 2Gi
          limits:
            cpu: 4000m
            memory: 4Gi
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
```

#### 性能监控集成
```go
// internal/metrics/metrics.go
package metrics

import (
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    // HTTP请求计数器
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    // HTTP请求持续时间
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    // 数据库连接池
    dbConnectionsActive = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_active",
            Help: "Number of active database connections",
        },
    )

    // 缓存命中率
    cacheHitRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cache_hit_rate",
            Help: "Cache hit rate",
        },
        []string{"cache_type"},
    )
)

// InitMetrics 初始化指标
func InitMetrics() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(dbConnectionsActive)
    prometheus.MustRegister(cacheHitRate)

    // 启动指标端点
    http.Handle("/metrics", promhttp.Handler())
    go func() {
        http.ListenAndServe(":9090", nil)
    }()
}

// RecordHTTPRequest 记录HTTP请求指标
func RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
    httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", statusCode)).Inc()
    httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}
```

---

## 8. 备份策略

### 8.1 数据库备份

#### 自动备份CronJob
```yaml
# backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: mysql-backup
  namespace: law-oa
spec:
  schedule: "0 2 * * *"  # 每天凌晨2点
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: mysql-backup
            image: mysql:8.0
            command:
            - /bin/bash
            - -c
            - |
              mysqldump -h mysql -u root -p$MYSQL_ROOT_PASSWORD \
                --single-transaction --routines --triggers \
                --all-databases > /backup/backup-$(date +%Y%m%d-%H%M%S).sql

              # 压缩备份文件
              gzip /backup/backup-$(date +%Y%m%d-%H%M%S).sql

              # 上传到云存储（可选）
              if [ ! -z "$AWS_S3_BUCKET" ]; then
                aws s3 cp /backup/backup-$(date +%Y%m%d-%H%M%S).sql.gz s3://$AWS_S3_BUCKET/mysql-backups/
              fi
            env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mysql-secret
                  key: root-password
            - name: AWS_S3_BUCKET
              value: "law-oa-backups"
            volumeMounts:
            - name: backup-storage
              mountPath: /backup
          volumes:
          - name: backup-storage
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
```

#### 增量备份配置
```yaml
# incremental-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: mysql-incremental-backup
  namespace: law-oa
spec:
  schedule: "0 */4 * * *"  # 每4小时一次
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: mysql-incremental-backup
            image: percona/percona-xtrabackup:8.0
            command:
            - /bin/bash
            - -c
            - |
              # 创建增量备份
              xtrabackup --backup \
                --host=mysql \
                --user=root \
                --password=$MYSQL_ROOT_PASSWORD \
                --target-dir=/backup/incremental-$(date +%Y%m%d-%H%M%S) \
                --incremental-basedir=/backup/full-latest

              # 准备备份
              xtrabackup --prepare \
                --target-dir=/backup/incremental-$(date +%Y%m%d-%H%M%S) \
                --incremental-basedir=/backup/full-latest
            env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mysql-secret
                  key: root-password
            volumeMounts:
            - name: backup-storage
              mountPath: /backup
          volumes:
          - name: backup-storage
              claimName: backup-pvc
          restartPolicy: OnFailure
```

### 8.2 应用备份

#### 配置备份
```bash
#!/bin/bash
# backup-configs.sh

# 创建备份目录
BACKUP_DIR="/backup/configs/$(date +%Y%m%d-%H%M%S)"
mkdir -p $BACKUP_DIR

# 备份Kubernetes配置
kubectl get configmap -n law-oa -o yaml > $BACKUP_DIR/configmaps.yaml
kubectl get secret -n law-oa -o yaml > $BACKUP_DIR/secrets.yaml
kubectl get deployment -n law-oa -o yaml > $BACKUP_DIR/deployments.yaml
kubectl get service -n law-oa -o yaml > $BACKUP_DIR/services.yaml
kubectl get ingress -n law-oa -o yaml > $BACKUP_DIR/ingress.yaml

# 备份Helm配置
helm get values law-oa -n law-oa > $BACKUP_DIR/helm-values.yaml

# 备份Docker Compose配置（如果使用）
docker compose config > $BACKUP_DIR/docker-compose.yml

# 压缩备份
tar -czf $BACKUP_DIR.tar.gz $BACKUP_DIR
rm -rf $BACKUP_DIR

# 上传到云存储
if [ ! -z "$AWS_S3_BUCKET" ]; then
    aws s3 cp $BACKUP_DIR.tar.gz s3://$AWS_S3_BUCKET/config-backups/
fi

echo "Configuration backup completed: $BACKUP_DIR.tar.gz"
```

#### 灾难恢复脚本
```bash
#!/bin/bash
# disaster-recovery.sh

BACKUP_FILE=$1
NAMESPACE="law-oa"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file.tar.gz>"
    exit 1
fi

# 解压备份文件
tar -xzf $BACKUP_FILE
BACKUP_DIR=$(basename $BACKUP_FILE .tar.gz)

# 停止应用服务
kubectl scale deployment law-oa-app --replicas=0 -n $NAMESPACE

# 等待Pod终止
kubectl wait --for=delete pod -l app=law-oa-app -n $NAMESPACE --timeout=300s

# 恢复命名空间
kubectl apply -f $BACKUP_DIR/namespace.yaml

# 恢复配置和密钥
kubectl apply -f $BACKUP_DIR/configmaps.yaml
kubectl apply -f $BACKUP_DIR/secrets.yaml

# 恢复应用
kubectl apply -f $BACKUP_DIR/deployments.yaml
kubectl apply -f $BACKUP_DIR/services.yaml
kubectl apply -f $BACKUP_DIR/ingress.yaml

# 等待应用就绪
kubectl wait --for=condition=available --timeout=600s deployment/law-oa-app -n $NAMESPACE

# 验证部署
kubectl get pods -n $NAMESPACE
kubectl get services -n $NAMESPACE

echo "Disaster recovery completed successfully"
```

### 8.3 备份监控

#### 备份监控配置
```yaml
# backup-monitoring.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: backup-monitoring
  namespace: law-oa
data:
  backup-check.sh: |
    #!/bin/bash
    BACKUP_DIR="/backup"
    MAX_AGE_HOURS=24
    ALERT_EMAIL="admin@lawoa.com"

    # 检查备份文件
    LATEST_BACKUP=$(find $BACKUP_DIR -name "backup-*.sql.gz" -type f -mmin -$((MAX_AGE_HOURS*60)) | head -1)

    if [ -z "$LATEST_BACKUP" ]; then
        echo "ALERT: No backup found in the last $MAX_AGE_HOURS hours" | \
        mail -s "Backup Alert - Law OA Go" $ALERT_EMAIL

        # 发送Slack通知
        curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"🚨 Backup Alert: No backup found in the last 24 hours for Law OA Go"}' \
        $SLACK_WEBHOOK_URL
    else
        echo "Latest backup: $LATEST_BACKUP"

        # 检查备份文件大小
        BACKUP_SIZE=$(stat -f%z "$LATEST_BACKUP" 2>/dev/null || stat -c%s "$LATEST_BACKUP" 2>/dev/null)
        if [ "$BACKUP_SIZE" -lt 1000000 ]; then  # 小于1MB认为异常
            echo "ALERT: Backup file size is too small: $BACKUP_SIZE bytes" | \
            mail -s "Backup Size Alert - Law OA Go" $ALERT_EMAIL
        fi
    fi
```

#### 备份清理CronJob
```yaml
# backup-cleanup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-cleanup
  namespace: law-oa
spec:
  schedule: "0 3 * * 0"  # 每周日凌晨3点
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup-cleanup
            image: alpine:3.18
            command:
            - /bin/sh
            - -c
            - |
              # 删除7天前的备份文件
              find /backup -name "backup-*.sql.gz" -mtime +7 -delete
              find /backup -name "incremental-*" -mtime +7 -delete

              # 删除空的目录
              find /backup -type d -empty -delete

              # 清理S3中的旧备份（如果使用AWS）
              if [ ! -z "$AWS_S3_BUCKET" ]; then
                aws s3 ls s3://$AWS_S3_BUCKET/mysql-backups/ | \
                while read -r line; do
                  createDate=$(echo $line | awk '{print $1" "$2}')
                  createDate=$(date -d "$createDate" +%s)
                  olderThan=$(date -d "7 days ago" +%s)
                  if [[ $createDate -lt $olderThan ]]; then
                    fileName=$(echo $line | awk '{print $4}')
                    aws s3 rm s3://$AWS_S3_BUCKET/mysql-backups/$fileName
                  fi
                done
              fi

              echo "Backup cleanup completed"
            env:
            - name: AWS_S3_BUCKET
              value: "law-oa-backups"
            volumeMounts:
            - name: backup-storage
              mountPath: /backup
          volumes:
          - name: backup-storage
              claimName: backup-pvc
          restartPolicy: OnFailure
```

---

## 9. 故障排除

### 9.1 常见问题诊断

#### 应用启动失败
```bash
# 检查Pod状态
kubectl get pods -n law-oa
kubectl describe pod <pod-name> -n law-oa

# 查看Pod日志
kubectl logs <pod-name> -n law-oa --previous
kubectl logs -f deployment/law-oa-app -n law-oa

# 检查资源使用
kubectl top pods -n law-oa
kubectl describe node <node-name>

# 检查事件
kubectl get events -n law-oa --sort-by=.metadata.creationTimestamp
```

#### 数据库连接问题
```bash
# 检查数据库服务
kubectl get svc mysql -n law-oa
kubectl describe pod mysql-0 -n law-oa

# 测试数据库连接
kubectl exec -it mysql-0 -n law-oa -- mysql -u root -p -e "SELECT 1;"

# 检查数据库日志
kubectl logs mysql-0 -n law-oa

# 检查网络连接
kubectl exec -it <app-pod> -n law-oa -- nc -zv mysql 3306
```

#### 性能问题诊断
```bash
# 检查资源使用情况
kubectl top nodes
kubectl top pods -n law-oa

# 查看应用指标
curl http://api.lawoa.com/metrics

# 分析慢查询
kubectl exec mysql-0 -n law-oa -- mysql -u root -p -e "SHOW PROCESSLIST;"
kubectl exec mysql-0 -n law-oa -- mysql -u root -p -e "SHOW FULL PROCESSLIST;"

# 检查缓存命中率
kubectl exec redis-master-0 -n law-oa -- redis-cli info stats | grep keyspace
```

### 9.2 日志分析

#### 结构化日志查询
```bash
# 使用jq分析JSON日志
kubectl logs -f deployment/law-oa-app -n law-oa | jq 'select(.level == "error")'

# 查找特定错误
kubectl logs deployment/law-oa-app -n law-oa | grep "connection refused"

# 分析响应时间
kubectl logs deployment/law-oa-app -n law-oa | jq '.duration_ms' | awk '{sum+=$1} END {print "Average:", sum/NR}'
```

#### 日志聚合分析
```bash
# 使用Elasticsearch查询日志
curl -X GET "elasticsearch:9200/law-oa-logs/_search" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must": [
        {"term": {"level": "error"}},
        {"range": {"@timestamp": {"gte": "now-1h"}}}
      ]
    }
  },
  "sort": [{"@timestamp": {"order": "desc"}}]
}
'

# 使用Kibana仪表板
# 访问 http://grafana.lawoa.com/dashboards
```

### 9.3 性能调优

#### 应用性能调优
```bash
# 调整Go运行时参数
kubectl patch deployment law-oa-app -n law-oa -p \
'{"spec":{"template":{"spec":{"containers":[{"name":"law-oa-app","env":[{"name":"GOMAXPROCS","value":"4"},{"name":"GOGC","value":"100"}]}]}}}}'

# 调整资源限制
kubectl patch deployment law-oa-app -n law-oa -p \
'{"spec":{"template":{"spec":{"containers":[{"name":"law-oa-app","resources":{"limits":{"cpu":"4000m","memory":"4Gi"}}}]}}}}'

# 启用性能分析
kubectl exec -it <pod-name> -n law-oa -- curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
```

#### 数据库性能调优
```sql
-- 查看慢查询
SELECT * FROM mysql.slow_log ORDER BY start_time DESC LIMIT 10;

-- 分析表结构
SHOW TABLE STATUS LIKE 'law_oa%';

-- 优化表
OPTIMIZE TABLE users, clients, cases;

-- 查看索引使用情况
SHOW INDEX FROM users;

-- 分析查询执行计划
EXPLAIN SELECT * FROM users WHERE email = 'test@example.com';
```

### 9.4 故障恢复

#### 自动故障恢复
```yaml
# failure-recovery-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: failure-recovery
  namespace: law-oa
spec:
  template:
    spec:
      containers:
      - name: recovery
        image: law-registry.com/law-oa-go:2.1.0
        command:
        - /bin/sh
        - -c
        - |
          # 检查应用健康状态
          if ! curl -f http://law-oa-service:8080/health; then
            echo "Application is unhealthy, attempting recovery..."

            # 重启应用
            kubectl rollout restart deployment/law-oa-app -n law-oa

            # 等待重启完成
            kubectl rollout status deployment/law-oa-app -n law-oa --timeout=300s

            # 再次检查健康状态
            if curl -f http://law-oa-service:8080/health; then
              echo "Application recovery successful"
            else
              echo "Application recovery failed, escalating..."
              # 发送告警
              curl -X POST -H 'Content-type: application/json' \
                --data '{"text":"🚨 Critical: Application recovery failed for Law OA Go"}' \
                $SLACK_WEBHOOK_URL
            fi
          fi
      restartPolicy: OnFailure
```

---

## 10. CI/CD集成

### 10.1 GitHub Actions配置

#### 完整的CI/CD流水线
```yaml
# .github/workflows/ci-cd.yml
name: 🚀 CI/CD Pipeline

on:
  push:
    branches: [ main, develop, release/* ]
    tags: ['v*']
  pull_request:
    branches: [ main, develop ]
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deployment environment'
        required: true
        default: 'staging'
        type: choice
        options:
          - development
          - staging
          - production

env:
  REGISTRY: law-registry.com
  IMAGE_NAME: law-oa-go
  GO_VERSION: '1.23'
  NODE_VERSION: '20'

jobs:
  # 代码质量检查
  quality:
    name: 🔍 Code Quality
    runs-on: ubuntu-latest
    timeout-minutes: 30

    steps:
      - name: 📥 检出代码
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: 🐹 设置Go环境
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: 🔍 代码格式检查
        run: |
          go install mvdan.cc/gofumpt@latest
          gofumpt -w -s .
          git diff --exit-code

      - name: 📋 静态分析
        run: |
          go vet ./...
          go mod verify
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
          golangci-lint run --timeout=5m

      - name: 🔒 安全扫描
        run: |
          go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
          gosec ./...
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: 🧪 运行测试
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html

      - name: 📈 上传覆盖率报告
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          token: ${{ secrets.CODECOV_TOKEN }}

  # 多架构构建
  build:
    name: 🏗️ Multi-Arch Build
    runs-on: ubuntu-latest
    timeout-minutes: 45
    needs: quality
    strategy:
      matrix:
        platform: [linux/amd64, linux/arm64]

    outputs:
      image-digest: ${{ steps.build.outputs.digest }}
      image-tag: ${{ steps.build.outputs.tag }}

    steps:
      - name: 📥 检出代码
        uses: actions/checkout@v4

      - name: 🏷️ 设置构建元数据
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=sha,prefix={{branch}}-

      - name: 🔐 登录容器仓库
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ secrets.REGISTRY_USERNAME }}
          password: ${{ secrets.REGISTRY_PASSWORD }}

      - name: 🔧 设置QEMU
        uses: docker/setup-qemu-action@v3
        if: matrix.platform == 'linux/arm64'

      - name: 🔧 设置Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: 🏗️ 构建和推送镜像
        id: build
        uses: docker/build-push-action@v6
        with:
          context: .
          platforms: ${{ matrix.platform }}
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            BUILD_COMMIT=${{ github.sha }}
            BUILD_DATE=${{ github.event.head_commit.timestamp }}
            BUILD_TARGET=production

      - name: 🔍 安全扫描镜像
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ steps.meta.outputs.tags }}
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: 📊 上传安全扫描结果
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: 'trivy-results.sarif'

  # 部署到Staging
  deploy-staging:
    name: 🚀 Deploy to Staging
    runs-on: ubuntu-latest
    timeout-minutes: 45
    needs: [quality, build]
    if: github.ref == 'refs/heads/develop' || github.event_name == 'workflow_dispatch'
    environment: staging

    steps:
      - name: 📥 检出代码
        uses: actions/checkout@v4

      - name: 🔧 设置kubectl
        uses: azure/setup-kubectl@v4
        with:
          version: 'v1.29.0'

      - name: 🔐 配置kubeconfig
        run: |
          mkdir -p $HOME/.kube
          echo "${{ secrets.KUBE_CONFIG_STAGING }}" | base64 -d > $HOME/.kube/config

      - name: 📦 安装Helm
        uses: azure/setup-helm@v4
        with:
          version: 'v3.14.0'

      - name: 🚀 部署应用
        run: |
          helm upgrade law-oa-staging helm/law-oa-go/ \
            --install \
            --namespace law-oa-staging \
            --create-namespace \
            --wait \
            --timeout=10m \
            --set image.tag=${{ needs.build.outputs.image-tag }} \
            --set environment=staging \
            --set ingress.enabled=true \
            --set ingress.host=staging.lawoa.com

      - name: 🔍 健康检查
        run: |
          kubectl wait --for=condition=ready pod -l app=law-oa-go -n law-oa-staging --timeout=5m
          sleep 30
          curl -f https://staging.lawoa.com/health || exit 1

  # 部署到Production
  deploy-production:
    name: 🚀 Deploy to Production
    runs-on: ubuntu-latest
    timeout-minutes: 60
    needs: [quality, build]
    if: startsWith(github.ref, 'refs/tags/v') || (github.event_name == 'workflow_dispatch' && github.event.inputs.environment == 'production')
    environment: production

    steps:
      - name: 📥 检出代码
        uses: actions/checkout@v4

      - name: 🔧 设置kubectl
        uses: azure/setup-kubectl@v4
        with:
          version: 'v1.29.0'

      - name: 🔐 配置kubeconfig
        run: |
          mkdir -p $HOME/.kube
          echo "${{ secrets.KUBE_CONFIG_PRODUCTION }}" | base64 -d > $HOME/.kube/config

      - name: 🚀 蓝绿部署
        run: |
          helm upgrade law-oa-prod helm/law-oa-go/ \
            --install \
            --namespace law-oa-production \
            --create-namespace \
            --wait \
            --timeout=15m \
            --set image.tag=${{ needs.build.outputs.image-tag }} \
            --set environment=production \
            --set ingress.enabled=true \
            --set ingress.host=api.lawoa.com \
            --set ingress.tls.enabled=true

      - name: 🔍 生产健康检查
        run: |
          kubectl wait --for=condition=ready pod -l app=law-oa-go -n law-oa-production --timeout=10m
          sleep 60
          curl -f https://api.lawoa.com/health || exit 1

  # 通知
  notify:
    name: 📢 Notification
    runs-on: ubuntu-latest
    needs: [quality, build, deploy-staging, deploy-production]
    if: always()

    steps:
      - name: 💬 Slack通知
        uses: 8398a7/action-slack@v3
        if: always()
        with:
          status: ${{ job.status }}
          text: |
            *Law OA Go CI/CD Pipeline*
            *Branch*: ${{ github.ref_name }}
            *Commit*: ${{ github.sha }}
            *Status*: ${{ job.status }}
            *Build Result*: ${{ needs.build.outcome }}
            *Quality Check*: ${{ needs.quality.outcome }}
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
          webhook_type: incoming-webhook
```

### 10.2 部署策略

#### 蓝绿部署
```yaml
# blue-green-deployment.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: law-oa-app
  namespace: law-oa
spec:
  replicas: 5
  strategy:
    blueGreen:
      activeService: law-oa-active
      previewService: law-oa-preview
      autoPromotionEnabled: false
      scaleDownDelaySeconds: 30
      prePromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: law-oa-preview
      postPromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: law-oa-active
      previewReplicaCount: 2
  selector:
    matchLabels:
      app: law-oa-app
  template:
    metadata:
      labels:
        app: law-oa-app
    spec:
      containers:
      - name: law-oa-app
        image: law-registry.com/law-oa-go:2.1.0
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
```

#### 金丝雀部署
```yaml
# canary-deployment.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: law-oa-app
  namespace: law-oa
spec:
  replicas: 5
  strategy:
    canary:
      steps:
      - setWeight: 20
      - pause: {duration: 10m}
      - setWeight: 40
      - pause: {duration: 10m}
      - setWeight: 60
      - pause: {duration: 10m}
      - setWeight: 80
      - pause: {duration: 10m}
      canaryService: law-oa-canary
      stableService: law-oa-stable
      trafficRouting:
        istio:
          virtualService:
            name: law-oa-vsvc
            routes:
            - primary
          destinationRule:
            name: law-oa-dr
            subsets:
            - name: stable
            - name: canary
      analysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: law-oa-canary
        - name: prometheus-url
          value: http://prometheus:9090
        startingStep: 2
        interval: 5m
```

---

<div align="center">

**Law OA Go 企业级运维指南**

© 2025 Law OA Go. All rights reserved.

</div>