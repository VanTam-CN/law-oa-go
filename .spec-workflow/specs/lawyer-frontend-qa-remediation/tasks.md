# Tasks Document

- [x] 1. 建立修复基线和 QA 映射
  - File: `.spec-workflow/specs/lawyer-frontend-qa-remediation/requirements.md`
  - File: `reports/frontend-lawyer-qa-2026-05-25.md`
  - 确认 12 个 QA 问题全部映射到需求编号和任务编号。
  - 确认 P1 阻断问题为 QA-003、QA-004、QA-005、QA-008。
  - 目的: 让后续执行模型不会遗漏 QA 记录中的问题。
  - _Leverage: `reports/frontend-lawyer-qa-2026-05-25.md`, `.spec-workflow/specs/lawyer-frontend-qa-remediation/requirements.md`_
  - _Requirements: All_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Lead | Task: Read the QA report and requirements document, verify every QA item QA-001 through QA-012 is represented in the spec, and add any missing traceability notes before implementation starts | Restrictions: Do not change application code in this task, do not remove QA items, do not downgrade priorities without product approval | Success: The spec has complete traceability from QA IDs to requirements and implementation tasks, and blockers are clearly identified_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 2. 修复工作台“新建立案”入口路由
  - File: `frontend/src/pages/dashboard/Dashboard.tsx`
  - File: `frontend/src/pages/dashboard/__tests__/DashboardQuickActions.test.ts`
  - File: `frontend/e2e/cases.spec.ts`
  - 检查所有可见的“新建立案”入口，确保点击后进入 `/case/create`。
  - 如果发现某个入口产品语义是“案件列表”，则把按钮文案改为“案件管理”或“进入案件管理”。
  - 更新或新增 E2E，断言律师从工作台点击“新建立案”后看到“新建案件立案工作台”。
  - _Leverage: `MVP_DASHBOARD_QUICK_ACTIONS`, existing Playwright helpers in `frontend/e2e/utils/test-helpers.ts`_
  - _Requirements: R1_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Frontend Router Developer | Task: Fix the dashboard create-case entry so every visible "新建立案" action either navigates to /case/create or has accurate copy if it goes to /case, then update dashboard/unit and Playwright assertions | Restrictions: Do not change unrelated dashboard statistics or cards, do not introduce new routes unless required, keep existing quick action structure | Success: Lawyer clicks "新建立案" from dashboard and lands on /case/create with heading "新建案件立案工作台"; tests assert the route_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 3. 移除正式新建案件流程中的样例预填数据
  - File: `frontend/src/pages/case/CreateCase.tsx`
  - File: `frontend/src/components/case/CompactCaseForm.tsx`
  - File: `frontend/src/components/case/CompactCaseFormWrapper.tsx`
  - File: `frontend/src/components/CreateCaseWizard.tsx`
  - File: `frontend/e2e/cases.spec.ts`
  - 搜索并隔离 `/case/create` 路径下的硬编码示例数据，尤其是“红杉资本投资管理咨询合同纠纷案”。
  - 将正式新建页初始化为 empty/default values。
  - 保留测试夹具和原型页中的演示数据，但不能作为正式创建表单默认值。
  - 新增 E2E：访问 `/case/create` 后，初始页面不得展示已预填的 demo 案件标题。
  - _Leverage: `frontend/e2e/utils/test-helpers.ts`, existing case form state and service_
  - _Requirements: R2_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Senior React Form Developer | Task: Remove demo prefilled business data from the formal /case/create workflow while preserving test fixtures and prototype/demo pages, then add Playwright coverage proving the create form starts blank | Restrictions: Do not delete E2E fixture data used for case lists or approvals, do not remove prototype/gallery demo content unless it is imported into the production route, do not weaken existing case-management assertions | Success: /case/create no longer shows "红杉资本投资管理咨询合同纠纷案" or other demo values on initial load, while list and approval fixture tests still pass_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 4. 防止“文件材料归档”中断立案流程
  - File: `frontend/src/pages/case/CreateCase.tsx`
  - File: `frontend/src/components/case/CompactCaseForm.tsx`
  - File: `frontend/e2e/cases.spec.ts`
  - 找到“打开文件材料归档”按钮的点击逻辑。
  - MVP 首选处理：在当前步骤内显示说明 modal/notice，并保留用户在 `/case/create`。
  - 如果必须跳转 `/file`，则附带 `returnTo` 和当前草稿上下文，并在目标页提供“返回本次立案”。
  - 新增 E2E：点击该按钮后不得把用户留在无返回入口的 `/file` 死端。
  - _Leverage: existing stepper and form state in case creation components_
  - _Requirements: R3_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Workflow UX Engineer | Task: Change the document/materials archive action so it preserves the current case creation context, preferably by showing an inline unavailable notice or modal instead of navigating to a dead-end /file page | Restrictions: Do not remove the document/materials step, do not discard form state, do not create a fake document upload backend | Success: A lawyer clicking "打开文件材料归档" stays in or can return to the same case-create workflow with current step and data preserved; Playwright covers the behavior_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 5. 修复律师账号审批详情动作权限展示
  - File: `frontend/src/pages/approval/ApprovalDetail.tsx`
  - File: `frontend/src/services/approval.ts`
  - File: `frontend/src/types/approval.ts`
  - File: `frontend/e2e/approval.spec.ts`
  - 梳理审批详情中的当前用户、当前审批人、审批状态和可用动作字段。
  - 如果后端返回 allowlist，则按 allowlist 渲染按钮。
  - 如果没有 allowlist，则使用保守 fallback：非当前审批人的律师账号不展示审批决策按钮。
  - 将现有 E2E 中“律师应该显示审批动作按钮”的断言改为“律师不应看到审批决策按钮”。
  - 如果有审批人 fixture，则新增审批人可见按钮用例；如果没有，记录为后续增强。
  - _Leverage: `seedAuthenticatedUser(page, 'lawyer')`, existing approval detail page_
  - _Requirements: R4_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Frontend Authorization Developer | Task: Correct approval detail action rendering so non-approver lawyer users cannot see or invoke approver-only decision actions, using backend available actions when present and a conservative current-approver fallback otherwise | Restrictions: Do not rely only on button CSS visibility, do not bypass backend authorization, do not make approver actions disappear for real approvers if that fixture exists | Success: Lawyer user no longer sees "同意并成案", "拒绝", or "退回修改" on approval detail; E2E asserts the corrected permission boundary_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 6. 修复客户管理“新增客户”入口反馈
  - File: `frontend/src/pages/client/ClientManagement.tsx`
  - File: `frontend/src/pages/client/ClientManagement.less`
  - File: `frontend/src/services/client.ts`
  - File: `frontend/e2e/clients.spec.ts`
  - 验证 `handleAdd` 是否被按钮触发、modal 是否挂载、是否被样式遮挡。
  - 确保点击“新增客户”后出现新增客户 modal/drawer 或明确提示。
  - 如果 modal 已存在，修复可见性、必填项和提交反馈。
  - 新增 E2E：点击“新增客户”后可见“新增客户”弹窗或明确不可用提示。
  - _Leverage: existing modal state and `clientService`_
  - _Requirements: R5_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: React Ant Design Developer | Task: Make the "新增客户" action produce visible, useful feedback by opening the existing customer modal/drawer or showing a clear unsupported message, then cover it with Playwright | Restrictions: Do not rewrite the whole client management page, do not remove existing edit/delete/view actions, preserve client list refresh behavior | Success: Lawyer clicks "新增客户" and immediately sees a visible creation UI or explicit message; no silent no-op remains_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 7. 为工作台全局搜索增加反馈状态
  - File: `frontend/src/pages/dashboard/Dashboard.tsx`
  - File: `frontend/src/pages/dashboard/Dashboard.module.css`
  - File: `frontend/e2e/dashboard.spec.ts`
  - 明确搜索触发方式：Enter、搜索按钮或即时搜索。
  - 添加 loading、results、empty、error 状态。
  - 如果没有统一搜索 API，先实现本地/路由级反馈：提示“按 Enter 搜索”或跳转到案件/客户列表并带 query。
  - 新增 E2E：输入“示例科技”后能看到结果、空态或可解释提示。
  - _Leverage: existing dashboard data service and route navigation_
  - _Requirements: R6_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Frontend Search UX Developer | Task: Add visible feedback to dashboard global search so lawyers can tell whether a query is idle, loading, empty, failed, or has results | Restrictions: Do not create a fake backend endpoint, do not block dashboard load, do not silently ignore input | Success: Searching "示例科技" produces an obvious UI response and Playwright asserts that response_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 8. 修复冲突检测页面横向溢出
  - File: `frontend/src/pages/conflict/ConflictWorkbench.tsx`
  - File: `frontend/src/pages/conflict/ConflictCheck.tsx`
  - File: `frontend/src/pages/conflict/ConflictCheck.less`
  - File: `frontend/src/pages/conflict/ConflictResult.less`
  - File: `frontend/e2e/layout.spec.ts`
  - 找到 `/conflict` 中撑破主内容区的表格、证据列、来源列或操作列。
  - 为 Ant Design Table 设置受控横向滚动容器，给长文本加 ellipsis/tooltip/detail。
  - 添加 CSS：主内容和表格父容器需要 `min-width: 0`，禁止全页横向滚动。
  - 新增布局 E2E：1366px 宽度下页面级 `scrollWidth` 不应明显超过 `innerWidth`。
  - _Leverage: existing conflict table components and styles_
  - _Requirements: R7_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Frontend Layout Engineer | Task: Contain the conflict detection table so long evidence/source/action columns do not expand the whole page, using Ant Design table scroll and responsive CSS | Restrictions: Do not hide required risk data without a tooltip/detail alternative, do not introduce page-level horizontal overflow, keep actions reachable | Success: /conflict remains readable and operable at 1366px with no page-level horizontal overflow; E2E verifies layout containment_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 9. 修复客户主档案横向溢出
  - File: `frontend/src/pages/client/ClientManagement.tsx`
  - File: `frontend/src/pages/client/ClientManagement.less`
  - File: `frontend/e2e/layout.spec.ts`
  - 调整客户统计、操作按钮、主档案内容区域的 grid/flex。
  - 按钮组在空间不足时换行或收进更多菜单。
  - 长客户名、联系人、地址等字段必须换行、截断或 tooltip，不得撑破容器。
  - 新增布局 E2E：1366px 和 1200px 宽度下 `/client` 无明显页面级横向溢出。
  - _Leverage: existing Ant Design Row/Col/Card/Table layout_
  - _Requirements: R8_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Responsive UI Engineer | Task: Fix client management layout overflow in stats, action buttons, and profile content using responsive grid/flex containment | Restrictions: Do not remove customer actions, do not make text unreadable, do not solve by globally hiding overflow on body | Success: /client has no page-level horizontal overflow at common desktop widths and all primary actions remain accessible_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 10. 修复修改密码表单字段级校验
  - File: `frontend/src/pages/profile/Profile.tsx`
  - File: `frontend/src/services/auth.ts`
  - File: `frontend/e2e/profile.spec.ts`
  - 为旧密码、新密码、确认密码添加 Ant Design Form rules。
  - 确认密码必须与新密码一致。
  - 客户端校验失败时不调用 API，不产生未捕获 console error。
  - API 失败时显示用户可见错误。
  - 新增 E2E：空表单提交后显示三个必填错误。
  - _Leverage: existing profile modal and auth service_
  - _Requirements: R10_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: React Form Validation Developer | Task: Add robust field-level validation to the change-password modal and ensure empty or mismatched submissions show clear errors without uncaught console failures | Restrictions: Do not change password API contract unless required, do not submit invalid forms to backend, do not remove existing profile editing behavior | Success: Empty password form shows field-level errors, mismatch is caught, valid submission still uses the existing service, and E2E covers validation_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 11. 修复帮助中心入口
  - File: `frontend/src/components/layout/Header.tsx`
  - File: `frontend/e2e/auth.spec.ts`
  - 点击用户菜单中的“帮助中心”后显示帮助 modal/drawer、帮助页，或明确“帮助中心建设中”提示。
  - 不允许静默返回工作台。
  - 新增 E2E：菜单点击后有可见帮助内容或建设中提示。
  - _Leverage: existing Header menu and Ant Design Modal/Drawer/message_
  - _Requirements: R9_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Frontend Navigation Developer | Task: Make the help center menu item produce visible help content or a clear unavailable notice instead of routing back to dashboard | Restrictions: Do not add a broken route, do not remove the help menu item, keep keyboard accessibility | Success: Clicking "帮助中心" gives a visible and testable response_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 12. 改进财务无权限页面说明
  - File: `frontend/src/pages/finance/FinanceManagement.tsx`
  - File: route guard or shared no-permission component if one exists
  - File: `frontend/e2e/finance.spec.ts`
  - 律师直访 `/finance` 时，展示无权限状态和所需角色/权限说明。
  - 保持侧边栏对律师隐藏财务入口，除非产品明确改变权限策略。
  - 提供返回工作台或返回上一页的安全动作。
  - _Leverage: existing finance page and finance E2E_
  - _Requirements: R11_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Access-State UX Developer | Task: Improve the /finance no-permission state for lawyer users by explaining the missing role or permission and offering a safe return action | Restrictions: Do not grant finance access to lawyers, do not add finance to lawyer sidebar, do not change backend permission checks | Success: Direct /finance access shows clear permission context and E2E asserts the copy_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 13. 修复通知中心空状态和“全部已读”动作
  - File: `frontend/src/components/layout/Header.tsx`
  - File: `frontend/src/hooks/useNotifications.ts`
  - File: `frontend/e2e/auth.spec.ts`
  - 当通知列表为空或无未读通知时，隐藏或禁用“全部已读”。
  - 如果有通知设置/历史入口，则在空状态中提供；如果没有，不展示无效动作。
  - 确保 mark-all-read 失败时仍有用户可见反馈。
  - _Leverage: existing notification popover and `useNotifications` hook_
  - _Requirements: R12_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Notification UI Developer | Task: Improve the notification popover empty state so "全部已读" is hidden or disabled when there is nothing to mark, and provide useful empty-state copy | Restrictions: Do not create fake notifications, do not call mark-all-read when there are no unread notifications, do not remove error handling | Success: Empty notifications show a coherent state without active no-op actions; tests verify the behavior_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 14. 更新 E2E 用例到当前路由、账号和选择器
  - File: `frontend/e2e/auth.spec.ts`
  - File: `frontend/e2e/cases.spec.ts`
  - File: `frontend/e2e/approval.spec.ts`
  - File: `frontend/e2e/finance.spec.ts`
  - File: `frontend/e2e/utils/test-helpers.ts`
  - File: new specs as needed: `frontend/e2e/clients.spec.ts`, `frontend/e2e/dashboard.spec.ts`, `frontend/e2e/layout.spec.ts`, `frontend/e2e/profile.spec.ts`
  - 使用律师账号 `lawyer / Demo@2026` 覆盖本 Spec 中的核心律师流程。
  - 将旧断言中与修复目标相反的期望改为新产品行为，例如审批按钮权限。
  - 避免 fragile selector，优先使用 role、label、heading、button name。
  - _Leverage: existing Playwright config and helpers_
  - _Requirements: All_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Playwright QA Automation Engineer | Task: Update and extend E2E coverage so it reflects the current routes, lawyer account, and corrected selectors for all QA remediation behavior | Restrictions: Do not skip tests to hide bugs, do not assert implementation details where accessible roles exist, keep tests deterministic and isolated | Success: Playwright covers the fixed lawyer workflow and all P1/P2 regressions, with stable selectors and current account setup_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

