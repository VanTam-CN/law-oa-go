# 接案协作 MVP 后端最终差异审查

结论：**不通过，冻结前先处理 B1。** 确认一条可绕过接案唯一成案约束的备用入口，以及重提后待办被过滤的问题；新增历史读取还继承了对象授权缺口。以下为源码调用链确认，不冒充真实 PostgreSQL 两案复现。

- 审查日期：2026-11-07，Asia/Shanghai。
- 目录：`/private/tmp/lawoa-intake-collaboration-20261107`。
- HEAD / 比较基线：`a806831a7cef839c2b971cd4ab260cce77f42b8b`；包括本轮未提交 diff、新增和被 ignore 的测试。
- 范围：transaction / snapshot / legacy identity / response projection、N4 待办及 bootstrap。未重复前端、未全库审计。
- `core-review.md`、`inbox/N4_HANDOFF.md` 仅作为定位线索。结论来自实际源码与本轮独立执行的限定检查。

## B1 — P1 / BLOCKING：接案来源的旧冲突审批仍能绕过快照与唯一槽成案

**主要位置：** `internal/handlers/demo_aggregate_handler.go:2982–3025`；消费端 `internal/services/approval_conflict_integration_service.go:562–576`、`:1385–1397`。

**触发条件：** 同一接案 I 的正式检测 C 允许创建冲突审批（例如有需复核命中的完整检测）。先通过 `POST /conflict/tasks/:task_id/approval` 创建审批 A；随后完成有效的独立复核，再从接案工作台正常创建 `case_creation` 审批 B。申请人与审批人均可为合法角色，不需要伪造他人的对象或修改数据库。

实际调用链：

1. `createConflictApproval` 虽在 `:2891` 附近校验检测与当前接案事实，却直接插入 `type=conflict_approval/status=submitted`。`:2982–2988` 的 metadata 没有 `intake_id`；`:3020–3022` 写旧 `conflict_approval` 快照，不设置 active 指针，也不调用 Freeze 或取得 intake tracking 槽。`:3033` 新增的待办交接不补足这些约束。
2. `approval_handler.go:506–511` 使用 `RequiresCaseCreationApproval` 路由决定；后者明确把 `conflict_approval` 视为可成案类型。具备有效复核及完整主体数据时，批准后会进入 `ensureApprovedApprovalHasCase` → `createCaseTxWorkflow`。
3. `validateCurrentIntakeSubmission` 在 `approval_workflow_transactions.go:230–232` 因 active 指针为空直接放行。成案入口 `approval_conflict_integration_service.go:562` 只从 metadata 取 `intake_id`；A 连整个槽位判断都不会进入。即使补上 metadata，`:570–576` 还允许 nil-pointer 历史单缺槽继续。因此仅修一个字段或仅修新建接口不够。
4. `ClaimCaseCreationOnTx`（`integration_repository.go:350`）锁的是各自 approval ID，只保证同一审批不重复成案。`CaseService.ValidateApprovedCase`、`RequireConflictDispositionForCase` 核对复核、客户和律师，没有接案唯一性；`cases.conflict_check_id` 也不是唯一列。追踪表的 intake 唯一索引允许 NULL，不能约束 A。
5. B 的提交检查 `integration_handler.go:107–119`、`:508–523` 只查 `type=case_creation` 且 metadata 关联 I 的审批，不会发现 A；新 Freeze/占槽可正常建立 B。A 不占 I 的槽，后续 A、B 分别批准时没有共享仲裁点。可形成两个正式案件；至少“先建 A、再提交 B、依次批准 A/B”的顺序不依赖并发竞态。

**证据边界：** 用户提供“浏览器已生成 `d08a805d…`，尚未批准/成案”的事实；本审查未查询该实时对象、未批准、未调用成案接口。上述是源码可达性确认，不能写成已经实际产生两案。此备用入口在基线已有，但本轮统一协议仍未覆盖，直接阻断 C11/C12 的整体结论。

**最小统一建议：**

- 对 intake 来源的检测，备用“创建冲突审批”入口应拒绝再生成可成案 legacy 单，并引导现有独立复核流程；复核完成后由原接案受控提交。不能在“尚待独立复核”阶段强行调用要求复核已完成的 Freeze。若确需此入口兼作接案提交，则只能在前置条件满足时委托同一受控服务，不能继续直接插表。
- 决定、首次成案、失败重试的最终服务边界均用持久检测/历史证据解析真实 intake（可复用 `ResolveApprovalIntakeID`），不能以 metadata 缺失或 active=nil 判断“非接案”。对未成案的 intake 关联审批，缺少有效快照或本审批持有的唯一槽时在写入前返回 409。真正不关联 intake 的旧业务另行保持原兼容语义。
- 该消费端守卫必须覆盖已创建的 `d08a805d…` 类单据；仅关闭新建按钮不够。不要在读取或批准时静默重写旧快照。已有正式案件的幂等读取/校验与“允许再创建一案”应分开。
- 修后真实隔离 PG 回归：备用单先建、正式单先建、备用单先成案，以及并发两批准；断言有效槽与案件总数至多一，失败方无审批决定/案件/主体/交接的部分写入。新增普通 HTTP 决定和 integration/retry 入口的负向检查，不能只测新 `case_creation` 路径。

