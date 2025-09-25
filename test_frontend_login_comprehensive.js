#!/usr/bin/env node

const http = require('http');
const https = require('https');
const url = require('url');

// 测试frontend2登录功能
async function testFrontendLogin() {
  console.log('🚀 开始测试frontend2登录功能...\n');
  
  // 测试结果
  const results = {
    frontendAccessible: false,
    loginPageAccessible: false,
    loginApiWorking: false,
    dashboardAccessible: false,
    errors: []
  };
  
  // 1. 测试前端服务是否可访问
  console.log('🔍 步骤1: 测试前端服务可访问性...');
  try {
    const frontendResponse = await makeRequest('http://localhost:3002/');
    console.log(`✅ 前端服务响应: ${frontendResponse.status} ${frontendResponse.statusText}`);
    results.frontendAccessible = true;
    
    // 检查响应内容
    if (frontendResponse.data.includes('html')) {
      console.log('✅ 前端页面内容正常');
    } else {
      console.log('⚠️  前端页面内容可能异常');
    }
    
  } catch (error) {
    console.error('❌ 前端服务不可访问:', error.message);
    results.errors.push('前端服务不可访问: ' + error.message);
  }
  
  // 2. 测试登录页面是否可访问
  console.log('\n🔍 步骤2: 测试登录页面可访问性...');
  try {
    const loginResponse = await makeRequest('http://localhost:3002/login');
    console.log(`✅ 登录页面响应: ${loginResponse.status} ${loginResponse.statusText}`);
    results.loginPageAccessible = true;
    
    // 检查登录页面内容
    if (loginResponse.data.includes('登录') || loginResponse.data.includes('Login') || loginResponse.data.includes('form')) {
      console.log('✅ 登录页面内容正常');
      
      // 检查是否包含登录表单元素
      const hasForm = loginResponse.data.includes('<form') || 
                     loginResponse.data.includes('input') || 
                     loginResponse.data.includes('button');
      
      if (hasForm) {
        console.log('✅ 检测到登录表单元素');
      } else {
        console.log('⚠️  未检测到登录表单元素');
      }
      
    } else {
      console.log('⚠️  登录页面内容可能异常');
      console.log('页面内容预览:', loginResponse.data.substring(0, 200));
    }
    
  } catch (error) {
    console.error('❌ 登录页面不可访问:', error.message);
    results.errors.push('登录页面不可访问: ' + error.message);
  }
  
  // 3. 测试登录API
  console.log('\n🔍 步骤3: 测试登录API...');
  try {
    const loginData = {
      email: 'lawyer1@lawoa.com',
      password: 'password'
    };
    
    const apiResponse = await makeRequest('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      data: loginData
    });
    
    console.log(`✅ 登录API响应: ${apiResponse.status} ${apiResponse.statusText}`);
    console.log('API响应数据:', JSON.stringify(apiResponse.data, null, 2));
    
    if (apiResponse.status === 200 && apiResponse.data.success) {
      console.log('✅ 登录API工作正常');
      results.loginApiWorking = true;
      
      // 检查是否返回了token
      if (apiResponse.data.data && apiResponse.data.data.token) {
        console.log('✅ 登录成功，获取到JWT token');
        
        // 4. 测试需要认证的页面/API
        console.log('\n🔍 步骤4: 测试需要认证的页面...');
        
        const token = apiResponse.data.data.token;
        
        // 测试dashboard页面
        try {
          const dashboardResponse = await makeRequest('http://localhost:3002/dashboard', {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Cookie': `token=${token}`
            }
          });
          
          console.log(`✅ Dashboard页面响应: ${dashboardResponse.status} ${dashboardResponse.statusText}`);
          results.dashboardAccessible = true;
          
          if (dashboardResponse.status === 200) {
            console.log('✅ Dashboard页面访问成功');
            
            // 检查页面内容
            if (dashboardResponse.data.includes('dashboard') || 
                dashboardResponse.data.includes('仪表板') ||
                dashboardResponse.data.includes('统计')) {
              console.log('✅ Dashboard页面内容正常');
            }
          }
          
        } catch (error) {
          console.error('❌ Dashboard页面访问失败:', error.message);
          results.errors.push('Dashboard页面访问失败: ' + error.message);
        }
        
        // 测试其他API端点
        console.log('\n🔍 步骤5: 测试其他API端点...');
        
        const apiEndpoints = [
          'http://localhost:8080/api/v1/dashboard/statistics',
          'http://localhost:8080/api/v1/users/profile',
          'http://localhost:8080/api/v1/clients',
          'http://localhost:8080/api/v1/cases'
        ];
        
        for (const endpoint of apiEndpoints) {
          try {
            const response = await makeRequest(endpoint, {
              headers: {
                'Authorization': `Bearer ${token}`
              }
            });
            
            console.log(`✅ ${endpoint}: ${response.status} ${response.statusText}`);
            
            if (response.status === 200) {
              console.log(`   数据: ${JSON.stringify(response.data).substring(0, 100)}...`);
            }
            
          } catch (error) {
            console.error(`❌ ${endpoint}: ${error.message}`);
          }
        }
        
      } else {
        console.log('❌ 登录响应中没有token');
        results.errors.push('登录响应中没有token');
      }
      
    } else {
      console.log('❌ 登录API返回错误');
      results.errors.push('登录API返回错误: ' + JSON.stringify(apiResponse.data));
    }
    
  } catch (error) {
    console.error('❌ 登录API测试失败:', error.message);
    results.errors.push('登录API测试失败: ' + error.message);
  }
  
  // 6. 测试前端路由
  console.log('\n🔍 步骤6: 测试前端路由...');
  
  const routes = [
    '/',
    '/login',
    '/dashboard',
    '/users',
    '/clients',
    '/cases',
    '/documents'
  ];
  
  for (const route of routes) {
    try {
      const response = await makeRequest(`http://localhost:3002${route}`);
      console.log(`✅ ${route}: ${response.status} ${response.statusText}`);
    } catch (error) {
      console.error(`❌ ${route}: ${error.message}`);
    }
  }
  
  // 7. 生成测试报告
  console.log('\n📋 测试报告');
  console.log('=' .repeat(50));
  
  console.log(`前端服务可访问: ${results.frontendAccessible ? '✅' : '❌'}`);
  console.log(`登录页面可访问: ${results.loginPageAccessible ? '✅' : '❌'}`);
  console.log(`登录API工作: ${results.loginApiWorking ? '✅' : '❌'}`);
  console.log(`Dashboard可访问: ${results.dashboardAccessible ? '✅' : '❌'}`);
  
  if (results.errors.length > 0) {
    console.log('\n❌ 发现的错误:');
    results.errors.forEach((error, index) => {
      console.log(`   ${index + 1}. ${error}`);
    });
  }
  
  const successCount = Object.values(results).filter(v => v === true).length;
  const totalChecks = 4; // 前端服务、登录页面、登录API、Dashboard
  
  console.log(`\n📊 测试结果: ${successCount}/${totalChecks} 项检查通过`);
  
  if (successCount === totalChecks) {
    console.log('🎉 所有测试通过！frontend2登录功能正常工作。');
  } else {
    console.log('⚠️  部分测试失败，请检查上述错误信息。');
  }
  
  return results;
}

