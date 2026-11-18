# Project Context

## Purpose
Law OA Go 是一个基于 Go 1.25+ 构建的现代化律师事务所办公自动化系统，为中小型律师事务所提供完整的数字化解决方案。项目采用单体架构设计，注重高性能、安全性和可维护性，支持案件管理、客户管理、用户管理等核心业务功能。

### 核心目标
- 提供完整的律师事务所办公自动化解决方案
- 实现高并发处理能力（API响应时间 < 100ms）
- 构建安全可靠的权限管理体系
- 支持多数据库环境（PostgreSQL + MySQL）
- 实现现代化的用户界面和交互体验

## Tech Stack

### 后端技术栈
- **语言**: Go 1.25+ (使用最新语言特性和PGO优化)
- **Web框架**: Gin v1.9.1 (高性能HTTP框架)
- **ORM**: GORM v1.30.0 (现代化ORM框架)
- **数据库**: PostgreSQL 15+ / MySQL 8.0+ (支持双环境)
- **缓存**: Redis 7+ (go-redis/v9客户端)
- **搜索引擎**: Elasticsearch 8.11+ (全文搜索)
- **认证**: JWT v5.0.0 (golang-jwt/jwt)
- **配置管理**: Viper v1.16.0
- **日志**: Zap v1.24.0 (结构化日志)
- **监控**: Prometheus client_golang v1.16.0
- **API文档**: Swagger/OpenAPI 3.0 (swaggo)

### 前端技术栈
- **框架**: React 18.2.0 + TypeScript 5.0.2
- **构建工具**: Vite 5.1.0
- **UI组件**: Ant Design 5.16.1
- **路由**: React Router 7.9.4
- **图表**: Recharts 3.1.2
- **HTTP客户端**: Axios 1.12.2
- **状态管理**: Zustand 5 + TanStack React Query 5
- **样式**: Less 4.4.1
- **工具**: Lodash 4.17.21, Dayjs 1.11.10

### 基础设施
- **容器化**: Docker & Docker Compose
- **数据库迁移**: golang-migrate/v4
- **测试**: testify 1.10.0, go-sqlmock 1.5.2
- **代码检查**: golangci-lint, ESLint, Prettier
- **反向代理**: Nginx

## Project Conventions

### Code Style

#### Go 代码规范
- 遵循 Go 官方代码规范和最佳实践
- 使用 `gofmt -s -w ./...` 格式化代码
- 使用 `golangci-lint run` 进行静态检查
- 文件命名采用 snake_case (如: `user_service.go`)
- 类型和函数使用 PascalCase (如: `UserService`, `GetUserByID`)
- 常量使用 UPPER_SNAKE_CASE (如: `MAX_RETRY_COUNT`)
- 私有函数使用下划线前缀 (如: `_validateInput`)
- 所有公共函数必须包含完整的 godoc 注释
- 单元测试覆盖率要求 ≥ 70%

#### TypeScript 代码规范
- 使用严格的 TypeScript 模式
- 组件命名使用 PascalCase (如: `UserManagement.tsx`)
- 变量和函数使用 camelCase (如: `getUserById`)
- 常量使用 UPPER_SNAKE_CASE (如: `API_BASE_URL`)
- 使用 ESLint 和 Prettier 进行代码格式化
- 组件优先使用函数式组件和 React Hooks
- 所有组件和函数必须有完整的类型定义

### Architecture Patterns

#### 分层架构设计
```
handlers/     # HTTP处理器层 - 处理HTTP请求和响应
├── services/  # 业务逻辑层 - 核心业务逻辑处理
├── repositories/ # 数据访问层 - 数据库操作封装
├── models/    # 数据模型层 - 数据模型定义
├── middleware/  # 中间件层 - 认证、日志、监控等
├── config/     # 配置管理 - 应用配置管理
└── utils/      # 工具函数 - 通用工具函数
```

#### Repository 模式
- 封装数据访问逻辑，提供统一的数据操作接口
- 支持事务管理和连接池优化
- 使用 GORM 进行数据库操作
- 实现缓存策略（Redis + 本地缓存）

#### 中间件架构
- **认证中间件**: JWT 令牌验证
- **日志中间件**: 结构化请求/响应日志
- **监控中间件**: Prometheus 指标收集
- **安全中间件**: CORS、请求限制、安全头设置
- **错误处理中间件**: 统一错误处理和响应

#### 单体应用架构
- 采用单体架构便于部署和维护
- 清晰的模块边界和职责分离
- 支持水平扩展和负载均衡
- 容器化部署支持

