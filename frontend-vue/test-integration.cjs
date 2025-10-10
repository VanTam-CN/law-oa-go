#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// 测试结果记录
const testResults = {
    startTime: new Date().toISOString(),
    tests: [],
    summary: {
        total: 0,
        passed: 0,
        failed: 0,
        skipped: 0
    }
};

// 测试工具函数
const TestUtils = {
    // 记录测试结果
    recordTest(testName, status, details = null) {
        const result = {
            name: testName,
            status: status, // 'passed', 'failed', 'skipped'
            timestamp: new Date().toISOString(),
            details: details
        };

        testResults.tests.push(result);
        testResults.summary.total++;

        if (status === 'passed') {
            testResults.summary.passed++;
        } else if (status === 'failed') {
            testResults.summary.failed++;
        } else {
            testResults.summary.skipped++;
        }

        console.log(`[${status.toUpperCase()}] ${testName}`);
        if (details) {
            console.log(`  Details: ${details}`);
        }
    },

    // 检查文件是否存在
    checkFileExists(filePath) {
        return fs.existsSync(filePath);
    },

    // 检查文件内容是否包含特定文本
    checkFileContains(filePath, searchText) {
        if (!this.checkFileExists(filePath)) return false;
        const content = fs.readFileSync(filePath, 'utf8');
        return content.includes(searchText);
    },

    // 检查JSON文件是否有效
    checkValidJson(filePath) {
        if (!this.checkFileExists(filePath)) return false;
        try {
            JSON.parse(fs.readFileSync(filePath, 'utf8'));
            return true;
        } catch (e) {
            return false;
        }
    },

    // 检查目录结构
    checkDirectoryStructure(basePath, requiredDirs) {
        return requiredDirs.every(dir =>
            fs.existsSync(path.join(basePath, dir))
        );
    },

    // 生成测试报告
    generateReport() {
        const reportPath = path.join(__dirname, 'test-report.json');
        fs.writeFileSync(reportPath, JSON.stringify(testResults, null, 2));

        console.log('\n=== 测试完成 ===');
        console.log(`总测试数: ${testResults.summary.total}`);
        console.log(`通过: ${testResults.summary.passed}`);
        console.log(`失败: ${testResults.summary.failed}`);
        console.log(`跳过: ${testResults.summary.skipped}`);
        console.log(`成功率: ${((testResults.summary.passed / testResults.summary.total) * 100).toFixed(2)}%`);
        console.log(`测试报告已保存到: ${reportPath}`);

        return testResults;
    }
};

