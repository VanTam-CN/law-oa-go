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
    JWT               JWTConfig
    Log               LogConfig
    CORS              CORSConfig
    ConflictDetection *ConflictDetectionConfig
    ExternalHealthCheck ExternalHealthCheckConfig
    OnlyOffice        OnlyOfficeConfig
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

#### 外部健康检查配置
```go
type ExternalHealthCheckConfig struct {
    Enabled bool
    URL     string
}
```

默认关闭。只有显式设置 `EXTERNAL_HEALTHCHECK_ENABLED=true` 且提供真实的 `EXTERNAL_HEALTHCHECK_URL` 时，启动流程才会注册外部 API 健康检查；`api.example.com` 这类示例域名会被配置验证拒绝。

#### OnlyOffice 配置
```go
type OnlyOfficeConfig struct {
    Enabled    bool
    URL        string
    Secret     string
    BackendURL string
}
```

OnlyOffice 默认关闭。启用时必须同时满足四项：`ONLYOFFICE_ENABLED=true`、`ONLYOFFICE_URL` 为纯 origin、`ONLYOFFICE_SECRET` 至少 32 字符、`BACKEND_URL` 为纯 origin。

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

# 当前代码未把 DB_LOC 绑定成独立环境变量；如需指定数据库时区，请写入配置文件：
# database:
#   loc: Asia/Shanghai

JWT_SECRET=${JWT_SECRET}
APP_SECRET=${APP_SECRET}
SUBJECT_DATA_KEY=${SUBJECT_DATA_KEY}
```

生产环境启动前还必须配置真实的 `CORS_ALLOWED_ORIGINS`。`SUBJECT_DATA_KEY` 必须解码为 32 字节；为避免密钥复用，建议它与 `JWT_SECRET`、`APP_SECRET`、`ONLYOFFICE_SECRET` 分离，但当前实现只对长度、编码和存在性做硬校验，没有做“不得同值”的比对门禁。后端还会检查唯一有效的双人批准冲突政策、权威档案覆盖、同步时效和数据质量登记，任一未完成时 `/health/ready` 不会就绪。登记字段和操作顺序见 [`利益冲突/production-release-policy-and-source-quality.md`](利益冲突/production-release-policy-and-source-quality.md)。

全新对方或第三人不得以自由文本直接进入正式主体版本。律师应在案件详情选择“报告主体变更并重新复核”→“登记全新主体”，提交法定名称、主体类型和可核验身份标识；系统会加密候选身份、阻断案件受控动作并向有权核查岗创建待办。核查岗在利益冲突检测清单的“新主体登记待确认”中选择新建、合并或驳回，申请律师收到结果待办后再运行主体重检。生产数据库必须执行 `000069_case_subject_revision_state_guard`；旧的 `trg_case_subject_revisions_append_only` 会阻止所有合法状态迁移，不能保留。

初次接案同样要求结构化身份。客户主档案须登记受保护的身份证件号码或统一社会信用代码；对方及相关方由律师在接案工作台登记主体类型、身份类型、身份号码和别名。迁移 `000070_case_intake_party_identity` 为接案当事人增加密文、摘要和别名字段。浏览器草稿不会保存身份号码，创建和工作台接口也只返回“已登记（受保护）”，不会返回原文、密文或摘要。缺少迁移、密钥或任一受检主体身份时，正式冲突检查失败关闭。

客户主档案的正式身份契约由 `000071_client_generic_identity` 提供：`identity_type` 明确区分身份证、护照、统一社会信用代码、营业执照等类型，`identity_number_ciphertext` 与 `identity_number_digest` 分别用于受保护存储和确定性检索。迁移会把历史 `id_card_*` 保护数据按客户类型回填；旧字段仅用于兼容回读，不得再作为企业身份的业务名称或审计依据。生产就绪检查会拒绝未完成该回填的数据库。

客户创建人、审批审计和独立联系人还要求执行 `000072_client_creator_access`、`000073_client_optional_email_null`、`000074_approval_audit_timestamps_timezone` 和 `000075_client_primary_contacts`。新建客户在首个案件产生前仅对创建律师和有权业务管理角色可见；客户公共邮箱未填写时保存为数据库 `NULL`，填写后仍保持唯一约束。主联系人姓名、职务、电话和邮箱写入独立 `client_contacts` 表，其中电话和邮箱使用 `SUBJECT_DATA_KEY` 派生的用途隔离密钥加密，禁止再把联系人职务追加到客户备注或把联系人邮箱覆盖客户公共邮箱。审批域历史无时区字段会按 `Asia/Shanghai` 解释并转换为 `timestamptz`，避免浏览器把审计时间重复增加 8 小时。中国律所部署必须在配置文件中设置 `database.loc=Asia/Shanghai`；其他司法辖区应在导入历史数据前书面确认业务时区并完成专项迁移评审。

正式冲突政策的双人确认工作流由 `000076_conflict_policy_endorsement_workflow` 提供。政策材料包与每次确认均为只追加记录；主任/管理合伙人与合规负责人必须使用两个不同账号确认同一 SHA-256 摘要，第二次有效确认后服务端才创建 `APPROVED` 政策版本。技术管理员不能代签，测试环境的虚构确认记录不得迁入生产。

四类权威档案来源可由冲突核查岗在“冲突治理”登记和复核。页面使用中文来源名称和百分比，提交时转换为服务端基点值；数据质量责任人按姓名选择。`source_version`、`index_run_id` 和核对凭证必须来自 `backfill-conflict-index --apply` 生成的导入对账报告，界面不会也不得伪造这些值。

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
2. 设置 `ENVIRONMENT`、数据库、JWT 等核心变量；Redis 仅在启用 `cache` profile 时配置
3. 用目标入口启动服务验证 `config.Load()`

### 配置热加载

当前代码未实现配置热加载；修改配置后重启后端服务。

## 版本历史

- **v2.4.0**: 2026-04-29 校准为当前 `godotenv` + `viper` 实现
- **v1.0.0**: 基于 `viper` 的配置系统

## 参考资料

- [Go 环境变量最佳实践](https://12factor.net/)
- [配置管理设计模式](https://12factor.net/config)
