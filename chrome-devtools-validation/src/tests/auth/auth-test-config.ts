/**
 * 认证测试配置
 */

import { AuthTestConfig, AuthTestUser } from './auth-test-suite';

export const AUTH_TEST_CONFIG: AuthTestConfig = {
  baseUrl: 'http://localhost:3000', // 开发环境URL
  defaultTimeout: 30000,
  screenshotOnFailure: true,

  // 有效测试用户
  validUser: {
    username: 'test.attorney',
    email: 'test.attorney@lawfirm.com',
    password: 'TestPassword123!',
    firstName: 'Test',
    lastName: 'Attorney',
    role: 'attorney',
    department: 'litigation',
    phone: '+1-555-0123'
  } as AuthTestUser,

  // 无效测试用户
  invalidUsers: {
    wrongPassword: {
      username: 'test.attorney',
      email: 'test.attorney@lawfirm.com',
      password: 'WrongPassword123!',
      firstName: 'Test',
      lastName: 'Attorney',
      role: 'attorney',
      department: 'litigation'
    } as AuthTestUser,

    nonExistent: {
      username: 'nonexistent.user',
      email: 'nonexistent@lawfirm.com',
      password: 'AnyPassword123!',
      firstName: 'Non',
      lastName: 'Existent',
      role: 'attorney',
      department: 'litigation'
    } as AuthTestUser,

    lockedAccount: {
      username: 'locked.attorney',
      email: 'locked.attorney@lawfirm.com',
      password: 'LockedPassword123!',
      firstName: 'Locked',
      lastName: 'Attorney',
      role: 'attorney',
      department: 'litigation'
    } as AuthTestUser
  }
};

// 测试环境配置
export const TEST_ENVIRONMENTS = {
  development: {
    baseUrl: 'http://localhost:3000',
    apiBaseUrl: 'http://localhost:3001/api',
    timeout: 30000,
    screenshots: true,
    headless: false
  },

  staging: {
    baseUrl: 'https://staging.lawfirm.com',
    apiBaseUrl: 'https://staging-api.lawfirm.com/api',
    timeout: 45000,
    screenshots: true,
    headless: true
  },

  production: {
    baseUrl: 'https://app.lawfirm.com',
    apiBaseUrl: 'https://api.lawfirm.com/api',
    timeout: 60000,
    screenshots: false,
    headless: true
  }
};

// 密码复杂度要求
export const PASSWORD_REQUIREMENTS = {
  minLength: 8,
  maxLength: 128,
  requireUppercase: true,
  requireLowercase: true,
  requireNumbers: true,
  requireSpecialChars: true,
  forbiddenPatterns: [
    'password',
    '123456',
    'qwerty',
    'admin',
    'user'
  ]
};

// 会话配置
export const SESSION_CONFIG = {
  timeout: 1800000, // 30分钟
  renewalThreshold: 300000, // 5分钟
  concurrentSessions: {
    allow: false,
    forceLogin: true,
    notifyPreviousSession: true
  }
};

// 安全配置
export const SECURITY_CONFIG = {
  maxLoginAttempts: 5,
  lockoutDuration: 900000, // 15分钟
  passwordExpiryDays: 90,
  mfaEnabled: false,
  captchaEnabled: true,
  ipWhitelist: ['127.0.0.1', '::1'],
  allowedOrigins: ['http://localhost:3000']
};

// 测试数据
export const TEST_USERS = {
  attorneys: [
    {
      username: 'attorney1',
      email: 'attorney1@lawfirm.com',
      password: 'TestPassword123!',
      firstName: 'John',
      lastName: 'Doe',
      role: 'attorney',
      department: 'litigation',
      phone: '+1-555-0101'
    },
    {
      username: 'attorney2',
      email: 'attorney2@lawfirm.com',
      password: 'TestPassword123!',
      firstName: 'Jane',
      lastName: 'Smith',
      role: 'attorney',
      department: 'corporate',
      phone: '+1-555-0102'
    }
  ],

  paralegals: [
    {
      username: 'paralegal1',
      email: 'paralegal1@lawfirm.com',
      password: 'TestPassword123!',
      firstName: 'Mike',
      lastName: 'Johnson',
      role: 'paralegal',
      department: 'litigation',
      phone: '+1-555-0201'
    }
  ],

  admins: [
    {
      username: 'admin1',
      email: 'admin1@lawfirm.com',
      password: 'AdminPassword123!',
      firstName: 'Admin',
      lastName: 'User',
      role: 'admin',
      department: 'it',
      phone: '+1-555-0301'
    }
  ]
};

