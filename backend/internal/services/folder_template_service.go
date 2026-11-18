package services

import (
	"context"
	"errors"
	"fmt"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

var (
	ErrFolderNotFound     = errors.New("folder not found")
	ErrFolderCaseMismatch = errors.New("folder does not belong to case")
)

// FolderTemplateService 卷宗目录模板服务接口
type FolderTemplateService interface {
	CreateTemplate(ctx context.Context, req *CreateFolderTemplateRequest) (*models.CaseFolderTemplate, error)
	GetTemplate(ctx context.Context, id uint) (*models.CaseFolderTemplate, error)
	UpdateTemplate(ctx context.Context, id uint, req *UpdateFolderTemplateRequest) error
	DeleteTemplate(ctx context.Context, id uint) error
	ListTemplates(ctx context.Context, params *repositories.FolderTemplateListParams) ([]*models.CaseFolderTemplate, int64, error)
	ApplyTemplate(ctx context.Context, caseID uint, templateID uint) ([]*models.FolderNode, error)
	GetCaseFolders(ctx context.Context, caseID uint) ([]*models.FolderNode, error)
	CreateCustomFolder(ctx context.Context, req *CreateCaseFolderRequest) (*models.CaseFolder, error)
	DeleteCaseFolder(ctx context.Context, caseID, folderID uint) error
}

// CreateFolderTemplateRequest 创建模板请求
type CreateFolderTemplateRequest struct {
	Name            string                 `json:"name" binding:"required"`
	Description     string                 `json:"description"`
	FolderStructure map[string]interface{} `json:"folder_structure" binding:"required"`
	CaseType        string                 `json:"case_type"`
	IsDefault       bool                   `json:"is_default"`
	TemplateFiles   map[string]interface{} `json:"template_files"`
	CreatedBy       uint                   `json:"created_by" binding:"required"`
}

// UpdateFolderTemplateRequest 更新模板请求
type UpdateFolderTemplateRequest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	FolderStructure map[string]interface{} `json:"folder_structure"`
	CaseType        string                 `json:"case_type"`
	IsDefault       *bool                  `json:"is_default"`
	IsActive        *bool                  `json:"is_active"`
	TemplateFiles   map[string]interface{} `json:"template_files"`
}

// CreateCaseFolderRequest 创建自定义文件夹请求
type CreateCaseFolderRequest struct {
	CaseID       uint   `json:"case_id" binding:"required"`
	ParentID     *uint  `json:"parent_id"`
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
}

type folderTemplateService struct {
	repo repositories.FolderTemplateRepository
}

// NewFolderTemplateService 创建卷宗目录模板服务
func NewFolderTemplateService(repo repositories.FolderTemplateRepository) FolderTemplateService {
	return &folderTemplateService{repo: repo}
}

// CreateTemplate 创建模板
func (s *folderTemplateService) CreateTemplate(ctx context.Context, req *CreateFolderTemplateRequest) (*models.CaseFolderTemplate, error) {
	template := &models.CaseFolderTemplate{
		Name:            req.Name,
		Description:     req.Description,
		FolderStructure: models.JSON(req.FolderStructure),
		CaseType:        req.CaseType,
		IsDefault:       req.IsDefault,
		IsActive:        true,
		TemplateFiles:   models.JSON(req.TemplateFiles),
		CreatedBy:       req.CreatedBy,
	}

	if err := s.repo.CreateTemplate(ctx, template); err != nil {
		return nil, err
	}

	// 如果设为默认，取消其他同类型模板的默认标记（在同一事务中）
	if req.IsDefault {
		if err := s.clearOtherDefaults(ctx, template.ID, template.CaseType); err != nil {
			// 回滚：删除刚创建的模板
			s.repo.DeleteTemplate(ctx, template.ID)
			return nil, fmt.Errorf("设置默认模板失败: %w", err)
		}
	}

	return template, nil
}

// GetTemplate 获取模板详情
func (s *folderTemplateService) GetTemplate(ctx context.Context, id uint) (*models.CaseFolderTemplate, error) {
	return s.repo.GetTemplateByID(ctx, id)
}

// UpdateTemplate 更新模板
func (s *folderTemplateService) UpdateTemplate(ctx context.Context, id uint, req *UpdateFolderTemplateRequest) error {
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.FolderStructure != nil {
		updates["folder_structure"] = models.JSON(req.FolderStructure)
	}
	if req.CaseType != "" {
		updates["case_type"] = req.CaseType
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
		if *req.IsDefault {
			// 获取模板的 case_type
			t, err := s.repo.GetTemplateByID(ctx, id)
			if err == nil && t != nil {
				_ = s.clearOtherDefaults(ctx, id, t.CaseType)
			}
		}
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.TemplateFiles != nil {
		updates["template_files"] = models.JSON(req.TemplateFiles)
	}

	return s.repo.UpdateTemplate(ctx, id, updates)
}

// DeleteTemplate 删除模板
func (s *folderTemplateService) DeleteTemplate(ctx context.Context, id uint) error {
	return s.repo.DeleteTemplate(ctx, id)
}

// ListTemplates 列表查询模板
func (s *folderTemplateService) ListTemplates(ctx context.Context, params *repositories.FolderTemplateListParams) ([]*models.CaseFolderTemplate, int64, error) {
	return s.repo.ListTemplates(ctx, params)
}

// ApplyTemplate 将模板应用到案件，创建实际文件夹结构
func (s *folderTemplateService) ApplyTemplate(ctx context.Context, caseID uint, templateID uint) ([]*models.FolderNode, error) {
	template, err := s.repo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("获取模板失败: %w", err)
	}
	if template == nil {
		return nil, fmt.Errorf("模板不存在")
	}

	// 解析文件夹结构
	structure := template.FolderStructure
	foldersRaw, ok := structure["folders"]
	if !ok {
		return nil, fmt.Errorf("模板中缺少 folders 字段")
	}

	foldersList, ok := foldersRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("folders 字段格式错误")
	}

	// 递归创建文件夹，收集错误
	var allFolders []*models.CaseFolder
	var createErrors []error
	for i, f := range foldersList {
		folderMap, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		created, errs := s.createFolderFromTemplate(ctx, caseID, nil, folderMap, i, "", templateID)
		allFolders = append(allFolders, created...)
		createErrors = append(createErrors, errs...)
	}

	// 如果有关键错误，返回失败
	if len(createErrors) > 0 {
		return nil, fmt.Errorf("创建文件夹时发生 %d 个错误，首个: %w", len(createErrors), createErrors[0])
	}

	// 构建树形结构
	return buildFolderTree(ctx, allFolders), nil
}

