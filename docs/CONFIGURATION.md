# 配置管理系统文档

## 概述

本项目当前配置加载由 `internal/config/config.go` 实现，核心是 `godotenv` + `viper`：优先加载 `.env`，再读取 `config.yaml`、`config/config.yaml` 或 `/etc/law-oa/config.yaml`，并用环境变量覆盖关键配置。

## 特性

### 🔧 核心特性
- **环境隔离**: 支持 development、production、test、staging 环境
- **类型安全**: 使用结构体标签确保配置类型安全
- **验证机制**: 内置配置验证和错误检查
- **默认值**: 智能默认值和环境特定覆盖
- **配置文件**: 支持根目录 `config.yaml`、`config/config.yaml` 和 `/etc/law-oa/config.yaml`
- **测试支持**: 完整的测试覆盖

### 🌟️ 环境特定优化
- **开发环境**: 调试模式、详细日志、宽松的安全限制
- **生产环境**: 优化性能、严格的安全配置、监控启用
- **测试环境**: 内存数据库、禁用外部服务、快速执行

## 配置结构

### 主配置结构

```go
type Config struct {
    Environment       string
    Port              string
    Database          DatabaseConfig
    Redis             RedisConfig
    Elasticsearch     ElasticsearchConfig
    JWT               JWTConfig
    Log               LogConfig
    CORS              CORSConfig
    ConflictDetection *ConflictDetectionConfig
}
```

### 子配置结构

#### 服务器配置
服务端口在顶层 `port` / `PORT` 上配置，默认 `8080`。

#### 数据库配置
```go
type DatabaseConfig struct {
    Driver   string // DB_DRIVER, default postgres
    Host     string // DB_HOST
    Port     string // DB_PORT
    Username string // DB_USERNAME
    Password string // DB_PASSWORD
    Database string // DB_DATABASE
    SSLMode  string
}
```

生产安装目前只支持 PostgreSQL schema bootstrap。`DB_DRIVER=postgres`、`DB_HOST`、`DB_PORT`、`DB_USERNAME`、`DB_PASSWORD`、`DB_DATABASE` 和 `DB_SSLMODE=require` 是生产数据库连接的最小配置；MySQL/SQLite 兼容代码不等于已经通过生产迁移验收。

## 环境配置文件

### 目录结构
```
.env.example
config.yaml
config/config.yaml
```

### 环境文件示例

#### 开发环境 (.env.development)
```bash
ENVIRONMENT=development
PORT=8080
LOG_LEVEL=debug

DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=law_oa_user
DB_PASSWORD=your_secure_password
DB_DATABASE=law_oa_dev

JWT_SECRET=your-super-secret-jwt-key-for-development-32-chars-min
```

#### 生产环境 (.env.production)
```bash
APP_ENV=production
DEBUG=false
LOG_LEVEL=warn

SERVER_HOST=0.0.0.0
SERVER_PORT=8080

DB_DRIVER=postgres
DB_HOST=${DB_HOST}
DB_PORT=5432
DB_SSLMODE=require
DB_USERNAME=${DB_USERNAME}
DB_PASSWORD=${DB_PASSWORD}
DB_DATABASE=${DB_DATABASE}

JWT_SECRET=${JWT_SECRET}
APP_SECRET=${APP_SECRET}
SUBJECT_DATA_KEY=${SUBJECT_DATA_KEY}
```

生产环境启动前还必须配置真实的 `CORS_ALLOWED_ORIGINS`。`SUBJECT_DATA_KEY` 必须解码为 32 字节，并与 `JWT_SECRET`、`APP_SECRET` 使用不同的密钥；它用于保护案件主体身份标识，丢失会使历史主体变更无法解密。后端还会检查权威档案覆盖登记，未完成时 `/health/ready` 不会就绪。

## 配置管理器

### 基本用法

```go
import "law-oa-go/internal/config"

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 应用逻辑...
}
```

当前代码未实现热加载监听器；变更配置后重启服务生效。

## 使用指南

### 加载配置

```go
cfg, err := config.Load()
```

### 环境切换

```bash
# 切换到生产环境
./scripts/config-simple.sh switch production

# 切换到测试环境
./scripts/config-simple.sh switch test
```

### 配置验证

```go
err := config.Validate()
if err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}
```

## 配置管理脚本

### 可用命令

```bash
./scripts/config-simple.sh help          # 显示帮助
./scripts/config-simple.sh env           # 显示当前环境
./scripts/config-simple.sh switch <env>  # 切换环境
./scripts/config-simple.sh generate <env> # 生成配置模板
./scripts/config-simple.sh list          # 列出可用环境
./scripts/config-simple.sh test          # 测试配置加载
```

### 示例用法

```bash
# 1. 查看当前环境
./scripts/config-simple.sh env

# 2. 生成新环境配置
./scripts/config-simple.sh generate staging

# 3. 切换环境
./scripts/config-simple.sh switch production

# 4. 测试配置
./scripts/config-simple.sh test

# 5. 列出所有环境
./scripts/config-simple.sh list
```

## 测试

### 运行配置测试

```bash
go test ./internal/config -v
```

### 测试覆盖

- ✅ 配置加载测试
- ✅ 配置验证测试
- ✅ 环境覆盖测试
- ✅ 配置覆盖测试
- ✅ 字段设置测试

## 最佳实践

### 1. 环境变量命名
- 使用大写字母和下划线
- 使用有意义的名称
- 为敏感信息使用环境变量替换

### 2. 配置安全
- 不在代码中硬编码敏感信息
- 使用 `.env.local` 文件覆盖敏感配置
- 生产环境使用环境变量或密钥管理服务

### 3. 配置验证
- 使用必需字段标记 (`required`)
- 验证配置格式和范围
- 提供有意义的错误信息

### 4. 默认值策略
- 为开发环境提供合理的默认值
- 生产环境要求显式配置
- 测试环境使用最小化配置

## 故障排除

### 常见问题

#### 配置加载失败
```bash
# 检查必需环境变量
./scripts/config-simple.sh env

# 验证配置文件
./scripts/config-simple.sh test
```

#### 环境变量问题
```bash
# 检查环境变量是否设置
env | grep -E "APP_ENV|DB_|JWT_"

# 加载环境文件
source .env
```

#### 配置类型错误
```bash
# 查看详细错误信息
go test ./internal/config -v
```

## 开发指南

### 添加新配置项

1. 在 `internal/config/config.go` 的 `Config` 或子配置结构中添加字段
2. 在 `Load()` 中添加默认值和环境变量绑定
3. 必要时更新 `Validate()`、`Get*()` helper 和测试
4. 更新 `.env.example`、`config.yaml` 和本文档

### 添加新环境

1. 创建或复制对应 `.env` / `config.yaml`
2. 设置 `ENVIRONMENT` 和数据库、Redis、ES、JWT 等核心变量
3. 用目标入口启动服务验证 `config.Load()`

### 配置热加载

当前代码未实现配置热加载；修改配置后重启后端服务。

## 版本历史

- **v2.4.0**: 2026-04-29 校准为当前 `godotenv` + `viper` 实现
- **v1.0.0**: 基于 `viper` 的配置系统

## 参考资料

- [Go 环境变量最佳实践](https://12factor.net/)
- [配置管理设计模式](https://12factor.net/config)
