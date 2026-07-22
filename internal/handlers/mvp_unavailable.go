package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
)

// MvpModuleUnavailable is used at the API boundary for modules that the
// current MVP intentionally does not expose. The guard prevents direct API
// callers from reaching legacy handlers that are not yet behind the matter
// isolation wall.
func MvpModuleUnavailable(moduleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		common.NewAPIError(c, http.StatusServiceUnavailable, "MVP_MODULE_UNAVAILABLE", moduleName+"未纳入当前 MVP 试用范围，相关 API 暂不可用")
		c.Abort()
	}
}
