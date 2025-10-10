# Project Structure Steering Document

## 目录组织结构

### 根目录结构
```
law-oa-go/
├── main.go                   # 应用程序入口点
├── go.mod                    # Go模块依赖文件
├── go.sum                    # Go依赖校验文件
├── Makefile                  # 构建命令配置
├── .env.example             # 环境变量模板
├── docker-compose.yml        # Docker Compose配置
├── docker-compose.dev.yml    # 开发环境Docker配置
├── Dockerfile               # 后端Docker镜像构建文件
├── .golangci.yml            # 代码检查配置
├── .air.toml                # 热重载配置
├── token.txt                # JWT密钥文件（开发用）
├── README.md                # 项目说明文档
├── .claude/                 # AI辅助配置目录
└── scripts/                 # 构建和部署脚本
```

### internal/ 目录结构 (内部包，不对外暴露)
```
internal/
├── handlers/               # HTTP处理器层
│   ├── auth_handler.go      # 认证相关处理器
│   ├── user_handler.go      # 用户管理处理器
│   ├── client_handler.go    # 客户管理处理器
│   ├── case_handler.go      # 案件管理处理器
│   └── health_handler.go   # 健康检查处理器
├── services/               # 业务逻辑层
│   ├── auth_service.go      # 认证服务
│   ├── user_service.go      # 用户服务
│   ├── client_service.go    # 客户服务
│   ├── case_service.go      # 案件服务
│   └── stats_service.go     # 统计服务
├── models/                 # 数据模型层
│   ├── models.go           # 数据模型定义
│   ├── user.go             # 用户模型
│   ├── client.go           # 客户模型
│   ├── case.go             # 案件模型
│   └── validators.go       # 验证器
├── repositories/           # 数据访问层
│   ├── user_repository.go  # 用户数据访问
│   ├── client_repository.go # 客户数据访问
│   ├── case_repository.go  # 案件数据访问
│   └── base_repository.go # 基础数据访问
├── middleware/             # 中间件层
│   ├── auth_middleware.go  # 认证中间件
│   ├── logging_middleware.go # 日志中间件
│   ├── cors_middleware.go  # CORS中间件
│   ├── recovery_middleware.go # 恢复中间件
│   └── metrics_middleware.go # 监控中间件
├── database/               # 数据库相关
│   ├── connection.go       # 数据库连接
│   ├── migration.go        # 数据库迁移
│   └── connection_pool.go  # 连接池管理
├── security/               # 安全模块
│   ├── jwt.go              # JWT处理
│   ├── encryption.go       # 加密工具
│   └── auth.go             # 认证逻辑
├── config/                 # 配置管理
│   ├── config.go           # 配置结构
│   └── database.go         # 数据库配置
├── validators/             # 验证器
│   ├── user_validator.go   # 用户验证器
│   ├── client_validator.go # 客户验证器
│   └── case_validator.go   # 案件验证器
├── utils/                  # 工具函数
│   ├── response.go        # 响应工具
│   ├── errors.go          # 错误工具
│   └── pagination.go      # 分页工具
├── router/                 # 路由配置
│   └── router.go          # 路由定义
└── logger/                 # 日志配置
    └── logger.go          # 日志初始化
```

### 前端目录结构

#### Bootstrap版本 (frontend/)
```
frontend/
├── public/                 # 静态资源
│   ├── index.html         # 主页面
│   ├── favicon.ico        # 网站图标
│   └── manifest.json      # PWA配置
├── src/                    # 源代码
│   ├── components/        # React组件
│   │   ├── common/        # 通用组件
│   │   ├── layout/        # 布局组件
│   │   └── forms/         # 表单组件
│   ├── pages/            # 页面组件
│   │   ├── auth/          # 认证页面
│   │   ├── dashboard/     # 仪表板
│   │   ├── users/         # 用户管理
│   │   ├── clients/       # 客户管理
│   │   └── cases/         # 案件管理
│   ├── services/         # API服务
│   │   ├── api.js         # API客户端
│   │   ├── auth.js        # 认证服务
│   │   └── utils.js       # 工具函数
│   ├── contexts/         # React上下文
│   │   ├── authContext.js # 认证上下文
│   │   └── appContext.js  # 应用上下文
│   ├── hooks/            # 自定义hooks
│   ├── utils/            # 工具函数
│   ├── types/            # TypeScript类型定义
│   ├── constants/        # 常量定义
│   ├── styles/           # 样式文件
│   │   ├── global.css     # 全局样式
│   │   └── components.css # 组件样式
│   ├── App.js            # 主应用组件
│   ├── index.js          # 应用入口
│   └── setupTests.js     # 测试配置
├── package.json           # npm依赖配置
├── craco.config.js        # CRACO配置
├── tsconfig.json          # TypeScript配置
├── .env                   # 环境变量
└── .env.example          # 环境变量模板
```

