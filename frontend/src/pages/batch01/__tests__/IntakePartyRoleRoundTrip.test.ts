import {
  buildIntakePartyRows,
  intakeDraftFromServer,
  IntakeFormState,
} from '../Batch01Prototype'

// Real save-contract round trip: rows built exactly like buildIntakePayload
// sends them, then served back the way /case-intakes/:id actually responds
// (metadata as a JSON string, identity ciphertext redacted, no idempotency
// or revision fields on party rows). A hand-written fixture with
// metadata.party_type would pass even when the save path forgot to write it,
// which is the gap Astra flagged on 932972c.
const baseForm: IntakeFormState = {
  intakeId: 'intake-rt',
  intakeCode: 'INT-RT',
  idempotencyKey: 'idem-rt',
  title: 'round-trip-title',
  caseType: 'civil',
  businessArea: '',
  subArea: '',
  priority: 'medium',
  disputeAmount: '',
  sourceChannel: '',
  sourceContact: '',
  investmentAgreementDate: '',
  disputeDate: '',
  breachDate: '',
  proposedFilingDate: '',
  jurisdiction: '',
  clientId: 33,
  clientName: 'Xinghe Zhilian',
  opponentName: 'Opposing Co',
  opponentEntityType: 'LEGAL_PERSON',
  opponentIdentityType: 'SOCIAL_CREDIT_CODE',
  opponentIdentityNumber: '91310000MA1FL0000X',
  opponentAliases: '',
  description: '',
  billingMethod: '',
  feeBase: '',
  contingencyRate: '',
  minimumFee: '',
  lawyerId: 7,
  revision: 'rev-rt',
}

const rows = buildIntakePartyRows(
  baseForm,
  '企业', // enterprise client label from the master-data picker
  [
    {
      name: 'Guarantor Co',
      role: '保证人',
      entityType: 'LEGAL_PERSON',
      identityType: 'SOCIAL_CREDIT_CODE',
      identityNumber: '91310000MA1FL1111Y',
      depth: 2,
    },
  ],
)

// Server response shape: metadata serialized, identity redacted.
const serverPartyRows = rows.map((row) => {
  const { identity_number: _redacted, ...served } = row as Record<string, unknown>
  return { ...served, metadata: JSON.stringify((row as any).metadata) }
})

describe('intake party save -> server -> restore round trip', () => {
  it('writes the business role and depth into the party metadata contract', () => {
    const related = rows.find((row) => row.party_role === 'related_party') as any
    expect(related.metadata.party_type).toBe('保证人')
    expect(related.relation_depth).toBe(2)
  })

  it('restores the related party role and depth from the served party rows', () => {
    const payload = {
      intake: {
        id: 'intake-rt',
        case_type: 'civil',
        priority: 'medium',
        client_id: 33,
        metadata: JSON.stringify({
          case_tags: ['担保', '追加监督'],
        }),
      },
      parties: serverPartyRows,
    }
    const restored = intakeDraftFromServer(payload as Record<string, any>)
    expect(restored.caseTags).toEqual(['担保', '追加监督'])
    expect(restored.relatedParties).toHaveLength(1)
    const related = restored.relatedParties[0]
    expect(related.name).toBe('Guarantor Co')
    expect(related.role).toBe('保证人')
    expect(related.depth).toBe(2)
    expect(related.identityNumber).toBe('')
  })

  it('flattens the rows back out with the same depth on a re-save', () => {
    const restored = intakeDraftFromServer({
      intake: { id: 'intake-rt', case_type: 'civil', priority: 'medium', metadata: '{}' },
      parties: serverPartyRows,
    } as Record<string, any>)
    const rebuilt = buildIntakePartyRows(restored.form, '企业', restored.relatedParties)
    const related = rebuilt.find((row) => row.party_role === 'related_party') as any
    expect(related.relation_depth).toBe(2)
    expect(related.metadata.party_type).toBe('保证人')
  })

  it('rejects the old contract shape where the save omitted metadata.party_type', () => {
    const legacyRows = serverPartyRows.map((row) => {
      const copy = { ...(row as Record<string, unknown>) } as any
      delete copy.metadata
      return copy
    })
    const restored = intakeDraftFromServer({
      intake: { id: 'intake-x', case_type: 'civil', priority: 'medium', metadata: '{}' },
      parties: legacyRows,
    } as Record<string, any>)
    // Documents the pre-fix behavior this contract closes: without the saved
    // metadata the role degrades to the pending label.
    expect(restored.relatedParties[0].role).toBe('待确认')
  })
})
