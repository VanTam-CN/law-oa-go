package editing

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/storage"
	"github.com/sirupsen/logrus"
)

// VersionControlService 文档版本控制服务接口
type VersionControlService interface {
	// 仓库管理
	InitializeRepository(ctx context.Context, docID string) error
	GetRepository(ctx context.Context, docID string) (*git.Repository, error)
	DeleteRepository(ctx context.Context, docID string) error

	// 版本操作
	CreateCommit(ctx context.Context, req *CommitRequest) (*CommitInfo, error)
	GetCommit(ctx context.Context, docID, commitID string) (*CommitInfo, error)
	GetCommitHistory(ctx context.Context, docID string, opts *HistoryOptions) ([]*CommitInfo, error)
	CompareCommits(ctx context.Context, docID, fromCommit, toCommit string) (*DiffResult, error)

	// 分支管理
	CreateBranch(ctx context.Context, docID, branchName, baseCommit string) error
	DeleteBranch(ctx context.Context, docID, branchName string) error
	ListBranches(ctx context.Context, docID string) ([]*BranchInfo, error)
	MergeBranch(ctx context.Context, docID, sourceBranch, targetBranch string) (*MergeResult, error)

	// 文件操作
	Checkout(ctx context.Context, docID string, opts *CheckoutOptions) error
	GetFileContent(ctx context.Context, docID, commitID, filePath string) ([]byte, error)
	ApplyPatch(ctx context.Context, docID, patch string) error

	// 版本管理
	RevertToCommit(ctx context.Context, docID, commitID string) error
	CherryPick(ctx context.Context, docID, commitID, targetBranch string) error
	GetFileHistory(ctx context.Context, docID, filePath string) ([]*CommitInfo, error)
}

// CommitRequest 提交请求
type CommitRequest struct {
	DocumentID     string            `json:"document_id"`
	BranchName     string            `json:"branch_name"`
	AuthorName     string            `json:"author_name"`
	AuthorEmail    string            `json:"author_email"`
	Message        string            `json:"message"`
	Files          []*FileChange     `json:"files"`
	ParentCommits  []string          `json:"parent_commits"`
	Metadata       map[string]string `json:"metadata"`
}

// FileChange 文件变更
type FileChange struct {
	Path       string            `json:"path"`
	Operation  string            `json:"operation"` // "create", "update", "delete"
	Content    []byte            `json:"content"`
	OldContent []byte            `json:"old_content"`
	Attributes map[string]string `json:"attributes"`
}

// CommitInfo 提交信息
type CommitInfo struct {
	Hash         string            `json:"hash"`
	AuthorName   string            `json:"author_name"`
	AuthorEmail  string            `json:"author_email"`
	Message      string            `json:"message"`
	Timestamp    time.Time         `json:"timestamp"`
	ParentHashes []string          `json:"parent_hashes"`
	FileChanges  []*FileChange     `json:"file_changes"`
	Branch       string            `json:"branch"`
	Metadata     map[string]string `json:"metadata"`
}

// HistoryOptions 历史查询选项
type HistoryOptions struct {
	BranchName  string `json:"branch_name"`
	Limit       int    `json:"limit"`
	Skip        int    `json:"skip"`
	Path        string `json:"path"` // 特定文件路径
	Since       *time.Time `json:"since"`
	Until       *time.Time `json:"until"`
}

// BranchInfo 分支信息
type BranchInfo struct {
	Name        string    `json:"name"`
	CommitHash  string    `json:"commit_hash"`
	CommitTime  time.Time `json:"commit_time"`
	IsDefault   bool      `json:"is_default"`
	Author      string    `json:"author"`
	Message     string    `json:"message"`
}

// MergeResult 合并结果
type MergeResult struct {
	Success      bool     `json:"success"`
	CommitHash   string   `json:"commit_hash"`
	Conflict     bool     `json:"conflict"`
	Conflicts    []*Conflict `json:"conflicts"`
	MergeType    string   `json:"merge_type"`
	Message      string   `json:"message"`
}

