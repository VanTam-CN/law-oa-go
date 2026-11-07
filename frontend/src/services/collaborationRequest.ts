import { getToken } from '@/utils/storage'

export class ApiRequestError extends Error {
  status: number
  code: string
  context: Record<string, unknown>

  constructor(
    message: string,
    options: { status: number; code: string; context?: Record<string, unknown> },
  ) {
    super(message)
    this.name = 'ApiRequestError'
    this.status = options.status
    this.code = options.code
    this.context = options.context || {}
  }
}

// Keep error codes and versions intact. Never log identity-bearing requests.
export async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken()
  const response = await fetch(`/api/v1${path}`, {
    ...options,
    headers: {
      ...(options.body instanceof FormData ? {} : { 'Content-Type': 'application/json' }),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })
  const body = await response.json().catch(() => ({}))
  if (!response.ok || body.success === false) {
    const detail = typeof body.error === 'object' && body.error !== null ? body.error : {}
    throw new ApiRequestError(
      body.message ||
        detail.message ||
        (typeof body.error === 'string' ? body.error : '') ||
        `请求失败（${response.status}）`,
      {
        status: response.status,
        code: detail.code || body.code || '',
        context: detail.context || body.context || {},
      },
    )
  }
  return (body.data ?? body) as T
}
