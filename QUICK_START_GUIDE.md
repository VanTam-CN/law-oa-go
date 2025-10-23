# Law OA Go 快速启动指南

## 🚀 项目启动方式

本项目提供了多种启动方式，您可以根据需求选择：

### 1. 快速启动（推荐）

```bash
# 启动所有服务
./quick-start.sh

# 停止所有服务
./quick-start.sh stop
```

**特点：**
- 使用现有的编译文件（如果有）
- 自动检测并选择最佳启动方式
- 简单快速，适合开发测试

### 2. 本地开发启动

```bash
# 启动前后端服务
./start-local.sh start

# 查看服务状态
./start-local.sh status

# 查看日志
./start-local.sh logs

# 停止服务
./start-local.sh stop

# 重启服务
./start-local.sh restart
```

**特点：**
- 完整的开发环境配置
- 支持热重载
- 详细的日志和状态监控
- 健康检查

### 3. 分别启动

**仅启动后端：**
```bash
./start-local.sh backend
```

**仅启动前端：**
```bash
./start-local.sh frontend
```

## 📋 前置要求

### 必需软件
- **Go 1.21+**
- **Node.js 18+**
- **npm**

### 可选软件（用于数据库和缓存）
- **MySQL 8.0+**
- **Redis 6.0+**

## 🔧 环境配置

### 1. 创建环境配置文件

项目已包含 `.env.local` 文件，包含基本的开发配置：

```bash
# 应用配置
ENVIRONMENT=development
DEBUG=true
VERSION=2.1.0
PORT=8080

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=law_oa

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT配置
JWT_SECRET=your-very-secure-jwt-secret-key-for-development-only
```

### 2. 数据库准备

如果使用MySQL，请确保：

1. MySQL服务正在运行
2. 创建数据库：
   ```sql
   CREATE DATABASE law_oa CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

## 🌐 访问地址

启动成功后，您可以通过以下地址访问：

- **前端应用**: http://localhost:3003
- **后端API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health
- **API文档**: http://localhost:8080/swagger/index.html
- **监控指标**: http://localhost:8080/metrics

## 📊 服务状态检查

### 查看进程状态
```bash
# 查看Go进程
ps aux | grep "main\|law-oa"

# 查看Node.js进程
ps aux | grep "node\|npm"

# 查看端口占用
lsof -i :8080  # 后端端口
lsof -i :3003  # 前端端口
```

### 查看日志
```bash
# 实时查看后端日志
tail -f backend.log

# 实时查看前端日志
tail -f frontend.log

# 查看最近的日志
tail -20 backend.log
tail -20 frontend.log
```

## 🛑 停止服务

### 方式1：使用脚本停止
```bash
./start-local.sh stop
# 或
./quick-start.sh stop
```

### 方式2：手动停止
```bash
# 查看PID
cat .backend.pid
cat .frontend.pid

# 停止进程
kill <backend_pid> <frontend_pid>
```

### 方式3：强制停止
```bash
# 停止占用端口的进程
lsof -ti:8080 | xargs kill -9
lsof -ti:3003 | xargs kill -9
```

## 🐛 常见问题

### 1. 端口被占用
```bash
# 查看端口占用
lsof -i :8080
lsof -i :3003

# 停止占用进程
lsof -ti:8080 | xargs kill -9
lsof -ti:3003 | xargs kill -9
```

### 2. 后端启动失败
```bash
# 检查Go环境
go version

# 重新编译
go build -o main main.go

# 检查配置文件
cat .env.local
```

### 3. 前端启动失败
```bash
# 进入前端目录
cd frontend

# 重新安装依赖
npm install

# 检查Node.js版本
node --version
npm --version
```

### 4. 数据库连接失败
```bash
# 检查MySQL服务
brew services list | grep mysql  # macOS
systemctl status mysql             # Linux

# 测试连接
mysql -h localhost -u root -p

# 检查数据库是否存在
mysql -u root -p -e "SHOW DATABASES;"
```

### 5. Redis连接失败
```bash
# 检查Redis服务
redis-cli ping

# 启动Redis
brew services start redis   # macOS
systemctl start redis      # Linux
```

## 🔄 开发工作流

### 典型的开发流程：

1. **启动服务**
   ```bash
   ./quick-start.sh
   ```

2. **开发过程中**
   - 后端修改后会自动重新编译（如果使用热重载）
   - 前端修改后会自动刷新浏览器

3. **查看日志**
   ```bash
   tail -f backend.log frontend.log
   ```

4. **停止服务**
   ```bash
   ./quick-start.sh stop
   ```

## 📝 日志文件位置

- **后端日志**: `backend.log`
- **前端日志**: `frontend.log`
- **进程ID**: `.backend.pid`, `.frontend.pid`

## 🎯 下一步

启动成功后，您可以：

1. 访问前端应用进行功能测试
2. 查看API文档了解接口
3. 运行测试用例验证功能
4. 开始开发新功能

如有问题，请查看日志文件或参考本文档的常见问题部分。