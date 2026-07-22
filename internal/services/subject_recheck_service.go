package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"law-oa-go/internal/models"
	"law-oa-go/internal/security"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SubjectWorkflowError is returned for business gates that must be visible to
// clients as a 409/403-style decision, rather than as a generic 500 error.
type SubjectWorkflowError struct {
	Code    string
	Message string
}

func (e *SubjectWorkflowError) Error() string { return e.Message }

func NewSubjectWorkflowError(code, message string) *SubjectWorkflowError {
	return &SubjectWorkflowError{Code: code, Message: message}
}

type SubjectRevisionRequest struct {
	ExpectedSubjectVersion int                    `json:"expected_subject_version"`
	ChangeType             string                 `json:"change_type"`
	Payload                map[string]interface{} `json:"payload"`
	Reason                 string                 `json:"reason"`
}

type SubjectRevisionResult struct {
	Revision           *models.CaseSubjectRevision   `json:"revision"`
	CaseSubjectVersion int                           `json:"case_subject_version"`
	CaseSubjectState   string                        `json:"case_subject_state"`
	ConflictCheck      *models.ConflictCheckResponse `json:"conflict_check,omitempty"`
	ActionGateMessage  string                        `json:"action_gate_message,omitempty"`
}

// SubjectRecheckService owns the state transition and action gate. It uses a
// row lock for every transition so two browser tabs cannot approve competing
// subject versions or move an old version back to EFFECTIVE.
type SubjectRecheckService struct {
	db              *gorm.DB
	conflictService ConflictDetectionService
}

func NewSubjectRecheckService(db *gorm.DB, conflictService ConflictDetectionService) *SubjectRecheckService {
	return &SubjectRecheckService{db: db, conflictService: conflictService}
}

func (s *SubjectRecheckService) CreateRevision(ctx context.Context, caseID, actorID uint, actorRole string, req *SubjectRevisionRequest) (*SubjectRevisionResult, error) {
	if s == nil || s.db == nil {
		return nil, NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "案件主体版本服务未初始化，已阻止变更")
	}
	if caseID == 0 || actorID == 0 || req == nil {
		return nil, NewSubjectWorkflowError("INVALID_SUBJECT_REVISION", "案件、操作人和主体变更内容不能为空")
	}
	if req.ExpectedSubjectVersion <= 0 {
		return nil, NewSubjectWorkflowError("SUBJECT_VERSION_REQUIRED", "提交主体变更时必须携带当前主体版本")
	}
	if len(req.Payload) == 0 {
		return nil, NewSubjectWorkflowError("SUBJECT_PAYLOAD_REQUIRED", "主体变更内容不能为空")
	}
	changeType := strings.ToUpper(strings.TrimSpace(req.ChangeType))
	if changeType == "" {
		return nil, NewSubjectWorkflowError("CHANGE_TYPE_REQUIRED", "必须说明主体变更类型")
	}
	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_PAYLOAD_INVALID", "主体变更内容无法保存")
	}
	payload, err = protectSubjectPayload(payload, req.Payload)
	if err != nil {
		return nil, err
	}

	revision := &models.CaseSubjectRevision{}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var caseModel models.Case
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", caseID).First(&caseModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NewSubjectWorkflowError("CASE_NOT_FOUND", "案件不存在")
			}
			return err
		}
		state := strings.ToUpper(strings.TrimSpace(caseModel.SubjectState))
		if state == "" {
			return NewSubjectWorkflowError("RECHECK_REQUIRED", "历史案件主体状态尚未完成初始化，需先补录主体版本并完成冲突复核")
		}
		if state != models.SubjectStateEffective || strings.TrimSpace(caseModel.PendingSubjectRevisionID) != "" {
			return NewSubjectWorkflowError("ACTIVE_RECHECK_EXISTS", "该案件已有待复核的主体变更，请先完成或拒绝现有重检")
		}
		if caseModel.SubjectVersion <= 0 {
			return NewSubjectWorkflowError("RECHECK_REQUIRED", "案件主体版本尚未初始化，需先完成历史主体迁移和冲突复核")
		}
		if caseModel.SubjectVersion != req.ExpectedSubjectVersion {
			return NewSubjectWorkflowError("SUBJECT_VERSION_CONFLICT", "案件主体已被其他人更新，请刷新后重新提交")
		}

		revision = &models.CaseSubjectRevision{
			ID:                 uuid.NewString(),
			CaseID:             caseID,
			BaseSubjectVersion: req.ExpectedSubjectVersion,
			ChangeType:         changeType,
			Status:             models.SubjectStateChangeProposed,
			Payload:            string(payload),
			Reason:             strings.TrimSpace(req.Reason),
			RequestedBy:        actorID,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		if err := tx.Create(revision).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Case{}).Where("id = ?", caseID).Updates(map[string]interface{}{
			"subject_state":               models.SubjectStateRecheckRequired,
			"pending_subject_revision_id": revision.ID,
			"updated_at":                  time.Now(),
		}).Error; err != nil {
			return err
		}
		return s.appendAuditTx(tx, actorID, actorRole, "SUBJECT_CHANGE_PROPOSED", "CASE", strconv.FormatUint(uint64(caseID), 10), models.SubjectStateEffective, models.SubjectStateRecheckRequired, caseModel.SubjectVersion, map[string]interface{}{
			"revision_id": revision.ID,
			"change_type": changeType,
			"reason":      strings.TrimSpace(req.Reason),
		})
	})
	if err != nil {
		return nil, err
	}
	return &SubjectRevisionResult{
		Revision:           revision,
		CaseSubjectVersion: req.ExpectedSubjectVersion,
		CaseSubjectState:   models.SubjectStateRecheckRequired,
		ActionGateMessage:  "主体变更已登记，完成独立复核前已阻断受控动作",
	}, nil
}