## B2 — P1：补正换检测后，合法的新轮次待办被授权投影过滤

**位置：** `internal/services/inbox_workflow_action.go:328–336`；关联 `:131–144`、`internal/services/inbox_service.go:558–563`。

**触发：** 首轮审批保存 C1；退回后修改主体事实，完成 C2 的正式检测与独立复核，再成功重提原审批。

Freeze 更新 active 指针及顶层 `ConflictCheckID=C2`，有意保留原始持久 metadata 中的 C1。`ResubmitIntakeApproval` 保存该对象并事务交接待办；详情通过 `ApplyApprovalSubmission` 才在内存投影 C2。N4 却直接批量读取原始 approval，`approvalInboxMetadata` 只解析旧 metadata，`inboxApprovalCheckID` 优先取顶层 C2，随后 `:333–336` 因旧 metadata 的 C1 不相等返回 false。所有该审批的 workflow 项被 `visibleInbox` 排除，连授权后计数也归零；首页共用该服务，同样看不到交接任务。数据库“新待办行已插入”不能证明用户收到可办理项。

**修复边界：** 批量加载并验证 active 指向的权威快照，使用同一轮次上下文进行待办授权和动作投影；保留历史 metadata，不通过删除一致性校验或逐条查询来掩盖问题。加入 C1→C2 实际重提后的 List/Stats/单项 GET/首页回归，覆盖审批人待处理、申请人等待、再次退回补充及成案结果。本轮已有测试主要检查待办行数，或使用顶层与 metadata 一致的 fixture，未证明这条交叉链。

## B3 — P1：新增历史快照入口沿用申请人/审批人无条件放行

**位置：** `internal/handlers/demo_aggregate_handler.go:3228–3238`、`:3275–3278`；新历史服务 `internal/services/approval_submission_history.go:23–51`。

**触发：** 某审批申请人或当前审批人仍保留审批行中的身份，但已因客户/案件隔离墙撤销对象访问权。直接请求 snapshot API（包括新增 `snapshot_id` 历史选择），`canAccessApprovalSnapshot` 在匹配申请人/审批人 ID 时立即返回 true，根本不检查业务对象。历史服务只校验行属于此审批；响应脱敏保留申请事实、主体名称等，独立核查角色还保留完整冲突证据。于是详情 GET 经 `authorizeApprovalConflictContext` 拒绝的对象，仍可通过 snapshot 读取。

**相对基线说明：** ID 直接放行是基线已有缺口，非本轮新写出的判断；本轮新增按 ID 读取任意历史轮次及历史列表，继续复用了该缺口，故记录为扩展入口的继承性权限阻断，不能宣称是新引入的整个权限漏洞。角色脱敏不等于对象授权。修复应在返回内容及历史元信息前执行与详情一致的业务上下文授权；显式历史读取还需避免仅以当前新客户的权限放行已无权查看的旧客户材料。须补充“先可见、后撤权”的真实 handler 负向测试。

## 原六项及本轮补修的复核结果

这些是限定源码复核结果，不是完整端到端 PASS。

| 检查项 | 本轮看到的实际实现及限制 |
| --- | --- |
| HTTP claim 与首次 Freeze | handler 传 `json:"-"` 的 server token；集成事务锁 intake 并按 token 消费 claim，再进入 Freeze；未直接放宽接受任意 `approval_submitting`。本轮未重跑真实 PG claim 交错。 |
| 普通 submit 占槽 | `submitApprovalInTx` 在 Freeze 后调用 `ensureInitialIntakeBinding`；通用/模板 create 拒绝受控 metadata。但 B1 的直接插表入口仍未经过该路径。 |
| 历史关联删除 | Update 使用 `ResolveApprovalIntakeID`；HTTP 重提按持久关联选严格 DTO；Freeze 再核对持久关联与 metadata，不因缺字段直接进入新协议的空操作。 |
| 成功响应脱敏 | 普通审批及 integration 的 `updatedApproval` 成功出口均使用共享投影；嵌套 JSON 字符串、身份摘要及密文被处理。B3 是对象授权问题，不能靠投影关闭。 |
| 重提重放同源 | 命中重放和首次成功均通过 `GetApproval` 投影当前快照；未再见原先直接返回裸持久 metadata 的重放分支。 |
| 旧历史快照读取 | 显式 ID 按所属审批精确取行，旧类型不冒充 active；无快照/跨审批不存在返回 404。授权边界仍见 B3。 |
| 成案主体/结果交接 | 新轮次 parties 取批准快照；首次成案及已成案校验在事务内调用交接，错误回滚。B1 的无指针路径仍走 legacy 分支。 |
| N4 业务完成语义 | workflow 手工完成/删除拒绝、归档与业务状态分离；结果区分 approved_pending_case / failed / incomplete / created；overdue 已接同一 InboxService。B2 影响重提后的可见性。 |
| Bootstrap | 已包含 000082–84 所需列、唯一索引、任务契约约束、client 幂等约束和新快照/决定不可变触发器；与迁移逐项只读对照。未执行 DDL，不宣称迁移或部署验收。 |

