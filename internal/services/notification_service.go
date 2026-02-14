package services

import (
	"context"
	"log"
)

// NotificationService 通知服务接口（简化版）
// 实际项目中应该连接到真实的通知系统
type notificationService struct {
	// 可以注入真实的通知服务依赖
}

// NewNotificationService 创建新的通知服务
func NewNotificationService() NotificationService {
	return &notificationService{}
}

// SendConflictAlert 发送冲突告警
func (s *notificationService) SendConflictAlert(ctx context.Context, conflicts []*NewConflictInfo) error {
	log.Printf("🚨 发送冲突告警: 发现 %d 个新冲突", len(conflicts))

	for _, conflict := range conflicts {
		log.Printf("  - 冲突详情: %+v", conflict.Details.ToMap())
	}

	// 这里应该调用实际的通知系统
	// 例如：发送邮件、短信、企业微信等

	return nil
}

// SendScanReport 发送扫描报告
func (s *notificationService) SendScanReport(ctx context.Context, jobID uint, result interface{}) error {
	log.Printf("📊 发送扫描报告: jobID=%d", jobID)
	return nil
}

// SendConflictWarning 发送冲突警告通知
func (s *notificationService) SendConflictWarning(ctx context.Context, lawyerID uint, conflictCount int) error {
	log.Printf("⚠️ 发送冲突警告: lawyerID=%d, count=%d", lawyerID, conflictCount)
	return nil
}
