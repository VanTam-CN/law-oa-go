/**
 * 登录页面Page Object
 */

import { BasePageObject, PageObjectConfig } from '../../core/base-page-object';
import { Logger } from '../../core/logger';

export interface LoginCredentials {
  username: string;
  password: string;
}

export interface UserInfo {
  username: string;
  email: string;
  fullName: string;
  role: string;
  department?: string;
}

export class LoginPage extends BasePageObject {
    protected override selectors = {
    loginForm: '#login-form',
    usernameInput: '#username',
    passwordInput: '#password',
    loginButton: '#login-button',
    rememberMe: '#remember-me',
    forgotPasswordLink: '#forgot-password',
    errorMessage: '.error-message',
    successMessage: '.success-message',
    loadingSpinner: '.loading-spinner',
    userMenu: '.user-menu',
    logoutButton: '#logout-button'
  };

  constructor(config: PageObjectConfig, logger?: Logger) {
    super(config, {
      loginForm: '#login-form',
      usernameInput: '#username',
      passwordInput: '#password',
      loginButton: '#login-button',
      rememberMe: '#remember-me',
      forgotPasswordLink: '#forgot-password',
      errorMessage: '.error-message',
      successMessage: '.success-message',
      loadingSpinner: '.loading-spinner',
      userMenu: '.user-menu',
      logoutButton: '#logout-button'
    }, logger);
  }

  /**
   * 导航到登录页面
   */
  override async navigateToLogin(): Promise<void> {
    await this.navigate('/login');
    await this.waitForElement(this.selectors.loginForm);
  }

  /**
   * 填写登录表单
   */
  override async fillLoginForm(credentials: LoginCredentials): Promise<void> {
    await this.waitForElement(this.selectors.usernameInput);
    await this.fill(this.selectors.usernameInput, credentials.username);

    await this.waitForElement(this.selectors.passwordInput);
    await this.fill(this.selectors.passwordInput, credentials.password);

    this.logger.debug('登录表单已填写', { username: credentials.username });
  }

  /**
   * 点击登录按钮
   */
  override async clickLoginButton(): Promise<void> {
    await this.waitForElement(this.selectors.loginButton);
    await this.click(this.selectors.loginButton);
    this.logger.debug('登录按钮已点击');
  }

  /**
   * 执行完整登录流程
   */
  override async login(credentials: LoginCredentials): Promise<void> {
    this.logger.info('开始登录流程', { username: credentials.username });

    await this.navigateToLogin();
    await this.fillLoginForm(credentials);
    await this.clickLoginButton();

    // 等待登录完成
    await this.waitForLoginComplete();

    this.logger.info('登录流程完成', { username: credentials.username });
  }

  /**
   * 等待登录完成
   */
  override async waitForLoginComplete(): Promise<void> {
    try {
      // 检查是否跳转到dashboard或显示成功消息
      await this.waitForEither([
        { selector: '.dashboard', timeout: 10000 },
        { selector: this.selectors.successMessage, timeout: 5000 }
      ]);
    } catch (error) {
      // 如果出现错误消息，抛出异常
      if (await this.isVisible(this.selectors.errorMessage)) {
        const errorMessage = await this.getText(this.selectors.errorMessage);
        throw new Error(`登录失败: ${errorMessage}`);
      }
      throw error;
    }
  }

  /**
   * 获取错误消息
   */
  override async getErrorMessage(): Promise<string> {
    if (await this.isVisible(this.selectors.errorMessage)) {
      return await this.getText(this.selectors.errorMessage);
    }
    return '';
  }

  /**
   * 获取成功消息
   */
  override async getSuccessMessage(): Promise<string> {
    if (await this.isVisible(this.selectors.successMessage)) {
      return await this.getText(this.selectors.successMessage);
    }
    return '';
  }

  /**
   * 检查是否已登录
   */
  override async isLoggedIn(): Promise<boolean> {
    try {
      return await this.isVisible(this.selectors.userMenu);
    } catch {
      return false;
    }
  }

