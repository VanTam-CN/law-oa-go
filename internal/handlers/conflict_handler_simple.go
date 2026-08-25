package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConflictHandlerSimple 冲突检测处理器
type ConflictHandlerSimple struct {
	conflictService       services.ConflictDetectionService
	conflictReviewService services.ConflictReviewService
	asyncConflictService  services.AsyncConflictCheckService
	authz                 *services.AuthorizationService
	db                    *gorm.DB
}

// NewConflictHandlerSimple 创建新的冲突处理器
func NewConflictHandlerSimple(conflictService services.ConflictDetectionService, asyncServices ...services.AsyncConflictCheckService) *ConflictHandlerSimple {
	var asyncService services.AsyncConflictCheckService
	if len(asyncServices) > 0 {
		asyncService = asyncServices[0]
	}
	handler := &ConflictHandlerSimple{
		conflictService:      conflictService,
		asyncConflictService: asyncService,
	}
	if reviewService, ok := conflictService.(services.ConflictReviewService); ok {
		handler.conflictReviewService = reviewService
	}
	return handler
}

func (h *ConflictHandlerSimple) SetAuthorizationService(authz *services.AuthorizationService) {
	h.authz = authz
}

// SetDatabase wires the database used to resolve the immutable case/intake
// context of legacy history rows. A client_id alone is not an authorization
// boundary because one client may have both visible and wall-protected matters.
func (h *ConflictHandlerSimple) SetDatabase(db *gorm.DB) {
	h.db = db
}

// CheckConflict 执行冲突检查
func (h *ConflictHandlerSimple) CheckConflict(c *gin.Context) {
	log.Println("🔍 处理冲突检查请求")

	// 解析请求
	var request models.ConflictCheckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ 解析请求失败: %v", err)
		common.APIBadRequest(c, "请求参数错误", "请求体解析失败")
		return
	}
	if !h.prepareConflictRequest(c, &request) {
		return
	}

	// P0 requires a full search of the system archive. Do not let a browser
	// request silently reduce the evidence window.
	request.SearchYears = 0
	request.IncludeCorporateRelations = true
	if strings.EqualFold(strings.TrimSpace(request.SearchDepth), "BASIC") || request.SearchDepth == "" {
		request.SearchDepth = "STANDARD"
	}
	request.RequestTime = time.Now()

	log.Printf("📋 冲突检测请求: 客户ID=%s, 案件=%s, 律师ID=%s",
		request.ClientID, request.CaseName, request.UserID)

	// 执行冲突检测
	var result *models.ConflictCheckResponse
	var err error

	if h.conflictService != nil {
		result, err = h.conflictService.PerformConflictCheck(c.Request.Context(), &request)
		if err != nil {
			log.Printf("❌ 冲突检测失败: %v", err)
			var conflictErr *models.ConflictError
			if errors.As(err, &conflictErr) {
				common.APIBadRequest(c, conflictErr.Message, conflictErr.Code)
				return
			}
			common.APIInternalServerError(c, "冲突检测失败", err.Error())
			return
		}
	} else {
		log.Printf("❌ 冲突检测服务未初始化")
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_SERVICE_UNAVAILABLE", "冲突检测服务未初始化")
		return
	}

	projectConflictResponseForViewer(c, result)
	// 构建统一响应格式
	data := gin.H{
		"checkId":            result.CheckID,
		"hasConflict":        result.HasConflict,
		"conflictCases":      result.ConflictCases,
		"checkStatistics":    result.CheckStatistics,
		"riskAssessment":     result.RiskAssessment,
		"recommendations":    result.Recommendations,
		"checkTime":          result.CheckTime,
		"duration":           result.Duration,
		"normalizedSubjects": result.NormalizedSubjects,
		"decision":           result.Decision,
		"review":             result.Review,
	}

	log.Printf("✅ 冲突检查完成，检测到 %d 个冲突案例，风险等级: %s",
		len(result.ConflictCases), result.RiskAssessment.OverallRisk)

	common.APISuccess(c, data)
}

