# MVP 试用收口版 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use TDD strictly. No production code may be changed before a failing test is written and observed. Execute this plan task-by-task with checkbox tracking.

**Goal:** 把当前系统收口成“律所主任内部 MVP 试用版”，保留工作台、案件、新建立案、利益冲突、客户、审批、信托账户这些主任高频路径，隐藏或降级未完成模块，并保证前端质量门禁通过。

**Architecture:** 前端采用“保留路由、收口入口、业务化空态”的方式，不删除后端接口和已有页面。MVP 菜单由统一配置控制；未纳入 MVP 的页面用统一组件展示“未纳入本次 MVP 试用范围”；利益冲突单独改成可试用的风控工作台，并和新建立案流程保持联动。

**Tech Stack:** Go backend, React 18, Vite, TypeScript, React Router 7, Ant Design 5, Jest, Testing Library, ESLint.

---

## Non-Negotiable Rules

- 不删除测试、不跳过测试、不降低 TypeScript 或 ESLint 规则。
- 不把 `strict`、`noImplicitAny`、`skipLibCheck` 等配置改松来换通过。
- 不用大面积 `any` 掩盖类型问题。
- 不删除后端接口、不删除未完成页面代码，只调整试用入口和展示策略。
- 利益冲突必须保留，是 MVP 核心风控模块。
- 新建立案里的“利益冲突检查”步骤必须保留。
- 所有行为变更必须先写失败测试，确认失败原因正确后再改实现。

## MVP Scope

### 保留为一级入口

- `/dashboard` 工作台
- `/case` 案件管理
- `/case/create` 新建立案
- `/conflict` 利益冲突
- `/client` 客户管理
- `/approval` 审批工作台
- `/trust` 信托账户

### 从一级入口隐藏，但保留直接访问路由

- `/file` 文档中心
- `/finance` 财务中心
- `/inbox` 收件箱
- `/settings` 系统设置

### 不纳入本次收口

- 律师管理
- 工具箱
- 用户管理
- 角色管理
- 权限管理
- 财务子页面
- 开发测试页面

---

## Expected Final Verification

GLM 完成全部任务后必须执行：

```bash
cd /Users/mac/Desktop/FT/law-oa-go
go build ./...

cd /Users/mac/Desktop/FT/law-oa-go/frontend
npm run type-check
npm run lint
npm run build
npm test -- --watchAll=false
```

Chrome 回归必须覆盖：

- 登录：`admin / Demo@2026`
- `/dashboard`
- `/case`
- `/case/create`
- `/conflict`
- `/client`
- `/approval`
- `/trust`
- 直接访问 `/file`、`/finance`、`/inbox`、`/settings`

---

## Task 1: Add Central MVP Route Configuration

**Files:**

- Create: `frontend/src/config/mvp.ts`
- Create: `frontend/src/config/__tests__/mvp.test.ts`

### TDD Steps

- [ ] **Step 1: Write the failing test**

Create `frontend/src/config/__tests__/mvp.test.ts`:

```ts
import {
  MVP_MENU_KEYS,
  MVP_UNAVAILABLE_PATHS,
  isMvpMenuKey,
  isMvpUnavailablePath,
} from '../mvp'

describe('mvp route configuration', () => {
  it('keeps director MVP menu keys including conflict', () => {
    expect(MVP_MENU_KEYS).toEqual([
      'dashboard',
      'case',
      'conflict',
      'client',
      'approval',
      'trust',
    ])
  })

  it('does not treat conflict as unavailable', () => {
    expect(isMvpMenuKey('conflict')).toBe(true)
    expect(isMvpUnavailablePath('/conflict')).toBe(false)
  })

  it('marks unfinished modules as unavailable pages', () => {
    expect(MVP_UNAVAILABLE_PATHS).toEqual(['/file', '/finance', '/inbox', '/settings'])
    expect(isMvpUnavailablePath('/finance')).toBe(true)
    expect(isMvpUnavailablePath('/finance/contracts/1')).toBe(true)
    expect(isMvpUnavailablePath('/dashboard')).toBe(false)
  })
})
```

- [ ] **Step 2: Run test and verify RED**

Run:

```bash
cd /Users/mac/Desktop/FT/law-oa-go/frontend
npm test -- src/config/__tests__/mvp.test.ts --watchAll=false
```

Expected: FAIL because `frontend/src/config/mvp.ts` does not exist.

