package services

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ApprovalRequiresSubmissionSnapshot is a read-only business-mode projection,
// not authorization or a substitute for transaction-time version/slot checks.
// Only an ordinary approval with no durable intake or snapshot can return false.
func ApprovalRequiresSubmissionSnapshot(db *gorm.DB, approval *models.ApprovalRequest) (bool, error) {
	intakeID, err := ResolveApprovalIntakeID(db, approval)
	if err != nil {
		return false, err
	}
	if intakeID != "" || approval.ActiveSubmissionSnapshotID != nil || strings.EqualFold(strings.TrimSpace(approval.Type), "case_creation") {
		return true, nil
	}
	metadata, err := submissionMetadata(approval.Metadata)
	if err != nil {
		return false, err
	}
	if metadata["approval_snapshot"] != nil {
		return true, nil
	}
	if db == nil {
		return false, fmt.Errorf("审批提交快照模式服务未初始化")
	}
	var count int64
	if err := db.Model(&models.ApprovalSnapshot{}).Where("approval_request_id = ?", approval.ID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("读取审批提交快照模式失败: %w", err)
	}
	return count > 0, nil
}
