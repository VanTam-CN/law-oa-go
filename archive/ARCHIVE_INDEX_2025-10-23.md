# 项目文件归档索引

**归档日期**: 2025-10-23
**归档目的**: 清理项目根目录，移除不需要的文件，保持项目结构整洁
**归档人员**: AI助手
**总清理文件数**: 700+ 文件
**节省磁盘空间**: 300MB+

## 归档目录结构

```
archive/
├── chrome-tests/          # Chrome测试相关文件 (170MB)
├── standalone-servers/    # 独立测试服务器
├── test-files/           # 各种测试和调试文件
├── reports/              # 修复报告和测试报告
├── logs/                 # 日志和临时数据文件
├── temporary-files/      # 临时配置和脚本文件
├── binaries/             # 已存在的二进制文件归档
├── configs/              # 已存在的配置文件归档
└── ARCHIVE_INDEX_2025-10-23.md  # 本索引文件
```

## 各目录详细说明

### 📁 chrome-tests/
- **大小**: ~170MB
- **内容**: Chrome测试浏览器和DevTools验证项目
- **包含**: `chrome/`, `chrome-devtools-validation/`
- **归档原因**: 对核心功能不必要，占用大量空间

### 📁 standalone-servers/
- **大小**: ~50KB
- **内容**: 独立的冲突检测服务器和演示
- **包含**: `standalone_conflict_server/`, `standalone_conflict_demo.go`
- **归档原因**: 独立测试项目，不是核心系统一部分

### 📁 test-files/
- **大小**: ~5MB
- **内容**: 各种测试、调试和演示文件
- **包含**: `test_*.go`, `debug_*.go`, `test_*.js`, `test_*.html` 等
- **归档原因**: 临时测试文件，与核心代码分离

### 📁 reports/
- **大小**: ~2MB
- **内容**: 修复报告、测试报告、临时文档
- **包含**: `*_FIX_REPORT.md`, `*_COMPLETE.md`, `*_TEST_REPORT.md` 等
- **归档原因**: 历史记录，不是核心文档

### 📁 logs/
- **大小**: ~100MB+
- **内容**: 日志文件、进程文件、临时数据
- **包含**: `*.log`, `*.pid`, `token.txt`, 各种JSON数据文件
- **归档原因**: 运行时产生的临时文件

### 📁 temporary-files/
- **大小**: ~10MB
- **内容**: 多余的配置文件、脚本、二进制文件
- **包含**: 多余的docker-compose文件、启动脚本、构建产物
- **归档原因**: 临时或重复的配置文件

## 保留的核心文件

### 🔧 核心应用文件
- `main.go` - 应用入口
- `go.mod`, `go.sum` - Go模块定义
- `internal/` - 核心业务逻辑
- `cmd/` - 命令行工具
- `frontend/` - 前端应用

### 📋 核心配置
- `docker-compose.yml` - 主要Docker编排
- `Dockerfile` - 主要容器配置
- `.env.example` - 环境变量示例
- `config/` - 配置目录

### 📚 核心文档
- `README.md` - 项目说明
- `CLAUDE.md` - AI助手指导
- `docs/` - 正式文档目录

## 恢复指南

### 如果需要恢复特定文件
1. **Chrome测试**: 从 `chrome-tests/` 恢复
2. **独立服务器**: 从 `standalone-servers/` 恢复
3. **测试文件**: 从 `test-files/` 恢复
4. **历史报告**: 从 `reports/` 查看
5. **日志数据**: 从 `logs/` 查看

### 如果需要完全恢复
1. 使用 `git checkout <commit-hash>` 恢复到特定提交
2. 从git历史查找被删除的文件
3. 根据需要重新创建配置

## 后续维护建议

### 定期清理
- 每月检查是否有新的临时文件需要归档
- 定期清理日志文件
- 及时归档完成的修复报告

### 归档规范
- 新的归档应该更新本索引文件
- 每个归档目录都应该有README.md说明
- 重要文件应该在git中有历史记录

### 注意事项
- 不要归档核心业务文件
- 重要的配置变更应该提交到git
- 保持归档结构的清晰和一致

---

**本次归档总结**: 成功清理了700+个文件，节省了300MB+的磁盘空间，显著改善了项目结构的整洁度和可维护性。