func (s *SubjectRecheckService) RunRecheck(ctx context.Context, caseID uint, revisionID string, actorID uint, actorRole string, request *models.ConflictCheckRequest) (*SubjectRevisionResult, error) {
	if s == nil || s.db == nil || s.conflictService == nil {
		return nil, NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "正式重检服务未初始化，已阻止主体变更生效")
	}
	if actorID == 0 {
		return nil, NewSubjectWorkflowError("ACTOR_REQUIRED", "无法识别当前操作人")
	}

	var revision models.CaseSubjectRevision
	var caseModel models.Case
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND case_id = ?", revisionID, caseID).First(&revision).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NewSubjectWorkflowError("REVISION_NOT_FOUND", "主体变更记录不存在")
			}
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", caseID).First(&caseModel).Error; err != nil {
			return err
		}
		if caseModel.PendingSubjectRevisionID != revision.ID || !isRecheckState(revision.Status) {
			return NewSubjectWorkflowError("RECHECK_STATE_INVALID", "该主体变更不在可执行重检状态")
		}
		if strings.TrimSpace(revision.ConflictCheckID) != "" && revision.Status == models.SubjectStateRecheckRunning {
			return NewSubjectWorkflowError("RECHECK_ALREADY_RUNNING", "该主体变更已有重检任务运行中")
		}
		now := time.Now()
		if err := tx.Model(&revision).Updates(map[string]interface{}{
			"status":     models.SubjectStateRecheckRunning,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&models.Case{}).Where("id = ?", caseID).Updates(map[string]interface{}{
			"subject_state": models.SubjectStateRecheckRunning,
			"updated_at":    now,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	// The revision payload is the authoritative subject change. Never trust a
	// caller-provided Parties/OtherParties list here: accepting it would allow a
	// stale browser tab to run a check against the old subjects and then approve
	// a different effective version.
	requestedSearchDepth := "STANDARD"
	if request != nil && strings.TrimSpace(request.SearchDepth) != "" {
		requestedSearchDepth = request.SearchDepth
	}
	request = &models.ConflictCheckRequest{SearchDepth: requestedSearchDepth}
	var payloadValues map[string]interface{}
	if strings.TrimSpace(revision.Payload) != "" {
		payload, payloadErr := unprotectSubjectPayload(revision.Payload)
		if payloadErr != nil {
			_ = s.restoreRecheckRequired(ctx, caseID, revision.ID, actorID, actorRole, payloadErr.Error())
			return nil, payloadErr
		}
		if err := json.Unmarshal(payload, &payloadValues); err != nil {
			_ = s.restoreRecheckRequired(ctx, caseID, revision.ID, actorID, actorRole, err.Error())
			return nil, NewSubjectWorkflowError("SUBJECT_PAYLOAD_INVALID", "主体变更内容无法用于重检")
		}
	}
	parties, partyErr := s.subjectPartiesForRevision(ctx, caseID, payloadValues)
	if partyErr != nil {
		_ = s.restoreRecheckRequired(ctx, caseID, revision.ID, actorID, actorRole, partyErr.Error())
		return nil, partyErr
	}
	request.Parties = parties

	effectiveClientID := caseModel.ClientID
	if clientID, ok := subjectUint(payloadValues, "client_id", "clientId"); ok {
		effectiveClientID = clientID
	}
	var client models.Client
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", effectiveClientID).First(&client).Error; err != nil {
		_ = s.restoreRecheckRequired(ctx, caseID, revision.ID, actorID, actorRole, err.Error())
		return nil, NewSubjectWorkflowError("SUBJECT_CLIENT_NOT_FOUND", "重检使用的客户主体不存在")
	}
	request.CheckID = "RECHECK_" + uuid.NewString()
	request.SubjectCaseID = strconv.FormatUint(uint64(caseID), 10)
	request.SubjectCaseNumber = caseModel.CaseNumber
	request.CaseName = subjectFirstNonEmpty(request.CaseName, caseModel.Title)
	request.CaseType = subjectFirstNonEmpty(request.CaseType, caseModel.CaseType)
	request.ClientID = strconv.FormatUint(uint64(effectiveClientID), 10)
	request.ClientName = client.Name
	request.ClientType = subjectClientType(client.Type)
	if idCard, idErr := client.DecryptedIDCard(); idErr != nil {
		_ = s.restoreRecheckRequired(ctx, caseID, revision.ID, actorID, actorRole, idErr.Error())
		return nil, NewSubjectWorkflowError("SUBJECT_IDENTITY_UNREADABLE", "客户身份标识无法安全读取，已阻止重检")
	} else if strings.TrimSpace(idCard) != "" {
		request.ClientIdentifiers = map[string]string{"id_card": idCard}
	}
	if request.UserID == "" {
		request.UserID = strconv.FormatUint(uint64(caseModel.LawyerID), 10)
	}
	request.SearchYears = 0
	request.SearchDepth = subjectFirstNonEmpty(request.SearchDepth, "STANDARD")
	request.ActorUserID = actorID
	request.ActorRole = actorRole
	request.RequestTime = time.Now()
	result, checkErr := s.conflictService.PerformConflictCheck(ctx, request)
	if checkErr != nil {
		_ = s.restoreRecheckRequired(ctx, caseID, revision.ID, actorID, actorRole, checkErr.Error())
		return nil, checkErr
	}
	if result == nil || strings.TrimSpace(result.CheckID) == "" {
		err := NewSubjectWorkflowError("CONFLICT_RESULT_UNAVAILABLE", "重检未返回可追溯的冲突检测记录，已阻止主体变更生效")
		_ = s.restoreRecheckRequired(ctx, caseID, revision.ID, actorID, actorRole, err.Error())
		return nil, err
	}

	coverageStatus := "COVERAGE_LIMITED"
	if result != nil && result.Decision != nil && result.Decision.CoverageStatus != "" {
		coverageStatus = result.Decision.CoverageStatus
	}
	now := time.Now()
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.CaseSubjectRevision{}).Where("id = ? AND case_id = ?", revision.ID, caseID).Updates(map[string]interface{}{
			"status":            models.SubjectStateRecheckRequired,
			"conflict_check_id": result.CheckID,
			"updated_at":        now,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Case{}).Where("id = ?", caseID).Updates(map[string]interface{}{
			"subject_state":            models.SubjectStateRecheckRequired,
			"conflict_check_id":        result.CheckID,
			"conflict_coverage_status": coverageStatus,
			"updated_at":               now,
		}).Error; err != nil {
			return err
		}
		evidenceCount := 0
		requiresReview := true
		if result.Decision != nil {
			evidenceCount = result.Decision.EvidenceCount
			requiresReview = result.Decision.RequiresManualReview
		}
		return s.appendAuditTx(tx, actorID, actorRole, "SUBJECT_RECHECK_COMPLETED", "CASE_SUBJECT_REVISION", revision.ID, models.SubjectStateRecheckRunning, models.SubjectStateRecheckRequired, caseModel.SubjectVersion, map[string]interface{}{
			"check_id":        result.CheckID,
			"coverage_status": coverageStatus,
			"evidence_count":  evidenceCount,
			"requires_review": requiresReview,
		})
	}); err != nil {
		return nil, err
	}
	revision.Status = models.SubjectStateRecheckRequired
	revision.ConflictCheckID = result.CheckID
	return &SubjectRevisionResult{Revision: &revision, CaseSubjectVersion: caseModel.SubjectVersion, CaseSubjectState: models.SubjectStateRecheckRequired, ConflictCheck: result, ActionGateMessage: "重检已完成，等待独立复核；主体变更尚未生效"}, nil
}

