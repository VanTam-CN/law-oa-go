/**
 * 用户认证测试套件
 */

import { AuthPage } from '../pages/auth/auth-pages';
import { BasePageObject } from '../core/base-page-object';
import { Logger } from '../core/logger';
import { TestExecutionEngine } from '../core/test-execution-engine';
import { TestCase, TestStep } from '../types/test-types';

export interface AuthTestUser {
  username: string;
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  role: string;
  department?: string;
  phone?: string;
}

export interface AuthTestConfig {
  baseUrl: string;
  validUser: AuthTestUser;
  invalidUsers: {
    wrongPassword: AuthTestUser;
    nonExistent: AuthTestUser;
    lockedAccount: AuthTestUser;
  };
  defaultTimeout?: number;
  screenshotOnFailure?: boolean;
}

export class AuthTestSuite {
  private authPage: AuthPage;
  private testEngine: TestExecutionEngine;
  private config: AuthTestConfig;
  private logger: Logger;

  constructor(config: AuthTestConfig, logger?: Logger) {
    this.config = config;
    this.logger = logger || new Logger('AuthTestSuite');

    const baseConfig = {
      baseUrl: config.baseUrl,
      defaultTimeout: config.defaultTimeout || 30000,
      screenshotOnFailure: config.screenshotOnFailure || true
    };

    this.authPage = new AuthPage(baseConfig, this.logger);
    this.testEngine = new TestExecutionEngine(baseConfig, this.logger);
  }

  /**
   * 运行完整的认证测试套件
   */
  override async runFullAuthTestSuite(): Promise<{
    passed: number;
    failed: number;
    total: number;
    results: any[];
    summary: string;
  }> {
    this.logger.info('开始运行完整的认证测试套件');

    const testCases = this.getAllTestCases();
    const results = await this.testEngine.executeTestCases(testCases);

    const passed = results.filter(r => r.status === 'passed').length;
    const failed = results.filter(r => r.status === 'failed').length;
    const total = results.length;

    const summary = `
认证测试套件执行完成：
- 总计测试用例: ${total}
- 通过: ${passed}
- 失败: ${failed}
- 成功率: ${((passed / total) * 100).toFixed(2)}%
    `.trim();

    this.logger.info(summary);

    return {
      passed,
      failed,
      total,
      results,
      summary
    };
  }

  /**
   * 获取所有认证测试用例
   */
  private getAllTestCases(): TestCase[] {
    return [
      this.createLoginPageValidationTestCase(),
      this.createValidLoginTestCase(),
      this.createInvalidPasswordTestCase(),
      this.createNonExistentUserTestCase(),
      this.createLockedAccountTestCase(),
      this.createRememberMeTestCase(),
      this.createPasswordVisibilityTestCase(),
      this.createForgotPasswordTestCase(),
      this.createRegistrationTestCase(),
      this.createLogoutTestCase(),
      this.createSessionTimeoutTestCase(),
      this.createConcurrentLoginTestCase(),
      this.createPasswordRequirementsTestCase(),
      this.createEmailValidationTestCase(),
      this.createRedirectAfterLoginTestCase()
    ];
  }

  /**
   * 创建登录页面验证测试用例
   */
  private createLoginPageValidationTestCase(): TestCase {
    return {
      id: 'AUTH-LP-001',
      name: '登录页面元素验证',
      description: '验证登录页面包含所有必需的元素',
      priority: 'high',
      tags: ['auth', 'login', 'ui-validation'],
      steps: [
        {
          id: 'step-1',
          name: '导航到登录页面',
          action: 'navigate',
          expected: '成功加载登录页面',
          selector: '#login-page'
        },
        {
          id: 'step-2',
          name: '验证页面标题',
          action: 'verify-element',
          expected: '页面标题包含"登录"',
          selector: '#login-page-title'
        },
        {
          id: 'step-3',
          name: '验证用户名输入框',
          action: 'verify-element',
          expected: '用户名输入框存在且可见',
          selector: '#login-username'
        },
        {
          id: 'step-4',
          name: '验证密码输入框',
          action: 'verify-element',
          expected: '密码输入框存在且可见',
          selector: '#login-password'
        },
        {
          id: 'step-5',
          name: '验证登录按钮',
          action: 'verify-element',
          expected: '登录按钮存在且可点击',
          selector: '#login-button'
        },
        {
          id: 'step-6',
          name: '验证记住我复选框',
          action: 'verify-element',
          expected: '记住我复选框存在',
          selector: '#login-remember-me'
        },
        {
          id: 'step-7',
          name: '验证忘记密码链接',
          action: 'verify-element',
          expected: '忘记密码链接存在',
          selector: '#forgot-password-link'
        },
        {
          id: 'step-8',
          name: '验证注册链接',
          action: 'verify-element',
          expected: '注册链接存在',
          selector: '#register-link'
        }
      ]
    };
  }

