package preview

import (
	"time"

	"gorm.io/gorm"
)

// DocumentVersion 文档版本
type DocumentVersion struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	DocumentID    uint           `json:"document_id" gorm:"not null;index"`
	VersionNumber int            `json:"version_number" gorm:"not null"`
	Title         string         `json:"title" gorm:"size:255;not null"`
	Content       string         `json:"content" gorm:"type:text"`
	ContentType   string         `json:"content_type" gorm:"size:100;not null"` // pdf, docx, xlsx, txt, md等
	FileSize      int64          `json:"file_size" gorm:"default:0"`
	FileHash      string         `json:"file_hash" gorm:"size:64;uniqueIndex"`
	StoragePath   string         `json:"storage_path" gorm:"size:500"`
	ThumbnailPath string         `json:"thumbnail_path" gorm:"size:500"`

	// 版本控制信息
	ParentVersionID *uint          `json:"parent_version_id" gorm:"index"`
	VersionTag     string         `json:"version_tag" gorm:"size:50"` // v1.0, v1.1, v2.0等
	ChangeLog      string         `json:"change_log" gorm:"type:text"`
	IsMajor        bool           `json:"is_major" gorm:"default:false"`
	IsDraft        bool           `json:"is_draft" gorm:"default:false"`
	IsPublished    bool           `json:"is_published" gorm:"default:false"`

	// 编辑元信息
	EditorID       *uint          `json:"editor_id" gorm:"index"`
	EditReason     string         `json:"edit_reason" gorm:"size:500"`
	EditDuration   int            `json:"edit_duration"` // 编辑时长(分钟)
	CharacterCount int            `json:"character_count"`
	WordCount      int            `json:"word_count"`
	PageCount      int            `json:"page_count"`

	// 渲染和预览相关
	RenderStatus   string         `json:"render_status" gorm:"size:20;default:'pending'"` // pending, processing, completed, failed
	RenderOptions  string         `json:"render_options" gorm:"type:json"` // JSON格式的渲染选项
	PreviewHTML    string         `json:"preview_html" gorm:"type:longtext"`
	PreviewJSON    string         `json:"preview_json" gorm:"type:json"` // 预览数据(页面信息、目录等)

	// 时间戳
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Document       Document       `json:"document" gorm:"foreignKey:DocumentID"`
	ParentVersion  *DocumentVersion `json:"parent_version" gorm:"foreignKey:ParentVersionID"`
	Editor         *User          `json:"editor" gorm:"foreignKey:EditorID"`
	ChildVersions  []DocumentVersion `json:"child_versions,omitempty" gorm:"foreignKey:ParentVersionID"`

	// 版本差异
	Differences    []VersionDiff  `json:"differences,omitempty" gorm:"foreignKey:VersionID"`
	Annotations    []Annotation   `json:"annotations,omitempty" gorm:"foreignKey:VersionID"`
	Collaborations []CollaborationSession `json:"collaborations,omitempty" gorm:"foreignKey:VersionID"`
}

// Document 文档基础信息
type Document struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	TenantID    string         `json:"tenant_id" gorm:"not null;index"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`

	// 文档分类和标签
	CategoryID  *uint          `json:"category_id" gorm:"index"`
	Category    *DocumentCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Tags        []DocumentTag  `json:"tags,omitempty" gorm:"many2many:document_tags;"`

	// 文档状态和权限
	Status      string         `json:"status" gorm:"size:20;default:'draft'"` // draft, published, archived
	Visibility  string         `json:"visibility" gorm:"size:20;default:'private'"` // private, public, restricted
	AccessLevel string         `json:"access_level" gorm:"size:20;default:'read'"` // read, write, admin

	// 创建者和所有者
	CreatedBy   uint           `json:"created_by" gorm:"not null;index"`
	OwnerID     uint           `json:"owner_id" gorm:"not null;index"`

	// 文档设置
	AllowComments bool          `json:"allow_comments" gorm:"default:true"`
	AllowDownloads bool         `json:"allow_downloads" gorm:"default:true"`
	AllowPrints   bool          `json:"allow_prints" gorm:"default:true"`
	ExpiryDate    *time.Time    `json:"expiry_date"`

	// 统计信息
	ViewCount    int64          `json:"view_count" gorm:"default:0"`
	DownloadCount int64         `json:"download_count" gorm:"default:0"`
	ShareCount   int64          `json:"share_count" gorm:"default:0"`

	// 时间戳
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Creator      User           `json:"creator" gorm:"foreignKey:CreatedBy"`
	Owner        User           `json:"owner" gorm:"foreignKey:OwnerID"`
	Versions     []DocumentVersion `json:"versions,omitempty" gorm:"foreignKey:DocumentID;order:VersionNumber desc"`
	Comments     []Comment      `json:"comments,omitempty" gorm:"foreignKey:DocumentID"`
	Shares       []DocumentShare `json:"shares,omitempty" gorm:"foreignKey:DocumentID"`
	Activities   []Activity     `json:"activities,omitempty" gorm:"foreignKey:DocumentID"`
}

