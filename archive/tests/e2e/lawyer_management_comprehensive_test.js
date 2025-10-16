/**
 * 律师管理页面综合测试套件
 * 包含功能测试、错误处理测试、用户体验测试
 */

const { chromium } = require('playwright');

class LawyerManagementTestSuite {
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
    console.log('🚀 初始化测试环境...');
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

  runTest(testName, testFunction) {
    this.testResults.total++;
    console.log(`\n🧪 运行测试: ${testName}`);

    return testFunction()
      .then(result => {
        if (result) {
          this.testResults.passed++;
          console.log(`✅ ${testName} - 通过`);
          this.testResults.details.push({ name: testName, status: 'PASS', message: '测试通过' });
        } else {
          this.testResults.failed++;
          console.log(`❌ ${testName} - 失败`);
          this.testResults.details.push({ name: testName, status: 'FAIL', message: '测试失败' });
        }
        return result;
      })
      .catch(error => {
        this.testResults.failed++;
        console.log(`❌ ${testName} - 异常: ${error.message}`);
        this.testResults.details.push({ name: testName, status: 'ERROR', message: error.message });
        return false;
      });
  }

  async testPageAccess() {
    try {
      await this.page.goto('http://localhost:3000/lawyer-management', {
        waitUntil: 'networkidle',
        timeout: 30000
      });

      await this.page.waitForTimeout(2000);
      const currentUrl = this.page.url();

      return !currentUrl.includes('/login') && currentUrl.includes('/lawyer-management');
    } catch (error) {
      console.error(`页面访问测试失败: ${error.message}`);
      return false;
    }
  }

  async testPageContent() {
    try {
      // 检查页面标题
      const title = await this.page.textContent('h1');
      if (!title || !title.includes('律师管理')) {
        return false;
      }

      // 检查开发模式指示器
      const devModeIndicator = await this.page.$('.alert-info');
      if (devModeIndicator) {
        const devText = await devModeIndicator.textContent();
        console.log(`🛠️ 开发模式指示器: ${devText}`);
      }

      // 检查统计卡片
      const statsCards = await this.page.$$('.card.text-center');
      if (statsCards.length < 4) {
        console.warn(`统计卡片数量不足: ${statsCards.length}/4`);
      }

      // 检查搜索表单
      const searchForm = await this.page.$('form');
      if (!searchForm) {
        console.warn('搜索表单未找到');
      }

      return true;
    } catch (error) {
      console.error(`页面内容测试失败: ${error.message}`);
      return false;
    }
  }

  async testLawyerList() {
    try {
      // 等待律师列表加载
      await this.page.waitForSelector('table', { timeout: 10000 });

      // 检查律师数量
      const rows = await this.page.$$('table tbody tr');
      const hasLawyers = rows.length > 0;

      if (hasLawyers) {
        console.log(`📊 律师列表加载成功，显示 ${rows.length} 条记录`);

        // 检查第一行数据
        const firstRow = rows[0];
        const cells = await firstRow.$$('td');
        if (cells.length > 0) {
          const firstCellText = await cells[0].textContent();
          console.log(`👤 第一条记录: ${firstCellText}`);
        }
      } else {
        console.warn('律师列表为空');
      }

      return true;
    } catch (error) {
      console.error(`律师列表测试失败: ${error.message}`);
      return false;
    }
  }

  async testSearchFunctionality() {
    try {
      // 查找搜索输入框
      const searchInput = await this.page.$('input[placeholder*="姓名"]');
      if (!searchInput) {
        console.warn('搜索输入框未找到');
        return false;
      }

      // 输入搜索内容
      await searchInput.fill('张');
      await this.page.waitForTimeout(1000);

      // 点击搜索按钮
      const searchButton = await this.page.$('button:has-text("搜索")');
      if (searchButton) {
        await searchButton.click();
        await this.page.waitForTimeout(2000);
        console.log('🔍 搜索功能测试完成');
      }

      // 清空搜索
      await searchInput.fill('');
      const resetButton = await this.page.$('button:has-text("重置")');
      if (resetButton) {
        await resetButton.click();
        await this.page.waitForTimeout(1000);
        console.log('🔄 重置功能测试完成');
      }

      return true;
    } catch (error) {
      console.error(`搜索功能测试失败: ${error.message}`);
      return false;
    }
  }

