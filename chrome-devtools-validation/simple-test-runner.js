#!/usr/bin/env node

/**
 * 简单测试运行器 - 绕过TypeScript编译问题直接运行测试
 */

const fs = require('fs');
const path = require('path');

class SimpleTestRunner {
  constructor() {
    this.results = {
      total: 0,
      passed: 0,
      failed: 0,
      skipped: 0,
      tests: []
    };
    this.startTime = Date.now();
  }

  log(message, level = 'info') {
    const timestamp = new Date().toISOString();
    const prefix = level === 'error' ? '❌' : level === 'warn' ? '⚠️' : level === 'success' ? '✅' : '📝';
    console.log(`[${timestamp}] ${prefix} ${message}`);
  }

  async runTest(testName, testFunction) {
    this.results.total++;
    const testStart = Date.now();

    try {
      this.log(`开始测试: ${testName}`);

      if (typeof testFunction === 'function') {
        await testFunction();
      }

      const duration = Date.now() - testStart;
      this.results.passed++;
      this.results.tests.push({
        name: testName,
        status: 'passed',
        duration,
        error: null
      });

      this.log(`测试通过: ${testName} (${duration}ms)`, 'success');

    } catch (error) {
      const duration = Date.now() - testStart;
      this.results.failed++;
      this.results.tests.push({
        name: testName,
        status: 'failed',
        duration,
        error: error.message
      });

      this.log(`测试失败: ${testName} - ${error.message}`, 'error');
    }
  }

