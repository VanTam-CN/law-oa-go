import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { ConflictCheckResults } from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  getToken: () => 'test-token',
  getRoles: () => null,
  getUserInfo: () => null,
}))

const today = new Date().toISOString()

function riskRow(overrides: Record<string, unknown>) {
  return {
    id: 'CCT-x',
    title: '冲突检测',
    client_name: '虚构客户',
    status: 'COMPLETED',
    risk_level: 'REVIEW_REQUIRED',
    created_at: today,
    ...overrides,
  }
}

function dashboardPayload(riskQueue: Array<Record<string, unknown>>) {
  return {
    generated_at: today,
    summary: {},
    todo_items: [],
    risk_queue: riskQueue,
    approval_queue: [],
    case_rows: [],
    case_stage_distribution: [],
    overdue_tasks: [],
    recent_activities: [],
  }
}

async function renderQueue(riskQueue: Array<Record<string, unknown>>) {
  global.fetch = jest.fn((input: RequestInfo | URL) => {
    if (String(input).includes('/dashboard/command-center')) {
      return Promise.resolve({
        ok: true,
        json: async () => ({ data: dashboardPayload(riskQueue) }),
      })
    }
    return Promise.resolve({ ok: true, json: async () => ({ data: {} }) })
  }) as unknown as typeof fetch
  render(
    <MemoryRouter initialEntries={['/conflict']}>
      <ConflictCheckResults />
    </MemoryRouter>,
  )
  // The overview card splits each line into <span>label</span><strong>N 条</strong>,
  // so a whole-sentence text match cannot work; callers assert on the row title
  // (data is loaded) and use expectPendingCount for the split count line.
  await waitFor(() => expect(screen.getByRole('button', { name: '筛选待人工复核' })).toBeEnabled())
}

function expectPendingCount(count: number) {
  const summaryLine = screen.getByText((_, element) => {
    if (element?.tagName !== 'P') return false
    const span = element.querySelector('span')
    const strong = element.querySelector('strong')
    return span?.textContent === '待人工复核' && strong?.textContent === `${count} 条`
  })
  expect(summaryLine).toBeInTheDocument()
}

describe('ConflictCheckResults review-state filtering', () => {
  afterEach(() => {
    cleanup()
  })

  it('excludes a reviewed check from the pending-human-review filter and count', async () => {
    await renderQueue([
      riskRow({
        id: 'CCT-reviewed',
        title: '已复核检测',
        review_state: 'reviewed',
        review_decision: 'no_conflict',
      }),
    ])

    await waitFor(() => expect(screen.getByText('已复核检测')).toBeInTheDocument())
    expectPendingCount(0)

    fireEvent.click(screen.getByRole('button', { name: '筛选待人工复核' }))

    expect(screen.queryByText('已复核检测')).not.toBeInTheDocument()
  })

  it('keeps a pending_review check in the pending-human-review filter and count', async () => {
    await renderQueue([
      riskRow({
        id: 'CCT-pending',
        title: '待人工复核检测',
        review_state: 'pending_review',
      }),
    ])

    await waitFor(() => expect(screen.getByText('待人工复核检测')).toBeInTheDocument())
    expectPendingCount(1)

    fireEvent.click(screen.getByRole('button', { name: '筛选待人工复核' }))

    expect(screen.getByText('待人工复核检测')).toBeInTheDocument()
  })

  it('falls back to the frozen-suggestion behavior when review_state is absent', async () => {
    await renderQueue([
      riskRow({
        id: 'CCT-legacy',
        title: '旧后端检测',
        check_result: JSON.stringify({ decision: { status: 'REVIEW_REQUIRED' } }),
      }),
    ])

    await waitFor(() => expect(screen.getByText('旧后端检测')).toBeInTheDocument())
    expectPendingCount(1)

    fireEvent.click(screen.getByRole('button', { name: '筛选待人工复核' }))

    expect(screen.getByText('旧后端检测')).toBeInTheDocument()
  })
})
