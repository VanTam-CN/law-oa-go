package security

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	casbinpersist "github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

// AccessControlService 访问控制服务
// 支持RBAC + ABAC混合模式，为律师事务所提供细粒度权限管理
type AccessControlService struct {
	enforcer        *casbin.Enforcer
	logger          *slog.Logger
	db              *gorm.DB
	cache           *PermissionCache
	auditLogger     *AccessAuditLogger
	policyManager   *PolicyManager
	attributeStore  *AttributeStore
	config          *AccessControlConfig
	mu              sync.RWMutex
}

// AccessControlConfig 访问控制配置
type AccessControlConfig struct {
	EnableRBAC       bool          `json:"enable_rbac"`
	EnableABAC       bool          `json:"enable_abac"`
	EnableCache      bool          `json:"enable_cache"`
	CacheTTL         time.Duration `json:"cache_ttl"`
	EnableAudit      bool          `json:"enable_audit"`
	DefaultDeny      bool          `json:"default_deny"`
	MaxHierarchyDepth int          `json:"max_hierarchy_depth"`
}

// Permission 权限定义
type Permission struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Resource    string            `json:"resource"`
	Action      string            `json:"action"`
	Attributes  map[string]string `json:"attributes"`
	Category    string            `json:"category"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Role 角色定义
type Role struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Level       int               `json:"level"`
	Department  string            `json:"department"`
	Attributes  map[string]string `json:"attributes"`
	IsSystem    bool              `json:"is_system"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// User 用户信息
type User struct {
	ID           string            `json:"id"`
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	FullName     string            `json:"full_name"`
	Roles        []string          `json:"roles"`
	Department   string            `json:"department"`
	Position     string            `json:"position"`
	Attributes   map[string]string `json:"attributes"`
	IsActive     bool              `json:"is_active"`
	LastLoginAt  *time.Time        `json:"last_login_at"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// Resource 资源定义
type Resource struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Owner       string            `json:"owner"`
	Attributes  map[string]string `json:"attributes"`
	Sensitivity string            `json:"sensitivity"` // public, internal, confidential, restricted
	Category    string            `json:"category"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// AccessRequest 访问请求
type AccessRequest struct {
	UserID     string            `json:"user_id"`
	ResourceID string            `json:"resource_id"`
	Action     string            `json:"action"`
	Context    map[string]string `json:"context"`
	IPAddress  string            `json:"ip_address"`
	UserAgent  string            `json:"user_agent"`
	RequestID  string            `json:"request_id"`
	Timestamp  time.Time         `json:"timestamp"`
}

// AccessDecision 访问决策
type AccessDecision struct {
	Allowed    bool              `json:"allowed"`
	Reason     string            `json:"reason"`
	Policies   []string          `json:"policies"`
	Attributes map[string]string `json:"attributes"`
	Duration   time.Duration     `json:"duration"`
	RequestID  string            `json:"request_id"`
	Timestamp  time.Time         `json:"timestamp"`
}

// AccessContext 访问上下文
type AccessContext struct {
	User       *User             `json:"user"`
	Resource   *Resource         `json:"resource"`
	Action     string            `json:"action"`
	Time       time.Time         `json:"time"`
	Location   string            `json:"location"`
	Device     string            `json:"device"`
	SessionID  string            `json:"session_id"`
	Attributes map[string]string `json:"attributes"`
}

// PolicyManager 策略管理器
type PolicyManager struct {
	enforcer *casbin.Enforcer
	logger   *slog.Logger
	db       *gorm.DB
}

// AttributeStore 属性存储
type AttributeStore struct {
	db     *gorm.DB
	logger *slog.Logger
	cache  map[string]map[string]string
	mu     sync.RWMutex
}

// PermissionCache 权限缓存
type PermissionCache struct {
	data    map[string]*CacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
	enabled bool
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Value     bool
	ExpiresAt time.Time
	Decision  *AccessDecision
}

// AccessAuditLogger 访问审计日志器
type AccessAuditLogger struct {
	logger *slog.Logger
	db     *gorm.DB
}

// AccessLog 访问日志
type AccessLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"user_id"`
	Resource  string    `gorm:"index" json:"resource"`
	Action    string    `gorm:"index" json:"action"`
	Allowed   bool      `gorm:"index" json:"allowed"`
	Reason    string    `json:"reason"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	RequestID string    `gorm:"uniqueIndex" json:"request_id"`
	Duration  int64     `json:"duration"` // 毫秒
	CreatedAt time.Time `json:"created_at"`
}

