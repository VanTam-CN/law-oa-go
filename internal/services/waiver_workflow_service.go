package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

const (
	WaiverStatusUnderReview = "UNDER_REVIEW"
	WaiverStatusApproved    = "APPROVED"
	WaiverStatusRejected    = "REJECTED"
	WaiverStatusExpired     = "EXPIRED"
	// A waiver may have no fixed expiry date, but it must still be reviewed
	// periodically. Once that review is overdue it is no longer a usable
	// exception and every controlled action must fail closed.
	WaiverStatusReviewOverdue = "WAIVER_REVIEW_OVERDUE"

	WaiverDecisionApprove        = "APPROVE"
	WaiverDecisionReject         = "REJECT"
	WaiverDecisionRequestChanges = "REQUEST_CHANGES"
	WaiverDecisionEscalate       = "ESCALATE"
)

var (
	ErrWaiverNotFound  = errors.New("waiver application not found")
	ErrWaiverForbidden = errors.New("waiver operation forbidden")
)

type WaiverWorkflowRepository interface {
	CreateWaiverApplication(ctx context.Context, application *models.WaiverApplication) error
	GetWaiverApplication(ctx context.Context, id string) (*models.WaiverApplication, error)
	GetWaiverApplicationsByConflictCheck(ctx context.Context, conflictCheckID string) ([]*models.WaiverApplication, error)
	UpdateWaiverApplication(ctx context.Context, application *models.WaiverApplication) error
	CreateWaiverApprovalRecord(ctx context.Context, record *models.WaiverApprovalRecord) error
	GetWaiverApprovalRecords(ctx context.Context, applicationID string) ([]*models.WaiverApprovalRecord, error)
}

type WaiverWorkflowService struct {
	waiverRepo   WaiverWorkflowRepository
	conflictRepo repositories.BasicConflictRepository
	inboxRepo    repositories.InboxRepository
	userRepo     repositories.UserRepository
}

func NewWaiverWorkflowService(
	waiverRepo WaiverWorkflowRepository,
	conflictRepo repositories.BasicConflictRepository,
	inboxRepo repositories.InboxRepository,
	userRepo repositories.UserRepository,
) *WaiverWorkflowService {
	return &WaiverWorkflowService{
		waiverRepo:   waiverRepo,
		conflictRepo: conflictRepo,
		inboxRepo:    inboxRepo,
		userRepo:     userRepo,
	}
}

type CreateWaiverRequest struct {
	ConflictCheckID             string                   `json:"conflict_check_id"`
	CaseID                      string                   `json:"case_id"`
	ClientID                    string                   `json:"client_id"`
	LawyerID                    string                   `json:"lawyer_id"`
	WaiverType                  string                   `json:"waiver_type"`
	WaiverCategory              string                   `json:"waiver_category"`
	ConflictSummary             string                   `json:"conflict_summary"`
	Conflicts                   []map[string]interface{} `json:"conflicts"`
	RiskAssessment              map[string]interface{}   `json:"risk_assessment"`
	ProposedConditions          []string                 `json:"proposed_conditions"`
	Limitations                 []string                 `json:"limitations"`
	MonitoringRequirements      []string                 `json:"monitoring_requirements"`
	ReportingRequirements       []string                 `json:"reporting_requirements"`
	Rationale                   string                   `json:"rationale" binding:"required"`
	SupportingEvidence          []map[string]interface{} `json:"supporting_evidence"`
	AlternativesConsidered      []string                 `json:"alternatives_considered"`
	ClientRepresentativeName    string                   `json:"client_representative_name"`
	ClientRepresentativeTitle   string                   `json:"client_representative_title"`
	ClientRepresentativeContact string                   `json:"client_representative_contact"`
	RequestingLawyerName        string                   `json:"requesting_lawyer_name"`
	RequestingLawyerTitle       string                   `json:"requesting_lawyer_title"`
	SupervisingLawyerName       string                   `json:"supervising_lawyer_name"`
	AssignedReviewer            string                   `json:"assigned_reviewer"`
	ReviewPriority              string                   `json:"review_priority"`
	DurationDays                *int                     `json:"duration_days"`
}

