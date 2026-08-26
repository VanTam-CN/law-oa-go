import { cleanup, render, screen, waitFor } from '@testing-library/react'
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

  render(<DashboardCommandCenter />)

  await waitFor(() => {
    expect(screen.getByText('实时数据')).toBeInTheDocument()
  })
}

describe('DashboardCommandCenter action permissions', () => {
  afterEach(() => {
    cleanup()
    localStorage.clear()
    useAppStore.setState({ user: null, isAuthenticated: false })
  })

  it('shows intake and permitted workflow actions for a lawyer', async () => {
    loginUser(['lawyer'])

    await renderDashboard()

    expect(screen.getByRole('button', { name: /新建立案/ })).toBeInTheDocument()
    expect(screen.getByText('冲突检测')).toBeInTheDocument()
    expect(screen.getAllByText('审批队列').some((item) => item.closest('.ng-shortcut'))).toBe(
      true,
    )
    expect(screen.queryByText('申请立案指引')).not.toBeInTheDocument()
  })

  it('hides inaccessible intake and workflow links from a plain user and shows guidance', async () => {
    loginUser(['user'])

    await renderDashboard()

    expect(screen.queryByRole('button', { name: /新建立案/ })).not.toBeInTheDocument()
    expect(screen.queryByText('新建立案')).not.toBeInTheDocument()
    expect(screen.queryByText('冲突检测')).not.toBeInTheDocument()
    expect(screen.getAllByText('审批队列').some((item) => item.closest('.ng-shortcut'))).toBe(
      false,
    )
    expect(screen.getByText('申请立案指引')).toBeInTheDocument()
    expect(screen.getByText('请联系管理员开通立案权限')).toBeInTheDocument()
    expect(screen.getByText('待办中心')).toBeInTheDocument()
  })
})
