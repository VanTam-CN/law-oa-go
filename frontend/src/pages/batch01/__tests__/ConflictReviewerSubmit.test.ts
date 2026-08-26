import { isAssignedConflictReviewer } from '../Batch01Prototype'

describe('assigned conflict reviewer submit gate', () => {
  it('allows only the currently assigned reviewer to submit', () => {
    const assignment = { id: 'assignment-1', reviewer_id: 21 }

    expect(isAssignedConflictReviewer(21, assignment)).toBe(true)
    expect(isAssignedConflictReviewer(20, assignment)).toBe(false)
    expect(isAssignedConflictReviewer('21', assignment)).toBe(true)
  })

  it('blocks submission before an active assignment is loaded', () => {
    expect(isAssignedConflictReviewer(21, {})).toBe(false)
    expect(isAssignedConflictReviewer(21, { reviewer_id: 0 })).toBe(false)
    expect(isAssignedConflictReviewer(undefined, { reviewerId: 21 })).toBe(false)
  })
})
