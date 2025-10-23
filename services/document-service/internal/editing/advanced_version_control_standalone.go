package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ============ 基础类型定义 ============

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

// AuthService 认证服务接口
type AuthService interface {
	AuthenticateUser(ctx context.Context, token string) (string, error)
	HasPermission(ctx context.Context, userID, resource, action string) error
}

// ============ 简单版本控制实现 ============

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

// ============ 高级版本控制类型定义 ============

// BranchInfo 分支信息
type BranchInfo struct {
	Name         string    `json:"name"`
	Head         string    `json:"head"`
	CommitsCount int       `json:"commits_count"`
	Author       string    `json:"author"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Protected    bool      `json:"protected"`
}

// FileChange 文件变更
type FileChange struct {
	Path         string    `json:"path"`
	Type         string    `json:"type"` // "add", "delete", "modify", "rename"
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
	OldPath      string    `json:"old_path,omitempty"`
	Content      []byte    `json:"content,omitempty"`
}

// BranchDiffResult 分支差异结果
type BranchDiffResult struct {
	SourceBranch  string            `json:"source_branch"`
	TargetBranch  string            `json:"target_branch"`
	DivergencePoint string          `json:"divergence_point"`
	SourceCommits []*VersionInfo    `json:"source_commits"`
	TargetCommits []*VersionInfo    `json:"target_commits"`
	FilesChanged  []*FileChange     `json:"files_changed"`
	IsAhead       bool              `json:"is_ahead"`
	IsBehind      bool              `json:"is_behind"`
	CommitsAhead  int               `json:"commits_ahead"`
	CommitsBehind int               `json:"commits_behind"`
	CanFastForward bool             `json:"can_fast_forward"`
}

// MergeStrategy 合并策略
type MergeStrategy string

const (
	MergeStrategyFastForward MergeStrategy = "fast-forward"
	MergeStrategyThreeWay    MergeStrategy = "three-way"
	MergeStrategySquash      MergeStrategy = "squash"
	MergeStrategyRebase      MergeStrategy = "rebase"
)

// MergeResult 合并结果
type MergeResult struct {
	Success      bool              `json:"success"`
	MergeType    MergeStrategy     `json:"merge_type"`
	MergeCommit  *VersionInfo      `json:"merge_commit,omitempty"`
	Conflicts    []*ConflictInfo   `json:"conflicts,omitempty"`
	Message      string            `json:"message"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// MergePreview 合并预览
type MergePreview struct {
	CanMerge      bool              `json:"can_merge"`
	Conflicts     []*ConflictInfo   `json:"conflicts,omitempty"`
	Preview       string            `json:"preview"`
	Warnings      []string          `json:"warnings,omitempty"`
	Strategy      MergeStrategy     `json:"strategy"`
	EstimatedTime time.Duration     `json:"estimated_time"`
}

// BranchProtectionConfig 分支保护配置
type BranchProtectionConfig struct {
	RequireReviews      bool     `json:"require_reviews"`
	RequiredReviewers   []string `json:"required_reviewers"`
	RequireStatusChecks  bool     `json:"require_status_checks"`
	AllowForcePushes    bool     `json:"allow_force_pushes"`
	RequireLinearHistory bool     `json:"require_linear_history"`
	RestrictPushes      bool     `json:"restrict_pushes"`
	AllowedPushers      []string `json:"allowed_pushers"`
	EnforceAdmins       bool     `json:"enforce_admins"`
}

// TagInfo 标签信息
type TagInfo struct {
	Name      string    `json:"name"`
	VersionID string    `json:"version_id"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

// ConflictInfo 冲突信息
type ConflictInfo struct {
	ID          string    `json:"id"`
	FilePath    string    `json:"file_path"`
	Type        string    `json:"type"` // "content", "permission", "format"
	Status      string    `json:"status"` // "pending", "resolving", "resolved"
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	Resolution  string    `json:"resolution,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// ConflictStatus 冲突状态
const (
	ConflictStatusPending  = "pending"
	ConflictStatusResolving = "resolving"
	ConflictStatusResolved = "resolved"
)

// VersionGraph 版本图
type VersionGraph struct {
	DocID       string           `json:"doc_id"`
	Nodes       []*VersionNode   `json:"nodes"`
	Edges       []*VersionEdge   `json:"edges"`
	Branches    map[string]string `json:"branches"` // branch -> head version
	Commits     int              `json:"commits"`
	BranchesCnt int              `json:"branches_count"`
}

// VersionNode 版本节点
type VersionNode struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Branch    string    `json:"branch"`
	Parents   []string  `json:"parents"`
	Children  []string  `json:"children"`
}

// VersionEdge 版本边
type VersionEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "parent", "merge", "branch"
}

// Contributor 贡献者
type Contributor struct {
	UserName   string    `json:"user_name"`
	Commits    int       `json:"commits"`
	Additions  int       `json:"additions"`
	Deletions  int       `json:"deletions"`
	LastActive time.Time `json:"last_active"`
	Branches   []string  `json:"branches"`
}

