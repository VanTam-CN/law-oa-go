/** @jest-environment ./src/test/fetchJsdomEnvironment.cjs */
import React from 'react'
import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { CaseIntakeWorkbench, scopedCaseIntakeDraftKey } from '../Batch01Prototype'
import { useAppStore } from '@/stores/useAppStore'

jest.mock('@/utils/storage', () => ({ getToken: () => 'fictional-test', getRoles: () => [], getUserInfo: () => ({ id: '7', role: 'lawyer' }) }))
jest.mock('@/utils/messageHelper', () => ({ message: { success: jest.fn(), error: jest.fn(), warning: jest.fn(), info: jest.fn() } }))

const draft = {
  revision: 'rev-1',
  intake: { id: 'intake-one', intake_code: 'INT-ONE', title: '恢复案件', case_type: 'civil', client_id: 33, created_by: '7', updated_at: '2026-11-07T02:03:00Z', metadata: { lawyer_id: 7, business_area: '争议解决', sub_area: '合同争议' } },
  parties: [{ entity_name: '客户甲', party_role: 'client', entity_type: 'LEGAL_PERSON' }, { entity_name: '对方甲', party_role: 'opposing_party', entity_type: 'LEGAL_PERSON', identity_type: 'SOCIAL_CREDIT_CODE' }],
}
const approval = { id: 'approval-one', intake_id: 'intake-one', applicant_id: '7', status: 'needs_revision', active_submission_snapshot_id: 'snapshot-one', records: [{ id: 'return-one', decision: 'request_changes', created_at: '2026-11-07', decision_reason: '请补充说明', imposed_requirements: JSON.stringify({ schema_version: 1, supplement_requests: [{ item_id: 'item-one', category: 'matter_explanation', field_key: 'description', description: '说明事项', required: true }] }) }] }
let writes: Array<{ path: string; body: any }>
let failSave = false
let legacyRevision = false
let failResubmitOnce = false
let savePending: null | (() => void) = null

beforeEach(() => {
  writes = []; failSave = false; savePending = null; legacyRevision = false; failResubmitOnce = false
  localStorage.clear()
  useAppStore.setState({ isAuthenticated: true, user: { id: '7', username: 'lawyer', email: 'test@example.invalid', realName: '测试律师', roles: ['lawyer'], permissions: [], isActive: true, createdAt: '' } })
  global.fetch = jest.fn(async (raw, options) => {
    const path = String(raw)
    const ok = (data: any) => ({ ok: true, status: 200, json: async () => ({ success: true, data }) })
    if (options?.method === 'PUT' || options?.method === 'POST') {
      writes.push({ path, body: JSON.parse(String(options.body)) })
      if (failResubmitOnce && path.endsWith('/resubmit')) { failResubmitOnce = false; throw new Error('重新提交响应中断') }
      if (failSave) return { ok: false, status: 409, json: async () => ({ error: { code: 'INTAKE_DRAFT_REVISION_CONFLICT', message: '草稿已有更新', context: { current_revision: 'rev-new' } } }) } as Response
      if (savePending) await new Promise<void>((resolve) => { savePending = resolve })
      return ok({ id: 'intake-one', revision: 'rev-2', intake_code: 'INT-ONE' }) as Response
    }
    if (path === '/api/v1/case-intakes/intake-one') return ok(draft) as Response
    if (path.startsWith('/api/v1/case-intakes?')) return ok({ items: [], total: 0 }) as Response
    if (path === '/api/v1/approvals/approval-one/snapshot') return ok({ active_submission_snapshot_id: legacyRevision ? null : 'snapshot-one', viewed_snapshot_id: legacyRevision ? 'persistent-legacy-row-uuid' : 'snapshot-one', snapshot: { snapshot_id: 'SNAP_DO_NOT_USE' } }) as Response
    if (path === '/api/v1/approvals/approval-one') return ok(legacyRevision ? { ...approval, active_submission_snapshot_id: null, return_record_id: 'return-one', return_reason: '旧文字补充意见', supplement_requests: [{ item_id: 'legacy-return-one', category: 'matter_explanation', description: '旧文字补充意见', required: true }] } : approval) as Response
    if (path.startsWith('/api/v1/clients?')) return ok({ clients: [{ id: 33, name: '客户甲', type: '企业' }], pagination: { total: 1 } }) as Response
    if (path === '/api/v1/clients/33') return ok({ id: 33, name: '客户甲', type: '企业' }) as Response
    if (path.startsWith('/api/v1/lawfirm/lawyers')) return ok({ lawyers: [{ id: 7, name: '测试律师' }] }) as Response
    return ok({}) as Response
  }) as jest.Mock
})
afterEach(() => { cleanup(); useAppStore.setState({ user: null, isAuthenticated: false }) })

function mount(query = '?intake_id=intake-one') {
  const router = createMemoryRouter([{ path: '/case/create', element: <CaseIntakeWorkbench /> }, { path: '/case', element: <p>案件列表目标</p> }, { path: '/approval/:id', element: <p>原审批目标</p> }], { initialEntries: ['/case/create' + query] })
  render(<RouterProvider router={router} />)
  return router
}

