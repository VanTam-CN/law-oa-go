# MySQL到PostgreSQL数据库迁移指南

## 概述

本指南详细说明了如何将现有的MySQL数据库迁移到PostgreSQL，并确保Go应用程序正常工作。

## 文件说明

### 1. 数据库结构文件
- `scripts/postgresql-complete-schema.sql` - 完整的PostgreSQL数据库结构
- `scripts/migrate-to-postgresql.sh` - 自动化迁移脚本

### 2. 数据模型文件
- `internal/models/complete_models.go` - 完整的Go数据模型

## 迁移步骤

### 步骤1: 准备PostgreSQL环境

1. 安装PostgreSQL（如果未安装）
```bash
# Ubuntu/Debian
sudo apt-get install postgresql postgresql-contrib

# macOS
brew install postgresql

# 启动PostgreSQL服务
sudo systemctl start postgresql  # Linux
brew services start postgresql  # macOS
```

2. 创建数据库和用户
```sql
-- 以postgres用户登录
sudo -u postgres psql

-- 创建数据库
CREATE DATABASE law_oa_go;

-- 创建用户
CREATE USER law_oa_user WITH PASSWORD 'your_password';

-- 授权
GRANT ALL PRIVILEGES ON DATABASE law_oa_go TO law_oa_user;
GRANT ALL PRIVILEGES ON ALL SCHEMAS IN DATABASE law_oa_go TO law_oa_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO law_oa_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO law_oa_user;
```

### 步骤2: 执行数据库迁移

1. 使用自动化脚本（推荐）
```bash
# 设置环境变量
export PG_HOST=localhost
export PG_PORT=5432
export PG_USER=law_oa_user
export PG_PASSWORD=your_password
export PG_DATABASE=law_oa_go

# 执行迁移
chmod +x scripts/migrate-to-postgresql.sh
./scripts/migrate-to-postgresql.sh
```

2. 或手动执行SQL文件
```bash
psql -h localhost -p 5432 -U law_oa_user -d law_oa_go -f scripts/postgresql-complete-schema.sql
```

### 步骤3: 更新Go应用程序

1. 更新数据库连接配置

在 `config/config.go` 或相应的配置文件中更新PostgreSQL连接：

```go
// config/config.go
package config

import (
    "fmt"
    "os"
)

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func GetDatabaseConfig() DatabaseConfig {
    return DatabaseConfig{
        Host:     getEnv("DB_HOST", "localhost"),
        Port:     getEnv("DB_PORT", "5432"),
        User:     getEnv("DB_USER", "law_oa_user"),
        Password: getEnv("DB_PASSWORD", ""),
        DBName:   getEnv("DB_NAME", "law_oa_go"),
        SSLMode:  getEnv("DB_SSL_MODE", "disable"),
    }
}

func GetDSN() string {
    cfg := GetDatabaseConfig()
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

2. 更新数据库初始化代码

在 `internal/database/database.go` 中更新：

```go
package database

import (
    "fmt"
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "your-project/internal/models"
    "your-project/config"
)

var DB *gorm.DB

func InitDatabase() {
    dsn := config.GetDSN()

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    DB = db

    // 自动迁移（仅用于开发环境）
    if os.Getenv("AUTO_MIGRATE") == "true" {
        autoMigrate()
    }

    log.Println("Database connected successfully")
}

func autoMigrate() {
    // 迁移完整模型
    err := DB.AutoMigrate(
        // 用户权限相关
        &models.UserComplete{},
        &models.Role{},
        &models.Permission{},
        &models.UserRole{},
        &models.RolePermission{},
        &models.Department{},

        // 业务相关
        &models.ClientComplete{},
        &models.Lawyer{},
        &models.CaseComplete{},
        &models.CaseProgress{},
        &models.CaseDocument{},

        // 冲突检测相关
        &models.LawEntity{},
        &models.LawEntityAlias{},
        &models.LawEntityRelation{},
        &models.ConflictCheckRecordComplete{},

        // 文档管理相关
        &models.Document{},
        &models.DocumentVersion{},
        &models.DocumentPermission{},
        &models.DocumentCategory{},

        // 系统管理相关
        &models.SystemConfig{},
        &models.OperationLog{},

        // 财务相关
        &models.FinancialRecord{},

        // 通知相关
        &models.Notification{},

        // 日程相关
        &models.Schedule{},

        // 分析相关
        &models.UserSession{},
        &models.PageView{},
        &models.UserEvent{},
    )

    if err != nil {
        log.Printf("Auto migration failed: %v", err)
    } else {
        log.Println("Auto migration completed")
    }
}
```

### 步骤4: 更新环境变量

在 `.env` 文件中更新：

```bash
# PostgreSQL配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=law_oa_user
DB_PASSWORD=your_password
DB_NAME=law_oa_go
DB_SSL_MODE=disable

