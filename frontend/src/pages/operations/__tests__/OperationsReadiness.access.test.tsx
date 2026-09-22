import React from 'react'
import { render, screen } from '@testing-library/react'
import OperationsReadiness from '../OperationsReadiness'
import { getOperationsReadinessSummary, ServerOperationsReadinessSummary } from '@/services/operationsReadiness'

jest.mock('@/services/operationsReadiness', () => ({
  ...jest.requireActual('@/services/operationsReadiness'),
  getOperationsReadinessSummary: jest.fn(),
  registerOperationsEvidence: jest.fn(),
}))

jest.mock('@/stores/useAppStore', () => ({
  useAppStore: () => ({
    user: {
      roles: ['compliance'],
      permissions: [],
    },
  }),
}))

const emptySummary: ServerOperationsReadinessSummary = {
  scope: 'controlled_pilot',
  ready: false,
  score: 0,
  maximumScore: 7,
  verifiedCount: 0,
  total: 5,
  productionReady: false,
  productionGate: 'production_external_evidence',
  items: [],
}

describe('OperationsReadiness role boundary', () => {
  it('lets compliance review gaps without showing the registration form', async () => {
    const getSummary = getOperationsReadinessSummary as jest.Mock
    getSummary.mockResolvedValue(emptySummary)

    render(<OperationsReadiness />)

    expect(screen.getByText('受控证据')).toBeInTheDocument()
    expect(screen.getByText('当前账号可复核证据')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '登记受控证据' })).not.toBeInTheDocument()
    await screen.findByText('0/7（0/5 项证据）')
    expect(screen.queryByText(/production ready/i)).not.toBeInTheDocument()
  })
})
