package models

import (
	"fmt"
	"time"
)

// HierarchyItem 层级结构项
type HierarchyItem struct {
	Depth int        `json:"depth"`
	LegalStatute
}

// CategoryStat 分类统计
type CategoryStat struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	Level        int    `json:"level"`
	Description  string `json:"description"`
	StatuteCount int64  `json:"statute_count"`
}

// CategoryTreeNode 分类树节点
type CategoryTreeNode struct {
	ID       int                `json:"id"`
	Name     string             `json:"name"`
	Code     string             `json:"code"`
	Level    int                `json:"level"`
	Children []*CategoryTreeNode `json:"children,omitempty"`
	StatuteCount int            `json:"statute_count"`
}

// LegalStatuteSummary 法条摘要
type LegalStatuteSummary struct {
	ID            int       `json:"id"`
	StatuteNumber string    `json:"statute_number"`
	Title         string    `json:"title"`
	LawName       string    `json:"law_name"`
	CategoryName  string    `json:"category_name"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// LegalSearchRequest 搜索请求
type LegalSearchRequest struct {
	Query             string   `json:"query" form:"query"`
	CategoryID        int      `json:"category_id" form:"category_id"`
	LawName           string   `json:"law_name" form:"law_name"`
	Status            string   `json:"status" form:"status"`
	EffectiveFrom     string   `json:"effective_from" form:"effective_from"`
	EffectiveTo       string   `json:"effective_to" form:"effective_to"`
	Tags              []string `json:"tags" form:"tags"`
	IncludeInactive   bool     `json:"include_inactive" form:"include_inactive"`
	SortBy            string   `json:"sort_by" form:"sort_by"`         // relevance, date, title
	SortOrder         string   `json:"sort_order" form:"sort_order"`   // asc, desc
	Page              int      `json:"page" form:"page"`
	PageSize          int      `json:"page_size" form:"page_size"`
}

// LegalSearchResponse 搜索响应
type LegalSearchResponse struct {
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
	Statutes   []*LegalStatute      `json:"statutes"`
	Categories []*CategoryStat      `json:"categories,omitempty"`
	Suggestions []string            `json:"suggestions,omitempty"`
	SearchTime  int                 `json:"search_time_ms"`
}

// LegalCategoryRequest 分类请求
type LegalCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	ParentID    *int   `json:"parent_id"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// LegalCategoryResponse 分类响应
type LegalCategoryResponse struct {
	ID           int                   `json:"id"`
	Name         string                `json:"name"`
	Code         string                `json:"code"`
	ParentID     *int                  `json:"parent_id"`
	Level        int                   `json:"level"`
	Description  string                `json:"description"`
	IsActive     bool                  `json:"is_active"`
	StatuteCount int                   `json:"statute_count"`
	Children     []*LegalCategoryResponse `json:"children,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// LegalTagRequest 标签请求
type LegalTagRequest struct {
	Name        string `json:"name" binding:"required"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// LegalTagResponse 标签响应
type LegalTagResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Description string    `json:"description"`
	UsageCount  int       `json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// LegalStatuteCreateRequest 创建法条请求
type LegalStatuteCreateRequest struct {
	StatuteNumber      string     `json:"statute_number" binding:"required"`
	Title              string     `json:"title" binding:"required"`
	Content            string     `json:"content" binding:"required"`
	CategoryID         int        `json:"category_id" binding:"required"`
	LawName            string     `json:"law_name" binding:"required"`
	Chapter            string     `json:"chapter"`
	Section            string     `json:"section"`
	Part               string     `json:"part"`
	EffectiveDate      *time.Time `json:"effective_date"`
	ExpiryDate         *time.Time `json:"expiry_date"`
	PublishingAuthority string     `json:"publishing_authority"`
	Status             string     `json:"status"`
	HierarchyLevel     int        `json:"hierarchy_level"`
	ParentStatuteID    *int       `json:"parent_statute_id"`
	OrderInHierarchy   *int       `json:"order_in_hierarchy"`
	Tags               []string   `json:"tags"`
	Keywords           []string   `json:"keywords"`
}

// LegalStatuteUpdateRequest 更新法条请求
type LegalStatuteUpdateRequest struct {
	Title              string     `json:"title"`
	Content            string     `json:"content"`
	CategoryID         int        `json:"category_id"`
	Chapter            string     `json:"chapter"`
	Section            string     `json:"section"`
	Part               string     `json:"part"`
	EffectiveDate      *time.Time `json:"effective_date"`
	ExpiryDate         *time.Time `json:"expiry_date"`
	PublishingAuthority string     `json:"publishing_authority"`
	Status             string     `json:"status"`
	OrderInHierarchy   *int       `json:"order_in_hierarchy"`
	Tags               []string   `json:"tags"`
	Keywords           []string   `json:"keywords"`
	ChangeDescription  string     `json:"change_description"` // 用于版本记录
}

// LegalStatuteResponse 法条响应
type LegalStatuteResponse struct {
	ID                 int                        `json:"id"`
	StatuteNumber      string                     `json:"statute_number"`
	Title              string                     `json:"title"`
	Content            string                     `json:"content"`
	Category           *LegalCategoryResponse     `json:"category"`
	LawName            string                     `json:"law_name"`
	Chapter            string                     `json:"chapter"`
	Section            string                     `json:"section"`
	Part               string                     `json:"part"`
	EffectiveDate      *time.Time                 `json:"effective_date"`
	ExpiryDate         *time.Time                 `json:"expiry_date"`
	PublishingAuthority string                     `json:"publishing_authority"`
	Status             string                     `json:"status"`
	HierarchyLevel     int                        `json:"hierarchy_level"`
	ParentStatuteID    *int                       `json:"parent_statute_id"`
	OrderInHierarchy   *int                       `json:"order_in_hierarchy"`
	Tags               []string                   `json:"tags"`
	Keywords           []string                   `json:"keywords"`
	IsFavorited        bool                       `json:"is_favorited"`
	ViewCount          int                        `json:"view_count"`
	FavoriteCount      int                        `json:"favorite_count"`
	CreatedAt          time.Time                  `json:"created_at"`
	UpdatedAt          time.Time                  `json:"updated_at"`

	// 扩展字段
	FullPath           string                     `json:"full_path"`
	IsActive           bool                       `json:"is_active"`
	ChildStatutes      []*LegalStatuteResponse    `json:"child_statutes,omitempty"`
	Versions           []*LegalStatuteVersion     `json:"versions,omitempty"`
}

// FavoriteRequest 收藏请求
type FavoriteRequest struct {
	StatuteID int `json:"statute_id" binding:"required"`
}

// FavoriteResponse 收藏响应
type FavoriteResponse struct {
	ID        int       `json:"id"`
	StatuteID int       `json:"statute_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Statute   *LegalStatuteResponse `json:"statute,omitempty"`
}

// SearchHistoryResponse 搜索历史响应
type SearchHistoryResponse struct {
	ID            int                    `json:"id"`
	UserID        *int                   `json:"user_id"`
	SearchQuery   string                 `json:"search_query"`
	SearchFilters map[string]interface{} `json:"search_filters"`
	ResultCount   int                    `json:"result_count"`
	SearchDuration *int                  `json:"search_duration"`
	CreatedAt     time.Time              `json:"created_at"`
}

// LegalStatuteBatchRequest 批量操作请求
type LegalStatuteBatchRequest struct {
	StatuteIDs []int `json:"statute_ids" binding:"required"`
}

// LegalStatuteBatchResponse 批量操作响应
type LegalStatuteBatchResponse struct {
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	SuccessIDs   []int    `json:"success_ids"`
	FailureIDs   []int    `json:"failure_ids"`
	Errors       []string `json:"errors,omitempty"`
}

// LegalExportRequest 导出请求
type LegalExportRequest struct {
	Format      string   `json:"format"`                // pdf, word, excel, json
	CategoryID  int      `json:"category_id"`
	LawName     string   `json:"law_name"`
	Status      string   `json:"status"`
	Tags        []string `json:"tags"`
	FromDate    string   `json:"from_date"`
	ToDate      string   `json:"to_date"`
	IncludeContent bool  `json:"include_content"`
}

// LegalExportResponse 导出响应
type LegalExportResponse struct {
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ExportTime  time.Time `json:"export_time"`
}

// 验证方法
func (req *LegalSearchRequest) Validate() error {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "relevance"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}
	return nil
}

func (req *LegalStatuteCreateRequest) Validate() error {
	if req.StatuteNumber == "" {
		return fmt.Errorf("法条编号不能为空")
	}
	if req.Title == "" {
		return fmt.Errorf("法条标题不能为空")
	}
	if req.Content == "" {
		return fmt.Errorf("法条内容不能为空")
	}
	if req.LawName == "" {
		return fmt.Errorf("法律名称不能为空")
	}
	if req.CategoryID <= 0 {
		return fmt.Errorf("分类ID无效")
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.HierarchyLevel <= 0 {
		req.HierarchyLevel = 1
	}
	return nil
}

func (req *LegalStatuteUpdateRequest) Validate() error {
	if req.Status == "" {
		req.Status = "active"
	}
	return nil
}

func (req *LegalCategoryRequest) Validate() error {
	if req.Name == "" {
		return fmt.Errorf("分类名称不能为空")
	}
	if req.Code == "" {
		return fmt.Errorf("分类代码不能为空")
	}
	return nil
}

func (req *LegalTagRequest) Validate() error {
	if req.Name == "" {
		return fmt.Errorf("标签名称不能为空")
	}
	if req.Color == "" {
		req.Color = "#1890ff"
	}
	return nil
}