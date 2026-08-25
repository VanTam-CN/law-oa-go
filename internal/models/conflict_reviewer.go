package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	ConflictReviewerAssignmentActive  = "ACTIVE"
	ConflictReviewerAssignmentRevoked = "REVOKED"
)

// ConflictReviewerAssignment is the auditable appointment that allows one
// independent reviewer (or a named delegate) to conclude a specific check.
// It is intentionally separate from the user's role: a role grants capacity,
// while this record grants assignment for this evidence set.
type ConflictReviewerAssignment struct {
	ID                 string     `json:"id" gorm:"primaryKey;column:id;size:36"`
	CheckID            string     `json:"check_id" gorm:"column:check_id;size:100;not null;index"`
	CaseID             *uint      `json:"case_id,omitempty" gorm:"column:case_id;index"`
	ReviewerID         uint       `json:"reviewer_id" gorm:"column:reviewer_id;not null;index"`
	DelegateForID      *uint      `json:"delegate_for_id,omitempty" gorm:"column:delegate_for_id;index"`
	AssignedBy         uint       `json:"assigned_by" gorm:"column:assigned_by;not null;index"`
	Status             string     `json:"status" gorm:"column:status;size:20;not null;index"`
	RecusalDeclared    bool       `json:"recusal_declared" gorm:"column:recusal_declared;not null;default:false"`
	IndependenceReason string     `json:"independence_reason" gorm:"column:independence_reason;type:text"`
	SLADueAt           *time.Time `json:"sla_due_at,omitempty" gorm:"column:sla_due_at"`
	EffectiveFrom      *time.Time `json:"effective_from,omitempty" gorm:"column:effective_from"`
	EffectiveTo        *time.Time `json:"effective_to,omitempty" gorm:"column:effective_to"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty" gorm:"column:revoked_at"`
	CreatedAt          time.Time  `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"column:updated_at;not null"`
}

// ConflictOfficerAppointment records the firm-level appointment and deputy
// arrangement. It grants eligibility only; every matter still requires a
// separate assignment and independence check.
type ConflictOfficerAppointment struct {
	ID                         string    `json:"id" gorm:"primaryKey;column:id;size:36"`
	OfficerID                  uint      `json:"officer_id" gorm:"column:officer_id;not null;index"`
	DeputyID                   *uint     `json:"deputy_id,omitempty" gorm:"column:deputy_id;index"`
	AppointedBy                uint      `json:"appointed_by" gorm:"column:appointed_by;not null;index"`
	EffectiveFrom              time.Time `json:"effective_from" gorm:"column:effective_from;not null;index"`
	EffectiveTo                time.Time `json:"effective_to" gorm:"column:effective_to;not null;index"`
	RecusalDeclaration         string    `json:"recusal_declaration" gorm:"column:recusal_declaration;type:text;not null"`
	ExternalMechanismReference string    `json:"external_mechanism_reference,omitempty" gorm:"column:external_mechanism_reference;type:text"`
	CreatedAt                  time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (ConflictOfficerAppointment) TableName() string { return "conflict_officer_appointments" }

func (ConflictOfficerAppointment) BeforeDelete(*gorm.DB) error {
	return errors.New("conflict officer appointments are append-only")
}

func (ConflictOfficerAppointment) BeforeUpdate(*gorm.DB) error {
	return errors.New("conflict officer appointments are append-only")
}

func (ConflictReviewerAssignment) TableName() string { return "conflict_reviewer_assignments" }

func (ConflictReviewerAssignment) BeforeDelete(*gorm.DB) error {
	return errors.New("conflict reviewer assignments are append-only; revoke instead")
}
