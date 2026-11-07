# N4 / B2 修复交接

2026-11-07；候选 `/private/tmp/lawoa-intake-collaboration-20261107`，HEAD `a806831a7cef839c2b971cd4ab260cce77f42b8b`，含未提交改动。B2 的作者修复与限定回归已完成；原独立审查报告不改判，等待独立复核。不是 MVP 最终 PASS，`production_ready=false`。

## 修复与共享契约

- 修改 [inbox_workflow_action.go](../../../../../../internal/services/inbox_workflow_action.go)：先批量加载审批，再按去重 active 指针批量加载快照；每批最多 400 个 ID。对缺失行、空指针、所属审批、快照类型、JSON 内外 ID、轮次、接案、检测 ID、同轮 `conflict_task_id` 一致性逐项验证。
- 仅在内存中的审批副本投影 frozen metadata 和 `ConflictCheckID`；后续批量对象加载、授权、动作全部使用同一副本。历史 metadata 中的 C1 不再与本轮 C2 混用。无效 active 不降级到旧 metadata，列表/count 排除，单项 GET 按原契约 404；schema/查询失败返回错误，不伪装空列表成功。
- `approvalVisible` 的来源一致性检查保持。无 active 的 legacy 行仍执行原检查。没有逐条调用数据库版 `ResolveApprovalSubmission`，没有 N+1，没有读请求写回历史数据。
- Kuhn 对齐事项：**共享 DTO、导出 service 签名均未变，无需新增字段或改业务调用点**。`WorkflowAction.approval_id` 继续为显式字符串，intake/supplement 同时提供真实 intake 与 approval ID；`ListInboxItems(...WorkflowBucket: "needs_action")` 的 Stats 仍是首页同源计数。审批决定、快照写入、成案守卫、handler/router 未由本切片修改。

本次生产源码只改上述 action 文件，`inbox_service.go` 无需追加改动。测试修改/新增为：

1. [inbox_workflow_contract_test.go](../../../../../../internal/services/inbox_workflow_contract_test.go)：原 active-pointer fixture 补齐实际快照行，保留原断言。
2. [inbox_active_snapshot_test.go](../../../../../../internal/services/inbox_active_snapshot_test.go)：14 类无效/越权 active、缺表失败、legacy 来源不一致、76 个不同审批批量查询。
3. [inbox_workflow_postgres_test.go](../../../../../../internal/services/inbox_workflow_postgres_test.go)：复用隔离 PG fixture，真实退回、C1→C2 重提、再次退回、再次重提、集成批准成案。
4. [inbox_workflow_postgres_http_test.go](../../../../../../internal/services/inbox_workflow_postgres_http_test.go)：真实 Gin Inbox handler 与 CommandCenter 首页 handler，未启动服务器。

## 实际证据

| 证据 | 实际结果与边界 |
|---|---|
| [02 修前真实 PG 失败](02-first-red-real-pg.log) | 实际完成 C1→C2 重提，历史 metadata 保留；合法审批人列表应有 1 项却为 0。B2 行为 first-red。 |
| [09 Inbox 完整限定回归](09-final-inbox-verification.jsonl) / [10 汇总](10-final-results.json) | 25 个顶层测试、33 个子用例通过，services/handlers 均通过，无 skip。包含 2 个真实隔离 PG 顶层测试；其余为临时 SQLite/HTTP 契约等。运行期间其他作者改了成案守卫，故追加下一行复验。 |
| [12 最新共享代码的 PG 复验](12-pg-recheck.jsonl) / [13 源码绑定](13-pg-recheck-result.json) | 两个 B2 顶层测试及两个角色 HTTP 子用例通过。16:15:44 +08:00 完成，所记录依赖在这次复验前后 SHA-256 一致。 |
| [06 vet](06-vet.log) | `go vet ./internal/services ./internal/handlers` exit 0；日志为空。发生在后续共享守卫改动前，不冒充其他作者最终静态检查。 |

