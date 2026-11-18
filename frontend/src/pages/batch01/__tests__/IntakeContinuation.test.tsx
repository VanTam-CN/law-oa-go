import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, useLocation, useNavigate } from 'react-router'
import { CaseIntakeWorkbench, scopedCaseIntakeDraftKey } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  getRoles: jest.fn(() => []), getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({ id: 7, role: 'lawyer' })),
}))
jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as typeof fetch
const json = (data: unknown) => ({ ok: true, json: async () => ({ success: true, data }) }) as Response
const revisionPath = '/case/create?mode=revision&intake_id=I1&approval_id=A1'
const revisionApproval = { approval_id: 'A1', snapshot_id: 'S1', status: 'needs_revision', round: 1 }
const intake = (id = 'I1') => ({
  intake: { id, intake_code: `INT-${id}`, title: `案件-${id}`, case_type: 'commercial',
    description: '本轮案情摘要', client_id: 42, created_by: '7',
    metadata: { lawyer_id: 7, business_area: '商事', sub_area: '合同' } },
  revision: 'rev-1',
  parties: [
    { entity_name: '虚构客户', entity_type: 'LEGAL_PERSON', party_role: 'client' },
    { entity_name: '虚构对方', entity_type: 'LEGAL_PERSON', party_role: 'opposing_party' },
  ],
})
const continuation = (id = 'I1', approved = false) => ({
  intake_id: id, revision: 'rev-1', status: 'conflict_ready', can_reuse_identity: true,
  can_reuse_check: true, can_submit: approved, recheck_required: false,
  reasons: approved ? [] : ['REVIEW_MISSING'], facts: { facts_version: 'fv1', confirmed_at: '2026-10-02' },
  conflict_binding: { check_id: `CHECK-${id}`, check_status: 'COMPLETED', coverage_status: 'COMPLETE',
    ...(approved ? { review_id: 11, reviewer_id: 9, review_decision: 'no_conflict' } : {}) },
  subjects: [{ kind: 'party', party_id: 12, name: '虚构对方', role: 'opposing_party',
    entity_type: 'LEGAL_PERSON', identity_type: 'SOCIAL_CREDIT_CODE', identity_masked: '****0001' }],
})
function LocationProbe() {
  const location = useLocation()
  const navigate = useNavigate()
  return <><output data-testid='location'>{location.pathname + location.search}</output>
    <button onClick={() => navigate('/case/create?draft_id=I2')}>打开第二份申请</button></>
}
async function open(path = revisionPath) {
  await act(async () => {
    render(<MemoryRouter initialEntries={[path]}><CaseIntakeWorkbench /><LocationProbe /></MemoryRouter>)
  })
}
let proof: unknown
let proofResponse: (() => Promise<Response>) | undefined
beforeEach(() => {
  window.localStorage.clear()
  proof = { ...continuation(), approval: revisionApproval }
  proofResponse = undefined
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/continuation')) return proofResponse ? proofResponse() : Promise.resolve(json(proof))
    if (/\/case-intakes\/I[12]$/.test(path)) return Promise.resolve(json(intake(path.endsWith('I2') ? 'I2' : 'I1')))
    if (path.endsWith('/approvals/A1')) return Promise.resolve(json({ id: 'A1', active_submission_snapshot_id: 'S1', records: [
      { id: 'R1', decision: 'request_changes', imposed_requirements: { items: [] } },
    ] }))
    if (path.endsWith('/approvals/A1/snapshot')) return Promise.resolve(json({ approval_id: 'A1', snapshot_id: 'S1', snapshot: {} }))
    if (path.includes('/clients?')) return Promise.resolve(json({ clients: [{ id: 42, name: '虚构客户', type: '企业' }] }))
    if (path.includes('/lawfirm/lawyers?')) return Promise.resolve(json({ lawyers: [{ id: 7, name: '律师甲' }] }))
    return Promise.resolve(json({}))
  })
})

it('resumes the original pending review without identity re-entry or another check; focus refresh picks up the verdict', async () => {
  await open()
  expect(screen.getByTestId('intake-continuation')).toHaveTextContent('等待独立复核')
  expect((screen.getByLabelText('对方身份标识 *') as HTMLInputElement).value).toBe('')
  expect(screen.getByText(/已由系统核验/)).toHaveTextContent('****0001')
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeDisabled()
  const step = document.querySelectorAll('.batch-stepper button')[1]
  fireEvent.click(step)
  fireEvent.click(screen.getByRole('button', { name: '查看本条检测复核记录' }))
  expect(screen.getByTestId('location').textContent).toContain('task_id=CHECK-I1')
  proof = { ...continuation('I1', true), approval: revisionApproval }
  await act(async () => { window.dispatchEvent(new Event('focus')) })
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeEnabled()
  expect(fetchMock.mock.calls.some(([p]) => String(p).endsWith('/conflict-check'))).toBe(false)
  expect(JSON.stringify(window.localStorage)).not.toContain('****0001')
})

