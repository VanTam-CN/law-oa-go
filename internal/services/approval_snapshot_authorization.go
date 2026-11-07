package services

import (
	"context"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// A returned intake may legitimately select a different client. Authorize
// both the old frozen context and the current editable intake; comparing the
// old check's client ID to the new intake would reject every such correction.
func AuthorizeApprovalResubmission(db *gorm.DB, userID string, approval *models.ApprovalRequest) error {
	if approval.ActiveSubmissionSnapshotID == nil || approval.Status != models.ApprovalStatusNeedsRevision {
		return authorizeApprovalSubject(db, userID, approval)
	}
	if approval.ApplicantID != userID {
		return ErrApprovalForbidden
	}
	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		return ErrApprovalForbidden
	}
	auth := NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil)
	actor := AuthActor{UserID: user.ID, Role: user.Role}
	return AuthorizeApprovalRevisionContext(db.Statement.Context, db, auth, actor, approval)
}

// During correction the editable intake and the active frozen submission are
// intentionally different versions. Require membership plus access to BOTH
// objects, rather than asserting that their client IDs must still be equal.
// All other approval states keep the strict check/intake consistency gate.
func AuthorizeApprovalRevisionContext(ctx context.Context, db *gorm.DB, auth *AuthorizationService, actor AuthActor, approval *models.ApprovalRequest) error {
	if approval == nil || approval.ActiveSubmissionSnapshotID == nil || approval.Status != models.ApprovalStatusNeedsRevision || auth == nil {
		return ErrApprovalForbidden
	}
	if _, err := NewApprovalService(db).GetApprovalForAuthorization(strconv.FormatUint(uint64(actor.UserID), 10), approval.ID); err != nil {
		return err
	}
	intakeID, err := ResolveApprovalIntakeID(db, approval)
	if err != nil {
		return err
	}
	ownerID, err := strconv.ParseUint(approval.ApplicantID, 10, 32)
	if err != nil || ownerID == 0 {
		return ErrApprovalForbidden
	}
	allowed, err := auth.CanReadConflictIntakeContext(ctx, actor, intakeID, uint(ownerID), "")
	if err != nil {
		return err
	}
	if !allowed {
		return ErrApprovalForbidden
	}
	snapshot, err := ResolveApprovalSubmission(db, approval)
	if err != nil {
		return err
	}
	return AuthorizeApprovalSnapshotContext(ctx, db, auth, actor, approval, snapshot)
}

// AuthorizeApprovalSnapshotContext is additional to current approval access.
// A new accessible client must not grant access to an old round's client/case.
// Check both frozen references and their persisted check context, including
// walls erected after the snapshot was written. This function never writes.
func AuthorizeApprovalSnapshotContext(ctx context.Context, db *gorm.DB, auth *AuthorizationService, actor AuthActor, approval *models.ApprovalRequest, snapshot map[string]interface{}) error {
	clients, cases, checks := map[uint]bool{}, map[uint]bool{}, map[string]bool{}
	addID := func(set map[uint]bool, value interface{}) error {
		text := submissionText(value)
		if text == "" || text == "0" {
			return nil
		}
		id, err := strconv.ParseUint(text, 10, 32)
		if err != nil || id == 0 {
			return ErrApprovalForbidden
		}
		set[uint(id)] = true
		return nil
	}
	metadata := submissionMap(snapshot["metadata"])
	objects := []map[string]interface{}{snapshot, metadata, submissionMap(snapshot["conflict_record"]), submissionMap(metadata["conflict_record"]), submissionMap(snapshot["case_creation_config"]), submissionMap(metadata["case_creation_config"])}
	for _, object := range objects {
		if err := addID(clients, object["client_id"]); err != nil {
			return err
		}
		if err := addID(clients, submissionMap(object["client"])["id"]); err != nil {
			return err
		}
		for _, source := range []map[string]interface{}{object, submissionMap(object["search_parameters"])} {
			for _, key := range []string{"subjectCaseId", "subject_case_id", "caseId", "case_id"} {
				if err := addID(cases, source[key]); err != nil {
					return err
				}
			}
		}
		for _, key := range []string{"conflict_check_id", "conflict_task_id", "check_id"} {
			if id := submissionText(object[key]); id != "" {
				checks[id] = true
			}
		}
	}
	for checkID := range checks {
		var record models.ConflictCheckRecord
		if err := db.WithContext(ctx).Where("check_id = ?", checkID).First(&record).Error; err != nil {
			return ErrApprovalForbidden
		}
		if err := addID(clients, record.ClientID); err != nil {
			return err
		}
		subject, err := ResolveConflictSubjectContext(ctx, db, &record)
		if err != nil {
			return err
		}
		if subject.CaseID > 0 {
			cases[subject.CaseID] = true
		}
	}
	controlled := approval.ActiveSubmissionSnapshotID != nil || submissionText(metadata["intake_id"]) != "" || strings.TrimSpace(approval.ConflictCheckID) != "" || approval.Type == "case_creation" || approval.Type == "conflict_approval" || approval.Type == "conflict" || approval.Type == "waiver"
	if len(clients)+len(cases) == 0 {
		if controlled {
			return ErrApprovalForbidden
		}
		return nil
	}
	if auth == nil || IsTechnicalAdminRole(actor.Role) || actor.UserID == 0 {
		return ErrApprovalForbidden
	}
	for clientID := range clients {
		var count int64
		if err := db.WithContext(ctx).Model(&models.Client{}).Where("id = ?", clientID).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return ErrApprovalForbidden
		}
		allowed, err := auth.canReadClientWithinEthicalWall(ctx, actor, clientID)
		if err != nil {
			return err
		}
		if !allowed {
			return ErrApprovalForbidden
		}
	}
	for caseID := range cases {
		allowed, err := auth.CanReadConflictContext(ctx, actor, caseID)
		if err != nil {
			return err
		}
		if !allowed {
			return ErrApprovalForbidden
		}
	}
	return nil
}
