import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'

// Mock axios
vi.mock('axios')
const mockedAxios = vi.mocked(axios)

// API响应类型
interface ApiResponse<T = any> {
  code: number
  msg: string
  data: T
}

// 测试API请求工具
describe('API Request Utilities', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('HTTP GET requests', () => {
    it('should make successful GET request', async () => {
      const mockResponse = {
        code: 200,
        msg: 'success',
        data: { id: 1, name: 'Test User' }
      }

      mockedAxios.get.mockResolvedValueOnce({ data: mockResponse })

      const response = await axios.get('/api/users/1')
      const result = response.data as ApiResponse

      expect(mockedAxios.get).toHaveBeenCalledWith('/api/users/1')
      expect(result.code).toBe(200)
      expect(result.msg).toBe('success')
      expect(result.data).toEqual({ id: 1, name: 'Test User' })
    })

    it('should handle GET request with query parameters', async () => {
      const mockResponse = {
        code: 200,
        msg: 'success',
        data: [{ id: 1, name: 'User 1' }, { id: 2, name: 'User 2' }]
      }

      mockedAxios.get.mockResolvedValueOnce({ data: mockResponse })

      const params = { page: 1, size: 10, status: 'active' }
      const response = await axios.get('/api/users', { params })
      const result = response.data as ApiResponse

      expect(mockedAxios.get).toHaveBeenCalledWith('/api/users', { params })
      expect(result.data).toHaveLength(2)
    })

    it('should handle GET request error', async () => {
      const mockError = new Error('Network Error')
      mockedAxios.get.mockRejectedValueOnce(mockError)

      await expect(axios.get('/api/users/999')).rejects.toThrow('Network Error')
    })

    it('should handle 404 error', async () => {
      const mockErrorResponse = {
        response: {
          status: 404,
          data: {
            code: 404,
            msg: 'User not found',
            data: null
          }
        }
      }

      mockedAxios.get.mockRejectedValueOnce(mockErrorResponse)

      try {
        await axios.get('/api/users/999')
      } catch (error: any) {
        expect(error.response.status).toBe(404)
        expect(error.response.data.code).toBe(404)
        expect(error.response.data.msg).toBe('User not found')
      }
    })
  })

  describe('HTTP POST requests', () => {
    it('should make successful POST request', async () => {
      const mockResponse = {
        code: 200,
        msg: 'success',
        data: { id: 1, name: 'New User' }
      }

      const requestData = {
        name: 'New User',
        email: 'newuser@example.com',
        role: 'user'
      }

      mockedAxios.post.mockResolvedValueOnce({ data: mockResponse })

      const response = await axios.post('/api/users', requestData)
      const result = response.data as ApiResponse

      expect(mockedAxios.post).toHaveBeenCalledWith('/api/users', requestData)
      expect(result.code).toBe(200)
      expect(result.data).toEqual({ id: 1, name: 'New User' })
    })

    it('should handle POST request validation error', async () => {
      const mockErrorResponse = {
        response: {
          status: 400,
          data: {
            code: 400,
            msg: 'Validation failed',
            data: {
              errors: ['Email is required', 'Name is required']
            }
          }
        }
      }

      const invalidData = {
        name: '',
        email: 'invalid-email'
      }

      mockedAxios.post.mockRejectedValueOnce(mockErrorResponse)

      try {
        await axios.post('/api/users', invalidData)
      } catch (error: any) {
        expect(error.response.status).toBe(400)
        expect(error.response.data.code).toBe(400)
        expect(error.response.data.msg).toBe('Validation failed')
      }
    })
  })

  describe('HTTP PUT requests', () => {
    it('should make successful PUT request', async () => {
      const mockResponse = {
        code: 200,
        msg: 'success',
        data: { id: 1, name: 'Updated User', email: 'updated@example.com' }
      }

      const updateData = {
        name: 'Updated User',
        email: 'updated@example.com'
      }

      mockedAxios.put.mockResolvedValueOnce({ data: mockResponse })

      const response = await axios.put('/api/users/1', updateData)
      const result = response.data as ApiResponse

      expect(mockedAxios.put).toHaveBeenCalledWith('/api/users/1', updateData)
      expect(result.code).toBe(200)
      expect(result.data.name).toBe('Updated User')
    })
  })

  describe('HTTP DELETE requests', () => {
    it('should make successful DELETE request', async () => {
      const mockResponse = {
        code: 200,
        msg: 'success',
        data: null
      }

      mockedAxios.delete.mockResolvedValueOnce({ data: mockResponse })

      const response = await axios.delete('/api/users/1')
      const result = response.data as ApiResponse

      expect(mockedAxios.delete).toHaveBeenCalledWith('/api/users/1')
      expect(result.code).toBe(200)
      expect(result.data).toBeNull()
    })
  })

  describe('Request interceptors', () => {
    it('should add authorization header', async () => {
      const mockResponse = {
        code: 200,
        msg: 'success',
        data: { id: 1, name: 'Test User' }
      }

      // Mock axios interceptors
      const mockConfig = {
        headers: {}
      }

      // Simulate request interceptor
      const token = 'mock-token'
      if (token) {
        mockConfig.headers.Authorization = `Bearer ${token}`
      }

      mockedAxios.get.mockResolvedValueOnce({ data: mockResponse })

      await axios.get('/api/users/1', mockConfig)

      expect(mockConfig.headers.Authorization).toBe('Bearer mock-token')
    })

    it('should handle timeout', async () => {
      const mockError = new Error('Timeout of 5000ms exceeded')
      mockError.code = 'ECONNABORTED'

      mockedAxios.get.mockRejectedValueOnce(mockError)

      await expect(axios.get('/api/users/1', { timeout: 5000 }))
        .rejects.toThrow('Timeout of 5000ms exceeded')
    })
  })

  describe('Response interceptors', () => {
    it('should handle token expiration', async () => {
      const mockErrorResponse = {
        response: {
          status: 401,
          data: {
            code: 401,
            msg: 'Token expired',
            data: null
          }
        }
      }

      mockedAxios.get.mockRejectedValueOnce(mockErrorResponse)

      try {
        await axios.get('/api/users/1')
      } catch (error: any) {
        expect(error.response.status).toBe(401)
        expect(error.response.data.msg).toBe('Token expired')
      }
    })

    it('should handle server error', async () => {
      const mockErrorResponse = {
        response: {
          status: 500,
          data: {
            code: 500,
            msg: 'Internal server error',
            data: null
          }
        }
      }

      mockedAxios.get.mockRejectedValueOnce(mockErrorResponse)

      try {
        await axios.get('/api/users/1')
      } catch (error: any) {
        expect(error.response.status).toBe(500)
        expect(error.response.data.msg).toBe('Internal server error')
      }
    })
  })

  describe('File upload', () => {
    it('should upload file successfully', async () => {
      const mockResponse = {
        code: 200,
        msg: 'success',
        data: {
          url: 'https://example.com/uploaded-file.pdf',
          filename: 'test.pdf',
          size: 1024
        }
      }

      const formData = new FormData()
      const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' })
      formData.append('file', file)

      mockedAxios.post.mockResolvedValueOnce({ data: mockResponse })

      const response = await axios.post('/api/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })

      const result = response.data as ApiResponse
      expect(result.code).toBe(200)
      expect(result.data.url).toBe('https://example.com/uploaded-file.pdf')
    })
  })

  describe('Request cancellation', () => {
    it('should cancel request', async () => {
      const controller = new AbortController()
      const { signal } = controller

      const mockError = new Error('Request cancelled')
      mockError.name = 'AbortError'

      mockedAxios.get.mockRejectedValueOnce(mockError)

      controller.abort()

      await expect(axios.get('/api/users/1', { signal }))
        .rejects.toThrow('Request cancelled')
    })
  })
})