import { ConfigManager } from '../../src/core/config';
import * as fs from 'fs';
import { jest } from '@jest/globals';

// Mock fs module
jest.mock('fs');
const mockFs = fs as jest.Mocked<typeof fs>;

describe('ConfigManager', () => {
  // ConfigManager instance for testing
  let originalEnv: NodeJS.ProcessEnv;

  beforeEach(() => {
    // Reset environment variables
    originalEnv = process.env;
    process.env = { ...originalEnv };

    // Clear mock calls
    mockFs.existsSync.mockClear();
    mockFs.readFileSync.mockClear();
    mockFs.writeFileSync.mockClear();

    // Reset singleton instance
    (ConfigManager as any).instance = null;

    // Mock file system
    mockFs.existsSync.mockReturnValue(false);
  });

  afterEach(() => {
    // Restore environment variables
    process.env = originalEnv;
  });

  describe('constructor and singleton', () => {
    it('should create singleton instance', () => {
      const instance1 = ConfigManager.getInstance();
      const instance2 = ConfigManager.getInstance();

      expect(instance1).toBe(instance2);
    });

    it('should load default configuration when no config file exists', () => {
      mockFs.existsSync.mockReturnValue(false);

      const manager = ConfigManager.getInstance();
      const config = manager.getConfig();

      expect(config.baseUrl).toBe('http://localhost:3000');
      expect(config.apiUrl).toBe('http://localhost:8080/api');
      expect(config.defaultTimeout).toBe(30000);
    });

    it('should load configuration from file when it exists', () => {
      const fileConfig = {
        baseUrl: 'https://test.example.com',
        apiUrl: 'https://api.test.example.com',
        defaultTimeout: 60000,
      };

      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue(JSON.stringify(fileConfig));

      const manager = ConfigManager.getInstance();
      const config = manager.getConfig();

      expect(config.baseUrl).toBe('https://test.example.com');
      expect(config.apiUrl).toBe('https://api.test.example.com');
      expect(config.defaultTimeout).toBe(60000);
      // Other defaults should remain
      expect(config.retryAttempts).toBe(3);
    });

    it('should merge file config with defaults', () => {
      const fileConfig = {
        baseUrl: 'https://test.example.com',
        customField: 'custom-value',
      };

      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue(JSON.stringify(fileConfig));

      const manager = ConfigManager.getInstance();
      const config = manager.getConfig();

      expect(config.baseUrl).toBe('https://test.example.com');
      expect((config as any).customField).toBe('custom-value');
      expect(config.apiUrl).toBe('http://localhost:8080/api'); // default
    });
  });

  describe('environment variables', () => {
    it('should override config with environment variables', () => {
      process.env['TEST_BASE_URL'] = 'https://env.example.com';
      process.env['TEST_TIMEOUT'] = '45000';
      process.env['TEST_HEADLESS'] = 'true';

      const manager = ConfigManager.getInstance();
      const config = manager.getConfig();

      expect(config.baseUrl).toBe('https://env.example.com');
      expect(config.defaultTimeout).toBe(45000);
      expect(config.headless).toBe(true);
    });

    it('should parse numeric environment variables correctly', () => {
      process.env['TEST_TIMEOUT'] = '45000';
      process.env['TEST_RETRY_ATTEMPTS'] = '5';

      const manager = ConfigManager.getInstance();
      const config = manager.getConfig();

      expect(config.defaultTimeout).toBe(45000);
      expect(config.retryAttempts).toBe(5);
    });

    it('should parse boolean environment variables correctly', () => {
      process.env['TEST_HEADLESS'] = 'false';

      const manager = ConfigManager.getInstance();
      const config = manager.getConfig();

      expect(config.headless).toBe(false);
    });
  });

  describe('getMCPConfig', () => {
    it('should return MCP configuration based on test config', () => {
      const manager = ConfigManager.getInstance();
      const mcpConfig = manager.getMCPConfig();

      expect(mcpConfig.chromeDevTools.enabled).toBe(true);
      expect(mcpConfig.chromeDevTools.config.headless).toBe(false);
      expect(mcpConfig.chromeDevTools.config.defaultTimeout).toBe(30000);
    });
  });

  describe('getChromeDevToolsConfig', () => {
    it('should return Chrome DevTools configuration', () => {
      const manager = ConfigManager.getInstance();
      const chromeConfig = manager.getChromeDevToolsConfig();

      expect(chromeConfig.enabled).toBe(true);
      expect(chromeConfig.headless).toBe(false);
      expect(chromeConfig.defaultTimeout).toBe(30000);
      expect(chromeConfig.userAgent).toBe('Chrome DevTools MCP Test Framework');
    });
  });

  describe('updateConfig', () => {
    it('should update configuration with new values', () => {
      const manager = ConfigManager.getInstance();
      const originalConfig = manager.getConfig();

      manager.updateConfig({
        baseUrl: 'https://updated.example.com',
        headless: true,
      });

      const updatedConfig = manager.getConfig();

      expect(updatedConfig.baseUrl).toBe('https://updated.example.com');
      expect(updatedConfig.headless).toBe(true);
      // Other values should remain unchanged
      expect(updatedConfig.apiUrl).toBe(originalConfig.apiUrl);
    });
  });

  describe('saveConfig', () => {
    it('should save configuration to file', () => {
      mockFs.writeFileSync.mockImplementation(() => {});

      const manager = ConfigManager.getInstance();
      const configPath = '/path/to/custom.config.json';

      manager.saveConfig(configPath);

      expect(mockFs.writeFileSync).toHaveBeenCalledWith(
        configPath,
        expect.stringContaining('"baseUrl": "http://localhost:3000"')
      );
    });

    it('should save to default path when no path provided', () => {
      mockFs.writeFileSync.mockImplementation(() => {});

      const manager = ConfigManager.getInstance();
      manager.saveConfig();

      expect(mockFs.writeFileSync).toHaveBeenCalledWith(
        expect.stringContaining('test.config.json'),
        expect.any(String)
      );
    });
  });

  describe('validateConfig', () => {
    it('should return valid for correct configuration', () => {
      const manager = ConfigManager.getInstance();
      const validation = manager.validateConfig();

      expect(validation.valid).toBe(true);
      expect(validation.errors).toHaveLength(0);
    });

    it('should return errors for missing required fields', () => {
      const manager = ConfigManager.getInstance();
      manager.updateConfig({
        baseUrl: '',
        apiUrl: '',
      });

      const validation = manager.validateConfig();

      expect(validation.valid).toBe(false);
      expect(validation.errors).toContain('baseUrl is required');
      expect(validation.errors).toContain('apiUrl is required');
    });

    it('should return errors for invalid numeric values', () => {
      const manager = ConfigManager.getInstance();
      manager.updateConfig({
        defaultTimeout: -1,
        retryAttempts: -1,
        concurrencyLevel: 0,
      });

      const validation = manager.validateConfig();

      expect(validation.valid).toBe(false);
      expect(validation.errors).toContain('defaultTimeout must be positive');
      expect(validation.errors).toContain('retryAttempts must be non-negative');
      expect(validation.errors).toContain('concurrencyLevel must be positive');
    });
  });

  describe('error handling', () => {
    it('should handle file read errors gracefully', () => {
      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockImplementation(() => {
        throw new Error('File read error');
      });

      expect(() => {
        ConfigManager.getInstance();
      }).not.toThrow();
    });

    it('should handle file write errors gracefully', () => {
      mockFs.writeFileSync.mockImplementation(() => {
        throw new Error('File write error');
      });

      const manager = ConfigManager.getInstance();

      expect(() => {
        manager.saveConfig();
      }).not.toThrow();
    });
  });
});