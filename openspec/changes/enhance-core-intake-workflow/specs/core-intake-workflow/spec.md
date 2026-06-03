## ADDED Requirements

### Requirement: Core Intake Workflow
The system SHALL provide a continuous intake workflow from dashboard priority review through client verification, case intake, conflict checking, approval decision, and approved case creation.

#### Scenario: Complete intake flow
- **WHEN** a responsible attorney starts a new intake from the dashboard or client profile
- **THEN** the system SHALL allow the user to verify client data, save intake details, run conflict checking, submit the result for approval, and proceed to case creation after approval

#### Scenario: Preserve workflow navigation
- **WHEN** a user moves between dashboard, client profile, intake, conflict result, and approval detail
- **THEN** the system SHALL preserve linked workflow context without requiring the user to manually reconstruct the matter from unrelated navigation entries

### Requirement: Dashboard Intake Prioritization
The dashboard SHALL prioritize intake-related work including pending conflict reviews, pending approvals, background conflict checks, overdue tasks, high-risk intake items, and positive handoffs.

#### Scenario: View daily risk workload
- **WHEN** a managing partner opens the dashboard
- **THEN** the system SHALL display intake and risk items with direct navigation to the relevant conflict result, approval detail, client profile, or intake workbench

### Requirement: Client Master Profile Completeness
The client master profile SHALL provide the data needed for intake and conflict checking, including completeness score, missing fields, aliases/history, related parties, historical matters, and a version value.

#### Scenario: Block conflict check for incomplete client data
- **WHEN** required client data for conflict checking is incomplete
- **THEN** the system SHALL show missing fields and prevent formal conflict check submission until the required data is completed

#### Scenario: Display related parties
- **WHEN** a user reviews a client profile before intake
- **THEN** the system SHALL display related-party summaries that are included in the default conflict check scope

### Requirement: Client Concurrent Edit Protection
The client profile update flow SHALL prevent stale writes using an optimistic concurrency token such as a version field.

#### Scenario: Reject stale update
- **WHEN** a user submits client changes with an outdated version
- **THEN** the system SHALL reject the update with a conflict response and require the user to refresh before retrying

### Requirement: Intake Workbench Required Data
The intake workbench SHALL collect the minimum data required for pre-engagement conflict checking, including client, counterparty, matter type, business domain, owner/team, fee model, summary, and material checklist.

#### Scenario: Save intake draft
- **WHEN** required conflict-check fields are incomplete
- **THEN** the system SHALL allow the user to save a draft without creating a conflict task

#### Scenario: Create conflict task from intake
- **WHEN** client, counterparty, and matter type are present and the user selects save-and-check
- **THEN** the system SHALL save the intake and create an asynchronous conflict check task

### Requirement: Async Conflict Check Task
Conflict checking SHALL be performed as an asynchronous persisted task with status visibility and result retrieval.

#### Scenario: Start async conflict check
- **WHEN** a user initiates a conflict check
- **THEN** the system SHALL return a task ID, initial status, and recommended polling interval without requiring the request to wait for final analysis

#### Scenario: Poll conflict task
- **WHEN** the frontend polls a conflict task
- **THEN** the system SHALL return one of queued, running, completed, or failed with the latest update timestamp

#### Scenario: Background conflict task visibility
- **WHEN** a user leaves the conflict result page while a task is still running
- **THEN** the system SHALL keep the task visible from the dashboard or related intake workflow

### Requirement: Conflict Result Decision Context
Conflict result pages SHALL display risk level, score, matched subject, match type, scope, confidence, penetration layer, evidence summary, source record, waiver eligibility, and recommendation actions.

#### Scenario: Review high-risk conflict
- **WHEN** a compliance reviewer opens a completed high-risk conflict result
- **THEN** the system SHALL display the evidence chain and recommendation actions including approval, waiver assessment, team adjustment, or matter exit

### Requirement: Conflict-Linked Approval Snapshot
Approval submissions created from conflict results SHALL persist immutable snapshots of the intake data and conflict result used for the decision.

#### Scenario: Submit approval snapshot
- **WHEN** a user submits a conflict result for approval
- **THEN** the system SHALL persist an immutable snapshot containing client, counterparty, intake fields, conflict JSON result, risk score, evidence summary, and optional report reference

#### Scenario: Source data changes after submission
- **WHEN** source client, matter, or conflict data changes after approval submission
- **THEN** the approval detail SHALL continue showing the approval-time snapshot and clearly indicate that source data changed after submission

### Requirement: Waiver Review Subflow
Waiver-eligible conflict findings SHALL support a waiver review subflow with reason, supporting attachments, approver assignment, and explicit outcome status.

#### Scenario: Request waiver review
- **WHEN** a user selects waiver assessment for a waiver-eligible finding
- **THEN** the system SHALL require waiver reason, supporting materials, and designated approver before creating the waiver request

#### Scenario: Complete waiver review
- **WHEN** a waiver approver decides on the request
- **THEN** the system SHALL transition the waiver request to waiver approved or waiver rejected and reflect that result in the related approval context

### Requirement: Positive Handoff Notification
The system SHALL support positive handoff notifications when assistants complete required client or intake data.

#### Scenario: Assistant hands off completed data
- **WHEN** an assistant marks required client or intake data as completed and hands it off
- **THEN** the system SHALL create a notification or inbox item for the responsible attorney or compliance reviewer

#### Scenario: Handoff appears on dashboard
- **WHEN** the recipient opens the dashboard or inbox
- **THEN** the system SHALL show the handoff item with navigation to the relevant client profile or intake workbench

### Requirement: Idempotent Intake Actions
Conflict task creation and approval submission SHALL be protected from duplicate execution caused by repeated clicks or network retries.

#### Scenario: Duplicate conflict task creation
- **WHEN** the same idempotency token is submitted again for conflict task creation
- **THEN** the system SHALL return the existing task reference instead of creating a duplicate task

#### Scenario: Duplicate approval submission
- **WHEN** the same idempotency token is submitted again for approval creation
- **THEN** the system SHALL return the existing approval reference instead of creating a duplicate approval
