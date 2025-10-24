import { MCPClient } from '../../src/mcp/mcp-client';
import { MCPServiceConfig } from '../../src/types';
import { jest } from '@jest/globals';

// Mock the delay function to speed up tests
jest.spyOn(global, 'setTimeout').mockImplementation((callback: Function) => {
  callback();
  return {} as NodeJS.Timeout;
});

describe('MCPClient', () => {
  let client: MCPClient;
  let config: MCPServiceConfig;

  beforeEach(() => {
    config = {
      serviceName: 'test-service',
      timeout: 5000,
      retryAttempts: 2,
      retryDelay: 100,
    };
    client = new MCPClient(config);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('constructor', () => {
    it('should create MCP client with config', () => {
      expect(client).toBeInstanceOf(MCPClient);
      expect(client.isHealthy()).toBe(false);
    });
  });

  describe('initialize', () => {
    it('should initialize successfully', async () => {
      await client.initialize();
      expect(client.isHealthy()).toBe(true);
      expect(client.getSession()).toBeDefined();
      expect(client.getSession()?.id).toMatch(/^session_/);
    });

    it('should set up session with correct capabilities', async () => {
      await client.initialize();
      const session = client.getSession();

      expect(session).toMatchObject({
        capabilities: {
          browserName: 'Chrome',
          browserVersion: '120.0.0',
          platform: 'Desktop',
        },
        pages: [],
        currentPage: 0,
      });
    });
  });

  describe('executeRequest', () => {
    beforeEach(async () => {
      await client.initialize();
    });

    it('should execute request successfully', async () => {
      const request = {
        method: 'browser.navigate',
        params: { url: 'https://example.com' },
      };

      const response = await client.executeRequest(request);

      expect(response.success).toBe(true);
      expect(response.metadata?.requestId).toMatch(/^req_/);
      expect(response.metadata?.timestamp).toBeInstanceOf(Date);
      expect(response.metadata?.duration).toBeGreaterThan(0);
    });

    it('should handle different request methods', async () => {
      const methods = [
        'browser.list_pages',
        'browser.new_page',
        'browser.click',
        'browser.fill',
        'browser.screenshot',
      ];

      for (const method of methods) {
        const response = await client.executeRequest({
          method,
          params: {},
        });

        expect(response.success).toBe(true);
      }
    });

    it('should throw error when not initialized', async () => {
      const uninitializedClient = new MCPClient(config);

      await expect(uninitializedClient.executeRequest({
        method: 'test',
        params: {},
      })).rejects.toThrow('MCP客户端未初始化');
    });

    it('should include request details in metadata', async () => {
      const request = {
        method: 'browser.navigate',
        params: { url: 'https://test.com' },
        timeout: 3000,
      };

      const response = await client.executeRequest(request);

      expect(response.metadata?.requestId).toBeDefined();
      expect(response.metadata?.timestamp).toBeInstanceOf(Date);
      expect(response.metadata?.duration).toBeGreaterThanOrEqual(0);
    });
  });

  describe('session management', () => {
    it('should return undefined session when not initialized', () => {
      const uninitializedClient = new MCPClient(config);
      expect(uninitializedClient.getSession()).toBeUndefined();
    });

    it('should update session activity on successful request', async () => {
      await client.initialize();
      const initialTime = client.getSession()?.lastActivity;

      // Add small delay to ensure time difference
      await new Promise(resolve => setTimeout(resolve, 10));

      await client.executeRequest({
        method: 'browser.navigate',
        params: { url: 'https://example.com' },
      });

      const updatedTime = client.getSession()?.lastActivity;
      expect(updatedTime?.getTime()).toBeGreaterThan(initialTime?.getTime() || 0);
    });
  });

  describe('connection health', () => {
    it('should report correct health status', () => {
      expect(client.isHealthy()).toBe(false);

      return client.initialize().then(() => {
        expect(client.isHealthy()).toBe(true);
      });
    });
  });

  describe('close', () => {
    it('should close connection successfully', async () => {
      await client.initialize();
      expect(client.isHealthy()).toBe(true);

      await client.close();
      expect(client.isHealthy()).toBe(false);
      expect(client.getSession()).toBeUndefined();
    });

    it('should handle close when not initialized', async () => {
      const uninitializedClient = new MCPClient(config);

      await expect(uninitializedClient.close()).resolves.not.toThrow();
    });
  });

  describe('request timeout', () => {
    beforeEach(async () => {
      await client.initialize();
    });

    it('should handle request with timeout', async () => {
      const request = {
        method: 'browser.navigate',
        params: { url: 'https://example.com' },
        timeout: 1000,
      };

      const startTime = Date.now();
      const response = await client.executeRequest(request);
      const duration = Date.now() - startTime;

      expect(response.success).toBe(true);
      expect(duration).toBeLessThan(2000); // Should be much faster than timeout
    });
  });

  describe('error handling', () => {
    beforeEach(async () => {
      await client.initialize();
    });

    it('should handle malformed request', async () => {
      // Test with invalid method (should still work in mock)
      const response = await client.executeRequest({
        method: 'invalid.method',
        params: {},
      });

      expect(response.success).toBe(true); // Mock returns success for any method
    });

    it('should handle request with empty params', async () => {
      const response = await client.executeRequest({
        method: 'browser.navigate',
        params: {} as any,
      });

      expect(response.success).toBe(true);
    });
  });
});