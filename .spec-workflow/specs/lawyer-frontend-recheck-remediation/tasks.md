# Tasks Document

- [x] 1. 建立复验问题到需求的追踪基线
  - File: `reports/frontend-lawyer-qa-recheck-2026-05-26.md`
  - File: `.spec-workflow/specs/lawyer-frontend-recheck-remediation/requirements.md`
  - File: `.spec-workflow/specs/lawyer-frontend-recheck-remediation/tasks.md`
  - 确认 QA-RC-001 到 QA-RC-006 全部映射到需求和任务。
  - 确认 P1 为 QA-RC-001，其余按 P2/P3 排序。
  - 不改业务代码，只做追踪确认。
  - _Leverage: `reports/frontend-lawyer-qa-recheck-2026-05-26.md`_
  - _Requirements: All_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Lead | Task: Verify that every recheck finding QA-RC-001 through QA-RC-006 is represented in requirements.md and tasks.md, and add missing traceability notes if needed | Restrictions: Do not change application code, do not delete or downgrade QA findings, do not mark issues fixed | Success: The spec has complete traceability from QA-RC IDs to requirements and implementation tasks_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 2. 修复立案冲突检查跳出工作流的问题
  - File: `frontend/src/pages/batch01/Batch01Prototype.tsx`
  - 找到 `CaseIntakeWorkbench` 中运行冲突检查后 `navigate('/conflict')` 的逻辑。
  - 将立案流程内的冲突检查改为留在 `/case/create`。
  - 成功后把返回的 conflict result 写入当前 `runtime.conflict`。
  - 成功后显示可见成功反馈，并保持 active step 为“利益冲突检查”。
  - 确保 `进入团队与费用` 在成功后可点击。
  - _Leverage: `runConflictCheck`, `setRuntime`, `runtime.conflict`, `message`, existing intake step UI_
  - _Requirements: R1, R2_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Senior React Workflow Engineer | Task: Fix the case intake conflict-check action so it runs inside the /case/create workflow, preserves the current draft, stores the returned conflict result in runtime state, and enables the Team & Fees step instead of navigating to /conflict | Restrictions: Do not remove the standalone /conflict workbench route, do not discard form state, do not fake a conflict result, do not change backend authorization | Success: A lawyer can click 运行利益冲突检查 and remain in /case/create with the current draft visible, conflict status completed, and 进入团队与费用 enabled_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 3. 校正立案冲突结果展示为当前草稿上下文
  - File: `frontend/src/pages/batch01/Batch01Prototype.tsx`
  - 在“利益冲突检查”步骤中显示当前草稿的案件名称、客户、对方当事人、相关方。
  - 如果 API 返回的 record 名称与当前草稿不一致，仍显示“本次立案草稿”摘要，避免用户误以为检查了历史案件。
  - 显示 `checkId` / `record.check_id`。
  - _Leverage: `form`, `runtime.intake`, `runtime.conflict`, `conflictRiskText`_
  - _Requirements: R1, R2_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Legal Workflow UX Engineer | Task: Make the conflict-check step clearly show the current case intake draft and returned check ID, preventing historical conflict records from being mistaken for the newly entered draft | Restrictions: Do not hide backend conflict data, do not invent check IDs, preserve existing risk display | Success: After conflict check, the UI shows the current draft title/client/opponent plus returned check ID and risk status in the same /case/create workflow_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 4. 验证并补齐立案后续步骤的可达性
  - File: `frontend/src/pages/batch01/Batch01Prototype.tsx`
  - File: `frontend/src/pages/batch01/Batch01Prototype.less`
  - 确认“团队与费用”“文档与材料”“立案提交”三个步骤都能从冲突成功状态进入。
  - 如果某一步按钮 disabled，明确对应的必填条件并展示说明。
  - 确认底部固定操作栏不会遮挡关键按钮和字段。
  - _Leverage: existing `activeStep`, step buttons, material unavailable modal, submit approval handler_
  - _Requirements: R2_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Frontend Form Flow Developer | Task: Ensure the lawyer can move from conflict result to Team & Fees, Materials, and Final Submit steps without losing draft state, and make disabled states explainable | Restrictions: Do not bypass required validation, do not remove existing material step, do not submit incomplete approvals silently | Success: The full intake stepper is reachable after conflict success and each disabled action has a clear reason_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 5. 修复工作台”查看全部”入口反馈
  - File: `frontend/src/pages/batch01/Batch01Prototype.tsx`
  - 为“我的待办”的 `查看全部待办` 增加导航或 modal/message。
  - 为“利益冲突待复核”的 `查看全部` 增加导航到 `/conflict` 或展开完整列表。
  - 所有按钮必须有可见状态变化。
  - _Leverage: `DashboardCommandCenter`, `navigate`, Ant Design `Modal`/`message`_
  - _Requirements: R3_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Dashboard UX Developer | Task: Wire dashboard view-all actions so each click navigates, expands content, or shows a clear MVP limitation message | Restrictions: Do not remove the buttons, do not route to broken pages, do not leave silent no-op handlers | Success: 查看全部待办 and 利益冲突待复核 查看全部 both produce visible, testable results_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 6. 修复审批列表操作列布局
  - File: `frontend/src/pages/batch01/Batch01Prototype.tsx`
  - File: `frontend/src/pages/batch01/Batch01Prototype.less`
  - 为审批队列表格设置受控宽度、表格内部横向滚动或固定操作列。
  - 长标题列要换行、截断或 tooltip，不得把操作列挤出页面。
  - 1366px 和 1200px 下不得依赖页面级横向滚动访问操作列。
  - _Leverage: `ApprovalWorkbench`, `.batch-table-wrap`, existing table styles_
  - _Requirements: R4_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Responsive Table Layout Engineer | Task: Make the approval list action column visible or clearly reachable at common desktop widths by constraining table layout and preserving row actions | Restrictions: Do not hide approval titles without a readable alternative, do not solve by globally hiding body overflow, do not remove row actions | Success: At 1366px and 1200px, lawyers can see or use an obvious table-internal scroll to reach 进入审批, and the page itself has no problematic horizontal overflow_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 7. 统一审批详情当前审批人与决策权限模型
  - File: `frontend/src/pages/batch01/Batch01Prototype.tsx`
  - 抽取或完善 `normalizeApprovalAccess` / `isCurrentUserApprover` 辅助逻辑。
  - Header、流程当前节点、底部操作区使用同一套 normalized approval access。
  - 优先使用 `available_actions`；没有时保守 fallback。
  - 修复“顶部当前审批人”和“流程当前节点人员”互相矛盾的展示。
  - _Leverage: `isCurrentUserApprover`, `ApprovalDecisionFlow`, current user local storage helpers_
  - _Requirements: R5_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Frontend Authorization Engineer | Task: Normalize approval current-approver and available-action state so the approval detail header, workflow node, and decision buttons use one consistent permission model | Restrictions: Do not grant decision actions when backend data is ambiguous, do not hide actions for a true current approver with allowed actions, do not rely on CSS-only hiding | Success: Non-current approver lawyers see readonly copy and no decision buttons; current approvers see allowed decision buttons; the displayed current approver information is not contradictory_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 8. 为客户主档案快捷操作增加明确反馈
  - File: `frontend/src/pages/batch01/Batch01Prototype.tsx`
  - `新增联系人`: 打开联系人 modal/drawer，或显示明确未开放提示。
  - `上传附件`: 打开上传 modal，或显示明确未开放提示。
  - `导出客户档案`: 触发 JSON/文本导出，或显示明确失败/未开放提示。
  - 每个动作 1 秒内必须有可见反馈。
  - _Leverage: `ClientMasterProfile`, selected `profile`, Ant Design `Modal`/`message`_
  - _Requirements: R6_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Client Profile UX Developer | Task: Make client profile quick actions produce explicit feedback for add contact, upload attachment, and export profile | Restrictions: Do not create fake backend persistence unless an endpoint exists, do not leave buttons as focus-only or scroll-only interactions, preserve existing start-case and start-conflict actions | Success: Each quick action opens a useful UI, triggers a download, or shows a clear MVP unavailable message within one second_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 9. 增加重复按钮的可访问名称上下文
  - File: `frontend/src/pages/batch01/Batch01Prototype.tsx`
  - 审批行按钮使用 `aria-label` 或按钮文本包含审批编号。
  - 冲突筛选按钮使用 `aria-label="筛选高风险"` 等。
  - 客户 Tab 与快捷操作在重复文本场景下提供上下文名称。
  - _Leverage: role/name selectors used in Playwright specs_
  - _Requirements: R7_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Accessibility Engineer | Task: Disambiguate repeated accessible names for approval row actions, conflict risk filters, and client profile buttons so keyboard/screen-reader users and Playwright tests can target them reliably | Restrictions: Do not make visible labels noisy unless needed, prefer aria-label when visible text should remain concise, do not break existing visual layout | Success: Repeated controls have unique contextual accessible names and E2E selectors no longer hit strict-mode ambiguity for these controls_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 10. 新增完整立案工作流 E2E
  - File: `frontend/e2e/case-create-full-workflow.spec.ts`
  - File: `frontend/e2e/utils/test-helpers.ts`
  - 覆盖律师从工作台进入 `/case/create`。
  - 填写必填项，运行冲突检查。
  - 断言 URL 仍是 `/case/create`。
  - 断言冲突状态完成，并可进入“团队与费用”。
  - _Leverage: `seedAuthenticatedUser(page, 'lawyer')`, existing route mocks_
  - _Requirements: R1, R2, R8_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Playwright Workflow QA Engineer | Task: Add an E2E test for the full lawyer case-intake path through conflict check and Team & Fees, using stable role/name selectors and current route/account setup | Restrictions: Do not skip the test, do not mock away the frontend bug being tested, do not assert brittle implementation details | Success: The test fails before the workflow-context fix and passes after the lawyer remains in /case/create and can enter Team & Fees_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 11. 新增 Dashboard、审批、客户快捷操作回归 E2E
  - File: `frontend/e2e/dashboard-actions.spec.ts`
  - File: `frontend/e2e/approval-layout.spec.ts`
  - File: `frontend/e2e/approval-permission-consistency.spec.ts`
  - File: `frontend/e2e/client-profile-actions.spec.ts`
  - 覆盖 QA-RC-002 到 QA-RC-006。
  - 使用当前账号 `lawyer / Demo@2026` 和必要的 admin/current approver fixture。
  - _Leverage: existing `approval.spec.ts`, `clients.spec.ts`, `layout.spec.ts`, `dashboard.spec.ts`_
  - _Requirements: R3, R4, R5, R6, R7, R8_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Playwright Regression Engineer | Task: Add or update E2E coverage for dashboard view-all actions, approval layout, approval permission consistency, client quick-action feedback, and accessible-name disambiguation | Restrictions: Do not weaken existing assertions, do not use fragile CSS selectors when role/name selectors are available, do not skip tests for unresolved bugs | Success: QA-RC-002 through QA-RC-006 have deterministic Playwright coverage using stable selectors and current test accounts_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