// VersionDiff 版本差异
type VersionDiff struct {
	ID              uint         `json:"id" gorm:"primaryKey"`
	VersionID       uint         `json:"version_id" gorm:"not null;index"`
	CompareVersionID *uint        `json:"compare_version_id" gorm:"index"`

	// 差异类型和位置
	DiffType        string       `json:"diff_type" gorm:"size:20;not null"` // insert, delete, modify, move
	LineNumber      int          `json:"line_number" gorm:"default:0"`
	CharPosition    int          `json:"char_position" gorm:"default:0"`

	// 差异内容
	OldContent      string       `json:"old_content" gorm:"type:text"`
	NewContent      string       `json:"new_content" gorm:"type:text"`
	OldLength       int          `json:"old_length" gorm:"default:0"`
	NewLength       int          `json:"new_length" gorm:"default:0"`

	// 差异统计
	InsertsCount    int          `json:"inserts_count" gorm:"default:0"`
	DeletesCount    int          `json:"deletes_count" gorm:"default:0"`
	ModifiesCount   int          `json:"modifies_count" gorm:"default:0"`

	// 时间戳
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`

	// 关联关系
	Version         DocumentVersion `json:"version" gorm:"foreignKey:VersionID"`
	CompareVersion  *DocumentVersion `json:"compare_version,omitempty" gorm:"foreignKey:CompareVersionID"`
}

// Annotation 文档注释
type Annotation struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	VersionID     uint           `json:"version_id" gorm:"not null;index"`
	DocumentID    uint           `json:"document_id" gorm:"not null;index"`

	// 注释位置信息
	PageNumber    int            `json:"page_number" gorm:"default:1"`
	PositionType  string         `json:"position_type" gorm:"size:20;not null"` // text, image, area, highlight
	PositionData  string         `json:"position_data" gorm:"type:json"` // JSON格式的位置数据
	SelectedText  string         `json:"selected_text" gorm:"type:text"`

	// 注释内容
	Content       string         `json:"content" gorm:"type:text;not null"`
	AnnotationType string        `json:"annotation_type" gorm:"size:20;not null"` // comment, note, highlight, bookmark
	Color         string         `json:"color" gorm:"size:7;default:'#ffff00'"` // 十六进制颜色

	// 注释状态
	Status        string         `json:"status" gorm:"size:20;default:'active'"` // active, resolved, deleted
	IsPrivate     bool           `json:"is_private" gorm:"default:false"`
	IsDeleted     bool           `json:"is_deleted" gorm:"default:false"`

	// 创建者和信息
	CreatedBy     uint           `json:"created_by" gorm:"not null;index"`
	RepliedTo     *uint          `json:"replied_to" gorm:"index"`

	// 时间戳
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Version       DocumentVersion `json:"version" gorm:"foreignKey:VersionID"`
	Document      Document       `json:"document" gorm:"foreignKey:DocumentID"`
	Creator       User           `json:"creator" gorm:"foreignKey:CreatedBy"`
	ReplyTo       *Annotation    `json:"reply_to,omitempty" gorm:"foreignKey:RepliedTo"`
	Replies       []Annotation   `json:"replies,omitempty" gorm:"foreignKey:RepliedTo"`

	// 附件
	Attachments   []AnnotationAttachment `json:"attachments,omitempty" gorm:"foreignKey:AnnotationID"`
}

// AnnotationAttachment 注释附件
type AnnotationAttachment struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	AnnotationID  uint         `json:"annotation_id" gorm:"not null;index"`

	// 文件信息
	FileName      string       `json:"file_name" gorm:"size:255;not null"`
	FileSize      int64        `json:"file_size" gorm:"not null"`
	FilePath      string       `json:"file_path" gorm:"size:500;not null"`
	ThumbnailPath string       `json:"thumbnail_path" gorm:"size:500"`
	MimeType      string       `json:"mime_type" gorm:"size:100;not null"`

	// 时间戳
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`

	// 关联关系
	Annotation    Annotation   `json:"annotation" gorm:"foreignKey:AnnotationID"`
}

