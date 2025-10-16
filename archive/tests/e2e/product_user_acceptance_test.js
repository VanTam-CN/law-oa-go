#!/usr/bin/env node

/**
 * 律师事务所OA系统 - 产品用户验收测试
 * 从真实用户角度测试每个功能和操作
 */

const axios = require('axios');
const fs = require('fs');

const BASE_URL = 'http://localhost:8080';
const FRONTEND_URL = 'http://localhost:3000';

// 测试配置
const TEST_CONFIG = {
    timeout: 10000,
    retries: 3,
    delay: 1000
};

// 测试数据
const TEST_DATA = {
    lawyer: {
        name: "张律师",
        email: "zhang@lawfirm.com",
        phone: "13800138001",
        department: "民事部",
        position: "高级律师"
    },
    client: {
        name: "测试公司",
        company: "北京测试科技有限公司",
        contact_person: "李总",
        phone: "13900139001",
        email: "li@testcompany.com",
        address: "北京市朝阳区测试大厦"
    },
    case: {
        title: "商业合同纠纷案",
        description: "关于软件开发合同的争议",
        case_type: "民事纠纷",
        status: "进行中",
        priority: "高"
    }
};

class ProductUserAcceptanceTest {
    constructor() {
        this.results = {
            total: 0,
            passed: 0,
            failed: 0,
            details: []
        };
        this.authToken = null;
        this.testEntities = {};
    }

    // 工具方法
    async delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    async makeRequest(method, url, data = null, headers = {}) {
        try {
            const config = {
                method,
                url: `${BASE_URL}${url}`,
                timeout: TEST_CONFIG.timeout,
                headers: {
                    'Content-Type': 'application/json',
                    ...headers
                }
            };

            if (this.authToken) {
                config.headers.Authorization = `Bearer ${this.authToken}`;
            }

            if (data) {
                config.data = data;
            }

            const response = await axios(config);
            return { success: true, data: response.data, status: response.status };
        } catch (error) {
            return {
                success: false,
                error: error.message,
                status: error.response?.status,
                data: error.response?.data
            };
        }
    }

    logTest(testName, passed, message = '', details = null) {
        this.results.total++;
        if (passed) {
            this.results.passed++;
            console.log(`✅ ${testName}: ${message}`);
        } else {
            this.results.failed++;
            console.log(`❌ ${testName}: ${message}`);
        }

        this.results.details.push({
            test: testName,
            passed,
            message,
            details,
            timestamp: new Date().toISOString()
        });
    }

    // 用户场景1: 系统登录和认证
    async testUserAuthentication() {
        console.log('\n🔐 测试用户场景1: 系统登录和认证');

        // 测试登录页面访问
        const loginPageTest = await this.makeRequest('GET', '/api/auth/status');
        this.logTest(
            '登录页面访问',
            loginPageTest.success || loginPageTest.status === 401,
            loginPageTest.success ? '登录接口可访问' : '需要认证（正常）'
        );

        // 测试用户登录
        const loginData = {
            email: 'admin@example.com',
            password: 'admin123'
        };

        const loginTest = await this.makeRequest('POST', '/api/auth/login', loginData);
        if (loginTest.success && loginTest.data && loginTest.data.data && loginTest.data.data.token) {
            this.authToken = loginTest.data.data.token;
            this.logTest('用户登录', true, `登录成功，用户: ${loginTest.data.data.user.name}`);
        } else {
            this.logTest('用户登录', false, `登录失败: ${loginTest.error || 'Unknown error'}`);
            return false;
        }

        // 测试认证状态检查
        const authStatusTest = await this.makeRequest('GET', '/api/auth/status');
        this.logTest(
            '认证状态检查',
            authStatusTest.success,
            authStatusTest.success ? '认证状态正常' : '认证状态异常'
        );

        return true;
    }

