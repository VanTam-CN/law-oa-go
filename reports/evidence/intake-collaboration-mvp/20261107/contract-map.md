# N0 contract-map：接案协作 MVP

## 0. 核对对象与证据边界

- 实际目录：`/private/tmp/lawoa-intake-collaboration-20261107`。
- 基线：`a806831a7cef839c2b971cd4ab260cce77f42b8b`，`git show --no-patch` 可读取，提交说明为 `fix(approval): atomic resubmit revival + intake-slot guard closes two-case window`；本次读取时 HEAD 仍为该提交，本轮实现位于未提交工作树。
- 本文以 **B** 表示 `git show a806831:<path>` 的基线对象，以 **W** 表示 2026-11-07 本次读取的共享工作树。B 的行号不等于 W 的行号；定位以文件和函数名为准。W 由主线程继续修复，不能据本文推断之后的代码状态。
- 初版 N0 仅进行源码、测试源码、路由及迁移文本核对，未运行测试/迁移/服务/数据库/浏览器。初版标为 **NOT_RUN** 的项目保留该证据边界；对矩阵中 19 个基线测试/迁移文件及命名符号的 `git show` 校验仅证明静态引用存在。
- 后续按授权完成 Inbox B2/R1 修复，本次仅同步其实际契约与证据，不将初版静态核对升级为整体通过。最新作者证据见 [B2报告](inbox/b2/REPORT.md)、[R1报告](inbox/r1/REPORT.md)；早期 [N4交接](inbox/N4_HANDOFF.md) 不代表最新共享代码、浏览器或最终冻结。第7节其他发现保留初版时点，不在本次范围内重新判定。

## 1. a806831 基线文件与测试差异矩阵

下表 B 路径及测试均实际用 `git show` 读取，非从主树未提交稿或历史报告推断。基线已有能力仍需在本轮集成后回归；不能将本轮新增代码当作基线事实。