// Conflict 冲突信息
type Conflict struct {
	FilePath    string `json:"file_path"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	BaseContent string `json:"base_content"`
	OurContent  string `json:"our_content"`
	TheirContent string `json:"their_content"`
	Description string `json:"description"`
}

// DiffResult 差异结果
type DiffResult struct {
	FromCommit  string           `json:"from_commit"`
	ToCommit    string           `json:"to_commit"`
	FileDiffs   []*FileDiff      `json:"file_diffs"`
	Summary     *DiffSummary     `json:"summary"`
	Timestamp   time.Time        `json:"timestamp"`
}

// FileDiff 文件差异
type FileDiff struct {
	Path          string          `json:"path"`
	ChangeType    string          `json:"change_type"` // "add", "modify", "delete", "rename"
	OldPath       string          `json:"old_path"`
	OldContent    []byte          `json:"old_content"`
	NewContent    []byte          `json:"new_content"`
	LinesAdded    int             `json:"lines_added"`
	LinesRemoved  int             `json:"lines_removed"`
	Chunks        []*DiffChunk    `json:"chunks"`
}

// DiffChunk 差异块
type DiffChunk struct {
	OldStart  int    `json:"old_start"`
	NewStart  int    `json:"new_start"`
	Context   string `json:"context"`
	Changes    []*DiffChange `json:"changes"`
}

// DiffChange 差异变更
type DiffChange struct {
	Type     string `json:"type"` // "context", "add", "delete"
	Content  string `json:"content"`
	OldLine  int    `json:"old_line"`
	NewLine  int    `json:"new_line"`
}

// DiffSummary 差异摘要
type DiffSummary struct {
	FilesChanged  int `json:"files_changed"`
	LinesAdded    int `json:"lines_added"`
	LinesRemoved  int `json:"lines_removed"`
	CharactersAdded int `json:"characters_added"`
	CharactersRemoved int `json:"characters_removed"`
}

// CheckoutOptions 检出选项
type CheckoutOptions struct {
	BranchName  string `json:"branch_name"`
	CommitHash  string `json:"commit_hash"`
	CreateBranch bool   `json:"create_branch"`
	Force       bool   `json:"force"`
}

// AuthService 认证服务接口
type AuthService interface {
	AuthenticateUser(ctx context.Context, token string) (string, error)
	HasPermission(ctx context.Context, userID, resource, action string) error
}

// VersionControlImpl 版本控制服务实现
type VersionControlImpl struct {
	repoPath    string
	storage     storage.Storer
	logger      *logrus.Logger
	authService AuthService
}

// NewVersionControlService 创建版本控制服务
func NewVersionControlService(
	repoPath string,
	logger *logrus.Logger,
	authService AuthService,
) VersionControlService {
	return &VersionControlImpl{
		repoPath: repoPath,
		logger:    logger,
		authService: authService,
	}
}

// InitializeRepository 初始化仓库
func (v *VersionControlImpl) InitializeRepository(ctx context.Context, docID string) error {
	repoDir := v.getRepoPath(docID)

	// 创建目录
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("创建仓库目录失败: %w", err)
	}

	// 检查是否已存在
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err == nil {
		// 仓库已存在
		return nil
	}

	// 初始化Git仓库
	repo, err := git.PlainInit(repoDir, false)
	if err != nil {
		return fmt.Errorf("初始化Git仓库失败: %w", err)
	}

	// 配置仓库
	err = v.configureRepository(repo, docID)
	if err != nil {
		return fmt.Errorf("配置仓库失败: %w", err)
	}

	// 创建初始提交
	err = v.createInitialCommit(repo, docID)
	if err != nil {
		return fmt.Errorf("创建初始提交失败: %w", err)
	}

	v.logger.WithFields(logrus.Fields{
		"document_id": docID,
		"repository": repoDir,
	}).Info("文档版本控制仓库初始化完成")

	return nil
}

// GetRepository 获取仓库
func (v *VersionControlImpl) GetRepository(ctx context.Context, docID string) (*git.Repository, error) {
	repoDir := v.getRepoPath(docID)

	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return nil, fmt.Errorf("打开仓库失败: %w", err)
	}

	return repo, nil
}

// DeleteRepository 删除仓库
func (v *VersionControlImpl) DeleteRepository(ctx context.Context, docID string) error {
	repoDir := v.getRepoPath(docID)

	// 检查权限
	if err := v.checkDeletePermission(ctx, docID); err != nil {
		return err
	}

	// 删除目录
	if err := os.RemoveAll(repoDir); err != nil {
		return fmt.Errorf("删除仓库失败: %w", err)
	}

	v.logger.WithFields(logrus.Fields{
		"document_id": docID,
		"repository": repoDir,
	}).Info("文档版本控制仓库已删除")

	return nil
}

// CreateCommit 创建提交
func (v *VersionControlImpl) CreateCommit(ctx context.Context, req *CommitRequest) (*CommitInfo, error) {
	// 验证权限
	if err := v.checkWritePermission(ctx, req.DocumentID); err != nil {
		return nil, err
	}

	repo, err := v.GetRepository(ctx, req.DocumentID)
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("获取工作区失败: %w", err)
	}

	// 切换到目标分支
	if err := v.checkoutBranch(worktree, req.BranchName, req.ParentCommits); err != nil {
		return nil, fmt.Errorf("切换分支失败: %w", err)
	}

	// 应用文件变更
	for _, fileChange := range req.Files {
		if err := v.applyFileChange(worktree, fileChange); err != nil {
			return nil, fmt.Errorf("应用文件变更失败 [%s]: %w", fileChange.Path, err)
		}
	}

	// 获取当前状态
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("获取状态失败: %w", err)
	}

	if status.IsClean() {
		return nil, fmt.Errorf("没有变更需要提交")
	}

	// 添加变更到暂存区
	for file, fileStatus := range status {
		if fileStatus.Worktree != plumbing.Unmodified {
			_, err := worktree.Add(file)
			if err != nil {
				v.logger.WithFields(logrus.Fields{
					"file": file,
					"error": err,
				}).Warn("添加文件到暂存区失败")
			}
		}
	}

	// 创建提交
	commitHash, err := worktree.Commit(req.Message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  req.AuthorName,
			Email: req.AuthorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建提交失败: %w", err)
	}

	// 获取提交对象
	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("获取提交对象失败: %w", err)
	}

	// 转换为CommitInfo
	commitInfo := v.convertCommitToInfo(commit, req.BranchName, req.Metadata)

	v.logger.WithFields(logrus.Fields{
		"document_id": req.DocumentID,
		"commit_hash": commitHash.String(),
		"branch": req.BranchName,
		"author": req.AuthorName,
		"files_changed": len(req.Files),
	}).Info("创建提交成功")

	return commitInfo, nil
}

// GetCommit 获取提交信息
func (v *VersionControlImpl) GetCommit(ctx context.Context, docID, commitID string) (*CommitInfo, error) {
	if err := v.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return nil, err
	}

	hash := plumbing.NewHash(commitID)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("获取提交失败: %w", err)
	}

	commitInfo := v.convertCommitToInfo(commit, "", nil)

	return commitInfo, nil
}

// GetCommitHistory 获取提交历史
func (v *VersionControlImpl) GetCommitHistory(ctx context.Context, docID string, opts *HistoryOptions) ([]*CommitInfo, error) {
	if err := v.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return nil, err
	}

	// 获取分支引用
	var ref *plumbing.Reference
	if opts != nil && opts.BranchName != "" {
		ref, err = repo.Reference(plumbing.ReferenceName(opts.BranchName), true)
		if err != nil {
			return nil, fmt.Errorf("获取分支引用失败: %w", err)
		}
	} else {
		ref, err = repo.Head()
		if err != nil {
			return nil, fmt.Errorf("获取HEAD引用失败: %w", err)
		}
	}

	// 获取提交历史
	commitIter, err := repo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("获取提交历史失败: %w", err)
	}
	defer commitIter.Close()

	var commits []*CommitInfo
	count := 0
	skip := 0
	if opts != nil {
		skip = opts.Skip
	}

	err = commitIter.ForEach(func(commit *object.Commit) error {
		if opts != nil && skip > 0 {
			skip--
			return nil
		}

		// 检查时间范围
		if opts != nil && opts.Since != nil && commit.Author.When.Before(*opts.Since) {
			return nil
		}
		if opts != nil && opts.Until != nil && commit.Author.When.After(*opts.Until) {
			return nil
		}

		// 检查文件路径
		if opts != nil && opts.Path != "" {
			if !v.isFileInCommit(commit, opts.Path) {
				return nil
			}
		}

		commitInfo := v.convertCommitToInfo(commit, ref.Name().Short(), nil)
		commits = append(commits, commitInfo)
		count++

		// 检查限制
		if opts != nil && opts.Limit > 0 && count >= opts.Limit {
			return fmt.Errorf("达到限制数量")
		}

		return nil
	})

	if err != nil && err.Error() != "达到限制数量" {
		return nil, fmt.Errorf("遍历提交历史失败: %w", err)
	}

	return commits, nil
}

// CompareCommits 比较两个提交
func (v *VersionControlImpl) CompareCommits(ctx context.Context, docID, fromCommit, toCommit string) (*DiffResult, error) {
	if err := v.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return nil, err
	}

	fromHash := plumbing.NewHash(fromCommit)
	toHash := plumbing.NewHash(toCommit)

	fromCommitObj, err := repo.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("获取源提交失败: %w", err)
	}

	toCommitObj, err := repo.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("获取目标提交失败: %w", err)
	}

	// 生成补丁
	patch, err := fromCommitObj.Patch(toCommitObj)
	if err != nil {
		return nil, fmt.Errorf("生成补丁失败: %w", err)
	}

	// 转换为DiffResult
	diffResult := v.convertPatchToDiffResult(patch, fromCommit, toCommit)

	return diffResult, nil
}

// CreateBranch 创建分支
func (v *VersionControlImpl) CreateBranch(ctx context.Context, docID, branchName, baseCommit string) error {
	if err := v.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return err
	}

	// 获取基础提交
	hash := plumbing.NewHash(baseCommit)
	if _, err := repo.CommitObject(hash); err != nil {
		return fmt.Errorf("无效的基础提交: %w", err)
	}

	// 创建分支引用
	branchRef := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewReference(plumbing.ReferenceName(branchRef.String()), hash.String())

	if err := repo.Storer.SetReference(ref); err != nil {
		return fmt.Errorf("创建分支失败: %w", err)
	}

	v.logger.WithFields(logrus.Fields{
		"document_id": docID,
		"branch": branchName,
		"base_commit": baseCommit,
	}).Info("创建分支成功")

	return nil
}

// DeleteBranch 删除分支
func (v *VersionControlImpl) DeleteBranch(ctx context.Context, docID, branchName string) error {
	if err := v.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	// 防止删除默认分支
	if branchName == "main" || branchName == "master" {
		return fmt.Errorf("不能删除默认分支")
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return err
	}

	// 删除分支引用
	branchRef := plumbing.NewBranchReferenceName(branchName)
	if err := repo.Storer.RemoveReference(branchRef); err != nil {
		return fmt.Errorf("删除分支失败: %w", err)
	}

	v.logger.WithFields(logrus.Fields{
		"document_id": docID,
		"branch": branchName,
	}).Info("删除分支成功")

	return nil
}

// ListBranches 列出所有分支
func (v *VersionControlImpl) ListBranches(ctx context.Context, docID string) ([]*BranchInfo, error) {
	if err := v.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return nil, err
	}

	branchIter, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("获取分支列表失败: %w", err)
	}
	defer branchIter.Close()

	var branchInfos []*BranchInfo
	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return nil
		}

		branchInfo := &BranchInfo{
			Name:       ref.Name().Short(),
			CommitHash: ref.Hash().String(),
			CommitTime: commit.Author.When,
			IsDefault: ref.Name().Short() == "main" || ref.Name().Short() == "master",
			Author:     commit.Author.Name,
			Message:    strings.Split(commit.Message, "\n")[0],
		}

		branchInfos = append(branchInfos, branchInfo)
		return nil
	})

	return branchInfos, nil
}

// MergeBranch 合并分支
func (v *VersionControlImpl) MergeBranch(ctx context.Context, docID, sourceBranch, targetBranch string) (*MergeResult, error) {
	if err := v.checkWritePermission(ctx, docID); err != nil {
		return nil, err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("获取工作区失败: %w", err)
	}

	// 切换到目标分支
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(targetBranch),
	})
	if err != nil {
		return nil, fmt.Errorf("切换到目标分支失败: %w", err)
	}

	// 获取源分支引用
	sourceRef, err := repo.Reference(plumbing.NewBranchReferenceName(sourceBranch), true)
	if err != nil {
		return nil, fmt.Errorf("获取源分支引用失败: %w", err)
	}

	// 执行合并
	err = worktree.Merge(sourceRef.Hash(), &git.MergeOptions{
		FastForwardOnly: false,
	})
	if err != nil {
		// 检查是否是冲突
		if strings.Contains(err.Error(), "conflict") {
			conflicts := v.detectConflicts(worktree)
			return &MergeResult{
				Success:   false,
				Conflict:  true,
				Conflicts: conflicts,
				Message:   fmt.Sprintf("合并冲突: %v", err),
			}, nil
		}
		return nil, fmt.Errorf("合并失败: %w", err)
	}

	// 获取合并后的HEAD提交
	headRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("获取HEAD引用失败: %w", err)
	}

	result := &MergeResult{
		Success:    true,
		CommitHash: headRef.Hash().String(),
		MergeType:  "merge",
		Message:    "合并成功",
	}

	v.logger.WithFields(logrus.Fields{
		"document_id":   docID,
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"commit_hash":   result.CommitHash,
	}).Info("分支合并成功")

	return result, nil
}

// Checkout 检出
func (v *VersionControlImpl) Checkout(ctx context.Context, docID string, opts *CheckoutOptions) error {
	if err := v.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作区失败: %w", err)
	}

	var checkoutOpts *git.CheckoutOptions

	if opts != nil && opts.BranchName != "" {
		// 检出分支
		branchName := plumbing.NewBranchReferenceName(opts.BranchName)
		checkoutOpts = &git.CheckoutOptions{
			Branch: branchName,
			Create: opts.CreateBranch,
			Force:  opts.Force,
		}
	} else if opts != nil && opts.CommitHash != "" {
		// 检出提交
		hash := plumbing.NewHash(opts.CommitHash)
		checkoutOpts = &git.CheckoutOptions{
			Hash:  hash,
			Force: opts.Force,
		}
	} else {
		return fmt.Errorf("必须指定分支名或提交哈希")
	}

	err = worktree.Checkout(checkoutOpts)
	if err != nil {
		return fmt.Errorf("检出失败: %w", err)
	}

	v.logger.WithFields(logrus.Fields{
		"document_id": docID,
		"options":    opts,
	}).Info("检出成功")

	return nil
}

// GetFileContent 获取文件内容
func (v *VersionControlImpl) GetFileContent(ctx context.Context, docID, commitID, filePath string) ([]byte, error) {
	if err := v.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return nil, err
	}

	hash := plumbing.NewHash(commitID)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("获取提交失败: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("获取树对象失败: %w", err)
	}

	file, err := tree.FindFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("查找文件失败: %w", err)
	}

	if file == nil {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	content, err := file.Contents()
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	return content, nil
}

// ApplyPatch 应用补丁
func (v *VersionControlImpl) ApplyPatch(ctx context.Context, docID, patch string) error {
	if err := v.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作区失败: %w", err)
	}

	// 这里应该实现补丁应用逻辑
	// 由于go-git的补丁应用比较复杂，暂时返回错误
	return fmt.Errorf("补丁应用功能暂未实现")
}

// RevertToCommit 回滚到指定提交
func (v *VersionControlImpl) RevertToCommit(ctx context.Context, docID, commitID string) error {
	if err := v.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作区失败: %w", err)
	}

	hash := plumbing.NewHash(commitID)
	err = worktree.Reset(&git.ResetOptions{
		Commit: hash,
		Mode:   git.HardReset,
	})
	if err != nil {
		return fmt.Errorf("回滚失败: %w", err)
	}

	v.logger.WithFields(logrus.Fields{
		"document_id": docID,
		"commit_id":  commitID,
	}).Info("回滚到指定提交成功")

	return nil
}

// CherryPick 应用提交
func (v *VersionControlImpl) CherryPick(ctx context.Context, docID, commitID, targetBranch string) error {
	if err := v.checkWritePermission(ctx, docID); err != nil {
		return err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("获取工作区失败: %w", err)
	}

	// 切换到目标分支
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(targetBranch),
	})
	if err != nil {
		return fmt.Errorf("切换到目标分支失败: %w", err)
	}

	// 获取要应用的提交
	hash := plumbing.NewHash(commitID)
	err = worktree.CherryPick(&git.CherryPickOptions{
		Commit: hash,
	})
	if err != nil {
		return fmt.Errorf("应用提交失败: %w", err)
	}

	v.logger.WithFields(logrus.Fields{
		"document_id":   docID,
		"commit_id":    commitID,
		"target_branch": targetBranch,
	}).Info("应用提交成功")

	return nil
}

// GetFileHistory 获取文件历史
func (v *VersionControlImpl) GetFileHistory(ctx context.Context, docID, filePath string) ([]*CommitInfo, error) {
	if err := v.checkReadPermission(ctx, docID); err != nil {
		return nil, err
	}

	repo, err := v.GetRepository(ctx, docID)
	if err != nil {
		return nil, err
	}

	// 获取HEAD引用
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("获取HEAD引用失败: %w", err)
	}

	// 获取文件日志
	commitIter, err := repo.Log(&git.LogOptions{
		From:    ref.Hash(),
		Order:   git.LogOrderCommitterTime,
		PathFilter: func(path string) bool {
			return path == filePath
		},
	})
	if err != nil {
		return nil, fmt.Errorf("获取文件历史失败: %w", err)
	}
	defer commitIter.Close()

	var commits []*CommitInfo
	err = commitIter.ForEach(func(commit *object.Commit) error {
		commitInfo := v.convertCommitToInfo(commit, "", map[string]string{
			"file_path": filePath,
		})
		commits = append(commits, commitInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历文件历史失败: %w", err)
	}

	return commits, nil
}

// 辅助方法

// getRepoPath 获取仓库路径
func (v *VersionControlImpl) getRepoPath(docID string) string {
	return filepath.Join(v.repoPath, v.hashDocumentID(docID))
}

// hashDocumentID 哈希文档ID
func (v *VersionControlImpl) hashDocumentID(docID string) string {
	h := sha1.New()
	h.Write([]byte(docID))
	return hex.EncodeToString(h.Sum(nil))
}

// configureRepository 配置仓库
func (v *VersionControlImpl) configureRepository(repo *git.Repository, docID string) error {
	cfg, err := repo.Config()
	if err != nil {
		return err
	}

	// 设置用户信息
	cfg.User.Name = "Document Service"
	cfg.User.Email = "docservice@lawoa.com"

	// 设置仓库描述
	cfg.Raw.Section("core").SetOption("description", fmt.Sprintf("Document version control for %s", docID))

	return cfg.SetConfig(cfg)
}

// createInitialCommit 创建初始提交
func (v *VersionControlImpl) createInitialCommit(repo *git.Repository, docID string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	// 创建README文件
	readmeContent := fmt.Sprintf(`# Document: %s

This document is version controlled by Law OA Document Service.

Created at: %s
`, docID, time.Now().Format(time.RFC3339))

	// 写入文件
	filePath := "README.md"
	err = os.WriteFile(filepath.Join(worktree.Filesystem.Root(), filePath), []byte(readmeContent), 0644)
	if err != nil {
		return err
	}

	// 添加到暂存区
	_, err = worktree.Add(filePath)
	if err != nil {
		return err
	}

	// 创建提交
	_, err = worktree.Commit(fmt.Sprintf("Initial commit for document %s", docID), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Document Service",
			Email: "docservice@lawoa.com",
			When:  time.Now(),
		},
	})

	return err
}

// checkoutBranch 检出分支
func (v *VersionControlImpl) checkoutBranch(worktree *git.Worktree, branchName string, parentCommits []string) error {
	// 检查分支是否存在
	branchRef := plumbing.NewBranchReferenceName(branchName)
	_, err := worktree.Reference(branchRef)
	if err == nil {
		// 分支存在，直接检出
		return worktree.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
		})
	}

	// 分支不存在，创建新分支
	if len(parentCommits) > 0 {
		// 从指定父提交创建分支
		hash := plumbing.NewHash(parentCommits[0])
		branchRefName := plumbing.NewReference(plumbing.ReferenceName(branchRef.String()), hash.String())
		if err := worktree.Storer.SetReference(branchRefName); err != nil {
			return err
		}
	} else {
		// 从当前HEAD创建分支
		headRef, err := worktree.Reference(plumbing.HEAD)
		if err != nil {
			return err
		}
		branchRefName := plumbing.NewReference(plumbing.ReferenceName(branchRef.String()), headRef.Hash().String())
		if err := worktree.Storer.SetReference(branchRefName); err != nil {
			return err
		}
	}

	// 检出新分支
	return worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Create: true,
	})
}

// applyFileChange 应用文件变更
func (v *VersionControlImpl) applyFileChange(worktree *git.Worktree, fileChange *FileChange) error {
	filePath := filepath.Join(worktree.Filesystem.Root(), fileChange.Path)

	switch fileChange.Operation {
	case "create", "update":
		// 创建或更新文件
		if err := os.WriteFile(filePath, fileChange.Content, 0644); err != nil {
			return err
		}
	case "delete":
		// 删除文件
		if err := os.Remove(filePath); err != nil {
			return err
		}
	default:
		return fmt.Errorf("不支持的操作类型: %s", fileChange.Operation)
	}

	return nil
}

// convertCommitToInfo 转换提交对象为信息结构
func (v *VersionControlImpl) convertCommitToInfo(commit *object.Commit, branch string, metadata map[string]string) *CommitInfo {
	parentHashes := make([]string, len(commit.ParentHashes))
	for i, hash := range commit.ParentHashes {
		parentHashes[i] = hash.String()
	}

	// 获取文件变更
	fileChanges := v.extractFileChanges(commit)

	return &CommitInfo{
		Hash:         commit.Hash.String(),
		AuthorName:   commit.Author.Name,
		AuthorEmail:  commit.Author.Email,
		Message:      commit.Message,
		Timestamp:    commit.Author.When,
		ParentHashes: parentHashes,
		FileChanges:  fileChanges,
		Branch:       branch,
		Metadata:     metadata,
	}
}

// extractFileChanges 提取文件变更
func (v *VersionControlImpl) extractFileChanges(commit *object.Commit) []*FileChange {
	// 这里应该实现文件变更提取逻辑
	// 由于复杂度较高，暂时返回空列表
	return []*FileChange{}
}

// convertPatchToDiffResult 转换补丁为差异结果
func (v *VersionControlImpl) convertPatchToDiffResult(patch *object.Patch, fromCommit, toCommit string) *DiffResult {
	filePatches := patch.FilePatches()
	fileDiffs := make([]*FileDiff, len(filePatches))

	var summary DiffResult

	for i, filePatch := range filePatches {
		var fileDiff FileDiff
		fileDiff.Path = filePatch.File().Path()
		fileDiff.ChangeType = v.getChangeType(filePatch)
		fileDiff.OldPath = filePatch.File().Path()

		// 统计变更
		fileDiff.LinesAdded = filePatch.NumAdditions()
		fileDiff.LinesRemoved = filePatch.NumDeletions()
		summary.LinesAdded += fileDiff.LinesAdded
		summary.LinesRemoved += fileDiff.LinesRemoved

		fileDiffs[i] = &fileDiff
	}

	summary.FilesChanged = len(filePatches)

	return &DiffResult{
		FromCommit: fromCommit,
		ToCommit:   toCommit,
		FileDiffs:  fileDiffs,
		Summary:    &summary,
		Timestamp:  time.Now(),
	}
}

// getChangeType 获取变更类型
func (v *VersionControlImpl) getChangeType(change string) string {
	// 简化的变更类型检测
	if strings.Contains(change, "deleted") {
		return "delete"
	}
	if strings.Contains(change, "added") {
		return "add"
	}
	return "modify"
}

// detectConflicts 检测冲突
func (v *VersionControlImpl) detectConflicts(worktree *git.Worktree) []*Conflict {
	// 这里应该实现冲突检测逻辑
	// 暂时返回空列表
	return []*Conflict{}
}

// isFileInCommit 检查文件是否在提交中
func (v *VersionControlImpl) isFileInCommit(commit *object.Commit, filePath string) bool {
	tree, err := commit.Tree()
	if err != nil {
		return false
	}

	_, err = tree.FindFile(filePath)
	return err == nil
}

// checkReadPermission 检查读权限
func (v *VersionControlImpl) checkReadPermission(ctx context.Context, docID string) error {
	// 这里应该实现权限检查逻辑
	return nil
}

// checkWritePermission 检查写权限
func (v *VersionControlImpl) checkWritePermission(ctx context.Context, docID string) error {
	// 这里应该实现权限检查逻辑
	return nil
}

// checkDeletePermission 检查删除权限
func (v *VersionControlImpl) checkDeletePermission(ctx context.Context, docID string) error {
	// 这里应该实现权限检查逻辑
	return nil
}