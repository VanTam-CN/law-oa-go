import { purgeIntakeIdentityCaches, intakeEditSignature } from '../intake/draftState'
import { currentSubmissionSnapshotID, latestSupplementReturn, saveSupplementResponses, resubmitIntakeApproval } from '@/services/intakeSupplement'
import { apiRequest } from '@/services/collaborationRequest'
import { workflowActionPath, reminderCompletionMode, type WorkflowAction } from '@/components/workflow/WorkflowActions'

jest.mock('@/services/collaborationRequest', () => ({ apiRequest: jest.fn() }))

it.each([
  { active_submission_snapshot_id: null, snapshot: { snapshot_id: 'SNAP_OLD' } },
  { active_submission_snapshot_id: null, viewed_snapshot_id: 'SNAP_OLD' },
  { active_submission_snapshot_id: 'active-row', viewed_snapshot_id: 'different-row' },
])('rejects an unverified current snapshot token: %j', (payload) => {
  expect(() => currentSubmissionSnapshotID(payload)).toThrow('无法核实当前提交版本')
})

it('removes old raw-identity fingerprints for every account without deleting unrelated local data', () => {
  localStorage.clear()
  localStorage.setItem('law-oa-case-intake-draft-v2:7:new:conflictFingerprint', '{"identityNumber":"FICTIONAL"}')
  localStorage.setItem('law-oa-case-intake-draft-v2:8:new:conflictFingerprint', 'protected')
  localStorage.setItem('unrelated', 'keep')
  purgeIntakeIdentityCaches(localStorage)
  expect(localStorage.length).toBe(1)
  expect(localStorage.getItem('unrelated')).toBe('keep')
})

it('does not count server token rotation as an edit but does detect an edited identity', () => {
  const form = { title: '草稿', revision: 'v1', opponentIdentityNumber: 'FICTIONAL-A' }
  const first = intakeEditSignature(form, [], [])
  expect(intakeEditSignature({ ...form, revision: 'v2' }, [], [])).toBe(first)
  expect(intakeEditSignature({ ...form, opponentIdentityNumber: 'FICTIONAL-B' }, [], [])).not.toBe(first)
})

it('reads only the latest return record and preserves its stable item ids', () => {
  const result = latestSupplementReturn({ records: [
    { id: 'old', decision: 'request_changes', created_at: '2026-01-01', imposed_requirements: '{}' },
    { id: 'new', decision: 'request_changes', created_at: '2026-02-01', decision_reason: '补充两个事项', imposed_requirements: JSON.stringify({ schema_version: 1, supplement_requests: [{ item_id: 'a', category: 'matter_explanation', description: '说明事项' }, { item_id: 'b', category: 'field_correction', field_key: 'title', description: '修改名称' }] }) },
  ] })
  expect(result?.recordId).toBe('new')
  expect(result?.items.map((item) => item.item_id)).toEqual(['a', 'b'])
})

it('saves replies without facts, metadata or browser assertions, then sends the original approval versions', async () => {
  const replies = [{ item_id: 'a', response: '已说明' }]
  await saveSupplementResponses('intake-1', 'approval-1', 'rev-1', 'return-1', replies)
  const save = (apiRequest as jest.Mock).mock.calls.at(-1)
  expect(JSON.parse(save[1].body)).toEqual({ approval_id: 'approval-1', expected_revision: 'rev-1', return_record_id: 'return-1', supplement_responses: replies })
  await resubmitIntakeApproval('approval-1', { expected_intake_revision: 'rev-2', expected_submission_snapshot_id: 'snapshot-1', return_record_id: 'return-1', supplement_responses: replies, revision_note: '已补充' })
  expect((apiRequest as jest.Mock).mock.calls.at(-1)[0]).toBe('/approvals/approval-1/resubmit')
})

const action: WorkflowAction = { resource_type: 'approval', resource_id: 'approval-uuid', action_code: 'decide_approval', action_label: '办理审批', workflow_status: 'submitted', completion_mode: 'workflow', action_version: 'exact-nanos' }
it('maps the server string id, refuses stale/unknown actions and never treats a business task as manual', () => {
  expect(workflowActionPath(action)).toBe('/approval/approval-uuid')
  expect(workflowActionPath({ ...action, blocked_code: 'TASK_SUPERSEDED' })).toBeNull()
  expect(workflowActionPath({ ...action, resource_type: 'case' })).toBeNull()
  expect(workflowActionPath({ ...action, resource_id: 7 as any })).toBeNull()
  expect(reminderCompletionMode({ completion_mode: 'manual', actions: [action] })).toBe('workflow')
})

it('uses the explicit approval association for supplement actions, never inventing it from intake id', () => {
  const supplement: WorkflowAction = { ...action, resource_type: 'intake', resource_id: 'intake-uuid', action_code: 'supplement' }
  expect(workflowActionPath(supplement)).toBeNull()
  expect(workflowActionPath({ ...supplement, approval_id: 'approval-uuid' })).toBe('/case/create?intake_id=intake-uuid&approval_id=approval-uuid&mode=revision')
})

it('uses server-normalized legacy requirements and keeps their server-assigned item ids', () => {
  const requirement = { item_id: 'legacy-return-1', category: 'matter_explanation', description: '旧文字意见', required: true }
  const returned = latestSupplementReturn({ return_record_id: 'return-1', return_reason: '请补充', supplement_requests: [requirement], records: [{ id: 'return-1', imposed_requirements: '旧文字意见' }] })
  expect(returned).toEqual({ recordId: 'return-1', legacyText: '请补充', items: [requirement] })
})