| 能力 | B 实现位置及可确认边界 | B 测试位置/断言 | W 增量及尚需验证的差异 |
|---|---|---|---|
| 草稿恢复 | `internal/handlers/demo_aggregate_handler.go:357` `IntakeWorkbench`、`:438` `ListIntakeDrafts`；`internal/router/router.go` 的 `GET /case-intakes`、`GET /case-intakes/:id`。`frontend/src/pages/batch01/Batch01Prototype.tsx:1475` `intakeDraftFromServer`、`:5201` `restoreServerDraft` 按所选 ID 重新读服务端，不仅取最新一条；恢复业务字段、标签、相关方，身份原文不回传 | `internal/handlers/intake_draft_list_test.go:72` `TestListIntakeDraftsScopesAndFields`；`intake_draft_revision_test.go:280` `TestDraftRestoreReturnsRevisionAndProtectedIdentityStatus`、`:312` `TestDraftListCarriesRevisionTokens`；前端 `__tests__/IntakeDraftRestore.test.ts` 的完整字段、角色、受保护身份及切换草稿用例 | W 增加补充模式 URL 定位、离开草稿提示、客户弹窗及回复恢复。基线恢复能力不等于 W 的补充清单/回复已经完整恢复。相关新增前端测试位于 `frontend/src/pages/batch01/__tests__/IntakeCollaborationPage.test.tsx`、`IntakeCollaborationContracts.test.ts`；本次未执行 |
| revision 与事实失效 | B `demo_aggregate_handler.go:2089` `UpdateCaseIntake` 要求 `expected_revision`，先对象/客户授权，再事务内锁行并比较 token；保存轮换 revision，事实修改撤销确认/检测绑定。`:2432` `ConfirmIntakeFacts`、`:2532` `StartIntakeConflictCheck` 另有版本门。`internal/repositories/conflict_repository.go:137` `CommitIntakeConflictResult` 在接案行锁下重核事实，原子写检测结果与关联 | `intake_draft_revision_test.go:163/181/251` 缺版本、过期版本拒绝且不覆盖；`intake_draft_revision_contract_test.go:47/110/173/214/267` 确认失效、旧草稿兼容、授权先于版本提示、metadata 保留/合并；`internal/repositories/conflict_intake_commit_race_test.go:52/110` 单赢家与旧事实拒绝 | W 普通事实保存拒绝受控 metadata；同一 PUT 的窄回复分支只改 replies 与 draft revision，不改变事实版本。`SaveIntakeSupplementResponsesInTx`、`FreezeIntakeSubmissionInTx` 分别核对写版本与事实版本。基线 parties/materials 是受保护的整组重建，不应描述为已经具有任意 child PATCH/changed_fields 协议 |
| 多方身份、角色与成案主体 | B `internal/handlers/intake_party_identity_guard.go` 的 `prepareCaseIntakePartyUpdateRow` 及客户/对方身份校验；`demo_aggregate_handler.go:2369` `prepareCaseIntakePartyRow` 序列化 role/depth metadata；`internal/services/conflict_p0_service.go:99` `normalizeConflictSubjects`；`internal/services/case_party_materialization.go` 和 `internal/repositories/integration_repository.go:463` `linkConflictCasePartiesOnTx` 按真实身份/角色物化，不按同名自动合并 | `intake_party_identity_guard_test.go:73/96/119` 空输入继承保护身份、新输入更新、新主体必须有身份；`intake_client_party_identity_test.go:54/78/98/116/129/147` 客户身份继承、同名/类型伪造拒绝；`intake_party_metadata_serialization_test.go:15` role metadata；前端 `IntakePartyRoleRoundTrip.test.ts`；`case_party_materialization_test.go:38/78/97` 全角色、幂等、未知角色拒绝 | W `validateSubmissionPartyRows` 冻结时核对受保护身份摘要与角色集合；`approval_frozen_parties.go` 的 `approvedCaseParties` 从批准快照物化，避免批准后 live 检测变更改变主体。对应 `TestApprovalSubmissionPreservesBusinessRole`、`TestApprovalCollaborationPostgresApprovedSnapshotSurvivesCheckMutation` 源码已存在，本次未运行 |
| 独立复核 | B `internal/services/conflict_reviewer_service.go` `ValidateConflictReviewer`、`AssignConflictReviewer`：启用账号、真实业务角色、当前有效指派、非申请人/负责人、直接管理回避、回避声明。B `conflict_p0_service.go:419` `ReviewConflict` 在检测行锁下要求 COMPLETED、当前接案事实、COMPLETE coverage、无既有 review，再追加不可变复核记录；不改原检测 JSON | `conflict_reviewer_service_test.go:47/65/92/119/144` 无指派、技术管理员/直接上级、过期指派拒绝，指派幂等审计及冲突审批同步；B 复核证据/coverage 门由 `conflict_detection_service_test.go` 的 `TestConflictReviewRejects...` 用例覆盖 | W 指派成功/幂等分支、正式 commit callback、复核记录写入后接 `ReconcileConflictInboxInTx`。基线本来有复核事务，缺的是与 Inbox 同事务交接。主线程报告已在真实链完成 officer 指派，本文未独立执行该运行；下一步复核需证明 review、旧任务关闭与新结果一起提交/回滚 |
| 唯一 intake→approval binding | B `migrations/000081_intake_approval_unique_binding.up.sql` 明确 `UNIQUE approval_case_creation_tracking(intake_id)`，NULL 历史行不回填；`integration_handler.go` 的 `claimIntakeForApproval`、`reacquireIntakeClaimOn`、`releaseIntakeClaimOn` 使用 claim token；`approval_conflict_integration_service.go` 创建审批时先占 tracking 槽，成案事务再次拒绝其他审批持槽 | `internal/handlers/intake_duplicate_approval_guard_test.go:107`；`intake_delayed_interleave_guard_test.go:20/98/163` 延迟交错、不同接案、失败残留 draft 重试；`internal/services/intake_bind_compensation_test.go:51`；`approval_resubmit_atomic_test.go:139/160/175` 竞争槽拒绝、自有槽/历史无接案兼容 | W 首次集成提交绑定同一事务；通用 Submit 冻结 active 快照后调用 `ensureInitialIntakeBinding`，新协议成案缺槽返回 `INTAKE_APPROVAL_SLOT_REQUIRED`；旧协议行为须单列。`ResolveApprovalIntakeID` 从 metadata/检测/tracking/历史快照交叉核对，防止删除 legacy metadata 关联降级绕行。不能把基线终态 best-effort release 宣称为全生命周期原子 |
| 原子成案 | B `approval_conflict_integration_service.go:426` `createCaseTxWorkflow`：同一 tx 内 claim、案件、主体/角色、tracking completed、审批关联；`integration_repository.go:350` `ClaimCaseCreationOnTx` 使用 `clause.Locking{Strength:"UPDATE"}`，不是无效的 GORM v1 query_option。B 已有 replay 真实案件/主体检查 | `approval_case_tx_test.go:268/290/312/331/358/387/408` happy path、tracking/关联失败回滚、重复/并发同案、缺主体修复、损坏快照拒绝；`approval_process_replay_test.go:35/91/110/136/184` replay、缺案、缺主体、主体不匹配 | W 成案/成功 replay 从 active submission 取冻结事实及 parties，并在关联落盘后、commit 前接审批 Inbox helper。新增 `approval_collaboration_postgres_test.go` 的主体失败回滚/批准快照不受 live check 修改影响等测试。这里的“原子”仅指成案事务内部；审批决定事务先提交，两者不是一个大事务 |
| 原子重提 | B `approval_service.go:543` `ResubmitApproval` 在事务外先读取/检查，再通过 `approvalRepo.Transaction` 将 `terminalBinding.rebindAfterResubmitInTx`、重新指派、resubmitted Save 一起提交；唯一槽冲突回滚复活状态。B 没有 active submission 协议/结构化逐项回复/旧页面决定版本门 | `approval_resubmit_atomic_test.go:99` `TestResubmitAtomic_CompetingSlotRollsBackStatus` 明确断言原 rejected 状态保留、竞争者槽保留；`:124` `TestResubmitAtomic_FreeSlotCommitsBoth`；`approval_terminal_release_test.go:88/117/130` 释放及重绑；不能用这些测试证明 snapshot/inbox 原子性 | W `approval_workflow_transactions.go` 锁 approval 后校验用户与 replay；`FreezeIntakeSubmissionInTx` 锁 intake、核对版本/回复/检测/复核、插入新 snapshot；重绑、重新指派、保存 active 指针/状态、Inbox 交接同一事务。新 `ResubmitIntakeApproval` 只收 needs_revision；普通旧 `ResubmitApproval` 不接受已有 active pointer 的新协议请求。对应 PG rollback、stale decision、resubmit-vs-new-approval、lock-order 测试源码见下文 |

复验基线的最小只读命令示例（不切换、不覆盖工作树）：

```sh
git -C /private/tmp/lawoa-intake-collaboration-20261107 show --no-patch --format='%H %s' a806831
git -C /private/tmp/lawoa-intake-collaboration-20261107 show a806831:internal/services/approval_resubmit_atomic_test.go
git -C /private/tmp/lawoa-intake-collaboration-20261107 show a806831:internal/services/approval_service.go
git -C /private/tmp/lawoa-intake-collaboration-20261107 show a806831:migrations/000081_intake_approval_unique_binding.up.sql
```

## 2. 本轮 DTO、入口与来源 ID

### 2.1 HTTP/DTO 合约

