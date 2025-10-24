// 测试案件类型转换逻辑
const { transformToConflictCheckRequest, debugConflictCheckRequest } = require('./frontend/src/utils/conflictTransform.ts');

// 模拟表单数据
const formData = {
  caseName: '字节跳动诉腾讯垄断纠纷案',
  caseType: 'COMMERCIAL', // 大写
  clientName: '字节跳动科技有限公司',
  opponentInfo: '腾讯',
  searchYears: 5,
  searchDepth: 'DEEP',
  includeCorporateRelations: true
};

console.log('=== 测试案件类型转换 ===');

// 执行转换
const { request: transformedRequest } = transformToConflictCheckRequest(formData);

console.log('转换前案件类型:', formData.caseType);
console.log('转换后案件类型:', transformedRequest.caseType);
console.log('转换后案件类型(类型):', typeof transformedRequest.caseType);

debugConflictCheckRequest(transformedRequest);