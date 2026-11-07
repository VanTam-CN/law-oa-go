# Inbox 最终证据索引（只读交付辅助）

采集时间：2026-11-07T17:09:43+08:00。目录 `/private/tmp/lawoa-intake-collaboration-20261107`；HEAD `a806831a7cef839c2b971cd4ab260cce77f42b8b`。本次只写本索引，未改源码、matrix、tasks、ignore 或 git index；没有运行测试、启动服务、连接数据库/浏览器或派代理。下文“通过”均指已存在日志的终态，不是本次重新执行或最终验收。

## 1. 被 ignore 的新增测试

实际执行 `git ls-files --others --ignored --exclude-standard -- internal/**/*_test.go`、`git check-ignore -v`，并逐个用 `git cat-file -e HEAD:<path>` 核验：**21 份文件均存在于工作树、未跟踪、HEAD 中不存在**，全部命中 `.gitignore:282:*_test.go`。这里“新增”指相对 HEAD 不存在，不推断其他作者具体创建时间。没有暂存它们。

### Inbox 切片：8 份

8 份当前 SHA-256 均与 R1 `10-final-results.json` 一致。`inbox_revision_postgres_test.go` 是 fixture，不能独立计作通过的测试。

| 文件 | 用途 | 当前 SHA-256 |
|---|---|---|
| [internal/handlers/inbox_workflow_security_test.go](../../../../../../internal/handlers/inbox_workflow_security_test.go) | HTTP 防伪、legacy、workflow mutation/snooze | `cdfcafd9a4ac6d9a082fe9e7a372a81880fc49c61f316a02c8efec2f45f02aca` |
| [internal/services/inbox_active_snapshot_test.go](../../../../../../internal/services/inbox_active_snapshot_test.go) | B2 active 绑定、拒绝、批量查询 | `55028f81e94cc3483b0c4f4bfb99d85a6be243351726df6cf2ae7d870d986a0c` |
| [internal/services/inbox_revision_authorization_test.go](../../../../../../internal/services/inbox_revision_authorization_test.go) | R1 状态/来源/冻结引用负向与批量 | `f0a7c2df05385eeae09b0aacb0818debee72df544a2244e98e4d1c2bb39f0cf1` |
| [internal/services/inbox_revision_postgres_http_test.go](../../../../../../internal/services/inbox_revision_postgres_http_test.go) | R1 真实 PUT 换客户、撤权/恢复、重提成案 | `cd1d5ec9b17975db7bbd21dd7463ce0791241c2cf80138487c5fece5cf09c977` |
| [internal/services/inbox_revision_postgres_test.go](../../../../../../internal/services/inbox_revision_postgres_test.go) | R1 PG fixture，无 Test 函数 | `fece673791423fa1da126f1362d62239e7d1c5ecdad005e4081f42f8a9affd15` |
| [internal/services/inbox_workflow_contract_test.go](../../../../../../internal/services/inbox_workflow_contract_test.go) | N4 服务契约、交接、授权、计数、归档/延后 | `cd58e79f54dba21f83f5e50f7fcaeee3b054e5803a9ae1a784e95550c9a0bdab` |
| [internal/services/inbox_workflow_postgres_http_test.go](../../../../../../internal/services/inbox_workflow_postgres_http_test.go) | B2 List/Stats/GET/home，申请人/审批人 | `aea925938d2bf1cfd06002a3bd5d6779dd1c17116df958127438d79122a4d052` |
| [internal/services/inbox_workflow_postgres_test.go](../../../../../../internal/services/inbox_workflow_postgres_test.go) | B2 真实重提服务链及 PG fixture | `04e82f376132e16cf6ffdc6aca1b7fa1bb6ee4f42cf7744f6ecd713d50373f23` |

### 同候选其他新增 ignored 测试：13 份

以下为只读包装清单，不把这些文件全部归入本切片或声称本次逐项执行。C13 的直接依赖为 collaboration 两个文件；bootstrap/模式测试仍由 Kuhn 汇合。

