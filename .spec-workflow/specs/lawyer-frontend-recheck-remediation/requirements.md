# Requirements Document

## Introduction

This specification covers the second remediation pass after the lawyer-user frontend recheck performed on 2026-05-26. The recheck confirmed that several prior fixes are working, but it also found remaining blockers and usability defects in the lawyer workflow. The most serious issue is that the formal case-intake workflow loses context after running conflict detection, preventing lawyers from completing the full path from dashboard entry to approval submission.

The goal of this spec is to make the lawyer-facing frontend workflow coherent, continuous, and testable:

1. A lawyer can start a case from the dashboard, fill the intake form, run conflict detection, continue to team/fee/material steps, and submit approval without being ejected from the workflow.
2. Dashboard and client-profile actions must either navigate, open the appropriate UI, start a download, or show a clear unavailable message.
3. Approval list and detail views must make primary actions discoverable and must show a consistent current-approver model.
4. Repeated button names must be disambiguated enough for keyboard users, screen readers, and Playwright selectors.

Source QA report:

- `reports/frontend-lawyer-qa-recheck-2026-05-26.md`

## Alignment with Product Vision

Law OA Go is a law-firm operations system. For the lawyer role, the primary product value is reducing friction in intake, conflict review, client management, and approval follow-up. The defects in this spec directly affect trust in those workflows:

- Losing the intake context after conflict detection breaks the main case-creation job.
- Silent buttons reduce confidence and make the product feel unfinished.
- Inconsistent approver state creates legal-operation risk because users cannot tell who is allowed to approve.
- Truncated approval actions slow down repeated legal review work.

This remediation supports the MVP goal of making a usable lawyer workbench backed by formal APIs.

## Requirements

### Requirement 1: Preserve Case Intake Context Through Conflict Detection

**User Story:** As a lawyer, I want conflict detection to run inside the current case-intake workflow, so that I can continue from basic information to team assignment, materials, and approval submission without losing my draft.

#### Acceptance Criteria

1. WHEN a lawyer fills `/case/create` and clicks `运行利益冲突检查` THEN the system SHALL keep the user in the case-intake workflow or return them to the same workflow state automatically.
2. WHEN conflict detection succeeds THEN the system SHALL store the conflict result in the current intake runtime state and SHALL enable `进入团队与费用`.
3. WHEN conflict detection succeeds THEN the displayed result SHALL reference the current intake draft data, including current case title, selected client, and opposing party, not an unrelated historical conflict task.
4. IF the backend returns a conflict task/check ID THEN the system SHALL display that ID in the intake conflict step and carry it into approval submission payloads.
5. IF conflict detection fails THEN the system SHALL show a visible error message and SHALL keep all user-entered intake form data.
6. WHEN a lawyer clicks browser Back or page navigation after conflict detection THEN the system SHALL not silently discard the active intake draft.

Traceability:

- QA-RC-001

### Requirement 2: Complete the Case Intake Workflow After Conflict Detection

**User Story:** As a lawyer, I want the post-conflict steps to remain reachable, so that I can finish team/fee configuration, materials handling, and approval submission in one continuous workflow.

#### Acceptance Criteria

1. WHEN the conflict step has a successful result THEN `进入团队与费用` SHALL be enabled and clickable.
2. WHEN the lawyer enters `团队与费用` THEN the system SHALL show controls for responsible lawyer, billing method, expected fee, and related fields already present in the current UI.
3. WHEN the lawyer enters `文档与材料` THEN the system SHALL keep the same intake draft and SHALL show material actions without navigating to a dead-end.
4. WHEN the lawyer reaches `立案提交` THEN the summary SHALL include current case title, client, opposing party, conflict status, team, fee, and material status.
5. WHEN the lawyer clicks `提交审批并等待成案` with required fields and conflict result present THEN the system SHALL call the existing approval integration and show success or actionable backend error.

Traceability:

- QA-RC-001

### Requirement 3: Make Dashboard “View All” Actions Explicit

**User Story:** As a lawyer, I want every dashboard `查看全部` action to produce a clear result, so that I know whether I navigated, expanded a list, or the feature is unavailable.

#### Acceptance Criteria

1. WHEN the lawyer clicks `利益冲突待复核` -> `查看全部` THEN the system SHALL navigate to the conflict workbench/list or visibly expand the section.
2. WHEN the lawyer clicks `我的待办` -> `查看全部待办` THEN the system SHALL navigate to a task/inbox/workbench route or visibly show a supported message.
3. IF a target route is not implemented THEN the system SHALL show an Ant Design modal/message explaining the current MVP limitation and next available action.
4. Silent no-op buttons SHALL NOT remain on the dashboard.

Traceability:

- QA-RC-002

### Requirement 4: Keep Approval List Actions Discoverable at Desktop Widths

**User Story:** As a lawyer, I want the approval list action column to be visible or predictably scrollable, so that I can open approval details without fighting the table layout.

#### Acceptance Criteria

