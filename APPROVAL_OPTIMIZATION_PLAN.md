# 案件审批流程闭环优化方案

## 1. 状态机设计

### 1.1 审批状态定义

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          审批状态机设计                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─────────┐   提交    ┌───────────┐                                    │
│  │  draft  │ ────────> │ submitted │                                    │
│  │  草稿    │           │  已提交    │                                    │
│  └─────────┘           └─────┬─────┘                                    │
│                                │                                          │
│                         指派审批人 │                                          │
│                                ▼                                          │
│  ┌─────────┐  撤回     ┌───────────┐   需要修改   ┌──────────────┐      │
│  │cancelled│ <──────── │under_review│ <─────────── │rejected      │      │
│  │已撤回   │           │  审核中    │             │(可重新提交)   │      │
│  └─────────┘           └─────┬─────┘             └──────┬───────┘      │
│                                │                         │              │
│                          审批操作                   修改后重新提交        │
│                   (通过/拒绝/要求修改)                 │              │
│                                │                         │              │
│                 ┌───────────────┴─────────┬───────────┘              │
│                 │                         │                          │
│                 ▼                         ▼                          │
│         ┌───────────┐           ┌───────────┐                      │
│         │ approved  │           │  rejected │                      │
│         │  已通过    │           │  已拒绝   │                      │
│         └───────────┘           └───────────┘                      │
│                                                                           │
│  状态转换规则:                                                            │
│  1. draft → submitted: 用户提交申请                                      │
│  2. submitted → under_review: 审批人开始处理                             │
│  3. under_review → approved: 所有审批阶段通过                            │
│  4. under_review → rejected: 任一审批阶段拒绝                            │
│  5. rejected → submitted: 申请人修改后重新提交                           │
│  6. submitted/under_review → cancelled: 申请人撤回(无审批记录时)        │
│  7. under_review → under_review: 审批人转派/要求修改                    │
│                                                                           │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 新增状态常量

为支持完整的闭环，需要新增以下状态：

```go
const (
    // 现有状态
    ApprovalStatusDraft        = "draft"
    ApprovalStatusSubmitted    = "submitted"
    ApprovalStatusUnderReview  = "under_review"
    ApprovalStatusApproved     = "approved"
    ApprovalStatusRejected     = "rejected"
    ApprovalStatusCancelled    = "cancelled"

    // 新增状态
    ApprovalStatusNeedsRevision = "needs_revision" // 需要修改
    ApprovalStatusResubmitted   = "resubmitted"   // 重新提交
)
```

## 2. API接口设计

### 2.1 审批操作端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/approvals | 创建审批申请 |
| PUT | /api/v1/approvals/:id | 更新审批申请(草稿/需修改状态) |
| POST | /api/v1/approvals/:id/submit | 提交审批申请 |
| POST | /api/v1/approvals/:id/approve | 审批通过 |
| POST | /api/v1/approvals/:id/reject | 审批拒绝 |
| POST | /api/v1/approvals/:id/request-changes | 要求修改 |
| POST | /api/v1/approvals/:id/cancel | 撤回申请 |
| POST | /api/v1/approvals/:id/resubmit | 重新提交 |
| POST | /api/v1/approvals/:id/reassign | 转派审批人 |
| GET | /api/v1/approvals/:id/records | 获取审批历史 |

### 2.2 请求/响应格式

#### 2.2.1 审批决定请求

```typescript
interface ApprovalDecisionRequest {
  decision: 'approve' | 'reject' | 'request_changes' | 'reassign'
  decisionReason: string        // 必填：决定理由
  decisionComments?: string     // 可选：补充意见
  nextApproverId?: string       // 转派时使用
  requiredChanges?: string[]    // 要求修改时的具体要求
}
```

#### 2.2.2 重新提交请求

```typescript
interface ResubmitApprovalRequest {
  revisionNote: string          // 修改说明
  modifiedFields?: string[]     // 修改的字段列表
  attachments?: Attachment[]     // 新增的附件
}
```

## 3. 后端实现计划

### 3.1 审批状态机服务 (新增)

