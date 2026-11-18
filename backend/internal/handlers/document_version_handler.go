package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// DocumentVersionHandler 文档版本管理处理器
type DocumentVersionHandler struct {
	db            *gorm.DB
	versionService *services.DocumentVersionService
	lockService   *services.DocumentLockService
}

// NewDocumentVersionHandler 创建文档版本处理器
func NewDocumentVersionHandler(
	db *gorm.DB,
	versionService *services.DocumentVersionService,
	lockService *services.DocumentLockService,
) *DocumentVersionHandler {
	return &DocumentVersionHandler{
		db:             db,
		versionService: versionService,
		lockService:    lockService,
	}
}

// CreateVersion 创建新版本
func (h *DocumentVersionHandler) CreateVersion(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 解析请求
	var req struct {
		ChangeDescription string `json:"change_description"`
		ChangeType        string `json:"change_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// 创建版本请求
	versionReq := &services.CreateVersionRequest{
		DocumentID:  uint(documentID),
		Name:        fmt.Sprintf("Version %s", req.ChangeType),
		Description: req.ChangeDescription,
		Changes:     req.ChangeDescription,
		CreatedBy:   userID.(uint),
	}

	// 从当前文档创建新版本
	version, err := h.versionService.CreateVersion(ctx, versionReq)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, version)
}

// GetVersions 获取文档的所有版本
func (h *DocumentVersionHandler) GetVersions(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	ctx := c.Request.Context()

	req := &services.VersionListRequest{
		DocumentID: uint(documentID),
		Page:       page,
		PageSize:   pageSize,
	}

	versions, total, err := h.versionService.GetVersions(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       versions,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetVersion 获取指定版本
func (h *DocumentVersionHandler) GetVersion(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	versionStr := c.Param("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version number"})
		return
	}

	ctx := c.Request.Context()

	versionInfo, err := h.versionService.GetVersion(ctx, int(documentID), version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, versionInfo)
}

// GetCurrentVersion 获取当前版本
func (h *DocumentVersionHandler) GetCurrentVersion(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	ctx := c.Request.Context()

	// 获取版本历史，然后返回最新的版本
	versions, err := h.versionService.GetVersionHistory(ctx, uint(documentID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 返回最新的版本（第一个）
	if len(versions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No versions found"})
		return
	}

	c.JSON(http.StatusOK, versions[0])
}

// RestoreVersion 恢复到指定版本
func (h *DocumentVersionHandler) RestoreVersion(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	versionStr := c.Param("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version number"})
		return
	}

	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx := c.Request.Context()

	// 验证编辑权限
	canEdit, _, err := h.lockService.ValidateEditPermission(ctx, uint(documentID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !canEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "No permission to edit this document"})
		return
	}

	// 恢复版本 - 服务方法需要 int 类型
	if err := h.versionService.RestoreVersion(ctx, int(documentID), version, userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Version restored successfully"})
}

// CompareVersions 比较两个版本
func (h *DocumentVersionHandler) CompareVersions(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	fromVersionStr := c.Query("from")
	toVersionStr := c.Query("to")

	if fromVersionStr == "" || toVersionStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both 'from' and 'to' versions are required"})
		return
	}

	fromVersion, err := strconv.Atoi(fromVersionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from version"})
		return
	}

	toVersion, err := strconv.Atoi(toVersionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to version"})
		return
	}

	ctx := c.Request.Context()

	// 构建比较请求
	req := &services.CompareVersionsRequest{
		DocumentID:  uint(documentID),
		FromVersion: fromVersion,
		ToVersion:   toVersion,
	}

	comparison, err := h.versionService.CompareVersions(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comparison)
}

// DeleteVersion 删除版本
func (h *DocumentVersionHandler) DeleteVersion(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	versionStr := c.Param("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version number"})
		return
	}

	ctx := c.Request.Context()

	// 删除版本 - 服务方法需要 int 类型
	if err := h.versionService.DeleteVersion(ctx, int(documentID), version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Version deleted successfully"})
}

// GetLockStatus 获取文档锁状态
func (h *DocumentVersionHandler) GetLockStatus(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx := c.Request.Context()

	status, err := h.lockService.GetLockStatus(ctx, uint(documentID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// AcquireLock 获取文档锁
func (h *DocumentVersionHandler) AcquireLock(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 获取用户信息
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 解析请求
	var req struct {
		IsCheckout bool `json:"is_checkout"`
	}

	c.ShouldBindJSON(&req)

	ctx := c.Request.Context()

	lockReq := &services.AcquireLockRequest{
		DocumentID: uint(documentID),
		UserID:     userID.(uint),
		UserName:   user.Name,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		IsCheckout: req.IsCheckout,
	}

	status, err := h.lockService.AcquireLock(ctx, lockReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ReleaseLock 释放文档锁
func (h *DocumentVersionHandler) ReleaseLock(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 解析请求
	var req struct {
		Force bool `json:"force"`
	}

	c.ShouldBindJSON(&req)

	ctx := c.Request.Context()

	releaseReq := &services.ReleaseLockRequest{
		DocumentID: uint(documentID),
		UserID:     userID.(uint),
		Force:      req.Force,
	}

	if err := h.lockService.ReleaseLock(ctx, releaseReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lock released successfully"})
}

// RenewLock 续期文档锁
func (h *DocumentVersionHandler) RenewLock(c *gin.Context) {
	documentIDStr := c.Param("document_id")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx := c.Request.Context()

	renewReq := &services.RenewLockRequest{
		DocumentID: uint(documentID),
		UserID:     userID.(uint),
	}

	status, err := h.lockService.RenewLock(ctx, renewReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetUserLocks 获取用户持有的所有锁
func (h *DocumentVersionHandler) GetUserLocks(c *gin.Context) {
	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx := c.Request.Context()

	locks, err := h.lockService.GetUserLocks(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  locks,
		"count": len(locks),
	})
}
