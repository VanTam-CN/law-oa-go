# Law OA Go 开发环境搭建指南

**版本**: v2.1.0
**更新日期**: 2025-09-30
**适用系统**: macOS, Linux, Windows

---

## 📋 概述

本文档提供Law OA Go项目开发环境的完整搭建指南，包括系统要求、依赖安装、项目配置、开发工具设置等内容。

---

## 🖥️ 系统要求

### 最低配置
- **操作系统**: macOS 10.15+, Ubuntu 18.04+, Windows 10+
- **CPU**: 2核心 2.0GHz
- **内存**: 8GB RAM
- **存储**: 20GB 可用空间
- **网络**: 稳定的互联网连接

### 推荐配置
- **操作系统**: macOS 12+, Ubuntu 20.04+, Windows 11
- **CPU**: 4核心 2.5GHz
- **内存**: 16GB RAM
- **存储**: 50GB SSD
- **网络**: 宽带连接

---

## 🛠️ 必需软件安装

### 1. Git 版本控制

#### macOS
```bash
# 使用 Homebrew 安装
brew install git

# 或下载安装包
# https://git-scm.com/download/mac
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install git
```

#### Windows
```bash
# 使用 Chocolatey 安装
choco install git

# 或下载安装包
# https://git-scm.com/download/win
```

#### 配置 Git
```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
git config --global init.defaultBranch main
git config --global pull.rebase false
```

### 2. Go 开发环境

#### 安装 Go 1.23

**macOS (使用 Homebrew)**:
```bash
brew install go@1.23
```

**Linux**:
```bash
# 下载 Go 1.23
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz

# 解压到 /usr/local
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# 添加到 PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc
```

**Windows**:
1. 下载安装包：https://go.dev/dl/go1.23.0.windows-amd64.msi
2. 运行安装程序，按默认设置安装
3. 重启命令提示符

#### 验证 Go 安装
```bash
go version
# 应该输出: go version go1.23.0 linux/amd64

go env GOPATH
# 应该输出: /home/username/go (Linux) 或 C:\Users\username\go (Windows)
```

### 3. 数据库

#### MySQL 8.0

**macOS (使用 Homebrew)**:
```bash
brew install mysql@8.0
brew services start mysql@8.0

# 安全配置
mysql_secure_installation
```

**Linux (Ubuntu)**:
```bash
sudo apt update
sudo apt install mysql-server-8.0
sudo systemctl start mysql
sudo systemctl enable mysql

# 安全配置
sudo mysql_secure_installation
```

**Windows**:
1. 下载 MySQL Installer：https://dev.mysql.com/downloads/installer/
2. 运行安装程序，选择 MySQL Server 8.0
3. 设置 root 密码并启动服务

#### 创建开发数据库
```sql
-- 登录 MySQL
mysql -u root -p

-- 创建数据库
CREATE DATABASE law_oa_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建开发用户
CREATE USER 'law_oa_dev'@'localhost' IDENTIFIED BY 'dev_password';
GRANT ALL PRIVILEGES ON law_oa_dev.* TO 'law_oa_dev'@'localhost';
FLUSH PRIVILEGES;
```

### 4. Redis

**macOS**:
```bash
brew install redis
brew services start redis
```

**Linux**:
```bash
sudo apt install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

**Windows**:
```bash
# 使用 Chocolatey
choco install redis-64

# 或使用 WSL
wsl --install
sudo apt update
sudo apt install redis-server
```

### 5. Node.js (前端开发)

**使用 nvm (推荐)**:
```bash
# 安装 nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash

# 重启终端或运行
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# 安装 Node.js 18 LTS
nvm install 18
nvm use 18
nvm alias default 18
```

**直接安装**:
```bash
# macOS
brew install node@18

# Linux
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
```

---

## 💻 项目设置

### 1. 克隆项目

```bash
# 克隆项目仓库
git clone https://github.com/your-org/law-oa-go.git
cd law-oa-go

# 查看项目结构
ls -la
```

### 2. 安装 Go 依赖

```bash
# 下载依赖模块
go mod download

# 验证依赖
go mod verify

# 整理依赖
go mod tidy
```

### 3. 安装开发工具

```bash
# 安装 golangci-lint (代码检查工具)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装 gosec (安全检查工具)
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# 安装 mockgen (Mock 生成工具)
go install github.com/golang/mock/mockgen@latest

# 安装 swag (API 文档生成)
go install github.com/swaggo/swag/cmd/swag@latest

