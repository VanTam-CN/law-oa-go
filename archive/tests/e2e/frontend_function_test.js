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

        // 设置视口大小
        await page.setViewport({ width: 1280, height: 720 });

        console.log('1. 测试登录页面...');

        // 访问登录页面
        await page.goto('http://localhost:3003/login');
        await page.waitForSelector('form');

        // 截图
        await page.screenshot({ path: 'login-page.png' });

        // 填写登录表单
        await page.type('input[type="email"]', 'admin@law-oa.com');
        await page.type('input[type="password"]', 'admin123');

        // 点击登录按钮
        await page.click('button[type="submit"]');

        // 等待登录成功
        await page.waitForNavigation({ timeout: 10000 });

        console.log('✅ 登录成功');

        // 截图登录后页面
        await page.screenshot({ path: 'dashboard-page.png' });

        console.log('2. 测试仪表盘页面...');

        // 检查仪表盘元素
        const dashboardElements = await page.$$eval('.ant-card, .ant-statistic, .ant-table', elements => {
            return elements.length;
        });
        console.log(`✅ 仪表盘页面加载完成，发现 ${dashboardElements} 个组件`);

        console.log('3. 测试案件管理页面...');

        // 导航到案件管理页面
        await page.click('a[href="/case"]');
        await page.waitForNavigation();

        // 截图案件管理页面
        await page.screenshot({ path: 'case-management.png' });

        // 检查案件列表
        const caseRows = await page.$$eval('.ant-table-tbody tr', rows => {
            return rows.length;
        });
        console.log(`✅ 案件管理页面加载完成，发现 ${caseRows} 个案件`);

        console.log('4. 测试客户管理页面...');

        // 导航到客户管理页面
        await page.click('a[href="/client"]');
        await page.waitForNavigation();

        // 截图客户管理页面
        await page.screenshot({ path: 'client-management.png' });

        // 检查客户列表
        const clientRows = await page.$$eval('.ant-table-tbody tr', rows => {
            return rows.length;
        });
        console.log(`✅ 客户管理页面加载完成，发现 ${clientRows} 个客户`);

        console.log('5. 测试律师管理页面...');

        // 导航到律师管理页面
        await page.click('a[href="/lawyer"]');
        await page.waitForNavigation();

        // 截图律师管理页面
        await page.screenshot({ path: 'lawyer-management.png' });

        // 检查律师列表
        const lawyerRows = await page.$$eval('.ant-table-tbody tr', rows => {
            return rows.length;
        });
        console.log(`✅ 律师管理页面加载完成，发现 ${lawyerRows} 个律师`);

        console.log('6. 测试响应式设计...');

        // 测试移动端视图
        await page.setViewport({ width: 375, height: 667 });
        await page.screenshot({ path: 'mobile-view.png' });

        // 恢复桌面视图
        await page.setViewport({ width: 1280, height: 720 });

        console.log('✅ 响应式设计测试完成');

        console.log('7. 测试页面交互功能...');

        // 测试搜索功能
        await page.type('input[placeholder*="搜索"], '测试');
        await page.waitForTimeout(1000);

        // 测试筛选功能
        const filters = await page.$$('.ant-select');
        if (filters.length > 0) {
            await filters[0].click();
            await page.waitForTimeout(500);
            await page.screenshot({ path: 'filter-test.png' });
        }

        console.log('✅ 页面交互功能测试完成');

        console.log('8. 测试用户菜单...');

        // 测试用户下拉菜单
        const userMenu = await page.$('.ant-dropdown-trigger');
        if (userMenu) {
            await userMenu.click();
            await page.waitForTimeout(1000);
            await page.screenshot({ path: 'user-menu.png' });
        }

        console.log('✅ 用户菜单测试完成');

        console.log('🎉 前端功能测试全部完成！');

        return {
            success: true,
            results: {
                login: '✅ 成功',
                dashboard: `✅ ${dashboardElements} 个组件`,
                caseManagement: `✅ ${caseRows} 个案件`,
                clientManagement: `✅ ${clientRows} 个客户`,
                lawyerManagement: `✅ ${lawyerRows} 个律师`,
                responsiveDesign: '✅ 成功',
                interactions: '✅ 成功',
                userMenu: '✅ 成功'
            }
        };

    } catch (error) {
        console.error('❌ 测试失败:', error.message);
        return {
            success: false,
            error: error.message
        };
    } finally {
        if (browser) {
            await browser.close();
        }
    }
}

// 执行测试
testFrontendFunctionality().then(result => {
    console.log('\n测试结果总结:');
    console.log(JSON.stringify(result, null, 2));
}).catch(error => {
    console.error('测试执行失败:', error);
});