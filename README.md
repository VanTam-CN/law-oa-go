# Law OA Go 项目

## 项目状态 ✅

**🎉 Phase 2 详细模块开发进行中！**

### Phase 2.1: 客户管理模块 ✅ 已完成
- ✅ 完整的客户 CRUD 操作
- ✅ 数据验证和业务规则检查
- ✅ 分页查询和条件筛选
- ✅ 客户搜索功能
- ✅ 批量导入（支持 CSV/Excel）
- ✅ 批量删除功能
- ✅ 客户统计信息
- ✅ 身份证号验证
- ✅ 重复数据检查
- ✅ 完整的 API 文档

### Phase 2 剩余模块
- 🔄 Phase 2.2: 律师管理模块 (待开发)
- ⏳ Phase 2.3: 案件管理模块 (待开发)
- ⏳ Phase 2.4: 利益冲突检查模块 (待开发)
- ⏳ Phase 2.5: 文档管理模块 (待开发)
- ⏳ Phase 2.6: 报表统计模块 (待开发)

**当前版本**: v2.1.0 (客户管理模块完成)

**构建状态**: ✅ 编译成功 (38.9MB 可执行文件)

**最新进展**: 客户管理模块已全面完成，包含完整的业务逻辑和高级功能

## 项目概述

Law OA Go 是一个基于 Go 语言开发的律师事务所办公自动化系统，提供案件管理、客户管理、利益冲突检查等核心功能。

## 技术栈

- **后端框架**: Gin (Go Web Framework)
- **数据库**: MySQL 8.0
- **缓存**: Redis 7
- **搜索引擎**: Elasticsearch 8.8
- **ORM**: GORM
- **认证**: JWT
- **容器化**: Docker & Docker Compose
- **代理**: Nginx

## 快速开始

### 前置条件

- Go 1.21+
- Docker & Docker Compose
- MySQL 8.0+
- Redis 7+
- Elasticsearch 8.8+

### 安装步骤

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd law-oa-go
   ```

2. **初始化项目**
   ```bash
   ./dev.sh init
   ```

3. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件，配置数据库连接等信息
   ```

4. **启动开发环境**
   ```bash
   ./dev.sh start
   ```

5. **访问应用**
   - API 地址: http://localhost:8080
   - Kibana 地址: http://localhost:5601

### 开发命令

```bash
# 初始化项目
./dev.sh init

# 启动开发环境
./dev.sh start

# 停止开发环境
./dev.sh stop

# 构建项目
./dev.sh build

# 运行测试
./dev.sh test

# 运行代码检查
./dev.sh lint

# 运行应用
./dev.sh run

# 生成文档
./dev.sh docs

# 清理项目
./dev.sh clean

# 显示帮助
./dev.sh help
```

## 项目结构

```
law-oa-go/
├── main.go                 # 主程序入口
├── go.mod                  # Go 模块文件
├── go.sum                  # Go 依赖校验
├── config.yaml             # 配置文件
├── .env.example           # 环境变量示例
├── internal/              # 内部包
│   ├── config/            # 配置模块
│   ├── database/          # 数据库模块
│   ├── handlers/          # 处理器
│   ├── middleware/        # 中间件
│   ├── models/            # 数据模型
│   └── router/            # 路由
├── scripts/               # 脚本文件
├── nginx/                 # Nginx 配置
├── docs/                  # 文档
├── logs/                  # 日志文件
├── Dockerfile             # Docker 镜像配置
├── docker-compose.yml     # Docker Compose 配置
└── dev.sh                 # 开发脚本
```

## API 文档

### 认证接口

#### 登录
- **URL**: `POST /api/v1/auth/login`
- **请求体**:
  ```json
  {
    "username": "admin",
    "password": "password"
  }
  ```
- **响应**:
  ```json
  {
    "code": 200,
    "message": "登录成功",
    "data": {
      "token": "jwt-token",
      "user": {
        "id": 1,
        "username": "admin",
        "email": "admin@lawfirm.com"
      }
    }
  }
  ```

### 案件管理接口

#### 获取案件列表
- **URL**: `GET /api/v1/cases`
- **查询参数**:
  - `page`: 页码 (默认: 1)
  - `size`: 每页大小 (默认: 10)
  - `case_name`: 案件名称 (可选)
  - `case_type`: 案件类型 (可选)
  - `status`: 状态 (可选)
- **响应**:
  ```json
  {
    "code": 200,
    "message": "获取成功",
    "data": {
      "cases": [...],
      "total": 100,
      "page": 1,
      "size": 10
    }
  }
  ```

#### 创建案件
- **URL**: `POST /api/v1/cases`
- **请求体**:
  ```json
  {
    "case_name": "案件名称",
    "case_type": "CIVIL",
    "project_type": "CASE",
    "principal_info": "委托人信息",
    "opponent_info": "对方当事人信息",
    "cause_of_action": "案由",
    "description": "案件描述",
    "contract_amount": 50000.00,
    "billing_method": "FIXED"
  }
  ```

