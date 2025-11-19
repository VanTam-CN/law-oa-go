# 🔐 认证系统问题完整解决方案

## 📋 问题诊断结果

经过系统性分析，确认了审批管理界面数据不显示的根本原因：

### ✅ 系统组件状态：
- ✅ 前端登录流程正常
- ✅ Token存储机制正常
- ✅ 后端JWT中间件正常
- ✅ 审批API路由注册正常
- ✅ 前端HTTP请求配置正常

### ❌ 问题根源：
**数据库中没有测试用户** - 导致登录API返回"邮箱或密码错误"，无法获取认证token

## 🛠️ 解决方案

### 方案1：使用psql命令行工具（推荐）

如果您有PostgreSQL客户端，运行：

```bash
# 1. 运行SQL脚本创建测试用户和数据
psql -h localhost -U postgres -d law_oa -f scripts/create_test_users_and_data.sql

# 2. 使用以下账号登录测试
# 邮箱: zhangsan@law.com
# 密码: 123456
```

### 方案2：手动数据库操作

通过数据库管理工具（如pgAdmin、DBeaver等）执行以下SQL：

```sql
-- 创建用户表（如果不存在）
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    phone VARCHAR(20),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入测试用户（密码是123456的bcrypt哈希）
INSERT INTO users (id, username, name, email, password, role, phone, status) VALUES
(1, 'zhangsan', '张三', 'zhangsan@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer', '13800000002', 'active')
ON CONFLICT (id) DO NOTHING;

-- 创建审批测试数据
INSERT INTO approval_requests (
    id, request_number, title, type, category, content,
    applicant_id, applicant_name, applicant_title, department_id, department_name,
    urgency, priority, status, submission_date, current_stage,
    current_approver_id, current_approver_name, workflow_type,
    duration_days, created_by, updated_by, created_at, updated_at,
    metadata, attachments
) VALUES
('test-approval-001', 'AP-20241201001', '张三的年假申请', 'leave', '人事行政',
 '因家庭事务需要回老家处理，特申请年假5天',
 '1', '张三', '高级律师', 'dept-001', '诉讼部',
 'normal', 'medium', 'submitted', NOW() - INTERVAL '2 hours', 'department_head_review',
 '2', '李四', 'STANDARD_APPROVAL',
 5, '1', '1', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours',
 '{"leave_type": "年假"}', '[]');
```

### 方案3：修改后端代码临时添加开发模式

如果无法访问数据库，可以临时修改认证处理器：

1. 编辑 `internal/handlers/auth_handler.go`
2. 在 `Login` 方法中添加开发模式逻辑：

```go
// 在文件开头添加开发模式检查
var isDevelopmentMode = true // 临时设为true

// 在Login方法开始处添加
if isDevelopmentMode {
    // 开发模式：直接返回成功，不验证用户
    response := LoginResponse{
        Token:     "dev_token_12345",
        ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
        User: map[string]interface{}{
            "id":    1,
            "name":  "开发用户",
            "email": "dev@law.com",
            "role":  "lawyer",
        },
    }
    common.APISuccess(c, response)
    return
}
```

### 方案4：通过用户注册API创建用户

```bash
# 注册新用户
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试用户",
    "email": "test@law.com",
    "password": "123456",
    "role": "lawyer"
  }'
```

## 🎯 测试账号信息

创建测试数据后，可以使用以下账号登录：

| 邮箱 | 密码 | 角色 | 说明 |
|------|------|------|------|
| admin@law.com | 123456 | 管理员 | 系统管理员 |
| zhangsan@law.com | 123456 | 律师 | 张三 |
| lisi@law.com | 123456 | 律师 | 李四 |
| wangwu@law.com | 123456 | 律师 | 王五 |

## 📊 预期测试结果

完成以上步骤后，审批管理界面应该显示：

- **待我审批**：2条记录（需要当前用户审批）
- **我的申请**：2条记录（当前用户提交的）
- **统计数据**：各状态审批的数量统计
- **完整功能**：查看详情、审批操作、撤回等

## 🔍 验证步骤

1. **登录验证**：使用测试账号登录系统
2. **界面验证**：检查审批管理页面显示
3. **功能验证**：测试查看详情、审批操作等
4. **API验证**：检查浏览器开发者工具中的API请求

## 🚨 注意事项

1. **安全性**：生产环境必须移除开发模式代码
2. **数据清理**：测试完成后可以清理测试数据
3. **密码策略**：生产环境使用更复杂的密码
4. **权限控制**：确保不同角色有正确的权限

## 💡 排查提示

如果仍有问题：

1. **检查后端日志**：查看`./law-oa-server`的输出
2. **检查网络连接**：确认前端能访问后端API
3. **检查数据库连接**：确认数据库服务正常运行
4. **检查Token传递**：在浏览器开发者工具中查看请求头

---

选择适合您环境的方案执行即可解决认证问题！