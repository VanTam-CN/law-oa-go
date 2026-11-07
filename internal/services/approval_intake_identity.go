package services

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ResolveApprovalIntakeID prevents editable legacy metadata from erasing a
// durable intake association. All available sources must agree; corruption is
// not an invitation to fall back to an uncontrolled workflow.
func ResolveApprovalIntakeID(db *gorm.DB, approval *models.ApprovalRequest) (string, error) {
	if approval == nil {
		return "", ErrApprovalRecordNotFound
	}
	resolved := ""
	accept := func(id string) error {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil
		}
		if resolved != "" && resolved != id {
			return snapshotUnresolved("审批的接案关联存在冲突")
		}
		resolved = id
		return nil
	}
	metadata, err := submissionMetadata(approval.Metadata)
	if err != nil {
		return "", err
	}
	if err := accept(submissionText(metadata["intake_id"])); err != nil {
		return "", err
	}
	if err := accept(submissionText(submissionMap(submissionMap(metadata["approval_snapshot"])["metadata"])["intake_id"])); err != nil {
		return "", err
	}
	if db == nil {
		return resolved, nil
	}
	checks := map[string]bool{}
	if checkID := strings.TrimSpace(approval.ConflictCheckID); checkID != "" {
		checks[checkID] = true
	}
	// Legacy conflict tickets did not carry intake_id in approval metadata.
	// Keep their check and immutable snapshot bindings authoritative even if
	// the editable top-level fields were removed before this protocol shipped.
	var inspect func(map[string]interface{}) error
	inspect = func(object map[string]interface{}) error {
		if object == nil {
			return nil
		}
		if err := accept(submissionText(object["intake_id"])); err != nil {
			return err
		}
		params := submissionMap(object["search_parameters"])
		for _, key := range []string{"intakeId", "intake_id"} {
			if err := accept(submissionText(params[key])); err != nil {
				return err
			}
		}
		for _, key := range []string{"conflict_check_id", "conflict_task_id", "check_id"} {
			if id := submissionText(object[key]); id != "" {
				checks[id] = true
			}
		}
		for _, key := range []string{"metadata", "approval_snapshot", "conflict_record", "conflict_result", "conflict_check"} {
			if err := inspect(submissionMap(object[key])); err != nil {
				return err
			}
		}
		return nil
	}
	if err := inspect(metadata); err != nil {
		return "", err
	}
	if db.Migrator().HasTable(&repositories.CaseCreationTracking{}) {
		var rows []repositories.CaseCreationTracking
		if err := db.Select("intake_id").Where("approval_request_id = ? AND intake_id IS NOT NULL", approval.ID).Find(&rows).Error; err != nil {
			return "", err
		}
		for _, row := range rows {
			if row.IntakeID != nil {
				if err := accept(*row.IntakeID); err != nil {
					return "", err
				}
			}
		}
	}
	if db.Migrator().HasTable(&models.ApprovalSnapshot{}) {
		var rows []models.ApprovalSnapshot
		if err := db.Where("approval_request_id = ?", approval.ID).Find(&rows).Error; err != nil {
			return "", err
		}
		for _, row := range rows {
			if err := inspect(submissionMap(row.SnapshotData)); err != nil {
				return "", err
			}
		}
	}
	if db.Migrator().HasTable(&repositories.ConflictAssociation{}) {
		var ids []string
		if err := db.Model(&repositories.ConflictAssociation{}).Where("approval_request_id = ?", approval.ID).Pluck("conflict_check_id", &ids).Error; err != nil {
			return "", err
		}
		for _, id := range ids {
			if strings.TrimSpace(id) != "" {
				checks[id] = true
			}
		}
	}
	for checkID := range checks {
		var check models.ConflictCheckRecord
		if err := db.Select("check_id, search_parameters").Where("check_id = ?", checkID).First(&check).Error; err != nil {
			return "", err
		}
		if err := accept(SubjectIntakeIDFromSearchParameters(check.SearchParameters)); err != nil {
			return "", err
		}
	}
	return resolved, nil
}

// Every protocol-enabled first submit owns the database intake slot, including
// internal callers that did not enter through the integrated HTTP endpoint.
// Freeze has already locked/validated the intake before this lookup.
func ensureInitialIntakeBinding(tx *gorm.DB, approval *models.ApprovalRequest) error {
	snapshot, err := ResolveApprovalSubmission(tx, approval)
	if err != nil {
		return err
	}
	intakeID := submissionText(snapshot["intake_id"])
	var row repositories.CaseCreationTracking
	err = tx.Where("intake_id = ?", intakeID).Take(&row).Error
	if err == nil {
		if row.ApprovalRequestID == approval.ID {
			return nil
		}
		return &ApprovalConflictGateError{Code: "INTAKE_APPROVAL_ALREADY_IN_PROGRESS", Message: "该接案已有进行中的成案审批"}
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := newTerminalBindingReleaser(tx).rebindAfterResubmitInTx(tx, approval); err != nil {
		if asUniqueConstraintError(err) != nil {
			return &ApprovalConflictGateError{Code: "INTAKE_APPROVAL_ALREADY_IN_PROGRESS", Message: "该接案已有进行中的成案审批"}
		}
		return err
	}
	return nil
}
