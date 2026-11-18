package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

// DeadlineHandler 时效管理处理器
type DeadlineHandler struct {
	calculator     *services.DeadlineCalculator
	authz          *services.AuthorizationService
	subjectRecheck *services.SubjectRecheckService
}

// NewDeadlineHandler 创建时效管理处理器
func NewDeadlineHandler(calculator *services.DeadlineCalculator, authz ...*services.AuthorizationService) *DeadlineHandler {
	h := &DeadlineHandler{calculator: calculator}
	if len(authz) > 0 {
		h.authz = authz[0]
	}
	return h
}

func (h *DeadlineHandler) SetSubjectRecheckService(service *services.SubjectRecheckService) {
	h.subjectRecheck = service
}

// CalculateDeadline godoc
// @Summary 计算期限
// @Description 根据案件类型和日期计算各类期限
// @Tags 时效管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body calculateDeadlineRequest true "期限计算请求"
// @Success 200 {object} common.APIResponse{data=calculateDeadlineResponse} "计算成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Router /deadlines/calculate [post]
func (h *DeadlineHandler) CalculateDeadline(c *gin.Context) {
	var req calculateDeadlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有必填字段")
		return
	}

	resp := calculateDeadlineResponse{}

	// 计算上诉期限
	if !req.JudgmentDate.IsZero() {
		appealDeadline, appealDays, err := h.calculator.CalculateAppealDeadline(req.JudgmentDate, req.CaseType)
		if err == nil {
			resp.AppealDeadline = appealDeadline
			resp.AppealDays = appealDays
		}
		// 计算执行申请期限
		executionDeadline, err := h.calculator.CalculateExecutionDeadline(req.JudgmentDate)
		if err == nil {
			resp.ExecutionDeadline = executionDeadline
		}
	}

	// 计算诉讼时效
	if !req.IncidentDate.IsZero() {
		solDeadline, solDays, err := h.calculator.CalculateStatuteOfLimitations(req.IncidentDate, req.CaseType)
		if err == nil {
			resp.StatuteOfLimitationsDeadline = solDeadline
			resp.StatuteOfLimitationsDays = solDays
		}
	}

	common.APISuccess(c, resp)
}