type WaiverDecisionRequest struct {
	Decision               string                   `json:"decision" binding:"required"`
	DecisionReason         string                   `json:"decision_reason" binding:"required"`
	DecisionComments       string                   `json:"decision_comments"`
	ApproverName           string                   `json:"approver_name"`
	ApproverTitle          string                   `json:"approver_title"`
	ApproverRole           string                   `json:"approver_role"`
	ApprovedConditions     map[string]interface{}   `json:"approved_conditions"`
	ImposedLimitations     map[string]interface{}   `json:"imposed_limitations"`
	MonitoringRequirements map[string]interface{}   `json:"monitoring_requirements"`
	ReportingRequirements  map[string]interface{}   `json:"reporting_requirements"`
	RiskMitigationPlan     map[string]interface{}   `json:"risk_mitigation_plan"`
	FollowUpActions        []map[string]interface{} `json:"follow_up_actions"`
}

func (s *WaiverWorkflowService) CreateWaiver(ctx context.Context, approvalID, actorID, actorName, actorRole string, req *CreateWaiverRequest) (*models.WaiverApplication, error) {
	if strings.TrimSpace(req.ConflictCheckID) == "" {
		return nil, errors.New("conflict_check_id is required")
	}
	if strings.TrimSpace(req.Rationale) == "" {
		return nil, errors.New("rationale is required")
	}
	if len(req.ProposedConditions) == 0 {
		return nil, errors.New("at least one risk-control condition is required")
	}
	conflictRecord, err := s.conflictRepo.GetCheckRecord(ctx, req.ConflictCheckID)
	if err != nil || conflictRecord == nil {
		return nil, ErrConflictTaskNotFound
	}
	if strings.TrimSpace(conflictRecord.ClientID) == "" || conflictRecord.UserID == 0 {
		return nil, fmt.Errorf("%w: 冲突检测记录缺少客户或承办律师绑定", ErrWaiverForbidden)
	}
	if strings.TrimSpace(req.ClientID) != "" && strings.TrimSpace(req.ClientID) != strings.TrimSpace(conflictRecord.ClientID) {
		return nil, fmt.Errorf("%w: 豁免申请客户与冲突检测记录不一致", ErrWaiverForbidden)
	}
	if strings.TrimSpace(req.LawyerID) != "" && strings.TrimSpace(req.LawyerID) != conflictRecordUserID(conflictRecord) {
		return nil, fmt.Errorf("%w: 豁免申请承办律师与冲突检测记录不一致", ErrWaiverForbidden)
	}
	if recordedCaseID := subjectCaseIDFromSearchParameters(conflictRecord.SearchParameters); recordedCaseID > 0 && strings.TrimSpace(req.CaseID) != "" && strings.TrimSpace(req.CaseID) != strconv.FormatUint(uint64(recordedCaseID), 10) {
		return nil, fmt.Errorf("%w: 豁免申请案件与冲突检测记录不一致", ErrWaiverForbidden)
	}
	if strconv.FormatUint(uint64(conflictRecord.UserID), 10) != actorID && !isConflictManagementRole(actorRole) {
		return nil, fmt.Errorf("%w: only the conflict task owner or a management reviewer may request a waiver", ErrWaiverForbidden)
	}
	decisionStatus := conflictRecordDecisionStatus(conflictRecord)
	if !conflictRecord.HasConflict && decisionStatus != "BLOCKED" && decisionStatus != "REVIEW_REQUIRED" {
		return nil, errors.New("waiver is only available for a blocked or review-required conflict check")
	}
	if conflictRecordHasNonWaivableDirectConflict(conflictRecord) {
		return nil, fmt.Errorf("%w: a confirmed direct conflict cannot be waived", ErrWaiverForbidden)
	}
	existing, err := s.waiverRepo.GetWaiverApplicationsByConflictCheck(ctx, req.ConflictCheckID)
	if err == nil {
		for _, application := range existing {
			if application != nil && application.Status != WaiverStatusRejected {
				if err := s.updateConflictWaiverState(ctx, conflictRecord, application); err != nil {
					return nil, err
				}
				return application, nil
			}
		}
	}
	if strings.TrimSpace(req.AssignedReviewer) == "" {
		req.AssignedReviewer = s.resolveWaiverReviewer(actorID)
	}
	if req.AssignedReviewer == "" || req.AssignedReviewer == actorID {
		return nil, fmt.Errorf("%w: an independent waiver reviewer is required", ErrWaiverForbidden)
	}
	clientID := firstNonEmpty(req.ClientID, conflictRecordValue(conflictRecord, "client_id"))
	lawyerID := firstNonEmpty(req.LawyerID, conflictRecordUserID(conflictRecord), actorID)
	caseID := req.CaseID
	waiverType := strings.ToUpper(firstNonEmpty(req.WaiverType, "INFORMED_CONSENT"))
	waiverCategory := strings.ToUpper(firstNonEmpty(req.WaiverCategory, "CLIENT_CONSENT"))
	reviewPriority := strings.ToUpper(firstNonEmpty(req.ReviewPriority, "MEDIUM"))
	if !allowedWaiverValue(waiverType, "INFORMED_CONSENT", "ETHICAL_BARRIER", "INFORMATION_SCREEN", "STRUCTURAL_BARRIER") {
		return nil, errors.New("unsupported waiver_type")
	}
	if !allowedWaiverValue(waiverCategory, "CLIENT_CONSENT", "BARRIER_IMPLEMENTATION", "MONITORING_ARRANGEMENT", "SPECIAL_CIRCUMSTANCES") {
		return nil, errors.New("unsupported waiver_category")
	}
	if !allowedWaiverValue(reviewPriority, "LOW", "MEDIUM", "HIGH", "URGENT") {
		return nil, errors.New("unsupported review_priority")
	}
	if req.DurationDays != nil && (*req.DurationDays < 1 || *req.DurationDays > 3650) {
		return nil, errors.New("duration_days must be between 1 and 3650")
	}
	now := time.Now()
	stage := "COMPLIANCE_REVIEW"

	application := &models.WaiverApplication{
		BaseModel: models.BaseModel{
			ID: uuid.NewString(),
		},
		ApplicationNumber:        fmt.Sprintf("WV-%s-%06d", now.Format("20060102"), now.UnixNano()%1000000),
		ConflictCheckID:          req.ConflictCheckID,
		ClientID:                 clientID,
		LawyerID:                 lawyerID,
		WaiverType:               waiverType,
		WaiverCategory:           waiverCategory,
		ConflictSummary:          buildConflictSummary(req, conflictRecord),
		Conflicts:                toModelJSON(req.Conflicts),
		RiskAssessment:           buildWaiverRiskAssessment(req, conflictRecord),
		ProposedConditions:       toModelJSON(req.ProposedConditions),
		Limitations:              toModelJSON(req.Limitations),
		MonitoringRequirements:   toModelJSON(req.MonitoringRequirements),
		ReportingRequirements:    toModelJSON(req.ReportingRequirements),
		RequestedEffectiveDate:   now,
		DurationDays:             req.DurationDays,
		Rationale:                req.Rationale,
		SupportingEvidence:       toModelJSON(req.SupportingEvidence),
		AlternativesConsidered:   toModelJSON(req.AlternativesConsidered),
		ClientRepresentativeName: firstNonEmpty(req.ClientRepresentativeName, "待补充"),
		RequestingLawyerName:     firstNonEmpty(req.RequestingLawyerName, actorName, "待补充"),
		Status:                   WaiverStatusUnderReview,
		SubmissionDate:           &now,
		ReviewPriority:           reviewPriority,
		CurrentStage:             &stage,
		AssignedReviewer:         &req.AssignedReviewer,
		ReviewDeadline:           waiverTimePtr(now.Add(48 * time.Hour)),
		CreatedBy:                actorID,
		UpdatedBy:                &actorID,
	}

	if caseID != "" {
		application.CaseID = &caseID
	}
	if req.DurationDays != nil && *req.DurationDays > 0 {
		expiresAt := now.Add(time.Duration(*req.DurationDays) * 24 * time.Hour)
		application.RequestedExpiryDate = &expiresAt
	}
	if req.ClientRepresentativeTitle != "" {
		application.ClientRepresentativeTitle = &req.ClientRepresentativeTitle
	}
	if req.ClientRepresentativeContact != "" {
		application.ClientRepresentativeContact = &req.ClientRepresentativeContact
	}
	if req.RequestingLawyerTitle != "" {
		application.RequestingLawyerTitle = &req.RequestingLawyerTitle
	}
	if req.SupervisingLawyerName != "" {
		application.SupervisingLawyerName = &req.SupervisingLawyerName
	}

	if err := s.waiverRepo.CreateWaiverApplication(ctx, application); err != nil {
		return nil, err
	}
	if err := s.updateConflictWaiverState(ctx, conflictRecord, application); err != nil {
		return nil, err
	}
	_ = s.createWaiverReviewInbox(ctx, application, approvalID)

	return application, nil
}

