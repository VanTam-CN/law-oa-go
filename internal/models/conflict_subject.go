package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	SubjectStateEffective                  = "EFFECTIVE"
	SubjectStateChangeProposed             = "CHANGE_PROPOSED"
	SubjectStateRecheckRequired            = "RECHECK_REQUIRED"
	SubjectStateRecheckRunning             = "RECHECK_RUNNING"
	SubjectStateChangeApprovedAndEffective = "CHANGE_APPROVED_AND_EFFECTIVE"
	SubjectStateChangeRejected             = "CHANGE_REJECTED"
)

const (
	ConflictIndexBuildRunning  = "RUNNING"
	ConflictIndexBuildComplete = "COMPLETED"
	ConflictIndexBuildFailed   = "FAILED"
)

// ConflictIndexBuildRun is the durable reconciliation evidence for one
// authoritative archive scope. A scope may only become COMPLETE after it is
// linked to a completed run with zero missing source records.
type ConflictIndexBuildRun struct {
	ID                 string     `json:"id" gorm:"primaryKey;column:id;size:100"`
	ScopeType          string     `json:"scope_type" gorm:"column:scope_type;size:50;not null;index"`
	SourceVersion      string     `json:"source_version" gorm:"column:source_version;size:120;not null"`
	Status             string     `json:"status" gorm:"column:status;size:20;not null;index"`
	SourceRecordCount  int64      `json:"source_record_count" gorm:"column:source_record_count;not null;default:0"`
	IndexedRecordCount int64      `json:"indexed_record_count" gorm:"column:indexed_record_count;not null;default:0"`
	MissingRecordCount int64      `json:"missing_record_count" gorm:"column:missing_record_count;not null;default:0"`
	ReconciliationHash string     `json:"reconciliation_hash" gorm:"column:reconciliation_hash;size:64;not null"`
	EvidenceReference  string     `json:"evidence_reference" gorm:"column:evidence_reference;type:text"`
	StartedAt          time.Time  `json:"started_at" gorm:"column:started_at;not null"`
	CompletedAt        *time.Time `json:"completed_at,omitempty" gorm:"column:completed_at"`
	CreatedBy          *uint      `json:"created_by,omitempty" gorm:"column:created_by"`
	ErrorMessage       string     `json:"error_message,omitempty" gorm:"column:error_message;type:text"`
	CreatedAt          time.Time  `json:"created_at" gorm:"column:created_at;not null"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"column:updated_at;not null"`
}

func (ConflictIndexBuildRun) TableName() string { return "conflict_index_build_runs" }

func (ConflictIndexBuildRun) BeforeDelete(*gorm.DB) error {
	return errors.New("conflict index build runs are append-only")
}

// ConflictSearchScope records one authoritative archive source and the
// reconciliation evidence approved by the law firm. A missing or incomplete
// active scope keeps every conflict decision in COVERAGE_LIMITED.
type ConflictSearchScope struct {
	ID                string     `json:"id" gorm:"primaryKey;column:id;size:100"`
	ScopeType         string     `json:"scope_type" gorm:"column:scope_type;size:50;not null"`
	Status            string     `json:"status" gorm:"column:status;size:20;not null"`
	CoverageStatus    string     `json:"coverage_status" gorm:"column:coverage_status;size:30;not null"`
	SourceVersion     string     `json:"source_version,omitempty" gorm:"column:source_version;size:120"`
	EvidenceReference string     `json:"evidence_reference,omitempty" gorm:"column:evidence_reference;type:text"`
	CoveredFrom       *time.Time `json:"covered_from,omitempty" gorm:"column:covered_from"`
	CoveredTo         *time.Time `json:"covered_to,omitempty" gorm:"column:covered_to"`
	MissingSources    string     `json:"missing_sources,omitempty" gorm:"column:missing_sources;type:text"`
	IndexRunID        string     `json:"index_run_id,omitempty" gorm:"column:index_run_id;size:100;index"`
	ApprovedBy        *uint      `json:"approved_by,omitempty" gorm:"column:approved_by"`
	ApprovedAt        *time.Time `json:"approved_at,omitempty" gorm:"column:approved_at"`
	CreatedAt         time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"column:updated_at"`
}

func (ConflictSearchScope) TableName() string { return "conflict_search_scopes" }

// CaseSubjectRevision stores a proposed subject-set change separately from the
// effective case snapshot. The old snapshot remains in force until an
// independent review explicitly approves the new version.
type CaseSubjectRevision struct {
	ID                 string     `json:"id" gorm:"primaryKey;column:id;size:36"`
	CaseID             uint       `json:"case_id" gorm:"column:case_id;not null;index"`
	BaseSubjectVersion int        `json:"base_subject_version" gorm:"column:base_subject_version;not null"`
	ChangeType         string     `json:"change_type" gorm:"column:change_type;size:80;not null"`
	Status             string     `json:"status" gorm:"column:status;size:50;not null;index"`
	Payload            string     `json:"payload" gorm:"column:payload;type:text;not null"`
	Reason             string     `json:"reason" gorm:"column:reason;type:text"`
	ConflictCheckID    string     `json:"conflict_check_id,omitempty" gorm:"column:conflict_check_id;size:100;index"`
	RequestedBy        uint       `json:"requested_by" gorm:"column:requested_by;not null;index"`
	ReviewedBy         *uint      `json:"reviewed_by,omitempty" gorm:"column:reviewed_by"`
	ReviewDecision     string     `json:"review_decision,omitempty" gorm:"column:review_decision;size:40"`
	ReviewNotes        string     `json:"review_notes,omitempty" gorm:"column:review_notes;type:text"`
	CreatedAt          time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"column:updated_at"`
	EffectiveAt        *time.Time `json:"effective_at,omitempty" gorm:"column:effective_at"`
}

func (CaseSubjectRevision) TableName() string { return "case_subject_revisions" }

// ComplianceAuditEvent is append-only application evidence. There is
// intentionally no DeletedAt field and no delete handler for this model.
type ComplianceAuditEvent struct {
	ID             string    `json:"id" gorm:"primaryKey;column:id;size:36"`
	ActorID        *uint     `json:"actor_id,omitempty" gorm:"column:actor_id"`
	ActorRole      string    `json:"actor_role,omitempty" gorm:"column:actor_role;size:50"`
	EventType      string    `json:"event_type" gorm:"column:event_type;size:80;not null"`
	ObjectType     string    `json:"object_type" gorm:"column:object_type;size:80;not null"`
	ObjectID       string    `json:"object_id" gorm:"column:object_id;size:100;not null"`
	RequestID      string    `json:"request_id,omitempty" gorm:"column:request_id;size:100"`
	FromState      string    `json:"from_state,omitempty" gorm:"column:from_state;size:50"`
	ToState        string    `json:"to_state,omitempty" gorm:"column:to_state;size:50"`
	SubjectVersion int       `json:"subject_version,omitempty" gorm:"column:subject_version"`
	Payload        string    `json:"payload,omitempty" gorm:"column:payload;type:text"`
	IntegrityHash  string    `json:"integrity_hash,omitempty" gorm:"column:integrity_hash;size:64"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
}

func (ComplianceAuditEvent) TableName() string { return "compliance_audit_events" }

func (ComplianceAuditEvent) BeforeUpdate(*gorm.DB) error {
	return errors.New("compliance audit events are append-only")
}

func (ComplianceAuditEvent) BeforeDelete(*gorm.DB) error {
	return errors.New("compliance audit events are append-only")
}

func (CaseSubjectRevision) BeforeDelete(*gorm.DB) error {
	return errors.New("case subject revisions are append-only")
}