// CreateConflictTask 创建异步冲突检测任务
func (h *ConflictHandlerSimple) CreateConflictTask(c *gin.Context) {
	var request models.ConflictCheckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请求体解析失败")
		return
	}

	if !h.prepareConflictRequest(c, &request) {
		return
	}
	if h.asyncConflictService == nil {
		common.APIInternalServerError(c, "异步冲突检测服务未初始化")
		return
	}

	task, err := h.asyncConflictService.CreateTask(c.Request.Context(), &request)
	if err != nil {
		common.APIBadRequest(c, "创建冲突检测任务失败", err.Error())
		return
	}
	if task == nil || !h.canAccessConflictTask(c, task) {
		common.NewAPIError(c, http.StatusForbidden, "CONFLICT_TASK_FORBIDDEN", "无权访问或创建该冲突检测任务")
		return
	}

	common.APISuccess(c, task)
}

// GetConflictTaskStatus 获取异步冲突检测任务状态
func (h *ConflictHandlerSimple) GetConflictTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		common.APIBadRequest(c, "任务ID不能为空")
		return
	}
	if h.asyncConflictService == nil {
		common.APIInternalServerError(c, "异步冲突检测服务未初始化")
		return
	}

	task, err := h.asyncConflictService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, services.ErrConflictTaskNotFound) {
			common.APINotFound(c, "任务不存在", "指定的冲突检测任务不存在")
			return
		}
		common.APIInternalServerError(c, "获取任务状态失败", err.Error())
		return
	}
	if !h.canAccessConflictTask(c, task) {
		common.NewAPIError(c, http.StatusForbidden, "CONFLICT_TASK_FORBIDDEN", "无权访问其他律师的冲突检测任务")
		return
	}

	common.APISuccess(c, task)
}

// GetConflictTaskResult 获取异步冲突检测任务结果
func (h *ConflictHandlerSimple) GetConflictTaskResult(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		common.APIBadRequest(c, "任务ID不能为空")
		return
	}
	if h.asyncConflictService == nil {
		common.APIInternalServerError(c, "异步冲突检测服务未初始化")
		return
	}

	result, err := h.asyncConflictService.GetTaskResult(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, services.ErrConflictTaskNotFound) {
			common.APINotFound(c, "任务不存在", "指定的冲突检测任务不存在")
			return
		}
		common.APIInternalServerError(c, "获取任务结果失败", err.Error())
		return
	}
	if result.Task == nil || !h.canAccessConflictTask(c, result.Task) {
		common.NewAPIError(c, http.StatusForbidden, "CONFLICT_TASK_FORBIDDEN", "无权访问其他律师的冲突检测结果")
		return
	}
	projectConflictTaskResultForViewer(c, result)

	common.APISuccess(c, result)
}

