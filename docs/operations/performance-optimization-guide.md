# 性能优化指南

## 概述

本指南提供法律事务所自动化系统的性能优化策略和最佳实践。涵盖应用层、数据层、缓存层和网络层的优化技术，帮助提升系统性能和用户体验。

## 性能优化原则

### 1. 优化原则

1. **测量优先**: 优化前必须建立性能基准
2. **瓶颈导向**: 优先优化最关键的性能瓶颈
3. **渐进优化**: 采用小步快跑的优化策略
4. **平衡取舍**: 在性能、成本和复杂度之间找到平衡

### 2. 性能目标

```yaml
# 响应时间目标
response_time_targets:
  api_p50: 100ms     # 50%请求在100ms内完成
  api_p90: 300ms     # 90%请求在300ms内完成
  api_p95: 500ms     # 95%请求在500ms内完成
  api_p99: 1000ms    # 99%请求在1秒内完成

# 吞吐量目标
throughput_targets:
  requests_per_second: 1000    # 每秒处理1000个请求
  concurrent_users: 5000       # 支持5000并发用户

# 可用性目标
availability_targets:
  overall: 99.9%        # 整体可用性99.9%
  critical_apis: 99.95%  # 关键API可用性99.95%
```

## 性能分析工具

### 1. 应用性能分析

#### Go语言内置工具

```bash
# CPU性能分析
go tool pprof http://localhost:8080/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine分析
go tool pprof http://localhost:8080/debug/pprof/goroutine

# 阻塞分析
go tool pprof http://localhost:8080/debug/pprof/block
```

#### Prometheus + Grafana

```yaml
# 关键监控指标
metrics:
  - http_requests_total
  - http_request_duration_seconds
  - http_requests_in_progress
  - go_goroutines
  - go_memstats_alloc_bytes
  - go_memstats_gc_cpu_fraction
```

### 2. 数据库性能分析

#### PostgreSQL性能分析

```bash
# 慢查询分析
psql -d lawoffice_db -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# 索引使用分析
psql -d lawoffice_db -c "SELECT * FROM pg_stat_user_indexes;"

# 表统计信息
psql -d lawoffice_db -c "ANALYZE; SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del FROM pg_stat_user_tables;"

# 锁等待分析
psql -d lawoffice_db -c "SELECT * FROM pg_locks WHERE granted = false;"
```

#### EXPLAIN分析

```bash
# 执行计划分析
psql -d lawoffice_db -c "EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM cases WHERE status = 'active';"

# 详细执行计划
psql -d lawoffice_db -c "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) SELECT * FROM cases WHERE status = 'active';"
```

### 3. 缓存性能分析

#### Redis性能分析

```bash
# Redis性能统计
redis-cli info stats

# 内存使用分析
redis-cli info memory

# 慢查询分析
redis-cli SLOWLOG GET 10

# 键空间分析
redis-cli --scan --pattern "case:*" | wc -l
```

## 应用层优化

### 1. 代码优化

#### 数据库查询优化

```go
// 优化前：N+1查询问题
func GetCasesWithClient(c *gin.Context) {
    var cases []Case
    db.Find(&cases)  // 查询所有案件
    
    for _, case := range cases {
        var client Client
        db.First(&client, case.ClientID)  // 每个案件都查询一次客户端
    }
}

// 优化后：使用预加载
func GetCasesWithClient(c *gin.Context) {
    var cases []Case
    db.Preload("Client").Find(&cases)  // 一次性加载案件和客户端信息
}
```

#### 批量处理优化

```go
// 优化前：逐条处理
func ProcessCasesBatch(cases []Case) {
    for _, case := range cases {
        UpdateCase(case)  // 逐条更新
    }
}

// 优化后：批量处理
func ProcessCasesBatch(cases []Case) {
    // 批量更新
    db.Model(&Case{}).Where("id IN ?", getCaseIDs(cases)).Updates(map[string]interface{}{
        "updated_at": time.Now(),
        "status":     "processed",
    })
}
```

#### 并发处理优化

```go
// 优化前：串行处理
func ProcessMultipleRequests(requests []Request) {
    for _, req := range requests {
        ProcessSingleRequest(req)
    }
}

// 优化后：并发处理
func ProcessMultipleRequests(requests []Request) {
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10)  // 限制并发数
    
    for _, req := range requests {
        wg.Add(1)
        go func(r Request) {
            defer wg.Done()
            semaphore <- struct{}{}    // 获取信号量
            defer func() { <-semaphore }()  // 释放信号量
            
            ProcessSingleRequest(r)
        }(req)
    }
    wg.Wait()
}
```

