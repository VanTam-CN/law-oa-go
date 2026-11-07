# 审批补正、待办事务与旧守卫专项测试交付

工作树：`/private/tmp/lawoa-intake-collaboration-20261107`，HEAD `a806831a7cef839c2b971cd4ab260cce77f42b8b`。这是共享未提交工作树上的本地测试证据，不是该 HEAD 原始内容的验收，也不是全 MVP 或生产验收。

**最终闭环版交接：34 PASS、0 FAIL、0 SKIP。** 已纳入 Freeze 的持久接案关联检查，以及成案/已完成重试在 commit 前的待办交接。最新日志为 `49-final-transition-race.jsonl` / `.log`，结果为 `52-final-transition-results-summary.json`；PG 30 项、旧 handler 守卫 4 项全部通过 `-race`，`50-final-transition-vet.log` 退出码 0。`48`/`51` 的 447 个源码、模块及迁移文件前后哈希一致。原 29 个叶子和原测试字节全部保留，仅在本人 PG 测试文件末尾追加 5 项。最后交接说明见 `53-final-handoff.md`；交接后本任务停止写入。

**编号修复阶段同源复测：29 PASS、0 FAIL、0 SKIP。** 该阶段日志为 `41-retest-race.jsonl` / `.log`，逐项结果为 `44-retest-results-summary.json`。真实 PostgreSQL 25 项和既有 handler 守卫 4 项均通过 `-race`，测试退出码 0；`42-retest-vet.log` 的 services/handlers go vet 退出码 0。编号修复复测阶段没有修改任何业务源码或测试。

`40-retest-source-before.json` 与 `43-retest-source-after.json` 由同一个执行进程紧邻测试前后自动采集，447 个 Go 源码、模块与相关迁移文件哈希完全一致。29 个叶子用例集合、四个专项测试文件与修复前红灯运行完全一致；`09`、`21` 原始红灯日志和 `24` 原结果摘要的哈希全部保持不变。

编号修复前合跑 `21-final-targeted-race.jsonl` / `.log`：29 个叶子用例，28 PASS、1 FAIL、0 SKIP；命令退出码 1。真实 PostgreSQL 25 项通过，旧 handler 守卫 3 PASS、1 FAIL。该原始红灯完整保留。四个文件 gofmt 无输出，两个既有文件 git diff --check 通过。

**审批编号碰撞已由主代理修复并复测。** 原 `generateRequestNumber` 使用日期加 `UnixNano()%1000000`，修复前 `TestIntakeConcurrentDistinctIntakesAllSucceed` 中第四个不同接案创建遇到 `UNIQUE constraint failed: approval_requests.request_number`，返回 500，违反四个并发请求都成功的断言。`09-full-real-pg-race.log` 先在真实 PG 捕获同一唯一约束 `idx_approval_requests_request_number` 的 SQLSTATE 23505。主代理改为 `AP-date-full UUID`，调用 `uuid.NewRandom()` 并返回随机源错误，48 字符未超 50 字符列长；本测试任务没有改业务编号、调度或断言。先前 `17` 偶然全通过不用于证明修复；修复后同源依据是 `40`—`44`。

## 本轮代码写入范围

| 文件 | 内容 |
| --- | --- |
| `internal/services/approval_collaboration_postgres_test.go` | 隔离 PG fixture、真实事务故障注入、补正/竞争/成案测试、HTTP 测试专用 fixture 导出 |
| `internal/services/approval_collaboration_postgres_http_test.go` | external test package 调用真实 Gin handler，首次提交/失败恢复/重复提交及双决定入口旧版本 409 |
| `internal/handlers/intake_duplicate_approval_guard_test.go` | 仅增强数据库 setup、有效申请律师、受保护客户身份、正式检查、独立复核、工作流、snapshots/inbox 等表 |
| `internal/handlers/intake_delayed_interleave_guard_test.go` | 仅将原有三份额外 intake 的 setup 改为完整独立检查/复核/主体 fixture |

没有修改业务实现、迁移、路由；没有 commit/push、子代理、依赖安装或外部 API。新 services 测试文件被仓库 `.gitignore:282 *_test.go` 忽略，复查/交接时需显式纳入文件清单；本轮没有暂存。

