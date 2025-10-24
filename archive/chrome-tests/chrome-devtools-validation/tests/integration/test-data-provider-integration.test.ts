import {
  MemoryDataProvider,
  FileDataProvider,
  EnvironmentDataProvider,
  CompositeDataProvider,
  CachedDataProvider
} from '../../src/core/test-data-provider';
import { ChromeDevToolsTestExecutionEngine } from '../../src/core/test-execution-engine';
import { TestSuite, TestCase, TestStep } from '../../src/types/test-types';
import { readFileSync, existsSync, unlinkSync, mkdirSync } from 'fs';
import { join } from 'path';
import { jest } from '@jest/globals';

// Mock ChromeDevToolsValidator
jest.mock('../../src/index');
const MockChromeDevToolsValidator = require('../../src/index').ChromeDevToolsValidator;

describe('Data Provider Integration Tests', () => {
  let testOutputDir: string;
  let mockValidator: any;

  beforeEach(() => {
    testOutputDir = join(__dirname, '../test-data');

    // Create test output directory
    if (!existsSync(testOutputDir)) {
      mkdirSync(testOutputDir, { recursive: true });
    }

    // Mock validator
    mockValidator = {
      navigate: jest.fn(),
      click: jest.fn(),
      fill: jest.fn(),
      wait: jest.fn(),
      screenshot: jest.fn(),
      executeScript: jest.fn(),
      close: jest.fn()
    };

    MockChromeDevToolsValidator.mockImplementation(() => mockValidator);
  });

  afterEach(() => {
    // Clean up test files
    const cleanup = (dir: string) => {
      if (existsSync(dir)) {
        const files = require('fs').readdirSync(dir);
        files.forEach((file: string) => {
          const filePath = join(dir, file);
          if (require('fs').statSync(filePath).isDirectory()) {
            cleanup(filePath);
            require('fs').rmdirSync(filePath);
          } else {
            unlinkSync(filePath);
          }
        });
      }
    };
    cleanup(testOutputDir);
    jest.clearAllMocks();
  });

  describe('MemoryDataProvider Integration', () => {
    let provider: MemoryDataProvider;
    let engine: ChromeDevToolsTestExecutionEngine;

    beforeEach(() => {
      provider = new MemoryDataProvider();
      engine = new ChromeDevToolsTestExecutionEngine();
      engine.setDataProvider(provider);
    });

    it('should integrate with test execution engine', async () => {
      // Setup test data
      const testData = {
        loginUrl: 'https://example.com/login',
        credentials: {
          username: 'testuser',
          password: 'testpass123'
        },
        expectedResults: {
          welcomeMessage: 'Welcome back!',
          dashboardElements: ['profile', 'settings', 'logout']
        }
      };

      await provider.setTestData('login-test-data', testData);

      const testSuite: TestSuite = {
        id: 'memory-data-test',
        name: 'Memory Data Provider Test',
        description: 'Test with memory data provider',
        tags: ['memory'],
        timeout: 30000,
        testCases: [
          {
            id: 'test-with-memory-data',
            name: 'Test with Memory Data',
            description: 'Test case using memory data',
            assertions: [],
            tags: ['memory'],
            timeout: 30000,
            priority: 'P1',
            steps: [
              {
                id: 'navigate-to-login',
                name: 'Navigate to Login Page',
                type: 'navigate',
                url: 'https://example.com/login'
              }
            ]
          }
        ]
      };

      mockValidator.executeScript.mockResolvedValue(true);

      const result = await engine.executeSuite(testSuite);

      expect(result).toBeDefined();
      expect(result.total).toBe(1);
      expect(result.passed).toBe(1);
    });

    it('should handle complex data structures', async () => {
      const complexData = {
        users: [
          {
            id: 1,
            name: 'Admin User',
            email: 'admin@example.com',
            roles: ['admin', 'superuser'],
            preferences: {
              theme: 'dark',
              notifications: true,
              language: 'en-US'
            }
          },
          {
            id: 2,
            name: 'Regular User',
            email: 'user@example.com',
            roles: ['user'],
            preferences: {
              theme: 'light',
              notifications: false,
              language: 'zh-CN'
            }
          }
        ],
        configurations: {
          appSettings: {
            timeout: 30000,
            retryAttempts: 3,
            debugMode: false
          },
          featureFlags: {
            newDashboard: true,
            betaFeatures: false
          }
        }
      };

      await provider.setTestData('complex-test-data', complexData);

      // Verify data retrieval
      const retrievedData = await provider.getTestData('complex-test-data');
      expect(retrievedData).toEqual(complexData);

      // Verify nested access
      expect(retrievedData.users[0].name).toBe('Admin User');
      expect(retrievedData.configurations.featureFlags.newDashboard).toBe(true);
    });

    it('should handle data cleanup', async () => {
      // Setup multiple data entries
      await provider.setTestData('temp-data-1', 'value1');
      await provider.setTestData('temp-data-2', 'value2');
      await provider.setTestData('temp-data-3', 'value3');

      expect(provider.getDataSize()).toBe(3);

      // Cleanup
      await provider.cleanupTestData();

      expect(provider.getDataSize()).toBe(0);
      expect(await provider.getTestData('temp-data-1')).toBeUndefined();
    });
  });

  describe('FileDataProvider Integration', () => {
    let provider: FileDataProvider;
    let engine: ChromeDevToolsTestExecutionEngine;

    beforeEach(() => {
      provider = new FileDataProvider(testOutputDir);
      engine = new ChromeDevToolsTestExecutionEngine();
      engine.setDataProvider(provider);
    });

    it('should persist data to file system', async () => {
      const testData = {
        testId: 'file-persistence-test',
        timestamp: new Date().toISOString(),
        data: {
          message: 'Hello, World!',
          count: 42,
          active: true
        }
      };

      await provider.setTestData('persistence-test', testData);

      // Verify file was created
      const filePath = join(testOutputDir, 'persistence-test.json');
      expect(existsSync(filePath)).toBe(true);

      // Verify file contents
      const fileContent = readFileSync(filePath, 'utf-8');
      const parsedContent = JSON.parse(fileContent);
      expect(parsedContent).toEqual(testData);

      // Verify data retrieval
      const retrievedData = await provider.getTestData('persistence-test');
      expect(retrievedData).toEqual(testData);
    });

    it('should handle file path sanitization', async () => {
      const testData = { message: 'Path sanitization test' };

      // Test various problematic keys
      const testKeys = [
        'test with spaces',
        'test@with#special$chars',
        'test/with/slashes',
        'test\\with\\backslashes',
        'test:with:colons',
        'test?with?questions'
      ];

      for (const key of testKeys) {
        await provider.setTestData(key, testData);
        const retrievedData = await provider.getTestData(key);
        expect(retrievedData).toEqual(testData);
      }
    });

    it('should handle concurrent data access', async () => {
      const testData = { timestamp: Date.now() };
      const testKey = 'concurrent-test';

      // Simulate concurrent access
      const promises = Array.from({ length: 10 }, (_, i) =>
        provider.setTestData(`${testKey}-${i}`, { ...testData, index: i })
      );

      await Promise.all(promises);

      // Verify all data was stored
      for (let i = 0; i < 10; i++) {
        const data = await provider.getTestData(`${testKey}-${i}`);
        expect(data).toEqual({ ...testData, index: i });
      }
    });
  });

  describe('EnvironmentDataProvider Integration', () => {
    let originalEnv: NodeJS.ProcessEnv;
    let provider: EnvironmentDataProvider;
    let engine: ChromeDevToolsTestExecutionEngine;

    beforeEach(() => {
      originalEnv = process.env;
      process.env = { ...originalEnv };

      // Setup test environment variables
      process.env['TEST_DATA_LOGIN_URL'] = 'https://test.example.com/login';
      process.env['TEST_DATA_USER_CREDENTIALS'] = JSON.stringify({
        username: 'testuser',
        password: 'testpass123'
      });
      process.env['TEST_CONFIG_TIMEOUT'] = '30000';
      process.env['TEST_CONFIG_RETRY_COUNT'] = '3';

      provider = new EnvironmentDataProvider('TEST_DATA_');
      engine = new ChromeDevToolsTestExecutionEngine();
      engine.setDataProvider(provider);
    });

    afterEach(() => {
      process.env = originalEnv;
    });

    it('should read from environment variables', async () => {
      const loginUrl = await provider.getTestData('login-url');
      expect(loginUrl).toBe('https://test.example.com/login');

      const credentials = await provider.getTestData('user-credentials');
      expect(credentials).toEqual({
        username: 'testuser',
        password: 'testpass123'
      });
    });

    it('should parse JSON environment variables', async () => {
      const credentials = await provider.getTestData('user-credentials');
      expect(typeof credentials).toBe('object');
      expect(credentials.username).toBe('testuser');
      expect(credentials.password).toBe('testpass123');
    });

    it('should handle missing environment variables', async () => {
      const missingData = await provider.getTestData('non-existent');
      expect(missingData).toBeNull();
    });

    it('should handle invalid JSON gracefully', async () => {
      process.env['TEST_DATA_INVALID_JSON'] = '{ invalid: json }';
      const invalidData = await provider.getTestData('invalid-json');
      expect(invalidData).toBe('{ invalid: json }');
    });

    it('should use custom prefix', async () => {
      const customProvider = new EnvironmentDataProvider('CUSTOM_');
      process.env['CUSTOM_TEST_KEY'] = 'custom-value';

      const data = await customProvider.getTestData('test-key');
      expect(data).toBe('custom-value');
    });
  });

  describe('CompositeDataProvider Integration', () => {
    let compositeProvider: CompositeDataProvider;
    let engine: ChromeDevToolsTestExecutionEngine;
    let mockProviders: any[];

    beforeEach(() => {
      mockProviders = [
        {
          getTestData: jest.fn(),
          setTestData: jest.fn(),
          cleanupTestData: jest.fn()
        },
        {
          getTestData: jest.fn(),
          setTestData: jest.fn(),
          cleanupTestData: jest.fn()
        },
        {
          getTestData: jest.fn(),
          setTestData: jest.fn(),
          cleanupTestData: jest.fn()
        }
      ];

      compositeProvider = new CompositeDataProvider(mockProviders);
      engine = new ChromeDevToolsTestExecutionEngine();
      engine.setDataProvider(compositeProvider);
    });

    it('should retrieve data from first available provider', async () => {
      mockProviders[0].getTestData.mockResolvedValue(null);
      mockProviders[1].getTestData.mockResolvedValue({ data: 'from-provider-2' });
      mockProviders[2].getTestData.mockResolvedValue({ data: 'from-provider-3' });

      const data = await compositeProvider.getTestData('test-key');
      expect(data).toEqual({ data: 'from-provider-2' });

      expect(mockProviders[0].getTestData).toHaveBeenCalledWith('test-key');
      expect(mockProviders[1].getTestData).toHaveBeenCalledWith('test-key');
      expect(mockProviders[2].getTestData).not.toHaveBeenCalled();
    });

    it('should set data on first successful provider', async () => {
      mockProviders[0].setTestData.mockRejectedValue(new Error('Provider 1 failed'));
      mockProviders[1].setTestData.mockResolvedValue(undefined);
      mockProviders[2].setTestData.mockResolvedValue(undefined);

      await expect(compositeProvider.setTestData('test-key', 'test-value')).resolves.not.toThrow();

      expect(mockProviders[0].setTestData).toHaveBeenCalledWith('test-key', 'test-value');
      expect(mockProviders[1].setTestData).toHaveBeenCalledWith('test-key', 'test-value');
      expect(mockProviders[2].setTestData).not.toHaveBeenCalled();
    });

    it('should cleanup all providers', async () => {
      mockProviders[0].cleanupTestData.mockResolvedValue(undefined);
      mockProviders[1].cleanupTestData.mockRejectedValue(new Error('Provider 2 failed'));
      mockProviders[2].cleanupTestData.mockResolvedValue(undefined);

      await expect(compositeProvider.cleanupTestData()).resolves.not.toThrow();

      expect(mockProviders[0].cleanupTestData).toHaveBeenCalled();
      expect(mockProviders[1].cleanupTestData).toHaveBeenCalled();
      expect(mockProviders[2].cleanupTestData).toHaveBeenCalled();
    });

    it('should handle provider errors gracefully', async () => {
      mockProviders[0].getTestData.mockRejectedValue(new Error('Provider 1 error'));
      mockProviders[1].getTestData.mockRejectedValue(new Error('Provider 2 error'));
      mockProviders[2].getTestData.mockResolvedValue({ data: 'fallback-data' });

      const data = await compositeProvider.getTestData('test-key');
      expect(data).toEqual({ data: 'fallback-data' });
    });
  });

  describe('CachedDataProvider Integration', () => {
    let underlyingProvider: MemoryDataProvider;
    let cachedProvider: CachedDataProvider;
    let engine: ChromeDevToolsTestExecutionEngine;

    beforeEach(() => {
      underlyingProvider = new MemoryDataProvider();
      cachedProvider = new CachedDataProvider(underlyingProvider, 5000); // 5 second TTL
      engine = new ChromeDevToolsTestExecutionEngine();
      engine.setDataProvider(cachedProvider);
    });

    it('should cache data from underlying provider', async () => {
      const testData = { message: 'Cached data test', timestamp: Date.now() };

      // Setup underlying provider
      await underlyingProvider.setTestData('cache-test', testData);

      // First access - should hit underlying provider
      const data1 = await cachedProvider.getTestData('cache-test');
      expect(data1).toEqual(testData);

      // Verify cache stats
      const stats = cachedProvider.getCacheStats();
      expect(stats.missCount).toBe(1);

      // Second access - should hit cache
      const data2 = await cachedProvider.getTestData('cache-test');
      expect(data2).toEqual(testData);

      // Verify cache stats
      const updatedStats = cachedProvider.getCacheStats();
      expect(updatedStats.hitCount).toBe(1);
      expect(updatedStats.missCount).toBe(1);
    });

    it('should respect TTL', async () => {
      const testData = { message: 'TTL test data' };

      await underlyingProvider.setTestData('ttl-test', testData);

      // Access data
      await cachedProvider.getTestData('ttl-test');

      // Verify cache has data
      expect(cachedProvider.getCacheStats().size).toBe(1);

      // Manually expire cache by changing the timestamp
      jest.spyOn(Date, 'now').mockReturnValue(Date.now() + 6000); // 6 seconds later

      // Access again - should hit underlying provider due to TTL expiry
      await cachedProvider.getTestData('ttl-test');

      // Verify cache stats
      const stats = cachedProvider.getCacheStats();
      expect(stats.missCount).toBe(2);
    });

    it('should not cache null or undefined values', async () => {
      await underlyingProvider.setTestData('null-test', null);

      // Access multiple times
      await cachedProvider.getTestData('null-test');
      await cachedProvider.getTestData('null-test');
      await cachedProvider.getTestData('null-test');

      // Should have missed cache each time
      const stats = cachedProvider.getCacheStats();
      expect(stats.missCount).toBe(3);
      expect(stats.hitCount).toBe(0);
    });

    it('should handle cache cleanup', async () => {
      await underlyingProvider.setTestData('cleanup-test', { data: 'cleanup-data' });

      // Access data to cache it
      await cachedProvider.getTestData('cleanup-test');

      // Verify cache has data
      expect(cachedProvider.getCacheStats().size).toBe(1);

      // Cleanup cache
      cachedProvider.cleanupExpiredCache();

      // Cache should be empty
      expect(cachedProvider.getCacheStats().size).toBe(0);
    });
  });

  describe('End-to-End Data Provider Integration', () => {
    let engine: ChromeDevToolsTestExecutionEngine;
    let compositeProvider: CompositeDataProvider;

    beforeEach(() => {
      // Create composite provider with multiple sources
      const memoryProvider = new MemoryDataProvider();
      const fileProvider = new FileDataProvider(testOutputDir);
      const envProvider = new EnvironmentDataProvider('INTEGRATION_TEST_');

      compositeProvider = new CompositeDataProvider([
        memoryProvider,
        fileProvider,
        envProvider
      ]);

      engine = new ChromeDevToolsTestExecutionEngine();
      engine.setDataProvider(compositeProvider);
    });

    it('should work with real test execution', async () => {
      // Setup test data in different providers
      const memoryProvider = compositeProvider['providers'][0];
      if (memoryProvider) {
        await memoryProvider.setTestData('test-url', 'https://integration-test.example.com');
      }

      // Setup environment variable
      process.env['INTEGRATION_TEST_CREDENTIALS'] = JSON.stringify({
        username: 'integration-user',
        password: 'integration-pass'
      });

      const testSuite: TestSuite = {
        id: 'integration-test-suite',
        name: 'Integration Test Suite',
        description: 'Integration test with composite data provider',
        tags: ['integration'],
        timeout: 30000,
        testCases: [
          {
            id: 'integration-test-case',
            name: 'Integration Test Case',
            description: 'Integration test case',
            assertions: [],
            tags: ['integration'],
            timeout: 30000,
            priority: 'P1',
            steps: [
              {
                id: 'step-1',
                name: 'Navigate to test URL',
                type: 'navigate',
                url: 'https://integration-test.example.com'
              },
              {
                id: 'step-2',
                name: 'Verify page element',
                type: 'verify',
                selector: 'body',
                expectedState: { visible: true }
              }
            ]
          }
        ]
      };

      mockValidator.navigate.mockResolvedValue(undefined);
      mockValidator.executeScript.mockResolvedValue(true);

      const result = await engine.executeSuite(testSuite);

      expect(result).toBeDefined();
      expect(result.total).toBe(1);
      expect(result.passed).toBe(1);
      expect(result.suites[0].testCases[0].steps).toHaveLength(2);
      expect(result.suites[0].testCases[0].steps.every((step: any) => step.status === 'passed')).toBe(true);
    });

    it('should handle complex data-driven scenarios', async () => {
      // Setup test data with multiple test cases
      const testDataSet = {
        loginTests: [
          {
            username: 'user1@example.com',
            password: 'password1',
            loginUrl: 'https://example.com/login/user1',
            expectedSuccess: true
          },
          {
            username: 'user2@example.com',
            password: 'password2',
            loginUrl: 'https://example.com/login/user2',
            expectedSuccess: true
          },
          {
            username: 'invalid@example.com',
            password: 'wrongpassword',
            loginUrl: 'https://example.com/login/invalid',
            expectedSuccess: false
          }
        ],
        config: {
          timeout: 30000,
          retryAttempts: 2
        }
      };

      const memoryProvider = compositeProvider['providers'][0];
      if (memoryProvider) {
        await memoryProvider.setTestData('login-dataset', testDataSet);
      }

      const testSuite: TestSuite = {
        id: 'data-driven-test-suite',
        name: 'Data Driven Test Suite',
        description: 'Data-driven test suite with multiple login scenarios',
        tags: ['data-driven', 'login'],
        timeout: 30000,
        testCases: testDataSet.loginTests.map((testData, index) => ({
          id: `login-test-${index}`,
          name: `Login Test ${index + 1}`,
          description: `Data-driven login test ${index + 1}`,
          assertions: [],
          tags: ['data-driven', 'login'],
          timeout: 30000,
          priority: 'P1',
          steps: [
            {
              id: `login-step-${index}`,
              name: 'Navigate to Login',
              type: 'navigate',
              url: testData.loginUrl
            }
          ]
        }))
      };

      mockValidator.navigate.mockResolvedValue(undefined);

      const result = await engine.executeSuite(testSuite);

      expect(result).toBeDefined();
      expect(result.total).toBe(3);
      expect(result.passed).toBe(3);
      expect(result.suites[0].testCases).toHaveLength(3);
    });
  });
});