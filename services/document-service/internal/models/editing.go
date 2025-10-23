package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EditSession 编辑会话模型
type EditSession struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID    uuid.UUID  `json:"document_id" gorm:"type:uuid;not null;index"`
	SessionToken  string     `json:"session_token" gorm:"size:255;uniqueIndex;not null"`
	UserID        uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	EditorType    string     `json:"editor_type" gorm:"size:50;not null;check:editor_type IN ('rich-text', 'code', 'markdown')"`
	CursorPosition *Position `json:"cursor_position" gorm:"type:jsonb"`
	SelectionRange *Range    `json:"selection_range" gorm:"type:jsonb"`
	ActivityAt    time.Time  `json:"activity_at" gorm:"default:now()"`
	CreatedAt     time.Time  `json:"created_at" gorm:"default:now()"`
	ExpiresAt     time.Time  `json:"expires_at" gorm:"index"`

	// 关联关系
	Document *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Position 光标位置
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Range 选择范围
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// EditOperation 编辑操作模型
type EditOperation struct {
	ID             uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID     uuid.UUID          `json:"document_id" gorm:"type:uuid;not null;index"`
	UserID         uuid.UUID          `json:"user_id" gorm:"type:uuid;not null;index"`
	SessionID      uuid.UUID          `json:"session_id" gorm:"type:uuid;not null;index"`
	OperationType  string             `json:"operation_type" gorm:"size:50;not null;check:operation_type IN ('insert', 'delete', 'retain', 'format', 'cursor', 'selection')"`
	OperationData  *OperationData     `json:"operation_data" gorm:"type:jsonb;not null"`
	YjsStateVector map[string]uint64  `json:"yjs_state_vector" gorm:"type:jsonb"`
	Timestamp      time.Time          `json:"timestamp" gorm:"default:now();index"`
	Applied        bool               `json:"applied" gorm:"default:false;index"`

	// 关联关系
	Document *Document    `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	User     *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Session  *EditSession `json:"session,omitempty" gorm:"foreignKey:SessionID"`
}

// OperationData 操作数据
type OperationData struct {
	Type       string                 `json:"type"`       // insert, delete, retain, format
	Position   int                    `json:"position"`   // 操作位置
	Content    string                 `json:"content"`    // 插入内容
	Length     int                    `json:"length"`     // 删除长度
	Attributes map[string]interface{} `json:"attributes"` // 格式属性
	Origin     string                 `json:"origin"`     // 操作来源
	Author     string                 `json:"author"`     // 操作者
}

// CollaborationSession 协作会话模型
type CollaborationSession struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID   uuid.UUID `json:"document_id" gorm:"type:uuid;not null;uniqueIndex"`
	RoomName     string    `json:"room_name" gorm:"size:255;uniqueIndex;not null"`
	ActiveUsers  int       `json:"active_users" gorm:"default:0"`
	MaxUsers     int       `json:"max_users" gorm:"default:10"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:now()"`
	LastActivity time.Time `json:"last_activity" gorm:"default:now();index"`

	// 关联关系
	Document *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
}

// DocumentVersion 文档版本模型
type DocumentVersion struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID      uuid.UUID  `json:"document_id" gorm:"type:uuid;not null;index"`
	VersionNumber   int        `json:"version_number" gorm:"not null"`
	Title           string     `json:"title" gorm:"size:255"`
	ContentHash     string     `json:"content_hash" gorm:"size:64;index"`
	ContentDelta    *DeltaData `json:"content_delta" gorm:"type:jsonb"`
	SnapshotPath    string     `json:"snapshot_path" gorm:"size:500"`
	EditorID        *uuid.UUID `json:"editor_id" gorm:"type:uuid;index"`
	EditSummary     string     `json:"edit_summary" gorm:"type:text"`
	IsMajorVersion  bool       `json:"is_major_version" gorm:"default:false"`
	IsPublished     bool       `json:"is_published" gorm:"default:false"`
	FileSize        int64      `json:"file_size"`
	CharacterCount  int        `json:"character_count"`
	CreatedAt       time.Time  `json:"created_at" gorm:"default:now();index"`

	// 关联关系
	Document *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	Editor   *User     `json:"editor,omitempty" gorm:"foreignKey:EditorID"`
}

// DeltaData Yjs增量数据
type DeltaData struct {
	Ops       []interface{} `json:"ops"`
	ClientID  uint64        `json:"client_id"`
	Clock     uint64        `json:"clock"`
	Origin    string        `json:"origin"`
	State     map[string]uint64 `json:"state"`
}

