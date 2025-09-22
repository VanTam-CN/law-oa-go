# 项目架构文档 - 实际情况

## 🏗️ 当前架构概述

### 架构类型：单体应用 + 分层架构

```
law-oa-go/
├── cmd/server/                    # 应用入口点
│   └── main.go                   # 主程序入口
├── internal/                     # 内部模块，不对外暴露
│   ├── api/                     # HTTP API层
│   │   ├── handlers/           # 请求处理器
│   │   ├── middleware/         # 中间件
│   │   ├── routes/             # 路由定义
│   │   └── response.go         # 统一响应格式
│   ├── services/               # 业务逻辑层
│   │   ├── user_service.go    # 用户业务逻辑
│   │   ├── case_service.go    # 案件业务逻辑
│   │   ├── client_service.go   # 客户业务逻辑
│   │   └── auth_service.go     # 认证业务逻辑
│   ├── repositories/           # 数据访问层
│   │   ├── user_repository.go # 用户数据访问
│   │   ├── case_repository.go  # 案件数据访问
│   │   ├── client_repository.go # 客户数据访问
│   │   └── interfaces.go      # 仓储接口定义
│   ├── models/                 # 数据模型
│   │   ├── user.go            # 用户模型
│   │   ├── case.go            # 案件模型
│   │   └── client.go          # 客户模型
│   ├── config/                 # 配置管理
│   │   ├── config.go          # 配置结构体
│   │   └── config.yaml        # 配置文件
│   ├── middleware/            # 中间件
│   │   ├── jwt.go            # JWT认证中间件
│   │   ├── logging.go        # 日志中间件
│   │   └── cors.go           # CORS中间件
│   └── errors/               # 错误处理
│       └── errors.go         # 统一错误处理
├── pkg/                        # 公共包，可对外暴露
│   ├── utils/                 # 工具函数
│   ├── database/              # 数据库连接
│   └── logging/               # 日志工具
├── migrations/                 # 数据库迁移文件
├── docs/                       # 项目文档
├── scripts/                    # 脚本文件
└── go.mod                     # Go模块文件
```

## 🔧 技术栈

### 后端技术栈
- **语言**: Go 1.23
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL/PostgreSQL
- **认证**: JWT
- **配置管理**: Viper
- **日志**: Zap Logger
- **错误处理**: 自定义错误系统

### 架构模式
- **分层架构**: Controller-Service-Repository
- **依赖注入**: 通过构造函数注入
- **中间件模式**: 用于认证、日志、CORS等
- **统一响应格式**: 标准化API响应

## 📊 数据流

```
HTTP请求 → 中间件 → 路由 → 处理器 → 服务层 → 仓储层 → 数据库
    ↓
统一响应格式 ← 错误处理 ← 业务逻辑 ← 数据模型 ← 数据库操作
```

## 🛠️ 核心模块

### 1. API层 (`internal/api/`)
- **职责**: 处理HTTP请求，调用业务逻辑，返回响应
- **特点**: 
  - 统一的错误处理
  - 统一的响应格式
  - 参数验证
  - 中间件支持

### 2. 服务层 (`internal/services/`)
- **职责**: 实现业务逻辑，协调多个仓储
- **特点**:
  - 业务规则验证
  - 事务管理
  - 跨多个实体的业务操作

### 3. 仓储层 (`internal/repositories/`)
- **职责**: 数据访问抽象，封装数据库操作
- **特点**:
  - 接口定义
  - CRUD操作
  - 查询方法
  - 事务支持

### 4. 模型层 (`internal/models/`)
- **职责**: 定义数据结构和业务规则
- **特点**:
  - GORM模型
  - 数据验证
  - 关系定义

## 🔐 安全架构

### 认证与授权
- **JWT认证**: 无状态认证
- **中间件**: 路由级别的认证控制
- **角色权限**: 基于角色的访问控制

### 数据安全
- **参数验证**: 输入数据验证
- **SQL注入防护**: GORM自动防护
- **敏感信息**: 不记录敏感日志

## 📈 性能特点

### 当前性能
- **架构**: 单体应用，低延迟
- **数据库**: 单一数据库连接池
- **缓存**: 当前无缓存层
- **并发**: 通过Goroutine处理并发

### 性能优化建议
1. **数据库优化**: 索引优化，查询优化
2. **缓存层**: Redis缓存热点数据
3. **连接池**: 数据库连接池优化
4. **监控**: 性能监控和告警

## 🚀 部署架构

### 当前部署方式
- **部署**: 单体应用部署
- **容器化**: Docker支持
- **数据库**: 单一数据库实例
- **文件存储**: 本地文件系统

### 建议部署方式
```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
    depends_on:
      - db
  
  db:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=root
      - MYSQL_DATABASE=law_oa
    volumes:
      - db_data:/var/lib/mysql

volumes:
  db_data:
```

## 🔄 未来演进方向

### 短期目标（3-6个月）
1. **完善监控**: 集成Prometheus和Grafana
2. **性能优化**: 数据库查询优化，添加缓存
3. **测试覆盖**: 提高单元测试覆盖率
4. **文档完善**: API文档和架构文档

### 中期目标（6-12个月）
1. **微前端**: 前端模块化
2. **消息队列**: 异步任务处理
3. **搜索功能**: 集成Elasticsearch
4. **文件管理**: 专业的文件存储服务

### 长期目标（1年以上）
1. **微服务拆分**: 根据业务增长逐步拆分
2. **云原生**: Kubernetes部署
3. **DevOps**: 完善的CI/CD流程
4. **高可用**: 多可用区部署

## 📝 开发规范

### 代码结构
- **包名**: 小写，简洁明了
- **文件名**: snake_case
- **函数名**: PascalCase（导出）或snake_case（私有）
- **变量名**: camelCase

### 错误处理
- **统一错误格式**: 使用自定义错误系统
- **错误日志**: 记录错误堆栈和上下文
- **用户友好**: 返回用户友好的错误信息

### API设计
- **RESTful**: 遵循REST设计原则
- **统一响应**: 标准化的响应格式
- **版本控制**: API版本管理
- **文档**: Swagger/OpenAPI文档

## 📋 检查清单

- [x] 单体架构实现
- [x] 分层架构设计
- [x] 统一错误处理
- [x] JWT认证
- [x] GORM数据访问
- [x] 配置管理
- [x] 日志系统
- [x] 中间件支持
- [ ] 完整测试覆盖
- [ ] 性能监控
- [ ] 缓存层
- [ ] API文档
- [ ] 部署自动化

---

**文档版本**: v1.0  
**最后更新**: 2025-01-14  
**维护者**: 开发团队