it('restores the exact URL draft and preserves local edits on a stale save without exiting', async () => {
  mount()
  fireEvent.change(await screen.findByDisplayValue('恢复案件'), { target: { value: '本地修改仍保留' } })
  failSave = true
  fireEvent.click(screen.getByRole('button', { name: '保存草稿' }))
  await screen.findByText('保存被拒绝：服务端已有更新的草稿版本')
  expect(screen.getByDisplayValue('本地修改仍保留')).toBeInTheDocument()
  expect(screen.getByText('保存失败，输入仍保留')).toBeInTheDocument()
  expect(writes[0].path).toBe('/api/v1/case-intakes/intake-one')
  expect(writes[0].body.expected_revision).toBe('rev-1')
  expect(screen.queryByText('案件列表目标')).not.toBeInTheDocument()
})

it('does not mark edits made during saving as saved', async () => {
  mount()
  const title = await screen.findByDisplayValue('恢复案件')
  fireEvent.change(title, { target: { value: '首次修改' } })
  savePending = () => undefined
  fireEvent.click(screen.getByRole('button', { name: '保存草稿' }))
  await screen.findByText('正在保存')
  fireEvent.change(title, { target: { value: '保存中再次修改' } })
  await act(async () => { savePending?.() })
  await screen.findByText('未保存修改')
  expect(screen.getByDisplayValue('保存中再次修改')).toBeInTheDocument()
  expect(writes[0].body.title).toBe('首次修改')
})

it('blocks leaving and keeps the page open when save-before-leave fails', async () => {
  const router = mount()
  fireEvent.change(await screen.findByDisplayValue('恢复案件'), { target: { value: '未保存的修改' } })
  await act(async () => { void router.navigate('/case') })
  await screen.findByText('尚有未保存修改')
  failSave = true
  fireEvent.click(screen.getByRole('button', { name: '保存后离开' }))
  await waitFor(() => expect(writes).toHaveLength(1))
  expect(screen.getByDisplayValue('未保存的修改')).toBeInTheDocument()
  expect(screen.queryByText('案件列表目标')).not.toBeInTheDocument()
})

it('saves only controlled replies then resubmits the same approval with the rotated revision', async () => {
  mount('?intake_id=intake-one&approval_id=approval-one&mode=revision')
  await screen.findByText('说明事项')
  fireEvent.change(screen.getByLabelText('补正回复（必填）'), { target: { value: '已说明本轮事项' } })
  fireEvent.click(screen.getByRole('button', { name: '保存逐项回复' }))
  await waitFor(() => expect(writes).toHaveLength(1))
  expect(writes[0].body).toEqual({ approval_id: 'approval-one', return_record_id: 'return-one', expected_revision: 'rev-1', supplement_responses: [{ item_id: 'item-one', response: '已说明本轮事项' }] })
  fireEvent.change(screen.getByLabelText('修改说明（随重新提交保存）'), { target: { value: '已完成补充说明' } })
  fireEvent.click(screen.getByRole('button', { name: '重新提交原审批' }))
  await waitFor(() => expect(writes).toHaveLength(2))
  expect(writes[1].path).toBe('/api/v1/approvals/approval-one/resubmit')
  expect(writes[1].body.expected_intake_revision).toBe('rev-2')
  expect(writes[1].body.expected_submission_snapshot_id).toBe('snapshot-one')
  await screen.findByText('原审批目标')
})

it('never copies a previous account draft into the new account cache', async () => {
  mount('')
  fireEvent.change(await screen.findByLabelText('案件名称 *'), { target: { value: '账号七未保存内容' } })
  await act(async () => { useAppStore.setState({ user: { ...useAppStore.getState().user!, id: '8' } }) })
  await waitFor(() => expect(screen.queryByDisplayValue('账号七未保存内容')).not.toBeInTheDocument())
  expect(localStorage.getItem(scopedCaseIntakeDraftKey('8')) || '').not.toContain('账号七未保存内容')
  expect(localStorage.getItem(scopedCaseIntakeDraftKey('7'))).toContain('账号七未保存内容')
})

it('resubmits a normalized legacy return with the verified persisted snapshot id and preserves it on an uncertain retry', async () => {
  legacyRevision = true; failResubmitOnce = true
  mount('?intake_id=intake-one&approval_id=approval-one&mode=revision')
  await screen.findByLabelText('补正回复（必填）')
  fireEvent.change(screen.getByLabelText('补正回复（必填）'), { target: { value: '旧文字要求的逐项回复' } })
  fireEvent.change(screen.getByLabelText('修改说明（随重新提交保存）'), { target: { value: '补充旧申请说明' } })
  fireEvent.click(screen.getByRole('button', { name: '重新提交原审批' }))
  await waitFor(() => expect(writes.filter((write) => write.path.endsWith('/resubmit'))).toHaveLength(1))
  await waitFor(() => expect(screen.getByRole('button', { name: '重新提交原审批' })).not.toHaveClass('ant-btn-loading'))
  fireEvent.click(screen.getByRole('button', { name: '重新提交原审批' }))
  await screen.findByText('原审批目标')
  const submissions = writes.filter((write) => write.path.endsWith('/resubmit'))
  expect(submissions).toHaveLength(2)
  expect(submissions[0].body).toEqual(submissions[1].body)
  expect(submissions[0].body.expected_submission_snapshot_id).toBe('persistent-legacy-row-uuid')
  expect(submissions[0].body.supplement_responses).toEqual([{ item_id: 'legacy-return-one', response: '旧文字要求的逐项回复' }])
})
