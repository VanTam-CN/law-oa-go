import React from 'react'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, useNavigate } from 'react-router'
import { boundApprovalMaterialEvidence, CaseIntakeWorkbench } from '../Batch01Prototype'
import SupplementChecklist, { type SupplementItemFE } from '../intake/SupplementChecklist'

jest.mock('@/utils/storage', () => ({
  getRoles: jest.fn(() => []), getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({ id: 7, role: 'lawyer' })),
}))
jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

const items: SupplementItemFE[] = [
  { item_id: 'material-3', material_id: 3, category: 'material_request', field_key: 'material', description: '请换交清晰的主体资料', required: true },
  { item_id: 'title', category: 'field_correction', field_key: 'case_title', description: '请核对名称', required: false },
]
const evidence = (state: string, extra = {}) => ({ item_id: 'material-3', material_id: 3, state, current_file_id: 11, message: '内部评估消息不直接展示', ...extra })
const data = (state = 'missing') => ({ items, replies: [{ item_id: 'material-3', response: '已保存回复' }], material_evidence: [evidence(state)] })
const json = (body: unknown) => ({ ok: true, json: async () => ({ success: true, data: body }) }) as Response
const fetchMock = jest.fn()
global.fetch = fetchMock as typeof fetch
let returned: unknown
let deferredRead: (() => Promise<Response>) | undefined
let ledger: Record<string, unknown>
beforeEach(() => {
  window.localStorage.clear()
  Element.prototype.scrollIntoView = jest.fn()
  returned = data()
  deferredRead = undefined
  ledger = { id: 3, intake_id: 'I1', name: '客户主体资料', material_type: 'identity', required: true, status: 'uploaded',
    uploaded_by: 7, file_count: 1, file_version: 1,
    current_file: { id: 10, file_name: '原主体资料.pdf', size_bytes: 3, uploaded_by: 7, version: 1 } }
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
    const path = String(input)
    if (/\/case-intakes\/I[12]$/.test(path)) return Promise.resolve(json({
      intake: { id: path.endsWith('I2') ? 'I2' : 'I1', title: path.endsWith('I2') ? '第二份申请' : '第一份申请',
        client_id: 42, case_type: 'commercial', description: '原案情', metadata: { lawyer_id: 7 } },
      revision: 'rev-1', parties: [],
    }))
    if (/\/approvals\/A[12]$/.test(path)) return Promise.resolve(json({ active_submission_snapshot_id: 'S1', records: [
      { id: path.endsWith('A2') ? 'R2' : 'R1', decision: 'request_changes', imposed_requirements: { items } },
    ] }))
    if (path.endsWith('/snapshot')) return Promise.resolve(json({ approval_id: path.includes('/A2/') ? 'A2' : 'A1', snapshot_id: 'S1', snapshot: {} }))
    if (path.includes('/return-items?')) return deferredRead ? deferredRead() : Promise.resolve(json(returned))
    if (path.endsWith('/materials')) return Promise.resolve(json({ materials: [ledger] }))
    if (path.endsWith('/case-materials/3/files') && init?.method === 'POST') {
      ledger = { ...ledger, file_count: 2, file_version: 2,
        current_file: { id: 11, file_name: '新主体资料.pdf', size_bytes: 7, uploaded_by: 7, version: 2 } }
      returned = data('complete')
      return Promise.resolve(json({}))
    }
    return Promise.resolve(json({}))
  })
})

it.each([
  { items, material_evidence: [] },
  { items, material_evidence: [evidence('uploaded')] },
  { items, material_evidence: [evidence('complete', { material_id: 99 })] },
  { items, material_evidence: [evidence('complete'), evidence('complete')] },
  { items: [], material_evidence: [evidence('complete')] },
  { items: [{ ...items[0], material_id: 99 }], material_evidence: [evidence('complete')] },
  { items, material_evidence: [evidence('complete', { current_file_id: undefined })] },
  { items, material_evidence: [evidence('complete', { current_file_id: 0 })] },
  { items, material_evidence: [evidence('complete', { current_file_id: -1 })] },
  { items, material_evidence: [evidence('complete', { current_file_id: 1.5 })] },
  { items, material_evidence: [evidence('complete', { current_file_id: '11' })] },
])('does not bind absent, ambiguous or mismatched evidence %#', (response) => {
  expect(boundApprovalMaterialEvidence(items, response)).toEqual({})
})

