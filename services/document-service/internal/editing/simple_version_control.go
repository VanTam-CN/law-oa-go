package editing

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SimpleVersionControlService 简化版本控制服务
type SimpleVersionControlService interface {
	// 基本仓库管理
	InitializeRepository(ctx context.Context, docID string) error
	GetRepository(ctx context.Context, docID string) (string, error)
	DeleteRepository(ctx context.Context, docID string) error

	// 版本操作
	SaveVersion(ctx context.Context, docID string, content []byte, author string, message string) (string, error)
	GetVersion(ctx context.Context, docID string, versionID string) ([]byte, error)
	GetVersions(ctx context.Context, docID string) ([]*VersionInfo, error)
	CompareVersions(ctx context.Context, docID string, fromVersion, toVersion string) (*DiffResult, error)

	// 分支操作
	CreateBranch(ctx context.Context, docID string, branchName string) error
	GetBranches(ctx context.Context, docID string) ([]string, error)
	SwitchBranch(ctx context.Context, docID string, branchName string) error
}

// AuthService 认证服务接口
type AuthService interface {
	AuthenticateUser(ctx context.Context, token string) (string, error)
	HasPermission(ctx context.Context, userID, resource, action string) error
}

// VersionInfo 版本信息
type VersionInfo struct {
	ID        string    `json:"id"`
	DocID     string    `json:"doc_id"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
	Branch    string    `json:"branch"`
}

// DiffResult 差异结果
type DiffResult struct {
	FromVersion string   `json:"from_version"`
	ToVersion   string   `json:"to_version"`
	Changes     []*Change `json:"changes"`
	Summary     *DiffSummary `json:"summary"`
	Timestamp   time.Time `json:"timestamp"`
}

// Change 变更
type Change struct {
	Type        string `json:"type"` // "add", "delete", "modify"
	LineNumber  int    `json:"line_number"`
	OldContent  string `json:"old_content"`
	NewContent  string `json:"new_content"`
}

// DiffSummary 差异摘要
type DiffSummary struct {
	LinesAdded    int `json:"lines_added"`
	LinesRemoved  int `json:"lines_removed"`
	CharactersAdded int `json:"characters_added"`
	CharactersRemoved int `json:"characters_removed"`
}

// SimpleVersionControlImpl 简化版本控制实现
type SimpleVersionControlImpl struct {
	basePath   string
	logger     *logrus.Logger
	authService AuthService
	repos      map[string]*SimpleRepo
	mutex      sync.RWMutex
}

// SimpleRepo 简化仓库
type SimpleRepo struct {
	DocID      string
	Branches   map[string]*SimpleBranch
	CurrentBranch string
	mutex      sync.RWMutex
	CreatedAt  time.Time
}

// SimpleBranch 简化分支
type SimpleBranch struct {
	Name     string
	Versions []*SimpleVersion
	Head     string // 当前版本ID
	mutex    sync.RWMutex
}

// SimpleVersion 简化版本
type SimpleVersion struct {
	ID        string
	Content   []byte
	Author    string
	Message   string
	Timestamp time.Time
	Parent    string
}

// NewSimpleVersionControlService 创建简化版本控制服务
func NewSimpleVersionControlService(
	basePath string,
	logger *logrus.Logger,
	authService AuthService,
) SimpleVersionControlService {
	return &SimpleVersionControlImpl{
		basePath:   basePath,
		logger:     logger,
		authService: authService,
		repos:      make(map[string]*SimpleRepo),
	}
}

// InitializeRepository 初始化仓库
func (s *SimpleVersionControlImpl) InitializeRepository(ctx context.Context, docID string) error {
	if err := s.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.repos[docID]; exists {
		return nil // 仓库已存在
	}

	// 创建新仓库
	repo := &SimpleRepo{
		DocID:        docID,
		Branches:     make(map[string]*SimpleBranch),
		CurrentBranch: "main",
		CreatedAt:    time.Now(),
	}

	// 创建main分支
	repo.Branches["main"] = &SimpleBranch{
		Name:  "main",
		Versions: make([]*SimpleVersion, 0),
	}

	s.repos[docID] = repo

	s.logger.WithField("doc_id", docID).Info("文档仓库已初始化")
	return nil
}

// GetRepository 获取仓库
func (s *SimpleVersionControlImpl) GetRepository(ctx context.Context, docID string) (string, error) {
	if err := s.checkReadPermission(ctx, docID); err != nil {
		return "", err
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	repo, exists := s.repos[docID]
	if !exists {
		return "", fmt.Errorf("仓库不存在")
	}

	return repo.DocID, nil
}

// DeleteRepository 删除仓库
func (s *SimpleVersionControlImpl) DeleteRepository(ctx context.Context, docID string) error {
	if err := s.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.repos, docID)

	s.logger.WithField("doc_id", docID).Info("文档仓库已删除")
	return nil
}

// SaveVersion 保存版本
func (s *SimpleVersionControlImpl) SaveVersion(ctx context.Context, docID string, content []byte, author string, message string) (string, error) {
	if err := s.checkWritePermission(ctx, docID); err != nil {
		return "", err
	}

	s.mutex.RLock()
	repo, exists := s.repos[docID]
	s.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("仓库不存在")
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	branch := repo.Branches[repo.CurrentBranch]
	if branch == nil {
		return "", fmt.Errorf("分支不存在")
	}

	// 创建新版本
	versionID := s.generateVersionID()
	version := &SimpleVersion{
		ID:        versionID,
		Content:   content,
		Author:    author,
		Message:   message,
		Timestamp: time.Now(),
		Parent:    branch.Head,
	}

	// 添加到分支
	branch.Versions = append(branch.Versions, version)
	branch.Head = versionID

	s.logger.WithFields(logrus.Fields{
		"doc_id":     docID,
		"version_id": versionID,
		"author":     author,
		"branch":     repo.CurrentBranch,
	}).Info("版本已保存")

	return versionID, nil
}

// GetVersion 获取版本
func (s *SimpleVersionControlImpl) GetVersion(ctx context.Context, docID string, versionID string) ([]byte, error) {
	if err := s.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	s.mutex.RLock()
	repo, exists := s.repos[docID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("仓库不存在")
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	// 在所有分支中查找版本
	for _, branch := range repo.Branches {
		for _, version := range branch.Versions {
			if version.ID == versionID {
				return version.Content, nil
			}
		}
	}

	return nil, fmt.Errorf("版本不存在")
}

// GetVersions 获取版本列表
func (s *SimpleVersionControlImpl) GetVersions(ctx context.Context, docID string) ([]*VersionInfo, error) {
	if err := s.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	s.mutex.RLock()
	repo, exists := s.repos[docID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("仓库不存在")
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	branch := repo.Branches[repo.CurrentBranch]
	if branch == nil {
		return nil, fmt.Errorf("分支不存在")
	}

	var versions []*VersionInfo
	for _, version := range branch.Versions {
		versionInfo := &VersionInfo{
			ID:        version.ID,
			DocID:     docID,
			Author:    version.Author,
			Message:   version.Message,
			Timestamp: version.Timestamp,
			Size:      int64(len(version.Content)),
			Branch:    branch.Name,
		}
		versions = append(versions, versionInfo)
	}

	return versions, nil
}

// CompareVersions 比较版本
func (s *SimpleVersionControlImpl) CompareVersions(ctx context.Context, docID string, fromVersion, toVersion string) (*DiffResult, error) {
	if err := s.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	// 获取两个版本的内容
	fromContent, err := s.GetVersion(ctx, docID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("获取源版本失败: %w", err)
	}

	toContent, err := s.GetVersion(ctx, docID, toVersion)
	if err != nil {
		return nil, fmt.Errorf("获取目标版本失败: %w", err)
	}

	// 简单的差异比较
	changes := s.simpleDiff(string(fromContent), string(toContent))

	summary := &DiffSummary{}
	for _, change := range changes {
		switch change.Type {
		case "add":
			summary.LinesAdded++
			summary.CharactersAdded += len(change.NewContent)
		case "delete":
			summary.LinesRemoved++
			summary.CharactersRemoved += len(change.OldContent)
		case "modify":
			summary.LinesAdded++
			summary.LinesRemoved++
			summary.CharactersAdded += len(change.NewContent)
			summary.CharactersRemoved += len(change.OldContent)
		}
	}

	return &DiffResult{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Changes:     changes,
		Summary:     summary,
		Timestamp:   time.Now(),
	}, nil
}

// CreateBranch 创建分支
func (s *SimpleVersionControlImpl) CreateBranch(ctx context.Context, docID string, branchName string) error {
	if err := s.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	s.mutex.RLock()
	repo, exists := s.repos[docID]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("仓库不存在")
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.Branches[branchName]; exists {
		return fmt.Errorf("分支已存在")
	}

	// 从当前分支创建新分支
	currentBranch := repo.Branches[repo.CurrentBranch]
	newBranch := &SimpleBranch{
		Name:     branchName,
		Versions: make([]*SimpleVersion, 0),
		Head:     currentBranch.Head,
	}

	repo.Branches[branchName] = newBranch

	s.logger.WithFields(logrus.Fields{
		"doc_id":  docID,
		"branch":  branchName,
		"from":    repo.CurrentBranch,
	}).Info("分支已创建")

	return nil
}

// GetBranches 获取分支列表
func (s *SimpleVersionControlImpl) GetBranches(ctx context.Context, docID string) ([]string, error) {
	if err := s.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	s.mutex.RLock()
	repo, exists := s.repos[docID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("仓库不存在")
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	var branches []string
	for branchName := range repo.Branches {
		branches = append(branches, branchName)
	}

	return branches, nil
}

// SwitchBranch 切换分支
func (s *SimpleVersionControlImpl) SwitchBranch(ctx context.Context, docID string, branchName string) error {
	if err := s.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	s.mutex.RLock()
	repo, exists := s.repos[docID]
	s.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("仓库不存在")
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.Branches[branchName]; !exists {
		return fmt.Errorf("分支不存在")
	}

	oldBranch := repo.CurrentBranch
	repo.CurrentBranch = branchName

	s.logger.WithFields(logrus.Fields{
		"doc_id":    docID,
		"old_branch": oldBranch,
		"new_branch": branchName,
	}).Info("分支已切换")

	return nil
}

// 辅助方法

// checkReadPermission 检查读权限
func (s *SimpleVersionControlImpl) checkReadPermission(ctx context.Context, docID string) error {
	// 简化实现，总是允许读
	return nil
}

// checkWritePermission 检查写权限
func (s *SimpleVersionControlImpl) checkWritePermission(ctx context.Context, docID string) error {
	// 简化实现，总是允许写
	return nil
}

// generateVersionID 生成版本ID
func (s *SimpleVersionControlImpl) generateVersionID() string {
	return fmt.Sprintf("v%d", time.Now().UnixNano())
}

// simpleDiff 简单的差异比较
func (s *SimpleVersionControlImpl) simpleDiff(oldContent, newContent string) []*Change {
	var changes []*Change

	if oldContent == newContent {
		return changes
	}

	if oldContent == "" && newContent != "" {
		// 全部新增
		changes = append(changes, &Change{
			Type:       "add",
			LineNumber: 1,
			OldContent: "",
			NewContent: newContent,
		})
	} else if oldContent != "" && newContent == "" {
		// 全部删除
		changes = append(changes, &Change{
			Type:       "delete",
			LineNumber: 1,
			OldContent: oldContent,
			NewContent: "",
		})
	} else {
		// 简单修改
		changes = append(changes, &Change{
			Type:       "modify",
			LineNumber: 1,
			OldContent: oldContent,
			NewContent: newContent,
		})
	}

	return changes
}

// getRepoPath 获取仓库路径
func (s *SimpleVersionControlImpl) getRepoPath(docID string) string {
	return filepath.Join(s.basePath, docID)
}