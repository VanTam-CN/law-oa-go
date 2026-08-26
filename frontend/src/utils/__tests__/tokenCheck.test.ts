import { checkTokenConsistency, checkUserInfoConsistency } from '../tokenCheck'

describe('token consistency checks', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
  })

  it('returns absent user info when storage is empty', () => {
    expect(checkUserInfoConsistency()).toEqual({
      userInfo: null,
      hasUserData: false,
    })
  })

  it('returns a valid user info object', () => {
    const userInfo = { id: 'user-1', name: '张律师' }
    localStorage.setItem('law_oa_user_info', JSON.stringify(userInfo))

    expect(checkUserInfoConsistency()).toEqual({
      userInfo,
      hasUserData: true,
    })
  })

  it('treats invalid JSON user info as absent without modifying storage', () => {
    localStorage.setItem('law_oa_user_info', '[object Object]')

    expect(checkUserInfoConsistency()).toEqual({
      userInfo: null,
      hasUserData: false,
    })
    expect(localStorage.getItem('law_oa_user_info')).toBe('[object Object]')
  })

  it.each(['"text"', '42', 'true', '["user"]', 'null'])(
    'treats non-object JSON user info as absent: %s',
    (storedValue) => {
      localStorage.setItem('law_oa_user_info', storedValue)

      expect(checkUserInfoConsistency()).toEqual({
        userInfo: null,
        hasUserData: false,
      })
    },
  )

  it('does not include token fragments in consistency details', () => {
    const encode = (value: object) =>
      btoa(JSON.stringify(value)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
    const header = { alg: 'HS256', typ: 'JWT', kid: 'token-leak-test-key' }
    const body = { sub: 'token-leak-test-subject', exp: Math.floor(Date.now() / 1000) + 3600 }
    const encodedHeader = encode(header)
    const encodedBody = encode(body)
    const signature = 'token-leak-signature-value'
    const token = `${encodedHeader}.${encodedBody}.${signature}`
    localStorage.setItem('auth_token', token)

    const details = checkTokenConsistency().details.join('\n')

    expect(details).not.toContain(token)
    expect(details).not.toContain(encodedHeader)
    expect(details).not.toContain(encodedBody)
    expect(details).not.toContain(signature)
    ;[JSON.stringify(header), JSON.stringify(body), encodedHeader.slice(0, 8), encodedBody.slice(-8), signature.slice(0, 8)].forEach(
      (fragment) => {
        expect(fragment.length).toBeGreaterThanOrEqual(8)
        expect(details).not.toContain(fragment)
      },
    )
    expect(details).toContain('当前token: 已设置')
  })
})