  async runSystemTests() {
    this.log('🚀 开始系统测试...');

    // 测试1: 检查项目结构
    await this.runTest('项目结构检查', async () => {
      const requiredDirs = [
        'src/core',
        'src/types',
        'src/tests',
        'src/utils',
        'src/pages'
      ];

      for (const dir of requiredDirs) {
        const dirPath = path.join(__dirname, dir);
        if (!fs.existsSync(dirPath)) {
          throw new Error(`缺少必需目录: ${dir}`);
        }
      }

      const requiredFiles = [
        'src/core/test-execution-engine.ts',
        'src/types/test-types.ts',
        'package.json',
        'tsconfig.json'
      ];

      for (const file of requiredFiles) {
        const filePath = path.join(__dirname, file);
        if (!fs.existsSync(filePath)) {
          throw new Error(`缺少必需文件: ${file}`);
        }
      }
    });

    // 测试2: 检查依赖包
    await this.runTest('依赖包检查', async () => {
      const packageJsonPath = path.join(__dirname, 'package.json');
      const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));

      const requiredDeps = ['typescript', 'ts-node', 'winston'];
      for (const dep of requiredDeps) {
        if (!packageJson.dependencies[dep] && !packageJson.devDependencies[dep]) {
          throw new Error(`缺少必需依赖: ${dep}`);
        }
      }
    });

    // 测试3: 检查配置文件
    await this.runTest('配置文件检查', async () => {
      const tsConfigPath = path.join(__dirname, 'tsconfig.json');
      const tsConfig = JSON.parse(fs.readFileSync(tsConfigPath, 'utf8'));

      if (!tsConfig.compilerOptions) {
        throw new Error('tsconfig.json 缺少 compilerOptions');
      }

      if (tsConfig.compilerOptions.strict !== false) {
        this.log('注意: TypeScript严格模式已启用，可能会有类型错误', 'warn');
      }
    });

    // 测试4: 检查测试文件
    await this.runTest('测试文件检查', async () => {
      const testDirs = [
        'src/tests/auth',
        'src/tests/case',
        'src/tests/e2e'
      ];

      for (const dir of testDirs) {
        const dirPath = path.join(__dirname, dir);
        if (!fs.existsSync(dirPath)) {
          throw new Error(`缺少测试目录: ${dir}`);
        }

        const files = fs.readdirSync(dirPath);
        const testFiles = files.filter(f => f.endsWith('.ts'));

        if (testFiles.length === 0) {
          throw new Error(`测试目录 ${dir} 中没有测试文件`);
        }

        this.log(`测试目录 ${dir} 包含 ${testFiles.length} 个测试文件`);
      }
    });

    // 测试5: 检查Page Object文件
    await this.runTest('Page Object文件检查', async () => {
      const pageObjectDir = path.join(__dirname, 'src/pages');
      if (!fs.existsSync(pageObjectDir)) {
        throw new Error('缺少Page Object目录');
      }

      const files = fs.readdirSync(pageObjectDir);
      const pageObjectFiles = files.filter(f => f.endsWith('.ts'));

      this.log(`发现 ${pageObjectFiles.length} 个Page Object文件`);

      let hasValidPageObjects = false;
      for (const file of pageObjectFiles) {
        const filePath = path.join(pageObjectDir, file);
        const content = fs.readFileSync(filePath, 'utf8');

        // 跳过索引文件，它只需要包含导出
        if (file === 'index.ts') {
          if (!content.includes('export')) {
            throw new Error(`Page Object索引文件 ${file} 格式不正确`);
          }
          continue;
        }

        // 检查其他Page Object文件
        if (content.includes('class') && content.includes('PageObject')) {
          hasValidPageObjects = true;
        } else if (!content.includes('export')) {
          throw new Error(`Page Object文件 ${file} 格式不正确`);
        }
      }

      if (!hasValidPageObjects) {
        throw new Error('未找到有效的Page Object类定义');
      }
    });

    // 测试6: 检查核心引擎文件
    await this.runTest('核心引擎文件检查', async () => {
      const enginePath = path.join(__dirname, 'src/core/test-execution-engine.ts');
      const content = fs.readFileSync(enginePath, 'utf8');

      if (!content.includes('TestExecutionEngine')) {
        throw new Error('测试执行引擎文件格式不正确');
      }

      if (!content.includes('executeSuite') && !content.includes('executeTestCase')) {
        throw new Error('测试执行引擎缺少必需方法');
      }

      this.log('测试执行引擎文件检查通过');
    });

    // 测试7: 检查类型定义文件
    await this.runTest('类型定义文件检查', async () => {
      const typesPath = path.join(__dirname, 'src/types/test-types.ts');
      const content = fs.readFileSync(typesPath, 'utf8');

      const requiredTypes = ['TestCase', 'TestStep', 'TestResult', 'Assertion'];
      for (const type of requiredTypes) {
        if (!content.includes(`interface ${type}`) && !content.includes(`type ${type}`)) {
          throw new Error(`缺少必需类型定义: ${type}`);
        }
      }

      this.log('类型定义文件检查通过');
    });

    // 测试8: 检查工具类文件
    await this.runTest('工具类文件检查', async () => {
      const utilsPath = path.join(__dirname, 'src/utils');
      if (!fs.existsSync(utilsPath)) {
        throw new Error('缺少工具类目录');
      }

      const files = fs.readdirSync(utilsPath);
      const utilFiles = files.filter(f => f.endsWith('.ts'));

      this.log(`发现 ${utilFiles.length} 个工具类文件`);

      for (const file of utilFiles) {
        const filePath = path.join(utilsPath, file);
        const content = fs.readFileSync(filePath, 'utf8');

        if (!content.includes('class') && !content.includes('function') && !content.includes('interface')) {
          throw new Error(`工具类文件 ${file} 格式不正确`);
        }
      }
    });
  }

  generateReport() {
    const duration = Date.now() - this.startTime;
    const successRate = this.results.total > 0 ? (this.results.passed / this.results.total * 100).toFixed(2) : 0;

    console.log('\n' + '='.repeat(60));
    console.log('📊 系统测试报告');
    console.log('='.repeat(60));
    console.log(`总测试数: ${this.results.total}`);
    console.log(`通过: ${this.results.passed}`);
    console.log(`失败: ${this.results.failed}`);
    console.log(`跳过: ${this.results.skipped}`);
    console.log(`成功率: ${successRate}%`);
    console.log(`执行时间: ${duration}ms`);
    console.log('='.repeat(60));

    if (this.results.failed > 0) {
      console.log('\n❌ 失败的测试:');
      this.results.tests
        .filter(t => t.status === 'failed')
        .forEach(t => {
          console.log(`   - ${t.name}: ${t.error}`);
        });
    }

    if (this.results.passed === this.results.total) {
      console.log('\n✅ 所有系统测试通过！');
    } else {
      console.log('\n⚠️ 部分测试失败，请检查上述错误');
    }

    return this.results;
  }
}

// 主函数
async function main() {
  const runner = new SimpleTestRunner();

  try {
    await runner.runSystemTests();
    const results = runner.generateReport();

    // 设置退出码
    process.exit(results.failed > 0 ? 1 : 0);

  } catch (error) {
    console.error('❌ 测试运行器发生错误:', error);
    process.exit(1);
  }
}

// 运行主函数
if (require.main === module) {
  main();
}

module.exports = SimpleTestRunner;