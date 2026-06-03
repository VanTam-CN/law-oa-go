## ADDED Requirements

### Requirement: Approval Detail Action Name Uniqueness
Approval detail pages SHALL provide unique visible text or accessible names for repeated action buttons within the same page so each action can be identified by purpose and location.

#### Scenario: Header and decision overflow actions exist
- **WHEN** an approval detail page contains an overflow action in the header and another overflow action in the decision area
- **THEN** the header action SHALL be identifiable as an approval-level or header-level action
- **AND** the decision-area action SHALL be identifiable as a decision or handling action
- **AND** locating either action by role and accessible name SHALL resolve to exactly one element

#### Scenario: Lawyer uses assistive technology
- **WHEN** a lawyer navigates the approval detail page by button names
- **THEN** the system SHALL expose names that describe the difference between header-level operations and decision-level operations

### Requirement: Related Case Action State
The approval detail page SHALL make the "view related case" action deterministic by enabling it only when a routeable related case exists and explaining the disabled state when no related case is available.

#### Scenario: Related case exists
- **WHEN** an approval has a valid related case identifier and the lawyer clicks "查看关联案件"
- **THEN** the system SHALL navigate to the related case detail page
- **AND** the destination SHALL identify the related case

#### Scenario: Related case does not exist
- **WHEN** an approval does not yet have a routeable related case
- **THEN** the "查看关联案件" action SHALL be disabled or replaced with a non-actionable status
- **AND** the page SHALL display a visible tooltip, helper text, or status message explaining why the case cannot be opened

#### Scenario: Related case is unavailable due to permission
- **WHEN** a related case exists but the lawyer lacks permission to view it
- **THEN** the system SHALL show a permission-specific explanation
- **AND** the system SHALL NOT present the state as a broken or inert button

## MODIFIED Requirements

### Requirement: Conflict-Linked Approval Snapshot
Approval submissions created from conflict results SHALL persist immutable snapshots of the intake data and conflict result used for the decision. Approval detail pages SHALL display the approval-time snapshot, use one normalized approval access model for current approver labels and action permissions, and expose deterministic action states for linked resources.

#### Scenario: Submit approval snapshot
- **WHEN** a user submits a conflict result for approval
- **THEN** the system SHALL persist an immutable snapshot containing client, counterparty, intake fields, conflict JSON result, risk score, evidence summary, and optional report reference

#### Scenario: Current approver identity is consistent
- **WHEN** a lawyer opens an approval detail page
- **THEN** the header current approver, current process node, and available decision buttons SHALL be derived from the same normalized approval access model

#### Scenario: Source data changes after submission
- **WHEN** source client, matter, or conflict data changes after approval submission
- **THEN** the approval detail SHALL continue showing the approval-time snapshot
- **AND** the page SHALL clearly indicate that source data changed after submission
