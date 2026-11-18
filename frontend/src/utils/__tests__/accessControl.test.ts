import { hasPermission } from '../accessControl'

describe('conflict officer approval access', () => {
  const conflictOfficer = {
    id: '35',
    username: 'conflict-officer',
    email: 'officer@example.test',
    realName: '独立冲突核查人',
    roles: ['conflict_officer'],
    permissions: [],
    isActive: true,
    createdAt: '2026-07-31T00:00:00Z',
  }

  it('can enter the approval queue without receiving general approval management', () => {
    expect(hasPermission(conflictOfficer, 'approval:view')).toBe(true)
    expect(hasPermission(conflictOfficer, 'approval:manage')).toBe(false)
    expect(hasPermission(conflictOfficer, 'case:manage')).toBe(false)
    expect(hasPermission(conflictOfficer, 'conflict:governance')).toBe(true)
  })
})

describe('conflict governance roles', () => {
  it.each(['director', 'partner', 'management', 'compliance', 'risk', 'risk_control'])(
    '%s can open the dashboard and conflict governance',
    (role) => {
      const user = { roles: [role], permissions: [] } as any
      expect(hasPermission(user, 'dashboard:view')).toBe(true)
      expect(hasPermission(user, 'conflict:governance')).toBe(true)
    },
  )

  it('keeps technical administrators out of professional conflict governance', () => {
    const user = { roles: ['admin'], permissions: [] } as any
    expect(hasPermission(user, 'conflict:governance')).toBe(false)
  })
})

describe('compliance reviewer approval access', () => {
  it('can view assigned conflict approvals without receiving general approval management', () => {
    const user = { roles: ['compliance'], permissions: [] } as any
    expect(hasPermission(user, 'approval:view')).toBe(true)
    expect(hasPermission(user, 'approval:manage')).toBe(false)
    expect(hasPermission(user, 'conflict:governance')).toBe(true)
  })
})
