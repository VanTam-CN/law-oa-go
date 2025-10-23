package preview

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gopkg.in/yaml.v3"

	"law-oa-go/internal/config"
	"law-oa-go/internal/storage"
)

// Config 预览编辑服务配置
type Config struct {
	Preview    PreviewConfig    `yaml:"preview" json:"preview"`
	Rendering  RenderingConfig  `yaml:"rendering" json:"rendering"`
	Collaboration CollaborationConfig `yaml:"collaboration" json:"collaboration"`
	Cache      CacheConfig      `yaml:"cache" json:"cache"`
	Security   SecurityConfig   `yaml:"security" json:"security"`
}

// PreviewConfig 预览配置
type PreviewConfig struct {
	// 基础配置
	MaxWidth         int           `yaml:"max_width" json:"max_width"`             // 最大宽度
	MaxHeight        int           `yaml:"max_height" json:"max_height"`           // 最大高度
	MaxFileSize      int64         `yaml:"max_file_size" json:"max_file_size"`     // 最大文件大小(字节)
	Timeout          time.Duration `yaml:"timeout" json:"timeout"`                 // 处理超时时间
	WorkerCount      int           `yaml:"worker_count" json:"worker_count"`       // 工作协程数量
	QueueSize        int           `yaml:"queue_size" json:"queue_size"`           // 队列大小

	// 支持的格式
	SupportedFormats []string      `yaml:"supported_formats" json:"supported_formats"`

	// 默认选项
	DefaultQuality   int           `yaml:"default_quality" json:"default_quality"`   // 默认图片质量
	DefaultFormat    string        `yaml:"default_format" json:"default_format"`     // 默认输出格式
	DefaultScale     float64       `yaml:"default_scale" json:"default_scale"`       // 默认缩放比例
	DefaultDPI       int           `yaml:"default_dpi" json:"default_dpi"`           // 默认DPI

	// 缩略图配置
	ThumbnailSizes   []int         `yaml:"thumbnail_sizes" json:"thumbnail_sizes"`   // 缩略图尺寸列表
	ThumbnailQuality int           `yaml:"thumbnail_quality" json:"thumbnail_quality"` // 缩略图质量
}

// RenderingConfig 渲染配置
type RenderingConfig struct {
	// PDF渲染
	PDF PDFRenderConfig `yaml:"pdf" json:"pdf"`

	// Office文档渲染
	Office OfficeRenderConfig `yaml:"office" json:"office"`

	// 图片渲染
	Image ImageRenderConfig `yaml:"image" json:"image"`

	// 文本渲染
	Text TextRenderConfig `yaml:"text" json:"text"`

	// 渲染引擎配置
	Engines map[string]EngineConfig `yaml:"engines" json:"engines"`
}

