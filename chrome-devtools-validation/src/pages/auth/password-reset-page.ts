/**
 * 密码重置页面Page Object
 */

import { BasePageObject, PageObjectConfig } from '../../core/base-page-object';
import { Logger } from '../../core/logger';

export interface PasswordResetData {
  email: string;
  newPassword: string;
  confirmPassword: string;
  verificationCode?: string;
}

export class PasswordResetPage extends BasePageObject {
    protected override selectors = {
    resetForm: '#password-reset-form',
    emailInput: '#email',
    sendCodeButton: '#send-code-button',
    verificationCodeInput: '#verification-code',
    newPasswordInput: '#new-password',
    confirmPasswordInput: '#confirm-password',
    resetButton: '#reset-password-button',
    cancelButton: '#cancel-button',
    errorMessage: '.error-message',
    successMessage: '.success-message',
    loadingSpinner: '.loading-spinner',
    loginLink: '#login-link',
    resendCodeLink: '#resend-code-link',
    codeExpiry: '#code-expiry',
    passwordStrength: '.password-strength'
  };

  constructor(config: PageObjectConfig, logger?: Logger) {
    super(config, {
      resetForm: '#reset-form',
      emailInput: '#email',
      sendCodeButton: '#send-code-button',
      verificationCodeInput: '#verification-code',
      newPasswordInput: '#new-password',
      confirmPasswordInput: '#confirm-password',
      resetButton: '#reset-button',
      cancelButton: '#cancel-button',
      backButton: '#back-button',
      successMessage: '.success-message',
      errorMessage: '.error-message',
      passwordStrength: '.password-strength'
    }, logger);
  }

  /**
   * 导航到密码重置页面
   */
  override async navigateToPasswordReset(): Promise<void> {
    await this.navigate('/password-reset');
    await this.waitForElement(this.selectors.resetForm);
  }

  /**
   * 输入邮箱
   */
  override async enterEmail(email: string): Promise<void> {
    await this.waitForElement(this.selectors.emailInput);
    await this.fill(this.selectors.emailInput, email);
    this.logger.debug('邮箱已输入', { email });
  }

  /**
   * 发送验证码
   */
  override async sendVerificationCode(): Promise<void> {
    await this.waitForElement(this.selectors.sendCodeButton);
    await this.click(this.selectors.sendCodeButton);
    this.logger.debug('验证码发送按钮已点击');

    // 等待发送完成
    await this.waitForSendCodeComplete();
  }

  /**
   * 等待验证码发送完成
   */
  override async waitForSendCodeComplete(): Promise<void> {
    try {
      await this.waitForEither([
        { selector: this.selectors.verificationCodeInput, timeout: 10000 },
        { selector: this.selectors.successMessage, timeout: 5000 }
      ]);
    } catch (error) {
      if (await this.isVisible(this.selectors.errorMessage)) {
        const errorMessage = await this.getText(this.selectors.errorMessage);
        throw new Error(`发送验证码失败: ${errorMessage}`);
      }
      throw error;
    }
  }

  /**
   * 输入验证码
   */
  override async enterVerificationCode(code: string): Promise<void> {
    await this.waitForElement(this.selectors.verificationCodeInput);
    await this.fill(this.selectors.verificationCodeInput, code);
    this.logger.debug('验证码已输入', { codeLength: code.length });
  }

  /**
   * 输入新密码
   */
  override async enterNewPassword(password: string): Promise<void> {
    await this.waitForElement(this.selectors.newPasswordInput);
    await this.fill(this.selectors.newPasswordInput, password);
    this.logger.debug('新密码已输入');
  }

  /**
   * 确认新密码
   */
  override async confirmNewPassword(password: string): Promise<void> {
    await this.waitForElement(this.selectors.confirmPasswordInput);
    await this.fill(this.selectors.confirmPasswordInput, password);
    this.logger.debug('密码确认已完成');
  }

  /**
   * 点击重置密码按钮
   */
  override async clickResetButton(): Promise<void> {
    await this.waitForElement(this.selectors.resetButton);
    await this.click(this.selectors.resetButton);
    this.logger.debug('重置密码按钮已点击');
  }

  /**
   * 执行完整密码重置流程
   */
  override async resetPassword(data: PasswordResetData): Promise<void> {
    this.logger.info('开始密码重置流程', { email: data.email });

    await this.navigateToPasswordReset();
    await this.enterEmail(data.email);
    await this.sendVerificationCode();

    if (data.verificationCode) {
      await this.enterVerificationCode(data.verificationCode);
    }

    await this.enterNewPassword(data.newPassword);
    await this.confirmNewPassword(data.confirmPassword);
    await this.clickResetButton();

    // 等待重置完成
    await this.waitForResetComplete();

    this.logger.info('密码重置流程完成', { email: data.email });
  }

