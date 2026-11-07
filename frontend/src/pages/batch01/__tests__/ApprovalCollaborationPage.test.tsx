import React from 'react'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { ApprovalDecisionFlow } from '../Batch01Prototype'
import { message } from '@/utils/messageHelper'

jest.mock('@/utils/storage', () => ({ getToken: () => 'test', getUserInfo: () => ({ id: '7', role: 'director' }), getRoles: () => [] }))
jest.mock('@/utils/messageHelper', () => ({ message: { success: jest.fn(), error: jest.fn(), warning: jest.fn(), info: jest.fn() } }))
const version = '2026-11-07T02:03:04.123456789Z'
let decisions: any[]
let failSnapshot = false
let legacyCurrent = false
let snapshotStatus = 0
let detail: any
let afterDecision: any
let integration: any
let decisionStatus = 200
beforeEach(() => {
  decisions = []; failSnapshot = false; legacyCurrent = false; snapshotStatus = 0; decisionStatus = 200
  detail = { type: 'case_creation', submission_snapshot_required: true }
  afterDecision = { status: 'under_review', current_stage: 'compliance_review', current_approver_name: '合规员', current_approver_id: '8' }
  integration = {}
  global.fetch = jest.fn(async (raw, options) => {
    const path = String(raw)
    let data: any = {}
    if (failSnapshot && path.includes('/snapshot')) throw new Error('版本读取中断')
    if (snapshotStatus && path.includes('/snapshot')) return { ok: false, status: snapshotStatus, json: async () => ({ message: 'snapshot unavailable' }) } as Response
    if (options?.method === 'POST') {
      decisions.push(JSON.parse(String(options.body)))
      if (decisionStatus !== 200) return { ok: false, status: decisionStatus, json: async () => ({ message: 'INTAKE_SUBMISSION_REQUIRED' }) } as Response
      detail = { ...detail, ...afterDecision }; data = detail
    }
    else if (path.includes('/snapshot')) data = {
      active_submission_snapshot_id: legacyCurrent ? null : 'current-snapshot',
      viewed_snapshot_id: path.includes('snapshot_id=old-snapshot') ? 'old-snapshot' : 'current-snapshot',
      submission_history: [{ id: 'current-snapshot', source_version: 2, snapshot_type: legacyCurrent ? 'conflict_approval' : 'intake_submission' }, { id: 'old-snapshot', source_version: 1, snapshot_type: 'conflict_approval' }],
      snapshot: { conflict_review: { decision: path.includes('snapshot_id=old-snapshot') ? 'insufficient_information' : 'no_conflict' }, conflict_result: { checkId: 'check-1', decision: { status: 'REVIEW_REQUIRED', coverageStatus: path.includes('snapshot_id=old-snapshot') ? 'PARTIAL' : 'COMPLETE' }, riskAssessment: { overallRisk: 'REVIEW_REQUIRED' } }, metadata: { case_creation_config: { title: path.includes('snapshot_id=old-snapshot') ? '历史提交标题' : '当前提交标题' } }, parties: [] },
    }
    else if (path.endsWith('/review')) data = { id: 'review-1', checkId: 'check-1', decision: 'no_conflict' }
    else if (path === '/api/v1/approvals/approval-uuid') data = {
      id: 'approval-uuid', title: '审批事项', applicant_id: '2', current_approver_id: '7', status: 'submitted', action_version: version, active_submission_snapshot_id: legacyCurrent ? null : 'current-snapshot', available_actions: ['approve', 'reject', 'request_changes'], conflict_check_id: 'check-1', conflict_result: { checkId: 'check-1', decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' }, riskAssessment: { overallRisk: 'REVIEW_REQUIRED' } }, records: [], ...detail,
    }
    if (path.includes('/integration/') && path.endsWith('/status')) data = integration
    return { ok: true, status: 200, json: async () => ({ data }) } as Response
  }) as jest.Mock
})
afterEach(cleanup)
function mount() {
  render(<MemoryRouter initialEntries={['/approval/approval-uuid']}><Routes><Route path='/approval/:id' element={<ApprovalDecisionFlow />} /></Routes></MemoryRouter>)
}
it('returns two structured requirements as an object with the exact server action version', async () => {
  mount()
  fireEvent.click(await screen.findByRole('button', { name: '退回修改' }))
  fireEvent.change(screen.getByLabelText('处理依据'), { target: { value: '请核对当前申请并分别补充以下两项说明' } })
  fireEvent.change(screen.getByLabelText('补正要求 1'), { target: { value: '说明第一项事项' } })
  fireEvent.click(screen.getByRole('button', { name: '增加补正要求' }))
  fireEvent.change(screen.getByLabelText('补正要求 2'), { target: { value: '说明第二项事项' } })
  fireEvent.click(screen.getByRole('button', { name: '确认退回修改' }))
  await waitFor(() => expect(decisions).toHaveLength(1))
  expect(decisions[0].expected_action_version).toBe(version)
  expect(decisions[0].expected_submission_snapshot_id).toBe('current-snapshot')
  expect(typeof decisions[0].imposed_requirements).toBe('object')
  expect(decisions[0].imposed_requirements.supplement_requests).toHaveLength(2)
  expect(decisions[0].imposed_requirements.supplement_requests[0]).toEqual(expect.objectContaining({ field_key: 'description', required: true }))
})

it('makes a historical snapshot read-only until switching back to the current submission', async () => {
  mount()
  await screen.findByRole('button', { name: '退回修改' })
  fireEvent.mouseDown(screen.getByRole('combobox', { name: '提交版本' }))
  fireEvent.click(await screen.findByText('第 1 次提交 · 历史只读'))
  await screen.findByText('正在查看历史提交版本，仅供查阅')
  expect(screen.getAllByText('历史提交标题').length).toBeGreaterThan(0)
  expect(screen.queryByText('当前提交标题')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: '退回修改' })).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '切回当前版本' }))
  await screen.findByRole('button', { name: '退回修改' })
  expect(decisions).toHaveLength(0)
})

