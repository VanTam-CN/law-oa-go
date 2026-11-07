# 接案协作 MVP 后端核心独立审查

- 日期：2026-11-07。
- 工作目录：`/private/tmp/lawoa-intake-collaboration-20261107`。
- 基线：`a806831a7cef839c2b971cd4ab260cce77f42b8b`，审查对象包括未提交 diff 和新增文件。
- 依据：适用 AGENTS、`docs/plans/2026-09-22-intake-collaboration-mvp-glm.md`。
- 下列为修复前的独立只读审查记录，行号对应当时源码。后续修复不覆盖原发现，也不将作者自测改称独立复审通过。

## 原始发现及复核建议

1. **P1：首次 HTTP 提交状态与冻结门禁冲突。** `internal/handlers/integration_handler.go:340` 先将 intake 设为 `approval_submitting`，服务返回后才恢复；`internal/services/intake_submission_snapshot.go:384` 却只接受 `conflict_ready`，正常首次提交被拒且 handler 映射为 500。复核应经过真实 HTTP claim 路径，确认首次提交成功、失败恢复、token 所有权和并发占槽；直接调用服务的成功不能代替该检查。
2. **P1：通用 submit 冻结接案却不取得唯一槽。** `internal/handlers/approval_handler.go:337` 仅拦截 type/workflow，普通类型仍能携带 `metadata.intake_id`；`internal/services/approval_workflow_transactions.go:44` 冻结、提交和交接待办时没有写 intake tracking。复核普通 create/update/submit 携带关联的直接 API 请求，以及两审批争同一 intake；失败方不能产生有效审批、快照或待办。
3. **P1：历史接案可删除 metadata 关联后进入旧重提。** 无 active pointer 的 needs_revision 审批可通过 `internal/services/approval_service.go:729` 后的普通更新替换 metadata，删除 intake_id、保留关联 check；`internal/handlers/approval_handler.go:582` 随后走旧重提，绕过版本、退回轮次、回复和新快照。复核应以检测记录、历史快照、tracking 等真实关联识别 intake，保护关联不能被普通更新清除，同时保留真正非接案业务的旧协议。
4. **P1：成功响应未使用详情页脱敏投影。** `internal/handlers/approval_handler.go:610` 直接返回新重提结果，其中 Metadata 已投影完整快照，包括身份摘要和冲突证据；集成首次成功结果也直接返回 ApprovalSnapshot。复核普通律师及独立核查角色的创建、决定、重提、重放、详情和历史响应；确认受保护标识不出现在任何响应，普通律师看不到受限历史事项，持久快照不被修改。
5. **P2：幂等重放返回旧 metadata。** `internal/services/approval_workflow_transactions.go:102` 重放直接返回原始 approval，首次成功则通过 GetApproval 投影 active snapshot。第二轮变更事实后可出现新指针与旧内嵌材料混合。复核同请求重复提交的权威事实和版本一致，且第二次无业务行增量。
6. **P2：旧历史快照按 ID 不可读。** `internal/handlers/demo_aggregate_handler.go:3238` 将查询 snapshot_id 当作 active pointer，而 `internal/services/intake_submission_snapshot.go:86` 只允许 intake_submission；历史列表中的 conflict_approval 因此返回 409。复核新旧历史版本精确读取、跨审批 ID 拒绝、当前指针不变、缺快照返回 404；旧内容不得改写或伪造成新协议快照。

## 证据边界

- 六项属于源码调用链确认，未声称运行 PostgreSQL 或新增 HTTP 探针复现。
- 原始独立执行：7 个顶层本地测试通过，范围为 `TestApprovalSubmissionResolverPointerBeatsCreatedTime`、`TestApprovalSubmissionResolverLegacyAgreement`、`TestApprovalSubmissionResolverNonIntakeLegacy`、`TestApprovalSubmissionPreservesBusinessRole`、`TestIntakeSupplementNormalization`、`TestIntakeSupplementRejectUncontrolledKeys`、`TestIntakeSupplementLegacyReadOnlyAndFingerprint`。使用纯逻辑及测试独占临时 SQLite、`GOPROXY=off`、`-mod=readonly` 和独立 Go 缓存。
- 原始审查未改源码或测试、未派子代理、未调用外部环境；未执行 PostgreSQL 并发、迁移、浏览器或生产验收。作者 baseline/snapshot/transactions 报告仅为线索，不是独立结论。
- 已观察到 intake 锁前移、supplement_return_record_id 拒绝、bootstrap 新触发器和普通审批兼容分支，不重复作为未实现项。
- 待验证风险：成案及完成重放的 party 仍从检测记录重建；检测记录后续修复是否导致偏离批准快照未做探针，不计入六项确认缺陷。

## 后续修复交接

用户在独立审查结束后授权原审查者参与限定 handler 修复；因此本节只记录实现和作者验证，修复后的独立复审须由其他审查者完成。授权范围：approval_handler.go、approval_conflict_redaction.go、demo_aggregate_handler.go、新 handlers 测试、新 services/approval_submission_history.go、本文。其余源码由主线程负责。

### 2026-11-07 15:40（Asia/Shanghai）限定修复结果

