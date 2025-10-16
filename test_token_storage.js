#!/usr/bin/env node

/**
 * 测试JWT token存储和获取
 * 检查当前系统中的token存储情况
 */

console.log('🔍 检查当前token存储状态...\n');

// 模拟localStorage和sessionStorage
let localStorage = {};
let sessionStorage = {};

// 如果在浏览器环境中，使用实际的localStorage
if (typeof window !== 'undefined' && window.localStorage) {
  localStorage = window.localStorage;
  sessionStorage = window.sessionStorage;
}

// 检查所有可能的token存储位置
const tokenKeys = [
  'auth_token',
  'law_oa_token',
  'token',
  'jwt_token',
  'access_token',
  'user_token'
];

console.log('📋 localStorage中的token:');
tokenKeys.forEach(key => {
  const value = localStorage.getItem ? localStorage.getItem(key) : localStorage[key];
  console.log(`  ${key}: ${value ? `存在 (${value.substring(0, 20)}...)` : '不存在'}`);
});

console.log('\n📋 sessionStorage中的token:');
tokenKeys.forEach(key => {
  const value = sessionStorage.getItem ? sessionStorage.getItem(key) : sessionStorage[key];
  console.log(`  ${key}: ${value ? `存在 (${value.substring(0, 20)}...)` : '不存在'}`);
});

// 模拟getAuthToken函数的逻辑
function simulateGetAuthToken() {
  console.log('\n🔧 模拟getAuthToken函数逻辑:');

  // 首先尝试从新的位置获取
  let token = localStorage.getItem ? localStorage.getItem('auth_token') : localStorage['auth_token'];
  console.log(`  1. 从auth_token获取: ${token ? '成功' : '失败'}`);

  // 如果没有，从旧的位置获取
  if (!token) {
    token = localStorage.getItem ? localStorage.getItem('law_oa_token') : localStorage['law_oa_token'];
    console.log(`  2. 从law_oa_token获取: ${token ? '成功' : '失败'}`);
  }

  // 如果还没有，尝试从sessionStorage获取
  if (!token) {
    token = sessionStorage.getItem ? sessionStorage.getItem('auth_token') : sessionStorage['auth_token'];
    if (!token) {
      token = sessionStorage.getItem ? sessionStorage.getItem('law_oa_token') : sessionStorage['law_oa_token'];
    }
    console.log(`  3. 从sessionStorage获取: ${token ? '成功' : '失败'}`);
  }

  console.log(`\n🎯 最终结果: ${token ? `有效token(${token.substring(0, 20)}...)` : '未找到token'}`);

  return token;
}

const token = simulateGetAuthToken();

if (!token) {
  console.log('\n❌ 未找到有效的JWT token');
  console.log('\n🔧 建议解决方案:');
  console.log('1. 确保用户已正确登录');
  console.log('2. 检查登录API是否正确存储token');
  console.log('3. 验证token存储的key名称是否正确');
  console.log('4. 检查浏览器是否允许使用localStorage');

  console.log('\n🧪 可以在浏览器控制台执行以下命令设置测试token:');
  console.log('// 设置测试token');
  console.log('localStorage.setItem("auth_token", "your-jwt-token-here");');
  console.log('localStorage.setItem("law_oa_token", "your-jwt-token-here");');

} else {
  console.log('\n✅ 成功找到JWT token');
  console.log('\n🔍 建议验证事项:');
  console.log('1. 验证token是否过期');
  console.log('2. 检查token格式是否正确');
  console.log('3. 确认后端是否接受此token格式');
}

console.log('\n📝 如果问题仍然存在，请提供以下信息:');
console.log('1. 浏览器控制台的完整错误信息');
console.log('2. 登录API的响应内容');
console.log('3. 网络请求的详细日志');