import React from 'react'
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { CaseIntakeWorkbench } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  getRoles: jest.fn(() => []),
  getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({ id: 7, role: 'lawyer', roles: ['lawyer'], name: '律师甲' })),
}))

jest.mock('@/utils/messageHelper', () => ({
  message: { error: jest.fn(), success: jest.fn(), warning: jest.fn(), info: jest.fn() },
  setAppMessage: jest.fn(),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function jsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

function listClientsHandler() {
  return Promise.resolve(
    jsonResponse({ clients: [{ id: 42, name: '库内客户乙', type: '企业' }] }),
  )
}

describe('CaseIntakeWorkbench blank-intake clue prefill and server draft collapse', () => {
  beforeEach(() => {
    window.localStorage.clear()
    window.sessionStorage.clear()
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/clients?page=1&page_size=50')) {
        return listClientsHandler()
      }
      if (path.endsWith('/clients') && method === 'POST') {
        return Promise.resolve(jsonResponse({ id: 99, name: '线索客户甲' }))
      }
      if (path.endsWith('/lawfirm/lawyers?page=1&page_size=50')) {
        return Promise.resolve(jsonResponse({ lawyers: [{ id: 7, name: '律师甲' }] }))
      }
      if (path.endsWith('/case-intakes?page=1&page_size=10')) {
        return Promise.resolve(jsonResponse({ items: [], total: 0 }))
      }
      return Promise.resolve(jsonResponse({}))
    })
  })

  it('prefills the inline client modal with the clue name and never creates or binds without confirmation', async () => {
    window.sessionStorage.setItem(
      'lawoa-intake-prefill-v1',
      JSON.stringify({
        title: '线索带入案件',
        clientName: '线索客户甲',
        caseType: 'commercial',
        partyNote: '',
        description: '线索事实摘要',
      }),
    )
    render(
      <MemoryRouter initialEntries={['/case/create']}>
        <CaseIntakeWorkbench />
      </MemoryRouter>,
    )

    // Clue facts land in the blank form; the client stays unresolved.
    const title = await screen.findByLabelText('案件名称 *')
    expect(title).toHaveValue('线索带入案件')
    expect(screen.getByLabelText('案情摘要 *')).toHaveValue('线索事实摘要')
    expect(screen.getByTestId('intake-clue-client-hint').textContent).toContain('线索客户甲')

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: /新建客户/ }))
    })
    const dialog = await screen.findByRole('dialog')
    const nameInput = await within(dialog).findByLabelText('客户名称 *')
    expect(nameInput).toHaveValue('线索客户甲')

    // The prefilled name stays editable and nothing writes while open.
    await act(async () => {
      fireEvent.change(nameInput, { target: { value: '线索客户甲（确认后创建）' } })
    })
    expect(nameInput).toHaveValue('线索客户甲（确认后创建）')
    expect(
      fetchMock.mock.calls.filter(
        ([input, init]) =>
          String(input).endsWith('/clients') && String(init?.method || '').toUpperCase() === 'POST',
      ),
    ).toHaveLength(0)
    expect(screen.getByTestId('intake-clue-client-hint').textContent).toContain('线索客户甲')

    // Cancel keeps the form summary untouched and still binds nothing.
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    })
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(screen.getByLabelText('案件名称 *')).toHaveValue('线索带入案件')
    expect(screen.getByLabelText('案情摘要 *')).toHaveValue('线索事实摘要')
    expect(screen.getByTestId('intake-clue-client-hint').textContent).toContain('线索客户甲')
    expect(
      fetchMock.mock.calls.filter(
        ([input, init]) =>
          String(input).endsWith('/clients') && String(init?.method || '').toUpperCase() === 'POST',
      ),
    ).toHaveLength(0)
  })

  it('collapses the server draft list by default and toggling never overwrites typed input', async () => {
    fetchMock.mockImplementation((input: RequestInfo | URL, init?: RequestInit) => {
      const path = String(input)
      const method = String(init?.method || 'GET').toUpperCase()
      if (path.endsWith('/clients?page=1&page_size=50')) {
        return listClientsHandler()
      }
      if (path.endsWith('/clients') && method === 'POST') {
        return Promise.resolve(jsonResponse({ id: 99, name: '不该被创建' }))
      }
      if (path.endsWith('/lawfirm/lawyers?page=1&page_size=50')) {
        return Promise.resolve(jsonResponse({ lawyers: [{ id: 7, name: '律师甲' }] }))
      }
      if (path.endsWith('/case-intakes?page=1&page_size=10')) {
        return Promise.resolve(
          jsonResponse({
            items: [
              { id: 'd1', intake_code: 'INT-D1', title: '服务端草稿一', revision: 'rev-d1' },
              { id: 'd2', intake_code: 'INT-D2', title: '服务端草稿二', revision: 'rev-d2' },
              { id: 'd3', intake_code: 'INT-D3', title: '服务端草稿三', revision: 'rev-d3' },
            ],
            total: 3,
          }),
        )
      }
      return Promise.resolve(jsonResponse({}))
    })
    render(
      <MemoryRouter initialEntries={['/case/create']}>
        <CaseIntakeWorkbench />
      </MemoryRouter>,
    )

    // Collapsed by default: only the count header is visible.
    const toggle = await screen.findByRole('button', { name: /服务端草稿（3\/3）/ })
    expect(toggle).toHaveAttribute('aria-expanded', 'false')
    expect(document.querySelector('#server-drafts-list')).toBeNull()
    expect(screen.queryByText('恢复此草稿')).toBeNull()

    await act(async () => {
      fireEvent.click(toggle)
    })
    expect(toggle).toHaveAttribute('aria-expanded', 'true')
    expect(await screen.findByText('服务端草稿一')).toBeInTheDocument()

    // Typing then collapsing/expanding must not clobber the lawyer's input.
    const title = screen.getByLabelText('案件名称 *')
    await act(async () => {
      fireEvent.change(title, { target: { value: '我正在填写的内容' } })
    })
    await act(async () => {
      fireEvent.click(toggle)
    })
    await act(async () => {
      fireEvent.click(toggle)
    })
    expect(screen.getByLabelText('案件名称 *')).toHaveValue('我正在填写的内容')
    expect(screen.getByText('服务端草稿一')).toBeInTheDocument()
  })
})