func (s *SubjectRecheckService) ApplyReviewDecision(ctx context.Context, caseID uint, revisionID string, reviewerID uint, reviewerRole, decision, notes string) (*SubjectRevisionResult, error) {
	decision = strings.ToLower(strings.TrimSpace(decision))
	if decision == "" || strings.TrimSpace(notes) == "" {
		return nil, NewSubjectWorkflowError("REVIEW_INPUT_REQUIRED", "复核结论和依据不能为空")
	}
	if !IsConflictReviewRole(reviewerRole) {
		return nil, NewSubjectWorkflowError("REVIEWER_FORBIDDEN", "只有独立冲突核查人或获授权管理人员可以形成复核结论")
	}
	if decision != "no_conflict" && decision != "false_positive" && decision != "confirmed_conflict" && decision != "insufficient_information" && decision != "waiver_requested" {
		return nil, NewSubjectWorkflowError("REVIEW_DECISION_INVALID", "不支持的主体重检复核结论")
	}
	var revision models.CaseSubjectRevision
	var caseModel models.Case
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND case_id = ?", revisionID, caseID).First(&revision).Error; err != nil {
			return NewSubjectWorkflowError("REVISION_NOT_FOUND", "主体变更记录不存在")
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", caseID).First(&caseModel).Error; err != nil {
			return err
		}
		if revision.Status != models.SubjectStateRecheckRequired || caseModel.PendingSubjectRevisionID != revision.ID || revision.ConflictCheckID == "" {
			return NewSubjectWorkflowError("REVIEW_STATE_INVALID", "该主体变更尚未完成重检，不能形成复核结论")
		}
		if err := validateSubjectConflictReviewer(ctx, tx, revision.ConflictCheckID, caseID, reviewerID, reviewerRole); err != nil {
			return err
		}
		if reviewerID == 0 || reviewerID == caseModel.LawyerID || strconv.FormatUint(uint64(reviewerID), 10) == strings.TrimSpace(caseModel.CreatedBy) {
			return NewSubjectWorkflowError("REVIEWER_CONFLICTED", "案件负责人或申请人不得复核自己的主体变更")
		}
		var latestReview models.ConflictReview
		if err := tx.Where("check_id = ?", revision.ConflictCheckID).Order("created_at DESC").First(&latestReview).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NewSubjectWorkflowError("REVIEW_REQUIRED", "请先提交独立冲突复核结论")
			}
			return err
		}
		if latestReview.ReviewerID != reviewerID || strings.ToLower(strings.TrimSpace(latestReview.Decision)) != decision {
			return NewSubjectWorkflowError("REVIEW_MISMATCH", "主体变更复核人与冲突复核记录不一致")
		}
		var record models.ConflictCheckRecord
		if err := tx.Where("check_id = ?", revision.ConflictCheckID).First(&record).Error; err != nil {
			return NewSubjectWorkflowError("REVIEW_REQUIRED", "关联的冲突检测记录不存在")
		}
		var response models.ConflictCheckResponse
		raw, _ := json.Marshal(record.CheckResult)
		if err := json.Unmarshal(raw, &response); err != nil || response.Decision == nil || strings.ToUpper(strings.TrimSpace(response.Decision.CoverageStatus)) != "COMPLETE" {
			return NewSubjectWorkflowError("COVERAGE_LIMITED", "冲突检索范围尚未由律所确认完整，不能批准主体变更")
		}
		if err := validateConflictReviewEvidence(&latestReview, &response); err != nil {
			return err
		}
		now := time.Now()
		updateRevision := map[string]interface{}{
			"reviewed_by":     reviewerID,
			"review_decision": decision,
			"review_notes":    strings.TrimSpace(notes),
			"updated_at":      now,
		}
		updateCase := map[string]interface{}{"updated_at": now}
		fromState := models.SubjectStateRecheckRequired
		toState := models.SubjectStateRecheckRequired
		if decision == "no_conflict" || decision == "false_positive" {
			fieldUpdates, err := subjectRevisionCaseUpdates(tx, revision)
			if err != nil {
				return err
			}
			for key, value := range fieldUpdates {
				updateCase[key] = value
			}
			updateRevision["status"] = models.SubjectStateChangeApprovedAndEffective
			updateRevision["effective_at"] = now
			updateCase["subject_state"] = models.SubjectStateEffective
			updateCase["subject_version"] = gorm.Expr("subject_version + ?", 1)
			updateCase["subject_snapshot"] = revision.Payload
			updateCase["pending_subject_revision_id"] = nil
			fromState = models.SubjectStateRecheckRequired
			toState = models.SubjectStateEffective
		} else if decision == "confirmed_conflict" {
			updateRevision["status"] = models.SubjectStateChangeRejected
			updateCase["subject_state"] = models.SubjectStateEffective
			updateCase["pending_subject_revision_id"] = nil
			fromState = models.SubjectStateRecheckRequired
			toState = models.SubjectStateEffective
		}
		if err := tx.Model(&models.CaseSubjectRevision{}).Where("id = ?", revision.ID).Updates(updateRevision).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Case{}).Where("id = ?", caseID).Updates(updateCase).Error; err != nil {
			return err
		}
		return s.appendAuditTx(tx, reviewerID, reviewerRole, "SUBJECT_REVIEW_DECISION", "CASE_SUBJECT_REVISION", revision.ID, fromState, toState, caseModel.SubjectVersion, map[string]interface{}{
			"check_id": revision.ConflictCheckID,
			"decision": decision,
			"notes":    strings.TrimSpace(notes),
		})
	})
	if err != nil {
		return nil, err
	}
	revision.Status = revisionStatusAfterDecision(decision)
	if decision == "no_conflict" || decision == "false_positive" {
		revision.EffectiveAt = subjectTimePtr(time.Now())
	}
	revision.ReviewedBy = &reviewerID
	revision.ReviewDecision = decision
	revision.ReviewNotes = strings.TrimSpace(notes)
	if decision == "no_conflict" || decision == "false_positive" {
		caseModel.SubjectVersion++
		caseModel.SubjectState = models.SubjectStateEffective
		caseModel.PendingSubjectRevisionID = ""
	} else if decision == "confirmed_conflict" {
		caseModel.SubjectState = models.SubjectStateEffective
		caseModel.PendingSubjectRevisionID = ""
	}
	return &SubjectRevisionResult{Revision: &revision, CaseSubjectVersion: caseModel.SubjectVersion, CaseSubjectState: caseModel.SubjectState, ActionGateMessage: actionGateMessage(caseModel.SubjectState)}, nil
}

