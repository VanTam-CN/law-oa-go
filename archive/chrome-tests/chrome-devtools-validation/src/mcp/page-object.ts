import { ChromeDevToolsService } from './devtools-service';
import { DevToolsElement } from '../types';
import { Logger } from '../core/logger';

/**
 * Page Object基类，提供页面元素操作的基础功能
 */
export abstract class PageObject {
  protected service: ChromeDevToolsService;
  protected logger: Logger;
  protected url?: string;

  constructor(service: ChromeDevToolsService, logger?: Logger) {
    this.service = service;
    this.logger = logger || new Logger(this.constructor.name);
  }

  /**
   * 导航到页面
   */
  override async navigate(): Promise<void> {
    if (!this.url) {
      throw new Error(`${this.constructor.name} 未定义URL`);
    }

    this.logger.info(`导航到${this.constructor.name}`, { url: this.url });
    await this.service.navigate(this.url);
    await this.waitForPageLoad();
  }

  /**
   * 等待页面加载完成
   */
  protected override async waitForPageLoad(): Promise<void> {
    this.logger.debug('等待页面加载完成');
    await this.service.wait({
      condition: 'navigation',
      timeout: 30000,
    });
  }

  /**
   * 获取元素
   */
  protected override async getElement(selector: string): Promise<DevToolsElement | null> {
    return await this.service.getElement(selector);
  }

  /**
   * 等待元素可见
   */
  protected override async waitForElement(selector: string, timeout = 10000): Promise<DevToolsElement> {
    this.logger.debug('等待元素可见', { selector, timeout });

    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      const element = await this.getElement(selector);
      if (element && element.visible) {
        return element;
      }
      await this.delay(500);
    }

