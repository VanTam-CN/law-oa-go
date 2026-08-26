import {
  caseIntakeConflictFingerprint,
  conflictMatchesCaseContext,
  conflictRecordMatchType,
  conflictCheckRequiredFieldLabels,
  getConflictCheckFallbackMessage,
  getMissingConflictCheckFields,
  firstRunGuidance,
  persistableCaseIntakeDraft,
  scopedCaseIntakeDraftKey,
} from '../Batch01Prototype'

describe('CaseIntakeWorkbench conflict action', () => {
  it('returns stable required fields before any draft or conflict request', () => {
    const missing = getMissingConflictCheckFields({
      title: ' ',
      clientId: 0,
      clientName: '',
      opponentName: '',
      opponentIdentityNumber: '  ',
      lawyerId: 0,
      caseType: '',
      businessArea: '',
      subArea: '',
    })

    expect(missing).toEqual([
      'title',
      'client',
      'opponentName',
      'opponentIdentityNumber',
      'lawyer',
      'caseType',
      'businessArea',
      'subArea',
    ])
    expect(missing.map((field) => conflictCheckRequiredFieldLabels[field])).toEqual([
      '案件名称',
      '客户',
      '对方当事人',
      '对方身份标识',
      '负责律师',
      '案件类型',
      '业务领域',
      '子领域',
    ])
  })

  it('returns no missing fields for a complete conflict-check input', () => {
    expect(getMissingConflictCheckFields({
      title: '股权争议',
      clientId: 42,
      clientName: '示例客户',
      opponentName: '甲公司',
      opponentIdentityNumber: '91310000TEST000001',
      lawyerId: 7,
      caseType: 'commercial',
      businessArea: '公司与并购',
      subArea: '投资与融资',
    })).toEqual([])
  })

  it('labels a coverage-limited record with no evidence as no match', () => {
    expect(conflictRecordMatchType({
      has_conflict: false,
      check_result: {
        decision: { status: 'REVIEW_REQUIRED', coverageStatus: 'COVERAGE_LIMITED' },
      },
    })).toBe('无匹配')
  })

  it('keeps a clear MVP fallback message for unavailable conflict checks', () => {
    expect(getConflictCheckFallbackMessage()).toBe(
      '试用版当前使用样例冲突复核流程，请在利益冲突工作台查看待复核事项。',
    )
  })

  it('provides a lawyer-readable first-run path and support guidance', () => {
    expect(firstRunGuidance.requiredChecklist).toEqual([
      '案件名称',
      '客户',
      '对方当事人',
      '对方身份标识',
      '负责律师',
      '案件类型',
      '业务领域',
      '子领域',
    ])
    expect(firstRunGuidance.nextStep).toContain('保存最新输入并检测')
    expect(firstRunGuidance.restoreNotice).toContain('不保留对方身份标识')
    expect(firstRunGuidance.restoreNotice).toContain('重新填写')
    expect(firstRunGuidance.saveFailure).toContain('已填写内容仍保留在本页')
    expect(firstRunGuidance.saveFailure).toContain('联系律所管理员')
    expect(firstRunGuidance.help.content).toContain('保存并退出')
    expect(firstRunGuidance.help.content).toContain('联系律所管理员')
    expect(JSON.stringify(firstRunGuidance)).not.toMatch(/API|ms|接口响应|诊断/)
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

  it('never persists an opponent identity number in a browser draft', () => {
    const draft = persistableCaseIntakeDraft({
      intakeId: '', intakeCode: '', idempotencyKey: '', title: '股权争议', caseType: 'commercial',
      businessArea: '', subArea: '', priority: 'medium', disputeAmount: '', sourceChannel: '',
      sourceContact: '', investmentAgreementDate: '', disputeDate: '', breachDate: '',
      proposedFilingDate: '', jurisdiction: '', clientId: 42, clientName: '示例客户',
      opponentName: '甲公司', opponentEntityType: 'LEGAL_PERSON',
      opponentIdentityType: 'SOCIAL_CREDIT_CODE', opponentIdentityNumber: '91310000TEST000001',
      opponentAliases: '', description: '', billingMethod: '', feeBase: '', contingencyRate: '',
      minimumFee: '', lawyerId: 7,
    })

    expect(draft.opponentIdentityNumber).toBe('')
    expect(JSON.stringify(draft)).not.toContain('91310000TEST000001')
  })

  it('invalidates a frozen result when a material conflict input changes', () => {
    const form = { clientId: 42, clientName: '示例客户', opponentName: '甲公司', opponentEntityType: 'LEGAL_PERSON', opponentIdentityType: 'SOCIAL_CREDIT_CODE', opponentIdentityNumber: '91310000TEST000001', opponentAliases: '', title: '股权争议', caseType: 'commercial', lawyerId: 7 }
    const related = { name: '保证人乙', role: 'GUARANTOR', entityType: 'LEGAL_PERSON', identityType: 'SOCIAL_CREDIT_CODE', identityNumber: '91310000TEST000002' }
    const frozen = caseIntakeConflictFingerprint(form, [related])
    expect(caseIntakeConflictFingerprint({ ...form, opponentName: '乙公司' }, [related])).not.toBe(frozen)
    expect(caseIntakeConflictFingerprint({ ...form, lawyerId: 8 }, [related])).not.toBe(frozen)
    expect(caseIntakeConflictFingerprint({ ...form, opponentIdentityNumber: '91310000TEST000099' }, [related])).not.toBe(frozen)
    expect(caseIntakeConflictFingerprint(form, [{ ...related, name: '新增实控人' }])).not.toBe(frozen)
  })

  it('keeps the same fingerprint when related parties are only reordered', () => {
    const form = { clientId: 42, clientName: '示例客户', opponentName: '甲公司', opponentEntityType: 'LEGAL_PERSON', opponentIdentityType: 'SOCIAL_CREDIT_CODE', opponentIdentityNumber: '91310000TEST000001', opponentAliases: '', title: '股权争议', caseType: 'commercial', lawyerId: 7 }
    const partyA = { name: '乙', role: 'GUARANTOR', entityType: 'LEGAL_PERSON', identityType: 'SOCIAL_CREDIT_CODE', identityNumber: '91310000TEST000002' }
    const partyB = { name: '丙', role: 'CONTROLLER', entityType: 'INDIVIDUAL', identityType: 'ID_CARD', identityNumber: '110101199001010011' }
    const left = caseIntakeConflictFingerprint(form, [partyA, partyB])
    const right = caseIntakeConflictFingerprint(form, [partyB, partyA])
    expect(right).toBe(left)
  })
})