// CollaborationParticipant 协作参与者
type CollaborationParticipant struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID   uuid.UUID  `json:"session_id" gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	SocketID    string     `json:"socket_id" gorm:"size:255;not null"`
	UserName    string     `json:"user_name" gorm:"size:100;not null"`
	UserAvatar  string     `json:"user_avatar" gorm:"size:500"`
	UserColor   string     `json:"user_color" gorm:"size:7;default:#2196F3"`
	Status      string     `json:"status" gorm:"size:20;default:online;check:status IN ('online', 'away', 'busy')"`
	LastSeen    time.Time  `json:"last_seen" gorm:"default:now()"`
	JoinedAt    time.Time  `json:"joined_at" gorm:"default:now()"`
	Permissions []string   `json:"permissions" gorm:"type:text[]"`

	// 关联关系
	Session *CollaborationSession `json:"session,omitempty" gorm:"foreignKey:SessionID"`
	User    *User                `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// EditorConfig 编辑器配置
type EditorConfig struct {
	ID           uuid.UUID           `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EditorType   string              `json:"editor_type" gorm:"size:50;not null;uniqueIndex"`
	Theme        string              `json:"theme" gorm:"size:50;default:default"`
	FontFamily   string              `json:"font_family" gorm:"size:100;default:monospace"`
	FontSize     int                 `json:"font_size" gorm:"default:14"`
	TabSize      int                 `json:"tab_size" gorm:"default:4"`
	WordWrap     bool                `json:"word_wrap" gorm:"default:true"`
	LineNumbers  bool                `json:"line_numbers" gorm:"default:true"`
	Settings     map[string]interface{} `json:"settings" gorm:"type:jsonb"`
	CreatedAt    time.Time           `json:"created_at" gorm:"default:now()"`
	UpdatedAt    time.Time           `json:"updated_at" gorm:"default:now()"`
}

// EditPermission 编辑权限
type EditPermission struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID uuid.UUID `json:"document_id" gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Role       string    `json:"role" gorm:"size:50;not null;check:role IN ('viewer', 'commenter', 'editor', 'admin', 'owner')"`
	Permissions []string `json:"permissions" gorm:"type:text[]"`
	GrantedBy  uuid.UUID `json:"granted_by" gorm:"type:uuid;not null;index"`
	GrantedAt  time.Time `json:"granted_at" gorm:"default:now()"`
	ExpiresAt  *time.Time `json:"expires_at"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`

	// 关联关系
	Document *Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Granter  *User     `json:"granter,omitempty" gorm:"foreignKey:GrantedBy"`
}

// TableName 指定表名
func (EditSession) TableName() string {
	return "edit_sessions"
}

func (EditOperation) TableName() string {
	return "edit_operations"
}

func (CollaborationSession) TableName() string {
	return "collaboration_sessions"
}

func (DocumentVersion) TableName() string {
	return "document_versions"
}

func (CollaborationParticipant) TableName() string {
	return "collaboration_participants"
}

func (EditorConfig) TableName() string {
	return "editor_configs"
}

func (EditPermission) TableName() string {
	return "edit_permissions"
}

// BeforeCreate GORM钩子
func (e *EditSession) BeforeCreate(tx *gorm.DB) error {
	if e.SessionToken == "" {
		e.SessionToken = uuid.New().String()
	}
	return nil
}

func (c *CollaborationSession) BeforeCreate(tx *gorm.DB) error {
	if c.RoomName == "" {
		c.RoomName = "doc_" + c.DocumentID.String()
	}
	return nil
}

func (c *CollaborationParticipant) BeforeCreate(tx *gorm.DB) error {
	if c.UserColor == "" {
		c.UserColor = "#2196F3" // 默认蓝色
	}
	return nil
}

// IsExpired 检查会话是否过期
func (e *EditSession) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// CanEdit 检查用户是否有编辑权限
func (p *EditPermission) CanEdit() bool {
	if !p.IsActive || p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
		return false
	}

	for _, perm := range p.Permissions {
		if perm == "write" || perm == "admin" || perm == "owner" {
			return true
		}
	}
	return false
}

// CanComment 检查用户是否有评论权限
func (p *EditPermission) CanComment() bool {
	if !p.IsActive || p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
		return false
	}

	for _, perm := range p.Permissions {
		if perm == "comment" || perm == "write" || perm == "admin" || perm == "owner" {
			return true
		}
	}
	return false
}