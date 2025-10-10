/**
 * 测试数据提供者实现
 */

import { TestDataProvider } from '../types/test-engine-types';
import { Logger } from '../core/logger';

/**
 * 内存数据提供者
 */
export class MemoryDataProvider implements TestDataProvider {
  private logger: Logger;
  private data: Map<string, any> = new Map();

  constructor(logger?: Logger) {
    this.logger = logger || new Logger('MemoryDataProvider');
  }

  async getTestData(dataKey: string): Promise<any> {
    try {
      const data = this.data.get(dataKey);
      this.logger.debug('获取测试数据', { dataKey, found: data !== undefined });
      return data;
    } catch (error) {
      this.logger.error('获取测试数据失败', {
        dataKey,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  async setTestData(dataKey: string, data: any): Promise<void> {
    try {
      this.data.set(dataKey, data);
      this.logger.debug('设置测试数据', { dataKey });
    } catch (error) {
      this.logger.error('设置测试数据失败', {
        dataKey,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  async clearTestData(): Promise<void> {
    try {
      const size = this.data.size;
      this.data.clear();
      this.logger.info('清理测试数据完成', { cleanedCount: size });
    } catch (error) {
      this.logger.error('清理测试数据失败', {
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  /**
   * 获取所有数据键
   */
  getDataKeys(): string[] {
    return Array.from(this.data.keys());
  }

  /**
   * 检查数据是否存在
   */
  hasData(dataKey: string): boolean {
    return this.data.has(dataKey);
  }

  /**
   * 获取数据大小
   */
  getDataSize(): number {
    return this.data.size;
  }
}

/**
 * 文件数据提供者
 */
export class FileDataProvider implements TestDataProvider {
  private logger: Logger;
  private basePath: string;

  constructor(basePath: string, logger?: Logger) {
    this.basePath = basePath;
    this.logger = logger || new Logger('FileDataProvider');
  }

  async getTestData(dataKey: string): Promise<any> {
    try {
      // 在实际实现中，这里会从文件系统读取数据
      this.logger.debug('从文件获取测试数据', { dataKey, path: this.getFilePath(dataKey) });

      // 模拟文件读取
      const mockData = await this.readMockFile(dataKey);
      return mockData;
    } catch (error) {
      this.logger.error('从文件获取测试数据失败', {
        dataKey,
        path: this.getFilePath(dataKey),
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  async setTestData(dataKey: string, data: any): Promise<void> {
    try {
      // 在实际实现中，这里会写入文件系统
      const filePath = this.getFilePath(dataKey);
      this.logger.debug('保存测试数据到文件', { dataKey, filePath });

      // 模拟文件写入
      await this.writeMockFile(dataKey, data);
    } catch (error) {
      this.logger.error('保存测试数据到文件失败', {
        dataKey,
        path: this.getFilePath(dataKey),
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  async (): Promise<void> {
    try {
      // 在实际实现中，这里会清理文件系统中的测试数据
      this.logger.info('清理文件测试数据完成', { basePath: this.basePath });
    } catch (error) {
      this.logger.error('清理文件测试数据失败', {
        basePath: this.basePath,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  /**
   * 获取文件路径
   */
  private getFilePath(dataKey: string): string {
    // 替换不安全的字符
    const safeKey = dataKey.replace(/[^a-zA-Z0-9-_]/g, '_');
    return `${this.basePath}/${safeKey}.json`;
  }

  /**
   * 模拟文件读取
   */
  private async readMockFile(dataKey: string): Promise<any> {
    // 模拟文件读取延迟
    await new Promise(resolve => setTimeout(resolve, 10));

    // 返回模拟数据
    const mockData: Record<string, any> = {
      'user-credentials': {
        username: 'testuser',
        password: 'testpass123',
        email: 'test@example.com'
      },
      'test-case-data': {
        caseTitle: '测试案例',
        priority: 'high',
        assignee: '测试人员'
      },
      'browser-config': {
        viewport: { width: 1920, height: 1080 },
        timeout: 30000,
        slowMo: 0
      }
    };

    return mockData[dataKey] || null;
  }

  /**
   * 模拟文件写入
   */
  private async writeMockFile(dataKey: string, data: any): Promise<void> {
    // 模拟文件写入延迟
    await new Promise(resolve => setTimeout(resolve, 10));
    this.logger.debug('模拟文件写入完成', { dataKey, dataSize: JSON.stringify(data).length });
  }
}

/**
 * 环境变量数据提供者
 */
export class EnvironmentDataProvider implements TestDataProvider {
  private logger: Logger;
  private prefix: string;

  constructor(prefix: string = 'TEST_DATA_', logger?: Logger) {
    this.prefix = prefix;
    this.logger = logger || new Logger('EnvironmentDataProvider');
  }

  async getTestData(dataKey: string): Promise<any> {
    try {
      const envKey = `${this.prefix}${dataKey.toUpperCase()}`;
      const envValue = process.env[envKey];

      this.logger.debug('从环境变量获取测试数据', { dataKey, envKey, found: envValue !== undefined });

      if (envValue === undefined) {
        return null;
      }

      // 尝试解析JSON
      try {
        return JSON.parse(envValue);
      } catch {
        // 如果不是JSON，直接返回字符串
        return envValue;
      }
    } catch (error) {
      this.logger.error('从环境变量获取测试数据失败', {
        dataKey,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  async setTestData(dataKey: string, data: any): Promise<void> {
    try {
      const envKey = `${this.prefix}${dataKey.toUpperCase()}`;
      // 在实际实现中，这里会设置环境变量
      // const value = typeof data === 'string' ? data : JSON.stringify(data);
      // process.env[envKey] = value;

      this.logger.debug('设置环境变量测试数据', { dataKey, envKey });
      this.logger.warn('环境变量数据提供者只支持读取，不支持写入');
    } catch (error) {
      this.logger.error('设置环境变量测试数据失败', {
        dataKey,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  async (): Promise<void> {
    try {
      // 环境变量提供者通常不需要清理
      this.logger.info('环境变量测试数据清理完成');
    } catch (error) {
      this.logger.error('清理环境变量测试数据失败', {
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }
}

/**
 * 组合数据提供者
 */
export class CompositeDataProvider implements TestDataProvider {
  private logger: Logger;
  private providers: TestDataProvider[];

  constructor(providers: TestDataProvider[], logger?: Logger) {
    this.providers = providers;
    this.logger = logger || new Logger('CompositeDataProvider');
  }

  async getTestData(dataKey: string): Promise<any> {
    let errors: Error[] = [];

    for (const provider of this.providers) {
      try {
        const data = await provider.getTestData(dataKey);
        if (data !== null && data !== undefined) {
          this.logger.debug('从提供者获取到测试数据', {
            dataKey,
            provider: provider.constructor.name
          });
          return data;
        }
      } catch (error) {
        errors.push(error instanceof Error ? error : new Error(String(error)));
      }
    }

    // 所有提供者都没有找到数据
    this.logger.warn('所有提供者都未找到测试数据', {
      dataKey,
      errors: (errors || []).map(e => e.message)
    });

    return null;
  }

  async setTestData(dataKey: string, data: any): Promise<void> {
    const errors: Error[] | undefined = undefined;

    for (const provider of this.providers) {
      try {
        await provider.setTestData(dataKey, data);
        this.logger.debug('设置测试数据到提供者', {
          dataKey,
          provider: provider.constructor.name
        });
        return; // 成功设置到第一个提供者就返回
      } catch (error) {
        errors.push(error instanceof Error ? error : new Error(String(error)));
      }
    }

    // 所有提供者都失败了
    throw new Error(`所有提供者设置测试数据失败: ${(errors || []).map(e => e.message).join(', ')}`);
  }

  async (): Promise<void> {
    const errors: Error[] | undefined = undefined;

    for (const provider of this.providers) {
      try {
        await provider.cleanupTestData();
      } catch (error) {
        errors.push(error instanceof Error ? error : new Error(String(error)));
      }
    }

    if ((errors || []).length > 0) {
      this.logger.warn('部分提供者清理测试数据失败', {
        errors: (errors || []).map(e => e.message)
      });
    }
  }
}

/**
 * 缓存数据提供者装饰器
 */
export class CachedDataProvider implements TestDataProvider {
  private logger: Logger;
  private provider: TestDataProvider;
  private cache: Map<string, { data: any; timestamp: number }> = new Map();
  private ttl: number; // 缓存时间（毫秒）

  constructor(provider: TestDataProvider, ttl: number = 300000, logger?: Logger) {
    this.provider = provider;
    this.ttl = ttl;
    this.logger = logger || new Logger('CachedDataProvider');
  }

  async getTestData(dataKey: string): Promise<any> {
    try {
      // 检查缓存
      const cached = this.cache.get(dataKey);
      if (cached && Date.now() - cached.timestamp < this.ttl) {
        this.logger.debug('从缓存获取测试数据', { dataKey, age: Date.now() - cached.timestamp });
        return cached.data;
      }

      // 从底层提供者获取数据
      const data = await this.provider.getTestData(dataKey);

      // 更新缓存
      this.cache.set(dataKey, { data, timestamp: Date.now() });

      this.logger.debug('更新测试数据缓存', { dataKey, cacheSize: this.cache.size });
      return data;
    } catch (error) {
      this.logger.error('获取缓存测试数据失败', {
        dataKey,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  async setTestData(dataKey: string, data: any): Promise<void> {
    try {
      // 设置到底层提供者
      await this.provider.setTestData(dataKey, data);

      // 更新缓存
      this.cache.set(dataKey, { data, timestamp: Date.now() });

      this.logger.debug('设置缓存测试数据', { dataKey });
    } catch (error) {
      this.logger.error('设置缓存测试数据失败', {
        dataKey,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  async (): Promise<void> {
    try {
      // 清理底层提供者
      await this.provider.cleanupTestData();

      // 清理缓存
      const size = this.cache.size;
      this.cache.clear();

      this.logger.info('清理缓存测试数据完成', { cleanedCount: size });
    } catch (error) {
      this.logger.error('清理缓存测试数据失败', {
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  /**
   * 清理过期缓存
   */
  cleanupExpiredCache(): number {
    const now = Date.now();
    const expiredKeys: string[] | undefined = undefined;

    const entries = Array.from(this.cache.entries());
    for (const [key, value] of entries) {
      if (now - value.timestamp > this.ttl) {
        expiredKeys.push(key);
      }
    }

    for (const key of expiredKeys) {
      this.cache.delete(key);
    }

    if (expiredKeys.length > 0) {
      this.logger.debug('清理过期缓存', { count: expiredKeys.length });
    }

    return expiredKeys.length;
  }

  /**
   * 获取缓存统计
   */
  getCacheStats(): { size: number; hitCount: number; missCount: number } {
    // 简单的统计实现
    return {
      size: this.cache.size,
      hitCount: 0, // 需要更复杂的实现来统计命中率
      missCount: 0
    };
  }
}