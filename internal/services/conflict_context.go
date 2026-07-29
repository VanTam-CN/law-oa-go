package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ConflictSubjectContext identifies the protected subject that owns a
// conflict check. A new intake has no formal case yet, so IntakeID is the
// authoritative context until approval creates the case.
type ConflictSubjectContext struct {
	CaseID   uint
	IntakeID string
}

// SubjectIntakeIDFromSearchParameters reads the immutable intake association
// saved with a conflict-check audit snapshot.
func SubjectIntakeIDFromSearchParameters(parameters models.JSON) string {
	if len(parameters) == 0 {
		return ""
	}
	var values struct {
		IntakeID      string `json:"intakeId"`
		IntakeIDSnake string `json:"intake_id"`
	}
	raw, err := json.Marshal(parameters)
	if err != nil || json.Unmarshal(raw, &values) != nil {
		return ""
	}
	return firstNonEmptyContextValue(values.IntakeID, values.IntakeIDSnake)
}

// ResolveConflictSubjectContext resolves the explicit case/intake association
// without treating a case title, client name, or applicant ID as a substitute
// for an object reference.
func ResolveConflictSubjectContext(ctx context.Context, db *gorm.DB, record *models.ConflictCheckRecord) (ConflictSubjectContext, error) {
	if record == nil {
		return ConflictSubjectContext{}, nil
	}
	result := ConflictSubjectContext{
		CaseID:   subjectCaseIDFromConflictParameters(record.SearchParameters),
		IntakeID: SubjectIntakeIDFromSearchParameters(record.SearchParameters),
	}
	if result.CaseID > 0 || db == nil || strings.TrimSpace(record.CheckID) == "" {
		return result, nil
	}
	if !db.Migrator().HasTable("cases") {
		return result, nil
	}
	var caseIDs []uint
	if err := db.WithContext(ctx).
		Table("cases").
		Where("conflict_check_id = ? AND deleted_at IS NULL", record.CheckID).
		Limit(2).
		Pluck("id", &caseIDs).Error; err != nil {
		return ConflictSubjectContext{}, err
	}
	if len(caseIDs) == 1 {
		result.CaseID = caseIDs[0]
	}
	return result, nil
}

func subjectCaseIDFromConflictParameters(parameters models.JSON) uint {
	if len(parameters) == 0 {
		return 0
	}
	var values struct {
		SubjectCaseID      string `json:"subjectCaseId"`
		SubjectCaseIDSnake string `json:"subject_case_id"`
		CaseID             string `json:"caseId"`
		CaseIDSnake        string `json:"case_id"`
	}
	raw, err := json.Marshal(parameters)
	if err != nil || json.Unmarshal(raw, &values) != nil {
		return 0
	}
	caseIDText := firstNonEmptyContextValue(values.SubjectCaseID, values.SubjectCaseIDSnake, values.CaseID, values.CaseIDSnake)
	caseID, err := strconv.ParseUint(caseIDText, 10, 32)
	if err != nil {
		return 0
	}
	return uint(caseID)
}

func firstNonEmptyContextValue(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && value != "<nil>" && value != "0" {
			return value
		}
	}
	return ""
}

// CanReadConflictIntakeContext applies the same ethical-wall boundary as a
// formal case while the intake is waiting for approval. Ownership alone is
// insufficient: the selected client must also be visible to the actor, and a
// reviewer/manager must be explicitly allowed to inspect that client context.
func (s *AuthorizationService) CanReadConflictIntakeContext(ctx context.Context, actor AuthActor, intakeID string, ownerID uint, clientID string) (bool, error) {
	if s == nil || IsTechnicalAdminRole(actor.Role) || actor.UserID == 0 || strings.TrimSpace(intakeID) == "" {
		return false, nil
	}
	if s.caseRepo == nil || s.caseRepo.GetDB() == nil {
		return false, fmt.Errorf("接案权限服务未初始化")
	}
	db := s.caseRepo.GetDB().WithContext(ctx)
	if !db.Migrator().HasTable("case_intakes") {
		return false, fmt.Errorf("接案表未初始化")
	}
	var intake map[string]interface{}
	if err := db.Table("case_intakes").Select("id, created_by, client_id").Where("id = ?", strings.TrimSpace(intakeID)).Take(&intake).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	actualOwnerID, _ := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(intake["created_by"])), 10, 32)
	actualClientID, _ := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(intake["client_id"])), 10, 32)
	if actualOwnerID == 0 || actualClientID == 0 {
		return false, nil
	}
	if ownerID > 0 && actualOwnerID != uint64(ownerID) {
		return false, nil
	}
	if expectedClientID := strings.TrimSpace(clientID); expectedClientID != "" && expectedClientID != "0" && expectedClientID != strconv.FormatUint(actualClientID, 10) {
		return false, nil
	}

	owner := actualOwnerID == uint64(actor.UserID)
	if !owner && !IsConflictReviewRole(actor.Role) && !IsBusinessMatterManagementRole(actor.Role) {
		return false, nil
	}
	visible, err := s.canReadClientWithinEthicalWall(ctx, actor, uint(actualClientID))
	if err != nil || !visible {
		return false, err
	}
	return true, nil
}
