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

	WaiverDecisionApprove        = "APPROVE"
	WaiverDecisionReject         = "REJECT"
	WaiverDecisionRequestChanges = "REQUEST_CHANGES"
	WaiverDecisionEscalate       = "ESCALATE"
)

var ErrWaiverNotFound = errors.New("waiver application not found")

type WaiverWorkflowRepository interface {
	CreateWaiverApplication(ctx context.Context, application *models.WaiverApplication) error
	GetWaiverApplication(ctx context.Context, id string) (*models.WaiverApplication, error)
	UpdateWaiverApplication(ctx context.Context, application *models.WaiverApplication) error
	CreateWaiverApprovalRecord(ctx context.Context, record *models.WaiverApprovalRecord) error
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
	ConflictCheckID             string                   `json:"conflict_check_id" binding:"required"`
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
	AssignedReviewer            string                   `json:"assigned_reviewer" binding:"required"`
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

func (s *WaiverWorkflowService) CreateWaiver(ctx context.Context, approvalID, actorID, actorName string, req *CreateWaiverRequest) (*models.WaiverApplication, error) {
	if strings.TrimSpace(req.ConflictCheckID) == "" {
		return nil, errors.New("conflict_check_id is required")
	}
	if strings.TrimSpace(req.Rationale) == "" {
		return nil, errors.New("rationale is required")
	}
	if strings.TrimSpace(req.AssignedReviewer) == "" {
		return nil, errors.New("assigned_reviewer is required")
	}

	conflictRecord, _ := s.conflictRepo.GetCheckRecord(ctx, req.ConflictCheckID)
	clientID := firstNonEmpty(req.ClientID, conflictRecordValue(conflictRecord, "client_id"))
	lawyerID := firstNonEmpty(req.LawyerID, conflictRecordUserID(conflictRecord), actorID)
	caseID := req.CaseID
	waiverType := firstNonEmpty(req.WaiverType, "INFORMED_CONSENT")
	waiverCategory := firstNonEmpty(req.WaiverCategory, "CLIENT_CONSENT")
	reviewPriority := firstNonEmpty(req.ReviewPriority, "MEDIUM")
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
	_ = s.createWaiverReviewInbox(ctx, application, approvalID)

	return application, nil
}

func (s *WaiverWorkflowService) GetWaiver(ctx context.Context, waiverID string) (*models.WaiverApplication, error) {
	application, err := s.waiverRepo.GetWaiverApplication(ctx, waiverID)
	if err != nil {
		return nil, ErrWaiverNotFound
	}
	return application, nil
}

func (s *WaiverWorkflowService) DecideWaiver(ctx context.Context, waiverID, actorID, actorName string, req *WaiverDecisionRequest) (*models.WaiverApplication, error) {
	application, err := s.waiverRepo.GetWaiverApplication(ctx, waiverID)
	if err != nil {
		return nil, ErrWaiverNotFound
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

	return s.waiverRepo.GetWaiverApplication(ctx, waiverID)
}

func (s *WaiverWorkflowService) createWaiverReviewInbox(ctx context.Context, application *models.WaiverApplication, approvalID string) error {
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
