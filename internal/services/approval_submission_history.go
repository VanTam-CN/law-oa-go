package services

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ReadApprovalSubmissionSnapshot is a read-only history API, not a replacement
// for the active authority resolver used by decisions and case creation.
// An explicit legacy version is read as stored; it is never promoted to active.
func ReadApprovalSubmissionSnapshot(db *gorm.DB, approval *models.ApprovalRequest, requestedID string) (map[string]interface{}, error) {
	if approval == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if db == nil {
		return nil, errors.New("审批快照读取服务未初始化")
	}
	requestedID = strings.TrimSpace(requestedID)
	if requestedID != "" {
		if _, err := uuid.Parse(requestedID); err != nil {
			return nil, gorm.ErrRecordNotFound
		}
		if !db.Migrator().HasTable(&models.ApprovalSnapshot{}) {
			return nil, gorm.ErrRecordNotFound
		}
		var row models.ApprovalSnapshot
		if err := db.Where("id = ? AND approval_request_id = ?", requestedID, approval.ID).First(&row).Error; err != nil {
			return nil, err
		}
		if row.SnapshotType == models.IntakeSubmissionSnapshotType {
			view := *approval
			view.ActiveSubmissionSnapshotID = &row.ID
			return ResolveApprovalSubmission(db, &view)
		}
		snapshot := submissionMap(row.SnapshotData)
		if snapshot == nil {
			return nil, snapshotUnresolved("历史快照无法解析")
		}
		// Old snapshot_id values may be SNAP_*, so the row ID is deliberately
		// returned separately by the handler rather than overwriting old JSON.
		for _, boundID := range []string{submissionText(snapshot["approval_request_id"]), submissionText(submissionMap(snapshot["approval"])["id"])} {
			if boundID != "" && boundID != approval.ID {
				return nil, snapshotUnresolved("历史快照的审批关联不一致")
			}
		}
		return snapshot, nil
	}
	if approval.ActiveSubmissionSnapshotID != nil {
		return ResolveApprovalSubmission(db, approval)
	}
	metadata, err := submissionMetadata(approval.Metadata)
	if err != nil {
		return nil, err
	}
	var count int64
	if db.Migrator().HasTable(&models.ApprovalSnapshot{}) {
		if err := db.Model(&models.ApprovalSnapshot{}).Where("approval_request_id = ?", approval.ID).Count(&count).Error; err != nil {
			return nil, err
		}
	}
	if count == 0 && submissionMap(metadata["approval_snapshot"]) == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return ResolveApprovalSubmission(db, approval)
}

type ApprovalSupplementContext struct {
	ReturnRecordID     string                     `json:"return_record_id"`
	SupplementRequests []models.SupplementRequest `json:"supplement_requests"`
	ReturnReason       string                     `json:"return_reason"`
}

// ReadApprovalSupplementContext reuses the same latest-return and round checks
// as saving replies/resubmitting. Callers must authorize the subject first.
func ReadApprovalSupplementContext(db *gorm.DB, approval *models.ApprovalRequest) (*ApprovalSupplementContext, error) {
	if approval == nil || approval.Status != models.ApprovalStatusNeedsRevision {
		return nil, nil
	}
	intakeID, err := ResolveApprovalIntakeID(db, approval)
	if err != nil {
		return nil, err
	}
	if intakeID == "" && approval.ActiveSubmissionSnapshotID == nil {
		return nil, nil
	}
	var latest models.ApprovalRecord
	if err := db.Where("approval_request_id = ? AND decision = ?", approval.ID, models.ApprovalDecisionRequestChanges).Order("created_at DESC, id DESC").First(&latest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, submissionStale("当前审批缺少可核验的退回记录")
		}
		return nil, err
	}
	record, err := currentIntakeReturn(db, approval, latest.ID)
	if err != nil {
		return nil, err
	}
	items, err := ReadSupplementRequests(record)
	if err != nil {
		return nil, err
	}
	return &ApprovalSupplementContext{ReturnRecordID: record.ID, SupplementRequests: items, ReturnReason: record.DecisionReason}, nil
}