// ReviewAndApply records the independent review and applies the subject
// transition in one database transaction. Keeping these writes atomic avoids
// a review row being left behind when the case state transition fails.
func (s *SubjectRecheckService) ReviewAndApply(ctx context.Context, caseID uint, revisionID, checkID string, reviewerID uint, reviewerRole, reviewerName, decision, notes string, nextReviewAt *time.Time) (*SubjectRevisionResult, error) {
	decision = strings.ToLower(strings.TrimSpace(decision))
	if !allowedConflictReviewDecisions[decision] || strings.TrimSpace(notes) == "" {
		return nil, NewSubjectWorkflowError("REVIEW_INPUT_REQUIRED", "复核结论和依据不能为空或不受支持")
	}
	if s == nil || s.db == nil {
		return nil, NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "主体版本服务未初始化，已阻止复核")
	}
	if reviewerID == 0 {
		return nil, NewSubjectWorkflowError("REVIEWER_REQUIRED", "无法识别独立复核人")
	}
	if !IsConflictReviewRole(reviewerRole) {
		return nil, NewSubjectWorkflowError("REVIEWER_FORBIDDEN", "只有独立冲突核查人或获授权管理人员可以形成复核结论")
	}

	var revision models.CaseSubjectRevision
	var caseModel models.Case
	var review models.ConflictReview
	now := time.Now()
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND case_id = ?", revisionID, caseID).First(&revision).Error; err != nil {
			return NewSubjectWorkflowError("REVISION_NOT_FOUND", "主体变更记录不存在")
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND deleted_at IS NULL", caseID).First(&caseModel).Error; err != nil {
			return err
		}
		if revision.Status != models.SubjectStateRecheckRequired || caseModel.PendingSubjectRevisionID != revision.ID || revision.ConflictCheckID == "" {
			return NewSubjectWorkflowError("REVIEW_STATE_INVALID", "该主体变更尚未完成重检，不能形成复核结论")
		}
		if strings.TrimSpace(checkID) == "" || strings.TrimSpace(checkID) != strings.TrimSpace(revision.ConflictCheckID) {
			return NewSubjectWorkflowError("REVIEW_MISMATCH", "复核记录与主体变更关联的检测记录不一致")
		}
		if reviewerID == caseModel.LawyerID || strconv.FormatUint(uint64(reviewerID), 10) == strings.TrimSpace(caseModel.CreatedBy) {
			return NewSubjectWorkflowError("REVIEWER_CONFLICTED", "案件负责人或申请人不得复核自己的主体变更")
		}

		var record models.ConflictCheckRecord
		if err := tx.Where("check_id = ?", revision.ConflictCheckID).First(&record).Error; err != nil {
			return NewSubjectWorkflowError("REVIEW_REQUIRED", "关联的冲突检测记录不存在")
		}
		if record.UserID == reviewerID {
			return NewSubjectWorkflowError("REVIEWER_CONFLICTED", "冲突检测申请人不得复核本人记录")
		}
		if err := validateSubjectConflictReviewer(ctx, tx, revision.ConflictCheckID, caseID, reviewerID, reviewerRole); err != nil {
			return err
		}
		var existingReview models.ConflictReview
		if err := tx.Where("check_id = ?", revision.ConflictCheckID).Order("created_at DESC").First(&existingReview).Error; err == nil {
			return NewSubjectWorkflowError("REVIEW_ALREADY_EXISTS", "该检测已有复核结论，不能重复提交；如主体仍有变化，请重新发起主体重检")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var response models.ConflictCheckResponse
		raw, _ := json.Marshal(record.CheckResult)
		if err := json.Unmarshal(raw, &response); err != nil {
			return NewSubjectWorkflowError("REVIEW_EVIDENCE_INVALID", "检测证据无法解析，已阻止复核")
		}
		if response.Decision == nil || strings.ToUpper(strings.TrimSpace(response.Decision.CoverageStatus)) != "COMPLETE" {
			return NewSubjectWorkflowError("COVERAGE_LIMITED", "冲突检索范围尚未由律所确认完整，不能批准主体变更")
		}
		review = models.ConflictReview{
			CheckID: revision.ConflictCheckID, Decision: decision, Notes: strings.TrimSpace(notes),
			ReviewerID: reviewerID, ReviewerName: subjectFirstNonEmpty(reviewerName, fmt.Sprintf("用户%d", reviewerID)),
			EvidenceHash: conflictEvidenceHash(&response), NextReviewAt: nextReviewAt, CreatedAt: now,
		}
		if err := tx.Create(&review).Error; err != nil {
			return err
		}
		// Keep the frozen check evidence immutable. The review row is the
		// append-only source for the professional conclusion.

		updateRevision := map[string]interface{}{
			"reviewed_by": reviewerID, "review_decision": decision, "review_notes": strings.TrimSpace(notes), "updated_at": now,
		}
		updateCase := map[string]interface{}{"updated_at": now}
		toState := models.SubjectStateRecheckRequired
		if decision == "no_conflict" || decision == "false_positive" {
			fieldUpdates, err := subjectRevisionCaseUpdates(tx, revision)
			if err != nil {
				return err
			}
			for key, value := range fieldUpdates {
				updateCase[key] = value
			}
			updateRevision["status"] = models.SubjectStateChangeApprovedAndEffective
			updateRevision["effective_at"] = now
			updateCase["subject_state"] = models.SubjectStateEffective
			updateCase["subject_version"] = gorm.Expr("subject_version + ?", 1)
			updateCase["subject_snapshot"] = revision.Payload
			updateCase["pending_subject_revision_id"] = nil
			toState = models.SubjectStateEffective
		} else if decision == "confirmed_conflict" {
			updateRevision["status"] = models.SubjectStateChangeRejected
			updateCase["subject_state"] = models.SubjectStateEffective
			updateCase["pending_subject_revision_id"] = nil
			toState = models.SubjectStateEffective
		}
		if err := tx.Model(&models.CaseSubjectRevision{}).Where("id = ?", revision.ID).Updates(updateRevision).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Case{}).Where("id = ?", caseID).Updates(updateCase).Error; err != nil {
			return err
		}
		return s.appendAuditTx(tx, reviewerID, reviewerRole, "SUBJECT_REVIEW_DECISION", "CASE_SUBJECT_REVISION", revision.ID, models.SubjectStateRecheckRequired, toState, caseModel.SubjectVersion, map[string]interface{}{
			"check_id": revision.ConflictCheckID, "decision": decision, "notes": strings.TrimSpace(notes), "review_id": review.ID,
		})
	})
	if err != nil {
		return nil, err
	}

	revision.Status = revisionStatusAfterDecision(decision)
	revision.ReviewedBy = &reviewerID
	revision.ReviewDecision = decision
	revision.ReviewNotes = strings.TrimSpace(notes)
	if decision == "no_conflict" || decision == "false_positive" {
		revision.EffectiveAt = &now
		caseModel.SubjectVersion++
		caseModel.SubjectState = models.SubjectStateEffective
		caseModel.PendingSubjectRevisionID = ""
	} else if decision == "confirmed_conflict" {
		caseModel.SubjectState = models.SubjectStateEffective
		caseModel.PendingSubjectRevisionID = ""
	}
	return &SubjectRevisionResult{Revision: &revision, CaseSubjectVersion: caseModel.SubjectVersion, CaseSubjectState: caseModel.SubjectState, ActionGateMessage: actionGateMessage(caseModel.SubjectState)}, nil
}

func validateSubjectConflictReviewer(ctx context.Context, tx *gorm.DB, checkID string, caseID, reviewerID uint, reviewerRole string) error {
	if err := ValidateConflictReviewer(ctx, tx, checkID, caseID, reviewerID, reviewerRole); err != nil {
		var reviewerErr *ConflictReviewerError
		if errors.As(err, &reviewerErr) {
			return NewSubjectWorkflowError(reviewerErr.Code, reviewerErr.Message)
		}
		return NewSubjectWorkflowError("REVIEWER_GATE_UNAVAILABLE", "独立复核资格校验失败，已阻止复核")
	}
	return nil
}

func (s *SubjectRecheckService) RequireEffectiveSubject(ctx context.Context, caseID uint, action string) error {
	if s == nil || s.db == nil {
		return NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "案件主体版本服务未初始化，已阻止受控动作")
	}
	var caseModel models.Case
	if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", caseID).First(&caseModel).Error; err != nil {
		return NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "无法读取案件主体版本，已阻止受控动作")
	}
	state := strings.ToUpper(strings.TrimSpace(caseModel.SubjectState))
	if state == "" {
		return NewSubjectWorkflowError("RECHECK_REQUIRED", fmt.Sprintf("历史案件主体状态尚未完成初始化，已阻止：%s", action))
	}
	if caseModel.SubjectVersion <= 0 {
		return NewSubjectWorkflowError("RECHECK_REQUIRED", fmt.Sprintf("案件主体版本尚未初始化，已阻止：%s", action))
	}
	if state != models.SubjectStateEffective || strings.TrimSpace(caseModel.PendingSubjectRevisionID) != "" {
		return NewSubjectWorkflowError("RECHECK_REQUIRED", fmt.Sprintf("案件主体变更尚未完成独立复核，已阻止：%s", action))
	}
	if strings.ToUpper(strings.TrimSpace(caseModel.ConflictCoverageStatus)) != "COMPLETE" {
		return NewSubjectWorkflowError("COVERAGE_LIMITED", fmt.Sprintf("案件冲突检索范围尚未完成确认，已阻止：%s", action))
	}
	if strings.TrimSpace(caseModel.ConflictCheckID) == "" {
		return NewSubjectWorkflowError("CONFLICT_CHECK_REQUIRED", fmt.Sprintf("案件尚未关联可用的冲突复核记录，已阻止：%s", action))
	}
	if err := s.RequireConflictDisposition(ctx, caseModel.ConflictCheckID, action); err != nil {
		return err
	}
	return nil
}

