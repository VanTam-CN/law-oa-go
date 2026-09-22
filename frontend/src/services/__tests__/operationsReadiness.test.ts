import { summarizeOperationsReadiness } from '../operationsReadiness'

const verifiedEvidence = {
  verificationStatus: 'verified' as const,
  reference: 'controlled-environment-record',
  verifiedAt: '2026-08-26T10:00:00+08:00',
  verifiedBy: 'Independent Reviewer',
}

describe('operations readiness evidence boundary', () => {
  it('keeps every operations control pending when no evidence exists', () => {
    const summary = summarizeOperationsReadiness()

    expect(summary.ready).toBe(false)
    expect(summary.verifiedCount).toBe(0)
    expect(summary.pendingCount).toBe(5)
    expect(summary.items.every((item) => item.status === 'pending-evidence')).toBe(true)
  })

  it('does not let a healthy service substitute for missing operations evidence', () => {
    const summary = summarizeOperationsReadiness({}, 'healthy')

    expect(summary.ready).toBe(false)
    expect(summary.pendingCount).toBe(5)
  })

  it('only becomes ready after all five controls have independently verified evidence', () => {
    const summary = summarizeOperationsReadiness(
      {
        backup: verifiedEvidence,
        restoreDrill: verifiedEvidence,
        incidentOwner: verifiedEvidence,
        upgrade: verifiedEvidence,
      },
      'healthy',
    )

    expect(summary.ready).toBe(false)
    expect(summary.pendingCount).toBe(1)
    expect(summary.items.find((item) => item.id === 'rollback')?.status).toBe('pending-evidence')
  })
})
