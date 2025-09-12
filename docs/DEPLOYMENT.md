# Go 版本部署指南

## 概述

Law OA Go 是基于 Go 语言重构的律师事务所办公自动化系统，提供更高的性能、更好的可维护性和更简单的部署流程。

## 快速开始

### 1. 环境要求

- Go 1.21+
- Docker & Docker Compose
- MySQL 8.0+
- Redis 7+
- Elasticsearch 8.8+

### 2. 克隆项目

```bash
git clone <repository-url>
cd law-oa-go
```

### 3. 初始化项目

```bash
./dev.sh init
```

### 4. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库连接等信息
```

### 5. 启动开发环境

```bash
./dev.sh start
```

### 6. 访问应用

- API 地址: http://localhost:8080
- 健康检查: http://localhost:8080/health

## 开发命令

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

## 生产环境部署

### 1. 使用 Docker 部署

```bash
# 构建镜像
docker build -t law-oa-go .

# 启动服务
docker-compose up -d
```

### 2. 直接部署

```bash
# 构建应用
./dev.sh build

# 配置生产环境变量
export ENVIRONMENT=production
export PORT=8080
export DB_HOST=production-db-host
export DB_PASSWORD=secure-password
export JWT_SECRET=very-secure-secret-key

# 运行应用
./main
```

## 配置说明

### 环境变量配置

```bash
# 服务器配置
ENVIRONMENT=production
PORT=8080
LOG_LEVEL=info

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=law_oa

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Elasticsearch 配置
ES_HOST=http://localhost:9200
ES_USERNAME=
ES_PASSWORD=

# JWT 配置
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRE=3600

# 限流配置
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=60
```

### 配置文件配置

项目支持 YAML 配置文件 `config.yaml`：

```yaml
# 生产环境配置
environment: production
port: "8080"
log_level: info
enable_swagger: false

# 数据库配置
database:
  host: localhost
  port: "3306"
  username: root
  password: password
  database: law_oa
  charset: utf8mb4
  parse_time: true
  loc: Local

# Redis 配置
redis:
  host: localhost
  port: "6379"
  password: ""
  db: 0

# JWT 配置
jwt:
  secret: your-secret-key-change-in-production
  expire: 3600
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
    "case_no": "CASE2024001",
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

## 监控和日志

### 应用监控

- **健康检查**: `GET /health`
- **统计信息**: `GET /api/v1/stats/dashboard`

### 日志管理

项目使用结构化日志记录，支持以下日志级别：
- DEBUG
- INFO
- WARN
- ERROR

日志输出到控制台，支持 JSON 格式。

## 性能优化

### 缓存策略

- 使用 Redis 缓存热点数据
- 优化数据库查询
- 使用 Elasticsearch 进行全文搜索

### 数据库优化

- 使用 GORM 进行数据库操作
- 支持数据库连接池
- 自动数据库迁移

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

4. **JWT 认证失败**
   - 检查 JWT 密钥配置
   - 验证令牌格式
   - 确认令牌未过期

### 日志查看

```bash
# 查看应用日志
./dev.sh run

# 查看 Docker 日志
docker-compose logs -f app

# 查看 Nginx 日志
docker-compose logs -f nginx
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

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 技术栈

- **后端框架**: Gin (Go Web Framework)
- **数据库**: MySQL 8.0
- **缓存**: Redis 7
- **搜索引擎**: Elasticsearch 8.8
- **ORM**: GORM
- **认证**: JWT
- **容器化**: Docker & Docker Compose
- **代理**: Nginx

## 许可证

MIT License