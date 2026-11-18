package services

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
)

// N3 / C08 decision-path contract: a request_changes decision persists the
// normalized structured items into the approval record, refuses empty
// reasons with a typed 400 error, and normalizes text-only returns into one
// legacy_text item.

func seedSupplementDecisionWorld(t *testing.T) (*gorm.DB, *ApprovalService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s", t.TempDir()+"/supplement_decision.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.ApprovalRequest{},
		&models.ApprovalRecord{},
		&models.ApprovalWorkflow{},
		&models.User{},
		&models.InboxItem{},
	))
	svc := NewApprovalService(db)
	return db, svc
}

func newSupplementDecisionApproval(t *testing.T, db *gorm.DB, applicantID, approverID string) *models.ApprovalRequest {
	t.Helper()
	approval := &models.ApprovalRequest{
		ID:                fmt.Sprintf("APR-SUP-%d", time.Now().UnixNano()),
		RequestNumber:     fmt.Sprintf("RN-SUP-%d", time.Now().UnixNano()),
		Title:             "structured supplement decision",
		Type:              "CASE_CREATION",
		Content:           "structured supplement decision path",
		ApplicantID:       applicantID,
		ApplicantName:     "申请人",
		Status:            models.ApprovalStatusSubmitted,
		WorkflowType:      "CASE_APPROVAL",
		CurrentApproverID: approverID,
	}
	require.NoError(t, db.Create(approval).Error)
	return approval
}

func structuredSupplementPayload() map[string]interface{} {
	return map[string]interface{}{
		"schema_version": 1,
		"items": []interface{}{
			map[string]interface{}{
				"category":    "field_correction",
				"field_key":   "case_title",
				"description": "案件名称与委托合同不一致，请核对后修改",
			},
			map[string]interface{}{
				"category":       "identity_verification",
				"field_key":      "party",
				"party_local_id": "party-1",
				"description":    "需核验对方当事人张三的身份",
			},
		},
	}
}

func TestSupplementDecision_StructuredItemsPersist(t *testing.T) {
	db, svc := seedSupplementDecisionWorld(t)
	approval := newSupplementDecisionApproval(t, db, "u-lawyer", "u-approver")

	updated, err := svc.ProcessApprovalDecision("u-approver", approval.ID, &models.ApprovalDecisionRequest{
		Decision:            models.ApprovalDecisionRequestChanges,
		DecisionReason:      "材料需要补充和更正",
		ImposedRequirements: structuredSupplementPayload(),
	})
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusNeedsRevision, updated.Status)

	records, err := svc.GetApprovalRecords(approval.ID)
	require.NoError(t, err)
	require.Len(t, records, 1)
	rec := records[0]
	require.Equal(t, models.ApprovalDecisionRequestChanges, rec.Decision)

	var stored map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(rec.ImposedRequirements), &stored))
	require.Equal(t, float64(1), stored["schema_version"])
	storedItems, ok := stored["items"].([]interface{})
	require.True(t, ok)
	require.Len(t, storedItems, 2)
	for _, raw := range storedItems {
		item, itemOK := raw.(map[string]interface{})
		require.True(t, itemOK)
		require.NotEmpty(t, item["item_id"], "server must assign a stable item_id before persistence")
	}
}

// Regression for the isolated-PG C17 blocker (migration 000083 makes
// supplement_replies JSONB NOT NULL): every decision record insert must
// persist the empty draft object instead of an explicit NULL, which GORM
// would otherwise send and the database would reject (SQLSTATE 23502).
func TestSupplementDecision_RecordInsertPersistsSupplementRepliesNonNull(t *testing.T) {
	db, svc := seedSupplementDecisionWorld(t)
	approval := newSupplementDecisionApproval(t, db, "u-lawyer", "u-approver")

	updated, err := svc.ProcessApprovalDecision("u-approver", approval.ID, &models.ApprovalDecisionRequest{
		Decision:       models.ApprovalDecisionRequestChanges,
		DecisionReason: "材料需要补充和更正",
	})
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusNeedsRevision, updated.Status)

	var stored *string
	require.NoError(t, db.Raw("SELECT supplement_replies FROM approval_records WHERE approval_request_id = ?", approval.ID).Scan(&stored).Error)
	require.NotNil(t, stored, "decision record must not write NULL into supplement_replies")
	require.JSONEq(t, "{}", *stored)
}

func TestSupplementDecision_EmptyReasonIsTyped400(t *testing.T) {
	db, svc := seedSupplementDecisionWorld(t)
	approval := newSupplementDecisionApproval(t, db, "u-lawyer", "u-approver")

	_, err := svc.ProcessApprovalDecision("u-approver", approval.ID, &models.ApprovalDecisionRequest{
		Decision: models.ApprovalDecisionRequestChanges,
	})
	require.Error(t, err)
	var supErr *SupplementRequestError
	require.ErrorAs(t, err, &supErr)
	require.Equal(t, "退回必须填写审批意见", supErr.Message)
}

func TestSupplementDecision_TextOnlyNormalizesToLegacyItem(t *testing.T) {
	db, svc := seedSupplementDecisionWorld(t)
	approval := newSupplementDecisionApproval(t, db, "u-lawyer", "u-approver")

	updated, err := svc.ProcessApprovalDecision("u-approver", approval.ID, &models.ApprovalDecisionRequest{
		Decision:       models.ApprovalDecisionRequestChanges,
		DecisionReason: "请补充案件摘要",
	})
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusNeedsRevision, updated.Status)

	records, err := svc.GetApprovalRecords(approval.ID)
	require.NoError(t, err)
	require.Len(t, records, 1)
	var stored map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(records[0].ImposedRequirements), &stored))
	storedItems, ok := stored["items"].([]interface{})
	require.True(t, ok)
	require.Len(t, storedItems, 1)
	item, itemOK := storedItems[0].(map[string]interface{})
	require.True(t, itemOK)
	require.Equal(t, "legacy_text", item["source"])
}
