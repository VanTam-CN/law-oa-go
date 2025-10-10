const fs = require('fs');
const path = require('path');

/**
 * 简单修复脚本
 */

function fixConfigFile() {
  const filePath = path.join(__dirname, 'src/core/config.ts');
  let content = fs.readFileSync(filePath, 'utf8');

  // 修复 errors 数组问题
  content = content.replace(
    'const errors: string[] | undefined = undefined;',
    'let errors: string[] = [];'
  );

  fs.writeFileSync(filePath, content, 'utf8');
  console.log('✅ 修复 config.ts');
}

function fixBasePageObject() {
  const filePath = path.join(__dirname, 'src/core/base-page-object.ts');
  let content = fs.readFileSync(filePath, 'utf8');

  // 修复数组初始化问题
  content = content.replace(
    'const passed: any[] | undefined = undefined;',
    'let passed: any[] = [];'
  );
  content = content.replace(
    'const failed: any[] | undefined = undefined;',
    'let failed: any[] = [];'
  );

  // 修复数组长度检查
  content = content.replace(
    'if (result && result.length > 0) {',
    'if (result && (result.length || 0) > 0) {'
  );

  fs.writeFileSync(filePath, content, 'utf8');
  console.log('✅ 修复 base-page-object.ts');
}

function fixTestHelpers() {
  const filePath = path.join(__dirname, 'src/utils/test-helpers.ts');
  let content = fs.readFileSync(filePath, 'utf8');

  // 添加缺失的导入
  if (!content.includes('import { TestCase }')) {
    content = content.replace(
      'import { Logger } from \'../core/logger\';',
      'import { Logger } from \'../core/logger\';\nimport { TestCase } from \'../types/test-types\';'
    );
  }

  if (!content.includes('import { TestExecutionResult }')) {
    content = content.replace(
      'import { Logger } from \'../core/logger\';',
      'import { Logger } from \'../core/logger\';\nimport { TestExecutionResult } from \'../types/test-types\';'
    );
  }

  fs.writeFileSync(filePath, content, 'utf8');
  console.log('✅ 修复 test-helpers.ts');
}

function fixE2ETestRunner() {
  const filePath = path.join(__dirname, 'src/tests/e2e/e2e-test-runner.ts');
  let content = fs.readFileSync(filePath, 'utf8');

  // 修复数组初始化
  content = content.replace(
    'const results: any[] | undefined = undefined;',
    'let results: any[] = [];'
  );

  content = content.replace(
    'const recommendations: string[] | undefined = undefined;',
    'let recommendations: string[] = [];'
  );

  // 修复 push 调用
  content = content.replace(
    '(recommendations = recommendations || []).push',
    'recommendations.push'
  );

  fs.writeFileSync(filePath, content, 'utf8');
  console.log('✅ 修复 e2e-test-runner.ts');
}

function fixTestDataProvider() {
  const filePath = path.join(__dirname, 'src/core/test-data-provider.ts');
  let content = fs.readFileSync(filePath, 'utf8');

  // 修复方法签名
  content = content.replace(
    'async (): Promise<void> {',
    'async clearTestData(): Promise<void> {'
  );

  content = content.replace(
    'async set(dataKey: string, _data: any): Promise<void> {',
    'async setTestData(dataKey: string, data: any): Promise<void> {'
  );

  // 修复数组初始化
  content = content.replace(
    'const errors: Error[] | undefined = undefined;',
    'let errors: Error[] = [];'
  );

  fs.writeFileSync(filePath, content, 'utf8');
  console.log('✅ 修复 test-data-provider.ts');
}

console.log('🔧 开始简单修复...\n');

fixConfigFile();
fixBasePageObject();
fixTestHelpers();
fixE2ETestRunner();
fixTestDataProvider();

console.log('\n✅ 简单修复完成！');