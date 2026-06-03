## Context

The retest was performed from the perspective of a lawyer using the internal browser against `http://127.0.0.1:3003/` with pages showing "正式 API". It confirmed that recent remediation improved core navigation, but several controls still violate the same interaction rule: a button must either perform its visible promise, clearly explain why it cannot, or be unavailable before the user clicks it.

This change is intentionally frontend-heavy. It focuses on deterministic UI state, clear button semantics, and E2E coverage. Backend changes should be limited to exposing missing data needed for correct frontend state, such as a related case ID or export failure reason.

## Goals

- Prevent formal conflict-check actions from creating partial side effects when required intake fields are missing.
- Make all tested lawyer-facing buttons produce one of four observable outcomes: route change, modal/drawer open, download/export feedback, or explicit disabled/unavailable reason.
- Remove same-page button name ambiguity where a lawyer, screen reader, or E2E test cannot distinguish actions.
- Keep behavior aligned with the existing Ant Design and React Router patterns in the repository.
- Add regression tests that can be run without relying on the internal browser text-entry capability.

## Non-Goals

- Do not redesign page layout or visual hierarchy beyond what is needed to expose state and feedback.
- Do not remove existing test assertions to make the suite pass.
- Do not silently bypass backend validation by inventing placeholder field values.
- Do not introduce a new global state library.

## Design Decisions

### Intake Preflight Validation

`Batch01Prototype.tsx` already tracks the intake form state and can display required-field completion in the right summary panel. The conflict-check action should use the same state to derive a preflight validation result before creating or updating an intake draft.

Required fields for running conflict check in this remediation:

- `caseTitle` or equivalent matter title field
- `clientId` and `clientName`
- at least one counterparty / opponent name
- responsible lawyer ID
- matter type or case type, if already marked required by the form
- business domain and subdomain, if already marked required by the form

The exact field names must be confirmed in code before editing. The implementation should not widen scope to new business rules that are not already represented in the UI.

### Side-Effect Ordering

The conflict-check button must follow this order:

1. Compute missing required fields locally.
2. If anything is missing, show inline errors and a top-level message, keep the user on the same step, and do not call create-draft, update-draft, or conflict-check APIs.
3. If all required fields are present, save or update the intake draft.
4. Create or reuse the conflict-check task.
5. Update the in-workflow status and allow the user to continue to team/fees.

### Dashboard Action Semantics

Dashboard "view all" actions must not be inert. Use existing routes where available:

- My todos: route to inbox/task center, or expand a full dashboard list if no route exists.
- Conflict reviews: route to conflict detection/review queue.

Each button must have a unique accessible name. Prefer visible labels that are unique. If visual copy must remain short, add a unique `aria-label`, but E2E tests must use the exact intended accessible name.

### Approval Detail Action Semantics

The approval detail page currently exposes repeated "更多操作" buttons and a "查看关联案件" button that may be disabled without explanation.

The UI should distinguish actions by location and purpose:

- Header overflow: "更多审批操作" or `aria-label="更多审批操作"`
- Decision area overflow: "更多处理方式" or `aria-label="更多处理方式"`
- Related case: enabled only when a routeable case ID exists; otherwise disabled with visible tooltip or adjacent helper text such as "暂无关联案件，审批通过后生成案件".

### Client Profile Quick Actions

Client profile quick actions should be deterministic:

- "新增联系人": open an Ant Design modal/drawer with required fields, or scroll to and focus an inline contact form if that is the established page pattern.
- "上传附件": open upload control/modal, or show "附件上传暂不可用" if not implemented.
- "导出客户档案": trigger existing export/download behavior. If export is unavailable or fails, show a toast/message with the reason.

Do not leave a quick action as a no-op. If implementation is intentionally deferred, expose that state in the UI and in tests.

## Implementation Notes

- Prefer existing helpers, route constants, services, and Ant Design message/modal patterns.
- Keep changes scoped to the four affected pages and their E2E tests.
- Add stable selectors only when role/name-based selection cannot express the intent. Prefer accessible names first.
- Avoid creating new demo data in E2E unless the test needs isolation; use existing authenticated setup patterns.
- Internal browser text entry is limited, so automated regression should cover both click-only negative paths and full Playwright text-entry paths.

## Risks

- Existing E2E tests may rely on old short labels like "查看全部" or "更多操作"; update tests to use the new exact accessible names.
- Backend API may create a draft before returning validation errors. If so, frontend preflight is still required to prevent avoidable calls, and backend validation should remain as defense in depth.
- Export behavior may differ by browser context. Tests should assert visible success/error feedback rather than only file-system download when download handling is not stable.
