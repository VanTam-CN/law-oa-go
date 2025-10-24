import { PageObject } from '../mcp/page-object';
import { ChromeDevToolsService } from '../mcp/devtools-service';
import { Logger } from '../core/logger';

/**
 * 登录页面Page Object
 */
export class LoginPage extends PageObject {
  protected override selectors = {
    loginForm: '#login-form',
    usernameInput: '#username',
    passwordInput: '#password',
    loginButton: '#login-button',
    errorMessage: '.error-message',
    successMessage: '.success-message',
    forgotPasswordLink: '#forgot-password',
    registerLink: '#register-link',
  };

  constructor(service: ChromeDevToolsService, logger?: Logger) {
    super(service, logger);
    this.url = '/login';
  }

  /**
   * 输入用户名
   */
  override async enterUsername(username: string): Promise<void> {
    await this.fill(this.selectors.usernameInput, username);
    this.logger.debug('输入用户名', { username });
  }

  /**
   * 输入密码
   */
  override async enterPassword(password: string): Promise<void> {
    await this.fill(this.selectors.passwordInput, password);
    this.logger.debug('输入密码', { password: '***' });
  }

  /**
   * 点击登录按钮
   */
  override async clickLogin(): Promise<void> {
    await this.click(this.selectors.loginButton, true); // 等待导航
    this.logger.debug('点击登录按钮');
  }

  /**
   * 执行登录操作
   */
  override async login(username: string, password: string): Promise<void> {
    this.logger.info('执行用户登录', { username });

    await this.navigate();
    await this.waitForPageLoad();

    await this.enterUsername(username);
    await this.enterPassword(password);
    await this.clickLogin();

    // 等待登录完成
    await this.waitForLoginComplete();
  }

  /**
   * 等待登录完成
   */
  private override async waitForLoginComplete(): Promise<void> {
    await this.delay(2000); // 等待登录处理

    // 检查是否登录成功（跳转到仪表板或显示成功消息）
    const currentUrl = await this.getUrl();
    if (currentUrl.includes('/dashboard')) {
      this.logger.info('登录成功，已跳转到仪表板');
      return;
    }

    // 检查是否有错误消息
    if (await this.exists(this.selectors.errorMessage)) {
      const errorMessage = await this.getText(this.selectors.errorMessage);
      throw new Error(`登录失败: ${errorMessage}`);
    }

    this.logger.info('登录完成');
  }

  /**
   * 获取错误消息
   */
  override async getErrorMessage(): Promise<string> {
    if (!(await this.isVisible(this.selectors.errorMessage))) {
      return '';
    }
    return await this.getText(this.selectors.errorMessage);
  }

  /**
   * 获取成功消息
   */
  override async getSuccessMessage(): Promise<string> {
    if (!(await this.isVisible(this.selectors.successMessage))) {
      return '';
    }
    return await this.getText(this.selectors.successMessage);
  }

  /**
   * 检查是否已登录
   */
  override async isLoggedIn(): Promise<boolean> {
    const currentUrl = await this.getUrl();
    return currentUrl.includes('/dashboard') || currentUrl.includes('/home');
  }

  /**
   * 点击忘记密码链接
   */
  override async clickForgotPassword(): Promise<void> {
    await this.click(this.selectors.forgotPasswordLink, true);
    this.logger.debug('点击忘记密码链接');
  }

  /**
   * 点击注册链接
   */
  override async clickRegister(): Promise<void> {
    await this.click(this.selectors.registerLink, true);
    this.logger.debug('点击注册链接');
  }

  /**
   * 验证登录页面元素
   */
  override async validateLoginPage(): Promise<void> {
    this.logger.info('验证登录页面元素');

    await this.expectVisible(this.selectors.usernameInput);
    await this.expectVisible(this.selectors.passwordInput);
    await this.expectVisible(this.selectors.loginButton);
    await this.expectVisible(this.selectors.forgotPasswordLink);
    await this.expectVisible(this.selectors.registerLink);

    this.logger.info('登录页面元素验证通过');
  }

  /**
   * 获取登录按钮状态
   */
  override async isLoginButtonEnabled(): Promise<boolean> {
    const disabled = await this.getAttribute(this.selectors.loginButton, 'disabled');
    return disabled !== 'true' && disabled !== '';
  }

  /**
   * 清空登录表单
   */
  override async clearForm(): Promise<void> {
    await this.setValue(this.selectors.usernameInput, '');
    await this.setValue(this.selectors.passwordInput, '');
    this.logger.debug('清空登录表单');
  }

  /**
   * 按回车键登录
   */
  override async loginWithEnter(username: string, password: string): Promise<void> {
    this.logger.info('使用回车键登录', { username });

    await this.navigate();
    await this.waitForPageLoad();

    await this.enterUsername(username);
    await this.enterPassword(password);

    // 在密码输入框按回车键
    const script = `
      const event = new KeyboardEvent('keypress', {
        key: 'Enter',
        code: 'Enter',
        which: 13,
        keyCode: 13,
        bubbles: true
      });
      document.querySelector('${this.selectors.passwordInput}').dispatchEvent(event);
    `;

    await this.executeScript(script);
    await this.waitForLoginComplete();
  }
}