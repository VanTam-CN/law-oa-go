/**
 * Page Object基类 - 提供通用的页面操作方法
 */

import { Logger } from './logger';
import { Assertion } from '../types/test-types';

export interface PageObjectConfig {
  baseUrl: string;
  defaultTimeout: number;
  screenshotOnFailure: boolean;
}

export interface SelectorMap {
  [key: string]: string;
}

export interface WaitOptions {
  timeout?: number;
  pollInterval?: number;
}

export class BasePageObject {
  protected logger: Logger;
  protected config: PageObjectConfig;
  protected selectors: SelectorMap;

  constructor(config: PageObjectConfig, selectors: SelectorMap = {}, logger?: Logger) {
    this.config = config;
    this.selectors = selectors;
    this.logger = logger || new Logger('BasePageObject');
  }

  /**
   * 导航到指定URL
   */
   async navigate(url: string): Promise<void> {
    try {
      this.logger.debug('导航到页面', { url });

      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // await this.mcpService.navigate(url);

      this.logger.info('页面导航完成', { url });
    } catch (error) {
      this.logger.error('页面导航失败', { url, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 获取当前页面URL
   */
   async getCurrentUrl(): Promise<string> {
    try {
      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const url = await this.mcpService.getCurrentUrl();
      const url = `${this.config.baseUrl}/current-page`; // 模拟返回

      this.logger.debug('获取当前URL', { url });
      return url;
    } catch (error) {
      this.logger.error('获取当前URL失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 等待元素出现
   */
   async waitForElement(selector: string, options?: WaitOptions): Promise<void> {
    const timeout = options?.timeout || this.config.defaultTimeout;
    const pollInterval = options?.pollInterval || 100;

    try {
      this.logger.debug('等待元素出现', { selector, timeout });

      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // await this.mcpService.waitForElement(selector, { timeout, pollInterval });

      this.logger.debug('元素已出现', { selector });
    } catch (error) {
      this.logger.error('等待元素超时', { selector, timeout, error: error instanceof Error ? error.message : error });
      throw new Error(`等待元素 ${selector} 超时`);
    }
  }

  /**
   * 检查元素是否可见
   */
   async isVisible(selector: string): Promise<boolean> {
    try {
      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const visible = await this.mcpService.isVisible(selector);
      const visible = true; // 模拟返回

      this.logger.debug('检查元素可见性', { selector, visible });
      return visible;
    } catch (error) {
      this.logger.error('检查元素可见性失败', { selector, error: error instanceof Error ? error.message : error });
      return false;
    }
  }

  /**
   * 检查元素是否存在
   */
   async isExists(selector: string): Promise<boolean> {
    try {
      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const exists = await this.mcpService.isExists(selector);
      const exists = true; // 模拟返回

      this.logger.debug('检查元素存在性', { selector, exists });
      return exists;
    } catch (error) {
      this.logger.error('检查元素存在性失败', { selector, error: error instanceof Error ? error.message : error });
      return false;
    }
  }

  /**
   * 获取元素文本内容
   */
   async getText(selector: string): Promise<string> {
    try {
      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const text = await this.mcpService.getText(selector);
      const text = '模拟文本内容'; // 模拟返回

      this.logger.debug('获取元素文本', { selector, text });
      return text;
    } catch (error) {
      this.logger.error('获取元素文本失败', { selector, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 获取元素属性值
   */
   async getAttribute(selector: string, attributeName: string): Promise<string | null> {
    try {
      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const value = await this.mcpService.getAttribute(selector, attributeName);
      const value = '模拟属性值'; // 模拟返回

      this.logger.debug('获取元素属性', { selector, attributeName, value });
      return value;
    } catch (error) {
      this.logger.error('获取元素属性失败', { selector, attributeName, error: error instanceof Error ? error.message : error });
      return null;
    }
  }

  /**
   * 输入文本到输入框
   */
   async fill(selector: string, value: string): Promise<void> {
    try {
      this.logger.debug('输入文本', { selector, value });

      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // await this.mcpService.fill(selector, value);

      this.logger.debug('文本输入完成', { selector });
    } catch (error) {
      this.logger.error('文本输入失败', { selector, value, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 点击元素
   */
   async click(selector: string): Promise<void> {
    try {
      this.logger.debug('点击元素', { selector });

      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // await this.mcpService.click(selector);

      this.logger.debug('元素点击完成', { selector });
    } catch (error) {
      this.logger.error('元素点击失败', { selector, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 选择下拉选项
   */
   async select(selector: string, value: string): Promise<void> {
    try {
      this.logger.debug('选择下拉选项', { selector, value });

      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // await this.mcpService.select(selector, value);

      this.logger.debug('下拉选项选择完成', { selector });
    } catch (error) {
      this.logger.error('下拉选项选择失败', { selector, value, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 等待指定时间
   */
   async wait(milliseconds: number): Promise<void> {
    try {
      this.logger.debug('等待', { milliseconds });
      await new Promise(resolve => setTimeout(resolve, milliseconds));
    } catch (error) {
      this.logger.error('等待失败', { milliseconds, error: error instanceof Error ? error.message : error });
    }
  }

  /**
   * 截取屏幕截图
   */
   async takeScreenshot(filename?: string): Promise<string> {
    try {
      this.logger.debug('截取屏幕截图', { filename });

      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const screenshotPath = await this.mcpService.takeScreenshot(filename);
      const screenshotPath = `screenshots/${filename || 'screenshot'}-${Date.now()}.png`;

      this.logger.debug('屏幕截图完成', { path: screenshotPath });
      return screenshotPath;
    } catch (error) {
      this.logger.error('屏幕截图失败', { filename, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 执行JavaScript脚本
   */
   async executeScript(script: string): Promise<any> {
    try {
      this.logger.debug('执行JavaScript脚本', { script: script.substring(0, 100) + '...' });

      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const result = await this.mcpService.executeScript(script);
      const result = { success: true }; // 模拟返回

      this.logger.debug('JavaScript脚本执行完成', { result });
      return result;
    } catch (error) {
      this.logger.error('JavaScript脚本执行失败', { script: script.substring(0, 100) + '...', error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 获取页面标题
   */
   async getPageTitle(): Promise<string> {
    try {
      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const title = await this.mcpService.getPageTitle();
      const title = '模拟页面标题'; // 模拟返回

      this.logger.debug('获取页面标题', { title });
      return title;
    } catch (error) {
      this.logger.error('获取页面标题失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 检查页面是否包含指定文本
   */
   async pageContainsText(text: string): Promise<boolean> {
    try {
      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const contains = await this.mcpService.pageContainsText(text);
      const contains = true; // 模拟返回

      this.logger.debug('检查页面文本', { text, contains });
      return contains;
    } catch (error) {
      this.logger.error('检查页面文本失败', { text, error: error instanceof Error ? error.message : error });
      return false;
    }
  }

  /**
   * 验证断言
   */
   async verifyAssertion(assertion: Assertion): Promise<boolean> {
    try {
      this.logger.debug('验证断言', { assertion });

      switch (assertion.type) {
        case 'element-exists':
          return await this.isExists(assertion.selector);

        case 'element-visible':
          return await this.isVisible(assertion.selector);

        case 'element-enabled':
          // 需要通过属性检查
          const disabled = await this.getAttribute(assertion.selector, 'disabled');
          return disabled !== 'disabled' && disabled !== 'true';

        case 'text-contains':
          const text = await this.getText(assertion.selector);
          return text.includes(assertion.expected);

        case 'value-equals':
          const value = await this.getAttribute(assertion.selector, 'value');
          return value === assertion.expected;

        case 'url-contains':
          const url = await this.getCurrentUrl();
          return url.includes(assertion.expected);

        default:
          this.logger.warn('未知的断言类型', { type: assertion.type });
          return false;
      }
    } catch (error) {
      this.logger.error('断言验证失败', { assertion, error: error instanceof Error ? error.message : error });
      return false;
    }
  }

  /**
   * 获取页面快照用于调试
   */
   async takeSnapshot(): Promise<any> {
    try {
      this.logger.debug('获取页面快照');

      // 在实际实现中，这里会调用Chrome DevTools MCP服务
      // const snapshot = await this.mcpService.takeSnapshot();
      const snapshot = {
        url: await this.getCurrentUrl(),
        title: await this.getPageTitle(),
        elements: [],
        timestamp: new Date()
      };

      this.logger.debug('页面快照获取完成');
      return snapshot;
    } catch (error) {
      this.logger.error('获取页面快照失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 批量验证多个断言
   */
   async verifyAssertions(assertions: any[]): Promise<{ passed: any[]; failed: any[] }> {
    let passed: any[] = [];
    let failed: any[] = [];

    for (const assertion of assertions) {
      try {
        const result = await this.verifyAssertion(assertion);
        if (result && (result.length || 0) > 0) {
          passed.push(assertion);
        } else {
          failed.push(assertion);
        }
      } catch (error) {
        this.logger.error('断言验证异常', { assertion, error: error instanceof Error ? error.message : error });
        failed.push(assertion);
      }
    }

    this.logger.info('批量断言验证完成', {
      total: assertions.length,
      passed: passed.length,
      failed: failed.length
    });

    return { passed, failed };
  }

  /**
   * 安全地执行操作，失败时自动截图
   */
  async safeExecute<T>(
    operation: () => Promise<T>,
    operationName: string,
    screenshotOnError: boolean = true
  ): Promise<T> {
    try {
      this.logger.debug('开始执行操作', { operationName });
      const result = await operation();
      this.logger.debug('操作执行成功', { operationName });
      return result;
    } catch (error) {
      this.logger.error('操作执行失败', { operationName, error: error instanceof Error ? error.message : error });

      if (screenshotOnError && this.config.screenshotOnFailure) {
        try {
          const screenshotPath = await this.takeScreenshot(`error-${operationName}-${Date.now()}`);
          this.logger.info('错误截图已保存', { path: screenshotPath });
        } catch (screenshotError) {
          this.logger.error('错误截图失败', { screenshotError: screenshotError instanceof Error ? screenshotError.message : screenshotError });
        }
      }

      throw error;
    }
  }

  /**
   * 获取选择器映射
   */
  getSelectors(): SelectorMap {
    return { ...this.selectors };
  }

  /**
   * 添加新的选择器
   */
  addSelector(name: string, selector: string): void {
    this.selectors[name] = selector;
    this.logger.debug('添加选择器', { name, selector });
  }

  /**
   * 获取配置信息
   */
  getConfig(): PageObjectConfig {
    return { ...this.config };
  }
}