#### Ant Design版本 (frontend-vue/)
```
frontend-vue/
├── src/                   # 源代码
│   ├── components/        # React组件
│   ├── pages/            # 页面组件
│   ├── api/              # API服务
│   ├── utils/            # 工具函数
│   ├── styles/           # 样式文件
│   └── main.tsx          # 应用入口
├── package.json          # npm依赖配置
├── vite.config.ts        # Vite配置
├── tsconfig.json         # TypeScript配置
├── .env                  # 环境变量
└── index.html            # 主页面
```

### 配置文件结构
```
config/
├── config.yaml           # 主配置文件
├── database.yaml         # 数据库配置
└── redis.yaml           # Redis配置
```

### 数据库迁移结构
```
migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_clients_table.up.sql
├── 000002_create_clients_table.down.sql
└── ...
```

### 脚本目录结构
```
scripts/
├── start-dev.sh          # 开发环境启动脚本
├── docker-manage.sh      # Docker管理脚本
├── database-init.sh      # 数据库初始化脚本
├── deployment-verification.sh # 部署验证脚本
├── project-summary.sh    # 项目总结脚本
├── test-integration.sh   # 集成测试脚本
├── my.cnf               # MySQL配置
├── redis.conf           # Redis配置
└── elasticsearch.yml    # Elasticsearch配置
```

### 测试目录结构
```
tests/
├── integration/          # 集成测试
├── e2e/                 # 端到端测试
├── performance/          # 性能测试
└── fixtures/            # 测试数据
```

## 文件命名约定

### Go文件命名
- **处理器**: `*_handler.go` (e.g., `user_handler.go`)
- **服务**: `*_service.go` (e.g., `user_service.go`)
- **模型**: `*.go` (e.g., `user.go`)
- **仓库**: `*_repository.go` (e.g., `user_repository.go`)
- **中间件**: `*_middleware.go` (e.g., `auth_middleware.go`)
- **验证器**: `*_validator.go` (e.g., `user_validator.go`)
- **测试文件**: `*_test.go` (e.g., `user_service_test.go`)

### 前端文件命名
- **React组件**: `PascalCase.tsx` (e.g., `UserManagement.tsx`)
- **页面组件**: `PascalCase.tsx` (e.g., `UserList.tsx`)
- **样式文件**: `*.less` or `*.css` (e.g., `UserList.less`)
- **类型定义**: `*.types.ts` (e.g., `user.types.ts`)
- **工具函数**: `*.js` or `*.ts` (e.g., `api.js`)
- **配置文件**: `*.config.js` (e.g., `craco.config.js`)

### 配置文件命名
- **环境变量**: `.env` 和 `.env.example`
- **Docker配置**: `docker-compose.yml` 和 `docker-compose.dev.yml`
- **构建配置**: `Makefile`, `package.json`
- **代码检查**: `.golangci.yml`, `.eslintrc.js`

## 关键文件位置

### 入口文件
- **后端入口**: `main.go`
- **前端入口**: `frontend/src/index.js` (Bootstrap)
- **前端入口**: `frontend-vue/src/main.tsx` (Ant Design)

### 配置文件
- **环境变量**: `.env.example` (模板)
- **Docker配置**: `docker-compose.yml` (生产), `docker-compose.dev.yml` (开发)
- **构建配置**: `Makefile` (后端), `package.json` (前端)
- **数据库配置**: `internal/config/database.go`

### 核心业务逻辑
- **用户管理**: `internal/services/user_service.go`
- **客户管理**: `internal/services/client_service.go`
- **案件管理**: `internal/services/case_service.go`
- **认证服务**: `internal/services/auth_service.go`

### 数据模型
- **用户模型**: `internal/models/user.go`
- **客户模型**: `internal/models/client.go`
- **案件模型**: `internal/models/case.go`
- **基础模型**: `internal/models/models.go`

### HTTP处理器
- **用户处理器**: `internal/handlers/user_handler.go`
- **客户处理器**: `internal/handlers/client_handler.go`
- **案件处理器**: `internal/handlers/case_handler.go`
- **认证处理器**: `internal/handlers/auth_handler.go`

### 前端核心组件
- **主应用**: `frontend/src/App.js` (Bootstrap)
- **主应用**: `frontend-vue/src/App.tsx` (Ant Design)
- **路由配置**: `frontend/src/App.js` (集成路由)
- **API服务**: `frontend/src/services/api.js`

## 目录组织原则

### 1. 内部包隔离
- `internal/` 目录包含所有内部代码，不对外暴露
- 外部只能通过 `main.go` 的导出函数访问

### 2. 分层架构
- `handlers/`: HTTP请求处理层
- `services/`: 业务逻辑层
- `repositories/`: 数据访问层
- `models/`: 数据模型层

### 3. 功能模块化
- 按功能模块组织代码 (用户、客户、案件)
- 每个模块包含完整的MVC结构

### 4. 配置集中化
- 所有配置文件集中在根目录
- 环境相关配置使用 `.env` 管理
- 构建配置使用标准工具配置

### 5. 脚本工具化
- 所有常用操作都提供脚本支持
- 脚本放在 `scripts/` 目录
- 脚本具有可执行权限

### 6. 前端模块化
- 组件按功能和层级分类
- 页面组件和通用组件分离
- 服务层和工具层独立