func (h *ConflictHandlerSimple) prepareConflictRequest(c *gin.Context, request *models.ConflictCheckRequest) bool {
	if !canRunConflictCheck(c) {
		common.NewAPIError(c, http.StatusForbidden, "CONFLICT_CHECK_FORBIDDEN", "仅执业律师或独立冲突核查人可以运行利益冲突检查")
		return false
	}
	if strings.TrimSpace(request.IntakeID) != "" {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_CONFLICT_CHECK_REQUIRED", "接案冲突检查必须从接案工作台发起，并先完成律师事实确认")
		return false
	}
	subjectCaseID := strings.TrimSpace(request.SubjectCaseID)
	if subjectCaseID == "" {
		common.NewAPIError(c, http.StatusConflict, "CASE_CONTEXT_REQUIRED", "冲突检查必须绑定正式案件上下文，请从立案工作台发起")
		return false
	}
	if h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_AUTHZ_UNAVAILABLE", "案件冲突上下文权限服务未初始化，已阻止冲突检测")
		return false
	}
	caseID, err := strconv.ParseUint(subjectCaseID, 10, 32)
	if err != nil || caseID == 0 {
		common.APIBadRequest(c, "案件上下文无效", "subjectCaseId必须是有效的案件ID")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	allowed, authErr := h.authz.CanReadConflictContext(c.Request.Context(), actor, uint(caseID))
	if authErr != nil {
		common.APIInternalServerError(c, "案件权限校验失败", authErr.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	jwtUserID, exists := c.Get("user_id")
	if !exists {
		if request.UserID == "" {
			common.APIUnauthorized(c, "未授权访问", "缺少用户认证信息")
			return false
		}
	} else {
		var jwtUserIDStr string
		if uid, ok := jwtUserID.(float64); ok {
			jwtUserIDStr = strconv.FormatUint(uint64(uid), 10)
		} else if uid, ok := jwtUserID.(uint); ok {
			jwtUserIDStr = strconv.FormatUint(uint64(uid), 10)
		} else if uid, ok := jwtUserID.(string); ok {
			jwtUserIDStr = uid
		}

		if request.UserID == "" {
			request.UserID = jwtUserIDStr
		} else if request.UserID != jwtUserIDStr && !canRunConflictCheckForOthers(c) {
			log.Printf("🚫 冲突检测被拒绝: 登录用户=%s, 请求检查律师=%s, 角色=%v", jwtUserIDStr, request.UserID, c.GetString("role"))
			common.NewAPIError(c, http.StatusForbidden, "CONFLICT_LAWYER_SCOPE_FORBIDDEN", "普通律师只能以本人作为承办律师执行冲突检查")
			return false
		}
		if parsed, err := strconv.ParseUint(jwtUserIDStr, 10, 32); err == nil {
			request.ActorUserID = uint(parsed)
		}
		request.ActorRole = strings.ToLower(strings.TrimSpace(c.GetString("role")))
	}

	if h.db == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_CONTEXT_UNAVAILABLE", "无法读取案件主体上下文，已阻止冲突检测")
		return false
	}

	var caseRecord struct {
		ID       uint
		ClientID uint
		LawyerID uint
		Title    string
		CaseType string
	}
	if err := h.db.WithContext(c.Request.Context()).
		Table("cases").
		Select("id, client_id, lawyer_id, title, case_type").
		Where("id = ? AND deleted_at IS NULL", uint(caseID)).
		Take(&caseRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.NewAPIError(c, http.StatusConflict, "CASE_CONTEXT_INVALID", "案件主体上下文不存在，已阻止冲突检测")
			return false
		}
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_CONTEXT_UNAVAILABLE", "无法读取案件主体上下文，已阻止冲突检测")
		return false
	}
	requestedClientID, clientErr := strconv.ParseUint(strings.TrimSpace(request.ClientID), 10, 32)
	requestedLawyerID, lawyerErr := strconv.ParseUint(strings.TrimSpace(request.UserID), 10, 32)
	if clientErr != nil || requestedClientID == 0 || lawyerErr != nil || requestedLawyerID == 0 {
		common.NewAPIError(c, http.StatusConflict, "CONFLICT_SUBJECT_BINDING_INVALID", "冲突检测必须绑定有效的案件客户和承办律师")
		return false
	}
	if uint(requestedClientID) != caseRecord.ClientID {
		common.NewAPIError(c, http.StatusConflict, "CONFLICT_CLIENT_SCOPE_MISMATCH", "冲突检测客户与案件客户不一致，已阻止检测")
		return false
	}
	if uint(requestedLawyerID) != caseRecord.LawyerID {
		common.NewAPIError(c, http.StatusConflict, "CONFLICT_LAWYER_SCOPE_MISMATCH", "冲突检测承办律师与案件承办律师不一致，已阻止检测")
		return false
	}
	if strings.TrimSpace(caseRecord.Title) == "" || strings.TrimSpace(caseRecord.CaseType) == "" {
		common.NewAPIError(c, http.StatusConflict, "CASE_CONTEXT_INVALID", "案件缺少可追溯的名称或类型，已阻止冲突检测")
		return false
	}
	var clientRecord models.Client
	if err := h.db.WithContext(c.Request.Context()).
		Table("clients").
		Select("id, name, type, status, id_card, id_card_digest, id_card_ciphertext, identity_type, identity_number_digest, identity_number_ciphertext, aliases").
		Where("id = ? AND deleted_at IS NULL", caseRecord.ClientID).
		Take(&clientRecord).Error; err != nil {
		common.NewAPIError(c, http.StatusConflict, "CLIENT_CONTEXT_INVALID", "案件客户主体不存在，已阻止冲突检测")
		return false
	}
	if strings.TrimSpace(clientRecord.Name) == "" {
		common.NewAPIError(c, http.StatusConflict, "CLIENT_CONTEXT_INVALID", "案件客户主体缺少可检索名称，已阻止冲突检测")
		return false
	}

	// The database owns the subject binding and audit labels. Never persist
	// browser-supplied names or a stale case type as conflict evidence.
	request.ClientName = clientRecord.Name
	request.ClientType = normalizeConflictClientType(clientRecord.Type)
	request.ClientIdentifiers = nil
	if identityNumber, err := clientRecord.DecryptedIdentity(); err != nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CLIENT_IDENTITY_UNAVAILABLE", "案件客户身份标识无法安全读取，已阻止冲突检测")
		return false
	} else if strings.TrimSpace(identityNumber) != "" {
		request.ClientIdentifiers = map[string]string{clientRecord.IdentityIdentifierKey(): identityNumber}
	} else {
		common.NewAPIError(c, http.StatusConflict, "CLIENT_IDENTITY_REQUIRED", "案件客户缺少可核验身份标识，已阻止冲突检测")
		return false
	}
	request.ClientAliases = splitConflictAliases(clientRecord.Aliases)
	request.CaseName = caseRecord.Title
	request.CaseType = caseRecord.CaseType
	request.SearchYears = 0
	if request.SearchDepth == "" {
		request.SearchDepth = "STANDARD"
	}
	request.RequestTime = time.Now()
	return true
}

