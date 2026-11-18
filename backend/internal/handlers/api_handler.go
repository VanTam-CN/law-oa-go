package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/errors"
)

// APICreateHandler a generic handler for creation operations using new API format.
func APICreateHandler[ReqT any, ResT any](
	createServiceFunc func(c *gin.Context, req *ReqT) (*ResT, error),
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReqT
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(errors.ValidationErrorWithDetails("request_binding", "Invalid request format: "+err.Error(), "Invalid request format", []string{"request_binding"}))
			return
		}

		res, err := createServiceFunc(c, &req)
		if err != nil {
			_ = c.Error(err)
			return
		}

		common.APISuccess(c, res)
	}
}

// APIListHandler a generic handler for listing operations using new API format.
func APIListHandler[ReqT any, ResT any](
	listServiceFunc func(c *gin.Context, req *ReqT) ([]*ResT, int64, error),
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ReqT
		if err := c.ShouldBindQuery(&req); err != nil {
			_ = c.Error(errors.ValidationErrorWithDetails("query_binding", "Invalid query parameters: "+err.Error(), "Invalid query parameters", []string{"query_binding"}))
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
			_ = c.Error(errors.ValidationErrorWithDetails("pagination", "Invalid pagination parameters", "Invalid pagination parameters", []string{"pagination"}))
			return
		}

		common.APISuccessWithPage(c, results, total, page, pageSize)
	}
}

// APIGetHandler a generic handler for fetching a single entity by ID using new API format.
func APIGetHandler[ResT any](
	getServiceFunc func(c *gin.Context, id uint) (*ResT, error),
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			_ = c.Error(errors.ValidationErrorWithDetails("id_validation", "Invalid ID: must be a valid number", "Invalid ID: must be a valid number", []string{"id_validation"}))
			return
		}

		res, err := getServiceFunc(c, uint(id))
		if err != nil {
			_ = c.Error(err)
			return
		}

		common.APISuccess(c, res)
	}
}

// APIUpdateHandler a generic handler for update operations using new API format.
func APIUpdateHandler[ReqT any, ResT any](
	updateServiceFunc func(c *gin.Context, id uint, req *ReqT) (*ResT, error),
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			_ = c.Error(errors.ValidationErrorWithDetails("id_validation", "Invalid ID: must be a valid number", "Invalid ID: must be a valid number", []string{"id_validation"}))
			return
		}

		var req ReqT
		if err := c.ShouldBindJSON(&req); err != nil {
			_ = c.Error(errors.ValidationErrorWithDetails("request_binding", "Invalid request format: "+err.Error(), "Invalid request format", []string{"request_binding"}))
			return
		}

		res, err := updateServiceFunc(c, uint(id), &req)
		if err != nil {
			_ = c.Error(err)
			return
		}

		common.APISuccess(c, res)
	}
}

// APIDeleteHandler a generic handler for delete operations using new API format.
func APIDeleteHandler(
	deleteServiceFunc func(c *gin.Context, id uint) error,
	errorContext string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			_ = c.Error(errors.ValidationErrorWithDetails("id_validation", "Invalid ID: must be a valid number", "Invalid ID: must be a valid number", []string{"id_validation"}))
			return
		}

		err = deleteServiceFunc(c, uint(id))
		if err != nil {
			_ = c.Error(err)
			return
		}

		common.APISuccess(c, gin.H{"message": errorContext + " deleted successfully"})
	}
}