// GetCaseFolders 获取案件的文件夹树
func (s *folderTemplateService) GetCaseFolders(ctx context.Context, caseID uint) ([]*models.FolderNode, error) {
	folders, err := s.repo.GetFoldersByCaseID(ctx, caseID)
	if err != nil {
		return nil, err
	}
	return buildFolderTree(ctx, folders), nil
}

// CreateCustomFolder 创建自定义文件夹
func (s *folderTemplateService) CreateCustomFolder(ctx context.Context, req *CreateCaseFolderRequest) (*models.CaseFolder, error) {
	if req == nil || req.CaseID == 0 {
		return nil, fmt.Errorf("案件ID不能为空")
	}
	if req.ParentID != nil && *req.ParentID > 0 {
		parent, err := s.repo.GetFolderByID(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrFolderNotFound
		}
		if parent.CaseID != req.CaseID {
			return nil, ErrFolderCaseMismatch
		}
	}
	folder := &models.CaseFolder{
		CaseID:       req.CaseID,
		ParentID:     req.ParentID,
		Name:         req.Name,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
	}

	if err := s.repo.CreateFolder(ctx, folder); err != nil {
		return nil, err
	}
	return folder, nil
}

// DeleteCaseFolder 删除案件文件夹（含子文件夹）。案件ID和文件夹ID
// 必须同时匹配，避免仅凭可猜测的 folder_id 删除其他案件的目录。
func (s *folderTemplateService) DeleteCaseFolder(ctx context.Context, caseID, folderID uint) error {
	if caseID == 0 || folderID == 0 {
		return fmt.Errorf("案件ID和文件夹ID不能为空")
	}
	folder, err := s.repo.GetFolderByID(ctx, folderID)
	if err != nil {
		return err
	}
	if folder == nil {
		return ErrFolderNotFound
	}
	if folder.CaseID != caseID {
		return ErrFolderCaseMismatch
	}
	return s.repo.DeleteFolder(ctx, folderID)
}

// createFolderFromTemplate 递归创建文件夹
// 返回创建的文件夹列表和遇到的错误列表
func (s *folderTemplateService) createFolderFromTemplate(ctx context.Context, caseID uint, parentID *uint, folderMap map[string]interface{}, order int, path string, templateID uint) ([]*models.CaseFolder, []error) {
	var result []*models.CaseFolder
	var errors []error

	name, _ := folderMap["name"].(string)
	description, _ := folderMap["description"].(string)
	currentPath := path + "/" + name

	folder := &models.CaseFolder{
		CaseID:       caseID,
		ParentID:     parentID,
		Name:         name,
		DisplayOrder: order,
		Description:  description,
		TemplateID:   &templateID,
		TemplatePath: currentPath,
	}

	if err := s.repo.CreateFolder(ctx, folder); err != nil {
		errors = append(errors, fmt.Errorf("创建文件夹失败 (%s): %w", name, err))
		return result, errors
	}

	result = append(result, folder)

	// 处理子文件夹
	if subfolders, ok := folderMap["subfolders"].([]interface{}); ok {
		for i, sub := range subfolders {
			if subMap, ok := sub.(map[string]interface{}); ok {
				childFolders, childErrs := s.createFolderFromTemplate(ctx, caseID, &folder.ID, subMap, i, currentPath, templateID)
				result = append(result, childFolders...)
				errors = append(errors, childErrs...)
			}
		}
	}

	return result, errors
}

// clearOtherDefaults 清除同类型其他模板的默认标记
func (s *folderTemplateService) clearOtherDefaults(ctx context.Context, currentID uint, caseType string) error {
	return s.repo.ClearOtherDefaults(ctx, caseType, currentID)
}

// buildFolderTree 将扁平文件夹列表构建为树形结构
func buildFolderTree(ctx context.Context, folders []*models.CaseFolder) []*models.FolderNode {
	nodeMap := make(map[uint]*models.FolderNode)
	var roots []*models.FolderNode

	// 创建所有节点
	for _, f := range folders {
		node := &models.FolderNode{
			ID:            f.ID,
			CaseID:        f.CaseID,
			ParentID:      f.ParentID,
			Name:          f.Name,
			DisplayOrder:  f.DisplayOrder,
			Description:   f.Description,
			TemplatePath:  f.TemplatePath,
			DocumentCount: f.DocumentCount,
			Children:      []*models.FolderNode{},
		}
		nodeMap[f.ID] = node
	}

	// 构建树
	for _, f := range folders {
		node := nodeMap[f.ID]
		if f.ParentID == nil {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[*f.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			// 父节点不在列表中，作为根节点
			roots = append(roots, node)
		}
	}

	return roots
}
