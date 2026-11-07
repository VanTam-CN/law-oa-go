## ADDED Requirements

### Requirement: Structured supplement requests and responses
The system SHALL record versioned supplement requests on the actual approval decision record, with stable item identifiers and allowlisted field references. Replies SHALL bind to the latest return record and item, preserve existing feedback, and MUST NOT themselves establish successful identity or conflict review.

#### Scenario: Return two corrections
- **WHEN** an authorized approver requests two supported corrections
- **THEN** the applicant SHALL receive one actionable supplement task, see both persisted items, navigate to their correct fields, and save individual replies

#### Scenario: Legacy text-only return
- **WHEN** a historical return contains only textual feedback
- **THEN** the page SHALL preserve and label that feedback without inventing historical structured requirements

#### Scenario: Forged or unsupported item
- **WHEN** a client supplies an unknown field, another round's item, or an unsupported material-completion claim
- **THEN** the server SHALL reject it without changing the current approval or claiming the requirement is fulfilled

### Requirement: Immutable authoritative submission rounds
The system SHALL append immutable submission snapshots and designate one authoritative current snapshot for each upgraded approval. Approval display, decisions, integration status, case creation, retries, and party materialization MUST use the same current submission source while retaining earlier snapshots.

#### Scenario: Revised facts are submitted
- **WHEN** the applicant validly resubmits revised intake facts
- **THEN** a new immutable snapshot SHALL bind the exact intake facts, formal check, review evidence, replies, actor, and round, while all prior snapshots remain unchanged

#### Scenario: Conflicting historical snapshot sources
- **WHEN** the service cannot establish an authoritative legacy submission
- **THEN** it SHALL block that item's upgraded resubmission with an explicit reason rather than reconstructing historical approval from the latest draft

#### Scenario: Attempt to overwrite an authoritative binding through a generic endpoint
- **WHEN** a client submits an active snapshot pointer or reviewed check binding through a generic update or submission API
- **THEN** the server SHALL reject unauthorized binding changes and require the same controlled submission path used by the collaboration flow

### Requirement: Fact-aware gated resubmission
Resubmission SHALL enforce existing intake revisions, object authorization, independent review, and conflict validity. Material fact changes MUST invalidate the prior result; explanation-only replies MAY reuse evidence only after the server verifies unchanged facts and current validity.

#### Scenario: Change a party's identity or role
- **WHEN** the applicant changes identity, role, or another conflict-relevant fact and attempts resubmission
- **THEN** the old check SHALL be rejected until new formal checking and the required independent review are complete

#### Scenario: Add explanation without changing facts
- **WHEN** only supplement replies change
- **THEN** reuse SHALL depend on server-side fact binding, validity, and review checks, not an applicant checkbox or a browser fingerprint

### Requirement: Atomic unique resubmission and stale decision rejection
Resubmission MUST atomically validate the return round, claim the existing unique intake approval binding, append and activate the new snapshot, reassign through the existing workflow, and hand off tasks. Replays SHALL be idempotent, competing claims MUST NOT revive a loser approval, and stale decisions MUST NOT affect a later round.

#### Scenario: Old page decides after a new round
- **WHEN** an approver submits a decision for a superseded snapshot or node
- **THEN** the server SHALL reject the stale decision and leave the new round unchanged

#### Scenario: Resubmission competes with a new approval
- **WHEN** an old approval is resubmitted while a new approval attempts to claim the same intake
- **THEN** at most one valid approval SHALL own the binding, and the losing transaction SHALL activate no snapshot, revived status, or next task

#### Scenario: Failure during snapshot or party persistence
- **WHEN** a required snapshot or case-party write fails
- **THEN** the system SHALL preserve transactional consistency and allow recovery without a false completed state or a second case
