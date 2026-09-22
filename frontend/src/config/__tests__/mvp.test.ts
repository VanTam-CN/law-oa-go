import {
  MVP_MENU_KEYS,
  MVP_UNAVAILABLE_PATHS,
  getMvpUnavailableModuleName,
  isMvpMenuKey,
  isMvpUnavailablePath,
} from '../mvp'

describe('mvp route configuration', () => {
  it('keeps director MVP menu keys including conflict', () => {
    expect(MVP_MENU_KEYS).toEqual([
      'dashboard',
      'case',
      'conflict',
      'conflict-governance',
      'client',
      'approval',
      'inbox',
      'trust',
      'operations-readiness',
    ])
  })

  it('does not treat conflict as unavailable', () => {
    expect(isMvpMenuKey('conflict')).toBe(true)
    expect(isMvpUnavailablePath('/conflict')).toBe(false)
  })

  it('marks unfinished modules as unavailable pages', () => {
    expect(MVP_UNAVAILABLE_PATHS).toEqual(['/file', '/finance', '/settings'])
    expect(isMvpUnavailablePath('/finance')).toBe(true)
    expect(isMvpUnavailablePath('/finance/contracts/1')).toBe(true)
    expect(isMvpUnavailablePath('/dashboard')).toBe(false)
  })

  it('maps direct unavailable routes to business module names', () => {
    expect(getMvpUnavailableModuleName('/file')).toBe('文档中心')
    expect(getMvpUnavailableModuleName('/finance/contracts/1')).toBe('财务中心')
    expect(getMvpUnavailableModuleName('/inbox')).toBeNull()
    expect(getMvpUnavailableModuleName('/settings')).toBe('系统设置')
    expect(getMvpUnavailableModuleName('/conflict')).toBeNull()
  })
})