### 2. 中间件优化

#### 认证中间件优化

```go
// 优化前：每次请求都查询数据库
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
            return
        }
        
        var user User
        if err := db.Where("token = ?", token).First(&user).Error; err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
            return
        }
        
        c.Set("user", user)
        c.Next()
    }
}

// 优化后：使用缓存
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
            return
        }
        
        // 先从缓存获取
        cacheKey := fmt.Sprintf("auth:%s", token)
        var user User
        if err := cache.Get(cacheKey, &user); err == nil {
            c.Set("user", user)
            c.Next()
            return
        }
        
        // 缓存未命中，查询数据库
        if err := db.Where("token = ?", token).First(&user).Error; err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
            return
        }
        
        // 缓存用户信息
        cache.Set(cacheKey, user, 5*time.Minute)
        c.Set("user", user)
        c.Next()
    }
}
```

#### 响应压缩优化

```go
// 启用Gzip压缩
func SetupGzipMiddleware() gin.HandlerFunc {
    return gzip.Gzip(gzip.DefaultCompression)
}

// 配置压缩级别
func SetupCompression() {
    gin.SetMode(gin.ReleaseMode)
    // 使用更高压缩级别
    gzip.DefaultCompression = gzip.BestCompression
}
```

### 3. 内存优化

#### 连接池优化

```go
// 数据库连接池优化
func SetupDatabase() {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        // 连接池配置
        ConnPool: &stdlib.ConnPool{
            Conn:     getDBConn(),
            MaxOpen:  25,     // 最大连接数
            MaxIdle:  5,      // 最大空闲连接数
            MaxLifetime: 5 * time.Minute,  // 连接最大生命周期
        },
    })
}

// HTTP客户端连接池优化
func SetupHTTPClient() *http.Client {
    return &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:        100,              // 最大空闲连接
            IdleConnTimeout:    90 * time.Second,  // 空闲连接超时
            MaxIdleConnsPerHost: 10,              // 每个主机最大空闲连接
            DisableCompression:  false,            // 启用压缩
        },
        Timeout: 30 * time.Second,  // 请求超时
    }
}
```

#### 对象池优化

```go
// 使用sync.Pool来重用对象
var requestPool = sync.Pool{
    New: func() interface{} {
        return &CaseRequest{}
    },
}

func ProcessRequest(c *gin.Context) {
    // 从池中获取对象
    req := requestPool.Get().(*CaseRequest)
    defer requestPool.Put(req)  // 使用后放回池中
    
    // 重置对象状态
    *req = CaseRequest{}
    
    // 绑定请求数据
    if err := c.ShouldBindJSON(req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 处理请求
    result := ProcessCase(req)
    c.JSON(200, result)
}
```

## 数据库优化

### 1. 查询优化

#### 索引优化

```sql
-- 创建复合索引
CREATE INDEX idx_cases_status_created_at ON cases(status, created_at);

-- 创建部分索引
CREATE INDEX idx_active_cases ON cases(created_at) WHERE status = 'active';

-- 创建表达式索引
CREATE INDEX idx_cases_search ON cases USING gin(to_tsvector('english', title || ' ' || description));

-- 分析索引使用情况
SELECT * FROM pg_stat_user_indexes WHERE schemaname = 'public';
```

#### 查询重写

```sql
-- 优化前：使用OR条件
SELECT * FROM cases WHERE status = 'active' OR priority = 'high';

-- 优化后：使用UNION ALL
SELECT * FROM cases WHERE status = 'active'
UNION ALL
SELECT * FROM cases WHERE priority = 'high' AND status != 'active';

-- 优化前：使用NOT IN
SELECT * FROM cases WHERE id NOT IN (SELECT case_id FROM case_assignments WHERE user_id = 1);

-- 优化后：使用LEFT JOIN
SELECT c.* FROM cases c
LEFT JOIN case_assignments ca ON c.id = ca.case_id AND ca.user_id = 1
WHERE ca.case_id IS NULL;
```

#### 分页优化