it.each(['complete', 'missing', 'unknown'])('renders %s only as the trusted correction state, not business completion', (state) => {
  render(<SupplementChecklist items={items} answers={{ 'material-3': '我已回复' }} onChange={() => undefined}
    materialStatuses={{ 3: 'uploaded' }} materialFileNames={{ 3: '原主体资料.pdf' }}
    materialEvidence={boundApprovalMaterialEvidence(items, data(state))} />)
  expect(screen.getByTestId('supplement-status-material-3')).toHaveTextContent(
    state === 'complete' ? '补正附件已上传，待审批复核' : state === 'missing' ? '已有旧附件，尚未补交新附件' : '未能确认补正，请刷新材料状态或联系项目组',
  )
  expect(screen.queryByText('内部评估消息不直接展示')).not.toBeInTheDocument()
  expect(screen.getByText('已回复')).toBeInTheDocument()
})

function SwitchIntake() {
  const navigate = useNavigate()
  return <button onClick={() => navigate('/case/create?mode=revision&intake_id=I2&approval_id=A2')}>切换申请</button>
}
async function open() {
  await act(async () => {
    render(<MemoryRouter initialEntries={['/case/create?mode=revision&intake_id=I1&approval_id=A1']}>
      <CaseIntakeWorkbench /><SwitchIntake />
    </MemoryRouter>)
  })
}

it('reads the current returned round and refreshes evidence without overwriting replies or facts', async () => {
  returned = data('complete')
  await open()
  expect(screen.getByTestId('supplement-status-material-3')).toHaveTextContent('补正附件已上传，待审批复核')
  expect(fetchMock.mock.calls.some(([path]) => String(path).endsWith('/approvals/A1/return-items?return_record_id=R1'))).toBe(true)
  fireEvent.change(screen.getByLabelText('回复：客户主体资料'), { target: { value: '我正在补充的回复' } })
  fireEvent.change(screen.getByLabelText('案情摘要 *'), { target: { value: '我正在修改的案情' } })
  returned = data('missing')
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '刷新材料状态' })) })
  expect(screen.getByTestId('supplement-status-material-3')).toHaveTextContent('已有旧附件，尚未补交新附件')
  expect(screen.getByLabelText('回复：客户主体资料')).toHaveValue('我正在补充的回复')
  expect(screen.getByLabelText('案情摘要 *')).toHaveValue('我正在修改的案情')
})

it('keeps old uploaded attachments and replies unconfirmed when evidence is missing or the read fails', async () => {
  returned = { items, replies: [{ item_id: 'material-3', response: '已回复' }] }
  await open()
  expect(screen.getByTestId('supplement-status-material-3')).toHaveTextContent('未能确认补正')
  expect(screen.getByLabelText('回复：客户主体资料')).toHaveValue('已回复')
  deferredRead = async () => { throw new Error('offline') }
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '刷新材料状态' })) })
  expect(screen.getByTestId('supplement-status-material-3')).toHaveTextContent('未能确认补正')
  expect(screen.getByLabelText('回复：客户主体资料')).toHaveValue('已回复')
})

it('ignores a refresh arriving after another refresh', async () => {
  await open()
  let resolveOld!: (response: Response) => void
  deferredRead = () => new Promise((resolve) => { resolveOld = resolve })
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '刷新材料状态' })) })
  deferredRead = undefined
  returned = data('unknown')
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '刷新材料状态' })) })
  await act(async () => { resolveOld(json(data('complete'))) })
  expect(screen.getByTestId('supplement-status-material-3')).toHaveTextContent('未能确认补正')
})