# 安装 migrate (数据库迁移工具)
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 验证工具安装
golangci-lint version
gosec -version
mockgen -version
swag -version
migrate -version
```

### 4. 配置开发环境

#### 创建开发配置文件
```bash
# 复制配置模板
cp config/config.example.yaml config/dev.yaml
cp .env.example .env.dev
```

#### 编辑开发配置
```yaml
# config/dev.yaml
app:
  name: "Law OA Go"
  version: "2.1.0"
  env: "development"
  port: 8080
  debug: true

database:
  host: "localhost"
  port: 3306
  name: "law_oa_dev"
  username: "law_oa_dev"
  password: "dev_password"
  charset: "utf8mb4"
  parse_time: true
  loc: "Local"
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: "1h"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 5

jwt:
  secret: "dev-secret-key-change-in-production"
  expiry: "24h"
  refresh_expiry: "168h"

logging:
  level: "debug"
  format: "console"
  output: "stdout"

security:
  cors_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
  cors_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  rate_limit:
    rps: 1000
    burst: 2000

upload:
  max_size: 10485760  # 10MB
  storage_path: "./uploads"

monitoring:
  enabled: true
  prometheus:
    enabled: true
    path: "/metrics"
  health_check:
    enabled: true
    path: "/health"
  pprof:
    enabled: true
    path: "/debug/pprof"
```

#### 编辑环境变量
```bash
# .env.dev
APP_ENV=development
APP_PORT=8080

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_NAME=law_oa_dev
DB_USER=law_oa_dev
DB_PASSWORD=dev_password

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT配置
JWT_SECRET=dev-secret-key-change-in-production
JWT_EXPIRY=24h

# 日志配置
LOG_LEVEL=debug
LOG_FORMAT=console

# 开发模式
DEBUG=true
ENABLE_PPROF=true
```

### 5. 数据库迁移

```bash
# 运行数据库迁移
migrate -path migrations -database "mysql://law_oa_dev:dev_password@tcp(localhost:3306)/law_oa_dev" up

# 或者使用应用内置迁移
go run cmd/server/main.go migrate up --config config/dev.yaml

# 创建初始数据
mysql -u law_oa_dev -p law_oa_dev < scripts/create_test_data.sql
```

---

## 🏃 运行项目

### 1. 启动后端服务

```bash
# 开发模式运行
go run cmd/server/main.go --config config/dev.yaml

# 或使用 Makefile
make dev

# 或使用 air (热重载)
# 安装 air
go install github.com/cosmtrek/air@latest

# 运行热重载
air -c .air.toml
```

### 2. 验证服务运行

```bash
# 检查健康状态
curl http://localhost:8080/health

# 检查 API 文档
curl http://localhost:8080/swagger/index.html

# 检查指标
curl http://localhost:8080/metrics
```

### 3. 启动前端开发服务器

```bash
cd frontend-vue

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 或使用 yarn
yarn install
yarn dev
```

---

## 🔧 开发工具配置

### 1. VS Code 配置

#### 安装推荐扩展
```bash
# 通过 VS Code UI 安装
# - Go (官方)
# - ESLint
# - Prettier
# - GitLens
# - Docker
# - Thunder Client (API 测试)
# - Remote Development
```

#### 工作区设置
```json
// .vscode/settings.json
{
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "go.testOnSave": true,
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    },
    "go.buildTags": "",
    "go.buildFlags": [],
    "go.testFlags": ["-v", "-race"],
    "go.testTimeout": "30s",
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    },
    "files.exclude": {
        "**/*.exe": true,
        "**/law-oa-server": true,
        "**/vendor": true
    }
}
```

#### 调试配置
```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Law OA Server",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/server/main.go",
            "args": ["--config", "config/dev.yaml"],
            "env": {
                "APP_ENV": "development"
            },
            "console": "integratedTerminal"
        },
        {
            "name": "Debug Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}",
            "env": {
                "APP_ENV": "test"
            },
            "args": [
                "-test.v",
                "-test.run", "^${input:testName}$"
            ]
        }
    ],
    "inputs": [
        {
            "id": "testName",
            "description": "Test name pattern",
            "default": "Test",
            "type": "promptString"
        }
    ]
}
```

#### 任务配置
```json
// .vscode/tasks.json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "go: build",
            "type": "shell",
            "command": "go",
            "args": ["build", "-o", "bin/law-oa-server", "./cmd/server"],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            },
            "problemMatcher": ["$go"]
        },
        {
            "label": "go: test",
            "type": "shell",
            "command": "go",
            "args": ["test", "-v", "./..."],
            "group": "test",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            },
            "problemMatcher": ["$go"]
        },
        {
            "label": "go: lint",
            "type": "shell",
            "command": "golangci-lint",
            "args": ["run"],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        }
    ]
}
```

### 2. Git 配置

#### .gitignore
```gitignore
# 二进制文件
*.exe
*.exe~
*.dll
*.so
*.dylib
law-oa-server
law-oa-server.exe

