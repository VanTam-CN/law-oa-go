import { login, logout } from '../auth'
import { post } from '../http'

jest.mock('../http', () => ({
  post: jest.fn(),
  get: jest.fn(),
}))

const postMock = post as jest.MockedFunction<typeof post>

describe('auth service', () => {
  it('sends the current token to the logout API', async () => {
    postMock.mockResolvedValue({ success: true })

    await logout('current-access-token')

    expect(postMock).toHaveBeenCalledWith('/auth/logout', { token: 'current-access-token' })
  })

  it('sends account instead of email to the login API', async () => {
    postMock.mockResolvedValue({ success: true })

    await login({ account: 'Lawyer.Wang', password: 'Password123!' })

    expect(postMock).toHaveBeenCalledWith('/auth/login', {
      account: 'Lawyer.Wang',
      password: 'Password123!',
    })
  })
})