func normalizeConflictClientType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "企业", "公司", "COMPANY", "LEGAL_PERSON", "ORGANIZATION":
		return "COMPANY"
	case "个人", "PERSON", "INDIVIDUAL":
		return "PERSON"
	default:
		return "ANY"
	}
}

func splitConflictAliases(value string) []string {
	items := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '；' || r == '\n'
	})
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func canRunConflictCheckForOthers(c *gin.Context) bool {
	return services.IsConflictReviewRole(c.GetString("role"))
}

func canRunConflictCheck(c *gin.Context) bool {
	role := strings.ToLower(strings.TrimSpace(c.GetString("role")))
	return role == "lawyer" || services.IsConflictReviewRole(role)
}

func (h *ConflictHandlerSimple) canAccessConflictTask(c *gin.Context, task *services.ConflictCheckTaskResponse) bool {
	return canAccessConflictTaskForActor(c, h.authz, task)
}

func canAccessConflictTaskForActor(c *gin.Context, authz *services.AuthorizationService, task *services.ConflictCheckTaskResponse) bool {
	if task == nil {
		return false
	}
	actorID, ok := conflictActorID(c)
	if !ok {
		return false
	}
	subjectCaseID := strings.TrimSpace(task.SubjectCaseID)
	if authz == nil && (subjectCaseID != "" || strings.TrimSpace(task.IntakeID) != "") {
		return false
	}
	if subjectCaseID == "" || subjectCaseID == "0" {
		if strings.TrimSpace(task.IntakeID) != "" {
			actor, actorOK := currentAuthActor(c)
			if !actorOK {
				return false
			}
			allowed, authErr := authz.CanReadConflictIntakeContext(c.Request.Context(), actor, task.IntakeID, task.OwnerID, task.ClientID)
			if authErr != nil || !allowed {
				return false
			}
		} else if canRunConflictCheckForOthers(c) {
			// An independent reviewer may not browse or act on a legacy global
			// record whose case/intake context cannot be checked against the
			// ethical wall.
			return false
		}
	}
	if canRunConflictCheckForOthers(c) && (subjectCaseID == "" || subjectCaseID == "0") && strings.TrimSpace(task.IntakeID) == "" {
		// An independent reviewer may not browse or act on a legacy global
		// record whose case context cannot be checked against the ethical wall.
		return false
	}
	if subjectCaseID != "" && subjectCaseID != "0" {
		if authz == nil {
			return false
		}
		caseID, err := strconv.ParseUint(subjectCaseID, 10, 32)
		if err != nil || caseID == 0 {
			return false
		}
		actor, actorOK := currentAuthActor(c)
		if !actorOK {
			return false
		}
		allowed, authErr := authz.CanReadConflictContext(c.Request.Context(), actor, uint(caseID))
		if authErr != nil || !allowed {
			return false
		}
	}
	return actorID == task.OwnerID || canRunConflictCheckForOthers(c)
}

