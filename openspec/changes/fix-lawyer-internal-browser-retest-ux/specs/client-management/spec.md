## ADDED Requirements

### Requirement: Client Profile Quick Action Feedback
Client profile quick actions SHALL produce deterministic visible feedback or navigation. A quick action SHALL NOT be implemented as a silent no-op.

#### Scenario: Start new matter from client profile
- **WHEN** a lawyer clicks "发起新案件" from a client profile
- **THEN** the system SHALL navigate to the intake workbench
- **AND** the intake workbench SHALL be ready to associate the selected client when the route or state supports preselection

#### Scenario: Start conflict check from client profile
- **WHEN** a lawyer clicks "发起冲突检查" from a client profile
- **THEN** the system SHALL navigate to the conflict detection or review workflow
- **AND** the destination SHALL make it clear which client context is being checked when context is available

#### Scenario: Add contact from client profile
- **WHEN** a lawyer clicks "新增联系人"
- **THEN** the system SHALL open a modal, drawer, or inline form for adding a contact
- **OR** the system SHALL show an explicit unavailable message if contact creation is not implemented
- **AND** the click SHALL NOT leave the page visually unchanged without explanation

#### Scenario: Upload attachment from client profile
- **WHEN** a lawyer clicks "上传附件"
- **THEN** the system SHALL open an upload control, modal, or drawer
- **OR** the system SHALL show an explicit unavailable message if attachment upload is not implemented
- **AND** the click SHALL NOT leave the page visually unchanged without explanation

#### Scenario: Export client profile
- **WHEN** a lawyer clicks "导出客户档案"
- **THEN** the system SHALL start the export/download workflow and show success or progress feedback
- **OR** the system SHALL show an explicit error or unavailable message
- **AND** the system SHALL NOT silently ignore the click

### Requirement: Client Profile Quick Action Testability
Client profile quick actions SHALL be identifiable by role and accessible name, and their results SHALL be testable through visible UI state, route changes, or download/error feedback.

#### Scenario: Automated test selects a quick action
- **WHEN** an E2E test locates a client profile quick action by role and accessible name
- **THEN** the locator SHALL resolve to exactly one intended action when used within the client profile action area
- **AND** the test SHALL be able to assert the action outcome without relying on hidden implementation details
