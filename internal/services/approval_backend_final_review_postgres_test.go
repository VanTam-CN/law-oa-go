package services

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// Persist the pre-protocol conflict approval shape, as still present in old
// installations: check binding, legacy immutable snapshot, no active pointer
// and no intake slot. This is compatibility input, not a replacement workflow.
func backendReviewLegacyApproval(t *testing.T, f *approvalCollaborationPostgresFixture) *models.ApprovalRequest {
	t.Helper()
	metadata := map[string]interface{}{"conflict_task_id": f.checkID, "source": "real_api"}
	a, err := NewApprovalService(f.db).CreateApproval("11", "虚构申请律师", &models.CreateApprovalRequest{
		Title: "旧冲突审批兼容样本", Type: "conflict_approval", Content: "已存在的旧冲突审批", WorkflowType: "COLLAB_PG", Metadata: metadata,
	})
	require.NoError(t, err)
	require.NoError(t, f.db.Model(a).Updates(map[string]interface{}{"status": models.ApprovalStatusSubmitted, "conflict_check_id": f.checkID, "current_approver_id": "22", "current_approver_name": "虚构审批合伙人", "current_stage": "partner_review", "submission_date": time.Now()}).Error)
	require.NoError(t, f.db.Create(&models.ApprovalSnapshot{ID: uuid.NewString(), ApprovalRequestID: a.ID, SnapshotType: "conflict_approval", SourceVersion: 1,
		SnapshotData: approvalCollaborationPostgresJSON(t, map[string]interface{}{"approval": map[string]interface{}{"id": a.ID}, "metadata": metadata})}).Error)
	return f.reload(t, a.ID)
}

func TestBackendFinalReviewPostgresLegacyApprovalOrders(t *testing.T) {
	for _, schedule := range []string{"formal_created_first", "formal_approved_first", "simultaneous_1", "simultaneous_2", "simultaneous_3"} {
		t.Run(schedule, func(t *testing.T) {
			f := newApprovalCollaborationPostgresIntakeFixture(t)
			var legacy *models.ApprovalRequest
			if schedule != "formal_created_first" {
				legacy = backendReviewLegacyApproval(t, f)
			}
			formal, err := f.integrated().CreateIntegratedApproval(context.Background(), "11", "虚构申请律师", f.newRequest())
			require.NoError(t, err)
			if legacy == nil {
				legacy = backendReviewLegacyApproval(t, f)
			}
			decision := f.decision(t, formal.ApprovalID, models.ApprovalDecisionApprove)
			var legacyErr, formalErr error
			approveOld := func() {
				_, legacyErr = f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批人", legacy.ID, &models.ApprovalDecisionRequest{Decision: models.ApprovalDecisionApprove})
			}
			approveFormal := func() {
				_, formalErr = f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批人", formal.ApprovalID, decision)
			}
			switch schedule {
			case "formal_created_first":
				before := f.state(t)
				approveOld()
				require.Equal(t, before, f.state(t))
				approveFormal()
			case "formal_approved_first":
				approveFormal()
				before := f.state(t)
				approveOld()
				require.Equal(t, before, f.state(t))
			default:
				start := make(chan struct{})
				var wg sync.WaitGroup
				wg.Add(2)
				go func() { defer wg.Done(); <-start; approveOld() }()
				go func() { defer wg.Done(); <-start; approveFormal() }()
				close(start)
				wg.Wait()
			}
			require.Error(t, legacyErr)
			require.NoError(t, formalErr)
			require.Equal(t, legacy, f.reload(t, legacy.ID), "losing old approval is byte-for-byte unchanged")
			require.Zero(t, f.count(t, "approval_records", "approval_request_id = ?", legacy.ID))
			require.EqualValues(t, 1, f.count(t, "approval_case_creation_tracking", "intake_id = ?", f.intakeID))
			f.assertOneCompleteCase(t, formal.ApprovalID)
		})
	}
}

func TestBackendFinalReviewPostgresLegacyServiceBoundaries(t *testing.T) {
	for _, entry := range []string{"ordinary_decision", "auto_case", "retry_case", "frozen_reference_only"} {
		t.Run(entry, func(t *testing.T) {
			f := newApprovalCollaborationPostgresIntakeFixture(t)
			a := backendReviewLegacyApproval(t, f)
			if entry == "frozen_reference_only" {
				require.NoError(t, f.db.Model(a).Updates(map[string]interface{}{"conflict_check_id": "", "metadata": "{}"}).Error)
				a = f.reload(t, a.ID)
				intakeID, err := ResolveApprovalIntakeID(f.db, a)
				require.NoError(t, err)
				require.Equal(t, f.intakeID, intakeID, "old persisted snapshot survives mutable binding removal")
			}
			if entry != "ordinary_decision" {
				require.NoError(t, f.db.Model(a).Updates(map[string]interface{}{"status": models.ApprovalStatusApproved, "case_creation_status": "failed"}).Error)
			}
			before := f.state(t)
			var err error
			switch entry {
			case "ordinary_decision":
				_, err = NewApprovalService(f.db).ProcessApprovalDecision("22", a.ID, &models.ApprovalDecisionRequest{Decision: models.ApprovalDecisionApprove})
			case "retry_case":
				_, err = f.integrated().RetryCaseCreationForUser(context.Background(), "11", a.ID)
			default:
				_, err = f.integrated().AutoCreateCaseFromApproval(context.Background(), a.ID, map[string]interface{}{"conflict_check_id": f.checkID})
			}
			require.ErrorContains(t, err, "旧冲突审批")
			require.Equal(t, before, f.state(t))
		})
	}
}