it('invalidates reuse immediately on a fact edit and keeps that input on focus', async () => {
  proof = { ...continuation('I1', true), approval: revisionApproval }
  await open()
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeEnabled()
  fireEvent.change(screen.getByLabelText('案情摘要 *'), { target: { value: '我正在更正的摘要' } })
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeDisabled()
  await act(async () => { window.dispatchEvent(new Event('focus')) })
  expect(screen.getByLabelText('案情摘要 *')).toHaveValue('我正在更正的摘要')
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeDisabled()
})

it.each(['failure', 'wrong_revision', 'wrong_intake', 'unknown'])(
  'fails closed for %s even when local storage contains validity flags', async (variant) => {
    window.localStorage.setItem(scopedCaseIntakeDraftKey('7'), JSON.stringify({
      ...intake().intake, intakeId: 'I1', intakeIdentityVerified: true, can_submit: true,
    }))
    proof = variant === 'unknown' ? {} : {
      ...continuation('I1', true), intake_id: variant === 'wrong_intake' ? 'OTHER' : 'I1',
      revision: variant === 'wrong_revision' ? 'rev-old' : 'rev-1',
    }
    if (variant === 'failure') proofResponse = async () => { throw new Error('offline') }
    await open()
    expect(screen.getByRole('button', { name: '重新提交审批' })).toBeDisabled()
    expect(screen.getByTestId('intake-continuation')).toHaveTextContent(/未能确认|无法确认/)
  },
)

it('opens a server draft on start and ignores a continuation arriving after switching scope', async () => {
  let resolveOld!: (response: Response) => void
  proofResponse = () => new Promise<Response>((resolve) => { resolveOld = resolve })
  await open('/case/create?draft_id=I1')
  expect(screen.getByLabelText('案件名称 *')).toHaveValue('案件-I1')
  proofResponse = async () => json(continuation('I2', false))
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '打开第二份申请' })) })
  await act(async () => { resolveOld(json(continuation('I1', true))) })
  expect(screen.getByLabelText('案件名称 *')).toHaveValue('案件-I2')
  expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
  expect(screen.getByTestId('intake-continuation')).toHaveTextContent('等待独立复核')
})

it('resubmits the same untouched intake and snapshot without a PUT, re-entry or another check', async () => {
  proof = { ...continuation('I1', true), approval: revisionApproval }
  await open()
  const button = screen.getByRole('button', { name: '重新提交审批' })
  expect(button).toBeEnabled()
  fireEvent.change(screen.getByLabelText('修改说明'), { target: { value: '本轮逐项说明已补充' } })
  await act(async () => { fireEvent.click(button) })
  const call = fetchMock.mock.calls.find(([path]) => String(path).endsWith('/approvals/A1/resubmit'))
  expect(call).toBeDefined()
  expect(JSON.parse(String(call?.[1]?.body))).toMatchObject({
    expected_intake_revision: 'rev-1', expected_submission_snapshot_id: 'S1', return_record_id: 'R1',
    revision_note: '本轮逐项说明已补充',
  })
  expect(fetchMock.mock.calls.some(([path, init]) => String(path).endsWith('/case-intakes/I1') && init?.method === 'PUT')).toBe(false)
  expect(fetchMock.mock.calls.some(([path]) => /facts-confirmation|conflict-check/.test(String(path)))).toBe(false)
})

it.each(['approval', 'snapshot'])('blocks a continuation for another %s context', async (variant) => {
  proof = { ...continuation('I1', true), approval: {
    ...revisionApproval,
    approval_id: variant === 'approval' ? 'A2' : 'A1', snapshot_id: variant === 'snapshot' ? 'S2' : 'S1',
  } }
  await open()
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeDisabled()
  expect(screen.getByTestId('intake-continuation')).toHaveTextContent('未能确认')
})

it('rejects an older focus response after an edit, including reverting the text', async () => {
  proof = { ...continuation('I1', true), approval: revisionApproval }
  await open()
  let resolveOld!: (response: Response) => void
  proofResponse = () => new Promise((resolve) => { resolveOld = resolve })
  await act(async () => { window.dispatchEvent(new Event('focus')) })
  fireEvent.change(screen.getByLabelText('案情摘要 *'), { target: { value: '正在修改' } })
  fireEvent.change(screen.getByLabelText('案情摘要 *'), { target: { value: '本轮案情摘要' } })
  await act(async () => { resolveOld(json(proof)) })
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeDisabled()
  expect(screen.getByLabelText('案情摘要 *')).toHaveValue('本轮案情摘要')
})

it('the revision save card names the available save-and-check and resubmit buttons', async () => {
  await open()
  fireEvent.change(screen.getByLabelText('案情摘要 *'), { target: { value: '未保存的摘要' } })
  expect(screen.getByText(/有未保存修改/)).toHaveTextContent('保存最新输入并检测')
  expect(screen.getByText(/有未保存修改/)).toHaveTextContent('重新提交审批')
  expect(screen.queryByText(/请点“保存草稿”|可在页面底部选择“保存草稿”/)).not.toBeInTheDocument()
})