| 入口/对象 | W 契约与来源 | 关键限制 |
|---|---|---|
| `POST /api/v1/clients` | 可选 `Idempotency-Key`；作用域操作者 + create operation + key，`client_creation_requests` 保存受保护 key/request digest 和结果 client ID。`ClientService.CreateClientWithIdempotencyKey`、迁移 000083 | 同 key 同规范化内容重放同结果；不同内容 409。不是全所同名主体合并。授权和 identity 保护仍执行 |
| `GET /case-intakes/:id`、列表 | intake 本身/parties/标签、revision、受保护身份状态；草稿 ID 精确恢复 | 授权后返回；身份原文不从服务端恢复 |
| `PUT /case-intakes/:id` 普通事实 | `expected_revision`；普通草稿白名单；本轮 `rejectControlledIntakeMetadata` 类门禁拒绝 submission、conflict、事实确认、回复等服务端字段（实际辅助函数见 `intake_supplement_handler.go`） | 普通事实写与回复写不能混包；不能通过 metadata 重写权威快照/检测/补充状态 |
| 同一 PUT 的回复分支 | 仅 `approval_id`, `return_record_id`, `expected_revision`, `supplement_responses:[{item_id,response}]`，由 `saveIntakeSupplement` → `SaveIntakeSupplementResponsesInTx` 处理 | 先申请人/对象授权、锁 approval→intake，验证最新退回轮次及草稿版本；旋转 draft revision，不伪造正式事实已重新确认 |
| `ApprovalDecisionRequest` | `decision`, `decision_reason`, `next_approver_id` 等既有项；新增 `expected_submission_snapshot_id`, `expected_action_version`；`imposed_requirements:{schema_version:1,supplement_requests:[...]}` | 新协议 active pointer 非空时必须带对应 snapshot ID 与 action version；action version 是数据库 `UpdatedAt.UTC().Format(RFC3339Nano)`。退回 item_id/source 由服务端生成，不能客户端自定 |
| `SupplementRequest` | `item_id`, `category`, `field_key`, `party_local_id?`, `description`, `required`, `source?`；category 为 field_correction / identity_verification / matter_explanation | 1–20 项、说明最多 1000 字；字段白名单；party_* 必须 stable local ID，不能数组下标或名称。身份类说明归一化，不把证件号码放入普通清单 |
| `POST /approvals/:id/resubmit` | `ResubmitApprovalRequest`：`expected_intake_revision`, `expected_submission_snapshot_id`, `return_record_id`, `supplement_responses`, `revision_note` | 服务端生成快照。按 durable intake association 选新协议，不能删除 metadata.intake_id 降级为旧文本重提；同 return 同规范化请求 replay，同 return 不同内容 409 |
| `GET /approvals/:id` | ActiveSubmissionSnapshotID 可空，ActionVersion 只读派生；记录保存实际 submission_snapshot_id/action_version；needs_revision 返回 supplement_context（return_record_id、清单、原因） | 先对象授权，再读关联记录/快照并按角色投影；原记录/快照不因脱敏被改写 |
| `GET /approvals/:id/snapshot?snapshot_id=` | `approval_id`, `active_submission_snapshot_id`, `viewed_snapshot_id`, `submission_history`, `snapshot`, `immutable` | 无参数读取 active authority；指定 UUID 只读该审批下的历史快照，历史旧类型不被提升为 active。历史可读不表示可作为当前写依据 |
| Inbox 列表/详情/统计 | `InboxItemResponse.workflow_action` 与 `completion_mode/source_ref/workflow_bucket/closed_reason/archived_at`；`ListInboxItemsResponse.Stats` 为授权后全量，Pagination.Total 为请求过滤后的总量 | 禁止客户端自定关联/模式/指派/完成状态；先批量验证 active 快照并投影本轮事实，再授权、筛选、分页。仅有效 active + needs_revision 分别核验当前草稿和冻结对象权限；其他状态仍要求 check/intake 一致。当前会加载本人全量提醒，未实现 SQL 内授权分页 |
| `PUT /inbox/:id/archive` | 个人/legacy 提醒或已关闭 workflow 可归档 | 未办 workflow 不允许手工完成、编辑、归档；删除始终拒绝留痕。snooze 不改业务期限，不把 needs_action 隐藏 |

DTO 定义：[intake_collaboration.go](../../../../internal/models/intake_collaboration.go)、[approval_models.go](../../../../internal/models/approval_models.go)、[inbox_workflow_action.go](../../../../internal/services/inbox_workflow_action.go)、[inbox_service.go](../../../../internal/services/inbox_service.go)。

### 2.2 WorkflowAction 与编号

字段：`resource_type`, `resource_id:string`, `approval_id?:string`, `action_code`, `action_label`, `workflow_status`, `assigned_to:{id:string,name:string}|null`, `blocked_code?`, `blocked_reason?`, `completion_mode`, `action_version?`。DTO 不是授权凭证，进入目标页必须重新读状态并在写入时再次授权/核验版本。

| 动作 | 真正的 resource_id | 其他 ID/授权依据 |
|---|---|---|
| `resume_intake` | 持久化 intake ID 的字符串表示 | 草稿所属人和客户对象权限；这是前端支持的动作，不能因此声称当前每个工作台行都生成该 DTO，需逐入口检查 |
| `review_conflict` | `conflict_check_records.check_id` 原字符串 | 当前有效 assignment 的 reviewer_id（用户数值 ID，在 assignee DTO 中转十进制字符串）；真实角色/回避/事实/coverage 门 |
| `decide_approval` | `approval_requests.id` 原字符串 | 当前 current_approver_id、current_stage、状态与版本；禁止 self approval，不能凭提醒接收人绕过对象授权 |
| `supplement`（resource_type=intake） | 有效 active 快照的 frozen metadata.intake_id，并核验实际 intake 关联、申请人及当前草稿/旧冻结对象双侧权限；legacy 保留严格来源检查 | **显式 `approval_id:string`** 来自对应审批 `a.ID`；与 intake ID 不同。前端 `/case/create?intake_id=...&approval_id=...&mode=revision`，缺失就禁用，不猜 source/title；换客户尚未重提不要求新旧客户相等，也不跳过任一侧撤权检查 |
| `view_only` | 已授权的 approval/check/intake/case 真实 ID | 不承诺可写；当前 blocked 状态和前端只读导航的关系仍有接入缺口，见 §7 |

