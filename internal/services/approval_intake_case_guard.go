package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// Historical checks remain attached to their intake even after a later check
// supersedes them. A case created by a pre-protocol approval still consumes
// that intake: absence of a tracking slot must never authorize another case.
func intakeExistingCaseIDs(tx *gorm.DB, intakeID string) ([]uint, error) {
	var predicate string
	switch tx.Dialector.Name() {
	case "postgres":
		predicate = "(r.search_parameters->>'intakeId' = ? OR r.search_parameters->>'intake_id' = ?)"
	case "sqlite":
		predicate = "(json_extract(r.search_parameters, '$.intakeId') = ? OR json_extract(r.search_parameters, '$.intake_id') = ?)"
	case "mysql":
		predicate = "(JSON_UNQUOTE(JSON_EXTRACT(r.search_parameters, '$.intakeId')) = ? OR JSON_UNQUOTE(JSON_EXTRACT(r.search_parameters, '$.intake_id')) = ?)"
	default:
		return nil, fmt.Errorf("接案成案校验不支持数据库 %q", tx.Dialector.Name())
	}
	var ids []uint
	caseIDCast := "CAST(c.id AS TEXT)"
	if tx.Dialector.Name() == "mysql" {
		caseIDCast = "CAST(c.id AS CHAR)"
	}
	// The tracking link also survives edits to a check's mutable parameters.
	err := tx.Table("cases AS c").Joins("LEFT JOIN conflict_check_records r ON r.check_id = c.conflict_check_id").
		Where("c.deleted_at IS NULL").Where("("+predicate+" OR EXISTS (SELECT 1 FROM approval_case_creation_tracking t WHERE t.intake_id = ? AND t.case_id = "+caseIDCast+"))", intakeID, intakeID, intakeID).
		Distinct("c.id").Limit(2).Pluck("c.id", &ids).Error
	return ids, err
}

func rejectExistingIntakeCase(tx *gorm.DB, intakeID string) error {
	ids, err := intakeExistingCaseIDs(tx, intakeID)
	if err != nil {
		return err
	}
	if len(ids) != 0 {
		return intakeCollaborationError("INTAKE_ALREADY_CASE_CREATED", "该接案已生成正式案件，请使用已有案件，不能再次提交成案", 409)
	}
	return nil
}

// Called under the approval transaction lock, before an approving decision
// or case-creation claim. All intake paths arbitrate on the same intake row.
// Existing legacy cases may be verified on the SAME linked case only; legacy
// conflict review tickets cannot silently become frozen case approvals.
func requireIntakeCaseAuthority(tx *gorm.DB, approval *models.ApprovalRequest, allowCompleted bool) error {
	intakeID, err := ResolveApprovalIntakeID(tx, approval)
	if err != nil || intakeID == "" {
		return err
	}
	var intake struct{ ID string }
	if err := tx.Table("case_intakes").Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("id = ?", intakeID).Take(&intake).Error; err != nil {
		return err
	}
	ids, err := intakeExistingCaseIDs(tx, intakeID)
	if err != nil {
		return err
	}
	if len(ids) > 0 {
		linked := strings.TrimSpace(approval.CreatedCaseID)
		if linked == "" && allowCompleted {
			var tracking repositories.CaseCreationTracking
			err := tx.Where("approval_request_id = ? AND creation_status = 'completed'", approval.ID).First(&tracking).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err == nil && tracking.CaseID != nil {
				linked = strings.TrimSpace(*tracking.CaseID)
			}
		}
		if allowCompleted && approval.Status == models.ApprovalStatusApproved && len(ids) == 1 && linked == strconv.FormatUint(uint64(ids[0]), 10) {
			return nil
		}
		return intakeCollaborationError("INTAKE_ALREADY_CASE_CREATED", "该接案已有正式案件，不能通过另一审批重复成案", 409)
	}
	if approval.ActiveSubmissionSnapshotID == nil {
		return intakeCollaborationError("INTAKE_SUBMISSION_REQUIRED", "此旧冲突审批仅保留历史记录，不能同意并成案；请完成独立冲突复核后从原接案工作台提交接案审批", 409)
	}
	snapshot, err := ResolveApprovalSubmission(tx, approval)
	if err != nil {
		return err
	}
	if submissionText(snapshot["intake_id"]) != intakeID {
		return snapshotUnresolved("批准快照与接案持久关联不一致")
	}
	var slot repositories.CaseCreationTracking
	if err := tx.Where("intake_id = ?", intakeID).Take(&slot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewSubjectWorkflowError("INTAKE_APPROVAL_SLOT_REQUIRED", "该审批没有接案唯一绑定，已阻止同意并成案")
		}
		return err
	}
	if slot.ApprovalRequestID != approval.ID {
		return NewSubjectWorkflowError("INTAKE_APPROVAL_ALREADY_IN_PROGRESS", "该接案已有其他持槽审批，不能重复成案")
	}
	return nil
}