  /**
   * 创建有效登录测试用例
   */
  private createValidLoginTestCase(): TestCase {
    return {
      id: 'AUTH-LG-001',
      name: '有效用户登录',
      description: '使用有效凭据成功登录系统',
      priority: 'critical',
      tags: ['auth', 'login', 'happy-path'],
      testData: {
        username: this.config.validUser.username,
        password: this.config.validUser.password
      },
      steps: [
        {
          id: 'step-1',
          name: '导航到登录页面',
          action: 'navigate',
          expected: '成功加载登录页面',
          url: `${this.config.baseUrl}/auth/login`
        },
        {
          id: 'step-2',
          name: '输入用户名',
          action: 'fill',
          selector: '#login-username',
          value: this.config.validUser.username,
          expected: '用户名输入成功'
        },
        {
          id: 'step-3',
          name: '输入密码',
          action: 'fill',
          selector: '#login-password',
          value: this.config.validUser.password,
          expected: '密码输入成功'
        },
        {
          id: 'step-4',
          name: '点击登录按钮',
          action: 'click',
          selector: '#login-button',
          expected: '登录成功，跳转到仪表板'
        },
        {
          id: 'step-5',
          name: '验证重定向到仪表板',
          action: 'verify-url',
          expected: 'URL包含/dashboard',
          pattern: '.*dashboard.*'
        },
        {
          id: 'step-6',
          name: '验证用户信息显示',
          action: 'verify-element',
          expected: '用户姓名在导航栏中显示',
          selector: '#user-profile-name'
        },
        {
          id: 'step-7',
          name: '验证登录状态',
          action: 'verify-element',
          expected: '用户登录状态为已登录',
          selector: '#user-login-status'
        }
      ]
    };
  }

  /**
   * 创建无效密码测试用例
   */
  private createInvalidPasswordTestCase(): TestCase {
    return {
      id: 'AUTH-LG-002',
      name: '无效密码登录',
      description: '使用错误密码尝试登录应显示错误消息',
      priority: 'high',
      tags: ['auth', 'login', 'negative'],
      testData: {
        username: this.config.invalidUsers.wrongPassword.username,
        password: 'wrongpassword123'
      },
      steps: [
        {
          id: 'step-1',
          name: '导航到登录页面',
          action: 'navigate',
          expected: '成功加载登录页面',
          url: `${this.config.baseUrl}/auth/login`
        },
        {
          id: 'step-2',
          name: '输入用户名',
          action: 'fill',
          selector: '#login-username',
          value: this.config.invalidUsers.wrongPassword.username,
          expected: '用户名输入成功'
        },
        {
          id: 'step-3',
          name: '输入错误密码',
          action: 'fill',
          selector: '#login-password',
          value: 'wrongpassword123',
          expected: '密码输入成功'
        },
        {
          id: 'step-4',
          name: '点击登录按钮',
          action: 'click',
          selector: '#login-button',
          expected: '显示密码错误消息'
        },
        {
          id: 'step-5',
          name: '验证错误消息',
          action: 'verify-element',
          expected: '错误消息显示"用户名或密码错误"',
          selector: '#login-error-message'
        },
        {
          id: 'step-6',
          name: '验证页面未跳转',
          action: 'verify-url',
          expected: '仍停留在登录页面',
          pattern: '.*login.*'
        }
      ]
    };
  }