// RequireConflictDisposition is the final gate used immediately before an
// approved case is created. A machine result, a missing scope, or a pending
// review can never be treated as permission to create a live case.
func (s *SubjectRecheckService) RequireConflictDisposition(ctx context.Context, checkID, action string) error {
	if s == nil || s.db == nil {
		return NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "冲突处置门禁未初始化，已阻止受控动作")
	}
	var record models.ConflictCheckRecord
	if err := s.db.WithContext(ctx).Where("check_id = ?", strings.TrimSpace(checkID)).First(&record).Error; err != nil {
		return NewSubjectWorkflowError("CONFLICT_CHECK_REQUIRED", "未找到可用于成案的冲突检测记录")
	}
	if strings.ToUpper(strings.TrimSpace(record.CheckStatus)) != "COMPLETED" {
		return NewSubjectWorkflowError("RECHECK_REQUIRED", "冲突检测尚未完成，已阻止受控动作")
	}
	decision := mapValue(record.CheckResult, "decision")
	coverageStatus := strings.ToUpper(subjectFirstNonEmpty(mapString(decision, "coverageStatus"), mapString(decision, "coverage_status")))
	if coverageStatus != "COMPLETE" {
		return NewSubjectWorkflowError("COVERAGE_LIMITED", "冲突检索范围尚未由律所确认完整，已阻止受控动作")
	}
	var response models.ConflictCheckResponse
	raw, _ := json.Marshal(record.CheckResult)
	if err := json.Unmarshal(raw, &response); err != nil {
		return NewSubjectWorkflowError("REVIEW_EVIDENCE_INVALID", "冲突检测证据无法解析，已阻止受控动作")
	}
	var review models.ConflictReview
	if err := s.db.WithContext(ctx).Where("check_id = ?", checkID).Order("created_at DESC").First(&review).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewSubjectWorkflowError("REVIEW_REQUIRED", "冲突检测尚未完成独立人工复核，已阻止受控动作")
		}
		return NewSubjectWorkflowError("REVIEW_REQUIRED", "无法读取独立复核记录，已阻止受控动作")
	}
	if review.ReviewerID == record.UserID {
		return NewSubjectWorkflowError("REVIEWER_CONFLICTED", "冲突检测申请人不得复核本人记录")
	}
	if err := validateConflictReviewEvidence(&review, &response); err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(review.Decision)) {
	case "no_conflict", "false_positive":
		return nil
	case "waiver_requested":
		if err := s.requireCurrentApprovedWaiver(ctx, checkID, &record); err == nil {
			return nil
		} else if workflowErr, ok := err.(*SubjectWorkflowError); ok && workflowErr.Code != "WAIVER_PENDING" {
			return workflowErr
		}
		return NewSubjectWorkflowError("WAIVER_PENDING", fmt.Sprintf("豁免处置尚未生效，已阻止：%s", action))
	case "confirmed_conflict":
		return NewSubjectWorkflowError("CONFLICT_CONFIRMED", fmt.Sprintf("已确认存在利益冲突，已阻止：%s", action))
	default:
		return NewSubjectWorkflowError("REVIEW_REQUIRED", fmt.Sprintf("冲突复核尚未形成可继续结论，已阻止：%s", action))
	}
}

// RequireConflictDispositionForCase prevents a valid review from being
// replayed for a different client or lawyer. The check record is the
// authoritative binding created by the conflict workflow; callers must not
// be able to swap case subjects after the independent review.
func (s *SubjectRecheckService) RequireConflictDispositionForCase(ctx context.Context, checkID string, clientID, lawyerID uint, action string) error {
	if clientID == 0 || lawyerID == 0 {
		return NewSubjectWorkflowError("CASE_SUBJECT_BINDING_REQUIRED", "正式成案必须绑定客户和承办律师后才能校验冲突复核记录")
	}
	if s == nil || s.db == nil {
		return NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "冲突处置门禁未初始化，已阻止受控动作")
	}

	var record models.ConflictCheckRecord
	if err := s.db.WithContext(ctx).Where("check_id = ?", strings.TrimSpace(checkID)).First(&record).Error; err != nil {
		return NewSubjectWorkflowError("CONFLICT_CHECK_REQUIRED", "未找到可用于成案的冲突检测记录")
	}
	if strings.TrimSpace(record.ClientID) != strconv.FormatUint(uint64(clientID), 10) {
		return NewSubjectWorkflowError("SUBJECT_CLIENT_MISMATCH", fmt.Sprintf("冲突复核记录与正式成案客户不一致，已阻止：%s", action))
	}
	if record.UserID != lawyerID {
		return NewSubjectWorkflowError("SUBJECT_LAWYER_MISMATCH", fmt.Sprintf("冲突复核记录与正式成案承办律师不一致，已阻止：%s", action))
	}
	return s.RequireConflictDisposition(ctx, checkID, action)
}

// requireCurrentApprovedWaiver is deliberately part of the final action gate,
// rather than relying on a prior GET /waiver call to refresh the denormalized
// conflict result. This prevents a stale WAIVED JSON status from surviving
// an expired or overdue waiver when a caller invokes case creation directly.
func (s *SubjectRecheckService) requireCurrentApprovedWaiver(ctx context.Context, checkID string, record *models.ConflictCheckRecord) error {
	if s == nil || s.db == nil {
		return NewSubjectWorkflowError("WAIVER_PENDING", "豁免状态服务不可用，已阻止受控动作")
	}
	if record != nil && conflictRecordHasNonWaivableDirectConflict(record) {
		return NewSubjectWorkflowError("CONFLICT_CONFIRMED", "该检测包含不可豁免的直接利益冲突，已阻止受控动作")
	}
	var application struct {
		ID                  string     `gorm:"column:id"`
		Status              string     `gorm:"column:status"`
		RequestedExpiryDate *time.Time `gorm:"column:requested_expiry_date"`
	}
	if err := s.db.WithContext(ctx).
		Table("waiver_applications").
		Where("conflict_check_id = ? AND deleted_at IS NULL", strings.TrimSpace(checkID)).
		Order("updated_at DESC").First(&application).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewSubjectWorkflowError("WAIVER_PENDING", "豁免尚未批准，已阻止受控动作")
		}
		return NewSubjectWorkflowError("WAIVER_PENDING", "无法读取豁免状态，已阻止受控动作")
	}
	if strings.ToUpper(strings.TrimSpace(application.Status)) != WaiverStatusApproved {
		if strings.EqualFold(strings.TrimSpace(application.Status), WaiverStatusReviewOverdue) {
			return NewSubjectWorkflowError("WAIVER_REVIEW_OVERDUE", "豁免年度复核已逾期，已阻止受控动作")
		}
		return NewSubjectWorkflowError("WAIVER_PENDING", "豁免尚未批准，已阻止受控动作")
	}
	now := time.Now()
	if application.RequestedExpiryDate != nil && !application.RequestedExpiryDate.After(now) {
		return NewSubjectWorkflowError("WAIVER_EXPIRED", "豁免已到期，已阻止受控动作")
	}
	var approval struct {
		Decision       string     `gorm:"column:decision"`
		Status         string     `gorm:"column:status"`
		ExpiryDate     *time.Time `gorm:"column:expiry_date"`
		NextReviewDate *time.Time `gorm:"column:next_review_date"`
	}
	if err := s.db.WithContext(ctx).
		Table("waiver_approval_records").
		Where("waiver_application_id = ? AND deleted_at IS NULL", application.ID).
		Order("approval_date DESC").First(&approval).Error; err != nil {
		return NewSubjectWorkflowError("WAIVER_REVIEW_OVERDUE", "豁免缺少有效审批记录，已阻止受控动作")
	}
	if strings.ToUpper(strings.TrimSpace(approval.Decision)) != WaiverDecisionApprove || strings.ToUpper(strings.TrimSpace(approval.Status)) != "ACTIVE" {
		return NewSubjectWorkflowError("WAIVER_PENDING", "豁免审批记录未生效，已阻止受控动作")
	}
	if approval.ExpiryDate != nil && !approval.ExpiryDate.After(now) {
		return NewSubjectWorkflowError("WAIVER_EXPIRED", "豁免已到期，已阻止受控动作")
	}
	if approval.NextReviewDate == nil || !approval.NextReviewDate.After(now) {
		return NewSubjectWorkflowError("WAIVER_REVIEW_OVERDUE", "豁免年度复核已逾期，已阻止受控动作")
	}
	return nil
}