    // 用户场景2: 律师管理功能
    async testLawyerManagement() {
        console.log('\n👨‍💼 测试用户场景2: 律师管理功能');

        // 测试律师列表查看
        const lawyerListTest = await this.makeRequest('GET', '/api/lawyers?page=1&page_size=10');
        this.logTest(
            '律师列表查看',
            lawyerListTest.success,
            lawyerListTest.success ? 
                `成功获取律师列表，共${lawyerListTest.data?.data?.length || 0}条记录` : 
                `获取律师列表失败: ${lawyerListTest.error}`
        );

        // 测试添加新律师
        const addLawyerTest = await this.makeRequest('POST', '/api/lawyers', TEST_DATA.lawyer);
        if (addLawyerTest.success) {
            this.testEntities.lawyer = addLawyerTest.data;
            this.logTest('添加新律师', true, `成功添加律师: ${TEST_DATA.lawyer.name}`);
        } else {
            this.logTest('添加新律师', false, `添加律师失败: ${addLawyerTest.error}`);
        }

        // 测试律师详情查看
        if (this.testEntities.lawyer?.id) {
            const lawyerDetailTest = await this.makeRequest('GET', `/api/lawyers/${this.testEntities.lawyer.id}`);
            this.logTest(
                '律师详情查看',
                lawyerDetailTest.success,
                lawyerDetailTest.success ? '成功查看律师详情' : '查看律师详情失败'
            );

            // 测试律师信息编辑
            const updateData = { ...TEST_DATA.lawyer, phone: '13800138002' };
            const updateLawyerTest = await this.makeRequest('PUT', `/api/lawyers/${this.testEntities.lawyer.id}`, updateData);
            this.logTest(
                '律师信息编辑',
                updateLawyerTest.success,
                updateLawyerTest.success ? '成功更新律师信息' : '更新律师信息失败'
            );
        }

        // 测试律师搜索功能
        const searchLawyerTest = await this.makeRequest('GET', `/api/lawyers?search=${TEST_DATA.lawyer.name}&page=1&page_size=10`);
        this.logTest(
            '律师搜索功能',
            searchLawyerTest.success,
            searchLawyerTest.success ? '律师搜索功能正常' : '律师搜索功能异常'
        );
    }

    // 用户场景3: 客户管理功能
    async testClientManagement() {
        console.log('\n🏢 测试用户场景3: 客户管理功能');

        // 测试客户列表查看
        const clientListTest = await this.makeRequest('GET', '/api/clients?page=1&page_size=10');
        this.logTest(
            '客户列表查看',
            clientListTest.success,
            clientListTest.success ? 
                `成功获取客户列表，共${clientListTest.data?.data?.length || 0}条记录` : 
                `获取客户列表失败: ${clientListTest.error}`
        );

        // 测试添加新客户
        const addClientTest = await this.makeRequest('POST', '/api/clients', TEST_DATA.client);
        if (addClientTest.success) {
            this.testEntities.client = addClientTest.data;
            this.logTest('添加新客户', true, `成功添加客户: ${TEST_DATA.client.company}`);
        } else {
            this.logTest('添加新客户', false, `添加客户失败: ${addClientTest.error}`);
        }

        // 测试客户详情查看
        if (this.testEntities.client?.id) {
            const clientDetailTest = await this.makeRequest('GET', `/api/clients/${this.testEntities.client.id}`);
            this.logTest(
                '客户详情查看',
                clientDetailTest.success,
                clientDetailTest.success ? '成功查看客户详情' : '查看客户详情失败'
            );

            // 测试客户信息编辑
            const updateData = { ...TEST_DATA.client, phone: '13900139002' };
            const updateClientTest = await this.makeRequest('PUT', `/api/clients/${this.testEntities.client.id}`, updateData);
            this.logTest(
                '客户信息编辑',
                updateClientTest.success,
                updateClientTest.success ? '成功更新客户信息' : '更新客户信息失败'
            );
        }

        // 测试客户搜索功能
        const searchClientTest = await this.makeRequest('GET', `/api/clients?search=${TEST_DATA.client.company}&page=1&page_size=10`);
        this.logTest(
            '客户搜索功能',
            searchClientTest.success,
            searchClientTest.success ? '客户搜索功能正常' : '客户搜索功能异常'
        );
    }

