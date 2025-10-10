const fs = require('fs');
const path = require('path');

/**
 * 最终修复脚本 - 处理剩余的TypeScript错误
 */

// 修复数组长度问题
function fixArrayLengthIssues(content) {
  // 修复数组长度检查
  content = content.replace(/if \(result && result\.length > 0\) \{/g, 'if (result && (result.length || 0) > 0) {');
  content = content.replace(/if \((\w+)\s*\?\.\s*length > 0\) \{/g, 'if (($1 || []).length > 0) {');

  return content;
}

// 修复常量重新赋值问题
function fixConstAssignment(content) {
  // 修复 const errors 重新赋值
  content = content.replace(/const\s+(errors):\s*(string\[\]\s*\|\s*undefined)\s*=\s*undefined;/g, 'let $1: string[] = [];');
  content = content.replace(/const\s+(errors):\s*string\[\]\s*=\s*\[\];/g, 'let $1: string[] = [];');

  // 修复 (errors = errors || []).push 模式
  content = content.replace(/\((\w+)\s*=\s*\1\s*\|\|\s*\[\])\)\.push/g, '$1.push');

  // 修复 recommendations 常量问题
  content = content.replace(/const\s+(recommendations):\s*(string\[\]\s*\|\s*undefined)\s*=\s*undefined;/g, 'let $1: string[] = [];');
  content = content.replace(/\(recommendations\s*=\s*recommendations\s*\|\|\s*\[\]\)\.push/g, 'recommendations.push');

  return content;
}

// 修复undefined数组问题
function fixUndefinedArrays(content) {
  // 修复数组初始化
  content = content.replace(/const\s+(\w+):\s*(\w+\[\])\s*\|\s*undefined\s*=\s*undefined;/g, 'let $1: $2 = [];');
  content = content.replace(/const\s+(\w+):\s*(any\[\])\s*\|\s*undefined\s*=\s*undefined;/g, 'let $1: any[] = [];');
  content = content.replace(/const\s+(\w+):\s*(Error\[\])\s*\|\s*undefined\s*=\s*undefined;/g, 'let $1: Error[] = [];');

  return content;
}

// 修复方法签名问题
function fixMethodSignatures(content) {
  // 修复缺失的方法名
  content = content.replace(/\s+async\s+\(\):/g, ' async clearTestData():');
  content = content.replace(/\s+async\s+set\(/g, ' async setTestData(');

  return content;
}

// 修复可能为undefined的访问
function fixOptionalAccess(content) {
  // 添加可选链操作符
  content = content.replace(/(\w+)\.length/g, '$1?.length || 0');
  content = content.replace(/(\w+)\.push/g, '$1?.push');

  return content;
}

// 修复类型导入问题
function fixTypeImports(content) {
  // 添加缺失的TestCase导入
  if (content.includes('validateTestCase(testCase: TestCase') && !content.includes('import { TestCase }')) {
    if (content.includes('import { Logger }')) {
      content = content.replace(
        /(import { Logger } from '[^']+';)/,
        '$1\nimport { TestCase } from \'../types/test-types\';'
      );
    } else if (content.includes('import')) {
      content = content.replace(
        /(import [^;]+;)/,
        '$1\nimport { TestCase } from \'../types/test-types\';'
      );
    }
  }

  // 添加缺失的TestExecutionResult导入
  if (content.includes('TestExecutionResult') && !content.includes('import.*TestExecutionResult')) {
    if (content.includes('import')) {
      content = content.replace(
        /(import [^;]+;)/,
        '$1\nimport { TestExecutionResult } from \'../types/test-types\';'
      );
    }
  }

  return content;
}

// 处理文件
function processFile(filePath) {
  try {
    let content = fs.readFileSync(filePath, 'utf8');
    const originalContent = content;

    // 应用所有修复
    content = fixArrayLengthIssues(content);
    content = fixConstAssignment(content);
    content = fixUndefinedArrays(content);
    content = fixMethodSignatures(content);
    content = fixOptionalAccess(content);
    content = fixTypeImports(content);

    // 只有当内容有变化时才写入
    if (content !== originalContent) {
      fs.writeFileSync(filePath, content, 'utf8');
      console.log(`✅ 修复完成: ${filePath}`);
    } else {
      console.log(`⏭️ 无需修复: ${filePath}`);
    }
  } catch (error) {
    console.error(`❌ 修复失败: ${filePath}`, error.message);
  }
}

// 主函数
function main() {
  // 需要修复的文件列表
  const filesToFix = [
    'src/core/base-page-object.ts',
    'src/core/config.ts',
    'src/core/test-data-provider.ts',
    'src/tests/e2e/e2e-test-runner.ts',
    'src/utils/test-helpers.ts'
  ];

  console.log('🔧 开始最终修复...\n');

  filesToFix.forEach(file => {
    const filePath = path.join(__dirname, file);
    if (fs.existsSync(filePath)) {
      processFile(filePath);
    } else {
      console.warn(`⚠️ 文件不存在: ${filePath}`);
    }
  });

  console.log('\n✅ 最终修复完成！');
}

main();