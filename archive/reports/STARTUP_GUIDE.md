# 🚀 Law OA Go 快速启动指南

## ✅ 问题解决

如果您之前遇到脚本执行错误，现在已经解决了！我们创建了两个版本的启动脚本，并且已经修复了Docker Compose兼容性问题：

### 📝 脚本版本说明

1. **start.sh** - 完整功能版本（推荐）
   - 支持多环境切换
   - 详细的日志输出
   - 高级功能（健康检查、性能监控等）

2. **start-dev.sh** - 简化版本（兼容性更好）
   - 专为zsh和bash优化
   - 简洁易用
   - 基础功能完整
   - ✅ **已修复Docker Compose兼容性问题**

### 🔧 Docker Compose兼容性修复

**原问题**: `docker-compose: command not found`
**解决方案**: 脚本现在自动检测并使用适当的Docker Compose命令：
- 优先使用 `docker-compose`（独立安装版本）
- 如果不可用，自动切换到 `docker compose`（Docker CLI插件）
- 智能检测，无需手动配置

## 🎯 快速开始

### 使用简化版本（推荐）
```bash
# 启动开发环境
./start-dev.sh

# 或者明确指定命令
./start-dev.sh start

# 查看服务状态
./start-dev.sh status

# 查看日志
./start-dev.sh logs

# 停止服务
./start-dev.sh stop
```

### 使用完整版本
```bash
# 启动开发环境
./start.sh dev

# 启动生产环境
./start.sh prod

# 查看帮助
./start.sh help
```

## 🔧 服务访问地址

启动成功后，可以通过以下地址访问：

| 服务 | 地址 | 功能 |
|------|------|------|
| 🎨 前端应用 | http://localhost:3003 | React前端界面 |
| 🔧 后端API | http://localhost:8080 | Go后端服务 |
| ❤️ 健康检查 | http://localhost:8080/health | 服务健康状态 |
| 📊 监控指标 | http://localhost:8080/metrics | Prometheus指标 |
| 🗄️ MySQL | localhost:33060 | 数据库服务 |
| 🔴 Redis | localhost:6379 | 缓存服务 |
| 🔍 搜索 | http://localhost:9200 | Elasticsearch |

## 🛠️ 常用命令

```bash
# 查看所有可用命令
./start-dev.sh help

# 启动所有服务
./start-dev.sh start

# 重启服务
./start-dev.sh restart

# 查看服务状态
./start-dev.sh status

# 查看实时日志
./start-dev.sh logs

# 清理数据（谨慎使用）
./start-dev.sh clean
```

## 🚨 故障排除

### 如果脚本无法执行：
```bash
# 检查文件权限
ls -la start-dev.sh

# 如果没有执行权限，添加权限
chmod +x start-dev.sh

# 再次尝试
./start-dev.sh help
```

### 如果Docker相关问题：
```bash
# 检查Docker状态
docker --version
docker-compose --version

# 如果Docker未启动
# macOS: 打开Docker Desktop应用
# Linux: sudo systemctl start docker
```

### 如果端口被占用：
```bash
# 检查端口占用
lsof -i :8080
lsof -i :3003

# 停止占用端口的服务
./start-dev.sh stop
```

### 如果服务启动失败：
```bash
# 查看详细错误日志
./start-dev.sh logs

# 清理并重新启动
./start-dev.sh clean
./start-dev.sh start
```

## 📋 系统要求

- Docker Desktop 4.0+
- 8GB+ RAM
- 10GB+ 可用磁盘空间
- macOS 或 Linux 系统

## 💡 开发技巧

### 查看服务状态
```bash
# 快速状态检查
./start-dev.sh status

# 查看资源使用
docker stats
```

### 调试日志
```bash
# 查看所有服务日志
./start-dev.sh logs

# 查看特定服务日志
docker-compose logs backend
docker-compose logs mysql
```

### 数据管理
```bash
# 连接数据库
docker-compose exec mysql mysql -u root -p

# 查看Redis
docker-compose exec redis redis-cli
```

## 🎉 开始使用

现在您已经准备好启动Law OA Go开发环境了！

```bash
# 一键启动
./start-dev.sh

# 等待服务启动完成
# 访问 http://localhost:3003 开始使用
```

---

**🚨 如果遇到任何问题，请检查上面的故障排除部分或查看服务日志。**