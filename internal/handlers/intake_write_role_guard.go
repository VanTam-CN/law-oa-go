package handlers

import (
	"net/http"

	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
)

// denyIntakeWriteForPlainUser rejects self-registered plain users before any
// payload parsing or database access. Lawyer, assistant, matter-management,
// and independent conflict-review roles keep their existing paths.
func denyIntakeWriteForPlainUser(c *gin.Context) bool {
	role, _ := middleware.GetCurrentRole(c)
	switch {
	case services.IsIntakeAssistantRole(role):
		return false
	case role == "lawyer":
		return false
	case services.IsBusinessMatterManagementRole(role):
		return false
	case services.IsConflictReviewRole(role):
		return false
	default:
		common.NewAPIError(c, http.StatusForbidden, "INTAKE_WRITE_ROLE_FORBIDDEN",
			"普通账号不能创建或修改接案记录，请由负责律师办理接案")
		return true
	}
}
