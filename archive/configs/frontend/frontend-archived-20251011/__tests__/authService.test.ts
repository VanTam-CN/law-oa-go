import { login, register } from '../src/services/authService';
import apiClient from '../src/services/api';

// Mock the apiClient
jest.mock('../src/services/api');

describe('authService', () => {
  beforeEach(() => {
    // Clear all mocks before each test
    jest.clearAllMocks();
  });

  describe('login', () => {
    it('should call apiClient.post with correct parameters', async () => {
      // Arrange
      const mockData = {
        email: 'test@example.com',
        password: 'password123'
      };
      
      const mockResponse = {
        token: 'mock-token',
        refreshToken: 'mock-refresh-token',
        user: {
          id: 1,
          name: 'Test User',
          email: 'test@example.com',
          role: 'user',
          status: 'active',
          createdAt: '2023-01-01T00:00:00Z',
          updatedAt: '2023-01-01T00:00:00Z'
        }
      };
      
      (apiClient.post as jest.Mock).mockResolvedValue(mockResponse);

      // Act
      const result = await login(mockData);

      // Assert
      expect(apiClient.post).toHaveBeenCalledWith('/auth/login', mockData);
      expect(result).toEqual(mockResponse);
    });
  });

  describe('register', () => {
    it('should call apiClient.post with correct parameters', async () => {
      // Arrange
      const mockData = {
        name: 'Test User',
        email: 'test@example.com',
        password: 'password123',
        role: 'user'
      };
      
      const mockResponse = {
        token: 'mock-token',
        refreshToken: 'mock-refresh-token',
        user: {
          id: 1,
          name: 'Test User',
          email: 'test@example.com',
          role: 'user',
          status: 'active',
          createdAt: '2023-01-01T00:00:00Z',
          updatedAt: '2023-01-01T00:00:00Z'
        }
      };
      
      (apiClient.post as jest.Mock).mockResolvedValue(mockResponse);

      // Act
      const result = await register(mockData);

      // Assert
      expect(apiClient.post).toHaveBeenCalledWith('/auth/register', mockData);
      expect(result).toEqual(mockResponse);
    });
  });
});