// Seed an already materialized pre-protocol case, not a successful business
// operation. Recovery below uses the real case verifier/party writer.
func backendReviewExistingLegacyCase(t *testing.T, f *approvalCollaborationPostgresFixture, a *models.ApprovalRequest, staleFlags bool) string {
	caseRow := &models.Case{CaseNumber: "B1-OLD-CASE", Title: "历史已成案件", ClientID: f.clientID, LawyerID: 11, CreatedBy: "11", ConflictCheckID: f.checkID}
	require.NoError(t, f.db.Create(caseRow).Error)
	caseID := strconv.FormatUint(uint64(caseRow.ID), 10)
	require.NoError(t, f.db.Model(a).Updates(map[string]interface{}{"status": models.ApprovalStatusApproved, "case_created": !staleFlags, "created_case_id": caseID, "case_creation_status": "failed"}).Error)
	require.NoError(t, f.db.Create(&repositories.CaseCreationTracking{ID: uuid.NewString(), ApprovalRequestID: a.ID, CaseID: &caseID, CreationStatus: "completed", CreatedBy: "11", ErrorDetails: "{}", DataMapping: "{}", MappedFields: "[]", UnmappedFields: "[]", AppliedConditions: "{}", ImposedRequirements: "{}", WorkflowActions: "[]"}).Error)
	return caseID
}

func TestBackendFinalReviewPostgresExistingLegacyCase(t *testing.T) {
	for _, mode := range []string{"completed_replay", "stale_flags_replay", "new_check_submit", "formal_already_submitted"} {
		t.Run(mode, func(t *testing.T) {
			f := newApprovalCollaborationPostgresIntakeFixture(t)
			a := backendReviewLegacyApproval(t, f)
			var formalID string
			if mode == "formal_already_submitted" {
				formal, err := f.integrated().CreateIntegratedApproval(context.Background(), "11", "虚构律师", f.newRequest())
				require.NoError(t, err)
				formalID = formal.ApprovalID
			}
			caseID := backendReviewExistingLegacyCase(t, f, a, mode == "stale_flags_replay")
			if mode == "new_check_submit" {
				f.reviseReviewedSubject(t, &models.ResubmitApprovalRequest{})
				before := f.state(t)
				_, err := f.integrated().CreateIntegratedApproval(context.Background(), "11", "虚构律师", f.newRequest())
				require.ErrorContains(t, err, "已生成正式案件")
				require.Equal(t, before, f.state(t))
			} else if mode == "formal_already_submitted" {
				before := f.state(t)
				_, err := f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批人", formalID, f.decision(t, formalID, models.ApprovalDecisionApprove))
				require.ErrorContains(t, err, "已有正式案件")
				require.Equal(t, before, f.state(t))
			} else {
				for i := 0; i < 2; i++ {
					result, err := f.integrated().RetryCaseCreationForUser(context.Background(), "11", a.ID)
					require.NoError(t, err)
					require.Equal(t, caseID, result.CaseID)
					require.Equal(t, "completed", result.Status)
				}
				f.assertOneCompleteCase(t, a.ID)
			}
			require.EqualValues(t, 1, f.count(t, "cases", "1=1"))
			t.Logf("legacy compatibility mode=%s case=%s case_count=1", mode, caseID)
		})
	}
}

func TestBackendFinalReviewPostgresNonIntakeLegacyCompatibility(t *testing.T) {
	f := newApprovalCollaborationPostgresIntakeFixture(t)
	require.NoError(t, f.db.Exec("UPDATE conflict_check_records SET search_parameters = (search_parameters::jsonb - 'intakeId')::json WHERE check_id = ?", f.checkID).Error)
	a := backendReviewLegacyApproval(t, f)
	require.NoError(t, f.db.Model(a).Updates(map[string]interface{}{"status": models.ApprovalStatusApproved, "case_creation_status": "failed"}).Error)
	result, err := f.integrated().RetryCaseCreationForUser(context.Background(), "11", a.ID)
	require.NoError(t, err, "a genuine non-intake legacy case still uses its original reviewed-party workflow")
	require.NotEmpty(t, result.CaseID)
	f.assertOneCompleteCase(t, a.ID)
}