// PDFRenderConfig PDF渲染配置
type PDFRenderConfig struct {
	Engine      string        `yaml:"engine" json:"engine"`           // 渲染引擎: pdfcpu, unipdf
	DPI         int           `yaml:"dpi" json:"dpi"`                 // 渲染DPI
	Quality     int           `yaml:"quality" json:"quality"`         // 渲染质量
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`         // 渲染超时
	EnableOCR   bool          `yaml:"enable_ocr" json:"enable_ocr"`   // 启用OCR
	OCRLanguage string        `yaml:"ocr_language" json:"ocr_language"` // OCR语言
	MaxPages    int           `yaml:"max_pages" json:"max_pages"`      // 最大页数限制
}

// OfficeRenderConfig Office文档渲染配置
type OfficeRenderConfig struct {
	Engine           string        `yaml:"engine" json:"engine"`                     // 渲染引擎
	Timeout          time.Duration `yaml:"timeout" json:"timeout"`                   // 转换超时
	MaxFileSize      int64         `yaml:"max_file_size" json:"max_file_size"`       // 最大文件大小
	TempDir          string        `yaml:"temp_dir" json:"temp_dir"`                 // 临时目录
	CleanupTemp      bool          `yaml:"cleanup_temp" json:"cleanup_temp"`         // 清理临时文件
	LibreOfficePath  string        `yaml:"libreoffice_path" json:"libreoffice_path"` // LibreOffice路径
}

// ImageRenderConfig 图片渲染配置
type ImageRenderConfig struct {
	MaxWidth       int      `yaml:"max_width" json:"max_width"`       // 最大宽度
	MaxHeight      int      `yaml:"max_height" json:"max_height"`     // 最大高度
	MaxFileSize    int64    `yaml:"max_file_size" json:"max_file_size"` // 最大文件大小
	SupportedTypes []string `yaml:"supported_types" json:"supported_types"` // 支持的图片类型
	AutoRotate     bool     `yaml:"auto_rotate" json:"auto_rotate"`     // 自动旋转
	PreserveEXIF   bool     `yaml:"preserve_exif" json:"preserve_exif"`   // 保留EXIF信息
}

// TextRenderConfig 文本渲染配置
type TextRenderConfig struct {
	MaxFileSize    int64    `yaml:"max_file_size" json:"max_file_size"` // 最大文件大小
	Encoding       string   `yaml:"encoding" json:"encoding"`           // 默认编码
	LineHeight     float64  `yaml:"line_height" json:"line_height"`     // 行高
	FontSize       int      `yaml:"font_size" json:"font_size"`         // 字体大小
	FontFamily     string   `yaml:"font_family" json:"font_family"`     // 字体
	WrapLength     int      `yaml:"wrap_length" json:"wrap_length"`     // 换行长度
	SyntaxHighlight bool    `yaml:"syntax_highlight" json:"syntax_highlight"` // 语法高亮
}

// EngineConfig 渲染引擎配置
type EngineConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Path        string            `yaml:"path" json:"path"`
	Version     string            `yaml:"version" json:"version"`
	Enabled     bool              `yaml:"enabled" json:"enabled"`
	Config      map[string]interface{} `yaml:"config" json:"config"`
}

// CollaborationConfig 协作配置
type CollaborationConfig struct {
	// 会话配置
	Session SessionConfig `yaml:"session" json:"session"`

	// WebSocket配置
	WebSocket WebSocketConfig `yaml:"websocket" json:"websocket"`

	// 操作配置
	Operation OperationConfig `yaml:"operation" json:"operation"`

	// 冲突解决配置
	Conflict ConflictConfig `yaml:"conflict" json:"conflict"`

	// 实时同步配置
	Sync SyncConfig `yaml:"sync" json:"sync"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	MaxParticipants    int           `yaml:"max_participants" json:"max_participants"`   // 最大参与者数量
	DefaultTimeout    time.Duration `yaml:"default_timeout" json:"default_timeout"`       // 默认会话超时
	MaxSessionDuration time.Duration `yaml:"max_session_duration" json:"max_session_duration"` // 最大会话持续时间
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval" json:"heartbeat_interval"`   // 心跳间隔
	IdleTimeout       time.Duration `yaml:"idle_timeout" json:"idle_timeout"`               // 空闲超时
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	ReadBufferSize    int           `yaml:"read_buffer_size" json:"read_buffer_size"`     // 读缓冲区大小
	WriteBufferSize   int           `yaml:"write_buffer_size" json:"write_buffer_size"`   // 写缓冲区大小
	PongWait         time.Duration `yaml:"pong_wait" json:"pong_wait"`                   // 等待Pong时间
	WriteWait        time.Duration `yaml:"write_wait" json:"write_wait"`                 // 写等待时间
	MaxMessageSize    int64         `yaml:"max_message_size" json:"max_message_size"`     // 最大消息大小
	EnableCompression bool          `yaml:"enable_compression" json:"enable_compression"` // 启用压缩
	CheckOrigin       bool          `yaml:"check_origin" json:"check_origin"`             // 检查Origin
	AllowedOrigins    []string      `yaml:"allowed_origins" json:"allowed_origins"`       // 允许的Origin
}

// OperationConfig 操作配置
type OperationConfig struct {
	MaxOperationSize  int           `yaml:"max_operation_size" json:"max_operation_size"`   // 最大操作大小
	BatchSize         int           `yaml:"batch_size" json:"batch_size"`                   // 批处理大小
	BatchTimeout      time.Duration `yaml:"batch_timeout" json:"batch_timeout"`             // 批处理超时
	ConflictWindowSize int          `yaml:"conflict_window_size" json:"conflict_window_size"` // 冲突窗口大小
	OperationHistory  int           `yaml:"operation_history" json:"operation_history"`     // 操作历史保留数量
}

