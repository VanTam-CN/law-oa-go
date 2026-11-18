package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ErrApprovalForbidden identifies an object-level access denial.
var ErrApprovalForbidden = errors.New("无权查看此审批记录")

// ErrApprovalSnapshotStale marks a decision whose expected round snapshot no
// longer matches the approval's active submission snapshot (N3/C11): the
// caller must reload the approval and decide on the current round.
var ErrApprovalSnapshotStale = errors.New("审批依据已更新（当前轮次快照与请求不一致），请刷新后重试")

// ErrApprovalStateStale marks a decision that lost a race against another
// decision or a resubmit on the same approval row (N3/C11).
var ErrApprovalStateStale = errors.New("审批状态已变化，请刷新后重试")

// ErrApprovalRecordNotFound marks a missing approval subject. It wraps
// gorm.ErrRecordNotFound so both errors.Is checks and the legacy
// strings.Contains handlers map it to the canonical 404 contract.
var ErrApprovalRecordNotFound = fmt.Errorf("审批记录不存在: %w", gorm.ErrRecordNotFound)

// ApprovalService 审批服务
type ApprovalService struct {
	approvalRepo *repositories.ApprovalRepository
	userRepo     repositories.UserRepository
	stateMachine *ApprovalStateMachine
	assigner     *ApprovalAssigner
	// terminalBinding releases the intake bind row when a case-creation
	// approval reaches a lifecycle end for its current submission
	// (rejected/needs-revision/cancelled) and rebinds it on Resubmit.
	// Without the release the bind row permanently locks the intake
	// behind INTAKE_APPROVAL_ALREADY_IN_PROGRESS even though no live
	// approval remains (probe-reproduced on 58a91ed).
	terminalBinding *terminalBindingReleaser
}

// NewApprovalService 创建审批服务
func NewApprovalService(db *gorm.DB) *ApprovalService {
	approvalRepo := repositories.NewApprovalRepository(db)
	userRepo := repositories.NewUserRepository(db)

	return &ApprovalService{
		approvalRepo:    approvalRepo,
		userRepo:        userRepo,
		stateMachine:    NewApprovalStateMachine(),
		assigner:        NewApprovalAssigner(db, approvalRepo, userRepo),
		terminalBinding: newTerminalBindingReleaser(db),
	}
}

// GetApprovalStats 获取审批统计。
//
// Firm-wide pending totals are only returned when the caller explicitly has a
// business matter-management appointment. The variadic argument preserves the
// old one-argument call shape for internal callers while making the safe,
// user-scoped view the default.
func (s *ApprovalService) GetApprovalStats(userID string, includeFirmTotals ...bool) (*models.ApprovalStats, error) {
	var stats models.ApprovalStats
	showFirmTotals := len(includeFirmTotals) > 0 && includeFirmTotals[0]

	// 获取总申请数
	totalRequests, err := s.approvalRepo.CountByUserID(userID, "")
	if err != nil {
		return nil, fmt.Errorf("获取总申请数失败: %v", err)
	}
	stats.TotalRequests = int(totalRequests)

	// 获取待处理审批数（需要当前用户审批的）
	myPendingRequests, err := s.approvalRepo.CountPendingByApproverID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取待处理审批数失败: %v", err)
	}
	stats.MyPendingRequests = int(myPendingRequests)

	// A normal lawyer must not learn the firm's global approval volume. Use the
	// applicant's own pending count for that view; management roles can opt in
	// to the firm-wide count after the handler has checked their role.
	var pendingRequests int64
	if showFirmTotals {
		pendingRequests, err = s.approvalRepo.CountPending()
		if err != nil {
			return nil, fmt.Errorf("获取全部待审批数失败: %v", err)
		}
	} else {
		pendingRequests, err = s.approvalRepo.CountPendingByApplicantID(userID)
		if err != nil {
			return nil, fmt.Errorf("获取用户待审批数失败: %v", err)
		}
	}
	stats.PendingRequests = int(pendingRequests)

	// 获取已通过申请数
	approvedRequests, err := s.approvalRepo.CountByUserID(userID, models.ApprovalStatusApproved)
	if err != nil {
		return nil, fmt.Errorf("获取已通过申请数失败: %v", err)
	}
	stats.ApprovedRequests = int(approvedRequests)

	// 获取已拒绝申请数
	rejectedRequests, err := s.approvalRepo.CountByUserID(userID, models.ApprovalStatusRejected)
	if err != nil {
		return nil, fmt.Errorf("获取已拒绝申请数失败: %v", err)
	}
	stats.RejectedRequests = int(rejectedRequests)

	return &stats, nil
}

