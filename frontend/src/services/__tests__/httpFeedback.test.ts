import { shouldShowGlobalError } from '../http'

describe('shouldShowGlobalError', () => {
  it('/auth/login 由登录页面独立展示错误', () => {
    expect(shouldShowGlobalError('/auth/login')).toBe(false)
  })

  it('注册请求仍沿用全局错误提示', () => {
    expect(shouldShowGlobalError('/auth/register')).toBe(true)
  })

  it('保留普通业务请求的全局错误提示', () => {
    expect(shouldShowGlobalError('/clients')).toBe(true)
  })

  it('保留既有静默轮询接口行为', () => {
    expect(shouldShowGlobalError('/notifications/stats')).toBe(false)
  })
})