```sql
-- 优化前：使用OFFSET
SELECT * FROM cases ORDER BY created_at DESC LIMIT 10 OFFSET 1000;

-- 优化后：使用游标分页
SELECT * FROM cases WHERE created_at < '2024-01-01T00:00:00Z' ORDER BY created_at DESC LIMIT 10;

-- 游标分页实现
func GetCasesWithCursor(cursor string, limit int) ([]Case, string, error) {
    var cases []Case
    query := db.Model(&Case{}).Order("created_at DESC")
    
    if cursor != "" {
        query = query.Where("created_at < ?", cursor)
    }
    
    if err := query.Limit(limit + 1).Find(&cases).Error; err != nil {
        return nil, "", err
    }
    
    var nextCursor string
    if len(cases) > limit {
        nextCursor = cases[len(cases)-1].CreatedAt.Format(time.RFC3339)
        cases = cases[:limit]
    }
    
    return cases, nextCursor, nil
}
```

### 2. 数据库配置优化

#### PostgreSQL配置

```yaml
# postgresql.conf
# 内存配置
shared_buffers = 256MB           # 25% of total RAM
effective_cache_size = 1GB      # 50-75% of total RAM
work_mem = 4MB                  # Per sort/hash operation
maintenance_work_mem = 64MB     # Maintenance operations

# 连接配置
max_connections = 100            # 根据应用需要调整
max_worker_processes = 8        # CPU核心数

# 日志配置
log_statement = 'all'           # 记录所有查询
log_min_duration_statement = 1000  # 记录超过1秒的查询
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '

# 检查点配置
checkpoint_segments = 32
checkpoint_timeout = 5min
checkpoint_completion_target = 0.9
```

#### 连接池配置

```go
// 数据库连接池优化配置
database:
  host: localhost
  port: 5432
  user: lawoffice
  password: ${DB_PASSWORD}
  dbname: lawoffice_db
  ssl_mode: disable
  
  # 连接池配置
  max_connections: 25              # 最大连接数
  max_idle_connections: 5          # 最大空闲连接数
  max_lifetime: 5m                # 连接最大生命周期
  max_idle_time: 10m              # 空闲连接最大时间
  
  # 连接参数
  connect_timeout: 10s            # 连接超时
  read_timeout: 30s               # 读取超时
  write_timeout: 30s              # 写入超时
```

### 3. 数据库维护

#### 定期维护任务

```bash
# 每日VACUUM和ANALYZE
0 2 * * * /usr/bin/vacuumdb -d lawoffice_db --analyze

# 每周REINDEX
0 3 * * 0 /usr/bin/reindexdb -d lawoffice_db

# 每月表统计信息更新
0 4 1 * * /usr/bin/vacuumdb -d lawoffice_db --full --analyze
```

#### 查询优化

```bash
# 查看慢查询
SELECT query, mean_time, calls, rows 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 20;

# 查看未使用的索引
SELECT schemaname, tablename, indexname, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
WHERE idx_tup_read = 0 AND idx_tup_fetch = 0;

# 查看表膨胀
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) as index_size,
    (pg_stat_get_dead_tuple(c.oid)::float/pg_stat_get_live_tuple(c.oid)::float)*100 as dead_ratio
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE relkind = 'r' AND nspname = 'public'
ORDER BY dead_ratio DESC;
```

## 缓存优化

### 1. Redis优化

#### Redis配置优化

```yaml
# redis.conf
# 内存配置
maxmemory 1gb                          # 最大内存使用
maxmemory-policy allkeys-lru          # 内存淘汰策略

# 持久化配置
save 900 1                             # 15分钟内有1个key改变就保存
save 300 10                            # 5分钟内有10个key改变就保存
save 60 10000                          # 1分钟内有10000个key改变就保存

# 网络配置
tcp-keepalive 300                      # TCP keepalive
timeout 0                              # 客户端空闲超时

# 慢查询配置
slowlog-log-slower-than 10000          # 超过10ms的查询记录
slowlog-max-len 128                    # 最多记录128条慢查询
```

#### 缓存策略优化

