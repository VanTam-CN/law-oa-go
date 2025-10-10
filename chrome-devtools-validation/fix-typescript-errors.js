#!/usr/bin/env node

/**
 * 批量修复TypeScript编译错误的脚本
 */

const fs = require('fs');
const path = require('path');

const filesToFix = [
  'src/pages/auth/password-reset-page.ts',
  'src/pages/auth/register-page.ts',
  'src/pages/cases/case-detail-page.ts',
  'src/pages/cases/case-list-page.ts',
  'src/pages/clients/client-list-page.ts',
  'src/pages/clients/client-detail-page.ts',
  'src/pages/documents/document-list-page.ts',
  'src/pages/documents/document-detail-page.ts',
  'src/pages/finance/finance-dashboard-page.ts'
];

const fixes = {
  // 修复selectors访问修饰符
  fixSelectorsAccess: (content) => {
    return content.replace(/private selectors = {/g, 'protected override selectors = {');
  },

  // 修复构造函数调用
  fixConstructorCall: (content) => {
    return content.replace(/super\(config, logger\);/g, 'super(config, this.selectors, logger);');
  },

  // 修复缺失的loginForm
  fixMissingLoginForm: (content) => {
    if (content.includes('waitForElement(this.selectors.loginForm)') && !content.includes('loginForm:')) {
      const selectorsMatch = content.match(/protected override selectors = {([^}]+)}/);
      if (selectorsMatch) {
        const newSelectors = 'protected override selectors = {\n    loginForm: \'#login-form\',\n    ' + selectorsMatch[1].trim();
        return content.replace(/protected override selectors = {[^}]+}/, newSelectors);
      }
    }
    return content;
  },

  // 修复缺失的registerForm
  fixMissingRegisterForm: (content) => {
    if (content.includes('waitForElement(this.selectors.registerForm)') && !content.includes('registerForm:')) {
      const selectorsMatch = content.match(/protected override selectors = {([^}]+)}/);
      if (selectorsMatch) {
        const newSelectors = 'protected override selectors = {\n    registerForm: \'#register-form\',\n    ' + selectorsMatch[1].trim();
        return content.replace(/protected override selectors = {[^}]+}/, newSelectors);
      }
    }
    return content;
  },

  // 修复缺失的resetForm
  fixMissingResetForm: (content) => {
    if (content.includes('waitForElement(this.selectors.resetForm)') && !content.includes('resetForm:')) {
      const selectorsMatch = content.match(/protected override selectors = {([^}]+)}/);
      if (selectorsMatch) {
        const newSelectors = 'protected override selectors = {\n    resetForm: \'#reset-form\',\n    ' + selectorsMatch[1].trim();
        return content.replace(/protected override selectors = {[^}]+}/, newSelectors);
      }
    }
    return content;
  }
};

function fixFile(filePath) {
  console.log(`修复文件: ${filePath}`);

  try {
    let content = fs.readFileSync(filePath, 'utf8');

    // 应用所有修复
    Object.values(fixes).forEach(fix => {
      content = fix(content);
    });

    fs.writeFileSync(filePath, content, 'utf8');
    console.log(`✅ 文件修复完成: ${filePath}`);
  } catch (error) {
    console.error(`❌ 修复文件失败: ${filePath}`, error.message);
  }
}

// 修复所有文件
filesToFix.forEach(fixFile);

console.log('批量修复完成！');