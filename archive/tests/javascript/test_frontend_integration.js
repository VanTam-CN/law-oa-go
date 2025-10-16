#!/usr/bin/env node

/**
 * 前端应用功能和API集成测试脚本
 * 用于验证前端应用与后端API的集成状态
 */

const http = require('http');
const https = require('https');
const fs = require('fs');
const path = require('path');

// 配置
const config = {
    frontend: {
        url: 'http://localhost:3003',
        health: '/index.html'
    },
    backend: {
        url: 'http://localhost:8080',
        health: '/health',
        api: '/api/v1'
    },
    timeout: 5000
};

// 颜色输出
const colors = {
    reset: '\x1b[0m',
    red: '\x1b[31m',
    green: '\x1b[32m',
    yellow: '\x1b[33m',
    blue: '\x1b[34m',
    magenta: '\x1b[35m',
    cyan: '\x1b[36m',
    bright: '\x1b[1m'
};

// 测试结果
const results = {
    frontend: {
        status: 'unknown',
        tests: [],
        errors: []
    },
    backend: {
        status: 'unknown',
        tests: [],
        errors: []
    },
    api: {
        status: 'unknown',
        tests: [],
        errors: []
    }
};

// 工具函数
function log(message, color = 'reset') {
    console.log(`${colors[color]}${message}${colors.reset}`);
}

function logTest(testName, passed, details = '') {
    const status = passed ? '✓' : '✗';
    const color = passed ? 'green' : 'red';
    log(`  ${status} ${testName}`, color);
    if (details) {
        log(`    ${details}`, passed ? 'green' : 'yellow');
    }
    return passed;
}

// HTTP请求工具
function makeRequest(url, options = {}) {
    return new Promise((resolve, reject) => {
        const protocol = url.startsWith('https') ? https : http;
        const reqOptions = {
            timeout: config.timeout,
            headers: {
                'User-Agent': 'Law-OA-Frontend-Test'
            },
            ...options
        };

        const req = protocol.request(url, reqOptions, (res) => {
            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => {
                resolve({
                    statusCode: res.statusCode,
                    headers: res.headers,
                    data: data
                });
            });
        });

        req.on('error', (error) => {
            reject(error);
        });

        req.on('timeout', () => {
            req.destroy();
            reject(new Error('Request timeout'));
        });

        if (options.body) {
            req.write(options.body);
        }

        req.end();
    });
}

// 测试前端应用
async function testFrontend() {
    log('🔍 测试前端应用...', 'cyan');
    results.frontend.status = 'testing';

    try {
        // 测试前端应用是否可访问
        const frontendUrl = config.frontend.url + config.frontend.health;
        const response = await makeRequest(frontendUrl);

        const accessible = logTest(
            '前端应用可访问',
            response.statusCode === 200,
            `状态码: ${response.statusCode}`
        );
        results.frontend.tests.push({ name: '前端应用可访问', passed: accessible });

        // 检查前端资源
        const resourcesToCheck = [
            '/static/js/main.js',
            '/static/css/main.css',
            '/favicon.ico'
        ];

        for (const resource of resourcesToCheck) {
            try {
                const resourceUrl = config.frontend.url + resource;
                const resourceResponse = await makeRequest(resourceUrl);
                const resourceAccessible = logTest(
                    `前端资源 ${resource}`,
                    resourceResponse.statusCode === 200,
                    `状态码: ${resourceResponse.statusCode}`
                );
                results.frontend.tests.push({
                    name: `前端资源 ${resource}`,
                    passed: resourceAccessible
                });
            } catch (error) {
                logTest(`前端资源 ${resource}`, false, error.message);
                results.frontend.tests.push({
                    name: `前端资源 ${resource}`,
                    passed: false
                });
            }
        }

        results.frontend.status = 'completed';
        return accessible;

    } catch (error) {
        log('❌ 前端应用测试失败:', 'red');
        log(`   错误: ${error.message}`, 'yellow');
        results.frontend.status = 'failed';
        results.frontend.errors.push(error.message);
        return false;
    }
}

// 测试后端应用
async function testBackend() {
    log('\n🔍 测试后端应用...', 'cyan');
    results.backend.status = 'testing';

    try {
        // 测试后端健康检查
        const healthUrl = config.backend.url + config.backend.health;
        const healthResponse = await makeRequest(healthUrl);

        const healthOk = logTest(
            '后端健康检查',
            healthResponse.statusCode === 200,
            `状态码: ${healthResponse.statusCode}`
        );
        results.backend.tests.push({ name: '后端健康检查', passed: healthOk });

        // 测试CORS配置
        const corsResponse = await makeRequest(healthUrl, {
            headers: {
                'Origin': config.frontend.url,
                'Access-Control-Request-Method': 'GET',
                'Access-Control-Request-Headers': 'Authorization'
            }
        });

        const corsOk = logTest(
            'CORS配置',
            corsResponse.headers['access-control-allow-origin'] === '*' ||
            corsResponse.headers['access-control-allow-origin'] === config.frontend.url,
            `CORS头部: ${corsResponse.headers['access-control-allow-origin'] || '未设置'}`
        );
        results.backend.tests.push({ name: 'CORS配置', passed: corsOk });

        results.backend.status = 'completed';
        return healthOk;

    } catch (error) {
        log('❌ 后端应用测试失败:', 'red');
        log(`   错误: ${error.message}`, 'yellow');
        results.backend.status = 'failed';
        results.backend.errors.push(error.message);
        return false;
    }
}

