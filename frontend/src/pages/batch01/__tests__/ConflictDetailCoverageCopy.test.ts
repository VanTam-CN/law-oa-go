import {
  conflictDetailCoverageCopy,
} from '../Batch01Prototype'

type DetailCoverageStatus = 'COMPLETE' | 'COVERAGE_LIMITED' | 'LIMITED' | 'UNKNOWN' | ''

describe('conflict detail no-match copy by coverage status', () => {
  it.each([
    ['COMPLETE', '未发现匹配证据（律所已登记来源覆盖完整，待独立复核确认）'],
    ['COVERAGE_LIMITED', '未发现匹配证据（检索覆盖受限，待独立复核确认）'],
    ['LIMITED', '未发现匹配证据（检索覆盖受限，待独立复核确认）'],
    ['UNKNOWN', '未发现匹配证据（检索覆盖受限，待独立复核确认）'],
    ['', '未发现匹配证据（检索覆盖受限，待独立复核确认）'],
  ] as Array<[DetailCoverageStatus, string]>)(
    'keeps review required for zero hits with coverage %j',
    (coverage, expectedEvidence) => {
      const result = conflictDetailCoverageCopy(coverage, true)
      expect(result.emptyEvidence).toBe(expectedEvidence)
      expect(result.pendingReview).toBe(
        coverage === 'COMPLETE'
          ? '尚未完成人工复核。当前未发现匹配记录，律所已登记来源覆盖完整；独立核查确认前不得继续提交立案。'
          : '尚未完成人工复核。当前未发现匹配记录，检索覆盖完整性尚未确认；独立核查确认前不得继续提交立案。',
      )
      expect(result.emptyEvidence).not.toContain('仅因检索覆盖受限进入')
      expect(result.pendingReview).not.toContain('检索覆盖受限进入')
    },
  )

  it('does not claim real-world completeness for COMPLETE coverage', () => {
    const result = conflictDetailCoverageCopy('COMPLETE', true)
    expect(result.emptyEvidence).not.toMatch(/现实|全量/)
  })
})

describe('conflict detail hit fallback and unknown coverage', () => {
  it('keeps the hit fallback unchanged for matched records', () => {
    const limitedHit = conflictDetailCoverageCopy('COVERAGE_LIMITED', false)
    const completeHit = conflictDetailCoverageCopy('COMPLETE', false)
    expect(limitedHit.emptyEvidence).toBe('历史记录没有结构化证据，请重新运行检测')
    expect(completeHit.emptyEvidence).toBe('历史记录没有结构化证据，请重新运行检测')
  })

  it('treats undefined, null, and unknown coverage as conservatively incomplete', () => {
    for (const coverage of [undefined, null, 'UNKNOWN'] as const) {
      const result = conflictDetailCoverageCopy(coverage, true)
      expect(result.emptyEvidence).toBe('未发现匹配证据（检索覆盖受限，待独立复核确认）')
      expect(result.pendingReview).toBe(
        '尚未完成人工复核。当前未发现匹配记录，检索覆盖完整性尚未确认；独立核查确认前不得继续提交立案。',
      )
      expect(result.pendingReview).not.toContain('来源覆盖完整')
    }
  })
})
