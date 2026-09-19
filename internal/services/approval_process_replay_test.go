package services

import (
	"context"
	"strconv"
	"testing"

	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// replayProbeApprovalService serves one fixed approval row for both the
// authorization lookup and the generic decision path.
type replayProbeApprovalService struct {
	txProbeApprovalService
	processCalls int
}

func (s *replayProbeApprovalService) GetApproval(userID string, id string) (*models.ApprovalRequest, error) {
	return s.approval, nil
}

func (s *replayProbeApprovalService) ProcessApproval(userID, userName, id string, decisionReq *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	s.processCalls++
	return nil, gorm.ErrInvalidDB
}

// An approve replay on an already-approved, already-linked case-creation
// approval must return the SAME complete case through the idempotent
// short-circuit without ever reaching the generic decision path. Reaching it
// would fail with a state error and turn every client retry into 500.
func TestProcessApprovalWithConflictApproveReplayIsIdempotent(t *testing.T) {
	fixture := seedApprovalCaseWorld(t)
	approval := fixture.approval
	approval.Status = models.ApprovalStatusApproved
	approval.CaseCreated = true
	approval.CreatedCaseID = "3"

	probe := &replayProbeApprovalService{txProbeApprovalService: txProbeApprovalService{approval: approval}}
	fixture.svc.(*approvalConflictIntegrationService).approvalService = probe

	// The fixture world already contains a case row; overwrite its binding to
	// the replayed case ID and make sure it is really present.
	var client models.Client
	if err := fixture.db.First(&client).Error; err != nil {
		t.Fatalf("load client: %v", err)
	}
	existingCase := &models.Case{CaseNumber: "CASE-REPLAY-1", Title: "成案重放案件", ClientID: client.ID, CaseType: "CIVIL", Status: "active"}
	if err := fixture.db.Create(existingCase).Error; err != nil {
		t.Fatalf("seed case: %v", err)
	}
	approval.CreatedCaseID = strconv.FormatUint(uint64(existingCase.ID), 10)

	decision := &models.ApprovalDecisionRequest{Decision: models.ApprovalDecisionApprove, DecisionReason: "重放同一同意请求"}
	result, err := fixture.svc.ProcessApprovalWithConflict(context.Background(), "director-1", "董事", approval.ID, decision)
	if err != nil {
		t.Fatalf("replay returned error: %v", err)
	}
	if result == nil || result.ID != approval.ID || result.Status != models.ApprovalStatusApproved {
		t.Fatalf("replay result = %+v, want same approved approval %s", result, approval.ID)
	}
	if probe.processCalls != 0 {
		t.Fatalf("generic decision path reached %d times, want 0", probe.processCalls)
	}

	var count int64
	if err := fixture.db.Model(&models.Case{}).Where("case_number = ?", "CASE-REPLAY-1").Count(&count).Error; err != nil {
		t.Fatalf("count cases: %v", err)
	}
	if count != 1 {
		t.Fatalf("case rows for CASE-REPLAY-1 = %d, want 1", count)
	}
}

// A replay whose approval has not completed case creation must not be
// published as a completed replay: it falls through to the generic decision
// path (or its failure), never fabricates a success.
func TestProcessApprovalWithConflictReplayWithoutCaseIsRejected(t *testing.T) {
	fixture := seedApprovalCaseWorld(t)
	approval := fixture.approval
	approval.Status = models.ApprovalStatusApproved
	approval.CaseCreated = true
	approval.CreatedCaseID = ""

	probe := &replayProbeApprovalService{txProbeApprovalService: txProbeApprovalService{approval: approval}}
	fixture.svc.(*approvalConflictIntegrationService).approvalService = probe

	decision := &models.ApprovalDecisionRequest{Decision: models.ApprovalDecisionApprove, DecisionReason: "未成案重放"}
	if _, err := fixture.svc.ProcessApprovalWithConflict(context.Background(), "director-1", "董事", approval.ID, decision); err == nil {
		t.Fatalf("replay without case must fail, got success")
	}
}

// A dangling created_case_id (approval says linked, case row missing) must be
// rejected loudly as CASE_CREATION_INCOMPLETE, never returned as a completed
// replay.
func TestProcessApprovalWithConflictReplayWithMissingCaseRow(t *testing.T) {
	fixture := seedApprovalCaseWorld(t)
	approval := fixture.approval
	approval.Status = models.ApprovalStatusApproved
	approval.CaseCreated = true
	approval.CreatedCaseID = "999999"

	probe := &replayProbeApprovalService{txProbeApprovalService: txProbeApprovalService{approval: approval}}
	fixture.svc.(*approvalConflictIntegrationService).approvalService = probe

	decision := &models.ApprovalDecisionRequest{Decision: models.ApprovalDecisionApprove, DecisionReason: "悬空链接重放"}
	_, err := fixture.svc.ProcessApprovalWithConflict(context.Background(), "director-1", "董事", approval.ID, decision)
	if err == nil {
		t.Fatalf("dangling case link replay must fail")
	}
	wfErr, ok := err.(*SubjectWorkflowError)
	if !ok || wfErr.Code != "CASE_CREATION_INCOMPLETE" {
		t.Fatalf("dangling case link error = %v, want CASE_CREATION_INCOMPLETE", err)
	}
}