- **#4 成功响应投影**：`approval_conflict_redaction.go:33` 新增通用 `projectApprovalResponseForViewer(value, role)`，供主线程在 `integration_handler.go` 的成功响应复用。审批列表、详情、创建、模板创建、提交、两决定分支、重提和更新均接入。先对所有角色删除受保护身份字段、identifiers、digest/ciphertext，并递归处理 JSON 字符串和 `hmac-sha256:` 值；再对非核查角色裁剪历史冲突证据。保持 metadata 的字符串契约，不修改调用方对象或持久数据。
- **#6 历史读取**：`approval_submission_history.go:15` 的 `ReadApprovalSubmissionSnapshot` 区分当前权威读取与显式历史行读取；显式历史按 `approval_request_id + id` 约束，旧 snapshot_type 和 SNAP_* JSON 标识原样保留。`demo_aggregate_handler.go:3239` 接入并分别返回真实 active ID 与 viewed ID。无快照、未知 ID、跨审批 ID 为 404；坏 active 指针、当前来源冲突、历史 JSON/审批关联损坏为 409。
- **详情补充契约**：`approval_handler.go:256` 在既有对象/冲突上下文授权之后，通过 `ReadApprovalSupplementContext` 返回顶层 `return_record_id/supplement_requests/return_reason`。仅接案关联或 active pointer 的 needs_revision 审批适用；复用 `currentIntakeReturn` 的排序、同时间歧义、版本绑定检查和 `ReadSupplementRequests` 的旧文本规范化，旧记录不写回。
- **入口和映射**：`approval_handler.go:365` 通用/模板 Create 拒绝 intake、check 各别名、快照和成案受控 metadata。`:626` Resubmit 使用主线程的 `ResolveApprovalIntakeID(h.db, subject)` 分流；真实非接案业务保留旧 note-only DTO，接案分支拒绝未知字段/尾随 JSON/缺版本。`:566` 统一已知业务错误映射，普通 Update 的关联门禁为 409、非申请人为 403、不可编辑状态为 400。

### 作者本地验证

执行命令（网络关闭、只读模块依赖、测试独占临时 SQLite、无共享 PostgreSQL）：

```sh
env GOCACHE=/private/tmp/lawoa-intake-independent-review-go-cache GOPROXY=off GOTOOLCHAIN=local \
  go test -mod=readonly ./internal/handlers -run '^(TestCoreReview|TestProjectApproval)' -count=1 -v
```

15:39:43 最终运行：**PASS，14 个顶层测试**（新增 10 个、原脱敏回归 4 个，含子用例），`ok law-oa-go/internal/handlers 1.687s`；限定已跟踪源码 `git diff --check` 通过。新增测试位于 `internal/handlers/approval_core_review_test.go`。

- 验证普通业务真实 HTTP handler 的 create → update → submit → request_changes → resubmit；普通 ImposedRequirements 保留，成功响应脱敏，存储仍保留完整数据；Update 403/400/409 和拒绝写入。
- 通用/模板 Create 共 26 个受控 metadata 反例，拒绝后审批行数为零；持久旧快照关联不能路由到旧 note-only DTO。
- 详情旧文本与结构化退回、两次读取 item ID 稳定、403 先于相关数据处理、同时间退回歧义和缺失/错轮记录 409、历史记录/metadata/快照字节不变。
- active/default/旧历史精确选取、跨审批和无快照 404、当前冲突不因历史读取放宽、无 active 指针改写；两角色的 parties.identifiers 与 search_parameters 保护值均不返回。
- 集成决定分支使用成功结果 stub 验证响应投影与对象不变；这不是集成服务成案事务的真实链路证明。原审查的独立证据与本节作者自测分开。
- 开发中曾因新增测试的未使用 import、未建 approval_records、JSON 类型构造及结构体/Map 序列化顺序断言失败；修正测试构造后重新执行上述完整选择集，未跳过或削弱业务断言。

本次未改 integration_handler、事务/冻结/identity 核心、模型、迁移、前端或既有测试；未派子代理，未运行外部环境、PG 锁/迁移、浏览器或生产验收，未提交/暂存文件。

### 主线程交接和证据限制

1. 集成响应复用：`projectApprovalResponseForViewer(result, c.GetString("role"))`；本线程不修改 integration_handler，主线程须覆盖首次/重复请求/决定/状态等实际成功出口并另行验证。
2. 截至本次读取，`intake_submission_snapshot.go:303` 的 Freeze 仍只从旧 metadata 取 intake_id，并在无 intake_id/active pointer 时提前返回。handler 已按持久关联选择严格 DTO；**服务层还应使用同一关联解析或明确拒绝关联丢失**，否则仅修 handler 不能证明 #3 在内部调用/伪造齐全 DTO 下闭环。该项已交主线程，未越界修改，也未称六项全部关闭。
3. `.gitignore:282` 忽略所有 `*_test.go`，新增测试实际存在且已执行，但不会出现在普通 git status。主线程提交时须明确 `git add -f internal/handlers/approval_core_review_test.go`；本线程未改 .gitignore 或索引。
4. 当前共享树仍在变动。本节只证明以下文件内容及当时依赖下的作者自测，后续独立复审应检查最终 diff 与新增文件。

| 文件 | 最终自测后 SHA-256 |
| --- | --- |
| internal/handlers/approval_handler.go | 5d24f4ee26ef091ec0611075d00b390a0f0058282293cba1139b833f4a8cf2df |
| internal/handlers/approval_conflict_redaction.go | 2cf009de47657e9e8f1ef22862f2d85444b4b58d88018a8c1352dda57d55f53e |
| internal/handlers/demo_aggregate_handler.go | fbc1f57e71cee17107b0234d567b13569ea6d545e44660e87a7cb1728a95930a |
| internal/handlers/approval_core_review_test.go | f9b7326ad901d7cd7a9298648099978a61be40655b0116054afde80b25f779e5 |
| internal/services/approval_submission_history.go | 5d17fb0f54914ced8e702a4797fbcf3f3240057ab67d6764503915c46ecf9204 |
