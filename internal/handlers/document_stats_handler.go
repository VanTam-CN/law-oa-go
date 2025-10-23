package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
	ctx := c.Request.Context()

	stats, err := h.statsService.GetDocumentOverview(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get document overview",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    stats,
		"error":   nil,
		"success": true,
	})
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
	ctx := c.Request.Context()

	stats, err := h.statsService.GetStorageUsage(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get storage usage",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    stats,
		"error":   nil,
		"success": true,
	})
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
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "User ID must be a valid integer",
		})
		return
	}

	ctx := c.Request.Context()

	stats, err := h.statsService.GetUserActivity(ctx, uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user activity",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    stats,
		"error":   nil,
		"success": true,
	})
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
	ctx := c.Request.Context()

	report, err := h.statsService.GetComplianceReport(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get compliance report",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    report,
		"error":   nil,
		"success": true,
	})
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
	statsType := c.Query("type")
	format := c.Query("format")

	if statsType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing type parameter",
			"message": "Type parameter is required (overview, storage, compliance)",
		})
		return
	}

	if format == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing format parameter",
			"message": "Format parameter is required (json, csv)",
		})
		return
	}

	// Validate parameters
	validTypes := map[string]bool{
		"overview":   true,
		"storage":    true,
		"compliance": true,
	}
	validFormats := map[string]bool{
		"json": true,
		"csv":  true,
	}

	if !validTypes[statsType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid type parameter",
			"message": "Type must be one of: overview, storage, compliance",
		})
		return
	}

	if !validFormats[format] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid format parameter",
			"message": "Format must be one of: json, csv",
		})
		return
	}

	ctx := c.Request.Context()

	data, err := h.statsService.ExportStats(ctx, statsType, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to export statistics",
			"message": err.Error(),
		})
		return
	}

	// Set appropriate headers for file download
	filename := "document_stats_" + statsType + "." + format
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", data)
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
	ctx := c.Request.Context()

	// Get overview stats
	overview, err := h.statsService.GetDocumentOverview(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get overview stats",
			"message": err.Error(),
		})
		return
	}

	// Get storage stats
	storage, err := h.statsService.GetStorageUsage(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get storage stats",
			"message": err.Error(),
		})
		return
	}

	// Create simplified dashboard data
	dashboardStats := gin.H{
		"total_documents": overview.TotalDocuments,
		"recent_uploads":  overview.RecentUploads,
		"storage_usage": gin.H{
			"total_space":      storage.TotalSpace,
			"used_space":       storage.UsedSpace,
			"available_space":  storage.AvailableSpace,
			"usage_percentage": storage.UsagePercentage,
		},
		"top_categories": func() []gin.H {
			var categories []gin.H
			for _, cat := range overview.DocumentsByCategory {
				categories = append(categories, gin.H{
					"name":  cat.Category,
					"count": cat.Count,
					"size":  cat.Size,
				})
			}
			return categories
		}(),
		"upload_trend": func() []gin.H {
			var trends []services.TrendData
			if len(overview.UploadTrends) > 7 {
				// Return last 7 days
				trends = overview.UploadTrends[len(overview.UploadTrends)-7:]
			} else {
				trends = overview.UploadTrends
			}

			var result []gin.H
			for _, trend := range trends {
				result = append(result, gin.H{
					"date":  trend.Date,
					"count": trend.Count,
				})
			}
			return result
		}(),
		"generated_at": overview.GeneratedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    dashboardStats,
		"error":   nil,
		"success": true,
	})
}