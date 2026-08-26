package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func canRegisterOperationsReadinessEvidence(role string) bool {
	return role == "admin" || role == "super_admin" || role == "director"
}

type OperationsReadinessEvidenceInput struct {
	Control           string    `json:"control" binding:"required"`
	Scope             string    `json:"scope" binding:"required"`
	EvidenceReference string    `json:"evidence_reference" binding:"required"`
	ReviewedAt        time.Time `json:"reviewed_at" binding:"required"`
	Notes             string    `json:"notes"`
}

type OperationsReadinessEvidenceView struct {
	ID                 string    `json:"id"`
	Control            string    `json:"control"`
	Scope              string    `json:"scope"`
	EvidenceReference  string    `json:"evidence_reference"`
	ReviewedBy         uint      `json:"reviewed_by"`
	ReviewedAt         time.Time `json:"reviewed_at"`
	Notes              string    `json:"notes,omitempty"`
	PreviousEvidenceID string    `json:"previous_evidence_id"`
	IntegrityHash      string    `json:"integrity_hash"`
	CreatedAt          time.Time `json:"created_at"`
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
	Control   string                           `json:"control"`
	Status    string                           `json:"status"`
	Integrity string                           `json:"integrity"`
	Evidence  *OperationsReadinessEvidenceView `json:"evidence,omitempty"`
}

type OperationsReadinessService struct {
	db *gorm.DB
}

type operationsEvidenceChainPayload struct {
	ID                 string `json:"id"`
	Control            string `json:"control"`
	Scope              string `json:"scope"`
	Result             string `json:"result"`
	EvidenceReference  string `json:"evidence_reference"`
	ReviewedBy         uint   `json:"reviewed_by"`
	ReviewedAt         string `json:"reviewed_at"`
	Notes              string `json:"notes"`
	PreviousEvidenceID string `json:"previous_evidence_id"`
}

func NewOperationsReadinessService(db *gorm.DB) *OperationsReadinessService {
	return &OperationsReadinessService{db: db}
}

func (s *OperationsReadinessService) Register(actor AuthActor, input OperationsReadinessEvidenceInput) (*OperationsReadinessEvidenceView, error) {
	if actor.UserID == 0 {
		return nil, errors.New("authenticated reviewer is required")
	}
	if !canRegisterOperationsReadinessEvidence(actor.Role) {
		return nil, errors.New("only directors or system administrators can register operations evidence")
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
	if len(reference) < 8 || len(reference) > 512 {
		return nil, errors.New("evidence reference must contain 8 to 512 characters")
	}
	if !validOperationsEvidenceReference(reference) {
		return nil, errors.New("evidence reference must use archive://, ticket://, qa://, controlled-pilot://, or https:// and contain an identifier")
	}
	if strings.Contains(strings.ToLower(reference), "password") || strings.Contains(strings.ToLower(reference), "secret") || strings.Contains(strings.ToLower(reference), "token") || strings.Contains(strings.ToLower(reference), "credential") {
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
	var previousID string
	transactionErr := s.db.Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			// Serialize hash-chain extension separately for each scope. The lock
			// is transaction-scoped and acquired before the latest row is read.
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", "operations-readiness-evidence:"+scope).Error; err != nil {
				return fmt.Errorf("lock operations evidence scope: %w", err)
			}
		}
		latest := new(models.OperationsReadinessEvidence)
		if err := tx.Where("scope = ? AND integrity_hash <> ?", scope, "").
			Order("created_at DESC, reviewed_at DESC, id DESC").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(latest).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		previousID = latest.ID
		record.PreviousEvidenceID = previousID
		record.IntegrityHash = operationsEvidenceHash(record)
		return tx.Create(&record).Error
	})
	if transactionErr != nil {
		return nil, fmt.Errorf("register operations evidence: %w", transactionErr)
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
		Order("created_at ASC, reviewed_at ASC, id ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("read operations evidence: %w", err)
	}
	chainIntegrity := "verified"
	legacyRows := false
	byControl := make(map[string]models.OperationsReadinessEvidence, len(records))
	previousID := ""
	for _, record := range records {
		if record.IntegrityHash == "" || record.PreviousEvidenceID == "" && previousID != "" {
			legacyRows = true
			continue
		}
		if record.PreviousEvidenceID != previousID || record.IntegrityHash != operationsEvidenceHash(record) {
			chainIntegrity = "failed"
			break
		}
		previousID = record.ID
		// The chain is registration-ordered; for repeated validation of one
		// control, the most recently registered row supersedes older evidence.
		byControl[record.Control] = record
	}
	if chainIntegrity == "failed" {
		byControl = make(map[string]models.OperationsReadinessEvidence, len(operationsReadinessControls))
	} else if legacyRows {
		chainIntegrity = "legacy-rows-present"
		byControl = make(map[string]models.OperationsReadinessEvidence, len(operationsReadinessControls))
	}
	summary := OperationsReadinessSummary{
		Scope: scope, Total: len(operationsReadinessControls),
		MaximumScore: OperationsReadinessMaximumScore, ProductionGate: "production_external_evidence",
	}
	for control := range operationsReadinessControls {
		item := OperationsReadinessControlSummary{Control: control, Status: "pending-evidence", Integrity: chainIntegrity}
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
	summary.Ready = chainIntegrity == "verified" && summary.VerifiedCount == summary.Total
	// Five individually verified controls are the minimum sustainable small-firm
	// operations baseline and therefore reach 7/10. Partial evidence must never
	// receive the completion bonus or imply that health checks add coverage.
	summary.Score = 0
	if chainIntegrity == "verified" {
		summary.Score = summary.VerifiedCount
	}
	if summary.Ready {
		summary.Score = OperationsReadinessMaximumScore
	}
	return &summary, nil
}

func operationsEvidenceView(record models.OperationsReadinessEvidence) OperationsReadinessEvidenceView {
	return OperationsReadinessEvidenceView{
		ID: record.ID, Control: record.Control, Scope: record.Scope,
		EvidenceReference: record.EvidenceReference, ReviewedBy: record.ReviewedBy,
		ReviewedAt: record.ReviewedAt, Notes: record.Notes,
		PreviousEvidenceID: record.PreviousEvidenceID, IntegrityHash: record.IntegrityHash,
		CreatedAt: record.CreatedAt,
	}
}

func operationsEvidenceHash(record models.OperationsReadinessEvidence) string {
	payload := operationsEvidenceChainPayload{
		ID: record.ID, Control: record.Control, Scope: record.Scope,
		Result: record.Result, EvidenceReference: record.EvidenceReference,
		ReviewedBy: record.ReviewedBy, ReviewedAt: record.ReviewedAt.UTC().Format(time.RFC3339Nano),
		Notes: record.Notes, PreviousEvidenceID: record.PreviousEvidenceID,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func validOperationsEvidenceReference(reference string) bool {
	parsed, err := url.ParseRequestURI(reference)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	identifier := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(reference, scheme+":"), "/"))
	switch scheme {
	case "archive", "ticket", "qa", "controlled-pilot":
		return identifier != "" && !strings.Contains(identifier, "://")
	case "https":
		return parsed.Host != "" && parsed.Path != ""
	default:
		return false
	}
}
