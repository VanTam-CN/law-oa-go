#!/usr/bin/env node

const axios = require('axios');

async function testConflictCheck() {
  try {
    console.log('🔐 登录获取token...');

    // 1. 登录获取token
    const loginResponse = await axios.post('http://localhost:8080/api/v1/auth/login', {
      email: 'zhangwei@law.com',
      password: 'law123456'
    });

    const token = loginResponse.data.data.token;
    console.log('✅ 登录成功，获得token');

    // 2. 测试冲突检测API
    console.log('\n🔍 测试冲突检测API...');

    const conflictRequest = {
      clientId: "57", // 字节跳动科技有限公司
      clientName: "字节跳动科技有限公司",
      caseName: "字节跳动诉腾讯垄断纠纷案",
      caseType: "civil", // 前端发送的小写英文
      clientType: "COMPANY",
      otherParties: ["腾讯控股有限公司"],
      searchYears: 5,
      includeCorporateRelations: true,
      searchDepth: "DEEP",
      userId: "45", // 张伟律师的ID
      requestTime: new Date().toISOString()
    };

    console.log('发送的请求体:', JSON.stringify(conflictRequest, null, 2));

    const conflictResponse = await axios.post(
      'http://localhost:8080/api/v1/conflict/check',
      conflictRequest,
      {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      }
    );

    console.log('✅ 冲突检测API调用成功！');
    console.log('响应状态:', conflictResponse.status);
    console.log('响应数据:', JSON.stringify(conflictResponse.data, null, 2));

    // 3. 检查冲突检测结果
    if (conflictResponse.data.success !== false) {
      const data = conflictResponse.data.data || conflictResponse.data;

      if (data.hasConflict) {
        console.log('\n🚨 检测到冲突!');
        console.log(`冲突案例数: ${data.conflictCases ? data.conflictCases.length : 0}`);
        console.log('风险等级:', data.riskAssessment?.overallRisk || 'UNKNOWN');

        if (data.conflictCases && data.conflictCases.length > 0) {
          console.log('\n冲突案例详情:');
          data.conflictCases.forEach((conflict, index) => {
            console.log(`  ${index + 1}. ${conflict.caseName} (${conflict.conflictType})`);
            console.log(`     风险等级: ${conflict.riskLevel}`);
            console.log(`     描述: ${conflict.description}`);
          });
        }
      } else {
        console.log('\n✅ 未检测到冲突');
      }

      if (data.recommendations && data.recommendations.length > 0) {
        console.log('\n建议:');
        data.recommendations.forEach((rec, index) => {
          console.log(`  ${index + 1}. ${rec}`);
        });
      }
    } else {
      console.log('❌ API返回失败:', conflictResponse.data);
    }

  } catch (error) {
    console.error('❌ 测试失败:', error.response?.data || error.message);
    if (error.response) {
      console.error('状态码:', error.response.status);
      console.error('响应详情:', JSON.stringify(error.response.data, null, 2));
    }
  }
}

testConflictCheck();