type conflictReviewRequest struct {
	Decision     string     `json:"decision" binding:"required"`
	Notes        string     `json:"notes" binding:"required"`
	NextReviewAt *time.Time `json:"nextReviewAt"`
}

// ReviewConflict records an immutable professional conclusion over frozen evidence.
func (h *ConflictHandlerSimple) ReviewConflict(c *gin.Context) {
	if h.conflictReviewService == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_REVIEW_UNAVAILABLE", "冲突复核服务未初始化")
		return
	}
	if !canReviewConflict(c) {
		common.NewAPIError(c, http.StatusForbidden, "CONFLICT_REVIEW_FORBIDDEN", "仅独立冲突核查人或获授权业务管理人员可以提交复核结论")
		return
	}
	if h.asyncConflictService == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_TASK_UNAVAILABLE", "冲突任务服务未初始化，已阻止提交复核")
		return
	}
	task, err := h.asyncConflictService.GetTask(c.Request.Context(), c.Param("task_id"))
	if err != nil {
		if errors.Is(err, services.ErrConflictTaskNotFound) {
			common.APINotFound(c, "冲突检测记录不存在", "check not found")
			return
		}
		common.APIInternalServerError(c, "获取冲突检测上下文失败", err.Error())
		return
	}
	if !h.canAccessConflictTask(c, task) {
		common.NewAPIError(c, http.StatusForbidden, "CONFLICT_TASK_FORBIDDEN", "无权访问或复核该冲突检测任务")
		return
	}
	var request conflictReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.APIBadRequest(c, "复核结论和意见均为必填项", "invalid conflict review payload")
		return
	}
	reviewerID, ok := conflictActorID(c)
	if !ok {
		common.APIUnauthorized(c, "无法识别复核人员", "missing authenticated user id")
		return
	}
	reviewerName := strings.TrimSpace(c.GetString("username"))
	if reviewerName == "" {
		reviewerName = fmt.Sprintf("用户%d", reviewerID)
	}
	review, err := h.conflictReviewService.ReviewConflict(c.Request.Context(), c.Param("task_id"), request.Decision, request.Notes, reviewerID, reviewerName, request.NextReviewAt)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.APINotFound(c, "冲突检测记录不存在", "check not found")
			return
		}
		var reviewerErr *services.ConflictReviewerError
		if errors.As(err, &reviewerErr) {
			status := http.StatusConflict
			if strings.HasSuffix(reviewerErr.Code, "FORBIDDEN") || reviewerErr.Code == "REVIEWER_ROLE_FORBIDDEN" || reviewerErr.Code == "REVIEWER_INACTIVE" || reviewerErr.Code == "REVIEWER_NOT_FOUND" {
				status = http.StatusForbidden
			}
			common.NewAPIError(c, status, reviewerErr.Code, reviewerErr.Message)
			return
		}
		common.APIBadRequest(c, "提交冲突复核失败", err.Error())
		return
	}
	common.APISuccess(c, review)
}

func (h *ConflictHandlerSimple) GetConflictReview(c *gin.Context) {
	if h.conflictReviewService == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_REVIEW_UNAVAILABLE", "冲突复核服务未初始化")
		return
	}
	if h.asyncConflictService == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_TASK_UNAVAILABLE", "冲突任务服务未初始化")
		return
	}
	task, err := h.asyncConflictService.GetTask(c.Request.Context(), c.Param("task_id"))
	if err != nil || task == nil {
		common.APINotFound(c, "冲突检测记录不存在", "check not found")
		return
	}
	if !h.canAccessConflictTask(c, task) {
		common.NewAPIError(c, http.StatusForbidden, "CONFLICT_TASK_FORBIDDEN", "无权访问其他律师的冲突复核记录")
		return
	}
	review, err := h.conflictReviewService.GetConflictReview(c.Request.Context(), c.Param("task_id"))
	if err != nil {
		common.APIInternalServerError(c, "获取冲突复核失败", err.Error())
		return
	}
	if review != nil && !canReviewConflict(c) {
		review.Notes = "复核结论已形成；涉及历史事项的依据仅向获授权核查人披露。"
		review.EvidenceHash = ""
	}
	common.APISuccess(c, review)
}

