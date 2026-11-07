# 客户创建幂等独立切片 N1/N2 交付证据

结论：已实际实现客户后端独立切片。指定基线加本次文件的隔离验证通过；共享工作树整体服务测试仍受其他切片未完成契约阻断。此结论不是整个接案协作 MVP 验收，也不是生产或真实律师试用结论。

## 基线、范围和文件

- 工作树：`/private/tmp/lawoa-intake-collaboration-20261107`。
- 基线：`a806831a7cef839c2b971cd4ab260cce77f42b8b`。
- 只读核对了本树 AGENTS、方案第4节与 C02/C03；主树未提交客户实现仅作参考。未改主树，未复制其整目录。
- 修改：`internal/handlers/client_handler.go`、`internal/repositories/client_repository.go`。
- 新增：`internal/services/client_idempotency.go`、`internal/models/client_creation_request.go`。
- 新增专项测试：`internal/services/client_idempotency_test.go`、`internal/handlers/client_idempotency_test.go`。
- 新增迁移仅：`migrations/000083_client_creation_idempotency.up.sql`、`migrations/000083_client_creation_idempotency.down.sql`。
- `internal/services/client_service.go` 无须修改：在同一事务绑定的仓储上调用原 `CreateClient`，复用校验、邮箱规则、NULL 邮箱、身份加密 hooks 和脱敏投影。
- 未改共享模型/路由、其他迁移或他人文件；未调用其他代理，未 commit/push，未安装依赖，未调用真实外部 API 或生产环境。
- 注意：仓库 `.gitignore:282` 的 `*_test.go` 会隐藏这两个新测试，后续整理提交必须显式纳入，不能只依赖普通 `git status`。

所有八个文件的 SHA-256 和验证副本位置见 `client-slice-manifest.json`；最终再次核对与验证副本逐字节一致。

## 接口与持久化契约

`POST /api/v1/clients` 保持原入口、响应外壳及成功状态 200。

1. 缺省 `Idempotency-Key` 时完全调用原创建路径，不依赖幂等表；身份选填等原有行为不变。
2. 提供请求头时，只接受单个 1—120 字节的可见 ASCII key（不含空格）；空值、超长、重复请求头或非法字符返回 `400 CLIENT_CREATE_KEY_INVALID`。key 区分大小写，不自动裁剪。
3. 作用域：认证操作者 + `create_client` + key。操作者来自认证上下文，覆盖请求体伪造的 `created_by`。
4. 规范化：字符串去首尾空白；身份类型补默认值并转大写；`identity_number` 优先，缺省时兼容 `id_card`；身份号码使用现有 `NormalizeIdentityNumber`。同一份规范化对象同时参与摘要计算和实际创建，避免摘要与落库语义分叉。
5. 同作用域、同规范化请求返回原持久化客户 ID。结果按当前客户记录和既有脱敏投影回读，不缓存明文响应；客户后续权限变化仍生效。
6. 同作用域异参返回 `409 CLIENT_CREATE_KEY_CONFLICT`。不同 key 不按名称或身份自动合并；原邮箱唯一规则继续执行。
7. 幂等表只保存操作者、操作、key HMAC、请求 HMAC、状态、客户 ID 和时间；不保存原 key、明文请求或身份号码。摘要复用 `SUBJECT_DATA_KEY` 和已有 HMAC 方法，并使用不同用途前缀。
8. 无合法身份保护密钥时返回 `503 SUBJECT_DATA_KEY_REQUIRED`，无降级哈希。依赖或数据库失败返回 `503 CLIENT_CREATE_UNAVAILABLE`，提示保留原 key 重试。原结果被删除/不完整时返回 `409 CLIENT_CREATE_RESULT_UNAVAILABLE`，不会重新创建。
9. 参数/业务校验拒绝使用 `400 CLIENT_CREATE_REJECTED` 通用文案，不暴露其他客户名称、邮箱存在性文案或数据库错误详情。
10. 创建前保留 `CanCreateClient`；所有带 key 的成功返回前再执行 `CanReadClient`，沿用对象授权和伦理墙，拒绝时不返回客户内容。

持久化并发由 `(operator_id, operation, key_digest)` 唯一约束协调。claim 插入前设置 savepoint；唯一冲突后回滚到 savepoint 并读取胜出记录，避免 PostgreSQL 的 aborted transaction。读取使用锁定当前读，兼顾 MySQL 的快照语义。新 claim、客户创建、结果 ID/完成状态全部在同一事务，任一步失败整体回滚。

迁移 000083 按候选现有 PostgreSQL 迁移边界编写；包括结果状态 CHECK、摘要长度 CHECK 和客户外键。未对真实环境执行迁移。摘要随既有身份保护密钥使用，密钥轮换及历史摘要迁移不是本切片范围。

## 行为验证及证据

