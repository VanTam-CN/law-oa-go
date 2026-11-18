package models

import "time"

// ApprovalSnapshot maps the existing approval_snapshots table for the N3
// intake_submission contract (plan §6.3). Older rows keep working through
// raw SQL readers; this struct only covers new typed rows.
type ApprovalSnapshot struct {
	ID                string    `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	ApprovalRequestID string    `gorm:"column:approval_request_id;type:varchar(36)" json:"approval_request_id"`
	SnapshotType      string    `gorm:"column:snapshot_type;type:varchar(80)" json:"snapshot_type"`
	SnapshotData      string    `gorm:"column:snapshot_data;type:jsonb" json:"snapshot_data"`
	SourceVersion     int       `gorm:"column:source_version" json:"source_version"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ApprovalSnapshot) TableName() string { return "approval_snapshots" }

// Snapshot types defined by the intake-collaboration MVP.
const (
	ApprovalSnapshotTypeIntakeSubmission = "intake_submission"
)

// Structured-supplement contract error codes (plan §8).
const (
	ErrCodeSupplementItemInvalid      = "SUPPLEMENT_ITEM_INVALID"
	ErrCodeApprovalRoundStale         = "APPROVAL_ROUND_STALE"
	ErrCodeApprovalSnapshotUnresolved = "APPROVAL_SNAPSHOT_UNRESOLVED"
)
