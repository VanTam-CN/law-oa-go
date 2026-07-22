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
    expect(screen.getByText('待办中心')).toBeInTheDocument()
    expect(screen.getByText('代管款管理')).toBeInTheDocument()

    expect(screen.queryByText('律师管理')).not.toBeInTheDocument()
    expect(screen.queryByText('文件管理')).not.toBeInTheDocument()
    expect(screen.queryByText('工具箱')).not.toBeInTheDocument()
    expect(screen.queryByText('财务管理')).not.toBeInTheDocument()
    expect(screen.queryByText('系统设置')).not.toBeInTheDocument()
  })
})
