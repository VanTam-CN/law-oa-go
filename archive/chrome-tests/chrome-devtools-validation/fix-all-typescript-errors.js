#!/usr/bin/env node

/**
 * 全面修复TypeScript编译错误的脚本
 */

const fs = require('fs');
const path = require('path');

// 需要修复的文件列表
const filesToFix = [
  'src/pages/auth/login-page.ts',
  'src/pages/auth/password-reset-page.ts',
  'src/pages/auth/register-page.ts',
  'src/pages/cases/case-detail-page.ts',
  'src/pages/cases/case-list-page.ts',
  'src/pages/cases/case-form-page.ts',
  'src/pages/clients/client-list-page.ts',
  'src/pages/clients/client-detail-page.ts',
  'src/pages/documents/document-list-page.ts',
  'src/pages/documents/document-detail-page.ts',
  'src/pages/finance/finance-dashboard-page.ts'
];

// 修复函数集合
const fixes = {
  // 修复构造函数中的super调用
  fixConstructorSuperCall: (content, fileName) => {
    if (fileName.includes('login-page')) {
      return content.replace(
        /constructor\(config: PageObjectConfig, logger\?: Logger\) {\s*super\(config, this\.selectors, logger\);\s*}/,
        `constructor(config: PageObjectConfig, logger?: Logger) {
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
  }`
      );
    }

    if (fileName.includes('password-reset-page')) {
      return content.replace(
        /constructor\(config: PageObjectConfig, logger\?: Logger\) {\s*super\(config, this\.selectors, logger\);\s*}/,
        `constructor(config: PageObjectConfig, logger?: Logger) {
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
  }`
      );
    }

    if (fileName.includes('register-page')) {
      return content.replace(
        /constructor\(config: PageObjectConfig, logger\?: Logger\) {\s*super\(config, this\.selectors, logger\);\s*}/,
        `constructor(config: PageObjectConfig, logger?: Logger) {
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
  }`
      );
    }

    return content;
  },

  // 修复selectors声明
  fixSelectorsDeclaration: (content) => {
    return content.replace(/protected override selectors = {/g, '  private selectors = {');
  },

  // 修复缺失的loginForm引用
  fixMissingLoginFormReference: (content) => {
    if (content.includes('waitForElement(this.selectors.loginForm)') && content.includes('resetForm:')) {
      return content.replace('waitForElement(this.selectors.loginForm)', 'waitForElement(this.selectors.resetForm)');
    }
    if (content.includes('waitForElement(this.selectors.loginForm)') && content.includes('registerForm:')) {
      return content.replace('waitForElement(this.selectors.loginForm)', 'waitForElement(this.selectors.registerForm)');
    }
    return content;
  },

  // 修复CaseListPage的selectors类型问题
  fixCaseListSelectors: (content) => {
    if (content.includes('caseRow: (id: string) => string')) {
      return content.replace(/caseRow: \(id: string\) => string,/g, 'caseRow: `tr[data-case-id="PLACEHOLDER"]`,');
    }
    return content;
  },

  // 修复函数式selectors的调用
  fixFunctionalSelectorCalls: (content) => {
    // 将 this.selectors.caseRow(id) 替换为 this.selectors.caseRow.replace('PLACEHOLDER', id)
    return content.replace(/this\.selectors\.caseRow\(([^)]+)\)/g, 'this.selectors.caseRow.replace(\'PLACEHOLDER\', $1)');
  }
};

function fixFile(filePath) {
  console.log(`修复文件: ${filePath}`);

  try {
    let content = fs.readFileSync(filePath, 'utf8');
    const fileName = path.basename(filePath);

    // 应用所有修复
    Object.values(fixes).forEach(fix => {
      content = fix(content, fileName);
    });

    fs.writeFileSync(filePath, content, 'utf8');
    console.log(`✅ 文件修复完成: ${filePath}`);
  } catch (error) {
    console.error(`❌ 修复文件失败: ${filePath}`, error.message);
  }
}

// 修复所有文件
filesToFix.forEach(fixFile);

console.log('全面修复完成！');
console.log('请运行 npm run build 检查是否还有错误');