it('disables decisions when the submission snapshot cannot be read', async () => {
  failSnapshot = true
  mount()
  await screen.findByText('提交版本读取失败，请重新加载后再处理审批')
  expect(screen.queryByRole('button', { name: '退回修改' })).not.toBeInTheDocument()
  expect(decisions).toHaveLength(0)
})

it('treats the server-verified legacy current snapshot as current and consumes the live review', async () => {
  legacyCurrent = true
  mount()
  const approve = await screen.findByRole('button', { name: /同意并通过本阶段/ })
  expect(approve).toBeEnabled()
  expect(screen.getByRole('button', { name: '退回修改' })).toBeEnabled()
  expect(screen.queryByText('正在查看历史提交版本，仅供查阅')).not.toBeInTheDocument()
  fireEvent.click(approve)
  fireEvent.change(screen.getByLabelText('处理依据'), { target: { value: '独立复核已完成并已核对当前申请事项' } })
  fireEvent.click(screen.getByRole('button', { name: '确认同意并通过本阶段' }))
  await waitFor(() => expect(decisions).toHaveLength(1))
  expect(decisions[0].expected_submission_snapshot_id).toBe('current-snapshot')
})

it('keeps explicitly selected legacy history read-only even with a null active pointer', async () => {
  legacyCurrent = true
  mount()
  await screen.findByRole('button', { name: '退回修改' })
  fireEvent.mouseDown(screen.getByRole('combobox', { name: '提交版本' }))
  fireEvent.click(await screen.findByText('第 1 次提交 · 历史只读'))
  await screen.findByText('正在查看历史提交版本，仅供查阅')
  expect(screen.queryByRole('button', { name: '退回修改' })).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '切回当前版本' }))
  await screen.findByRole('button', { name: '退回修改' })
})