func ApprovalBackendFinalReviewLegacyHTTPFixture(t *testing.T, approved bool) (*gorm.DB, *models.ApprovalRequest, string, func() map[string]string) {
	f := newApprovalCollaborationPostgresIntakeFixture(t)
	a := backendReviewLegacyApproval(t, f)
	if approved {
		require.NoError(t, f.db.Model(a).Updates(map[string]interface{}{"status": models.ApprovalStatusApproved, "case_creation_status": "failed"}).Error)
	}
	return f.db, a, f.checkID, func() map[string]string { return f.state(t) }
}

func backendReviewPendingClientChange(t *testing.T) (*approvalCollaborationPostgresFixture, *models.ResubmitApprovalRequest, uint, string) {
	f := newApprovalCollaborationPostgresFixture(t)
	oldClientID, oldCheckID := f.clientID, f.checkID
	req := f.returnForRevision(t)
	f.reviseReviewedSubject(t, req)
	client := &models.Client{Name: "补正后新客户", Type: "个人", CreatedBy: 11, IdentityType: models.IdentityTypeIDCard, IdentityNumber: "310110199001010044"}
	require.NoError(t, repositories.NewClientRepository(f.db).Create(context.Background(), client))
	f.clientID, f.clientDigest = client.ID, client.IdentityNumberDigest
	var check models.ConflictCheckRecord
	require.NoError(t, f.db.First(&check, "check_id = ?", f.checkID).Error)
	check.ClientID, check.ClientName = fmt.Sprint(client.ID), client.Name
	check.SearchParameters["clientIdentifiers"] = map[string]interface{}{"id_card": "hmac-sha256:" + client.IdentityNumberDigest}
	require.NoError(t, f.db.Save(&check).Error)
	require.NoError(t, f.db.Table("case_intakes").Where("id = ?", f.intakeID).Update("client_id", client.ID).Error)
	return f, req, oldClientID, oldCheckID
}

func ApprovalBackendFinalReviewPendingClientFixture(t *testing.T) (*gorm.DB, *models.ApprovalRequest, *models.ResubmitApprovalRequest, string, uint, uint, func() map[string]string) {
	f, req, oldClient, _ := backendReviewPendingClientChange(t)
	return f.db, f.reload(t, f.approval.ID), req, f.intakeID, oldClient, f.clientID, func() map[string]string { return f.state(t) }
}

func ApprovalBackendFinalReviewHistoricalClientFixture(t *testing.T) (*gorm.DB, *models.ApprovalRequest, string, uint, func() map[string]string) {
	f, req, oldClientID, oldCheckID := backendReviewPendingClientChange(t)
	oldSnapshot := *f.approval.ActiveSubmissionSnapshotID
	_, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
	require.NoError(t, err, "real resubmit must freeze C2/new client before historical access is tested")
	matter := &models.Case{CaseNumber: "B3-OLD-CLIENT", Title: "旧客户隔离样本", ClientID: oldClientID, LawyerID: 11, CreatedBy: "11", ConflictCheckID: oldCheckID}
	require.NoError(t, f.db.Create(matter).Error)
	return f.db, f.reload(t, f.approval.ID), oldSnapshot, matter.ID, func() map[string]string { return f.state(t) }
}

func TestBackendFinalReviewPostgresLegacyCannotCreateSecondCase(t *testing.T) {
	f := newApprovalCollaborationPostgresIntakeFixture(t)
	legacy := backendReviewLegacyApproval(t, f)
	formal, err := f.integrated().CreateIntegratedApproval(context.Background(), "11", "虚构申请律师", f.newRequest())
	require.NoError(t, err)
	before := f.state(t)
	_, legacyErr := f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批合伙人", legacy.ID, &models.ApprovalDecisionRequest{Decision: models.ApprovalDecisionApprove})
	afterLegacy := f.state(t)
	_, formalErr := f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批合伙人", formal.ApprovalID, f.decision(t, formal.ApprovalID, models.ApprovalDecisionApprove))
	t.Logf("B1 real PG legacy_error=%v formal_error=%v cases=%d legacy_records=%d", legacyErr, formalErr, f.count(t, "cases", "1=1"), f.count(t, "approval_records", "approval_request_id = ?", legacy.ID))
	require.Error(t, legacyErr, "legacy conflict approval must not become a second case approval")
	require.Equal(t, before, afterLegacy, "rejected old approval must make no decision/case/party/inbox writes")
	require.NoError(t, formalErr)
	f.assertOneCompleteCase(t, formal.ApprovalID)
}

// Genuine active snapshot and case context for external-package HTTP tests.
func ApprovalBackendFinalReviewSnapshotFixture(t *testing.T) (*gorm.DB, *models.ApprovalRequest, uint, func() map[string]string) {
	t.Helper()
	f := newApprovalCollaborationPostgresFixture(t)
	matter := &models.Case{CaseNumber: "B3-FICTIONAL", Title: "虚构授权撤销样本", ClientID: f.clientID, LawyerID: 11, CreatedBy: "11", ConflictCheckID: f.checkID}
	require.NoError(t, f.db.Create(matter).Error)
	return f.db, f.approval, matter.ID, func() map[string]string { return f.state(t) }
}