1. WHEN `/approval` is viewed at 1366px wide THEN the `操作` column and `进入审批` buttons SHALL be visible or reachable via an obvious table-internal horizontal scroll area.
2. WHEN `/approval` is viewed at 1200px wide THEN the page SHALL NOT rely on page-level horizontal overflow to access `进入审批`.
3. The approval table SHALL use stable column widths or responsive wrapping/truncation so long titles do not push the action column out of view.
4. The action column SHOULD be fixed to the right when table width exceeds the content container.

Traceability:

- QA-RC-003

### Requirement 5: Make Approval Current-Approver State Consistent

**User Story:** As a lawyer, I want approval details to show one consistent current approver and permission model, so that I can trust whether I may approve or only view progress.

#### Acceptance Criteria

1. WHEN an approval detail loads THEN the header current approver, workflow current node, and bottom action area SHALL be derived from a single normalized approval state.
2. IF the current user is not the current approver THEN decision buttons such as `同意并成案`, `拒绝`, and `退回修改` SHALL NOT be shown.
3. IF the current user is the current approver and the approval is in an actionable state THEN decision buttons SHALL be shown.
4. IF backend `available_actions` is present THEN the frontend SHALL use it as the primary permission source.
5. IF backend `available_actions` is absent THEN the frontend SHALL conservatively compare current user ID/email/name against normalized current-approver fields.
6. The UI SHALL not show contradictory values such as `当前审批人：demo.lawyer@example.test` while the workflow current node names another approver unless both are explicitly labeled as separate concepts.

Traceability:

- QA-RC-004

### Requirement 6: Add Feedback to Client Profile Quick Actions

**User Story:** As a lawyer, I want client quick actions to open useful UI or explain their status, so that I know whether my click worked.

#### Acceptance Criteria

1. WHEN the lawyer clicks `新增联系人` THEN the system SHALL open a contact modal/drawer or show a clear `功能建设中` style message.
2. WHEN the lawyer clicks `上传附件` THEN the system SHALL open an upload UI or show a clear unavailable message.
3. WHEN the lawyer clicks `导出客户档案` THEN the system SHALL trigger a download or show a success/failure/unavailable message.
4. Each quick action SHALL provide visible feedback within 1 second.
5. Quick actions SHALL NOT only move focus or scroll without explanation.

Traceability:

- QA-RC-005

### Requirement 7: Disambiguate Repeated Accessible Names

**User Story:** As a keyboard, screen-reader, or automated-test user, I want repeated buttons to have contextual names, so that controls can be reliably identified.

#### Acceptance Criteria

1. Repeated approval row actions SHALL expose an accessible name containing the approval number or title, for example `进入审批 AP-20260525-563000`.
2. Conflict risk filters SHALL expose contextual names, for example `筛选高风险`, `筛选中风险`, `筛选低风险`.
3. Client tabs and client quick actions SHALL expose distinct accessible names when duplicate visible text exists.
4. E2E selectors SHOULD prefer role/name selectors and SHOULD NOT depend on brittle CSS selectors when accessible names can be improved.

Traceability:

- QA-RC-006

### Requirement 8: Update Regression Coverage for the Recheck Issues

**User Story:** As an engineering team, we want automated regression tests for every recheck finding, so that future changes do not reintroduce the same defects.

#### Acceptance Criteria

1. A Playwright test SHALL cover the full lawyer case-intake workflow through conflict check and `团队与费用`.
2. A Playwright test SHALL cover dashboard `查看全部` actions and assert navigation or visible feedback.
3. A Playwright test SHALL cover approval table action-column accessibility at 1366px and 1200px.
4. A Playwright test SHALL cover approval current-approver consistency for current approver and non-current approver cases.
5. A Playwright test SHALL cover client quick-action feedback.
6. The final QA report SHALL update statuses for QA-RC-001 through QA-RC-006.

Traceability:

- All QA-RC items

## Non-Functional Requirements

### Code Architecture and Modularity

- Keep changes narrowly scoped to the existing frontend routes and components unless the implementation proves shared helpers are necessary.
- Prefer existing `Batch01Prototype` patterns for this MVP surface.
- Extract local helper functions only when they reduce duplication or make permission/context rules clearer.
- Do not introduce new global state libraries for this remediation.

### Performance

- Conflict detection UI feedback SHALL appear immediately when the request starts.
- Case-intake state updates SHALL not cause whole-page reloads.
- Table layout fixes SHALL not materially slow down approval list rendering.

### Security

- Frontend permission hiding is not a backend authorization substitute.
- Approval decision buttons SHALL remain gated by backend authorization.
- Do not grant finance, admin, or approval permissions to the lawyer role to make tests pass.

### Reliability

- API failure states SHALL preserve user-entered form data.
- Navigation changes SHALL not discard active intake runtime state without explicit user confirmation.
- Regression tests SHALL use deterministic seeded data or existing E2E route mocks.

### Usability

- Every visible button tested in this spec must produce navigation, visible state change, download, modal/drawer, toast, or explicit unavailable copy.
- Primary legal workflow actions must be visible in common desktop widths.
- User-facing copy must be short, concrete, and action-oriented.