| 文件 | 与本索引关系 |
|---|---|
| [internal/handlers/approval_core_review_test.go](../../../../../../internal/handlers/approval_core_review_test.go) | 共享审批/快照/客户专项；按对应作者交接 |
| [internal/handlers/client_idempotency_test.go](../../../../../../internal/handlers/client_idempotency_test.go) | 共享审批/快照/客户专项；按对应作者交接 |
| [internal/services/approval_backend_final_review_http_test.go](../../../../../../internal/services/approval_backend_final_review_http_test.go) | 共享审批/快照/客户专项；按对应作者交接 |
| [internal/services/approval_backend_final_review_postgres_test.go](../../../../../../internal/services/approval_backend_final_review_postgres_test.go) | 共享审批/快照/客户专项；按对应作者交接 |
| [internal/services/approval_bootstrap_postgres_contract_test.go](../../../../../../internal/services/approval_bootstrap_postgres_contract_test.go) | bootstrap / snapshot mode 联合候选 |
| [internal/services/approval_bootstrap_postgres_fixture_test.go](../../../../../../internal/services/approval_bootstrap_postgres_fixture_test.go) | bootstrap / snapshot mode 联合候选 |
| [internal/services/approval_bootstrap_postgres_http_test.go](../../../../../../internal/services/approval_bootstrap_postgres_http_test.go) | bootstrap / snapshot mode 联合候选 |
| [internal/services/approval_collaboration_postgres_http_test.go](../../../../../../internal/services/approval_collaboration_postgres_http_test.go) | 共享事务/HTTP；含 C13 故障用例 |
| [internal/services/approval_collaboration_postgres_test.go](../../../../../../internal/services/approval_collaboration_postgres_test.go) | 共享事务/HTTP；含 C13 故障用例 |
| [internal/services/approval_snapshot_mode_postgres_http_test.go](../../../../../../internal/services/approval_snapshot_mode_postgres_http_test.go) | bootstrap / snapshot mode 联合候选 |
| [internal/services/client_idempotency_test.go](../../../../../../internal/services/client_idempotency_test.go) | 共享审批/快照/客户专项；按对应作者交接 |
| [internal/services/intake_submission_snapshot_test.go](../../../../../../internal/services/intake_submission_snapshot_test.go) | 共享审批/快照/客户专项；按对应作者交接 |
| [internal/services/intake_supplement_test.go](../../../../../../internal/services/intake_supplement_test.go) | 共享审批/快照/客户专项；按对应作者交接 |

已跟踪的 [inbox_handler_security_test.go](../../../../../../internal/handlers/inbox_handler_security_test.go) 是修改文件，**不属于新增 ignored 清单**，但下面 C06/C07 使用其行为证据。

## 2. 已核对的原始日志

| 代号 | 原始证据 | 核对到的范围 |
|---|---|
| I | [R1限定 race 日志](09-final-race.jsonl)；[结果与源码哈希](10-final-results.json) | 30 个顶层、77 个叶子 pass，0 fail/skip，16:58:10 完成；Inbox 全组 + Kuhn PendingClientChangeHTTP。仅其中 6 个叶子 realPG，其余不能混标。**没有选中 C13 ReturnRollback/ResubmitRollback。** |
| T | [事务原始日志](../../transactions/49-final-transition-race.jsonl)；[34项逐项结果](../../transactions/52-final-transition-results-summary.json) | 15:49:52 完成；含下列 C13 全部 6 项实际 PG 故障回滚/重试。 |
| K | [Kuhn扩展回归](../../backend-final-fixes/32-final-race.jsonl) | 16:38:02 完成；再次含 C13 全部 6 项 pass。 |
| J | [bootstrap联合回归](../../bootstrap/09-final-race.jsonl)；[联合结果](../../bootstrap/14-final-results.json) | 17:04:24 services 等包终态 pass；29 个 Inbox 顶层与 C13 全部 6 项均实际 pass，含 R1 HTTP。此为作者已有联合记录，不是本次执行。 |
| L | [后续候选回归流](../../bootstrap/21-final-race.jsonl) | 本次读取末条时间 2026-11-07T17:07:50.657408+08:00，services终态pass=true。是正在汇合的新记录；是否最终冻结由主线程依据终态、exit code和前后哈希确定，本索引不跟随轮询或代填结论。 |

