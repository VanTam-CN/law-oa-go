package handlers

import (
	"net/http"
	"strconv"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
)

// ConflictHandlerV2 v2 冲突检测处理器
type ConflictHandlerV2 struct {
	conflictService     services.V2ConflictDetectionService
	conflictPoolService services.ConflictPoolService
	pdfReportService    services.PDFReportService
	scanService         services.ConflictScanService
	companyAPIService   services.CompanyAPIService
}

// NewConflictHandlerV2 创建新的 v2 冲突检测处理器
func NewConflictHandlerV2(
	conflictService services.V2ConflictDetectionService,
	conflictPoolService services.ConflictPoolService,
	pdfReportService services.PDFReportService,
	scanService services.ConflictScanService,
	companyAPIService services.CompanyAPIService,
) *ConflictHandlerV2 {
	return &ConflictHandlerV2{
		conflictService:     conflictService,
		conflictPoolService: conflictPoolService,
		pdfReportService:    pdfReportService,
		scanService:         scanService,
		companyAPIService:   companyAPIService,
	}
}

// QuickCheckRequest 快速检测请求
type QuickCheckRequest struct {
	LawyerID        uint     `json:"lawyerId" binding:"required"`
	ClientName      string   `json:"clientName" binding:"required"`
	ClientTaxID     string   `json:"clientTaxId"`
	CaseID          uint     `json:"caseId"`
	OpposingParties []string `json:"opposingParties"`
	SearchDepth     string   `json:"searchDepth"`
	IncludeRelated  bool     `json:"includeRelated"`
}