// NewAccessControlService 创建访问控制服务
func NewAccessControlService(db *gorm.DB, logger *slog.Logger, config *AccessControlConfig) (*AccessControlService, error) {
	// 初始化Casbin模型
	model := casbinmodel.NewModel()

	// 请求定义: sub, obj, act, ctx (subject, object, action, context)
	model.AddDef("r", "r", "sub, obj, act, ctx")

	// 策略定义: sub, obj, act, eft (subject, object, action, effect)
	model.AddDef("p", "p", "sub, obj, act, eft")

	// 角色定义: _, _ (role inheritance)
	model.AddDef("g", "g", "_, _")

	// 角色用户定义: _, _ (user-role mapping)
	model.AddDef("g2", "g2", "_, _")

	// 策略效果: 允许优先
	model.AddDef("e", "e", "some(where (p.eft == allow))")

	// 匹配器：支持RBAC和ABAC混合模式
	matcher := `
		# RBAC检查
		g(r.sub, p.sub) && g2(r.sub, p.sub)

		# 基础权限检查
		&& r.obj == p.obj && r.act == p.act

		# ABAC属性检查
		&& eval(abac_rule(r.ctx, p.sub, r.obj, r.act))
	`
	model.AddDef("m", "m", matcher)

	// 创建适配器
	adapter := NewGormAdapter(db)

	// 创建执行器
	enforcer, err := casbin.NewEnforcer(model, adapter)
	if err != nil {
		return nil, fmt.Errorf("创建Casbin执行器失败: %w", err)
	}

	// 设置默认配置
	if config == nil {
		config = &AccessControlConfig{
			EnableRBAC:       true,
			EnableABAC:       true,
			EnableCache:      true,
			CacheTTL:         5 * time.Minute,
			EnableAudit:      true,
			DefaultDeny:      true,
			MaxHierarchyDepth: 5,
		}
	}

	service := &AccessControlService{
		enforcer:       enforcer,
		logger:         logger,
		db:             db,
		config:         config,
		policyManager:  NewPolicyManager(enforcer, logger, db),
		attributeStore: NewAttributeStore(db, logger),
		auditLogger:    NewAccessAuditLogger(logger, db),
	}

	// 初始化缓存
	if config.EnableCache {
		service.cache = NewPermissionCache(config.CacheTTL)
	}

	// 自动迁移数据库表
	if err := service.autoMigrate(); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 加载默认策略
	if err := service.loadDefaultPolicies(); err != nil {
		return nil, fmt.Errorf("加载默认策略失败: %w", err)
	}

	logger.Info("访问控制服务初始化完成",
		"rbac_enabled", config.EnableRBAC,
		"abac_enabled", config.EnableABAC,
		"cache_enabled", config.EnableCache,
		"audit_enabled", config.EnableAudit,
	)

	return service, nil
}