下文 I/J 为已有 author 运行；标注 realPG 的路径才使用隔离 PostgreSQL。Gin httptest 属于实际 handler + DB 测试，**不是 browser**。对日志的只读核对也不升级为 independent 行为执行。

## 3. C05–C07：条件 → 具体用例

| 条件 | 具体用例名 | 实际断言与证据层级 |
|---|---|
| C05 补充动作与双 ID | [`TestInboxSupplementActionJSONContract`](../../../../../../internal/services/inbox_workflow_contract_test.go#L151) | JSON 中 intake resource_id 与显式 approval_id；伪造关联后 GET 不可见、List/count=0。I/J，author SQLite。 |
| C05 真实重提后列表/计数/GET | [`TestInboxB2PostgresRevisedSubmission`](../../../../../../internal/services/inbox_workflow_postgres_test.go#L45) | C1→C2 真重提，审批人 needs_action、申请人 waiting；再退回 supplement、再重提及成案结果。I/J，author + realPG。 |
| C05 List/Stats/GET/home 同源 | [`TestInboxB2PostgresHTTPListStatsGetAndHome`](../../../../../../internal/services/inbox_workflow_postgres_http_test.go#L17) `/partner`、`/lawyer` | 实际四入口的 ID、action、stats 一致；waiting 不冒充首页待处理。I/J，author + realPG + Gin。 |
| C05 尚未重提的补充入口 | [`TestInboxR1PostgresClientEditBeforeResubmitHTTP`](../../../../../../internal/services/inbox_revision_postgres_http_test.go#L18) | 真 PUT 客户 A→B 后停在 needs_revision；四入口 supplement=1，随后实际重提与批准。I/J，author + realPG + Gin。 |
| C05 核查任务/交接 | [`TestInboxConflictAssignmentCompletionAndRollback`](../../../../../../internal/services/inbox_workflow_contract_test.go#L360)；[`TestInboxConflictReassignmentAndReturnToEarlierActiveAssignment`](../../../../../../internal/services/inbox_workflow_contract_test.go#L417) | 读到 review_conflict、有效指派转换/旧项关闭/返回原责任人新任务；不是三角色浏览器点击证据。I/J，author SQLite。 |
| C06 授权先于分页/计数 | [`TestInboxAuthorizationBeforePaginationAndCount`](../../../../../../internal/services/inbox_workflow_contract_test.go#L185) | 伪造指派行不授予权限；隔离墙排除受限项且统计=0；白名单恢复，低权/技术角色不获可办理动作。I/J，author SQLite。 |
| C06 直接对象访问/登录门 | [`TestInboxItemHandlersRejectCrossUserAccess`](../../../../../../internal/handlers/inbox_handler_security_test.go#L24) `/get,/update,/mark_read,/complete,/snooze,/delete`；[`TestInboxItemHandlersRequireAuthenticatedUser`](../../../../../../internal/handlers/inbox_handler_security_test.go#L94) | 跨用户返回404且原行不变；未登录401。I/J，author SQLite + Gin；日志子测试中的空格规范化为下划线。 |
| C06 防伪来源/模式/指派 | [`TestInboxPublicAPIDeniesForgedWorkflowFields`](../../../../../../internal/handlers/inbox_workflow_security_test.go#L20)；[`TestInboxRejectsForgedServiceFields`](../../../../../../internal/services/inbox_workflow_contract_test.go#L249) | POST/PUT 拒绝受控字段、数值伪来源及冒用 user_id，不产生业务能力。I/J，author HTTP/服务负向。 |
| C06 无效 active 不回退 | [`TestInboxActiveSnapshotBindingFailClosed`](../../../../../../internal/services/inbox_active_snapshot_test.go#L15)；[`TestInboxActiveSnapshotSchemaErrorIsNotEmptySuccess`](../../../../../../internal/services/inbox_active_snapshot_test.go#L84)；[`TestInboxLegacyMetadataConsistencyStillRequired`](../../../../../../internal/services/inbox_active_snapshot_test.go#L94) | 14 类绑定/对象负向、缺表错误、legacy 来源不一致。I/J，author SQLite。 |
| C06 R1 双侧撤权、来源边界 | [`TestInboxR1PostgresClientEditBeforeResubmitHTTP`](../../../../../../internal/services/inbox_revision_postgres_http_test.go#L18)；[`TestInboxR1OnlyActiveRevisionMayChangeClient`](../../../../../../internal/services/inbox_revision_authorization_test.go#L25)；[`TestInboxR1FrozenReferencesRemainAuthorized`](../../../../../../internal/services/inbox_revision_authorization_test.go#L55) | A/B 分别撤权→List/Stats/home=0、GET404，恢复可见；9类状态/绑定、10类冻结引用负向。HTTP一项 realPG，其余 SQLite，均 I/J。 |
| C06 转交旧项与历史文案 | [`TestInboxReassignmentAndCaseFailureProjection`](../../../../../../internal/services/inbox_workflow_contract_test.go#L259)；[`TestInboxLegacyHTTPArchiveContract`](../../../../../../internal/handlers/inbox_workflow_security_test.go#L71) | 转交旧项 view_only，新责任人获得任务；受限历史标题/意见不输出。I/J，author SQLite，非全字段身份脱敏证明。 |
| C07 workflow 不可手工完成 | [`TestInboxWorkflowCannotBeManuallyCompleted`](../../../../../../internal/services/inbox_workflow_contract_test.go#L27)；[`TestInboxWorkflowHTTPMutationAndSnoozeContract`](../../../../../../internal/handlers/inbox_workflow_security_test.go#L95) | complete/archive/edit/delete 拒绝；snooze 后任务仍需办、Inbox due_date 与审批 TimeoutAt 都不变。HTTP fixture 是 contract 类型审批，非真实 PG 接案链。I/J，author SQLite + Gin。 |
| C07 个人提醒可完成、延后计数 | [`TestInboxMoreThan1000AndSnoozeKeepsActionableCount`](../../../../../../internal/services/inbox_workflow_contract_test.go#L225)；[`TestInboxItemHandlersAllowOwnerAccess`](../../../../../../internal/handlers/inbox_handler_security_test.go#L52) | 1107条分页/统计同源；完成一条后 pending=1106、completed=1；真实 HTTP 个人编辑/read/snooze/complete 落盘。I/J，author SQLite。 |
| C07 legacy归档不冒充业务完成 | [`TestInboxLegacyArchiveIsNotBusinessCompletion`](../../../../../../internal/services/inbox_workflow_contract_test.go#L37)；[`TestInboxLegacyHTTPArchiveContract`](../../../../../../internal/handlers/inbox_workflow_security_test.go#L71) | archived_at 有值、IsCompleted 仍 false；归档≠审批/成案完成。I/J，author SQLite/HTTP。 |

防 N+1 的辅助证据：`TestInboxProjectionUsesBatchQueries`、`TestInboxActiveSnapshotsBatchAcrossApprovals`、`TestInboxR1BatchAuthorizationAcrossRevisions`。I/J 实际通过；76 个 active 提交仍为 11 次查询、76 个补正为 12 次；并非 SQL 内授权分页或全规模压测结论。

## 4. C13：六个真实 PG 故障用例

源码：[`TestApprovalCollaborationPostgresReturnRollback`](../../../../../../internal/services/approval_collaboration_postgres_test.go#L290)、[`TestApprovalCollaborationPostgresResubmitRollback`](../../../../../../internal/services/approval_collaboration_postgres_test.go#L321)。T/K/J 均有下列完整叶子 pass 终态，且有 fault hits 日志，不以测试名称存在代替实际命中。

| 完整叶子名称 | 真实故障点与恢复断言 |
|---|---|
| `TestApprovalCollaborationPostgresReturnRollback/ordinary_inbox_update` | 退回入口：故障命中事务内 Inbox update/create，12表原状回滚；清除故障后原决定请求成功，1条退回记录、1条未完成supplement、槽已释放。 |
| `TestApprovalCollaborationPostgresReturnRollback/ordinary_inbox_create` | 退回入口：故障命中事务内 Inbox update/create，12表原状回滚；清除故障后原决定请求成功，1条退回记录、1条未完成supplement、槽已释放。 |
| `TestApprovalCollaborationPostgresReturnRollback/integrated_inbox_update` | 退回入口：故障命中事务内 Inbox update/create，12表原状回滚；清除故障后原决定请求成功，1条退回记录、1条未完成supplement、槽已释放。 |
| `TestApprovalCollaborationPostgresReturnRollback/integrated_inbox_create` | 退回入口：故障命中事务内 Inbox update/create，12表原状回滚；清除故障后原决定请求成功，1条退回记录、1条未完成supplement、槽已释放。 |
| `TestApprovalCollaborationPostgresResubmitRollback/inbox_items_update` | 重提入口：Inbox update/create失败后pointer/snapshot/slot/status/inbox一起回滚；原请求重试得到2快照、1槽、1办理任务、旧supplement关闭；同请求重放无写、不同内容409。 |
| `TestApprovalCollaborationPostgresResubmitRollback/inbox_items_create` | 重提入口：Inbox update/create失败后pointer/snapshot/slot/status/inbox一起回滚；原请求重试得到2快照、1槽、1办理任务、旧supplement关闭；同请求重放无写、不同内容409。 |

`TestApprovalCollaborationPostgresResubmitRollback/approval_snapshots_create` 是相邻 C14 快照故障项，不重复计入上述 6 项。`TestInboxApprovalHandoffIdempotentAndRollback` 是 SQLite 辅助，不替代 C13 realPG；`TestApprovalCollaborationPostgresCaseCompletionInboxRollback/*` 是成案/重放交接故障，也不能替代退回/重提六项。

## 5. 文档声称仍不能由现有索引直接证明的部分

| 声称/条件 | 已有证据边界 | 仍缺的交付证据（不是本次新增实现任务） |
|---|---|
| C05「律师/核查/审批三类待办直达」 | B2/R1 的实际 PG HTTP 主要为律师与审批人；核查交接有 SQLite helper 用例，检测/复核输入是 fixture。 | 最终候选下三类真实账号从工作台/待办点击到正确页面并操作的连贯 browser 记录；不能把 JSON action/ID 正确等同于导航及角色完整主链完成。 |
| C06「不泄漏受限标题/身份/意见」 | 已有标题/意见不出现、404、空列表与零计数；R1 双侧撤权真实 PG 已覆盖。 | 全角色、全部受限身份字段及列表/详情/首页响应的最终候选负向取证，特别是核查角色权限撤销；当前用例没有逐字段覆盖完整身份载荷。 |
| C07「业务任务不消失、延后不改期限」 | generic workflow 的 HTTP SQLite 与个人提醒服务/HTTP已直接断言期限、完成标记和计数。 | 若交付写成“真实PG接案/三角色浏览器均验证”，仍需相应证据；现有明确断言可支撑有限的后端契约，不支撑该扩展表述。 |
| C13「同一候选整体闭环」 | 六项服务级 realPG 故障/恢复确有 T/K/J pass；R1 77叶子本身没有它们。 | 最终共同源码哈希/退出码/完整选择集的冻结绑定；不得用R1的读取回归替代C13，也不得把服务故障探针表述为HTTP或浏览器异常恢复。 |
| R1「完成」是否等于独立验收 | 作者修复报告和 I/J 行为证据均已存在；[现存独立报告](../../review/backend-final-recheck.md)仍记录16:36发现R1的时点。 | 本次只读发现的报告中，没有修复后 R1 独立签字/最终冻结结论；旧独立 finding 不是当前实现未修的证明，也不能被作者报告自动清零。 |
| 当前文件是否都被旧日志绑定 | Inbox八测试及Inbox生产文件与R1哈希相符；共享 approval_handler 与 bootstrap fixture/HTTP 测试在R1之后变化，另有新snapshot_mode文件。J为17:04已完成联合回归，L是后续候选。 | 以主线程/Kuhn最终来源哈希与对应终态为准。不要称R1旧源码快照覆盖所有后续共享修复，不把缺最终签字误报为这些专项从未运行。 |
| ignored测试是否已交付 | 21份均未跟踪且不在HEAD；fixture并非独立测试。 | 冻结/打包时显式纳入这些文件及配套实现。普通git diff不会包含其内容；本次未暂存或提交。 |

本索引没有给 C05/C06/C07/C13 或其他 C01–C18 填最终 PASS，没有改 matrix/tasks。只补证据导航，源码保持冻结；完成后不继续轮询、测试或修复。
