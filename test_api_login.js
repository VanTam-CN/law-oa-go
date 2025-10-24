#!/usr/bin/env node

const axios = require('axios');

async function testLogin() {
  try {
    console.log('🔐 测试登录...');

    const response = await axios.post('http://localhost:8080/api/v1/auth/login', {
      email: 'zhangwei@law.com',
      password: 'law123456'
    });

    console.log('✅ 登录成功');
    console.log('Token:', response.data.data.token);

    // 测试冲突检测
    await testConflictCheck(response.data.data.token);

  } catch (error) {
    console.error('❌ 登录失败:', error.response?.data || error.message);
  }
}

async function testConflictCheck(token) {
  try {
    console.log('\n🔍 测试冲突检测...');

    const conflictRequest = {
      clientId: '57', // 字节跳动科技有限公司
      clientName: '字节跳动科技有限公司',
      caseName: '字节跳动诉腾讯垄断纠纷案',
      caseType: 'civil',
      clientType: 'COMPANY',
      otherParties: ['腾讯控股有限公司'],
      searchYears: 5,
      includeCorporateRelations: true,
      searchDepth: 'DEEP',
      userId: '45', // 张伟律师的ID
      requestTime: new Date().toISOString()
    };

    const response = await axios.post(
      'http://localhost:8080/api/v1/conflict/check',
      conflictRequest,
      {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      }
    );

    console.log('✅ 冲突检测完成');
    console.log('响应数据:', JSON.stringify(response.data, null, 2));

    // 检查是否检测到冲突
    if (response.data.data.hasConflict) {
      console.log('\n🚨 检测到冲突!');
      console.log(`冲突案例数: ${response.data.data.conflictCases.length}`);
      console.log('风险等级:', response.data.data.riskAssessment.overallRisk);

      console.log('\n冲突案例详情:');
      response.data.data.conflictCases.forEach((conflict, index) => {
        console.log(`  ${index + 1}. ${conflict.caseName} (${conflict.conflictType})`);
        console.log(`     风险等级: ${conflict.riskLevel}`);
        console.log(`     描述: ${conflict.description}`);
        console.log(`     详情: ${conflict.conflictDetails}`);
      });
    } else {
      console.log('\n✅ 未检测到冲突');
    }

    console.log('\n建议:');
    response.data.data.recommendations.forEach((rec, index) => {
      console.log(`  ${index + 1}. ${rec}`);
    });

  } catch (error) {
    console.error('❌ 冲突检测失败:', error.response?.data || error.message);
    if (error.response) {
      console.error('状态码:', error.response.status);
      console.error('响应详情:', JSON.stringify(error.response.data, null, 2));
    }
  }
}

testLogin();