func (s *WaiverWorkflowService) GetConflictWaiver(ctx context.Context, conflictCheckID, actorID, actorRole string) (*models.WaiverApplication, error) {
	record, err := s.conflictRepo.GetCheckRecord(ctx, conflictCheckID)
	if err != nil || record == nil {
		return nil, ErrConflictTaskNotFound
	}
	if strconv.FormatUint(uint64(record.UserID), 10) != actorID && !isConflictWaiverViewerRole(actorRole) {
		return nil, fmt.Errorf("%w: only the conflict task owner or a management reviewer may view its waiver", ErrWaiverForbidden)
	}
	applications, err := s.waiverRepo.GetWaiverApplicationsByConflictCheck(ctx, conflictCheckID)
	if err != nil || len(applications) == 0 {
		return nil, ErrWaiverNotFound
	}
	return s.expireWaiverIfNeeded(ctx, applications[0])
}

func (s *WaiverWorkflowService) GetWaiver(ctx context.Context, waiverID, actorID, actorRole string) (*models.WaiverApplication, error) {
	application, err := s.waiverRepo.GetWaiverApplication(ctx, waiverID)
	if err != nil {
		return nil, ErrWaiverNotFound
	}
	if application.CreatedBy != actorID && valueOrEmpty(application.AssignedReviewer) != actorID && !isConflictManagementRole(actorRole) {
		return nil, fmt.Errorf("%w: waiver is not accessible to current user", ErrWaiverForbidden)
	}
	return s.expireWaiverIfNeeded(ctx, application)
}