- [ ] **Step 3: Implement minimal config**

Create `frontend/src/config/mvp.ts`:

```ts
export const MVP_MENU_KEYS = [
  'dashboard',
  'case',
  'conflict',
  'client',
  'approval',
  'trust',
] as const

export const MVP_UNAVAILABLE_PATHS = ['/file', '/finance', '/inbox', '/settings'] as const

type MvpMenuKey = (typeof MVP_MENU_KEYS)[number]

export const isMvpMenuKey = (key: string): key is MvpMenuKey => {
  return MVP_MENU_KEYS.includes(key as MvpMenuKey)
}

export const isMvpUnavailablePath = (pathname: string): boolean => {
  return MVP_UNAVAILABLE_PATHS.some((path) => pathname === path || pathname.startsWith(`${path}/`))
}
```

- [ ] **Step 4: Run test and verify GREEN**

Run:

```bash
npm test -- src/config/__tests__/mvp.test.ts --watchAll=false
```

Expected: PASS.

---

## Task 2: Restrict Sidebar to MVP Menu Items

**Files:**

- Modify: `frontend/src/components/layout/Sidebar.tsx`
- Create: `frontend/src/components/layout/__tests__/Sidebar.mvp.test.tsx`
- Use config: `frontend/src/config/mvp.ts`

### Required Behavior

After login, sidebar must show only:

- 工作台
- 案件管理
- 利益冲突检查
- 客户管理
- 审批中心
- 代管款管理

Sidebar must not show:

- 律师管理
- 文件管理
- 工具箱
- 财务管理
- 待办中心
- 用户管理
- 角色管理
- 权限管理
- 系统设置

### TDD Steps

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/layout/__tests__/Sidebar.mvp.test.tsx`.

Test intent:

```tsx
import React from 'react'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import Sidebar from '../Sidebar'

jest.mock('@/stores/useAppStore', () => ({
  useAppStore: () => ({
    user: {
      permissions: [
        'dashboard:view',
        'case:manage',
        'client:manage',
        'conflict:check',
        'approval:manage',
        'trust:manage',
        'lawyer:manage',
        'document:manage',
        'tools:view',
        'finance:view',
        'user:manage',
        'role:manage',
        'permission:manage',
        'system:manage',
      ],
    },
  }),
}))

jest.mock('@/utils/accessControl', () => ({
  hasPermission: () => true,
}))

describe('Sidebar MVP mode', () => {
  it('renders only director MVP menu entries and keeps conflict visible', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Sidebar collapsed={false} setCollapsed={jest.fn()} />
      </MemoryRouter>,
    )

    expect(screen.getByText('工作台')).toBeInTheDocument()
    expect(screen.getByText('案件管理')).toBeInTheDocument()
    expect(screen.getByText('利益冲突检查')).toBeInTheDocument()
    expect(screen.getByText('客户管理')).toBeInTheDocument()
    expect(screen.getByText('审批中心')).toBeInTheDocument()
    expect(screen.getByText('代管款管理')).toBeInTheDocument()

    expect(screen.queryByText('律师管理')).not.toBeInTheDocument()
    expect(screen.queryByText('文件管理')).not.toBeInTheDocument()
    expect(screen.queryByText('工具箱')).not.toBeInTheDocument()
    expect(screen.queryByText('财务管理')).not.toBeInTheDocument()
    expect(screen.queryByText('待办中心')).not.toBeInTheDocument()
    expect(screen.queryByText('系统设置')).not.toBeInTheDocument()
  })
})
```

- [ ] **Step 2: Run test and verify RED**

Run:

```bash
npm test -- src/components/layout/__tests__/Sidebar.mvp.test.tsx --watchAll=false
```

Expected: FAIL because current sidebar still renders non-MVP entries when permission allows.

- [ ] **Step 3: Implement minimal sidebar filtering**

In `frontend/src/components/layout/Sidebar.tsx`:

1. Import config:

```ts
import { isMvpMenuKey } from '@/config/mvp'
```

2. In the menu filtering logic, add MVP filter before permission filter:

```ts
.filter((item) => {
  if (!isMvpMenuKey(item.key)) {
    return false
  }

  if (item.children) {
    const filteredChildren = filterMenuItems(item.children)
    return filteredChildren.length > 0
  }

  if (item.permission) {
    return hasPermission(item.permission)
  }

  return true
})
```

3. Ensure menu order is:

```ts
dashboard -> case -> conflict -> client -> approval -> trust
```

If current `baseMenuItems` order differs, move the existing conflict item before client or move client after conflict.

- [ ] **Step 4: Run test and verify GREEN**

Run:

```bash
npm test -- src/components/layout/__tests__/Sidebar.mvp.test.tsx --watchAll=false
```

Expected: PASS.

---

## Task 3: Add Unified MVP Unavailable Page

**Files:**

- Create: `frontend/src/components/mvp/MvpUnavailable.tsx`
- Create: `frontend/src/components/mvp/__tests__/MvpUnavailable.test.tsx`
- Modify later: `frontend/src/App.tsx`

### Required Behavior

For `/file`、`/finance`、`/inbox`、`/settings`, direct access must show:

- `{moduleName}未纳入本次 MVP 试用范围`
- `当前试用版聚焦主任工作台、案件、利益冲突、客户、审批和信托账户流程。`
- Button: `返回工作台`

Do not show:

- `开发中`
- pure empty tables
- raw database empty state text

### TDD Steps

- [ ] **Step 1: Write the failing test**

Create `frontend/src/components/mvp/__tests__/MvpUnavailable.test.tsx`:

```tsx
import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import MvpUnavailable from '../MvpUnavailable'