it('ignores evidence for the previous intake/approval/return round', async () => {
  await open()
  let resolveOld!: (response: Response) => void
  deferredRead = () => new Promise((resolve) => { resolveOld = resolve })
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '刷新材料状态' })) })
  deferredRead = undefined
  returned = data('unknown')
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '切换申请' })) })
  await act(async () => { resolveOld(json(data('complete'))) })
  expect(screen.getByLabelText('案件名称 *')).toHaveValue('第二份申请')
  expect(screen.getByTestId('supplement-status-material-3')).toHaveTextContent('未能确认补正')
  expect(fetchMock.mock.calls.some(([path]) => String(path).endsWith('/approvals/A2/return-items?return_record_id=R2'))).toBe(true)
})

it('the real ledger upload causes a fresh trusted return-item read and keeps typed replies', async () => {
  await open()
  fireEvent.change(screen.getByLabelText('回复：客户主体资料'), { target: { value: '新附件的补充说明' } })
  await act(async () => { fireEvent.click(document.querySelectorAll('.batch-stepper button')[3]) })
  const row = await screen.findByTestId('material-row-3')
  const input = row.querySelector<HTMLInputElement>('input[type="file"]')
  expect(input).not.toBeNull()
  const readsBefore = fetchMock.mock.calls.filter(([path]) => String(path).includes('/return-items?')).length
  await act(async () => {
    fireEvent.change(input!, { target: { files: [new File(['newdata'], '新主体资料.pdf', { type: 'application/pdf' })] } })
  })
  await waitFor(() => expect(screen.getByTestId('supplement-status-material-3')).toHaveTextContent('补正附件已上传，待审批复核'))
  expect(fetchMock.mock.calls.filter(([path]) => String(path).includes('/return-items?')).length).toBeGreaterThan(readsBefore)
  expect(screen.getByLabelText('回复：客户主体资料')).toHaveValue('新附件的补充说明')
})

it('requires confirmed attachment evidence even when the material reply is optional', async () => {
  const originalFetch = fetchMock.getMockImplementation()!
  const optionalItem = { ...items[0], required: false }
  returned = { items: [optionalItem], replies: [], material_evidence: [evidence('missing')] }
  fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
    const path = String(input)
    if (path.endsWith('/approvals/A1')) return Promise.resolve(json({ active_submission_snapshot_id: 'S1', records: [
      { id: 'R1', decision: 'request_changes', imposed_requirements: { items: [optionalItem] } },
    ] }))
    if (path.endsWith('/continuation')) return Promise.resolve(json({
      intake_id: 'I1', revision: 'rev-1', status: 'conflict_ready', can_reuse_identity: true, can_reuse_check: true,
      can_submit: true, recheck_required: false, reasons: [], facts: { facts_version: 'fv1', confirmed_at: '2026-10-02' },
      conflict_binding: { check_id: 'C1', check_status: 'COMPLETED', coverage_status: 'COMPLETE', review_id: 11, reviewer_id: 9, review_decision: 'no_conflict' },
      approval: { approval_id: 'A1', status: 'needs_revision', snapshot_id: 'S1' },
    }))
    return originalFetch(input, init)
  })
  await open()
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeDisabled()
  expect(screen.getByRole('textbox', { name: '回复：客户主体资料' })).toHaveValue('')
  expect(screen.getByText(/以下补正附件尚未确认/)).toHaveTextContent('第 4 步上传补正附件')
  returned = { items: [optionalItem], replies: [], material_evidence: [evidence('complete')] }
  await act(async () => { fireEvent.click(screen.getByRole('button', { name: '刷新材料状态' })) })
  expect(screen.getByRole('button', { name: '重新提交审批' })).toBeEnabled()
  expect(screen.getByRole('textbox', { name: '回复：客户主体资料' })).toHaveValue('')
  expect(fetchMock.mock.calls.some(([path]) => /facts-confirmation|conflict-check/.test(String(path)))).toBe(false)
})