### Testing Strategy

#### 后端测试
- **单元测试**: 使用 testify 进行单元测试，覆盖率 ≥ 70%
- **集成测试**: 使用 go-sqlmock 进行数据库集成测试
- **API测试**: 使用 gin-test 框架进行API端点测试
- **性能测试**: 使用 Go 标准库进行基准测试
- **Fuzzing测试**: 使用 Go 1.25+ 的模糊测试功能

#### 前端测试
- **组件测试**: React 组件单元测试
- **集成测试**: API 集成测试
- **端到端测试**: 使用 Playwright 进行 E2E 测试
- **类型检查**: TypeScript 编译时类型检查

#### 测试环境
- 使用 Docker Compose 提供独立的测试环境
- 数据库使用临时实例进行隔离测试
- Redis 和 Elasticsearch 使用内存模式进行测试

### Git Workflow

#### 分支策略
- **main**: 主分支，生产环境代码
- **vue**: 当前开发分支，前端重构
- **feature/***: 功能开发分支
- **hotfix/***: 紧急修复分支
- **release/***: 发布准备分支

#### 提交规范
使用约定式提交格式：
- `feat`: 新功能
- `fix`: 修复问题
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具变动

#### 代码审查
- 所有代码变更必须通过 Pull Request
- 至少需要一人审查才能合并
- 自动化测试必须通过
- 代码覆盖率不能降低

## Domain Context

### 律师事务所业务模型
- **案件管理**: 案件信息、状态跟踪、律师分配、文档管理
- **客户管理**: 客户档案、联系信息、案件关联、统计分析
- **用户管理**: 律师信息、权限分配、资料管理、操作日志
- **统计分析**: 业务报表、数据可视化、性能指标
- **文档管理**: 案件文档、合同文件、证据材料、归档管理

### 业务规则
- **权限管理**: 基于RBAC模型，支持管理员、律师、助理等角色
- **数据验证**: 所有输入参数必须经过验证，敏感数据加密存储
- **响应格式**: 统一API响应格式 {data, error, message, timestamp}
- **错误处理**: 结构化错误响应，完整的错误日志记录
- **审计日志**: 所有关键操作必须记录审计日志

### 数据模型关系
- 用户 (User) 可以管理多个案件 (Case)
- 客户 (Client) 可以关联多个案件
- 案件 (Case) 分配给特定律师 (User)
- 支持多对多关系的案件标签和分类

## Important Constraints

### 技术约束
- **数据库**: 必须支持 PostgreSQL 和 MySQL 双环境
- **性能**: API 响应时间必须 < 100ms
- **并发**: 支持高并发访问，使用 Go 协程
- **缓存**: 必须使用 Redis 作为分布式缓存
- **搜索**: 需要集成 Elasticsearch 进行全文搜索

### 业务约束
- **安全性**: 必须实现 JWT 认证和 RBAC 权限控制
- **数据隐私**: 客户和案件信息必须加密存储
- **审计**: 所有关键操作必须有完整的审计日志
- **备份**: 需要支持数据备份和恢复功能
- **合规**: 符合律师事务所行业的数据保护要求

### 部署约束
- **容器化**: 必须支持 Docker 容器化部署
- **配置管理**: 使用环境变量进行配置管理
- **监控**: 需要集成 Prometheus 监控和日志系统
- **健康检查**: 必须提供完整的健康检查端点
- **扩展性**: 支持水平扩展和负载均衡

## External Dependencies

### 核心依赖
- **PostgreSQL 15+**: 主数据库，支持 JSON 和全文搜索
- **MySQL 8.0+**: 备选数据库，兼容性支持
- **Redis 7+**: 缓存和会话存储
- **Elasticsearch 8.11+**: 全文搜索引擎
- **Nginx**: 反向代理和静态文件服务

### 第三方服务
- **JWT**: 用于无状态认证
- **Prometheus**: 监控指标收集
- **Swagger**: API 文档生成
- **Docker**: 容器化部署
- **GitHub Actions**: CI/CD 流水线

### 开发工具
- **golangci-lint**: Go 代码检查
- **ESLint**: JavaScript/TypeScript 代码检查
- **Prettier**: 代码格式化
- **Playwright**: 端到端测试
- **Swag**: API 文档生成

### API 集成
- 邮件服务（规划中）
- 短信服务（规划中）
- 文件存储服务（规划中）
- 第三方法律数据库服务（规划中）