- [x] 12. 执行完整验证并回填复验报告
  - File: `reports/frontend-lawyer-qa-recheck-2026-05-26.md`
  - File: optionally create `reports/frontend-lawyer-qa-recheck-remediation-2026-05-26.md`
  - 运行 E2E、前端构建、lint、type-check、Go build。
  - QA-RC-001 到 QA-RC-006 必须标记最终状态。
  - 如果 type-check 仍因历史债务失败，记录失败摘要并确认本轮改动是否新增错误。
  - _Leverage: repository verification commands_
  - _Requirements: R8_
  - _Prompt: Implement the task for spec lawyer-frontend-recheck-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Release QA Engineer | Task: Run final verification for the recheck remediation, update QA-RC statuses, and create a concise remediation report with command results and residual risk | Restrictions: Do not mark an item fixed without E2E or manual browser evidence, do not ignore failing commands, do not delete the original recheck report | Success: Verification results are recorded, QA-RC-001 through QA-RC-006 each have explicit final status, and remaining risks are documented_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion if the log tool is available, then set to `[x]`.

## Execution Order

1. Task 1: establish traceability.
2. Tasks 2-4: fix the P1 case-intake workflow blocker.
3. Tasks 5-9: fix dashboard, approval, client feedback, and accessibility issues.
4. Tasks 10-11: add and stabilize regression E2E.
5. Task 12: run verification and update reports.

## Definition of Done

- QA-RC-001 is fixed: conflict detection no longer ejects lawyers from `/case/create`, and the lawyer can enter `团队与费用` after a successful check.
- QA-RC-002 is fixed: dashboard view-all actions navigate, expand, or show explicit feedback.
- QA-RC-003 is fixed: approval list row actions are reachable at 1366px and 1200px without page-level overflow.
- QA-RC-004 is fixed: approval current approver and decision buttons use one consistent permission model.
- QA-RC-005 is fixed: client quick actions show modal/drawer/download/toast/unavailable feedback.
- QA-RC-006 is fixed: repeated buttons have contextual accessible names and stable role/name E2E selectors.
- Playwright E2E covers all QA-RC findings.
- `npm run build`, `npm run lint`, `npm run test:e2e`, `go build ./...`, and feasible type-check verification have been run and documented.
