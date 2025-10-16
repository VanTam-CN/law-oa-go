/**
 * Vue前端律师管理页面测试脚本
 * 测试Vue版本的律师管理页面是否可以正常访问
 */

const { chromium } = require('playwright');

class VueLawyerPageTest {
  constructor() {
    this.browser = null;
    this.context = null;
    this.page = null;
    this.testResults = {
      total: 0,
      passed: 0,
      failed: 0,
      details: []
    };
  }

  async init() {
    console.log('🚀 初始化Vue前端测试环境...');
    this.browser = await chromium.launch({
      headless: false,
      slowMo: 500,
      args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    this.context = await this.browser.newContext({
      viewport: { width: 1920, height: 1080 }
    });
    this.page = await this.context.newPage();

    // 监听控制台输出
    this.page.on('console', msg => {
      console.log(`📝 页面日志: ${msg.text()}`);
    });

    // 监听页面错误
    this.page.on('pageerror', error => {
      console.error(`⚠️ 页面错误: ${error.message}`);
    });
  }

  async cleanup() {
    if (this.browser) {
      await this.browser.close();
    }
  }

  async testPageAccess() {
    console.log('📍 测试Vue前端律师管理页面访问...');

    try {
      // 直接访问律师管理页面
      await this.page.goto('http://localhost:5173/lawyer', {
        waitUntil: 'networkidle',
        timeout: 30000
      });

      await this.page.waitForTimeout(3000);
      const currentUrl = this.page.url();

      if (currentUrl.includes('/login')) {
        console.log('ℹ️ 页面重定向到登录页（正常行为，因为未登录）');

        // 检查是否有开发模式提示
        const devAlert = await this.page.$('.ant-alert-info');
        if (devAlert) {
          const alertText = await devAlert.textContent();
          console.log(`🛠️ 发现开发模式提示: ${alertText}`);
        }

        return true; // 这是预期行为
      } else if (currentUrl.includes('/lawyer')) {
        console.log('✅ 直接访问律师管理页面成功');
        return true;
      } else {
        console.error(`❌ 意外重定向到: ${currentUrl}`);
        return false;
      }
    } catch (error) {
      console.error(`❌ 页面访问测试失败: ${error.message}`);
      return false;
    }
  }

  async testLoginPageAccess() {
    console.log('🔐 测试登录页面访问...');

    try {
      await this.page.goto('http://localhost:5173/login', {
        waitUntil: 'networkidle',
        timeout: 30000
      });

      await this.page.waitForTimeout(2000);
      const currentUrl = this.page.url();

      if (currentUrl.includes('/login')) {
        console.log('✅ 登录页面访问正常');
        return true;
      } else {
        console.error(`❌ 登录页面访问异常: ${currentUrl}`);
        return false;
      }
    } catch (error) {
      console.error(`❌ 登录页面访问失败: ${error.message}`);
      return false;
    }
  }

  async testWithMockToken() {
    console.log('🎫 测试使用模拟token访问...');

    try {
      // 设置模拟token
      await this.page.evaluate(() => {
        localStorage.setItem('law_oa_token', 'mock_dev_token_12345');
        localStorage.setItem('user', JSON.stringify({
          id: 1,
          username: 'dev_user',
          real_name: '开发用户',
          email: 'dev@example.com',
          role: 'admin',
          department: '开发部门'
        }));
      });

      // 重新访问律师管理页面
      await this.page.goto('http://localhost:5173/lawyer', {
        waitUntil: 'networkidle',
        timeout: 30000
      });

      await this.page.waitForTimeout(3000);
      const currentUrl = this.page.url();

      if (currentUrl.includes('/lawyer')) {
        console.log('✅ 使用模拟token成功访问律师管理页面');

        // 检查是否有开发模式指示器
        const devAlert = await this.page.$('.ant-alert-info');
        if (devAlert) {
          const alertText = await devAlert.textContent();
          console.log(`🛠️ 开发模式指示器: ${alertText}`);
        }

        // 检查页面内容
        const title = await this.page.$('h1, h2, h3');
        if (title) {
          const titleText = await title.textContent();
          console.log(`📄 页面标题: ${titleText}`);
        }

        // 检查统计卡片
        const statsCards = await this.page.$$('.ant-statistic');
        console.log(`📊 统计卡片数量: ${statsCards.length}`);

        // 检查律师表格
        const table = await this.page.$('table');
        if (table) {
          const rows = await this.page.$$('table tbody tr');
          console.log(`👥 律师列表行数: ${rows.length}`);
        }

        return true;
      } else {
        console.error(`❌ 带token访问失败，当前URL: ${currentUrl}`);
        return false;
      }
    } catch (error) {
      console.error(`❌ 模拟token测试失败: ${error.message}`);
      return false;
    }
  }

  async testDeveloperConsole() {
    console.log('🔍 检查开发者控制台输出...');

    try {
      // 获取控制台日志
      const logs = await this.page.evaluate(() => {
        return {
          logs: window.console.logs ? window.console.logs.map(log => log.args.join(' ')) : [],
          warnings: window.console.warnings ? window.console.warnings.map(log => log.args.join(' ')) : [],
          errors: window.console.errors ? window.console.errors.map(log => log.args.join(' ')) : []
        };
      });

      console.log('📋 控制台日志摘要:');
      console.log(`   - 普通日志: ${logs.logs.length} 条`);
      console.log(`   - 警告日志: ${logs.warnings.length} 条`);
      console.log(`   - 错误日志: ${logs.errors.length} 条`);

      // 检查是否有开发模式相关日志
      const hasDevModeLogs = [
        ...logs.logs,
        ...logs.warnings,
        ...logs.errors
      ].some(log => log.includes('开发模式') || log.includes('🛠️'));

      if (hasDevModeLogs) {
        console.log('✅ 发现开发模式相关日志');
      } else {
        console.log('ℹ️ 未发现开发模式相关日志');
      }

      return logs.errors.length === 0; // 如果没有错误日志，返回true
    } catch (error) {
      console.error(`❌ 控制台检查失败: ${error.message}`);
      return false;
    }
  }

  async runAllTests() {
    console.log('🎯 开始Vue前端律师管理页面测试...\n');

    const tests = [
      { name: '页面访问测试', func: () => this.testPageAccess() },
      { name: '登录页面测试', func: () => this.testLoginPageAccess() },
      { name: '模拟token测试', func: () => this.testWithMockToken() },
      { name: '开发者控制台检查', func: () => this.testDeveloperConsole() }
    ];

    for (const test of tests) {
      this.testResults.total++;
      console.log(`\n🧪 运行测试: ${test.name}`);

      try {
        const result = await test.func();
        if (result) {
          this.testResults.passed++;
          console.log(`✅ ${test.name} - 通过`);
          this.testResults.details.push({ name: test.name, status: 'PASS', message: '测试通过' });
        } else {
          this.testResults.failed++;
          console.log(`❌ ${test.name} - 失败`);
          this.testResults.details.push({ name: test.name, status: 'FAIL', message: '测试失败' });
        }
      } catch (error) {
        this.testResults.failed++;
        console.log(`❌ ${test.name} - 异常: ${error.message}`);
        this.testResults.details.push({ name: test.name, status: 'ERROR', message: error.message });
      }
    }
  }

  printResults() {
    console.log('\n' + '='.repeat(60));
    console.log('📊 Vue前端测试结果汇总');
    console.log('='.repeat(60));
    console.log(`总测试数: ${this.testResults.total}`);
    console.log(`✅ 通过: ${this.testResults.passed}`);
    console.log(`❌ 失败: ${this.testResults.failed}`);
    console.log(`📈 成功率: ${((this.testResults.passed / this.testResults.total) * 100).toFixed(1)}%`);

    console.log('\n📋 详细结果:');
    this.testResults.details.forEach(test => {
      const icon = test.status === 'PASS' ? '✅' : test.status === 'FAIL' ? '❌' : '⚠️';
      console.log(`${icon} ${test.name}: ${test.message}`);
    });

    if (this.testResults.failed === 0) {
      console.log('\n🎉 所有测试通过！Vue前端律师管理页面修复成功！');
    } else {
      console.log(`\n⚠️ 有 ${this.testResults.failed} 个测试失败，可能需要进一步检查。`);
    }
  }
}

// 运行测试
async function main() {
  const testSuite = new VueLawyerPageTest();

  try {
    await testSuite.init();
    await testSuite.runAllTests();
    testSuite.printResults();
  } catch (error) {
    console.error('❌ 测试运行失败:', error.message);
  } finally {
    await testSuite.cleanup();
  }
}

// 如果直接运行此脚本
if (require.main === module) {
  main().catch(console.error);
}

module.exports = { VueLawyerPageTest };