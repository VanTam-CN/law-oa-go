import { intakeDraftFromServer } from '../Batch01Prototype'

const savedMetadata = JSON.stringify({
  business_area: '争议解决',
  sub_area: '商事诉讼',
  dispute_amount: '120000',
  source_channel: '客户转介',
  jurisdiction: '上海',
  billing_method: '固定费',
  fee_base: '计件',
  lawyer_id: 7,
})

// Mirrors the 5ddbdba three-party regression sample: business metadata plus
// client, opposing party and one related party. Restoring and saving again
// must reproduce the same payload, so no field may be dropped on restore.
const threePartyPayload = {
  revision: 'rev-server-1',
  intake: {
    id: 'intake-three-party',
    intake_code: 'INT-3P',
    title: '三方担保纠纷',
    case_type: 'civil',
    priority: 'medium',
    description: '担保责任争议',
    client_id: 33,
    created_by: '7',
    status: 'draft',
    metadata: savedMetadata,
  },
  client: { id: 33, name: '星河智联科技有限公司' },
  parties: [
    {
      entity_name: '星河智联科技有限公司',
      entity_type: 'LEGAL_PERSON',
      party_role: 'client',
      relation_depth: 0,
      identity_type: 'SOCIAL_CREDIT_CODE',
    },
    {
      entity_name: '对手公司',
      entity_type: 'LEGAL_PERSON',
      party_role: 'opposing_party',
      relation_depth: 0,
      identity_type: 'SOCIAL_CREDIT_CODE',
      aliases: JSON.stringify(['对手公司上海分部']),
    },
    {
      entity_name: 'Astra虚构保证人公司',
      entity_type: 'LEGAL_PERSON',
      party_role: 'related_party',
      relation_depth: 1,
      identity_type: 'SOCIAL_CREDIT_CODE',
      metadata: JSON.stringify({ party_type: '保证人' }),
    },
  ],
}

describe('intakeDraftFromServer full-fidelity restore', () => {
  it('restores every saved business field after a server draft reload', () => {
    const restored = intakeDraftFromServer(threePartyPayload)
    expect(restored.form.businessArea).toBe('争议解决')
    expect(restored.form.subArea).toBe('商事诉讼')
    expect(restored.form.disputeAmount).toBe('120000')
    expect(restored.form.sourceChannel).toBe('客户转介')
    expect(restored.form.jurisdiction).toBe('上海')
    expect(restored.form.billingMethod).toBe('固定费')
    expect(restored.form.feeBase).toBe('计件')
    expect(restored.form.lawyerId).toBe(7)
    expect(restored.form.revision).toBe('rev-server-1')
    expect(restored.form.clientName).toBe('星河智联科技有限公司')
    expect(restored.form.opponentName).toBe('对手公司')
    expect(restored.form.opponentIdentityNumber).toBe('')
  })

  it('restores the related party row with its stored role and no plaintext identity', () => {
    const restored = intakeDraftFromServer(threePartyPayload)
    expect(restored.relatedParties).toHaveLength(1)
    const related = restored.relatedParties[0]
    expect(related.name).toBe('Astra虚构保证人公司')
    expect(related.role).toBe('保证人')
    expect(related.entityType).toBe('LEGAL_PERSON')
    expect(related.identityType).toBe('SOCIAL_CREDIT_CODE')
    expect(related.identityNumber).toBe('')
  })

  it('restores opponent aliases so a protected identity can re-save safely', () => {
    const restored = intakeDraftFromServer(threePartyPayload)
    expect(restored.form.opponentAliases).toBe('对手公司上海分部')
  })

  it('keeps a switched-in draft clean with no phantom related rows', () => {
    const empty = intakeDraftFromServer({
      revision: 'rev-2',
      intake: { id: 'intake-single', metadata: '{}' },
      parties: [],
    })
    expect(empty.relatedParties).toHaveLength(0)
    expect(empty.form.businessArea).toBe('')
    expect(empty.form.opponentAliases).toBe('')
  })
})
