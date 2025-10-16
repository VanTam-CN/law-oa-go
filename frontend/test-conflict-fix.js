/**
 * 测试冲突检查功能修复
 */

import { performConflictCheck } from './src/services/conflict';

async function testConflictCheck() {
  console.log('🧪 开始测试冲突检查功能...');

  // 测试数据 - 使用新的表单格式
  const formData = {
    caseName: '测试案件名称',
    caseType: 'COMMERCIAL', // 这应该被转换为小写 'commercial'
    clientName: '测试客户有限公司',
    opponentInfo: '对方当事人A,对方当事人B',
    searchYears: 5,
    searchDepth: 'STANDARD',
    includeCorporateRelations: true
  };

  try {
    console.log('📤 发送请求:', formData);
    const result = await performConflictCheck(formData);

    console.log('✅ 请求成功!');
    console.log('📥 响应结果:', JSON.stringify(result, null, 2));

    if (result.success) {
      console.log('🎉 冲突检查调用成功，问题已修复!');
    } else {
      console.log('❌ 冲突检查调用失败:', result.message);
      if (result.error) {
        console.log('错误详情:', result.error);
      }
    }
  } catch (error) {
    console.error('💥 测试失败:', error);
  }
}

testConflictCheck();