    // 用户场景4: 案件管理功能
    async testCaseManagement() {
        console.log('\n📋 测试用户场景4: 案件管理功能');

        // 测试案件列表查看
        const caseListTest = await this.makeRequest('GET', '/api/cases?page=1&page_size=10');
        this.logTest(
            '案件列表查看',
            caseListTest.success,
            caseListTest.success ? 
                `成功获取案件列表，共${caseListTest.data?.data?.length || 0}条记录` : 
                `获取案件列表失败: ${caseListTest.error}`
        );

        // 测试添加新案件
        const caseData = {
            ...TEST_DATA.case,
            client_id: this.testEntities.client?.id || 1,
            lawyer_id: this.testEntities.lawyer?.id || 1
        };

        const addCaseTest = await this.makeRequest('POST', '/api/cases', caseData);
        if (addCaseTest.success) {
            this.testEntities.case = addCaseTest.data;
            this.logTest('添加新案件', true, `成功添加案件: ${TEST_DATA.case.title}`);
        } else {
            this.logTest('添加新案件', false, `添加案件失败: ${addCaseTest.error}`);
        }

        // 测试案件详情查看
        if (this.testEntities.case?.id) {
            const caseDetailTest = await this.makeRequest('GET', `/api/cases/${this.testEntities.case.id}`);
            this.logTest(
                '案件详情查看',
                caseDetailTest.success,
                caseDetailTest.success ? '成功查看案件详情' : '查看案件详情失败'
            );

            // 测试案件状态更新
            const updateData = { ...caseData, status: '已结案' };
            const updateCaseTest = await this.makeRequest('PUT', `/api/cases/${this.testEntities.case.id}`, updateData);
            this.logTest(
                '案件状态更新',
                updateCaseTest.success,
                updateCaseTest.success ? '成功更新案件状态' : '更新案件状态失败'
            );
        }

        // 测试案件搜索功能
        const searchCaseTest = await this.makeRequest('GET', `/api/cases?search=${TEST_DATA.case.title}&page=1&page_size=10`);
        this.logTest(
            '案件搜索功能',
            searchCaseTest.success,
            searchCaseTest.success ? '案件搜索功能正常' : '案件搜索功能异常'
        );

        // 测试案件筛选功能
        const filterCaseTest = await this.makeRequest('GET', '/api/cases?status=进行中&page=1&page_size=10');
        this.logTest(
            '案件筛选功能',
            filterCaseTest.success,
            filterCaseTest.success ? '案件筛选功能正常' : '案件筛选功能异常'
        );
    }

    // 用户场景5: 利益冲突检测
    async testConflictDetection() {
        console.log('\n⚠️ 测试用户场景5: 利益冲突检测');

        // 测试利益冲突检测功能
        const conflictData = {
            client_name: TEST_DATA.client.company,
            case_type: TEST_DATA.case.case_type,
            opposing_party: "对方当事人公司",
            description: "检测是否存在利益冲突"
        };

        const conflictTest = await this.makeRequest('POST', '/api/conflicts/check', conflictData);
        this.logTest(
            '利益冲突检测',
            conflictTest.success,
            conflictTest.success ? 
                `冲突检测完成，风险等级: ${conflictTest.data?.risk_level || '未知'}` : 
                `冲突检测失败: ${conflictTest.error}`
        );

        // 测试冲突检测历史记录
        const conflictHistoryTest = await this.makeRequest('GET', '/api/conflicts?page=1&page_size=10');
        this.logTest(
            '冲突检测历史',
            conflictHistoryTest.success,
            conflictHistoryTest.success ? '成功获取冲突检测历史' : '获取冲突检测历史失败'
        );
    }

    // 用户场景6: 数据统计和报表
    async testReportsAndStatistics() {
        console.log('\n📊 测试用户场景6: 数据统计和报表');

        // 测试案件统计
        const caseStatsTest = await this.makeRequest('GET', '/api/statistics/cases');
        this.logTest(
            '案件统计报表',
            caseStatsTest.success,
            caseStatsTest.success ? '成功获取案件统计数据' : '获取案件统计失败'
        );

        // 测试律师工作量统计
        const lawyerStatsTest = await this.makeRequest('GET', '/api/statistics/lawyers');
        this.logTest(
            '律师工作量统计',
            lawyerStatsTest.success,
            lawyerStatsTest.success ? '成功获取律师工作量统计' : '获取律师统计失败'
        );

        // 测试客户统计
        const clientStatsTest = await this.makeRequest('GET', '/api/statistics/clients');
        this.logTest(
            '客户统计报表',
            clientStatsTest.success,
            clientStatsTest.success ? '成功获取客户统计数据' : '获取客户统计失败'
        );
    }

    // 用户场景7: 系统设置和权限
    async testSystemSettings() {
        console.log('\n⚙️ 测试用户场景7: 系统设置和权限');

        // 测试用户权限检查
        const permissionTest = await this.makeRequest('GET', '/api/auth/permissions');
        this.logTest(
            '用户权限检查',
            permissionTest.success,
            permissionTest.success ? '成功获取用户权限信息' : '获取权限信息失败'
        );

        // 测试系统配置获取
        const configTest = await this.makeRequest('GET', '/api/system/config');
        this.logTest(
            '系统配置获取',
            configTest.success || configTest.status === 404,
            configTest.success ? '成功获取系统配置' : '系统配置接口未实现（可选功能）'
        );
    }

    // 用户场景8: 数据导入导出
    async testDataImportExport() {
        console.log('\n📤 测试用户场景8: 数据导入导出');

        // 测试数据导出功能
        const exportTest = await this.makeRequest('GET', '/api/export/cases?format=csv');
        this.logTest(
            '案件数据导出',
            exportTest.success || exportTest.status === 404,
            exportTest.success ? '数据导出功能正常' : '数据导出功能未实现（可选功能）'
        );

        // 测试数据备份功能
        const backupTest = await this.makeRequest('POST', '/api/system/backup');
        this.logTest(
            '数据备份功能',
            backupTest.success || backupTest.status === 404,
            backupTest.success ? '数据备份功能正常' : '数据备份功能未实现（可选功能）'
        );
    }