# 测试二进制文件
*.test

# 覆盖率文件
*.out
coverage.html

# 依赖目录
vendor/

# Go 工作空间文件
go.work
go.work.sum

# IDE 文件
.vscode/
.idea/
*.swp
*.swo
*~

# 操作系统文件
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# 配置文件（包含敏感信息）
.env.local
.env.production
config/local.yaml
config/production.yaml

# 日志文件
*.log
logs/

# 上传文件
uploads/
temp/

# 数据库文件
*.db
*.sqlite
*.sqlite3

# 临时文件
tmp/
temp/

# Node.js (前端)
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# 构建输出
dist/
build/
```

#### Pre-commit Hooks
```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "运行 pre-commit 检查..."

# 检查代码格式
if ! gofmt -l . | grep -q .; then
    echo "✅ 代码格式检查通过"
else
    echo "❌ 代码格式检查失败，请运行 gofmt -w ."
    exit 1
fi

# 静态分析
if ! golangci-lint run; then
    echo "❌ 静态分析检查失败"
    exit 1
fi

# 运行测试
if ! go test -race ./...; then
    echo "❌ 测试失败"
    exit 1
fi

echo "✅ 所有检查通过"
```

### 3. Makefile

```makefile
# Makefile
.PHONY: help build test lint clean dev docker-build docker-run migrate-up migrate-down

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  dev         - 开发模式运行"
	@echo "  build       - 构建应用"
	@echo "  test        - 运行测试"
	@echo "  lint        - 代码检查"
	@echo "  clean       - 清理构建文件"
	@echo "  migrate-up  - 数据库迁移"
	@echo "  migrate-down- 回滚数据库"
	@echo "  docker-build- 构建 Docker 镜像"
	@echo "  docker-run  - 运行 Docker 容器"

# 开发模式
dev:
	go run cmd/server/main.go --config config/dev.yaml

# 构建应用
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-w -s -X main.version=$(shell git describe --tags --always) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" \
		-o bin/law-oa-server cmd/server/main.go

# 运行测试
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 代码检查
lint:
	golangci-lint run
	gosec ./...

# 清理
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

# 数据库迁移
migrate-up:
	migrate -path migrations -database "mysql://law_oa_dev:dev_password@tcp(localhost:3306)/law_oa_dev" up

migrate-down:
	migrate -path migrations -database "mysql://law_oa_dev:dev_password@tcp(localhost:3306)/law_oa_dev" down

# Docker 操作
docker-build:
	docker build -t law-oa-go:dev .

docker-run:
	docker run -p 8080:8080 --rm law-oa-go:dev

# 生成 API 文档
docs:
	swag init -g cmd/server/main.go -o docs/swagger

# 生成 Mock 文件
mocks:
	mockgen -source=internal/repositories/interfaces.go -destination=internal/repositories/mocks/interfaces.go
	mockgen -source=internal/services/interfaces.go -destination=internal/services/mocks/interfaces.go

# 安装开发工具
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/cosmtrek/air@latest
```

---

## 🧪 测试环境

### 1. 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/services

# 运行测试并显示覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 运行基准测试
go test -bench=. -benchmem ./...

# 运行测试并生成报告
go test -v -race -coverprofile=coverage.out ./... 2>&1 | tee test.log
```

### 2. 集成测试

```bash
# 启动测试数据库
docker run -d --name mysql-test -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=law_oa_test -p 3307:3306 mysql:8.0

# 运行集成测试
DB_HOST=localhost DB_PORT=3307 DB_NAME=law_oa_test go test -tags=integration ./tests/integration/...

# 清理测试容器
docker stop mysql-test && docker rm mysql-test
```

### 3. API 测试

```bash
# 使用 curl 测试
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 使用 httpie (更友好的 HTTP 客户端)
http POST localhost:8080/api/v1/auth/login \
  username=admin password=admin123

# 使用 Postman
# 导入 API 文档: http://localhost:8080/swagger/doc.json
```

---

## 🐳 Docker 开发环境

