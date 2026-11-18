import {
  conflictIntakePrefill,
  conflictPrefillFromRequest,
  intakeFormFromPrefill,
  readConflictIntakePrefill,
  clearConflictIntakePrefill,
  INTAKE_PREFILL_STORAGE_KEY,
} from '../Batch01Prototype'

describe('conflict intake prefill (review fix 2026-09-30)', () => {
  afterEach(() => {
    window.sessionStorage.clear()
    clearConflictIntakePrefill()
  })

  it('builds the prefill only from the check request snapshot and never from matched_subject', () => {
    const conflict = {
      title: '历史队列标题',
      client_name: '历史客户列',
      case_type: 'civil',
      matched_subject: '历史命中主体甲公司',
      search_parameters: {
        caseName: '甲公司股权回购争议初查',
        clientName: '乙咨询',
        caseType: 'commercial',
        parties: [
          { name: '甲公司', role: 'opposing_party' },
          { name: '乙咨询', role: 'client' },
        ],
      },
    }

    const prefill = conflictIntakePrefill(conflict, { description: '拟接受乙咨询委托' })

    expect(prefill).not.toBeNull()
    expect(prefill?.title).toBe('甲公司股权回购争议初查')
    expect(prefill?.clientName).toBe('乙咨询')
    expect(prefill?.caseType).toBe('commercial')
    expect(prefill?.partyNote).toBe('对方当事人：甲公司；委托人：乙咨询')
    expect(prefill?.description).toBe('拟接受乙咨询委托')
    // The historical hit subject must not leak into the neutral prefill.
    expect(JSON.stringify(prefill)).not.toContain('历史命中主体')
    expect(JSON.stringify(prefill)).not.toContain('历史队列标题')
  })

  it('does not fall back to queue row columns when the snapshot has no facts', () => {
    const conflict = {
      title: '队列标题（可能来自历史）',
      client_name: '队列客户（可能来自历史）',
      case_type: 'civil',
      matched_subject: '历史命中主体',
    }

    // Without snapshot facts AND without the lawyer's own summary there is
    // nothing neutral to carry: no queue-column fallback is allowed.
    expect(conflictPrefillFromRequest(conflict)).toBeNull()
    // The lawyer's own typed summary alone is a valid neutral prefill; the
    // workbench then reminds them to re-enter the missing client/case type.
    const descriptionOnly = conflictPrefillFromRequest(conflict, { description: '摘要' })
    expect(descriptionOnly).toEqual({
      title: '',
      clientName: '',
      caseType: '',
      partyNote: '',
      description: '摘要',
    })
  })

  it('maps partyNote into description and leaves opponentName/clientId unresolved', () => {
    const mapped = intakeFormFromPrefill({
      title: '甲公司股权回购争议初查',
      clientName: '乙咨询',
      caseType: 'commercial',
      partyNote: '对方当事人：甲公司',
      description: '拟接受乙咨询委托',
    })

    expect(mapped.title).toBe('甲公司股权回购争议初查')
    expect(mapped.clientName).toBe('乙咨询')
    expect(mapped.caseType).toBe('commercial')
    expect(mapped.description).toBe('对方当事人：甲公司\n拟接受乙咨询委托')
    // The structured opponent field stays blank for the lawyer to fill
    // explicitly; no client link is guessed from a name.
    expect('opponentName' in mapped).toBe(false)
    expect((mapped as Record<string, unknown>).clientId).toBeUndefined()
  })

  it('reads back only the five whitelisted string fields from sessionStorage', () => {
    window.sessionStorage.setItem(
      INTAKE_PREFILL_STORAGE_KEY,
      JSON.stringify({
        title: '初查标题',
        partyNote: '对方当事人：甲公司',
        riskLevel: 'HIGH',
        identityNumber: '91310000TAMPER0000',
        checkResult: { decision: 'CLEAR' },
      }),
    )

    expect(readConflictIntakePrefill()).toEqual({
      title: '初查标题',
      clientName: '',
      caseType: '',
      partyNote: '对方当事人：甲公司',
      description: '',
    })
  })
})
