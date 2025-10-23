package security

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"
)

// NewAttributeStore 创建属性存储
func NewAttributeStore(db *gorm.DB, logger *slog.Logger) *AttributeStore {
	store := &AttributeStore{
		db:     db,
		logger: logger,
		cache:  make(map[string]map[string]string),
	}

	// 自动迁移表
	if err := store.autoMigrate(); err != nil {
		logger.With("error", err).Error("属性存储表迁移失败")
	}

	return store
}

// GetUser 获取用户信息
func (as *AttributeStore) GetUser(userID string) (*User, error) {
	// 检查缓存
	if cached, exists := as.cache[fmt.Sprintf("user:%s", userID)]; exists {
		user := &User{
			ID:         userID,
			Attributes: cached,
		}

		// 从缓存中解析其他字段
		if name, ok := cached["name"]; ok {
			user.FullName = name
		}
		if email, ok := cached["email"]; ok {
			user.Email = email
		}
		if department, ok := cached["department"]; ok {
			user.Department = department
		}
		if position, ok := cached["position"]; ok {
			user.Position = position
		}
		if active, ok := cached["active"]; ok {
			user.IsActive = active == "true"
		}

		// 解析角色
		if rolesStr, ok := cached["roles"]; ok {
			var roles []string
			if err := json.Unmarshal([]byte(rolesStr), &roles); err == nil {
				user.Roles = roles
			}
		}

		return user, nil
	}

	var user User
	if err := as.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %s", userID)
	}

	// 缓存用户信息
	userCache := make(map[string]string)
	for k, v := range user.Attributes {
		userCache[k] = v
	}
	userCache["name"] = user.FullName
	userCache["email"] = user.Email
	userCache["department"] = user.Department
	userCache["position"] = user.Position
	userCache["active"] = fmt.Sprintf("%t", user.IsActive)

	if rolesJSON, err := json.Marshal(user.Roles); err == nil {
		userCache["roles"] = string(rolesJSON)
	}

	as.cache[fmt.Sprintf("user:%s", userID)] = userCache

	return &user, nil
}

// UpdateUser 更新用户信息
func (as *AttributeStore) UpdateUser(user *User) error {
	// 更新数据库
	if err := as.db.Save(user).Error; err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:%s", user.ID)
	delete(as.cache, cacheKey)

	as.logger.Info("用户信息更新成功", "user_id", user.ID)
	return nil
}