// CollaborationSession 协作会话
type CollaborationSession struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	VersionID     uint           `json:"version_id" gorm:"not null;index"`
	DocumentID    uint           `json:"document_id" gorm:"not null;index"`

	// 会话信息
	SessionToken  string         `json:"session_token" gorm:"size:64;uniqueIndex;not null"`
	Title         string         `json:"title" gorm:"size:255"`
	Description   string         `json:"description" gorm:"type:text"`

	// 会话状态
	Status        string         `json:"status" gorm:"size:20;default:'active'"` // active, paused, ended
	SessionType   string         `json:"session_type" gorm:"size:20;default:'edit'"` // edit, review, comment

	// 会话设置
	MaxParticipants int          `json:"max_participants" gorm:"default:10"`
	AllowAnonymous  bool         `json:"allow_anonymous" gorm:"default:false"`
	RequireApproval bool         `json:"require_approval" gorm:"default:false"`

	// 会话控制
	OwnerID       uint           `json:"owner_id" gorm:"not null;index"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`

	// 时间设置
	ScheduledStart *time.Time    `json:"scheduled_start"`
	ScheduledEnd   *time.Time    `json:"scheduled_end"`
	ActualStart    *time.Time    `json:"actual_start"`
	ActualEnd      *time.Time    `json:"actual_end"`

	// 时间戳
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Version       DocumentVersion `json:"version" gorm:"foreignKey:VersionID"`
	Document      Document       `json:"document" gorm:"foreignKey:DocumentID"`
	Owner         User           `json:"owner" gorm:"foreignKey:OwnerID"`

	// 参与者
	Participants  []CollaborationParticipant `json:"participants,omitempty" gorm:"foreignKey:SessionID"`
	Operations    []CollaborationOperation   `json:"operations,omitempty" gorm:"foreignKey:SessionID"`
	Changes       []CollaborationChange      `json:"changes,omitempty" gorm:"foreignKey:SessionID"`
}

// CollaborationParticipant 协作参与者
type CollaborationParticipant struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	SessionID     uint           `json:"session_id" gorm:"not null;index"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`

	// 参与者信息
	DisplayName   string         `json:"display_name" gorm:"size:100;not null"`
	Email         string         `json:"email" gorm:"size:255"`
	Avatar        string         `json:"avatar" gorm:"size:500"`

	// 权限和角色
	Role          string         `json:"role" gorm:"size:20;default:'participant'"` // owner, editor, reviewer, participant
	Permissions   string         `json:"permissions" gorm:"type:json"` // JSON格式的权限列表

	// 连接状态
	Status        string         `json:"status" gorm:"size:20;default:'offline'"` // online, offline, away, busy
	LastSeen      *time.Time     `json:"last_seen"`
	Cursor        Cursor         `json:"cursor" gorm:"serializer:json"` // 当前光标位置
	Selection     Selection      `json:"selection" gorm:"serializer:json"` // 当前选择

	// 活动统计
	JoinTime      time.Time      `json:"join_time"`
	LeaveTime     *time.Time     `json:"leave_time"`
	TotalTime     int            `json:"total_time"` // 参与时长(分钟)
	EditsCount    int            `json:"edits_count" gorm:"default:0"`
	CommentsCount int            `json:"comments_count" gorm:"default:0"`

	// 时间戳
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	// 关联关系
	Session       CollaborationSession `json:"session" gorm:"foreignKey:SessionID"`
	User          User                 `json:"user" gorm:"foreignKey:UserID"`
}

