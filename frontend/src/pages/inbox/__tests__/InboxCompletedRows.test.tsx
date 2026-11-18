import React from 'react'
import { cleanup, render, screen, within } from '@testing-library/react'
import InboxList from '../InboxList'

jest.mock('@/utils/storage', () => ({
  getToken: () => 'test-token',
  getRoles: () => null,
  getUserInfo: () => ({ id: 7 }),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

function inboxItem(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    user_id: 7,
    source_type: 'case',
    source_id: 31,
    title: '案件任务待办',
    content: '核对补充材料清单',
    priority: 'medium',
    due_date: null,
    due_date_type: 'none',
    is_read: false,
    read_at: null,
    is_completed: false,
    completed_at: null,
    reminder_sent: false,
    reminder_count: 0,
    escalated: false,
    escalated_at: null,
    snoozed_until: null,
    snoozed_count: 0,
    created_at: '2026-09-30T08:00:00Z',
    updated_at: '2026-09-30T08:00:00Z',
    workflow_action: {
      resource_type: 'case_task',
      resource_id: '77',
      related_case_id: '31',
      action_code: 'open_case_task',
      action_label: '去办理案件任务',
      workflow_status: 'open',
      completion_mode: 'workflow',
      action_version: '1',
    },
    ...overrides,
  }
}

function jsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

describe('InboxList completed-row presentation', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const path = String(input)
      if (path.startsWith('/api/v1/inbox/stats')) {
        return Promise.resolve(jsonResponse({
          total: 2, unread: 1, pending: 1, completed: 1, critical: 0,
          high: 0, overdue: 0, due_today: 0, due_this_week: 0,
        }))
      }
      if (path.startsWith('/api/v1/inbox?')) {
        return Promise.resolve(jsonResponse({
          items: [
            inboxItem({ id: 1, is_completed: false, completed_at: null }),
            inboxItem({
              id: 2,
              title: '已完成的案件任务',
              is_completed: true,
              completed_at: '2026-09-30T09:00:00Z',
              workflow_action: {
                resource_type: 'case_task',
                resource_id: '78',
                related_case_id: '31',
                action_code: 'open_case_task',
                action_label: '去办理案件任务',
                workflow_status: 'completed',
                completion_mode: 'workflow',
                action_version: '1',
              },
            }),
          ],
          pagination: { page: 1, page_size: 20, total: 2, total_page: 1 },
        }))
      }
      return Promise.resolve(jsonResponse({}))
    })
  })

  afterEach(() => cleanup())

  it('marks completed rows done and never offers a business button on them', async () => {
    render(<InboxList />)
    await screen.findByText('已完成的案件任务')

    const completedRow = screen.getByText('已完成的案件任务').closest('tr') as HTMLElement
    // 阻塞/办理 both say 已完成; no 去办理 button anywhere in the row.
    expect(within(completedRow).getAllByText('已完成').length).toBeGreaterThanOrEqual(2)
    expect(within(completedRow).queryByRole('button', { name: /去办理/ })).not.toBeInTheDocument()

    // The pending row keeps its actionable button.
    const pendingRow = screen.getByText('案件任务待办').closest('tr') as HTMLElement
    expect(within(pendingRow).getByRole('button', { name: '去办理案件任务' })).toBeInTheDocument()
    expect(within(pendingRow).getByText('可办理')).toBeInTheDocument()
  })
})

