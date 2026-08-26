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

// OperationsReadinessEvidence is a controlled-environment registration record.
// It records where independently retained evidence can be reviewed, but does
// not itself prove a production release is safe.
type OperationsReadinessEvidence struct {
	ID                string    `json:"id" gorm:"primaryKey;column:id;size:36"`
	Control           string    `json:"control" gorm:"column:control;size:40;not null;uniqueIndex:uq_operations_evidence_control_scope"`
	Scope             string    `json:"scope" gorm:"column:scope;size:30;not null;uniqueIndex:uq_operations_evidence_control_scope"`
	Result            string    `json:"result" gorm:"column:result;size:20;not null"`
	EvidenceReference string    `json:"evidence_reference" gorm:"column:evidence_reference;type:text;not null"`
	ReviewedBy        uint      `json:"reviewed_by" gorm:"column:reviewed_by;not null;index"`
	ReviewedAt        time.Time `json:"reviewed_at" gorm:"column:reviewed_at;not null"`
	Notes             string    `json:"notes,omitempty" gorm:"column:notes;type:text"`
	CreatedAt         time.Time `json:"created_at" gorm:"column:created_at;not null"`
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
