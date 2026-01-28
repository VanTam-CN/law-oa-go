# Findings & Decisions - 案件审批流程闭环优化

## Requirements
<!-- 从用户请求中捕获的需求 -->
- 实现案件审批流程的完整闭环
- 审批流程应包括：提交、审批、驳回、重新提交、通过/拒绝
- 每个状态转换都应有清晰的业务逻辑
- 用户应获得及时的状态反馈
- 需要权限控制（谁可以审批）
- 需要审批历史记录

## Research Findings
<!-- 探索过程中的关键发现 -->

### 已探索领域
- [x] 后端Go代码：审批相关handlers, services, models
- [x] 前端React代码：审批UI组件和API调用
- [x] 数据库schema：审批状态表设计
- [x] 当前状态机实现
- [x] 权限控制逻辑

### 核心发现

#### 1. 审批状态定义 (approval_models.go:313-321)
```go
const (
    ApprovalStatusDraft        = "draft"
    ApprovalStatusSubmitted    = "submitted"
    ApprovalStatusUnderReview  = "under_review"
    ApprovalStatusApproved     = "approved"
    ApprovalStatusRejected     = "rejected"
    ApprovalStatusCancelled    = "cancelled"
    ApprovalStatusExpired      = "expired"
)
```

#### 2. 审批决定类型 (approval_models.go:324-331)
```go
const (
    ApprovalDecisionApprove        = "approve"
    ApprovalDecisionReject         = "reject"
    ApprovalDecisionRequestChanges = "request_changes"
    ApprovalDecisionDefer          = "defer"
    ApprovalDecisionEscalate       = "escalate"
    ApprovalDecisionReassign       = "reassign"
)
```

#### 3. 当前存在的问题

##### 问题1: 审批流程不完整 - 驳回后无法重新提交
**位置**: `internal/services/approval_service.go:288-397`
- 当前 `ProcessApproval` 函数在拒绝后直接将状态设为 `rejected`
- 没有机制让申请人修改后重新提交
- `request_changes` 决策存在但未实现完整的重新提交流程

##### 问题2: 缺少审批人分配逻辑
**位置**: `internal/services/approval_service.go:162-190`
- `CreateApproval` 函数创建申请时，`CurrentApproverID` 为空
- 没有自动分配审批人的逻辑
- 工作流配置存在但未实际使用

##### 问题3: 多级审批未实现
**位置**: `internal/services/approval_service.go:362-369`
- 审批通过后直接设置为 `approved` 状态
- 没有检查是否需要下一级审批
- 工作流阶段 (`current_stage`) 没有推进逻辑

##### 问题4: 驳回理由未被有效传递给申请人
**位置**: `internal/handlers/approval_handler.go`
- 前端 `handleApproval` 函数只是模拟返回成功
- 没有实际的 API 端点处理审批决定

##### 问题5: 缺少审批历史完整记录
**位置**: 数据库 `approval_records` 表
- 审批记录表存在，但前端没有完整展示
- 无法看到完整的审批链路和每个环节的决定理由

##### 问题6: 没有撤回功能的API端点
**位置**: `internal/router/router.go:327-337`
- 前端有 `cancelApproval` 调用，但后端路由没有对应的撤回端点
- `approval_handler.go` 没有实现 `CancelApproval` 处理器方法

##### 问题7: 权限校验不完整
**位置**: `internal/services/approval_service.go:133-135`
- 只检查 `approval.CurrentApproverID == userID`
- 没有基于角色的权限控制
- 没有检查审批人是否在职、是否有权限审批该类型申请

##### 问题8: 前端审批操作是模拟的
**位置**: `frontend/src/services/approval.ts:215-226`
```typescript
export const handleApproval = (
  id: string,
  action: 'approve' | 'reject',
  comment: string,
): Promise<any> => {
  // 开发环境返回模拟数据
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({ success: true, message: `审批${action === 'approve' ? '通过' : '拒绝'}成功` })
    }, 300)
  })
}
```

##### 问题9: 状态转换逻辑不一致
- `UpdateApproval` 只允许更新草稿状态
- 但 `CancelApproval` 允许撤回已提交和审核中的状态
- 缺少明确的状态机图和转换规则

##### 问题10: 通知机制未实现
- `approval_notifications` 表存在但没有实际使用
- 审批状态变更时没有通知相关人员
- 待审批任务没有提醒机制

## Technical Decisions
<!-- 技术决策及其理由 -->
| Decision | Rationale |
|----------|-----------|
|          |           |

## Issues Encountered
<!-- 遇到的问题和解决方法 -->
| Issue | Resolution |
|-------|------------|
|       |            |

## Resources
<!-- URLs, 文件路径, API参考 -->
- 项目根目录: /Users/mac/Desktop/FT/law-oa-go
- 后端审批模型: `internal/models/approval_models.go`
- 后端审批服务: `internal/services/approval_service.go`
- 后端审批处理器: `internal/handlers/approval_handler.go`
- 后端路由配置: `internal/router/router.go`
- 前端审批服务: `frontend/src/services/approval.ts`
- 前端审批列表: `frontend/src/pages/approval/ApprovalList.tsx`
- 前端审批详情: `frontend/src/pages/approval/ApprovalDetail.tsx`
- 数据库迁移: `migrations/000019_approval_system.up.sql`

## Visual/Browser Findings
<!-- 视觉/浏览器发现的信息 -->

---
*定期更新此文件，特别是在每2个view/browser/search操作后*
