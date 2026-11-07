# N4 待办切片交接

交接目录：`/private/tmp/lawoa-intake-collaboration-20261107`。本切片停止修改，主线程可以接手。基线和本切片文件 SHA-256 见 `source-before.json`、`source-after.json`；这两份记录不代表共享工作树其他文件已冻结。

## 当前结论

- N4 后端接口、实现、迁移文件和本地测试已交付，没有发现阻止主线程继续 officer 独立复核 API 的新增 Inbox 阻碍。
- 主线程报告已完成 officer reviewer assignment。本切片没有重新执行该真实 PG/页面链，也不把用户报告归为本切片测试证据。
- 页面与工作台仍有以下接入剩余项，完整 N4 闭环未验收。未启动服务、连接真实环境、调用代理、commit 或 push。

## 写入范围及可接收文件

以下相对路径均相对于交接目录：

| 文件 | 交付内容 |
|---|---|
| `internal/models/v2_2_0_models.go` | 仅 InboxItem：真实字符串来源、模式、幂等任务键/组/用途/版本、关闭原因及归档时间 |
| `internal/repositories/inbox_repository.go` | 个人任务完成/编辑限制、归档、批量读取、避免后台更新覆盖业务完成状态 |
| `internal/services/inbox_service.go` | 授权后过滤/分页/计数、legacy 脱敏及归档、禁止伪造字段、统一响应 |
| `internal/services/inbox_workflow_action.go` | 批量对象授权与 WorkflowAction 投影，含补充动作 approval_id |
| `internal/services/intake_inbox_handoff.go` | 审批/正式独立复核事务交接、幂等生成、旧任务关闭 |
| `internal/handlers/inbox_handler.go` | 严格请求字段、明确 400/409、归档 handler |
| `internal/handlers/inbox_handler_security_test.go` | 保留跨用户安全检查；个人任务 fixture 明确登记 manual |
| `internal/services/inbox_workflow_contract_test.go` | 13 个服务契约测试 |
| `internal/handlers/inbox_workflow_security_test.go` | 3 个 HTTP 安全/完成/归档/延后测试 |
| `migrations/000084_intake_workflow_inbox.up.sql` | Inbox 字段、契约约束及幂等唯一索引；不将旧记录回填成可操作 workflow/manual |
| `migrations/000084_intake_workflow_inbox.down.sql` | 对应撤销文件；本轮未执行 |

注意：新增的两个测试文件被现有 `.gitignore:282` 的 `*_test.go` 忽略。文件存在且已运行；主线程后续整理候选时必须显式纳入，不能只依赖 `git status`。本切片未修改 ignore 或暂存区。

审批全部业务事务、snapshot/slot、检测与复核 hook、router、production schema、demo handler、前端均由其他写入者负责；本切片不回退其修改。

## 稳定接入契约

```go
func ReconcileApprovalInboxInTx(tx *gorm.DB, approval *models.ApprovalRequest) error
func ReconcileConflictInboxInTx(tx *gorm.DB, checkID string) error

func (s *InboxService) ListInboxItems(ctx context.Context, userID uint, req *ListInboxItemsRequest) (*ListInboxItemsResponse, error)
func (s *InboxService) GetInboxStats(ctx context.Context, userID uint) (*InboxStatsResponse, error)
func (s *InboxService) ArchiveInboxItem(ctx context.Context, id, userID uint) error
```

- 两个交接函数必须使用调用方已锁定的同一事务，在业务状态保存后调用。函数不自行提交；错误必须回传，使业务更新、旧任务关闭、新任务创建一起回滚。不在锁内外发通知。
- 审批函数会从当前事务重读已保存对象，使用持久化 UpdatedAt，避免 PG 微秒与调用方纳秒精度不一致或过期对象重新激活任务。ActiveSubmissionSnapshotID 优先参与轮次身份。
- helper 本身假定 schema 齐备，缺表/列返回错误，不静默跳过。旧无 inbox 的 fixture 兼容判断属于主调用方。
- `ListInboxItems` 的 `Stats` 是授权后的全量统计，不受本次 bucket/page 限制；`Pagination.Total` 是授权并按请求条件过滤后的总数。`GetInboxStats` 共用规则。
- 工作台需要处理项使用 `WorkflowBucket:"needs_action"`；逾期条目可传 `IsCompleted:&false`、RFC3339 格式 `DueBefore`。授权在筛选/计数/分页之前，批量加载关联对象，没有逐条查询；当前实现会先读取本人所有提醒，不是数据库层授权分页。
- `intake/supplement` 返回真实 `resource_id`（intake）与显式字符串 `approval_id`。同时核对审批、接案关联和对象权限，关联不匹配时不返回动作，不猜来源。前端按已冻结的 `/case/create?intake_id=...&approval_id=...&mode=revision` 契约接入。

## approved 与成案结果语义

审批任务完成与案件创建成功分开。对需成案的审批：