const navigateMock = jest.fn()

jest.mock('react-router', () => {
  const actual = jest.requireActual('react-router')
  return {
    ...actual,
    useNavigate: () => navigateMock,
  }
})

describe('MvpUnavailable', () => {
  beforeEach(() => {
    navigateMock.mockClear()
  })

  it('explains the MVP scope and returns to dashboard', async () => {
    render(
      <MemoryRouter>
        <MvpUnavailable moduleName='财务中心' />
      </MemoryRouter>,
    )

    expect(screen.getByText('财务中心未纳入本次 MVP 试用范围')).toBeInTheDocument()
    expect(
      screen.getByText('当前试用版聚焦主任工作台、案件、利益冲突、客户、审批和信托账户流程。'),
    ).toBeInTheDocument()
    expect(screen.queryByText(/开发中/)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: '返回工作台' }))
    expect(navigateMock).toHaveBeenCalledWith('/dashboard')
  })
})
```

- [ ] **Step 2: Run test and verify RED**

Run:

```bash
npm test -- src/components/mvp/__tests__/MvpUnavailable.test.tsx --watchAll=false
```

Expected: FAIL because component does not exist.

- [ ] **Step 3: Implement component**

Create `frontend/src/components/mvp/MvpUnavailable.tsx`:

```tsx
import React from 'react'
import { Button, Result, Typography } from 'antd'
import { useNavigate } from 'react-router'

const { Paragraph } = Typography

interface MvpUnavailableProps {
  moduleName: string
}

const MvpUnavailable: React.FC<MvpUnavailableProps> = ({ moduleName }) => {
  const navigate = useNavigate()

  return (
    <Result
      status='info'
      title={`${moduleName}未纳入本次 MVP 试用范围`}
      subTitle={
        <Paragraph style={{ marginBottom: 0 }}>
          当前试用版聚焦主任工作台、案件、利益冲突、客户、审批和信托账户流程。
        </Paragraph>
      }
      extra={
        <Button type='primary' onClick={() => navigate('/dashboard')}>
          返回工作台
        </Button>
      }
    />
  )
}