  /**
   * 创建不存在用户测试用例
   */
  private createNonExistentUserTestCase(): TestCase {
    return {
      id: 'AUTH-LG-003',
      name: '不存在用户登录',
      description: '使用不存在的用户名尝试登录应显示错误消息',
      priority: 'high',
      tags: ['auth', 'login', 'negative'],
      testData: {
        username: 'nonexistentuser',
        password: 'anypassword123'
      },
      steps: [
        {
          id: 'step-1',
          name: '导航到登录页面',
          action: 'navigate',
          expected: '成功加载登录页面',
          url: `${this.config.baseUrl}/auth/login`
        },
        {
          id: 'step-2',
          name: '输入不存在的用户名',
          action: 'fill',
          selector: '#login-username',
          value: 'nonexistentuser',
          expected: '用户名输入成功'
        },
        {
          id: 'step-3',
          name: '输入密码',
          action: 'fill',
          selector: '#login-password',
          value: 'anypassword123',
          expected: '密码输入成功'
        },
        {
          id: 'step-4',
          name: '点击登录按钮',
          action: 'click',
          selector: '#login-button',
          expected: '显示用户不存在错误消息'
        },
        {
          id: 'step-5',
          name: '验证错误消息',
          action: 'verify-element',
          expected: '错误消息显示"用户不存在"',
          selector: '#login-error-message'
        }
      ]
    };
  }

  /**
   * 创建锁定账户测试用例
   */
  private createLockedAccountTestCase(): TestCase {
    return {
      id: 'AUTH-LG-004',
      name: '锁定账户登录',
      description: '尝试登录已锁定的账户应显示账户锁定消息',
      priority: 'medium',
      tags: ['auth', 'login', 'security'],
      testData: {
        username: this.config.invalidUsers.lockedAccount.username,
        password: this.config.invalidUsers.lockedAccount.password
      },
      steps: [
        {
          id: 'step-1',
          name: '导航到登录页面',
          action: 'navigate',
          expected: '成功加载登录页面',
          url: `${this.config.baseUrl}/auth/login`
        },
        {
          id: 'step-2',
          name: '输入锁定账户用户名',
          action: 'fill',
          selector: '#login-username',
          value: this.config.invalidUsers.lockedAccount.username,
          expected: '用户名输入成功'
        },
        {
          id: 'step-3',
          name: '输入密码',
          action: 'fill',
          selector: '#login-password',
          value: this.config.invalidUsers.lockedAccount.password,
          expected: '密码输入成功'
        },
        {
          id: 'step-4',
          name: '点击登录按钮',
          action: 'click',
          selector: '#login-button',
          expected: '显示账户锁定消息'
        },
        {
          id: 'step-5',
          name: '验证锁定消息',
          action: 'verify-element',
          expected: '错误消息显示"账户已锁定"',
          selector: '#login-error-message'
        },
        {
          id: 'step-6',
          name: '验证解锁链接',
          action: 'verify-element',
          expected: '显示账户解锁链接',
          selector: '#account-unlock-link'
        }
      ]
    };
  }

  /**
   * 创建记住我功能测试用例
   */
  private createRememberMeTestCase(): TestCase {
    return {
      id: 'AUTH-LG-005',
      name: '记住我功能',
      description: '测试记住我功能的正常工作',
      priority: 'medium',
      tags: ['auth', 'login', 'feature'],
      steps: [
        {
          id: 'step-1',
          name: '导航到登录页面',
          action: 'navigate',
          expected: '成功加载登录页面',
          url: `${this.config.baseUrl}/auth/login`
        },
        {
          id: 'step-2',
          name: '输入用户名',
          action: 'fill',
          selector: '#login-username',
          value: this.config.validUser.username,
          expected: '用户名输入成功'
        },
        {
          id: 'step-3',
          name: '输入密码',
          action: 'fill',
          selector: '#login-password',
          value: this.config.validUser.password,
          expected: '密码输入成功'
        },
        {
          id: 'step-4',
          name: '勾选记住我',
          action: 'click',
          selector: '#login-remember-me',
          expected: '记住我复选框被选中'
        },
        {
          id: 'step-5',
          name: '点击登录按钮',
          action: 'click',
          selector: '#login-button',
          expected: '登录成功'
        },
        {
          id: 'step-6',
          name: '退出登录',
          action: 'click',
          selector: '#logout-button',
          expected: '成功退出'
        },
        {
          id: 'step-7',
          name: '重新访问登录页面',
          action: 'navigate',
          expected: '用户名已自动填充',
          url: `${this.config.baseUrl}/auth/login`
        },
        {
          id: 'step-8',
          name: '验证用户名自动填充',
          action: 'verify-value',
          expected: `用户名字段值为${this.config.validUser.username}`,
          selector: '#login-username',
          expectedValue: this.config.validUser.username
        }
      ]
    };
  }

