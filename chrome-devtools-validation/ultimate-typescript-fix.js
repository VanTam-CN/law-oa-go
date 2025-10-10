const fs = require('fs');
const path = require('path');

/**
 * 修复TypeScript严格模式导致的类型错误
 */

// 修复数组初始化问题
function fixArrayInitialization(content) {
  // 修复 undefined[] 类型问题
  content = content.replace(/(\w+\s*\??):\s*(\w+\[\])\s*=\s*\[\];/g, '$1: $2 | undefined = undefined;');
  content = content.replace(/(\w+\s*\??):\s*(\w+\[\])\s*=\s*\[\]\s*\|\|\s*\[\];/g, '$1: $2 | undefined = undefined;');

  // 修复 const 数组重新赋值问题
  content = content.replace(/const\s+(\w+):\s*(\w+\[\])\s*=\s*\[\];/g, 'let $1: $2 = [];');
  content = content.replace(/const\s+(errors):\s*(string\[\])\s*=\s*undefined;/g, 'let $1: string[] = [];');

  return content;
}

// 修复 override 修饰符问题
function fixOverrideModifiers(content) {
  // 移除不必要的 override 修饰符
  content = content.replace(/(\s+)override\s+(async\s+)?(get|set|clear)\w+/g, '$1$2$3');
  content = content.replace(/(\s+)override\s+(async\s+)?\w+/g, '$1$2');

  return content;
}

// 修复可选属性赋值问题
function fixOptionalPropertyAssignment(content) {
  // 修复可选属性的类型问题
  content = content.replace(/(\w+\??):\s*(\w+\[\])\s*=\s*undefined;/g, '$1: $2 | undefined = undefined;');
  content = content.replace(/(\w+\??):\s*(\w+)\s*=\s*undefined;/g, '$1: $2 | undefined = undefined;');

  return content;
}

// 修复常量重新赋值问题
function fixConstReassignment(content) {
  // 将 errors 数组改为 let 声明
  content = content.replace(/const\s+(errors):\s*(string\[\])\s*=\s*\[\];/g, 'let $1: string[] = [];');
  content = content.replace(/const\s+(errors):\s*string\[\]\s*=\s*undefined;/g, 'let $1: string[] = [];');

  // 修复常量重新赋值
  content = content.replace(/\(errors\s*=\s*errors\s*\|\|\s*\[\]\)\.push/g, 'errors.push');
  content = content.replace(/\(errors\s*=\s*errors\s*\|\|\s*\[\]\)/g, 'errors');

  return content;
}

// 修复类型断言问题
function fixTypeAssertions(content) {
  // 添加类型断言
  content = content.replace(/\.length\s*\?\s*\w+\.length\s*:\s*0/g, '.length || 0');
  content = content.replace(/(\w+)\s*\?\.\s*length/g, '$1?.length || 0');

  return content;
}

// 修复导入问题
function fixImports(content) {
  // 添加缺失的导入
  if (content.includes('TestCase') && !content.includes('import.*TestCase')) {
    content = content.replace(/(import.*\{[^}]*\}.*from.*test-types['"`];?)/, '$1\nimport { TestCase } from \'../types/test-types\';');
  }

  if (content.includes('TestExecutionResult') && !content.includes('import.*TestExecutionResult')) {
    content = content.replace(/(import.*\{[^}]*\}.*from.*test-types['"`];?)/, '$1\nimport { TestExecutionResult } from \'../types/test-types\';');
  }

  return content;
}

// 处理文件
function processFile(filePath) {
  try {
    let content = fs.readFileSync(filePath, 'utf8');

    // 应用所有修复
    content = fixArrayInitialization(content);
    content = fixOverrideModifiers(content);
    content = fixOptionalPropertyAssignment(content);
    content = fixConstReassignment(content);
    content = fixTypeAssertions(content);
    content = fixImports(content);

    fs.writeFileSync(filePath, content, 'utf8');
    console.log(`✅ 修复完成: ${filePath}`);
  } catch (error) {
    console.error(`❌ 修复失败: ${filePath}`, error.message);
  }
}

// 主函数
function main() {
  const srcDir = path.join(__dirname, 'src');

  // 需要修复的文件列表
  const filesToFix = [
    'src/core/base-page-object.ts',
    'src/core/config.ts',
    'src/core/logger.ts',
    'src/core/page-object-factory.ts',
    'src/core/test-data-provider.ts',
    'src/core/test-execution-engine.ts',
    'src/tests/e2e/e2e-test-runner.ts',
    'src/utils/test-data-generator.ts',
    'src/utils/test-helpers.ts'
  ];

  console.log('🚀 开始修复TypeScript编译错误...\n');

  filesToFix.forEach(file => {
    const filePath = path.join(__dirname, file);
    if (fs.existsSync(filePath)) {
      processFile(filePath);
    } else {
      console.warn(`⚠️ 文件不存在: ${filePath}`);
    }
  });

  console.log('\n✅ TypeScript修复完成！');
  console.log('📝 请运行 "npm run build" 检查修复结果');
}

main();