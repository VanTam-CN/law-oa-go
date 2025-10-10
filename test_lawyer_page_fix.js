/**
 * 律师管理页面修复验证测试
 * 测试律师列表页面是否还会重定向到登录界面
 */

const { chromium } = require('playwright');

async function testLawyerPageAccess() {
  console.log('🚀 开始测试律师管理页面访问...');

  const browser = await chromium.launch({
    headless: false, // 显示浏览器便于观察
    slowMo: 1000    // 减慢操作速度
  });

  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    // 监听控制台输出
    page.on('console', msg => {
      console.log(`📝 页面日志: ${msg.text()}`);
    });

    // 监听页面错误
    page.on('pageerror', error => {
      console.error(`⚠️ 页面错误: ${error.message}`);
    });

    // 监听请求
    page.on('request', request => {
      if (request.url().includes('/users/profile') || request.url().includes('/admin/users')) {
        console.log(`🌐 API请求: ${request.method()} ${request.url()}`);
      }
    });

    page.on('response', response => {
      if (response.url().includes('/users/profile') || response.url().includes('/admin/users')) {
        console.log(`📡 API响应: ${response.status()} ${response.url()}`);
      }
    });

    console.log('📍 导航到律师管理页面...');

    // 直接访问律师管理页面
    await page.goto('http://localhost:3000/lawyer-management', {
      waitUntil: 'networkidle',
      timeout: 30000
    });

    console.log('⏳ 等待页面加载完成...');
    await page.waitForTimeout(3000);

    // 检查当前URL，看是否被重定向到登录页面
    const currentUrl = page.url();
    console.log(`🔍 当前URL: ${currentUrl}`);

    if (currentUrl.includes('/login')) {
      console.error('❌ 测试失败：页面被重定向到登录界面');
      return false;
    } else if (currentUrl.includes('/lawyer-management')) {
      console.log('✅ 测试通过：成功访问律师管理页面');

      // 检查页面内容
      const pageTitle = await page.textContent('h1');
      console.log(`📄 页面标题: ${pageTitle}`);

      if (pageTitle && pageTitle.includes('律师管理')) {
        console.log('✅ 页面内容加载正常');

        // 检查是否有开发模式指示器
        const devModeIndicator = await page.$('.alert-info');
        if (devModeIndicator) {
          const devModeText = await devModeIndicator.textContent();
          console.log(`🛠️ 开发模式指示器: ${devModeText}`);
        }

        // 检查律师列表是否加载
        const lawyerTable = await page.$('table');
        if (lawyerTable) {
          console.log('✅ 律师列表表格已加载');

          // 统计律师数量
          const rows = await page.$$('table tbody tr');
          console.log(`📊 显示的律师数量: ${rows.length}`);
        } else {
          console.warn('⚠️ 律师列表表格未找到');
        }

        // 检查统计卡片
        const statsCards = await page.$$('.card.text-center');
        console.log(`📈 统计卡片数量: ${statsCards.length}`);

        return true;
      } else {
        console.error('❌ 页面内容异常');
        return false;
      }
    } else {
      console.error(`❌ 测试失败：页面被重定向到未知页面: ${currentUrl}`);
      return false;
    }

  } catch (error) {
    console.error(`❌ 测试过程中发生错误: ${error.message}`);
    return false;
  } finally {
    console.log('🔚 关闭浏览器...');
    await browser.close();
  }
}

// 测试多次访问，确保稳定性
async function runMultipleTests() {
  console.log('🔄 开始多次测试以确保稳定性...');

  const results = [];
  const testCount = 3;

  for (let i = 1; i <= testCount; i++) {
    console.log(`\n--- 第 ${i}/${testCount} 次测试 ---`);
    const result = await testLawyerPageAccess();
    results.push(result);

    if (i < testCount) {
      console.log('⏱️ 等待 2 秒后进行下一次测试...');
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
  }

  // 统计结果
  const successCount = results.filter(r => r === true).length;
  console.log(`\n📊 测试结果统计:`);
  console.log(`✅ 成功: ${successCount}/${testCount}`);
  console.log(`❌ 失败: ${testCount - successCount}/${testCount}`);
  console.log(`📈 成功率: ${((successCount / testCount) * 100).toFixed(1)}%`);

  if (successCount === testCount) {
    console.log('🎉 所有测试通过！律师管理页面修复成功！');
  } else {
    console.log('⚠️ 部分测试失败，可能需要进一步检查');
  }
}

// 运行测试
if (require.main === module) {
  runMultipleTests().catch(console.error);
}

module.exports = { testLawyerPageAccess, runMultipleTests };