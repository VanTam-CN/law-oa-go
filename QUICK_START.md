# Law OA Go 快速启动指南

## 🚀 一键启动

基于最新的Docker容器化和Kubernetes部署最佳实践，我们提供了完整的一键启动解决方案。

### 快速开始

```bash
# 克隆项目
git clone <repository-url>
cd law-oa-go

# 一键启动开发环境
./start.sh dev
```

## 📋 系统要求

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.23+ (本地开发)
- Node.js 20+ (前端开发)
- 8GB+ RAM
- 10GB+ 可用磁盘空间

## 🔧 启动命令

### 开发环境
```bash
# 启动开发环境（推荐）
./start.sh dev

# 或者分步启动
./start.sh start -e development
```

### 生产环境
```bash
# 启动生产环境
./start.sh prod

# 或者
./start.sh start -e production --health-check
```

### 其他命令
```bash
# 查看服务状态
./start.sh status

# 查看日志
./start.sh logs
./start.sh logs backend

# 重启服务
./start.sh restart

# 停止服务
./start.sh stop

# 清理数据
./start.sh clean

# 运行测试
./start.sh test

# 构建镜像
./start.sh build

# 查看帮助
./start.sh help
```

## 🌐 访问地址

启动成功后，可以通过以下地址访问服务：

| 服务 | 地址 | 说明 |
|------|------|------|
| 🎨 前端应用 | http://localhost:3003 | React前端界面 |
| 🔧 后端API | http://localhost:8080 | Go后端API |
| ❤️ 健康检查 | http://localhost:8080/health | 服务健康状态 |
| 📊 监控指标 | http://localhost:8080/metrics | Prometheus指标 |
| 🗄️ MySQL | localhost:33060 | 数据库服务 |
| 🔴 Redis | localhost:6379 | 缓存服务 |
| 🔍 Elasticsearch | http://localhost:9200 | 搜索引擎 |
| 📊 Kibana | http://localhost:5601 | 日志分析 |

## 📁 项目结构

```
law-oa-go/
├── start.sh                    # 一键启动脚本
├── Dockerfile.optimized         # 优化的Docker镜像
├── docker-compose.yml           # Docker编排配置
├── docker-compose.prod.yml      # 生产环境配置
├── .env.example                 # 环境配置模板
├── .air.toml                    # 热重载配置
├── scripts/
│   └── build.sh                 # 构建脚本
├── k8s/                         # Kubernetes配置
│   ├── namespaces/
│   ├── configmaps/
│   ├── secrets/
│   ├── deployments/
│   └── services/
├── .github/workflows/
│   └── ci-cd.yml               # CI/CD流水线
├── internal/                    # Go源码
├── frontend/                    # 前端源码
└── docs/                       # 文档
```

## 🔧 环境配置

### 1. 创建环境配置文件
```bash
# 复制环境配置模板
cp .env.example .env.local

# 根据需要修改配置
vim .env.local
```

### 2. 关键配置项

```bash
# 应用基础配置
ENVIRONMENT=development
DEBUG=true
PORT=8080

# 数据库配置
DB_HOST=mysql
DB_USER=lawuser
DB_PASSWORD=lawpass
DB_NAME=law_oa

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# JWT密钥（生产环境必须修改）
JWT_SECRET=your-secure-jwt-secret-key
```

## 🐳 Docker命令

### 构建镜像
```bash
# 构建所有镜像
docker-compose build

# 构建特定服务
docker-compose build backend

# 无缓存构建
docker-compose build --no-cache
```

### 服务管理
```bash
# 启动所有服务
docker-compose up -d

# 启动特定服务
docker-compose up -d backend mysql redis

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
docker-compose logs -f backend
```

### 数据管理
```bash
# 连接MySQL
docker-compose exec mysql mysql -u root -p

# 连接Redis
docker-compose exec redis redis-cli

# 备份数据库
docker-compose exec mysql mysqldump -u root -p law_oa > backup.sql

# 恢复数据库
docker-compose exec -T mysql mysql -u root -p law_oa < backup.sql
```

## ☸️ Kubernetes部署

### 准备工作
```bash
# 安装kubectl
# 配置kubeconfig文件

# 创建命名空间
kubectl apply -f k8s/namespaces/

# 创建配置映射
kubectl apply -f k8s/configmaps/

# 创建密钥
kubectl apply -f k8s/secrets/
```

### 部署应用
```bash
# 部署到Kubernetes
kubectl apply -f k8s/deployments/
kubectl apply -f k8s/services/
kubectl apply -f k8s/ingress/

# 查看部署状态
kubectl get pods -n law-oa
kubectl get services -n law-oa

# 查看日志
kubectl logs -f deployment/law-oa-backend -n law-oa
```

## 🔍 开发工具

### 本地开发
```bash
# 安装Air热重载工具
go install github.com/cosmtrek/air@latest

# 启动热重载
air

# 安装依赖
go mod download
go mod tidy

# 运行测试
go test ./...

# 生成测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 前端开发
```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm start

# 构建生产版本
npm run build

# 运行测试
npm test
```

## 📊 监控和调试

### 健康检查
```bash
# 检查后端服务
curl http://localhost:8080/health

# 检查前端服务
curl http://localhost:3003

# 查看详细健康信息
curl http://localhost:8080/api/v1/health/detailed
```

### 日志查看
```bash
# 实时查看所有日志
./start.sh logs

# 查看特定服务日志
./start.sh logs backend
./start.sh logs mysql

# 查看最近的日志
docker-compose logs --tail=100 backend
```

### 性能监控
```bash
# 查看资源使用情况
docker stats

# 查看系统指标
curl http://localhost:8080/metrics

# 访问监控仪表板
# Grafana: http://localhost:3001 (如果启用)
# Jaeger: http://localhost:16686 (如果启用)
```

## 🚨 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 检查端口占用
   lsof -i :8080
   lsof -i :3003

   # 停止占用端口的服务
   ./start.sh stop
   ```

2. **Docker权限问题**
   ```bash
   # 将用户添加到docker组
   sudo usermod -aG docker $USER

   # 重新登录或执行
   newgrp docker
   ```

3. **内存不足**
   ```bash
   # 增加Docker内存限制
   # 在Docker Desktop中调整内存设置到8GB+
   ```

4. **数据库连接失败**
   ```bash
   # 检查MySQL服务状态
   docker-compose ps mysql

   # 查看MySQL日志
   docker-compose logs mysql

   # 重启数据库服务
   docker-compose restart mysql
   ```

### 日志分析
```bash
# 查看错误日志
docker-compose logs backend | grep ERROR

# 查看最近的错误
docker-compose logs --tail=50 backend | grep -i error

# 查看启动日志
docker-compose logs --timestamps backend
```

### 重置环境
```bash
# 完全重置（删除所有数据）
./start.sh clean -f

# 重新启动
./start.sh dev
```

## 📚 更多文档

- [API文档](http://localhost:8080/swagger/index.html)
- [开发指南](./docs/DEVELOPMENT.md)
- [部署指南](./docs/DEPLOYMENT.md)
- [配置参考](./docs/CONFIGURATION.md)

## 🆘 获取帮助

```bash
# 查看启动脚本帮助
./start.sh help

# 查看Docker Compose帮助
docker-compose --help

# 查看项目信息
./start.sh status
```

---

**🎉 现在您可以开始使用Law OA Go系统了！**

如有问题，请查看日志文件或提交Issue。