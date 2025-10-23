package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// auditService 审计服务实现
type auditService struct {
	auditRepo repositories.DocumentAuditRepository
	userRepo  repositories.UserRepository
	docRepo   repositories.DocumentRepository
	logger    *logrus.Logger
}

// NewAuditService 创建新的审计服务
func NewAuditService(
	auditRepo repositories.DocumentAuditRepository,
	userRepo repositories.UserRepository,
	docRepo repositories.DocumentRepository,
	logger *logrus.Logger,
) AuditService {
	return &auditService{
		auditRepo: auditRepo,
		userRepo:  userRepo,
		docRepo:   docRepo,
		logger:    logger,
	}
}

// LogAction 记录操作日志
func (s *auditService) LogAction(ctx context.Context, req *LogActionRequest) error {
	// 解析用户ID
	var userID uint
	if req.UserID != "" {
		id, err := s.parseUserID(req.UserID)
		if err != nil {
			s.logger.WithError(err).Error("Invalid user ID in audit request")
			return fmt.Errorf("invalid user ID: %w", err)
		}
		userID = id
	}

	// 解析文档ID
	var documentID uint
	if req.DocumentID != "" {
		id, err := s.parseDocumentID(req.DocumentID)
		if err != nil {
			s.logger.WithError(err).Error("Invalid document ID in audit request")
			return fmt.Errorf("invalid document ID: %w", err)
		}
		documentID = id
	}

	// 创建审计记录
	audit := &models.DocumentAudit{
		UserID:     userID,
		DocumentID: documentID,
		TenantID:   req.TenantID,
		Action:     req.Action,
		Details:    req.Details,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.auditRepo.Create(ctx, audit); err != nil {
		return fmt.Errorf("failed to create audit record: %w", err)
	}

	return nil
}

// GetAuditLogs 获取审计日志
func (s *auditService) GetAuditLogs(ctx context.Context, filter *AuditFilter) (*AuditListResponse, error) {
	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// 计算偏移量
	offset := (filter.Page - 1) * filter.PageSize

	// 构建查询选项
	options := repositories.AuditListOptions{
		TenantID:   filter.TenantID,
		UserID:     0,
		DocumentID: 0,
		Action:     filter.Action,
		StartDate:  filter.StartDate,
		EndDate:    filter.EndDate,
		Limit:      filter.PageSize,
		Offset:     offset,
		SortBy:     filter.SortBy,
		SortOrder:  filter.SortOrder,
	}

	// 解析ID过滤条件
	if filter.UserID != "" {
		userID, err := s.parseUserID(filter.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID filter: %w", err)
		}
		options.UserID = userID
	}

	if filter.DocumentID != "" {
		documentID, err := s.parseDocumentID(filter.DocumentID)
		if err != nil {
			return nil, fmt.Errorf("invalid document ID filter: %w", err)
		}
		options.DocumentID = documentID
	}

	// 获取审计日志
	audits, total, err := s.auditRepo.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	// 转换为响应格式
	responses := make([]*AuditResponse, len(audits))
	for i, audit := range audits {
		responses[i] = s.convertToAuditResponse(audit)
	}

	return &AuditListResponse{
		Audits:   responses,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetAuditLog 获取单个审计日志
func (s *auditService) GetAuditLog(ctx context.Context, auditID string) (*AuditResponse, error) {
	id, err := s.parseAuditID(auditID)
	if err != nil {
		return nil, fmt.Errorf("invalid audit ID: %w", err)
	}

	audit, err := s.auditRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return s.convertToAuditResponse(audit), nil
}

// GetUserAuditLogs 获取用户审计日志
func (s *auditService) GetUserAuditLogs(ctx context.Context, userID string, filter *AuditFilter) (*AuditListResponse, error) {
	// 解析用户ID
	userIDInt, err := s.parseUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// 计算偏移量
	offset := (filter.Page - 1) * filter.PageSize

	// 构建查询选项
	options := repositories.AuditListOptions{
		TenantID:   filter.TenantID,
		UserID:     userIDInt,
		DocumentID: 0,
		Action:     filter.Action,
		StartDate:  filter.StartDate,
		EndDate:    filter.EndDate,
		Limit:      filter.PageSize,
		Offset:     offset,
		SortBy:     filter.SortBy,
		SortOrder:  filter.SortOrder,
	}

	// 解析文档ID过滤条件
	if filter.DocumentID != "" {
		documentID, err := s.parseDocumentID(filter.DocumentID)
		if err != nil {
			return nil, fmt.Errorf("invalid document ID filter: %w", err)
		}
		options.DocumentID = documentID
	}

	// 获取审计日志
	audits, total, err := s.auditRepo.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get user audit logs: %w", err)
	}

	// 转换为响应格式
	responses := make([]*AuditResponse, len(audits))
	for i, audit := range audits {
		responses[i] = s.convertToAuditResponse(audit)
	}

	return &AuditListResponse{
		Audits:   responses,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetDocumentAuditLogs 获取文档审计日志
func (s *auditService) GetDocumentAuditLogs(ctx context.Context, documentID string, filter *AuditFilter) (*AuditListResponse, error) {
	// 解析文档ID
	docID, err := s.parseDocumentID(documentID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// 计算偏移量
	offset := (filter.Page - 1) * filter.PageSize

	// 构建查询选项
	options := repositories.AuditListOptions{
		TenantID:   filter.TenantID,
		UserID:     0,
		DocumentID: docID,
		Action:     filter.Action,
		StartDate:  filter.StartDate,
		EndDate:    filter.EndDate,
		Limit:      filter.PageSize,
		Offset:     offset,
		SortBy:     filter.SortBy,
		SortOrder:  filter.SortOrder,
	}

	// 解析用户ID过滤条件
	if filter.UserID != "" {
		userID, err := s.parseUserID(filter.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID filter: %w", err)
		}
		options.UserID = userID
	}

	// 获取审计日志
	audits, total, err := s.auditRepo.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get document audit logs: %w", err)
	}

	// 转换为响应格式
	responses := make([]*AuditResponse, len(audits))
	for i, audit := range audits {
		responses[i] = s.convertToAuditResponse(audit)
	}

	return &AuditListResponse{
		Audits:   responses,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetAuditStats 获取审计统计
func (s *auditService) GetAuditStats(ctx context.Context, filter *AuditStatsFilter) (*AuditStats, error) {
	// 构建查询选项
	options := repositories.AuditListOptions{
		TenantID:  filter.TenantID,
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Limit:     10000, // 大量数据用于统计
		Offset:    0,
	}

	// 解析用户ID过滤条件
	if filter.UserID != "" {
		userID, err := s.parseUserID(filter.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID filter: %w", err)
		}
		options.UserID = userID
	}

	// 解析文档ID过滤条件
	if filter.DocumentID != "" {
		documentID, err := s.parseDocumentID(filter.DocumentID)
		if err != nil {
			return nil, fmt.Errorf("invalid document ID filter: %w", err)
		}
		options.DocumentID = documentID
	}

	// 获取审计日志用于统计
	audits, _, err := s.auditRepo.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs for stats: %w", err)
	}

	stats := &AuditStats{
		TotalActions:      int64(len(audits)),
		ActionsByType:     make(map[string]int64),
		ActionsByUser:     make(map[string]int64),
		ActionsByDate:     make(map[string]int64),
		TopDocuments:      make(map[string]int64),
		TopUsers:          make(map[string]int64),
	}

	// 统计数据
	for _, audit := range audits {
		// 按操作类型统计
		stats.ActionsByType[audit.Action]++

		// 按用户统计
		if audit.UserID != nil {
			userKey := strconv.FormatUint(uint64(*audit.UserID), 10)
			stats.ActionsByUser[userKey]++
			stats.TopUsers[userKey]++
		}

		// 按日期统计
		dateKey := audit.CreatedAt.Format("2006-01-02")
		stats.ActionsByDate[dateKey]++

		// 按文档统计
		if audit.DocumentID != nil {
			docKey := strconv.FormatUint(uint64(*audit.DocumentID), 10)
			stats.TopDocuments[docKey]++
		}
	}

	// 获取用户名和文档名（如果需要）
	if len(stats.TopUsers) > 0 {
		for userID := range stats.TopUsers {
			if user, err := s.userRepo.GetByID(ctx, uint(strconv.ParseUint(userID, 10, 32))); err == nil {
				stats.UserNames[userID] = user.Username
			}
		}
	}

	if len(stats.TopDocuments) > 0 {
		for docID := range stats.TopDocuments {
			if doc, err := s.docRepo.GetByID(ctx, uint(strconv.ParseUint(docID, 10, 32))); err == nil {
				stats.DocumentNames[docID] = doc.Name
			}
		}
	}

	return stats, nil
}

// DeleteAuditLogs 删除审计日志
func (s *auditService) DeleteAuditLogs(ctx context.Context, req *DeleteAuditLogsRequest) error {
	// 验证删除条件
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required for deletion")
	}

	// 检查日期范围
	if req.EndDate == nil {
		return fmt.Errorf("end_date is required for deletion")
	}

	// 执行删除
	deletedCount, err := s.auditRepo.DeleteBeforeDate(ctx, req.TenantID, *req.EndDate)
	if err != nil {
		return fmt.Errorf("failed to delete audit logs: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"tenant_id":    req.TenantID,
		"end_date":     req.EndDate,
		"deleted_count": deletedCount,
	}).Info("Deleted audit logs")

	// 记录删除操作的审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "delete_audit_logs",
		Details:    fmt.Sprintf("Deleted %d audit logs before %s", deletedCount, req.EndDate.Format("2006-01-02")),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.LogAction(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit deletion")
	}

	return nil
}

// ExportAuditLogs 导出审计日志
func (s *auditService) ExportAuditLogs(ctx context.Context, filter *AuditFilter) (*AuditExportResponse, error) {
	// 设置大页面大小用于导出
	originalPageSize := filter.PageSize
	filter.PageSize = 10000

	// 获取所有符合条件的审计日志
	response, err := s.GetAuditLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs for export: %w", err)
	}

	// 恢复原始页面大小
	filter.PageSize = originalPageSize

	// 生成导出数据
	exportData := make([]map[string]interface{}, len(response.Audits))
	for i, audit := range response.Audits {
		exportData[i] = map[string]interface{}{
			"id":          audit.ID,
			"user_id":     audit.UserID,
			"username":    audit.Username,
			"document_id": audit.DocumentID,
			"document_name": audit.DocumentName,
			"action":      audit.Action,
			"details":     audit.Details,
			"ip_address":  audit.IPAddress,
			"user_agent":  audit.UserAgent,
			"created_at":  audit.CreatedAt,
		}
	}

	return &AuditExportResponse{
		Data:      exportData,
		Total:     response.Total,
		ExportedAt: time.Now(),
		Format:    "json",
	}, nil
}

// SearchAuditLogs 搜索审计日志
func (s *auditService) SearchAuditLogs(ctx context.Context, req *SearchAuditLogsRequest) (*AuditListResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize

	// 构建查询选项
	options := repositories.AuditListOptions{
		TenantID:   req.TenantID,
		UserID:     0,
		DocumentID: 0,
		Action:     "",
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Limit:      req.PageSize,
		Offset:     offset,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
		Search:     req.Query,
	}

	// 解析ID过滤条件
	if req.UserID != "" {
		userID, err := s.parseUserID(req.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID filter: %w", err)
		}
		options.UserID = userID
	}

	if req.DocumentID != "" {
		documentID, err := s.parseDocumentID(req.DocumentID)
		if err != nil {
			return nil, fmt.Errorf("invalid document ID filter: %w", err)
		}
		options.DocumentID = documentID
	}

	// 获取审计日志
	audits, total, err := s.auditRepo.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to search audit logs: %w", err)
	}

	// 转换为响应格式
	responses := make([]*AuditResponse, len(audits))
	for i, audit := range audits {
		responses[i] = s.convertToAuditResponse(audit)
	}

	return &AuditListResponse{
		Audits:   responses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 辅助方法

// parseUserID 解析用户ID
func (s *auditService) parseUserID(userID string) (uint, error) {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %s", userID)
	}
	return uint(id), nil
}

// parseDocumentID 解析文档ID
func (s *auditService) parseDocumentID(documentID string) (uint, error) {
	id, err := strconv.ParseUint(documentID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid document ID format: %s", documentID)
	}
	return uint(id), nil
}

// parseAuditID 解析审计ID
func (s *auditService) parseAuditID(auditID string) (uint, error) {
	id, err := strconv.ParseUint(auditID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid audit ID format: %s", auditID)
	}
	return uint(id), nil
}

// convertToAuditResponse 转换为审计响应格式
func (s *auditService) convertToAuditResponse(audit *models.DocumentAudit) *AuditResponse {
	response := &AuditResponse{
		ID:         audit.ID,
		UserID:     "",
		DocumentID: "",
		TenantID:   audit.TenantID,
		Action:     audit.Action,
		Details:    audit.Details,
		IPAddress:  audit.IPAddress,
		UserAgent:  audit.UserAgent,
		CreatedAt:  audit.CreatedAt,
	}

	// 转换用户ID为字符串
	if audit.UserID != nil {
		response.UserID = strconv.FormatUint(uint64(*audit.UserID), 10)
		// 尝试获取用户名
		if user, err := s.userRepo.GetByID(context.Background(), *audit.UserID); err == nil {
			response.Username = user.Username
		}
	}

	// 转换文档ID为字符串
	if audit.DocumentID != nil {
		response.DocumentID = strconv.FormatUint(uint64(*audit.DocumentID), 10)
		// 尝试获取文档名
		if doc, err := s.docRepo.GetByID(context.Background(), *audit.DocumentID); err == nil {
			response.DocumentName = doc.Name
		}
	}

	return response
}