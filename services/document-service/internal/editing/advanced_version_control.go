package editing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

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

	// 高级版本操作
	CreateTag(ctx context.Context, docID string, tagName, versionID, message string) error
	GetTags(ctx context.Context, docID string) ([]*TagInfo, error)
	CherryPick(ctx context.Context, docID, versionID, targetBranch string) error
	RebaseBranch(ctx context.Context, docID, branchName, baseBranch string) error

	// 冲突解决
	DetectConflicts(ctx context.Context, docID, sourceBranch, targetBranch string) ([]*ConflictInfo, error)
	ResolveConflict(ctx context.Context, docID, conflictID string, resolution string) error
	GetConflicts(ctx context.Context, docID string) ([]*ConflictInfo, error)

	// 版本历史和统计
	GetVersionGraph(ctx context.Context, docID string) (*VersionGraph, error)
	GetContributors(ctx context.Context, docID string) ([]*Contributor, error)
	GetActivityLog(ctx context.Context, docID, options *ActivityLogOptions) ([]*ActivityEntry, error)
}

// MergeStrategy 合并策略
type MergeStrategy string

const (
	MergeStrategyFastForward MergeStrategy = "fast-forward"
	MergeStrategyThreeWay    MergeStrategy = "three-way"
	MergeStrategySquash       MergeStrategy = "squash"
	MergeStrategyRebase       MergeStrategy = "rebase"
)

// BranchDiffResult 分支差异结果
type BranchDiffResult struct {
	SourceBranch     string           `json:"source_branch"`
	TargetBranch     string           `json:"target_branch"`
	DivergencePoint   string           `json:"divergence_point"`
	SourceCommits    []*VersionInfo   `json:"source_commits"`
	TargetCommits    []*VersionInfo   `json:"target_commits"`
	CommonAncestors   []*VersionInfo   `json:"common_ancestors"`
	HasConflicts     bool             `json:"has_conflicts"`
	FileChanges      []*FileChange    `json:"file_changes"`
	DivergenceTime   time.Time        `json:"divergence_time"`
}

// MergePreview 合并预览
type MergePreview struct {
	CanMerge     bool               `json:"can_merge"`
	Strategy     MergeStrategy       `json:"strategy"`
	Conflicts    []*ConflictInfo     `json:"conflicts"`
	Changes      []*FileChange       `json:"changes"`
	Preview      string              `json:"preview"`
	Warnings     []string            `json:"warnings"`
	EstimatedTime time.Duration       `json:"estimated_time"`
}

// ConflictInfo 冲突信息
type ConflictInfo struct {
	ID          string            `json:"id"`
	FilePath    string            `json:"file_path"`
	ConflictType ConflictType      `json:"conflict_type"`
	BaseContent string            `json:"base_content"`
	SourceContent string           `json:"source_content"`
	TargetContent string           `json:"target_content"`
	Description string            `json:"description"`
	Severity    ConflictSeverity   `json:"severity"`
	CreatedAt   time.Time         `json:"created_at"`
	Status      ConflictStatus     `json:"status"`
	Resolution  string            `json:"resolution,omitempty"`
}

// ConflictType 冲突类型
type ConflictType string

const (
	ConflictTypeContent   ConflictType = "content"
	ConflictTypeStructure ConflictType = "structure"
	ConflictTypeBinary   ConflictType = "binary"
	ConflictTypePermission ConflictType = "permission"
)

// ConflictSeverity 冲突严重程度
type ConflictSeverity string

const (
	ConflictSeverityLow      ConflictSeverity = "low"
	ConflictSeverityMedium   ConflictSeverity = "medium"
	ConflictSeverityHigh     ConflictSeverity = "high"
	ConflictSeverityCritical ConflictSeverity = "critical"
)

// ConflictStatus 冲突状态
type ConflictStatus string

const (
	ConflictStatusPending    ConflictStatus = "pending"
	ConflictStatusResolved  ConflictStatus = "resolved"
	ConflictStatusIgnored    ConflictStatus = "ignored"
)

// BranchProtectionConfig 分支保护配置
type BranchProtectionConfig struct {
	RequireReviews      bool          `json:"require_reviews"`
	RequiredReviewers   []string      `json:"required_reviewers"`
	RequireStatusChecks  bool          `json:"require_status_checks"`
	RequiredStatusChecks []string      `json:"required_status_checks"`
	AllowForcePushes    bool          `json:"allow_force_pushes"`
	RequireLinearHistory bool          `json:"require_linear_history"`
	RestrictPushes      bool          `json:"restrict_pushes"`
	AllowedPushers      []string      `json:"allowed_pushers"`
	EnforceAdmins       bool          `json:"enforce_admins"`
	DismissStaleReviews bool          `json:"dismiss_stale_reviews"`
}

// TagInfo 标签信息
type TagInfo struct {
	Name       string    `json:"name"`
	VersionID  string    `json:"version_id"`
	Message    string    `json:"message"`
	Author     string    `json:"author"`
	CreatedAt  time.Time `json:"created_at"`
	Tags       map[string]string `json:"tags"`
}

// VersionGraph 版本图
type VersionGraph struct {
	DocID       string                    `json:"doc_id"`
	Nodes       []*VersionGraphNode      `json:"nodes"`
	Edges       []*VersionGraphEdge      `json:"edges"`
	Branches    map[string]*BranchGraph  `json:"branches"`
	Commits     int                       `json:"total_commits"`
	BranchesCnt int                       `json:"total_branches"`
}

// VersionGraphNode 版本图节点
type VersionGraphNode struct {
	ID        string            `json:"id"`
	VersionID string            `json:"version_id"`
	Author    string            `json:"author"`
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Branch    string            `json:"branch"`
	Parents   []string          `json:"parents"`
	Metadata  map[string]string `json:"metadata"`
}

