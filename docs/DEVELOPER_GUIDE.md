# 律师事务所OA系统开发者指南

## 开发环境搭建

### 系统要求

- Go 1.25+
- Node.js 18+
- MySQL 8.0+ 或 PostgreSQL 15+
- Redis 6.0+
- Git

### 本地开发环境

1. **克隆项目**
```bash
git clone https://github.com/VanTam-CN/law-oa-go.git
cd law-oa-go
```

2. **后端环境配置**
```bash
# 复制配置文件
cp .env.example .env

# 编辑配置文件
vim .env
```

**环境变量配置 (`.env`)**:
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=law_oa

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT配置
JWT_SECRET=your-secret-key
JWT_EXPIRE_HOURS=24

# API配置
API_PORT=8080
API_HOST=localhost
API_VERSION=v1

# 环境配置
GIN_MODE=debug
LOG_LEVEL=debug

# MCP配置 (冲突检测)
MCP_API_KEY=your-mcp-key
MCP_API_URL=https://mcp.example.com
```

3. **数据库初始化**
```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS law_oa CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 运行迁移
go run ./cmd/migrate -migrations ./migrations -command up
```

4. **前端环境配置**
```bash
cd frontend

# 安装依赖
npm install

# 复制配置文件
cp .env.example .env.local

# 启动开发服务器
npm run dev
```

### IDE配置

#### VSCode配置

**推荐扩展:**
- Go (官方)
- Thunder Client (API测试)
- Database Client
- Redis Client
- GitLens

**工作区设置 (`.vscode/settings.json`)**:
```json
{
  "go.useLanguageServer": true,
  "go.toolsEnvVars": {
    "GO111MODULE": "on"
  },
  "files.exclude": [
    "**/node_modules",
    "**/vendor",
    "**/.git"
  ]
}
```

## 项目结构

```
law-oa-go/
├── main.go                 # 默认应用入口
├── cmd/                    # 辅助命令和备用入口
│   ├── server/
│   └── migrate/
├── internal/               # 内部包
│   ├── api/               # API helper / CRUD 基类
│   ├── handlers/          # 处理器层
│   ├── services/          # 业务逻辑层
│   ├── repositories/      # 数据访问层
│   ├── models/            # 数据模型
│   ├── middleware/        # 中间件
│   ├── common/            # 通用工具
│   ├── errors/            # 错误处理
│   ├── router/            # 路由注册
│   └── config/            # 配置管理
├── frontend/              # 前端代码
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── services/
│   │   ├── utils/
│   │   └── styles/
│   ├── public/
│   └── package.json
├── docs/                   # 文档
├── tests/                  # 测试文件
├── scripts/               # 脚本文件
├── configs/               # 配置文件
└── migrations/             # 数据库迁移
```

## 开发工作流

### 1. 功能开发流程

#### 创建新功能模块

1. **定义数据模型** (`internal/models/`)
```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type NewFeature struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Name      string         `json:"name" gorm:"not null;size:255"`
    Status    string         `json:"status" gorm:"not null;default:active"`
    CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (nf *NewFeature) TableName() string {
    return "new_features"
}

func (nf *NewFeature) Validate() error {
    if nf.Name == "" {
        return fmt.Errorf("名称不能为空")
    }
    return nil
}
```

2. **实现数据仓库** (`internal/repositories/`)
```go
package repositories

import (
    "context"
    "law-oa-go/internal/models"
    "gorm.io/gorm"
)

type NewFeatureRepository struct {
    db *gorm.DB
}

func NewNewFeatureRepository(db *gorm.DB) *NewFeatureRepository {
    return &NewFeatureRepository{db: db}
}

func (r *NewFeatureRepository) Create(ctx context.Context, feature *models.NewFeature) error {
    return r.db.WithContext(ctx).Create(feature).Error
}

func (r *NewFeatureRepository) GetByID(ctx context.Context, id uint) (*models.NewFeature, error) {
    var feature models.NewFeature
    err := r.db.WithContext(ctx).First(&feature, id).Error
    if err != nil {
        return nil, err
    }
    return &feature, nil
}

