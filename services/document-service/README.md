# 独立文档管理系统 (Document Management Service)

## 概述

这是一个独立可复用的文档管理微服务，专为律师事务所和其他企业级应用设计。该服务提供完整的文档生命周期管理功能，包括版本控制、全文搜索、权限控制、在线预览编辑等核心能力。

## 核心特性

- 🔄 **文档版本管理** - 全量保存、版本对比、历史追踪
- 🔍 **智能全文搜索** - Elasticsearch引擎、OCR文字识别
- 🛡️ **高级权限控制** - Casbin ABAC模型、分级保密
- 👁️ **在线预览编辑** - 多格式支持、富文本编辑
- ⚡ **批量操作自动化** - 消息队列、工作流引擎
- 🌐 **多租户支持** - 支持多个业务系统接入
- 🔐 **安全合规** - 审计追踪、数据加密、访问控制

## 技术架构

### 服务架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Mobile App    │    │   Other Systems │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   API Gateway   │
                    │   (统一入口)      │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Auth Service   │
                    │  (认证授权)      │
                    └─────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Document Service │    │  Search Service │    │ Permission Svc  │
│   (文档核心)      │    │   (搜索服务)     │    │   (权限服务)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      MySQL      │    │  Elasticsearch  │    │      Redis      │
│     (元数据)      │    │   (搜索索引)     │    │    (缓存)       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Object Storage │
                    │   MinIO/S3      │
                    │   (文件存储)     │
                    └─────────────────┘
```

### 核心组件

1. **Document Service** - 文档CRUD、版本管理、元数据处理
2. **Search Service** - 全文搜索、OCR处理、智能推荐
3. **Permission Service** - 权限控制、访问验证、审计日志
4. **Preview Service** - 文档预览、格式转换、在线编辑
5. **Workflow Service** - 工作流编排、批量处理、自动化任务

## 项目结构

```
document-service/
├── cmd/
│   └── server/
│       └── main.go                    # 服务入口点
├── internal/
│   ├── api/
│   │   ├── handlers/                  # HTTP处理器
│   │   ├── middleware/                # 中间件
│   │   └── routes/                    # 路由定义
│   ├── services/                     # 业务逻辑服务
│   │   ├── document/
│   │   ├── version/
│   │   ├── search/
│   │   ├── permission/
│   │   └── preview/
│   ├── repositories/                  # 数据访问层
│   ├── models/                        # 数据模型
│   ├── config/                        # 配置管理
│   └── utils/                         # 工具函数
├── pkg/                              # 公共库
├── api/                              # API定义
│   └── openapi/                       # OpenAPI规范
├── scripts/                          # 脚本文件
├── configs/                          # 配置文件
├── docker/                           # Docker配置
├── migrations/                       # 数据库迁移
├── tests/                            # 测试文件
├── docs/                             # 文档
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## API设计

### 核心API端点

```
文档管理:
POST   /api/v1/documents              # 上传文档
GET    /api/v1/documents              # 获取文档列表
GET    /api/v1/documents/{id}         # 获取文档详情
PUT    /api/v1/documents/{id}         # 更新文档
DELETE /api/v1/documents/{id}         # 删除文档

版本管理:
GET    /api/v1/documents/{id}/versions    # 获取版本列表
GET    /api/v1/documents/{id}/versions/{version}  # 获取特定版本
POST   /api/v1/documents/{id}/versions     # 创建新版本
PUT    /api/v1/documents/{id}/versions/{version}/restore  # 恢复版本

搜索服务:
GET    /api/v1/search                     # 全文搜索
POST   /api/v1/search/index              # 重建索引
GET    /api/v1/search/suggest             # 搜索建议

权限管理:
GET    /api/v1/permissions               # 获取权限列表
POST   /api/v1/permissions               # 创建权限
PUT    /api/v1/permissions/{id}          # 更新权限
DELETE /api/v1/permissions/{id}          # 删除权限

预览编辑:
GET    /api/v1/documents/{id}/preview    # 文档预览
PUT    /api/v1/documents/{id}/edit       # 在线编辑
GET    /api/v1/documents/{id}/download   # 文档下载

批量操作:
POST   /api/v1/batch/upload               # 批量上传
POST   /api/v1/batch/delete               # 批量删除
POST   /api/v1/batch/permissions          # 批量设置权限
```

## 数据模型

### 文档模型
```go
type Document struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    UUID         string    `json:"uuid" gorm:"uniqueIndex;not null"`
    TenantID     string    `json:"tenant_id" gorm:"not null;index"`
    Name         string    `json:"name" gorm:"not null"`
    Description  string    `json:"description"`
    OriginalName string    `json:"original_name"`
    MIMEType     string    `json:"mime_type"`
    Size         int64     `json:"size"`
    Category     string    `json:"category" gorm:"index"`
    Tags         string    `json:"tags"`
    EntityType   string    `json:"entity_type" gorm:"index"`
    EntityID     uint      `json:"entity_id" gorm:"index"`
    CurrentVersion uint    `json:"current_version"`
    Status       string    `json:"status" gorm:"default:'active'"`
    CreatedBy    uint      `json:"created_by"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
```

### 版本模型
```go
type DocumentVersion struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    DocumentID   uint      `json:"document_id" gorm:"not null;index"`
    Version      int       `json:"version" gorm:"not null"`
    UUID         string    `json:"uuid" gorm:"uniqueIndex;not null"`
    StoragePath  string    `json:"storage_path" gorm:"not null"`
    FileHash     string    `json:"file_hash" gorm:"not null;index"`
    Size         int64     `json:"size"`
    Description  string    `json:"description"`
    CreatedBy    uint      `json:"created_by"`
    CreatedAt    time.Time `json:"created_at"`

    // 关联
    Document     Document `json:"document" gorm:"foreignKey:DocumentID"`
}
```

## 部署

### Docker部署
```bash
# 构建镜像
docker build -t document-service:latest .

# 运行服务
docker-compose up -d
```

### Kubernetes部署
```bash
# 应用配置
kubectl apply -f k8s/
```

## 开发

### 环境要求
- Go 1.23+
- MySQL 8.0+
- Elasticsearch 8+
- Redis 7+
- MinIO/S3

### 启动开发环境
```bash
# 启动依赖服务
docker-compose -f docker-compose.dev.yml up -d

# 运行迁移
make migrate

# 启动服务
make run
```

## 监控

- **Health Check**: `/health`
- **Metrics**: `/metrics` (Prometheus格式)
- **Logging**: 结构化JSON日志
- **Tracing**: OpenTelemetry支持

## 安全

- JWT认证
- RBAC权限控制
- API限流
- 数据加密
- 审计日志

## 许可证

MIT License