// GetPendingApprovals 获取待审批列表
func (s *ApprovalService) GetPendingApprovals(userID string, req *models.ApprovalListRequest) (*models.ApprovalListResponse, error) {
	approvals, err := s.approvalRepo.FindPendingByApproverID(userID, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("获取待审批列表失败: %v", err)
	}

	total, err := s.approvalRepo.CountPendingByApproverID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取待审批总数失败: %v", err)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &models.ApprovalListResponse{
		List:       approvals,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListApprovals 获取审批列表
func (s *ApprovalService) ListApprovals(userID string, req *models.ApprovalListRequest) (*models.ApprovalListResponse, error) {
	queryUserID := userID
	if req.ApplicantID != "" {
		queryUserID = req.ApplicantID
	}

	approvals, err := s.approvalRepo.FindByUserID(queryUserID, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("获取审批列表失败: %v", err)
	}

	total, err := s.approvalRepo.CountByUserID(queryUserID, req.Status)
	if err != nil {
		return nil, fmt.Errorf("获取审批总数失败: %v", err)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &models.ApprovalListResponse{
		List:       approvals,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetApproval 获取单个审批详情
func (s *ApprovalService) GetApproval(userID string, id string) (*models.ApprovalRequest, error) {
	approval, err := s.GetApprovalForAuthorization(userID, id)
	if err != nil || approval == nil {
		return approval, err
	}
	if err := s.LoadApprovalRecords(approval); err != nil {
		return nil, err
	}
	return approval, nil
}

// GetApprovalForAuthorization loads only the approval subject. Handlers use it
// before ethical-wall authorization so related audit rows are not read until
// the caller's object access has been established.
func (s *ApprovalService) GetApprovalForAuthorization(userID string, id string) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取审批详情失败: %v", err)
	}
	if approval == nil {
		return nil, ErrApprovalRecordNotFound
	}

	// The approval detail is an object-scoped resource. Technical admin is not
	// a business approval role, so it must not bypass the applicant/approver
	// boundary by guessing an approval ID.
	if approval != nil {
		allowed := approval.ApplicantID == userID || approval.CurrentApproverID == userID
		if !allowed {
			wasApprover, historyErr := s.approvalRepo.HasApproverRecord(approval.ID, userID)
			if historyErr != nil {
				return nil, historyErr
			}
			allowed = wasApprover
		}
		if !allowed {
			return nil, ErrApprovalForbidden
		}
	}

	return approval, nil
}

// LoadApprovalRecords populates immutable decision history after object-level
// authorization. Legacy development schemas may lack the table; production
// readiness independently requires the complete approval schema.
func (s *ApprovalService) LoadApprovalRecords(approval *models.ApprovalRequest) error {
	if approval == nil || !s.approvalRepo.DB().Migrator().HasTable(&models.ApprovalRecord{}) {
		return nil
	}
	records, err := s.approvalRepo.FindRecordsByApprovalID(approval.ID)
	if err != nil {
		return err
	}
	approval.Records = records
	return nil
}

func (s *ApprovalService) GetApprovalByID(id string) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取审批详情失败: %v", err)
	}
	if approval == nil {
		return nil, ErrApprovalRecordNotFound
	}
	return approval, nil
}

// GetApprovalRecords 获取审批记录
func (s *ApprovalService) GetApprovalRecords(approvalID string) ([]models.ApprovalRecord, error) {
	records, err := s.approvalRepo.FindRecordsByApprovalID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取审批记录失败: %v", err)
	}
	return records, nil
}

