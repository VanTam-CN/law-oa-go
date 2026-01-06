# 项目文件清理计划

## 清理日期: 2025-01-06

---

## 一、可执行文件 (约 520MB)

这些是重复编译的二进制文件，只需要保留一个：

| 文件名 | 大小 | 类型 | 操作 |
|--------|------|------|------|
| app | 74.5MB | Mach-O executable | 备份后删除 |
| law-oa-go-final | 74.4MB | Mach-O executable | 备份后删除 |
| law-oa-go-fixed | 74.4MB | Mach-O executable | 备份后删除 |
| law-oa-server | 74.6MB | Mach-O executable | 保留 |
| law-oa-server-clean | 74.6MB | Mach-O executable | 备份后删除 |
| law-oa-server-debug | 74.6MB | Mach-O executable | 备份后删除 |
| law-oa-server-final | 74.6MB | Mach-O executable | 备份后删除 |
| law-oa-server-final-fixed | 74.6MB | Mach-O executable | 备份后删除 |
| law-oa-server-fixed | 74.6MB | Mach-O executable | 备份后删除 |
| law-oa-server-fixed-final | 74.6MB | Mach-O executable | 备份后删除 |
| law-oa-server-new | 74.6MB | Mach-O executable | 备份后删除 |
| law-oa-server-ultimate-fix | 74.6MB | Mach-O executable | 备份后删除 |
| main | 74.4MB | Mach-O executable | 备份后删除 |

**可释放空间**: 约 520MB

---

## 二、备份文件 (.bak)

| 文件 | 原因 |
|------|------|
| _test_api.go.bak | 旧的测试文件备份 |
| _test_approval_system.go.bak | 旧的测试文件备份 |
| _test_approval_stats.go.bak | 旧的测试文件备份 |
| frontend/src/assets/styles/components.css.bak | 样式文件备份 |
| frontend/src/pages/lawyer/LawyerManagementDebug.tsx.bak | 调试组件备份 |
| internal/metrics/monitor_service.go.bak | 监控服务备份 |
| internal/metrics/business_monitor.go.bak | 业务监控备份 |
| internal/middleware/error_handler.go.bak | 错误处理备份 |
| internal/middleware/signature.go.bak | 签名中间件备份 |
| internal/rbac/initializer.go.bak | RBAC初始化备份 |
| internal/cache/cached_service.go.bak | 缓存服务备份 |
| internal/config/config_legacy.go.bak | 旧配置备份 |
| internal/config/test_config.go.bak | 测试配置备份 |
| internal/adapters/conflict_waiver_adapter.go.bak | 冲突豁免适配器备份 |
| internal/errors/errors_legacy.go.bak | 旧错误处理备份 |
| internal/handlers/waiver_approval_handler.go.bak | 豁免审批处理备份 |
| internal/handlers/auth_handler.go.bak | 认证处理备份 |
| internal/handlers/case_handler.go.bak | 案件处理备份 |
| internal/handlers/approval_handler.go.bak | 审批处理备份 |
| internal/services/conflict_waiver_service.go.bak | 冲突豁免服务备份 |
| internal/services/waiver_approval_service.go.bak | 豁免审批服务备份 |
| internal/services/user_service.go.bak | 用户服务备份 |
| _debug_approval_data.go.bak | 调试数据备份 |

---

## 三、日志文件 (约 14MB)

| 文件 | 大小 | 说明 |
|------|------|------|
| backend.log | 6.5MB | 后端日志 |
| go_server.log | 593KB | Go服务器日志 |
| server-fixed.log | 44KB | 修复服务器日志 |
| server_debug.log | 38KB | 调试服务器日志 |
| frontend.log | 49B | 前端日志 |
| performance_test_results.log | 36KB | 性能测试结果 |
| unit_test_results.log | 218KB | 单元测试结果 |

---

## 四、诊断/测试 HTML 文件 (约 200KB)

| 文件 | 说明 |
|------|------|
| check_browser_storage.html | 浏览器存储检查 |
| check_frontend_token.html | 前端令牌检查 |
| debug_frontend_params.js | 前端参数调试 |
| debug_user_id.html | 用户ID调试 |
| fix_user_id_issue.html | 用户ID问题修复 |
| set_token_and_test.html | 设置令牌测试 |
| test_api_login.js | API登录测试 |
| test_approval_auth_fix.js | 审批认证修复测试 |
| test_approval_data.html | 审批数据测试 |
| test_approval_detail_api.js | 审批详情API测试 |
| test_conflict_api_fix.js | 冲突API修复测试 |
| test_conflict_fix_verification.js | 冲突修复验证 |
| test_conflict_with_auth.js | 带认证的冲突测试 |
| test_frontend_conflict_fix.js | 前端冲突修复测试 |
| test_frontend_fix.html | 前端修复测试 |
| test_frontend_integration.js | 前端集成测试 |
| test_frontend_service_fix.js | 前端服务修复测试 |
| test_integration.html | 集成测试 |
| test_real_scenario.js | 真实场景测试 |
| verify_frontend_fix.js | 前端修复验证 |
| verify_postgres_data | PostgreSQL数据验证 (7MB) |
| waiver_approval_test.html | 豁免审批测试 |

---

## 五、临时文件

| 文件 | 说明 |
|------|------|
| current_token.txt | 当前令牌 |
| new_token.txt | 新令牌 |
| .backend.pid | 后端进程ID |
| .frontend.pid | 前端进程ID |
| law_oa.db | 空数据库文件 |

---

## 六、启动脚本 (保留主要的)

| 文件 | 说明 | 操作 |
|------|------|------|
| run.sh | 主启动脚本 | 保留 |
| dev.sh | 开发启动脚本 | 保留 |
| run_backend.sh | 后端启动脚本 | 保留 |
| rebuild_frontend.sh | 前端重建脚本 | 保留 |
| start.sh | 启动脚本(与run.sh重复) | 备份后删除 |
| start-final.sh | 最终启动脚本 | 备份后删除 |
| test_login.sh | 登录测试 | 备份 |
| test_approval_fix.sh | 审批修复测试 | 备份 |
| test_approval_system.sh | 审批系统测试 | 备份 |
| test_complete_approval_flow.sh | 完整审批流测试 | 备份 |

---

## 总计可释放空间: 约 750MB