`14-fixture-boundary-audit.json` 与两份 `.before` 原文记录了旧守卫修改边界：router、HTTP 请求 helper、四个原测试除 clone setup 外的内容逐字节一致，原并发顺序、assertion、数量和真实创建调用全部保留。

## 覆盖与实际断言

| 对应计划 | 用例 | 实际证据 |
| --- | --- | --- |
| C11 | 普通/集成决定入口 × 真实上一轮 snapshot/上一节点 version，服务层 4 项与真实 HTTP 4 项 | 退回再重提后发送旧版本；精确错误码、HTTP 409、12 张相关表整行 JSONB 内容零变化 |
| C11 | 重提成功后的同参 replay / 异参旧退回请求 | 同参返回相同 snapshot 且零写入；异参 typed 409 且零写入；在重提故障恢复 3 项中验证 |
| C12 | 同起跑 3 轮，先重提后新建、先新建后重提各 1 轮 | 每轮恰好一个成功、失败方受控占槽错误；1 个有效审批/槽位/审批任务；胜方实际批准和成案、重试后同一完整案件及 2 个主体 |
| C12 | 真实锁重叠 1 项 | 回调只暂停真实 SQL；从 `pg_stat_activity` 和 `pg_blocking_pids` 观察 Lock 等待；新建等待 intake 锁时尚未插槽，释放后仍唯一成案 |
| C13 | 普通/集成退回 × inbox update/create 故障 4 项 | callback 确认进入真实事务、注入点确实命中；record/status/slot/inbox 等 12 表完整回滚，恢复后 1 条退回记录、1 个补充任务 |
| C13/C14 | 重提 snapshot create / inbox update / inbox create 故障 3 项 | pointer/snapshot/slot/status/inbox 等全回滚；恢复后总计 2 轮 snapshot、1 槽、1 个当前审批任务；无活动补充任务 |
| C14 | 成案主体 create 故障 1 项 | 第二轮改变主体并完成新检查/独立复核，实际审批后在真实成案事务中注入写主体失败；无残留案件/主体/假 completed；恢复和 replay 唯一完整案件，当前主体匹配第二轮快照 |
| C14 | 批准后 mutable check 两次被修改 1 项 | 第二轮批准后改 `SearchParameters.parties` 的名称、角色、身份；首次成案及 completed replay 各前改一次，仍同一案件/2 个批准主体、0 个被改身份实体；所有 snapshot 内容不变 |
| HTTP 首次提交 | 成功/重复、snapshot 写失败/原请求恢复 2 项 | 从没有审批的有效接案开始，真实 handler 获取 `approval_submitting` claim（观察 callback，不替换行为）；失败时审批/snapshot/slot/inbox 无残留，接案除合法更新时间外所有列恢复；成功唯一审批/快照/槽/待办；重复 409 零写入 |
| C13/C14 最终追加 | 首次成案/已完成重试 × inbox update/create 故障 4 项 | 真实审批先提交，再注入成案事务中的待办故障，12 表精确回滚；首次成案无残留案件/主体/假 completed，已完成重试保留原案件；恢复后唯一完整案件，结果待办保持同一 ID/key，整张 inbox 状态不变 |
| C11 最终追加 | 遗留 draft 删除 metadata.intake_id，仍保留正式 check 1 项 | 真实 Submit/Freeze 返回 APPROVAL_SNAPSHOT_UNRESOLVED/409，12 表零写入；恢复原关联后同一 draft 真实提交成功、1 快照、1 槽 |

12 张比较表：`approval_requests`、`approval_records`、`approval_notifications`、`approval_snapshots`、`approval_case_creation_tracking`、`approval_conflict_associations`、`inbox_items`、`case_intakes`、`case_intake_parties`、`cases`、`case_parties`、`entities`。用整行 JSONB 聚合按内容排序比较，不只比较计数。序列消耗不算业务行变化。初次 HTTP claim 在业务事务外取得并在成功/失败时释放，所以该用例单独允许接案 `updated_at` 刷新，其余列完整比较。

## 运行方式与隔离

