#!/usr/bin/env node

const puppeteer = require('puppeteer');

// 测试frontend2登录功能
async function testLogin() {
  console.log('🚀 开始测试frontend2登录功能...\n');
  
  let browser;
  try {
    // 启动浏览器
    browser = await puppeteer.launch({
      headless: false, // 设置为true可以无头模式运行
      args: ['--start-maximized'],
      defaultViewport: null
    });
    
    const page = await browser.newPage();
    
    // 1. 访问登录页面
    console.log('🔍 步骤1: 访问登录页面...');
    await page.goto('http://localhost:3002/login', { 
      waitUntil: 'networkidle2',
      timeout: 10000 
    });
    
    // 检查页面标题
    const title = await page.title();
    console.log(`✅ 页面标题: ${title}`);
    
    // 检查页面URL
    const url = page.url();
    console.log(`✅ 当前URL: ${url}`);
    
    // 截图
    await page.screenshot({ path: 'login-page.png' });
    console.log('✅ 登录页面截图已保存');
    
    // 等待页面加载完成
    await page.waitForSelector('form, .login, [type="email"], [type="password"], button', { timeout: 5000 });
    
    // 2. 填写登录表单
    console.log('\n🔍 步骤2: 填写登录表单...');
    
    // 查找并填写邮箱字段
    const emailSelector = 'input[type="email"], input[name="email"], input[placeholder*="邮箱"], input[placeholder*="Email"]';
    const emailFound = await page.$(emailSelector);
    
    if (emailFound) {
      await page.type(emailSelector, 'lawyer1@lawoa.com', { delay: 100 });
      console.log('✅ 已填写邮箱: lawyer1@lawoa.com');
    } else {
      console.log('⚠️  未找到邮箱输入框，尝试通用选择器...');
      await page.evaluate(() => {
        const inputs = document.querySelectorAll('input');
        for (let input of inputs) {
          if (input.type === 'email' || input.name?.toLowerCase().includes('email') || input.placeholder?.toLowerCase().includes('email')) {
            input.value = 'lawyer1@lawoa.com';
            input.dispatchEvent(new Event('input', { bubbles: true }));
            break;
          }
        }
      });
    }
    
    // 查找并填写密码字段
    const passwordSelector = 'input[type="password"], input[name="password"], input[placeholder*="密码"], input[placeholder*="Password"]';
    const passwordFound = await page.$(passwordSelector);
    
    if (passwordFound) {
      await page.type(passwordSelector, 'password', { delay: 100 });
      console.log('✅ 已填写密码: ********');
    } else {
      console.log('⚠️  未找到密码输入框，尝试通用选择器...');
      await page.evaluate(() => {
        const inputs = document.querySelectorAll('input');
        for (let input of inputs) {
          if (input.type === 'password' || input.name?.toLowerCase().includes('password') || input.placeholder?.toLowerCase().includes('password')) {
            input.value = 'password';
            input.dispatchEvent(new Event('input', { bubbles: true }));
            break;
          }
        }
      });
    }
    
    // 截图填写后的表单
    await page.screenshot({ path: 'form-filled.png' });
    console.log('✅ 表单填写完成截图已保存');
    
    // 3. 提交登录表单
    console.log('\n🔍 步骤3: 提交登录表单...');
    
    // 查找并点击登录按钮
    const loginButtonSelector = 'button[type="submit"], button:contains("登录"), button:contains("Login"), .btn-login, .login-btn';
    
    try {
      // 尝试多种方式找到登录按钮
      await Promise.race([
        // 方法1：直接点击提交按钮
        page.click('button[type="submit"]'),
        // 方法2：点击包含登录文本的按钮
        page.evaluate(() => {
          const buttons = document.querySelectorAll('button');
          for (let btn of buttons) {
            if (btn.textContent?.includes('登录') || btn.textContent?.includes('Login') || btn.textContent?.includes('Sign In')) {
              btn.click();
              return true;
            }
          }
          return false;
        }),
        // 方法3：提交表单
        page.evaluate(() => {
          const form = document.querySelector('form');
          if (form) {
            form.submit();
            return true;
          }
          return false;
        })
      ]);
      
      console.log('✅ 登录表单已提交');
      
      // 等待页面跳转
      await page.waitForNavigation({ timeout: 10000 });
      
    } catch (error) {
      console.log('⚠️  表单提交超时，检查当前状态...');
      await page.screenshot({ path: 'login-error.png' });
    }
    
    // 4. 检查登录结果
    console.log('\n🔍 步骤4: 检查登录结果...');
    
    const currentUrl = page.url();
    console.log(`✅ 登录后URL: ${currentUrl}`);
    
    // 检查是否跳转到首页或dashboard
    const isLoggedIn = currentUrl.includes('/dashboard') || 
                     currentUrl === 'http://localhost:3002/' ||
                     currentUrl.includes('/home');
    
    if (isLoggedIn) {
      console.log('✅ 登录成功！已跳转到首页/仪表板');
      
      // 截图登录成功后的页面
      await page.screenshot({ path: 'login-success.png' });
      console.log('✅ 登录成功页面截图已保存');
      
      // 检查页面内容
      const pageContent = await page.evaluate(() => {
        return document.body.innerText.substring(0, 200);
      });
      console.log(`✅ 页面内容预览: ${pageContent}...`);
      
      // 检查是否有用户信息显示
      const hasUserInfo = await page.evaluate(() => {
        return document.querySelector('.user-info, .profile, .avatar, [class*="user"]') !== null;
      });
      
      if (hasUserInfo) {
        console.log('✅ 检测到用户信息显示');
      }
      
      // 检查是否有导航菜单
      const hasNavigation = await page.evaluate(() => {
        return document.querySelector('nav, .navbar, .sidebar, .menu') !== null;
      });
      
      if (hasNavigation) {
        console.log('✅ 检测到导航菜单');
      }
      
      console.log('\n🎉 登录测试通过！');
      
    } else {
      console.log('❌ 登录失败，仍在登录页面或跳转到错误页面');
      
      // 检查是否有错误消息
      const errorMessage = await page.evaluate(() => {
        const errorElements = document.querySelectorAll('.error, .alert-error, .login-error, [class*="error"]');
        return Array.from(errorElements).map(el => el.textContent).join('; ');
      });
      
      if (errorMessage) {
        console.log(`❌ 错误消息: ${errorMessage}`);
      }
      
      // 截图错误状态
      await page.screenshot({ path: 'login-failed.png' });
      console.log('✅ 登录失败页面截图已保存');
      
      // 检查控制台错误
      const consoleErrors = await page.evaluate(() => {
        return console.errors || [];
      });
      
      if (consoleErrors.length > 0) {
        console.log('❌ 控制台错误:', consoleErrors);
      }
    }
    
    // 5. 测试页面功能
    console.log('\n🔍 步骤5: 测试页面功能...');
    
    // 检查是否能正常加载页面内容
    const pageLoaded = await page.evaluate(() => {
      return document.readyState === 'complete';
    });
    
    if (pageLoaded) {
      console.log('✅ 页面加载完成');
    }
    
    // 检查网络请求
    const networkRequests = await page.evaluate(() => {
      return window.performance.getEntriesByType('resource').length;
    });
    
    console.log(`✅ 网络请求数量: ${networkRequests}`);
    
  } catch (error) {
    console.error('❌ 测试过程中出现错误:', error);
    
    // 如果页面仍然存在，截图错误状态
    if (browser && browser.pages().length > 0) {
      const page = browser.pages()[0];
      await page.screenshot({ path: 'test-error.png' });
      console.log('✅ 错误状态截图已保存');
    }
    
  } finally {
    // 关闭浏览器
    if (browser) {
      await browser.close();
    }
  }
  
  console.log('\n📋 测试完成！');
  console.log('📁 生成的截图文件:');
  console.log('   - login-page.png: 登录页面截图');
  console.log('   - form-filled.png: 表单填写完成截图');
  console.log('   - login-success.png: 登录成功页面截图');
  console.log('   - login-failed.png: 登录失败页面截图');
  console.log('   - login-error.png: 错误状态截图');
  console.log('   - test-error.png: 测试错误截图');
}

// 运行测试
testLogin().catch(console.error);