- [x] 15. 执行完整验证并回填 QA 状态
  - File: `reports/frontend-lawyer-qa-2026-05-25.md`
  - File: optionally create `reports/frontend-lawyer-qa-remediation-2026-05-25.md`
  - 运行后端构建、前端构建、E2E 和可行的类型检查/静态检查。
  - 对 QA-001 到 QA-012 标记：已修复、延期、产品确认不改、仍失败。
  - 如果 `npm run type-check` 存在历史失败，记录失败摘要，并确认本轮改动没有新增相关错误。
  - _Leverage: repository verification commands and QA report_
  - _Requirements: All_
  - _Prompt: Implement the task for spec lawyer-frontend-qa-remediation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Release QA Engineer | Task: Run full verification after remediation, document command results, and update the QA issue status for QA-001 through QA-012 | Restrictions: Do not mark issues fixed without reproducing or test evidence, do not ignore failing verification commands, do not delete the original QA report | Success: Verification results are recorded, all QA items have explicit final status, and remaining risks are documented_
  - Instructions: Set this task to `[-]` when starting, log implementation notes after completion, then set to `[x]`.

## Execution Order

1. Task 1: establish traceability.
2. Tasks 3, 4, 5, 6: fix P1 blockers.
3. Tasks 2, 7, 8, 9, 10: fix P2 usability and layout issues.
4. Tasks 11, 12, 13: fix P3 polish issues.
5. Task 14: keep E2E current throughout implementation; finalize after all behavior changes.
6. Task 15: run verification and update QA status.

## Definition of Done

- All P1 QA issues QA-003, QA-004, QA-005, and QA-008 are fixed and covered by E2E.
- P2 issues have either fixes or explicit deferral notes with product rationale.
- P3 issues have coherent empty/no-permission/help states.
- No visible button tested in the lawyer workflow remains a silent no-op.
- `/case/create` no longer initializes with formal-business demo data.
- Lawyer account no longer sees approval-only decision actions unless it is also the current approver.
- `/conflict` and `/client` do not produce page-level horizontal overflow at common desktop widths.
- Verification commands have been run and results documented.
