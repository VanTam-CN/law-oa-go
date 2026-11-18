import React from 'react'
import { cleanup, render, screen, within } from '@testing-library/react'
import dayjs from 'dayjs'
import InboxList from '../InboxList'

jest.mock('@/utils/storage', () => ({
  getToken: () => 'test-token',
  getRoles: () => null,
  getUserInfo: () => ({ id: 7 }),
}))

const fetchMock = jest.fn()
global.fetch = fetchMock as unknown as typeof fetch

// Fixed clock: local 2026-10-04 01:10. The real browser FAIL happened at
// this wall time: a due date of 2026-10-05 00:00 is ~22h50m away and was
// wrongly labeled 今天到期 by whole-day truncation.
const NOW = dayjs('2026-10-04 01:10:00')

function inboxItem(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    user_id: 7,
    source_type: 'case',
    source_id: 31,
    title: '日期标签用例',
    content: '核对补充材料清单',
    priority: 'medium',
    due_date: null,
    due_date_type: 'case_task',
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
    workflow_action: null,
    ...overrides,
  }
}

function jsonResponse(data: unknown) {
  return { ok: true, json: async () => ({ success: true, data }) } as Response
}

function mockList(items: ReturnType<typeof inboxItem>[]) {
  fetchMock.mockReset()
  fetchMock.mockImplementation((input: RequestInfo | URL) => {
    const path = String(input)
    if (path.startsWith('/api/v1/inbox/stats')) {
      return Promise.resolve(jsonResponse({
        total: items.length, unread: 0, pending: items.length, completed: 0,
        critical: 0, high: 0, overdue: 0, due_today: 0, due_this_week: 0,
      }))
    }
    if (path.startsWith('/api/v1/inbox?')) {
      return Promise.resolve(jsonResponse({
        items,
        pagination: { page: 1, page_size: 20, total: items.length, total_page: 1 },
      }))
    }
    return Promise.resolve(jsonResponse({}))
  })
}

async function renderRow(item: ReturnType<typeof inboxItem>) {
  mockList([item])
  render(<InboxList />)
  await screen.findByText(item.title as string)
  return screen.getByText(item.title as string).closest('tr') as HTMLElement
}

async function rowText(item: ReturnType<typeof inboxItem>) {
  const row = await renderRow(item)
  return row.textContent ?? ''
}

describe('InboxList due-date natural-day labels with a fixed local clock', () => {
  beforeEach(() => {
    // Freeze Date construction too: dayjs() uses new Date(), not Date.now().
    // Keep scheduling real so fetch, RTL waits and UI effects can finish.
    jest.useFakeTimers({
      doNotFake: [
        'setTimeout', 'clearTimeout', 'setInterval', 'clearInterval',
        'setImmediate', 'clearImmediate', 'nextTick', 'queueMicrotask',
        'performance', 'hrtime', 'requestAnimationFrame', 'cancelAnimationFrame',
        'requestIdleCallback', 'cancelIdleCallback',
      ],
    })
    jest.setSystemTime(NOW.toDate())
    expect(dayjs().valueOf()).toBe(NOW.valueOf())
  })

  afterEach(() => {
    cleanup()
    jest.useRealTimers()
    jest.restoreAllMocks()
  })

  it('labels due 2026-10-05 00:00 as tomorrow, not today (regression: 22h50m truncation)', async () => {
    const row = await renderRow(inboxItem({ id: 1, title: '明天零点任务', due_date: '2026-10-05 00:00:00' }))
    expect(within(row).getByText('明天到期')).toBeInTheDocument()
    expect(within(row).queryByText('今天到期')).not.toBeInTheDocument()
  })

  it('labels a later time today as due today', async () => {
    const row = await renderRow(inboxItem({ id: 2, title: '今天晚些任务', due_date: '2026-10-04 18:00:00' }))
    expect(within(row).getByText('今天到期')).toBeInTheDocument()
  })

  it('labels yesterday as overdue by 1 day', async () => {
    const row = await renderRow(inboxItem({ id: 3, title: '昨天到期任务', due_date: '2026-10-03 12:00:00' }))
    expect(within(row).getByText('已超期 1 天')).toBeInTheDocument()
  })

  it('labels a same-day already-passed deadline as overdue, never as not-yet-due', async () => {
    const row = await renderRow(inboxItem({ id: 4, title: '今天已过点任务', due_date: '2026-10-04 00:30:00' }))
    expect(within(row).getByText('已超期')).toBeInTheDocument()
    expect(within(row).queryByText('0 天')).not.toBeInTheDocument()
    expect(within(row).queryByText('今天到期')).not.toBeInTheDocument()
  })

  it('labels due 2026-10-06 as 2 days remaining', async () => {
    const row = await renderRow(inboxItem({ id: 5, title: '后天到期任务', due_date: '2026-10-06 12:00:00' }))
    expect(within(row).getByText('剩余 2 天')).toBeInTheDocument()
  })

  it('shows no error label without deadline', async () => {
    const row = await renderRow(inboxItem({ id: 6, title: '无期限任务', due_date: null }))
    expect(within(row).queryByText(/到期|超期|剩余/)).not.toBeInTheDocument()
  })

  it('shows no due-date label for invalid date strings', async () => {
    const row = await renderRow(inboxItem({ id: 7, title: '非法日期任务', due_date: 'not-a-date' }))
    expect(within(row).queryByText(/到期|超期|剩余/)).not.toBeInTheDocument()
  })

  it('shows completed rows without any due-date label', async () => {
    const row = await renderRow(inboxItem({
      id: 8,
      title: '已完成任务',
      is_completed: true,
      completed_at: '2026-10-01T09:00:00Z',
      due_date: '2026-10-03 12:00:00',
    }))
    expect(within(row).queryByText(/到期|超期|剩余/)).not.toBeInTheDocument()
    expect(within(row).getAllByText('已完成').length).toBeGreaterThanOrEqual(1)
  })

  it('handles the near-midnight cross-day boundary consistently on calendar days', async () => {
    const midnightNow = NOW.hour(23).minute(30)
    jest.setSystemTime(midnightNow.toDate())
    expect(dayjs().valueOf()).toBe(midnightNow.valueOf())
    // At 23:30, a deadline 20 minutes away is today; 40 minutes away is tomorrow.
    const todayRow = await renderRow(inboxItem({ id: 9, title: '接近午夜今天', due_date: '2026-10-04 23:50:00' }))
    expect(within(todayRow).getByText('今天到期')).toBeInTheDocument()
    expect(within(todayRow).queryByText('明天到期')).not.toBeInTheDocument()
    cleanup()
    const tomorrowRow = await renderRow(inboxItem({ id: 10, title: '接近午夜明天', due_date: '2026-10-05 00:10:00' }))
    expect(within(tomorrowRow).getByText('明天到期')).toBeInTheDocument()
    expect(within(tomorrowRow).queryByText('今天到期')).not.toBeInTheDocument()
    expect(dayjs().valueOf()).toBe(midnightNow.valueOf())
  })
})
