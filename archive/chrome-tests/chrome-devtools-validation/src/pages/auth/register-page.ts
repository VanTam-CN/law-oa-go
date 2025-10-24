/**
 * 用户注册页面Page Object
 */

import { BasePageObject, PageObjectConfig } from '../../core/base-page-object';
import { Logger } from '../../core/logger';

export interface RegistrationData {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  firstName: string;
  lastName: string;
  phone?: string;
  department?: string;
  role?: string;
}

export class RegisterPage extends BasePageObject {
    protected override selectors = {
    registerForm: '#register-form',
    usernameInput: '#username',
    emailInput: '#email',
    passwordInput: '#password',
    confirmPasswordInput: '#confirm-password',
    firstNameInput: '#first-name',
    lastNameInput: '#last-name',
    phoneInput: '#phone',
    departmentSelect: '#department',
    roleSelect: '#role',
    registerButton: '#register-button',
    cancelButton: '#cancel-button',
    errorMessage: '.error-message',
    successMessage: '.success-message',
    loadingSpinner: '.loading-spinner',
    validationErrors: '.validation-error',
    termsCheckbox: '#terms-checkbox',
    loginLink: '#login-link'
  };

  constructor(config: PageObjectConfig, logger?: Logger) {
    super(config, {
      registerForm: '#register-form',
      usernameInput: '#username',
      emailInput: '#email',
      passwordInput: '#password',
      confirmPasswordInput: '#confirm-password',
      firstNameInput: '#first-name',
      lastNameInput: '#last-name',
      phoneInput: '#phone',
      departmentSelect: '#department',
      registerButton: '#register-button',
      cancelButton: '#cancel-button',
      successMessage: '.success-message',
      errorMessage: '.error-message',
      passwordStrength: '.password-strength',
      loginLink: '.login-link'
    }, logger);
  }

  /**
   * 导航到注册页面
   */
  override async navigateToRegister(): Promise<void> {
    await this.navigate('/register');
    await this.waitForElement(this.selectors.registerForm);
  }

  /**
   * 填写注册表单
   */
  override async fillRegistrationForm(data: RegistrationData): Promise<void> {
    await this.waitForElement(this.selectors.usernameInput);
    await this.fill(this.selectors.usernameInput, data.username);

    await this.waitForElement(this.selectors.emailInput);
    await this.fill(this.selectors.emailInput, data.email);

    await this.waitForElement(this.selectors.passwordInput);
    await this.fill(this.selectors.passwordInput, data.password);

    await this.waitForElement(this.selectors.confirmPasswordInput);
    await this.fill(this.selectors.confirmPasswordInput, data.confirmPassword);

    await this.waitForElement(this.selectors.firstNameInput);
    await this.fill(this.selectors.firstNameInput, data.firstName);

    await this.waitForElement(this.selectors.lastNameInput);
    await this.fill(this.selectors.lastNameInput, data.lastName);

    if (data.phone) {
      await this.waitForElement(this.selectors.phoneInput);
      await this.fill(this.selectors.phoneInput, data.phone);
    }

    if (data.department) {
      await this.waitForElement(this.selectors.departmentSelect);
      await this.select(this.selectors.departmentSelect, data.department);
    }

    if (data.role) {
      await this.waitForElement(this.selectors.roleSelect);
      await this.select(this.selectors.roleSelect, data.role);
    }

    this.logger.debug('注册表单已填写', { username: data.username, email: data.email });
  }

  /**
   * 同意条款
   */
  override async acceptTerms(): Promise<void> {
    await this.waitForElement(this.selectors.termsCheckbox);
    await this.click(this.selectors.termsCheckbox);
    this.logger.debug('条款已接受');
  }

  /**
   * 点击注册按钮
   */
  override async clickRegisterButton(): Promise<void> {
    await this.waitForElement(this.selectors.registerButton);
    await this.click(this.selectors.registerButton);
    this.logger.debug('注册按钮已点击');
  }

  /**
   * 执行完整注册流程
   */
  override async register(data: RegistrationData): Promise<void> {
    this.logger.info('开始注册流程', { username: data.username, email: data.email });

    await this.navigateToRegister();
    await this.fillRegistrationForm(data);
    await this.acceptTerms();
    await this.clickRegisterButton();

    // 等待注册完成
    await this.waitForRegistrationComplete();

    this.logger.info('注册流程完成', { username: data.username });
  }