export default MvpUnavailable
```

- [ ] **Step 4: Run test and verify GREEN**

Run:

```bash
npm test -- src/components/mvp/__tests__/MvpUnavailable.test.tsx --watchAll=false
```

Expected: PASS.

---

## Task 4: Route Unfinished Modules to MVP Unavailable Page

**Files:**

- Modify: `frontend/src/App.tsx`
- Test: `frontend/src/App.mvp.test.tsx` or `frontend/src/__tests__/App.mvp.test.tsx`

### Required Behavior

Direct access should render the unavailable page:

- `/file` -> `文档中心未纳入本次 MVP 试用范围`
- `/finance` -> `财务中心未纳入本次 MVP 试用范围`
- `/inbox` -> `收件箱未纳入本次 MVP 试用范围`
- `/settings` -> `系统设置未纳入本次 MVP 试用范围`

Do not remove the original page imports unless TypeScript/lint requires cleanup. If imports become unused, remove only the unused imports.

### TDD Steps

- [ ] **Step 1: Write failing route tests**

Create an App route test that renders authenticated app at each path and asserts the MVP unavailable title appears. If existing app setup is too heavy, test a small exported route mapping helper instead. Preferred route test behavior:

```tsx
it.each([
  ['/file', '文档中心未纳入本次 MVP 试用范围'],
  ['/finance', '财务中心未纳入本次 MVP 试用范围'],
  ['/inbox', '收件箱未纳入本次 MVP 试用范围'],
  ['/settings', '系统设置未纳入本次 MVP 试用范围'],
])('renders MVP unavailable page for %s', async (path, title) => {
  // Render App inside MemoryRouter with authenticated store.
  // Assert screen.findByText(title).
})
```

If mocking `useAppStore` is easier, mock it as authenticated:

```ts
jest.mock('@/stores/useAppStore', () => ({
  useAppStore: () => ({
    isLoading: false,
    isAuthenticated: true,
    user: {
      id: 1,
      username: 'admin',
      real_name: '张律师',
      permissions: ['dashboard:view', 'case:manage', 'client:view', 'conflict:check', 'approval:manage', 'trust:manage', 'finance:view', 'file:view', 'system:manage'],
    },
  }),
}))
```

- [ ] **Step 2: Run test and verify RED**

Run:

```bash
npm test -- App.mvp.test.tsx --watchAll=false
```

Expected: FAIL because routes still render original pages.

- [ ] **Step 3: Change routes**

In `frontend/src/App.tsx`, import:

```ts
import MvpUnavailable from './components/mvp/MvpUnavailable'
```

Change protected routes:

```tsx
<Route path='file' element={withAccess(<MvpUnavailable moduleName='文档中心' />, 'file:view')} />
<Route path='finance' element={withAccess(<MvpUnavailable moduleName='财务中心' />, 'finance:view')} />
<Route path='settings' element={withAccess(<MvpUnavailable moduleName='系统设置' />, 'system:manage')} />
<Route path='inbox' element={<MvpUnavailable moduleName='收件箱' />} />
```

Finance child routes should also be unavailable during MVP:

```tsx
<Route path='finance/contracts/:id' element={withAccess(<MvpUnavailable moduleName='财务中心' />, 'finance:view')} />
<Route path='finance/bad-debts' element={withAccess(<MvpUnavailable moduleName='财务中心' />, 'finance:manage')} />
<Route path='finance/commission-rules' element={withAccess(<MvpUnavailable moduleName='财务中心' />, 'finance:manage')} />
```

- [ ] **Step 4: Run tests**

Run:

```bash
npm test -- App.mvp.test.tsx --watchAll=false
npm test -- src/components/mvp/__tests__/MvpUnavailable.test.tsx --watchAll=false
```

Expected: PASS.

---

## Task 5: Convert `/conflict` into an MVP Conflict Workbench

**Files:**

- Create: `frontend/src/pages/conflict/ConflictWorkbench.tsx`
- Create: `frontend/src/pages/conflict/__tests__/ConflictWorkbench.test.tsx`
- Modify: `frontend/src/App.tsx`

### Required Behavior

`/conflict` must not look like an empty database record page. It must present a director-facing conflict workbench:

- Page title: `利益冲突工作台`
- Primary action: `发起利益冲突检查`
- Primary action navigates to `/case/create`
- Explanation: `新案件立案前必须完成利益冲突检查。`
- Section: `待复核冲突事项`
- Empty state: `当前暂无待复核冲突事项`
- Empty state description: `新案件立案时，系统会在提交前进行利益冲突检查。`
- Button: `新建立案`, navigates to `/case/create`

If there is available API data later, table columns should be:

- 案件名称
- 客户名称
- 相对方
- 冲突类型
- 风险等级
- 当前状态
- 提交时间
- 负责人
- 操作

For this MVP task, do not invent fake conflict records. Empty state is acceptable if it is business-friendly and provides the intake path.

### TDD Steps

- [ ] **Step 1: Write failing component test**

Create `frontend/src/pages/conflict/__tests__/ConflictWorkbench.test.tsx`:

```tsx
import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import ConflictWorkbench from '../ConflictWorkbench'

const navigateMock = jest.fn()

jest.mock('react-router', () => {
  const actual = jest.requireActual('react-router')
  return {
    ...actual,
    useNavigate: () => navigateMock,
  }
})

