package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// CRUDService CRUD服务接口
type CRUDService[T any, ID any] interface {
	Create(ctx context.Context, entity *T) (*T, error)
	GetByID(ctx context.Context, id ID) (*T, error)
	Update(ctx context.Context, id ID, entity *T) (*T, error)
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context, req *ListRequest) ([]*T, int64, error)
	Count(ctx context.Context, filter map[string]interface{}) (int64, error)
}

// CRUDHandler 通用CRUD处理器
type CRUDHandler[T any, ID uint | string] struct {
	BaseHandlerImpl
	service  CRUDService[T, ID]
	resource string
}

// NewCRUDHandler 创建CRUD处理器
func NewCRUDHandler[T any, ID uint | string](service CRUDService[T, ID], resource string) *CRUDHandler[T, ID] {
	return &CRUDHandler[T, ID]{
		service:  service,
		resource: resource,
	}
}

// GetTimeout 获取超时时间
func (h *CRUDHandler[T, ID]) GetTimeout() time.Duration {
	return DefaultTimeout
}

// HandleRequest 处理请求 - 必须被重写
func (h *CRUDHandler[T, ID]) HandleRequest(c *gin.Context) *APIResponse {
	return NewErrorResponse("NOT_IMPLEMENTED", "Method not implemented", "")
}

// ValidateRequest 验证请求 - 必须被重写
func (h *CRUDHandler[T, ID]) ValidateRequest(c *gin.Context) error {
	return nil
}

// Create 创建资源
func (h *CRUDHandler[T, ID]) Create(c *gin.Context) {
	h.HandleRequestWithContext(c, func(ctx context.Context) *APIResponse {
		var req T
		if err := c.ShouldBindJSON(&req); err != nil {
			return NewErrorResponse("VALIDATION_ERROR", "Invalid request format", err.Error())
		}

		// 预处理验证
		if err := h.validateCreateRequest(&req); err != nil {
			return NewErrorResponse("VALIDATION_ERROR", "Validation failed", err.Error())
		}

		// 调用服务层
		result, err := h.service.Create(ctx, &req)
		if err != nil {
			return h.handleServiceError(err, "create")
		}

		return NewSuccessResponse(result)
	}, h.GetTimeout())
}

// Get 获取单个资源
func (h *CRUDHandler[T, ID]) Get(c *gin.Context) {
	h.HandleRequestWithContext(c, func(ctx context.Context) *APIResponse {
		// 先获取uint类型的ID
		uintID, err := h.ValidateID(c, "id")
		if err != nil {
			return NewErrorResponse("VALIDATION_ERROR", "Invalid ID", err.Error())
		}

		// 转换为ID类型
		var id ID
		switch any(id).(type) {
		case uint:
			id = any(uintID).(ID)
		case string:
			id = any(fmt.Sprintf("%d", uintID)).(ID)
		default:
			return NewErrorResponse("VALIDATION_ERROR", "Unsupported ID type", "")
		}

		result, err := h.service.GetByID(ctx, id)
		if err != nil {
			return h.handleNotFoundError(err, id)
		}

		return NewSuccessResponse(result)
	}, h.GetTimeout())
}

// Update 更新资源
func (h *CRUDHandler[T, ID]) Update(c *gin.Context) {
	h.HandleRequestWithContext(c, func(ctx context.Context) *APIResponse {
		// 先获取uint类型的ID
		uintID, err := h.ValidateID(c, "id")
		if err != nil {
			return NewErrorResponse("VALIDATION_ERROR", "Invalid ID", err.Error())
		}

		// 转换为ID类型
		var id ID
		switch any(id).(type) {
		case uint:
			id = any(uintID).(ID)
		case string:
			id = any(fmt.Sprintf("%d", uintID)).(ID)
		default:
			return NewErrorResponse("VALIDATION_ERROR", "Unsupported ID type", "")
		}

		var req T
		if err := c.ShouldBindJSON(&req); err != nil {
			return NewErrorResponse("VALIDATION_ERROR", "Invalid request format", err.Error())
		}

		// 预处理验证
		if err := h.validateUpdateRequest(id, &req); err != nil {
			return NewErrorResponse("VALIDATION_ERROR", "Validation failed", err.Error())
		}

		result, err := h.service.Update(ctx, id, &req)
		if err != nil {
			return h.handleServiceError(err, "update")
		}

		return NewSuccessResponse(result)
	}, h.GetTimeout())
}

// Delete 删除资源
func (h *CRUDHandler[T, ID]) Delete(c *gin.Context) {
	h.HandleRequestWithContext(c, func(ctx context.Context) *APIResponse {
		// 先获取uint类型的ID
		uintID, err := h.ValidateID(c, "id")
		if err != nil {
			return NewErrorResponse("VALIDATION_ERROR", "Invalid ID", err.Error())
		}

		// 转换为ID类型
		var id ID
		switch any(id).(type) {
		case uint:
			id = any(uintID).(ID)
		case string:
			id = any(fmt.Sprintf("%d", uintID)).(ID)
		default:
			return NewErrorResponse("VALIDATION_ERROR", "Unsupported ID type", "")
		}

		if err := h.service.Delete(ctx, id); err != nil {
			return h.handleServiceError(err, "delete")
		}

		return NewSuccessResponse(map[string]string{
			"message": fmt.Sprintf("%s deleted successfully", h.resource),
		})
	}, h.GetTimeout())
}

// List 获取资源列表
func (h *CRUDHandler[T, ID]) List(c *gin.Context) {
	h.HandleRequestWithContext(c, func(ctx context.Context) *APIResponse {
		var req ListRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			return NewErrorResponse("VALIDATION_ERROR", "Invalid query parameters", err.Error())
		}

		// 验证和设置默认值
		ValidateListRequest(&req)

		// 调用服务层
		result, total, err := h.service.List(ctx, &req)
		if err != nil {
			return h.handleServiceError(err, "list")
		}

		// 构建分页信息
		pagination := CalculatePagination(req.Page, req.PageSize, total)

		// 创建分页响应
		pageResp := &PageResponse{
			Success:    true,
			Data:       result,
			Pagination: pagination,
			Meta:       NewResponseMeta(),
		}

		// 转换为 APIResponse
		return PageToAPIResponse(pageResp)
	}, h.GetTimeout())
}

// validateCreateRequest 验证创建请求 - 可被子类重写
func (h *CRUDHandler[T, ID]) validateCreateRequest(req *T) error {
	return nil
}

// validateUpdateRequest 验证更新请求 - 可被子类重写
func (h *CRUDHandler[T, ID]) validateUpdateRequest(id ID, req *T) error {
	return nil
}

// handleServiceError 处理服务层错误
func (h *CRUDHandler[T, ID]) handleServiceError(err error, operation string) *APIResponse {
	// 这里可以根据错误类型返回不同的响应
	// 现在简单处理
	return NewErrorResponse("INTERNAL_ERROR",
		fmt.Sprintf("Failed to %s %s", operation, h.resource),
		err.Error())
}

// handleNotFoundError 处理资源未找到错误
func (h *CRUDHandler[T, ID]) handleNotFoundError(err error, id ID) *APIResponse {
	if errors.Is(err, ErrNotFound) {
		return NewErrorResponse("NOT_FOUND",
			fmt.Sprintf("%s not found", h.resource),
			fmt.Sprintf("No %s found with ID: %v", h.resource, id))
	}
	return h.handleServiceError(err, "get")
}

// 自定义错误类型
var ErrNotFound = errors.New("resource not found")
