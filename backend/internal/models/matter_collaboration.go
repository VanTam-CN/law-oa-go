package models

import (
	"time"

	"gorm.io/gorm"
)

// Matter collaboration vertical (2026-09-30 five-item fill, B/C slice).
// These models extend the existing case_materials ledger with ownership,
// upload state, and per-file attachments, plus a first-class case task
// whose completion is transactionally mirrored into inbox_items.

const (
	// MaterialStatusMissing marks a required ledger row with no file yet.
	MaterialStatusMissing = "missing"
	// MaterialStatusUploaded marks a row whose attachments landed in
	// storage. Only the upload transaction writes this value; clients can
	// never POST it directly (anti-forgery requirement 2026-09-30).
	MaterialStatusUploaded = "uploaded"
	// MaterialStatusNotRequired marks a row the lawyer marked unnecessary.
	MaterialStatusNotRequired = "not_required"
)

// CaseMaterial is the existing case_materials ledger row (000028 schema)
// extended by migration 000084 with owner assignment and upload metadata.
type CaseMaterial struct {
	ID           uint      `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CaseID       *uint     `json:"case_id,omitempty" gorm:"column:case_id;index"`
	IntakeID     *string   `json:"intake_id,omitempty" gorm:"column:intake_id;type:uuid;index"`
	Name         string    `json:"name" gorm:"column:name;size:255;not null"`
	MaterialType string    `json:"material_type" gorm:"column:material_type;size:80"`
	Status       string    `json:"status" gorm:"column:status;size:40;not null;default:'missing'"`
	Required     bool      `json:"required" gorm:"column:required;not null;default:true"`
	StorageURL   string    `json:"storage_url,omitempty" gorm:"column:storage_url;type:text"`
	// OwnerID is the responsible person assigned inside the case team.
	OwnerID *uint `json:"owner_id,omitempty" gorm:"column:owner_id;index"`
	// UploadedBy / UploadedAt are set only by the upload transaction.
	UploadedBy *uint      `json:"uploaded_by,omitempty" gorm:"column:uploaded_by"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty" gorm:"column:uploaded_at"`
	Metadata   string     `json:"metadata" gorm:"column:metadata;type:jsonb;not null;default:'{}'"`
}

func (CaseMaterial) TableName() string { return "case_materials" }

// MaterialFile is one physical attachment of a case_material row. Files
// live in their own safe storage tree; they never masquerade as documents
// rows and never bind numeric documents IDs to intake UUIDs.
type MaterialFile struct {
	ID         uint           `json:"id" gorm:"primarykey"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	MaterialID uint           `json:"material_id" gorm:"column:material_id;not null;index"`
	// SHA256 hex digest of the stored bytes; the resubmit snapshot stores
	// this digest so later tampering is detectable.
	ContentSHA256 string `json:"content_sha256" gorm:"column:content_sha256;size:64;not null"`
	FileName      string `json:"file_name" gorm:"column:file_name;size:255;not null"`
	SizeBytes     int64  `json:"size_bytes" gorm:"column:size_bytes;not null"`
	ContentType   string `json:"content_type" gorm:"column:content_type;size:120"`
	StoragePath   string `json:"-" gorm:"column:storage_path;size:500;not null"`
	UploadedBy    uint   `json:"uploaded_by" gorm:"column:uploaded_by;not null;index"`
}

func (MaterialFile) TableName() string { return "material_files" }

const (
	// CaseTaskStatusOpen is the only actionable state.
	CaseTaskStatusOpen = "open"
	// CaseTaskStatusCompleted is terminal; completion closes the mirrored
	// todo in the same transaction.
	CaseTaskStatusCompleted = "completed"
	// CaseTaskStatusCancelled is terminal without completion.
	CaseTaskStatusCancelled = "cancelled"
)

// CaseTask is the general-purpose case matter task (user item 5). DueDate
// is the lawyer working deadline; it never claims to be a statutory legal
// date and never enters the deadline calculator chain.
type CaseTask struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CaseID      uint      `json:"case_id" gorm:"column:case_id;not null;index"`
	Title       string    `json:"title" gorm:"column:title;size:255;not null"`
	Detail      string    `json:"detail,omitempty" gorm:"column:detail;type:text"`
	AssigneeID  uint      `json:"assignee_id" gorm:"column:assignee_id;not null;index"`
	CreatedBy   uint      `json:"created_by" gorm:"column:created_by;not null"`
	// DueDate is an explicit lawyer working deadline, confirmed by the
	// assigning lawyer; it is NOT a computed statutory deadline.
	DueDate     *time.Time `json:"due_date,omitempty" gorm:"column:due_date"`
	Status      string     `json:"status" gorm:"column:status;size:20;not null;default:'open';index"`
	CompletedBy *uint      `json:"completed_by,omitempty" gorm:"column:completed_by"`
	CompletedAt *time.Time `json:"completed_at,omitempty" gorm:"column:completed_at"`
	// IdempotencyKey is the client-supplied dedup key; when omitted the
	// server mints a unique per-request key (auto-<uuid>), so empty strings
	// never exist and keyless retries intentionally create new tasks.
	IdempotencyKey string `json:"idempotency_key,omitempty" gorm:"column:idempotency_key;size:120"`
	// IdempotencyFingerprint is the SHA-256 over (case, title, detail,
	// assignee, due date) at creation time. A same-key retry with any
	// differing field is a conflict, never a silent reuse.
	IdempotencyFingerprint string `gorm:"column:idempotency_fingerprint;size:64;not null;default:''"`
	// CompletionNote is the free-text completion record persisted by the
	// same transaction that flips the task status. The HTTP response's
	// completion_note_recorded flag reflects this column, never the raw
	// request body, so a success can never claim a note that was dropped.
	CompletionNote string `json:"completion_note,omitempty" gorm:"column:completion_note;type:text"`
	// Version is the optimistic-concurrency token for adjust/cancel and
	// complete. Every write CASes on it and bumps it by one in the same
	// UPDATE; gorm tag kept in sync with migration 000085 so dev-mode
	// AutoMigrate and SQL migrations produce the same column.
	Version uint `json:"version" gorm:"column:version;not null;default:1"`
}

func (CaseTask) TableName() string { return "case_tasks" }
