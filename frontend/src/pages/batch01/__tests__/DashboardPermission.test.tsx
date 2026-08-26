import { cleanup, render, screen, waitFor, within } from '@testing-library/react'
import { DashboardCommandCenter } from '../Batch01Prototype'
import { useAppStore } from '@/stores/useAppStore'

jest.mock('react-router', () => ({
  ...jest.requireActual('react-router'),
  useNavigate: () => jest.fn(),
}))

jest.mock('@/utils/storage', () => ({
  getToken: () => null,
  getRoles: () => null,
  getUserInfo: () => null,
}))

const dashboardPayload = {
  generated_at: '2026-08-26T10:00:00+08:00',
  summary: {},
  todo_items: [],
  risk_queue: [],
  approval_queue: [],
  case_rows: [],
  case_stage_distribution: [],
  overdue_tasks: [],
  recent_activities: [],
}

function loginUser(roles: string[], permissions: string[] = []) {
  useAppStore.setState({
    user: {
      id: '1',
      username: 'test-user',
      email: 'test@example.com',
      realName: '测试用户',
      roles,
      permissions,
      isActive: true,
      createdAt: '2026-08-26T00:00:00+08:00',
    },
    isAuthenticated: true,
  })
}

async function renderDashboard() {
  global.fetch = jest.fn().mockResolvedValue({
    ok: true,
    json: async () => ({ data: dashboardPayload }),
  })

  const { container } = render(<DashboardCommandCenter />)

  await waitFor(() => {
    expect(screen.getByText('实时数据')).toBeInTheDocument()
  })

  return within(container.querySelector('.ng-shortcuts') as HTMLElement)
}

describe('DashboardCommandCenter action permissions', () => {
  afterEach(() => {
    cleanup()
    localStorage.clear()
    useAppStore.setState({ user: null, isAuthenticated: false })
  })

  it('shows intake and permitted workflow actions for a lawyer', async () => {
    loginUser(['lawyer'])

    const shortcuts = await renderDashboard()

    expect(shortcuts.getByRole('button', { name: /新建立案 录入新案件信息/ })).toBeInTheDocument()
    expect(shortcuts.getByRole('button', { name: /冲突检测 复核利益冲突/ })).toBeInTheDocument()
    expect(shortcuts.getByRole('button', { name: /审批队列 待审批事项/ })).toBeInTheDocument()
    expect(shortcuts.getAllByRole('button')).toHaveLength(4)
    expect(shortcuts.queryByText('数据看板')).not.toBeInTheDocument()
    expect(shortcuts.queryByText('经营分析')).not.toBeInTheDocument()
    expect(screen.queryByText('申请立案指引')).not.toBeInTheDocument()
  })

  it('hides inaccessible intake and workflow links from a plain user and shows guidance', async () => {
    loginUser(['user'])

    const shortcuts = await renderDashboard()

    expect(shortcuts.queryByRole('button', { name: /新建立案/ })).not.toBeInTheDocument()
    expect(shortcuts.queryByRole('button', { name: /冲突检测/ })).not.toBeInTheDocument()
    expect(shortcuts.queryByRole('button', { name: /审批队列/ })).not.toBeInTheDocument()
    expect(shortcuts.queryByRole('button', { name: /数据看板/ })).not.toBeInTheDocument()
    expect(shortcuts.queryByText('数据看板')).not.toBeInTheDocument()
    expect(shortcuts.queryByText('经营分析')).not.toBeInTheDocument()
    const guidance = shortcuts.getByRole('note')
    expect(guidance).toBeInTheDocument()
    expect(within(guidance).getByText('申请立案指引')).toBeInTheDocument()
    expect(guidance.closest('button')).toBeNull()
    expect(shortcuts.getByText('请联系管理员开通立案权限')).toBeInTheDocument()
    expect(shortcuts.getByRole('button', { name: /待办中心 处理 inbox 待办/ })).toBeInTheDocument()
  })
})
