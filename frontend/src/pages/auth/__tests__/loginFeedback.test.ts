import { getLoginErrorMessage } from '../loginFeedback'

describe('getLoginErrorMessage', () => {
  it.each([
    [401, '账号或密码错误'],
    [403, '账号已停用或无权登录，请联系系统管理员'],
    [429, '登录尝试过于频繁，请稍后再试'],
    [500, '系统服务暂不可用，请稍后重试或联系系统管理员'],
    [503, '系统服务暂不可用，请稍后重试或联系系统管理员'],
  ])('状态码 %s 返回准确反馈', (status, expected) => {
    expect(getLoginErrorMessage({ response: { status } })).toBe(expected)
  })

  it('400 优先显示接口提供的字段提示', () => {
    expect(
      getLoginErrorMessage({
        response: {
          status: 400,
          data: { error: { details: ['请输入账号或邮箱'] } },
        },
      }),
    ).toBe('请输入账号或邮箱')
  })

  it('网络不可达时不误报为账号密码错误', () => {
    expect(getLoginErrorMessage({ code: 'ERR_NETWORK', message: 'Network Error' })).toBe(
      '无法连接系统服务，请检查网络或联系系统管理员',
    )
  })

  it('未知错误使用中性兜底提示', () => {
    expect(getLoginErrorMessage(new Error('internal details'))).toBe('登录失败，请稍后重试')
  })
})