`source_ref` 保存真实字符串对象 ID，旧 `source_id:uint` 兼容保留；不能将 approval/check 字符串强转 0、截断 UUID 或解析标题。Intake/case 即使数据库为数值 ID，resource_id 仍使用字符串；不要把所有 resource_id 都称为 UUID。`party_local_id` 是稳定表单/主体定位标识，与审批 UUID、intake ID、数组下标不同。

旧记录未明确证明为 manual 的 task 也返回 `legacy_unresolved`，不授予动作；标题/正文和目标引用按 legacy 规则脱敏，归档不计为业务完成。内部 trusted handoff 的旧数字来源同样不自动升级为 workflow。

### 2.3 主要错误码

错误来源和 HTTP 映射均来自 W handler/service，下面列协作链关键码而非全库错误大全。部分 integration handler 使用 `{error,message}`，普通入口使用 common API envelope；不能假设两个入口 JSON 外形完全一致。

| HTTP/层次 | 错误码 | 触发与处理 |
|---|---|---|
| 400 | `CLIENT_CREATE_KEY_INVALID`, `CLIENT_CREATE_REJECTED` | 非法 key/创建请求；不泄漏受限客户是否存在 |
| 409；503 | `CLIENT_CREATE_KEY_CONFLICT`, `CLIENT_CREATE_RESULT_UNAVAILABLE`；`CLIENT_CREATE_UNAVAILABLE` | 请求冲突/旧结果不可用；服务不可用保留原 key 重试 |
| 409 | `INTAKE_DRAFT_REVISION_REQUIRED`, `INTAKE_DRAFT_REVISION_CONFLICT` | 草稿版本缺失/过期；先授权再给版本提示 |
| 403 | `INTAKE_WRITE_ROLE_FORBIDDEN`, `FORBIDDEN` | 角色不允许、非申请人或对象/接案关联不允许；不是前端隐藏按钮即可满足 |
| 400 | `SUPPLEMENT_ITEM_INVALID` | 项数、字段、类别、item_id、文本/URL、回复等受控清单不合法 |
| 409 | `APPROVAL_ROUND_STALE`, `APPROVAL_SNAPSHOT_UNRESOLVED` | 退回/重提轮次过期或快照来源不完整/冲突，不能猜最新稿继续 |
| 409 | `APPROVAL_SUBMISSION_VERSION_CONFLICT`, `APPROVAL_ACTION_VERSION_CONFLICT` | 决定针对旧 snapshot 或旧节点/时间版本 |
| 409 | `INTAKE_REVISION_ROUTE_REQUIRED`, `INTAKE_RESUBMIT_VERSION_REQUIRED`, `INTAKE_RESUBMIT_STATE_CONFLICT` | 新协议不能走无版本通用提交/重提，或不处于 needs_revision |
| 409 | `INTAKE_APPROVAL_ALREADY_IN_PROGRESS`, `INTAKE_APPROVAL_SLOT_REQUIRED` | 唯一槽被另一审批占用，或新协议成案缺少持久 binding |
| 409 | `INTAKE_FACTS_STALE`, `CONFLICT_CHECK_LINK_REQUIRED`, `INTAKE_CONFLICT_LINK_MISMATCH`, `SUBJECT_LAWYER_MISMATCH` | 事实/检测/承办人不一致，需要重核/重检 |
| 409 | `CONFLICT_COVERAGE_INCOMPLETE`, `CONFLICT_REVIEW_REQUIRED`, `CONFLICT_REVIEW_STALE`, `CONFLICT_REVIEW_ASSIGNMENT_STALE` | 冻结提交时检测 coverage、独立复核或指派依据不足 |
| 409 | `INTAKE_CLIENT_IDENTITY_STALE`, `INTAKE_PARTIES_REQUIRED`, `INTAKE_PARTY_IDENTITY_REQUIRED`, `INTAKE_PARTY_ROLE_INVALID` | 主档案身份、当事人身份/角色不能与正式检查对应 |
| reviewer gate，按 handler 映射 | `REVIEWER_ASSIGNMENT_REQUIRED/MISMATCH`, `REVIEWER_ROLE_FORBIDDEN`, `REVIEWER_INACTIVE`, `REVIEWER_CONFLICTED`, `REVIEWER_RECUSAL_REQUIRED`, `STALE_CONFLICT_SNAPSHOT`, `REVIEW_ALREADY_EXISTS`, `COVERAGE_LIMITED` | 职务不等于有效指派；禁止本人/直接管理冲突、旧检测及重复复核。不把所有 reviewer 错误统一标成 409，实际映射由 reviewer/review handler 决定 |
| 400；409 | `INBOX_PROTECTED_FIELD`；`INBOX_WORKFLOW_ACTION_REQUIRED`, `INBOX_ARCHIVE_REQUIRED`, `INBOX_ITEM_DELETE_UNAVAILABLE` | 伪造关联/模式字段；业务任务应去原业务办理；legacy 仅归档；删除不可用 |
| 409（业务 gate） | `CASE_CREATION_INCOMPLETE`, `CASE_CREATION_NOT_APPROVED`, `CASE_CREATION_GATE_UNAVAILABLE` | 真实案件/主体不完整、未 approved 或成案服务不可用；不可伪报 completed |
| **DTO blocked_code，不是 HTTP 错误** | `CASE_CREATION_PENDING`, `CASE_CREATION_FAILED`, `CASE_CREATION_INCOMPLETE`, `TASK_SUPERSEDED`, `LEGACY_SOURCE_UNRESOLVED` 等 | 描述当前动作状态，不等同于整页无读取权限；前端不能无条件将全部 blocked 理解成详情也不可打开 |

## 3. 角色与动作权限边界