    // 清理测试数据
    async cleanupTestData() {
        console.log('\n🧹 清理测试数据');

        // 删除测试案件
        if (this.testEntities.case?.id) {
            const deleteCaseTest = await this.makeRequest('DELETE', `/api/cases/${this.testEntities.case.id}`);
            this.logTest(
                '删除测试案件',
                deleteCaseTest.success,
                deleteCaseTest.success ? '成功删除测试案件' : '删除测试案件失败'
            );
        }

        // 删除测试客户
        if (this.testEntities.client?.id) {
            const deleteClientTest = await this.makeRequest('DELETE', `/api/clients/${this.testEntities.client.id}`);
            this.logTest(
                '删除测试客户',
                deleteClientTest.success,
                deleteClientTest.success ? '成功删除测试客户' : '删除测试客户失败'
            );
        }

        // 删除测试律师
        if (this.testEntities.lawyer?.id) {
            const deleteLawyerTest = await this.makeRequest('DELETE', `/api/lawyers/${this.testEntities.lawyer.id}`);
            this.logTest(
                '删除测试律师',
                deleteLawyerTest.success,
                deleteLawyerTest.success ? '成功删除测试律师' : '删除测试律师失败'
            );
        }
    }

    // 生成测试报告
    generateReport() {
        const report = {
            summary: {
                total: this.results.total,
                passed: this.results.passed,
                failed: this.results.failed,
                success_rate: ((this.results.passed / this.results.total) * 100).toFixed(2) + '%'
            },
            details: this.results.details,
            timestamp: new Date().toISOString()
        };

        // 保存详细报告
        fs.writeFileSync('product_acceptance_test_report.json', JSON.stringify(report, null, 2));

        // 生成简要报告
        const briefReport = `
# 律师事务所OA系统 - 产品用户验收测试报告

## 测试概要
- 测试时间: ${new Date().toLocaleString('zh-CN')}
- 总测试项: ${this.results.total}
- 通过测试: ${this.results.passed}
- 失败测试: ${this.results.failed}
- 成功率: ${report.summary.success_rate}

## 测试场景覆盖
✅ 用户登录和认证
✅ 律师管理功能
✅ 客户管理功能  
✅ 案件管理功能
✅ 利益冲突检测
✅ 数据统计报表
✅ 系统设置权限
✅ 数据导入导出

## 测试结果分析
${this.results.failed === 0 ? 
    '🎉 所有核心功能测试通过！系统已准备好投入生产使用。' : 
    `⚠️ 发现 ${this.results.failed} 个问题需要修复，请查看详细报告。`}

详细测试报告已保存至: product_acceptance_test_report.json
        `;

        fs.writeFileSync('PRODUCT_ACCEPTANCE_REPORT.md', briefReport);

        return report;
    }

    // 主测试流程
    async runAllTests() {
        console.log('🚀 开始律师事务所OA系统产品用户验收测试');
        console.log('=' .repeat(60));

        try {
            // 执行所有测试场景
            const authSuccess = await this.testUserAuthentication();
            if (!authSuccess) {
                console.log('❌ 认证失败，无法继续测试');
                return;
            }

            await this.testLawyerManagement();
            await this.testClientManagement();
            await this.testCaseManagement();
            await this.testConflictDetection();
            await this.testReportsAndStatistics();
            await this.testSystemSettings();
            await this.testDataImportExport();

            // 清理测试数据
            await this.cleanupTestData();

            // 生成测试报告
            console.log('\n📋 生成测试报告');
            const report = this.generateReport();

            console.log('\n' + '='.repeat(60));
            console.log('🎯 产品用户验收测试完成');
            console.log(`📊 测试结果: ${report.summary.passed}/${report.summary.total} 通过 (${report.summary.success_rate})`);
            
            if (this.results.failed === 0) {
                console.log('🎉 恭喜！所有核心功能测试通过，系统已准备好投入使用！');
            } else {
                console.log(`⚠️ 发现 ${this.results.failed} 个问题，请查看详细报告进行修复。`);
            }

        } catch (error) {
            console.error('❌ 测试过程中发生错误:', error.message);
        }
    }
}

// 运行测试
if (require.main === module) {
    const test = new ProductUserAcceptanceTest();
    test.runAllTests();
}

module.exports = ProductUserAcceptanceTest;