func (r *NewFeatureRepository) List(ctx context.Context, page, pageSize int) ([]*models.NewFeature, int64, error) {
    var features []*models.NewFeature
    var total int64

    offset := (page - 1) * pageSize

    if err := r.db.WithContext(ctx).Model(&models.NewFeature{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&features).Error; err != nil {
        return nil, 0, err
    }

    return features, total, nil
}
```

3. **实现业务服务** (`internal/services/`)
```go
package services

import (
    "context"
    "errors"
    "law-oa-go/internal/models"
    "law-oa-go/internal/repositories"
)

type NewFeatureService struct {
    repo *repositories.NewFeatureRepository
}

func NewNewFeatureService(repo *repositories.NewFeatureRepository) *NewFeatureService {
    return &NewFeatureService{repo: repo}
}

func (s *NewFeatureService) Create(ctx context.Context, req *CreateNewFeatureRequest) (*models.NewFeature, error) {
    // 验证请求
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // 创建实体
    feature := &models.NewFeature{
        Name:   req.Name,
        Status: req.Status,
    }

    // 保存到数据库
    if err := s.repo.Create(ctx, feature); err != nil {
        return nil, err
    }

    return feature, nil
}

func (s *NewFeatureService) GetByID(ctx context.Context, id uint) (*models.NewFeature, error) {
    return s.repo.GetByID(ctx, id)
}
```

4. **实现处理器** (`internal/handlers/`)
```go
package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "law-oa-go/internal/common"
    "law-oa-go/internal/services"
)

type NewFeatureHandler struct {
    service *services.NewFeatureService
}

func NewNewFeatureHandler(service *services.NewFeatureService) *NewFeatureHandler {
    return &NewFeatureHandler{service: service}
}

// CreateNewFeature 创建新功能
// @Summary 创建新功能
// @Description 创建一个新的功能模块
// @Tags new-feature
// @Accept json
// @Produce json
// @Param request body CreateNewFeatureRequest true "创建请求"
// @Success 200 {object} common.UnifiedResponse
// @Failure 400 {object} common.UnifiedResponse
// @Router /new-features [post]
func (h *NewFeatureHandler) CreateNewFeature(c *gin.Context) {
    var req CreateNewFeatureRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.BadRequest(c, "请求格式错误: "+err.Error())
        return
    }

    result, err := h.service.Create(c.Request.Context(), &req)
    if err != nil {
        common.InternalServerError(c, "创建失败: "+err.Error())
        return
    }

    common.Success(c, result)
}
```

5. **注册路由** (`internal/router/router.go`)
```go
// 在Init函数中添加
newFeatureHandler := handlers.NewNewFeatureHandler(newFeatureService)

newFeatureV1 := apiV1.Group("/new-features")
{
    newFeatureV1.POST("", newFeatureHandler.CreateNewFeature)
    newFeatureV1.GET("", newFeatureHandler.ListNewFeatures)
    newFeatureV1.GET("/:id", newFeatureHandler.GetNewFeature)
    newFeatureV1.PUT("/:id", newFeatureHandler.UpdateNewFeature)
    newFeatureV1.DELETE("/:id", newFeatureHandler.DeleteNewFeature)
}
```

### 2. 测试开发

#### 单元测试
```go
package services

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "law-oa-go/internal/models"
)

func TestNewFeatureService_Create(t *testing.T) {
    // 创建模拟仓库
    mockRepo := &MockNewFeatureRepository{}
    service := NewNewFeatureService(mockRepo)

    // 准备测试数据
    req := &CreateNewFeatureRequest{
        Name:   "测试功能",
        Status: "active",
    }

    // 模拟仓库响应
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.NewFeature")).Return(nil)

    // 执行测试
    result, err := service.Create(context.Background(), req)

    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "测试功能", result.Name)
    assert.Equal(t, "active", result.Status)

    mockRepo.AssertExpectations(t)
}
```

#### 集成测试
```go
package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type NewFeatureIntegrationTestSuite struct {
    suite.Suite
    router *gin.Engine
}

func (suite *NewFeatureIntegrationTestSuite) SetupTest() {
    suite.router = setupTestRouter()
}