function ordinary() {
  detail = { type: 'leave', submission_snapshot_required: false, active_submission_snapshot_id: null, conflict_check_id: undefined, conflict_result: undefined }
  snapshotStatus = 404
  afterDecision = { status: 'approved' }
}
async function approve() {
  fireEvent.click(await screen.findByRole('button', { name: /同意并通过本阶段$/ }))
  fireEvent.change(screen.getByLabelText('处理依据'), { target: { value: '已经核对本阶段申请事项及相关条件' } })
  fireEvent.click(screen.getByRole('button', { name: '确认同意并通过本阶段' }))
  await waitFor(() => expect(decisions).toHaveLength(1))
}
it('allows an explicitly snapshot-free ordinary approval without case promises or a snapshot token', async () => {
  ordinary(); mount()
  await approve()
  expect(decisions[0]).not.toHaveProperty('expected_submission_snapshot_id')
  expect(decisions[0].expected_action_version).toBe(version)
  expect(decisions[0].decision_comments).not.toMatch(/成案/)
  await waitFor(() => expect(message.success).toHaveBeenCalledWith('审批已通过'))
  expect(screen.queryByText(/待生成正式案件|成案与归档/)).not.toBeInTheDocument()
})
it.each([
  ['missing capability', { submission_snapshot_required: undefined }, 404],
  ['required snapshot', { submission_snapshot_required: true }, 404],
  ['active pointer contradicts optional', { active_submission_snapshot_id: 'missing-snapshot' }, 404],
  ['intake source contradicts optional', { metadata: { intake_id: 'old-intake' } }, 404],
  ['case type contradicts optional', { type: 'case_creation' }, 404],
  ['forbidden', {}, 403], ['conflict', {}, 409], ['server error', {}, 500],
])('keeps decisions blocked for %s', async (_name, overrides, code) => {
  ordinary(); detail = { ...detail, ...(overrides as object) }; snapshotStatus = code as number
  mount()
  await screen.findByText('提交版本读取失败，请重新加载后再处理审批')
  expect(screen.queryByRole('button', { name: /同意并通过本阶段$/ })).not.toBeInTheDocument()
  expect(decisions).toHaveLength(0)
})
it('reports only the completed stage for a two-node approval', async () => {
  mount(); await approve()
  await waitFor(() => expect(message.success).toHaveBeenCalledWith('本阶段已通过，下一节点：合规复核（合规员）'))
  await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  expect(message.success).not.toHaveBeenCalledWith('审批已通过')
  expect(decisions[0].decision_comments).not.toMatch(/触发成案/)
})
it('reports a created case only after final approval and a real case id', async () => {
  afterDecision = { status: 'approved' }; integration = { case_creation: { case_id: 'case-1', case_number: 'CASE-001' } }
  mount(); await approve()
  await waitFor(() => expect(message.success).toHaveBeenCalledWith('最终审批已通过，已成案：CASE-001'))
})
it('does not retry or fall back when a readable old intake approval is refused by the server', async () => {
  legacyCurrent = true; detail.type = 'conflict_approval'; decisionStatus = 409
  mount(); await approve()
  await waitFor(() => expect(message.error).toHaveBeenCalledWith('INTAKE_SUBMISSION_REQUIRED'))
  expect(decisions).toHaveLength(1)
  expect(decisions[0].expected_submission_snapshot_id).toBe('current-snapshot')
})
it('separates frozen automatic results, complete coverage and the current independent review', async () => {
  mount()
  expect(await screen.findByText('当前可读独立复核：无冲突')).toBeInTheDocument()
  expect(screen.getByText('检索覆盖：完整')).toBeInTheDocument()
  expect(screen.getByText('冻结自动检索结果：待人工复核')).toBeInTheDocument()
  expect(screen.getByText('未发现匹配证据')).toBeInTheDocument()
  expect(screen.queryByText(/总体风险等级：待人工复核|未发现匹配证据（检索覆盖受限）/)).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: /同意并通过本阶段$/ })).toBeEnabled()
})
it('keeps incomplete coverage blocking despite a no-conflict review', async () => {
  detail.conflict_result = { checkId: 'check-1', decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'PARTIAL' } }
  mount()
  expect(await screen.findByText('当前可读独立复核：无冲突')).toBeInTheDocument()
  expect(screen.getByText('检索覆盖：未完整（PARTIAL）')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: /同意并通过本阶段$/ })).toBeDisabled()
})
it('shows only the chosen historical review and coverage until returning to current', async () => {
  mount(); await screen.findByText('当前可读独立复核：无冲突')
  fireEvent.mouseDown(screen.getByRole('combobox', { name: '提交版本' }))
  fireEvent.click(await screen.findByText('第 1 次提交 · 历史只读'))
  await screen.findByText('正在查看历史提交版本，仅供查阅')
  expect(screen.getByText('该版本冻结独立复核：信息不足')).toBeInTheDocument()
  expect(screen.getByText('检索覆盖：未完整（PARTIAL）')).toBeInTheDocument()
  expect(screen.queryByText('当前可读独立复核：无冲突')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /同意并通过本阶段$/ })).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '切回当前版本' }))
  expect(await screen.findByText('当前可读独立复核：无冲突')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: /同意并通过本阶段$/ })).toBeEnabled()
})

it('keeps ordinary approvals blocked on snapshot transport failure', async () => {
  ordinary(); snapshotStatus = 0; failSnapshot = true; mount()
  await screen.findByText('提交版本读取失败，请重新加载后再处理审批')
  expect(screen.queryByRole('button', { name: /同意并通过本阶段$/ })).not.toBeInTheDocument()
})
it('still requires the raw action version for snapshot-free decisions', async () => {
  ordinary(); detail.action_version = ''; mount()
  fireEvent.click(await screen.findByRole('button', { name: /同意并通过本阶段$/ }))
  fireEvent.change(screen.getByLabelText('处理依据'), { target: { value: '已经核对本阶段申请事项及相关条件' } })
  fireEvent.click(screen.getByRole('button', { name: '确认同意并通过本阶段' }))
  expect(message.error).toHaveBeenCalledWith('审批版本尚未加载，请刷新详情后处理')
  expect(decisions).toHaveLength(0)
})
