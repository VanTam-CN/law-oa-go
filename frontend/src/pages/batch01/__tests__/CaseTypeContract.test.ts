import {
  intakeCaseTypeContract,
  isIntakeCaseType,
  type IntakeCaseType,
} from '../intake/caseTypeContract'

describe('intake case type contract', () => {
  it('covers every intake UI option with no duplicates', () => {
    expect(new Set(intakeCaseTypeContract).size).toBe(intakeCaseTypeContract.length)
    const uiOptions: { value: string; label: string }[] = [
      { value: 'commercial', label: '商事诉讼' },
      { value: 'civil', label: '民事' },
      { value: 'civil_litigation', label: '民事诉讼' },
      { value: 'construction', label: '建设工程' },
      { value: 'labor', label: '劳动争议' },
      { value: 'intellectual', label: '知识产权' },
      { value: 'criminal', label: '刑事' },
      { value: 'administrative', label: '行政' },
      { value: 'financial', label: '金融商事' },
      { value: 'ma', label: '并购重组' },
    ]
    for (const option of uiOptions) {
      expect(isIntakeCaseType(option.value)).toBe(true)
    }
  })

  it('rejects codes outside the shared formal-check contract', () => {
    expect(isIntakeCaseType('civil_litigation')).toBe(true)
    expect(isIntakeCaseType('ma')).toBe(true)
    expect(isIntakeCaseType('未知类型')).toBe(false)
    expect(isIntakeCaseType('')).toBe(false)
    expect(isIntakeCaseType(null)).toBe(false)
    expect(isIntakeCaseType(123)).toBe(false)
  })

  it('keeps every code non-empty and lowercase', () => {
    for (const code of intakeCaseTypeContract) {
      expect(code.length).toBeGreaterThan(0)
      expect(code).toBe(code.toLowerCase())
    }
  })
});
