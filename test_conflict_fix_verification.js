#!/usr/bin/env node

/**
 * 冲突检测API修复验证测试
 * 测试前后端API通信是否已修复
 */

const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080/api/v1';

// 测试用户认证
const testUser = {
    email: 'admin@lawoa.com',
    password: 'admin123'
};

// 测试的冲突检测请求数据
const conflictTestData = {
    clientId: '1',
    clientName: '测试客户公司', // 添加客户名称
    clientType: 'COMPANY', // 添加客户类型
    caseName: '测试冲突检测修复',
    caseType: '民事诉讼', // 添加案件类型
    userId: '1',
    searchYears: 5,
    searchDepth: 'STANDARD',
    includeCorporateRelations: true,
    relatedParties: [
        {
            name: '测试公司A',
            type: 'corporate',
            relationship: 'client'
        }
    ]
};

async function login() {
    try {
        console.log('🔐 正在登录...');
        const response = await axios.post(`${API_BASE_URL}/auth/login`, testUser);

        if (response.data && response.data.data && response.data.data.token) {
            console.log('✅ 登录成功');
            return response.data.data.token;
        } else {
            throw new Error('登录响应格式异常');
        }
    } catch (error) {
        console.error('❌ 登录失败:', error.response?.data || error.message);
        throw error;
    }
}

async function testConflictCheck(token) {
    try {
        console.log('🔍 正在测试冲突检测API...');

        const config = {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        };

        const response = await axios.post(
            `${API_BASE_URL}/conflict/check`,
            conflictTestData,
            config
        );

        console.log('✅ 冲突检测API调用成功');
        console.log('📋 响应数据:', JSON.stringify(response.data, null, 2));

        // 验证响应结构
        if (response.data && response.data.code === 200) {
            console.log('✅ API响应格式正确');

            if (response.data.data) {
                const data = response.data.data;
                console.log(`✅ 返回数据完整 - 检查ID: ${data.checkId}`);
                console.log(`✅ 冲突状态: ${data.hasConflict ? '发现冲突' : '无冲突'}`);
                console.log(`✅ 风险等级: ${data.riskAssessment?.overallRisk || 'LOW'}`);

                return true;
            }
        } else {
            console.log('❌ API响应格式异常');
            console.log('响应内容:', response.data);
            return false;
        }
    } catch (error) {
        console.error('❌ 冲突检测API调用失败:');
        if (error.response) {
            console.error('状态码:', error.response.status);
            console.error('响应数据:', error.response.data);
        } else {
            console.error('错误信息:', error.message);
        }
        return false;
    }
}

async function testConflictHealth(token) {
    try {
        console.log('🏥 正在测试冲突检测服务健康检查...');

        // 健康检查使用认证token，以防需要
        const config = {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        };

        const response = await axios.get(`${API_BASE_URL}/conflict/health`, config);

        console.log('✅ 健康检查API调用成功');
        console.log('📋 健康状态:', response.data);

        if (response.data && response.data.data && response.data.data.status === 'healthy') {
            console.log('✅ 冲突检测服务状态健康');
            return true;
        } else {
            console.log('❌ 冲突检测服务状态异常');
            return false;
        }
    } catch (error) {
        console.error('❌ 健康检查API调用失败:', error.response?.data || error.message);
        return false;
    }
}

async function runTests() {
    console.log('🚀 开始冲突检测API修复验证测试\n');

    let successCount = 0;
    let totalTests = 3;

    // 测试1: 登录获取token
    try {
        const token = await login();

        // 测试2: 冲突检测API
        const conflictSuccess = await testConflictCheck(token);
        if (conflictSuccess) successCount++;

        // 测试3: 健康检查API
        const healthSuccess = await testConflictHealth(token);
        if (healthSuccess) successCount++;

        successCount++; // 登录成功

    } catch (error) {
        console.error('❌ 测试过程中出现致命错误:', error.message);
    }

    console.log('\n📊 测试结果汇总:');
    console.log(`✅ 成功: ${successCount}/${totalTests}`);
    console.log(`❌ 失败: ${totalTests - successCount}/${totalTests}`);

    if (successCount === totalTests) {
        console.log('🎉 所有测试通过！冲突检测API修复成功！');
        process.exit(0);
    } else {
        console.log('💥 部分测试失败，需要进一步检查');
        process.exit(1);
    }
}

// 运行测试
if (require.main === module) {
    runTests().catch(console.error);
}

module.exports = { runTests, testConflictCheck, testConflictHealth };