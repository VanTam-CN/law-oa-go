# 配置管理系统文档

## 概述

本项目采用了现代化的配置管理系统，支持多环境隔离、热加载和配置验证。基于 `go-envconfig` 库实现，遵循 Go 最佳实践。

## 特性

### 🔧 核心特性
- **环境隔离**: 支持 development、production、test、staging 环境
- **类型安全**: 使用结构体标签确保配置类型安全
- **验证机制**: 内置配置验证和错误检查
- **默认值**: 智能默认值和环境特定覆盖
- **配置管理器**: 支持热加载和变更监听
- **测试支持**: 完整的测试覆盖

### 🌟️ 环境特定优化
- **开发环境**: 调试模式、详细日志、宽松的安全限制
- **生产环境**: 优化性能、严格的安全配置、监控启用
- **测试环境**: 内存数据库、禁用外部服务、快速执行

## 配置结构

### 主配置结构

```go
type EnhancedConfig struct {
    Environment string                    `env:"APP_ENV, default=development"`
    Debug       bool                     `env:"DEBUG, default=false"`
    LogLevel    string                   `env:"LOG_LEVEL, default=info"`
    Server      ServerConfig             `env:", prefix=SERVER_"`
    Database    DatabaseConfig           `env:", prefix=DB_"`
    Redis       RedisConfig              `env:", prefix=REDIS_"`
    Elasticsearch ElasticsearchConfig     `env:", prefix=ES_"`
    JWT         JWTConfig                 `env:", prefix=JWT_"`
    CORS        CORSConfig                `env:", prefix=CORS_"`
    Monitoring  MonitoringConfig         `env:", prefix=MONITORING_"`
    Security    SecurityConfig           `env:", prefix=SECURITY_"`
    Cache       CacheConfig              `env:", prefix=CACHE_"`
    Storage     StorageConfig            `env:", prefix=STORAGE_"`
}
```

### 子配置结构

#### 服务器配置
```go
type ServerConfig struct {
    Host           string        `env:"HOST, default=localhost"`
    Port           int           `env:"PORT, default=8080"`
    ReadTimeout    time.Duration `env:"READ_TIMEOUT, default=30s"`
    WriteTimeout   time.Duration `env:"WRITE_TIMEOUT, default=30s"`
    IdleTimeout    time.Duration `env:"IDLE_TIMEOUT, default=60s"`
    GracefulShutdownTimeout time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT, default=30s"`
}
```

#### 数据库配置
```go
type DatabaseConfig struct {
    Diver       string        `env:"DRIVER, default=postgres"`
    Host        string        `env:"HOST, default=localhost"`
    Port        int           `env:"PORT, default=5432"`
    Username    string        `env:"USERNAME, required"`
    Password    string        `env:"PASSWORD, required"`
    Database    string        `env:"NAME, required"`
    SSLMode     string        `env:"SSLMODE, default=disable"`
    MaxOpenConns int          `env:"MAX_OPEN_CONNS, default=25"`
    MaxIdleConns int          `env:"MAX_IDLE_CONNS, default=5"`
    ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME, default=5m"`
}
```

## 环境配置文件

### 目录结构
```
configs/environments/
├── .env.development
├── .env.production
├── .env.test
└── .env.staging
```

### 环境文件示例

#### 开发环境 (.env.development)
```bash
APP_ENV=development
DEBUG=true
LOG_LEVEL=debug

SERVER_HOST=localhost
SERVER_PORT=8080

DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=law_oa_user
DB_PASSWORD=your_secure_password
DB_NAME=law_oa_dev

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
DB_USERNAME=${DB_USERNAME}
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=${DB_NAME}

JWT_SECRET=${JWT_SECRET}
```

## 配置管理器

### 基本用法

```go
import (
    "law-oa-go/internal/config"
)

func main() {
    ctx := context.Background()

    // 创建配置管理器
    manager := config.NewManager(nil)

    // 加载配置
    err := manager.Load(ctx, ".env")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 获取配置
    cfg := manager.GetConfig()

    // 启动配置监听
    go func() {
        if err := manager.WatchConfig(ctx); err != nil {
            log.Printf("Config watcher error: %v", err)
        }
    }()

    // 应用逻辑...
}
```

### 配置监听器

```go
// 添加配置变更监听器
manager.AddWatcher(func(newConfig *config.EnhancedConfig) {
    log.Printf("Configuration changed: %+v", newConfig.Debug)

    // 重新初始化依赖配置的组件
    if err := reinitializeServices(newConfig); err != nil {
        log.Printf("Failed to reinitialize services: %v", err)
    }
})
```

## 使用指南

### 加载配置

```go
// 方式1: 使用配置管理器
manager := config.NewManager(logger)
err := manager.Load(ctx, ".env")

// 方式2: 直接加载
config, err := config.LoadEnhancedConfig(ctx)
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

1. 在 `EnhancedConfig` 结构中添加字段
2. 添加相应的 `env` 标签
3. 设置合适的默认值
4. 更新测试用例
5. 更新环境配置文件模板

### 添加新环境

1. 创建 `configs/environments/.env.<env_name>` 文件
2. 添加到验证函数中的有效环境列表
3. 更新配置生成脚本
4. 测试环境切换

### 配置热加载

```go
// 添加配置监听器
manager.AddWatcher(func(config *config.EnhancedConfig) {
    // 处理配置变更
})
```

## 版本历史

- **v2.0.0**: 实现基于 `go-envconfig` 的现代化配置系统
- **v1.0.0**: 基于 `viper` 的原始配置系统

## 参考资料

- [go-envconfig 文档](https://github.com/sethvargo/go-envconfig)
- [Go 环境变量最佳实践](https://12factor.net/)
- [配置管理设计模式](https://12factor.net/config)