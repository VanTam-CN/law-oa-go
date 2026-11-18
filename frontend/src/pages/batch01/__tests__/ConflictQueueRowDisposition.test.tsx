import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { ConflictCheckResults, conflictQueueRowDisposition } from '../Batch01Prototype'

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

describe('Conflict queue row disposition reflects the current review', () => {
  afterEach(cleanup)

  it.each(['no_conflict', 'confirmed_conflict'])(
    'labels a reviewed %s row with its real conclusion instead of the frozen pending suggestion',
    async (decision) => {
      await renderQueue([
        riskRow({
          id: 'CCT-done',
          title: '已处理检测',
          review_state: 'reviewed',
          review_decision: decision,
        }),
      ])
      const title = await screen.findByText('已处理检测')
      const cells = title.closest('tr')!.querySelectorAll('td')
      expectPendingCount(0)
      expect(cells[1]).not.toHaveTextContent('待人工复核')
      if (decision === 'confirmed_conflict') {
        expect(cells[1]).toHaveTextContent(/冲突|暂停/)
        expect(cells[1]).not.toHaveTextContent(/无冲突|放行/)
      } else {
        expect(cells[1]).toHaveTextContent(/已复核|无冲突|已通过/)
      }
    },
  )

  it('does not present an unreadable review state as a verified pending or reviewed conclusion', async () => {
    await renderQueue([
      riskRow({
        id: 'CCT-unknown',
        title: '状态未知检测',
        review_state: 'unknown',
      }),
    ])
    const title = await screen.findByText('状态未知检测')
    const cells = title.closest('tr')!.querySelectorAll('td')
    expectPendingCount(0)
    expect(cells[1]).not.toHaveTextContent('待人工复核')
    expect(cells[1]).toHaveTextContent(/未知|刷新|无法确认/)
  })

  it('exports rows with the same disposition wording as the visible queue', () => {
    expect(conflictQueueRowDisposition({ review_state: 'reviewed', review_decision: 'no_conflict' })).not.toContain('待人工复核')
    expect(conflictQueueRowDisposition({ review_state: 'reviewed', review_decision: 'confirmed_conflict' })).toMatch(/确认冲突/)
    expect(conflictQueueRowDisposition({ review_state: 'reviewed', review_decision: 'confirmed_conflict' })).not.toMatch(/无冲突|放行/)
    expect(conflictQueueRowDisposition({})).toBe('待人工复核')
  })
})