func (suite *NewFeatureIntegrationTestSuite) TestCreateNewFeature() {
    // 准备请求数据
    reqBody := map[string]interface{}{
        "name":   "集成测试功能",
        "status": "active",
    }

    jsonData, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/new-features", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")

    // 执行请求
    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)

    // 验证响应
    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
}
```

### 3. 数据库迁移

创建迁移文件 `migrations/001_create_new_features_table.go`:
```go
package migrations

import (
    "github.com/go-gormigrate/gorm"
    "law-oa-go/internal/models"
)

func CreateNewFeaturesTable(db *gorm.DB) error {
    return db.AutoMigrate(&models.NewFeature{})
}
```

## 代码规范

### Go代码规范

1. **命名规范**
   - 包名：小写，单词间用下划线
   - 结构体：PascalCase
   - 函数：驼峰命名
   - 常量：全大写，下划线分隔

2. **注释规范**
   - 公开函数必须有注释
   - 复杂逻辑必须添加注释
   - 使用GoDoc格式

```go
// Package services 提供业务逻辑服务
package services

// NewFeatureService 新功能服务
// 负责处理新功能相关的业务逻辑
type NewFeatureService struct {
    repo *repositories.NewFeatureRepository
}

// Create 创建新功能
// ctx: 请求上下文
// req: 创建请求参数
// 返回: 创建的功能实体和可能的错误
func (s *NewFeatureService) Create(ctx context.Context, req *CreateNewFeatureRequest) (*models.NewFeature, error) {
    // 实现逻辑
}
```

3. **错误处理**
   - 使用自定义错误类型
   - 提供详细的错误信息
   - 记录错误日志

```go
// 使用统一错误处理
if err := validateInput(req); err != nil {
    return nil, errors.NewValidationError("input_validation", "输入验证失败", err.Error())
}
```

### 前端代码规范

1. **组件命名**
   - 组件文件：PascalCase
   - 组件变量：camelCase
   - CSS类：kebab-case

2. **类型定义**
   - 接口：PascalCase + Props
   - 类型：使用TypeScript

```typescript
// 组件接口定义
interface NewFeatureFormProps {
  onSubmit: (data: NewFeatureData) => void;
  initialValues?: Partial<NewFeatureData>;
}

// 类型定义
interface NewFeatureData {
  name: string;
  status: 'active' | 'inactive';
  description?: string;
}
```

3. **状态管理**
   - 使用自定义Hooks
   - 保持状态最小化
   - 异步状态处理

```typescript
// 自定义Hook
const useNewFeature = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<NewFeatureData | null>(null);
  const [error, setError] = useState<string | null>(null);

  const createNewFeature = async (formData: NewFeatureData) => {
    try {
      setLoading(true);
      const response = await api.newFeatures.create(formData);
      setData(response.data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return {
    createNewFeature,
    loading,
    data,
    error,
  };
};
```

## 性能优化

### 数据库优化

1. **索引策略**
   - 为查询字段添加索引
   - 避免过多索引
   - 定期分析查询性能

2. **查询优化**
   - 使用预加载减少N+1问题
   - 实现分页查询
   - 使用连接池

3. **缓存策略**
   - Redis缓存热点数据
   - 应用层缓存
   - CDN缓存静态资源

```go
// 优化查询示例
func (s *CaseService) ListCases(ctx context.Context, req *CaseListRequest) ([]*models.CaseResponse, int64, error) {
    query := s.db.WithContext(ctx).
        Preload("Client").
        Preload("Lawyer").
        Model(&models.Case{})

    // 添加筛选条件
    if req.Status != "" {
        query = query.Where("status = ?", req.Status)
    }

    // 使用索引优化的排序
    query = query.Order("CASE status WHEN 'pending' THEN 1 WHEN 'active' THEN 2 ELSE 3 END, created_at DESC")

    // 执行查询
    // ...
}
```

### API优化

1. **响应压缩**
   - 启用Gzip压缩
   - 压缩JSON响应
   - 缓存策略

2. **限流保护**
   - 实现速率限制
   - 用户级别限流
   - IP级别限流

3. **并发控制**
   - 连接池管理
   - 请求队列
   - 超时控制

```go
// 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        userID := c.GetString("user_id")
        key := fmt.Sprintf("rate_limit:user:%s", userID)

        if !limiter.Allow(key, 100, time.Minute) {
            c.JSON(429, gin.H{
                "error": "请求频率超限",
                "message": "请稍后再试",
            })
            c.Abort()
            return
        }
        c.Next()
    })
}
```

## 部署指南

### Docker部署

1. **创建Dockerfile**
```dockerfile
# 多阶段构建
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# 运行时镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./main"]
```

2. **docker-compose配置**
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      - mysql
      - redis

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: law_oa
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"

  redis:
    image: redis:6.2-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

volumes:
  mysql_data:
  redis_data:
```

