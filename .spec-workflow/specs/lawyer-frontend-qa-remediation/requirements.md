# Requirements Document

## Introduction

本 Spec 基于 2026-05-25 律师视角前端功能测试记录，修复律师用户在工作台、立案、冲突检测、审批、客户管理、个人中心、通知和帮助入口中遇到的可用性与权限展示问题。

本轮目标不是重做律师端 UI，而是把已经被手工点击验证过的问题转化为可验收、可测试、可分派的前端修复范围，确保其他执行模型可以按任务逐项落地。

原始 QA 记录：

- `reports/frontend-lawyer-qa-2026-05-25.md`

测试基线：

- Frontend: `http://127.0.0.1:3003/`
- Backend: `http://127.0.0.1:8080/`
- Account: `lawyer / Demo@2026`
- Role: 律师使用者
- Method: Chrome / Playwright / Computer Use 手工点击

## Alignment with Product Vision

Law OA Go 的律师端核心价值是帮助律师快速完成接案、冲突检查、审批提交、客户资料维护和后续案件管理。当前 QA 记录显示，核心链路可以跑通，但若正式环境展示样例数据、流程中断、权限按钮误导或入口静默无反馈，会直接降低律师对系统的信任。

本 Spec 支持以下产品目标：

- 降低律师从工作台发起接案的操作成本。
- 避免正式环境误提交样例案件数据。
- 保持立案五步流程的上下文连续性。
- 明确律师与审批人的权限边界。
- 让所有可点击入口都有明确反馈。
- 提高冲突检测与客户管理页面在常见桌面宽度下的可读性。

## Source QA Mapping

| QA ID | Title | Priority | Requirement |
| --- | --- | --- | --- |
| QA-001 | 工作台“新建立案”入口没有直达新建页 | P2 | R1 |
| QA-002 | 工作台全局搜索缺少结果反馈 | P2 | R6 |
| QA-003 | 正式环境新建案件页预填样例数据 | P1 | R2 |
| QA-004 | 文档材料归档按钮中断立案流程 | P1 | R3 |
| QA-005 | 律师账号可见审批决策按钮 | P1 | R4 |
| QA-006 | 冲突检测表格超出内容区 | P2 | R7 |
| QA-007 | 客户主档案主内容横向溢出 | P2 | R8 |
| QA-008 | 新增客户入口点击后无明显反馈 | P1 | R5 |
| QA-009 | 帮助中心入口没有帮助内容 | P3 | R9 |
| QA-010 | 修改密码空表单反馈不明显 | P2 | R10 |
| QA-011 | 律师直访财务入口显示无权访问，但侧边栏没有财务入口 | P3 | R11 |
| QA-012 | 通知中心为空但没有入口动作 | P3 | R12 |

## Requirements

### Requirement 1: 工作台新建立案入口必须进入正确流程

**User Story:** As a 律师, I want 工作台的“新建立案”入口直接打开新建案件流程, so that 我可以从首页快速开始接案而不需要再在案件列表中寻找入口。

#### Acceptance Criteria

1. WHEN 律师登录后点击工作台中的“新建立案”入口 THEN the system SHALL navigate to `/case/create`.
2. IF 产品决定该入口只进入案件列表 THEN the system SHALL rename the button to “案件管理” or “进入案件管理”.
3. WHEN the system navigates to `/case/create` THEN the page SHALL display heading “新建案件立案工作台”.
4. WHEN the quick action route is changed THEN existing dashboard quick action tests SHALL be updated to assert the current route.

### Requirement 2: 正式和 QA 业务环境的新建案件表单不得预填样例数据

**User Story:** As a 律师, I want 新建案件页默认展示空白业务表单, so that 我不会误提交演示数据或污染真实案件数据。

#### Acceptance Criteria

1. WHEN 律师访问 `/case/create` THEN the system SHALL render an empty case creation form by default.
2. WHEN the form first loads THEN the system SHALL NOT prefill case title, client, opposing party, facts, description, fee, lawyer, or other business fields with demo data such as “红杉资本投资管理咨询合同纠纷案”.
3. IF demo data is required for prototype or visual gallery pages THEN the system SHALL isolate it behind a clearly named demo route, mock fixture, or explicit demo flag.
4. IF existing E2E fixtures require “红杉资本” data for lists or approval queues THEN those fixtures SHALL remain in test helpers and SHALL NOT be used as default values for `/case/create`.
5. WHEN this requirement is implemented THEN an E2E test SHALL assert that the create case form does not contain the known demo case title on initial load.

### Requirement 3: 文件材料归档入口不得中断当前立案流程