// ConflictConfig 冲突解决配置
type ConflictConfig struct {
	Strategy       string        `yaml:"strategy" json:"strategy"`                 // 解决策略
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`                   // 解决超时
	AutoResolve    bool          `yaml:"auto_resolve" json:"auto_resolve"`         // 自动解决
	RetryAttempts  int           `yaml:"retry_attempts" json:"retry_attempts"`     // 重试次数
	RetryDelay     time.Duration `yaml:"retry_delay" json:"retry_delay"`           // 重试延迟
	ManualNotify   bool          `yaml:"manual_notify" json:"manual_notify"`       // 手动解决通知
}

// SyncConfig 同步配置
type SyncConfig struct {
	EnableSync       bool          `yaml:"enable_sync" json:"enable_sync"`           // 启用同步
	SyncInterval     time.Duration `yaml:"sync_interval" json:"sync_interval"`       // 同步间隔
	MaxPendingSyncs  int           `yaml:"max_pending_syncs" json:"max_pending_syncs"` // 最大待同步数
	SyncTimeout      time.Duration `yaml:"sync_timeout" json:"sync_timeout"`         // 同步超时
	PersistenceMode  string        `yaml:"persistence_mode" json:"persistence_mode"` // 持久化模式
	CompressionLevel int           `yaml:"compression_level" json:"compression_level"` // 压缩级别
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enable     bool          `yaml:"enable" json:"enable"`                     // 启用缓存
	Type       string        `yaml:"type" json:"type"`                         // 缓存类型: redis, memory, file
	TTL        time.Duration `yaml:"ttl" json:"ttl"`                           // 默认TTL
	MaxSize    int64         `yaml:"max_size" json:"max_size"`                 // 最大缓存大小
	EvictionPolicy string     `yaml:"eviction_policy" json:"eviction_policy"`   // 淘汰策略
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"` // 清理间隔

	// Redis配置
	Redis RedisConfig `yaml:"redis" json:"redis"`

	// 文件缓存配置
	File FileCacheConfig `yaml:"file" json:"file"`
}

// RedisConfig Redis缓存配置
type RedisConfig struct {
	Addr     string `yaml:"addr" json:"addr"`                 // Redis地址
	Password string `yaml:"password" json:"password"`         // Redis密码
	DB       int    `yaml:"db" json:"db"`                     // 数据库编号
	PoolSize int    `yaml:"pool_size" json:"pool_size"`       // 连接池大小
	KeyPrefix string `yaml:"key_prefix" json:"key_prefix"`     // 键前缀
}

// FileCacheConfig 文件缓存配置
type FileCacheConfig struct {
	Dir         string `yaml:"dir" json:"dir"`                     // 缓存目录
	MaxFiles    int    `yaml:"max_files" json:"max_files"`         // 最大文件数
	MaxFileSize int64  `yaml:"max_file_size" json:"max_file_size"` // 最大文件大小
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 访问控制
	AccessControl AccessControlConfig `yaml:"access_control" json:"access_control"`

	// 内容安全
	ContentSecurity ContentSecurityConfig `yaml:"content_security" json:"content_security"`

	// 下载控制
	DownloadControl DownloadControlConfig `yaml:"download_control" json:"download_control"`

	// 审计配置
	Audit AuditConfig `yaml:"audit" json:"audit"`
}

// AccessControlConfig 访问控制配置
type AccessControlConfig struct {
	EnableAuth          bool     `yaml:"enable_auth" json:"enable_auth"`           // 启用认证
	RequiredRoles       []string `yaml:"required_roles" json:"required_roles"`     // 必需角色
	IPWhitelist         []string `yaml:"ip_whitelist" json:"ip_whitelist"`         // IP白名单
	IPBlacklist         []string `yaml:"ip_blacklist" json:"ip_blacklist"`         // IP黑名单
	RateLimitPerMinute  int      `yaml:"rate_limit_per_minute" json:"rate_limit_per_minute"` // 每分钟速率限制
	MaxConcurrentAccess int      `yaml:"max_concurrent_access" json:"max_concurrent_access"` // 最大并发访问
}