// buildApprovalRequest constructs the draft approval row (request number,
// workflow, metadata, optional fields) shared by the two-phase
// CreateApproval and the atomic CreateAndSubmitApproval.
func (s *ApprovalService) buildApprovalRequest(userID, userName string, req *models.CreateApprovalRequest) (*models.ApprovalRequest, error) {
	// 生成申请编号
	requestNumber, err := s.generateRequestNumber()
	if err != nil {
		return nil, fmt.Errorf("生成申请编号失败: %v", err)
	}

	// 如果userName为空，从数据库查询用户信息
	if userName == "" {
		user, err := s.userRepo.FindByStringID(userID)
		if err == nil && user != nil {
			userName = user.Name
		} else {
			userName = "未知用户"
		}
	}

	// 设置默认工作流类型
	workflowType := req.WorkflowType
	if workflowType == "" {
		workflowType = "STANDARD_APPROVAL"
	}

	// 创建审批申请 - 初始状态为草稿
	approval := &models.ApprovalRequest{
		ID:                    generateUUID(),
		RequestNumber:         requestNumber,
		Title:                 req.Title,
		Type:                  req.Type,
		Category:              req.Category,
		Content:               req.Content,
		ApplicantID:           userID,
		ApplicantName:         userName,
		ApplicantTitle:        "",
		DepartmentID:          "",
		DepartmentName:        "",
		Urgency:               req.Urgency,
		Priority:              req.Priority,
		ExpectedEffectiveDate: nil,
		ExpectedExpiryDate:    nil,
		DurationDays:          req.DurationDays,
		Status:                models.ApprovalStatusDraft,
		SubmissionDate:        nil,
		CurrentStage:          "",
		CurrentApproverID:     "",
		CurrentApproverName:   "",
		WorkflowType:          workflowType,
		WorkflowConfig:        "{}",
		Attachments:           "{}",
		Metadata:              "{}",
		ConflictResult:        "{}",
		CreatedBy:             userID,
		UpdatedBy:             userID,
	}

	// 设置默认值
	if approval.Urgency == "" {
		approval.Urgency = models.ApprovalUrgencyNormal
	}
	if approval.Priority == "" {
		approval.Priority = models.ApprovalPriorityMedium
	}

	// 处理附件和元数据
	if len(req.Attachments) > 0 {
		attachmentsBytes, _ := json.Marshal(req.Attachments)
		approval.Attachments = string(attachmentsBytes)
	}
	if req.Metadata != nil {
		metadataBytes, _ := json.Marshal(req.Metadata)
		approval.Metadata = string(metadataBytes)
	}

	// 处理时间字段
	if req.ExpectedEffectiveDate != "" {
		if effectiveDate, err := time.Parse("2006-01-02", req.ExpectedEffectiveDate); err == nil {
			approval.ExpectedEffectiveDate = &effectiveDate
		}
	}
	if req.ExpectedExpiryDate != "" {
		if expiryDate, err := time.Parse("2006-01-02", req.ExpectedExpiryDate); err == nil {
			approval.ExpectedExpiryDate = &expiryDate
		}
	}
	return approval, nil
}

// CreateApproval 创建审批申请
func (s *ApprovalService) CreateApproval(userID string, userName string, req *models.CreateApprovalRequest) (*models.ApprovalRequest, error) {
	approval, err := s.buildApprovalRequest(userID, userName, req)
	if err != nil {
		return nil, err
	}
	// 保存草稿
	if err := s.approvalRepo.Create(approval); err != nil {
		return nil, fmt.Errorf("保存审批申请失败: %v", err)
	}

	return approval, nil
}

// CreateAndSubmitApproval creates the draft row, binds the intake slot,
// assigns the approver, persists the intake conflict binding, and opens
// the reviewer handoff inside ONE database transaction (C4 initial
// submission). A gate or assignment failure rolls the draft back: unlike
// the legacy CreateApproval → SubmitApproval sequence, no orphan draft and
// no approver-visible approval without its binding can survive a failed
// intake submission. The intake bind UNIQUE index keeps its typed
// INTAKE_APPROVAL_ALREADY_IN_PROGRESS contract for takeover losers.
func (s *ApprovalService) CreateAndSubmitApproval(userID, userName string, req *models.CreateApprovalRequest) (*models.ApprovalRequest, error) {
	approval, err := s.buildApprovalRequest(userID, userName, req)
	if err != nil {
		return nil, err
	}
	if err := s.stateMachine.ValidateTransition(models.ApprovalStatusDraft, models.ApprovalStatusSubmitted); err != nil {
		return nil, err
	}
	if err := s.approvalRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(approval).Error; err != nil {
			return fmt.Errorf("保存审批申请失败: %w", err)
		}
		// Intake-slot binding: same UNIQUE durable duplicate guard as the
		// legacy path, now inside the atomic transaction.
		if intakeID := metadataStringFromApproval(approval); intakeID != "" {
			bound := intakeID
			tracking := &repositories.CaseCreationTracking{
				ID:                  fmt.Sprintf("INTAKE_BIND_%s", approval.ID),
				ApprovalRequestID:   approval.ID,
				IntakeID:            &bound,
				CreationStatus:      "pending",
				ErrorDetails:        "{}",
				DataMapping:         "{}",
				MappedFields:        "{}",
				UnmappedFields:      "{}",
				AppliedConditions:   "{}",
				ImposedRequirements: "{}",
				WorkflowActions:     "[]",
				CreatedBy:           userID,
				CreatedAt:           time.Now(),
			}
			if err := tx.Create(tracking).Error; err != nil {
				if asUniqueConstraintError(err) != nil {
					return &ApprovalConflictGateError{
						Code:    "INTAKE_APPROVAL_ALREADY_IN_PROGRESS",
						Message: "该接案已有进行中的成案审批，不能重复提交；请在原审批单上继续处理",
					}
				}
				return fmt.Errorf("接案绑定跟踪记录创建失败：%w", err)
			}
		}
		if err := s.assigner.AssignApprover(approval); err != nil {
			return fmt.Errorf("分配审批人失败: %v", err)
		}
		if err := s.initialIntakeApprovalBinding(tx, approval); err != nil {
			return err
		}
		now := time.Now()
		approval.SubmissionDate = &now
		approval.UpdatedBy = userID
		if err := s.approvalRepo.SaveInTx(tx, approval); err != nil {
			return err
		}
		return s.openApprovalDecisionTask(tx, approval)
	}); err != nil {
		var gateErr *ApprovalConflictGateError
		if errors.As(err, &gateErr) {
			return nil, gateErr
		}
		return nil, err
	}
	return s.approvalRepo.FindByID(approval.ID)
}

