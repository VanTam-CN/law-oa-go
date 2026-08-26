import {
  clearStorage,
  getToken,
  setStoragePersistence,
  setToken,
} from '../storage'

function validToken() {
  const encode = (value: object) =>
    btoa(JSON.stringify(value)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
  return `${encode({ alg: 'none', typ: 'JWT' })}.${encode({ exp: Math.floor(Date.now() / 1000) + 3600 })}.`
}

describe('authentication storage persistence', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
  })

  afterEach(() => {
    clearStorage()
  })

  it('keeps an unremembered login in session storage only', () => {
    const token = validToken()
    setStoragePersistence(false)
    setToken(token)

    expect(sessionStorage.getItem('auth_token')).toBe(token)
    expect(localStorage.getItem('auth_token')).toBeNull()
    expect(getToken()).toBe(token)
  })

  it('persists a remembered login and clears both stores on logout', () => {
    const token = validToken()
    setStoragePersistence(true)
    setToken(token)

    expect(localStorage.getItem('auth_token')).toBe(token)
    expect(sessionStorage.getItem('auth_token')).toBeNull()

    clearStorage()
    expect(localStorage.getItem('auth_token')).toBeNull()
    expect(sessionStorage.getItem('auth_token')).toBeNull()
  })

  it('ignores malformed user info without deleting local authentication data', () => {
    localStorage.setItem('law_oa_user_info', '[object Object]')

    expect(getUserInfo()).toBeNull()
    expect(localStorage.getItem('law_oa_user_info')).toBe('[object Object]')
  })
})
