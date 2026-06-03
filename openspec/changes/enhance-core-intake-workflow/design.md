## Context

Batch 01 is the first productized workflow around matter intake and pre-engagement risk control. The existing system already has dashboard, client, case, conflict, approval, integration, and inbox modules, but the intake path crosses these modules without a single workflow contract.

The design turns these independent capabilities into one workflow without moving the system away from the current monolithic Gin/GORM/React architecture.

## Goals

- Provide a continuous workflow from dashboard priority to client verification, intake, conflict check, approval decision, and case creation.
- Make conflict checks asynchronous by default to avoid HTTP gateway timeouts during recursive entity penetration or expensive matching.
- Preserve approval decision context through immutable snapshots.
- Prevent client profile overwrites and duplicate approval submissions.
- Support positive handoffs when assistants complete missing data.

## Non-Goals

- Do not implement Batch 02-04 capabilities such as ethical wall control, document library, finance, or analytics in this change.
- Do not replace the existing approval system or conflict detection engine.
- Do not introduce a message queue requirement; async tasks may be implemented in-process initially and persisted in the database.
- Do not require SSE or WebSocket in the first implementation; polling is the default.

## Technical Decisions

### Async Conflict Tasks

Conflict checks SHALL be modeled as persisted tasks with statuses:

- `queued`
- `running`
- `completed`
- `failed`

Task creation SHALL return task ID, initial status, and a recommended polling interval. Frontend polling is the default interaction, with SSE left as a future enhancement.

### Approval Snapshots

When a conflict result is submitted for approval, the system SHALL persist an immutable snapshot of:

- intake fields
- client and counterparty key fields
- conflict result JSON
- risk score and risk level
- evidence summary
- optional PDF/report reference

Approval details SHALL display the snapshot by default. If source data changes later, the UI SHALL indicate that source data changed after submission without mutating the snapshot.

### Client Optimistic Locking

Client update requests SHALL include a `version` or equivalent concurrency token. If the stored version differs from the submitted version, the system SHALL reject the update with a conflict response and require the user to refresh.

### Idempotency

Approval submission and conflict task creation SHALL use an idempotency token to prevent duplicate actions caused by double clicks, retries, or network timeouts.

### Waiver Subflow

Waiver-eligible conflict findings SHALL enter a waiver request subflow instead of behaving as a simple approval button. The subflow SHALL capture waiver reason, attachments, designated approver, and status.

### Positive Handoff

When an assistant completes required client or intake fields, the system SHALL create a handoff notification for the responsible attorney or compliance reviewer. The notification SHALL be visible in dashboard todos and inbox.

## Data Flow

1. User starts from dashboard or client profile.
2. Client master data is reviewed and updated with version protection.
3. Intake workbench saves a draft or creates/updates enhanced case data.
4. "Save and check conflict" creates an async conflict task.
5. Dashboard and result page poll task status.
6. Completed conflict result is reviewed and submitted to approval.
7. Approval submission persists immutable snapshot and idempotency record.
8. Approval decision either creates the case, rejects, requests supplements, or routes to waiver review.

## Failure Handling

- Missing required intake fields SHALL block conflict task creation with explicit missing-field messages.
- Conflict task failures SHALL expose reason, retry entry, and latest update timestamp.
- Client version conflicts SHALL prevent overwrites and instruct the user to refresh.
- Duplicate approval submissions SHALL return the existing approval reference instead of creating another record.
- Source data changes after approval submission SHALL not alter approval snapshot content.