// CheckPermission 检查权限
func (acs *AccessControlService) CheckPermission(ctx context.Context, req *AccessRequest) (*AccessDecision, error) {
	startTime := time.Now()

	// 验证请求
	if err := acs.validateRequest(req); err != nil {
		return &AccessDecision{
			Allowed:   false,
			Reason:    fmt.Sprintf("请求验证失败: %v", err),
			RequestID: req.RequestID,
			Timestamp: time.Now(),
		}, nil
	}

	// 检查缓存
	if acs.config.EnableCache && acs.cache != nil {
		if cached := acs.cache.Get(req.CacheKey()); cached != nil {
			acs.logAccess(req, cached, time.Since(startTime))
			return cached, nil
		}
	}

	// 构建访问上下文
	accessCtx, err := acs.buildAccessContext(ctx, req)
	if err != nil {
		return &AccessDecision{
			Allowed:   false,
			Reason:    fmt.Sprintf("构建访问上下文失败: %v", err),
			RequestID: req.RequestID,
			Timestamp: time.Now(),
		}, nil
	}

	// 执行权限检查
	decision, err := acs.evaluatePermission(accessCtx)
	if err != nil {
		return &AccessDecision{
			Allowed:   false,
			Reason:    fmt.Sprintf("权限评估失败: %v", err),
			RequestID: req.RequestID,
			Timestamp: time.Now(),
		}, nil
	}

	// 更新缓存
	if acs.config.EnableCache && acs.cache != nil {
		acs.cache.Set(req.CacheKey(), decision)
	}

	// 记录审计日志
	decision.Duration = time.Since(startTime)
	acs.logAccess(req, decision, decision.Duration)

	return decision, nil
}

