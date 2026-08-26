import React from 'react'
import { render, screen, waitFor, within } from '@testing-library/react'
import OperationsReadiness from '../OperationsReadiness'
import { getOperationsReadinessSummary, OPERATIONS_READINESS_REQUIREMENTS } from '@/services/operationsReadiness'

jest.mock('@/services/operationsReadiness', () => {
  const actual = jest.requireActual('@/services/operationsReadiness')
  return {
    ...actual,
    getOperationsReadinessSummary: jest.fn(),
    registerOperationsEvidence: jest.fn(),
  }
})

const emptySummary = {
  scope: 'controlled_pilot', ready: false, score: 0, maximumScore: 7,
  verifiedCount: 0, total: 5, productionReady: false, productionGate: 'production_external_evidence',
  items: ['backup', 'restoreDrill', 'incidentOwner', 'upgrade', 'rollback'].map((control) => ({
    control, status: 'pending-evidence',
  })),
}

describe('OperationsReadiness', () => {
  beforeEach(() => {
    (getOperationsReadinessSummary as jest.Mock).mockResolvedValue(emptySummary)
  })

  it('shows every required control as pending evidence with a next action', async () => {
    render(<OperationsReadiness />)

    expect(screen.getByText('健康检查通过不等于运维已就绪')).toBeInTheDocument()
    await waitFor(() => expect(getOperationsReadinessSummary).toHaveBeenCalledWith('controlled_pilot'))
    expect(await screen.findByText('0/7（0/5 项证据）')).toBeInTheDocument()
    const table = await screen.findByRole('table')
    expect(await within(table).findAllByText('待补证据')).toHaveLength(5)

    for (const requirement of OPERATIONS_READINESS_REQUIREMENTS) {
      expect(within(table).getAllByText(requirement.title).length).toBeGreaterThan(0)
      expect(within(table).getByText(requirement.nextAction)).toBeInTheDocument()
    }
  })

  it('does not display a completed readiness state while evidence is missing', async () => {
    render(<OperationsReadiness />)

    await screen.findByText('0/7（0/5 项证据）')
    expect(screen.queryByText(/production ready/i)).not.toBeInTheDocument()
    expect(screen.getByText(/本页最高 7\/10/)).toBeInTheDocument()
  })
})
