## Why

Batch 01 PRD defines a focused "core intake workflow" that connects dashboard priorities, client master data, case intake, conflict checking, and approval decisions. The current system has these capabilities as separate modules, but the user flow is fragmented and lacks a reliable end-to-end intake control path.

The change creates a coherent intake-before-engagement workflow so attorneys, assistants, compliance reviewers, and partners can move from a new matter lead to a conflict-aware approval decision with traceable state, immutable review context, and clear handoffs.

## What Changes

- Add a core intake workflow across five Batch 01 surfaces: dashboard command center, client master profile, case intake workbench, conflict check results, and conflict-linked approval.
- Add client-domain requirements for rendering the client master profile, including completeness scoring, missing-field indicators, related-party summaries, and optimistic concurrency protection.
- Define conflict checks as asynchronous tasks with status polling, background progress visibility, failure handling, and result retrieval.
- Require approval submissions to persist immutable snapshots of client, matter, counterparty, conflict findings, risk scores, and evidence summaries.
- Add a waiver review subflow for waiver-eligible conflict results, including waiver reason, supporting attachments, designated approvers, and explicit waiver outcomes.
- Add positive handoff notifications so assistants can transfer completed client or intake data back to attorneys or compliance reviewers.
- Add anti-duplication safeguards for approval submission and conflict task creation.

## Impact

- Affected specs:
  - `core-intake-workflow`
  - `client-management`
  - `case-management`
  - `conflict-detection`
  - `approval-system`
  - `dashboard`
  - `inbox`
- Affected code:
  - `frontend/src/pages/dashboard/Dashboard.tsx`
  - `frontend/src/pages/client/ClientManagement.tsx`
  - `frontend/src/pages/case/CreateCase.tsx`
  - `frontend/src/pages/conflict/ConflictDetectionV2.tsx`
  - `frontend/src/pages/approval/ApprovalDetail.tsx`
  - client, enhanced case, conflict, approval, integration, dashboard, and inbox services/handlers/routes
  - database migrations for conflict task status, approval snapshots, waiver requests, client versioning, and idempotency tokens

**BREAKING**: Client update and approval submission APIs may require additional version/idempotency fields. Existing callers must be updated to send these fields or handle conflict responses.
