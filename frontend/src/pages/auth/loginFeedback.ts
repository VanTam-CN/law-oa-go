interface ApiErrorPayload {
  error?: {
    details?: string | string[]
    message?: string
  }
  message?: string
}

interface LoginErrorLike {
  code?: string
  message?: string
  response?: {
    status?: number
    data?: ApiErrorPayload
  }
}

const getApiMessage = (error: LoginErrorLike) => {
  const details = error.response?.data?.error?.details

  if (Array.isArray(details)) {
    return details.find(Boolean)
  }

  return details || error.response?.data?.error?.message || error.response?.data?.message
}

/**
 * 将认证失败转换为可操作、且不会泄露服务端细节的用户提示。
 */
export const getLoginErrorMessage = (error: unknown): string => {
  const loginError = (error || {}) as LoginErrorLike
  const status = loginError.response?.status

  switch (status) {
    case 400:
      return getApiMessage(loginError) || '登录信息格式有误，请检查后重试'
    case 401:
      return '账号或密码错误'
    case 403:
      return '账号已停用或无权登录，请联系系统管理员'
    case 429:
      return '登录尝试过于频繁，请稍后再试'
    default:
      if (status && status >= 500) {
        return '系统服务暂不可用，请稍后重试或联系系统管理员'
      }
  }

  if (
    !loginError.response &&
    (loginError.code === 'ERR_NETWORK' ||
      loginError.code === 'ECONNABORTED' ||
      loginError.message === 'Network Error')
  ) {
    return '无法连接系统服务，请检查网络或联系系统管理员'
  }

  return getApiMessage(loginError) || '登录失败，请稍后重试'
}
