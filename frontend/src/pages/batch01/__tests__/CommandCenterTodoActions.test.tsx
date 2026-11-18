import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import {
  clearConflictIntakePrefill,
  buildCluePrefill,
  DashboardCommandCenter,
  readConflictIntakePrefill,
} from '../Batch01Prototype'
import { useAppStore } from '@/stores/useAppStore'

jest.mock('@/utils/storage', () => ({
  getToken: () => null,
  getRoles: () => null,
  getUserInfo: () => null,
}))

const navigateMock = jest.fn()
jest.mock('react-router', () => ({
  ...jest.requireActual('react-router'),
  useNavigate: () => navigateMock,
}))

const today = new Date().toISOString()
const futureDate = new Date(Date.now() + 7 * 24 * 3600 * 1000).toISOString()

function dashboardPayloadSnapshotless() {
  return {
    generated_at: today,
    summary: {},
    todo_items: [
      // Undated approval todo: must stay visible in the default "today"
      // view (review fix: undated does NOT mean deferred).
      {
        id: 1,
        type: 'approval',
        title: '无期限审批待办',
        content: '请审批补充材料申请',
        priority: 'high',
        source_type: 'approval',
        workflow_action: {
          resource_type: 'approval',
          resource_id: 'APM-1',
          action_code: 'decide_approval',
          action_label: '去审批',
          workflow_status: 'submitted',
          completion_mode: 'workflow',
          action_version: '1',
        },
      },
      // Future-dated todo: the only kind the default view may hide.
      {
        id: 2,
        type: 'task',
        title: '下周才到期的任务',
        content: '未来任务',
        priority: 'low',
        due_at: futureDate,
        source_type: 'task',
        workflow_action: null,
      },
      // case_task todo: routing must use related_case_id, never resource_id.
      {
        id: 3,
        type: 'task',
        title: '案件任务待办',
        content: '核对补充材料清单',
        priority: 'medium',
        source_type: 'task',
        workflow_action: {
          resource_type: 'case_task',
          resource_id: '77',
          related_case_id: '31',
          related_title: '核对补充材料清单',
          next_step: '打开所属案件任务面板核对进度',
          action_code: 'open_case_task',
          action_label: '去办理案件任务',
          workflow_status: 'open',
          completion_mode: 'workflow',
          action_version: '1',
        },
      },
    ],
    risk_queue: [],
    approval_queue: [],
    case_rows: [],
    case_stage_distribution: [],
    overdue_tasks: [],
    recent_activities: [],
  }
}

function dashboardPayload() {
  // Default payload for every existing test: same todos, empty risk queue.
  return dashboardPayloadSnapshotless()
}

function dashboardPayloadWithConflictSnapshot() {
  const base = dashboardPayloadSnapshotless()
  return {
    ...base,
    risk_queue: [
      // Queue row with a usable request snapshot: the carry button must
      // appear and transport only the five neutral prefill fields.
      {
        id: 'CCT_snap',
        title: '快照初查任务',
        client_name: '甲公司',
        status: 'COMPLETED',
        risk_level: 'MINIMAL',
        created_at: today,
        search_parameters: JSON.stringify({
          caseName: '甲公司股权回购争议',
          clientName: '甲公司',
          caseType: '合同',
          parties: [
            { name: '甲公司', role: 'client' },
            { name: '乙公司', role: 'opposing_party' },
          ],
        }),
      },
      // Row without a snapshot: NO carry button may render for it.
      {
        id: 'CCT_bare',
        title: '无快照任务',
        client_name: '丙公司',
        status: 'REVIEW_REQUIRED',
        risk_level: 'HIGH',
        created_at: today,
      },
    ],
  }
}

function dashboardPayloadWithMachineDoneDesensitizedRow() {
  const base = dashboardPayloadSnapshotless()
  return {
    ...base,
    risk_queue: [
      // Machine run finished AND the desensitized risk value is
      // REVIEW_REQUIRED. Neither fact may render as "已完成" (a human
      // pass) nor as "待人工复核" (a claimed pending human review) on the
      // homepage; the verdict stays behind the detail page.
      {
        id: 'CCT_done',
        title: '机器完成脱敏行',
        client_name: '甲公司',
        status: 'COMPLETED',
        risk_level: 'REVIEW_REQUIRED',
        created_at: today,
      },
      // No machine status at all: the row must not invent a state.
      {
        id: 'CCT_nostate',
        title: '无状态行',
        client_name: '乙公司',
        status: '',
        risk_level: 'HIGH',
        created_at: today,
      },
    ],
  }
}

