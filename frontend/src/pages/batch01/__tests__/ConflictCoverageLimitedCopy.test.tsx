import {
  deriveConflictDecisionViewModel,
} from '../Batch01Prototype'

function conflictBase(overrides: Record<string, unknown> = {}) {
  return {
    has_conflict: false,
    check_result: {
      decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COVERAGE_LIMITED' },
      conflictCases: [] as Array<unknown>,
    },
    ...overrides,
  }
}

describe('conflict coverageLimited guidance branches', () => {
  it('names detected hits when coverage is limited AND a conflict exists', () => {
    const view = deriveConflictDecisionViewModel(
      conflictBase({
        has_conflict: true,
        check_result: {
          decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COVERAGE_LIMITED' },
          conflictCases: [{ case_number: 'HIT-1', case_name: '在办冲突案' }],
        },
      }),
    )
    expect(view.coverageLimited).toBe(true)
    expect(view.guidance).toContain('已检出潜在冲突命中')
    expect(view.guidance).not.toContain('未发现匹配证据')
  })

  it('keeps the no-hit wording only for the true no-hit branch', () => {
    const view = deriveConflictDecisionViewModel(conflictBase())
    expect(view.coverageLimited).toBe(true)
    expect(view.guidance).toContain('未发现匹配证据')
  })

  it('still requires human review for a machine-CLEAR check with complete coverage and no verdict', () => {
    const view = deriveConflictDecisionViewModel({
      has_conflict: false,
      check_result: {
        decision: { status: 'CLEAR', coverageStatus: 'COMPLETE' },
        conflictCases: [],
      },
    })
    expect(view.decision).toBe('REVIEW_REQUIRED')
    expect(view.needsHumanReview).toBe(true)
    expect(view.showReviewAction).toBe(true)
    expect(view.canSubmit).toBe(false)
    expect(view.headline).toContain('机器检测未见匹配')
    expect(view.guidance).toContain('独立人工复核尚未给出结论')
  })

  it('unlocks submission only after a human no_conflict verdict', () => {
    const view = deriveConflictDecisionViewModel(
      {
        has_conflict: false,
        check_result: {
          decision: { status: 'CLEAR', coverageStatus: 'COMPLETE' },
          conflictCases: [],
        },
      },
      { review: { decision: 'no_conflict' } },
    )
    expect(view.decision).toBe('CLEAR')
    expect(view.canSubmit).toBe(true)
    expect(view.needsHumanReview).toBe(false)
    expect(view.showReviewAction).toBe(false)
  })
})
