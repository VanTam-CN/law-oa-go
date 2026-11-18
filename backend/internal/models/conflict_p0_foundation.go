package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ConflictSubjectVersion is an immutable snapshot of a subject that is
// eligible for firm-wide conflict indexing. A new source snapshot creates a
// new row; callers must never rewrite a prior version in place.
type ConflictSubjectVersion struct {
	ID             string    `json:"id" gorm:"primaryKey;column:id;size:100"`
	SubjectKey     string    `json:"subject_key" gorm:"column:subject_key;size:200;not null;index:idx_conflict_subject_versions_key"`
	SourceType     string    `json:"source_type" gorm:"column:source_type;size:50;not null;index"`
	SourceID       string    `json:"source_id" gorm:"column:source_id;size:100;not null;index"`
	CaseID         string    `json:"case_id,omitempty" gorm:"column:case_id;size:100;index"`
	ClientID       string    `json:"client_id,omitempty" gorm:"column:client_id;size:100;index"`
	SubjectRole    string    `json:"subject_role" gorm:"column:subject_role;size:50;not null"`
	SubjectType    string    `json:"subject_type" gorm:"column:subject_type;size:50;not null"`
	OriginalName   string    `json:"original_name" gorm:"column:original_name;size:255;not null"`
	NormalizedName string    `json:"normalized_name" gorm:"column:normalized_name;size:255;not null;index"`
	AliasSnapshot  string    `json:"alias_snapshot,omitempty" gorm:"column:alias_snapshot;type:text"`
	SourceVersion  string    `json:"source_version" gorm:"column:source_version;size:120;not null"`
	VersionNumber  int       `json:"version_number" gorm:"column:version_number;not null;default:1"`
	Verification   string    `json:"verification_status" gorm:"column:verification_status;size:40;not null"`
	Snapshot       string    `json:"snapshot,omitempty" gorm:"column:snapshot;type:text"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (ConflictSubjectVersion) TableName() string { return "conflict_subject_versions" }

func (ConflictSubjectVersion) BeforeUpdate(*gorm.DB) error {
	return errors.New("conflict subject versions are append-only; create a new version")
}

func (ConflictSubjectVersion) BeforeDelete(*gorm.DB) error {
	return errors.New("conflict subject versions are append-only")
}

// ConflictSubjectIdentifier stores only protected identity material for one
// subject version. Digest is used for matching; ciphertext is available only
// to the explicitly authorized reviewer workflow.
type ConflictSubjectIdentifier struct {
	ID               string    `json:"id" gorm:"primaryKey;column:id;size:100"`
	SubjectVersionID string    `json:"subject_version_id" gorm:"column:subject_version_id;size:100;not null;index"`
	IdentifierType   string    `json:"identifier_type" gorm:"column:identifier_type;size:50;not null"`
	Digest           string    `json:"digest,omitempty" gorm:"column:digest;size:64;index"`
	Ciphertext       string    `json:"ciphertext,omitempty" gorm:"column:ciphertext;type:text"`
	MaskedValue      string    `json:"masked_value,omitempty" gorm:"column:masked_value;size:80"`
	Verification     string    `json:"verification_status" gorm:"column:verification_status;size:40;not null"`
	SourceReference  string    `json:"source_reference,omitempty" gorm:"column:source_reference;type:text"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (ConflictSubjectIdentifier) TableName() string { return "conflict_subject_identifiers" }

func (ConflictSubjectIdentifier) BeforeUpdate(*gorm.DB) error {
	return errors.New("conflict subject identifiers are append-only; create a new version")
}

func (ConflictSubjectIdentifier) BeforeDelete(*gorm.DB) error {
	return errors.New("conflict subject identifiers are append-only")
}

// ConflictMatchEvidenceV2 is the normalized evidence row for one check. The
// evidence snapshot is immutable and its hash lets the reviewer prove that a
// later response was not silently substituted for the reviewed result.
type ConflictMatchEvidenceV2 struct {
	ID               string    `json:"id" gorm:"primaryKey;column:id;size:100"`
	CheckID          string    `json:"check_id" gorm:"column:check_id;size:100;not null;index"`
	SubjectVersionID string    `json:"subject_version_id,omitempty" gorm:"column:subject_version_id;size:100;index"`
	MatchType        string    `json:"match_type" gorm:"column:match_type;size:40;not null"`
	SourceType       string    `json:"source_type" gorm:"column:source_type;size:50;not null"`
	SourceObjectID   string    `json:"source_object_id,omitempty" gorm:"column:source_object_id;size:100;index"`
	Restricted       bool      `json:"restricted" gorm:"column:restricted;not null;default:false"`
	EvidenceSnapshot string    `json:"evidence_snapshot" gorm:"column:evidence_snapshot;type:text;not null"`
	EvidenceHash     string    `json:"evidence_hash" gorm:"column:evidence_hash;size:64;not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (ConflictMatchEvidenceV2) TableName() string { return "conflict_match_evidence_v2" }

func (ConflictMatchEvidenceV2) BeforeUpdate(*gorm.DB) error {
	return errors.New("conflict match evidence is append-only")
}

func (ConflictMatchEvidenceV2) BeforeDelete(*gorm.DB) error {
	return errors.New("conflict match evidence is append-only")
}
