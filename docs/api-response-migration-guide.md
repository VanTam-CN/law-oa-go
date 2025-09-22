# API响应格式迁移指南

## 概述

本指南说明如何从旧版API响应格式迁移到新的统一格式。

## 响应格式对比

### 旧版格式（即将废弃）
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {...}
}
```

### 新版统一格式（推荐）
```json
{
  "success": true,
  "data": {...},
  "error": null,
  "meta": {
    "timestamp": "2025-01-15T10:30:00Z",
    "request_id": "req-123",
    "version": "v1",
    "server": "law-oa-go",
    "environment": "development"
  }
}
```

## 迁移步骤

### 1. 导入新的响应包
```go
import "law-oa-go/internal/common"
```

### 2. 替换响应函数调用

#### 成功响应
**旧版:**
```go
common.Success(c, data)
common.SuccessWithPage(c, data, total, page, size)
```

**新版:**
```go
common.APISuccess(c, data)
common.APISuccessWithPage(c, data, total, page, size)
```

#### 错误响应
**旧版:**
```go
common.Error(c, code, message)
common.BadRequest(c, message)
common.NotFound(c, message)
```

**新版:**
```go
common.APIError(c, statusCode, errCode, message, details...)
common.APIBadRequest(c, message, details...)
common.APINotFound(c, message, details...)
```

### 3. 使用增强功能

#### 带上下文的错误响应
```go
context := map[string]interface{}{
    "field_errors": fieldErrors,
    "user_id":      userID,
}
common.APIErrorWithContext(c, http.StatusBadRequest, "VALIDATION_ERROR", "验证失败", context)
```

#### 带建议的错误响应
```go
suggestions := []string{
    "检查用户名格式",
    "确认密码强度",
    "验证邮箱地址",
}
common.APIErrorWithSuggestions(c, http.StatusBadRequest, "VALIDATION_ERROR", "注册失败", suggestions)
```

### 4. 分页响应更新
**旧版:**
```go
common.SuccessWithPage(c, data, total, page, size)
```

**新版:**
```go
common.APISuccessWithPage(c, data, total, page, size)
```

## 向后兼容性

为了确保平滑迁移，所有旧版函数仍然可用，但建议在新代码中使用新版API函数。

## 示例

### 控制器示例
```go
// 旧版写法
func (h *UserHandler) GetUsers(c *gin.Context) {
    users, total, err := h.userService.ListUsers(c)
    if err != nil {
        common.Error(c, 500, "获取用户列表失败")
        return
    }
    common.SuccessWithPage(c, users, total, page, size)
}

// 新版写法
func (h *UserHandler) GetUsers(c *gin.Context) {
    users, total, err := h.userService.ListUsers(c)
    if err != nil {
        common.APIInternalServerError(c, "获取用户列表失败", err.Error())
        return
    }
    common.APISuccessWithPage(c, users, total, page, size)
}
```

### 错误处理示例
```go
// 旧版写法
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.ValidationError(c, "参数验证失败")
        return
    }
    // ... 处理逻辑
}

// 新版写法
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        fieldErrors := make(map[string]string)
        // 填充字段错误信息
        common.APIValidationError(c, "参数验证失败", fieldErrors)
        return
    }
    // ... 处理逻辑
}
```

## 响应格式优势

1. **一致性**: 所有API使用统一的响应格式
2. **丰富信息**: 包含请求ID、时间戳等元数据
3. **错误处理**: 支持错误上下文和建议
4. **分页支持**: 内置分页信息格式
5. **可追踪性**: 通过request_id追踪请求

## 客户端适配

### JavaScript示例
```javascript
// 旧版响应处理
axios.get('/api/users')
  .then(response => {
    if (response.data.code === 200) {
      console.log(response.data.data);
    }
  });

// 新版响应处理
axios.get('/api/users')
  .then(response => {
    if (response.data.success) {
      console.log(response.data.data);
      console.log('Request ID:', response.data.meta.request_id);
    } else {
      console.error('Error:', response.data.error.message);
      if (response.data.error.suggestions) {
        console.log('Suggestions:', response.data.error.suggestions);
      }
    }
  });
```

## 测试建议

1. **单元测试**: 更新所有API测试以适配新格式
2. **集成测试**: 确保客户端能正确处理新格式
3. **性能测试**: 验证新格式不影响性能
4. **兼容性测试**: 确保旧客户端仍能正常工作

## 部署计划

1. **阶段1**: 部署新响应格式，保持向后兼容
2. **阶段2**: 逐步更新现有API使用新格式
3. **阶段3**: 在新版本中废弃旧格式
4. **阶段4**: 移除旧格式代码

## 注意事项

- 新API函数会自动添加时间戳和版本信息
- 建议在生产环境中使用环境变量配置version和environment
- 错误代码建议使用大写字母和下划线
- 详细的错误信息应放在details字段中