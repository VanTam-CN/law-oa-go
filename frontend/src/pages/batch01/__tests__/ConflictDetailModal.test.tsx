import React from 'react'
import fs from 'node:fs'
import path from 'node:path'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { ConfigProvider } from 'antd'
import { antdTheme } from '@/config/theme'
import { ConflictCheckResults } from '../Batch01Prototype'

let mockRole = 'conflict_officer'
jest.mock('@/utils/storage', () => ({ getToken: () => 'fictional-token', getRoles: () => [], getUserInfo: () => ({ id: '7', role: mockRole }) }))
jest.mock('@/utils/messageHelper', () => ({ message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() } }))

let style: HTMLStyleElement
let requests: Array<{ path: string; body: unknown }>
let intakeSource = false
let automaticClear = false
let candidateID = 11
let assignment: Record<string, any>
let review: Record<string, any>
let includeUnopenedCheck = false
beforeAll(async () => {
  // Jest normally mocks stylesheet imports. Compile the real page LESS so the
  // scrolling rules are checked with the real portal, not a mocked Modal.
  const less = require('less')
  const pageCSS = await less.render(fs.readFileSync(path.resolve(__dirname, '../Batch01Prototype.less'), 'utf8'))
  style = document.createElement('style')
  style.textContent = fs.readFileSync(path.resolve(__dirname, '../../../assets/styles/design-tokens.css'), 'utf8') + pageCSS.css
  document.head.append(style)
})
afterAll(() => style.remove())
afterEach(cleanup)
beforeEach(() => {
  requests = []; mockRole = 'conflict_officer'; intakeSource = false; automaticClear = false; candidateID = 11; assignment = {}; review = {}; includeUnopenedCheck = false
  global.fetch = jest.fn(async (raw, options) => {
    const url = String(raw)
    let data: any = {}
    if (url.includes('command-center')) data = {
      summary: {}, risk_queue: [{ id: 'check-fictional', intake_id: intakeSource ? 'intake-fictional' : undefined, title: '长冲突详情', description: '用于复核的虚构说明。'.repeat(400), status: 'COMPLETED', has_conflict: false, risk_level: automaticClear ? 'LOW' : 'REVIEW_REQUIRED', check_result: { decision: { status: automaticClear ? 'CLEAR' : 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' } } }, ...(includeUnopenedCheck ? [{ id: 'check-unopened', title: '未打开的另一条检测', status: 'COMPLETED', risk_level: 'REVIEW_REQUIRED', check_result: { decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COMPLETE' } } }] : [])],
    }
    else if (url.endsWith('/conflict/reviewer-candidates')) data = [{ id: candidateID, name: '独立复核乙', role: 'conflict_officer' }]
    else if (options?.method === 'POST') {
      requests.push({ path: url, body: JSON.parse(String(options.body)) })
      if (url.endsWith('/reviewer-assignment')) { assignment = { id: 'assignment-fictional', reviewer_id: candidateID }; data = assignment }
      else if (url.endsWith('/review')) { review = { id: 'review-fictional', checkId: 'check-fictional', decision: 'no_conflict' }; data = review }
    }
    else if (url.endsWith('/reviewer-assignment')) data = assignment
    else if (url.endsWith('/review')) data = review
    return { ok: true, status: 200, json: async () => ({ data }) } as Response
  }) as jest.Mock
})

it('updates only the opened queue row from its already-fetched independent review, without a review request per row', async () => {
  intakeSource = true; includeUnopenedCheck = true
  review = { id: 'review-existing', checkId: 'check-fictional', decision: 'no_conflict' }
  render(<MemoryRouter initialEntries={['/conflict?task_id=check-fictional']}><ConflictCheckResults /></MemoryRouter>)
  const row = (await screen.findByText('长冲突详情', { selector: 'td' })).closest('tr')!
  await waitFor(() => expect(within(row).getAllByRole('cell')[1]).toHaveTextContent(/^未发现可识别冲突$/))
  const unopened = screen.getByText('未打开的另一条检测', { selector: 'td' }).closest('tr')!
  expect(within(unopened).getAllByText('待人工复核').length).toBeGreaterThan(0)
  const reviewReads = (global.fetch as jest.Mock).mock.calls.filter(([url, options]) => String(url).endsWith('/review') && !options?.method)
  expect(reviewReads.map(([url]) => url)).toEqual(['/api/v1/conflict/tasks/check-fictional/review'])
})

it('describes missing matches without inventing limited search coverage', async () => {
  intakeSource = true
  render(<MemoryRouter initialEntries={['/conflict?task_id=check-fictional']}><ConflictCheckResults /></MemoryRouter>)
  const dialog = await screen.findByRole('dialog', { name: '冲突检测详情' })
  expect(within(dialog).getByText('未发现匹配证据')).toBeInTheDocument()
  expect(within(dialog).queryByText(/当前仅因检索覆盖受限|检索覆盖完整性尚未确认/)).not.toBeInTheDocument()
})

