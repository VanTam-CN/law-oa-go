## Why

The 2026-05-26 lawyer-facing internal browser retest found that the main intake, dashboard, approval, and client-profile flows are navigable, but several buttons still create ambiguous or silent outcomes. These issues do not require a new product surface, but they do affect the trustworthiness of the lawyer workflow and must be specified before the next remediation pass.

The most important failure is in the case intake conflict-check step: a lawyer can trigger conflict checking while required matter data is incomplete, causing the UI to create an intake draft and then report a validation error while the conflict status remains undetected. This leaves users uncertain whether the draft is valid, whether the check ran, and whether they may continue.

## What Changes

- Tighten intake conflict-check preconditions so incomplete required fields block the action before any draft or conflict side effect is created.
- Require visible, deterministic feedback for all dashboard "view all" actions.
- Require unique visible or accessible names for repeated dashboard and approval-detail buttons.
- Clarify disabled approval actions, especially "view related case", with a reason and expected enabled behavior.
- Require customer-profile quick actions to open a form, start a download, navigate, or show an explicit unavailable/error message.
- Add E2E coverage for the retest findings so the internal browser issues are captured by automated regression tests.

## Impact

- Affected specs:
  - `core-intake-workflow`
  - `dashboard`
  - `approval-system`
  - `client-management`
- Affected frontend code:
  - `frontend/src/pages/batch01/Batch01Prototype.tsx`
  - `frontend/src/pages/dashboard/Dashboard.tsx`
  - `frontend/src/pages/approval/ApprovalDetail.tsx`
  - `frontend/src/pages/client/ClientManagement.tsx`
  - related frontend service helpers and route constants if navigation targets are centralized
- Affected tests:
  - `frontend/e2e/case-create-full-workflow.spec.ts`
  - `frontend/e2e/dashboard-actions.spec.ts`
  - `frontend/e2e/approval-layout.spec.ts`
  - `frontend/e2e/approval-permission-consistency.spec.ts`
  - `frontend/e2e/client-profile-actions.spec.ts`
- QA source:
  - `reports/frontend-lawyer-qa-internal-browser-retest-2026-05-26.md`

## Out of Scope

- Rebuilding the entire intake workbench.
- Changing backend approval or conflict domain models unless the existing API cannot expose required disabled-state reasons or related-case identifiers.
- Implementing real document export infrastructure beyond the existing frontend/backend capability; if export is not available, the required behavior is an explicit unavailable/error message.