    throw new Error(`元素 ${selector} 在 ${timeout}ms 内未变为可见`);
  }

  /**
   * 等待元素消失
   */
  protected override async waitForElementHidden(selector: string, timeout = 10000): Promise<void> {
    this.logger.debug('等待元素消失', { selector, timeout });

    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      const element = await this.getElement(selector);
      if (!element || !element.visible) {
        return;
      }
      await this.delay(500);
    }

    throw new Error(`元素 ${selector} 在 ${timeout}ms 内未消失`);
  }

  /**
   * 点击元素
   */
  protected override async click(selector: string, waitForNavigation = false): Promise<void> {
    this.logger.debug('点击元素', { selector });
    await this.service.click(selector);

    if (waitForNavigation) {
      await this.waitForPageLoad();
    }
  }

  /**
   * 填写输入框
   */
  protected override async fill(selector: string, value: string, clear = true): Promise<void> {
    this.logger.debug('填写输入框', { selector, value, clear });
    await this.service.fill(selector, value, { element: { uid: '', tagName: '', attributes: {}, visible: true, enabled: true, x: 0, y: 0, width: 0, height: 0 }, value, clear, delay: 0 });
  }

  /**
   * 选择下拉选项
   */
  protected override async select(selector: string, values: string[]): Promise<void> {
    this.logger.debug('选择下拉选项', { selector, values });
    await this.service.select(selector, values);
  }

  /**
   * 获取元素文本
   */
  protected override async getText(selector: string): Promise<string> {
    this.logger.debug('获取元素文本', { selector });

    const script = `
      const element = document.querySelector('${selector}');
      return element ? element.textContent || element.innerText || '' : '';
    `;

    return await this.service.executeScript(script);
  }

  /**
   * 获取元素属性
   */
  protected override async getAttribute(selector: string, attributeName: string): Promise<string | null> {
    this.logger.debug('获取元素属性', { selector, attributeName });

    const script = `
      const element = document.querySelector('${selector}');
      return element ? element.getAttribute('${attributeName}') : null;
    `;

    return await this.service.executeScript(script);
  }

  /**
   * 检查元素是否可见
   */
  protected override async isVisible(selector: string): Promise<boolean> {
    this.logger.debug('检查元素是否可见', { selector });

    const script = `
      const element = document.querySelector('${selector}');
      if (!element) return false;

      const style = window.getComputedStyle(element);
      return style.display !== 'none' &&
             style.visibility !== 'hidden' &&
             style.opacity !== '0';
    `;

    return await this.service.executeScript(script);
  }

  /**
   * 检查元素是否存在
   */
  protected override async exists(selector: string): Promise<boolean> {
    this.logger.debug('检查元素是否存在', { selector });

    const script = `
      return document.querySelector('${selector}') !== null;
    `;

    return await this.service.executeScript(script);
  }

  /**
   * 获取元素值（主要用于输入框）
   */
  protected override async getValue(selector: string): Promise<string> {
    this.logger.debug('获取元素值', { selector });

    const script = `
      const element = document.querySelector('${selector}');
      return element ? element.value || '' : '';
    `;

    return await this.service.executeScript(script);
  }

  /**
   * 设置元素值（主要用于输入框）
   */
  protected override async setValue(selector: string, value: string): Promise<void> {
    this.logger.debug('设置元素值', { selector, value });

    const script = `
      const element = document.querySelector('${selector}');
      if (element) {
        element.value = '${value.replace(/'/g, "\\'")}';
        element.dispatchEvent(new Event('input', { bubbles: true }));
        element.dispatchEvent(new Event('change', { bubbles: true }));
      }
    `;

    await this.service.executeScript(script);
  }

  /**
   * 截图
   */
  protected override async screenshot(filename?: string): Promise<string> {
    this.logger.debug('截图', { filename });
    return await this.service.screenshot(filename ? { filename } : {});
  }

  /**
   * 执行自定义JavaScript
   */
  protected async executeScript<T = any>(script: string, args?: any[]): Promise<T> {
    this.logger.debug('执行自定义JavaScript', { scriptLength: script.length });
    return await this.service.executeScript(script, args);
  }

  /**
   * 获取页面标题
   */
  override async getTitle(): Promise<string> {
    this.logger.debug('获取页面标题');
    return await this.service.executeScript('return document.title;');
  }

  /**
   * 获取页面URL
   */
  override async getUrl(): Promise<string> {
    this.logger.debug('获取页面URL');
    return await this.service.executeScript('return window.location.href;');
  }

  /**
   * 刷新页面
   */
  override async refresh(): Promise<void> {
    this.logger.info('刷新页面');
    await this.service.executeScript('window.location.reload();');
    await this.waitForPageLoad();
  }

  /**
   * 返回上一页
   */
  override async back(): Promise<void> {
    this.logger.info('返回上一页');
    await this.service.executeScript('window.history.back();');
    await this.waitForPageLoad();
  }

  /**
   * 前进一页
   */
  override async forward(): Promise<void> {
    this.logger.info('前进一页');
    await this.service.executeScript('window.history.forward();');
    await this.waitForPageLoad();
  }

  /**
   * 工具方法：延迟
   */
  protected override async delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * 工具方法：等待条件满足
   */
  protected async waitFor(
    condition: () => Promise<boolean>,
    timeout = 10000,
    interval = 500
  ): Promise<void> {
    this.logger.debug('等待条件满足', { timeout, interval });

    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      if (await condition()) {
        return;
      }
      await this.delay(interval);
    }

    throw new Error(`条件在 ${timeout}ms 内未满足`);
  }

  /**
   * 验证当前页面URL
   */
  protected override async expectUrl(pattern: string | RegExp): Promise<void> {
    const url = await this.getUrl();

    if (typeof pattern === 'string') {
      if (!url.includes(pattern)) {
        throw new Error(`期望URL包含 '${pattern}'，实际URL: '${url}'`);
      }
    } else if (!pattern.test(url)) {
      throw new Error(`期望URL匹配模式 '${pattern}'，实际URL: '${url}'`);
    }

    this.logger.debug('URL验证通过', { url, pattern });
  }

  /**
   * 验证元素包含文本
   */
  protected override async expectTextContains(selector: string, expectedText: string): Promise<void> {
    const actualText = await this.getText(selector);

    if (!actualText.includes(expectedText)) {
      throw new Error(`期望元素 '${selector}' 包含文本 '${expectedText}'，实际文本: '${actualText}'`);
    }

    this.logger.debug('文本验证通过', { selector, expectedText, actualText });
  }

  /**
   * 验证元素可见
   */
  protected override async expectVisible(selector: string): Promise<void> {
    if (!(await this.isVisible(selector))) {
      throw new Error(`期望元素 '${selector}' 可见，但实际不可见`);
    }

    this.logger.debug('可见性验证通过', { selector });
  }

  /**
   * 验证元素隐藏
   */
  protected override async expectHidden(selector: string): Promise<void> {
    if (await this.isVisible(selector)) {
      throw new Error(`期望元素 '${selector}' 隐藏，但实际可见`);
    }

    this.logger.debug('隐藏验证通过', { selector });
  }
}