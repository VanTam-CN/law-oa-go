import { getOperationsReadinessSummary, registerOperationsEvidence, summarizeOperationsReadiness } from '../operationsReadiness'
import { get, post } from '../http'

jest.mock('../http', () => ({
  get: jest.fn(),
  post: jest.fn(),
}))

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

  it('maps UI control IDs to stable auditable API control IDs', async () => {
    (get as jest.Mock).mockResolvedValue({
      scope: 'qa', ready: true, score: 5, maximumScore: 7, verifiedCount: 5, total: 5,
      productionReady: false, productionGate: 'production_external_evidence',
      items: [{ control: 'restore_drill', status: 'verified', evidence: { control: 'restore_drill' } }],
    })
    const summary = await getOperationsReadinessSummary('qa')
    expect(get).toHaveBeenCalledWith('/operations/readiness/evidence', { scope: 'qa' })
    expect(summary.items[0].control).toBe('restoreDrill')

    ;(post as jest.Mock).mockResolvedValue({ control: 'incident_owner' })
    const evidence = await registerOperationsEvidence({
      control: 'incidentOwner', scope: 'qa', evidenceReference: 'qa://incident-owner',
      reviewedAt: '2026-08-26T10:00:00+08:00',
    })
    expect(post).toHaveBeenCalledWith('/operations/readiness/evidence', {
      control: 'incident_owner', scope: 'qa', evidenceReference: 'qa://incident-owner',
      reviewedAt: '2026-08-26T10:00:00+08:00',
    })
    expect(evidence.control).toBe('incidentOwner')
  })
})