  /**
   * 创建密码可见性测试用例
   */
  private createPasswordVisibilityTestCase(): TestCase {
    return {
      id: 'AUTH-LG-006',
      name: '密码可见性切换',
      description: '测试密码显示/隐藏切换功能',
      priority: 'low',
      tags: ['auth', 'login', 'ui'],
      steps: [
        {
          id: 'step-1',
          name: '导航到登录页面',
          action: 'navigate',
          expected: '成功加载登录页面',
          url: `${this.config.baseUrl}/auth/login`
        },
        {
          id: 'step-2',
          name: '输入密码',
          action: 'fill',
          selector: '#login-password',
          value: 'testpassword123',
          expected: '密码输入成功'
        },
        {
          id: 'step-3',
          name: '验证密码默认隐藏',
          action: 'verify-attribute',
          expected: '密码字段类型为password',
          selector: '#login-password',
          attribute: 'type',
          expectedValue: 'password'
        },
        {
          id: 'step-4',
          name: '点击显示密码按钮',
          action: 'click',
          selector: '#password-visibility-toggle',
          expected: '密码显示按钮被点击'
        },
        {
          id: 'step-5',
          name: '验证密码可见',
          action: 'verify-attribute',
          expected: '密码字段类型为text',
          selector: '#login-password',
          attribute: 'type',
          expectedValue: 'text'
        },
        {
          id: 'step-6',
          name: '再次点击隐藏密码',
          action: 'click',
          selector: '#password-visibility-toggle',
          expected: '密码隐藏按钮被点击'
        },
        {
          id: 'step-7',
          name: '验证密码隐藏',
          action: 'verify-attribute',
          expected: '密码字段类型为password',
          selector: '#login-password',
          attribute: 'type',
          expectedValue: 'password'
        }
      ]
    };
  }

  /**
   * 创建忘记密码测试用例
   */
  private createForgotPasswordTestCase(): TestCase {
    return {
      id: 'AUTH-FP-001',
      name: '忘记密码流程',
      description: '测试忘记密码功能完整流程',
      priority: 'medium',
      tags: ['auth', 'password-recovery'],
      testData: {
        email: this.config.validUser.email
      },
      steps: [
        {
          id: 'step-1',
          name: '导航到登录页面',
          action: 'navigate',
          expected: '成功加载登录页面',
          url: `${this.config.baseUrl}/auth/login`
        },
        {
          id: 'step-2',
          name: '点击忘记密码链接',
          action: 'click',
          selector: '#forgot-password-link',
          expected: '跳转到忘记密码页面'
        },
        {
          id: 'step-3',
          name: '验证忘记密码页面',
          action: 'verify-element',
          expected: '忘记密码页面标题显示',
          selector: '#forgot-password-title'
        },
        {
          id: 'step-4',
          name: '输入注册邮箱',
          action: 'fill',
          selector: '#forgot-password-email',
          value: this.config.validUser.email,
          expected: '邮箱输入成功'
        },
        {
          id: 'step-5',
          name: '点击发送重置链接',
          action: 'click',
          selector: '#send-reset-link-button',
          expected: '显示重置链接发送成功消息'
        },
        {
          id: 'step-6',
          name: '验证成功消息',
          action: 'verify-element',
          expected: '显示"重置链接已发送到您的邮箱"',
          selector: '#reset-success-message'
        },
        {
          id: 'step-7',
          name: '验证返回登录链接',
          action: 'verify-element',
          expected: '显示返回登录链接',
          selector: '#back-to-login-link'
        }
      ]
    };
  }