// CollaborationOperation 协作操作
type CollaborationOperation struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	SessionID     uint         `json:"session_id" gorm:"not null;index"`
	UserID        uint         `json:"user_id" gorm:"not null;index"`

	// 操作信息
	OperationID   string       `json:"operation_id" gorm:"size:64;uniqueIndex;not null"`
	OperationType string       `json:"operation_type" gorm:"size:20;not null"` // insert, delete, retain, format
	Position      Position     `json:"position" gorm:"serializer:json"` // 操作位置

	// 操作内容
	Content       string       `json:"content" gorm:"type:text"`
	Attributes    string       `json:"attributes" gorm:"type:json"` // JSON格式的属性信息(格式、样式等)
	Length        int          `json:"length" gorm:"default:0"`

	// 操作状态
	Status        string       `json:"status" gorm:"size:20;default:'pending'"` // pending, applied, rejected, conflicted
	AppliedAt     *time.Time   `json:"applied_at"`

	// 时间戳
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`

	// 关联关系
	Session       CollaborationSession `json:"session" gorm:"foreignKey:SessionID"`
	User          User                 `json:"user" gorm:"foreignKey:UserID"`
}

// CollaborationChange 协作变更
type CollaborationChange struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	SessionID     uint         `json:"session_id" gorm:"not null;index"`
	UserID        uint         `json:"user_id" gorm:"not null;index"`
	OperationID   uint         `json:"operation_id" gorm:"not null;index"`

	// 变更信息
	ChangeType    string       `json:"change_type" gorm:"size:20;not null"` // content, format, structure, metadata
	FieldName     string       `json:"field_name" gorm:"size:100"`
	OldValue      string       `json:"old_value" gorm:"type:text"`
	NewValue      string       `json:"new_value" gorm:"type:text"`

	// 变更统计
	LinesAdded    int          `json:"lines_added" gorm:"default:0"`
	LinesRemoved  int          `json:"lines_removed" gorm:"default:0"`
	CharsAdded    int          `json:"chars_added" gorm:"default:0"`
	CharsRemoved  int          `json:"chars_removed" gorm:"default:0"`

	// 时间戳
	CreatedAt     time.Time    `json:"created_at"`

	// 关联关系
	Session       CollaborationSession `json:"session" gorm:"foreignKey:SessionID"`
	User          User                 `json:"user" gorm:"foreignKey:UserID"`
	Operation     CollaborationOperation `json:"operation" gorm:"foreignKey:OperationID"`
}

// Cursor 光标位置
type Cursor struct {
	Position  int    `json:"position"`
	Anchor    int    `json:"anchor"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
}

// Selection 选择区域
type Selection struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
	Text  string    `json:"text"`
}

// Position 位置信息
type Position struct {
	Line      int `json:"line"`
	Column    int `json:"column"`
	Character int `json:"character"`
}

// DocumentCategory 文档分类
type DocumentCategory struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	TenantID    string         `json:"tenant_id" gorm:"not null;index"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Description string         `json:"description" gorm:"type:text"`
	ParentID    *uint          `json:"parent_id" gorm:"index"`
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
	Color       string         `json:"color" gorm:"size:7;default:'#007bff'"`
	Icon        string         `json:"icon" gorm:"size:50"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Parent      *DocumentCategory `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children    []DocumentCategory `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Documents   []Document        `json:"documents,omitempty" gorm:"foreignKey:CategoryID"`
}

