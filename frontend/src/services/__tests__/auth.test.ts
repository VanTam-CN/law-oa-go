import { logout } from '../auth'
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
})
