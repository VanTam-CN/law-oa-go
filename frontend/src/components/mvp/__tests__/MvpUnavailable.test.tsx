import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import MvpUnavailable from '../MvpUnavailable'

const navigateMock = jest.fn()

jest.mock('react-router', () => {
  const actual = jest.requireActual<typeof import('react-router')>('react-router')
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
