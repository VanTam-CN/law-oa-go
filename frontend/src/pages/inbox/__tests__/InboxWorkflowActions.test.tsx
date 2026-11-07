import React from 'react'
import { App } from 'antd'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router'
import InboxList from '../InboxList'

jest.mock('@/utils/storage', () => ({ getToken: () => 'test' }))
const workflow = { resource_type: 'approval', resource_id: 'approval-uuid', action_code: 'decide_approval', action_label: '办理审批', completion_mode: 'workflow', workflow_status: 'submitted', action_version: 'v1' }
let calls: string[]
let legacy = false
let resultStatus = ''
beforeEach(() => {
  calls = []; legacy = false; resultStatus = ''
  global.fetch = jest.fn(async (raw, options) => {
    const path = String(raw); calls.push(path)
    const row = { id: 1, title: '协作事项', content: '', source_type: 'approval', source_id: 0, is_read: true, is_completed: false, priority: 'medium', created_at: '2026-11-07', completion_mode: legacy ? 'legacy_unresolved' : 'workflow', workflow_action: legacy ? { ...workflow, completion_mode: 'legacy_unresolved', blocked_code: 'LEGACY_SOURCE_UNRESOLVED' } : resultStatus ? { ...workflow, action_code: 'view_only', action_label: '查看状态', workflow_status: resultStatus, blocked_code: 'CASE_CREATION_PENDING', blocked_reason: '请查看成案进度' } : workflow }
    let data: any = {}
    if (path.includes('/inbox?')) data = { items: [row], pagination: { total: 1 } }
    else if (path === '/api/v1/inbox/1' && !options?.method) data = { ...row, workflow_action: { ...workflow, blocked_code: 'TASK_SUPERSEDED', blocked_reason: '事项已转交' } }
    return { ok: true, json: async () => ({ success: true, data }) } as Response
  }) as jest.Mock
})
afterEach(cleanup)
function mount() {
  render(<App><MemoryRouter initialEntries={['/inbox']}><Routes><Route path='/inbox' element={<InboxList />} /><Route path='/approval/:id' element={<p>精确审批目标</p>} /></Routes></MemoryRouter></App>)
}
it('uses the authorized string target and never offers manual business completion', async () => {
  mount()
  fireEvent.click(await screen.findByRole('button', { name: '办理审批' }))
  await screen.findByText('精确审批目标')
  expect(calls.some((path) => path.endsWith('/complete'))).toBe(false)
})
it('re-reads the detail action and disables a transferred task', async () => {
  mount()
  fireEvent.doubleClick((await screen.findByText('协作事项')).closest('tr')!)
  const dialog = await screen.findByRole('dialog')
  await within(dialog).findByText('事项已转交')
  expect(within(dialog).getByRole('button', { name: '办理审批' })).toBeDisabled()
  expect(within(dialog).queryByRole('button', { name: '标记完成' })).not.toBeInTheDocument()
  expect(calls).toContain('/api/v1/inbox/1')
})
it('archives an unresolved reminder through the archive endpoint, never complete', async () => {
  legacy = true
  mount()
  fireEvent.click(await screen.findByRole('button', { name: '归档提醒：协作事项' }))
  await waitFor(() => expect(calls).toContain('/api/v1/inbox/1/archive'))
  expect(calls).not.toContain('/api/v1/inbox/1/complete')
})

it.each([['approved_pending_case', '审批已通过，待成案'], ['case_creation_failed', '成案失败，可在审批详情重试']])('allows authorized read-only navigation for %s and displays the case creation substatus', async (status, label) => {
  resultStatus = status
  mount()
  await screen.findByText(label)
  const view = screen.getByRole('button', { name: '查看状态' })
  expect(view).toBeEnabled()
  fireEvent.click(view)
  await screen.findByText('精确审批目标')
  expect(calls.some((path) => path.endsWith('/complete'))).toBe(false)
})