// metadataStringFromApproval extracts the intake_id anchor from the draft's
// serialized metadata without a second JSON round trip in the caller.
func metadataStringFromApproval(approval *models.ApprovalRequest) string {
	metadata := map[string]interface{}{}
	if approval != nil && strings.TrimSpace(approval.Metadata) != "" {
		_ = json.Unmarshal([]byte(approval.Metadata), &metadata)
	}
	return strings.TrimSpace(metadataString(metadata, "intake_id"))
}

// SubmitApproval 提交审批申请
func (s *ApprovalService) SubmitApproval(userID string, approvalID string) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}
	if approval == nil {
		return nil, ErrApprovalRecordNotFound
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return nil, fmt.Errorf("无权提交此审批记录")
	}

	// 只能提交草稿或被拒绝/需要修改状态的审批
	if approval != nil && !s.stateMachine.CanEdit(approval.Status) {
		return nil, fmt.Errorf("当前状态 %s 不允许提交", approval.Status)
	}

	// 验证状态转换
	if err := s.stateMachine.ValidateTransition(approval.Status, models.ApprovalStatusSubmitted); err != nil {
		return nil, err
	}

	// 分配审批人
	if err := s.assigner.AssignApprover(approval); err != nil {
		return nil, fmt.Errorf("分配审批人失败: %v", err)
	}

	// 设置提交时间
	now := time.Now()
	approval.SubmissionDate = &now
	approval.UpdatedBy = userID

	// N4/C05+C13: the submission write and the approver's workflow handoff
	// commit together. A failed handoff rolls the submission back, so an
	// approval can never exist without its reviewer todo (and vice versa).
	if err := s.approvalRepo.Transaction(func(tx *gorm.DB) error {
		// C4: intake-scoped approvals must persist their server-owned
		// conflict binding (active association, immutable intake_submission
		// snapshot, active pointer, conflict_check_id anchor) in the same
		// transaction that submits them. A failed gate rolls the whole
		// submission back, so no approver-visible submitted approval can
		// exist without a provable binding.
		if err := s.initialIntakeApprovalBinding(tx, approval); err != nil {
			return err
		}
		if err := s.approvalRepo.SaveInTx(tx, approval); err != nil {
			return err
		}
		return s.openApprovalDecisionTask(tx, approval)
	}); err != nil {
		return nil, fmt.Errorf("提交审批失败: %v", err)
	}

	// 重新加载数据
	return s.approvalRepo.FindByID(approvalID)
}

