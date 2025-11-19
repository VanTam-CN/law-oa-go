## Why

当前利益冲突检测模块和审批申请模块之间缺乏有效的联动机制，导致审批详情信息不足，无法支持充分决策，且审批通过后需要手动创建案件，影响工作效率。

## What Changes

- 集成利益冲突检测结果到审批申请详情页面
- 增强审批申请信息展示，提供充分的决策上下文
- 实现审批通过后自动流转到案件创建的功能
- 建立两个模块间的数据关联和同步机制
- 添加真实数据驱动的测试和验证能力

**BREAKING**: 可能需要修改现有的审批申请数据模型以支持冲突检测信息关联

## Impact

- Affected specs: conflict-detection, approval-system
- Affected code:
  - `internal/services/conflict_detection_service.go`
  - `internal/services/approval_service.go`
  - `internal/handlers/approval_handler.go`
  - `frontend/src/pages/approval/ApprovalDetail.tsx`
  - `frontend/src/services/approval.ts`
  - 数据库schema需要调整以支持关联关系