// validateConflictReviewEvidence prevents an old professional conclusion from
// being reused after the underlying machine result has changed. The review
// hash covers only the evidence snapshot, so the review summary itself does
// not create a circular hash dependency.
func validateConflictReviewEvidence(review *models.ConflictReview, response *models.ConflictCheckResponse) error {
	if review == nil || response == nil {
		return NewSubjectWorkflowError("REVIEW_EVIDENCE_INVALID", "冲突复核证据不完整，已阻止受控动作")
	}
	expected := conflictEvidenceHash(response)
	actual := strings.TrimSpace(review.EvidenceHash)
	if actual == "" || !strings.EqualFold(actual, expected) {
		return NewSubjectWorkflowError("REVIEW_EVIDENCE_STALE", "冲突检测证据已变化，原复核结论失效，请重新进行独立复核")
	}
	return nil
}

func (s *SubjectRecheckService) restoreRecheckRequired(ctx context.Context, caseID uint, revisionID string, actorID uint, actorRole, reason string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.CaseSubjectRevision{}).Where("id = ? AND case_id = ?", revisionID, caseID).Updates(map[string]interface{}{"status": models.SubjectStateRecheckRequired, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Case{}).Where("id = ?", caseID).Updates(map[string]interface{}{"subject_state": models.SubjectStateRecheckRequired, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return s.appendAuditTx(tx, actorID, actorRole, "SUBJECT_RECHECK_FAILED", "CASE_SUBJECT_REVISION", revisionID, models.SubjectStateRecheckRunning, models.SubjectStateRecheckRequired, 0, map[string]interface{}{"reason": reason})
	})
}

func (s *SubjectRecheckService) appendAuditTx(tx *gorm.DB, actorID uint, actorRole, eventType, objectType, objectID, fromState, toState string, subjectVersion int, payload map[string]interface{}) error {
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	actor := actorID
	event := &models.ComplianceAuditEvent{
		ID: uuid.NewString(), ActorID: &actor, ActorRole: actorRole, EventType: eventType,
		ObjectType: objectType, ObjectID: objectID, FromState: fromState, ToState: toState,
		SubjectVersion: subjectVersion, Payload: string(raw), IntegrityHash: hex.EncodeToString(sum[:]), CreatedAt: time.Now(),
	}
	return tx.Create(event).Error
}

func subjectRevisionCaseUpdates(tx *gorm.DB, revision models.CaseSubjectRevision) (map[string]interface{}, error) {
	payload, err := unprotectSubjectPayload(revision.Payload)
	if err != nil {
		return nil, err
	}
	var values map[string]interface{}
	if err := json.Unmarshal(payload, &values); err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_PAYLOAD_INVALID", "主体变更内容无法应用到案件")
	}
	updates := map[string]interface{}{}
	if clientID, ok := subjectUint(values, "client_id", "clientId"); ok {
		var client models.Client
		if err := tx.Where("id = ? AND deleted_at IS NULL", clientID).First(&client).Error; err != nil {
			return nil, NewSubjectWorkflowError("SUBJECT_CLIENT_NOT_FOUND", "主体变更指定的客户不存在，不能生效")
		}
		updates["client_id"] = clientID
	}
	if lawyerID, ok := subjectUint(values, "lawyer_id", "lawyerId", "lead_lawyer_id", "leadLawyerId"); ok {
		var lawyer models.User
		if err := tx.Where("id = ? AND deleted_at IS NULL", lawyerID).First(&lawyer).Error; err != nil {
			return nil, NewSubjectWorkflowError("SUBJECT_LAWYER_NOT_FOUND", "主体变更指定的承办律师不存在，不能生效")
		}
		updates["lawyer_id"] = lawyerID
	}
	if err := applySubjectPartyUpdates(tx, revision, values); err != nil {
		return nil, err
	}
	return updates, nil
}

// applySubjectPartyUpdates keeps the effective case party relation in sync
// with an approved revision. A free-text party name is sufficient for a
// search request, but it is not sufficient to make a formal case subject
// effective; that path fails closed until an entity reference is supplied.
func applySubjectPartyUpdates(tx *gorm.DB, revision models.CaseSubjectRevision, values map[string]interface{}) error {
	if raw, ok := firstSubjectValue(values, "case_parties", "caseParties"); ok {
		parties, valid := subjectMapSlice(raw)
		if !valid {
			return NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "案件当事人变更格式无效")
		}
		records, err := buildCasePartyRecords(tx, revision.CaseID, parties)
		if err != nil {
			return err
		}
		if err := tx.Where("case_id = ? AND deleted_at IS NULL", revision.CaseID).Delete(&models.CaseParty{}).Error; err != nil {
			return NewSubjectWorkflowError("SUBJECT_PARTY_UPDATE_FAILED", "无法替换案件当事人关系，主体变更未生效")
		}
		for _, party := range records {
			if err := tx.Create(&party).Error; err != nil {
				return NewSubjectWorkflowError("SUBJECT_PARTY_UPDATE_FAILED", "无法写入案件当事人关系，主体变更未生效")
			}
		}
		return nil
	}

	if raw, ok := firstSubjectValue(values, "add_parties", "addParties"); ok {
		parties, valid := subjectMapSlice(raw)
		if !valid {
			return NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "新增案件当事人格式无效")
		}
		records, err := buildCasePartyRecords(tx, revision.CaseID, parties)
		if err != nil {
			return err
		}
		for _, party := range records {
			if err := tx.Where("case_id = ? AND entity_id = ? AND deleted_at IS NULL", revision.CaseID, party.EntityID).First(&models.CaseParty{}).Error; err == nil {
				continue
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return NewSubjectWorkflowError("SUBJECT_PARTY_UPDATE_FAILED", "无法校验重复案件当事人，主体变更未生效")
			}
			if err := tx.Create(&party).Error; err != nil {
				return NewSubjectWorkflowError("SUBJECT_PARTY_UPDATE_FAILED", "无法新增案件当事人关系，主体变更未生效")
			}
		}
		return nil
	}

	if raw, ok := firstSubjectValue(values, "remove_party_ids", "removePartyIds"); ok {
		ids, valid := subjectUintSlice(raw)
		if !valid || len(ids) == 0 {
			return NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "移除案件当事人格式无效")
		}
		if err := tx.Where("case_id = ? AND entity_id IN ? AND deleted_at IS NULL", revision.CaseID, ids).Delete(&models.CaseParty{}).Error; err != nil {
			return NewSubjectWorkflowError("SUBJECT_PARTY_UPDATE_FAILED", "无法移除案件当事人关系，主体变更未生效")
		}
		return nil
	}

	// Do not allow a free-text or relationship-only payload to advance the
	// effective version. It would make the next check use stale case parties.
	if hasSubjectKey(values, "parties", "related_parties", "relatedParties", "related_entity_ids", "relatedEntityIds") {
		return NewSubjectWorkflowError("SUBJECT_PARTY_REFERENCE_REQUIRED", "主体变更必须引用已登记的结构化主体，不能只提交名称或关系描述")
	}
	return nil
}

