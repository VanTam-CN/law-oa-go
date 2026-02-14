package handlers

import (
	"fmt"
	"law-oa-go/internal/services"
	"law-oa-go/internal/common"
	"law-oa-go/internal/errors"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// AvatarHandler 头像处理器
type AvatarHandler struct {
	userService *services.UserService
}

// NewAvatarHandler 创建新的头像处理器
func NewAvatarHandler(userService *services.UserService) *AvatarHandler {
	return &AvatarHandler{
		userService: userService,
	}
}

// UploadAvatar godoc
// @Summary 上传用户头像
// @Description 上传当前登录用户的头像
// @Tags 用户
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "头像文件"
// @Success 200 {object} common.APIResponse{data=AvatarUploadResponse} "上传成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /users/avatar [post]
func (h *AvatarHandler) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		_ = c.Error(errors.UnauthorizedError("upload_avatar", "avatar"))
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		_ = c.Error(errors.ValidationError("avatar", "Missing avatar file: "+err.Error()))
		return
	}

	// 验证文件类型
	if !isValidImageType(file.Filename) {
		_ = c.Error(errors.ValidationError("avatar", "Invalid file type: Only JPG, PNG, and GIF images are allowed"))
		return
	}

	// 验证文件大小 (限制为2MB)
	if file.Size > 2*1024*1024 {
		_ = c.Error(errors.ValidationError("avatar", "File too large: Avatar must be less than 2MB"))
		return
	}

	// 生成文件路径
	ext := strings.ToLower(filepath.Ext(file.Filename))
	avatarPath := fmt.Sprintf("/uploads/avatars/%d%s", userID.(uint), ext)

	// 保存文件
	if err := c.SaveUploadedFile(file, "."+avatarPath); err != nil {
		_ = c.Error(errors.InternalError("file_save_failed", err))
		return
	}

	// 更新用户头像 - 直接调用repository方法
	user, err := h.userService.UpdateUserAvatar(c.Request.Context(), userID.(uint), avatarPath)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// 返回响应
	response := AvatarUploadResponse{
		URL:      avatarPath,
		Message:  "Avatar uploaded successfully",
		User:     user,
	}

	common.APISuccess(c, response)
}

// AvatarUploadResponse 头像上传响应
type AvatarUploadResponse struct {
	URL     string              `json:"url"`
	Message string              `json:"message"`
	User    *services.UserProfile `json:"user"`
}

// isValidImageType 验证图片类型
func isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}