// GetCaseDeadlines godoc
// @Summary 获取案件期限信息
// @Description 获取案件的所有期限信息
// @Tags 时效管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param case_id path int true "案件ID"
// @Success 200 {object} common.APIResponse{data=services.DeadlineInfo} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Router /deadlines/cases/{case_id} [get]
func (h *DeadlineHandler) GetCaseDeadlines(c *gin.Context) {
	caseID, err := strconv.ParseUint(c.Param("case_id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "参数错误", "case_id必须是正整数")
		return
	}
	if !h.authorizeCase(c, uint(caseID), false) {
		return
	}

	info, err := h.calculator.GetCaseDeadlines(c.Request.Context(), uint(caseID))
	if err != nil {
		common.APINotFound(c, "获取案件期限失败", err.Error())
		return
	}

	common.APISuccess(c, info)
}

// CreateDeadlineReminder godoc
// @Summary 创建期限提醒
// @Description 为案件创建期限提醒待办事项
// @Tags 时效管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body createDeadlineReminderRequest true "创建期限提醒请求"
// @Success 200 {object} common.APIResponse "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Router /deadlines/reminders [post]
func (h *DeadlineHandler) CreateDeadlineReminder(c *gin.Context) {
	var req createDeadlineReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有必填字段")
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	uid := actor.UserID
	if !h.authorizeCase(c, req.CaseID, true) {
		return
	}
	if h.subjectRecheck == nil {
		common.NewAPIError(c, 503, "SUBJECT_GATE_UNAVAILABLE", "案件主体门禁未初始化，当前不能创建案件期限提醒")
		return
	}
	if err := h.subjectRecheck.RequireEffectiveSubject(c.Request.Context(), req.CaseID, "deadline_reminder_create"); err != nil {
		writeSubjectWorkflowError(c, err)
		return
	}

	switch req.DeadlineType {
	case "appeal":
		if req.JudgmentDate.IsZero() {
			common.APIBadRequest(c, "参数错误", "上诉期提醒需要提供judgment_date")
			return
		}
		err := h.calculator.CreateAppealDeadlineInbox(c.Request.Context(), req.CaseID, req.JudgmentDate, req.CaseType, uid)
		if err != nil {
			common.APIInternalServerError(c, "创建上诉期提醒失败", err.Error())
			return
		}
	case "statute_of_limitations":
		if req.IncidentDate.IsZero() {
			common.APIBadRequest(c, "参数错误", "诉讼时效提醒需要提供incident_date")
			return
		}
		err := h.calculator.CreateStatuteOfLimitationsInbox(c.Request.Context(), req.CaseID, req.IncidentDate, req.CaseType, uid)
		if err != nil {
			common.APIInternalServerError(c, "创建诉讼时效提醒失败", err.Error())
			return
		}
	case "evidence":
		err := h.calculator.CreateEvidenceDeadlineInbox(c.Request.Context(), req.CaseID, req.CaseType, uid)
		if err != nil {
			common.APIInternalServerError(c, "创建举证期限提醒失败", err.Error())
			return
		}
	case "execution":
		if req.JudgmentDate.IsZero() {
			common.APIBadRequest(c, "参数错误", "执行申请期限提醒需要提供judgment_date")
			return
		}
		err := h.calculator.CreateExecutionDeadlineInbox(c.Request.Context(), req.CaseID, req.JudgmentDate, uid)
		if err != nil {
			common.APIInternalServerError(c, "创建执行申请期限提醒失败", err.Error())
			return
		}
	default:
		common.APIBadRequest(c, "参数错误", "不支持的期限类型")
		return
	}

	common.APISuccess(c, gin.H{"message": "期限提醒创建成功"})
}

// GetDeadlineTypes godoc
// @Summary 获取支持的期限类型
// @Description 获取系统支持的所有期限类型和配置
// @Tags 时效管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse "获取成功"
// @Router /deadlines/types [get]
func (h *DeadlineHandler) GetDeadlineTypes(c *gin.Context) {
	types := []gin.H{
		{
			"type":        "appeal",
			"name":        "上诉期限",
			"description": "判决书送达后提起上诉的期限",
			"rules":       services.CaseTypeDeadlines,
		},
		{
			"type":        "evidence",
			"name":        "举证期限",
			"description": "提交证据材料的期限",
			"rules":       services.EvidenceDeadlines,
		},
		{
			"type":        "execution",
			"name":        "执行申请期限",
			"description": "判决生效后申请执行的期限（2年）",
		},
		{
			"type":        "statute_of_limitations",
			"name":        "诉讼时效",
			"description": "提起诉讼的有效期限",
			"rules": gin.H{
				"默认":     "3年",
				"身体伤害":   "1年",
				"商品质量":   "1年",
				"国际货物买卖": "4年",
			},
		},
	}
	common.APISuccess(c, types)
}

func (h *DeadlineHandler) authorizeCase(c *gin.Context, caseID uint, write bool) bool {
	if h.authz == nil {
		common.NewAPIError(c, 503, "CASE_AUTHZ_UNAVAILABLE", "案件权限服务未初始化")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	allowed, err := h.authz.CanReadCase(c.Request.Context(), actor, caseID)
	if write {
		allowed, err = h.authz.CanManageCase(c.Request.Context(), actor, caseID)
	}
	if err != nil {
		common.APIInternalServerError(c, "案件权限校验失败", err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

// 请求/响应结构体

type calculateDeadlineRequest struct {
	CaseType     string    `json:"case_type" binding:"required"`
	JudgmentDate time.Time `json:"judgment_date"`
	IncidentDate time.Time `json:"incident_date"`
}

type calculateDeadlineResponse struct {
	AppealDeadline               time.Time `json:"appeal_deadline,omitempty"`
	AppealDays                   int       `json:"appeal_days,omitempty"`
	ExecutionDeadline            time.Time `json:"execution_deadline,omitempty"`
	StatuteOfLimitationsDeadline time.Time `json:"statute_of_limitations_deadline,omitempty"`
	StatuteOfLimitationsDays     int       `json:"statute_of_limitations_days,omitempty"`
}

type createDeadlineReminderRequest struct {
	CaseID       uint      `json:"case_id" binding:"required"`
	DeadlineType string    `json:"deadline_type" binding:"required"`
	CaseType     string    `json:"case_type"`
	JudgmentDate time.Time `json:"judgment_date"`
	IncidentDate time.Time `json:"incident_date"`
}
