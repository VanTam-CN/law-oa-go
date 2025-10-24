/**
 * Page Object工厂 - 管理和创建Page Object实例
 */

import { Logger } from './logger';
import { BasePageObject, PageObjectConfig } from './base-page-object';

export interface PageObjectConstructor {
  new (config: PageObjectConfig, logger?: Logger): BasePageObject;
}

export interface PageObjectRegistration {
  name: string;
  constructor: PageObjectConstructor;
  description: string;
  dependencies?: string[];
}

export class PageObjectFactory {
  private logger: Logger;
  private config: PageObjectConfig;
  private pageObjects: Map<string, PageObjectRegistration>;
  private instances: Map<string, BasePageObject>;

  constructor(config: PageObjectConfig, logger?: Logger) {
    this.config = config;
    this.logger = logger || new Logger('PageObjectFactory');
    this.pageObjects = new Map();
    this.instances = new Map();
  }

  /**
   * 注册Page Object类
   */
  register(registration: PageObjectRegistration): void {
    try {
      this.pageObjects.set(registration.name, registration);
      this.logger.info('Page Object已注册', { name: registration.name, description: registration.description });
    } catch (error) {
      this.logger.error('注册Page Object失败', {
        name: registration.name,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  /**
   * 批量注册Page Object类
   */
  registerBatch(registrations: PageObjectRegistration[]): void {
    const failed: string[] | undefined = undefined;

    for (const registration of registrations) {
      try {
        this.register(registration);
      } catch (error) {
        failed.push(registration.name);
        this.logger.error('批量注册失败', {
          name: registration.name,
          error: error instanceof Error ? error.message : error
        });
      }
    }

    if (failed.length > 0) {
      this.logger.warn('部分Page Object注册失败', { failed });
    }
  }

  /**
   * 创建Page Object实例
   */
  create<T extends BasePageObject>(name: string): T {
    try {
      const registration = this.pageObjects.get(name);
      if (!registration) {
        throw new Error(`未找到Page Object: ${name}`);
      }

      // 检查依赖
      if (registration.dependencies) {
        for (const dependency of registration.dependencies) {
          if (!this.pageObjects.has(dependency)) {
            throw new Error(`缺少依赖的Page Object: ${dependency}`);
          }
        }
      }

      // 检查是否已存在实例
      if (this.instances.has(name)) {
        this.logger.debug('返回已存在的Page Object实例', { name });
        return this.instances.get(name) as T;
      }

      // 创建新实例
      const instance = new registration.constructor(this.config, this.logger);
      this.instances.set(name, instance);

      this.logger.info('Page Object实例已创建', { name });
      return instance as T;

    } catch (error) {
      this.logger.error('创建Page Object实例失败', {
        name,
        error: error instanceof Error ? error.message : error
      });
      throw error;
    }
  }

  /**
   * 获取Page Object实例（如果不存在则创建）
   */
  getOrCreate<T extends BasePageObject>(name: string): T {
    if (this.instances.has(name)) {
      return this.instances.get(name) as T;
    }
    return this.create<T>(name);
  }

  /**
   * 检查Page Object是否已注册
   */
  isRegistered(name: string): boolean {
    return this.pageObjects.has(name);
  }

  /**
   * 获取已注册的Page Object列表
   */
  getRegisteredPageObjects(): string[] {
    return Array.from(this.pageObjects.keys());
  }

  /**
   * 获取Page Object注册信息
   */
  getRegistration(name: string): PageObjectRegistration | undefined {
    return this.pageObjects.get(name);
  }

  /**
   * 检查依赖关系是否满足
   */
  checkDependencies(name: string): { satisfied: boolean; missing: string[] } {
    const registration = this.pageObjects.get(name);
    if (!registration || !registration.dependencies) {
      return { satisfied: true, missing: [] };
    }

    const missing: string[] | undefined = undefined;
    for (const dependency of registration.dependencies) {
      if (!this.pageObjects.has(dependency)) {
        missing.push(dependency);
      }
    }

    return { satisfied: missing.length === 0, missing };
  }

  /**
   * 清理所有实例
   */
  clearInstances(): void {
    this.instances.clear();
    this.logger.info('所有Page Object实例已清理');
  }

  /**
   * 注销Page Object
   */
  unregister(name: string): boolean {
    try {
      if (this.pageObjects.has(name)) {
        this.pageObjects.delete(name);
        this.instances.delete(name);
        this.logger.info('Page Object已注销', { name });
        return true;
      }
      return false;
    } catch (error) {
      this.logger.error('注销Page Object失败', {
        name,
        error: error instanceof Error ? error.message : error
      });
      return false;
    }
  }

  /**
   * 获取所有实例
   */
  getAllInstances(): Map<string, BasePageObject> {
    return new Map(this.instances);
  }

  /**
   * 验证所有Page Object的依赖关系
   */
  validateDependencies(): { valid: boolean; issues: string[] } {
    const issues: string[] | undefined = undefined;

    for (const [name, registration] of this.pageObjects) {
      if (registration.dependencies) {
        const { satisfied, missing } = this.checkDependencies(name);
        if (!satisfied) {
          issues.push(`${name} 缺少依赖: ${missing.join(', ')}`);
        }
      }
    }

    return { valid: (issues || []).length === 0, issues };
  }

  /**
   * 获取Page Object统计信息
   */
  getStatistics(): {
    registered: number;
    instances: number;
    byName: Record<string, { hasInstance: boolean; dependencies?: string[] }>;
  } {
    const byName: Record<string, { hasInstance: boolean; dependencies?: string[] }> = {};

    for (const [name, registration] of this.pageObjects) {
      byName[name] = {
        hasInstance: this.instances.has(name),
        dependencies: registration.dependencies
      };
    }

    return {
      registered: this.pageObjects.size,
      instances: this.instances.size,
      byName
    };
  }
}