### 客户管理接口

#### 获取客户列表
- **URL**: `GET /api/v1/clients`
- **查询参数**:
  - `page`: 页码
  - `size`: 每页大小
  - `client_name`: 客户名称
  - `client_type`: 客户类型

#### 创建客户
- **URL**: `POST /api/v1/clients`
- **请求体**:
  ```json
  {
    "client_name": "客户名称",
    "phone": "联系电话",
    "email": "邮箱",
    "client_type": "individual",
    "address": "地址"
  }
  ```

### 律师管理接口

#### 获取律师列表
- **URL**: `GET /api/v1/lawyers`
- **查询参数**:
  - `page`: 页码
  - `size`: 每页大小
  - `lawyer_name`: 律师姓名
  - `department`: 部门

#### 创建律师
- **URL**: `POST /api/v1/lawyers`
- **请求体**:
  ```json
  {
    "lawyer_name": "律师姓名",
    "phone": "联系电话",
    "email": "邮箱",
    "license_no": "执业证号",
    "position": "职位",
    "department": "部门",
    "specialty": "专长领域"
  }
  ```

### 利益冲突检查接口

#### 执行冲突检查
- **URL**: `POST /api/v1/conflict/check`
- **请求体**:
  ```json
  {
    "client_info": "客户信息",
    "opponent_info": "对方当事人信息",
    "case_type": "案件类型",
    "cause_of_action": "案由"
  }
  ```
- **响应**:
  ```json
  {
    "code": 200,
    "message": "检查完成",
    "data": {
      "conflicts": [...],
      "total_cases": 100,
      "risk_level": "medium"
    }
  }
  ```

## 数据模型

### 用户 (User)
- id: 用户ID
- username: 用户名
- password: 密码
- email: 邮箱
- real_name: 真实姓名
- status: 状态

### 客户 (Client)
- id: 客户ID
- client_name: 客户名称
- phone: 联系电话
- email: 邮箱
- client_type: 客户类型
- address: 地址

### 律师 (Lawyer)
- id: 律师ID
- lawyer_name: 律师姓名
- phone: 联系电话
- email: 邮箱
- license_no: 执业证号
- position: 职位
- department: 部门

### 案件 (Case)
- id: 案件ID
- case_no: 案件编号
- case_name: 案件名称
- case_type: 案件类型
- client_id: 客户ID
- lawyer_id: 律师ID
- status: 状态
- description: 案件描述
- contract_amount: 合同金额

### 利益冲突检查记录 (ConflictCheckRecord)
- id: 记录ID
- case_id: 案件ID
- check_type: 检查类型
- conflict_level: 冲突级别
- conflict_desc: 冲突描述
- status: 状态

## 部署

### Docker 部署

1. **构建镜像**
   ```bash
   docker build -t law-oa-go .
   ```

2. **启动服务**
   ```bash
   docker-compose up -d
   ```

### 生产环境部署

1. **环境配置**
   ```bash
   # 生产环境配置
   ENVIRONMENT=production
   PORT=8080
   DB_HOST=production-db-host
   DB_PASSWORD=secure-password
   JWT_SECRET=very-secure-secret-key
   ```

2. **构建和运行**
   ```bash
   ./dev.sh build
   ./main
   ```

## 开发指南

### 代码规范

- 遵循 Go 官方代码规范
- 使用 golangci-lint 进行代码检查
- 编写单元测试
- 使用 godoc 格式编写注释

### 数据库操作

- 使用 GORM 进行数据库操作
- 遵循数据库迁移最佳实践
- 使用事务保证数据一致性

### 错误处理

- 使用自定义错误类型
- 记录错误日志
- 提供友好的错误信息

### 性能优化

- 使用 Redis 缓存热点数据
- 优化数据库查询
- 使用 Elasticsearch 进行全文搜索

## 监控和日志

### 应用监控

- 使用 Prometheus 进行指标监控
- 使用 Grafana 进行可视化
- 集成健康检查接口

### 日志管理

- 结构化日志记录
- 按级别分类记录
- 日志轮转和归档

## 安全考虑

### 认证和授权

- JWT 令牌认证
- 角色权限控制
- 令牌过期管理

### 数据安全

- 密码加密存储
- SQL 注入防护
- XSS 防护

### 网络安全

- HTTPS 加密传输
- CORS 跨域配置
- 请求频率限制

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务是否启动
   - 验证连接参数是否正确
   - 确认网络连接正常

2. **Redis 连接失败**
   - 检查 Redis 服务状态
   - 验证 Redis 配置
   - 检查防火墙设置

3. **Elasticsearch 连接失败**
   - 检查 ES 服务状态
   - 验证 ES 配置
   - 检查索引是否存在

### 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 查看 Docker 日志
docker-compose logs -f app

# 查看 Nginx 日志
docker-compose logs -f nginx
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 联系方式

- 项目地址: [GitHub Repository]
- 问题反馈: [Issues]
- 邮箱: [Contact Email]