// 测试套件
const TestSuite = {
    // 项目结构测试
    testProjectStructure() {
        console.log('\n=== 测试项目结构 ===');

        const requiredDirs = [
            'src', 'src/pages', 'src/components', 'src/services',
            'src/hooks', 'src/layouts', 'src/utils', 'src/assets',
            'api', 'pages', 'styles', 'utils'
        ];

        const structureOk = TestUtils.checkDirectoryStructure('.', requiredDirs);
        TestUtils.recordTest('项目目录结构', structureOk ? 'passed' : 'failed',
            structureOk ? '所有必需目录存在' : '缺少必需目录');

        // 检查关键文件
        const keyFiles = [
            'package.json',
            'vite.config.ts',
            'tsconfig.json',
            'index.html',
            'src/App.tsx'
        ];

        keyFiles.forEach(file => {
            const exists = TestUtils.checkFileExists(file);
            TestUtils.recordTest(`关键文件: ${file}`, exists ? 'passed' : 'failed',
                exists ? '文件存在' : '文件不存在');
        });
    },

    // 依赖包测试
    testDependencies() {
        console.log('\n=== 测试依赖包 ===');

        const packageJson = TestUtils.checkFileExists('package.json');
        if (packageJson) {
            const packageData = JSON.parse(fs.readFileSync('package.json', 'utf8'));

            const requiredDeps = [
                'react', 'react-dom', 'antd', 'axios',
                'react-router-dom', 'typescript', 'vite'
            ];

            requiredDeps.forEach(dep => {
                const hasDep = packageData.dependencies && packageData.dependencies[dep] ||
                             packageData.devDependencies && packageData.devDependencies[dep];
                TestUtils.recordTest(`依赖包: ${dep}`, hasDep ? 'passed' : 'failed',
                    hasDep ? `版本: ${packageData.dependencies?.[dep] || packageData.devDependencies?.[dep]}` : '依赖缺失');
            });
        } else {
            TestUtils.recordTest('package.json', 'failed', '文件不存在');
        }
    },

    // 页面组件测试
    testPageComponents() {
        console.log('\n=== 测试页面组件 ===');

        const pages = [
            'src/pages/auth/Login.tsx',
            'src/pages/dashboard/Dashboard.tsx',
            'src/pages/case/CaseManagement.tsx',
            'src/pages/client/ClientManagement.tsx',
            'src/pages/lawyer/LawyerManagement.tsx',
            'src/pages/finance/FinanceManagement.tsx',
            'src/pages/approval/ApprovalList.tsx',
            'src/pages/conflict/ConflictCheck.tsx',
            'src/pages/tools/ToolsPage.tsx'
        ];

        pages.forEach(page => {
            const exists = TestUtils.checkFileExists(page);
            TestUtils.recordTest(`页面组件: ${path.basename(page)}`, exists ? 'passed' : 'failed',
                exists ? '组件文件存在' : '组件文件不存在');
        });
    },

    // 服务层测试
    testServices() {
        console.log('\n=== 测试服务层 ===');

        const services = [
            'src/services/auth.ts',
            'src/services/http.ts',
            'src/services/api.ts'
        ];

        services.forEach(service => {
            const exists = TestUtils.checkFileExists(service);
            TestUtils.recordTest(`服务文件: ${path.basename(service)}`, exists ? 'passed' : 'failed',
                exists ? '服务文件存在' : '服务文件不存在');
        });

        // 检查服务内容
        if (TestUtils.checkFileExists('src/services/auth.ts')) {
            const hasLogin = TestUtils.checkFileContains('src/services/auth.ts', 'login');
            const hasLogout = TestUtils.checkFileContains('src/services/auth.ts', 'logout');

            TestUtils.recordTest('登录服务', hasLogin ? 'passed' : 'failed', hasLogin ? '包含登录功能' : '缺少登录功能');
            TestUtils.recordTest('登出服务', hasLogout ? 'passed' : 'failed', hasLogout ? '包含登出功能' : '缺少登出功能');
        }
    },

    // 路由配置测试
    testRouting() {
        console.log('\n=== 测试路由配置 ===');

        const appTsx = TestUtils.checkFileExists('src/App.tsx');
        if (appTsx) {
            const hasRouter = TestUtils.checkFileContains('src/App.tsx', 'BrowserRouter');
            const hasRoutes = TestUtils.checkFileContains('src/App.tsx', 'Routes');
            const hasRoute = TestUtils.checkFileContains('src/App.tsx', 'Route');

            TestUtils.recordTest('路由器配置', hasRouter ? 'passed' : 'failed', hasRouter ? '包含BrowserRouter' : '缺少BrowserRouter');
            TestUtils.recordTest('路由配置', hasRoutes ? 'passed' : 'failed', hasRoutes ? '包含Routes配置' : '缺少Routes配置');
            TestUtils.recordTest('路由定义', hasRoute ? 'passed' : 'failed', hasRoute ? '包含Route定义' : '缺少Route定义');
        } else {
            TestUtils.recordTest('App.tsx', 'failed', '文件不存在');
        }
    },

    // 状态管理测试
    testStateManagement() {
        console.log('\n=== 测试状态管理 ===');

        const authContext = TestUtils.checkFileExists('src/context/AuthContext.tsx');
        const authHook = TestUtils.checkFileExists('src/hooks/useAuth.ts');

        TestUtils.recordTest('认证上下文', authContext ? 'passed' : 'failed', authContext ? 'AuthContext存在' : 'AuthContext不存在');
        TestUtils.recordTest('认证Hook', authHook ? 'passed' : 'failed', authHook ? 'useAuth存在' : 'useAuth不存在');

        if (authContext) {
            const hasAuthProvider = TestUtils.checkFileContains('src/context/AuthContext.tsx', 'AuthProvider');
            TestUtils.recordTest('AuthProvider', hasAuthProvider ? 'passed' : 'failed',
                hasAuthProvider ? '包含AuthProvider' : '缺少AuthProvider');
        }
    },

    // API配置测试
    testApiConfig() {
        console.log('\n=== 测试API配置 ===');

        const httpService = TestUtils.checkFileExists('src/services/http.ts');
        const apiConfig = TestUtils.checkFileExists('src/config/api.ts');

        TestUtils.recordTest('HTTP服务', httpService ? 'passed' : 'failed', httpService ? 'HTTP服务存在' : 'HTTP服务不存在');
        TestUtils.recordTest('API配置', apiConfig ? 'passed' : 'failed', apiConfig ? 'API配置存在' : 'API配置不存在');

        if (httpService) {
            const hasAxios = TestUtils.checkFileContains('src/services/http.ts', 'axios');
            const hasInterceptors = TestUtils.checkFileContains('src/services/http.ts', 'interceptors');

            TestUtils.recordTest('Axios配置', hasAxios ? 'passed' : 'failed', hasAxios ? '包含Axios配置' : '缺少Axios配置');
            TestUtils.recordTest('拦截器配置', hasInterceptors ? 'passed' : 'failed',
                hasInterceptors ? '包含拦截器配置' : '缺少拦截器配置');
        }
    },

    // 样式配置测试
    testStyling() {
        console.log('\n=== 测试样式配置 ===');

        const globalStyles = TestUtils.checkFileExists('src/assets/styles/design-tokens.css');
        const loginStyles = TestUtils.checkFileExists('src/pages/auth/Login.less');

        TestUtils.recordTest('设计令牌', globalStyles ? 'passed' : 'failed', globalStyles ? '设计令牌存在' : '设计令牌不存在');
        TestUtils.recordTest('登录页样式', loginStyles ? 'passed' : 'failed', loginStyles ? '登录页样式存在' : '登录页样式不存在');

        if (globalStyles) {
            const hasCssVars = TestUtils.checkFileContains('src/assets/styles/design-tokens.css', '--');
            TestUtils.recordTest('CSS变量', hasCssVars ? 'passed' : 'failed',
                hasCssVars ? '包含CSS变量' : '缺少CSS变量');
        }
    },

    // 构建配置测试
    testBuildConfig() {
        console.log('\n=== 测试构建配置 ===');

        const viteConfig = TestUtils.checkFileExists('vite.config.ts');
        const tsConfig = TestUtils.checkFileExists('tsconfig.json');

        TestUtils.recordTest('Vite配置', viteConfig ? 'passed' : 'failed', viteConfig ? 'Vite配置存在' : 'Vite配置不存在');
        TestUtils.recordTest('TypeScript配置', tsConfig ? 'passed' : 'failed', tsConfig ? 'TS配置存在' : 'TS配置不存在');

        if (viteConfig) {
            const hasReactPlugin = TestUtils.checkFileContains('vite.config.ts', '@vitejs/plugin-react');
            const hasProxyConfig = TestUtils.checkFileContains('vite.config.ts', 'proxy');

            TestUtils.recordTest('React插件', hasReactPlugin ? 'passed' : 'failed',
                hasReactPlugin ? '包含React插件' : '缺少React插件');
            TestUtils.recordTest('代理配置', hasProxyConfig ? 'passed' : 'failed',
                hasProxyConfig ? '包含代理配置' : '缺少代理配置');
        }
    },

    // PRD功能覆盖测试
    testPRDFeatureCoverage() {
        console.log('\n=== 测试PRD功能覆盖 ===');

        const prdFeatures = [
            { file: 'src/pages/auth/Login.tsx', feature: '用户认证' },
            { file: 'src/pages/dashboard/Dashboard.tsx', feature: '仪表盘' },
            { file: 'src/pages/case/CaseManagement.tsx', feature: '案件管理' },
            { file: 'src/pages/client/ClientManagement.tsx', feature: '客户管理' },
            { file: 'src/pages/lawyer/LawyerManagement.tsx', feature: '律师管理' },
            { file: 'src/pages/finance/FinanceManagement.tsx', feature: '财务管理' },
            { file: 'src/pages/approval/ApprovalList.tsx', feature: '审批流程' },
            { file: 'src/pages/conflict/ConflictCheck.tsx', feature: '利益冲突检查' },
            { file: 'src/pages/tools/ToolsPage.tsx', feature: '法律工具' },
            { file: 'src/pages/file/FileManagement.tsx', feature: '文件管理' },
            { file: 'src/pages/user/UserManagement.tsx', feature: '用户管理' }
        ];

        prdFeatures.forEach(({ file, feature }) => {
            const exists = TestUtils.checkFileExists(file);
            TestUtils.recordTest(`PRD功能: ${feature}`, exists ? 'passed' : 'failed',
                exists ? '功能已实现' : '功能未实现');
        });
    },

    // 响应式设计测试
    testResponsiveDesign() {
        console.log('\n=== 测试响应式设计 ===');

        const designTokens = TestUtils.checkFileExists('src/assets/styles/design-tokens.css');
        if (designTokens) {
            const hasMediaQueries = TestUtils.checkFileContains('src/assets/styles/design-tokens.css', '@media');
            const hasBreakpoints = TestUtils.checkFileContains('src/assets/styles/design-tokens.css', 'breakpoint');

            TestUtils.recordTest('媒体查询', hasMediaQueries ? 'passed' : 'failed',
                hasMediaQueries ? '包含媒体查询' : '缺少媒体查询');
            TestUtils.recordTest('断点配置', hasBreakpoints ? 'passed' : 'failed',
                hasBreakpoints ? '包含断点配置' : '缺少断点配置');
        }
    },

    // 安全性测试
    testSecurity() {
        console.log('\n=== 测试安全性 ===');

        const authContext = TestUtils.checkFileExists('src/context/AuthContext.tsx');
        if (authContext) {
            const hasTokenManagement = TestUtils.checkFileContains('src/context/AuthContext.tsx', 'token');
            const hasPermissionCheck = TestUtils.checkFileContains('src/context/AuthContext.tsx', 'permission');

            TestUtils.recordTest('Token管理', hasTokenManagement ? 'passed' : 'failed',
                hasTokenManagement ? '包含Token管理' : '缺少Token管理');
            TestUtils.recordTest('权限检查', hasPermissionCheck ? 'passed' : 'failed',
                hasPermissionCheck ? '包含权限检查' : '缺少权限检查');
        }

        const httpService = TestUtils.checkFileExists('src/services/http.ts');
        if (httpService) {
            const hasAuthHeader = TestUtils.checkFileContains('src/services/http.ts', 'Authorization');
            TestUtils.recordTest('认证头', hasAuthHeader ? 'passed' : 'failed',
                hasAuthHeader ? '包含认证头设置' : '缺少认证头设置');
        }
    }
};

// 执行所有测试
function runAllTests() {
    console.log('=== 开始律所OA系统前端集成测试 ===');
    console.log(`测试开始时间: ${testResults.startTime}`);

    TestSuite.testProjectStructure();
    TestSuite.testDependencies();
    TestSuite.testPageComponents();
    TestSuite.testServices();
    TestSuite.testRouting();
    TestSuite.testStateManagement();
    TestSuite.testApiConfig();
    TestSuite.testStyling();
    TestSuite.testBuildConfig();
    TestSuite.testPRDFeatureCoverage();
    TestSuite.testResponsiveDesign();
    TestSuite.testSecurity();

    return TestUtils.generateReport();
}

// 主函数
function main() {
    try {
        const results = runAllTests();
        process.exit(results.summary.failed > 0 ? 1 : 0);
    } catch (error) {
        console.error('测试执行失败:', error);
        process.exit(1);
    }
}

// 如果直接运行此脚本
if (require.main === module) {
    main();
}

module.exports = { TestUtils, TestSuite, runAllTests };