function loginUser() {
  useAppStore.setState({
    user: {
      id: '7',
      username: 'lawyer-a',
      email: 'a@example.com',
      realName: '测试律师',
      roles: ['lawyer'],
      permissions: [],
      isActive: true,
      createdAt: today,
    },
    isAuthenticated: true,
  })
}

function loginFinanceOnlyUser() {
  useAppStore.setState({
    user: {
      id: '9',
      username: 'finance-a',
      email: 'f@example.com',
      realName: '财务查看',
      roles: ['finance_viewer'],
      permissions: [],
      isActive: true,
      createdAt: today,
    },
    isAuthenticated: true,
  })
}

describe('DashboardCommandCenter todo actions', () => {
  beforeEach(() => {
    navigateMock.mockClear()
    clearConflictIntakePrefill()
    loginUser()
    global.fetch = jest.fn((input: RequestInfo | URL) => {
      const url = String(input)
      if (url.includes('/dashboard/command-center')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({ data: dashboardPayload() }),
        })
      }
      if (url.includes('/case-intakes')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({
            data: {
              items: [
                {
                  id: 'draft-owned',
                  intake_code: 'INTAKE-1',
                  title: '我的草稿',
                  case_type: 'civil',
                  created_by: '7',
                  updated_at: today,
                },
                {
                  id: 'draft-other',
                  intake_code: 'INTAKE-2',
                  title: '他人草稿',
                  case_type: 'civil',
                  created_by: '9',
                  updated_at: today,
                },
              ],
            },
          }),
        })
      }
      return Promise.resolve({ ok: true, json: async () => ({ data: {} }) })
    }) as unknown as typeof fetch
  })

  afterEach(() => {
    cleanup()
    clearConflictIntakePrefill()
    useAppStore.setState({ user: null, isAuthenticated: false })
  })

  it('keeps undated todos in the today view and hides only future-dated ones', async () => {
    await renderWithRouter()
    await waitFor(() => expect(screen.getByText('无期限审批待办')).toBeInTheDocument())
    await waitFor(() => expect(screen.getByText('我的草稿')).toBeInTheDocument())

    fireEvent.click(screen.getByRole('button', { name: /今日需处理/ }))

    expect(screen.getByText('无期限审批待办')).toBeInTheDocument()
    expect(screen.getByText('案件任务待办')).toBeInTheDocument()
    expect(screen.queryByText('下周才到期的任务')).not.toBeInTheDocument()
  })

  it('routes a case_task todo by related_case_id with the task id as a query param', async () => {
    await renderWithRouter()
    await waitFor(() => expect(screen.getByText('案件任务待办')).toBeInTheDocument())

    fireEvent.click(screen.getByRole('button', { name: '去办理：案件任务待办' }))

    expect(navigateMock).toHaveBeenCalledWith('/case/31?task_id=77')
    expect(navigateMock).not.toHaveBeenCalledWith(expect.stringContaining('/case/77'))
  })

  it('routes an approval todo to its decision page and shows only own drafts', async () => {
    await renderWithRouter()
    await waitFor(() => expect(screen.getByText('无期限审批待办')).toBeInTheDocument())
    await waitFor(() => expect(screen.getByText('我的草稿')).toBeInTheDocument())

    fireEvent.click(screen.getByRole('button', { name: '去办理：无期限审批待办' }))
    expect(navigateMock).toHaveBeenCalledWith('/approval/APM-1')

    expect(screen.queryByText('他人草稿')).not.toBeInTheDocument()
  })

  it('pre-intake clue panel carries only neutral facts into the intake prefill', async () => {
    await renderWithRouter()
    await waitFor(() => expect(screen.getByText('实时数据')).toBeInTheDocument())

    fireEvent.click(screen.getByRole('button', { name: '接案线索登记' }))
    const panel = document.getElementById('pre-intake-clue-panel')
    expect(panel).not.toBeNull()

    fireEvent.change(screen.getByLabelText('初查客户名称'), {
      target: { value: '甲公司' },
    })
    fireEvent.change(screen.getByLabelText('初查对方当事人名称'), {
      target: { value: '乙公司' },
    })
    fireEvent.mouseDown(screen.getAllByLabelText('初查案件类型')[0]!.querySelector('input')!)
    const option = await waitFor(() => {
      const el = Array.from(document.querySelectorAll('.ant-select-item-option')).find(
        (node) => node.textContent === '商事诉讼',
      )
      expect(el).toBeDefined()
      return el as HTMLElement
    })
    fireEvent.click(option)
    fireEvent.change(screen.getByLabelText('初查事项摘要'), {
      target: { value: '拟就投资协议争议咨询并可能委托' },
    })

    fireEvent.click(screen.getByRole('button', { name: '带入立案' }))

    expect(navigateMock).toHaveBeenCalledWith('/case/create')
    const prefill = readConflictIntakePrefill()
    expect(prefill).not.toBeNull()
    expect(prefill?.clientName).toBe('甲公司')
    expect(prefill?.title).toBe('拟就投资协议争议咨询并可能委托')
    expect(prefill?.partyNote).toBe('对方当事人：乙公司')
    expect(Object.keys(prefill || {}).sort()).toEqual([
      'caseType',
      'clientName',
      'description',
      'partyNote',
      'title',
    ])
    // No check id / risk conclusion / status may ride along.
    expect(JSON.stringify(prefill)).not.toContain('CCT_')
    expect(JSON.stringify(prefill)).not.toContain('risk')
  })

  it('clue carry stays disabled until client name is present (case type optional)', async () => {
    await renderWithRouter()
    await waitFor(() => expect(screen.getByText('实时数据')).toBeInTheDocument())

    fireEvent.click(screen.getByRole('button', { name: '接案线索登记' }))
    const carry = screen.getByRole('button', { name: '带入立案' }) as HTMLButtonElement
    expect(carry.disabled).toBe(true)

    fireEvent.change(screen.getByLabelText('初查客户名称'), { target: { value: '甲公司' } })
    expect(carry.disabled).toBe(false)
  })

  describe('homepage conflict card semantics (round 2 usability)', () => {
    it('labels the card by attention, not by a pending-human-review claim', async () => {
      await renderWithRouter()
      await waitFor(() => expect(screen.getByText('实时数据')).toBeInTheDocument())

      // KPI card, panel title, and data-overview row all use the neutral
      // attention label; none may still claim a pending human review.
      expect(screen.getAllByText('需关注的冲突检查').length).toBeGreaterThan(0)
      expect(screen.queryByText('利益冲突待复核')).not.toBeInTheDocument()
    })

    it('shows machine-completed wording and keeps the review verdict behind details', async () => {
      global.fetch = jest.fn((input: RequestInfo | URL) => {
        const url = String(input)
        if (url.includes('/dashboard/command-center')) {
          return Promise.resolve({
            ok: true,
            json: async () => ({ data: dashboardPayloadWithMachineDoneDesensitizedRow() }),
          })
        }
        if (url.includes('/case-intakes')) {
          return Promise.resolve({
            ok: true,
            json: async () => ({ data: { items: [] } }),
          })
        }
        return Promise.resolve({ ok: true, json: async () => ({ data: {} }) })
      }) as unknown as typeof fetch

      await renderWithRouter()
      await waitFor(() => expect(screen.getByText('机器完成脱敏行')).toBeInTheDocument())

      // Machine state renders as machine wording only.
      expect(screen.getByText('自动检查：已完成')).toBeInTheDocument()
      // Every machine-completed row points to the detail page for the
      // human review verdict; the homepage itself never claims one.
      expect(screen.getAllByText('人工复核结果请查看详情').length).toBeGreaterThan(0)
      // Neither tag may claim a human verdict or a human queue state.
      expect(screen.queryByText('已完成')).not.toBeInTheDocument()
      // The desensitized risk value shows a neutral attention label.
      expect(screen.getByText('风险结论需查看')).toBeInTheDocument()
      expect(screen.queryByText('待人工复核')).not.toBeInTheDocument()

      fireEvent.click(screen.getByRole('button', { name: '查看冲突详情：机器完成脱敏行' }))
      expect(navigateMock).toHaveBeenCalledWith('/conflict?task_id=CCT_done')
    })

    it('does not invent a state for a row without machine status', async () => {
      global.fetch = jest.fn((input: RequestInfo | URL) => {
        const url = String(input)
        if (url.includes('/dashboard/command-center')) {
          return Promise.resolve({
            ok: true,
            json: async () => ({ data: dashboardPayloadWithMachineDoneDesensitizedRow() }),
          })
        }
        if (url.includes('/case-intakes')) {
          return Promise.resolve({
            ok: true,
            json: async () => ({ data: { items: [] } }),
          })
        }
        return Promise.resolve({ ok: true, json: async () => ({ data: {} }) })
      }) as unknown as typeof fetch

      await renderWithRouter()
      await waitFor(() => expect(screen.getByText('无状态行')).toBeInTheDocument())

      expect(screen.getAllByText('人工复核结果请查看详情').length).toBeGreaterThan(0)
      expect(screen.queryByText('已完成')).not.toBeInTheDocument()
    })

    it('uses the attention wording for an empty conflict queue', async () => {
      global.fetch = jest.fn((input: RequestInfo | URL) => {
        const url = String(input)
        if (url.includes('/dashboard/command-center')) {
          return Promise.resolve({
            ok: true,
            json: async () => ({ data: dashboardPayloadWithMachineDoneDesensitizedRow() }),
          })
        }
        if (url.includes('/case-intakes')) {
          return Promise.resolve({
            ok: true,
            json: async () => ({ data: { items: [] } }),
          })
        }
        return Promise.resolve({ ok: true, json: async () => ({ data: {} }) })
      }) as unknown as typeof fetch

      await renderWithRouter()
      await waitFor(() =>
        expect(screen.getByText('暂无需关注的冲突检查任务')).toBeInTheDocument(),
      )
      expect(screen.queryByText('暂无数据库冲突复核任务')).not.toBeInTheDocument()
    })
  })

  it('hides the clue entry entirely for users without case-create permission', async () => {
    loginFinanceOnlyUser()
    await renderWithRouter()
    await waitFor(() => expect(screen.getByText('实时数据')).toBeInTheDocument())

    expect(screen.queryByRole('button', { name: '接案线索登记' })).toBeNull()
    expect(document.getElementById('pre-intake-clue-panel')).toBeNull()
  })

  describe('buildCluePrefill pure function', () => {
    it('returns exactly the five-field neutral whitelist with role restated', () => {
      const prefill = buildCluePrefill({
        client: ' 甲公司 ',
        opponent: '乙公司',
        caseType: '商事诉讼',
        summary: '拟就投资协议争议咨询并可能委托',
      })
      expect(prefill).not.toBeNull()
      expect(prefill).toEqual({
        title: '拟就投资协议争议咨询并可能委托',
        clientName: '甲公司',
        caseType: '商事诉讼',
        partyNote: '对方当事人：乙公司',
        description: '拟就投资协议争议咨询并可能委托',
      })
      // No conclusion-ish or identity-ish keys may exist.
      expect(JSON.stringify(prefill)).not.toContain('CCT_')
      expect(JSON.stringify(prefill)).not.toContain('risk')
      expect(JSON.stringify(prefill)).not.toContain('checkId')
    })

    it('returns null for a blank client name', () => {
      expect(buildCluePrefill({ client: '   ' })).toBeNull()
      expect(buildCluePrefill({ client: '' })).toBeNull()
    })

    it('keeps case type optional and party note empty without an opponent', () => {
      const prefill = buildCluePrefill({ client: '甲公司', summary: '咨询' })
      expect(prefill).not.toBeNull()
      expect(prefill?.caseType).toBe('')
      expect(prefill?.partyNote).toBe('')
      expect(prefill?.description).toBe('咨询')
    })
  })
}) 

async function renderWithRouter() {
  render(
    <MemoryRouter initialEntries={["/"]}>
      <DashboardCommandCenter />
    </MemoryRouter>,
  )
  await waitFor(() => expect(screen.getByText('实时数据')).toBeInTheDocument())
}
