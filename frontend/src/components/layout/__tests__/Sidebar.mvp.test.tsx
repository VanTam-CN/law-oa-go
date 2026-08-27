import React from 'react'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import Sidebar from '../Sidebar'

jest.mock('@/stores/useAppStore', () => ({
  useAppStore: jest.fn(() => ({
    user: {
      roles: ['director'],
      permissions: [
        'dashboard:view',
        'case:manage',
        'client:manage',
        'conflict:check',
        'approval:manage',
        'trust:manage',
      ],
    },
  })),
}))

describe('Sidebar MVP mode', () => {
  it('renders the director MVP menu including operations readiness', () => {
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Sidebar collapsed={false} setCollapsed={jest.fn()} />
      </MemoryRouter>,
    )

    expect(screen.getByText('工作台')).toBeInTheDocument()
    expect(screen.getByText('案件管理')).toBeInTheDocument()
    expect(screen.getByText('利益冲突检查')).toBeInTheDocument()
    expect(screen.getByText('冲突治理')).toBeInTheDocument()
    expect(screen.getByText('客户管理')).toBeInTheDocument()
    expect(screen.queryByText('审批中心')).not.toBeInTheDocument()
    expect(screen.getByText('待办中心')).toBeInTheDocument()
    expect(screen.getByText('代管款管理')).toBeInTheDocument()
    expect(screen.getByText('运维准备度')).toBeInTheDocument()

    expect(screen.queryByText('律师管理')).not.toBeInTheDocument()
    expect(screen.queryByText('文件管理')).not.toBeInTheDocument()
    expect(screen.queryByText('工具箱')).not.toBeInTheDocument()
    expect(screen.queryByText('财务管理')).not.toBeInTheDocument()
    expect(screen.queryByText('系统设置')).not.toBeInTheDocument()
  })

  it('keeps operations readiness out of the lawyer sidebar', () => {
    require('@/stores/useAppStore').useAppStore.mockReturnValueOnce({
      user: { roles: ['lawyer'], permissions: [] },
    })

    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Sidebar collapsed={false} setCollapsed={jest.fn()} />
      </MemoryRouter>,
    )

    expect(screen.queryByText('运维准备度')).not.toBeInTheDocument()
  })
})