// ActivityEntry 活动记录
type ActivityEntry struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Target    string                 `json:"target"`
	Message   string                 `json:"message"`
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ActivityLogOptions 活动日志选项
type ActivityLogOptions struct {
	Limit       int               `json:"limit"`
	Offset      int               `json:"offset"`
	Since       *time.Time        `json:"since,omitempty"`
	Until       *time.Time        `json:"until,omitempty"`
	Actions     []string          `json:"actions,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	Branch      string            `json:"branch,omitempty"`
	SortBy      string            `json:"sort_by"` // "timestamp", "action"
	SortOrder   string            `json:"sort_order"` // "asc", "desc"
}

// ============ 高级版本控制实现 ============

// AdvancedVersionControlService 高级版本控制服务
type AdvancedVersionControlService interface {
	// 扩展基本版本控制功能
	SimpleVersionControlService

	// 高级分支管理
	BranchInfo(ctx context.Context, docID, branchName string) (*BranchInfo, error)
	CompareBranches(ctx context.Context, docID, sourceBranch, targetBranch string) (*BranchDiffResult, error)

	// 合并操作
	MergeBranch(ctx context.Context, docID, sourceBranch, targetBranch string, strategy MergeStrategy) (*MergeResult, error)
	PreviewMerge(ctx context.Context, docID, sourceBranch, targetBranch string) (*MergePreview, error)

	// 分支保护
	ProtectBranch(ctx context.Context, docID, branchName string, config *BranchProtectionConfig) error
	UnprotectBranch(ctx context.Context, docID, branchName string) error
	IsBranchProtected(ctx context.Context, docID, branchName string) bool

	// 标签管理
	CreateTag(ctx context.Context, docID string, tagName, versionID, message string) error
	GetTags(ctx context.Context, docID string) ([]*TagInfo, error)

	// 变基和挑选
	CherryPick(ctx context.Context, docID, versionID, targetBranch string) error
	RebaseBranch(ctx context.Context, docID, branchName, baseBranch string) error

	// 冲突检测和解决
	DetectConflicts(ctx context.Context, docID, sourceBranch, targetBranch string) ([]*ConflictInfo, error)
	ResolveConflict(ctx context.Context, docID, conflictID string, resolution string) error
	GetConflicts(ctx context.Context, docID string) ([]*ConflictInfo, error)

	// 版本图和统计
	GetVersionGraph(ctx context.Context, docID string) (*VersionGraph, error)
	GetContributors(ctx context.Context, docID string) ([]*Contributor, error)
	GetActivityLog(ctx context.Context, docID string, options *ActivityLogOptions) ([]*ActivityEntry, error)
}

// AdvancedVersionControlImpl 高级版本控制实现
type AdvancedVersionControlImpl struct {
	*SimpleVersionControlImpl
	protectedBranches map[string]*BranchProtectionConfig
	branchProtector   BranchProtector
	conflictResolver  ConflictResolver
	eventManager      *EventManager
	activities        map[string][]*ActivityEntry
	tags              map[string]map[string]*TagInfo // docID -> tagName -> TagInfo
	conflicts         map[string][]*ConflictInfo     // docID -> conflicts
}

// BranchProtector 分支保护器
type BranchProtector interface {
	ProtectBranch(branchName string, config *BranchProtectionConfig) error
	UnprotectBranch(branchName string) error
	IsProtected(branchName string) bool
	CanOperate(userID, branchName, operation string) bool
}

// ConflictResolver 冲突解决器
type ConflictResolver interface {
	DetectConflicts(source, target []byte) ([]*ConflictInfo, error)
	ResolveConflict(conflict *ConflictInfo, resolution string) ([]byte, error)
	SuggestResolution(conflict *ConflictInfo) (string, error)
}

// EventManager 事件管理器
type EventManager struct {
	listeners map[string][]EventListener
}

// EventListener 事件监听器
type EventListener interface {
	OnEvent(eventType string, data interface{}) error
}

// NewAdvancedVersionControlService 创建高级版本控制服务
func NewAdvancedVersionControlService(
	basePath string,
	logger *logrus.Logger,
	authService AuthService,
) AdvancedVersionControlService {
	simple := NewSimpleVersionControlService(basePath, logger, authService).(*SimpleVersionControlImpl)

	return &AdvancedVersionControlImpl{
		SimpleVersionControlImpl: simple,
		protectedBranches:        make(map[string]*BranchProtectionConfig),
		branchProtector:          NewBranchProtector(logger),
		conflictResolver:         NewSmartConflictResolver(logger),
		eventManager:             NewEventManager(),
		activities:               make(map[string][]*ActivityEntry),
		tags:                     make(map[string]map[string]*TagInfo),
		conflicts:                make(map[string][]*ConflictInfo),
	}
}

// AdvancedRepo 高级仓库（扩展SimpleRepo）
type AdvancedRepo struct {
	*SimpleRepo
	Tags       []*TagInfo
	Conflicts  []*ConflictInfo
	Activities []*ActivityEntry
	Metadata   map[string]interface{}
}

// BranchInfo 获取分支信息
func (a *AdvancedVersionControlImpl) BranchInfo(ctx context.Context, docID, branchName string) (*BranchInfo, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	a.mutex.RLock()
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("仓库不存在")
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	branch := repo.Branches[branchName]
	if branch == nil {
		return nil, fmt.Errorf("分支不存在")
	}

	var author string
	var createdAt time.Time
	if len(branch.Versions) > 0 {
		firstVersion := branch.Versions[0]
		author = firstVersion.Author
		createdAt = firstVersion.Timestamp
	}

	return &BranchInfo{
		Name:         branchName,
		Head:         branch.Head,
		CommitsCount: len(branch.Versions),
		Author:       author,
		CreatedAt:    createdAt,
		UpdatedAt:    time.Now(),
		Protected:    a.IsBranchProtected(ctx, docID, branchName),
	}, nil
}

// CompareBranches 比较分支
func (a *AdvancedVersionControlImpl) CompareBranches(ctx context.Context, docID, sourceBranch, targetBranch string) (*BranchDiffResult, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	a.mutex.RLock()
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("仓库不存在")
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	source := repo.Branches[sourceBranch]
	target := repo.Branches[targetBranch]

	if source == nil || target == nil {
		return nil, fmt.Errorf("分支不存在")
	}

	// 找到分歧点
	divergencePoint := a.findDivergencePoint(source, target)

	// 获取分歧后的提交
	sourceCommits := a.getCommitsSince(source, divergencePoint)
	targetCommits := a.getCommitsSince(target, divergencePoint)

	// 分析文件变更
	fileChanges := a.analyzeFileChanges(sourceCommits, targetCommits)

	// 检测潜在冲突
	_ = a.detectPotentialConflicts(fileChanges) // 冲突检测，结果暂不使用

	return &BranchDiffResult{
		SourceBranch:    sourceBranch,
		TargetBranch:    targetBranch,
		DivergencePoint: divergencePoint,
		SourceCommits:   a.convertToVersionInfo(sourceCommits),
		TargetCommits:   a.convertToVersionInfo(targetCommits),
		FilesChanged:    fileChanges,
		IsAhead:         len(sourceCommits) > len(targetCommits),
		IsBehind:        len(sourceCommits) < len(targetCommits),
		CommitsAhead:    len(sourceCommits),
		CommitsBehind:   len(targetCommits),
		CanFastForward:  len(targetCommits) == 0,
	}, nil
}

// MergeBranch 合并分支
func (a *AdvancedVersionControlImpl) MergeBranch(ctx context.Context, docID, sourceBranch, targetBranch string, strategy MergeStrategy) (*MergeResult, error) {
	if err := a.checkWritePermission(ctx, docID); err != nil {
		return nil, err
	}

	// 检查分支保护
	if a.IsBranchProtected(ctx, docID, targetBranch) {
		return &MergeResult{
			Success: false,
			Message: "目标分支受保护，无法合并",
		}, nil
	}

	a.mutex.RLock()
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("仓库不存在")
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	source := repo.Branches[sourceBranch]
	target := repo.Branches[targetBranch]

	if source == nil || target == nil {
		return nil, fmt.Errorf("分支不存在")
	}

	// 内部比较分支（避免死锁）
	diff := a.compareBranchesInternal(source, target, sourceBranch, targetBranch)

	// 检测冲突
	conflicts := a.detectPotentialConflicts(diff.FilesChanged)
	if len(conflicts) > 0 && strategy == MergeStrategyFastForward {
		return &MergeResult{
			Success:   false,
			Message:   "存在冲突，无法快进合并",
			Conflicts: conflicts,
		}, nil
	}

	// 执行合并
	advancedRepo := &AdvancedRepo{SimpleRepo: repo}
	result, err := a.executeMerge(advancedRepo, source, target, strategy)
	if err != nil {
		return nil, fmt.Errorf("执行合并失败: %w", err)
	}

	// 记录活动
	a.logActivity(advancedRepo, "merge", targetBranch, "分支已合并", map[string]interface{}{
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"strategy":      strategy,
		"success":       result.Success,
	})

	return result, nil
}

// compareBranchesInternal 内部分支比较（不获取额外锁）
func (a *AdvancedVersionControlImpl) compareBranchesInternal(source, target *SimpleBranch, sourceBranch, targetBranch string) *BranchDiffResult {
	// 找到分歧点
	divergencePoint := a.findDivergencePoint(source, target)

	// 获取分歧后的提交
	sourceCommits := a.getCommitsSince(source, divergencePoint)
	targetCommits := a.getCommitsSince(target, divergencePoint)

	// 分析文件变更
	fileChanges := a.analyzeFileChanges(sourceCommits, targetCommits)

	// 检测潜在冲突
	_ = a.detectPotentialConflicts(fileChanges) // 冲突检测，结果暂不使用

	return &BranchDiffResult{
		SourceBranch:    sourceBranch,
		TargetBranch:    targetBranch,
		DivergencePoint: divergencePoint,
		SourceCommits:   a.convertToVersionInfo(sourceCommits),
		TargetCommits:   a.convertToVersionInfo(targetCommits),
		FilesChanged:    fileChanges,
		IsAhead:         len(sourceCommits) > len(targetCommits),
		IsBehind:        len(sourceCommits) < len(targetCommits),
		CommitsAhead:    len(sourceCommits),
		CommitsBehind:   len(targetCommits),
		CanFastForward:  len(targetCommits) == 0,
	}
}

// PreviewMerge 预览合并
func (a *AdvancedVersionControlImpl) PreviewMerge(ctx context.Context, docID, sourceBranch, targetBranch string) (*MergePreview, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	diff, err := a.CompareBranches(ctx, docID, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("比较分支失败: %w", err)
	}

	// 检测冲突
	conflicts := a.detectPotentialConflicts(diff.FilesChanged)

	// 选择最佳合并策略
	strategy := a.detectBestMergeStrategy(diff)

	// 生成预览内容
	preview := a.generateMergePreview(diff, conflicts, strategy)

	// 生成警告
	warnings := a.generateMergeWarnings(diff, conflicts)

	// 估算合并时间
	estimatedTime := a.estimateMergeTime(diff, conflicts)

	return &MergePreview{
		CanMerge:      len(conflicts) == 0,
		Conflicts:     conflicts,
		Preview:       preview,
		Warnings:      warnings,
		Strategy:      strategy,
		EstimatedTime: estimatedTime,
	}, nil
}

// ProtectBranch 保护分支
func (a *AdvancedVersionControlImpl) ProtectBranch(ctx context.Context, docID, branchName string, config *BranchProtectionConfig) error {
	if err := a.checkAdminPermission(ctx, docID); err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%s", docID, branchName)
	a.protectedBranches[key] = config

	// 使用分支保护器
	err := a.branchProtector.ProtectBranch(branchName, config)
	if err != nil {
		return fmt.Errorf("保护分支失败: %w", err)
	}

	a.logger.WithFields(logrus.Fields{
		"doc_id":     docID,
		"branch":     branchName,
		"protected":  true,
	}).Info("分支保护设置成功")

	return nil
}

// UnprotectBranch 取消保护分支
func (a *AdvancedVersionControlImpl) UnprotectBranch(ctx context.Context, docID, branchName string) error {
	if err := a.checkAdminPermission(ctx, docID); err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%s", docID, branchName)
	delete(a.protectedBranches, key)

	// 使用分支保护器
	err := a.branchProtector.UnprotectBranch(branchName)
	if err != nil {
		return fmt.Errorf("取消保护分支失败: %w", err)
	}

	a.logger.WithFields(logrus.Fields{
		"doc_id":     docID,
		"branch":     branchName,
		"protected":  false,
	}).Info("分支保护取消成功")

	return nil
}

// IsBranchProtected 检查分支是否受保护
func (a *AdvancedVersionControlImpl) IsBranchProtected(ctx context.Context, docID, branchName string) bool {
	key := fmt.Sprintf("%s:%s", docID, branchName)
	_, protected := a.protectedBranches[key]
	return protected || a.branchProtector.IsProtected(branchName)
}

// CreateTag 创建标签
func (a *AdvancedVersionControlImpl) CreateTag(ctx context.Context, docID string, tagName, versionID, message string) error {
	if err := a.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	// 验证版本存在（先获取锁验证版本）
	err := a.validateVersionExists(ctx, docID, versionID)
	if err != nil {
		return fmt.Errorf("版本不存在: %w", err)
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, exists := a.tags[docID]; !exists {
		a.tags[docID] = make(map[string]*TagInfo)
	}

	if _, exists := a.tags[docID][tagName]; exists {
		return fmt.Errorf("标签已存在")
	}

	tag := &TagInfo{
		Name:      tagName,
		VersionID: versionID,
		Message:   message,
		Author:    "system", // 这里应该从上下文获取
		Timestamp: time.Now(),
	}

	a.tags[docID][tagName] = tag

	a.logger.WithFields(logrus.Fields{
		"doc_id":     docID,
		"tag_name":   tagName,
		"version_id": versionID,
	}).Info("标签创建成功")

	return nil
}

// validateVersionExists 验证版本是否存在
func (a *AdvancedVersionControlImpl) validateVersionExists(ctx context.Context, docID, versionID string) error {
	a.mutex.RLock()
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("仓库不存在")
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	// 在所有分支中查找版本
	for _, branch := range repo.Branches {
		for _, version := range branch.Versions {
			if version.ID == versionID {
				return nil
			}
		}
	}

	return fmt.Errorf("版本不存在")
}

// GetTags 获取标签列表
func (a *AdvancedVersionControlImpl) GetTags(ctx context.Context, docID string) ([]*TagInfo, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	docTags, exists := a.tags[docID]
	if !exists {
		return []*TagInfo{}, nil
	}

	tags := make([]*TagInfo, 0, len(docTags))
	for _, tag := range docTags {
		tags = append(tags, tag)
	}

	return tags, nil
}

// CherryPick 挑选提交
func (a *AdvancedVersionControlImpl) CherryPick(ctx context.Context, docID, versionID, targetBranch string) error {
	if err := a.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	a.mutex.RLock()
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("仓库不存在")
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	// 找到要挑选的版本
	var sourceVersion *SimpleVersion
	for _, branch := range repo.Branches {
		for _, version := range branch.Versions {
			if version.ID == versionID {
				sourceVersion = version
				break
			}
		}
		if sourceVersion != nil {
			break
		}
	}

	if sourceVersion == nil {
		return fmt.Errorf("版本不存在")
	}

	target := repo.Branches[targetBranch]
	if target == nil {
		return fmt.Errorf("目标分支不存在")
	}

	// 创建新的版本
	newVersion := &SimpleVersion{
		ID:        a.generateVersionID(),
		Content:   sourceVersion.Content,
		Author:    sourceVersion.Author,
		Message:   fmt.Sprintf("Cherry-pick: %s", sourceVersion.Message),
		Timestamp: time.Now(),
		Parent:    target.Head,
	}

	// 添加到目标分支
	target.Versions = append(target.Versions, newVersion)
	target.Head = newVersion.ID

	// 记录活动
	advancedRepo := &AdvancedRepo{SimpleRepo: repo}
	a.logActivity(advancedRepo, "cherry-pick", targetBranch, "提交已挑选", map[string]interface{}{
		"version_id":   versionID,
		"new_version":  newVersion.ID,
		"target_branch": targetBranch,
	})

	a.logger.WithFields(logrus.Fields{
		"doc_id":       docID,
		"version_id":   versionID,
		"target_branch": targetBranch,
		"new_version":  newVersion.ID,
	}).Info("Cherry-pick成功")

	return nil
}

// RebaseBranch 变基分支
func (a *AdvancedVersionControlImpl) RebaseBranch(ctx context.Context, docID, branchName, baseBranch string) error {
	if err := a.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	a.mutex.RLock()
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("仓库不存在")
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	branch := repo.Branches[branchName]
	base := repo.Branches[baseBranch]

	if branch == nil || base == nil {
		return fmt.Errorf("分支不存在")
	}

	// 找到共同祖先
	divergencePoint := a.findDivergencePoint(branch, base)
	if divergencePoint == "" {
		return fmt.Errorf("无法找到共同祖先")
	}

	// 重新应用分支的提交到基础分支
	var newVersions []*SimpleVersion
	foundDivergence := false

	for _, version := range branch.Versions {
		if version.ID == divergencePoint {
			foundDivergence = true
			continue
		}

		// 跳过分歧点之前的提交
		if !foundDivergence {
			continue
		}

		// 创建新的版本
		newVersion := &SimpleVersion{
			ID:        a.generateVersionID(),
			Content:   version.Content,
			Author:    version.Author,
			Message:   fmt.Sprintf("Rebase: %s", version.Message),
			Timestamp: time.Now(),
			Parent:    base.Head,
		}

		newVersions = append(newVersions, newVersion)
		base.Head = newVersion.ID
	}

	// 更新分支
	branch.Versions = newVersions
	if len(newVersions) > 0 {
		branch.Head = newVersions[len(newVersions)-1].ID
	}

	// 记录活动
	advancedRepo := &AdvancedRepo{SimpleRepo: repo}
	a.logActivity(advancedRepo, "rebase", branchName, "分支已重新基于", map[string]interface{}{
		"base_branch": baseBranch,
		"commits":     len(newVersions),
	})

	a.logger.WithFields(logrus.Fields{
		"doc_id":       docID,
		"branch_name":  branchName,
		"base_branch":  baseBranch,
		"commits":      len(newVersions),
	}).Info("Rebase成功")

	return nil
}

// DetectConflicts 检测冲突
func (a *AdvancedVersionControlImpl) DetectConflicts(ctx context.Context, docID, sourceBranch, targetBranch string) ([]*ConflictInfo, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	diff, err := a.CompareBranches(ctx, docID, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("比较分支失败: %w", err)
	}

	conflicts := a.detectPotentialConflicts(diff.FilesChanged)

	// 保存冲突信息
	a.mutex.Lock()
	if a.conflicts[docID] == nil {
		a.conflicts[docID] = make([]*ConflictInfo, 0)
	}
	a.conflicts[docID] = append(a.conflicts[docID], conflicts...)
	a.mutex.Unlock()

	return conflicts, nil
}

// ResolveConflict 解决冲突
func (a *AdvancedVersionControlImpl) ResolveConflict(ctx context.Context, docID, conflictID string, resolution string) error {
	if err := a.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()

	conflicts := a.conflicts[docID]
	if conflicts == nil {
		return fmt.Errorf("不存在冲突")
	}

	// 找到并解决冲突
	for i, conflict := range conflicts {
		if conflict.ID == conflictID {
			conflict.Resolution = resolution
			conflict.Status = ConflictStatusResolved
			now := time.Now()
			conflict.ResolvedAt = &now

			// 从活跃冲突列表中移除
			a.conflicts[docID] = append(conflicts[:i], conflicts[i+1:]...)

			a.logger.WithFields(logrus.Fields{
				"doc_id":     docID,
				"conflict_id": conflictID,
				"resolution":  resolution,
			}).Info("冲突已解决")

			return nil
		}
	}

	return fmt.Errorf("冲突不存在")
}

// GetConflicts 获取冲突列表
func (a *AdvancedVersionControlImpl) GetConflicts(ctx context.Context, docID string) ([]*ConflictInfo, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	conflicts := a.conflicts[docID]
	if conflicts == nil {
		return []*ConflictInfo{}, nil
	}

	return conflicts, nil
}

// GetVersionGraph 获取版本图
func (a *AdvancedVersionControlImpl) GetVersionGraph(ctx context.Context, docID string) (*VersionGraph, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	a.mutex.RLock()
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("仓库不存在")
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	graph := &VersionGraph{
		DocID:    docID,
		Nodes:    make([]*VersionNode, 0),
		Edges:    make([]*VersionEdge, 0),
		Branches: make(map[string]string),
		Commits:  0,
	}

	// 构建节点和边
	versionMap := make(map[string]*VersionNode)

	for branchName, branch := range repo.Branches {
		graph.Branches[branchName] = branch.Head

		for _, version := range branch.Versions {
			node := &VersionNode{
				ID:        version.ID,
				Author:    version.Author,
				Message:   version.Message,
				Timestamp: version.Timestamp,
				Branch:    branchName,
				Parents:   []string{},
				Children:  []string{},
			}

			if version.Parent != "" {
				node.Parents = append(node.Parents, version.Parent)
			}

			graph.Nodes = append(graph.Nodes, node)
			versionMap[version.ID] = node
			graph.Commits++
		}
	}

	// 构建边
	for _, node := range graph.Nodes {
		for _, parentID := range node.Parents {
			if parent, exists := versionMap[parentID]; exists {
				edge := &VersionEdge{
					From: parentID,
					To:   node.ID,
					Type: "parent",
				}
				graph.Edges = append(graph.Edges, edge)
				parent.Children = append(parent.Children, node.ID)
			}
		}
	}

	graph.BranchesCnt = len(repo.Branches)

	return graph, nil
}

// GetContributors 获取贡献者统计
func (a *AdvancedVersionControlImpl) GetContributors(ctx context.Context, docID string) ([]*Contributor, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	a.mutex.RLock()
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("仓库不存在")
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	contributorMap := make(map[string]*Contributor)

	for _, branch := range repo.Branches {
		for _, version := range branch.Versions {
			contributor, exists := contributorMap[version.Author]
			if !exists {
				contributor = &Contributor{
					UserName:  version.Author,
					Commits:   0,
					Additions: 0,
					Deletions: 0,
					Branches:  []string{},
				}
				contributorMap[version.Author] = contributor
			}

			contributor.Commits++
			contributor.Additions += len(version.Content)
			contributor.LastActive = version.Timestamp

			// 添加分支信息
			if !contains(contributor.Branches, branch.Name) {
				contributor.Branches = append(contributor.Branches, branch.Name)
			}
		}
	}

	contributors := make([]*Contributor, 0, len(contributorMap))
	for _, contributor := range contributorMap {
		contributors = append(contributors, contributor)
	}

	return contributors, nil
}

// GetActivityLog 获取活动日志
func (a *AdvancedVersionControlImpl) GetActivityLog(ctx context.Context, docID string, options *ActivityLogOptions) ([]*ActivityEntry, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	activities := a.activities[docID]
	if activities == nil {
		return []*ActivityEntry{}, nil
	}

	// 应用过滤条件
	filtered := make([]*ActivityEntry, 0)
	for _, activity := range activities {
		// 时间过滤
		if options.Since != nil && activity.Timestamp.Before(*options.Since) {
			continue
		}
		if options.Until != nil && activity.Timestamp.After(*options.Until) {
			continue
		}

		// 操作过滤
		if len(options.Actions) > 0 && !contains(options.Actions, activity.Action) {
			continue
		}

		// 用户过滤
		if options.UserID != "" && activity.UserID != options.UserID {
			continue
		}

		// 分支过滤
		if options.Branch != "" {
			if branch, ok := activity.Metadata["branch"].(string); !ok || branch != options.Branch {
				continue
			}
		}

		filtered = append(filtered, activity)
	}

	// 应用分页
	if options.Offset >= len(filtered) {
		return []*ActivityEntry{}, nil
	}

	end := options.Offset + options.Limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[options.Offset:end], nil
}

// ============ 辅助方法实现 ============

// checkAdminPermission 检查管理员权限
func (a *AdvancedVersionControlImpl) checkAdminPermission(ctx context.Context, docID string) error {
	// 简化实现，总是允许管理员操作
	return nil
}

// findDivergencePoint 找到分歧点
func (a *AdvancedVersionControlImpl) findDivergencePoint(branch1, branch2 *SimpleBranch) string {
	// 简化实现，返回最新的共同父版本
	branch1Versions := make(map[string]bool)
	for _, version := range branch1.Versions {
		branch1Versions[version.ID] = true
	}

	for _, version := range branch2.Versions {
		if branch1Versions[version.ID] {
			return version.ID
		}
	}

	return ""
}

// getCommitsSince 获取指定版本之后的提交
func (a *AdvancedVersionControlImpl) getCommitsSince(branch *SimpleBranch, sinceVersion string) []*SimpleVersion {
	if sinceVersion == "" {
		return branch.Versions
	}

	var versions []*SimpleVersion
	found := false
	for _, version := range branch.Versions {
		if version.ID == sinceVersion {
			found = true
			continue
		}
		if found {
			versions = append(versions, version)
		}
	}

	return versions
}

// convertToVersionInfo 转换为版本信息
func (a *AdvancedVersionControlImpl) convertToVersionInfo(versions []*SimpleVersion) []*VersionInfo {
	info := make([]*VersionInfo, len(versions))
	for i, version := range versions {
		info[i] = &VersionInfo{
			ID:        version.ID,
			Author:    version.Author,
			Message:   version.Message,
			Timestamp: version.Timestamp,
			Size:      int64(len(version.Content)),
		}
	}
	return info
}

// analyzeFileChanges 分析文件变更
func (a *AdvancedVersionControlImpl) analyzeFileChanges(sourceCommits, targetCommits []*SimpleVersion) []*FileChange {
	// 简化实现，假设只有一个文件
	var changes []*FileChange

	if len(sourceCommits) > 0 && len(targetCommits) == 0 {
		changes = append(changes, &FileChange{
			Path:      "document.txt",
			Type:      "add",
			Additions: len(sourceCommits[len(sourceCommits)-1].Content),
		})
	} else if len(sourceCommits) == 0 && len(targetCommits) > 0 {
		changes = append(changes, &FileChange{
			Path:      "document.txt",
			Type:      "delete",
			Deletions: len(targetCommits[len(targetCommits)-1].Content),
		})
	} else if len(sourceCommits) > 0 && len(targetCommits) > 0 {
		sourceContent := sourceCommits[len(sourceCommits)-1].Content
		targetContent := targetCommits[len(targetCommits)-1].Content

		changes = append(changes, &FileChange{
			Path:      "document.txt",
			Type:      "modify",
			Additions: len(sourceContent),
			Deletions: len(targetContent),
		})
	}

	return changes
}

// detectPotentialConflicts 检测潜在冲突
func (a *AdvancedVersionControlImpl) detectPotentialConflicts(changes []*FileChange) []*ConflictInfo {
	var conflicts []*ConflictInfo

	for _, change := range changes {
		if change.Type == "modify" {
			conflict := &ConflictInfo{
				ID:          generateConflictID(change.Path),
				FilePath:    change.Path,
				Type:        "content",
				Status:      ConflictStatusPending,
				Description: "文件内容冲突",
				Source:      "source branch",
				Target:      "target branch",
				CreatedAt:   time.Now(),
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// executeMerge 执行合并
func (a *AdvancedVersionControlImpl) executeMerge(repo *AdvancedRepo, source, target *SimpleBranch, strategy MergeStrategy) (*MergeResult, error) {
	switch strategy {
	case MergeStrategyFastForward:
		return a.executeFastForwardMerge(repo, source, target)
	case MergeStrategyThreeWay:
		return a.executeThreeWayMerge(repo, source, target)
	case MergeStrategySquash:
		return a.executeSquashMerge(repo, source, target)
	case MergeStrategyRebase:
		return a.executeRebaseMerge(repo, source, target)
	default:
		return nil, fmt.Errorf("不支持的合并策略: %s", strategy)
	}
}

// executeFastForwardMerge 执行快进合并
func (a *AdvancedVersionControlImpl) executeFastForwardMerge(repo *AdvancedRepo, source, target *SimpleBranch) (*MergeResult, error) {
	if len(target.Versions) > 0 && len(source.Versions) > 0 {
		lastTargetVersion := target.Versions[len(target.Versions)-1]
		lastSourceVersion := source.Versions[len(source.Versions)-1]

		if lastTargetVersion.ID != lastSourceVersion.ID {
			return &MergeResult{
				Success: false,
				Message: "无法快进合并：分支已分歧",
			}, nil
		}
	}

	// 将源分支的所有版本添加到目标分支
	for _, version := range source.Versions {
		target.Versions = append(target.Versions, version)
	}
	target.Head = source.Head

	return &MergeResult{
		Success:   true,
		MergeType: MergeStrategyFastForward,
	}, nil
}

// executeThreeWayMerge 执行三方合并
func (a *AdvancedVersionControlImpl) executeThreeWayMerge(repo *AdvancedRepo, source, target *SimpleBranch) (*MergeResult, error) {
	// 简化实现，创建合并提交
	var mergedContent []byte
	if len(source.Versions) > 0 {
		mergedContent = source.Versions[len(source.Versions)-1].Content
	}

	mergeVersion := &SimpleVersion{
		ID:        a.generateVersionID(),
		Content:   mergedContent,
		Author:    "system",
		Message:   fmt.Sprintf("Merge branch '%s' into '%s'", source.Name, target.Name),
		Timestamp: time.Now(),
		Parent:    target.Head,
	}

	target.Versions = append(target.Versions, mergeVersion)
	target.Head = mergeVersion.ID

	versionInfo := &VersionInfo{
		ID:        mergeVersion.ID,
		Author:    mergeVersion.Author,
		Message:   mergeVersion.Message,
		Timestamp: mergeVersion.Timestamp,
		Size:      int64(len(mergeVersion.Content)),
	}

	return &MergeResult{
		Success:     true,
		MergeType:   MergeStrategyThreeWay,
		MergeCommit: versionInfo,
	}, nil
}

// executeSquashMerge 执行压缩合并
func (a *AdvancedVersionControlImpl) executeSquashMerge(repo *AdvancedRepo, source, target *SimpleBranch) (*MergeResult, error) {
	// 简化实现，将所有变更压缩为一个提交
	var squashContent []byte
	if len(source.Versions) > 0 {
		squashContent = source.Versions[len(source.Versions)-1].Content
	}

	squashVersion := &SimpleVersion{
		ID:        a.generateVersionID(),
		Content:   squashContent,
		Author:    "system",
		Message:   fmt.Sprintf("Squash merge branch '%s' into '%s'", source.Name, target.Name),
		Timestamp: time.Now(),
		Parent:    target.Head,
	}

	target.Versions = append(target.Versions, squashVersion)
	target.Head = squashVersion.ID

	versionInfo := &VersionInfo{
		ID:        squashVersion.ID,
		Author:    squashVersion.Author,
		Message:   squashVersion.Message,
		Timestamp: squashVersion.Timestamp,
		Size:      int64(len(squashVersion.Content)),
	}

	return &MergeResult{
		Success:     true,
		MergeType:   MergeStrategySquash,
		MergeCommit: versionInfo,
	}, nil
}

// executeRebaseMerge 执行变基合并
func (a *AdvancedVersionControlImpl) executeRebaseMerge(repo *AdvancedRepo, source, target *SimpleBranch) (*MergeResult, error) {
	// 简化实现，将源分支的提交重新应用到目标分支
	var rebasedVersions []*SimpleVersion

	for _, version := range source.Versions {
		rebasedVersion := &SimpleVersion{
			ID:        a.generateVersionID(),
			Content:   version.Content,
			Author:    version.Author,
			Message:   version.Message,
			Timestamp: time.Now(),
			Parent:    target.Head,
		}
		rebasedVersions = append(rebasedVersions, rebasedVersion)
		target.Head = rebasedVersion.ID
	}

	target.Versions = append(target.Versions, rebasedVersions...)

	return &MergeResult{
		Success:   true,
		MergeType: MergeStrategyRebase,
	}, nil
}

// detectBestMergeStrategy 检测最佳合并策略
func (a *AdvancedVersionControlImpl) detectBestMergeStrategy(diff *BranchDiffResult) MergeStrategy {
	if diff.CanFastForward {
		return MergeStrategyFastForward
	}
	if len(diff.SourceCommits) == 1 {
		return MergeStrategyThreeWay
	}
	if len(diff.SourceCommits) > 5 {
		return MergeStrategySquash
	}
	return MergeStrategyThreeWay
}

// generateMergePreview 生成合并预览
func (a *AdvancedVersionControlImpl) generateMergePreview(diff *BranchDiffResult, conflicts []*ConflictInfo, strategy MergeStrategy) string {
	preview := fmt.Sprintf("合并预览 (%s)\n", strategy)
	preview += fmt.Sprintf("源分支: %s (%d个提交)\n", diff.SourceBranch, len(diff.SourceCommits))
	preview += fmt.Sprintf("目标分支: %s (%d个提交)\n", diff.TargetBranch, len(diff.TargetCommits))
	preview += fmt.Sprintf("文件变更: %d个\n", len(diff.FilesChanged))
	preview += fmt.Sprintf("冲突: %d个\n", len(conflicts))

	if len(conflicts) > 0 {
		preview += "\n冲突列表:\n"
		for _, conflict := range conflicts {
			preview += fmt.Sprintf("- %s: %s\n", conflict.FilePath, conflict.Description)
		}
	}

	return preview
}

// generateMergeWarnings 生成合并警告
func (a *AdvancedVersionControlImpl) generateMergeWarnings(diff *BranchDiffResult, conflicts []*ConflictInfo) []string {
	var warnings []string

	if len(conflicts) > 0 {
		warnings = append(warnings, "存在需要手动解决的冲突")
	}

	if len(diff.SourceCommits) > 10 {
		warnings = append(warnings, "大量提交需要合并，建议使用压缩合并")
	}

	if diff.CommitsBehind > 20 {
		warnings = append(warnings, "目标分支严重落后，建议先更新目标分支")
	}

	return warnings
}

// estimateMergeTime 估算合并时间
func (a *AdvancedVersionControlImpl) estimateMergeTime(diff *BranchDiffResult, conflicts []*ConflictInfo) time.Duration {
	baseTime := time.Second * 5

	// 根据提交数量调整
	commitTime := time.Duration(len(diff.SourceCommits)) * time.Millisecond * 100

	// 根据冲突数量调整
	conflictTime := time.Duration(len(conflicts)) * time.Second * 10

	return baseTime + commitTime + conflictTime
}

// logActivity 记录活动
func (a *AdvancedVersionControlImpl) logActivity(repo *AdvancedRepo, action, target, message string, metadata map[string]interface{}) {
	activity := &ActivityEntry{
		ID:        generateActivityID(),
		Action:    action,
		Target:    target,
		Message:   message,
		UserID:    "system", // 应该从上下文获取
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	if a.activities[repo.DocID] == nil {
		a.activities[repo.DocID] = make([]*ActivityEntry, 0)
	}
	a.activities[repo.DocID] = append(a.activities[repo.DocID], activity)

	// 触发事件
	a.eventManager.OnEvent(action, activity)
}

// ============ 默认实现 ============

// DefaultBranchProtector 默认分支保护器
type DefaultBranchProtector struct {
	logger     *logrus.Logger
	protected  map[string]*BranchProtectionConfig
}

// NewBranchProtector 创建分支保护器
func NewBranchProtector(logger *logrus.Logger) BranchProtector {
	return &DefaultBranchProtector{
		logger:    logger,
		protected: make(map[string]*BranchProtectionConfig),
	}
}

func (p *DefaultBranchProtector) ProtectBranch(branchName string, config *BranchProtectionConfig) error {
	p.protected[branchName] = config
	return nil
}

func (p *DefaultBranchProtector) UnprotectBranch(branchName string) error {
	delete(p.protected, branchName)
	return nil
}

func (p *DefaultBranchProtector) IsProtected(branchName string) bool {
	_, protected := p.protected[branchName]
	return protected
}

func (p *DefaultBranchProtector) CanOperate(userID, branchName, operation string) bool {
	config, protected := p.protected[branchName]
	if !protected {
		return true
	}

	// 检查是否在允许的操作者列表中
	for _, allowedPusher := range config.AllowedPushers {
		if allowedPusher == userID {
			return true
		}
	}

	return false
}

// SmartConflictResolver 智能冲突解决器
type SmartConflictResolver struct {
	logger *logrus.Logger
}

// NewSmartConflictResolver 创建智能冲突解决器
func NewSmartConflictResolver(logger *logrus.Logger) ConflictResolver {
	return &SmartConflictResolver{
		logger: logger,
	}
}

func (r *SmartConflictResolver) DetectConflicts(source, target []byte) ([]*ConflictInfo, error) {
	var conflicts []*ConflictInfo

	// 简化的冲突检测
	if string(source) != string(target) {
		conflict := &ConflictInfo{
			ID:          generateConflictID("document.txt"),
			FilePath:    "document.txt",
			Type:        "content",
			Status:      ConflictStatusPending,
			Description: "文档内容不一致",
			Source:      string(source),
			Target:      string(target),
			CreatedAt:   time.Now(),
		}
		conflicts = append(conflicts, conflict)
	}

	return conflicts, nil
}

func (r *SmartConflictResolver) ResolveConflict(conflict *ConflictInfo, resolution string) ([]byte, error) {
	// 简化的冲突解决
	return []byte(resolution), nil
}

func (r *SmartConflictResolver) SuggestResolution(conflict *ConflictInfo) (string, error) {
	// 简化的建议生成
	return "建议手动合并冲突内容", nil
}

// NewEventManager 创建事件管理器
func NewEventManager() *EventManager {
	return &EventManager{
		listeners: make(map[string][]EventListener),
	}
}

func (m *EventManager) AddListener(eventType string, listener EventListener) {
	if m.listeners[eventType] == nil {
		m.listeners[eventType] = make([]EventListener, 0)
	}
	m.listeners[eventType] = append(m.listeners[eventType], listener)
}

func (m *EventManager) OnEvent(eventType string, data interface{}) error {
	listeners := m.listeners[eventType]
	for _, listener := range listeners {
		err := listener.OnEvent(eventType, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// ============ 工具函数 ============

// generateConflictID 生成冲突ID
func generateConflictID(filePath string) string {
	return fmt.Sprintf("conflict_%s_%d", filePath, time.Now().UnixNano())
}

// generateActivityID 生成活动ID
func generateActivityID() string {
	return fmt.Sprintf("activity_%d", time.Now().UnixNano())
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}