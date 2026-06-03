## MODIFIED Requirements

### Requirement: Intake Workbench Required Data
The intake workbench SHALL collect and validate the minimum data required for pre-engagement conflict checking, including client, counterparty, matter title, matter type, business domain when required by the form, owner or responsible lawyer, fee/team fields needed by the current workflow, summary, and material checklist. The system SHALL allow saving incomplete intake drafts without conflict checking, but SHALL NOT create a formal conflict-check side effect until required conflict-check fields are complete.

#### Scenario: Save incomplete intake draft without conflict check
- **WHEN** a lawyer has not completed all required conflict-check fields
- **AND** the lawyer selects a draft-only save action
- **THEN** the system SHALL allow the incomplete draft to be saved
- **AND** the system SHALL keep conflict-check status as not detected
- **AND** the system SHALL NOT create a conflict task

#### Scenario: Block conflict check before side effects when required fields are missing
- **WHEN** a lawyer selects a client and responsible lawyer
- **AND** matter title or counterparty is missing
- **AND** the lawyer clicks "运行利益冲突检查" or "保存并进行利益冲突检查"
- **THEN** the system SHALL show validation feedback identifying the missing fields
- **AND** the system SHALL keep the lawyer in the intake workflow
- **AND** the system SHALL NOT create a new intake draft, intake number, conflict task, or success toast for the blocked action
- **AND** the conflict-check status SHALL remain not detected

#### Scenario: Run conflict check only after required fields are complete
- **WHEN** client, counterparty, matter title, responsible lawyer, and form-required matter classification fields are complete
- **AND** the lawyer clicks "运行利益冲突检查"
- **THEN** the system SHALL save or update the intake draft
- **AND** the system SHALL create or reuse the conflict-check task
- **AND** the system SHALL keep the lawyer inside the intake workflow context
- **AND** the system SHALL expose a clear next action to continue to team and fee setup

#### Scenario: Prevent confusing mixed success and validation messages
- **WHEN** a conflict-check action is blocked by preflight validation
- **THEN** the system SHALL NOT display messages that imply successful draft creation or successful conflict-check execution
- **AND** the visible message SHALL describe the required correction before retrying
