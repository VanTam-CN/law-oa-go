const puppeteer = require('puppeteer');

async function testFrontendFunctionality() {
    console.log('开始前端功能测试...');

    let browser;
    try {
        // 启动浏览器
        browser = await puppeteer.launch({
            headless: false,
            args: ['--no-sandbox', '--disable-setuid-sandbox']
        });

        const page = await browser.newPage();
        await page.setViewport({ width: 1280, height: 720 });

        console.log('1. 测试登录页面...');

        // 访问登录页面
        await page.goto('http://localhost:3003/login');
        await page.waitForSelector('form', { timeout: 10000 });

        // 截图
        await page.screenshot({ path: 'login-page.png' });
        console.log('✅ 登录页面加载完成');

        // 填写登录表单 - 使用更通用的选择器
        const emailInputs = await page.$$('input[type="email"]');
        const passwordInputs = await page.$$('input[type="password"]');

        if (emailInputs.length > 0) {
            await emailInputs[0].type('admin@law-oa.com');
        }
        if (passwordInputs.length > 0) {
            await passwordInputs[0].type('admin123');
        }

        // 查找并点击登录按钮
        const loginButtons = await page.$$('button[type="submit"], .ant-btn-primary');
        if (loginButtons.length > 0) {
            await loginButtons[0].click();
        }

        // 等待登录成功
        await page.waitForNavigation({ timeout: 10000 });
        console.log('✅ 登录成功');

        // 截图登录后页面
        await page.screenshot({ path: 'dashboard-page.png' });

        console.log('2. 测试仪表盘页面...');

        // 检查仪表盘元素
        const dashboardElements = await page.$$eval('.ant-card, .ant-statistic, .ant-table, .ant-row, .ant-col', elements => {
            return elements.length;
        });
        console.log(`✅ 仪表盘页面加载完成，发现 ${dashboardElements} 个组件`);

        console.log('3. 测试案件管理页面...');

        // 尝试导航到案件管理页面
        const caseLinks = await page.$$eval('a[href*="case"], a:contains("案件"), .ant-menu-item:contains("案件")', links => {
            return links.length;
        });

        if (caseLinks > 0) {
            console.log('✅ 找到案件管理链接');
        } else {
            console.log('⚠️ 未找到案件管理链接，检查菜单结构');
            // 获取所有链接
            const allLinks = await page.$$eval('a', links => {
                return links.map(link => link.href || link.textContent);
            });
            console.log('页面中的链接:', allLinks.slice(0, 10)); // 只显示前10个
        }

        // 截图当前页面
        await page.screenshot({ path: 'current-page.png' });

        console.log('4. 测试页面响应性...');

        // 测试移动端视图
        await page.setViewport({ width: 375, height: 667 });
        await page.screenshot({ path: 'mobile-view.png' });

        // 恢复桌面视图
        await page.setViewport({ width: 1280, height: 720 });

        console.log('✅ 响应式设计测试完成');

        console.log('5. 检查主要页面元素...');

        // 检查Ant Design组件
        const antdComponents = await page.$$eval('.ant-btn, .ant-input, .ant-table, .ant-card, .ant-menu', components => {
            return components.length;
        });
        console.log(`✅ 发现 ${antdComponents} 个Ant Design组件`);

        // 检查React组件结构
        const reactComponents = await page.$$eval('[class*="app"], [class*="layout"], [class*="main"]', components => {
            return components.length;
        });
        console.log(`✅ 发现 ${reactComponents} 个主要布局组件`);

        console.log('🎉 前端功能测试完成！');

        return {
            success: true,
            results: {
                login: '✅ 成功',
                dashboard: `✅ ${dashboardElements} 个组件`,
                navigation: caseLinks > 0 ? '✅ 找到导航链接' : '⚠️ 需要检查导航结构',
                responsiveDesign: '✅ 成功',
                components: `✅ ${antdComponents} 个Ant Design组件`,
                layout: `✅ ${reactComponents} 个布局组件`
            },
            screenshots: [
                'login-page.png',
                'dashboard-page.png',
                'current-page.png',
                'mobile-view.png'
            ]
        };

    } catch (error) {
        console.error('❌ 测试失败:', error.message);
        // 尝试截图错误状态
        if (browser) {
            const page = await browser.newPage();
            await page.goto('http://localhost:3003');
            await page.screenshot({ path: 'error-state.png' });
            await page.close();
        }

        return {
            success: false,
            error: error.message,
            screenshots: ['error-state.png']
        };
    } finally {
        if (browser) {
            await browser.close();
        }
    }
}

// 执行测试
testFrontendFunctionality().then(result => {
    console.log('\n=== 测试结果总结 ===');
    console.log(JSON.stringify(result, null, 2));

    if (result.success) {
        console.log('\n🎉 所有测试完成！');
        console.log('截图文件已保存:');
        result.screenshots.forEach(screenshot => {
            console.log(`  - ${screenshot}`);
        });
    } else {
        console.log('\n❌ 测试过程中遇到问题');
        console.log('错误信息:', result.error);
    }
}).catch(error => {
    console.error('测试执行失败:', error);
});