```sh
cd /private/tmp/lawoa-intake-collaboration-20261107
env GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache \
  GOPROXY=off \
  LAW_OA_COLLAB_PG_DSN='host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_intake_mvp_test sslmode=disable' \
  go test -race -count=1 -timeout=180s -json ./internal/services ./internal/handlers \
  -run '^(TestApprovalCollaborationPostgres.*|TestIntakeDuplicateApprovalGuard|TestIntakeDelayedInterleaveDuplicateApproval|TestIntakeConcurrentDistinctIntakesAllSucceed|TestIntakeResidualDraftRetryAfterFailedSubmit)$'

env GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off \
  go vet ./internal/services ./internal/handlers
```

实际 Go 1.25.6 darwin/amd64，PostgreSQL 16.10。本地 socket/port/db/user 均在 fixture 内校验；每个 PG 用例独立 `approval_collab_<UUID>` schema，独立连接池，清理只 DROP 自己创建的 schema。使用虚构账号、客户和主体，正式 check/review 是有效持久 fixture 输入，没有调用真实检索或外部服务。实际审批创建、冻结、占槽、决定、补正、待办、成案与 replay 均走现有业务实现。事务 adapter 仅按生产 adapter 转发与绑定，不替换业务逻辑。

PG schema 用实际 GORM 模型和最小 intake/party/material 表，执行真实 000081/000082/000084 迁移，不是全库迁移验收。没有连接 MySQL、Redis 或外部 API。未设置 PG env 时测试明确 SKIP/NOT_RUN；最终证据使用上述 env 且 SKIP 为 0。

修复前 `20-final-source-before.json`、`23-final-source-after.json` 覆盖 127 个相关源码/迁移文件，原运行前后无哈希变化，逐项红灯结果见 `24-results-summary.json`。修复后同源复测扩大到 447 个 Go 源码、模块和迁移文件，编号修复阶段见 `44-retest-results-summary.json`，最终闭环版见 `52-final-transition-results-summary.json`。

## 保留的红灯与边界

- `01-compile-probe.log` 缺 PG env 仅编译探针，跳过不计 PASS；`02` 缺 approval_records、`03` 缺 case_materials 是本人测试 setup 缺口，已补齐。
- `07-controlled-lock-overlap.log` 是按旧锁顺序设计的观测探针，主代理统一锁顺序后该旧探针超时，不作为业务死锁证据。`08` 起直接观察 PG 真锁等待。
- `09` 捕获真实审批编号碰撞，同时暴露本人测试在提前退出时尚未等待协程就移除 callback 的清理竞态。已改为释放并等待所有 worker 后清理；`11` 及后续 `-race` 通过，编号碰撞在修复前 `21` 再现，修复后 `41` 全通过。
- `10-handler-fixtures.log` 的 `approval_submitting` 与 Freeze `conflict_ready` 冲突经主代理确认并修复；`16`、`17` 及最终 `21` 的真实 HTTP 场景通过。`12` 首次 HTTP fixture 曾遗漏 department_name，400 正确拒绝；已补齐 fixture，不能把该 400 当 claim 回归证据。
- `13-go-vet.log` 捕获共享文件正在落盘时的 `ensureInitialIntakeBinding` / `IntakeClaimToken` 临时缺符号，没有更改别人文件绕过。后续 `18`、`22` go vet 通过。
- `19-approved-snapshot-check-mutation.log` 单项回归和最终 `21` 均证明 C14 批准快照主体保持；日志中的 macOS linker LC_DYSYMTAB warning 未阻断测试。
- 编号修复后的首次复测 `26`/`27` 被共享测试 `approval_core_review_test.go` 的未使用 `net/http` 导入阻断；`35`/`36` 又被该文件字符串转 `models.JSON` 的临时编译错误阻断。均由文件所有者修复，没有在本任务内更改或排除该文件。`30` 的 29 项已全通过，但取证窗口内其他 review 测试变化，故没有将它作为最终全源码同源证据。编译恢复后 `41`/`42` 全通过且前后哈希一致，完成此次同源复测。
- 此专项不涵盖 C11 通用 Update 的全部负向、全部角色矩阵、历史迁移全量数据、真实浏览器闭环或生产可用性。没有把本次测试结果当全 MVP 验收。