// ContentSecurityConfig 内容安全配置
type ContentSecurityConfig struct {
	EnableScan      bool     `yaml:"enable_scan" json:"enable_scan"`           // 启用内容扫描
	BlockedTypes    []string `yaml:"blocked_types" json:"blocked_types"`       // 阻止的文件类型
	MaxContentSize  int64    `yaml:"max_content_size" json:"max_content_size"` // 最大内容大小
	AllowExternalLinks bool   `yaml:"allow_external_links" json:"allow_external_links"` // 允许外部链接
	WatermarkEnabled bool     `yaml:"watermark_enabled" json:"watermark_enabled"` // 启用水印
}

// DownloadControlConfig 下载控制配置
type DownloadControlConfig struct {
	EnableDownload     bool          `yaml:"enable_download" json:"enable_download"`         // 启用下载
	RequireAuth        bool          `yaml:"require_auth" json:"require_auth"`               // 下载需要认证
	DownloadLimit      int           `yaml:"download_limit" json:"download_limit"`           // 下载限制
	DownloadWindow     time.Duration `yaml:"download_window" json:"download_window"`         // 下载窗口
	TrackDownloads     bool          `yaml:"track_downloads" json:"track_downloads"`         // 跟踪下载
	PreventScreenshots bool          `yaml:"prevent_screenshots" json:"prevent_screenshots"` // 防止截图
}

