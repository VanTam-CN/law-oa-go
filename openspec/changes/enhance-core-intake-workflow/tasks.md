## 1. Data Model and API Contract

- [ ] 1.1 Confirm and expose client-domain APIs required by the client master profile: list, detail, create, update, delete, stats, completeness, related parties, and version field.
- [ ] 1.2 Add database support for client optimistic locking, conflict task status/result storage, approval immutable snapshots, waiver requests, and idempotency tokens.
- [ ] 1.3 Define API response contracts for client completeness, conflict task status, conflict task result, approval snapshot, waiver request, and positive handoff notification.
- [ ] 1.4 Update API documentation and generated types consumed by frontend services.

## 2. Client Master Profile

- [x] 2.1 Extend client detail response to include completeness score, missing required fields, related-party summary, historical matter summary, and version.
- [x] 2.2 Implement optimistic concurrency checks for client updates and return a conflict response when the submitted version is stale.
- [ ] 2.3 Update the client profile UI to show completeness, missing fields, aliases/history, related parties, linked matters, and intake actions.
- [ ] 2.4 Add positive handoff action from client profile after assistant completion.

## 3. Case Intake Workbench

- [ ] 3.1 Build or refactor the intake workbench to collect client, counterparties, matter type, business domain, owner/team, fee model, summary, and material checklist.
- [ ] 3.2 Enforce required fields before conflict check: client, counterparty, and matter type.
- [ ] 3.3 Save draft intake records without running conflict checks.
- [ ] 3.4 Implement "save and check conflict" using enhanced case creation/update followed by async conflict task creation.
- [ ] 3.5 Surface conflict task progress and allow users to leave the page while the task continues.

## 4. Async Conflict Check

- [x] 4.1 Implement async conflict task creation that returns task ID, initial status, and recommended polling interval.
- [x] 4.2 Implement task status and task result endpoints for queued, running, completed, and failed states.
- [ ] 4.3 Persist structured result data with risk level, score, match category, matched subject, confidence, penetration layer, evidence summary, source record, and recommendations.
- [ ] 4.4 Update conflict result UI to display risk summary, filters, evidence chain, three-layer penetration, waiver eligibility, and action recommendations.
- [ ] 4.5 Add retry and failure messaging for failed conflict tasks.

## 5. Conflict-Linked Approval and Waiver Flow

- [ ] 5.1 Create approval submissions from conflict results with idempotency protection.
- [x] 5.2 Persist immutable approval snapshots containing intake data, client/counterparty fields, conflict JSON result, risk score, evidence summary, and optional report reference.
- [ ] 5.3 Update approval details to display the approval-time snapshot by default and clearly mark when source data has changed after submission.
- [x] 5.4 Implement waiver request creation for waiver-eligible conflicts with reason, supporting attachments, approver assignment, and status transitions.
- [ ] 5.5 Support approval decisions: approve, reject, request supplement, resubmit, cancel, and approved case creation.

## 6. Dashboard and Positive Handoff

- [ ] 6.1 Update dashboard metrics and queues to prioritize pending conflict reviews, pending approvals, background conflict checks, overdue tasks, and high-risk intake items.
- [ ] 6.2 Add direct navigation from dashboard risk items to conflict results, approval details, client profile, or intake workbench.
- [x] 6.3 Implement positive handoff notifications from assistant to attorney/compliance after client or intake data is completed.
- [x] 6.4 Ensure handoff notifications appear in dashboard todos and inbox.

## 7. Testing and Acceptance

- [ ] 7.1 Add backend unit tests for client version conflicts, conflict task state transitions, approval snapshot persistence, waiver status transitions, and idempotency handling.
- [ ] 7.2 Add integration tests for the full intake flow: client profile -> intake draft -> async conflict check -> conflict result -> approval -> case creation.
- [ ] 7.3 Add frontend tests for dashboard risk navigation, required-field blocking, async conflict polling, approval snapshot display, and waiver request creation.
- [ ] 7.4 Validate three business scenarios from the PRD: new client intake, high-risk conflict hit, and partner daily risk workload review.
- [ ] 7.5 Update the Batch 01 PRD links or release notes after implementation.
