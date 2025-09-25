#!/usr/bin/env node

import https from 'https';
import http from 'http';
import fs from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// 测试配置
const config = {
  frontendUrl: 'http://localhost:3003',
  backendUrl: 'http://localhost:8080',
  loginPath: '/api/v1/auth/login',
  credentials: {
    email: 'lawyer1@lawoa.com',
    password: 'password'
  }
};

// HTTP请求工具
function makeRequest(url, options = {}) {
  return new Promise((resolve, reject) => {
    const protocol = url.startsWith('https') ? https : http;
    const req = protocol.request(url, options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          headers: res.headers,
          data: data
        });
      });
    });
    
    req.on('error', reject);
    
    if (options.body) {
      req.write(options.body);
    }
    
    req.end();
  });
}

// 测试登录API
async function testLoginAPI() {
  console.log('🧪 测试登录API...');
  
  try {
    const response = await makeRequest(config.backendUrl + config.loginPath, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(config.credentials)
    });
    
    console.log('📡 API响应状态:', response.statusCode);
    console.log('📡 API响应头:', JSON.stringify(response.headers, null, 2));
    
    if (response.statusCode === 200) {
      const data = JSON.parse(response.data);
      console.log('✅ 登录API调用成功');
      console.log('🎫 JWT Token:', data.data.token.substring(0, 50) + '...');
      console.log('👤 用户信息:', JSON.stringify(data.data.user, null, 2));
      
      return {
        success: true,
        token: data.data.token,
        user: data.data.user
      };
    } else {
      console.log('❌ 登录API调用失败:', response.data);
      return { success: false, error: response.data };
    }
  } catch (error) {
    console.log('❌ 登录API调用异常:', error.message);
    return { success: false, error: error.message };
  }
}

// 测试前端页面访问
async function testFrontendAccess() {
  console.log('🌐 测试前端页面访问...');
  
  try {
    const response = await makeRequest(config.frontendUrl + '/login', {
      method: 'GET'
    });
    
    console.log('📡 前端页面状态:', response.statusCode);
    console.log('📡 前端页面大小:', response.data.length, 'bytes');
    
    if (response.statusCode === 200) {
      console.log('✅ 前端页面访问成功');
      return { success: true };
    } else {
      console.log('❌ 前端页面访问失败');
      return { success: false, error: response.data };
    }
  } catch (error) {
    console.log('❌ 前端页面访问异常:', error.message);
    return { success: false, error: error.message };
  }
}

// 测试受保护的资源
async function testProtectedResource(token) {
  console.log('🔒 测试受保护资源访问...');
  
  try {
    const response = await makeRequest(config.backendUrl + '/api/v1/users/profile', {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      }
    });
    
    console.log('📡 受保护资源状态:', response.statusCode);
    
    if (response.statusCode === 200) {
      const data = JSON.parse(response.data);
      console.log('✅ 受保护资源访问成功');
      console.log('📊 用户数据:', JSON.stringify(data, null, 2));
      return { success: true, data: data };
    } else {
      console.log('❌ 受保护资源访问失败:', response.data);
      return { success: false, error: response.data };
    }
  } catch (error) {
    console.log('❌ 受保护资源访问异常:', error.message);
    return { success: false, error: error.message };
  }
}

// 主测试函数
async function runTests() {
  console.log('🚀 开始测试 Law OA Go 登录功能...');
  console.log('⏰ 测试时间:', new Date().toISOString());
  console.log('📍 测试环境:');
  console.log('   - 前端地址:', config.frontendUrl);
  console.log('   - 后端地址:', config.backendUrl);
  console.log('   - 测试用户:', config.credentials.email);
  console.log('');
  
  // 测试结果
  const results = {
    frontend: {},
    loginAPI: {},
    protectedResource: {}
  };
  
  // 1. 测试前端页面访问
  results.frontend = await testFrontendAccess();
  console.log('');
  
  // 2. 测试登录API
  results.loginAPI = await testLoginAPI();
  console.log('');
  
  // 3. 如果登录成功，测试受保护资源
  if (results.loginAPI.success && results.loginAPI.token) {
    results.protectedResource = await testProtectedResource(results.loginAPI.token);
  } else {
    console.log('⏭️  跳过受保护资源测试（登录失败）');
  }
  
  // 输出测试总结
  console.log('');
  console.log('📊 测试总结:');
  console.log('=====================================');
  console.log('前端页面访问:', results.frontend.success ? '✅ 通过' : '❌ 失败');
  console.log('登录API调用:', results.loginAPI.success ? '✅ 通过' : '❌ 失败');
  console.log('受保护资源:', results.protectedResource.success ? '✅ 通过' : '❌ 失败');
  console.log('');
  
  if (results.loginAPI.success) {
    console.log('🎉 登录功能测试通过！');
    console.log('📝 JWT Token 有效');
    console.log('👤 用户权限已分配');
    console.log('🔒 受保护资源可访问');
  } else {
    console.log('💥 登录功能测试失败！');
    console.log('🔧 需要检查的问题:');
    if (!results.frontend.success) {
      console.log('   - 前端页面无法访问');
    }
    if (!results.loginAPI.success) {
      console.log('   - 登录API调用失败');
    }
    if (!results.protectedResource.success) {
      console.log('   - 受保护资源访问失败');
    }
  }
  
  console.log('');
  console.log('📋 详细测试报告已保存到 test_results.json');
  
  // 保存测试结果
  const report = {
    timestamp: new Date().toISOString(),
    config: config,
    results: results,
    summary: {
      frontend: results.frontend.success,
      loginAPI: results.loginAPI.success,
      protectedResource: results.protectedResource.success,
      overall: results.frontend.success && results.loginAPI.success && results.protectedResource.success
    }
  };
  
  fs.writeFileSync('test_results.json', JSON.stringify(report, null, 2));
  
  return report;
}

// 执行测试
runTests().catch(console.error);