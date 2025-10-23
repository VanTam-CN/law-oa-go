package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

// minioStorageService MinIO存储服务实现
type minioStorageService struct {
	client *minio.Client
	logger *logrus.Logger
	bucket string
}

// NewMinioStorageService 创建新的MinIO存储服务
func NewMinioStorageService(
	endpoint string,
	accessKey string,
	secretKey string,
	bucket string,
	secure bool,
	logger *logrus.Logger,
) (StorageService, error) {
	// 创建MinIO客户端
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// 检查bucket是否存在，如果不存在则创建
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &minioStorageService{
		client: client,
		logger: logger,
		bucket: bucket,
	}, nil
}

// UploadFile 上传文件
func (s *minioStorageService) UploadFile(ctx context.Context, req *UploadFileRequest) (*FileResponse, error) {
	// 生成文件ID
	fileID := uuid.New().String()

	// 设置对象名称
	objectName := s.generateObjectName(req.TenantID, fileID, req.FileName)

	// 上传文件
	info, err := s.client.PutObject(ctx, s.bucket, objectName, req.Data, req.ContentType, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 计算文件哈希
	fileHash := s.calculateFileHash(req.Data)

	response := &FileResponse{
		ID:          fileID,
		FileName:    req.FileName,
		ContentType: req.ContentType,
		Size:        req.Size,
		TenantID:    req.TenantID,
		StoragePath: objectName,
		FileHash:    fileHash,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 添加元数据
	if req.Metadata != nil {
		response.Metadata = req.Metadata
	}

	return response, nil
}

// DownloadFile 下载文件
func (s *minioStorageService) DownloadFile(ctx context.Context, fileID string) ([]byte, *FileMetadata, error) {
	// 从fileID推断对象名称
	objectName, err := s.parseObjectName(fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid file ID: %w", err)
	}

	// 获取对象信息
	objInfo, err := s.client.StatObject(ctx, s.bucket, objectName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object info: %w", err)
	}

	// 下载对象
	object, err := s.client.GetObject(ctx, s.bucket, objectName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// 读取数据
	data := make([]byte, objInfo.Size)
	_, err = object.Read(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read object data: %w", err)
	}

	metadata := &FileMetadata{
		FileName:    objectName,
		ContentType: objInfo.ContentType,
		Size:        objInfo.Size,
		FileHash:    objInfo.ETag(),
		LastModified: objInfo.LastModified,
	}

	return data, metadata, nil
}

// DeleteFile 删除文件
func (s *minioStorageService) DeleteFile(ctx context.Context, fileID string) error {
	objectName, err := s.parseObjectName(fileID)
	if err != nil {
		return fmt.Errorf("invalid file ID: %w", err)
	}

	err = s.client.RemoveObject(ctx, s.bucket, objectName)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// CopyFile 复制文件
func (s *minioStorageService) CopyFile(ctx context.Context, sourceID, targetID string) error {
	sourceObjectName, err := s.parseObjectName(sourceID)
	if err != nil {
		return fmt.Errorf("invalid source file ID: %w", err)
	}

	targetObjectName := s.generateObjectName("", targetID, sourceObjectName)

	// 获取源对象信息
	srcOpts := minio.CopySrcOptions{
		Bucket: s.bucket,
		Object: sourceObjectName,
	}

	// 复制对象
	err = s.client.CopyObject(ctx, s.bucket, targetObjectName, srcOpts)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// GetFileInfo 获取文件信息
func (s *minioStorageService) GetFileInfo(ctx context.Context, fileID string) (*FileResponse, error) {
	objectName, err := s.parseObjectName(fileID)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %w", err)
	}

	objInfo, err := s.client.StatObject(ctx, s.bucket, objectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	// 从对象名称解析文件信息
	fileName, tenantID := s.parseObjectPath(objectName)
	fileHash := objInfo.ETag()

	response := &FileResponse{
		ID:          fileID,
		FileName:    fileName,
		ContentType: objInfo.ContentType,
		Size:        objInfo.Size,
		TenantID:    tenantID,
		StoragePath: objectName,
		FileHash:    fileHash,
		CreatedAt:   objInfo.LastModified,
		UpdatedAt:   objInfo.LastModified,
	}

	return response, nil
}

// ListFiles 列出文件
func (s *minioStorageService) ListFiles(ctx context.Context, filter *FileFilter) (*FileListResponse, error) {
	// 构建前缀
	prefix := ""
	if filter.TenantID != "" {
		prefix = filter.TenantID + "/"
	}

	// 列出对象
	for object := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix: prefix,
		Recursive: false,
	}) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		// 过滤和处理结果
		// TODO: 实现过滤逻辑

	}

	// 这里简化处理，实际应该实现完整的列表逻辑
	return &FileListResponse{
		Files:    []*FileResponse{},
		Total:   0,
		Page:    filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetFileURL 获取文件URL
func (s *minioStorageService) GetFileURL(ctx context.Context, fileID string, expiry time.Duration) (string, error) {
	objectName, err := s.parseObjectName(fileID)
	if err != nil {
		return "", fmt.Errorf("invalid file ID: %w", err)
	}

	// 生成预签名URL
	url, err := s.client.PresignedGetObject(ctx, s.bucket, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// GetStorageStats 获取存储统计
func (s *minioStorageService) GetStorageStats(ctx context.Context, tenantID string) (*StorageStats, error) {
	stats := &StorageStats{
		TotalFiles:   0,
		TotalSize:    0,
		FilesByType:  make(map[string]int64),
		SizeByType:   make(map[string]int64),
		DailyGrowth:  make(map[string]int64),
	}

	prefix := ""
	if tenantID != "" {
		prefix = tenantID + "/"
	}

	// 遍历对象收集统计信息
	for object := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if object.Err != nil {
			s.logger.WithError(object.Err).Error("Failed to list objects for stats")
			continue
		}

		stats.TotalFiles++
		stats.TotalSize += object.Size

		// 按类型统计
		contentType := object.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		stats.FilesByType[contentType]++
		stats.SizeByType[contentType] += object.Size

		// 按日期统计
		dateKey := object.LastModified.Format("2006-01-02")
		stats.DailyGrowth[dateKey]++
	}

	return stats, nil
}

// GetStorageUsage 获取存储使用情况
func (s *minioStorageService) GetStorageUsage(ctx context.Context, filter *UsageFilter) (*StorageUsage, error) {
	usage := &StorageUsage{
		TenantID:    filter.TenantID,
		FileCount:   0,
		TotalSize:    0,
		UsageByType:  make(map[string]int64),
		UsageByDate:  make(map[string]int64),
	}

	prefix := ""
	if filter.TenantID != "" {
		prefix = filter.TenantID + "/"
	}

	// 遍历对象收集使用情况
	for object := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if object.Err != nil {
			s.logger.WithError(object.Err).Error("Failed to list objects for usage")
			continue
		}

		usage.FileCount++
		usage.TotalSize += object.Size

		// 按类型统计
		contentType := object.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		usage.UsageByType[contentType]++

		// 按日期统计
		dateKey := object.LastModified.Format("2006-01-02")
		usage.UsageByDate[dateKey]++
	}

	return usage, nil
}

// 辅助方法

// generateObjectName 生成对象名称
func (s *minioStorageService) generateObjectName(tenantID, fileID, fileName string) string {
	if tenantID != "" {
		return fmt.Sprintf("%s/%s/%s", tenantID, fileID, fileName)
	}
	return fmt.Sprintf("%s/%s", fileID, fileName)
}

// parseObjectName 从对象名称解析文件信息
func (s *minioStorageService) parseObjectName(fileID string) (string, error) {
	// 简化实现，实际应该解析完整的对象路径
	return fileID, nil
}

// parseObjectPath 从对象名称解析文件路径
func (s *minioStorageService) parseObjectPath(objectName string) (string, string) {
	parts := strings.Split(objectName, "/")
	if len(parts) >= 2 {
		return parts[1], parts[0] // 返回文件名和租户ID
	}
	return parts[0], ""
}

// calculateFileHash 计算文件哈希
func (s *minioStorageService) calculateFileHash(data []byte) string {
	// 这里简化处理，实际应该使用SHA256等算法
	return fmt.Sprintf("%x", len(data))
}

// memoryStorageService 内存存储服务实现（用于测试）
type memoryStorageService struct {
	files  map[string]*memoryFile
	logger *logrus.Logger
}

type memoryFile struct {
	Data        []byte
	ContentType string
	Size        int64
	Metadata    map[string]interface{}
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewMemoryStorageService 创建新的内存存储服务
func NewMemoryStorageService(logger *logrus.Logger) StorageService {
	return &memoryStorageService{
		files:  make(map[string]*memoryFile),
		logger: logger,
	}
}

// UploadFile 上传文件到内存
func (s *memoryStorageService) UploadFile(ctx context.Context, req *UploadFileRequest) (*FileResponse, error) {
	fileID := uuid.New().String()

	file := &memoryFile{
		Data:        req.Data,
		ContentType: req.ContentType,
		Size:        req.Size,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.files[fileID] = file

	response := &FileResponse{
		ID:          fileID,
		FileName:    req.FileName,
		ContentType: req.ContentType,
		Size:        req.Size,
		TenantID:    req.TenantID,
		StoragePath: fileID,
		FileHash:    s.calculateFileHash(req.Data),
		CreatedAt:   file.CreatedAt,
		UpdatedAt:   file.UpdatedAt,
	}

	if req.Metadata != nil {
		response.Metadata = req.Metadata
	}

	return response, nil
}

// DownloadFile 从内存下载文件
func (s *memoryStorageService) DownloadFile(ctx context.Context, fileID string) ([]byte, *FileMetadata, error) {
	file, exists := s.files[fileID]
	if !exists {
		return nil, nil, fmt.Errorf("file not found: %s", fileID)
	}

	metadata := &FileMetadata{
		FileName:    fileID,
		ContentType: file.ContentType,
		Size:        file.Size,
		FileHash:    s.calculateFileHash(file.Data),
		LastModified: file.UpdatedAt,
	}

	return file.Data, metadata, nil
}

// DeleteFile 从内存删除文件
func (s *memoryStorageService) DeleteFile(ctx context.Context, fileID string) error {
	delete(s.files, fileID)
	return nil
}

// CopyFile 在内存中复制文件
func (s *memoryService) CopyFile(ctx context.Context, sourceID, targetID string) error {
	source, exists := s.files[sourceID]
	if !exists {
		return fmt.Errorf("source file not found: %s", sourceID)
	}

	target := &memoryFile{
		Data:        make([]byte, len(source.Data)),
		ContentType: source.ContentType,
		Size:        source.Size,
		Metadata:    source.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	copy(source.Data, target.Data)

	s.files[targetID] = target
	return nil
}

// GetFileInfo 获取内存文件信息
func (s *memoryStorageService) GetFileInfo(ctx context.Context, fileID string) (*FileResponse, error) {
	file, exists := s.files[fileID]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", fileID)
	}

	response := &FileResponse{
		ID:          fileID,
		FileName:    fileID,
		ContentType: file.ContentType,
		Size:        file.Size,
		TenantID:    "", // 内存存储不区分租户
		StoragePath: fileID,
		FileHash:    s.calculateFileHash(file.Data),
		CreatedAt:   file.CreatedAt,
		UpdatedAt:   file.UpdatedAt,
	}

	if file.Metadata != nil {
		response.Metadata = file.Metadata
	}

	return response, nil
}

// ListFiles 列出内存文件
func (s *memoryStorageService) ListFiles(ctx context.Context, filter *FileFilter) (*FileListResponse, error) {
	files := make([]*FileResponse, 0, len(s.files))

	for id, file := range s.files {
		// TODO: 实现过滤逻辑
		fileResponse := &FileResponse{
			ID:          id,
			FileName:    id,
			ContentType: file.ContentType,
			Size:        file.Size,
			StoragePath: id,
			FileHash:    s.calculateFileHash(file.Data),
			CreatedAt:   file.CreatedAt,
			UpdatedAt:   file.UpdatedAt,
		}
		if file.Metadata != nil {
			fileResponse.Metadata = file.Metadata
		}
		files = append(files, fileResponse)
	}

	return &FileListResponse{
		Files:    files,
		Total:   int64(len(files)),
		Page:    filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// GetFileURL 内存存储不支持URL生成
func (s *memoryStorageService) GetFileURL(ctx context.Context, fileID string, expiry time.Duration) (string, error) {
	return "", fmt.Errorf("memory storage does not support URL generation")
}

// GetStorageStats 获取内存存储统计
func (s *memoryStorageService) GetStorageStats(ctx context.Context, tenantID string) (*StorageStats, error) {
	stats := &StorageStats{
		TotalFiles:   int64(len(s.files)),
		TotalSize:    0,
		FilesByType:  make(map[string]int64),
		SizeByType:   make(map[string]int64),
		DailyGrowth:  make(map[string]int64),
	}

	for _, file := range s.files {
		stats.TotalSize += file.Size
		contentType := file.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		stats.FilesByType[contentType]++
		stats.SizeByType[contentType] += file.Size

		dateKey := file.CreatedAt.Format("2006-01-02")
		stats.DailyGrowth[dateKey]++
	}

	return stats, nil
}

// GetStorageUsage 获取内存存储使用情况
func (s *memoryStorageService) GetStorageUsage(ctx context.Context, filter *UsageFilter) (*StorageUsage, error) {
	usage := &StorageUsage{
		TenantID:    filter.TenantID,
		FileCount:   int64(len(s.files)),
		TotalSize:    0,
		UsageByType:  make(map[string]int64),
		UsageByDate:  make(map[string]int64),
	}

	for _, file := range s.files {
		usage.TotalSize += file.Size
		contentType := file.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		usage.UsageByType[contentType]++

		dateKey := file.CreatedAt.Format("2006-01-02")
		usage.UsageByDate[dateKey]++
	}

	return usage, nil
}