```go
// 缓存配置
type CacheConfig struct {
    TTL            time.Duration `yaml:"ttl"`              // 默认TTL
    MaxSize        int           `yaml:"max_size"`         // 最大缓存数量
    CleanupInterval time.Duration `yaml:"cleanup_interval"` // 清理间隔
    Compress       bool          `yaml:"compress"`         // 是否压缩
}

// 缓存键策略
const (
    CaseKeyPrefix     = "case:"
    ClientKeyPrefix   = "client:"
    UserKeyPrefix     = "user:"
    StatsKeyPrefix    = "stats:"
    SearchKeyPrefix   = "search:"
)

// 生成缓存键
func GenerateCacheKey(prefix string, id string) string {
    return fmt.Sprintf("%s%s", prefix, id)
}

// 缓存穿透保护
func GetWithCache(key string, fetchFunc func() (interface{}, error), ttl time.Duration) (interface{}, error) {
    // 尝试从缓存获取
    var result interface{}
    if err := cache.Get(key, &result); err == nil {
        return result, nil
    }
    
    // 使用互斥锁防止缓存击穿
    mutex := &sync.Mutex{}
    mutex.Lock()
    defer mutex.Unlock()
    
    // 双重检查
    if err := cache.Get(key, &result); err == nil {
        return result, nil
    }
    
    // 从数据源获取
    result, err := fetchFunc()
    if err != nil {
        return nil, err
    }
    
    // 设置缓存
    cache.Set(key, result, ttl)
    return result, nil
}
```

### 2. 多级缓存

#### L1 + L2缓存

```go
// L1缓存（内存缓存）
var l1Cache = lru.New(1000)

// L2缓存（Redis）
type L2Cache struct {
    client *redis.Client
}

// 多级缓存实现
type MultiLevelCache struct {
    l1 *lru.Cache
    l2 *L2Cache
}

func (c *MultiLevelCache) Get(key string) (interface{}, error) {
    // 先查L1缓存
    if value, ok := c.l1.Get(key); ok {
        return value, nil
    }
    
    // 查L2缓存
    value, err := c.l2.Get(key)
    if err == nil {
        // 回填L1缓存
        c.l1.Add(key, value)
        return value, nil
    }
    
    return nil, errors.New("key not found")
}

func (c *MultiLevelCache) Set(key string, value interface{}, ttl time.Duration) error {
    // 设置L1缓存
    c.l1.Add(key, value)
    
    // 设置L2缓存
    return c.l2.Set(key, value, ttl)
}
```

#### 缓存预热

```go
// 缓存预热策略
func CacheWarmup() {
    // 预热热门案件
    popularCases := GetPopularCases(100)
    for _, case := range popularCases {
        cacheKey := GenerateCacheKey(CaseKeyPrefix, case.ID)
        cache.Set(cacheKey, case, 10*time.Minute)
    }
    
    // 预热统计数据
    stats := GetSystemStats()
    cache.Set(GenerateCacheKey(StatsKeyPrefix, "system"), stats, 5*time.Minute)
    
    // 预热用户权限
    users := GetActiveUsers()
    for _, user := range users {
        permissions := GetUserPermissions(user.ID)
        cacheKey := GenerateCacheKey(UserKeyPrefix, fmt.Sprintf("perms:%s", user.ID))
        cache.Set(cacheKey, permissions, 30*time.Minute)
    }
}
```

### 3. 缓存失效策略

```go
// 缓存失效管理
type CacheInvalidationManager struct {
    pubsub *redis.PubSub
    cache  *MultiLevelCache
}

func (m *CacheInvalidationManager) Invalidate(key string) error {
    // 发布失效消息
    err := m.cache.l2.client.Publish("cache:invalidate", key).Err()
    if err != nil {
        return err
    }
    
    // 立即失效本地缓存
    m.cache.l1.Remove(key)
    
    return nil
}

func (m *CacheInvalidationManager) ListenForInvalidations() {
    // 订阅失效频道
    ch := m.pubsub.Channel()
    
    for msg := range ch {
        key := msg.Payload
        m.cache.l1.Remove(key)
        
        // 可选：重新加载数据
        go m.ReloadData(key)
    }
}
```

## 前端优化

### 1. 静态资源优化

#### 资源压缩和合并

```javascript
// Webpack配置优化
module.exports = {
    optimization: {
        splitChunks: {
            chunks: 'all',
            minSize: 20000,
            maxSize: 244000,
            minChunks: 1,
            maxAsyncRequests: 30,
            maxInitialRequests: 30,
            automaticNameDelimiter: '~',
            enforceSizeThreshold: 50000,
            cacheGroups: {
                vendors: {
                    test: /[\\/]node_modules[\\/]/,
                    priority: -10
                },
                default: {
                    minChunks: 2,
                    priority: -20,
                    reuseExistingChunk: true
                }
            }
        },
        minimizer: [
            new TerserPlugin({
                parallel: true,
                extractComments: false,
                terserOptions: {
                    compress: {
                        drop_console: true,
                        drop_debugger: true
                    }
                }
            })
        ]
    }
};
```

