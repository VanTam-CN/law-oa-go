## ADDED Requirements

### Requirement: Authorized actionable workflow items
The system SHALL use a shared server-authorized workflow action descriptor in the dashboard and inbox. Business object identifiers SHALL be represented without lossy numeric conversion, and route targets MUST be validated again by the destination service.

#### Scenario: Open a string-identified approval
- **WHEN** a user opens an approval task whose source object has a string identifier
- **THEN** the action SHALL address that exact authorized approval without truncating the identifier or guessing from the title

#### Scenario: Unknown historical reference
- **WHEN** a legacy item's source cannot be verified
- **THEN** the system SHALL display an unresolved historical reference and SHALL NOT offer a guessed business action

#### Scenario: Permission changes after a task was listed
- **WHEN** an actor loses object access or assignment before acting
- **THEN** the destination service SHALL refuse the operation and the refreshed item SHALL disclose no unauthorized case, identity, or supplement details

#### Scenario: Forge workflow metadata through a personal task endpoint
- **WHEN** a client attempts to change a workflow task into a manual reminder or supplies a privileged object reference
- **THEN** the server SHALL reject the protected-field mutation and SHALL NOT grant access to the referenced business object

### Requirement: Transactional idempotent workflow handoff
For the supported intake workflow, state transitions and associated task creation or closure MUST commit atomically. Task deduplication SHALL distinguish resource, round or node, recipient, and purpose.

#### Scenario: Return then resubmit an approval
- **WHEN** an authorized decision returns an approval and the applicant later resubmits it
- **THEN** each successful transition SHALL close its previous action task and create exactly one appropriate next task, with retries producing no duplicate handoff

#### Scenario: Task persistence fails
- **WHEN** a required next-task write fails during a business transition
- **THEN** the business state and task changes SHALL roll back together and the system SHALL allow a safe retry after recovery

### Requirement: Business completion cannot be replaced by reminder completion
The system SHALL distinguish workflow-controlled tasks, manual personal reminders, and unresolved historical notifications. Completing, deleting, or snoozing a reminder MUST NOT falsely complete or hide a required business action.

#### Scenario: Directly complete or delete a pending workflow task
- **WHEN** a user calls the completion or deletion endpoint for an unperformed workflow action
- **THEN** the server SHALL reject the request and keep the actual pending action visible

#### Scenario: Snooze a business reminder
- **WHEN** a user snoozes its reminder
- **THEN** the underlying deadline, responsibility, and business state SHALL remain unchanged and the action SHALL remain discoverable

#### Scenario: Approved but case creation failed
- **WHEN** an approval succeeds but its case creation fails
- **THEN** the system SHALL distinguish the finished approval from failed case creation and SHALL NOT label the matter as successfully created

### Requirement: Consistent personal work queues
Dashboard and inbox counts, filters, and pages SHALL use the same authorized classification of actionable, waiting, completed, and historical items.

#### Scenario: Paginate a filtered work queue
- **WHEN** the user filters or loads another page
- **THEN** counts and items SHALL reflect the same permission-filtered source set rather than counts computed from the first page alone