权威角色来自数据库用户及当前对象关系；JWT/URL/DTO/旧提醒中的角色或责任人不能代替服务端核验。[authorization_service.go](../../../../internal/services/authorization_service.go) 的业务管理集合是 director / partner / compliance / risk / risk_control / management；独立复核集合在其上增加 conflict_officer；admin/super_admin 属技术管理，不自动取得业务权限。

| 角色/身份 | 可用范围 | 不能推导出的权限 |
|---|---|---|
| 申请律师 lawyer | 自己的接案与可访问客户；事实确认、提交、needs_revision 补充与重提；读取本人可见结果 | 不能复核/审批自己；不能因自己的 reminder 读到别人的业务对象 |
| assistant / intake_assistant | 基线助理草稿白名单及授权范围 | 回复窄入口显式拒绝 assistant；不能取得律师提交/独立复核权 |
| 普通 user | 自己的个人提醒和已有授权功能 | `denyIntakeWriteForPlainUser` 拒绝接案写入；不能自行 claim reviewer 或审批人 |
| conflict_officer | 授权上下文中的检测/独立复核；可按既有规则指派适格独立复核人 | 不是通用业务管理角色；不能据角色自动修改接案事实。真正 review 还需当前有效 assignment、独立性、回避声明、coverage 和事实新鲜度 |
| 业务管理角色 | 按对象权限、隔离墙及具体 workflow 的管理/审批路径；可指定复核人 | 不免除自审批和直接管理回避，不绕过真实 current approver/assignment |
| 当前审批人 | 当前节点/状态且对象仍有权的决定操作，决定版本必须匹配 | 历史审批记录/旧待办只说明过去参与，不能保留转交后的写权限 |
| 技术管理员 admin/super_admin | 技术管理职责 | Inbox workflow 投影不因技术管理角色放行；不构成律师或独立复核任命 |

`ValidateConflictReviewer`、`AssignConflictReviewer`、`authorizeApprovalSubject`、`lockSubmissionIntake`、Inbox 批量授权各有职责，不能把某个 UI 按钮白名单当全链路统一授权。独立复核使用原检测证据和追加 review；隔离墙响应按 viewer 投影，不破坏用于审查的原证据。

## 4. 快照主要读写者映射

权威链：`approval_requests.active_submission_snapshot_id` → 同审批、类型 `intake_submission` 的 `approval_snapshots` 行；校验行 ID、审批 ID、source_version、intake/check 绑定。不得按 created_at 最新替代 active 指针。

| 写者/读者 | 文件、函数 | 数据来源及事务/输出要求 |
|---|---|---|
| 首次提交写者 | `approval_service.go` `submitApprovalInTx` → `FreezeIntakeSubmissionInTx(...,nil)` → `ensureInitialIntakeBinding` | 由服务端锁 intake，验证当前事实/正式检测/唯一独立复核及主体，插入第 1 轮 snapshot，将 pointer 放入 approval 后随状态 Save；集成首次提交通过 transaction binder 包裹创建/占槽/快照/任务 |
| 重提写者 | `approval_workflow_transactions.go` `ResubmitIntakeApproval` → `FreezeIntakeSubmissionInTx(...,req)` | 基于已核验前轮、最新退回与回复生成新 row，记录 previous_snapshot_id、return_record_id、规范化 fingerprint；不改历史 row；绑定/指派/approval Save/Inbox 同 tx |
| 退回/决定写者 | `approval_service.go` `processApprovalDecisionInTx`；`NormalizeSupplementRequirements` | `approval_records` 保存实际 SubmissionSnapshotID、ActionVersion、CurrentStage、原意见及规范化清单；不更新提交快照 |
| 独立回复写者 | `intake_supplement.go` `SaveIntakeSupplementResponsesInTx` | 只写 intake 协作字段和 draft revision；不旋转 active pointer，不将回复写成正式检查结论 |
| 权威读取器 | `intake_submission_snapshot.go` `ResolveApprovalSubmission` | active pointer 优先且严格绑定。无 active 的历史来源须一致：embedded/persisted 多份矛盾则 `APPROVAL_SNAPSHOT_UNRESOLVED`；新轮次行存在而 pointer 缺失也拒绝 |
| 兼容关联读取器 | `approval_intake_identity.go` `ResolveApprovalIntakeID` | metadata、embedded snapshot、检测参数、tracking、历史 snapshots 多源核对 intake ID；防止移除 legacy metadata 关联后降级到普通业务。只负责关联，不等同于选择当前快照 |
| 只读投影器 | `ApplyApprovalSubmission` | 在内存将权威事实/检测信息投影到审批；持久化调用者不得把投影覆盖历史 Metadata。active 新协议走快照 metadata；legacy 仅在可核验条件下兼容 |
| 审批详情/列表及 mutation response | `ApprovalService.GetApproval/GetApprovalByID`、`ApprovalHandler.GetApproval`、`projectApprovalResponseForViewer` | 使用权威投影，先对象授权再读记录，按 viewer 脱敏。普通决定及 integration 决定响应分别走 handler 投影；不可保存脱敏副本回数据库 |
| snapshot/history API | `approval_submission_history.go` `ReadApprovalSubmissionSnapshot`；`demo_aggregate_handler.go` `GetApprovalSnapshot` | 默认 active；显式 snapshot_id 可读同审批历史旧类型，返回 active/viewed IDs 分离；历史 reader 不放松当前写入 authority resolver。记录列表读取不改变快照 |
| 当前退回上下文 | `ReadApprovalSupplementContext` → `currentIntakeReturn` | needs_revision 的最新退回记录必须对应当前轮次，输出清单/原因供补充 UI 使用；不能把任意历史 return 当当前任务 |
| 决定校验 | `processApprovalDecisionInTx` → `ApplyApprovalSubmission` + `validateCurrentIntakeSubmission` | 锁 approval 后验证 snapshot/action version；approve 再锁 intake、检查当前事实/检测仍匹配批准依据；旧页面不可对新轮次决定 |
| integration 状态/自动成案/重试 | `approval_conflict_integration_service.go` 经 `GetApproval/GetApprovalByID`、`ApplyApprovalSubmission`；`createCaseTxWorkflow` 锁 approval 后再取批准依据 | 不能用最新 draft 或客户端 case_data 覆盖 active 新协议。成功 replay 也必须进入真实案件/主体集合校验 |
| 主体物化 | `approval_frozen_parties.go` `approvedCaseParties` → `ResolveApprovalSubmission` | active 新协议按批准时冻结主体物化/比较；旧协议保留兼容校验，不将 live check 修改当作新批准依据 |
| Inbox 读者/写者 | `intake_inbox_handoff.go`；`inbox_workflow_action.go` 的 `inboxSubmissionProjection`、`inboxFrozenRevisionScope`、`revisionVisible` | 写 helper 在当前 tx 重读持久化 approval，pointer 参与轮次/幂等身份。读侧按去重 active ID 批量取快照，验证所属审批、类型、JSON ID/轮次/intake/check 绑定，再把 frozen metadata 与 ConflictCheckID 同轮投影到内存副本；历史 metadata 不写回，无效 active 不回退。有效 active + needs_revision 对当前可编辑 intake 和冻结客户/案件/检测分别授权，允许合法换客户；其他状态及 legacy 保留严格一致性检查。对象、客户、案件、隔离墙均批量加载，逐项判断不发 SQL，不逐条调用数据库版 ResolveApprovalSubmission/AuthorizeApprovalRevisionContext。动作状态、任务代次及成案结果仍读取真实持久状态，不声称全部字段都来自快照 |
| 旧快照写者 | `createIntegratedApprovalInTx` 兼容分支、`demo_aggregate_handler.go` `createConflictApproval` 的 conflict_approval row、旧 metadata.approval_snapshot | 保留旧类型/旧原文；新提交协议不靠重写它们建立权威。非接案冲突审批与 intake_submission 不应混为一种快照 |