describe('ConflictWorkbench', () => {
  beforeEach(() => {
    navigateMock.mockClear()
  })

  it('shows director-facing conflict check entry and business empty state', async () => {
    render(
      <MemoryRouter>
        <ConflictWorkbench />
      </MemoryRouter>,
    )

    expect(screen.getByText('利益冲突工作台')).toBeInTheDocument()
    expect(screen.getByText('新案件立案前必须完成利益冲突检查。')).toBeInTheDocument()
    expect(screen.getByText('待复核冲突事项')).toBeInTheDocument()
    expect(screen.getByText('当前暂无待复核冲突事项')).toBeInTheDocument()
    expect(
      screen.getByText('新案件立案时，系统会在提交前进行利益冲突检查。'),
    ).toBeInTheDocument()
    expect(screen.queryByText(/暂无数据库冲突检测记录/)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: '发起利益冲突检查' }))
    expect(navigateMock).toHaveBeenCalledWith('/case/create')
  })
})
```

- [ ] **Step 2: Run test and verify RED**

Run:

```bash
npm test -- src/pages/conflict/__tests__/ConflictWorkbench.test.tsx --watchAll=false
```

Expected: FAIL because `ConflictWorkbench` does not exist.

- [ ] **Step 3: Implement minimal workbench**

Create `frontend/src/pages/conflict/ConflictWorkbench.tsx`:

```tsx
import React from 'react'
import { Button, Card, Empty, Space, Table, Tag, Typography } from 'antd'
import { SafetyCertificateOutlined, PlusOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router'
import type { ColumnsType } from 'antd/es/table'

const { Title, Paragraph } = Typography

interface ConflictReviewItem {
  id: string
  caseName: string
  clientName: string
  opposingParty: string
  conflictType: string
  riskLevel: string
  status: string
  submittedAt: string
  owner: string
}

const columns: ColumnsType<ConflictReviewItem> = [
  { title: '案件名称', dataIndex: 'caseName', key: 'caseName' },
  { title: '客户名称', dataIndex: 'clientName', key: 'clientName' },
  { title: '相对方', dataIndex: 'opposingParty', key: 'opposingParty' },
  { title: '冲突类型', dataIndex: 'conflictType', key: 'conflictType' },
  {
    title: '风险等级',
    dataIndex: 'riskLevel',
    key: 'riskLevel',
    render: (value: string) => <Tag color={value === '高' ? 'red' : 'orange'}>{value}</Tag>,
  },
  { title: '当前状态', dataIndex: 'status', key: 'status' },
  { title: '提交时间', dataIndex: 'submittedAt', key: 'submittedAt' },
  { title: '负责人', dataIndex: 'owner', key: 'owner' },
  {
    title: '操作',
    key: 'action',
    render: () => <Button type='link'>查看</Button>,
  },
]

const ConflictWorkbench: React.FC = () => {
  const navigate = useNavigate()
  const data: ConflictReviewItem[] = []

  return (
    <Space direction='vertical' size={16} style={{ width: '100%' }}>
      <Card>
        <Space align='start' style={{ width: '100%', justifyContent: 'space-between' }}>
          <div>
            <Title level={3} style={{ marginTop: 0 }}>
              利益冲突工作台
            </Title>
            <Paragraph style={{ marginBottom: 0 }}>
              新案件立案前必须完成利益冲突检查。
            </Paragraph>
          </div>
          <Button
            type='primary'
            icon={<PlusOutlined />}
            onClick={() => navigate('/case/create')}
          >
            发起利益冲突检查
          </Button>
        </Space>
      </Card>

      <Card
        title={
          <Space>
            <SafetyCertificateOutlined />
            <span>待复核冲突事项</span>
          </Space>
        }
      >
        <Table
          rowKey='id'
          columns={columns}
          dataSource={data}
          pagination={false}
          locale={{
            emptyText: (
              <Empty
                description={
                  <Space direction='vertical' size={8}>
                    <span>当前暂无待复核冲突事项</span>
                    <span>新案件立案时，系统会在提交前进行利益冲突检查。</span>
                    <Button type='primary' onClick={() => navigate('/case/create')}>
                      新建立案
                    </Button>
                  </Space>
                }
              />
            ),
          }}
        />
      </Card>
    </Space>
  )
}

export default ConflictWorkbench
```

- [ ] **Step 4: Route `/conflict` to workbench**

In `frontend/src/App.tsx`:

```ts
import ConflictWorkbench from './pages/conflict/ConflictWorkbench'
```

Change:

```tsx
<Route path='conflict' element={withAccess(<ConflictCheckResults />, 'conflict:check')} />
```

to:

```tsx
<Route path='conflict' element={withAccess(<ConflictWorkbench />, 'conflict:check')} />
```

Keep `/conflict/v2` as-is for existing advanced check flow.

- [ ] **Step 5: Run tests**

Run:

```bash
npm test -- src/pages/conflict/__tests__/ConflictWorkbench.test.tsx --watchAll=false
```

Expected: PASS.

---

## Task 6: Make Dashboard Quick Actions Match MVP

**Files:**

- Modify: `frontend/src/pages/dashboard/Dashboard.tsx`
- Create: `frontend/src/pages/dashboard/__tests__/DashboardQuickActions.test.tsx`

### Required Behavior

Dashboard quick actions should only expose MVP paths:

- 案件管理 -> `/case`
- 新建立案 -> `/case/create`
- 冲突检测 -> `/conflict`
- 客户管理 -> `/client`
- 审批管理 -> `/approval`
- 信托账户 -> `/trust`

Remove or hide from quick action area:

- 律师管理
- 文件管理
- 法规检索
- 期限计算

Important: Current “新建立案” behavior observed in Chrome was confusing. A button labeled `新建立案` must navigate directly to `/case/create`, not `/case`.

### TDD Steps

- [ ] **Step 1: Write failing quick action test**

Because current `Dashboard.tsx` fetches data and is large, prefer extracting quick action definitions to a small pure export:

Create or modify in `Dashboard.tsx`:

```ts
export const MVP_DASHBOARD_QUICK_ACTIONS = [...]
```

First write test expecting this export:

```tsx
import { MVP_DASHBOARD_QUICK_ACTIONS } from '../Dashboard'

describe('Dashboard MVP quick actions', () => {
  it('contains only MVP quick actions with correct destinations', () => {
    expect(MVP_DASHBOARD_QUICK_ACTIONS.map((item) => item.label)).toEqual([
      '案件管理',
      '新建立案',
      '冲突检测',
      '客户管理',
      '审批管理',
      '信托账户',
    ])

    expect(MVP_DASHBOARD_QUICK_ACTIONS.map((item) => item.path)).toEqual([
      '/case',
      '/case/create',
      '/conflict',
      '/client',
      '/approval',
      '/trust',
    ])
  })
})
```

- [ ] **Step 2: Run test and verify RED**

Run:

```bash
npm test -- src/pages/dashboard/__tests__/DashboardQuickActions.test.tsx --watchAll=false
```

Expected: FAIL because export does not exist or quick actions still include non-MVP items.

- [ ] **Step 3: Implement minimal quick action config**

In `frontend/src/pages/dashboard/Dashboard.tsx`, add:

```ts
export const MVP_DASHBOARD_QUICK_ACTIONS = [
  { key: 'case-management', label: '案件管理', path: '/case' },
  { key: 'case-create', label: '新建立案', path: '/case/create' },
  { key: 'conflict-check', label: '冲突检测', path: '/conflict' },
  { key: 'client-management', label: '客户管理', path: '/client' },
  { key: 'approval-management', label: '审批管理', path: '/approval' },
  { key: 'trust-management', label: '信托账户', path: '/trust' },
] as const
```

Update `handleQuickActionClick` to use the path from this config:

```ts
const handleQuickActionClick = (action: string) => {
  const quickAction = MVP_DASHBOARD_QUICK_ACTIONS.find((item) => item.key === action)
  if (quickAction) {
    navigate(quickAction.path)
    return
  }
  console.log(`未知功能: ${action}`)
}
```

Then update rendered quick action blocks to only render the six MVP actions. If the current markup is duplicated, replace it with a `.map()` over `MVP_DASHBOARD_QUICK_ACTIONS`. Keep styling consistent with the current dashboard.

- [ ] **Step 4: Run test and verify GREEN**

Run:

```bash
npm test -- src/pages/dashboard/__tests__/DashboardQuickActions.test.tsx --watchAll=false
```

Expected: PASS.

---

## Task 7: Tighten New Case Intake Conflict Action

**Files:**

- Modify: `frontend/src/pages/case/CreateCase.tsx` or the active component behind `/case/create`
- Add test near active component:
  - `frontend/src/pages/case/__tests__/CreateCaseConflictStep.test.tsx`
  - or `frontend/src/components/case/__tests__/CaseIntakeConflictAction.test.tsx`

### Required Behavior

`/case/create` must keep the conflict step and make the conflict action explicit:

- Step title `利益冲突检查` must remain visible.
- Button `保存并进行利益冲突检查` must not be a no-op.
- If real conflict API is wired: call the API and display result.
- If real conflict API is not reliable yet: show a clear non-fake message and navigate to `/conflict`.
- Never show fake “提交成功” or fake “冲突检查通过”.

Recommended fallback copy:

```text
试用版当前使用样例冲突复核流程，请在利益冲突工作台查看待复核事项。
```

### TDD Steps

- [ ] **Step 1: Identify active component**

`frontend/src/App.tsx` currently routes `/case/create` to `CaseIntakeWorkbench` from `frontend/src/pages/batch01/Batch01Prototype`. However there is also `frontend/src/pages/case/CreateCase.tsx`.

Before coding, verify in Chrome or with source which component is actually rendered at `/case/create`. Modify the active component only.

- [ ] **Step 2: Write failing test**

Test behavior:

```tsx
it('keeps conflict check step and routes conflict action to conflict workbench when API is unavailable', async () => {
  // Render the active /case/create component.
  // Assert text: 利益冲突检查
  // Click: 保存并进行利益冲突检查
  // Assert message copy or navigation to /conflict.
})
```

If the component is too large to test directly, extract a small function:

```ts
export const getConflictCheckFallbackMessage = () =>
  '试用版当前使用样例冲突复核流程，请在利益冲突工作台查看待复核事项。'
```

Then test this pure function first, followed by a smoke component test.

- [ ] **Step 3: Run test and verify RED**

Run the exact test file:

```bash
npm test -- src/pages/case/__tests__/CreateCaseConflictStep.test.tsx --watchAll=false
```

Expected: FAIL because current action is missing, no-op, or not safely handled.

- [ ] **Step 4: Implement minimal behavior**

Implementation rules:

- Keep the `利益冲突检查` step.
- Wire `保存并进行利益冲突检查` to one of:
  - real conflict API path, if already working;
  - fallback message plus `navigate('/conflict')`.
- Do not create or submit a case record unless the user explicitly clicks final submission and API is confirmed.

Fallback implementation shape:

```ts
message.info('试用版当前使用样例冲突复核流程，请在利益冲突工作台查看待复核事项。')
navigate('/conflict')
```

- [ ] **Step 5: Run test and verify GREEN**

Run:

```bash
npm test -- src/pages/case/__tests__/CreateCaseConflictStep.test.tsx --watchAll=false
```

Expected: PASS.

---

## Task 8: Fix TypeScript Errors Without Weakening Types

**Files:**

- Modify only files reported by `npm run type-check`.

### TDD / Type Gate Steps

- [ ] **Step 1: Capture current failures**

Run:

```bash
cd /Users/mac/Desktop/FT/law-oa-go/frontend
npm run type-check
```

Save the first 30 errors into the GLM work log. Do not start random edits.

- [ ] **Step 2: Fix P0 compile blockers first**

Priority:

1. Missing modules/import paths.
2. Component props mismatches.
3. API response shape mismatches.
4. Nullable value usage.
5. Enum/status literal mismatches.

- [ ] **Step 3: Add tests for every behavior-changing type fix**

If a type fix changes runtime behavior, add a Jest test first. Examples:

- API response unwrap logic
- route config behavior
- unavailable module behavior
- conflict fallback behavior

- [ ] **Step 4: Use typed adapters instead of page-level guessing**

If API response shapes vary, create a typed utility like:

```ts
export function unwrapListResponse<T>(response: unknown): T[] {
  if (Array.isArray(response)) {
    return response
  }

  if (response && typeof response === 'object') {
    const value = response as {
      data?: unknown
      items?: unknown
      records?: unknown
    }

    if (Array.isArray(value.items)) return value.items as T[]
    if (Array.isArray(value.records)) return value.records as T[]
    if (Array.isArray(value.data)) return value.data as T[]

    if (value.data && typeof value.data === 'object') {
      const nested = value.data as { items?: unknown; records?: unknown }
      if (Array.isArray(nested.items)) return nested.items as T[]
      if (Array.isArray(nested.records)) return nested.records as T[]
    }
  }

  return []
}
```

Add tests before using it:

```ts
expect(unwrapListResponse<number>({ data: { items: [1, 2] } })).toEqual([1, 2])
expect(unwrapListResponse<number>({ records: [3] })).toEqual([3])
expect(unwrapListResponse<number>(null)).toEqual([])
```

- [ ] **Step 5: Run type-check until clean**

Run:

```bash
npm run type-check
```

Expected: PASS.

---

## Task 9: Fix ESLint Without Disabling Rules

**Files:**

- Modify only files reported by `npm run lint`.

### TDD / Lint Gate Steps

- [ ] **Step 1: Capture current lint failures**

Run:

```bash
cd /Users/mac/Desktop/FT/law-oa-go/frontend
npm run lint
```

Fix reported errors directly. Do not edit ESLint config to hide failures.

- [ ] **Step 2: Fix known accessibility/security failures**

Expected fixes:

```tsx
<a target='_blank' rel='noopener noreferrer'>
```

```tsx
<label htmlFor='clientName'>客户名称</label>
<input id='clientName' />
```

Replace clickable `div` with `button` where possible:

```tsx
<button type='button' onClick={handleClick} className='quick-action-button'>
  ...
</button>
```

If using Ant Design Card:

```tsx
<Card variant='borderless'>
```

instead of deprecated:

```tsx
<Card bordered={false}>
```

- [ ] **Step 3: Run lint until clean**

Run:

```bash
npm run lint
```

Expected: PASS.

---

## Task 10: Full Build and Chrome Acceptance

**Files:**

- No new files expected unless fixing build errors.

### Verification Steps

- [ ] **Step 1: Backend build**

Run:

```bash
cd /Users/mac/Desktop/FT/law-oa-go
go build ./...
```

Expected: PASS.

- [ ] **Step 2: Frontend gates**

Run:

```bash
cd /Users/mac/Desktop/FT/law-oa-go/frontend
npm run type-check
npm run lint
npm run build
npm test -- --watchAll=false
```

Expected: all PASS.

- [ ] **Step 3: Start local app for Chrome test**

Use current project conventions. Expected ports:

- Backend: `8080`
- Frontend: `3003`

Login:

```text
账号：admin
密码：Demo@2026
```

- [ ] **Step 4: Chrome route acceptance**

Verify:

- `/dashboard`: menu only has MVP entries; quick actions only have MVP entries.
- `/dashboard`: clicking `新建立案` goes to `/case/create`.
- `/case`: case list loads.
- `/case/create`: `利益冲突检查` step remains visible.
- `/case/create`: `保存并进行利益冲突检查` is not a no-op and does not fake success.
- `/conflict`: shows `利益冲突工作台`, not raw empty database text.
- `/conflict`: `发起利益冲突检查` goes to `/case/create`.
- `/client`: client list loads.
- `/approval`: approval workbench loads.
- `/trust`: trust account data loads.
- `/file`: shows MVP unavailable page.
- `/finance`: shows MVP unavailable page and no `开发中`.
- `/inbox`: shows MVP unavailable page.
- `/settings`: shows MVP unavailable page.

- [ ] **Step 5: Console and network check**

Chrome Console:

- No red runtime errors.
- No Vite overlay.
- Ant Design deprecation warnings should be removed where touched.

Network:

- Core MVP requests should not return widespread `401`、`403`、`500`.

---

## Final Delivery Report Required From GLM

GLM must output:

```text
MVP 试用收口版已完成。

修改文件：
- ...

保留模块：
- 工作台
- 案件管理
- 新建立案
- 利益冲突
- 客户管理
- 审批工作台
- 信托账户

隐藏/降级模块：
- 文档中心
- 财务中心
- 收件箱
- 系统设置

验证结果：
- go build ./...：通过/失败
- npm run type-check：通过/失败
- npm run lint：通过/失败
- npm run build：通过/失败
- npm test -- --watchAll=false：通过/失败
- Chrome 核心路径验证：通过/失败

遗留说明：
- ...
```

If any verification fails, GLM must include:

- failing command
- exact error summary
- files suspected
- next concrete fix

Do not say “基本完成” if any required gate fails.