func (s *WaiverWorkflowService) DecideWaiver(ctx context.Context, waiverID, actorID, actorName, actorRole string, req *WaiverDecisionRequest) (*models.WaiverApplication, error) {
	application, err := s.waiverRepo.GetWaiverApplication(ctx, waiverID)
	if err != nil {
		return nil, ErrWaiverNotFound
	}
	if valueOrEmpty(application.AssignedReviewer) != actorID && !isConflictManagementRole(actorRole) {
		return nil, fmt.Errorf("%w: only the assigned reviewer or a management reviewer may decide this waiver", ErrWaiverForbidden)
	}
	if application.CreatedBy == actorID {
		return nil, fmt.Errorf("%w: waiver requester cannot approve their own request", ErrWaiverForbidden)
	}
	if application.Status != WaiverStatusUnderReview {
		return nil, errors.New("only an under-review waiver may be decided")
	}

	decision := normalizeWaiverDecision(req.Decision)
	if decision == "" {
		return nil, errors.New("unsupported waiver decision")
	}
	if strings.TrimSpace(req.DecisionReason) == "" {
		return nil, errors.New("decision_reason is required")
	}

	now := time.Now()
	approverTitle := req.ApproverTitle
	comments := req.DecisionComments
	record := &models.WaiverApprovalRecord{
		BaseModel: models.BaseModel{
			ID: uuid.NewString(),
		},
		WaiverApplicationID:    application.ID,
		ApprovalStage:          firstNonEmpty(valueOrEmpty(application.CurrentStage), "COMPLIANCE_REVIEW"),
		ApproverID:             actorID,
		ApproverName:           firstNonEmpty(req.ApproverName, actorName, "未知审批人"),
		ApproverTitle:          optionalString(approverTitle),
		ApproverRole:           firstNonEmpty(req.ApproverRole, "COMPLIANCE_OFFICER"),
		Decision:               decision,
		DecisionReason:         req.DecisionReason,
		DecisionComments:       optionalString(comments),
		ApprovedConditions:     toModelJSON(req.ApprovedConditions),
		ImposedLimitations:     toModelJSON(req.ImposedLimitations),
		MonitoringRequirements: toModelJSON(req.MonitoringRequirements),
		ReportingRequirements:  toModelJSON(req.ReportingRequirements),
		RiskAssessment:         application.RiskAssessment,
		RiskMitigationPlan:     toModelJSON(req.RiskMitigationPlan),
		FollowUpActions:        toModelJSON(req.FollowUpActions),
		ApprovalDate:           now,
		Status:                 "ACTIVE",
	}
	if decision == WaiverDecisionApprove {
		record.EffectiveDate = &now
		record.ExpiryDate = application.RequestedExpiryDate
		nextReviewAt := now.AddDate(1, 0, 0)
		if application.RequestedExpiryDate != nil && application.RequestedExpiryDate.Before(nextReviewAt) {
			nextReviewAt = *application.RequestedExpiryDate
		}
		record.NextReviewDate = &nextReviewAt
	}
	if err := s.waiverRepo.CreateWaiverApprovalRecord(ctx, record); err != nil {
		return nil, err
	}

	switch decision {
	case WaiverDecisionApprove:
		application.Status = WaiverStatusApproved
		finalStage := "FINAL_APPROVAL"
		application.CurrentStage = &finalStage
	case WaiverDecisionReject:
		application.Status = WaiverStatusRejected
		finalStage := "FINAL_APPROVAL"
		application.CurrentStage = &finalStage
	case WaiverDecisionRequestChanges:
		application.Status = WaiverStatusUnderReview
		stage := "INITIAL_REVIEW"
		application.CurrentStage = &stage
	case WaiverDecisionEscalate:
		application.Status = WaiverStatusUnderReview
		stage := "MANAGEMENT_APPROVAL"
		application.CurrentStage = &stage
	}
	application.UpdatedBy = &actorID

	if err := s.waiverRepo.UpdateWaiverApplication(ctx, application); err != nil {
		return nil, err
	}
	conflictRecord, _ := s.conflictRepo.GetCheckRecord(ctx, application.ConflictCheckID)
	if conflictRecord != nil {
		if err := s.updateConflictWaiverState(ctx, conflictRecord, application); err != nil {
			return nil, err
		}
	}

	return s.waiverRepo.GetWaiverApplication(ctx, waiverID)
}

