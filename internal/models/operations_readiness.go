package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	OperationsEvidenceScopeQA              = "qa"
	OperationsEvidenceScopeControlledPilot = "controlled_pilot"
	OperationsEvidenceResultPassed         = "passed"
)

// OperationsReadinessEvidence is one append-only controlled-environment
// registration. Multiple rows per control and scope preserve revalidation
// history; readiness reads only the latest row for each control. A row records
// where independently retained evidence can be reviewed, but does not itself
// prove a production release is safe.
type OperationsReadinessEvidence struct {
	ID                 string    `json:"id" gorm:"primaryKey;column:id;size:36"`
	Control            string    `json:"control" gorm:"column:control;size:40;not null;index:idx_operations_evidence_control_scope_time,priority:1"`
	Scope              string    `json:"scope" gorm:"column:scope;size:30;not null;index:idx_operations_evidence_control_scope_time,priority:2"`
	Result             string    `json:"result" gorm:"column:result;size:20;not null"`
	EvidenceReference  string    `json:"evidence_reference" gorm:"column:evidence_reference;type:text;not null"`
	ReviewedBy         uint      `json:"reviewed_by" gorm:"column:reviewed_by;not null;index"`
	ReviewedAt         time.Time `json:"reviewed_at" gorm:"column:reviewed_at;not null;index:idx_operations_evidence_control_scope_time,priority:3"`
	Notes              string    `json:"notes,omitempty" gorm:"column:notes;type:text"`
	PreviousEvidenceID string    `json:"previous_evidence_id" gorm:"column:previous_evidence_id;size:36;not null;default:''"`
	IntegrityHash      string    `json:"integrity_hash" gorm:"column:integrity_hash;size:64;not null;default:''"`
	CreatedAt          time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (OperationsReadinessEvidence) TableName() string {
	return "operations_readiness_evidence"
}

func (OperationsReadinessEvidence) BeforeUpdate(*gorm.DB) error {
	return errors.New("operations readiness evidence is append-only")
}

func (OperationsReadinessEvidence) BeforeDelete(*gorm.DB) error {
	return errors.New("operations readiness evidence is append-only")
}
