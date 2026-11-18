import React from 'react'
import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import CaseTaskPanel from '../CaseTaskPanel'

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

function setupFetch(rows: Record<string, unknown>[]) {
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.endsWith('/cases/12/tasks')) {
      return Promise.resolve(jsonResponse({ tasks: rows }))
    }
    return Promise.resolve(jsonResponse({ changes: [] }))
  })
}

const assigneeOptions = [{ id: 7, name: '律师甲' }]

function renderPanel(props: Partial<React.ComponentProps<typeof CaseTaskPanel>> = {}) {
  return render(
    <CaseTaskPanel caseID={12} assigneeOptions={assigneeOptions} currentUserId={7} {...props} />,
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

// Main-review probe extensions: stale-highlight and denied-history guards.
// The deep-link focus effect must clear its highlight when the requested id
// changes or disappears, and must never auto-read history for a task whose
// capabilities deny can_view_changes.
describe('CaseTaskPanel focus stale guards', () => {
  afterEach(() => cleanup())

  it.each([undefined, 999])(
    'clears an old highlight after same-case focus changes to %s',
    async (nextID) => {
      setupFetch([taskRow({ id: 21 }), taskRow({ id: 22, title: '其他任务' })])
      const { rerender } = renderPanel({ focusTaskID: 21 })
      await findFocusedRow(21)
      rerender(
        <CaseTaskPanel caseID={12} assigneeOptions={assigneeOptions} currentUserId={7} focusTaskID={nextID} />,
      )
      await waitFor(() => {
        expect(document.querySelectorAll('tr.task-row-focused').length).toBe(0)
      })
    },
  )

  it('does not open denied history or paint a late previous-task history under a new target', async () => {
    let resolveHistory!: (response: Response) => void
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const p = String(input)
      if (p.endsWith('/cases/12/tasks')) {
        return Promise.resolve(
          jsonResponse({
            tasks: [
              taskRow({ id: 21 }),
              taskRow({
                id: 22,
                title: '仅可查看任务',
                capabilities: {
                  can_adjust: false,
                  can_cancel: false,
                  can_complete: false,
                  can_view_changes: false,
                },
              }),
            ],
          }),
        )
      }
      if (p.endsWith('/case-tasks/21/changes')) {
        return new Promise<Response>((resolve) => {
          resolveHistory = resolve
        })
      }
      return Promise.resolve(jsonResponse({ changes: [] }))
    })
    const { rerender } = renderPanel({ focusTaskID: 21 })
    await findFocusedRow(21)
    await waitFor(() => expect(resolveHistory).toBeDefined())
    rerender(
      <CaseTaskPanel caseID={12} assigneeOptions={assigneeOptions} currentUserId={7} focusTaskID={22} />,
    )
    await findFocusedRow(22)
    await act(async () => {
      resolveHistory(
        jsonResponse({
          changes: [
            {
              id: 100,
              task_id: 21,
              change_type: 'reassign',
              changed_by: 7,
              changed_at: '2026-10-07T00:00:00Z',
              reason: '旧任务专属调整说明不应出现在新任务下',
            },
          ],
        }),
      )
    })
    expect(screen.queryByText('旧任务专属调整说明不应出现在新任务下')).not.toBeInTheDocument()
    expect(
      fetchMock.mock.calls.some(([p]) => String(p).endsWith('/case-tasks/22/changes')),
    ).toBe(false)
  })
})