创建 `internal/services/approval_state_machine.go`：

```go
type ApprovalStateMachine struct {
    approvalRepo *repositories.ApprovalRepository
    userRepo     *repositories.UserRepository
}

// CanTransition 检查状态转换是否合法
func (s *ApprovalStateMachine) CanTransition(from, to string) bool {
    // 定义合法的状态转换
    transitions := map[string][]string{
        models.ApprovalStatusDraft:        {models.ApprovalStatusSubmitted, models.ApprovalStatusCancelled},
        models.ApprovalStatusSubmitted:    {models.ApprovalStatusUnderReview, models.ApprovalStatusCancelled},
        models.ApprovalStatusUnderReview:  {models.ApprovalStatusApproved, models.ApprovalStatusRejected, models.ApprovalStatusNeedsRevision, models.ApprovalStatusCancelled},
        models.ApprovalStatusRejected:     {models.ApprovalStatusResubmitted},
        models.ApprovalStatusNeedsRevision:{models.ApprovalStatusResubmitted},
        models.ApprovalStatusResubmitted:  {models.ApprovalStatusUnderReview},
    }
    // ...
}
```

### 3.2 审批人分配服务 (新增)

创建 `internal/services/approval_assigner.go`：

```go
type ApprovalAssigner struct {
    userRepo  *repositories.UserRepository
    workflowRepo *repositories.ApprovalWorkflowRepository
}

// AssignApprover 根据工作流自动分配审批人
func (s *ApprovalAssigner) AssignApprover(approval *models.ApprovalRequest) error {
    // 1. 获取工作流配置
    // 2. 根据申请类型、金额、部门等确定审批人
    // 3. 设置 current_approver_id 和 current_stage
}
```

### 3.3 需要修改的文件

| 文件 | 修改内容 |
|------|----------|
| `internal/services/approval_service.go` | 增强状态转换逻辑，添加重新提交支持 |
| `internal/handlers/approval_handler.go` | 添加新的处理方法 |
| `internal/router/router.go` | 添加新的路由 |
| `internal/models/approval_models.go` | 添加新的状态常量 |
| `internal/repositories/approval_repository.go` | 添加新的查询方法 |

## 4. 前端实现计划

### 4.1 需要修改的文件

