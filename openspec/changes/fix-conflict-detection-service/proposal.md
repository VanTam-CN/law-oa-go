# 利益冲突检测服务修复提案

## Why
修复新建案件中利益冲突检测功能失效问题。当前冲突检测服务被设置为nil，导致系统始终返回"无发现冲突"的模拟响应，无法进行真实的利益冲突检测。

## What Changes
- **核心修复**: 在路由初始化中正确配置冲突检测服务实例
- **模型补全**: 为User模型添加缺失的Department和Seniority字段
- **接口统一**: 修复仓储方法名不匹配问题
- **编译问题**: 解决所有编译错误，确保系统能够正常构建
- **服务集成**: 确保冲突检测服务能够正确集成到现有路由系统中

## Impact
- **受影响的规范**: conflict-detection
- **受影响的代码**:
  - `internal/router/router.go` (路由初始化)
  - `internal/models/models.go` (User模型)
  - `internal/services/conflict_detection_service.go` (服务调用修复)
  - `internal/handlers/conflict_handler_simple.go` (处理器验证)
  - `internal/repositories/user_repository.go` (仓储接口)
- **业务影响**: 恢复案件创建流程中的利益冲突检测功能，确保律师能够正确识别和规避利益冲突风险