# 应用程序配置
GIN_MODE=release
PORT=8080
JWT_SECRET=your_jwt_secret
JWT_EXPIRE=7200
```

### 步骤5: 测试应用程序

1. 重建并运行应用程序
```bash
# 构建应用程序
go build -o law-oa-go ./main.go

# 运行应用程序
./law-oa-go
```

2. 测试数据库连接
```bash
# 测试API连接
curl -X GET http://localhost:8080/api/health

# 测试用户登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## 主要差异处理

### 1. 数据类型映射

| MySQL类型 | PostgreSQL类型 | Go类型 |
|----------|---------------|---------|
| INT | INTEGER/SERIAL | int/uint |
| BIGINT | BIGINT/BIGSERIAL | int64/uint64 |
| VARCHAR | VARCHAR | string |
| TEXT | TEXT | string |
| JSON | JSONB | JSONB |
| DATETIME | TIMESTAMP | time.Time |
| BOOLEAN | BOOLEAN | bool |
| DECIMAL | DECIMAL | float64 |

### 2. 索引处理

PostgreSQL使用不同的索引语法：

```sql
-- MySQL
CREATE INDEX idx_name ON table_name(column_name);

-- PostgreSQL
CREATE INDEX IF NOT EXISTS idx_name ON table_name(column_name);

-- 全文搜索索引
CREATE INDEX IF NOT EXISTS idx_content_search ON table_name USING GIN(to_tsvector('chinese', content));
```

### 3. 自增字段

```sql
-- MySQL
id INT AUTO_INCREMENT PRIMARY KEY

-- PostgreSQL
id SERIAL PRIMARY KEY
-- 或者
id BIGSERIAL PRIMARY KEY
```

### 4. JSON字段

```sql
-- MySQL
data JSON

-- PostgreSQL
data JSONB
```

## 验证清单

### 数据库层面
- [ ] 所有表都已创建
- [ ] 所有字段类型正确
- [ ] 索引已创建
- [ ] 外键约束正确
- [ ] 触发器已创建
- [ ] 初始数据已插入

### 应用程序层面
- [ ] 数据库连接成功
- [ ] 用户认证正常
- [ ] API响应正常
- [ ] 数据CRUD操作正常
- [ ] 文件上传下载正常
- [ ] 权限控制正常

### 功能层面
- [ ] 用户管理功能
- [ ] 客户管理功能
- [ ] 案件管理功能
- [ ] 文档管理功能
- [ ] 冲突检测功能
- [ ] 搜索功能
- [ ] 报表功能

## 性能优化建议

### 1. 连接池配置

```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    MaxIdleConns: 10,
    MaxOpenConns: 100,
    ConnMaxLifetime: time.Hour,
})
```

### 2. 查询优化

```go
// 使用索引提示
db.Where("email = ?", email).First(&user)

// 批量操作
db.CreateInBatches(users, 100)

// 预加载关联
db.Preload("Client").Preload("Lawyer").Find(&cases)
```

### 3. 慢查询监控

```sql
-- 启用慢查询日志
ALTER SYSTEM SET log_min_duration_statement = 1000; -- 1秒
ALTER SYSTEM SET log_statement = 'all';
SELECT pg_reload_conf();
```

## 故障排除

### 常见问题

1. **连接失败**
   - 检查PostgreSQL服务状态
   - 验证用户权限
   - 检查防火墙设置

2. **字符编码问题**
   - 确保数据库使用UTF-8编码
   - 检查客户端连接设置

3. **时区问题**
   - 统一使用UTC或本地时间
   - 设置正确的时区配置

4. **性能问题**
   - 检查查询执行计划
   - 优化索引
   - 调整连接池大小

### 回滚方案

如果迁移出现问题，可以按以下步骤回滚：

1. 从备份恢复MySQL数据库
2. 切换应用程序配置回MySQL
3. 验证功能正常

## 维护建议

1. **定期备份**
   ```bash
   # 每日备份脚本
   pg_dump -h localhost -U law_oa_user law_oa_go > backup_$(date +%Y%m%d).sql
   ```

2. **监控日志**
   - 监控错误日志
   - 设置告警机制
   - 定期检查日志大小

3. **性能监控**
   - 监控数据库连接数
   - 监控查询响应时间
   - 定期分析慢查询

## 联系支持

如果在迁移过程中遇到问题，请：

1. 检查日志文件
2. 查看错误消息
3. 参考故障排除章节
4. 联系技术支持

---

**注意：** 本迁移指南基于项目当前状态编写，请在执行前仔细阅读并根据实际情况调整。