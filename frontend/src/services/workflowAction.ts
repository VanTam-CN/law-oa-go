// WorkflowAction contract (plan §5.1). The server generates the action
// after object-level authorization; the frontend only maps whitelisted
// action codes to internal routes and never consumes raw redirect URLs.
import type { WorkflowAction } from './approval'

export interface ActionTarget {
  path: string
  query: Record<string, string>
}

export type ActionRouteResult =
  | { kind: 'route'; target: ActionTarget }
  | { kind: 'view_only'; reason: string }
  | { kind: 'legacy'; reason: string }

const legacyReason = '历史事项，关联对象待确认'

// actionRoute maps a projected action to the internal navigation target.
// Unknown codes, missing targets, and blocked actions degrade to view-only
// instead of guessing a route.
export function actionRoute(action: WorkflowAction | null | undefined): ActionRouteResult {
  if (!action || !action.action_code) {
    return { kind: 'view_only', reason: '暂无可办理动作' }
  }
  if (action.blocked_code || action.action_code === 'view_only') {
    const reason = action.blocked_reason || '该事项暂不可办理，请查看详情'
    // §5.2: legacy source rows whose real object cannot be verified must
    // show the historical-item notice and never offer a business button.
    if (action.blocked_code === 'legacy_unresolved') {
      return { kind: 'legacy', reason: legacyReason }
    }
    return { kind: 'view_only', reason }
  }
  switch (action.action_code) {
    case 'decide_approval':
      if (!action.resource_id) return { kind: 'view_only', reason: '审批对象缺失，请查看详情' }
      return { kind: 'route', target: { path: `/approval/${action.resource_id}`, query: {} } }
    case 'review_conflict':
      if (!action.resource_id) return { kind: 'view_only', reason: '复核对象缺失，请查看详情' }
      return { kind: 'route', target: { path: '/conflict', query: { task_id: action.resource_id } } }
    case 'supplement': {
      // §6.2: the workbench gate enters revision mode only with
      // intake_id + approval_id together; without the related approval
      // the todo fails closed instead of opening a blank new-intake form.
      if (!action.resource_id || !action.related_approval_id) {
        return { kind: 'view_only', reason: '补充对象缺失，请查看详情' }
      }
      return {
        kind: 'route',
        target: {
          path: '/case/create',
          query: {
            intake_id: action.resource_id,
            approval_id: action.related_approval_id,
            mode: 'revision',
          },
        },
      }
    }
    case 'open_case_task': {
      // The task id (resource_id) is NOT a case id. The projected action
      // carries the owning case explicitly in related_case_id (review note
      // 2026-09-30); without it the todo degrades to view-only instead of
      // navigating to a wrong matter.
      if (!action.related_case_id) {
        return { kind: 'view_only', reason: '关联案件缺失，请查看任务详情' }
      }
      return {
        kind: 'route',
        target: {
          path: `/case/${encodeURIComponent(action.related_case_id)}`,
          query: { task_id: action.resource_id },
        },
      }
    }
    default:
      return { kind: 'view_only', reason: '该事项暂不支持直达办理，请查看详情' }
  }
}