func firstSubjectValue(values map[string]interface{}, keys ...string) (interface{}, bool) {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value, true
		}
	}
	return nil, false
}

func hasSubjectKey(values map[string]interface{}, keys ...string) bool {
	_, ok := firstSubjectValue(values, keys...)
	return ok
}

func subjectMapSlice(value interface{}) ([]map[string]interface{}, bool) {
	items, ok := value.([]interface{})
	if !ok {
		return nil, false
	}
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		mapped, ok := item.(map[string]interface{})
		if !ok {
			return nil, false
		}
		result = append(result, mapped)
	}
	return result, true
}

func buildCasePartyRecords(tx *gorm.DB, caseID uint, values []map[string]interface{}) ([]models.CaseParty, error) {
	if len(values) == 0 {
		return nil, NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "案件当事人列表不能为空")
	}
	records := make([]models.CaseParty, 0, len(values))
	seen := make(map[uint]struct{}, len(values))
	for index, value := range values {
		entityID, ok := subjectUint(value, "entity_id", "entityId")
		if !ok {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_REFERENCE_REQUIRED", fmt.Sprintf("第%d个案件当事人缺少结构化主体ID", index+1))
		}
		if _, exists := seen[entityID]; exists {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "案件当事人列表包含重复主体")
		}
		seen[entityID] = struct{}{}
		var entity models.Entity
		if err := tx.Where("id = ? AND deleted_at IS NULL", entityID).First(&entity).Error; err != nil {
			return nil, NewSubjectWorkflowError("SUBJECT_ENTITY_NOT_FOUND", "主体变更引用的结构化主体不存在")
		}
		role := strings.ToUpper(strings.TrimSpace(subjectString(value, "role")))
		partyType := strings.ToUpper(strings.TrimSpace(subjectString(value, "party_type", "partyType")))
		if role == "" || partyType == "" {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "案件当事人必须同时提供角色和主体类型")
		}
		records = append(records, models.CaseParty{
			CaseID:       caseID,
			EntityID:     entityID,
			Role:         models.PartyRole(role),
			PartyType:    models.PartyType(partyType),
			Description:  strings.TrimSpace(subjectString(value, "description")),
			DisplayOrder: index,
		})
	}
	return records, nil
}

// subjectPartiesForRevision builds the exact party set that the recheck must
// inspect. It starts from the currently effective case parties, then applies
// the pending replacement/addition/removal from the encrypted revision. This
// keeps the search input and the later effective version in lockstep.
func (s *SubjectRecheckService) subjectPartiesForRevision(ctx context.Context, caseID uint, values map[string]interface{}) ([]models.ConflictPartyInfo, error) {
	parties := make([]models.ConflictPartyInfo, 0)
	var removedIDs map[uint]struct{}
	if raw, ok := firstSubjectValue(values, "remove_party_ids", "removePartyIds"); ok {
		ids, valid := subjectUintSlice(raw)
		if !valid || len(ids) == 0 {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "移除案件当事人格式无效")
		}
		removedIDs = make(map[uint]struct{}, len(ids))
		for _, id := range ids {
			removedIDs[id] = struct{}{}
		}
	}
	if raw, ok := firstSubjectValue(values, "case_parties", "caseParties"); ok {
		if len(removedIDs) > 0 {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "替换当事人与移除当事人不能在同一次变更中混用")
		}
		items, valid := subjectMapSlice(raw)
		if !valid {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "案件当事人格式无法用于重检")
		}
		resolved, err := s.resolveSubjectPartyValues(ctx, items)
		if err != nil {
			return nil, err
		}
		parties = append(parties, resolved...)
	} else {
		var current []models.CaseParty
		if err := s.db.WithContext(ctx).Preload("Entity").Where("case_id = ? AND deleted_at IS NULL", caseID).Order("display_order ASC").Find(&current).Error; err != nil {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_READ_FAILED", "无法读取案件当前当事人，已阻止重检")
		}
		for _, party := range current {
			if _, removed := removedIDs[party.EntityID]; removed {
				continue
			}
			if party.Entity.ID == 0 || strings.TrimSpace(party.Entity.Name) == "" {
				return nil, NewSubjectWorkflowError("SUBJECT_PARTY_REFERENCE_REQUIRED", "案件已有当事人缺少结构化主体，不能进行主体重检")
			}
			partyInfo, err := conflictPartyInfoFromEntity(party.Entity, string(party.Role))
			if err != nil {
				return nil, err
			}
			parties = append(parties, partyInfo)
		}
	}

	if raw, ok := firstSubjectValue(values, "add_parties", "addParties"); ok {
		items, valid := subjectMapSlice(raw)
		if !valid {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "新增案件当事人格式无效")
		}
		resolved, err := s.resolveSubjectPartyValues(ctx, items)
		if err != nil {
			return nil, err
		}
		parties = append(parties, resolved...)
	}

	return parties, nil
}

func (s *SubjectRecheckService) resolveSubjectPartyValues(ctx context.Context, items []map[string]interface{}) ([]models.ConflictPartyInfo, error) {
	parties := make([]models.ConflictPartyInfo, 0, len(items))
	seen := make(map[uint]struct{}, len(items))
	for _, item := range items {
		entityID, ok := subjectUint(item, "entity_id", "entityId")
		if !ok {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_REFERENCE_REQUIRED", "重检主体缺少结构化主体ID")
		}
		if _, exists := seen[entityID]; exists {
			return nil, NewSubjectWorkflowError("SUBJECT_PARTY_PAYLOAD_INVALID", "重检主体列表包含重复主体")
		}
		seen[entityID] = struct{}{}
		var entity models.Entity
		if err := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", entityID).First(&entity).Error; err != nil {
			return nil, NewSubjectWorkflowError("SUBJECT_ENTITY_NOT_FOUND", "重检主体引用的结构化主体不存在")
		}
		role := strings.TrimSpace(subjectString(item, "role"))
		if role == "" {
			role = strings.TrimSpace(subjectString(item, "party_type", "partyType"))
		}
		partyInfo, err := conflictPartyInfoFromEntity(entity, role)
		if err != nil {
			return nil, err
		}
		parties = append(parties, partyInfo)
	}
	return parties, nil
}

