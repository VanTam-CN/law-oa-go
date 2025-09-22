# Law OA Go - 统一API规范设计

## 📋 设计原则

### 核心原则
- **一致性**: 所有API响应格式统一
- **标准化**: 遵循OpenAPI 3.0和REST最佳实践
- **类型安全**: 强类型检查，编译时错误发现
- **可观测性**: 完善的日志、追踪和监控
- **可维护性**: 清晰的代码结构和文档

### Go 1.23+ 现代标准
- 使用 generics 减少代码重复
- 采用 context 超时控制
- 实现结构化日志记录
- 遵循 Go 官方错误处理规范

## 🏗️ API架构设计

### 响应格式统一化

#### 标准成功响应
```json
{
  "success": true,
  "data": {},
  "meta": {
    "timestamp": "2025-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

#### 标准分页响应
```json
{
  "success": true,
  "data": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  },
  "meta": {
    "timestamp": "2025-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

#### 标准错误响应
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": "邮箱格式不正确",
    "context": {
      "field": "email",
      "value": "invalid-email"
    },
    "suggestions": [
      "请检查邮箱格式",
      "确保邮箱地址有效"
    ]
  },
  "meta": {
    "timestamp": "2025-01-01T00:00:00Z",
    "request_id": "req_123456",
    "version": "1.0.0"
  }
}
```

### API版本控制策略

#### URL版本控制
```
/api/v1/          # 当前稳定版本
/api/v2/          # 未来版本
/api/latest/      # 最新版本（生产慎用）
```

#### 兼容性处理
- 同一主版本内保持向后兼容
- 废弃API提供6个月过渡期
- 使用Deprecation头标识废弃接口

## 🔧 中间件架构重新设计

### 请求生命周期中间件链
```go
Request → [CORS] → [RateLimit] → [RequestID] → [Timeout] → [Auth] → [Role] → [Validation] → Handler
```

### 中间件职责划分

#### 1. 基础中间件
- **RequestID**: 生成唯一请求标识
- **Timeout**: 请求超时控制（默认30s）
- **CORS**: 跨域请求处理
- **Security**: 安全头部设置

#### 2. 业务中间件  
- **Auth**: JWT身份验证
- **Role**: 基于角色的访问控制
- **Validation**: 请求参数验证
- **Audit**: 操作审计日志

#### 3. 监控中间件
- **Metrics**: Prometheus指标收集
- **Logging**: 结构化请求日志
- **Tracing**: 分布式追踪

## 📦 统一Handler框架

### 基础Handler接口
```go
type BaseHandler interface {
    HandleRequest(c *gin.Context) *APIResponse
    ValidateRequest(c *gin.Context) error
    GetTimeout() time.Duration
}
```

### 通用CRUD Handler
```go
type CRUDHandler[T, ID any] struct {
    service CRUDService[T, ID]
    resource string
}

func (h *CRUDHandler[T, ID]) Create(c *gin.Context)
func (h *CRUDHandler[T, ID]) Get(c *gin.Context) 
func (h *CRUDHandler[T, ID]) Update(c *gin.Context)
func (h *CRUDHandler[T, ID]) Delete(c *gin.Context)
func (h *CRUDHandler[T, ID]) List(c *gin.Context)
```

### 专用Handler继承
```go
type UserHandler struct {
    CRUDHandler[User, uint]
    userService UserService
}

type CaseHandler struct {
    CRUDHandler[Case, uint] 
    caseService CaseService
}
```

## 🎯 路由设计规范

### RESTful URL规范
```
# 用户管理
GET    /api/v1/users                    # 获取用户列表
POST   /api/v1/users                    # 创建用户
GET    /api/v1/users/{id}               # 获取用户详情
PUT    /api/v1/users/{id}               # 更新用户
DELETE /api/v1/users/{id}               # 删除用户

# 客户管理  
GET    /api/v1/clients                  # 获取客户列表
POST   /api/v1/clients                  # 创建客户
GET    /api/v1/clients/{id}             # 获取客户详情
PUT    /api/v1/clients/{id}             # 更新客户
DELETE /api/v1/clients/{id}             # 删除客户

# 案件管理
GET    /api/v1/cases                    # 获取案件列表
POST   /api/v1/cases                    # 创建案件
GET    /api/v1/cases/{id}               # 获取案件详情
PUT    /api/v1/cases/{id}               # 更新案件
DELETE /api/v1/cases/{id}               # 删除案件
POST   /api/v1/cases/{id}/assign        # 分配律师
PUT    /api/v1/cases/{id}/status        # 更新状态