// buildAccessContext 构建访问上下文
func (acs *AccessControlService) buildAccessContext(ctx context.Context, req *AccessRequest) (*AccessContext, error) {
	// 获取用户信息
	user, err := acs.attributeStore.GetUser(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 获取资源信息
	resource, err := acs.attributeStore.GetResource(req.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("获取资源信息失败: %w", err)
	}

	// 构建上下文
	accessCtx := &AccessContext{
		User:       user,
		Resource:   resource,
		Action:     req.Action,
		Time:       time.Now(),
		Location:   req.Context["location"],
		Device:     req.Context["device"],
		SessionID:  req.Context["session_id"],
		Attributes: req.Context,
	}

	return accessCtx, nil
}

// evaluatePermission 评估权限
func (acs *AccessControlService) evaluatePermission(accessCtx *AccessContext) (*AccessDecision, error) {
	decision := &AccessDecision{
		Allowed:    acs.config.DefaultDeny,
		Reason:     "默认拒绝",
		Policies:   []string{},
		Attributes: make(map[string]string),
		Timestamp:  time.Now(),
	}

	// 1. 检查用户状态
	if !accessCtx.User.IsActive {
		decision.Reason = "用户已被禁用"
		return decision, nil
	}

	// 2. RBAC权限检查
	if acs.config.EnableRBAC {
		rbacResult, err := acs.evaluateRBAC(accessCtx)
		if err != nil {
			return nil, fmt.Errorf("RBAC评估失败: %w", err)
		}

		if rbacResult.Allowed {
			decision.Allowed = true
			decision.Reason = rbacResult.Reason
			decision.Policies = append(decision.Policies, rbacResult.Policies...)
		}
	}

	// 3. ABAC属性检查
	if acs.config.EnableABAC {
		abacResult, err := acs.evaluateABAC(accessCtx)
		if err != nil {
			return nil, fmt.Errorf("ABAC评估失败: %w", err)
		}

		// ABAC可以覆盖RBAC的决策
		if abacResult.Allowed {
			decision.Allowed = true
			decision.Reason = abacResult.Reason
			decision.Policies = append(decision.Policies, abacResult.Policies...)
		} else if !abacResult.Allowed && decision.Allowed {
			// ABAC拒绝可以覆盖RBAC允许
			decision.Allowed = false
			decision.Reason = abacResult.Reason
		}

		decision.Attributes = abacResult.Attributes
	}

	// 4. 特殊规则检查
	specialResult := acs.evaluateSpecialRules(accessCtx)
	if specialResult != nil {
		decision.Allowed = specialResult.Allowed
		decision.Reason = specialResult.Reason
		decision.Policies = append(decision.Policies, specialResult.Policies...)
	}

	return decision, nil
}

// evaluateRBAC 评估RBAC权限
func (acs *AccessControlService) evaluateRBAC(accessCtx *AccessContext) (*AccessDecision, error) {
	decision := &AccessDecision{
		Allowed:  false,
		Reason:   "RBAC权限不足",
		Policies: []string{},
	}

	// 检查直接权限
	for _, role := range accessCtx.User.Roles {
		allowed, err := acs.enforcer.Enforce(role, accessCtx.Resource.ID, accessCtx.Action)
		if err != nil {
			return nil, fmt.Errorf("Casbin权限检查失败: %w", err)
		}

		if allowed {
			decision.Allowed = true
			decision.Reason = fmt.Sprintf("角色 %s 拥有权限", role)
			decision.Policies = append(decision.Policies, fmt.Sprintf("role:%s", role))
			break
		}
	}

	// 检查继承权限（层级角色）
	if !decision.Allowed {
		implicitRoles, err := acs.enforcer.GetImplicitRolesForUser(accessCtx.User.ID)
		if err != nil {
			return nil, fmt.Errorf("获取继承角色失败: %w", err)
		}

		for _, role := range implicitRoles {
			allowed, err := acs.enforcer.Enforce(role, accessCtx.Resource.ID, accessCtx.Action)
			if err != nil {
				return nil, fmt.Errorf("继承权限检查失败: %w", err)
			}

			if allowed {
				decision.Allowed = true
				decision.Reason = fmt.Sprintf("继承角色 %s 拥有权限", role)
				decision.Policies = append(decision.Policies, fmt.Sprintf("inherited_role:%s", role))
				break
			}
		}
	}

	return decision, nil
}

// evaluateABAC 评估ABAC权限
func (acs *AccessControlService) evaluateABAC(accessCtx *AccessContext) (*AccessDecision, error) {
	decision := &AccessDecision{
		Allowed:    false,
		Reason:     "ABAC权限不足",
		Policies:   []string{},
		Attributes: make(map[string]string),
	}

	// 构建属性上下文
	attributes := make(map[string]interface{})

	// 用户属性
	for k, v := range accessCtx.User.Attributes {
		attributes[fmt.Sprintf("user.%s", k)] = v
	}
	attributes["user.id"] = accessCtx.User.ID
	attributes["user.roles"] = accessCtx.User.Roles
	attributes["user.department"] = accessCtx.User.Department
	attributes["user.position"] = accessCtx.User.Position

	// 资源属性
	for k, v := range accessCtx.Resource.Attributes {
		attributes[fmt.Sprintf("resource.%s", k)] = v
	}
	attributes["resource.id"] = accessCtx.Resource.ID
	attributes["resource.type"] = accessCtx.Resource.Type
	attributes["resource.owner"] = accessCtx.Resource.Owner
	attributes["resource.sensitivity"] = accessCtx.Resource.Sensitivity
	attributes["resource.category"] = accessCtx.Resource.Category

	// 环境属性
	attributes["action"] = accessCtx.Action
	attributes["time.hour"] = accessCtx.Time.Hour()
	attributes["time.weekday"] = accessCtx.Time.Weekday()
	attributes["time.date"] = accessCtx.Time.Format("2006-01-02")
	attributes["location"] = accessCtx.Location
	attributes["device"] = accessCtx.Device

	// 序列化属性上下文
	contextJSON, err := json.Marshal(attributes)
	if err != nil {
		return nil, fmt.Errorf("序列化属性上下文失败: %w", err)
	}

	// 执行ABAC规则检查
	allowed, err := acs.enforcer.Enforce(
		accessCtx.User.ID,
		accessCtx.Resource.ID,
		accessCtx.Action,
		string(contextJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("ABAC规则检查失败: %w", err)
	}

	if allowed {
		decision.Allowed = true
		decision.Reason = "ABAC规则允许访问"
		decision.Policies = append(decision.Policies, "abac_rules")
	}

	// 保存属性到决策中
	for k, v := range attributes {
		if str, ok := v.(string); ok {
			decision.Attributes[k] = str
		}
	}

	return decision, nil
}

// evaluateSpecialRules 评估特殊规则
func (acs *AccessControlService) evaluateSpecialRules(accessCtx *AccessContext) *AccessDecision {
	// 规则1: 资源所有者拥有完全访问权限
	if accessCtx.User.ID == accessCtx.Resource.Owner {
		return &AccessDecision{
			Allowed:  true,
			Reason:   "资源所有者权限",
			Policies: []string{"resource_owner"},
		}
	}

	// 规则2: 管理员拥有完全访问权限
	for _, role := range accessCtx.User.Roles {
		if role == "admin" || role == "super_admin" {
			return &AccessDecision{
				Allowed:  true,
				Reason:   fmt.Sprintf("管理员权限 (%s)", role),
				Policies: []string{fmt.Sprintf("admin_role:%s", role)},
			}
		}
	}

	// 规则3: 律师事务所特殊规则 - 同部门律师可以访问内部文档
	if accessCtx.Resource.Sensitivity == "internal" {
		if accessCtx.User.Department == accessCtx.Resource.Attributes["department"] {
			for _, role := range accessCtx.User.Roles {
				if role == "lawyer" || role == "senior_lawyer" || role == "partner" {
					return &AccessDecision{
						Allowed:  true,
						Reason:   "同部门律师权限",
						Policies: []string{"department_access"},
					}
				}
			}
		}
	}

	// 规则4: 时间限制规则 - 工作时间外限制敏感资源访问
	if accessCtx.Resource.Sensitivity == "confidential" || accessCtx.Resource.Sensitivity == "restricted" {
		hour := accessCtx.Time.Hour()
		if hour < 8 || hour > 18 {
			// 检查是否有加班权限
			if accessCtx.User.Attributes["overtime_allowed"] != "true" {
				return &AccessDecision{
					Allowed:  false,
					Reason:   "非工作时间禁止访问敏感资源",
					Policies: []string{"time_restriction"},
				}
			}
		}
	}

	// 规则5: 位置限制规则 - 敏感资源只能在办公室访问
	if accessCtx.Resource.Sensitivity == "restricted" {
		if accessCtx.Location != "office" && accessCtx.Location != "secure_network" {
			return &AccessDecision{
				Allowed:  false,
				Reason:   "敏感资源仅允许在安全网络环境访问",
				Policies: []string{"location_restriction"},
			}
		}
	}

	return nil
}

// validateRequest 验证请求
func (acs *AccessControlService) validateRequest(req *AccessRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("用户ID不能为空")
	}
	if req.ResourceID == "" {
		return fmt.Errorf("资源ID不能为空")
	}
	if req.Action == "" {
		return fmt.Errorf("操作不能为空")
	}
	if req.RequestID == "" {
		return fmt.Errorf("请求ID不能为空")
	}
	return nil
}

// logAccess 记录访问日志
func (acs *AccessControlService) logAccess(req *AccessRequest, decision *AccessDecision, duration time.Duration) {
	if !acs.config.EnableAudit {
		return
	}

	go func() {
		log := &AccessLog{
			UserID:    req.UserID,
			Resource:  req.ResourceID,
			Action:    req.Action,
			Allowed:   decision.Allowed,
			Reason:    decision.Reason,
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
			RequestID: req.RequestID,
			Duration:  duration.Milliseconds(),
			CreatedAt: time.Now(),
		}

		if err := acs.auditLogger.Log(log); err != nil {
			acs.logger.With("error", err, "request_id", req.RequestID).Warn("记录访问日志失败")
		}
	}()
}

// CacheKey 返回缓存键
func (req *AccessRequest) CacheKey() string {
	return fmt.Sprintf("perm:%s:%s:%s:%s", req.UserID, req.ResourceID, req.Action, req.ContextHash())
}

// ContextHash 返回上下文哈希
func (req *AccessRequest) ContextHash() string {
	// 简化实现，实际应该使用真正的哈希
	return fmt.Sprintf("%x", len(req.Context))
}

// loadDefaultPolicies 加载默认策略
func (acs *AccessControlService) loadDefaultPolicies() error {
	// 清除现有策略
	acs.enforcer.ClearPolicy()

	// 添加默认角色
	defaultRoles := []struct {
		role        string
		description string
	}{
		{"super_admin", "超级管理员"},
		{"admin", "管理员"},
		{"partner", "合伙人"},
		{"senior_lawyer", "高级律师"},
		{"lawyer", "律师"},
		{"assistant", "助理"},
		{"client", "客户"},
		{"guest", "访客"},
	}

	// 添加默认权限策略
	defaultPolicies := [][]string{
		// 超级管理员权限
		{"super_admin", "*", "*", "allow"},
		// 管理员权限
		{"admin", "user_management", "*", "allow"},
		{"admin", "role_management", "*", "allow"},
		{"admin", "policy_management", "*", "allow"},
		{"admin", "audit_log", "read", "allow"},
		// 合伙人权限
		{"partner", "case_management", "*", "allow"},
		{"partner", "client_management", "*", "allow"},
		{"partner", "document_management", "*", "allow"},
		{"partner", "billing_management", "*", "allow"},
		{"partner", "report_access", "*", "allow"},
		// 高级律师权限
		{"senior_lawyer", "case_management", "read", "allow"},
		{"senior_lawyer", "case_management", "write", "allow"},
		{"senior_lawyer", "document_management", "*", "allow"},
		{"senior_lawyer", "client_communication", "*", "allow"},
		// 律师权限
		{"lawyer", "assigned_cases", "read", "allow"},
		{"lawyer", "assigned_cases", "write", "allow"},
		{"lawyer", "document_management", "read", "allow"},
		{"lawyer", "document_management", "write", "allow"},
		{"lawyer", "client_communication", "*", "allow"},
		// 助理权限
		{"assistant", "assigned_cases", "read", "allow"},
		{"assistant", "document_management", "read", "allow"},
		{"assistant", "document_upload", "*", "allow"},
		// 客户权限
		{"client", "own_cases", "read", "allow"},
		{"client", "own_documents", "read", "allow"},
		{"client", "communication", "*", "allow"},
	}

	// 批量添加策略
	if _, err := acs.enforcer.AddPolicies(defaultPolicies); err != nil {
		return fmt.Errorf("添加默认策略失败: %w", err)
	}

	// 添加默认用户角色映射
	defaultUserRoles := [][]string{
		{"admin001", "super_admin"},
		{"partner001", "partner"},
		{"lawyer001", "senior_lawyer"},
		{"lawyer002", "lawyer"},
		{"assistant001", "assistant"},
	}

	if _, err := acs.enforcer.AddGroupingPolicies(defaultUserRoles); err != nil {
		return fmt.Errorf("添加默认用户角色映射失败: %w", err)
	}

	// 保存策略到数据库
	if err := acs.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("保存策略到数据库失败: %w", err)
	}

	acs.logger.Info("默认策略加载完成",
		"policies_count", len(defaultPolicies),
		"user_roles_count", len(defaultUserRoles),
	)

	return nil
}

// autoMigrate 自动迁移数据库表
func (acs *AccessControlService) autoMigrate() error {
	tables := []interface{}{
		&AccessLog{},
		&User{},
		&Role{},
		&Resource{},
		&Permission{},
	}

	for _, table := range tables {
		if err := acs.db.AutoMigrate(table); err != nil {
			return fmt.Errorf("迁移表 %T 失败: %w", table, err)
		}
	}

	acs.logger.Info("数据库表迁移完成")
	return nil
}