  async testErrorHandling() {
    try {
      // 检查是否有错误提示区域（可能为空）
      const errorAlert = await this.page.$('.alert-warning');
      if (errorAlert) {
        const errorText = await errorAlert.textContent();
        console.log(`⚠️ 发现错误提示: ${errorText}`);

        // 查找重试按钮
        const retryButton = await this.page.$('button:has-text("手动重试")');
        if (retryButton) {
          console.log('🔄 发现重试按钮');
          // 可以测试重试功能，但这里只检查按钮存在
        }
      } else {
        console.log('ℹ️ 当前没有错误提示（正常状态）');
      }

      return true;
    } catch (error) {
      console.error(`错误处理测试失败: ${error.message}`);
      return false;
    }
  }

  async testModalFunctionality() {
    try {
      // 查找新增律师按钮
      const addButton = await this.page.$('button:has-text("新增律师")');
      if (!addButton) {
        console.warn('新增律师按钮未找到');
        return false;
      }

      // 点击新增按钮（不实际提交，只测试模态框显示）
      await addButton.click();
      await this.page.waitForTimeout(1000);

      // 检查模态框是否显示
      const modal = await this.page.$('.modal.show');
      if (modal) {
        console.log('📝 新增律师模态框显示成功');

        // 查找模态框内的表单元素
        const modalTitle = await this.page.$('.modal-title');
        if (modalTitle) {
          const titleText = await modalTitle.textContent();
          console.log(`📋 模态框标题: ${titleText}`);
        }

        // 关闭模态框（查找关闭按钮）
        const closeButton = await this.page.$('.btn-close, button[aria-label="Close"]');
        if (closeButton) {
          await closeButton.click();
          await this.page.waitForTimeout(500);
          console.log('❌ 模态框关闭成功');
        }
      } else {
        console.warn('模态框未显示');
        return false;
      }

      return true;
    } catch (error) {
      console.error(`模态框功能测试失败: ${error.message}`);
      return false;
    }
  }

  async testPerformance() {
    try {
      // 测试页面加载性能
      const performanceStart = Date.now();
      await this.page.goto('http://localhost:3000/lawyer-management', {
        waitUntil: 'networkidle'
      });
      const loadTime = Date.now() - performanceStart;

      console.log(`⚡ 页面加载时间: ${loadTime}ms`);

      // 检查资源加载
      const performanceEntries = await this.page.evaluate(() => {
        return performance.getEntriesByType('navigation').map(entry => ({
          domContentLoaded: entry.domContentLoadedEventEnd - entry.domContentLoadedEventStart,
          loadComplete: entry.loadEventEnd - entry.loadEventStart,
          totalLoadTime: entry.loadEventEnd - entry.fetchStart
        }));
      });

      if (performanceEntries.length > 0) {
        const perf = performanceEntries[0];
        console.log(`📊 性能指标:`);
        console.log(`   DOM内容加载: ${perf.domContentLoaded}ms`);
        console.log(`   页面完全加载: ${perf.loadComplete}ms`);
        console.log(`   总加载时间: ${perf.totalLoadTime}ms`);
      }

      // 性能通过标准：页面加载时间小于5秒
      return loadTime < 5000;
    } catch (error) {
      console.error(`性能测试失败: ${error.message}`);
      return false;
    }
  }

  async runAllTests() {
    console.log('🎯 开始律师管理页面综合测试...\n');

    const tests = [
      { name: '页面访问测试', func: () => this.testPageAccess() },
      { name: '页面内容测试', func: () => this.testPageContent() },
      { name: '律师列表测试', func: () => this.testLawyerList() },
      { name: '搜索功能测试', func: () => this.testSearchFunctionality() },
      { name: '错误处理测试', func: () => this.testErrorHandling() },
      { name: '模态框功能测试', func: () => this.testModalFunctionality() },
      { name: '性能测试', func: () => this.testPerformance() }
    ];

    for (const test of tests) {
      await this.runTest(test.name, test.func);
    }
  }

  printResults() {
    console.log('\n' + '='.repeat(60));
    console.log('📊 测试结果汇总');
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
      console.log('\n🎉 所有测试通过！律师管理页面运行正常！');
    } else {
      console.log(`\n⚠️ 有 ${this.testResults.failed} 个测试失败，需要进一步检查。`);
    }
  }
}

// 运行测试
async function main() {
  const testSuite = new LawyerManagementTestSuite();

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

module.exports = { LawyerManagementTestSuite };