#### CDN配置

```javascript
// CDN配置
const CDN_BASE_URL = 'https://cdn.lawoffice.com';

// 静态资源URL配置
const assetPath = (path) => {
    return `${CDN_BASE_URL}/${path}?v=${process.env.ASSET_VERSION}`;
};

// 图片CDN配置
const imageUrl = (path, width, height) => {
    return `${CDN_BASE_URL}/images/${path}?w=${width}&h=${height}&q=80`;
};
```

### 2. API调用优化

#### 请求合并

```javascript
// 优化前：多个独立请求
async function loadDashboard() {
    const [cases, clients, stats] = await Promise.all([
        fetch('/api/v1/cases'),
        fetch('/api/v1/clients'),
        fetch('/api/v1/stats')
    ]);
    
    return {
        cases: await cases.json(),
        clients: await clients.json(),
        stats: await stats.json()
    };
}

// 优化后：单个批量请求
async function loadDashboard() {
    const response = await fetch('/api/v1/dashboard', {
        method: 'POST',
        body: JSON.stringify({
            include: ['cases', 'clients', 'stats']
        })
    });
    
    return response.json();
}
```

#### 请求缓存

```javascript
// 请求缓存实现
const requestCache = new Map();

async function cachedFetch(url, options = {}) {
    const cacheKey = JSON.stringify({ url, options });
    
    if (requestCache.has(cacheKey)) {
        return requestCache.get(cacheKey);
    }
    
    const promise = fetch(url, options);
    requestCache.set(cacheKey, promise);
    
    try {
        const response = await promise;
        
        // 缓存成功响应
        if (response.ok) {
            setTimeout(() => {
                requestCache.delete(cacheKey);
            }, 5000); // 5秒后清除缓存
        }
        
        return response;
    } catch (error) {
        requestCache.delete(cacheKey);
        throw error;
    }
}
```

### 3. 渲染优化

#### 虚拟滚动

```javascript
// 大数据列表虚拟滚动
import { FixedSizeList } from 'react-window';

const CaseList = ({ cases }) => (
    <FixedSizeList
        height={600}
        width="100%"
        itemCount={cases.length}
        itemSize={80}
    >
        {({ index, style }) => (
            <div style={style}>
                <CaseItem case={cases[index]} />
            </div>
        )}
    </FixedSizeList>
);
```

#### 懒加载

```javascript
// 路由懒加载
const Dashboard = React.lazy(() => import('./Dashboard'));
const CaseManagement = React.lazy(() => import('./CaseManagement'));

// 图片懒加载
const LazyImage = ({ src, alt, ...props }) => {
    const [loaded, setLoaded] = useState(false);
    const imgRef = useRef();
    
    useEffect(() => {
        const observer = new IntersectionObserver(
            ([entry]) => {
                if (entry.isIntersecting) {
                    setLoaded(true);
                    observer.unobserve(imgRef.current);
                }
            },
            { threshold: 0.1 }
        );
        
        observer.observe(imgRef.current);
        return () => observer.disconnect();
    }, []);
    
    return (
        <div ref={imgRef} style={{ minHeight: '200px' }}>
            {loaded ? (
                <img src={src} alt={alt} {...props} />
            ) : (
                <div className="image-placeholder">Loading...</div>
            )}
        </div>
    );
};
```

## 网络优化

### 1. HTTP/2优化

```go
// HTTP/2服务器配置
func SetupHTTP2Server() *http.Server {
    return &http.Server{
        Addr: ":443",
        Handler: app,
        TLSConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
            CurvePreferences: []tls.CurveID{
                tls.CurveP256,
                tls.X25519,
            },
        },
    }
}
```

### 2. 连接复用

```javascript
// HTTP客户端连接复用
const axios = require('axios');

const apiClient = axios.create({
    baseURL: process.env.API_BASE_URL,
    timeout: 30000,
    maxContentLength: 10 * 1024 * 1024, // 10MB
    maxBodyLength: 10 * 1024 * 1024,    // 10MB
    httpAgent: new http.Agent({
        keepAlive: true,
        maxSockets: 50,
        maxFreeSockets: 10,
        keepAliveMsecs: 30000,
    }),
    httpsAgent: new https.Agent({
        keepAlive: true,
        maxSockets: 50,
        maxFreeSockets: 10,
        keepAliveMsecs: 30000,
    }),
});
```