  /**
   * 等待密码重置完成
   */
  override async waitForResetComplete(): Promise<void> {
    try {
      await this.waitForEither([
        { selector: this.selectors.successMessage, timeout: 10000 },
        { selector: this.selectors.loginForm, timeout: 10000 }
      ]);
    } catch (error) {
      if (await this.isVisible(this.selectors.errorMessage)) {
        const errorMessage = await this.getText(this.selectors.errorMessage);
        throw new Error(`密码重置失败: ${errorMessage}`);
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
   * 检查是否可以重发验证码
   */
  override async canResendCode(): Promise<boolean> {
    return await this.isVisible(this.selectors.resendCodeLink);
  }

  /**
   * 重发验证码
   */
  override async resendCode(): Promise<void> {
    if (await this.canResendCode()) {
      await this.click(this.selectors.resendCodeLink);
      this.logger.debug('验证码已重发');
    } else {
      throw new Error('无法重发验证码');
    }
  }

  /**
   * 获取验证码过期时间
   */
  override async getCodeExpiry(): Promise<string> {
    if (await this.isVisible(this.selectors.codeExpiry)) {
      return await this.getText(this.selectors.codeExpiry);
    }
    return '';
  }

  /**
   * 检查密码强度
   */
  override async getPasswordStrength(): Promise<{
    score: number;
    text: string;
    color: string;
  }> {
    // 模拟密码强度检查
    const password = await this.getAttribute(this.selectors.newPasswordInput, 'value') || '';

    let score = 0;
    let text = '弱';
    let color = 'red';

    if (password.length >= 8) score += 1;
    if (/[a-z]/.test(password)) score += 1;
    if (/[A-Z]/.test(password)) score += 1;
    if (/[0-9]/.test(password)) score += 1;
    if (/[^A-Za-z0-9]/.test(password)) score += 1;

    if (score >= 4) {
      text = '强';
      color = 'green';
    } else if (score >= 2) {
      text = '中等';
      color = 'orange';
    }

    return { score, text, color };
  }

  /**
   * 检查重置表单是否可见
   */
  override async isResetFormVisible(): Promise<boolean> {
    return await this.isVisible(this.selectors.resetForm);
  }

  /**
   * 检查是否正在加载
   */
  override async isLoading(): Promise<boolean> {
    return await this.isVisible(this.selectors.loadingSpinner);
  }

  /**
   * 点击取消按钮
   */
  override async clickCancelButton(): Promise<void> {
    await this.waitForElement(this.selectors.cancelButton);
    await this.click(this.selectors.cancelButton);
    this.logger.debug('取消按钮已点击');
  }

  /**
   * 点击登录链接
   */
  override async clickLoginLink(): Promise<void> {
    await this.waitForElement(this.selectors.loginLink);
    await this.click(this.selectors.loginLink);
    this.logger.debug('登录链接已点击');
  }

  /**
   * 验证密码重置页面元素
   */
  override async validatePasswordResetPage(): Promise<{ valid: boolean; missingElements: string[] }> {
    const requiredElements = [
      { name: 'resetForm', selector: this.selectors.resetForm },
      { name: 'emailInput', selector: this.selectors.emailInput },
      { name: 'sendCodeButton', selector: this.selectors.sendCodeButton }
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
    const fields = [
      this.selectors.emailInput,
      this.selectors.verificationCodeInput,
      this.selectors.newPasswordInput,
      this.selectors.confirmPasswordInput
    ];

    for (const field of fields) {
      if (await this.isVisible(field)) {
        await this.fill(field, '');
      }
    }

    this.logger.debug('密码重置表单已清除');
  }

  /**
   * 检查验证码格式
   */
  validateVerificationCode(code: string): boolean {
    // 验证码通常是6位数字
    return /^\d{6}$/.test(code);
  }

  /**
   * 检查密码规则
   */
  validatePasswordRules(password: string): {
    isValid: boolean;
    errors: string[];
  } {
    const errors: string[] | undefined = undefined;

    if (password.length < 8) {
      errors.push('密码长度至少8位');
    }

    if (!/[a-z]/.test(password)) {
      errors.push('需要包含小写字母');
    }

    if (!/[A-Z]/.test(password)) {
      errors.push('需要包含大写字母');
    }

    if (!/[0-9]/.test(password)) {
      errors.push('需要包含数字');
    }

    if (!/[^A-Za-z0-9]/.test(password)) {
      errors.push('需要包含特殊字符');
    }

    return {
      isValid: errors.length === 0,
      errors
    };
  }

  /**
   * 获取表单状态
   */
  override async getFormState(): Promise<{
    email: string;
    verificationCode: string;
    newPassword: string;
    confirmPassword: string;
    sendCodeEnabled: boolean;
    resetEnabled: boolean;
  }> {
    const email = await this.getAttribute(this.selectors.emailInput, 'value') || '';
    const verificationCode = await this.getAttribute(this.selectors.verificationCodeInput, 'value') || '';
    const newPassword = await this.getAttribute(this.selectors.newPasswordInput, 'value') || '';
    const confirmPassword = await this.getAttribute(this.selectors.confirmPasswordInput, 'value') || '';
    const sendCodeEnabled = !(await this.getAttribute(this.selectors.sendCodeButton, 'disabled'));
    const resetEnabled = !(await this.getAttribute(this.selectors.resetButton, 'disabled'));

    return {
      email: email || '',
      verificationCode: verificationCode || '',
      newPassword: newPassword || '',
      confirmPassword: confirmPassword || '',
      sendCodeEnabled,
      resetEnabled
    };
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
}