// CreateUser 创建用户
func (as *AttributeStore) CreateUser(user *User) error {
	// 设置默认值
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	if user.Attributes == nil {
		user.Attributes = make(map[string]string)
	}

	// 保存到数据库
	if err := as.db.Create(user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	as.logger.Info("用户创建成功", "user_id", user.ID, "username", user.Username)
	return nil
}

// DeleteUser 删除用户
func (as *AttributeStore) DeleteUser(userID string) error {
	// 删除数据库记录
	if err := as.db.Delete(&User{}, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:%s", userID)
	delete(as.cache, cacheKey)

	as.logger.Info("用户删除成功", "user_id", userID)
	return nil
}

// ListUsers 列出用户
func (as *AttributeStore) ListUsers(filter map[string]interface{}) ([]*User, error) {
	var users []*User
	query := as.db.Model(&User{})

	// 应用过滤条件
	for key, value := range filter {
		switch key {
		case "department":
			query = query.Where("department = ?", value)
		case "position":
			query = query.Where("position = ?", value)
		case "is_active":
			query = query.Where("is_active = ?", value)
		case "role":
			// 对于角色过滤，需要查询JSON字段
			query = query.Where("roles @> ?", fmt.Sprintf(`["%s"]`, value))
		}
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	return users, nil
}

// GetResource 获取资源信息
func (as *AttributeStore) GetResource(resourceID string) (*Resource, error) {
	// 检查缓存
	if cached, exists := as.cache[fmt.Sprintf("resource:%s", resourceID)]; exists {
		resource := &Resource{
			ID:         resourceID,
			Attributes: cached,
		}

		// 从缓存中解析其他字段
		if name, ok := cached["name"]; ok {
			resource.Name = name
		}
		if type_, ok := cached["type"]; ok {
			resource.Type = type_
		}
		if owner, ok := cached["owner"]; ok {
			resource.Owner = owner
		}
		if sensitivity, ok := cached["sensitivity"]; ok {
			resource.Sensitivity = sensitivity
		}
		if category, ok := cached["category"]; ok {
			resource.Category = category
		}

		return resource, nil
	}

	var resource Resource
	if err := as.db.Where("id = ?", resourceID).First(&resource).Error; err != nil {
		return nil, fmt.Errorf("资源不存在: %s", resourceID)
	}

	// 缓存资源信息
	resourceCache := make(map[string]string)
	for k, v := range resource.Attributes {
		resourceCache[k] = v
	}
	resourceCache["name"] = resource.Name
	resourceCache["type"] = resource.Type
	resourceCache["owner"] = resource.Owner
	resourceCache["sensitivity"] = resource.Sensitivity
	resourceCache["category"] = resource.Category

	as.cache[fmt.Sprintf("resource:%s", resourceID)] = resourceCache

	return &resource, nil
}

// UpdateResource 更新资源信息
func (as *AttributeStore) UpdateResource(resource *Resource) error {
	// 更新数据库
	if err := as.db.Save(resource).Error; err != nil {
		return fmt.Errorf("更新资源失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("resource:%s", resource.ID)
	delete(as.cache, cacheKey)

	as.logger.Info("资源信息更新成功", "resource_id", resource.ID)
	return nil
}

// CreateResource 创建资源
func (as *AttributeStore) CreateResource(resource *Resource) error {
	// 设置默认值
	if resource.CreatedAt.IsZero() {
		resource.CreatedAt = time.Now()
	}
	resource.UpdatedAt = time.Now()

	if resource.Attributes == nil {
		resource.Attributes = make(map[string]string)
	}

	// 保存到数据库
	if err := as.db.Create(resource).Error; err != nil {
		return fmt.Errorf("创建资源失败: %w", err)
	}

	as.logger.Info("资源创建成功", "resource_id", resource.ID, "resource_name", resource.Name)
	return nil
}

// DeleteResource 删除资源
func (as *AttributeStore) DeleteResource(resourceID string) error {
	// 删除数据库记录
	if err := as.db.Delete(&Resource{}, "id = ?", resourceID).Error; err != nil {
		return fmt.Errorf("删除资源失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("resource:%s", resourceID)
	delete(as.cache, cacheKey)

	as.logger.Info("资源删除成功", "resource_id", resourceID)
	return nil
}

// ListResources 列出资源
func (as *AttributeStore) ListResources(filter map[string]interface{}) ([]*Resource, error) {
	var resources []*Resource
	query := as.db.Model(&Resource{})

	// 应用过滤条件
	for key, value := range filter {
		switch key {
		case "type":
			query = query.Where("type = ?", value)
		case "owner":
			query = query.Where("owner = ?", value)
		case "sensitivity":
			query = query.Where("sensitivity = ?", value)
		case "category":
			query = query.Where("category = ?", value)
		case "department":
			// 对于部门过滤，需要查询JSON属性字段
			query = query.Where("attributes @> ?", fmt.Sprintf(`{"department":"%s"}`, value))
		}
	}

	if err := query.Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("查询资源列表失败: %w", err)
	}

	return resources, nil
}

// SetUserAttribute 设置用户属性
func (as *AttributeStore) SetUserAttribute(userID, key, value string) error {
	// 获取现有用户
	user, err := as.GetUser(userID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 更新属性
	if user.Attributes == nil {
		user.Attributes = make(map[string]string)
	}
	user.Attributes[key] = value
	user.UpdatedAt = time.Now()

	// 保存到数据库
	if err := as.db.Model(user).Update("attributes", user.Attributes).Error; err != nil {
		return fmt.Errorf("更新用户属性失败: %w", err)
	}

	// 更新缓存
	cacheKey := fmt.Sprintf("user:%s", userID)
	if cached, exists := as.cache[cacheKey]; exists {
		cached[key] = value
		as.cache[cacheKey] = cached
	}

	return nil
}

// GetUserAttribute 获取用户属性
func (as *AttributeStore) GetUserAttribute(userID, key string) (string, error) {
	user, err := as.GetUser(userID)
	if err != nil {
		return "", fmt.Errorf("获取用户失败: %w", err)
	}

	if value, exists := user.Attributes[key]; exists {
		return value, nil
	}

	return "", fmt.Errorf("属性不存在: %s", key)
}

// SetResourceAttribute 设置资源属性
func (as *AttributeStore) SetResourceAttribute(resourceID, key, value string) error {
	// 获取现有资源
	resource, err := as.GetResource(resourceID)
	if err != nil {
		return fmt.Errorf("获取资源失败: %w", err)
	}

	// 更新属性
	if resource.Attributes == nil {
		resource.Attributes = make(map[string]string)
	}
	resource.Attributes[key] = value
	resource.UpdatedAt = time.Now()

	// 保存到数据库
	if err := as.db.Model(resource).Update("attributes", resource.Attributes).Error; err != nil {
		return fmt.Errorf("更新资源属性失败: %w", err)
	}

	// 更新缓存
	cacheKey := fmt.Sprintf("resource:%s", resourceID)
	if cached, exists := as.cache[cacheKey]; exists {
		cached[key] = value
		as.cache[cacheKey] = cached
	}

	return nil
}

// GetResourceAttribute 获取资源属性
func (as *AttributeStore) GetResourceAttribute(resourceID, key string) (string, error) {
	resource, err := as.GetResource(resourceID)
	if err != nil {
		return "", fmt.Errorf("获取资源失败: %w", err)
	}

	if value, exists := resource.Attributes[key]; exists {
		return value, nil
	}

	return "", fmt.Errorf("属性不存在: %s", key)
}

// ClearCache 清除缓存
func (as *AttributeStore) ClearCache() {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.cache = make(map[string]map[string]string)
	as.logger.Info("属性缓存已清空")
}

// GetCacheStats 获取缓存统计信息
func (as *AttributeStore) GetCacheStats() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()

	totalEntries := len(as.cache)
	userEntries := 0
	resourceEntries := 0

	for key := range as.cache {
		if len(key) > 5 && key[:5] == "user:" {
			userEntries++
		} else if len(key) > 9 && key[:9] == "resource:" {
			resourceEntries++
		}
	}

	return map[string]interface{}{
		"total_entries":     totalEntries,
		"user_entries":      userEntries,
		"resource_entries":  resourceEntries,
		"cache_keys":        len(as.cache),
	}
}

// autoMigrate 自动迁移数据库表
func (as *AttributeStore) autoMigrate() error {
	tables := []interface{}{
		&User{},
		&Resource{},
		&Role{},
		&Permission{},
	}

	for _, table := range tables {
		if err := as.db.AutoMigrate(table); err != nil {
			return fmt.Errorf("迁移表 %T 失败: %w", table, err)
		}
	}

	as.logger.Info("属性存储表迁移完成")
	return nil
}