| 文件 | 修改内容 |
|------|----------|
| `frontend/src/services/approval.ts` | 实现真实的API调用，添加重新提交方法 |
| `frontend/src/pages/approval/ApprovalDetail.tsx` | 添加重新提交、修改响应功能 |
| `frontend/src/pages/approval/CreateApproval.tsx | 添加草稿保存功能 |

### 4.2 新增组件

| 组件 | 说明 |
|------|------|
| `ApprovalActionButtons.tsx` | 统一的审批操作按钮组件 |
| `ApprovalTimeline.tsx` | 审批时间线组件 |
| `RevisionForm.tsx` | 修改并重新提交表单 |

## 5. 实施步骤

### Phase 3.1: 状态机和审批人分配 (后端)
- [ ] 创建 `approval_state_machine.go`
- [ ] 创建 `approval_assigner.go`
- [ ] 实现状态转换验证
- [ ] 实现自动审批人分配

### Phase 3.2: API端点完善 (后端)
- [ ] 添加 `ProcessApprovalDecision` 处理器方法
- [ ] 添加 `ResubmitApproval` 处理器方法
- [ ] 添加 `CancelApproval` 处理器方法
- [ ] 添加 `ReassignApproval` 处理器方法
- [ ] 更新路由配置

### Phase 3.3: 单元测试 (后端)
- [ ] 测试状态转换逻辑
- [ ] 测试审批人分配
- [ ] 测试完整审批流程

### Phase 4.1: API服务层 (前端)
- [ ] 实现 `handleApproval` 真实API调用
- [ ] 实现 `resubmitApproval` 方法
- [ ] 实现 `cancelApproval` 真实API调用

### Phase 4.2: UI组件增强 (前端)
- [ ] 优化 `ApprovalDetail` 组件
- [ ] 添加驳回理由展示
- [ ] 添加重新提交表单
- [ ] 完善审批历史展示

### Phase 5: 集成测试
- [ ] 完整审批流程测试
- [ ] 驳回-修改-重新提交流程测试
- [ ] 多级审批流程测试
- [ ] 权限控制测试

## 6. 关键代码片段

### 6.1 审批决定处理 (后端)

```go
func (s *ApprovalService) ProcessDecision(userID, approvalID string, req *ApprovalDecisionRequest) error {
    approval, err := s.approvalRepo.FindByID(approvalID)
    if err != nil {
        return err
    }

    // 验证权限
    if approval.CurrentApproverID != userID {
        return errors.New("无权处理此审批")
    }

    // 验证状态转换
    if !s.stateMachine.CanTransition(approval.Status, getTargetStatus(req.Decision)) {
        return errors.New("当前状态不允许此操作")
    }

    // 创建审批记录
    record := &models.ApprovalRecord{
        ApprovalRequestID: approval.ID,
        ApproverID:       userID,
        Decision:         req.Decision,
        DecisionReason:   req.DecisionReason,
        ApprovalDate:     time.Now(),
    }

    // 更新审批状态
    switch req.Decision {
    case models.ApprovalDecisionApprove:
        if s.hasNextStage(approval) {
            s.moveToNextStage(approval)
        } else {
            approval.Status = models.ApprovalStatusApproved
        }
    case models.ApprovalDecisionReject:
        approval.Status = models.ApprovalStatusRejected
    case models.ApprovalDecisionRequestChanges:
        approval.Status = models.ApprovalStatusNeedsRevision
    }

    return s.approvalRepo.Update(approval)
}
```

### 6.2 重新提交处理 (后端)

```go
func (s *ApprovalService) ResubmitApproval(userID, approvalID string, req *ResubmitRequest) error {
    approval, err := s.approvalRepo.FindByID(approvalID)
    if err != nil {
        return err
    }

    // 只有申请人可以重新提交
    if approval.ApplicantID != userID {
        return errors.New("只有申请人可以重新提交")
    }

    // 只有被拒绝或要求修改的申请可以重新提交
    if approval.Status != models.ApprovalStatusRejected &&
       approval.Status != models.ApprovalStatusNeedsRevision {
        return errors.New("当前状态不允许重新提交")
    }

    // 更新状态为重新提交
    approval.Status = models.ApprovalStatusResubmitted
    approval.UpdatedBy = userID

    // 保存修改说明
    // 重新分配审批人
    s.assigner.AssignApprover(approval)

    return s.approvalRepo.Update(approval)
}
```

### 6.3 前端审批操作 (前端)

```typescript
export const handleApproval = async (
  id: string,
  decision: 'approve' | 'reject' | 'request_changes',
  decisionReason: string,
  decisionComments?: string
): Promise<any> => {
  return await post(`/approvals/${id}/decision`, {
    decision,
    decisionReason,
    decisionComments,
  })
}

export const resubmitApproval = async (
  id: string,
  revisionNote: string,
  modifiedFields?: Record<string, any>
): Promise<any> => {
  return await post(`/approvals/${id}/resubmit`, {
    revisionNote,
    modifiedFields,
  })
}
```

## 7. 风险评估

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 状态转换逻辑复杂 | 可能引入状态不一致 | 严格的状态机验证，单元测试覆盖 |
| 多级审批嵌套 | 性能问题 | 限制最大审批层级，使用缓存 |
| 权限控制 | 安全风险 | 基于角色的权限检查，审计日志 |
| 数据迁移 | 兼容性问题 | 提供数据迁移脚本，保持向后兼容 |

## 8. 测试策略

### 8.1 单元测试
- 状态机转换测试
- 审批人分配测试
- 权限验证测试

### 8.2 集成测试
- 完整审批流程测试
- 驳回-修改-重新提交测试
- 多级审批测试

### 8.3 E2E测试
- 用户提交审批
- 审批人审批通过
- 审批人要求修改
- 申请人修改并重新提交

---

文档版本: 1.0
创建日期: 2026-01-13