// 测试API接口
async function testAPI() {
    log('\n🔍 测试API接口...', 'cyan');
    results.api.status = 'testing';

    try {
        const apiBaseUrl = config.backend.url + config.backend.api;

        // 测试公共API端点
        const publicEndpoints = [
            { path: '/ping', method: 'GET' },
            { path: '/auth/register', method: 'POST' },
            { path: '/auth/login', method: 'POST' }
        ];

        for (const endpoint of publicEndpoints) {
            try {
                const url = apiBaseUrl + endpoint.path;
                const options = {
                    method: endpoint.method,
                    headers: { 'Content-Type': 'application/json' }
                };

                if (endpoint.method === 'POST') {
                    options.body = JSON.stringify({
                        name: 'Test User',
                        email: 'test@example.com',
                        password: 'testpass123'
                    });
                }

                const response = await makeRequest(url, options);
                const endpointOk = logTest(
                    `API端点 ${endpoint.method} ${endpoint.path}`,
                    response.statusCode >= 200 && response.statusCode < 500,
                    `状态码: ${response.statusCode}`
                );
                results.api.tests.push({
                    name: `API端点 ${endpoint.method} ${endpoint.path}`,
                    passed: endpointOk
                });
            } catch (error) {
                logTest(`API端点 ${endpoint.method} ${endpoint.path}`, false, error.message);
                results.api.tests.push({
                    name: `API端点 ${endpoint.method} ${endpoint.path}`,
                    passed: false
                });
            }
        }

        results.api.status = 'completed';
        return true;

    } catch (error) {
        log('❌ API接口测试失败:', 'red');
        log(`   错误: ${error.message}`, 'yellow');
        results.api.status = 'failed';
        results.api.errors.push(error.message);
        return false;
    }
}

// 生成测试报告
function generateReport() {
    log('\n📊 测试报告', 'cyan');
    log('=' .repeat(50), 'bright');

    // 计算总体状态
    const allTests = [
        ...results.frontend.tests,
        ...results.backend.tests,
        ...results.api.tests
    ];

    const passedTests = allTests.filter(test => test.passed).length;
    const totalTests = allTests.length;
    const successRate = totalTests > 0 ? (passedTests / totalTests * 100).toFixed(1) : 0;

    log(`总体成功率: ${passedTests}/${totalTests} (${successRate}%)`,
        passedTests === totalTests ? 'green' : 'yellow');

    // 详细报告
    log('\n📱 前端应用测试', 'cyan');
    log(`状态: ${results.frontend.status}`,
        results.frontend.status === 'completed' ? 'green' : 'red');

    if (results.frontend.errors.length > 0) {
        log('错误:', 'red');
        results.frontend.errors.forEach(error => {
            log(`  - ${error}`, 'red');
        });
    }

    log('\n🔧 后端应用测试', 'cyan');
    log(`状态: ${results.backend.status}`,
        results.backend.status === 'completed' ? 'green' : 'red');

    if (results.backend.errors.length > 0) {
        log('错误:', 'red');
        results.backend.errors.forEach(error => {
            log(`  - ${error}`, 'red');
        });
    }

    log('\n🔗 API接口测试', 'cyan');
    log(`状态: ${results.api.status}`,
        results.api.status === 'completed' ? 'green' : 'red');

    if (results.api.errors.length > 0) {
        log('错误:', 'red');
        results.api.errors.forEach(error => {
            log(`  - ${error}`, 'red');
        });
    }

    // 保存测试结果
    const reportPath = path.join(__dirname, 'frontend-test-report.json');
    const reportData = {
        timestamp: new Date().toISOString(),
        config: config,
        results: results,
        summary: {
            totalTests,
            passedTests,
            successRate: parseFloat(successRate)
        }
    };

    fs.writeFileSync(reportPath, JSON.stringify(reportData, null, 2));
    log(`\n📄 详细测试报告已保存到: ${reportPath}`, 'blue');

    return reportData;
}

// 主函数
async function main() {
    log('🚀 开始前端应用功能和API集成测试', 'bright');
    log('测试配置:', 'cyan');
    log(`  前端URL: ${config.frontend.url}`, 'blue');
    log(`  后端URL: ${config.backend.url}`, 'blue');
    log(`  超时时间: ${config.timeout}ms`, 'blue');

    // 运行测试
    const frontendOk = await testFrontend();
    const backendOk = await testBackend();
    const apiOk = await testAPI();

    // 生成报告
    const report = generateReport();

    // 总体结果
    log('\n' + '='.repeat(50), 'bright');
    const overallSuccess = frontendOk && backendOk && apiOk;
    log(`🎯 总体测试结果: ${overallSuccess ? '通过' : '失败'}`,
        overallSuccess ? 'green' : 'red');

    if (!overallSuccess) {
        log('\n💡 建议:', 'yellow');
        if (!frontendOk) {
            log('  - 确保前端应用正在运行在端口 3003', 'yellow');
            log('  - 检查前端应用的构建状态', 'yellow');
        }
        if (!backendOk) {
            log('  - 确保后端应用正在运行在端口 8080', 'yellow');
            log('  - 检查后端应用的启动状态', 'yellow');
        }
        if (!apiOk) {
            log('  - 检查API路由配置', 'yellow');
            log('  - 确认数据库连接正常', 'yellow');
        }
    }

    process.exit(overallSuccess ? 0 : 1);
}

// 运行主函数
if (require.main === module) {
    main().catch(error => {
        log('💥 测试运行失败:', 'red');
        log(`错误: ${error.message}`, 'yellow');
        process.exit(1);
    });
}

module.exports = {
    testFrontend,
    testBackend,
    testAPI,
    generateReport,
    config
};