// VersionGraphEdge 版本图边
type VersionGraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "parent", "merge", "rebase"
}

// BranchGraph 分支图
type BranchGraph struct {
	Name      string `json:"name"`
	Head      string `json:"head"`
	Base      string `json:"base"`
	Commits   int    `json:"commits"`
	Protected bool   `json:"protected"`
}

// Contributor 贡献者信息
type Contributor struct {
	UserID       string    `json:"user_id"`
	UserName     string    `json:"user_name"`
	Email        string    `json:"email"`
	Commits      int       `json:"commits"`
	Changes      int       `json:"changes"`
	Additions    int       `json:"additions"`
	Deletions    int       `json:"deletions"`
	FirstCommit  time.Time `json:"first_commit"`
	LastCommit   time.Time `json:"last_commit"`
}

// ActivityLogOptions 活动日志选项
type ActivityLogOptions struct {
	Branch     string    `json:"branch,omitempty"`
	Author     string    `json:"author,omitempty"`
	Since      time.Time `json:"since,omitempty"`
	Until      time.Time `json:"until,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Offset     int       `json:"offset,omitempty"`
	Actions    []string  `json:"actions,omitempty"`
}

// ActivityEntry 活动条目
type ActivityEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Author    string    `json:"author"`
	Branch    string    `json:"branch"`
	Version   string    `json:"version,omitempty"`
	Message   string    `json:"message"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// AdvancedVersionControlImpl 高级版本控制实现
type AdvancedVersionControlImpl struct {
	*SimpleVersionControlImpl
	logger               *logrus.Logger
	authService          AuthService
	conflictResolver     ConflictResolver
	branchProtector      BranchProtector
	eventManager         *EventManager
	repos                map[string]*AdvancedRepo
	mutex                sync.RWMutex
}

// AdvancedRepo 高级仓库
type AdvancedRepo struct {
	*SimpleRepo
	ProtectedBranches map[string]*BranchProtectionConfig `json:"protected_branches"`
	Tags             map[string]*TagInfo               `json:"tags"`
	Conflicts        map[string]*ConflictInfo            `json:"conflicts"`
	ActivityLog      []*ActivityEntry                    `json:"activity_log"`
	VersionGraph     *VersionGraph                       `json:"version_graph"`
	Contributors     map[string]*Contributor             `json:"contributors"`
	mutex            sync.RWMutex
}

// ConflictResolver 冲突解决器接口
type ConflictResolver interface {
	DetectConflicts(source, target []byte) ([]*ConflictInfo, error)
	ResolveConflict(conflict *ConflictInfo, resolution string) ([]byte, error)
	SuggestResolution(conflict *ConflictInfo) (string, error)
}

// BranchProtector 分支保护器接口
type BranchProtector interface {
	ProtectBranch(branchName string, config *BranchProtectionConfig) error
	UnprotectBranch(branchName string) error
	IsProtected(branchName string) bool
	CanOperate(userID, branchName, operation string) bool
}

// EventManager 事件管理器
type EventManager struct {
	listeners map[string][]EventListener
	mutex     sync.RWMutex
}

// EventListener 事件监听器接口
type EventListener interface {
	OnEvent(eventType string, data interface{}) error
}

// NewAdvancedVersionControlService 创建高级版本控制服务
func NewAdvancedVersionControlService(
	basePath string,
	logger *logrus.Logger,
	authService AuthService,
) AdvancedVersionControlService {
	simpleService := NewSimpleVersionControlService(basePath, logger, authService)

	advanced := &AdvancedVersionControlImpl{
		SimpleVersionControlImpl: simpleService.(*SimpleVersionControlImpl),
		logger:               logger,
		authService:          authService,
		conflictResolver:     NewSmartConflictResolver(logger),
		branchProtector:      NewBranchProtector(logger),
		eventManager:         NewEventManager(),
		repos:                make(map[string]*AdvancedRepo),
	}

	return advanced
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

	// 计算分支统计信息
	commits := len(branch.Versions)
	var latestCommit *SimpleVersion
	if commits > 0 {
		latestCommit = branch.Versions[commits-1]
	}

	// 计算分支年龄
	var age time.Duration
	if latestCommit != nil {
		age = time.Since(latestCommit.Timestamp)
	}

	// 获取贡献者
	contributors := a.calculateBranchContributors(repo, branchName)

	return &BranchInfo{
		Name:         branchName,
		CommitHash:   branch.Head,
		CommitTime:   time.Now(),
		IsDefault:    branchName == "main" || branchName == "master",
		Author:       "",
		Message:      "",
		CommitsCount: commits,
		Age:          age,
		Contributors: contributors,
		Protected:    repo.ProtectedBranches[branchName] != nil,
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
		return nil, fmt.Errorf("源分支或目标分支不存在")
	}

	// 找到共同祖先
	divergencePoint := a.findDivergencePoint(source, target)

	// 获取各自的提交历史
	sourceCommits := a.getCommitsSince(source, divergencePoint)
	targetCommits := a.getCommitsSince(target, divergencePoint)

	// 分析文件变更
	fileChanges := a.analyzeFileChanges(sourceCommits, targetCommits)

	// 检测潜在冲突
	conflicts := a.detectPotentialConflicts(fileChanges)

	return &BranchDiffResult{
		SourceBranch:   sourceBranch,
		TargetBranch:   targetBranch,
		DivergencePoint: divergencePoint,
		SourceCommits:  a.convertToVersionInfo(sourceCommits),
		TargetCommits:  a.convertToVersionInfo(targetCommits),
		CommonAncestors: a.getCommonAncestors(source, target),
		HasConflicts:   len(conflicts) > 0,
		FileChanges:    fileChanges,
		DivergenceTime: time.Now(),
	}, nil
}

// MergeBranch 合并分支
func (a *AdvancedVersionControlImpl) MergeBranch(ctx context.Context, docID, sourceBranch, targetBranch string, strategy MergeStrategy) (*MergeResult, error) {
	if err := a.checkWritePermission(ctx, docID); err != nil {
		return nil, err
	}

	// 检查分支保护
	if !a.canMergeBranch(ctx, docID, sourceBranch, targetBranch) {
		return nil, fmt.Errorf("分支保护策略禁止此合并操作")
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
		return nil, fmt.Errorf("源分支或目标分支不存在")
	}

	// 预检查冲突
	conflicts, err := a.DetectConflicts(ctx, docID, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("冲突检测失败: %w", err)
	}

	if len(conflicts) > 0 {
		return &MergeResult{
			Success:   false,
			Conflict:  true,
			Conflicts: a.convertToConflictInfo(conflicts),
			Message:   fmt.Sprintf("发现 %d 个冲突，请先解决冲突", len(conflicts)),
		}, nil
	}

	// 根据策略执行合并
	result, err := a.executeMerge(repo, source, target, strategy)
	if err != nil {
		return nil, fmt.Errorf("合并失败: %w", err)
	}

	// 记录活动
	a.logActivity(repo, "merge", sourceBranch, result.Message, map[string]interface{}{
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"strategy":      string(strategy),
		"commit_hash":   result.CommitHash,
	})

	a.logger.WithFields(logrus.Fields{
		"doc_id":        docID,
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"strategy":      strategy,
		"success":       result.Success,
	}).Info("分支合并完成")

	return result, nil
}

// PreviewMerge 预览合并
func (a *AdvancedVersionControlImpl) PreviewMerge(ctx context.Context, docID, sourceBranch, targetBranch string) (*MergePreview, error) {
	if err := a.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	// 检测冲突
	conflicts, err := a.DetectConflicts(ctx, docID, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("冲突检测失败: %w", err)
	}

	// 分析变更
	diff, err := a.CompareBranches(ctx, docID, sourceBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("分支比较失败: %w", err)
	}

	// 检测最佳合并策略
	strategy := a.detectBestMergeStrategy(diff)

	// 生成预览
	preview := a.generateMergePreview(diff, conflicts, strategy)

	return &MergePreview{
		CanMerge:     len(conflicts) == 0,
		Strategy:     strategy,
		Conflicts:    a.convertToConflictInfo(conflicts),
		Changes:      diff.FileChanges,
		Preview:      preview,
		Warnings:     a.generateMergeWarnings(diff, conflicts),
		EstimatedTime: a.estimateMergeTime(diff, conflicts),
	}, nil
}

// ProtectBranch 保护分支
func (a *AdvancedVersionControlImpl) ProtectBranch(ctx context.Context, docID, branchName string, config *BranchProtectionConfig) error {
	if err := a.checkAdminPermission(ctx, docID); err != nil {
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

	if _, exists := repo.Branches[branchName]; !exists {
		return fmt.Errorf("分支不存在")
	}

	// 应用保护配置
	repo.ProtectedBranches[branchName] = config

	// 记录活动
	a.logActivity(repo, "protect_branch", branchName, "分支保护设置已更新", map[string]interface{}{
		"config": config,
	})

	a.logger.WithFields(logrus.Fields{
		"doc_id":     docID,
		"branch":     branchName,
		"protected":  true,
	}).Info("分支保护已启用")

	return nil
}

// UnprotectBranch 取消分支保护
func (a *AdvancedVersionControlImpl) UnprotectBranch(ctx context.Context, docID, branchName string) error {
	if err := a.checkAdminPermission(ctx, docID); err != nil {
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

	delete(repo.ProtectedBranches, branchName)

	// 记录活动
	a.logActivity(repo, "unprotect_branch", branchName, "分支保护已移除", nil)

	a.logger.WithFields(logrus.Fields{
		"doc_id":    docID,
		"branch":    branchName,
		"protected": false,
	}).Info("分支保护已移除")

	return nil
}

// IsBranchProtected 检查分支是否受保护
func (a *AdvancedVersionControlImpl) IsBranchProtected(ctx context.Context, docID, branchName string) bool {
	a.mutex.RLock()
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return false
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	return repo.ProtectedBranches[branchName] != nil
}

// CreateTag 创建标签
func (a *AdvancedVersionControlImpl) CreateTag(ctx context.Context, docID string, tagName, versionID, message string) error {
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

	// 验证版本存在
	var foundVersion *SimpleVersion
	for _, branch := range repo.Branches {
		for _, version := range branch.Versions {
			if version.ID == versionID {
				foundVersion = version
				break
			}
		}
		if foundVersion != nil {
			break
		}
	}

	if foundVersion == nil {
		return fmt.Errorf("版本不存在")
	}

	// 创建标签
	tag := &TagInfo{
		Name:      tagName,
		VersionID: versionID,
		Message:   message,
		Author:    foundVersion.Author,
		CreatedAt: time.Now(),
		Tags:      make(map[string]string),
	}

	if repo.Tags == nil {
		repo.Tags = make(map[string]*TagInfo)
	}
	repo.Tags[tagName] = tag

	// 记录活动
	a.logActivity(repo, "create_tag", tagName, message, map[string]interface{}{
		"version_id": versionID,
	})

	a.logger.WithFields(logrus.Fields{
		"doc_id":     docID,
		"tag_name":   tagName,
		"version_id": versionID,
	}).Info("标签已创建")

	return nil
}

// GetTags 获取标签列表
func (a *AdvancedVersionControlImpl) GetTags(ctx context.Context, docID string) ([]*TagInfo, error) {
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

	var tags []*TagInfo
	for _, tag := range repo.Tags {
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

	// 查找源版本
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
		return fmt.Errorf("源版本不存在")
	}

	target := repo.Branches[targetBranch]
	if target == nil {
		return fmt.Errorf("目标分支不存在")
	}

	// 创建新版本（模拟cherry-pick）
	newVersion := &SimpleVersion{
		ID:        a.generateVersionID(),
		Content:   sourceVersion.Content,
		Author:    sourceVersion.Author,
		Message:   fmt.Sprintf("Cherry-pick: %s", sourceVersion.Message),
		Timestamp: time.Now(),
		Parent:    target.Head,
	}

	target.Versions = append(target.Versions, newVersion)
	target.Head = newVersion.ID

	// 记录活动
	a.logActivity(repo, "cherry_pick", targetBranch, newVersion.Message, map[string]interface{}{
		"source_version": versionID,
		"new_version":    newVersion.ID,
	})

	a.logger.WithFields(logrus.Fields{
		"doc_id":         docID,
		"target_branch": targetBranch,
		"source_version": versionID,
		"new_version":    newVersion.ID,
	}).Info("Cherry-pick 完成")

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
	a.logActivity(repo, "rebase", branchName, "分支已重新基于", map[string]interface{}{
		"base_branch": baseBranch,
		"commits":     len(newVersions),
	})

	a.logger.WithFields(logrus.Fields{
		"doc_id":       docID,
		"branch":       branchName,
		"base_branch":  baseBranch,
		"commits":      len(newVersions),
	}).Info("Rebase 完成")

	return nil
}

// DetectConflicts 检测冲突
func (a *AdvancedVersionControlImpl) DetectConflicts(ctx context.Context, docID, sourceBranch, targetBranch string) ([]*ConflictInfo, error) {
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
		return nil, fmt.Errorf("源分支或目标分支不存在")
	}

	// 获取最新版本内容
	var sourceContent, targetContent []byte
	if len(source.Versions) > 0 {
		sourceContent = source.Versions[len(source.Versions)-1].Content
	}
	if len(target.Versions) > 0 {
		targetContent = target.Versions[len(target.Versions)-1].Content
	}

	// 使用冲突解决器检测冲突
	conflicts, err := a.conflictResolver.DetectConflicts(sourceContent, targetContent)
	if err != nil {
		return nil, fmt.Errorf("冲突检测失败: %w", err)
	}

	// 存储冲突信息
	if repo.Conflicts == nil {
		repo.Conflicts = make(map[string]*ConflictInfo)
	}

	for _, conflict := range conflicts {
		repo.Conflicts[conflict.ID] = conflict
	}

	return conflicts, nil
}

// ResolveConflict 解决冲突
func (a *AdvancedVersionControlImpl) ResolveConflict(ctx context.Context, docID, conflictID string, resolution string) error {
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

	conflict, exists := repo.Conflicts[conflictID]
	if !exists {
		return fmt.Errorf("冲突不存在")
	}

	// 使用冲突解决器解决冲突
	resolvedContent, err := a.conflictResolver.ResolveConflict(conflict, resolution)
	if err != nil {
		return fmt.Errorf("冲突解决失败: %w", err)
	}

	// 更新冲突状态
	conflict.Status = ConflictStatusResolved
	conflict.Resolution = resolution
	repo.Conflicts[conflictID] = conflict

	// 记录活动
	a.logActivity(repo, "resolve_conflict", conflictID, "冲突已解决", map[string]interface{}{
		"file_path": conflict.FilePath,
		"resolution": resolution,
	})

	a.logger.WithFields(logrus.Fields{
		"doc_id":      docID,
		"conflict_id": conflictID,
		"file_path":   conflict.FilePath,
	}).Info("冲突已解决")

	return nil
}

// GetConflicts 获取冲突列表
func (a *AdvancedVersionControlImpl) GetConflicts(ctx context.Context, docID string) ([]*ConflictInfo, error) {
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

	var conflicts []*ConflictInfo
	for _, conflict := range repo.Conflicts {
		if conflict.Status == ConflictStatusPending {
			conflicts = append(conflicts, conflict)
		}
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

	// 构建版本图
	graph := &VersionGraph{
		DocID:    docID,
		Nodes:    make([]*VersionGraphNode, 0),
		Edges:    make([]*VersionGraphEdge, 0),
		Branches: make(map[string]*BranchGraph),
	}

	var totalCommits int

	// 构建分支图
	for branchName, branch := range repo.Branches {
		branchGraph := &BranchGraph{
			Name:      branchName,
			Head:      branch.Head,
			Commits:   len(branch.Versions),
			Protected: repo.ProtectedBranches[branchName] != nil,
		}

		graph.Branches[branchName] = branchGraph
		totalCommits += len(branch.Versions)

		// 构建节点和边
		for i, version := range branch.Versions {
			node := &VersionGraphNode{
				ID:        version.ID,
				VersionID: version.ID,
				Author:    version.Author,
				Message:   version.Message,
				Timestamp: version.Timestamp,
				Branch:    branchName,
				Parents:   make([]string, 0),
				Metadata:  make(map[string]string),
			}

			if i > 0 {
				node.Parents = append(node.Parents, branch.Versions[i-1].ID)
			}

			graph.Nodes = append(graph.Nodes, node)

			// 添加边
			if i > 0 {
				edge := &VersionGraphEdge{
					From: branch.Versions[i-1].ID,
					To:   version.ID,
					Type: "parent",
				}
				graph.Edges = append(graph.Edges, edge)
			}
		}
	}

	graph.Commits = totalCommits
	graph.BranchesCnt = len(repo.Branches)

	return graph, nil
}

// GetContributors 获取贡献者列表
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

	// 统计贡献者
	contributorStats := make(map[string]*Contributor)
	var firstCommit, lastCommit time.Time

	for _, branch := range repo.Branches {
		for _, version := range branch.Versions {
			if firstCommit.IsZero() || version.Timestamp.Before(firstCommit) {
				firstCommit = version.Timestamp
			}
			if version.Timestamp.After(lastCommit) {
				lastCommit = version.Timestamp
			}

			contributor, exists := contributorStats[version.Author]
			if !exists {
				contributor = &Contributor{
					UserID:    version.Author,
					UserName:  version.Author,
					Commits:   0,
					Changes:   0,
					Additions: 0,
					Deletions: 0,
				}
				contributorStats[version.Author] = contributor
			}

			contributor.Commits++
			contributor.Changes++
			contributor.Additions += len(version.Content)
			contributor.FirstCommit = firstCommit
			contributor.LastCommit = lastCommit
		}
	}

	var contributors []*Contributor
	for _, contributor := range contributorStats {
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
	repo, exists := a.repos[docID]
	a.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("仓库不存在")
	}

	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	// 过滤活动日志
	var activities []*ActivityEntry
	for _, activity := range repo.ActivityLog {
		// 应用过滤条件
		if options.Branch != "" && activity.Branch != options.Branch {
			continue
		}
		if options.Author != "" && activity.Author != options.Author {
			continue
		}
		if !options.Since.IsZero() && activity.Timestamp.Before(options.Since) {
			continue
		}
		if !options.Until.IsZero() && activity.Timestamp.After(options.Until) {
			continue
		}
		if len(options.Actions) > 0 && !contains(options.Actions, activity.Action) {
			continue
		}

		activities = append(activities, activity)
	}

	// 应用限制和偏移
	if options.Offset > 0 && options.Offset < len(activities) {
		activities = activities[options.Offset:]
	}

	if options.Limit > 0 && len(activities) > options.Limit {
		activities = activities[:options.Limit]
	}

	return activities, nil
}

// 辅助方法实现

// canMergeBranch 检查是否可以合并
func (a *AdvancedVersionControlImpl) canMergeBranch(ctx context.Context, docID, sourceBranch, targetBranch string) bool {
	// 检查目标分支保护
	if a.IsBranchProtected(ctx, docID, targetBranch) {
		// 获取用户ID（简化实现）
		userID := "system"
		return a.branchProtector.CanOperate(userID, targetBranch, "merge")
	}
	return true
}

// findDivergencePoint 找到分歧点
func (a *AdvancedVersionControlImpl) findDivergencePoint(branch1, branch2 *SimpleBranch) string {
	// 简化实现：返回共同的最老版本
	if len(branch1.Versions) == 0 || len(branch2.Versions) == 0 {
		return ""
	}

	// 查找共同的版本
	for i := len(branch1.Versions) - 1; i >= 0; i-- {
		for j := len(branch2.Versions) - 1; j >= 0; j-- {
			if branch1.Versions[i].ID == branch2.Versions[j].ID {
				return branch1.Versions[i].ID
			}
		}
	}

	return ""
}

// getCommitsSince 获取指定版本之后的提交
func (a *AdvancedVersionControlImpl) getCommitsSince(branch *SimpleBranch, sinceVersion string) []*SimpleVersion {
	var commits []*SimpleVersion
	found := false

	for i := len(branch.Versions) - 1; i >= 0; i-- {
		version := branch.Versions[i]
		if version.ID == sinceVersion {
			found = true
			continue
		}
		if found {
			commits = append([]*SimpleVersion{version}, commits...)
		}
	}

	return commits
}

// convertToVersionInfo 转换为版本信息
func (a *AdvancedVersionControlImpl) convertToVersionInfo(versions []*SimpleVersion) []*VersionInfo {
	var result []*VersionInfo
	for _, version := range versions {
		result = append(result, &VersionInfo{
			ID:        version.ID,
			Author:    version.Author,
			Message:   version.Message,
			Timestamp: version.Timestamp,
			Size:      int64(len(version.Content)),
			Branch:    "",
		})
	}
	return result
}

// getCommonAncestors 获取共同祖先
func (a *AdvancedVersionControlImpl) getCommonAncestors(branch1, branch2 *SimpleBranch) []*VersionInfo {
	// 简化实现：返回共同祖先列表
	ancestors := make(map[string]*SimpleVersion)

	for _, version := range branch1.Versions {
		for _, otherVersion := range branch2.Versions {
			if version.ID == otherVersion.ID {
				ancestors[version.ID] = version
				break
			}
		}
	}

	var result []*VersionInfo
	for _, ancestor := range ancestors {
		result = append(result, &VersionInfo{
			ID:        ancestor.ID,
			Author:    ancestor.Author,
			Message:   ancestor.Message,
			Timestamp: ancestor.Timestamp,
			Size:      int64(len(ancestor.Content)),
			Branch:    "",
		})
	}

	return result
}

// analyzeFileChanges 分析文件变更
func (a *AdvancedVersionControlImpl) analyzeFileChanges(sourceCommits, targetCommits []*SimpleVersion) []*FileChange {
	// 简化实现：假设每次提交都包含一个文件变更
	var changes []*FileChange

	for _, commit := range sourceCommits {
		change := &FileChange{
			Path:       "document.txt", // 简化实现
			Operation:  "modify",
			Content:    commit.Content,
			Attributes: make(map[string]string),
		}
		changes = append(changes, change)
	}

	return changes
}

// detectPotentialConflicts 检测潜在冲突
func (a *AdvancedVersionControlImpl) detectPotentialConflicts(changes []*FileChange) []*ConflictInfo {
	// 简化实现：基于文件路径检测冲突
	filePaths := make(map[string]int)
	for _, change := range changes {
		filePaths[change.Path]++
	}

	var conflicts []*ConflictInfo
	for filePath, count := range filePaths {
		if count > 1 {
			conflict := &ConflictInfo{
				ID:          generateConflictID(filePath),
				FilePath:    filePath,
				ConflictType: ConflictTypeContent,
				BaseContent: "",
				SourceContent: "",
				TargetContent: "",
				Description: fmt.Sprintf("文件 %s 在两个分支中都有修改", filePath),
				Severity:    ConflictSeverityMedium,
				CreatedAt:   time.Now(),
				Status:      ConflictStatusPending,
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// executeMerge 执行合并
func (a *AdvancedVersionControlImpl) executeMerge(repo *AdvancedRepo, source, target *SimpleBranch, strategy MergeStrategy) (*MergeResult, error) {
	var result *MergeResult

	switch strategy {
	case MergeStrategyFastForward:
		result = a.executeFastForwardMerge(repo, source, target)
	case MergeStrategyThreeWay:
		result = a.executeThreeWayMerge(repo, source, target)
	case MergeStrategySquash:
		result = a.executeSquashMerge(repo, source, target)
	case MergeStrategyRebase:
		result = a.executeRebaseMerge(repo, source, target)
	default:
		return nil, fmt.Errorf("不支持的合并策略: %s", strategy)
	}

	return result, nil
}

// executeFastForwardMerge 执行快进合并
func (a *AdvancedVersionControlImpl) executeFastForwardMerge(repo *AdvancedRepo, source, target *SimpleBranch) (*MergeResult, error) {
	// 简化实现：直接复制源分支的最新版本到目标分支
	if len(source.Versions) == 0 {
		return nil, fmt.Errorf("源分支没有提交")
	}

	latestSource := source.Versions[len(source.Versions)-1]

	newVersion := &SimpleVersion{
		ID:        a.generateVersionID(),
		Content:   latestSource.Content,
		Author:    latestSource.Author,
		Message:   fmt.Sprintf("Merge fast-forward from %s: %s", source.Name, latestSource.Message),
		Timestamp: time.Now(),
		Parent:    target.Head,
	}

	target.Versions = append(target.Versions, newVersion)
	target.Head = newVersion.ID

	return &MergeResult{
		Success:    true,
		CommitHash: newVersion.ID,
		MergeType:  string(MergeStrategyFastForward),
		Message:    "快进合并成功",
	}, nil
}

// executeThreeWayMerge 执行三方合并
func (a *AdvancedVersionControlImpl) executeThreeWayMerge(repo *AdvancedRepo, source, target *SimpleBranch) (*MergeResult, error) {
	// 简化实现：合并两个分支的最新内容
	var sourceContent, targetContent []byte

	if len(source.Versions) > 0 {
		sourceContent = source.Versions[len(source.Versions)-1].Content
	}
	if len(target.Versions) > 0 {
		targetContent = target.Versions[len(target.Versions)-1].Content
	}

	// 简单合并：以目标内容为基础，追加源内容
	mergedContent := targetContent
	if len(sourceContent) > 0 {
		mergedContent = append(mergedContent, []byte("\n--- Merge from source branch ---\n")...)
		mergedContent = append(mergedContent, sourceContent...)
	}

	newVersion := &SimpleVersion{
		ID:        a.generateVersionID(),
		Content:   mergedContent,
		Author:    "system",
		Message:   fmt.Sprintf("Three-way merge: %s into %s", source.Name, target.Name),
		Timestamp: time.Now(),
		Parent:    target.Head,
	}

	target.Versions = append(target.Versions, newVersion)
	target.Head = newVersion.ID

	return &MergeResult{
		Success:    true,
		CommitHash: newVersion.ID,
		MergeType:  string(MergeStrategyThreeWay),
		Message:    "三方合并成功",
	}, nil
}

// executeSquashMerge 执行压缩合并
func (a *AdvancedVersionControlImpl) executeSquashMerge(repo *AdvancedRepo, source, target *SimpleBranch) (*MergeResult, error) {
	// 简化实现：压缩所有源分支的提交为一个
	var messages []string
	var mergedContent []byte

	for _, version := range source.Versions {
		messages = append(messages, version.Message)
		if len(mergedContent) == 0 {
			mergedContent = version.Content
		} else {
			mergedContent = append(mergedContent, []byte("\n")...)
			mergedContent = append(mergedContent, version.Content...)
		}
	}

	squashMessage := fmt.Sprintf("Squash merge %d commits from %s", len(source.Versions), source.Name)
	for _, msg := range messages {
		squashMessage += fmt.Sprintf("\n- %s", msg)
	}

	newVersion := &SimpleVersion{
		ID:        a.generateVersionID(),
		Content:   mergedContent,
		Author:    "system",
		Message:   squashMessage,
		Timestamp: time.Now(),
		Parent:    target.Head,
	}

	target.Versions = append(target.Versions, newVersion)
	target.Head = newVersion.ID

	return &MergeResult{
		Success:    true,
		CommitHash: newVersion.ID,
		MergeType:  string(MergeStrategySquash),
		Message:    "压缩合并成功",
	}, nil
}

// executeRebaseMerge 执行变基合并
func (a *AdvancedVersionControlImpl) executeRebaseMerge(repo *AdvancedRepo, source, target *SimpleBranch) (*MergeResult, error) {
	// 简化实现：将源分支的提交重新应用到目标分支
	var newVersions []*SimpleVersion
	currentHead := target.Head

	for _, version := range source.Versions {
		newVersion := &SimpleVersion{
			ID:        a.generateVersionID(),
			Content:   version.Content,
			Author:    version.Author,
			Message:   fmt.Sprintf("Rebase: %s", version.Message),
			Timestamp: time.Now(),
			Parent:    currentHead,
		}

		newVersions = append(newVersions, newVersion)
		currentHead = newVersion.ID
	}

	// 更新源分支
	source.Versions = newVersions
	if len(newVersions) > 0 {
		source.Head = newVersions[len(newVersions)-1].ID
	}

	return &MergeResult{
		Success:    true,
		CommitHash: source.Head,
		MergeType:  string(MergeStrategyRebase),
		Message:    "变基合并成功",
	}, nil
}

// detectBestMergeStrategy 检测最佳合并策略
func (a *AdvancedVersionControlImpl) detectBestMergeStrategy(diff *BranchDiffResult) MergeStrategy {
	// 简化实现：基于差异情况选择策略
	if len(diff.SourceCommits) == 0 {
		return MergeStrategyFastForward
	}
	if len(diff.SourceCommits) < 5 {
		return MergeStrategyThreeWay
	}
	return MergeStrategySquash
}

// generateMergePreview 生成合并预览
func (a *AdvancedVersionControlImpl) generateMergePreview(diff *BranchDiffResult, conflicts []*ConflictInfo, strategy MergeStrategy) string {
	preview := fmt.Sprintf("合并预览 (%s)\n", strategy)
	preview += fmt.Sprintf("源分支提交: %d\n", len(diff.SourceCommits))
	preview += fmt.Sprintf("目标分支提交: %d\n", len(diff.TargetCommits))
	preview += fmt.Sprintf("文件变更: %d\n", len(diff.FileChanges))

	if len(conflicts) > 0 {
		preview += fmt.Sprintf("冲突数量: %d\n", len(conflicts))
		preview += "冲突详情:\n"
		for _, conflict := range conflicts {
			preview += fmt.Sprintf("- %s: %s\n", conflict.FilePath, conflict.Description)
		}
	} else {
		preview += "无冲突\n"
	}

	return preview
}

// generateMergeWarnings 生成合并警告
func (a *AdvancedVersionControlImpl) generateMergeWarnings(diff *BranchDiffResult, conflicts []*ConflictInfo) []string {
	var warnings []string

	if len(conflicts) > 0 {
		warnings = append(warnings, fmt.Sprintf("发现 %d 个冲突，需要手动解决", len(conflicts)))
	}

	if len(diff.FileChanges) > 10 {
		warnings = append(warnings, "大量文件变更，请仔细检查")
	}

	if diff.DivergenceTime.Sub(time.Now().Add(-30*24*time.Hour)) < 0 {
		warnings = append(warnings, "分支分歧时间较长，可能需要解决更多冲突")
	}

	return warnings
}

// estimateMergeTime 估算合并时间
func (a *AdvancedVersionControlImpl) estimateMergeTime(diff *BranchDiffResult, conflicts []*ConflictInfo) time.Duration {
	baseTime := 5 * time.Second
	conflictTime := time.Duration(len(conflicts)) * 30 * time.Second
	changeTime := time.Duration(len(diff.FileChanges)) * 2 * time.Second

	return baseTime + conflictTime + changeTime
}

// convertToConflictInfo 转换冲突信息
func (a *AdvancedVersionControlImpl) convertToConflictInfo(conflicts []*ConflictInfo) []*ConflictInfo {
	return conflicts
}

// calculateBranchContributors 计算分支贡献者
func (a *AdvancedVersionControlImpl) calculateBranchContributors(repo *AdvancedRepo, branchName string) []string {
	contributors := make(map[string]bool)
	branch := repo.Branches[branchName]

	for _, version := range branch.Versions {
		contributors[version.Author] = true
	}

	var result []string
	for contributor := range contributors {
		result = append(result, contributor)
	}

	return result
}

// logActivity 记录活动
func (a *AdvancedVersionControlImpl) logActivity(repo *AdvancedRepo, action, target, message string, metadata map[string]interface{}) {
	entry := &ActivityEntry{
		ID:        generateActivityID(),
		Timestamp: time.Now(),
		Action:    action,
		Author:    "system", // 简化实现
		Branch:    target,
		Message:   message,
		Metadata:  metadata,
	}

	repo.ActivityLog = append(repo.ActivityLog, entry)

	// 限制日志数量
	if len(repo.ActivityLog) > 1000 {
		repo.ActivityLog = repo.ActivityLog[len(repo.ActivityLog)-1000:]
	}
}

// checkReadPermission 检查读权限
func (a *AdvancedVersionControlImpl) checkReadPermission(ctx context.Context, docID string) error {
	return nil // 简化实现
}

// checkWritePermission 检查写权限
func (a *AdvancedVersionControlImpl) checkWritePermission(ctx context.Context, docID string) error {
	return nil // 简化实现
}

// checkAdminPermission 检查管理员权限
func (a *AdvancedVersionControlImpl) checkAdminPermission(ctx context.Context, docID string) error {
	return nil // 简化实现
}

// generateVersionID 生成版本ID
func (a *AdvancedVersionControlImpl) generateVersionID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:8])
}

// generateConflictID 生成冲突ID
func generateConflictID(filePath string) string {
	hash := sha256.Sum256([]byte(filePath + time.Now().String()))
	return hex.EncodeToString(hash[:8])
}

// generateActivityID 生成活动ID
func generateActivityID() string {
	hash := sha256.Sum256([]byte(time.Now().String()))
	return hex.EncodeToString(hash[:8])
}

// contains 检查字符串数组是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 简单版本控制服务的高级方法扩展
func (a *AdvancedVersionControlImpl) InitializeRepository(ctx context.Context, docID string) error {
	err := a.SimpleVersionControlImpl.InitializeRepository(ctx, docID)
	if err != nil {
		return err
	}

	// 升级为高级仓库
	a.mutex.Lock()
	defer a.mutex.Unlock()

	simpleRepo := a.repos[docID].(*SimpleRepo)
	advancedRepo := &AdvancedRepo{
		SimpleRepo:        simpleRepo,
		ProtectedBranches: make(map[string]*BranchProtectionConfig),
		Tags:             make(map[string]*TagInfo),
		Conflicts:        make(map[string]*ConflictInfo),
		ActivityLog:      make([]*ActivityEntry, 0),
		VersionGraph:     nil,
		Contributors:     make(map[string]*Contributor),
	}

	a.repos[docID] = advancedRepo
	return nil
}

// NewSmartConflictResolver 创建智能冲突解决器
func NewSmartConflictResolver(logger *logrus.Logger) ConflictResolver {
	return &SmartConflictResolver{logger: logger}
}

// SmartConflictResolver 智能冲突解决器
type SmartConflictResolver struct {
	logger *logrus.Logger
}

// DetectConflicts 检测冲突
func (r *SmartConflictResolver) DetectConflicts(source, target []byte) ([]*ConflictInfo, error) {
	// 简化的冲突检测实现
	sourceStr := string(source)
	targetStr := string(target)

	if sourceStr == targetStr {
		return nil, nil
	}

	conflict := &ConflictInfo{
		ID:          generateConflictID("document.txt"),
		FilePath:    "document.txt",
		ConflictType: ConflictTypeContent,
		BaseContent: "",
		SourceContent: sourceStr,
		TargetContent: targetStr,
		Description:  "文档内容存在差异",
		Severity:    ConflictSeverityMedium,
		CreatedAt:    time.Now(),
		Status:       ConflictStatusPending,
	}

	return []*ConflictInfo{conflict}, nil
}

// ResolveConflict 解决冲突
func (r *SmartConflictResolver) ResolveConflict(conflict *ConflictInfo, resolution string) ([]byte, error) {
	// 根据解决方案生成解决后的内容
	switch resolution {
	case "source":
		return []byte(conflict.SourceContent), nil
	case "target":
		return []byte(conflict.TargetContent), nil
	case "manual":
		return []byte(resolution), nil
	case "auto":
		// 简单的自动合并：选择较长的内容
		if len(conflict.SourceContent) > len(conflict.TargetContent) {
			return []byte(conflict.SourceContent), nil
		}
		return []byte(conflict.TargetContent), nil
	default:
		return nil, fmt.Errorf("未知的解决方案: %s", resolution)
	}
}

// SuggestResolution 建议解决方案
func (r *SmartConflictResolver) SuggestResolution(conflict *ConflictInfo) (string, error) {
	// 基于冲突类型和内容提供建议
	switch conflict.ConflictType {
	case ConflictTypeContent:
		if conflict.SourceContent == "" {
			return "target", nil
		}
		if conflict.TargetContent == "" {
			return "source", nil
		}
		if len(conflict.SourceContent) > len(conflict.TargetContent)*2 {
			return "source", nil
		}
		if len(conflict.TargetContent) > len(conflict.SourceContent)*2 {
			return "target", nil
		}
		return "auto", nil
	default:
		return "manual", nil
	}
}

// NewBranchProtector 创建分支保护器
func NewBranchProtector(logger *logrus.Logger) BranchProtector {
	return &DefaultBranchProtector{logger: logger}
}

// DefaultBranchProtector 默认分支保护器
type DefaultBranchProtector struct {
	logger *logrus.Logger
}

// ProtectBranch 保护分支
func (p *DefaultBranchProtector) ProtectBranch(branchName string, config *BranchProtectionConfig) error {
	p.logger.WithField("branch", branchName).Info("分支保护已设置")
	return nil
}

// UnprotectBranch 取消分支保护
func (p *DefaultBranchProtector) UnprotectBranch(branchName string) error {
	p.logger.WithField("branch", branchName).Info("分支保护已移除")
	return nil
}

// IsProtected 检查分支是否受保护
func (p *DefaultBranchProtector) IsProtected(branchName string) bool {
	protectedBranches := []string{"main", "master", "develop", "release/*"}
	for _, protected := range protectedBranches {
		if matchBranch(branchName, protected) {
			return true
		}
	}
	return false
}

// CanOperate 检查是否可以操作
func (p *DefaultBranchProtector) CanOperate(userID, branchName, operation string) bool {
	// 简化实现：管理员可以操作所有分支
	admins := []string{"admin", "system"}
	for _, admin := range admins {
		if userID == admin {
			return true
		}
	}

	// 非管理员不能操作受保护的分支
	if p.IsProtected(branchName) {
		return false
	}

	return true
}

// matchBranch 匹配分支模式
func matchBranch(branch, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(branch, prefix)
	}
	return branch == pattern
}

// NewEventManager 创建事件管理器
func NewEventManager() *EventManager {
	return &EventManager{
		listeners: make(map[string][]EventListener),
	}
}

// AddListener 添加监听器
func (m *EventManager) AddListener(eventType string, listener EventListener) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.listeners[eventType] = append(m.listeners[eventType], listener)
}

// OnEvent 触发事件
func (m *EventManager) OnEvent(eventType string, data interface{}) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, listener := range m.listeners[eventType] {
		if err := listener.OnEvent(eventType, data); err != nil {
			// 记录错误但继续执行
			continue
		}
	}

	return nil
}