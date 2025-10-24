const axios = require('axios');

// 测试前端与后端冲突检测功能的集成
async function testConflictDetectionIntegration() {
  console.log('🧪 开始测试前后端冲突检测功能集成...');

  const baseURL = 'http://localhost:8080';
  const conflictEndpoint = '/api/v1/conflict/check';

  // 模拟前端的冲突检测请求
  const conflictRequest = {
    clientId: "55",  // 阿里巴巴集团
    clientName: "阿里巴巴集团控股有限公司",
    clientType: "COMPANY",
    otherParties: ["字节跳动"],
    caseName: "测试案件",
    caseType: "civil",
    searchYears: 5,
    includeCorporateRelations: true,
    searchDepth: "STANDARD",
    userId: "45",  // 张伟律师
    requestTime: new Date().toISOString()
  };

  try {
    console.log('📤 发送冲突检测请求:', conflictRequest);

    const response = await axios.post(`${baseURL}${conflictEndpoint}`, conflictRequest, {
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 10000
    });

    console.log('✅ 冲突检测请求成功!');
    console.log('响应状态:', response.status);
    console.log('响应数据:', JSON.stringify(response.data, null, 2));

    // 验证响应结构
    const responseData = response.data;
    if (responseData.data) {
      const { checkId, hasConflict, conflictCases, riskAssessment, recommendations } = responseData.data;

      console.log('\n🎯 冲突检测结果分析:');
      console.log(`- 检查ID: ${checkId}`);
      console.log(`- 是否有冲突: ${hasConflict}`);
      console.log(`- 冲突案例数量: ${conflictCases ? conflictCases.length : 0}`);
      console.log(`- 风险评估: ${riskAssessment?.overallRisk || '未提供'}`);
      console.log(`- 建议数量: ${recommendations ? recommendations.length : 0}`);

      if (conflictCases && conflictCases.length > 0) {
        console.log('\n🔥 发现的冲突案例:');
        conflictCases.forEach((conflict, index) => {
          console.log(`  ${index + 1}. ID: ${conflict.caseId}, 案件名: ${conflict.caseName}, 风险: ${conflict.riskLevel}`);
          console.log(`     描述: ${conflict.description}`);
          console.log(`     冲突类型: ${conflict.conflictType}`);
        });
      }

      if (recommendations && recommendations.length > 0) {
        console.log('\n💡 建议:');
        recommendations.forEach((rec, index) => {
          console.log(`  ${index + 1}. ${rec}`);
        });
      }
    }

    return { success: true, data: response.data };

  } catch (error) {
    console.error('❌ 冲突检测请求失败:', error.message);

    if (error.response) {
      console.error('状态码:', error.response.status);
      console.error('响应数据:', error.response.data);
      console.error('响应头:', error.response.headers);
    } else if (error.request) {
      console.error('请求发送失败，无响应:', error.request);
    } else {
      console.error('请求配置错误:', error.config);
    }

    return { success: false, error: error.message };
  }
}

// 测试健康检查端点
async function testHealthCheck() {
  console.log('\n🏥 测试冲突检测服务健康状态...');

  try {
    const response = await axios.get('http://localhost:8080/api/v1/conflict/health');
    console.log('✅ 健康检查通过:', response.data);
    return true;
  } catch (error) {
    console.error('❌ 健康检查失败:', error.message);
    return false;
  }
}

// 主测试函数
async function runIntegrationTests() {
  console.log('🚀 开始前后端冲突检测功能集成测试\n');

  // 1. 健康检查
  const healthOk = await testHealthCheck();

  if (!healthOk) {
    console.log('⚠️ 健康检查失败，但继续进行功能测试...');
  }

  // 2. 冲突检测功能测试
  const result = await testConflictDetectionIntegration();

  console.log('\n📊 测试总结:');
  console.log(`- 健康检查: ${healthOk ? '✅ 通过' : '❌ 失败'}`);
  console.log(`- 冲突检测: ${result.success ? '✅ 通过' : '❌ 失败'}`);

  if (result.success) {
    console.log('\n🎉 前后端冲突检测功能集成测试成功完成！');
    console.log('现在可以在前端界面中正常使用冲突检测功能了。');
  } else {
    console.log('\n⚠️ 冲突检测功能存在问题，需要进一步调试。');
    console.log('请检查后端日志和前端控制台错误信息。');
  }
}

// 运行测试
if (require.main === module) {
  runIntegrationTests().catch(console.error);
}

module.exports = { testConflictDetectionIntegration, testHealthCheck, runIntegrationTests };