// 测试场景
export const TEST_SCENARIOS = {
  positive: [
    'valid_login',
    'remember_me',
    'password_visibility',
    'successful_registration',
    'successful_logout',
    'redirect_after_login'
  ],

  negative: [
    'invalid_password',
    'nonexistent_user',
    'locked_account',
    'weak_password',
    'invalid_email',
    'duplicate_email'
  ],

  security: [
    'session_timeout',
    'concurrent_login',
    'brute_force_protection',
    'sql_injection_attempt',
    'xss_attempt'
  ],

  performance: [
    'login_response_time',
    'registration_performance',
    'password_reset_performance'
  ]
};

// 测试断言配置
export const ASSERTION_CONFIG = {
  timeout: 5000,
  retry: 3,
  retryDelay: 1000,
  screenshots: {
    onFailure: true,
    onSuccess: false,
    path: './test-screenshots/auth/'
  },
  logging: {
    level: 'info',
    file: './test-logs/auth-tests.log'
  }
};

// 测试报告配置
export const REPORT_CONFIG = {
  format: 'json',
  outputDir: './test-reports/auth/',
  includeScreenshots: true,
  includeLogs: true,
  includeMetrics: true,
  customFields: {
    environment: process.env.TEST_ENV || 'development',
    browser: process.env.TEST_BROWSER || 'chrome',
    version: process.env.APP_VERSION || 'latest'
  }
};

// 邮件服务配置（用于测试邮件相关功能）
export const EMAIL_CONFIG = {
  service: 'ethereal.email',
  host: 'smtp.ethereal.email',
  port: 587,
  secure: false,
  auth: {
    user: 'test@ethereal.email',
    pass: 'testpassword'
  },
  testAccount: {
    user: 'test.lawfirm@ethereal.email',
    pass: 'testpassword123'
  }
};

// 数据库配置（用于测试数据准备和清理）
export const DATABASE_CONFIG = {
  host: 'localhost',
  port: 3306,
  username: 'test_user',
  password: 'test_password',
  database: 'law_firm_test',
  pool: {
    min: 1,
    max: 5,
    idle: 30000,
    acquire: 60000
  }
};

// 文件上传配置
export const FILE_UPLOAD_CONFIG = {
  maxFileSize: 10485760, // 10MB
  allowedTypes: [
    'image/jpeg',
    'image/png',
    'image/gif',
    'application/pdf',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
  ],
  testFiles: {
    profileImage: './test-data/test-profile.jpg',
    document: './test-data/test-document.pdf',
    invalidFile: './test-data/test-invalid.exe'
  }
};

// API端点配置
export const API_ENDPOINTS = {
  auth: {
    login: '/api/auth/login',
    logout: '/api/auth/logout',
    register: '/api/auth/register',
    refreshToken: '/api/auth/refresh',
    forgotPassword: '/api/auth/forgot-password',
    resetPassword: '/api/auth/reset-password',
    validateToken: '/api/auth/validate'
  },
  users: {
    profile: '/api/users/profile',
    updateProfile: '/api/users/profile',
    changePassword: '/api/users/change-password',
    deleteAccount: '/api/users/account'
  }
};

// 测试工具函数
export const TestUtils = {
  /**
   * 生成随机用户名
   */
  generateRandomUsername(prefix: string = 'test'): string {
    const timestamp = Date.now();
    const random = Math.floor(Math.random() * 1000);
    return `${prefix}${timestamp}${random}`;
  },

  /**
   * 生成随机邮箱
   */
  generateRandomEmail(domain: string = 'test.com'): string {
    const username = this.generateRandomUsername('user');
    return `${username}@${domain}`;
  },

  /**
   * 生成符合要求的密码
   */
  generateValidPassword(): string {
    const uppercase = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    const lowercase = 'abcdefghijklmnopqrstuvwxyz';
    const numbers = '0123456789';
    const special = '!@#$%^&*()_+-=[]{}|;:,.<>?';

    let password = '';
    password += uppercase[Math.floor(Math.random() * uppercase.length)];
    password += lowercase[Math.floor(Math.random() * lowercase.length)];
    password += numbers[Math.floor(Math.random() * numbers.length)];
    password += special[Math.floor(Math.random() * special.length)];

    // 添加随机字符以达到最小长度
    const allChars = uppercase + lowercase + numbers + special;
    while (password.length < 12) {
      password += allChars[Math.floor(Math.random() * allChars.length)];
    }

    return password.split('').sort(() => Math.random() - 0.5).join('');
  },

  /**
   * 等待指定时间
   */
  override async wait(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  },

  /**
   * 重试函数
   */
  async retry<T>(
    fn: () => Promise<T>,
    maxAttempts: number = 3,
    delay: number = 1000
  ): Promise<T> {
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        return await fn();
      } catch (error: unknown) {
        if (attempt === maxAttempts) {
          throw error;
        }
        await this.wait(delay);
      }
    }
    throw new Error('重试失败');
  }
};

export default AUTH_TEST_CONFIG;