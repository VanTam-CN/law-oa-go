## ADDED Requirements

### Requirement: Context preserving inline client creation
The system SHALL let an authorized intake writer create a client in the intake page through the existing authorized client service, select the returned client ID, and preserve all current intake inputs. It MUST preserve identity protection and MUST NOT merge parties solely by name.

#### Scenario: Create or cancel without losing intake inputs
- **WHEN** the user creates a client or cancels the inline dialog
- **THEN** the intake fields and current step SHALL remain intact, and a successful creation SHALL select the returned client without requiring a first-page list refresh

#### Scenario: Retry an uncertain client creation
- **WHEN** an actor retries the same client creation with the same idempotency key and normalized payload
- **THEN** the server SHALL return the same persisted client, while a changed payload under that key SHALL be rejected with a conflict and no duplicate write

#### Scenario: Unauthorized or unavailable identity protection
- **WHEN** the actor lacks client creation permission or required identity protection is unavailable
- **THEN** the server SHALL reject creation without relaxing permission or storing unprotected identity, and the page SHALL retain recoverable intake input

### Requirement: Discoverable authoritative draft continuation
The system SHALL expose authorized server drafts from the workbench, restore the exact chosen intake, and distinguish unsaved input from successful server persistence. Existing revision protection and account isolation MUST remain effective.

#### Scenario: Resume after signing in again
- **WHEN** the same authorized user returns in a new session and selects a saved draft
- **THEN** the system SHALL restore its permitted facts and roles from the server without relying on a local draft or exposing protected identity plaintext

#### Scenario: Save a stale page
- **WHEN** a page saves an outdated intake revision
- **THEN** the server SHALL reject the update, and the page SHALL retain its unsaved input until the user explicitly chooses a recovery action

#### Scenario: Search beyond the initial client page
- **WHEN** an authorized client is outside the initially loaded client list
- **THEN** server-side search or pagination SHALL allow its selection without exposing another user's restricted clients

### Requirement: Trial documentation matches the delivered candidate
The operation manual SHALL describe the delivered candidate's actual inline client creation, draft recovery, workflow actions, and supplement flow, and SHALL accurately label unavailable material features.

#### Scenario: Follow a documented operation
- **WHEN** a trial user follows an updated manual step
- **THEN** the corresponding control and result SHALL exist in that candidate, and any associated screenshot SHALL show the stated page with fictional data
