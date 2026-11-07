import { apiRequest } from '@/services/collaborationRequest'
import { objectFromJSON } from '@/services/intakeSupplement'

// Follow the server's current binding, never the newest client history row.
export async function readBoundIntakeConflict(payload: Record<string, any>) {
  const intake = objectFromJSON(payload.intake)
  if (intake.status !== 'conflict_ready') return null
  const metadata = objectFromJSON(intake.metadata)
  const checkID = metadata.conflict_check_id
  if (typeof checkID !== 'string' || !checkID || !metadata.lawyer_facts_version) {
    throw new Error('当前草稿的检测绑定不完整，请刷新后核对')
  }
  const [response, reviewResponse] = await Promise.all([
    apiRequest<any>(`/conflict/tasks/${encodeURIComponent(checkID)}/result`),
    apiRequest<any>(`/conflict/tasks/${encodeURIComponent(checkID)}/review`),
  ])
  const task = objectFromJSON(response.task)
  const result = objectFromJSON(response.result)
  // GET /review returns models.ConflictReview directly (camelCase checkId),
  // unlike the reviewer-assignment DTO's snake_case check_id.
  const review = objectFromJSON(reviewResponse?.review || reviewResponse)
  if (task.status !== 'COMPLETED' || task.checkId !== checkID || task.intakeId !== intake.id || String(task.clientId) !== String(intake.client_id) || !Object.keys(result).length || (result.checkId && result.checkId !== checkID) || (review.id && review.checkId !== checkID)) {
    throw new Error('检测结果与当前草稿关联不一致，请刷新后核对')
  }
  return { conflict: { ...result, checkId: checkID }, conflictTask: task, conflictDecisionExtras: { review } }
}
