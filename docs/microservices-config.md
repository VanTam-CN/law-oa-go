# 🚫 文档已标记为不正确 - 项目实际为单体架构

## ⚠️ 重要提醒

**本文档描述的微服务配置与项目实际情况完全不符。项目目前采用的是单体架构，不存在服务间通信配置。**

保留此文档仅作为未来架构演进参考，请勿将其作为当前项目配置的准确描述。

## 当前实际配置

### 应用配置
- **配置文件**: `config/config.yaml`
- **环境变量**: 通过`.env`文件管理
- **数据库配置**: 单一数据库连接
- **认证配置**: JWT密钥配置
- **日志配置**: 统一日志格式和输出

### 配置结构
```yaml
# config/config.yaml
app:
  name: "law-oa-go"
  version: "1.0.0"
  debug: true

database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  database: "law_oa"
  username: "root"
  password: ""

server:
  port: 8080
  timeout: 30s

jwt:
  secret: "your-secret-key"
  expires_in: 24h

logging:
  level: "info"
  format: "json"
```

---

## 原文档内容（仅供参考）

以下是原始文档内容，**请勿用于当前项目**：

### gRPC服务定义
- 用户服务proto文件
- 案件服务proto文件
- 客户服务proto文件

### 服务间通信
- gRPC配置
- 消息队列配置
- 服务发现配置

**注意：上述内容为规划中的配置，并非当前实现。**