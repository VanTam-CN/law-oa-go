package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"
)

// DocumentStatsHandler handles document statistics API endpoints
type DocumentStatsHandler struct {
	statsService *services.DocumentStatsService
}

// NewDocumentStatsHandler creates a new document stats handler
func NewDocumentStatsHandler(statsService *services.DocumentStatsService) *DocumentStatsHandler {
	return &DocumentStatsHandler{
		statsService: statsService,
	}
}

func (h *DocumentStatsHandler) requireManagement(c *gin.Context) bool {
	role, _ := middleware.GetCurrentRole(c)
	if !services.IsBusinessMatterManagementRole(role) {
		common.NewAPIError(c, http.StatusForbidden, "DOCUMENT_AGGREGATE_FORBIDDEN", "文档聚合统计仅限业务管理角色查看")
		return false
	}
	return true
}

// unavailable prevents the legacy statistics implementation from returning
// fabricated/estimated values. A real deployment must wire these endpoints to
// audited storage and activity metrics before exposing them to users.
func (h *DocumentStatsHandler) unavailable(c *gin.Context) {
	common.NewAPIError(c, http.StatusServiceUnavailable, "DOCUMENT_STATS_UNAVAILABLE", "文档统计尚未接入真实审计与存储指标，当前不可用")
}

// GetOverview returns document overview statistics
// @Summary Get document overview statistics
// @Description Returns comprehensive document overview statistics including counts, trends, and user activity
// @Tags document-stats
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} services.DocumentOverviewStats
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/documents/stats/overview [get]
func (h *DocumentStatsHandler) GetOverview(c *gin.Context) {
	if !h.requireManagement(c) {
		return
	}
	h.unavailable(c)
}

// GetStorageUsage returns storage usage statistics
// @Summary Get storage usage statistics
// @Description Returns detailed storage usage statistics including breakdown by category and file type
// @Tags document-stats
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} services.StorageUsageStats
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/documents/stats/storage [get]
func (h *DocumentStatsHandler) GetStorageUsage(c *gin.Context) {
	if !h.requireManagement(c) {
		return
	}
	h.unavailable(c)
}

// GetUserActivity returns user activity statistics
// @Summary Get user activity statistics
// @Description Returns activity statistics for a specific user
// @Tags document-stats
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path int true "User ID"
// @Success 200 {object} services.UserActivityStats
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/documents/stats/users/{user_id} [get]
func (h *DocumentStatsHandler) GetUserActivity(c *gin.Context) {
	if !h.requireManagement(c) {
		return
	}
	h.unavailable(c)
}

// GetComplianceReport returns compliance report
// @Summary Get compliance report
// @Description Returns compliance-related statistics and reports
// @Tags document-stats
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} services.ComplianceReport
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/documents/stats/compliance [get]
func (h *DocumentStatsHandler) GetComplianceReport(c *gin.Context) {
	if !h.requireManagement(c) {
		return
	}
	h.unavailable(c)
}

// ExportStats exports statistics in specified format
// @Summary Export statistics
// @Description Exports statistics in specified format (json, csv)
// @Tags document-stats
// @Accept json
// @Produce application/octet-stream
// @Security ApiKeyAuth
// @Param type query string true "Statistics type" Enums(overview, storage, compliance)
// @Param format query string true "Export format" Enums(json, csv)
// @Success 200 {file} file
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/documents/stats/export [get]
func (h *DocumentStatsHandler) ExportStats(c *gin.Context) {
	if !h.requireManagement(c) {
		return
	}
	h.unavailable(c)
}

// GetDashboardStats returns simplified dashboard statistics
// @Summary Get dashboard statistics
// @Description Returns simplified statistics suitable for dashboard display
// @Tags document-stats
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/documents/stats/dashboard [get]
func (h *DocumentStatsHandler) GetDashboardStats(c *gin.Context) {
	if !h.requireManagement(c) {
		return
	}
	h.unavailable(c)
}