  /**
   * 等待注册完成
   */
  override async waitForRegistrationComplete(): Promise<void> {
    try {
      // 检查是否显示成功消息或跳转到登录页面
      await this.waitForEither([
        { selector: this.selectors.successMessage, timeout: 10000 },
        { selector: this.selectors.loginForm, timeout: 10000 }
      ]);
    } catch (error) {
      // 如果出现错误消息，抛出异常
      if (await this.isVisible(this.selectors.errorMessage)) {
        const errorMessage = await this.getText(this.selectors.errorMessage);
        throw new Error(`注册失败: ${errorMessage}`);
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
   * 获取验证错误
   */
  override async getValidationErrors(): Promise<Array<{ field: string; message: string }>> {
    const errors: Array<{ field: string; message: string }> = [];

    if (await this.isVisible(this.selectors.validationErrors)) {
      // 在实际实现中，这里会解析具体的验证错误
      errors.push({ field: 'general', message: '表单验证失败' });
    }

    return errors;
  }

  /**
   * 检查注册表单是否可见
   */
  override async isRegisterFormVisible(): Promise<boolean> {
    return await this.isVisible(this.selectors.registerForm);
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
   * 验证注册页面元素
   */
  override async validateRegisterPage(): Promise<{ valid: boolean; missingElements: string[] }> {
    const requiredElements = [
      { name: 'registerForm', selector: this.selectors.registerForm },
      { name: 'usernameInput', selector: this.selectors.usernameInput },
      { name: 'emailInput', selector: this.selectors.emailInput },
      { name: 'passwordInput', selector: this.selectors.passwordInput },
      { name: 'confirmPasswordInput', selector: this.selectors.confirmPasswordInput },
      { name: 'firstNameInput', selector: this.selectors.firstNameInput },
      { name: 'lastNameInput', selector: this.selectors.lastNameInput },
      { name: 'registerButton', selector: this.selectors.registerButton }
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
      this.selectors.usernameInput,
      this.selectors.emailInput,
      this.selectors.passwordInput,
      this.selectors.confirmPasswordInput,
      this.selectors.firstNameInput,
      this.selectors.lastNameInput,
      this.selectors.phoneInput
    ];

    for (const field of fields) {
      if (await this.isVisible(field)) {
        await this.fill(field, '');
      }
    }

    this.logger.debug('注册表单已清除');
  }

  /**
   * 检查密码强度
   */
  override async checkPasswordStrength(password: string): Promise<{
    score: number;
    feedback: string[];
    isStrong: boolean;
  }> {
    // 模拟密码强度检查
    const feedback: string[] | undefined = undefined;
    let score = 0;

    if (password.length >= 8) score += 1;
    else feedback.push('密码长度至少8位');

    if (/[a-z]/.test(password)) score += 1;
    else feedback.push('需要包含小写字母');

    if (/[A-Z]/.test(password)) score += 1;
    else feedback.push('需要包含大写字母');

    if (/[0-9]/.test(password)) score += 1;
    else feedback.push('需要包含数字');

    if (/[^A-Za-z0-9]/.test(password)) score += 1;
    else feedback.push('需要包含特殊字符');

    return {
      score,
      feedback,
      isStrong: score >= 4
    };
  }

  /**
   * 检查用户名可用性
   */
  override async checkUsernameAvailability(username: string): Promise<{ available: boolean; message: string }> {
    // 模拟用户名可用性检查
    await this.wait(500); // 模拟网络延迟

    const unavailableUsernames = ['admin', 'root', 'test', 'user'];
    const available = !unavailableUsernames.includes(username.toLowerCase());

    return {
      available,
      message: available ? '用户名可用' : '用户名已被使用'
    };
  }

  /**
   * 检查邮箱可用性
   */
  override async checkEmailAvailability(email: string): Promise<{ available: boolean; message: string }> {
    // 模拟邮箱可用性检查
    await this.wait(500); // 模拟网络延迟

    const unavailableEmails = ['admin@example.com', 'test@example.com'];
    const available = !unavailableEmails.includes(email.toLowerCase());

    return {
      available,
      message: available ? '邮箱可用' : '邮箱已被注册'
    };
  }

  /**
   * 获取表单状态
   */
  override async getFormState(): Promise<{
    username: string;
    email: string;
    password: string;
    confirmPassword: string;
    firstName: string;
    lastName: string;
    phone: string;
    department: string;
    role: string;
    termsAccepted: boolean;
    submitEnabled: boolean;
  }> {
    const username = await this.getAttribute(this.selectors.usernameInput, 'value') || '';
    const email = await this.getAttribute(this.selectors.emailInput, 'value') || '';
    const password = await this.getAttribute(this.selectors.passwordInput, 'value') || '';
    const confirmPassword = await this.getAttribute(this.selectors.confirmPasswordInput, 'value') || '';
    const firstName = await this.getAttribute(this.selectors.firstNameInput, 'value') || '';
    const lastName = await this.getAttribute(this.selectors.lastNameInput, 'value') || '';
    const phone = await this.getAttribute(this.selectors.phoneInput, 'value') || '';
    const department = await this.getAttribute(this.selectors.departmentSelect, 'value') || '';
    const role = await this.getAttribute(this.selectors.roleSelect, 'value') || '';
    const termsAccepted = await this.getAttribute(this.selectors.termsCheckbox, 'checked') === 'checked';
    const submitEnabled = !(await this.getAttribute(this.selectors.registerButton, 'disabled'));

    return {
      username: username || '',
      email: email || '',
      password: password || '',
      confirmPassword: confirmPassword || '',
      firstName: firstName || '',
      lastName: lastName || '',
      phone: phone || '',
      department: department || '',
      role: role || '',
      termsAccepted,
      submitEnabled
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