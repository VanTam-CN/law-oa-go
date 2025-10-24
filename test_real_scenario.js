#!/usr/bin/env node

/**
 * 真实场景端到端测试
 * 模拟前端发送的真实请求，验证修复效果
 */

const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080/api/v1';

// 模拟前端真实发送的请求格式（基于用户提供的日志）
const realRequest = {
  "clientId": "57",
  "clientName": "字节跳动科技有限公司",
  "caseName": "字节跳动诉腾讯垄断纠纷案",
  "caseType": "civil",
  "clientType": "PERSON",
  "otherParties": [
    "腾讯垄断纠纷案"
  ],
  "searchYears": 5,
  "includeCorporateRelations": true,
  "searchDepth": "DEEP",
  "userId": "45",
  "requestTime": "2025-10-23T05:45:56.631Z",
  "causeOfAction": "字节跳动诉腾讯垄断纠纷案"
};

async function login() {
  try {
    console.log('🔐 正在登录获取token...');
    const response = await axios.post(`${API_BASE_URL}/auth/login`, {
      email: 'admin@lawoa.com',
      password: 'admin123'
    });

    if (response.data && response.data.data && response.data.data.token) {
      console.log('✅ 登录成功，获取到token');
      return response.data.data.token;
    } else {
      throw new Error('登录响应格式异常');
    }
  } catch (error) {
    console.error('❌ 登录失败:', error.response?.data || error.message);
    throw error;
  }
}

async function testRealScenario() {
  try {
    console.log('🚀 开始真实场景测试...\n');

    // 1. 登录获取token
    const token = await login();

    // 2. 发送真实的前端请求
    console.log('📤 发送真实前端请求到冲突检测API...');
    console.log('请求体:', JSON.stringify(realRequest, null, 2));

    const config = {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    };

    const startTime = Date.now();
    const response = await axios.post(`${API_BASE_URL}/conflict/check`, realRequest, config);
    const endTime = Date.now();

    console.log('✅ API调用成功!');
    console.log('📊 响应状态码:', response.status);
    console.log('⏱️ 请求耗时:', endTime - startTime, 'ms');
    console.log('📋 响应结构:', JSON.stringify(response.data, null, 2));

    // 3. 验证响应格式
    console.log('\n🔍 验证响应格式...');

    const hasRequiredFields = [
      'success',
      'data',
      'message'
    ].every(field => response.data.hasOwnProperty(field));

    if (hasRequiredFields) {
      console.log('✅ 响应包含必需字段: success, data, message');
    } else {
      console.log('❌ 响应缺少必需字段');
      return false;
    }

    // 4. 验证数据结构
    const data = response.data.data;
    const hasDataFields = [
      'checkId',
      'hasConflict',
      'conflictCases',
      'checkStatistics',
      'riskAssessment',
      'recommendations'
    ].every(field => data.hasOwnProperty(field));

    if (hasDataFields) {
      console.log('✅ 数据包含必需字段');
      console.log(`   - 检查ID: ${data.checkId}`);
      console.log(`   - 冲突状态: ${data.hasConflict ? '发现冲突' : '无冲突'}`);
      console.log(`   - 冲突案例数: ${data.conflictCases.length}`);
      console.log(`   - 风险等级: ${data.riskAssessment?.overallRisk || '未知'}`);
      console.log(`   - 建议数量: ${data.recommendations?.length || 0}`);
    } else {
      console.log('❌ 数据缺少必需字段');
      return false;
    }

    // 5. 测试健康检查
    console.log('\n🏥 测试健康检查API...');
    const healthResponse = await axios.get(`${API_BASE_URL}/conflict/health`, config);
    console.log('✅ 健康检查成功:', healthResponse.data.data?.status);

    return true;

  } catch (error) {
    console.error('❌ 真实场景测试失败:');

    if (error.response) {
      console.error('状态码:', error.response.status);
      console.error('响应数据:', error.response.data);
      console.error('请求URL:', error.config.url);
      console.error('请求方法:', error.config.method?.toUpperCase());
      console.error('请求头:', error.config.headers);
    } else {
      console.error('网络错误:', error.message);
    }

    return false;
  }
}

async function runFinalTest() {
  console.log('🎯 真实场景端到端测试\n');
  console.log('=' .repeat(50));

  const success = await testRealScenario();

  console.log('\n' + '=' .repeat(50));
  if (success) {
    console.log('🎉 所有测试通过！前后端集成修复成功！');
    console.log('\n✨ 修复总结:');
    console.log('1. ✅ JWT认证上下文键名不匹配问题已修复');
    console.log('2. ✅ API响应格式标准化已完成');
    console.log('3. ✅ 前后端字段映射已对齐');
    console.log('4. ✅ 真实场景验证通过');
    process.exit(0);
  } else {
    console.log('💥 测试失败，需要进一步检查');
    process.exit(1);
  }
}

// 运行最终测试
if (require.main === module) {
  runFinalTest().catch(console.error);
}

module.exports = { runFinalTest, testRealScenario };