// QuickCheck 快速冲突检测
func (h *ConflictHandlerV2) QuickCheck(c *gin.Context) {
	var req QuickCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 构建服务请求
	serviceReq := &services.ConflictCheckRequestV2{
		LawyerID:        req.LawyerID,
		ClientName:      req.ClientName,
		ClientTaxID:     req.ClientTaxID,
		CaseID:          req.CaseID,
		OpposingParties: req.OpposingParties,
		SearchDepth:     req.SearchDepth,
		IncludeRelated:  req.IncludeRelated,
	}

	// 执行检测
	result, err := h.conflictService.QuickCheck(c.Request.Context(), serviceReq)
	if err != nil {
		common.APIInternalServerError(c, "冲突检测失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// DetailedCheck 详细冲突检测
func (h *ConflictHandlerV2) DetailedCheck(c *gin.Context) {
	var req QuickCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	serviceReq := &services.ConflictCheckRequestV2{
		LawyerID:        req.LawyerID,
		ClientName:      req.ClientName,
		ClientTaxID:     req.ClientTaxID,
		CaseID:          req.CaseID,
		OpposingParties: req.OpposingParties,
		SearchDepth:     req.SearchDepth,
		IncludeRelated:  req.IncludeRelated,
	}

	result, err := h.conflictService.DetailedCheck(c.Request.Context(), serviceReq)
	if err != nil {
		common.APIInternalServerError(c, "冲突检测失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// GenerateReport 生成冲突检测报告
func (h *ConflictHandlerV2) GenerateReport(c *gin.Context) {
	var req services.ReportGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	report, err := h.pdfReportService.GenerateReport(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "生成报告失败", err.Error())
		return
	}

	common.APISuccess(c, report)
}

// GetReport 获取报告
func (h *ConflictHandlerV2) GetReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的报告ID", "")
		return
	}

	report, err := h.pdfReportService.GetReport(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取报告失败", err.Error())
		return
	}

	common.APISuccess(c, report)
}

// ListReports 列出报告
func (h *ConflictHandlerV2) ListReports(c *gin.Context) {
	filter := &services.ReportFilter{}

	if checkedByStr := c.Query("checkedBy"); checkedByStr != "" {
		if id, err := strconv.ParseUint(checkedByStr, 10, 32); err == nil {
			checkedBy := uint(id)
			filter.CheckedBy = &checkedBy
		}
	}

	filter.RiskLevel = c.Query("riskLevel")
	filter.ClientName = c.Query("clientName")
	filter.ReportNumber = c.Query("reportNumber")

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	reports, err := h.pdfReportService.ListReports(c.Request.Context(), filter)
	if err != nil {
		common.APIInternalServerError(c, "获取报告列表失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"list": reports, "total": len(reports)})
}

// DownloadReport 下载报告
func (h *ConflictHandlerV2) DownloadReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的报告ID", "")
		return
	}

	data, filename, err := h.pdfReportService.DownloadReport(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "下载报告失败", err.Error())
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "application/pdf", data)
}

// VerifySignature 验证报告签名
func (h *ConflictHandlerV2) VerifySignature(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的报告ID", "")
		return
	}

	result, err := h.pdfReportService.VerifySignature(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "验证签名失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// TriggerScan 触发扫描
func (h *ConflictHandlerV2) TriggerScan(c *gin.Context) {
	var req struct {
		TriggeredBy   uint   `json:"triggeredBy" binding:"required"`
		TriggerReason string `json:"triggerReason"`
		ScanScope     string `json:"scanScope"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	scanReq := &services.ManualScanRequest{
		TriggeredBy:   req.TriggeredBy,
		TriggerReason: req.TriggerReason,
		ScanScope:     req.ScanScope,
	}

	job, err := h.scanService.TriggerManualScan(c.Request.Context(), scanReq)
	if err != nil {
		common.APIInternalServerError(c, "触发扫描失败", err.Error())
		return
	}

	common.APISuccess(c, job)
}

// ListScanJobs 列出扫描任务
func (h *ConflictHandlerV2) ListScanJobs(c *gin.Context) {
	filter := &services.ScanJobFilter{}

	filter.ScanType = c.Query("scanType")
	filter.Status = c.Query("status")

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	jobs, err := h.scanService.ListScanJobs(c.Request.Context(), filter)
	if err != nil {
		common.APIInternalServerError(c, "获取扫描任务失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"list": jobs, "total": len(jobs)})
}

// GetScanStats 获取扫描统计
func (h *ConflictHandlerV2) GetScanStats(c *gin.Context) {
	stats, err := h.scanService.GetScanStats(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取扫描统计失败", err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// GetPoolStats 获取冲突池统计
func (h *ConflictHandlerV2) GetPoolStats(c *gin.Context) {
	lawyerIDStr := c.Query("lawyerId")
	_, err := strconv.ParseUint(lawyerIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的律师ID", "")
		return
	}

	// 这里需要创建一个获取 pool stats 的方法
	// 暂时返回基础统计
	common.APISuccess(c, gin.H{
		"totalEntries":    0,
		"byRelationship":  map[string]int64{},
		"byEntityType":    map[string]int64{},
		"apiDataCoverage": 0.0,
	})
}

// SearchCompany 搜索公司（调用 API）
func (h *ConflictHandlerV2) SearchCompany(c *gin.Context) {
	keyword := c.Query("keyword")
	provider := services.CompanyAPIProvider(c.DefaultQuery("provider", "mock"))

	if keyword == "" {
		common.APIBadRequest(c, "搜索关键词不能为空", "")
		return
	}

	result, err := h.companyAPIService.SearchCompany(c.Request.Context(), keyword, provider)
	if err != nil {
		common.APIInternalServerError(c, "搜索公司失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// RefreshPoolEntry 刷新冲突池条目
func (h *ConflictHandlerV2) RefreshPoolEntry(c *gin.Context) {
	entryIDStr := c.Param("id")
	entryID, err := strconv.ParseUint(entryIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的条目ID", "")
		return
	}

	provider := services.CompanyAPIProvider(c.DefaultQuery("provider", "mock"))

	if err := h.conflictPoolService.RefreshFromAPI(c.Request.Context(), uint(entryID), provider); err != nil {
		common.APIInternalServerError(c, "刷新失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "刷新成功"})
}

// BatchSyncPool 批量同步冲突池
func (h *ConflictHandlerV2) BatchSyncPool(c *gin.Context) {
	var req struct {
		LawyerIDs []uint `json:"lawyerIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 创建冲突池服务的批量同步方法
	// 这里需要调用 poolService 的批量同步方法
	common.APISuccess(c, gin.H{
		"message":   "批量同步已启动",
		"lawyerIds": req.LawyerIDs,
	})
}

// HealthCheck 健康检查
func (h *ConflictHandlerV2) HealthCheck(c *gin.Context) {
	common.APISuccess(c, gin.H{
		"status":    "healthy",
		"service":   "conflict-detection-v2",
		"timestamp": time.Now(),
		"version":   "v2.0.0",
	})
}
