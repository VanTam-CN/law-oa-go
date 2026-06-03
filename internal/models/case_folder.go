package models

import "time"

// CaseFolder 案件文件夹（卷宗目录实例）
// 根据模板创建的案件实际文件夹结构
type CaseFolder struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	CaseID       uint   `json:"case_id" gorm:"index:idx_folder_case;not null;comment:关联案件ID"`
	ParentID     *uint  `json:"parent_id,omitempty" gorm:"index:idx_folder_parent;comment:父文件夹ID"`
	Name         string `json:"name" gorm:"size:200;not null;comment:文件夹名称"`
	DisplayOrder int    `json:"display_order" gorm:"default:0;comment:排序序号"`
	Description  string `json:"description" gorm:"size:500;comment:文件夹描述"`

	// 模板关联
	TemplateID   *uint  `json:"template_id,omitempty" gorm:"comment:来源模板ID"`
	TemplatePath string `json:"template_path,omitempty" gorm:"size:500;comment:模板中的路径"`

	// 文件夹内文档数量（查询时通过 LEFT JOIN 统计，不持久化到表结构）
	DocumentCount int `json:"document_count,omitempty" gorm:"->:0;<-:false;column:document_count"`
}

func (CaseFolder) TableName() string {
	return "case_folders"
}

// FolderNode 文件夹树节点（用于返回层级结构）
type FolderNode struct {
	ID           uint          `json:"id"`
	CaseID       uint          `json:"case_id"`
	ParentID     *uint         `json:"parent_id,omitempty"`
	Name         string        `json:"name"`
	DisplayOrder int           `json:"display_order"`
	Description  string        `json:"description"`
	TemplatePath string        `json:"template_path,omitempty"`
	DocumentCount int          `json:"document_count"`
	Children     []*FolderNode `json:"children,omitempty"`
}
