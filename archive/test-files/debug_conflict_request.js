#!/usr/bin/env node

/**
 * 调试冲突检查请求数据格式
 * 模拟前端发送的数据并与后端期望格式进行对比
 */

console.log('🔍 冲突检查请求数据格式调试');
console.log('='.repeat(60));

// 1. 后端期望的请求结构（从CheckConflictRequest struct）
console.log('\n📋 后端期望的请求结构:');
const expectedStructure = {
  clientId: { type: 'string', required: true, example: '14' },
  clientName: { type: 'string', required: true, example: '字节跳动科技有限公司' },
  clientType: { type: 'string', required: true, example: 'COMPANY', enum: ['PERSON', 'COMPANY'] },
  caseName: { type: 'string', required: true, example: '字节跳动诉腾讯垄断纠纷案' },
  caseType: { type: 'string', required: true, example: 'commercial' },
  otherParties: { type: 'array', required: false, example: ['腾讯科技有限公司'] },
  searchYears: { type: 'number', required: false, example: 5 },
  includeCorporateRelations: { type: 'boolean', required: false, example: true },
  searchDepth: { type: 'string', required: false, example: 'DEEP', enum: ['BASIC', 'STANDARD', 'DEEP'] },
  userId: { type: 'string', required: false, example: '1' },
  requestTime: { type: 'string', required: false, example: '2025-10-15T10:21:20.000Z' }
};

console.table(expectedStructure);

// 2. 模拟前端发送的数据（基于当前代码）
console.log('\n📤 模拟前端发送的数据:');
const frontendData = {
  clientId: '14',
  clientName: '字节跳动科技有限公司',
  clientType: 'COMPANY',
  caseName: '字节跳动诉腾讯垄断纠纷案',
  caseType: 'commercial',
  otherParties: ['腾讯科技有限公司'],
  searchYears: 5,
  includeCorporateRelations: true,
  searchDepth: 'DEEP',
  userId: '1',
  requestTime: new Date().toISOString()
};

console.log(JSON.stringify(frontendData, null, 2));

// 3. 数据类型验证
console.log('\n🔍 数据类型验证:');
const typeValidation = {
  clientId: typeof frontendData.clientId === 'string',
  clientName: typeof frontendData.clientName === 'string',
  clientType: typeof frontendData.clientType === 'string',
  caseName: typeof frontendData.caseName === 'string',
  caseType: typeof frontendData.caseType === 'string',
  otherParties: Array.isArray(frontendData.otherParties),
  searchYears: typeof frontendData.searchYears === 'number',
  includeCorporateRelations: typeof frontendData.includeCorporateRelations === 'boolean',
  searchDepth: typeof frontendData.searchDepth === 'string',
  userId: typeof frontendData.userId === 'string',
  requestTime: typeof frontendData.requestTime === 'string'
};

Object.entries(typeValidation).forEach(([field, isValid]) => {
  console.log(`  ${field}: ${isValid ? '✅' : '❌'} (${typeof frontendData[field]})`);
});

// 4. 必需字段检查
console.log('\n🔒 必需字段检查:');
const requiredFields = ['clientId', 'clientName', 'clientType', 'caseName', 'caseType'];
const requiredValidation = requiredFields.map(field => ({
  field,
  present: field in frontendData && frontendData[field] !== null && frontendData[field] !== undefined && frontendData[field] !== '',
  value: frontendData[field]
}));

requiredValidation.forEach(({ field, present, value }) => {
  console.log(`  ${field}: ${present ? '✅' : '❌'} (${value})`);
});

// 5. 枚举值验证
console.log('\n🏷️ 枚举值验证:');
const enumValidation = {
  clientType: ['PERSON', 'COMPANY'].includes(frontendData.clientType),
  searchDepth: ['BASIC', 'STANDARD', 'DEEP'].includes(frontendData.searchDepth)
};

Object.entries(enumValidation).forEach(([field, isValid]) => {
  console.log(`  ${field}: ${isValid ? '✅' : '❌'} (${frontendData[field]})`);
});

// 6. 生成测试用的JSON
console.log('\n📝 测试用的JSON数据:');
console.log('可以复制以下数据在浏览器控制台中测试API调用:');
console.log('```javascript');
console.log('const testData = ' + JSON.stringify(frontendData, null, 2) + ';');
console.log('');

console.log('fetch("/api/conflict/check", {');
console.log('  method: "POST",');
console.log('  headers: {');
console.log('    "Content-Type": "application/json",');
console.log('    "Authorization": "Bearer YOUR_TOKEN_HERE"');
console.log('  },');
console.log('  body: JSON.stringify(testData)');
console.log('});');
console.log('```');

// 7. 可能的问题点
console.log('\n⚠️ 可能的问题点:');
console.log('1. 客户类型枚举值是否正确（PERSON/COMPANY）');
console.log('2. 搜索深度枚举值是否正确（BASIC/STANDARD/DEEP）');
console.log('3. 时间戳格式是否为ISO 8601');
console.log('4. 请求头是否包含正确的Authorization');
console.log('5. 数组字段是否为空数组而非null/undefined');

// 8. 完整的调试步骤
console.log('\n🔧 完整的调试步骤:');
console.log('1. 打开浏览器开发者工具');
console.log('2. 进入新建案件的第4步（利益冲突检查）');
console.log('3. 填写表单数据并点击"下一步"');
console.log('4. 查看控制台中的详细日志输出');
console.log('5. 检查Network标签页中的请求详情');
console.log('6. 对比请求数据与期望格式是否一致');

console.log('\n🎯 如果仍然出现400错误，请检查:');
console.log('- 浏览器控制台中的错误详情');
console.log('- Network面板中的请求和响应');
console.log('- 后端日志中的具体错误信息');
console.log('- 确认所有必需字段都有值且类型正确');