具体观察：

- 审批人：重提后 `needs_action=1`，List/Stats/GET 的真实 approval ID 与 `decide_approval` 一致；申请人：`waiting=1`、`view_only`。这些读操作前后比较 fixture 的 12 张表，内容不变。
- 再次退回：申请人 `supplement` 仍可见，`resource_id=intake_id`、`approval_id` 均正确；再次重提及真实集成批准后，一个正式案件、一个结果任务，`workflow_status=case_created`。原本 pending/failed/incomplete/created 区分的回归仍通过；没有把 approved 直接改成已成案。
- HTTP 首页：审批人 `todo_items=1`，动作与 Inbox GET 一致；申请人等待任务不进入首页待处理区，`todo_items=0`，但首页 `inbox_stats.waiting=1`。两个角色首页 stats 均与 `/inbox/stats` 相同。
- 批量：1 个审批与 76 个不同审批都是总计 11 次查询，其中快照查询 1 次。14 类无效绑定/无权对象均不泄漏列表或计数，单项不可读；主动删除 fixture 的快照表时明确返回错误。

运行均为 **author**；使用 PostgreSQL 的两项同时标记 **realPG**。Gin httptest 不是 **browser**，本次没有独立复核。C2 检测与独立复核记录是合法持久化 fixture 输入；退回/重提/决定/成案及 Inbox/home 读取使用真实业务代码。本证据不声称实际调用检测器、复核 UI 或 C16/C17 浏览器主链。

测试使用既有隔离 PostgreSQL 16.10：本地 socket `/private/tmp/lawoa-intake-collaboration-runtime/socket`、端口 55437、数据库 `lawoa_intake_mvp_test`。fixture 严格校验目标，每项创建随机 schema，结束清理自己的 schema；没有启动服务或连接真实业务环境。

```sh
env GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off \
  LAW_OA_COLLAB_PG_DSN='host=/private/tmp/lawoa-intake-collaboration-runtime/socket port=55437 user=mac dbname=lawoa_intake_mvp_test sslmode=disable' \
  go test -p 1 ./internal/services ./internal/handlers -run '^TestInbox' -count=1 -json
# 共享守卫变化后的最小复验
# 同一环境：go test -p 1 ./internal/services -run '^TestInboxB2Postgres' -count=1 -json
```

保留的中间失败：[01](01-first-red-real-pg.log) 是本机 socket 权限限制，不算业务 first-red；[03](03-inbox-verification.jsonl) services 因新增测试误用错误常量包名而编译失败，handlers 6 项通过。修正为现有 repositories 常量后 [04](04-inbox-services.jsonl) services 19 项通过；中间日志没有覆盖或删除。

## 并发作者状态与剩余事项

[07 并发状态](07-concurrent-author-state.json) 记录共享集成服务及首页文件变化，随后又观察到 `approval_intake_case_guard.go` 变化；本切片未编辑、回退或暂存这些文件。已在最新可见代码下补跑受影响的两项 PG 测试。当前可见 B1 成案守卫、B3 快照对象授权与对应测试仍属 Kuhn 范围，本报告不宣布其结果。

没有已知阻断 B2 交接的剩余实现项。还需其他审查者复核本补丁；主线程完成 B1/B3、前端及真实浏览器冻结。原 [contract-map](../../contract-map.md) 和 [acceptance-matrix](../../acceptance-matrix.md) 已先保存当时状态，本轮优先修 B2，未回填最终判定或改 tasks；其中 B2 修复状态以本增量报告为准。

**打包提醒：上述 4 个 Inbox 测试文件被现有 `*_test.go` ignore 规则匹配。** 本切片没有改变 ignore、暂存、commit 或 push；主线程冻结候选时应显式纳入这些实际存在的测试，不能只看普通 `git diff`。源码/测试 SHA-256 见 [13](13-pg-recheck-result.json)。