func (s *WaiverWorkflowService) resolveWaiverReviewer(actorID string) string {
	for _, role := range []string{"compliance", "risk_control", "risk", "director", "partner", "admin", "super_admin"} {
		users, err := s.userRepo.FindByRole(role, 10)
		if err != nil {
			continue
		}
		for _, user := range users {
			candidate := strconv.FormatUint(uint64(user.ID), 10)
			if candidate != actorID {
				return candidate
			}
		}
	}
	return ""
}

func (s *WaiverWorkflowService) updateConflictWaiverState(ctx context.Context, record *models.ConflictCheckRecord, application *models.WaiverApplication) error {
	if record.CheckResult == nil {
		record.CheckResult = models.JSON{}
	}
	waiver := models.JSON{
		"id": application.ID, "applicationNumber": application.ApplicationNumber,
		"status": application.Status, "currentStage": valueOrEmpty(application.CurrentStage),
		"assignedReviewer": valueOrEmpty(application.AssignedReviewer), "updatedAt": time.Now(),
		"expiryAt": application.RequestedExpiryDate,
	}
	record.CheckResult["waiver"] = waiver
	decision, _ := record.CheckResult["decision"].(map[string]interface{})
	if decision == nil {
		if typed, ok := record.CheckResult["decision"].(models.JSON); ok {
			decision = map[string]interface{}(typed)
		} else {
			decision = map[string]interface{}{}
		}
	}
	switch application.Status {
	case WaiverStatusApproved:
		decision["status"] = "WAIVED"
		decision["recommendation"] = "豁免已经批准，可按批准条件继续接案；相关限制、监督和伦理墙要求必须持续执行。"
	case WaiverStatusReviewOverdue:
		decision["status"] = "BLOCKED"
		decision["recommendation"] = "豁免年度复核已逾期，当前受控动作已阻止；完成新的独立复核并重新批准前不得继续。"
	case WaiverStatusRejected:
		decision["status"] = "BLOCKED"
		decision["recommendation"] = "豁免申请已拒绝，当前接案继续保持阻止状态。"
	case WaiverStatusExpired:
		decision["status"] = "BLOCKED"
		decision["recommendation"] = "原豁免已经到期，必须重新评估风险并取得新的批准后才能继续接案。"
	default:
		decision["status"] = "WAIVER_PENDING"
		decision["recommendation"] = "豁免申请正在复核，批准前不得继续接案。"
	}
	record.CheckResult["decision"] = decision
	record.UpdatedAt = time.Now()
	return s.conflictRepo.UpdateCheckRecord(ctx, record)
}