  /**
   * 创建用户注册测试用例
   */
  private createRegistrationTestCase(): TestCase {
    return {
      id: 'AUTH-REG-001',
      name: '新用户注册',
      description: '测试新用户注册完整流程',
      priority: 'high',
      tags: ['auth', 'registration'],
      testData: {
        newUser: {
          username: 'newtestuser',
          email: 'newtestuser@example.com',
          password: 'NewPassword123!',
          firstName: 'Test',
          lastName: 'User',
          role: 'attorney',
          department: 'litigation'
        }
      },
      steps: [
        {
          id: 'step-1',
          name: '导航到登录页面',
          action: 'navigate',
          expected: '成功加载登录页面',
          url: `${this.config.baseUrl}/auth/login`
        },
        {
          id: 'step-2',
          name: '点击注册链接',
          action: 'click',
          selector: '#register-link',
          expected: '跳转到注册页面'
        },
        {
          id: 'step-3',
          name: '验证注册页面',
          action: 'verify-element',
          expected: '注册页面标题显示',
          selector: '#register-title'
        },
        {
          id: 'step-4',
          name: '填写注册表单',
          action: 'fill-form',
          expected: '注册表单填写成功',
          form: {
            '#register-username': 'newtestuser',
            '#register-email': 'newtestuser@example.com',
            '#register-password': 'NewPassword123!',
            '#register-confirm-password': 'NewPassword123!',
            '#register-first-name': 'Test',
            '#register-last-name': 'User',
            '#register-role': 'attorney',
            '#register-department': 'litigation'
          }
        },
        {
          id: 'step-5',
          name: '点击注册按钮',
          action: 'click',
          selector: '#register-button',
          expected: '注册成功'
        },
        {
          id: 'step-6',
          name: '验证注册成功消息',
          action: 'verify-element',
          expected: '显示注册成功消息',
          selector: '#registration-success-message'
        },
        {
          id: 'step-7',
          name: '验证自动登录',
          action: 'verify-element',
          expected: '用户已自动登录并跳转到仪表板',
          selector: '#user-profile-name'
        }
      ]
    };
  }