// ProcessApprovalDecision 处理审批决定（通过/拒绝/要求修改/转派）
func (s *ApprovalService) ProcessApprovalDecision(userID string, approvalID string, req *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}
	if approval == nil {
		return nil, ErrApprovalRecordNotFound
	}

	// 确保是当前审批人
	if approval != nil && approval.CurrentApproverID != userID {
		return nil, errors.New("无权审批此申请")
	}
	if approval != nil && approval.ApplicantID == userID {
		return nil, errors.New("申请人不能审批自己的申请")
	}

	// 检查审批状态
	if approval != nil && approval.Status != models.ApprovalStatusSubmitted &&
		approval.Status != models.ApprovalStatusUnderReview &&
		approval.Status != models.ApprovalStatusResubmitted {
		return nil, errors.New("审批状态不允许处理")
	}

	// 获取下一个状态
	nextStatus, err := s.stateMachine.GetNextState(approval.Status, req.Decision)
	if err != nil {
		return nil, err
	}

	// N3/C08: structured supplement contract. A request_changes decision
	// requires a written reason; structured items pass the server-side
	// whitelist (stable item_id, category/field_key/description limits, no
	// HTML/links/identity numbers). Text-only returns normalize to one
	// legacy_text item. Invalid structured input is a 400, not a silent pass.
	if req.Decision == models.ApprovalDecisionRequestChanges {
		if err := NormalizeSupplementRequest(req); err != nil {
			return nil, &SupplementRequestError{Message: err.Error()}
		}
		// material_request items must reference materials of THIS
		// approval's own intake: the live metadata mapping is the
		// authoritative scope, never a client-supplied whitelist. A
		// cross-application material_id fails the decision here (400).
		if err := ValidateSupplementMaterialItems(s.approvalRepo.DB(), approvalID, req.ImposedRequirements); err != nil {
			return nil, &SupplementRequestError{Message: err.Error()}
		}
	}

	// N3/C11: stale-page guard. A client that pins the round snapshot it is
	// deciding on gets an explicit 409 instead of silently acting on a newer
	// round. Intake-bound approvals close the legacy hole: their decision
	// must name the active submission snapshot (spec review 2026-09-27, HOLD
	// item 2), so an old page can never act on a newer round by omitting the
	// field. Non-intake approvals keep the optional legacy behavior.
	var requestedSnapshot string
	if req != nil {
		requestedSnapshot = strings.TrimSpace(req.ExpectedSubmissionSnapshotID)
	}
	if requestedSnapshot == "" && approvalIntakeScoped(approval) {
		return nil, ErrApprovalSnapshotStale
	}
	if requestedSnapshot != "" {
		entrySnapshot := ""
		if approval.ActiveSubmissionSnapshotID != nil {
			entrySnapshot = strings.TrimSpace(*approval.ActiveSubmissionSnapshotID)
		}
		if entrySnapshot != requestedSnapshot {
			return nil, ErrApprovalSnapshotStale
		}
	}

	// N3/C11: the decision record, the approval write and the intake bind
	// release commit together, serialized on the approval row lock, so a
	// decision cannot interleave with a resubmit or a second decision.
	pendingNotice := ""
	if err := s.approvalRepo.Transaction(func(tx *gorm.DB) error {
		lockRow := struct{ ID string }{}
		if err := tx.Table("approval_requests").Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id").Where("id = ?", approvalID).Take(&lockRow).Error; err != nil {
			return fmt.Errorf("锁定审批记录失败: %w", err)
		}
		// Re-read the live state under the lock: a concurrent decision or
		// resubmit that committed after our entry read invalidates this one.
		var live struct {
			Status                     string
			ActiveSubmissionSnapshotID *string
		}
		if err := tx.Table("approval_requests").Select("status, active_submission_snapshot_id").
			Where("id = ?", approvalID).Take(&live).Error; err != nil {
			return fmt.Errorf("读取审批状态失败: %w", err)
		}
		if live.Status != approval.Status {
			return ErrApprovalStateStale
		}
		currentSnapshot := ""
		if live.ActiveSubmissionSnapshotID != nil {
			currentSnapshot = strings.TrimSpace(*live.ActiveSubmissionSnapshotID)
		}
		if requestedSnapshot != "" && currentSnapshot != requestedSnapshot {
			return ErrApprovalSnapshotStale
		}

		if approval.Status != models.ApprovalStatusUnderReview {
			approval.Status = models.ApprovalStatusUnderReview
		}

		approverName := approval.CurrentApproverName
		approverRole := ""
		if approver, lookupErr := s.userRepo.FindByStringID(userID); lookupErr == nil && approver != nil {
			if approver.Name != "" {
				approverName = approver.Name
			}
			approverRole = approver.Role
		}

		// 创建审批记录
		record := &models.ApprovalRecord{
			ID:                  generateUUID(),
			ApprovalRequestID:   approval.ID,
			Stage:               approval.CurrentStage,
			StageOrder:          1,
			ApproverID:          userID,
			ApproverName:        approverName,
			ApproverTitle:       "",
			ApproverRole:        approverRole,
			Decision:            req.Decision,
			DecisionReason:      req.DecisionReason,
			DecisionComments:    req.DecisionComments,
			ApprovedConditions:  "{}",
			ImposedRequirements: "{}",
			SupplementReplies:   ptrToJSONBString([]byte("{}")),
			FollowUpActions:     "[]",
			ApprovalDate:        time.Now(),
			EffectiveDate:       nil,
			NextReviewDate:      nil,
			SupportingDocuments: "[]",
			EvidenceReferences:  "[]",
			Status:              "active",
		}

		// 处理JSON字段
		// N3/C11: anchor the decision to the exact submission snapshot it
		// validated against, so evidence always carries its round version
		// (plan §5.3) without changing the old wire fields.
		anchor := map[string]interface{}{"decision_snapshot_id": currentSnapshot}
		for i, ref := range req.EvidenceReferences {
			anchor[fmt.Sprintf("client_ref_%d", i)] = ref
		}
		if req.ApprovedConditions != nil {
			if data, err := json.Marshal(req.ApprovedConditions); err == nil {
				record.ApprovedConditions = string(data)
			}
		}
		if req.ImposedRequirements != nil {
			if data, err := json.Marshal(req.ImposedRequirements); err == nil {
				record.ImposedRequirements = string(data)
			}
		}
		if len(req.FollowUpActions) > 0 {
			if data, err := json.Marshal(req.FollowUpActions); err == nil {
				record.FollowUpActions = string(data)
			}
		}
		if len(req.SupportingDocuments) > 0 {
			if data, err := json.Marshal(req.SupportingDocuments); err == nil {
				record.SupportingDocuments = string(data)
			}
		}
		if data, err := json.Marshal(anchor); err == nil {
			record.EvidenceReferences = string(data)
		}

		if err := s.approvalRepo.CreateRecordInTx(tx, record); err != nil {
			return err
		}

		// N4/C07: the approver's decision handoff closes with the decision
		// in the same transaction, so a business write and its visible todo
		// can never disagree about the current step.
		if err := closeApprovalWorkflowItems(tx, approval.ID, inboxPurposeApprovalDecision); err != nil {
			return err
		}

		// 根据决定处理状态
		switch req.Decision {
		case models.ApprovalDecisionApprove:
			// 检查是否有多级审批
			isComplete, err := s.assigner.MoveToNextStage(approval)
			if err != nil {
				return fmt.Errorf("移动到下一阶段失败: %v", err)
			}
			if !isComplete {
				// 还有下一级审批，提交后发送通知
				pendingNotice = "下一阶段审批"
			}

		case models.ApprovalDecisionReject:
			approval.Status = nextStatus
			approval.CurrentApproverID = ""
			approval.CurrentApproverName = ""
			pendingNotice = "审批被拒绝"

		case models.ApprovalDecisionRequestChanges:
			approval.Status = nextStatus
			pendingNotice = "需要修改"

		case models.ApprovalDecisionReassign:
			// 转派到其他审批人
			if req.NextApproverID != "" {
				// 验证新审批人存在
				newApprover, err := s.userRepo.FindByStringID(req.NextApproverID)
				if err != nil || newApprover == nil {
					return errors.New("指定的审批人不存在")
				}
				approval.CurrentApproverID = req.NextApproverID
				approval.CurrentApproverName = newApprover.Name
			}
		}

		approval.UpdatedBy = userID
		if err := s.approvalRepo.SaveInTx(tx, approval); err != nil {
			return err
		}

		// N4/C05+C06: the next actor's handoff commits with the decision.
		switch req.Decision {
		case models.ApprovalDecisionRequestChanges:
			if err := s.openApplicantRevisionTask(tx, approval); err != nil {
				return err
			}
		case models.ApprovalDecisionReassign:
			if err := s.openApprovalDecisionTask(tx, approval); err != nil {
				return err
			}
		}

		// Release the intake bind row in the SAME transaction: a failed
		// release rolls the decision back instead of leaving a terminal
		// approval that still holds the intake slot (fail-closed).
		if req.Decision == models.ApprovalDecisionReject ||
			req.Decision == models.ApprovalDecisionRequestChanges {
			s.terminalBinding.release(repositories.WithApprovalCaseTx(context.Background(), tx), approval.ID)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Notifications fire only after the decision is durably committed.
	switch pendingNotice {
	case "下一阶段审批", "审批被拒绝", "需要修改":
		_ = s.sendApprovalNotification(approval, pendingNotice)
	}

	// 重新加载数据
	updatedApproval, err := s.GetApproval(userID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("重新加载审批数据失败: %v", err)
	}

	return updatedApproval, nil
}

// ResubmitApproval 重新提交被拒绝或要求修改的审批
func (s *ApprovalService) ResubmitApproval(userID string, approvalID string, revisionNote string) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}
	if approval == nil {
		return nil, ErrApprovalRecordNotFound
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return nil, fmt.Errorf("无权重新提交此审批记录")
	}

	// 检查是否可以重新提交
	if !s.stateMachine.CanResubmit(approval.Status) {
		return nil, fmt.Errorf("当前状态 %s 不允许重新提交", approval.Status)
	}

	// 验证状态转换
	if err := s.stateMachine.ValidateTransition(approval.Status, models.ApprovalStatusResubmitted); err != nil {
		return nil, err
	}

	// 更新元数据，添加修改说明
	var metadata map[string]interface{}
	if approval.Metadata != "{}" && approval.Metadata != "" {
		json.Unmarshal([]byte(approval.Metadata), &metadata)
	} else {
		metadata = make(map[string]interface{})
	}
	metadata["revision_note"] = revisionNote
	metadata["resubmitted_at"] = time.Now().Format(time.RFC3339)
	metadataBytes, _ := json.Marshal(metadata)
	approval.Metadata = string(metadataBytes)
	approval.UpdatedBy = userID

	// Revival is atomic (audit-25f1d4b P1): the resubmitted status and the
	// intake rebind commit together or not at all. The previous best-effort
	// rebind left a revived approval whose case-creation capability was
	// never armed when a competing submission owned the UNIQUE slot — the
	// revived approval then created a second case for the same intake.
	// A UNIQUE rejection rolls everything back and surfaces the standard
	// 409 contract: the slot owner's approval is the one to continue on.
	rebind := func(tx *gorm.DB) error {
		if err := s.terminalBinding.rebindAfterResubmitInTx(tx, approval); err != nil {
			if asUniqueConstraintError(err) != nil {
				return &ApprovalConflictGateError{
					Code:    "INTAKE_APPROVAL_ALREADY_IN_PROGRESS",
					Message: "该接案已有进行中的成案审批，不能重新提交；请在原审批单上继续处理",
				}
			}
			return err
		}
		// Re-assign inside the transaction: the metadata/status writes above
		// only exist on this tx handle until commit, and the approver
		// assignment only touches approval fields in memory.
		if err := s.assigner.AssignApprover(approval); err != nil {
			return fmt.Errorf("重新分配审批人失败: %w", err)
		}
		approval.Status = models.ApprovalStatusResubmitted
		if err := tx.Save(approval).Error; err != nil {
			return fmt.Errorf("重新提交失败: %w", err)
		}
		// N4/C05+C07: legacy-path handoff symmetry with the structured
		// resubmit (close applicant todo, open approver todo for the new
		// round).
		if err := closeApprovalWorkflowItems(tx, approval.ID, inboxPurposeApplicantRevision); err != nil {
			return err
		}
		if err := s.openApprovalDecisionTask(tx, approval); err != nil {
			return err
		}
		return nil
	}
	if err := s.approvalRepo.Transaction(rebind); err != nil {
		var gateErr *ApprovalConflictGateError
		if errors.As(err, &gateErr) {
			return nil, gateErr
		}
		return nil, err
	}

	// 发送重新提交通知
	_ = s.sendApprovalNotification(approval, "重新提交")

	return s.approvalRepo.FindByID(approvalID)
}

