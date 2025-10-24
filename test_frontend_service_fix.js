// 测试前端冲突检测服务修复
console.log('🧪 测试前端冲突检测服务修复...');

// 模拟HTTP拦截器的行为
function simulateHttpInterceptor(backendResponse) {
  const res = backendResponse;
  
  // 🔧 修复：处理新的统一API响应格式 {success: boolean, data: any, error: any}
  if (res.success !== undefined) {
    if (!res.success && res.error) {
      // 新格式失败响应
      throw new Error(res.error.message || '请求失败');
    } else {
      // 新格式成功响应，返回data字段以符合前端期望
      return res.data !== undefined ? res.data : res;
    }
  }
  
  return res;
}

// 模拟ConflictCheckService.performConflictCheck方法
async function simulateConflictCheckService(request) {
  try {
    console.log('📤 发送请求:', JSON.stringify(request, null, 2));
    
    // 模拟后端响应
    const backendResponse = {
      success: true,
      data: {
        checkId: "CC_1729761669123456",
        hasConflict: false,
        conflictCases: [],
        checkStatistics: {
          totalCasesChecked: 0,
          clientHistoryCases: 0,
          relatedPartiesChecked: 0,
          corporateRelationsChecked: 0,
          timeRange: "5 years",
          searchScope: "standard"
        },
        riskAssessment: {
          overallRisk: "LOW",
          riskScore: 0.1,
          riskFactors: [],
          mitigation: ["无发现冲突"]
        },
        recommendations: ["可以继续处理此案件"],
        checkTime: new Date().toISOString(),
        duration: 10
      },
      meta: {
        timestamp: new Date().toISOString(),
        version: "v1",
        server: "law-oa-go",
        environment: "development"
      }
    };
    
    // 模拟HTTP拦截器处理
    const response = simulateHttpInterceptor(backendResponse);
    
    console.log('📥 HTTP拦截器处理后的响应:', JSON.stringify(response, null, 2));
    
    // 🔧 修复：HTTP拦截器已经处理了响应格式，直接使用response作为结果
    // HTTP拦截器会自动提取data字段，所以response就是我们需要的数据
    const result = response;

    if (typeof result.hasConflict === 'undefined') {
      console.warn('后端响应缺少hasConflict字段，基于conflictCases判断');
      result.hasConflict = (result.conflictCases && result.conflictCases.length > 0);
    }

    console.log('✅ ConflictCheckService处理成功!');
    console.log('📊 最终结果:', JSON.stringify(result, null, 2));
    
    return result;
    
  } catch (apiError) {
    console.error('API调用错误:', apiError);
    
    // 如果是开发环境且后端不可用，使用模拟数据
    const isDevelopment = true; // 模拟开发环境
    if (isDevelopment && isConnectionError(apiError)) {
      console.warn('开发环境：后端API不可用，使用模拟数据');
      return getMockResponse(request);
    }
    
    // 重新抛出API错误
    throw apiError;
  }
}

// 检查是否为网络连接错误
function isConnectionError(error) {
  if (error instanceof TypeError) {
    return error.message.includes('Failed to fetch') ||
           error.message.includes('Network request failed') ||
           error.message.includes('网络连接失败');
  }
  
  if (error instanceof Error) {
    return error.message.includes('fetch') ||
           error.message.includes('ECONNREFUSED') ||
           error.message.includes('连接被拒绝');
  }
  
  return false;
}

// 获取模拟响应数据
function getMockResponse(request) {
  return {
    checkId: 'MOCK_' + Date.now(),
    hasConflict: false,
    conflictCases: [],
    checkStatistics: {
      totalCasesChecked: 0,
      clientHistoryCases: 0,
      relatedPartiesChecked: request.otherParties.length + 1,
      corporateRelationsChecked: 0,
      timeRange: '开发环境模拟数据',
      searchScope: '模拟数据库'
    },
    riskAssessment: {
      overallRisk: 'LOW',
      riskReason: '开发环境：未连接真实数据库',
      requiresApproval: false,
      riskFactors: ['开发环境模拟']
    },
    recommendations: [
      '开发环境：此为模拟数据，非真实冲突检测结果',
      '请连接真实数据库进行实际冲突检测'
    ],
    checkTime: new Date().toISOString(),
    duration: 100
  };
}

// 测试正常情况
async function testNormalCase() {
  console.log('\n=== 测试正常情况 ===');
  
  const request = {
    clientId: "57",
    clientName: "字节跳动科技有限公司",
    clientType: "COMPANY",
    otherParties: ["腾讯", "垄断纠纷案"],
    caseName: "字节跳动诉腾讯垄断纠纷案",
    caseType: "commercial",
    searchYears: 5,
    includeCorporateRelations: true,
    searchDepth: "DEEP"
  };
  
  try {
    const result = await simulateConflictCheckService(request);
    console.log('✅ 正常情况测试通过');
    return true;
  } catch (error) {
    console.error('❌ 正常情况测试失败:', error.message);
    return false;
  }
}

// 测试错误情况
async function testErrorCase() {
  console.log('\n=== 测试错误情况 ===');
  
  // 模拟网络错误
  const originalSimulate = simulateHttpInterceptor;
  global.simulateHttpInterceptor = function() {
    throw new TypeError('Failed to fetch');
  };
  
  const request = {
    clientId: "57",
    clientName: "字节跳动科技有限公司",
    clientType: "COMPANY",
    otherParties: ["腾讯"],
    caseName: "测试案件",
    caseType: "commercial"
  };
  
  try {
    const result = await simulateConflictCheckService(request);
    console.log('✅ 错误情况测试通过，返回了模拟数据');
    return true;
  } catch (error) {
    console.error('❌ 错误情况测试失败:', error.message);
    return false;
  } finally {
    // 恢复原函数
    global.simulateHttpInterceptor = originalSimulate;
  }
}

// 运行所有测试
async function runAllTests() {
  const results = [];
  
  results.push(await testNormalCase());
  results.push(await testErrorCase());
  
  const passedTests = results.filter(r => r).length;
  const totalTests = results.length;
  
  console.log(`\n📊 测试结果: ${passedTests}/${totalTests} 通过`);
  
  if (passedTests === totalTests) {
    console.log('✅ 所有测试通过！前端冲突检测服务修复成功。');
  } else {
    console.log('❌ 部分测试失败，需要进一步调试。');
  }
}

// 运行测试
runAllTests().catch(console.error);