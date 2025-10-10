import { ChromeDevToolsService } from '../../src/mcp/devtools-service';
import { ChromeDevToolsConfig, DevToolsElement } from '../../src/types';
import { jest } from '@jest/globals';

// Mock MCPClient
jest.mock('../../src/mcp/mcp-client');
const { MCPClient } = require('../../src/mcp/mcp-client');

// Mock setTimeout to speed up tests
jest.spyOn(global, 'setTimeout').mockImplementation((callback: Function, _delay?: number) => {
  if (callback) callback();
  return {} as NodeJS.Timeout;
});

describe('ChromeDevToolsService', () => {
  let service: ChromeDevToolsService;
  let config: ChromeDevToolsConfig;
  let mockMCPClient: any;

  beforeEach(() => {
    config = {
      enabled: true,
      headless: false,
      slowMo: 100,
      defaultTimeout: 30000,
      viewport: { width: 1280, height: 720 },
      userAgent: 'Test Agent',
      ignoreHTTPSErrors: true,
    };

    // Create mock MCPClient instance
    mockMCPClient = {
      initialize: jest.fn(),
      executeRequest: jest.fn(),
      isHealthy: jest.fn().mockReturnValue(true),
      close: jest.fn(),
    };

    // Mock the MCPClient constructor
    MCPClient.mockImplementation(() => mockMCPClient);

    service = new ChromeDevToolsService(config);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('constructor', () => {
    it('should create service with config', () => {
      expect(service).toBeInstanceOf(ChromeDevToolsService);
      expect(MCPClient).toHaveBeenCalledWith(
        expect.objectContaining({
          serviceName: 'chrome-devtools',
          timeout: config.defaultTimeout,
          retryAttempts: 3,
          retryDelay: 1000,
        }),
        expect.any(Object)
      );
    });

    it('should use default config when none provided', () => {
      const defaultService = new ChromeDevToolsService();
      expect(defaultService).toBeInstanceOf(ChromeDevToolsService);
    });
  });

  describe('initialize', () => {
    it('should initialize successfully', async () => {
      mockMCPClient.initialize.mockResolvedValue(undefined);

      await service.initialize();

      expect(mockMCPClient.initialize).toHaveBeenCalledTimes(1);
    });

    it('should throw error when initialize fails', async () => {
      mockMCPClient.initialize.mockRejectedValue(new Error('Connection failed'));

      await expect(service.initialize()).rejects.toThrow('Connection failed');
    });
  });

  describe('createPage', () => {
    beforeEach(async () => {
      mockMCPClient.initialize.mockResolvedValue(undefined);
      await service.initialize();
    });

    it('should create page successfully', async () => {
      const mockResponse = {
        success: true,
        data: { pageId: 'test-page-1' },
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      const pageId = await service.createPage('https://example.com');

      expect(pageId).toBe('test-page-1');
      expect(mockMCPClient.executeRequest).toHaveBeenCalledWith({
        method: 'browser.new_page',
        params: { url: 'https://example.com' },
      });
    });

    it('should throw error when page creation fails', async () => {
      const mockResponse = {
        success: false,
        error: { message: 'Page creation failed' },
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      await expect(service.createPage()).rejects.toThrow('创建页面失败: Page creation failed');
    });
  });

  describe('navigate', () => {
    beforeEach(async () => {
      mockMCPClient.initialize.mockResolvedValue(undefined);
      await service.initialize();

      // Create a page first
      const mockCreateResponse = {
        success: true,
        data: { pageId: 'test-page-1' },
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockCreateResponse);
      await service.createPage();
    });

    it('should navigate successfully', async () => {
      const mockResponse = { success: true };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      await service.navigate('https://example.com');

      expect(mockMCPClient.executeRequest).toHaveBeenCalledWith({
        method: 'browser.navigate',
        params: {
          pageId: 'test-page-1',
          url: 'https://example.com',
          waitUntil: 'networkidle',
          timeout: 30000,
        },
      });
    });

    it('should use custom options', async () => {
      const mockResponse = { success: true };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      await service.navigate('https://example.com', {
        url: 'https://example.com',
        waitUntil: 'load',
        timeout: 10000,
      });

      expect(mockMCPClient.executeRequest).toHaveBeenCalledWith({
        method: 'browser.navigate',
        params: {
          pageId: 'test-page-1',
          url: 'https://example.com',
          waitUntil: 'load',
          timeout: 10000,
        },
      });
    });

    it('should throw error when no active page', async () => {
      // Create new service without creating page
      const newService = new ChromeDevToolsService(config);
      mockMCPClient.initialize.mockResolvedValue(undefined);
      await newService.initialize();

      await expect(newService.navigate('https://example.com')).rejects.toThrow('没有活动的页面');
    });
  });

  describe('click', () => {
    beforeEach(async () => {
      mockMCPClient.initialize.mockResolvedValue(undefined);
      await service.initialize();

      const mockCreateResponse = {
        success: true,
        data: { pageId: 'test-page-1' },
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockCreateResponse);
      await service.createPage();
    });

    it('should click element successfully', async () => {
      const mockResponse = { success: true };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      await service.click('#submit-button');

      expect(mockMCPClient.executeRequest).toHaveBeenCalledWith({
        method: 'browser.click',
        params: {
          pageId: 'test-page-1',
          selector: '#submit-button',
          button: 'left',
          clickCount: 1,
          delay: 0,
        },
      });
    });

    it('should use custom click options', async () => {
      const mockResponse = { success: true };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      const mockElement: DevToolsElement = {
        uid: 'test-uid',
        tagName: 'button',
        attributes: {},
        visible: true,
        enabled: true,
        x: 0,
        y: 0,
        width: 100,
        height: 50,
      };

      await service.click('#submit-button', {
        element: mockElement,
        button: 'right',
        clickCount: 2,
        delay: 100,
      });

      expect(mockMCPClient.executeRequest).toHaveBeenCalledWith({
        method: 'browser.click',
        params: {
          pageId: 'test-page-1',
          selector: '#submit-button',
          button: 'right',
          clickCount: 2,
          delay: 100,
        },
      });
    });
  });

  describe('fill', () => {
    beforeEach(async () => {
      mockMCPClient.initialize.mockResolvedValue(undefined);
      await service.initialize();

      const mockCreateResponse = {
        success: true,
        data: { pageId: 'test-page-1' },
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockCreateResponse);
      await service.createPage();
    });

    it('should fill input successfully', async () => {
      const mockResponse = { success: true };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      await service.fill('#username', 'testuser');

      expect(mockMCPClient.executeRequest).toHaveBeenCalledWith({
        method: 'browser.fill',
        params: {
          pageId: 'test-page-1',
          selector: '#username',
          value: 'testuser',
          clear: true,
          delay: 0,
        },
      });
    });
  });

  describe('screenshot', () => {
    beforeEach(async () => {
      mockMCPClient.initialize.mockResolvedValue(undefined);
      await service.initialize();

      const mockCreateResponse = {
        success: true,
        data: { pageId: 'test-page-1' },
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockCreateResponse);
      await service.createPage();
    });

    it('should take screenshot successfully', async () => {
      const mockResponse = {
        success: true,
        data: { screenshot: 'base64-image-data' },
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      const screenshot = await service.screenshot();

      expect(screenshot).toBe('base64-image-data');
      expect(mockMCPClient.executeRequest).toHaveBeenCalledWith({
        method: 'browser.screenshot',
        params: {
          pageId: 'test-page-1',
          fullPage: false,
          format: 'png',
          quality: 90,
        },
      });
    });

    it('should use custom screenshot options', async () => {
      const mockResponse = {
        success: true,
        data: { screenshot: 'base64-image-data' },
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      await service.screenshot({
        fullPage: true,
        format: 'jpeg',
        quality: 80,
        selector: '#main-content',
      });

      expect(mockMCPClient.executeRequest).toHaveBeenCalledWith({
        method: 'browser.screenshot',
        params: {
          pageId: 'test-page-1',
          fullPage: true,
          format: 'jpeg',
          quality: 80,
          selector: '#main-content',
        },
      });
    });
  });

  describe('getPageSnapshot', () => {
    beforeEach(async () => {
      mockMCPClient.initialize.mockResolvedValue(undefined);
      await service.initialize();

      const mockCreateResponse = {
        success: true,
        data: { pageId: 'test-page-1' },
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockCreateResponse);
      await service.createPage();
    });

    it('should get page snapshot successfully', async () => {
      const mockSnapshot = {
        url: 'https://example.com',
        title: 'Test Page',
        elements: [],
        timestamp: new Date(),
        viewport: { width: 1280, height: 720 },
      };
      const mockResponse = {
        success: true,
        data: mockSnapshot,
      };
      mockMCPClient.executeRequest.mockResolvedValue(mockResponse);

      const snapshot = await service.getPageSnapshot();

      expect(snapshot).toEqual(mockSnapshot);
      expect(mockMCPClient.executeRequest).toHaveBeenCalledWith({
        method: 'browser.get_snapshot',
        params: {
          pageId: 'test-page-1',
        },
      });
    });
  });

  describe('close', () => {
    it('should close service successfully', async () => {
      mockMCPClient.initialize.mockResolvedValue(undefined);
      await service.initialize();

      await service.close();

      expect(mockMCPClient.close).toHaveBeenCalledTimes(1);
    });

    it('should handle close errors gracefully', async () => {
      mockMCPClient.initialize.mockResolvedValue(undefined);
      mockMCPClient.close.mockRejectedValue(new Error('Close failed'));

      await service.initialize();

      await expect(service.close()).rejects.toThrow('Close failed');
    });
  });

  describe('health check', () => {
    it('should report health status', () => {
      mockMCPClient.isHealthy.mockReturnValue(true);
      expect(service.isHealthy()).toBe(true);

      mockMCPClient.isHealthy.mockReturnValue(false);
      expect(service.isHealthy()).toBe(false);
    });
  });
});