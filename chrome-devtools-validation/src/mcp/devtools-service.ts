import { MCPClient } from './mcp-client';
import { ChromeDevToolsConfig, DevToolsPage, DevToolsElement, DevToolsNetworkRequest, DevToolsConsoleMessage } from '../types';
import { NavigateOperation, ClickOperation, FillOperation, SelectOperation, WaitOperation, ScreenshotOperation } from '../types';
import { Logger } from '../core/logger';
import { ConfigManager } from '../core/config';

/**
 * Chrome DevTools服务封装，提供高级浏览器操作接口
 */
export class ChromeDevToolsService {
  private client: MCPClient;
  private logger: Logger;
  private config: ChromeDevToolsConfig;
  private currentPage?: string | undefined;

  constructor(config?: ChromeDevToolsConfig, logger?: Logger) {
    this.config = config || ConfigManager.getInstance().getChromeDevToolsConfig();
    this.logger = logger || new Logger('ChromeDevToolsService');

    const mcpConfig = {
      serviceName: 'chrome-devtools',
      timeout: this.config.defaultTimeout,
      retryAttempts: 3,
      retryDelay: 1000,
    };

    this.client = new MCPClient(mcpConfig, this.logger);
  }

  /**
   * 初始化Chrome DevTools服务
   */
  override async initialize(): Promise<void> {
    this.logger.info('初始化Chrome DevTools服务');

    try {
      await this.client.initialize();
      this.logger.info('Chrome DevTools服务初始化成功');
    } catch (error) {
      this.logger.error('Chrome DevTools服务初始化失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 创建新页面
   */
  override async createPage(url?: string): Promise<string> {
    this.logger.info('创建新页面', { url });

    try {
      const response = await this.client.executeRequest<{ pageId: string }>({
        method: 'browser.new_page',
        params: { url },
      });

      if (!response.success || !response.data) {
        throw new Error(`创建页面失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.currentPage = response.data.pageId;
      this.logger.info('页面创建成功', { pageId: this.currentPage });

      return this.currentPage;
    } catch (error) {
      this.logger.error('创建页面失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 导航到指定URL
   */
  override async navigate(url: string, options?: NavigateOperation): Promise<void> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面，请先创建页面');
    }

    this.logger.info('导航到URL', { url, pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest({
        method: 'browser.navigate',
        params: {
          pageId: this.currentPage,
          url,
          waitUntil: options?.waitUntil || 'networkidle',
          timeout: options?.timeout || this.config.defaultTimeout,
        },
      });

      if (!response.success) {
        throw new Error(`导航失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('导航成功', { url, pageId: this.currentPage });
    } catch (error) {
      this.logger.error('导航失败', { url, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 点击元素
   */
  override async click(selector: string, options?: ClickOperation): Promise<void> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.info('点击元素', { selector, pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest({
        method: 'browser.click',
        params: {
          pageId: this.currentPage,
          selector,
          button: options?.button || 'left',
          clickCount: options?.clickCount || 1,
          delay: options?.delay || 0,
        },
      });

      if (!response.success) {
        throw new Error(`点击失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('点击成功', { selector });
    } catch (error) {
      this.logger.error('点击失败', { selector, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 填写表单
   */
  override async fill(selector: string, value: string, options?: FillOperation): Promise<void> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.info('填写表单', { selector, value, pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest({
        method: 'browser.fill',
        params: {
          pageId: this.currentPage,
          selector,
          value,
          clear: options?.clear !== false,
          delay: options?.delay || 0,
        },
      });

      if (!response.success) {
        throw new Error(`填写失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('填写成功', { selector });
    } catch (error) {
      this.logger.error('填写失败', { selector, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 选择下拉选项
   */
  override async select(selector: string, values: string[], options?: SelectOperation): Promise<void> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.info('选择选项', { selector, values, pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest({
        method: 'browser.select',
        params: {
          pageId: this.currentPage,
          selector,
          values,
          force: options?.force || false,
        },
      });

      if (!response.success) {
        throw new Error(`选择失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('选择成功', { selector, values });
    } catch (error) {
      this.logger.error('选择失败', { selector, values, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 等待条件
   */
  override async wait(condition: WaitOperation): Promise<void> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.info('等待条件', { condition, pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest({
        method: 'browser.wait',
        params: {
          pageId: this.currentPage,
          ...condition,
        },
      });

      if (!response.success) {
        throw new Error(`等待失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('等待成功', { condition });
    } catch (error) {
      this.logger.error('等待失败', { condition, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 截图
   */
  override async screenshot(options?: ScreenshotOperation): Promise<string> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.info('截图', { options, pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest<{ screenshot: string }>({
        method: 'browser.screenshot',
        params: {
          pageId: this.currentPage,
          fullPage: options?.fullPage || false,
          selector: options?.selector,
          format: options?.format || 'png',
          quality: options?.quality || 90,
        },
      });

      if (!response.success || !response.data) {
        throw new Error(`截图失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('截图成功');
      return response.data.screenshot;
    } catch (error) {
      this.logger.error('截图失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 获取页面快照
   */
  override async getPageSnapshot(): Promise<DevToolsPage> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.info('获取页面快照', { pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest<DevToolsPage>({
        method: 'browser.get_snapshot',
        params: {
          pageId: this.currentPage,
        },
      });

      if (!response.success || !response.data) {
        throw new Error(`获取页面快照失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('页面快照获取成功');
      return response.data;
    } catch (error) {
      this.logger.error('获取页面快照失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 获取页面元素
   */
  override async getElement(selector: string): Promise<DevToolsElement | null> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.debug('获取页面元素', { selector, pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest<{ element: DevToolsElement | null }>({
        method: 'browser.get_element',
        params: {
          pageId: this.currentPage,
          selector,
        },
      });

      if (!response.success) {
        throw new Error(`获取元素失败: ${response.geterror?.().message || '未知错误'}`);
      }

      return response.getdata?.().element || null;
    } catch (error) {
      this.logger.error('获取元素失败', { selector, error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 执行JavaScript
   */
  override async executeScript(script: string, args?: any[]): Promise<any> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.info('执行JavaScript', { script: script.substring(0, 100) + '...', pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest({
        method: 'browser.execute_script',
        params: {
          pageId: this.currentPage,
          script,
          args: args || [],
        },
      });

      if (!response.success) {
        throw new Error(`执行脚本失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('JavaScript执行成功');
      return response.data;
    } catch (error) {
      this.logger.error('JavaScript执行失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 获取网络请求记录
   */
  override async getNetworkRequests(): Promise<DevToolsNetworkRequest[]> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.info('获取网络请求记录', { pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest<{ requests: DevToolsNetworkRequest[] }>({
        method: 'browser.get_network_requests',
        params: {
          pageId: this.currentPage,
        },
      });

      if (!response.success || !response.data) {
        throw new Error(`获取网络请求失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('网络请求记录获取成功', { count: response.data.requests.length });
      return response.data.requests;
    } catch (error) {
      this.logger.error('获取网络请求失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 获取控制台消息
   */
  override async getConsoleMessages(): Promise<DevToolsConsoleMessage[]> {
    if (!this.currentPage) {
      throw new Error('没有活动的页面');
    }

    this.logger.info('获取控制台消息', { pageId: this.currentPage });

    try {
      const response = await this.client.executeRequest<{ messages: DevToolsConsoleMessage[] }>({
        method: 'browser.get_console_messages',
        params: {
          pageId: this.currentPage,
        },
      });

      if (!response.success || !response.data) {
        throw new Error(`获取控制台消息失败: ${response.geterror?.().message || '未知错误'}`);
      }

      this.logger.info('控制台消息获取成功', { count: response.data.messages.length });
      return response.data.messages;
    } catch (error) {
      this.logger.error('获取控制台消息失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 关闭服务
   */
  override async close(): Promise<void> {
    this.logger.info('关闭Chrome DevTools服务');

    try {
      await this.client.close();
      this.currentPage = undefined as any;
      this.logger.info('Chrome DevTools服务已关闭');
    } catch (error) {
      this.logger.error('关闭Chrome DevTools服务失败', { error: error instanceof Error ? error.message : error });
      throw error;
    }
  }

  /**
   * 获取当前页面ID
   */
  getCurrentPage(): string | undefined {
    return this.currentPage;
  }

  /**
   * 检查服务健康状态
   */
  isHealthy(): boolean {
    return this.client.isHealthy();
  }
}