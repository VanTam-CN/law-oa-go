import React from 'react'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import CaseTaskPanel from '../CaseTaskPanel'
import { CaseDetailCenter } from '../../../pages/batch01/Batch01Prototype'
import { useAppStore } from '@/stores/useAppStore'

jest.mock('@/utils/storage', () => ({
  getToken: () => 'test-token',
  getUserInfo: () => ({ id: 7, role: 'lawyer' }),
  getRoles: () => null,
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

const today = new Date().toISOString()

function taskRow(overrides: Record<string, unknown> = {}) {
  return {
    id: 21,
    case_id: 12,
    title: '整理证据目录',
    assignee_id: 7,
    created_by: 7,
    status: 'open',
    capabilities: { can_adjust: true, can_cancel: true, can_complete: true, can_view_changes: true },
    ...overrides,
  }
}

function setupFetch(
  rows: Record<string, unknown>[],
  rowsByCase: Record<number, Record<string, unknown>[]> = {},
) {
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/cases/12/tasks')) {
      return Promise.resolve(jsonResponse({ tasks: rows }))
    }
    for (const [caseKey, caseRows] of Object.entries(rowsByCase)) {
      if (path.endsWith('/cases/' + caseKey + '/tasks')) {
        return Promise.resolve(jsonResponse({ tasks: caseRows }))
      }
    }
    return Promise.resolve(jsonResponse({ changes: [] }))
  })
}

const assigneeOptions = [{ id: 7, name: '律师甲' }]

function renderPanel(props: Partial<React.ComponentProps<typeof CaseTaskPanel>> = {}) {
  return render(
    <CaseTaskPanel
      caseID={12}
      assigneeOptions={assigneeOptions}
      currentUserId={7}
      {...props}
    />,
  )
}

async function findFocusedRow(rowID: number) {
  return waitFor(() => {
    const el = document.querySelector("[data-testid='task-row-" + rowID + "']") as HTMLElement | null
    expect(el).not.toBeNull()
    expect(el).toHaveClass('task-row-focused')
    return el as HTMLElement
  })
}

async function expectNoFocusedRow() {
  await waitFor(() => {
    expect(screen.getByText('整理证据目录')).toBeInTheDocument()
    const rows = document.querySelectorAll('tr.task-row-focused')
    expect(rows.length).toBe(0)
  })
}

describe('CaseTaskPanel focus deep link', () => {
  afterEach(() => {
    cleanup()
  })

  it('expands and highlights the focused task after rows load', async () => {
    setupFetch([taskRow({ id: 21, title: '整理证据目录' }), taskRow({ id: 22, title: '联系客户' })])
    renderPanel({ focusTaskID: 21 })
    await screen.findByText('整理证据目录')
    await findFocusedRow(21)
    // The focus effect must also expand the row (history section) through
    // the normal changes toggle path.
    await waitFor(() => {
      const path = String(fetchMock.mock.calls.map((c) => String(c[0])))
      expect(path).toContain('/case-tasks/21/changes')
    })
  })

  it('ignores a focus id that does not match any row of this case', async () => {
    setupFetch([taskRow({ id: 21 })])
    renderPanel({ focusTaskID: 999 })
    await expectNoFocusedRow()
    // No cross-case history read may be triggered by the unknown id.
    await waitFor(() => {
      const requested = fetchMock.mock.calls.map((c) => String(c[0]))
      expect(requested).toEqual(['/api/v1/cases/12/tasks'])
    })
  })

  it('ignores a focus row whose case_id belongs to another case', async () => {
    setupFetch([taskRow({ id: 21, case_id: 99 })])
    renderPanel({ focusTaskID: 21 })
    await expectNoFocusedRow()
    await waitFor(() => {
      const requested = fetchMock.mock.calls.map((c) => String(c[0]))
      expect(requested).toEqual(['/api/v1/cases/12/tasks'])
    })
  })

  it('keeps later scope switches focused only on the current focus id', async () => {
    setupFetch([taskRow({ id: 21 })])
    const { rerender } = renderPanel({ focusTaskID: 21 })
    await findFocusedRow(21)
    setupFetch([taskRow({ id: 21 })], {
      31: [taskRow({ id: 30, case_id: 31, title: '新案任务' })],
    })
    rerender(
      <CaseTaskPanel
        caseID={31}
        assigneeOptions={assigneeOptions}
        currentUserId={7}
        focusTaskID={undefined}
      />,
    )
    await screen.findByText('新案任务')
    await waitFor(() => {
      expect(document.querySelectorAll('tr.task-row-focused').length).toBe(0)
    })
  })
})

function renderCaseDetail(initialEntry: string) {
  const user = {
    id: '7',
    username: 'lawyer-a',
    email: 'a@example.com',
    realName: '测试律师',
    roles: ['lawyer'],
    permissions: [],
    isActive: true,
    createdAt: today,
  }
  useAppStore.setState({ user, isAuthenticated: true })
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path='/case/:id' element={<CaseDetailCenter />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('CaseDetailCenter task_id deep link', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.endsWith('/cases/12')) {
        return Promise.resolve(
          jsonResponse({
            id: 12,
            case_number: 'CASE-2026-012',
            title: '浏览器练习案（虚构）',
            client_id: 3,
            lawyer_id: 7,
            status: 'in_progress',
          }),
        )
      }
      if (path.includes('/dashboard/command-center')) {
        return Promise.resolve(jsonResponse({}))
      }
      if (path.includes('/teams/case/12')) {
        return Promise.resolve(jsonResponse({ lead_lawyer: { id: 7, name: '律师甲' } }))
      }
      if (path.endsWith('/cases/12/tasks')) {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              taskRow({ id: 21, title: '直达任务' }),
              taskRow({ id: 22, title: '其他任务' }),
            ],
          }),
        )
      }
      return Promise.resolve(jsonResponse({ changes: [] }))
    })
  })

  afterEach(() => {
    cleanup()
    useAppStore.setState({ user: null, isAuthenticated: false })
  })

  it('focuses the requested task when task_id is present and valid', async () => {
    renderCaseDetail('/case/12?task_id=21')
    await screen.findByText('直达任务')
    await findFocusedRow(21)
    const other = document.querySelector("[data-testid='task-row-22']") as HTMLElement | null
    expect(other).not.toHaveClass('task-row-focused')
  })

  it('does not focus anything without task_id or with an invalid task_id', async () => {
    renderCaseDetail('/case/12')
    await screen.findByText('直达任务')
    await waitFor(() => {
      expect(document.querySelectorAll('tr.task-row-focused').length).toBe(0)
    })
    cleanup()
    renderCaseDetail('/case/12?task_id=abc')
    await screen.findByText('直达任务')
    await waitFor(() => {
      expect(document.querySelectorAll('tr.task-row-focused').length).toBe(0)
    })
  })
})
