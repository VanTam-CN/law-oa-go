# 🚨 PostgreSQL数据库紧急修复说明

## 问题诊断

你的截图显示PostgreSQL数据库确实只有基本表结构，缺少：
- lawyers 表（律师管理）
- departments 表（部门管理）
- documents 表（文档管理）
- 关键字段（案件编号、律师关联、合同金额等）

## 立即解决方案

### 方法1: 使用简单Bash脚本（推荐）

```bash
# 进入脚本目录
cd /Users/mac/Desktop/FT/law-oa-go/scripts

# 设置数据库连接参数（根据你的实际情况修改）
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_postgres_password
export DB_NAME=law_oa_go

# 执行修复脚本
chmod +x quick-fix.sh
./quick-fix.sh
```

### 方法2: 使用Go程序

```bash
# 进入脚本目录
cd /Users/mac/Desktop/FT/law-oa-go/scripts

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_postgres_password
export DB_NAME=law_oa_go

# 运行修复程序
go run simple-fix.go
```

### 方法3: 直接SQL修复（如果psql可用）

```bash
# 进入脚本目录
cd /Users/mac/Desktop/FT/law-oa-go/scripts

# 直接执行SQL（需要输入密码）
psql -h localhost -p 5432 -U postgres -d law_oa_go < postgresql-critical-fix.sql
```

## 修复内容

执行后将会添加：

### 📋 新增表
- **lawyers** - 律师管理表（12个字段）
- **departments** - 部门管理表（11个字段）
- **documents** - 文档管理表（20个字段）

### 📝 新增字段
**users表：**
- real_name (真实姓名)
- last_login_at (最后登录时间)
- last_login_ip (最后登录IP)
- role_id (角色ID)
- department_id (部门ID)
- remark (备注)

**clients表：**
- client_name (客户姓名)
- lawyer_id (负责律师ID)
- remark (备注)

**cases表：**
- case_no (案件编号)
- case_name (案件名称)
- assisting_lawyer_id (协办律师ID)
- contract_amount (合同金额)
- opponent_info (对方信息)
- remark (备注)

### 📊 示例数据
- 3个示例律师记录
- 相关索引和约束

## 验证步骤

1. **执行修复脚本**
2. **刷新数据库管理工具**
3. **检查新增表是否出现**
4. **测试Go应用程序连接**

## 常见问题

### Q: psql命令不存在
```bash
# macOS
brew install postgresql

# Ubuntu/Debian
sudo apt-get install postgresql-client
```

### Q: 连接被拒绝
- 检查PostgreSQL服务是否启动
- 验证连接参数是否正确
- 确认用户权限足够

### Q: Go程序运行失败
- 确保已安装PostgreSQL驱动：`go get github.com/lib/pq`
- 检查环境变量是否设置正确

## 预期结果

修复完成后，你的数据库管理工具应该显示：

- **68张表**（包括新增的lawyers、departments、documents等）
- **完整的字段结构**（每个表都有所有必需字段）
- **示例数据**（3个律师记录）
- **正确的索引**（提高查询性能）

## 联系支持

如果遇到问题：
1. 检查错误日志
2. 确认PostgreSQL服务状态
3. 验证网络连接
4. 检查防火墙设置

---

**重要：** 请选择以上任一方法立即执行修复，然后刷新数据库管理工具查看结果！