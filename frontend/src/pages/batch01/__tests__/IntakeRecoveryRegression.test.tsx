/** @jest-environment ./src/test/fetchJsdomEnvironment.cjs */
import React from 'react'
import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { CaseIntakeWorkbench, intakeDraftFromServer, scopedCaseIntakeDraftKey } from '../Batch01Prototype'
import { useAppStore } from '@/stores/useAppStore'

jest.mock('@/utils/storage', () => ({ getToken: () => 'fictional', getRoles: () => [], getUserInfo: () => ({ id: '7', role: 'lawyer' }) }))
jest.mock('@/utils/messageHelper', () => ({ message: { error: jest.fn(), success: jest.fn(), info: jest.fn(), warning: jest.fn() } }))
const draft = (id: string) => ({ revision: 'rev-' + id, intake: { id, title: '草稿' + id, intake_code: 'INT-' + id, client_id: 33, case_type: 'civil', description: '完整虚构事项说明', created_by: '7', status: 'conflict_ready', metadata: { lawyer_id: 7, business_area: '争议解决', sub_area: '合同争议', lawyer_facts_version: 'facts-' + id, conflict_check_id: 'check-' + id } }, parties: [{ entity_name: '客户甲', party_role: 'client', entity_type: 'LEGAL_PERSON' }, { entity_name: '对方甲', party_role: 'opposing_party', entity_type: 'LEGAL_PERSON', identity_type: 'SOCIAL_CREDIT_CODE', identity_status: '已登记（受保护）' }] })
const result = (id: string) => ({ task: { checkId: 'check-' + id, intakeId: id, clientId: '33', status: 'COMPLETED' }, result: { checkId: 'check-' + id, decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' }, riskAssessment: { overallRisk: 'REVIEW_REQUIRED' } } })
let calls: Array<{ path: string; method: string; body: any }>
let pending: Map<string, Promise<any>>
let serverDraft: any
let invalidResult = false
let failReview = false
let failSave = false
let wrongReviewBinding = false
let automaticClearWithoutReview = false
function defer(id: string) {
  let resolve: (value: unknown) => void = () => undefined
  pending.set('/api/v1/case-intakes/' + id, new Promise((done) => { resolve = done }))
  return () => resolve(draft(id))
}
beforeEach(() => {
  calls = []; pending = new Map(); serverDraft = null; invalidResult = false; failReview = false; failSave = false; wrongReviewBinding = false; automaticClearWithoutReview = false
  localStorage.clear()
  useAppStore.setState({ isAuthenticated: true, user: { id: '7', username: 'lawyer', email: '', realName: '律师甲', roles: ['lawyer'], permissions: [], isActive: true, createdAt: '' } })
  global.fetch = jest.fn(async (raw, options) => {
    const path = String(raw), method = options?.method || 'GET'
    calls.push({ path, method, body: options?.body ? JSON.parse(String(options.body)) : null })
    if (method === 'PUT' && failSave) return { ok: false, status: 409, json: async () => ({ error: { code: 'INTAKE_DRAFT_REVISION_CONFLICT', message: '草稿已有更新', context: { current_revision: 'rev-new' } } }) } as Response
    let data: any = {}
    if (method === 'POST' && path === '/api/v1/case-intakes') data = { id: 'new-B', revision: 'rev-new-B', intake_code: 'INT-NEW-B' }
    else if (pending.has(path)) data = await pending.get(path)
    else if (path.startsWith('/api/v1/case-intakes?')) data = { items: ['A', 'B'].map((id) => ({ ...draft(id).intake, revision: 'rev-' + id })), total: 2 }
    else if (/\/case-intakes\/[AB]$/.test(path)) data = serverDraft || draft(path.slice(-1))
    else if (path.startsWith('/api/v1/clients?')) data = { clients: [{ id: 33, name: '客户甲', type: '企业' }], pagination: { total: 1 } }
    else if (path === '/api/v1/clients/33') data = { id: 33, name: '客户甲', type: '企业' }
    else if (path.startsWith('/api/v1/lawfirm/lawyers')) data = { lawyers: [{ id: 7, name: '律师甲' }] }
    else if (path.endsWith('/result')) { data = result(path.includes('check-B') ? 'B' : 'A'); if (invalidResult) data.task.intakeId = 'other-intake'; if (automaticClearWithoutReview) data.result.decision.status = 'CLEAR' }
    else if (path.endsWith('/review')) { if (failReview) throw new Error('复核读取失败'); data = automaticClearWithoutReview ? {} : { id: 'review-current', checkId: wrongReviewBinding ? 'check-other' : path.includes('check-B') ? 'check-B' : 'check-A', decision: 'no_conflict', notes: '虚构核验依据', reviewerId: 8, reviewerName: '复核乙', evidenceHash: 'fictional-hash', createdAt: '2026-11-07T07:00:00Z' } }
    else if (path === '/api/v1/integration/approvals/with-conflict') data = { approval_id: 'approval-A' }
    return { ok: true, status: 200, json: async () => ({ data }) } as Response
  }) as jest.Mock
})
afterEach(() => { cleanup(); useAppStore.setState({ user: null, isAuthenticated: false }) })
function mount(query = '?intake_id=A') {
  const router = createMemoryRouter([{ path: '/case/create', element: <CaseIntakeWorkbench /> }, { path: '/approval/:id', element: <p>原对象审批目标</p> }], { initialEntries: ['/case/create' + query] })
  render(<RouterProvider router={router} />)
  return router
}
const writes = () => calls.filter((call) => call.method !== 'GET')

it('re-reads the bound check and live review, submits without protected identity re-entry or a redundant fact save/check', async () => {
  mount()
  const submit = await screen.findByRole('button', { name: '提交审批并等待成案' })
  await waitFor(() => expect(submit).toBeEnabled())
  expect(screen.getByLabelText('身份标识')).toHaveValue('')
  fireEvent.click(screen.getByRole('button', { name: '保存草稿' }))
  expect(writes()).toHaveLength(0)
  fireEvent.click(submit)
  await screen.findByText('原对象审批目标')
  expect(writes()).toHaveLength(1)
  expect(writes()[0]).toEqual(expect.objectContaining({ path: '/api/v1/integration/approvals/with-conflict', body: expect.objectContaining({ metadata: expect.objectContaining({ intake_id: 'A', conflict_result: expect.objectContaining({ checkId: 'check-A' }) }) }) }))
  expect(calls.some((call) => call.path.includes('/conflict/history'))).toBe(false)
  expect(Array.from({ length: localStorage.length }, (_, i) => localStorage.key(i)).some((key) => key?.endsWith(':conflictFingerprint'))).toBe(false)
})
it('invalidates the restored check immediately when the lawyer changes a fact', async () => {
  mount()
  const submit = await screen.findByRole('button', { name: '提交审批并等待成案' })
  await waitFor(() => expect(submit).toBeEnabled())
  fireEvent.change(screen.getByDisplayValue('草稿A'), { target: { value: '已改变的案件' } })
  expect(submit).toBeDisabled()
  expect(screen.getAllByText('结果已过期').length).toBeGreaterThan(0)
  expect(writes()).toHaveLength(0)
})
it('requires independent review even when the restored intake detection is automatically clear', async () => {
  automaticClearWithoutReview = true
  mount()
  await screen.findByRole('button', { name: '进入本案冲突复核' })
  expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
  expect(screen.getAllByText('接案须先完成独立复核，请进入本案冲突复核').length).toBeGreaterThan(0)
  expect(writes()).toHaveLength(0)
})
it('also invalidates a restored check for edited matter details outside the party fingerprint', async () => {
  mount()
  const submit = await screen.findByRole('button', { name: '提交审批并等待成案' })
  await waitFor(() => expect(submit).toBeEnabled())
  fireEvent.change(screen.getByDisplayValue('完整虚构事项说明'), { target: { value: '修改后的事项摘要' } })
  expect(submit).toBeDisabled()
  expect(writes()).toHaveLength(0)
})
it('does not revive a historical check when the authoritative draft is no longer conflict_ready', async () => {
  serverDraft = draft('A'); serverDraft.intake.status = 'draft'
  mount(); await screen.findByDisplayValue('草稿A')
  expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
  expect(calls.some((call) => call.path.endsWith('/result'))).toBe(false)
})
it.each(['mismatch', 'review-mismatch', 'unavailable'])('fails closed when restoring the check is %s, without creating another check', async (reason) => {
  invalidResult = reason === 'mismatch'; wrongReviewBinding = reason === 'review-mismatch'; failReview = reason === 'unavailable'
  mount()
  await screen.findByText(reason === 'unavailable' ? '复核读取失败' : '检测结果与当前草稿关联不一致，请刷新后核对')
  expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeDisabled()
  expect(writes()).toHaveLength(0)
})
it('keeps the latest selected draft when two restore responses arrive in reverse order', async () => {
  const finishA = defer('A'), finishB = defer('B')
  mount('')
  const buttons = await screen.findAllByRole('button', { name: '恢复此草稿' })
  fireEvent.click(buttons[0]); fireEvent.click(buttons[1])
  await act(async () => { finishB() }); await screen.findByDisplayValue('草稿B')
  await act(async () => { finishA() })
  expect(screen.getByDisplayValue('草稿B')).toBeInTheDocument()
  expect(screen.queryByDisplayValue('草稿A')).not.toBeInTheDocument()
})
it('does not let an in-flight restore overwrite input entered while it was loading', async () => {
  const finishA = defer('A'); mount('')
  fireEvent.click((await screen.findAllByRole('button', { name: '恢复此草稿' }))[0])
  fireEvent.change(screen.getByLabelText('案件名称 *'), { target: { value: '读取期间的新输入' } })
  await act(async () => { finishA() })
  expect(screen.getByDisplayValue('读取期间的新输入')).toBeInTheDocument()
  expect(screen.queryByDisplayValue('草稿A')).not.toBeInTheDocument()
})
it('drops a late previous-account restore without putting its facts into the new account cache', async () => {
  const finishA = defer('A'); mount('')
  fireEvent.click((await screen.findAllByRole('button', { name: '恢复此草稿' }))[0])
  await act(async () => { useAppStore.setState({ user: { ...useAppStore.getState().user!, id: '8' } }) })
  await act(async () => { finishA() })
  expect(screen.getByLabelText('案件名称 *')).toHaveValue('')
  expect(localStorage.getItem(scopedCaseIntakeDraftKey('8')) || '').not.toContain('草稿A')
})
it('guards a server-backed local cache restore against an account change while it is loading', async () => {
  localStorage.setItem(scopedCaseIntakeDraftKey('7'), JSON.stringify(intakeDraftFromServer(draft('A')).form))
  const finishA = defer('A'); mount('')
  fireEvent.click(await screen.findByRole('button', { name: '继续未完成草稿' }))
  await act(async () => { useAppStore.setState({ user: { ...useAppStore.getState().user!, id: '8' } }) })
  await act(async () => { finishA() })
  expect(screen.getByLabelText('案件名称 *')).toHaveValue('')
  expect(localStorage.getItem(scopedCaseIntakeDraftKey('8')) || '').not.toContain('草稿A')
})
it('preserves edits made while an explicit 409 reload is still in flight', async () => {
  mount()
  await waitFor(() => expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeEnabled())
  fireEvent.change(screen.getByDisplayValue('草稿A'), { target: { value: '首次本地编辑' } })
  failSave = true
  fireEvent.click(screen.getByRole('button', { name: '保存草稿' }))
  const reload = await screen.findByRole('button', { name: '重新加载服务端草稿' })
  const finishA = defer('A')
  fireEvent.click(reload)
  fireEvent.change(screen.getByDisplayValue('首次本地编辑'), { target: { value: '重载期间的新输入' } })
  await act(async () => { finishA() })
  expect(screen.getByDisplayValue('重载期间的新输入')).toBeInTheDocument()
  expect(screen.queryByDisplayValue('草稿A')).not.toBeInTheDocument()
})
it('clears object and revision state on SPA navigation from a concrete draft to the blank entry', async () => {
  const router = mount()
  await screen.findByDisplayValue('草稿A')
  await waitFor(() => expect(screen.getByRole('button', { name: '提交审批并等待成案' })).toBeEnabled())
  await act(async () => { await router.navigate('/case/create') })
  expect(screen.getByLabelText('案件名称 *')).toHaveValue('')
  fireEvent.change(screen.getByLabelText('案件名称 *'), { target: { value: '新案件B' } })
  fireEvent.mouseDown(screen.getByRole('combobox', { name: '我方当事人（客户）' }))
  fireEvent.click(await screen.findByText('客户甲 · 企业 · 编号 33'))
  fireEvent.click(screen.getByRole('button', { name: /团队与费用/ }))
  fireEvent.mouseDown(screen.getByRole('combobox', { name: '负责律师 *' }))
  fireEvent.click(await screen.findByText('律师甲'))
  fireEvent.click(screen.getByRole('button', { name: '保存草稿' }))
  await waitFor(() => expect(writes()).toHaveLength(1))
  expect(writes()[0]).toEqual(expect.objectContaining({ method: 'POST', path: '/api/v1/case-intakes', body: expect.objectContaining({ title: '新案件B' }) }))
  expect(writes()[0].body.expected_revision).toBeUndefined()
  expect(writes().some((call) => call.path === '/api/v1/case-intakes/A')).toBe(false)
})
