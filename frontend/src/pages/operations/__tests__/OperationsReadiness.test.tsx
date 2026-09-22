import React from 'react'
import { render, screen } from '@testing-library/react'
import OperationsReadiness from '../OperationsReadiness'
import { OPERATIONS_READINESS_REQUIREMENTS } from '@/services/operationsReadiness'

describe('OperationsReadiness', () => {
  it('shows every required control as pending evidence with a next action', () => {
    render(<OperationsReadiness />)

    expect(screen.getByText('健康检查通过不等于运维已就绪')).toBeInTheDocument()
    expect(screen.getByText('0/5 项已有验证证据')).toBeInTheDocument()
    expect(screen.getAllByText('待补证据')).toHaveLength(5)

    for (const requirement of OPERATIONS_READINESS_REQUIREMENTS) {
      expect(screen.getByText(requirement.title)).toBeInTheDocument()
      expect(screen.getByText(requirement.nextAction)).toBeInTheDocument()
    }
  })

  it('does not display a completed readiness state while evidence is missing', () => {
    render(<OperationsReadiness />)

    expect(screen.queryByText('已就绪')).not.toBeInTheDocument()
    expect(screen.queryByText('已完成')).not.toBeInTheDocument()
  })
})
