package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// LegalStatute 法条基本信息模型
type LegalStatute struct {
	ID                 int       `json:"id" gorm:"primaryKey;column:id"`
	StatuteNumber      string    `json:"statute_number" gorm:"uniqueIndex;not null;column:statute_number;size:100"`
	Title              string    `json:"title" gorm:"not null;column:title;type:text"`
	Content            string    `json:"content" gorm:"not null;column:content;type:text"`
	CategoryID         int       `json:"category_id" gorm:"column:category_id;index"`
	LawName            string    `json:"law_name" gorm:"not null;column:law_name;size:200"`
	Chapter            string    `json:"chapter" gorm:"column:chapter;size:200"`
	Section            string    `json:"section" gorm:"column:section;size:200"`
	Part               string    `json:"part" gorm:"column:part;size:200"`
	EffectiveDate      *time.Time `json:"effective_date" gorm:"column:effective_date;index"`
	ExpiryDate         *time.Time `json:"expiry_date" gorm:"column:expiry_date"`
	PublishingAuthority string    `json:"publishing_authority" gorm:"column:publishing_authority;size:200"`
	Status             string    `json:"status" gorm:"column:status;default:active;size:20;index"`
	HierarchyLevel     int       `json:"hierarchy_level" gorm:"column:hierarchy_level;default:1;index"`
	ParentStatuteID    *int      `json:"parent_statute_id" gorm:"column:parent_statute_id;index"`
	OrderInHierarchy   *int      `json:"order_in_hierarchy" gorm:"column:order_in_hierarchy"`
	Tags               StringArray `json:"tags" gorm:"column:tags;type:text[]"`
	Keywords           StringArray `json:"keywords" gorm:"column:keywords;type:text[]"`
	CreatedAt          time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	// 关联字段
	Category     *LegalCategory      `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	ParentStatute *LegalStatute     `json:"parent_statute,omitempty" gorm:"foreignKey:ParentStatuteID"`
	ChildStatutes []LegalStatute    `json:"child_statutes,omitempty" gorm:"foreignKey:ParentStatuteID"`
	Versions     []LegalStatuteVersion `json:"versions,omitempty" gorm:"foreignKey:StatuteID"`
	TagsRelation []LegalStatuteTag  `json:"tags_relation,omitempty" gorm:"foreignKey:StatuteID"`
}

// LegalCategory 法条分类模型
type LegalCategory struct {
	ID          int       `json:"id" gorm:"primaryKey;column:id"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null;column:name;size:100"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null;column:code;size:50"`
	ParentID    *int      `json:"parent_id" gorm:"column:parent_id;index"`
	Level       int       `json:"level" gorm:"column:level;default:1;index"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	// 关联字段
	Parent   *LegalCategory   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []LegalCategory  `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Statutes []LegalStatute   `json:"statutes,omitempty" gorm:"foreignKey:CategoryID"`
}

// LegalHierarchy 法条层级关系模型
type LegalHierarchy struct {
	ID           int       `json:"id" gorm:"primaryKey;column:id"`
	AncestorID   int       `json:"ancestor_id" gorm:"not null;column:ancestor_id;index"`
	DescendantID int       `json:"descendant_id" gorm:"not null;column:descendant_id;index"`
	Depth        int       `json:"depth" gorm:"not null;column:depth;index"`
	Path         string    `json:"path" gorm:"column:path;type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`

	// 关联字段
	Ancestor   *LegalStatute `json:"ancestor,omitempty" gorm:"foreignKey:AncestorID"`
	Descendant *LegalStatute `json:"descendant,omitempty" gorm:"foreignKey:DescendantID"`
}

// LegalStatuteVersion 法条版本历史模型
type LegalStatuteVersion struct {
	ID                int       `json:"id" gorm:"primaryKey;column:id"`
	StatuteID         int       `json:"statute_id" gorm:"not null;column:statute_id;index"`
	VersionNumber     int       `json:"version_number" gorm:"not null;column:version_number"`
	Title             string    `json:"title" gorm:"not null;column:title;type:text"`
	Content           string    `json:"content" gorm:"not null;column:content;type:text"`
	EffectiveDate     *time.Time `json:"effective_date" gorm:"column:effective_date"`
	ExpiryDate        *time.Time `json:"expiry_date" gorm:"column:expiry_date"`
	ChangeDescription string    `json:"change_description" gorm:"column:change_description;type:text"`
	CreatedBy         int       `json:"created_by" gorm:"column:created_by;index"`
	CreatedAt         time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`

	// 关联字段
	Statute *LegalStatute `json:"statute,omitempty" gorm:"foreignKey:StatuteID"`
	Creator *User         `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// UserLegalFavorite 用户法条收藏模型
type UserLegalFavorite struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id"`
	UserID    int       `json:"user_id" gorm:"not null;column:user_id;index"`
	StatuteID int       `json:"statute_id" gorm:"not null;column:statute_id;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`

	// 关联字段
	User    *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Statute *LegalStatute `json:"statute,omitempty" gorm:"foreignKey:StatuteID"`
}

