package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/errors"
)

// CreateHandler a generic handler for creation operations.
func CreateHandler[ReqT any, ResT any](
	createServiceFunc func(c *gin.Context, req *ReqT) (*ResT, error),
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReqT
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format: "+err.Error(), "Invalid request format"))
			return
		}

		res, err := createServiceFunc(c, &req)
		if err != nil {
			_ = c.Error(err)
			return
		}

		common.Success(c, res)
	}
}

// ListHandler a generic handler for listing operations.
func ListHandler[ReqT any, ResT any](
	listServiceFunc func(c *gin.Context, req *ReqT) ([]*ResT, int64, error),
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReqT
		if err := c.ShouldBindQuery(&req); err != nil {
			_ = c.Error(errors.NewValidationError("query_binding", "query_binding", "Invalid query parameters: "+err.Error(), "Invalid query parameters"))
			return
		}

		results, total, err := listServiceFunc(c, &req)
		if err != nil {
			_ = c.Error(err)
			return
		}

		page, err1 := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, err2 := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if err1 != nil || err2 != nil {
			_ = c.Error(errors.NewValidationError("pagination", "pagination", "Invalid pagination parameters", "Invalid pagination parameters"))
			return
		}

		response := common.PageResponse{
			Data:  results,
			Total: total,
			Page:  page,
			Size:  pageSize,
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetHandler a generic handler for fetching a single entity by ID.
func GetHandler[ResT any](
	getServiceFunc func(c *gin.Context, id uint) (*ResT, error),
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			_ = c.Error(errors.NewValidationError("id_validation", "id_validation", "Invalid ID: must be a valid number", "Invalid ID: must be a valid number"))
			return
		}

		res, err := getServiceFunc(c, uint(id))
		if err != nil {
			_ = c.Error(err)
			return
		}

		common.Success(c, res)
	}
}

// UpdateHandler a generic handler for update operations.
func UpdateHandler[ReqT any, ResT any](
	updateServiceFunc func(c *gin.Context, id uint, req *ReqT) (*ResT, error),
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			_ = c.Error(errors.NewValidationError("id_validation", "id_validation", "Invalid ID: must be a valid number", "Invalid ID: must be a valid number"))
			return
		}

		var req ReqT
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format: "+err.Error(), "Invalid request format"))
			return
		}

		res, err := updateServiceFunc(c, uint(id), &req)
		if err != nil {
			_ = c.Error(err)
			return
		}

		common.Success(c, res)
	}
}

// DeleteHandler a generic handler for delete operations.
func DeleteHandler(
	deleteServiceFunc func(c *gin.Context, id uint) error,
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			_ = c.Error(errors.NewValidationError("id_validation", "id_validation", "Invalid ID: must be a valid number", "Invalid ID: must be a valid number"))
			return
		}

		err = deleteServiceFunc(c, uint(id))
		if err != nil {
			_ = c.Error(err)
			return
		}

		common.Success(c, gin.H{"message": errorContext + " deleted successfully"})
	}
}