  /**
   * 获取当前用户信息
   */
  override async getCurrentUser(): Promise<UserInfo | null> {
    if (!await this.isLoggedIn()) {
      return null;
    }

    try {
      // 模拟获取用户信息
      return {
        username: 'testuser',
        email: 'test@example.com',
        fullName: '测试用户',
        role: 'attorney',
        department: 'litigation'
      };
    } catch (error) {
      this.logger.error('获取用户信息失败', { error: error instanceof Error ? error.message : error });
      return null;
    }
  }

  /**
   * 设置记住我
   */
  override async setRememberMe(remember: boolean = true): Promise<void> {
    if (await this.isVisible(this.selectors.rememberMe)) {
      const isChecked = await this.getAttribute(this.selectors.rememberMe, 'checked');
      if ((remember && isChecked !== 'checked') || (!remember && isChecked === 'checked')) {
        await this.click(this.selectors.rememberMe);
        this.logger.debug('记住我状态已设置', { remember });
      }
    }
  }

  /**
   * 点击忘记密码链接
   */
  override async clickForgotPassword(): Promise<void> {
    await this.waitForElement(this.selectors.forgotPasswordLink);
    await this.click(this.selectors.forgotPasswordLink);
    this.logger.debug('忘记密码链接已点击');
  }

  /**
   * 检查登录表单是否可见
   */
  override async isLoginFormVisible(): Promise<boolean> {
    return await this.isVisible(this.selectors.loginForm);
  }

  /**
   * 检查是否正在加载
   */
  override async isLoading(): Promise<boolean> {
    return await this.isVisible(this.selectors.loadingSpinner);
  }

  /**
   * 登出
   */
  override async logout(): Promise<void> {
    if (await this.isLoggedIn()) {
      await this.waitForElement(this.selectors.userMenu);
      await this.click(this.selectors.userMenu);

      await this.waitForElement(this.selectors.logoutButton);
      await this.click(this.selectors.logoutButton);

      await this.waitForElement(this.selectors.loginForm);
      this.logger.info('用户已登出');
    }
  }

  /**
   * 验证登录页面元素
   */
  override async validateLoginPage(): Promise<{ valid: boolean; missingElements: string[] }> {
    const requiredElements = [
      { name: 'usernameInput', selector: this.selectors.usernameInput },
      { name: 'passwordInput', selector: this.selectors.passwordInput },
      { name: 'loginButton', selector: this.selectors.loginButton },
      { name: 'loginForm', selector: this.selectors.loginForm }
    ];

    const missingElements: string[] | undefined = undefined;

    for (const element of requiredElements) {
      if (!await this.isExists(element.selector)) {
        missingElements.push(element.name);
      }
    }

    return {
      valid: missingElements.length === 0,
      missingElements
    };
  }

  /**
   * 清除表单
   */
  override async clearForm(): Promise<void> {
    if (await this.isVisible(this.selectors.usernameInput)) {
      await this.fill(this.selectors.usernameInput, '');
    }
    if (await this.isVisible(this.selectors.passwordInput)) {
      await this.fill(this.selectors.passwordInput, '');
    }
    this.logger.debug('登录表单已清除');
  }

  /**
   * 等待任一元素出现
   */
  private override async waitForEither(selectors: Array<{ selector: string; timeout?: number }>): Promise<void> {
    const defaultTimeout = 10000;

    for (const { selector, timeout = defaultTimeout } of selectors) {
      try {
        await this.waitForElement(selector, { timeout });
        return;
      } catch {
        // 继续尝试下一个选择器
      }
    }

    throw new Error('等待元素超时');
  }

  /**
   * 获取表单状态
   */
  override async getFormState(): Promise<{
    username: string;
    password: string;
    rememberMe: boolean;
    submitEnabled: boolean;
  }> {
    const username = await this.getAttribute(this.selectors.usernameInput, 'value') || '';
    const password = await this.getAttribute(this.selectors.passwordInput, 'value') || '';
    const rememberMe = await this.getAttribute(this.selectors.rememberMe, 'checked') === 'checked';
    const submitEnabled = !(await this.getAttribute(this.selectors.loginButton, 'disabled'));

    return {
      username: username || '',
      password: password || '',
      rememberMe,
      submitEnabled
    };
  }
}