# 统计信息
GET    /api/v1/users/stats              # 用户统计
GET    /api/v1/clients/stats            # 客户统计  
GET    /api/v1/cases/stats              # 案件统计
```

### 查询参数标准化
```go
// 分页参数
page=1&pageSize=20

// 过滤参数
status=active&role=admin&priority=high

// 搜索参数
search=keyword&fields=name,email

// 排序参数  
sortBy=createdAt&sortOrder=desc

// 时间范围
startDate=2025-01-01&endDate=2025-12-31
```

## 🛡️ 安全设计

### 认证授权
```go
// JWT Token结构
type Claims struct {
    UserID    uint   `json:"user_id"`
    Username  string `json:"username"`
    Role      string `json:"role"`
    ClientID  string `json:"client_id"`
    DeviceID  string `json:"device_id"`
    IssuedAt  int64  `json:"iat"`
    ExpiresAt int64  `json:"exp"`
}
```

### 权限控制矩阵
```
/admin/*     - admin角色
/clients/*   - admin, lawyer, staff
/cases/*     - admin, lawyer, staff  
/users/*     - admin
/stats/*     - admin, manager
```

### 请求限流
```go
// 基于IP的限流
RateLimitByIP(100 requests/minute)

// 基于用户的限流  
RateLimitByUser(1000 requests/hour)

// 基于API的限流
RateLimitByEndpoint(50 requests/minute)
```

## 📊 监控和可观测性

### 指标收集
```go
// HTTP请求指标
http_requests_total{method, endpoint, status}
http_request_duration_seconds{method, endpoint}
http_response_size_bytes{method, endpoint}

// 业务指标
active_users_total
cases_created_total
clients_acquired_total
```

### 结构化日志
```json
{
  "timestamp": "2025-01-01T00:00:00Z",
  "level": "INFO",
  "request_id": "req_123456",
  "method": "POST",
  "endpoint": "/api/v1/users",
  "user_id": "user_123",
  "client_ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "duration_ms": 156,
  "status_code": 200,
  "message": "User created successfully"
}
```

### 分布式追踪
```go
// OpenTelemetry集成
tracer := otel.Tracer("law-oa-api")
ctx, span := tracer.Start(c.Request.Context(), "CreateUser")
defer span.End()
```

## 🔧 配置管理

### 环境配置
```go
type APIConfig struct {
    Version       string        `env:"API_VERSION" envDefault:"1.0.0"`
    Timeout       time.Duration `env:"API_TIMEOUT" envDefault:"30s"`
    RateLimit     int           `env:"RATE_LIMIT" envDefault:"100"`
    PageSize      int           `env:"PAGE_SIZE" envDefault:"20"`
    MaxPageSize   int           `env:"MAX_PAGE_SIZE" envDefault:"100"`
    EnableSwagger bool          `env:"ENABLE_SWAGGER" envDefault:"true"`
}
```

### 特性开关
```go
type FeatureFlags struct {
    EnableNewAuthSystem  bool `env:"ENABLE_NEW_AUTH" envDefault:"false"`
    EnableAdvancedSearch bool `env:"ENABLE_ADVANCED_SEARCH" envDefault:"false"`
    EnableAuditLogging   bool `env:"ENABLE_AUDIT_LOGGING" envDefault:"true"`
}
```

## 📋 实施计划

### 第一阶段：基础设施
1. 创建统一的API响应结构
2. 实现新的中间件链
3. 建立基础Handler框架
4. 配置监控和日志系统

### 第二阶段：重构核心API
1. 重构用户管理API
2. 重构客户管理API
3. 重构案件管理API
4. 统一认证授权系统

### 第三阶段：优化和测试
1. 添加API测试覆盖
2. 性能优化和压力测试
3. 完善文档和示例
4. 部署和监控配置

## 🎯 成功指标

### 技术指标
- API响应时间 < 200ms (P95)
- 错误率 < 0.1%
- 代码覆盖率 > 80%
- 所有API响应格式100%一致

### 业务指标  
- 开发效率提升 50%
- Bug修复时间减少 60%
- 新功能开发时间减少 40%
- API文档覆盖率 100%

---

这个统一API规范将彻底解决当前项目的混乱状态，建立一个现代化、可维护、高性能的API系统。