// DocumentTag 文档标签
type DocumentTag struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	TenantID  string         `json:"tenant_id" gorm:"not null;index"`
	Name      string         `json:"name" gorm:"size:50;not null"`
	Color     string         `json:"color" gorm:"size:7;default:'#6c757d'"`
	UsageCount int           `json:"usage_count" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// DocumentShare 文档分享
type DocumentShare struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	DocumentID    uint           `json:"document_id" gorm:"not null;index"`
	SharedBy      uint           `json:"shared_by" gorm:"not null;index"`
	SharedWith    *uint          `json:"shared_with" gorm:"index"` // 分享给的用户ID，null表示公开分享

	// 分享信息
	ShareToken    string         `json:"share_token" gorm:"size:64;uniqueIndex;not null"`
	ShareType     string         `json:"share_type" gorm:"size:20;not null"` // link, email, domain, public
	ShareScope    string         `json:"share_scope" gorm:"size:20;default:'document'"` // document, version, page

	// 权限控制
	Permissions   string         `json:"permissions" gorm:"type:json"` // JSON格式的权限列表
	AccessLevel   string         `json:"access_level" gorm:"size:20;default:'read'"` // read, write, comment

	// 分享设置
	AllowDownload bool           `json:"allow_download" gorm:"default:true"`
	AllowPrint    bool           `json:"allow_print" gorm:"default:true"`
	AllowCopy     bool           `json:"allow_copy" gorm:"default:true"`
	ShowComments  bool           `json:"show_comments" gorm:"default:false"`

	// 访问控制
	Password      string         `json:"password" gorm:"size:255"`
	MaxViews      int            `json:"max_views" gorm:"default:0"` // 0表示无限制
	ViewCount     int            `json:"view_count" gorm:"default:0"`

	// 时间控制
	ExpiresAt     *time.Time     `json:"expires_at"`
	LastAccessed  *time.Time     `json:"last_accessed"`

	// 状态
	Status        string         `json:"status" gorm:"size:20;default:'active'"` // active, suspended, revoked

	// 时间戳
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Document      Document       `json:"document" gorm:"foreignKey:DocumentID"`
	SharedByUser  User           `json:"shared_by_user" gorm:"foreignKey:SharedBy"`
	SharedWithUser *User         `json:"shared_with_user,omitempty" gorm:"foreignKey:SharedWith"`
}

// Comment 评论
type Comment struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	DocumentID    uint           `json:"document_id" gorm:"not null;index"`
	VersionID     *uint          `json:"version_id" gorm:"index"`
	ParentID      *uint          `json:"parent_id" gorm:"index"`

	// 评论内容
	Content       string         `json:"content" gorm:"type:text;not null"`
	CommentType   string         `json:"comment_type" gorm:"size:20;default:'general'"` // general, suggestion, issue, praise

	// 评论位置
	ContextType   string         `json:"context_type" gorm:"size:20"` // document, page, line, selection
	ContextData   string         `json:"context_data" gorm:"type:json"` // JSON格式的上下文数据
	PageNumber    int            `json:"page_number" gorm:"default:0"`
	LineNumber    int            `json:"line_number" gorm:"default:0"`

	// 评论状态
	Status        string         `json:"status" gorm:"size:20;default:'active'"` // active, resolved, deleted
	IsEdited      bool           `json:"is_edited" gorm:"default:false"`

	// 创建者和信息
	CreatedBy     uint           `json:"created_by" gorm:"not null;index"`
	EditorID      *uint          `json:"editor_id" gorm:"index"`

	// 时间戳
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Document      Document       `json:"document" gorm:"foreignKey:DocumentID"`
	Version       *DocumentVersion `json:"version,omitempty" gorm:"foreignKey:VersionID"`
	Parent        *Comment       `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies       []Comment      `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
	Creator       User           `json:"creator" gorm:"foreignKey:CreatedBy"`
	Editor        *User          `json:"editor,omitempty" gorm:"foreignKey:EditorID"`

	// 反应
	Reactions     []CommentReaction `json:"reactions,omitempty" gorm:"foreignKey:CommentID"`
}

// CommentReaction 评论反应
type CommentReaction struct {
	ID         uint         `json:"id" gorm:"primaryKey"`
	CommentID  uint         `json:"comment_id" gorm:"not null;index"`
	UserID     uint         `json:"user_id" gorm:"not null;index"`
	ReactionType string     `json:"reaction_type" gorm:"size:20;not null"` // like, dislike, heart, laugh, angry
	CreatedAt  time.Time    `json:"created_at"`

	// 关联关系
	Comment    Comment      `json:"comment" gorm:"foreignKey:CommentID"`
	User       User         `json:"user" gorm:"foreignKey:UserID"`
}

// Activity 活动记录
type Activity struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	DocumentID    uint           `json:"document_id" gorm:"not null;index"`
	VersionID     *uint          `json:"version_id" gorm:"index"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`

	// 活动信息
	ActivityType  string         `json:"activity_type" gorm:"size:50;not null"` // create, update, delete, view, share, comment
	Action        string         `json:"action" gorm:"size:50;not null"`
	Description   string         `json:"description" gorm:"size:500"`

	// 活动详情
	Details       string         `json:"details" gorm:"type:json"` // JSON格式的详细信息
	Metadata      string         `json:"metadata" gorm:"type:json"` // JSON格式的元数据

	// IP和设备信息
	IPAddress     string         `json:"ip_address" gorm:"size:45"`
	UserAgent     string         `json:"user_agent" gorm:"size:500"`

	// 时间戳
	CreatedAt     time.Time      `json:"created_at"`

	// 关联关系
	Document      Document       `json:"document" gorm:"foreignKey:DocumentID"`
	Version       *DocumentVersion `json:"version,omitempty" gorm:"foreignKey:VersionID"`
	User          User           `json:"user" gorm:"foreignKey:UserID"`
}

// User 用户信息(简化版)
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TenantID  string    `json:"tenant_id" gorm:"not null;index"`
	Username  string    `json:"username" gorm:"size:100;not null;uniqueIndex"`
	Email     string    `json:"email" gorm:"size:255;not null;uniqueIndex"`
	Nickname  string    `json:"nickname" gorm:"size:100"`
	Avatar    string    `json:"avatar" gorm:"size:500"`
	Status    string    `json:"status" gorm:"size:20;default:'active'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}