// CancelApproval 撤回审批
func (s *ApprovalService) CancelApproval(userID string, approvalID string) error {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return fmt.Errorf("无权撤回此审批记录")
	}

	// 检查是否有审批记录（如果有则说明已经开始处理）
	records, err := s.approvalRepo.FindRecordsByApprovalID(approvalID)
	if err != nil {
		return fmt.Errorf("检查审批记录失败: %v", err)
	}

	// 检查是否可以撤回
	if !s.stateMachine.CanCancel(approval.Status, len(records) > 0) {
		return errors.New("当前状态不允许撤回或已有审批记录")
	}

	// 验证状态转换
	if err := s.stateMachine.ValidateTransition(approval.Status, models.ApprovalStatusCancelled); err != nil {
		return err
	}

	// 更新状态为已撤回
	approval.Status = models.ApprovalStatusCancelled
	approval.CurrentApproverID = ""
	approval.CurrentApproverName = ""
	approval.UpdatedBy = userID

	// N4/C07: the cancellation write and the handoff closure commit
	// together; a failed closure rolls the cancellation back so no open
	// approver todo outlives a cancelled approval.
	if err := s.approvalRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.approvalRepo.SaveInTx(tx, approval); err != nil {
			return err
		}
		return closeApprovalWorkflowItems(tx, approval.ID, inboxPurposeApprovalDecision)
	}); err != nil {
		return fmt.Errorf("撤回审批失败: %v", err)
	}

	// Cancelled is a state-machine terminal state: the only way back to a
	// live case-creation attempt is a NEW approval through the intake
	// submission, which owns the UNIQUE bind slot itself. Release the old
	// bind so that new submission is not blocked forever. Best-effort:
	// a failed release leaves the bind in place (fail-closed 409) and
	// never fails the cancellation itself.
	s.terminalBinding.release(context.Background(), approval.ID)

	return nil
}

