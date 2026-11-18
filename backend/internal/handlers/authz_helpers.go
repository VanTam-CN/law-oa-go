package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"
)

func currentAuthActor(c *gin.Context) (services.AuthActor, bool) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok || userID == 0 {
		common.APIUnauthorized(c, "未授权访问", "用户信息无效")
		return services.AuthActor{}, false
	}
	role, _ := middleware.GetCurrentRole(c)
	return services.AuthActor{UserID: userID, Role: role}, true
}

func currentUserIDString(c *gin.Context) (string, bool) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return "", false
	}
	return strconv.FormatUint(uint64(actor.UserID), 10), true
}

func canViewAllMatterData(c *gin.Context) bool {
	role, _ := middleware.GetCurrentRole(c)
	// Technical administrators manage configuration and accounts; that role is
	// deliberately excluded from business-matter aggregation and conflict data.
	return services.IsBusinessMatterManagementRole(role)
}

func forbidObjectAccess(c *gin.Context) {
	common.APIForbidden(c, "无权访问该资源", "当前账号无权访问或操作该资源")
}
