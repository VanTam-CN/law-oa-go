# Technology Steering Document

## 技术栈概述

### 后端技术栈
- **语言**: Go 1.23+ (使用最新语言特性)
- **Web框架**: Gin v1.9.1 (高性能HTTP框架)
- **ORM**: GORM v1.30.0 (现代化ORM框架)
- **数据库**: MySQL 8.0+ (主数据库)
- **缓存**: Redis 7+ (go-redis/v9客户端)
- **搜索引擎**: Elasticsearch 8+ (go-elasticsearch/v8)
- **认证**: JWT v5.0.0 (golang-jwt/jwt)
- **配置管理**: Viper v1.16.0
- **日志**: Zap v1.24.0 (结构化日志)
- **监控**: Prometheus client_golang v1.16.0

### 前端技术栈

#### Bootstrap版本 (frontend/)
- **框架**: React 18.2.0 + TypeScript 5.9.2
- **UI库**: Bootstrap 5.3.1 + React Bootstrap 2.8.0
- **路由**: React Router DOM 6.15.0
- **状态管理**: Redux Toolkit 2.9.0
- **HTTP客户端**: Axios 1.5.0
- **构建工具**: CRACO 7.1.0
- **国际化**: i18next 25.5.2
- **图标**: Heroicons、Font Awesome、React Icons

#### Ant Design版本 (frontend-vue/)
- **框架**: React 18.2.0 + TypeScript 5.0.2
- **UI库**: Ant Design 5.16.1
- **路由**: React Router DOM 7.8.2
- **图表**: ECharts 5.6.0 + Recharts 3.1.2
- **构建工具**: Vite 5.1.0
- **样式**: Less 4.4.1
- **工具**: Lodash 4.17.21, Dayjs 1.11.10

### 开发工具
- **测试**: testify 1.10.0, go-sqlmock 1.5.2
- **代码检查**: golangci-lint
- **API文档**: Swagger/OpenAPI 3.0 (swaggo)
- **依赖管理**: Go Modules + npm
- **构建工具**: Make + Docker

### 运维部署
- **容器化**: Docker 20.10+ + Docker Compose
- **健康检查**: 内置健康检查端点
- **监控**: Prometheus指标导出
- **日志**: JSON结构化日志 (lumberjack)
- **数据库迁移**: golang-migrate

## 核心架构模式

### 分层架构
```
handlers/     # HTTP处理器层
├── services/  # 业务逻辑层
├── models/    # 数据模型层
├── repositories/ # 数据访问层
├── middleware/  # 中间件层
├── config/     # 配置管理
├── database/   # 数据库连接
├── security/   # 安全模块
└── utils/      # 工具函数
```

### 中间件架构
- **认证中间件**: JWT令牌验证
- **日志中间件**: 请求/响应日志记录
- **监控中间件**: 性能指标收集
- **安全中间件**: CORS、请求限制、安全头
- **错误处理中间件**: 统一错误处理

### 数据访问模式
- **Repository模式**: 封装数据访问逻辑
- **UnitOfWork模式**: 事务管理
- **连接池**: 数据库连接池管理
- **缓存策略**: Redis缓存 + 本地缓存

## 构建系统

### 后端构建 (Make)
```bash
# 基础命令
make build          # 构建应用
make run            # 运行应用
make dev            # 开发模式运行
make clean          # 清理构建文件
make deps           # 安装依赖

# 代码质量
make fmt            # 代码格式化
make lint           # 代码检查
make test           # 运行测试
make security       # 安全检查

# 高级功能
make pgo-build      # PGO优化构建
make profile        # 性能分析
make fuzz-all       # Fuzzing测试
```

### 前端构建 (Bootstrap版本)
```bash
cd frontend/
npm start           # 启动开发服务器
npm run build       # 构建生产版本
npm test            # 运行测试
npm run format      # 代码格式化
npm run lint        # 代码检查
npm run type-check  # TypeScript类型检查
```

### 前端构建 (Ant Design版本)
```bash
cd frontend-vue/
npm run dev         # 启动开发服务器
npm run build       # 构建生产版本
npm run preview     # 预览构建结果
```

## 测试策略

### 后端测试
- **单元测试**: go test ./internal/...
- **集成测试**: go test ./tests/integration/...
- **端到端测试**: go test ./tests/e2e/...
- **性能测试**: go test -bench=. ./tests/performance/...
- **Fuzzing测试**: go test -fuzz=Fuzz_...

### 测试覆盖要求
- 单元测试覆盖率 ≥ 70%
- 集成测试覆盖主要功能
- 所有测试必须通过
- 性能测试基准达标

## 数据库管理

### 迁移系统
```bash
make migrate-up      # 执行迁移
make migrate-down    # 回滚迁移
make migrate-create  # 创建迁移文件
```

### 数据库特性
- MySQL 8.0+ 主数据库
- GORM ORM支持
- 自动迁移和版本控制
- 连接池优化
- 查询性能监控

## 部署配置

### Docker容器化
```bash
# 构建镜像
docker build -t law-oa-go:latest .

# 开发环境
./scripts/start-dev.sh start

# 生产部署
docker-compose up -d
```

### 环境配置
- **开发环境**: docker-compose.dev.yml
- **生产环境**: docker-compose.yml
- **环境变量**: .env文件管理
- **配置文件**: config/config.yaml

## 性能优化

### PGO (Profile-Guided Optimization)
```bash
make pgo-build      # PGO优化构建
make pgo-full       # 完整PGO流程
make profile        # 性能剖析
```

### 缓存策略
- **Redis缓存**: 分布式缓存
- **本地缓存**: 内存缓存
- **缓存失效**: TTL + 主动失效
- **缓存预热**: 启动时预加载热点数据

## 监控和日志

### 结构化日志
- **日志格式**: JSON
- **日志级别**: DEBUG/INFO/WARN/ERROR
- **日志轮转**: lumberjack
- **上下文追踪**: correlation ID

### 监控指标
- **应用指标**: Prometheus导出
- **健康检查**: /health端点
- **性能指标**: 响应时间、QPS、错误率
- **系统指标**: CPU、内存、磁盘、网络

## 安全配置

### 认证授权
- **JWT认证**: 无状态令牌
- **密码加密**: bcrypt
- **权限控制**: RBAC模型
- **会话管理**: 令牌过期和刷新

### 安全防护
- **输入验证**: validator库
- **SQL注入防护**: GORM参数化查询
- **XSS防护**: 输入过滤和输出编码
- **CORS配置**: 跨域请求控制

## 开发规范

### Go代码规范
- 遵循官方Go代码规范
- 使用golangci-lint进行代码检查
- 编写完整的单元测试
- 使用godoc格式编写注释

### TypeScript规范
- 严格的类型检查
- 使用ESLint和Prettier
- 组件化和模块化开发
- 完整的props类型定义