## 独立执行与边界

本轮执行：

```sh
env GOCACHE=/private/tmp/lawoa-intake-independent-review-go-cache GOPROXY=off GOTOOLCHAIN=local \
  go test -mod=readonly ./internal/services ./internal/handlers ./internal/repositories \
  -run '^(TestInbox|TestCoreReview|TestProjectApproval|TestApprovalSubmissionResolver|TestApprovalSubmissionPreservesBusinessRole|TestIntakeSupplement)' -count=1
```

结果：services PASS（2.202s），handlers PASS（1.683s）；repositories 编译完成但 `[no tests to run]`，不计行为测试通过。选择集使用现有纯逻辑/临时 SQLite fixture，没有连接共享 PostgreSQL。`git diff --check -- internal` 通过，仅覆盖已跟踪差异。未新增或修改测试、未更改源码、未操作浏览器、未派代理、未提交或暂存。唯一交付写入是本报告。

B1/B2/B3 未新增动态复现；限定现有检查通过不能证明这些交叉路径正确。未独立执行 C12–C14 PG 并发/故障注入、迁移 up/down/up、真实律师试用或生产验收。作者已有报告不升级为本轮独立证据。`production_ready=false`。

## 审查绑定

15:51:15 +08:00 读取的关键源码 SHA-256；随后对 B1 追加只读调用链核查。共享树没有冻结，后续修复需重新核对受影响结论。

| 文件 | SHA-256 |
| --- | --- |
| internal/services/approval_workflow_transactions.go | 007921dcf74b67f5ddb21ac90ceb256d67056b0a31d6a407e856ced0850d538d |
| internal/services/intake_submission_snapshot.go | de65e4cdf70ce146d3ca0d0b37bad2be6456d773ff7a940a1c28e52b67b79095 |
| internal/services/approval_intake_identity.go | 1936fbfc17cb71b528e378d35e0bbfca8af2f33e67ce44b3b41e55ebf8c39cc3 |
| internal/services/approval_frozen_parties.go | 0f0838262165eb38e8931b42c3cc4a7b243c97dad9f29e8c37b214640328df7b |
| internal/services/approval_service.go | daedcc7d0883dedaf7b8f8d60cb4d529a639bf85aa51aafe8ff1f95948f7b796 |
| internal/services/approval_conflict_integration_service.go | 0008493049ecb44ad19dd9fb7fe4ba629532ed3bd8899a08288c009443019273 |
| internal/services/inbox_workflow_action.go | 792296f17c0b3ca35a2ba127fdcadf077c68f2704424dd4bfc11c74f47e6990f |
| internal/services/intake_inbox_handoff.go | 68ae6b1fbeee91f0957ead8cc3aaf33b5bc5c8bf30f098a2ec9e4830284d6687 |
| internal/services/inbox_service.go | 2561a0ff9393af189e9d246ddbd84cac9f4990ebdff67989fa362c711985ec83 |
| internal/handlers/approval_handler.go | 5d24f4ee26ef091ec0611075d00b390a0f0058282293cba1139b833f4a8cf2df |
| internal/handlers/approval_conflict_redaction.go | 2cf009de47657e9e8f1ef22862f2d85444b4b58d88018a8c1352dda57d55f53e |
| internal/handlers/integration_handler.go | b4a89f161ce6b89f3fb706f26e697c36e97ec20cb9bada2ecca80d13a8dea2ae |
| internal/handlers/demo_aggregate_handler.go | 161b4dad779602396f8821c891c30ff6660f1ee16670370a0565745bddcf0e64 |
| internal/database/intake_collaboration_schema.go | 1c81ec7ff8292d1dcc132e8d854a1142c4beab8a22d3afe4b4a6e45e0c462ff2 |
