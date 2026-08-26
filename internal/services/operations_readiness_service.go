package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

const OperationsReadinessMaximumScore = 7

var operationsReadinessControls = map[string]bool{
	"backup": true, "restore_drill": true, "incident_owner": true,
	"upgrade": true, "rollback": true,
}

var operationsReadinessScopes = map[string]bool{
	models.OperationsEvidenceScopeQA:              true,
	models.OperationsEvidenceScopeControlledPilot: true,
}

type OperationsReadinessEvidenceInput struct {
	Control           string    `json:"control" binding:"required"`
	Scope             string    `json:"scope" binding:"required"`
	EvidenceReference string    `json:"evidence_reference" binding:"required"`
	ReviewedAt        time.Time `json:"reviewed_at" binding:"required"`
	Notes             string    `json:"notes"`
}

type OperationsReadinessEvidenceView struct {
	ID                string    `json:"id"`
	Control           string    `json:"control"`
	Scope             string    `json:"scope"`
	EvidenceReference string    `json:"evidence_reference"`
	ReviewedBy        uint      `json:"reviewed_by"`
	ReviewedAt        time.Time `json:"reviewed_at"`
	Notes             string    `json:"notes,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type OperationsReadinessSummary struct {
	Scope           string                              `json:"scope"`
	Ready           bool                                `json:"ready"`
	Score           int                                 `json:"score"`
	MaximumScore    int                                 `json:"maximum_score"`
	VerifiedCount   int                                 `json:"verified_count"`
	Total           int                                 `json:"total"`
	ProductionReady bool                                `json:"production_ready"`
	ProductionGate  string                              `json:"production_gate"`
	Items           []OperationsReadinessControlSummary `json:"items"`
}

type OperationsReadinessControlSummary struct {
	Control  string                           `json:"control"`
	Status   string                           `json:"status"`
	Evidence *OperationsReadinessEvidenceView `json:"evidence,omitempty"`
}

type OperationsReadinessService struct {
	db *gorm.DB
}

func NewOperationsReadinessService(db *gorm.DB) *OperationsReadinessService {
	return &OperationsReadinessService{db: db}
}

func (s *OperationsReadinessService) Register(actor AuthActor, input OperationsReadinessEvidenceInput) (*OperationsReadinessEvidenceView, error) {
	if actor.UserID == 0 {
		return nil, errors.New("authenticated reviewer is required")
	}
	if actor.Role != "admin" && actor.Role != "super_admin" {
		return nil, errors.New("only system administrators can register operations evidence")
	}
	control := strings.ToLower(strings.TrimSpace(input.Control))
	scope := strings.ToLower(strings.TrimSpace(input.Scope))
	reference := strings.TrimSpace(input.EvidenceReference)
	if !operationsReadinessControls[control] {
		return nil, fmt.Errorf("unknown operations control %q", input.Control)
	}
	if !operationsReadinessScopes[scope] {
		return nil, errors.New("scope must be qa or controlled_pilot")
	}
	if len(reference) < 8 || len(reference) > 1000 {
		return nil, errors.New("evidence reference must contain 8 to 1000 characters")
	}
	if strings.Contains(strings.ToLower(reference), "password") || strings.Contains(strings.ToLower(reference), "secret") || strings.Contains(strings.ToLower(reference), "token") {
		return nil, errors.New("evidence reference must not contain credentials")
	}
	if input.ReviewedAt.IsZero() || input.ReviewedAt.After(time.Now().Add(time.Minute)) {
		return nil, errors.New("review timestamp cannot be empty or in the future")
	}
	if len(strings.TrimSpace(input.Notes)) > 1000 {
		return nil, errors.New("notes must contain at most 1000 characters")
	}

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("generate evidence id: %w", err)
	}
	record := models.OperationsReadinessEvidence{
		ID: hex.EncodeToString(idBytes), Control: control, Scope: scope,
		Result: models.OperationsEvidenceResultPassed, EvidenceReference: reference,
		ReviewedBy: actor.UserID, ReviewedAt: input.ReviewedAt.UTC(),
		Notes: strings.TrimSpace(input.Notes),
	}
	err := s.db.Create(&record).Error
	if err != nil {
		return nil, fmt.Errorf("register operations evidence: %w", err)
	}
	view := operationsEvidenceView(record)
	return &view, nil
}

func (s *OperationsReadinessService) Summary(scope string) (*OperationsReadinessSummary, error) {
	scope = strings.ToLower(strings.TrimSpace(scope))
	if !operationsReadinessScopes[scope] {
		return nil, errors.New("scope must be qa or controlled_pilot")
	}
	var records []models.OperationsReadinessEvidence
	if err := s.db.
		Where("scope = ? AND result = ?", scope, models.OperationsEvidenceResultPassed).
		Order("control ASC, reviewed_at DESC, created_at DESC, id DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("read operations evidence: %w", err)
	}
	byControl := make(map[string]models.OperationsReadinessEvidence, len(records))
	for _, record := range records {
		if _, exists := byControl[record.Control]; !exists {
			byControl[record.Control] = record
		}
	}
	summary := OperationsReadinessSummary{
		Scope: scope, Total: len(operationsReadinessControls),
		MaximumScore: OperationsReadinessMaximumScore, ProductionGate: "production_external_evidence",
	}
	for control := range operationsReadinessControls {
		item := OperationsReadinessControlSummary{Control: control, Status: "pending-evidence"}
		if record, ok := byControl[control]; ok {
			item.Status = "verified"
			view := operationsEvidenceView(record)
			item.Evidence = &view
			summary.VerifiedCount++
		}
		summary.Items = append(summary.Items, item)
	}
	if summary.Items == nil {
		summary.Items = []OperationsReadinessControlSummary{}
	}
	summary.Ready = summary.VerifiedCount == summary.Total
	// Five individually verified controls are the minimum sustainable small-firm
	// operations baseline and therefore reach 7/10. Partial evidence must never
	// receive the completion bonus or imply that health checks add coverage.
	summary.Score = summary.VerifiedCount
	if summary.Ready {
		summary.Score = OperationsReadinessMaximumScore
	}
	return &summary, nil
}

func operationsEvidenceView(record models.OperationsReadinessEvidence) OperationsReadinessEvidenceView {
	return OperationsReadinessEvidenceView{
		ID: record.ID, Control: record.Control, Scope: record.Scope,
		EvidenceReference: record.EvidenceReference, ReviewedBy: record.ReviewedBy,
		ReviewedAt: record.ReviewedAt, Notes: record.Notes, CreatedAt: record.CreatedAt,
	}
}
