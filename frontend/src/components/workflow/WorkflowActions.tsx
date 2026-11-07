import React from 'react'
import { Button, Space, Tag } from 'antd'
import { useNavigate } from 'react-router'

export interface WorkflowAction {
  resource_type: 'intake' | 'conflict' | 'approval' | 'case'
  resource_id: string
  approval_id?: string
  action_code: 'resume_intake' | 'review_conflict' | 'decide_approval' | 'supplement' | 'view_only'
  action_label: string
  workflow_status: string
  assigned_to?: { id?: string; name?: string } | string | null
  blocked_code?: string
  blocked_reason?: string
  completion_mode: 'workflow' | 'manual' | 'legacy_unresolved'
  action_version: string
}
export function workflowStatusLabel(status: string) {
  const labels: Record<string, string> = { approved_pending_case: '审批已通过，待成案', case_creation_failed: '成案失败，可在审批详情重试', case_created: '已成案', case_creation_incomplete: '成案关联待核实', needs_revision: '待补正', submitted: '待审批', under_review: '审批中', resubmitted: '已重新提交', approved: '审批已通过', rejected: '审批已拒绝' }
  return labels[status] || status
}
export function workflowActionPath(action: WorkflowAction): string | null {
  if (
    ((action.blocked_code || action.blocked_reason) && action.action_code !== 'view_only') ||
    action.completion_mode === 'legacy_unresolved' ||
    typeof action.resource_id !== 'string' ||
    !action.resource_id.trim()
  )
    return null
  const id = encodeURIComponent(action.resource_id)
  if (action.resource_type === 'intake' && action.action_code === 'supplement') {
    if (typeof action.approval_id !== 'string' || !action.approval_id.trim()) return null
    return (
      '/case/create?' +
      new URLSearchParams({
        intake_id: action.resource_id,
        approval_id: action.approval_id,
        mode: 'revision',
      }).toString()
    )
  }
  if (
    action.resource_type === 'intake' &&
    ['resume_intake', 'view_only'].includes(action.action_code)
  )
    return `/case/create?intake_id=${id}`
  if (
    action.resource_type === 'approval' &&
    ['decide_approval', 'supplement', 'view_only'].includes(action.action_code)
  )
    return `/approval/${id}`
  if (
    action.resource_type === 'conflict' &&
    ['review_conflict', 'view_only'].includes(action.action_code)
  )
    return `/conflict?task_id=${id}`
  if (action.resource_type === 'case' && action.action_code === 'view_only') return `/case/${id}`
  return null
}
export function workflowActionsOf(item: {
  actions?: WorkflowAction[]
  workflow_action?: WorkflowAction | null
}) {
  return Array.isArray(item.actions)
    ? item.actions
    : item.workflow_action
      ? [item.workflow_action]
      : []
}
export function reminderCompletionMode(item: {
  completion_mode?: string
  actions?: WorkflowAction[]
  workflow_action?: WorkflowAction | null
}) {
  if (
    item.completion_mode === 'workflow' ||
    workflowActionsOf(item).some((action) => action.completion_mode === 'workflow')
  )
    return 'workflow'
  if (item.completion_mode === 'manual') return 'manual'
  return 'legacy_unresolved'
}
export default function WorkflowActions({ actions = [] }: { actions?: WorkflowAction[] }) {
  const navigate = useNavigate()
  return (
    <Space wrap>
      {actions.map((action, index) => {
        const path = workflowActionPath(action)
        return (
          <Space direction='vertical' key={`${action.action_code}-${action.resource_id}-${index}`}>
            <Button
              disabled={!path}
              onClick={() => {
                if (path) navigate(path)
              }}
            >
              {action.action_label || '查看事项'}
            </Button>
            {action.workflow_status && <Tag>{workflowStatusLabel(action.workflow_status)}</Tag>}
            {(action.blocked_reason || !path) && (
              <span>{action.blocked_reason || '当前动作缺少可用关联信息，请刷新后核对'}</span>
            )}
            {action.assigned_to && (
              <Tag>
                处理人：
                {typeof action.assigned_to === 'string'
                  ? action.assigned_to
                  : action.assigned_to.name || '待确认'}
              </Tag>
            )}
          </Space>
        )
      })}
    </Space>
  )
}