| 事实 | workflow_status | blocked_code |
|---|---|---|
| approved，尚未创建案件 | approved_pending_case | CASE_CREATION_PENDING |
| approved，成案失败 | case_creation_failed | CASE_CREATION_FAILED |
| 标记已创建，但正式案件或检测关联未验证 | case_creation_incomplete | CASE_CREATION_INCOMPLETE |
| 正式案件存在且关联检测匹配 | case_created | 空 |

上述结果提醒的 `is_completed=true`、`workflow_bucket=completed` 表示审批处理结束，不能解释为已成案；允许归档，归档不修改业务。当前结果动作是 `view_only`，不另建 `needs_action` 成案重试任务，因此只取 needs_action 的首页不会显示该失败结果。重试仍走审批详情的受控业务入口。

同轮结果身份独立于 UpdatedAt 和成案进度，使用轮次、终态、接收人及用途生成。pending → failed → created 保持同一结果任务，重复调用不重置已读、延后、归档状态。成案事务 link 完成后可再次调用审批 helper，不产生第二条相同 approved 结果提醒。

## 主线程下一步复核的观察点

1. 复核提交前：当前适格 reviewer 有一个 `review_conflict` 需要处理项；申请人是等待项。
2. 真实复核记录写入后，在同一事务调用 `ReconcileConflictInboxInTx`：旧 review/waiting 任务关闭；reviewer 的该检测 needs_action 归零；检测发起人收到一条 `reviewed` 结果提醒。
3. 重复回调或复核完成后调整指派，不应重复生成该复核结果。A→B→A 转交或撤销新指派回落到旧 ACTIVE 指派，应产生正确的新任务代次，不复活原来已关闭的记录。
4. 待办写入失败必须让复核记录及旧任务关闭一起回滚。本地注入失败测试已覆盖；真实 PG 同事务证据仍由主线程取得。

## 已保存的测试与证据

`verification-results.json` 记录命令、起止时间、退出码和日志路径，三项均 exit 0：

```sh
GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off go test -json -p 1 ./internal/services ./internal/handlers ./internal/repositories -run '^TestInbox' -count=1
GOCACHE=/private/tmp/lawoa-intake-collaboration-runtime/go-cache GOPROXY=off go vet ./internal/services ./internal/handlers ./internal/repositories ./internal/models
git diff --check -- <本切片文件>
```

- `01-inbox-local-tests.jsonl`：原始 Go 测试事件；`01-inbox-local-tests.log`：通过的顶层测试清单。13 个 service + 3 个新增 handler + 3 个原有 ownership handler = 19 个通过。repository 包编译通过但没有匹配测试，不计为 repository 行为测试通过。
- `02-inbox-vet.log`：相关四个包 vet，exit 0；空输出正常。
- `03-scope-check.log`：已跟踪改动差异检查，exit 0；此命令不覆盖未跟踪文件。新增 Go 文件已 gofmt，且参与编译/测试。
- `source-before.json`、`source-after.json`：本切片 11 个文件的 SHA-256；本次检查期间没有变化。

覆盖：禁止手工完成 workflow；legacy 归档不完成业务；公共字段伪造拒绝；未授权对象在分页计数之前排除；1107 条数据计数不截断；延后不隐藏办理项；审批和复核幂等/故障回滚；转交及回转；批量查询；持久化时间重读；成案待处理/失败/成功不重复通知；补充动作列表和详情 JSON 明确字符串 approval_id、错误关联拒绝。

这些是本地 SQLite 作者测试。未执行真实 PG 迁移 up/down/up、真实环境、浏览器协作链或独立验收，不据此宣称 C13 或完整 MVP 通过。此前 build/其他回归结果未重新生成到此目录，不作为本次冻结证据。

## 剩余项与责任边界

| 剩余项 | 当前事实与影响 | 接手方 |
|---|---|---|
| 结果提醒文案 | `frontend/src/pages/inbox/InboxList.tsx` 的 completed 分类仍叫“业务已完成”，应避免将审批结束理解为已成案，并展示成案子状态 | 主线程/前端 |
| 结果详情入口 | `frontend/src/components/workflow/WorkflowActions.tsx` 遇任意 blocked_code/reason 就禁用跳转；pending/failed 虽显示原因，却无法从结果动作进入审批详情。允许经服务端授权的只读查看时，仍须禁止真正受阻的写动作 | 主线程/前端 |
| 工作台逾期授权同源 | `internal/handlers/demo_aggregate_handler.go` 的 overdueRows 仍直接查 inbox；需改用统一 ListInboxItems，避免绕过当前对象授权/legacy 脱敏 | 主线程 |
| 真实复核、重试、replay 链 | 主线程正在修复并验证共享业务；本切片没有替代其 PG/页面证据 | 主线程 |
| 候选整理 | 显式纳入两个被 ignore 的新增测试和 000084 迁移；主线程核对 production schema/router 对齐 | 主线程 |

这些页面/工作台问题影响完整交付，不要求暂停当前真实复核 API 验证。本切片到此交接，不继续扩展修改共享文件。