迁移约束：000082 为新类型设置 `(approval_request_id,snapshot_type,source_version)` 唯一索引、按 return_record_id 唯一索引及新快照/带 snapshot 决定的 UPDATE/DELETE 触发器；GORM model 也拒绝改/删。000083 客户幂等独立，000084 Inbox 幂等独立。迁移文本存在不等于数据库已具备约束，本次未执行验证。

残留代码边界：`reconcileCompletedCaseCreation` 与 `CaseCreationApprovalHook` 是旧实现；本次 `rg` 未找到前者定义外调用、后者在当前审批决定链的执行调用。实际新成案 replay 走 `settleClaimedCaseCreation` → `verifyExistingCaseCompletion` → `approvedCaseParties`。不能用未接入旧 hook 充当本轮事务/快照证据，也不要未经调用图核对就把旧定义认定为在用漏洞。

## 5. 两个审批决定入口与实际事务顺序

### 5.1 入口合流

| HTTP | handler | 后端分流及响应 |
|---|---|---|
| `POST /api/v1/approvals/:id/decision` | `ApprovalHandler.ProcessApprovalDecision` | 主 router 显式 `SetIntegratedDecisionService`；对象授权后，`RequiresCaseCreationApproval` 为真则走集成服务；其他审批走 `ApprovalService.ProcessApprovalDecision`。只在测试/未注入构造中可能没有集成服务，不能据此判断生产路由 |
| `POST /api/v1/integration/approvals/:id/decision` | `IntegrationHandler.ProcessApprovalWithConflict` | 直接走同名集成 service；service 再检查对象/版本。两 handler 均需 viewer 响应投影；HTTP 错误 envelope 历史上不同，不可假定字节级一致 |

`RequiresCaseCreationApproval` 明确识别 case_creation、conflict_approval，及 legacy case_creation_config / integration_type=both；普通合同/财务审批不能仅因 approved 自动成案。

### 5.2 T_decision：决定与任务交接原子

`approval_workflow_transactions.go:ProcessApprovalDecision` 的顺序：

1. `approvalTransaction` 开 tx，`SELECT approval FOR UPDATE`，创建绑定该 tx 的 ApprovalService。
2. `processApprovalDecisionInTx` 读当前审批，核验 current approver、非 self、状态、snapshot/action version、底层业务对象权限。
3. 读取权威快照；approve 时 `validateCurrentIntakeSubmission` 再锁 intake，核验事实版本、当前检测及复核门。
4. 构建并写 ApprovalRecord（真实 snapshot、node/action version）；request_changes 规范化补充项。
5. 按决定推进下一节点/退回/拒绝/转派并 Save；reject/request_changes 释放本审批 intake binding，释放错误返回事务，不再采用基线 best-effort 语义。
6. `reconcileApprovalTransition` 重读持久化 approval，调用 `ReconcileApprovalInboxInTx`：关闭旧任务，幂等创建新办理/等待/补充/结果任务；错误向外抛出。
7. Commit 后返回。业务记录、approval 状态、binding 释放和 Inbox 属于同一 T_decision；打印型通知不是交接成功证据。

### 5.3 集成决定：T_decision 后再 T_case

1. 集成 service 先读取授权审批、Apply 权威投影、核对 action/snapshot version。
2. approved 且已关联案件的 approve replay：验证真实案件存在，走同案主体核验/补齐；不是再写一条新决定。
3. 首次 approve 的成案审批先做正式复核/coverage/风险/归属及 case data preflight，避免可预见门禁失败先把审批记为通过。
4. 调 adapter → 上述 **T_decision**。若审批未走完多级节点，实际最终状态仍需按业务结果判断；不能把一次 approve 请求本身当成最终 approved。
5. 对成案路径调用 `ensureApprovedApprovalHasCase` → **T_case**：Begin → approval 行锁 + conditional claim → 重读 active 冻结事实 → 核验 slot → 创建 case → 写批准 parties/roles → tracking completed → approval case link → reconcile 同一审批结果任务 → Commit。
6. T_case 失败会回滚 case/parties/claim/link/该事务 Inbox，但**不会回滚已经提交的 T_decision**。因此可以持久存在 approved + 尚未成案；外层记录 failed tracking，失败状态同源问题见 §7.1。
7. 成功后重新读取审批投影返回；完整成功须同时有真实 case 和批准主体集合。case_created 标志不是单独充分条件。