// HTTP请求工具函数
function makeRequest(urlString, options = {}) {
  return new Promise((resolve, reject) => {
    const parsedUrl = url.parse(urlString);
    const isHttps = parsedUrl.protocol === 'https:';
    const httpModule = isHttps ? https : http;
    
    const requestOptions = {
      hostname: parsedUrl.hostname,
      port: parsedUrl.port || (isHttps ? 443 : 80),
      path: parsedUrl.path,
      method: options.method || 'GET',
      headers: {
        'User-Agent': 'Law-OA-Test/1.0',
        ...options.headers
      }
    };
    
    const req = httpModule.request(requestOptions, (res) => {
      let data = '';
      
      res.on('data', (chunk) => {
        data += chunk;
      });
      
      res.on('end', () => {
        try {
          let parsedData = data;
          
          // 尝试解析JSON
          try {
            parsedData = JSON.parse(data);
          } catch (e) {
            // 不是JSON，保持原样
          }
          
          resolve({
            status: res.statusCode,
            statusText: res.statusMessage,
            data: parsedData,
            headers: res.headers
          });
        } catch (error) {
          reject(error);
        }
      });
    });
    
    req.on('error', (error) => {
      reject(error);
    });
    
    // 发送请求体
    if (options.data) {
      req.write(JSON.stringify(options.data));
    }
    
    req.end();
  });
}

// 运行测试
testFrontendLogin().catch(console.error);