### 生产部署

1. **环境配置**
```bash
# 生产环境配置
export GIN_MODE=release
export LOG_LEVEL=info
export DB_HOST=prod-db.example.com
export REDIS_HOST=prod-redis.example.com
```

2. **健康检查**
```go
// 健康检查端点
func (h *HealthHandler) Check(c *gin.Context) {
    status := "healthy"

    // 检查数据库连接
    if sqlDB, err := h.db.DB(); err != nil {
        status = "unhealthy"
    }

    // 检查Redis连接
    if h.redisClient != nil {
        if _, err := h.redisClient.Ping(context.Background()).Result(); err != nil {
            status = "degraded"
        }
    }

    c.JSON(200, gin.H{
        "status":    status,
        "timestamp": time.Now(),
        "services": gin.H{
            "database": "ok",
            "redis":    "ok",
        },
    })
}
```

3. **监控配置**
```go
// Prometheus指标
var (
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP请求持续时间",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )

    requestCount = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "HTTP请求总数",
        },
        []string{"method", "endpoint", "status"},
    )
)
```

## 调试技巧

### 1. 日志记录

```go
// 结构化日志
logger := slog.With(
    "user_id", userID,
    "request_id", requestID,
    "method", c.Request.Method,
    "path", c.Request.URL.Path,
)

logger.Info("处理请求",
    "duration", time.Since(start),
    "status", statusCode,
)
```

### 2. 性能分析

```go
// 添加性能分析中间件
func PerformanceMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        start := time.Now()

        c.Next()

        duration := time.Since(start)

        // 记录慢查询
        if duration > 1*time.Second {
            logger.Warn("慢请求",
                "method", c.Request.Method,
                "path", c.Request.URL.Path,
                "duration", duration,
            )
        }
    })
}
```

### 3. 调试工具

```go
// 调试中间件
func DebugMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        if os.Getenv("DEBUG") == "true" {
            // 记录请求详情
            logRequestDetails(c)
        }
        c.Next()
    })
}

func logRequestDetails(c *gin.Context) {
    body, _ := c.GetRawData()

    logger.Debug("请求详情",
        "headers", c.Request.Header,
        "query", c.Request.URL.Query(),
        "body", string(body),
    )
}
```

## 安全最佳实践

### 1. 输入验证

```go
// 严格验证输入
func validateCreateCaseInput(req *CreateCaseRequest) error {
    if strings.TrimSpace(req.Title) == "" {
        return errors.NewValidationError("title_required", "案件标题不能为空")
    }

    if len(req.Title) > 200 {
        return errors.NewValidationError("title_too_long", "案件标题不能超过200个字符")
    }

    if req.ClientID == 0 {
        return errors.NewValidationError("client_required", "客户ID不能为空")
    }

    return nil
}
```

### 2. 权限控制

```go
// RBAC中间件
func RequirePermission(permission string) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            c.JSON(401, gin.H{"error": "未认证"})
            c.Abort()
            return
        }

        // 检查权限
        if !hasPermission(userID, permission) {
            c.JSON(403, gin.H{"error": "权限不足"})
            c.Abort()
            return
        }

        c.Next()
    })
}
```

### 3. SQL注入防护

```go
// 使用参数化查询
func (r *CaseRepository) FindByStatus(status string) ([]*models.Case, error) {
    var cases []*models.Case
    err := r.db.Where("status = ?", status).Find(&cases).Error
    return cases, err
}
```

这个开发者指南提供了完整的开发流程、代码规范、性能优化、部署和调试技巧。建议开发团队严格按照这些规范进行开发，以确保代码质量和系统稳定性。