it.each(['submitted', 'under_review', 'resubmitted', 'approved', 'rejected', 'cancelled', 'expired', 'draft', 'unknown', 'request_changes'])(
  'retains reusable identity/check but blocks duplicate submission for approval phase %s', async (status) => {
    const inProgress = ['submitted', 'under_review', 'resubmitted'].includes(status)
    const closed = ['approved', 'rejected', 'cancelled', 'expired'].includes(status)
    proof = { ...continuation('I1', true), can_submit: false,
      reasons: [inProgress ? 'APPROVAL_IN_PROGRESS' : closed ? 'APPROVAL_CLOSED' : 'APPROVAL_STATE_UNKNOWN'],
      approval: { ...revisionApproval, status },
    }
    await open('/case/create?draft_id=I1')
    expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
    expect(screen.getByTestId('intake-continuation')).toHaveTextContent(
      inProgress ? '原审批正在处理中' : closed ? '原审批已结束' : '原审批状态尚未确认',
    )
    expect(screen.getByTestId('intake-continuation')).not.toHaveTextContent('等待独立复核')
    expect(screen.getByTestId('intake-continuation')).not.toHaveTextContent('保存最新输入并检测')
    expect(screen.getByText(/已由系统核验/)).toHaveTextContent('****0001')
    fireEvent.change(screen.getByLabelText('案情摘要 *'), { target: { value: '仍保留我的未保存资料' } })
    await act(async () => { window.dispatchEvent(new Event('focus')) })
    expect(screen.getByLabelText('案情摘要 *')).toHaveValue('仍保留我的未保存资料')
    expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: '查看原审批进度' })) })
    expect(screen.getByTestId('location')).toHaveTextContent('/approval/A1')
    expect(fetchMock.mock.calls.some(([path]) => /facts-confirmation|conflict-check/.test(String(path)))).toBe(false)
  },
)

it('requires the original revision route rather than opening a second initial approval', async () => {
  proof = { ...continuation('I1', true), approval: revisionApproval }
  await open('/case/create?draft_id=I1')
  expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '回原申请补正' })) })
  expect(screen.getByTestId('location')).toHaveTextContent(revisionPath)
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeEnabled()
})

it.each([false, true])('automatically refreshes continuation after check completion (review approved: %s)', async (reviewApproved) => {
  const originalFetch = fetchMock.getMockImplementation()!
  let serverRevision = 'rev-1'
  let checkComplete = false
  let resolveCheck!: (response: Response) => void
  fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
    const path = String(input)
    if (path.endsWith('/case-intakes/I1') && init?.method === 'PUT') {
      serverRevision = 'rev-2'
      return Promise.resolve(json({ id: 'I1', intake_code: 'INT-I1', revision: serverRevision }))
    }
    if (path.endsWith('/facts-confirmation')) {
      serverRevision = 'rev-3'
      return Promise.resolve(json({ revision: serverRevision }))
    }
    if (path.endsWith('/conflict-check')) return new Promise<Response>((resolve) => { resolveCheck = resolve })
    if (path.endsWith('/continuation')) return Promise.resolve(json({
      ...continuation('I1', checkComplete && reviewApproved), revision: serverRevision,
      approval: revisionApproval, can_reuse_check: checkComplete, recheck_required: !checkComplete,
      reasons: checkComplete ? reviewApproved ? [] : ['REVIEW_MISSING'] : ['CONFLICT_CHECK_INCOMPLETE'],
    }))
    return originalFetch(input, init)
  })
  await open()
  fireEvent.change(screen.getByLabelText('对方身份标识 *'), { target: { value: '91310000TEST000001' } })
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '保存并进行利益冲突检查' })) })
  await waitFor(() => expect(resolveCheck).toBeDefined())
  await waitFor(() => expect(screen.getByTestId('intake-continuation')).toHaveTextContent('检测尚未完成'))
  const readsBeforeCompletion = fetchMock.mock.calls.filter(([path]) => String(path).endsWith('/continuation')).length
  checkComplete = true
  await act(async () => { resolveCheck(json({ task_id: 'CHECK-I1', status: 'COMPLETED', result: {
    checkId: 'CHECK-I1', decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' },
    riskAssessment: { overallRisk: 'HIGH', requiresApproval: true },
  } })) })
  await waitFor(() => expect(screen.getByTestId('intake-continuation')).toHaveTextContent(
    reviewApproved ? '独立复核已由系统确认' : '等待独立复核',
  ))
  expect(fetchMock.mock.calls.filter(([path]) => String(path).endsWith('/continuation')).length).toBeGreaterThan(readsBeforeCompletion)
  if (reviewApproved) expect(screen.getByRole('button', { name: '重新提交审批' })).toBeEnabled()
  else expect(screen.getByRole('button', { name: '重新提交审批' })).toBeDisabled()
})