func conflictPartyInfoFromEntity(entity models.Entity, role string) (models.ConflictPartyInfo, error) {
	identifiers := map[string]string{}
	identityNumber := strings.TrimSpace(entity.IdentityNumber)
	if identityNumber == "" && strings.TrimSpace(entity.IdentityNumberCiphertext) != "" {
		var err error
		identityNumber, err = security.DecryptIdentityNumber(entity.IdentityNumberCiphertext)
		if err != nil {
			return models.ConflictPartyInfo{}, NewSubjectWorkflowError("SUBJECT_IDENTITY_UNREADABLE", "案件主体身份标识无法安全读取，已阻止重检")
		}
	}
	if identityNumber != "" {
		identifiers[strings.ToLower(string(entity.IdentityType))] = identityNumber
	}
	return models.ConflictPartyInfo{
		Name:        entity.Name,
		Role:        role,
		EntityType:  string(entity.EntityType),
		Identifiers: identifiers,
		Aliases:     splitSubjectAliases(entity.Alias),
	}, nil
}

func subjectClientType(value string) string {
	if strings.Contains(strings.ToUpper(strings.TrimSpace(value)), "COMPANY") || strings.Contains(strings.TrimSpace(value), "企业") {
		return "COMPANY"
	}
	return "PERSON"
}

func splitSubjectAliases(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == '，' || r == ';' || r == '；' })
	aliases := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			aliases = append(aliases, trimmed)
		}
	}
	return aliases
}

func subjectString(values map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok && value != nil {
			if text, ok := value.(string); ok {
				return text
			}
		}
	}
	return ""
}

func subjectUintSlice(value interface{}) ([]uint, bool) {
	items, ok := value.([]interface{})
	if !ok {
		return nil, false
	}
	ids := make([]uint, 0, len(items))
	for _, item := range items {
		mapped := map[string]interface{}{"value": item}
		id, valid := subjectUint(mapped, "value")
		if !valid {
			return nil, false
		}
		ids = append(ids, id)
	}
	return ids, true
}

func subjectUint(values map[string]interface{}, keys ...string) (uint, bool) {
	for _, key := range keys {
		value, exists := values[key]
		if !exists || value == nil {
			continue
		}
		switch typed := value.(type) {
		case float64:
			if typed > 0 {
				return uint(typed), true
			}
		case int:
			if typed > 0 {
				return uint(typed), true
			}
		case uint:
			if typed > 0 {
				return typed, true
			}
		case string:
			parsed, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 32)
			if err == nil && parsed > 0 {
				return uint(parsed), true
			}
		}
	}
	return 0, false
}

const encryptedSubjectPayloadPrefix = "enc:v1:"

// protectSubjectPayload prevents identity numbers from being persisted in
// plaintext revision rows. Non-sensitive subject data remains queryable for
// local development, while a production request containing an identifier must
// have SUBJECT_DATA_KEY configured.
func protectSubjectPayload(payload []byte, value interface{}) ([]byte, error) {
	if !containsSubjectIdentifier(value) {
		return payload, nil
	}
	key, err := loadSubjectDataKey()
	if err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_DATA_KEY_REQUIRED", "主体身份标识需要加密保存，当前环境未配置主体数据密钥")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_DATA_KEY_INVALID", "主体数据密钥不可用，已阻止保存")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_DATA_KEY_INVALID", "主体数据加密器初始化失败，已阻止保存")
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_DATA_ENCRYPTION_FAILED", "主体数据加密失败，已阻止保存")
	}
	ciphertext := gcm.Seal(nil, nonce, payload, nil)
	encoded := encryptedSubjectPayloadPrefix + base64.RawURLEncoding.EncodeToString(nonce) + ":" + base64.RawURLEncoding.EncodeToString(ciphertext)
	return []byte(encoded), nil
}

func unprotectSubjectPayload(stored string) ([]byte, error) {
	stored = strings.TrimSpace(stored)
	if !strings.HasPrefix(stored, encryptedSubjectPayloadPrefix) {
		return []byte(stored), nil
	}
	parts := strings.Split(strings.TrimPrefix(stored, encryptedSubjectPayloadPrefix), ":")
	if len(parts) != 2 {
		return nil, NewSubjectWorkflowError("SUBJECT_PAYLOAD_INVALID", "加密主体变更记录格式无效")
	}
	key, err := loadSubjectDataKey()
	if err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_DATA_KEY_REQUIRED", "主体数据密钥未配置，无法执行主体重检")
	}
	nonce, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_PAYLOAD_INVALID", "加密主体变更记录无法读取")
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_PAYLOAD_INVALID", "加密主体变更记录无法读取")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_DATA_KEY_INVALID", "主体数据密钥不可用，已阻止重检")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(nonce) != gcm.NonceSize() {
		return nil, NewSubjectWorkflowError("SUBJECT_PAYLOAD_INVALID", "加密主体变更记录校验失败")
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, NewSubjectWorkflowError("SUBJECT_PAYLOAD_INVALID", "加密主体变更记录校验失败")
	}
	return plaintext, nil
}

func loadSubjectDataKey() ([]byte, error) {
	raw := strings.TrimSpace(os.Getenv("SUBJECT_DATA_KEY"))
	if raw == "" {
		return nil, errors.New("SUBJECT_DATA_KEY is not configured")
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if decoded, err := hex.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if len(raw) == 32 {
		return []byte(raw), nil
	}
	return nil, errors.New("SUBJECT_DATA_KEY must decode to 32 bytes")
}

func containsSubjectIdentifier(value interface{}) bool {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, nested := range typed {
			if isSubjectIdentifierKey(key) {
				return true
			}
			if containsSubjectIdentifier(nested) {
				return true
			}
		}
	case []interface{}:
		for _, nested := range typed {
			if containsSubjectIdentifier(nested) {
				return true
			}
		}
	}
	return false
}

func isSubjectIdentifierKey(key string) bool {
	key = strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.TrimSpace(key)))
	switch key {
	case "identifiers", "clientidentifiers", "identitynumber", "identityno", "idcard", "idcardnumber", "unifiedsocialcreditcode", "uscc", "creditcode", "passportnumber", "passportno":
		return true
	default:
		return false
	}
}

func mapValue(value models.JSON, key string) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	if nested, ok := value[key].(map[string]interface{}); ok {
		return nested
	}
	return map[string]interface{}{}
}

func mapString(values map[string]interface{}, key string) string {
	if values == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(values[key]))
}

func subjectFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" && trimmed != "<nil>" {
			return trimmed
		}
	}
	return ""
}

func isRecheckState(status string) bool {
	return status == models.SubjectStateChangeProposed || status == models.SubjectStateRecheckRequired
}

func revisionStatusAfterDecision(decision string) string {
	switch decision {
	case "no_conflict", "false_positive":
		return models.SubjectStateChangeApprovedAndEffective
	case "confirmed_conflict":
		return models.SubjectStateChangeRejected
	default:
		return models.SubjectStateRecheckRequired
	}
}

func actionGateMessage(state string) string {
	if state == models.SubjectStateEffective {
		return "主体版本已生效"
	}
	return "主体变更仍待独立复核，受控动作继续阻断"
}

func subjectTimePtr(value time.Time) *time.Time { return &value }

func isControlledCaseStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "approved", "signed", "in_progress", "trial", "closed", "archived":
		return true
	default:
		return false
	}
}