func canReviewConflict(c *gin.Context) bool {
	return services.IsConflictReviewRole(c.GetString("role"))
}

// projectConflictResponseForViewer applies the server-side disclosure rule.
// A requesting lawyer may see the involved subject and conflict type for a
// non-isolated candidate, but never the historical matter or working product.
// An ethical-wall hit is projected to a generic notice only.
func projectConflictResponseForViewer(c *gin.Context, response *models.ConflictCheckResponse) {
	if response == nil || canReviewConflict(c) {
		return
	}
	// Historical records may predate the P0 policy and contain a machine-level
	// BLOCKED/CLEAR value. A requesting lawyer must never treat that value as a
	// professional conclusion; the independent review record remains authoritative.
	if response.Decision == nil {
		response.Decision = &models.ConflictDecisionSummary{}
	}
	response.Decision.Status = "REVIEW_REQUIRED"
	response.Decision.RequiresManualReview = true
	response.Decision.RuleCodes = nil
	response.Decision.EvidenceCount = 0
	response.Decision.RestrictedCount = 0
	for index := range response.NormalizedSubjects {
		response.NormalizedSubjects[index].Identifiers = maskConflictIdentifiers(response.NormalizedSubjects[index].Identifiers)
	}
	response.Review = nil
	restrictedCount := 0
	for _, conflict := range response.ConflictCases {
		if conflict == nil {
			continue
		}
		if conflict.Restricted || conflictHasRestrictedEvidence(conflict) {
			restrictedCount++
		}
	}
	response.Decision.RestrictedCount = restrictedCount

	notice := "检索已完成，但需独立冲突核查人确认档案覆盖和主体信息完整性后才能形成无冲突结论。"
	if restrictedCount > 0 {
		notice = "存在受隔离记录，请联系独立冲突核查人；确认前不得视为已确认冲突或无冲突。"
	} else if len(response.ConflictCases) > 0 {
		notice = "发现潜在主体匹配，请联系独立冲突核查人确认主体身份、角色和处置。"
	}
	response.Decision.Recommendation = notice
	response.Recommendations = []string{notice}
	response.RiskAssessment = &models.RiskAssessment{
		OverallRisk:      "REVIEW_REQUIRED",
		RiskReason:       notice,
		RequiresApproval: true,
		RiskFactors:      []string{"独立冲突核查尚未完成"},
		Mitigation:       []string{"联系独立冲突核查人"},
	}
	response.CheckStatistics = &models.CheckStatistics{
		TimeRange:   "已按律所规则检索",
		SearchScope: "需独立复核",
	}
	for _, conflict := range response.ConflictCases {
		if conflict == nil {
			continue
		}
		if conflict.Restricted || conflictHasRestrictedEvidence(conflict) {
			projectRestrictedConflictCase(conflict)
			continue
		}

		// The requester may see which subject and conflict category require a
		// conversation, but not the historical matter, lawyer, status, or source.
		conflict.ID = ""
		conflict.CheckID = ""
		conflict.CaseID = ""
		conflict.CaseNo = ""
		conflict.CaseName = ""
		conflict.CaseType = ""
		conflict.RiskLevel = ""
		conflict.ClientID = ""
		conflict.CaseStatus = ""
		conflict.CreatedAt = time.Time{}
		conflict.OpposingParties = nil
		conflict.MatchType = ""
		conflict.RuleCode = ""
		conflict.RequiresManualReview = true
		conflict.Description = "发现潜在主体匹配，请联系独立冲突核查人确认主体身份、角色和处置。"
		conflict.ConflictDetails = "仅显示冲突类型和相关主体；历史案件详情需由独立冲突核查人查看。"
		for index := range conflict.Evidence {
			evidence := &conflict.Evidence[index]
			evidence.EvidenceID = ""
			evidence.RuleCode = ""
			evidence.MatchType = ""
			evidence.SourceType = ""
			evidence.PartyRole = ""
			evidence.HistoricalRole = ""
			evidence.SourceCaseID = ""
			evidence.SourceCaseNumber = ""
			evidence.SourceCaseName = ""
			evidence.LawyerName = ""
			evidence.SourceUpdatedAt = time.Time{}
			evidence.Summary = "发现相关主体，请联系独立冲突核查人确认。"
		}
	}
}