| 检查 | 结果 | 证据 |
|---|---|---|
| 原有客户服务专项回归 | PASS | `03-behavior-isolated-postgres.log` |
| C02：响应丢失、重新连接/重建服务后同规范请求重试 | SQLite/PostgreSQL PASS；同一客户 ID，一条 claim | 同上 |
| C02：12 路同参并发 | SQLite/PostgreSQL PASS；全部返回同一 ID | 同上 |
| C02：12 路异参竞争 | SQLite/PostgreSQL PASS；一组6个成功，另一组6个冲突 | 同上 |
| 操作者/key/操作作用域、不同 key 不合并 | SQLite/PostgreSQL PASS | 同上 |
| 结果写入失败晚期回滚、邮箱拒绝后释放 key | SQLite/PostgreSQL PASS；无半成品或遗留 claim | 同上 |
| 缺失/无效密钥、已删除结果、无 key 兼容 | SQLite/PostgreSQL PASS | 同上 |
| HMAC、身份密文、响应及 SQL 日志无测试身份原文/原 key | PASS；断言直接检查真实落库和捕获日志 | 同上 |
| 000083 up/down/up | 真实 PostgreSQL 16.10 PASS；再创建成功 | 同上 |
| 真实 HTTP handler/service/repository 链 | PASS；200/400/401/403/409/503、重放授权及脱敏 | 同上 |
| C03：65 个同名不同身份客户及授权分页 | PASS；65 个不同 ID，分页50+15，排除另一操作者客户 | 同上 |
| 构建 | PASS，退出码0 | `04-build-isolated.log` |
| 相关包 go vet | PASS，退出码0 | `05-vet-isolated.log` |
| 新专项 race 检测 | PASS，含真实 PostgreSQL；无竞争告警 | `06-race-isolated.log` |
| 格式、差异空白、验证文件一致性 | PASS | `07-scope-format-check.log`、`client-slice-manifest.json` |
| 最后一次共享树集成检查 | 客户 HTTP 三项 PASS；services 测试编译未通过 | `08-shared-worktree-integration.log` |

使用本机 Unix socket `/private/tmp/lawoa-intake-collaboration-runtime/socket`，端口 55437，专用临时库 `lawoa_client_idempotency_20261107`，全部虚构数据。每项 PostgreSQL 测试使用独立随机 schema，结束后清理 schema。未重启或停止主代理的 PostgreSQL，专用空数据库保留供复核。

本机普通权限访问 socket 起初报 `Operation not permitted`；经工具批准的本机连接后继续完成真实 PostgreSQL 测试。Docker 探测显示 `docker.sock: connect: no such file or directory`，没有安装或启动 Docker，实际验证使用已有本机 PostgreSQL。

初次自身编译错误为 handler 错误类别常量名写错，已修正，初始红灯日志保留于 `01-compile-and-existing.log`。其后共享树审批文件编辑中出现未定义辅助函数，保留于 `02-behavior-local.log`。因此建立基线验证副本，不修改这些外部文件。

最后一次共享树阻断准确位置：

- `internal/services/inbox_workflow_contract_test.go:40:49`: `response.CompletionMode undefined (type *InboxItemResponse has no field or method CompletionMode)`。
- `internal/services/inbox_workflow_contract_test.go:44:25`: `svc.ArchiveInboxItem undefined (type *InboxService has no field or method ArchiveInboxItem)`。

这些是其他并行切片的未完成契约，未由本切片修复或排除测试。共享树整体 PASS 不成立；需主代理完成待办切片后再做整体验证。客户独立证据不依赖这些未完成改动。

## 可复现命令

验证副本由 `git archive a806831a7cef839c2b971cd4ab260cce77f42b8b` 生成，仅叠加上列八个文件，绝不读取主树未提交实现作为测试源码。位置见 `verification-snapshot.json`。

```sh
export GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache
export GOPROXY=off
export CLIENT_IDEMPOTENCY_TEST_DSN='host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_client_idempotency_20261107 sslmode=disable'
go test ./internal/services ./internal/handlers -run 'TestClientCreation|TestClientService_' -count=1 -v
go build ./...
go vet ./internal/services ./internal/handlers ./internal/repositories ./internal/models
go test -race ./internal/services ./internal/handlers -run '^TestClientCreation' -count=1 -v
```

`verification-results.json` 记录各次退出码与范围；构建和 vet 成功时日志为空是正常行为。两个新测试文件是测试运行的实际输入，未因 gitignore 排除编译。

## 未验证边界

没有执行 MySQL 实库验证、前端浏览器交互、独立人工复核、真实律师试用、生产迁移/发布或完整 C01—C18。C02/C03 仅声明上表已跑过的客户后端部分，不能代替弹窗保留草稿/新客户回填等前端验收。生产路由无需修改，现有仓储构造自动提供事务能力；实际环境须先完成迁移及既有身份保护密钥配置。
