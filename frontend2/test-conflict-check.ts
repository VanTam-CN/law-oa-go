import { ConflictCheckService } from '../src/services/conflictCheck';

/**
 * 冲突检索功能测试
 */
async function testConflictCheck() {
  console.log('🧪 开始测试冲突检索功能...\n');
  
  const testRequest = {
    clientId: 'TEST_CLIENT_001',
    clientName: '测试客户',
    clientType: 'PERSON' as const,
    otherParties: ['对方当事人A', '对方当事人B'],
    caseName: '测试案件 - 合同纠纷',
    caseType: 'CIVIL',
    searchYears: 5,
    includeCorporateRelations: false,
    searchDepth: 'STANDARD' as const
  };
  
  try {
    console.log('📋 测试请求参数:');
    console.log(JSON.stringify(testRequest, null, 2));
    console.log('\n🔄 发起冲突检索请求...');
    
    const startTime = Date.now();
    const result = await ConflictCheckService.performConflictCheck(testRequest);
    const duration = Date.now() - startTime;
    
    console.log(`✅ 检索完成！耗时: ${duration}ms\n`);
    
    console.log('📊 检索结果:');
    console.log(`- 检索ID: ${result.checkId}`);
    console.log(`- 是否有冲突: ${result.hasConflict ? '是' : '否'}`);
    console.log(`- 冲突案件数量: ${result.conflictCases?.length || 0}`);
    console.log(`- 风险等级: ${result.riskAssessment?.overallRisk || 'N/A'}`);
    console.log(`- 需要审批: ${result.riskAssessment?.requiresApproval ? '是' : '否'}`);
    
    if (result.conflictCases && result.conflictCases.length > 0) {
      console.log('\n⚠️ 发现的冲突案件:');
      result.conflictCases.forEach((conflict, index) => {
        console.log(`  ${index + 1}. ${conflict.caseName}`);
        console.log(`     案件编号: ${conflict.caseNo}`);
        console.log(`     冲突类型: ${conflict.conflictType}`);
        console.log(`     风险等级: ${conflict.riskLevel}`);
        console.log(`     描述: ${conflict.description}`);
      });
    }
    
    if (result.recommendations && result.recommendations.length > 0) {
      console.log('\n💡 处理建议:');
      result.recommendations.forEach((rec, index) => {
        console.log(`  ${index + 1}. ${rec}`);
      });
    }
    
    console.log('\n📈 检索统计:');
    if (result.checkStatistics) {
      console.log(`- 总检索案件数: ${result.checkStatistics.totalCasesChecked}`);
      console.log(`- 委托人历史案件: ${result.checkStatistics.clientHistoryCases}`);
      console.log(`- 检索当事人数: ${result.checkStatistics.relatedPartiesChecked}`);
      console.log(`- 检索时间范围: ${result.checkStatistics.timeRange}`);
    }
    
    console.log('\n🎉 测试成功完成！');
    
    return {
      success: true,
      result: result,
      duration: duration
    };
    
  } catch (error) {
    console.error('❌ 测试失败:', error);
    return {
      success: false,
      error: error instanceof Error ? error.message : '未知错误'
    };
  }
}

// 如果是在Node.js环境中直接运行
if (typeof window === 'undefined') {
  console.log('⚡ 在Node.js环境中运行冲突检索测试\n');
  testConflictCheck().then(result => {
    if (result.success) {
      console.log('\n✅ 所有测试通过！');
      process.exit(0);
    } else {
      console.log('\n❌ 测试失败:', result.error);
      process.exit(1);
    }
  });
} else {
  console.log('⚡ 在浏览器环境中运行冲突检索测试');
  (window as any).testConflictCheck = testConflictCheck;
}

export { testConflictCheck };