func (s *WaiverWorkflowService) expireWaiverIfNeeded(ctx context.Context, application *models.WaiverApplication) (*models.WaiverApplication, error) {
	if application == nil || application.Status != WaiverStatusApproved {
		return application, nil
	}
	now := time.Now()
	if application.RequestedExpiryDate != nil && !application.RequestedExpiryDate.After(now) {
		application.Status = WaiverStatusExpired
		if err := s.waiverRepo.UpdateWaiverApplication(ctx, application); err != nil {
			return nil, err
		}
		return s.refreshConflictWaiverState(ctx, application)
	}
	records, err := s.waiverRepo.GetWaiverApprovalRecords(ctx, application.ID)
	if err != nil {
		return nil, err
	}
	var latestApproval *models.WaiverApprovalRecord
	for _, candidate := range records {
		if candidate == nil || candidate.Decision != WaiverDecisionApprove || candidate.Status != "ACTIVE" {
			continue
		}
		latestApproval = candidate
		break
	}
	// A legacy APPROVED application without an approval record or review date
	// cannot be treated as a valid exception. This is deliberately a blocking
	// migration state instead of an implicit one-year extension.
	if latestApproval == nil || latestApproval.NextReviewDate == nil || !latestApproval.NextReviewDate.After(now) {
		application.Status = WaiverStatusReviewOverdue
		if err := s.waiverRepo.UpdateWaiverApplication(ctx, application); err != nil {
			return nil, err
		}
		return s.refreshConflictWaiverState(ctx, application)
	}
	return application, nil
}

func (s *WaiverWorkflowService) refreshConflictWaiverState(ctx context.Context, application *models.WaiverApplication) (*models.WaiverApplication, error) {
	record, _ := s.conflictRepo.GetCheckRecord(ctx, application.ConflictCheckID)
	if record != nil {
		if err := s.updateConflictWaiverState(ctx, record, application); err != nil {
			return nil, err
		}
	}
	return application, nil
}

func isConflictManagementRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "super_admin", "director", "partner", "compliance", "risk", "risk_control", "management":
		return true
	default:
		return false
	}
}

func isConflictWaiverViewerRole(role string) bool {
	return isConflictManagementRole(role) || strings.EqualFold(strings.TrimSpace(role), "conflict_officer")
}

func (s *WaiverWorkflowService) createWaiverReviewInbox(ctx context.Context, application *models.WaiverApplication, approvalID string) error {
	if s.inboxRepo == nil {
		return nil
	}
	reviewerID, err := strconv.ParseUint(valueOrEmpty(application.AssignedReviewer), 10, 32)
	if err != nil || reviewerID == 0 {
		return nil
	}

	content := fmt.Sprintf("豁免申请 %s 已提交，请评估冲突检查 %s。", application.ApplicationNumber, application.ConflictCheckID)
	if approvalID != "" {
		content += " 关联审批：" + approvalID
	}

	return s.inboxRepo.Create(ctx, &models.InboxItem{
		UserID:      uint(reviewerID),
		SourceType:  "waiver",
		SourceID:    0,
		Title:       truncateInboxTitle("豁免评估待审批：" + application.ConflictSummary),
		Content:     content,
		Priority:    waiverPriorityToInbox(application.ReviewPriority),
		DueDate:     application.ReviewDeadline,
		DueDateType: "waiver_review",
	})
}

func buildConflictSummary(req *CreateWaiverRequest, record *models.ConflictCheckRecord) string {
	if strings.TrimSpace(req.ConflictSummary) != "" {
		return req.ConflictSummary
	}
	if record != nil {
		return fmt.Sprintf("案件 %s 的冲突检查结果需要豁免评估", record.CaseName)
	}
	return "冲突检查结果需要豁免评估"
}

