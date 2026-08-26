import { initializeApp, useAppStore } from '../useAppStore'

function validToken() {
  const encode = (value: object) =>
    btoa(JSON.stringify(value)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
  return `${encode({ alg: 'none', typ: 'JWT' })}.${encode({
    exp: Math.floor(Date.now() / 1000) + 3600,
  })}.`
}

describe('app session initialization', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
    useAppStore.setState({
      user: null,
      isAuthenticated: false,
      isLoading: false,
    })
  })

  it('ignores invalid user info without deleting local authentication data', () => {
    const token = validToken()
    localStorage.setItem('auth_token', token)
    localStorage.setItem('law_oa_user_info', '[object Object]')

    initializeApp()

    expect(useAppStore.getState().isAuthenticated).toBe(false)
    expect(useAppStore.getState().user).toBeNull()
    expect(localStorage.getItem('auth_token')).toBe(token)
    expect(localStorage.getItem('law_oa_user_info')).toBe('[object Object]')
  })
})
