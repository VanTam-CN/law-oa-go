import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import { Role } from '@/services/role'
import {
  ConflictGovernanceCenter,
  currentOfficerAppointments,
  officerAppointmentGaps,
  policyGovernanceState,
} from '../Batch01Prototype'

jest.mock('@/utils/storage', () => ({
  ...jest.requireActual('@/utils/storage'),
  getRoles: jest.fn(() => [{ code: 'director', name: '主任' } as Role]),
  getToken: jest.fn(() => 'test-token'),
  getUserInfo: jest.fn(() => ({})),
}))

jest.mock('@/utils/messageHelper', () => ({
  message: {
    error: jest.fn(),
    success: jest.fn(),
  },
  setAppMessage: jest.fn(),
}))

global.fetch = jest.fn(async (input: RequestInfo | URL) => {
  const path = String(input)
  if (path.endsWith('/conflict-v2/governance/policies')) {
    return {
      ok: true,
      json: async () => ({
        success: true,
        data: [{
          package: { id: 'policy-1', policy_version: 'V2026.1', integrity_hash: 'a'.repeat(64) },
          endorsements: [{ id: 'e1', endorsement_type: 'MANAGEMENT', endorser_name: '管理甲' }],
          status: 'PENDING_COMPLIANCE',
        }],
      }),
    } as Response
  }
  if (path.endsWith('/conflict/officer-appointments')) {
    return {
      ok: true,
      json: async () => ({
        success: true,
        data: [{
          id: 'appointment-1',
          officer_name: '合规乙',
          deputy_name: '律师丙',
          appointer_name: '主任丁',
          current: true,
          effective_from: '2026-01-01T00:00:00Z',
          effective_to: '2026-12-31T23:59:59Z',
          recusal_declaration: '本人与本案当事人及承办律师无利益冲突，已回避相关关系。',
          deputy_id: 3,
        }],
      }),
    } as Response
  }
  return { ok: true, json: async () => ({ success: true, data: [] }) } as Response
}) as unknown as typeof fetch

describe('ConflictGovernanceCenter closure', () => {
  beforeEach(() => {
    sessionStorage.setItem('law_oa_session_only', '1')
    sessionStorage.setItem('law_oa_roles', JSON.stringify([{ code: 'director', name: '主任' }]))
  })

  it('derives the next policy action without manufacturing approval', () => {
    const state = policyGovernanceState({
      package: { id: 'p1' },
      endorsements: [{ endorsement_type: 'MANAGEMENT', endorser_name: '管理甲' }],
      status: 'PENDING_COMPLIANCE',
    })

    expect(state.managementStatus).toBe('管理确认：管理甲')
    expect(state.complianceStatus).toBe('合规确认：待合规负责人处理')
    expect(state.nextAction).toContain('另一名合规负责人确认')
  })

  it('keeps approved packages subject to production health and audit evidence', () => {
    const state = policyGovernanceState({
      package: { id: 'p1' },
      endorsements: [
        { endorsement_type: 'MANAGEMENT', endorser_name: '管理甲' },
        { endorsement_type: 'COMPLIANCE', endorser_name: '合规乙' },
      ],
      status: 'APPROVED',
    })

    expect(state.nextAction).toContain('生产健康检查和审计记录')
  })

  it('filters current officer terms and reports appointment gaps', () => {
    expect(currentOfficerAppointments([{ id: 'a', current: false }, { id: 'b', current: true }])
      .map((item) => item.id)).toEqual(['b'])
    expect(officerAppointmentGaps({ id: 'a', recusal_declaration: '已回避' })).toEqual(['缺少独立代理人', '回避声明不足'])
    expect(officerAppointmentGaps({
      id: 'a',
      deputy_id: 3,
      recusal_declaration: '本人与案件相关方不存在需要披露的利益冲突关系。',
    })).toEqual([])
  })

  it('renders dual endorsement, next action, appointment, and recusal evidence', async () => {
    render(<ConflictGovernanceCenter />)

    expect(await screen.findByText('合规确认：待合规负责人处理')).toBeInTheDocument()
    expect(screen.getByText(/另一名合规负责人确认同一 SHA-256 材料包/)).toBeInTheDocument()
    expect(screen.getByText(/主核查人：合规乙 · 代理人：律师丙/)).toBeInTheDocument()
    expect(screen.getByText(/任命人：主任丁/)).toBeInTheDocument()
    expect(screen.getByText(/回避声明：本人与本案当事人及承办律师无利益冲突/)).toBeInTheDocument()
    expect(screen.getByText('任命与回避完备')).toBeInTheDocument()
    await waitFor(() => expect(screen.queryByText(/加载失败/)).not.toBeInTheDocument())
  })
})