// UpdateApproval 更新审批申请（草稿或需要修改状态）
func (s *ApprovalService) UpdateApproval(userID string, id string, req *models.UpdateApprovalRequest) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return nil, fmt.Errorf("无权更新此审批记录")
	}

	// 只能更新可编辑状态的审批
	if !s.stateMachine.CanEdit(approval.Status) {
		return nil, fmt.Errorf("当前状态 %s 不允许编辑", approval.Status)
	}

	// 更新字段
	if req.Title != "" {
		approval.Title = req.Title
	}
	if req.Content != "" {
		approval.Content = req.Content
	}

	// 处理附件和元数据
	if len(req.Attachments) > 0 {
		if attachmentsBytes, err := json.Marshal(req.Attachments); err == nil {
			approval.Attachments = string(attachmentsBytes)
		}
	}
	if req.Metadata != nil {
		if metadataBytes, err := json.Marshal(req.Metadata); err == nil {
			approval.Metadata = string(metadataBytes)
		}
	}

	approval.UpdatedBy = userID

	// 执行更新
	if err := s.approvalRepo.Update(approval); err != nil {
		return nil, fmt.Errorf("更新审批申请失败: %v", err)
	}

	// 重新加载数据
	return s.approvalRepo.FindByID(id)
}

// DeleteApproval 删除审批（仅草稿）
func (s *ApprovalService) DeleteApproval(userID string, id string) error {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return fmt.Errorf("无权删除此审批记录")
	}

	// 只能删除草稿状态的审批
	if approval != nil && approval.Status != models.ApprovalStatusDraft {
		return errors.New("只能删除草稿状态的审批")
	}

	if err := s.approvalRepo.Delete(id); err != nil {
		return fmt.Errorf("删除审批记录失败: %v", err)
	}

	return nil
}