it.each([false, true])('assigns and submits an intake independent review directly, including automatic-clear=%s', async (clear) => {
  intakeSource = true; automaticClear = clear; candidateID = 7
  render(<MemoryRouter initialEntries={['/conflict?task_id=check-fictional']}><ConflictCheckResults /></MemoryRouter>)
  const select = await screen.findByRole('combobox', { name: '选择独立复核人' })
  expect(screen.queryByRole('button', { name: '发起冲突审批' })).not.toBeInTheDocument()
  expect(screen.getByText('接案检测须先完成独立复核')).toBeInTheDocument()
  fireEvent.mouseDown(select)
  fireEvent.click(await screen.findByText('独立复核乙 · conflict_officer'))
  fireEvent.click(screen.getByRole('button', { name: '指定并声明回避' }))
  const conclusion = screen.getByRole('combobox', { name: '独立复核结论' })
  await waitFor(() => expect(conclusion).toBeEnabled())
  fireEvent.mouseDown(conclusion)
  fireEvent.click(await screen.findByText('无冲突'))
  fireEvent.change(screen.getByRole('textbox', { name: '独立复核依据' }), { target: { value: '核对虚构主体及检索结果，确认无冲突' } })
  fireEvent.click(screen.getByRole('button', { name: '提交人工复核结论' }))
  await waitFor(() => expect(requests).toHaveLength(2))
  expect(requests[0].path).toBe('/api/v1/conflict/tasks/check-fictional/reviewer-assignment')
  expect(requests[1]).toEqual({ path: '/api/v1/conflict/tasks/check-fictional/review', body: { decision: 'no_conflict', notes: '核对虚构主体及检索结果，确认无冲突' } })
  expect(requests.some((request) => request.path.endsWith('/approval'))).toBe(false)
})

it('guides the intake applicant to independent review without offering the legacy approval writer', async () => {
  intakeSource = true; mockRole = 'lawyer'
  render(<MemoryRouter initialEntries={['/conflict?task_id=check-fictional']}><ConflictCheckResults /></MemoryRouter>)
  await screen.findByRole('dialog', { name: '冲突检测详情' })
  expect(screen.queryByRole('button', { name: '发起冲突审批' })).not.toBeInTheDocument()
  expect(screen.getByText('请等待冲突核查岗在本页指定独立复核人并完成复核，无需先发起冲突审批。')).toBeInTheDocument()
  expect(requests).toHaveLength(0)
})

it('keeps the long detail and reviewer popup above fixed navigation, with a separately scrolling body', async () => {
  render(<ConfigProvider theme={antdTheme}><MemoryRouter initialEntries={['/conflict?task_id=check-fictional']}><ConflictCheckResults /></MemoryRouter></ConfigProvider>)
  const select = await screen.findByRole('combobox', { name: '选择独立复核人' })
  const dialog = screen.getByRole('dialog', { name: '冲突检测详情' })
  const overlay = dialog.closest('.ant-modal-wrap') as HTMLElement
  const fixedNavigationZ = Number(getComputedStyle(document.documentElement).getPropertyValue('--z-index-fixed'))
  expect(Number(getComputedStyle(overlay).zIndex)).toBeGreaterThan(fixedNavigationZ)
  expect(getComputedStyle(dialog.querySelector('.ant-modal-body')!).overflowY).toBe('auto')
  expect(getComputedStyle(dialog.querySelector('.ant-modal-content')!).maxHeight).toContain('100dvh')
  expect(getComputedStyle(dialog.querySelector('.batch-conflict-reviewer-assignment')!).flexWrap).toBe('wrap')

  fireEvent.mouseDown(select)
  const candidate = await screen.findByText('独立复核乙 · conflict_officer')
  const popup = candidate.closest('.ant-select-dropdown') as HTMLElement
  expect(Number(getComputedStyle(popup).zIndex)).toBeGreaterThan(Number(getComputedStyle(overlay).zIndex))
  fireEvent.click(candidate)
  fireEvent.click(within(dialog).getByRole('button', { name: '指定并声明回避' }))
  await waitFor(() => expect(requests).toEqual([{ path: '/api/v1/conflict/tasks/check-fictional/reviewer-assignment', body: expect.objectContaining({ reviewer_id: 11, recusal_declared: true }) }]))
  fireEvent.click(within(dialog).getByRole('button', { name: '关 闭' }))
  await waitFor(() => expect(screen.queryByRole('dialog', { name: '冲突检测详情' })).not.toBeInTheDocument())
})
