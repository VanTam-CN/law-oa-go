import {
  MemoryDataProvider,
  FileDataProvider,
  EnvironmentDataProvider,
  CompositeDataProvider,
  CachedDataProvider
} from '../../src/core/test-data-provider';
import { jest } from '@jest/globals';

// Mock setTimeout for consistent test timing
jest.spyOn(global, 'setTimeout').mockImplementation((callback: Function, delay?: number) => {
  if (callback) callback();
  return {} as NodeJS.Timeout;
});

describe('Test Data Providers', () => {
  describe('MemoryDataProvider', () => {
    let provider: MemoryDataProvider;

    beforeEach(() => {
      provider = new MemoryDataProvider();
    });

    afterEach(() => {
      jest.clearAllMocks();
    });

    describe('getTestData', () => {
      it('should return undefined for non-existent data', async () => {
        const result = await provider.getTestData('non-existent-key');
        expect(result).toBeUndefined();
      });

      it('should return stored data', async () => {
        const testData = { username: 'testuser', password: 'testpass' };
        await provider.setTestData('user-credentials', testData);

        const result = await provider.getTestData('user-credentials');
        expect(result).toEqual(testData);
      });

      it('should return null if stored as null', async () => {
        await provider.setTestData('null-data', null);

        const result = await provider.getTestData('null-data');
        expect(result).toBeNull();
      });
    });

    describe('setTestData', () => {
      it('should store data correctly', async () => {
        const testData = { value: 'test-data' };
        await expect(provider.setTestData('test-key', testData)).resolves.not.toThrow();

        const result = await provider.getTestData('test-key');
        expect(result).toEqual(testData);
      });

      it('should overwrite existing data', async () => {
        await provider.setTestData('test-key', 'old-value');
        await provider.setTestData('test-key', 'new-value');

        const result = await provider.getTestData('test-key');
        expect(result).toBe('new-value');
      });

      it('should store different data types', async () => {
        const testCases = [
          { key: 'string-data', data: 'test-string' },
          { key: 'number-data', data: 123 },
          { key: 'boolean-data', data: true },
          { key: 'object-data', data: { nested: { value: 'test' } } },
          { key: 'array-data', data: [1, 2, 3] },
          { key: 'null-data', data: null }
        ];

        for (const testCase of testCases) {
          await provider.setTestData(testCase.key, testCase.data);
          const result = await provider.getTestData(testCase.key);
          expect(result).toEqual(testCase.data);
        }
      });
    });

    describe('cleanupTestData', () => {
      it('should clear all data', async () => {
        await provider.setTestData('key1', 'value1');
        await provider.setTestData('key2', 'value2');

        expect(provider.getDataSize()).toBe(2);

        await provider.cleanupTestData();

        expect(provider.getDataSize()).toBe(0);
        expect(await provider.getTestData('key1')).toBeUndefined();
        expect(await provider.getTestData('key2')).toBeUndefined();
      });

      it('should handle cleanup when no data exists', async () => {
        await expect(provider.cleanupTestData()).resolves.not.toThrow();
        expect(provider.getDataSize()).toBe(0);
      });
    });

    describe('utility methods', () => {
      it('should return correct data keys', async () => {
        await provider.setTestData('key1', 'value1');
        await provider.setTestData('key2', 'value2');

        const keys = provider.getDataKeys();
        expect(keys).toContain('key1');
        expect(keys).toContain('key2');
        expect(keys).toHaveLength(2);
      });

      it('should check data existence correctly', async () => {
        await provider.setTestData('existing-key', 'value');

        expect(provider.hasData('existing-key')).toBe(true);
        expect(provider.hasData('non-existing-key')).toBe(false);
      });

      it('should return correct data size', async () => {
        expect(provider.getDataSize()).toBe(0);

        await provider.setTestData('key1', 'value1');
        expect(provider.getDataSize()).toBe(1);

        await provider.setTestData('key2', 'value2');
        expect(provider.getDataSize()).toBe(2);
      });
    });
  });

  describe('FileDataProvider', () => {
    let provider: FileDataProvider;

    beforeEach(() => {
      provider = new FileDataProvider('/tmp/test-data');
    });

    afterEach(() => {
      jest.clearAllMocks();
    });

    describe('getTestData', () => {
      it('should return mock data for known keys', async () => {
        const result = await provider.getTestData('user-credentials');
        expect(result).toEqual({
          username: 'testuser',
          password: 'testpass123',
          email: 'test@example.com'
        });
      });

      it('should return null for unknown keys', async () => {
        const result = await provider.getTestData('unknown-key');
        expect(result).toBeNull();
      });

      it('should handle mock file reading delay', async () => {
        const startTime = Date.now();
        await provider.getTestData('user-credentials');
        const endTime = Date.now();

        // Should have some delay due to mock
        expect(endTime - startTime).toBeGreaterThanOrEqual(0);
      });
    });

    describe('setTestData', () => {
      it('should simulate file writing', async () => {
        const testData = { key: 'value' };
        await expect(provider.setTestData('test-key', testData)).resolves.not.toThrow();
      });

      it('should handle different data types', async () => {
        const testCases = [
          { key: 'string-data', data: 'test-string' },
          { key: 'object-data', data: { complex: { nested: 'object' } } }
        ];

        for (const testCase of testCases) {
          await expect(provider.setTestData(testCase.key, testCase.data)).resolves.not.toThrow();
        }
      });
    });

    describe('cleanupTestData', () => {
      it('should handle cleanup operation', async () => {
        await expect(provider.cleanupTestData()).resolves.not.toThrow();
      });
    });

    describe('file path generation', () => {
      it('should generate safe file paths', async () => {
        // Access private method for testing
        const getFilePath = (provider as any).getFilePath.bind(provider);

        expect(getFilePath('test-key')).toBe('/tmp/test-data/test-key.json');
        expect(getFilePath('key with spaces')).toBe('/tmp/test-data/key_with_spaces.json');
        expect(getFilePath('key@with#special$chars')).toBe('/tmp/test-data/key_with_special_chars.json');
      });
    });
  });

  describe('EnvironmentDataProvider', () => {
    let provider: EnvironmentDataProvider;
    let originalEnv: NodeJS.ProcessEnv;

    beforeEach(() => {
      originalEnv = process.env;
      process.env = { ...originalEnv };
      provider = new EnvironmentDataProvider('TEST_DATA_');
    });

    afterEach(() => {
      process.env = originalEnv;
      jest.clearAllMocks();
    });

    describe('getTestData', () => {
      it('should return null for non-existent environment variables', async () => {
        const result = await provider.getTestData('non-existent');
        expect(result).toBeNull();
      });

      it('should return string value from environment', async () => {
        process.env.TEST_DATA_STRING_VALUE = 'test-string';
        const result = await provider.getTestData('string-value');
        expect(result).toBe('test-string');
      });

      it('should parse JSON values from environment', async () => {
        const jsonData = { key: 'value', number: 123 };
        process.env.TEST_DATA_JSON_VALUE = JSON.stringify(jsonData);
        const result = await provider.getTestData('json-value');
        expect(result).toEqual(jsonData);
      });

      it('should return raw string for invalid JSON', async () => {
        process.env.TEST_DATA_INVALID_JSON = '{ invalid: json }';
        const result = await provider.getTestData('invalid-json');
        expect(result).toBe('{ invalid: json }');
      });

      it('should use custom prefix', async () => {
        const customProvider = new EnvironmentDataProvider('CUSTOM_');
        process.env.CUSTOM_TEST_KEY = 'custom-value';

        const result = await customProvider.getTestData('test-key');
        expect(result).toBe('custom-value');
      });

      it('should use default prefix', async () => {
        const defaultProvider = new EnvironmentDataProvider();
        process.env.TEST_DATA_DEFAULT_KEY = 'default-value';

        const result = await defaultProvider.getTestData('default-key');
        expect(result).toBe('default-value');
      });
    });

    describe('setTestData', () => {
      it('should log warning about write-only nature', async () => {
        await expect(provider.setTestData('test-key', 'test-value')).resolves.not.toThrow();
      });
    });

    describe('cleanupTestData', () => {
      it('should handle cleanup without errors', async () => {
        await expect(provider.cleanupTestData()).resolves.not.toThrow();
      });
    });
  });

  describe('CompositeDataProvider', () => {
    let mockProvider1: any;
    let mockProvider2: any;
    let mockProvider3: any;
    let compositeProvider: CompositeDataProvider;

    beforeEach(() => {
      mockProvider1 = {
        getTestData: jest.fn(),
        setTestData: jest.fn(),
        cleanupTestData: jest.fn()
      };

      mockProvider2 = {
        getTestData: jest.fn(),
        setTestData: jest.fn(),
        cleanupTestData: jest.fn()
      };

      mockProvider3 = {
        getTestData: jest.fn(),
        setTestData: jest.fn(),
        cleanupTestData: jest.fn()
      };

      compositeProvider = new CompositeDataProvider([
        mockProvider1, mockProvider2, mockProvider3
      ]);
    });

    afterEach(() => {
      jest.clearAllMocks();
    });

    describe('getTestData', () => {
      it('should return data from first provider that has it', async () => {
        mockProvider1.getTestData.mockResolvedValue(null);
        mockProvider2.getTestData.mockResolvedValue({ data: 'value' });
        mockProvider3.getTestData.mockResolvedValue(null);

        const result = await compositeProvider.getTestData('test-key');
        expect(result).toEqual({ data: 'value' });
        expect(mockProvider1.getTestData).toHaveBeenCalledWith('test-key');
        expect(mockProvider2.getTestData).toHaveBeenCalledWith('test-key');
        expect(mockProvider3.getTestData).not.toHaveBeenCalled();
      });

      it('should return null if no provider has data', async () => {
        mockProvider1.getTestData.mockResolvedValue(null);
        mockProvider2.getTestData.mockResolvedValue(null);
        mockProvider3.getTestData.mockResolvedValue(null);

        const result = await compositeProvider.getTestData('test-key');
        expect(result).toBeNull();
      });

      it('should return data from first provider even if others have errors', async () => {
        mockProvider1.getTestData.mockResolvedValue({ data: 'value' });
        mockProvider2.getTestData.mockRejectedValue(new Error('Provider 2 error'));
        mockProvider3.getTestData.mockRejectedValue(new Error('Provider 3 error'));

        const result = await compositeProvider.getTestData('test-key');
        expect(result).toEqual({ data: 'value' });
        expect(mockProvider1.getTestData).toHaveBeenCalledWith('test-key');
        expect(mockProvider2.getTestData).toHaveBeenCalledWith('test-key');
        expect(mockProvider3.getTestData).not.toHaveBeenCalled();
      });

      it('should handle all providers throwing errors', async () => {
        const error1 = new Error('Provider 1 error');
        const error2 = new Error('Provider 2 error');
        const error3 = new Error('Provider 3 error');

        mockProvider1.getTestData.mockRejectedValue(error1);
        mockProvider2.getTestData.mockRejectedValue(error2);
        mockProvider3.getTestData.mockRejectedValue(error3);

        const result = await compositeProvider.getTestData('test-key');
        expect(result).toBeNull();
      });
    });

    describe('setTestData', () => {
      it('should set data on first provider that succeeds', async () => {
        mockProvider1.setTestData.mockRejectedValue(new Error('Provider 1 error'));
        mockProvider2.setTestData.mockResolvedValue(undefined);
        mockProvider3.setTestData.mockResolvedValue(undefined);

        await expect(compositeProvider.setTestData('test-key', 'test-value')).resolves.not.toThrow();
        expect(mockProvider1.setTestData).toHaveBeenCalledWith('test-key', 'test-value');
        expect(mockProvider2.setTestData).toHaveBeenCalledWith('test-key', 'test-value');
        expect(mockProvider3.setTestData).not.toHaveBeenCalled();
      });

      it('should throw error if all providers fail', async () => {
        mockProvider1.setTestData.mockRejectedValue(new Error('Provider 1 error'));
        mockProvider2.setTestData.mockRejectedValue(new Error('Provider 2 error'));
        mockProvider3.setTestData.mockRejectedValue(new Error('Provider 3 error'));

        await expect(compositeProvider.setTestData('test-key', 'test-value')).rejects.toThrow(
          '所有提供者设置测试数据失败: Provider 1 error, Provider 2 error, Provider 3 error'
        );
      });
    });

    describe('cleanupTestData', () => {
      it('should cleanup all providers', async () => {
        await compositeProvider.cleanupTestData();

        expect(mockProvider1.cleanupTestData).toHaveBeenCalled();
        expect(mockProvider2.cleanupTestData).toHaveBeenCalled();
        expect(mockProvider3.cleanupTestData).toHaveBeenCalled();
      });

      it('should handle partial cleanup failures', async () => {
        mockProvider1.cleanupTestData.mockResolvedValue(undefined);
        mockProvider2.cleanupTestData.mockRejectedValue(new Error('Provider 2 error'));
        mockProvider3.cleanupTestData.mockResolvedValue(undefined);

        await expect(compositeProvider.cleanupTestData()).resolves.not.toThrow();
      });
    });
  });

  describe('CachedDataProvider', () => {
    let mockProvider: any;
    let cachedProvider: CachedDataProvider;

    beforeEach(() => {
      mockProvider = {
        getTestData: jest.fn(),
        setTestData: jest.fn(),
        cleanupTestData: jest.fn()
      };

      cachedProvider = new CachedDataProvider(mockProvider, 1000); // 1 second TTL
    });

    afterEach(() => {
      jest.clearAllMocks();
    });

    describe('getTestData', () => {
      it('should cache data from underlying provider', async () => {
        const testData = { key: 'value' };
        mockProvider.getTestData.mockResolvedValue(testData);

        const result1 = await cachedProvider.getTestData('test-key');
        const result2 = await cachedProvider.getTestData('test-key');

        expect(result1).toEqual(testData);
        expect(result2).toEqual(testData);
        expect(mockProvider.getTestData).toHaveBeenCalledTimes(1); // Only called once due to cache
      });

      it('should bypass cache if TTL expired', async () => {
        const testData = { key: 'value' };
        mockProvider.getTestData.mockResolvedValue(testData);

        // Manually expire cache
        jest.spyOn(Date, 'now')
          .mockReturnValueOnce(1000)
          .mockReturnValueOnce(3000); // 2 seconds later (TTL is 1 second)

        await cachedProvider.getTestData('test-key');
        await cachedProvider.getTestData('test-key');

        expect(mockProvider.getTestData).toHaveBeenCalledTimes(2);
      });

      it('should not cache null or undefined values', async () => {
        mockProvider.getTestData.mockResolvedValue(null);

        await cachedProvider.getTestData('test-key');
        await cachedProvider.getTestData('test-key');

        expect(mockProvider.getTestData).toHaveBeenCalledTimes(2); // Called twice because null is not cached
      });
    });

    describe('setTestData', () => {
      it('should set data on underlying provider and update cache', async () => {
        const testData = { key: 'value' };
        mockProvider.setTestData.mockResolvedValue(undefined);

        await cachedProvider.setTestData('test-key', testData);

        expect(mockProvider.setTestData).toHaveBeenCalledWith('test-key', testData);

        // Verify cache is updated
        mockProvider.getTestData.mockResolvedValue('different-value');
        const result = await cachedProvider.getTestData('test-key');
        expect(result).toEqual(testData); // Should return cached value
        expect(mockProvider.getTestData).not.toHaveBeenCalled();
      });
    });

    describe('cleanupTestData', () => {
      it('should cleanup underlying provider and clear cache', async () => {
        await cachedProvider.setTestData('test-key', 'test-value');
        expect(cachedProvider.getCacheStats().size).toBe(1);

        await cachedProvider.cleanupTestData();

        expect(mockProvider.cleanupTestData).toHaveBeenCalled();
        expect(cachedProvider.getCacheStats().size).toBe(0);
      });
    });

    describe('cache management', () => {
      it('should clean up expired cache entries', async () => {
        // Manually set expired cache entries
        jest.spyOn(Date, 'now').mockReturnValue(1000);
        (cachedProvider as any).cache.set('expired-key', {
          data: 'expired-value',
          timestamp: 500 // 500ms ago (TTL is 1000ms)
        });

        expect(cachedProvider.getCacheStats().size).toBe(1);

        jest.spyOn(Date, 'now').mockReturnValue(2000); // Now is 2000, timestamp was 500
        const cleanedCount = cachedProvider.cleanupExpiredCache();

        expect(cleanedCount).toBe(1);
        expect(cachedProvider.getCacheStats().size).toBe(0);
      });

      it('should return cache statistics', async () => {
        const stats = cachedProvider.getCacheStats();
        expect(stats).toHaveProperty('size');
        expect(stats).toHaveProperty('hitCount');
        expect(stats).toHaveProperty('missCount');
        expect(typeof stats.size).toBe('number');
      });
    });
  });
});