  /**
   * 创建退出登录测试用例
   */
  private createLogoutTestCase(): TestCase {
    return {
      id: 'AUTH-LO-001',
      name: '用户退出登录',
      description: '测试用户正常退出登录流程',
      priority: 'high',
      tags: ['auth', 'logout'],
      steps: [
        {
          id: 'step-1',
          name: '使用有效凭据登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.validUser.username,
            password: this.config.validUser.password
          }
        },
        {
          id: 'step-2',
          name: '点击用户菜单',
          action: 'click',
          selector: '#user-menu-toggle',
          expected: '用户菜单展开'
        },
        {
          id: 'step-3',
          name: '点击退出登录',
          action: 'click',
          selector: '#logout-button',
          expected: '显示退出确认对话框'
        },
        {
          id: 'step-4',
          name: '确认退出',
          action: 'click',
          selector: '#logout-confirm-button',
          expected: '成功退出并跳转到登录页面'
        },
        {
          id: 'step-5',
          name: '验证跳转到登录页面',
          action: 'verify-url',
          expected: 'URL为登录页面',
          pattern: '.*login.*'
        },
        {
          id: 'step-6',
          name: '验证会话清除',
          action: 'verify-element',
          expected: '用户信息不再显示',
          selector: '#user-profile-name',
          expectedState: 'not-exists'
        },
        {
          id: 'step-7',
          name: '尝试访问受保护页面',
          action: 'navigate',
          expected: '重定向到登录页面',
          url: `${this.config.baseUrl}/dashboard`
        }
      ]
    };
  }

  /**
   * 创建会话超时测试用例
   */
  private createSessionTimeoutTestCase(): TestCase {
    return {
      id: 'AUTH-ST-001',
      name: '会话超时处理',
      description: '测试用户会话超时后的处理',
      priority: 'medium',
      tags: ['auth', 'session', 'timeout'],
      steps: [
        {
          id: 'step-1',
          name: '使用有效凭据登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.validUser.username,
            password: this.config.validUser.password
          }
        },
        {
          id: 'step-2',
          name: '导航到仪表板',
          action: 'navigate',
          expected: '成功加载仪表板',
          url: `${this.config.baseUrl}/dashboard`
        },
        {
          id: 'step-3',
          name: '等待会话超时',
          action: 'wait',
          expected: '等待会话超时时间',
          duration: 1800000 // 30分钟
        },
        {
          id: 'step-4',
          name: '尝试访问受保护页面',
          action: 'navigate',
          expected: '重定向到登录页面并显示会话超时消息',
          url: `${this.config.baseUrl}/cases`
        },
        {
          id: 'step-5',
          name: '验证会话超时消息',
          action: 'verify-element',
          expected: '显示"会话已超时，请重新登录"消息',
          selector: '#session-timeout-message'
        }
      ]
    };
  }

  /**
   * 创建并发登录测试用例
   */
  private createConcurrentLoginTestCase(): TestCase {
    return {
      id: 'AUTH-CL-001',
      name: '并发登录控制',
      description: '测试同一用户在不同设备的并发登录控制',
      priority: 'medium',
      tags: ['auth', 'concurrent', 'security'],
      steps: [
        {
          id: 'step-1',
          name: '在第一个标签页登录',
          action: 'login',
          expected: '成功登录系统',
          credentials: {
            username: this.config.validUser.username,
            password: this.config.validUser.password
          }
        },
        {
          id: 'step-2',
          name: '打开新标签页',
          action: 'open-new-tab',
          expected: '新标签页打开成功'
        },
        {
          id: 'step-3',
          name: '在新标签页尝试登录相同账户',
          action: 'login',
          expected: '系统提示已在其他地方登录',
          credentials: {
            username: this.config.validUser.username,
            password: this.config.validUser.password
          }
        },
        {
          id: 'step-4',
          name: '验证并发登录消息',
          action: 'verify-element',
          expected: '显示"该账户已在其他设备上登录"消息',
          selector: '#concurrent-login-message'
        },
        {
          id: 'step-5',
          name: '选择强制登录选项',
          action: 'click',
          selector: '#force-login-button',
          expected: '成功强制登录，第一个会话被踢出'
        }
      ]
    };
  }

  /**
   * 创建密码要求测试用例
   */
  private createPasswordRequirementsTestCase(): TestCase {
    return {
      id: 'AUTH-PR-001',
      name: '密码要求验证',
      description: '测试密码复杂度要求验证',
      priority: 'high',
      tags: ['auth', 'password', 'validation'],
      steps: [
        {
          id: 'step-1',
          name: '导航到注册页面',
          action: 'navigate',
          expected: '成功加载注册页面',
          url: `${this.config.baseUrl}/auth/register`
        },
        {
          id: 'step-2',
          name: '输入太短的密码',
          action: 'fill',
          selector: '#register-password',
          value: 'short',
          expected: '显示密码太短错误'
        },
        {
          id: 'step-3',
          name: '验证密码长度错误',
          action: 'verify-element',
          expected: '显示"密码长度至少8位"',
          selector: '#password-length-error'
        },
        {
          id: 'step-4',
          name: '输入无大写字母密码',
          action: 'fill',
          selector: '#register-password',
          value: 'password123!',
          expected: '显示缺少大写字母错误'
        },
        {
          id: 'step-5',
          name: '验证大写字母错误',
          action: 'verify-element',
          expected: '显示"密码必须包含大写字母"',
          selector: '#password-uppercase-error'
        },
        {
          id: 'step-6',
          name: '输入符合要求的密码',
          action: 'fill',
          selector: '#register-password',
          value: 'ValidPassword123!',
          expected: '密码验证通过'
        },
        {
          id: 'step-7',
          name: '验证密码要求提示',
          action: 'verify-element',
          expected: '密码要求提示显示且符合要求的项目被标记为绿色',
          selector: '#password-requirements'
        }
      ]
    };
  }

  /**
   * 创建邮箱验证测试用例
   */
  private createEmailValidationTestCase(): TestCase {
    return {
      id: 'AUTH-EV-001',
      name: '邮箱格式验证',
      description: '测试邮箱地址格式验证',
      priority: 'medium',
      tags: ['auth', 'email', 'validation'],
      steps: [
        {
          id: 'step-1',
          name: '导航到注册页面',
          action: 'navigate',
          expected: '成功加载注册页面',
          url: `${this.config.baseUrl}/auth/register`
        },
        {
          id: 'step-2',
          name: '输入无效邮箱格式',
          action: 'fill',
          selector: '#register-email',
          value: 'invalid-email',
          expected: '显示邮箱格式错误'
        },
        {
          id: 'step-3',
          name: '验证邮箱格式错误',
          action: 'verify-element',
          expected: '显示"请输入有效的邮箱地址"',
          selector: '#email-format-error'
        },
        {
          id: 'step-4',
          name: '输入已注册邮箱',
          action: 'fill',
          selector: '#register-email',
          value: this.config.validUser.email,
          expected: '显示邮箱已注册错误'
        },
        {
          id: 'step-5',
          name: '验证邮箱已存在错误',
          action: 'verify-element',
          expected: '显示"该邮箱已被注册"',
          selector: '#email-exists-error'
        },
        {
          id: 'step-6',
          name: '输入有效新邮箱',
          action: 'fill',
          selector: '#register-email',
          value: 'newvalidemail@example.com',
          expected: '邮箱验证通过'
        }
      ]
    };
  }

  /**
   * 创建登录重定向测试用例
   */
  private createRedirectAfterLoginTestCase(): TestCase {
    return {
      id: 'AUTH-RD-001',
      name: '登录后重定向',
      description: '测试登录后重定向到原始请求页面',
      priority: 'medium',
      tags: ['auth', 'redirect'],
      steps: [
        {
          id: 'step-1',
          name: '尝试访问受保护页面',
          action: 'navigate',
          expected: '重定向到登录页面并保存原始URL',
          url: `${this.config.baseUrl}/cases/123`
        },
        {
          id: 'step-2',
          name: '验证重定向到登录',
          action: 'verify-url',
          expected: '重定向到登录页面',
          pattern: '.*login.*'
        },
        {
          id: 'step-3',
          name: '验证重定向参数',
          action: 'verify-url-parameter',
          expected: 'URL包含redirect参数',
          parameter: 'redirect',
          expectedValue: '/cases/123'
        },
        {
          id: 'step-4',
          name: '使用有效凭据登录',
          action: 'login',
          expected: '成功登录',
          credentials: {
            username: this.config.validUser.username,
            password: this.config.validUser.password
          }
        },
        {
          id: 'step-5',
          name: '验证重定向到原始页面',
          action: 'verify-url',
          expected: '重定向到原始请求的页面',
          pattern: '.*cases/123.*'
        }
      ]
    };
  }

  /**
   * 运行特定的认证测试
   */
  override async runSpecificTest(testId: string): Promise<any> {
    const testCases = this.getAllTestCases();
    const testCase = testCases.find(tc => tc.id === testId);

    if (!testCase) {
      throw new Error(`测试用例 ${testId} 未找到`);
    }

    this.logger.info(`运行特定测试: ${testCase.name} (${testId})`);
    const results = await this.testEngine.executeTestCases([testCase]);

    return results[0];
  }

  /**
   * 运行特定类别的认证测试
   */
  override async runTestsByCategory(category: string): Promise<{
    passed: number;
    failed: number;
    total: number;
    results: any[];
  }> {
    const testCases = this.getAllTestCases();
    const categoryTests = testCases.filter(tc => tc.tags.includes(category));

    this.logger.info(`运行 ${category} 类别测试，共 ${categoryTests.length} 个测试用例`);
    const results = await this.testEngine.executeTestCases(categoryTests);

    const passed = results.filter(r => r.status === 'passed').length;
    const failed = results.filter(r => r.status === 'failed').length;
    const total = results.length;

    return {
      passed,
      failed,
      total,
      results
    };
  }

  /**
   * 生成测试报告
   */
  override async generateTestReport(results: any[]): Promise<string> {
    const report = {
      timestamp: new Date().toISOString(),
      suite: '用户认证测试',
      summary: {
        total: results.length,
        passed: results.filter(r => r.status === 'passed').length,
        failed: results.filter(r => r.status === 'failed').length,
        skipped: results.filter(r => r.status === 'skipped').length
      },
      results: results.map(result => ({
        id: result.testCase.id,
        name: result.testCase.name,
        status: result.status,
        duration: result.duration,
        error: result.error,
        steps: result.getstepResults?.().map((step: any) => ({
          id: step.step.id,
          name: step.step.name,
          status: step.status,
          error: step.error
        }))
      }))
    };

    return JSON.stringify(report, null, 2);
  }
}