// GetApprovalWorkflows 获取审批工作流列表
func (s *ApprovalService) GetApprovalWorkflows() ([]models.ApprovalWorkflow, error) {
	workflows, err := s.approvalRepo.FindWorkflows()
	if err != nil {
		return nil, fmt.Errorf("获取审批工作流失败: %v", err)
	}
	return workflows, nil
}

// GetApprovalTemplates 获取审批模板列表
func (s *ApprovalService) GetApprovalTemplates(templateType string, category string) ([]models.ApprovalTemplateV2, error) {
	templates, err := s.approvalRepo.FindTemplates(templateType, category)
	if err != nil {
		return nil, fmt.Errorf("获取审批模板失败: %v", err)
	}
	return templates, nil
}

// sendApprovalNotification 发送审批通知（简化版）
func (s *ApprovalService) sendApprovalNotification(approval *models.ApprovalRequest, message string) error {
	// TODO: 实现实际的通知发送逻辑
	// 可以通过WebSocket、邮件等方式发送通知
	fmt.Printf("通知发送: 申请 %s - %s\n", approval.RequestNumber, message)
	return nil
}

// generateRequestNumber 生成申请编号
func (s *ApprovalService) generateRequestNumber() (string, error) {
	timestamp := time.Now().Format("20060102")
	// 并发创建（如多个不同接案同时成案）可能落在同一时钟读数内，
	// 纳秒取模后缀会撞 request_number 唯一索引并把合法请求变成 500。
	// UUID 前 16 位十六进制提供 64 位熵，且不依赖时钟粒度。
	random := strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	return fmt.Sprintf("AP-%s-%s", timestamp, random), nil
}

// generateUUID 生成UUID
func generateUUID() string {
	return uuid.NewString()
}