func conflictHasRestrictedEvidence(conflict *models.ConflictCase) bool {
	if conflict == nil {
		return false
	}
	for _, evidence := range conflict.Evidence {
		if evidence.Restricted {
			return true
		}
	}
	return false
}

func projectRestrictedConflictCase(conflict *models.ConflictCase) {
	conflict.Restricted = true
	conflict.ID = ""
	conflict.CheckID = ""
	conflict.CaseID = ""
	conflict.CaseNo = "受限"
	conflict.CaseName = "受限历史事项"
	conflict.CaseType = ""
	conflict.ConflictType = "受限历史记录"
	conflict.RiskLevel = ""
	conflict.ClientID = ""
	conflict.CaseStatus = ""
	conflict.CreatedAt = time.Time{}
	conflict.OpposingParties = nil
	conflict.MatchType = ""
	conflict.RuleCode = ""
	conflict.Description = "存在受隔离记录，请联系独立冲突核查人。"
	conflict.ConflictDetails = "存在受隔离记录，请联系独立冲突核查人。"
	for index := range conflict.Evidence {
		evidence := &conflict.Evidence[index]
		evidence.Restricted = true
		evidence.EvidenceID = ""
		evidence.RuleCode = ""
		evidence.MatchType = ""
		evidence.SourceType = ""
		evidence.RequestedParty = ""
		evidence.MatchedEntity = ""
		evidence.PartyRole = ""
		evidence.HistoricalRole = ""
		evidence.SourceCaseID = ""
		evidence.SourceCaseNumber = ""
		evidence.SourceCaseName = ""
		evidence.LawyerName = ""
		evidence.SourceUpdatedAt = time.Time{}
		evidence.Summary = "存在受隔离记录，请联系独立冲突核查人。"
	}
}

func maskConflictIdentifiers(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	masked := make(map[string]string, len(values))
	for key := range values {
		masked[key] = "已登记（完整标识仅向获授权冲突核查人显示）"
	}
	return masked
}

func projectConflictTaskResultForViewer(c *gin.Context, result *services.ConflictCheckTaskResultResponse) {
	if result == nil || len(result.Result) == 0 || canReviewConflict(c) {
		return
	}
	raw, err := json.Marshal(result.Result)
	if err != nil {
		return
	}
	var response models.ConflictCheckResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return
	}
	projectConflictResponseForViewer(c, &response)
	projected, err := json.Marshal(response)
	if err != nil {
		return
	}
	var projectedResult models.JSON
	if err := json.Unmarshal(projected, &projectedResult); err == nil {
		result.Result = projectedResult
	}
}

func conflictActorID(c *gin.Context) (uint, bool) {
	value, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	switch id := value.(type) {
	case uint:
		return id, id > 0
	case float64:
		return uint(id), id > 0
	case string:
		parsed, err := strconv.ParseUint(id, 10, 32)
		return uint(parsed), err == nil && parsed > 0
	default:
		return 0, false
	}
}

// GetCheckHistory 获取冲突检测历史
func (h *ConflictHandlerSimple) GetCheckHistory(c *gin.Context) {
	clientID := c.Query("clientId")
	if clientID == "" {
		common.APIBadRequest(c, "客户ID不能为空", "查询参数clientId是必需的")
		return
	}
	if !h.canAccessConflictClient(c, clientID) {
		return
	}

	// 获取历史记录
	history, err := h.conflictService.GetCheckHistory(c.Request.Context(), clientID, 10)
	if err != nil {
		log.Printf("❌ 获取冲突检测历史失败: %v", err)
		common.APIInternalServerError(c, "获取历史记录失败", err.Error())
		return
	}
	history = h.filterConflictHistoryByContext(c, history)

	projectConflictHistoryForViewer(c, history)
	common.APISuccess(c, history)
}