func buildWaiverRiskAssessment(req *CreateWaiverRequest, record *models.ConflictCheckRecord) models.JSON {
	if len(req.RiskAssessment) > 0 {
		return toModelJSON(req.RiskAssessment)
	}
	if record != nil && record.CheckResult != nil {
		if value, ok := record.CheckResult["riskAssessment"]; ok {
			return toModelJSON(value)
		}
		return models.JSON{"riskLevel": record.RiskLevel, "hasConflict": record.HasConflict}
	}
	return models.JSON{"riskLevel": "MEDIUM"}
}

func conflictRecordValue(record *models.ConflictCheckRecord, field string) string {
	if record == nil {
		return ""
	}
	switch field {
	case "client_id":
		return record.ClientID
	default:
		return ""
	}
}

func conflictRecordUserID(record *models.ConflictCheckRecord) string {
	if record == nil || record.UserID == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(record.UserID), 10)
}

func conflictRecordDecisionStatus(record *models.ConflictCheckRecord) string {
	if record == nil || record.CheckResult == nil {
		return ""
	}
	decision, ok := record.CheckResult["decision"].(map[string]interface{})
	if !ok {
		if typed, typedOK := record.CheckResult["decision"].(models.JSON); typedOK {
			decision = map[string]interface{}(typed)
		}
	}
	return strings.ToUpper(strings.TrimSpace(fmt.Sprint(decision["status"])))
}

// conflictRecordHasNonWaivableDirectConflict keeps the hard policy boundary in
// the waiver service. A UI must not be able to turn a confirmed adverse-party
// conflict into an exception request merely by posting a waiver payload.
func conflictRecordHasNonWaivableDirectConflict(record *models.ConflictCheckRecord) bool {
	if record == nil {
		return false
	}
	return containsNonWaivableDirectConflict(record.CheckResult)
}

func containsNonWaivableDirectConflict(value interface{}) bool {
	switch typed := value.(type) {
	case models.JSON:
		return containsNonWaivableDirectConflict(map[string]interface{}(typed))
	case map[string]interface{}:
		ruleCode := strings.ToUpper(strings.TrimSpace(fmt.Sprint(firstJSONValue(typed, "ruleCode", "rule_code"))))
		if ruleCode == "DIRECT_ADVERSE_CURRENT_CLIENT" || ruleCode == "STRUCTURED_IDENTITY_EXACT" || ruleCode == "DIRECT_CONFLICT" {
			return true
		}
		conflictType := strings.ToUpper(strings.TrimSpace(fmt.Sprint(firstJSONValue(typed, "conflictType", "conflict_type"))))
		if strings.Contains(conflictType, "DIRECT") || strings.Contains(conflictType, "直接冲突") {
			return true
		}
		matchType := strings.ToUpper(strings.TrimSpace(fmt.Sprint(firstJSONValue(typed, "matchType", "match_type"))))
		partyRole := strings.ToUpper(strings.TrimSpace(fmt.Sprint(firstJSONValue(typed, "partyRole", "party_role"))))
		historicalRole := strings.ToUpper(strings.TrimSpace(fmt.Sprint(firstJSONValue(typed, "historicalRole", "historical_role"))))
		if matchType == "EXACT" && partyRole == "OPPOSING_PARTY" && historicalRole == "CLIENT" {
			return true
		}
		for _, child := range typed {
			if containsNonWaivableDirectConflict(child) {
				return true
			}
		}
	case []interface{}:
		for _, child := range typed {
			if containsNonWaivableDirectConflict(child) {
				return true
			}
		}
	}
	return false
}

func firstJSONValue(values map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if value, ok := values[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func normalizeWaiverDecision(decision string) string {
	switch strings.ToUpper(strings.TrimSpace(decision)) {
	case "APPROVE", "APPROVED":
		return WaiverDecisionApprove
	case "REJECT", "REJECTED":
		return WaiverDecisionReject
	case "REQUEST_CHANGES", "CHANGE", "CHANGES":
		return WaiverDecisionRequestChanges
	case "ESCALATE":
		return WaiverDecisionEscalate
	default:
		return ""
	}
}

func allowedWaiverValue(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func waiverPriorityToInbox(priority string) string {
	switch strings.ToUpper(priority) {
	case "URGENT":
		return "critical"
	case "HIGH":
		return "high"
	case "LOW":
		return "low"
	default:
		return "medium"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func waiverTimePtr(value time.Time) *time.Time {
	return &value
}