结果提醒 `is_completed=true` 表示“审批任务处理结束”，不是“整个接案已办结”。W DTO 能区分 `approved_pending_case` / `case_creation_failed` / `case_creation_incomplete` / `case_created`；同轮结果 key 不随成案进度/UpdatedAt 改变。补充/办理任务的 generation 使用真实 snapshot（旧请求稳定 fallback）及当前节点/处理人/数据库时间版本，A→B→A 不复活旧关闭任务。

### 5.4 其他真实事务交接点

- 首次集成提交：transaction binder 让审批创建、持久 intake slot、首次 snapshot、指派/状态和 Inbox 在同一外层 tx；通用 Submit 也必须通过 `ensureInitialIntakeBinding`。
- 回复保存：approval 行锁 → intake 行锁 → 最新退回/版本校验 → 回复及 draft revision。与普通事实写分开。
- 重提：approval 行锁 → 对象授权 → replay 检查 → Freeze 内 intake 行锁/事实与复核校验/新 snapshot → rebind unique slot → AssignApprover → Save active/status/submission date → Inbox → Commit。相同规范化 replay 返回已有轮次，不再写快照/任务。
- 正式检测：`CommitIntakeConflictResult` 锁 intake 后原子写正式 record/cases/evidence/绑定；本轮可变参数 callback `ReconcileFormalConflictInbox` 在同 tx 收尾。旧 async 检测未在本轮主入口闭环内，不能拿 direct formal 证据覆盖它。
- 独立指派：`AssignConflictReviewer` 锁 check → 校验角色/独立性 → 写 assignment/audit、同步仍开放的冲突审批 approver → 同 tx Inbox；幂等分支也 reconcile。
- 独立复核：`ReviewConflict` 锁 check → 完成状态/事实/assignment/coverage/唯一 review 校验 → 追加 review → `reconcileConflictTransition` → Commit。复核结果 reminder 与审批结果 reminder 是不同业务组。
- 上述列的是各入口局部锁顺序，**不是已经证明全系统无死锁**。跨入口重叠锁验证见 W `TestApprovalCollaborationPostgresOverlappingLockOrder`，本次仅核对其源码存在。

## 6. 本轮测试映射与未运行项

| 合约 | 当前测试源码 | 本次状态 |
|---|---|---|
| active pointer 优先于时间、legacy 一致性/非 intake 兼容 | `internal/services/intake_submission_snapshot_test.go`：`TestApprovalSubmissionResolverPointerBeatsCreatedTime`、`...LegacyAgreement`、`...NonIntakeLegacy` | NOT_RUN |
| 清单严格字段、旧文字只读、规范化 fingerprint | `internal/services/intake_supplement_test.go`：Normalization、RejectUncontrolledKeys、LegacyReadOnlyAndFingerprint | NOT_RUN |
| PG freeze/replay/immutable、负向门禁、迁移、并发/故障回滚 | `intake_submission_snapshot_test.go` 的 `TestIntakeSubmissionPG...` 系列 | NOT_RUN；名字包含 PG 不代表本次真实执行 |
| 退回/重提回滚、旧决定无写、重提争槽、case-party 失败、批准快照稳定、重叠锁顺序 | `internal/services/approval_collaboration_postgres_test.go`：`TestApprovalCollaborationPostgresReturnRollback`、`...ResubmitRollback`、`...StaleDecisionNoWrites`、`...ResubmitVersusNewApproval`、`...CasePartyRollback`、`...ApprovedSnapshotSurvivesCheckMutation`、`...OverlappingLockOrder` | NOT_RUN |
| 真实 HTTP 初次提交、两入口旧决定 | `internal/services/approval_collaboration_postgres_http_test.go`：`...HTTPInitialSubmission`、`...HTTPStaleDecision` | NOT_RUN |
| 响应脱敏、durable intake 关联、历史快照和当前 authority 区分 | `internal/handlers/approval_core_review_test.go`：`TestCoreReviewProjectionProtectsAllRolesAndCopies`、`...ResubmitUsesDurableIntakeAssociation`、`...SnapshotHistoryAndMissing`、`...IntegratedDecisionResponseProjection`、`...HistoryDoesNotRelaxAuthority` | NOT_RUN |
| Inbox 授权后分页/count、1107 条、篡改字段、归档/延后、事务回滚、重新指派、持久时间精度、成案投影、approval_id JSON | `internal/services/inbox_workflow_contract_test.go`、`internal/handlers/inbox_workflow_security_test.go`、既有 `inbox_handler_security_test.go` | 初版 NOT_RUN；后续作者限定回归见 B2/R1 报告，不外推独立/浏览器/最终验收 |
| B2：C1→C2 真实重提后的 List/Stats/GET/home；active 绑定负向及批量 | `inbox_active_snapshot_test.go`、`inbox_workflow_postgres_test.go`、`inbox_workflow_postgres_http_test.go` | author + realPG：[B2报告](inbox/b2/REPORT.md)；后续 R1 回归也覆盖此组 |
| R1：真实 PUT 换客户但未重提，补充待办及双侧撤权/恢复；严格状态边界、冻结引用、批量 | `inbox_revision_postgres_http_test.go`、`inbox_revision_postgres_test.go`、`inbox_revision_authorization_test.go` | author + realPG：[R1报告](inbox/r1/REPORT.md)；Gin HTTP 不是 browser，独立复核和冻结仍待主线程 |

新文件可能被根 `.gitignore` 的 `*_test.go` 忽略。测试源码存在且本地运行过，不等于已经进入 git 候选；本次不暂存或调整 ignore。

## 7. 关键代码缺口与交接项

### 7.1 关键：成案失败状态存在两个来源，Inbox 未消费真实失败来源