// filterConflictHistoryByContext prevents the compatibility client-history
// endpoint from returning records from another matter that happens to share
// the same client_id. The modern task endpoints carry an explicit case or
// intake context; legacy rows without one are excluded rather than guessed.
func (h *ConflictHandlerSimple) filterConflictHistoryByContext(c *gin.Context, history []*models.ConflictCheckRecord) []*models.ConflictCheckRecord {
	if h == nil || h.db == nil || h.authz == nil {
		return nil
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return nil
	}
	visible := make([]*models.ConflictCheckRecord, 0, len(history))
	for _, record := range history {
		if record == nil {
			continue
		}
		contextValue, err := services.ResolveConflictSubjectContext(c.Request.Context(), h.db, record)
		if err != nil {
			continue
		}
		allowed := false
		if contextValue.CaseID > 0 {
			allowed, err = h.authz.CanReadConflictContext(c.Request.Context(), actor, contextValue.CaseID)
		} else if contextValue.IntakeID != "" {
			allowed, err = h.authz.CanReadConflictIntakeContext(c.Request.Context(), actor, contextValue.IntakeID, record.UserID, record.ClientID)
		}
		if err == nil && allowed {
			visible = append(visible, record)
		}
	}
	return visible
}

func projectConflictHistoryForViewer(c *gin.Context, history []*models.ConflictCheckRecord) {
	if canReviewConflict(c) {
		return
	}
	for _, record := range history {
		if record == nil || len(record.CheckResult) == 0 {
			continue
		}
		raw, err := json.Marshal(record.CheckResult)
		if err != nil {
			continue
		}
		var response models.ConflictCheckResponse
		if err := json.Unmarshal(raw, &response); err != nil {
			continue
		}
		projectConflictResponseForViewer(c, &response)
		if projected, err := json.Marshal(response); err == nil {
			_ = json.Unmarshal(projected, &record.CheckResult)
		}
	}
}

// GetConflictStats 获取冲突检测统计
func (h *ConflictHandlerSimple) GetConflictStats(c *gin.Context) {
	clientID := c.Query("clientId")
	if clientID == "" && services.IsPrivilegedRole(c.GetString("role")) && !canReviewConflict(c) {
		forbidObjectAccess(c)
		return
	}
	if clientID == "" && !canReviewConflict(c) {
		forbidObjectAccess(c)
		return
	}
	if clientID != "" && !h.canAccessConflictClient(c, clientID) {
		return
	}
	// The compatibility statistics repository aggregates by client_id and has
	// no case/intake dimension. Returning it would reveal counts from a wall-
	// protected matter that shares the same client. The case-bound workbench is
	// the supported path until a context-aware aggregate query is available.
	common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_CLIENT_STATS_UNAVAILABLE", "旧版客户冲突统计缺少案件上下文，当前不可用；请从具体冲突任务查看结果")
}

// HealthCheck 冲突检查服务健康检查
func (h *ConflictHandlerSimple) HealthCheck(c *gin.Context) {
	result := gin.H{
		"status":    "healthy",
		"service":   "conflict-check",
		"timestamp": time.Now(),
		"version":   "v1.0.0",
	}

	common.APISuccess(c, result)
}

func (h *ConflictHandlerSimple) canAccessConflictClient(c *gin.Context, clientID string) bool {
	if services.IsPrivilegedRole(c.GetString("role")) && !canReviewConflict(c) {
		forbidObjectAccess(c)
		return false
	}
	if h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_AUTHZ_UNAVAILABLE", "冲突数据权限服务未初始化，当前已阻止历史记录访问")
		return false
	}
	if canReviewConflict(c) {
		// The legacy history/statistics service is client-scoped and cannot
		// prove which wall-protected matter a reviewer is allowed to inspect.
		// Do not let a client_id query become a firm-wide bypass; the modern
		// task endpoint carries a verifiable subject-case context instead.
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_CLIENT_HISTORY_UNAVAILABLE", "独立复核请从带案件上下文的冲突任务进入，旧版客户历史接口当前不可用")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	parsedClientID, err := strconv.ParseUint(clientID, 10, 64)
	if err != nil || parsedClientID == 0 {
		common.APIBadRequest(c, "客户ID无效", "clientId必须是有效的数字")
		return false
	}
	allowed, err := h.authz.CanReadClient(c.Request.Context(), actor, uint(parsedClientID))
	if err != nil {
		common.APIInternalServerError(c, "权限校验失败", err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}