// AuditConfig 审计配置
type AuditConfig struct {
	EnableAudit    bool          `yaml:"enable_audit" json:"enable_audit"`     // 启用审计
	LogLevel       string        `yaml:"log_level" json:"log_level"`           // 日志级别
	RetentionPeriod time.Duration `yaml:"retention_period" json:"retention_period"` // 保留期
	LogFile        string        `yaml:"log_file" json:"log_file"`             // 日志文件
	LogFormat      string        `yaml:"log_format" json:"log_format"`         // 日志格式
	AuditEvents    []string      `yaml:"audit_events" json:"audit_events"`     // 审计事件
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	var config Config

	// 设置默认值
	config = getDefaultConfig()

	// 如果指定了配置文件，则加载
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}

		// 根据文件扩展名选择解析器
		ext := filepath.Ext(configPath)
		switch ext {
		case ".yaml", ".yml":
			err = yaml.Unmarshal(data, &config)
		case ".json":
			err = json.Unmarshal(data, &config)
		default:
			return nil, fmt.Errorf("不支持的配置文件格式: %s", ext)
		}

		if err != nil {
			return nil, fmt.Errorf("解析配置文件失败: %w", err)
		}
	}

	// 从环境变量覆盖配置
	err := config.loadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("加载环境变量配置失败: %w", err)
	}

	// 验证配置
	err = config.validate()
	if err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() Config {
	return Config{
		Preview: PreviewConfig{
			MaxWidth:      2000,
			MaxHeight:     2000,
			MaxFileSize:   100 * 1024 * 1024, // 100MB
			Timeout:       30 * time.Second,
			WorkerCount:   5,
			QueueSize:     100,
			SupportedFormats: []string{
				"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx",
				"txt", "md", "rtf",
				"jpg", "jpeg", "png", "gif", "bmp", "webp",
			},
			DefaultQuality:   90,
			DefaultFormat:    "jpg",
			DefaultScale:     1.0,
			DefaultDPI:       150,
			ThumbnailSizes:   []int{150, 300, 600},
			ThumbnailQuality: 80,
		},
		Rendering: RenderingConfig{
			PDF: PDFRenderConfig{
				Engine:      "pdfcpu",
				DPI:         150,
				Quality:     90,
				Timeout:     30 * time.Second,
				EnableOCR:   false,
				OCRLanguage: "chi_sim+eng",
				MaxPages:    1000,
			},
			Office: OfficeRenderConfig{
				Engine:          "libreoffice",
				Timeout:         60 * time.Second,
				MaxFileSize:     50 * 1024 * 1024, // 50MB
				TempDir:         "/tmp/preview",
				CleanupTemp:     true,
				LibreOfficePath: "/usr/bin/libreoffice",
			},
			Image: ImageRenderConfig{
				MaxWidth:       4000,
				MaxHeight:      4000,
				MaxFileSize:    20 * 1024 * 1024, // 20MB
				SupportedTypes: []string{"jpg", "jpeg", "png", "gif", "bmp", "webp"},
				AutoRotate:     true,
				PreserveEXIF:   false,
			},
			Text: TextRenderConfig{
				MaxFileSize:     10 * 1024 * 1024, // 10MB
				Encoding:        "utf-8",
				LineHeight:      1.5,
				FontSize:        12,
				FontFamily:      "monospace",
				WrapLength:      80,
				SyntaxHighlight: true,
			},
		},
		Collaboration: CollaborationConfig{
			Session: SessionConfig{
				MaxParticipants:    10,
				DefaultTimeout:    2 * time.Hour,
				MaxSessionDuration: 24 * time.Hour,
				HeartbeatInterval: 30 * time.Second,
				IdleTimeout:       10 * time.Minute,
			},
			WebSocket: WebSocketConfig{
				ReadBufferSize:    1024,
				WriteBufferSize:   1024,
				PongWait:         60 * time.Second,
				WriteWait:        10 * time.Second,
				MaxMessageSize:    8192,
				EnableCompression: true,
				CheckOrigin:       true,
				AllowedOrigins:    []string{"*"},
			},
			Operation: OperationConfig{
				MaxOperationSize:   1024,
				BatchSize:          10,
				BatchTimeout:       1 * time.Second,
				ConflictWindowSize: 100,
				OperationHistory:   1000,
			},
			Conflict: ConflictConfig{
				Strategy:      "merge",
				Timeout:       30 * time.Second,
				AutoResolve:   true,
				RetryAttempts: 3,
				RetryDelay:    1 * time.Second,
				ManualNotify:  true,
			},
			Sync: SyncConfig{
				EnableSync:        true,
				SyncInterval:      1 * time.Second,
				MaxPendingSyncs:   100,
				SyncTimeout:       10 * time.Second,
				PersistenceMode:   "memory",
				CompressionLevel:  6,
			},
		},
		Cache: CacheConfig{
			Enable:          true,
			Type:            "memory",
			TTL:             24 * time.Hour,
			MaxSize:         1 * 1024 * 1024 * 1024, // 1GB
			EvictionPolicy:  "lru",
			CleanupInterval: 1 * time.Hour,
			Redis: RedisConfig{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
				PoolSize: 10,
				KeyPrefix: "preview:",
			},
			File: FileCacheConfig{
				Dir:         "/tmp/preview-cache",
				MaxFiles:    10000,
				MaxFileSize: 10 * 1024 * 1024, // 10MB
			},
		},
		Security: SecurityConfig{
			AccessControl: AccessControlConfig{
				EnableAuth:          true,
				RequiredRoles:       []string{"user"},
				RateLimitPerMinute:  100,
				MaxConcurrentAccess: 10,
			},
			ContentSecurity: ContentSecurityConfig{
				EnableScan:         false,
				MaxContentSize:     100 * 1024 * 1024, // 100MB
				AllowExternalLinks: true,
				WatermarkEnabled:   false,
			},
			DownloadControl: DownloadControlConfig{
				EnableDownload:     true,
				RequireAuth:        true,
				DownloadLimit:      100,
				DownloadWindow:     24 * time.Hour,
				TrackDownloads:     true,
				PreventScreenshots: false,
			},
			Audit: AuditConfig{
				EnableAudit:    true,
				LogLevel:       "info",
				RetentionPeriod: 30 * 24 * time.Hour, // 30天
				LogFile:        "/var/log/preview/audit.log",
				LogFormat:      "json",
				AuditEvents:    []string{"access", "download", "collaboration"},
			},
		},
	}
}