### 1. 开发容器

```dockerfile
# Dockerfile.dev
FROM golang:1.23-alpine

# 安装必要的工具
RUN apk add --no-cache git make curl bash

# 设置工作目录
WORKDIR /app

# 安装 air (热重载)
RUN go install github.com/cosmtrek/air@latest

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["air", "-c", ".air.toml"]
```

### 2. Docker Compose 开发环境

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    volumes:
      - .:/app
      - /app/vendor
    environment:
      - APP_ENV=development
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - law-oa-dev

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=root
      - MYSQL_DATABASE=law_oa_dev
      - MYSQL_USER=law_oa_dev
      - MYSQL_PASSWORD=dev_password
    ports:
      - "3307:3306"
    volumes:
      - mysql_dev_data:/var/lib/mysql
      - ./scripts:/docker-entrypoint-initdb.d
    networks:
      - law-oa-dev
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      timeout: 20s
      retries: 10

  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    volumes:
      - redis_dev_data:/data
    networks:
      - law-oa-dev
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      timeout: 10s
      retries: 5

  adminer:
    image: adminer
    ports:
      - "8081:8080"
    networks:
      - law-oa-dev

volumes:
  mysql_dev_data:
  redis_dev_data:

networks:
  law-oa-dev:
    driver: bridge
```

### 3. 运行 Docker 开发环境

```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 查看日志
docker-compose -f docker-compose.dev.yml logs -f app

# 停止环境
docker-compose -f docker-compose.dev.yml down

# 重建镜像
docker-compose -f docker-compose.dev.yml build --no-cache app
```

---

## 🔍 调试和故障排除

### 1. 常见问题

#### 端口冲突
```bash
# 查看端口占用
lsof -i :8080
netstat -tulpn | grep :8080

# 杀死占用端口的进程
kill -9 <PID>
```

#### 数据库连接失败
```bash
# 检查 MySQL 服务状态
sudo systemctl status mysql
brew services list | grep mysql

# 测试连接
mysql -h localhost -u law_oa_dev -p law_oa_dev

# 重启 MySQL
sudo systemctl restart mysql
brew services restart mysql@8.0
```

#### Go 模块问题
```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download && go mod tidy

# 检查模块依赖
go mod graph
go mod why <package>
```

### 2. 性能分析

#### CPU 性能分析
```bash
# 启动 pprof
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.pprof

# 分析结果
go tool pprof -http=:8081 cpu.pprof
```

#### 内存分析
```bash
# 获取内存堆信息
curl http://localhost:8080/debug/pprof/heap > heap.pprof

# 分析内存使用
go tool pprof -http=:8081 heap.pprof
```

#### Goroutine 分析
```bash
# 获取 Goroutine 信息
curl http://localhost:8080/debug/pprof/goroutine > goroutine.pprof

# 分析 Goroutine
go tool pprof -http=:8081 goroutine.pprof
```

### 3. 日志分析

```bash
# 实时查看应用日志
tail -f logs/app.log

# 搜索错误日志
grep -i error logs/app.log

# 分析访问日志
awk '{print $1}' logs/access.log | sort | uniq -c | sort -nr
```

---

## 📚 学习资源

### 官方文档
- [Go 官方文档](https://golang.org/doc/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [MySQL 文档](https://dev.mysql.com/doc/)
- [Redis 文档](https://redis.io/documentation)

### 推荐教程
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go 语言设计与实现](https://draveness.me/golang/)

### 开发工具
- [GoLand](https://www.jetbrains.com/go/) - JetBrains IDE
- [VS Code](https://code.visualstudio.com/) - 轻量级编辑器
- [Postman](https://www.postman.com/) - API 测试工具
- [DBeaver](https://dbeaver.io/) - 数据库管理工具

---

## 📞 技术支持

### 项目相关问题
- **项目仓库**: https://github.com/your-org/law-oa-go
- **Issues**: https://github.com/your-org/law-oa-go/issues
- **Wiki**: https://github.com/your-org/law-oa-go/wiki

### 开发团队
- **技术负责人**: tech-lead@law-oa.com
- **开发团队**: dev-team@law-oa.com
- **DevOps 团队**: devops@law-oa.com

### 即时支持
- **企业微信群**: Law OA Go 开发群
- **Slack 频道**: #law-oa-go-dev
- **技术论坛**: https://forum.law-oa.com

---

**文档版本**: v2.1.0
**最后更新**: 2025-09-30
**下次审查**: 2025-12-30
**维护团队**: 开发团队