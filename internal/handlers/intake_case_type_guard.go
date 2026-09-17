package handlers

import (
	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"net/http"
	"strings"
)

// Validate the already-bound payload without consuming the request body twice.
// Omitted/empty types remain supported for incomplete drafts.
func validateIntakeCaseType(c *gin.Context, payload map[string]interface{}) bool {
	value, present := payload["case_type"]
	if !present {
		return true
	}
	caseType, ok := value.(string)
	if ok && (strings.TrimSpace(caseType) == "" || models.IsValidConflictCaseType(caseType)) {
		return true
	}
	common.NewAPIError(c, http.StatusBadRequest, "VALIDATION_005", "案件类型不在规范列表中，请选择有效案件类型")
	return false
}