**User Story:** As a 律师, I want 在立案第 4 步处理文档材料时保留当前立案上下文, so that 我不会因为查看文档功能而丢失已填写内容或当前步骤。

#### Acceptance Criteria

1. WHEN 律师在 `/case/create` 的“文档与材料”步骤点击“打开文件材料归档” THEN the system SHALL keep the current case creation context available.
2. IF 文档中心当前未纳入 MVP THEN the system SHALL show an inline notice, disabled state, modal, or drawer explaining that the feature is unavailable.
3. IF the system navigates away from `/case/create` THEN it SHALL preserve a return path and show a visible “返回本次立案” action.
4. WHEN 律师 returns to the case creation flow THEN the system SHALL keep the current step and previously entered form data.
5. WHEN the feature is unavailable THEN the button SHALL NOT silently navigate to a dead-end page.

### Requirement 4: 律师账号不得看到无权限执行的审批决策按钮

**User Story:** As a 律师, I want 只看到我当前可以执行的审批动作, so that 我不会误以为自己可以审批自己的提交或他人的审批事项。

#### Acceptance Criteria

1. WHEN a lawyer views an approval detail and is not the current approver THEN the system SHALL hide or disable decision actions including “同意并成案”, “拒绝”, “退回修改”, and approval-only “更多操作”.
2. WHEN approval actions are hidden or disabled THEN the system SHALL show status text describing the current approval state and, when available, the current approver.
3. IF backend approval detail includes allowed actions THEN the frontend SHALL render actions from that allowlist.
4. IF backend approval detail does not include allowed actions THEN the frontend SHALL use a conservative permission fallback and hide approval-only actions for non-approver lawyer accounts.
5. WHEN an actual approver views the same detail THEN permitted approval actions SHALL remain visible and functional.
6. WHEN the lawyer account runs E2E tests THEN approval detail tests SHALL assert that lawyer users cannot see or click approver-only decision buttons.

### Requirement 5: 新增客户入口必须有明确反馈

**User Story:** As a 律师, I want 点击“新增客户”后立即看到新增客户表单或明确提示, so that 我可以在接案前维护客户资料。

#### Acceptance Criteria

1. WHEN 律师 clicks “新增客户” on `/client` THEN the system SHALL open a customer creation modal/drawer or navigate to a customer creation route.
2. IF customer creation is unavailable in the current version THEN the system SHALL show a visible message explaining that it is not currently supported.
3. WHEN the modal opens THEN it SHALL include required fields for customer name and customer type at minimum.
4. WHEN the user cancels the modal THEN the page SHALL return to the original customer list/profile state without data loss.
5. WHEN customer creation succeeds THEN the system SHALL refresh or update the customer list and show a success message.

### Requirement 6: 工作台全局搜索必须提供结果反馈

**User Story:** As a 律师, I want 使用工作台全局搜索时看到搜索是否生效, so that 我可以快速定位案件、客户或审批事项。

#### Acceptance Criteria

1. WHEN 律师 enters a query in dashboard global search THEN the system SHALL provide an obvious interaction model such as Enter-to-search, search button, or live results.
2. WHEN matching results exist THEN the system SHALL show grouped results or navigate to a result page/list with the query applied.
3. WHEN no matching results exist THEN the system SHALL show an empty state such as “未找到相关案件、客户或文档”.
4. WHEN search is loading THEN the system SHALL show a loading indicator.
5. WHEN search fails THEN the system SHALL show a recoverable error message.

### Requirement 7: 冲突检测表格不得撑破主内容区

**User Story:** As a 律师, I want 冲突检测命中结果在普通桌面宽度下保持可读和可操作, so that 我可以高效复核风险命中。

#### Acceptance Criteria

1. WHEN 律师 visits `/conflict` on 1200px to 1440px desktop widths THEN the main content SHALL NOT horizontally overflow the viewport.
2. WHEN conflict result text is long THEN the table SHALL use controlled wrapping, ellipsis, tooltip, or detail drawer.
3. WHEN columns exceed available width THEN the table SHALL use a bounded scroll container instead of expanding the whole page.
4. WHEN actions are shown in the table THEN action buttons SHALL remain reachable.
5. WHEN E2E layout checks run THEN `document.documentElement.scrollWidth` SHALL NOT exceed the viewport by more than the accepted tolerance for the page shell.

### Requirement 8: 客户主档案内容不得横向溢出

**User Story:** As a 律师, I want 客户管理页面的统计、操作和主档案信息保持在可视区域内, so that 我可以发现并点击常用客户操作。