**静态确认的调用差异；未运行故障注入复现。**

- `approval_conflict_integration_service.go:ProcessApprovalWithConflict` 在 `ensureApprovedApprovalHasCase` 失败后调用 `_ = s.updateCaseTracking(...,"failed",...)`。
- `updateCaseTracking` → `IntegrationRepository.UpdateCaseCreationTracking` 仅写 `approval_case_creation_tracking.creation_status`，没有写 `approval_requests.case_creation_status`；失败更新错误也被丢弃。
- `createCaseTxWorkflow` 的 processing claim 在失败时整体回滚，approval 行可能仍保留原来的空/pending 状态。
- integration status 的 `buildCaseCreationStatus` 会读取最新 tracking 并覆盖状态；**Inbox `inbox_workflow_action.go` 却只以 `a.CaseCreationStatus == "failed"` 生成 `case_creation_failed`**，不读取 tracking。`MarkCaseCreationFailed` 在本次检索的运行代码中只有定义/接口，没有调用。
- 因而真实成案失败后，integration 状态可能为 failed，Inbox 仍是 approved_pending_case。此前 N4 `TestInboxReassignmentAndCaseFailureProjection` 直接将 approval.CaseCreationStatus 设为 failed，只证明字段投影，不证明真实失败链的状态同步。

建议主线程统一失败状态来源（受锁的失败状态同步，或共享/批量读取权威成案状态），保证回滚后的失败登记不会覆盖并发成功，也不增加重复结果 reminder；补“决定成功→case 写入故障→Inbox 与 integration 都 failed→重试成功仅一个 case/一个结果任务”的隔离 PG 故障测试。这是状态真实性的关键交付缺口；本次不改源码。

### 7.2 页面入口与文案

本次读取时 `frontend/src/components/workflow/WorkflowActions.tsx` 对任意 blocked_code/blocked_reason 都返回无导航，pending/failed 结果无法直接进入审批详情；`InboxList.tsx` completed 分类仍写“业务已完成”。两者需与“审批结束、成案失败仍待处理”的真实语义对齐。不能通过清空真实 blocked reason 隐藏失败；授权后的只读查看与受阻写动作应分别处理。

### 7.3 已更新及仍待验证边界

- **此前 overdueRows 直接查表的问题已在本次读取中变更**：`demo_aggregate_handler.go:overdueRows` 现已使用 `ListInboxItems(IsCompleted=false,DueBefore=RFC3339)` 并返回 error。这里只确认实现接入，不声称本次跑过相关工作台测试。
- 通用 Submit 现有 `ensureInitialIntakeBinding`，历史关联删除/历史类型读取也有 `ResolveApprovalIntakeID`、`ReadApprovalSubmissionSnapshot` 及对应测试源码；不能继续沿用“尚未实现”的旧结论，也不能未经运行宣布完整修复。
- 只读结果/history/replay response 与当前 authority 已有分离；两个决定入口仍需在同一真实浏览器/PG 样本中核验权限、过期版本、失败与重放一致性。
- 主线程继续真实 browser 及必要修复；本次不占用或改动其源码/测试。本文是 contract-map 与缺口定位，不是 C01–C18、N4、试点或生产验收报告。

### 7.4 多级节点的成案触发条件需单独核验

静态观察：`processApprovalDecisionInTx` 明确支持 `MoveToNextStage` 返回未完成；但集成服务在决定返回后，仅判断请求为 approve 且 requiresCaseCreation 就调用 `ensureApprovedApprovalHasCase`，未检查 `updatedApproval.Status == approved`。若实际 workflow 有后续节点，可能在成功转交后误尝试成案并记录失败，返回“审批已记录但成案失败”。此项是由可达分支推导的条件风险，**未运行多级流程复现**；主线程应核实当前 MVP workflow 节点数，不能将一次节点同意等同最终通过。

## 8. 初版静态核对的工作树定位

采集时间：2026-11-07T15:54:26.833890+08:00。以下哈希保留初版静态核对时点，不是 B2/R1 修复后的冻结值。最新 Inbox 及直接依赖绑定见 [R1运行后哈希](inbox/r1/10-final-results.json)；共享工作树继续变化时，以文件内容和函数定位复核，本文不锁定他人的写入。

基线实施函数另核验了 10 个源文件的 `git show` 内容，均可读取且具备表中定位符号；没有运行代码。

| W 文件（相对实际目录） | SHA-256 |
|---|---|
| `internal/services/approval_workflow_transactions.go` | `007921dcf74b67f5ddb21ac90ceb256d67056b0a31d6a407e856ced0850d538d` |
| `internal/services/approval_service.go` | `daedcc7d0883dedaf7b8f8d60cb4d529a639bf85aa51aafe8ff1f95948f7b796` |
| `internal/services/approval_conflict_integration_service.go` | `0008493049ecb44ad19dd9fb7fe4ba629532ed3bd8899a08288c009443019273` |
| `internal/repositories/integration_repository.go` | `5d9d7530e035b8a7c001c776a9aff5d35377e60e190ac5eb77bff75aa1cb862e` |
| `internal/services/intake_submission_snapshot.go` | `de65e4cdf70ce146d3ca0d0b37bad2be6456d773ff7a940a1c28e52b67b79095` |
| `internal/services/approval_submission_history.go` | `5d17fb0f54914ced8e702a4797fbcf3f3240057ab67d6764503915c46ecf9204` |
| `internal/services/inbox_workflow_action.go` | `792296f17c0b3ca35a2ba127fdcadf077c68f2704424dd4bfc11c74f47e6990f` |
| `internal/handlers/demo_aggregate_handler.go` | `161b4dad779602396f8821c891c30ff6660f1ee16670370a0565745bddcf0e64` |
| `frontend/src/components/workflow/WorkflowActions.tsx` | `bbb0df785be86ba1066a78a97a1d402bb691c292815af4bfa1f855b106cbb26f` |
