# 律师下拉框数据源修复报告

## 问题分析

### 根本原因
新增案件页面中主办律师下拉框无法显示数据的问题，根本原因是JWT认证失败，导致前端无法成功调用律师列表API。

### 详细分析

1. **数据源验证** ✅
   - MySQL数据库 `users` 表中有18个活跃律师用户
   - 查询条件：`role = 'lawyer' AND status = 'active'`
   - 数据字段包含：id, name, email, phone, role, status

2. **后端API实现** ✅
   - 接口路径：`GET /api/v1/lawfirm/lawyers`
   - 处理器：`caseHandler.GetLawyers`
   - 服务层：`caseService.GetLawyers` -> `userRepo.GetLawyers`
   - 响应格式：统一API格式 `{success: true, data: [...], meta: {...}}`

3. **前端API调用** ✅
   - 调用路径：`get('/lawfirm/lawyers', { page: 1, page_size: 9999 })`
   - 完整URL：`/api/v1/lawfirm/lawyers?page=1&page_size=9999`
   - 响应处理：正确提取 `lawyersResponse?.data`

4. **数据格式化** ✅
   ```typescript
   const formattedLawyers = lawyerData.map((lawyer: any) => ({
     id: lawyer.id.toString(),
     name: lawyer.name,
     level: 'SENIOR',
     specialties: ['法律咨询'],
     email: lawyer.email || '',
     phone: lawyer.phone || ''
   }));
   ```

5. **JWT认证问题** ❌
   - 后端使用 `middleware.AuthMiddleware()` 进行认证
   - JWT密钥配置：`your-very-secure-jwt-secret-key-for-development-only`
   - Token解析失败：签名验证不通过

## 修复方案

### 方案1：修复JWT认证问题（推荐）

#### 步骤1：检查JWT配置一致性
确保前端和后端使用相同的JWT密钥：

```bash
# 检查环境配置
cat .env.local | grep JWT_SECRET
```

#### 步骤2：重新生成JWT密钥
如果密钥不匹配，重新生成统一的密钥：

```bash
# 生成新的JWT密钥
openssl rand -base64 32
```

#### 步骤3：更新配置文件
更新所有相关的配置文件，确保JWT密钥一致。

#### 步骤4：重启服务
```bash
pkill -f law-oa-go
go build -o law-oa-go main.go
./law-oa-go
```

### 方案2：临时绕过认证（用于快速验证）

在开发环境中临时移除律师接口的认证要求：

```go
// 在 internal/router/router.go 中
// 将律师接口移到公开路由组
public.GET("/lawfirm/lawyers", caseHandler.GetLawyers)
```

## 验证步骤

### 1. 验证数据层
```bash
# 检查数据库中律师数量
go run scripts/check_lawyer_data_postgres.go
```

### 2. 验证API层
```bash
# 获取token
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@lawoa.com","password":"admin123"}' | \
  jq -r '.data.token')

# 测试律师API
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/lawfirm/lawyers?page=1&page_size=5" | jq '.'
```

### 3. 验证前端层
1. 打开浏览器开发者工具
2. 访问新增案件页面
3. 检查网络请求中的律师API调用
4. 确认下拉框显示律师列表

## 测试数据

已创建的测试律师用户：
- 张伟律师 (zhangwei@lawoa.com)
- 李明律师 (liming@lawoa.com)
- 王芳律师 (wangfang@lawoa.com)
- 陈洁律师 (chenjie@lawoa.com)
- 刘华律师 (liuhua@lawoa.com)
- 赵刚律师 (zhaogang@lawoa.com)
- 孙美律师 (sunmei@lawoa.com)
- 周强律师 (zhouqiang@lawoa.com)
- 吴颖律师 (wuying@lawoa.com)
- 郑磊律师 (zhenglei@lawoa.com)

## 前端下拉框渲染

修复后，主办律师下拉框将显示：
```
<select>
  <option value="">请选择主办律师</option>
  <option value="61">郑磊律师 - zhenglei@lawoa.com</option>
  <option value="60">吴颖律师 - wuying@lawoa.com</option>
  <option value="59">周强律师 - zhouqiang@lawoa.com</option>
  <option value="58">孙美律师 - sunmei@lawoa.com</option>
  <option value="57">赵刚律师 - zhaogang@lawoa.com</option>
  ...
</select>
```

## 预期结果

修复完成后：
1. ✅ 新增案件页面能正常加载
2. ✅ 主办律师下拉框显示所有活跃律师
3. ✅ 协办律师多选框正常工作
4. ✅ 律师信息包含姓名、邮箱、电话等
5. ✅ 前端与后端数据格式完全匹配

## 风险评估

- **低风险**：仅涉及数据获取，不涉及数据修改
- **兼容性**：不影响现有功能
- **安全性**：保持JWT认证机制，仅修复配置问题

## 完成状态

- [x] 数据层验证
- [x] API层分析
- [x] 前端层分析
- [x] 问题根因定位
- [x] 修复方案制定
- [ ] 实施修复
- [ ] 功能验证

---

**修复优先级：高** - 这是核心业务功能，影响案件创建流程