#### Acceptance Criteria

1. WHEN 律师 visits `/client` on 1200px to 1440px desktop widths THEN the customer management main content SHALL NOT overflow horizontally.
2. WHEN action buttons do not fit on one row THEN they SHALL wrap, collapse, or move into a menu.
3. WHEN statistics cards are shown THEN the grid SHALL use responsive columns with stable min/max widths.
4. WHEN long customer names or contact fields are shown THEN text SHALL wrap, truncate, or use tooltip without breaking layout.
5. WHEN layout E2E tests run THEN the page SHALL remain usable at common desktop and tablet widths.

### Requirement 9: 帮助中心入口不得返回无关页面

**User Story:** As a 律师, I want 点击“帮助中心”后看到帮助内容或明确说明, so that 我遇到流程问题时能自助排查。

#### Acceptance Criteria

1. WHEN 律师 clicks “帮助中心” in the user menu THEN the system SHALL show help content, help drawer/modal, or an explicit “帮助中心建设中” message.
2. WHEN help content is not implemented THEN the system SHALL NOT silently route back to dashboard.
3. WHEN the help entry is displayed THEN it SHALL be keyboard accessible.

### Requirement 10: 修改密码空表单必须展示字段级校验

**User Story:** As a 律师, I want 修改密码表单在缺少必填项时直接提示具体字段, so that 我知道如何修正输入。

#### Acceptance Criteria

1. WHEN 律师 opens 修改密码 modal and submits empty fields THEN the system SHALL show field-level required validation for old password, new password, and confirm password.
2. WHEN new password and confirm password do not match THEN the system SHALL show a field-level mismatch message.
3. WHEN backend change-password request fails THEN the system SHALL show a user-visible error message.
4. WHEN validation fails on the client THEN the system SHALL NOT produce uncaught console errors.
5. WHEN validation passes THEN the existing change-password API integration SHALL remain functional.

### Requirement 11: 财务无权限页面必须说明所需权限

**User Story:** As a 律师, I want 访问无权限财务页面时知道缺少什么权限, so that 我可以判断是否需要联系管理员。

#### Acceptance Criteria

1. WHEN 律师 directly visits `/finance` THEN the system SHALL show a no-permission state.
2. WHEN the no-permission state is shown THEN it SHALL include required role or permission context such as “需要财务角色或管理员授权”.
3. WHEN the lawyer role lacks finance access THEN the sidebar SHALL continue to hide the finance entry unless product rules change.
4. WHEN possible THEN the no-permission state SHALL provide a safe return action to dashboard or previous page.

### Requirement 12: 通知中心空状态不得展示无效动作

**User Story:** As a 律师, I want 通知中心为空时看到合理空状态和可用下一步, so that 我不会点击无效的“全部已读”动作。

#### Acceptance Criteria

1. WHEN notification list is empty THEN the system SHALL show “暂无通知” or equivalent empty state.
2. WHEN there are no unread notifications THEN the “全部已读” action SHALL be hidden, disabled, or accompanied by no-op-safe behavior.
3. IF notification history or settings exists THEN the empty state SHALL provide a link to history/settings.
4. IF notification history or settings does not exist THEN the empty state SHALL avoid presenting dead-end actions.

## Non-Functional Requirements

### Code Architecture and Modularity

- Keep fixes scoped to the affected pages/components and shared helpers.
- Prefer existing Ant Design patterns already used in the frontend.
- Do not introduce a new UI framework.
- Do not duplicate API clients when an existing service already exists.
- Keep permission rendering isolated in small helper functions or component-level selectors.

### Performance

- Dashboard search feedback SHALL avoid blocking first paint.
- Layout fixes SHALL not introduce heavy runtime measurement loops.
- E2E layout checks SHALL use deterministic assertions rather than screenshot-only checks.

### Security

- Frontend permission hiding is only a UX layer. Backend authorization SHALL remain the source of truth.
- Do not expose approval actions for non-approvers in the UI.
- Do not add credentials or secrets to test files.

### Reliability

- Existing Playwright E2E tests SHALL continue to pass after route, account, and selector updates.
- Tests SHALL not be weakened to pass. If product behavior changes, tests must assert the new behavior explicitly.
- Demo/test fixtures SHALL remain isolated from production-like form defaults.

### Usability

- Every visible button or menu item SHALL produce a visible result: navigation, modal/drawer, toast, inline validation, disabled explanation, or empty state.
- Main content SHALL remain readable and operable at 1200px, 1366px, 1440px, and common tablet widths.
- Error and empty states SHALL be written for lawyer users, not developers.