// LegalSearchHistory 法条搜索历史模型
type LegalSearchHistory struct {
	ID            int       `json:"id" gorm:"primaryKey;column:id"`
	UserID        *int      `json:"user_id" gorm:"column:user_id;index"`
	SearchQuery   string    `json:"search_query" gorm:"not null;column:search_query;type:text"`
	SearchFilters JSON      `json:"search_filters" gorm:"column:search_filters;type:jsonb"`
	ResultCount   int       `json:"result_count" gorm:"column:result_count;default:0"`
	SearchDuration *int     `json:"search_duration" gorm:"column:search_duration"` // 毫秒
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime;index"`

	// 关联字段
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// LegalTag 法条标签模型
type LegalTag struct {
	ID          int    `json:"id" gorm:"primaryKey;column:id"`
	Name        string `json:"name" gorm:"uniqueIndex;not null;column:name;size:50"`
	Color       string `json:"color" gorm:"column:color;size:7;default:#1890ff"`
	Description string `json:"description" gorm:"column:description;type:text"`
	UsageCount  int    `json:"usage_count" gorm:"column:usage_count;default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`

	// 关联字段
	StatutesRelation []LegalStatuteTag `json:"statutes_relation,omitempty" gorm:"foreignKey:TagID"`
}

// LegalStatuteTag 法条标签关联模型
type LegalStatuteTag struct {
	ID        int       `json:"id" gorm:"primaryKey;column:id"`
	StatuteID int       `json:"statute_id" gorm:"not null;column:statute_id;index"`
	TagID     int       `json:"tag_id" gorm:"not null;column:tag_id;index"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`

	// 关联字段
	Statute *LegalStatute `json:"statute,omitempty" gorm:"foreignKey:StatuteID"`
	Tag     *LegalTag     `json:"tag,omitempty" gorm:"foreignKey:TagID"`
}

// StringArray 自定义类型，用于处理PostgreSQL的text[]类型
type StringArray []string

// Value 实现driver.Valuer接口
func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return "{}", nil
	}
	return json.Marshal(sa)
}

// Scan 实现sql.Scanner接口
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = StringArray{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// PostgreSQL数组通常以 {tag1,tag2}格式返回，需要转换为JSON
		str := string(v)
		if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
			// PostgreSQL数组格式：{tag1,tag2}
			str = strings.Trim(str, "{}")
			if str == "" {
				*sa = StringArray{}
				return nil
			}
			tags := strings.Split(str, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
			*sa = StringArray(tags)
		} else {
			// 尝试JSON解析
			return json.Unmarshal(v, sa)
		}
	case string:
		// PostgreSQL数组格式：{tag1,tag2}
		str := v
		if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
			str = strings.Trim(str, "{}")
			if str == "" {
				*sa = StringArray{}
				return nil
			}
			tags := strings.Split(str, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
			*sa = StringArray(tags)
		} else {
			// 尝试JSON解析
			return json.Unmarshal([]byte(v), sa)
		}
	default:
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
	return nil
}

// TableName 指定表名
func (LegalStatute) TableName() string {
	return "legal_statutes"
}

func (LegalCategory) TableName() string {
	return "legal_categories"
}

func (LegalHierarchy) TableName() string {
	return "legal_hierarchy"
}

func (LegalStatuteVersion) TableName() string {
	return "legal_statute_versions"
}

func (UserLegalFavorite) TableName() string {
	return "user_legal_favorites"
}

func (LegalSearchHistory) TableName() string {
	return "legal_search_history"
}

func (LegalTag) TableName() string {
	return "legal_tags"
}

func (LegalStatuteTag) TableName() string {
	return "legal_statute_tags"
}

// BeforeCreate GORM钩子 - 创建前
func (ls *LegalStatute) BeforeCreate(tx *gorm.DB) error {
	if ls.Status == "" {
		ls.Status = "active"
	}
	if ls.HierarchyLevel == 0 {
		ls.HierarchyLevel = 1
	}
	return nil
}

// BeforeUpdate GORM钩子 - 更新前
func (ls *LegalStatute) BeforeUpdate(tx *gorm.DB) error {
	// 如果内容发生变化，创建版本记录
	if tx.Statement.Changed("Content") || tx.Statement.Changed("Title") {
		// 这里可以添加版本记录逻辑
	}
	return nil
}

// GetFullPath 获取法条的完整路径
func (ls *LegalStatute) GetFullPath() string {
	path := ls.StatuteNumber
	if ls.Part != "" {
		path = ls.Part + " > " + path
	}
	if ls.Section != "" {
		path = ls.Section + " > " + path
	}
	if ls.Chapter != "" {
		path = ls.Chapter + " > " + path
	}
	return path
}

// IsActive 检查法条是否生效
func (ls *LegalStatute) IsActive() bool {
	now := time.Now()

	// 检查状态
	if ls.Status != "active" {
		return false
	}

	// 检查生效日期
	if ls.EffectiveDate != nil && now.Before(*ls.EffectiveDate) {
		return false
	}

	// 检查失效日期
	if ls.ExpiryDate != nil && now.After(*ls.ExpiryDate) {
		return false
	}

	return true
}

// AddTag 添加标签
func (ls *LegalStatute) AddTag(tag string) {
	for _, existingTag := range ls.Tags {
		if existingTag == tag {
			return // 标签已存在
		}
	}
	ls.Tags = append(ls.Tags, tag)
}

// RemoveTag 移除标签
func (ls *LegalStatute) RemoveTag(tag string) {
	for i, existingTag := range ls.Tags {
		if existingTag == tag {
			ls.Tags = append(ls.Tags[:i], ls.Tags[i+1:]...)
			return
		}
	}
}

// HasTag 检查是否包含指定标签
func (ls *LegalStatute) HasTag(tag string) bool {
	for _, existingTag := range ls.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// GetCategoryPath 获取分类路径
func (lc *LegalCategory) GetCategoryPath() []string {
	path := []string{lc.Name}
	current := lc.Parent

	for current != nil {
		path = append([]string{current.Name}, path...)
		// 这里需要额外查询来获取父分类信息
		// 在实际使用中可以通过预加载或递归查询实现
		break // 避免无限循环
	}

	return path
}

// GetFullCode 获取完整分类代码
func (lc *LegalCategory) GetFullCode() string {
	if lc.ParentID == nil {
		return lc.Code
	}
	// 这里需要查询父分类来构建完整代码
	return lc.Code
}