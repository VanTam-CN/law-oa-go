# B1/B3 后端修复交接（含 R1 读取及补充，待 Inbox R1 汇合）

候选目录 `/private/tmp/lawoa-intake-collaboration-20261107`；HEAD 仍为 `a806831a7cef839c2b971cd4ab260cce77f42b8b`，实际交付包括未提交和被 ignore 的测试。没有 commit/push、安装依赖、真实 API、生产数据操作或子代理调用。

**本切片通过：262 个叶子用例，0 fail、0 skip；其中实际 PostgreSQL 93 项。race / vet / go build ./... / git diff --check 均退出 0。** 31/36 的 593 个源码、模块及迁移文件哈希一致。此为作者本地证据，不替代独立审查、浏览器两链或 MVP 整体验收。

**尚待汇合：Euclid 的 Inbox R1（换客户尚未重提期间的 supplement List/Stats/GET/home 可见性）。** 本切片未改 inbox 文件。R3 已告知主线程不作为最终验收候选；R1 完整修复后需新构建并重跑两条浏览器主链。

## 最终契约

- B1：接案来源的旧 `conflict_approval` 保留历史，但不能缺 active 冻结快照/本审批唯一槽而批准成案。普通决定、集成决定、AutoCreate、retry 的最终事务边界解析 metadata、check、tracking、历史快照及 association 的持久接案关联，锁同一 intake 仲裁；拒绝时不写决定、案件、主体或待办。备用冲突审批新建入口对 intake 返回 409 `INTAKE_CONFLICT_REVIEW_REQUIRED`，指引先完成指定独立复核再从原接案提交。没有将未复核材料自动冻结。
- 已生成的 d08a805d 类旧单由消费端直接覆盖，不依赖重新创建。未操作实际浏览器旧单。已有正式案件只校验/补齐原案件，旧标志失配也不重新 claim；旧检测 C1 已成案后，即使改为 C2，也不能另一审批再次成案。真正非 intake 的 legacy 仍走原兼容流程。
- B3：快照读取先核验审批参与资格与当前上下文，再核验所选 frozen 客户/案件及持久检测上下文。旧客户撤权后显式历史读取 403，历史清单与 previous_snapshot_id 不泄漏不可见版本；新客户可读不能替代旧客户授权。读取不会重写审批或快照。
- R1：仅 active-pointer + needs_revision 的受控审批，分别授权当前可编辑 intake 与旧 frozen 客户/案件，不要求跨轮客户相等。详情、snapshot、重提共用 `AuthorizeApprovalRevisionContext`；事务内保存补充回复也核验两侧。其余状态、ordinary legacy 和通常 check/intake 一致性不放宽。

## 行为证据

- `01-first-red.log`：真实 PG 旧 A + 正式 B 都批准成功，cases=2；申请人/审批人撤权后 snapshot 仍 200。保留原红灯。
- `04-expanded-regressions.log`：区分新增 fixture JSON/HTTP 必填失败与真实“换客户重提被旧检测客户比较误拒”问题；未将未完成 Freeze 计为历史测试通过。
- `08-pg-combined.jsonl`：当时 59 项全部 PASS、0 skip，包含 B1 先后两序列及三次并发、已有旧案/过期标志恢复、C1→C2 重复成案拒绝、普通/集成/retry 真 HTTP 409、新旧客户撤权与白名单恢复。
- `12-final-race.jsonl` / `17-final-results.json` 保留扩大检查首红：13 fail、19 skip。客户 PG 数据库前缀配置错误、缺快照 PG 环境、旧成案 fixture 缺 intake/active snapshot、F04B 测试替身计数器的数据竞争分别修复；没有删断言、改业务绕过或把 skip 记 PASS。
- `18-fixture-protocol-recheck.jsonl` 保留增强 fixture 时 AutoMigrate 引入用户 NOT NULL 字段的 setup 错误；改为沿原 minimal-schema 风格补 cases/tracking，不扩张 users。`21-fixture-env-recheck.jsonl` 对此前失败/跳过组 49 叶子全部 PASS、0 skip。
- `23-final-race.jsonl`：R1 发现前 260 叶子 PASS；独立审查随后发现未覆盖中间状态，不把该次通过当全链完成。
- `29-pending-client-first-red.log`：改客户尚未重提，真实 HTTP 详情返回 403；申请人与审批人均复现。
- `30-pending-client-recheck.log`：两角色真实 HTTP 详情/default 及 explicit snapshot 可见、补充条目正确；旧/新客户分别撤权，读取以及申请人的 PUT 补充/POST 重提均 403 且 12 表无变化；恢复后实际保存补充并重提得到新指针，读取恢复。
- 最终本切片验证：`32-final-race.jsonl` / `.log`、`33-final-vet.log`、`34-final-build.log`、`35-final-diff-check.log`、`37-final-results.json`。三套 PG 环境显式设置，无 skip。`38-scope-and-evidence-audit.json` 包含实际 PG 93 项分类、首红文件哈希和本切片源码清单。

真实 PG 分类：B1/B3/R1 27 项、原事务及 HTTP 30 项、B2 Inbox/首页 3 项、快照专项 23 项、客户专项 10 项。37 的 `pg_leaves` 使用较窄的 Postgres 命名匹配，38 另纳入 `TestIntakeSubmissionPG`，没有改写原终态日志。

## 本切片实际改动路径

生产代码：

- `internal/services/approval_intake_case_guard.go`
- `internal/services/approval_intake_identity.go`
- `internal/services/approval_snapshot_authorization.go`
- `internal/services/approval_workflow_transactions.go`
- `internal/services/approval_conflict_integration_service.go`
- `internal/services/intake_submission_snapshot.go`
- `internal/services/intake_supplement.go`
- `internal/handlers/demo_aggregate_handler.go`
- `internal/handlers/approval_handler.go`
- `internal/handlers/approval_subject_authorization.go`
- `internal/handlers/integration_handler.go`

测试（原断言、并发顺序及数量要求保留；旧测试仅增强 fixture/修正测试替身计数器）：

- `internal/services/approval_backend_final_review_postgres_test.go`
- `internal/services/approval_backend_final_review_http_test.go`
- `internal/services/approval_resubmit_atomic_test.go`
- `internal/services/intake_submission_snapshot_test.go`
- `internal/handlers/approval_core_review_test.go`
- `internal/handlers/conflict_approval_workbench_test.go`
- `internal/handlers/intake_f04b_version_test.go`

两个新增 backend_final_review 测试文件以及部分原专项测试受 `*_test.go` ignore 规则匹配；交付时必须显式纳入，不能只看普通 git diff。原 client / transactions 交接路径不在本轮重新修改范围。

## 冻结点及运行环境

与主线程 `browser/candidate-r3.json` 相比，目前生产增量仅 `approval_subject_authorization.go`、`intake_supplement.go`、`approval_snapshot_authorization.go`，已在写入前通知主线程冻结影响；其余生产代码未变。最新 31/36/38 提供内容哈希。

`LAW_OA_COLLAB_PG_DSN` 和 `INTAKE_COLLAB_PG_DSN` 指向原 socket:55437 的 `lawoa_intake_mvp_test`；`CLIENT_IDEMPOTENCY_TEST_DSN` 指向同 socket 的专用 `lawoa_client_idempotency_20261107`。user=mac，每测试随机 schema，结束仅清理自己的 schema。`GOPROXY=off`，缓存为指定临时 GOCACHE。完整复跑命令和退出码见 37；可复跑脚本为 `run_final_checks_v3.py`。

当前停止本切片业务源码写入，等待 Inbox R1 汇合后做最终共同哈希绑定与必要复测。