### 3. 压缩优化

```go
// Brotli压缩配置
func SetupBrotliCompression() gin.HandlerFunc {
    return brotli.Brotli(brotli.DefaultCompression)
}

// 动态压缩策略
func SetupAdaptiveCompression() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查客户端是否支持压缩
        acceptEncoding := c.GetHeader("Accept-Encoding")
        
        if strings.Contains(acceptEncoding, "br") {
            c.Header("Content-Encoding", "br")
        } else if strings.Contains(acceptEncoding, "gzip") {
            c.Header("Content-Encoding", "gzip")
        }
        
        c.Next()
    }
}
```

## 性能监控和调优

### 1. 实时性能监控

```go
// 性能监控中间件
func PerformanceMonitor() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 记录请求开始
        c.Set("request_start", start)
        
        c.Next()
        
        // 计算请求耗时
        duration := time.Since(start)
        
        // 记录指标
        metrics.RecordRequestDuration(c.Request.Method, c.Request.URL.Path, duration)
        
        // 记录慢请求
        if duration > 500*time.Millisecond {
            log.Printf("Slow request: %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
        }
        
        // 添加响应头
        c.Header("X-Response-Time", duration.String())
    }
}
```

### 2. 性能基准测试

```go
// 基准测试示例
func BenchmarkGetCases(b *testing.B) {
    // 设置测试环境
    setupTestDB()
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        req, _ := http.NewRequest("GET", "/api/v1/cases", nil)
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        if w.Code != http.StatusOK {
            b.Errorf("Expected status 200, got %d", w.Code)
        }
    }
}
```

### 3. 负载测试

```yaml
# k6负载测试脚本
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '2m', target: 100 },  // 爬坡到100用户
        { duration: '5m', target: 100 },  // 稳定100用户
        { duration: '2m', target: 200 },  // 爬坡到200用户
        { duration: '5m', target: 200 },  // 稳定200用户
        { duration: '2m', target: 0 },    // 降级到0用户
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95%的请求在500ms内完成
        http_req_failed: ['rate<0.01'],  // 错误率小于1%
    },
};

export default function () {
    let res = http.get('http://localhost:8080/api/v1/cases');
    
    check(res, {
        'status was 200': (r) => r.status == 200,
        'response time was < 500ms': (r) => r.timings.duration < 500,
    });
    
    sleep(1);
}
```

## 性能优化checklist

### 开发阶段

- [ ] 使用性能分析工具识别瓶颈
- [ ] 实现数据库查询优化和索引
- [ ] 添加适当的缓存策略
- [ ] 实现连接池和资源复用
- [ ] 优化前端资源加载和渲染
- [ ] 实现代码分割和懒加载

### 测试阶段

- [ ] 进行负载测试和压力测试
- [ ] 监控内存使用和GC行为
- [ ] 测试数据库连接池性能
- [ ] 验证缓存命中率
- [ ] 测试并发处理能力
- [ ] 验证性能指标达标

### 生产环境

- [ ] 配置实时性能监控
- [ ] 设置性能告警阈值
- [ ] 定期进行性能评估
- [ ] 实施自动化性能测试
- [ ] 建立性能基准线
- [ ] 制定性能优化计划

## 性能优化最佳实践

### 1. 数据库优化

- 使用适当的索引和查询优化
- 实施数据库连接池
- 定期进行数据库维护
- 监控慢查询和性能指标

### 2. 缓存优化

- 实施多级缓存策略
- 使用合适的缓存失效策略
- 监控缓存命中率
- 实施缓存预热机制

### 3. 应用优化

- 使用并发和异步处理
- 优化内存使用和垃圾回收
- 实施连接池和资源复用
- 优化代码结构和算法

### 4. 前端优化

- 优化静态资源加载
- 实施懒加载和虚拟滚动
- 优化API调用策略
- 使用CDN加速

### 5. 监控和调优

- 建立完整的性能监控体系
- 设置合理的告警阈值
- 定期进行性能评估
- 持续优化和改进

## 总结

性能优化是一个持续的过程，需要建立完整的监控、测试和优化体系。通过实施本指南中的优化策略，可以显著提升法律事务所自动化系统的性能和用户体验。

关键要点：
- 建立性能基准和监控体系
- 采用分层优化策略
- 优先解决关键性能瓶颈
- 持续监控和调优
- 建立性能优化流程和最佳实践