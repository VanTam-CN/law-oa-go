import {
  caseIntakeConflictFingerprint,
  conflictMatchesCaseContext,
  getConflictCheckFallbackMessage,
  scopedCaseIntakeDraftKey,
} from '../Batch01Prototype'

describe('CaseIntakeWorkbench conflict action', () => {
  it('keeps a clear MVP fallback message for unavailable conflict checks', () => {
    expect(getConflictCheckFallbackMessage()).toBe(
      '试用版当前使用样例冲突复核流程，请在利益冲突工作台查看待复核事项。',
    )
  })

  it('does not treat a matched conflict case as the subject case', () => {
    const task = {
      id: 'CHECK-OTHER',
      title: '其他案件的检测任务',
      conflict_cases: [{ case_id: '35', case_no: 'DEMO-MVP-2026-002', case_name: '当前案件' }],
    }

    expect(conflictMatchesCaseContext(task, '35', 'DEMO-MVP-2026-002', '当前案件')).toBe(false)
  })

  it('matches only explicit subject case identifiers', () => {
    const task = {
      id: 'CHECK-35',
      case_id: '35',
      case_number: 'DEMO-MVP-2026-002',
      title: '当前案件',
      conflict_cases: [{ case_id: '99', case_name: '历史命中案件' }],
    }

    expect(conflictMatchesCaseContext(task, '35', 'DEMO-MVP-2026-002', '当前案件')).toBe(true)
    expect(conflictMatchesCaseContext(task, '99', '', '历史命中案件')).toBe(false)
  })

  it('matches subject identifiers nested in the persisted search parameters', () => {
    const task = {
      id: 'CHECK-44',
      title: '主体案件名称',
      search_parameters: {
        subjectCaseId: '44',
        subjectCaseNumber: 'DEMO-SUBJECT-2026-001',
      },
      conflict_cases: [{ case_id: '44', case_no: '历史命中案件' }],
    }

    expect(conflictMatchesCaseContext(task, '44', 'DEMO-SUBJECT-2026-001', '主体案件名称')).toBe(true)
  })

  it('matches the command-center task title used by a case-context link', () => {
    const task = {
      id: 'CHECK-CURRENT',
      title: '当前案件检测任务',
      conflict_cases: [{ case_id: '99', case_name: '历史命中案件' }],
    }

    expect(conflictMatchesCaseContext(task, '', '', '当前案件检测任务')).toBe(true)
  })

  it('does not use a historical hit case as the subject context', () => {
    const task = {
      id: 'CHECK-OTHER',
      title: '另一项检测',
      conflict_cases: [{ case_id: '44', case_no: 'DEMO-SUBJECT-2026-001', case_name: '当前案件' }],
    }

    expect(conflictMatchesCaseContext(task, '44', 'DEMO-SUBJECT-2026-001', '当前案件')).toBe(false)
  })

  it('isolates local drafts by user and case context', () => {
    expect(scopedCaseIntakeDraftKey('7')).toBe('law-oa-case-intake-draft-v2:7:new')
    expect(scopedCaseIntakeDraftKey('7', '35')).toBe('law-oa-case-intake-draft-v2:7:case-35')
    expect(scopedCaseIntakeDraftKey('8', '35')).not.toBe(scopedCaseIntakeDraftKey('7', '35'))
  })

  it('invalidates a frozen result when a material conflict input changes', () => {
    const form = { clientId: 42, clientName: '示例客户', opponentName: '甲公司', title: '股权争议', caseType: 'commercial', lawyerId: 7 }
    const frozen = caseIntakeConflictFingerprint(form, [{ name: '保证人乙', role: 'GUARANTOR' }])
    expect(caseIntakeConflictFingerprint({ ...form, opponentName: '乙公司' }, [{ name: '保证人乙', role: 'GUARANTOR' }])).not.toBe(frozen)
    expect(caseIntakeConflictFingerprint({ ...form, lawyerId: 8 }, [{ name: '保证人乙', role: 'GUARANTOR' }])).not.toBe(frozen)
    expect(caseIntakeConflictFingerprint(form, [{ name: '新增实控人', role: 'CONTROLLER' }])).not.toBe(frozen)
  })

  it('keeps the same fingerprint when related parties are only reordered', () => {
    const form = { clientId: 42, clientName: '示例客户', opponentName: '甲公司', title: '股权争议', caseType: 'commercial', lawyerId: 7 }
    const left = caseIntakeConflictFingerprint(form, [{ name: '乙', role: 'GUARANTOR' }, { name: '丙', role: 'CONTROLLER' }])
    const right = caseIntakeConflictFingerprint(form, [{ name: '丙', role: 'CONTROLLER' }, { name: '乙', role: 'GUARANTOR' }])
    expect(right).toBe(left)
  })
})
