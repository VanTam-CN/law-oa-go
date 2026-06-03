## ADDED Requirements

### Requirement: Dashboard Action Observability
Every clickable dashboard action exposed to a lawyer SHALL produce an observable outcome: route navigation, visible filter state change, list expansion, refreshed timestamp/data, modal/drawer display, download/export feedback, or explicit unavailable/error feedback.

#### Scenario: View all todo items
- **WHEN** a lawyer clicks the dashboard action for viewing all todo items
- **THEN** the system SHALL navigate to the inbox or task-center route, or visibly expand the complete todo list on the dashboard
- **AND** the outcome SHALL be observable through URL, heading, selected state, expanded content, or a visible message

#### Scenario: View all conflict review items
- **WHEN** a lawyer clicks the dashboard action for viewing all conflict review items
- **THEN** the system SHALL navigate to the conflict review or conflict detection route
- **AND** the destination SHALL show conflict-related content instead of leaving the user on an unchanged dashboard

#### Scenario: Button action unavailable
- **WHEN** a dashboard button cannot perform its advertised action because the feature is unavailable
- **THEN** the system SHALL show an explicit unavailable message
- **AND** the system SHALL NOT silently ignore the click

### Requirement: Dashboard Button Name Uniqueness
Dashboard buttons that appear in the same page context SHALL have unique visible text or unique accessible names so lawyers, assistive technology, and automated tests can distinguish their purpose.

#### Scenario: Multiple view-all buttons exist
- **WHEN** the dashboard contains more than one "view all" style action
- **THEN** each action SHALL have a unique accessible name such as "查看全部待办" and "查看全部冲突任务"
- **AND** E2E tests SHALL target the exact intended accessible name

#### Scenario: Role-based locator is used
- **WHEN** an automated test locates a dashboard button by role and accessible name
- **THEN** the locator SHALL resolve to exactly one actionable element