// loadFromEnv 从环境变量加载配置
func (c *Config) loadFromEnv() error {
	// 预览配置
	if width := os.Getenv("PREVIEW_MAX_WIDTH"); width != "" {
		if w, err := parseInt(width); err == nil {
			c.Preview.MaxWidth = w
		}
	}

	if height := os.Getenv("PREVIEW_MAX_HEIGHT"); height != "" {
		if h, err := parseInt(height); err == nil {
			c.Preview.MaxHeight = h
		}
	}

	// 渲染配置
	if engine := os.Getenv("RENDER_PDF_ENGINE"); engine != "" {
		c.Rendering.PDF.Engine = engine
	}

	// 缓存配置
	if cacheType := os.Getenv("CACHE_TYPE"); cacheType != "" {
		c.Cache.Type = cacheType
	}

	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		c.Cache.Redis.Addr = redisAddr
	}

	// 协作配置
	if maxParticipants := os.Getenv("COLLABORATION_MAX_PARTICIPANTS"); maxParticipants != "" {
		if mp, err := parseInt(maxParticipants); err == nil {
			c.Collaboration.Session.MaxParticipants = mp
		}
	}

	return nil
}

// validate 验证配置
func (c *Config) validate() error {
	// 验证预览配置
	if c.Preview.MaxWidth <= 0 || c.Preview.MaxHeight <= 0 {
		return fmt.Errorf("预览宽度和高度必须大于0")
	}

	if c.Preview.MaxFileSize <= 0 {
		return fmt.Errorf("最大文件大小必须大于0")
	}

	if c.Preview.WorkerCount <= 0 {
		return fmt.Errorf("工作协程数量必须大于0")
	}

	// 验证渲染配置
	if !isValidEngine(c.Rendering.PDF.Engine) {
		return fmt.Errorf("无效的PDF渲染引擎: %s", c.Rendering.PDF.Engine)
	}

	if c.Rendering.PDF.DPI <= 0 {
		return fmt.Errorf("PDF DPI必须大于0")
	}

	// 验证协作配置
	if c.Collaboration.Session.MaxParticipants <= 0 {
		return fmt.Errorf("最大参与者数量必须大于0")
	}

	// 验证缓存配置
	if c.Cache.Enable {
		if c.Cache.TTL <= 0 {
			return fmt.Errorf("缓存TTL必须大于0")
		}

		if c.Cache.Type == "redis" && c.Cache.Redis.Addr == "" {
			return fmt.Errorf("Redis缓存需要指定地址")
		}
	}

	return nil
}

// isValidEngine 检查是否为有效的渲染引擎
func isValidEngine(engine string) bool {
	validEngines := []string{"pdfcpu", "unipdf", "libreoffice"}
	for _, valid := range validEngines {
		if engine == valid {
			return true
		}
	}
	return false
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// SaveConfig 保存配置
func (c *Config) SaveConfig(configPath string) error {
	var data []byte
	var err error

	// 根据文件扩展名选择格式
	ext := filepath.Ext(configPath)
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(c)
	case ".json":
		data, err = json.MarshalIndent(c, "", "  ")
	default:
		return fmt.Errorf("不支持的配置文件格式: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 写入文件
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// ToYAML 转换为YAML格式
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// ToJSON 转换为JSON格式
func (c *Config) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// LogConfig 记录配置
func (c *Config) LogConfig(logger *logrus.Logger) {
	logger.Info("预览编辑服务配置:")

	logger.WithFields(logrus.Fields{
		"max_width":     c.Preview.MaxWidth,
		"max_height":    c.Preview.MaxHeight,
		"max_file_size": c.Preview.MaxFileSize,
		"worker_count":  c.Preview.WorkerCount,
	}).Info("预览配置")

	logger.WithFields(logrus.Fields{
		"pdf_engine":     c.Rendering.PDF.Engine,
		"pdf_dpi":        c.Rendering.PDF.DPI,
		"pdf_quality":    c.Rendering.PDF.Quality,
		"pdf_timeout":    c.Rendering.PDF.Timeout,
	}).Info("渲染配置")

	logger.WithFields(logrus.Fields{
		"cache_enabled": c.Cache.Enable,
		"cache_type":    c.Cache.Type,
		"cache_ttl":     c.Cache.TTL,
		"cache_max_size": c.Cache.MaxSize,
	}).Info("缓存配置")

	logger.WithFields(logrus.Fields{
		"max_participants":    c.Collaboration.Session.MaxParticipants,
		"session_timeout":     c.Collaboration.Session.DefaultTimeout,
		"websocket_enabled":   true,
		"conflict_strategy":   c.Collaboration.Conflict.Strategy,
	}).Info("协作配置")
}