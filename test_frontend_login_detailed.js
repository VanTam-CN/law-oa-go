#!/usr/bin/env node

const http = require('http');
const https = require('https');
const url = require('url');
const fs = require('fs');

// 测试frontend2详细登录功能
async function testFrontendLoginDetailed() {
  console.log('🚀 开始详细测试frontend2登录功能...\n');
  
  // 测试结果
  const results = {
    frontendService: false,
    loginPage: false,
    loginForm: false,
    loginApi: false,
    authentication: false,
    frontendRoutes: false,
    pageContent: false,
    errors: []
  };
  
  // 1. 详细测试前端服务
  console.log('🔍 步骤1: 详细测试前端服务...');
  try {
    const response = await makeRequest('http://localhost:3002/');
    console.log(`✅ 前端服务状态: ${response.status} ${response.statusText}`);
    
    // 保存前端页面内容
    fs.writeFileSync('frontend_index.html', response.data);
    console.log('✅ 前端首页内容已保存到 frontend_index.html');
    
    // 分析页面内容
    const hasReactRoot = response.data.includes('id="root"') || response.data.includes('id="app"');
    const hasScripts = response.data.includes('<script');
    const hasReact = response.data.includes('react') || response.data.includes('React');
    
    console.log(`✅ React根元素: ${hasReactRoot ? '检测到' : '未检测到'}`);
    console.log(`✅ 脚本标签: ${hasScripts ? '检测到' : '未检测到'}`);
    console.log(`✅ React相关: ${hasReact ? '检测到' : '未检测到'}`);
    
    results.frontendService = true;
    results.pageContent = hasReactRoot && hasScripts;
    
  } catch (error) {
    console.error('❌ 前端服务测试失败:', error.message);
    results.errors.push('前端服务测试失败: ' + error.message);
  }
  
  // 2. 详细测试登录页面
  console.log('\n🔍 步骤2: 详细测试登录页面...');
  try {
    const response = await makeRequest('http://localhost:3002/login');
    console.log(`✅ 登录页面状态: ${response.status} ${response.statusText}`);
    
    // 保存登录页面内容
    fs.writeFileSync('frontend_login.html', response.data);
    console.log('✅ 登录页面内容已保存到 frontend_login.html');
    
    // 分析登录页面内容
    const pageContent = response.data.toLowerCase();
    const hasLoginForm = pageContent.includes('<form') || pageContent.includes('input');
    const hasEmailInput = pageContent.includes('email') || pageContent.includes('邮箱');
    const hasPasswordInput = pageContent.includes('password') || pageContent.includes('密码');
    const hasLoginButton = pageContent.includes('button') && (pageContent.includes('login') || pageContent.includes('登录'));
    const hasReactApp = pageContent.includes('react') || pageContent.includes('id="root"');
    
    console.log(`✅ 登录表单: ${hasLoginForm ? '检测到' : '未检测到'}`);
    console.log(`✅ 邮箱输入框: ${hasEmailInput ? '检测到' : '未检测到'}`);
    console.log(`✅ 密码输入框: ${hasPasswordInput ? '检测到' : '未检测到'}`);
    console.log(`✅ 登录按钮: ${hasLoginButton ? '检测到' : '未检测到'}`);
    console.log(`✅ React应用: ${hasReactApp ? '检测到' : '未检测到'}`);
    
    results.loginPage = true;
    results.loginForm = hasLoginForm && hasEmailInput && hasPasswordInput;
    
    if (!results.loginForm) {
      console.log('⚠️  登录表单可能正在由React动态渲染');
    }
    
  } catch (error) {
    console.error('❌ 登录页面测试失败:', error.message);
    results.errors.push('登录页面测试失败: ' + error.message);
  }
  
  // 3. 详细测试登录API
  console.log('\n🔍 步骤3: 详细测试登录API...');
  try {
    const loginData = {
      email: 'lawyer1@lawoa.com',
      password: 'password'
    };
    
    const response = await makeRequest('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      data: loginData
    });
    
    console.log(`✅ 登录API状态: ${response.status} ${response.statusText}`);
    
    // 保存API响应
    fs.writeFileSync('login_api_response.json', JSON.stringify(response, null, 2));
    console.log('✅ 登录API响应已保存到 login_api_response.json');
    
    // 分析API响应
    if (response.status === 200 && response.data.success) {
      const token = response.data.data.token;
      const user = response.data.data.user;
      
      console.log(`✅ 登录成功: ${user.name} (${user.email})`);
      console.log(`✅ 用户角色: ${user.role}`);
      console.log(`✅ 用户ID: ${user.id}`);
      console.log(`✅ JWT Token: ${token.substring(0, 50)}...`);
      console.log(`✅ 过期时间: ${response.data.data.expires_at}`);
      
      results.loginApi = true;
      
      // 4. 测试认证功能
      console.log('\n🔍 步骤4: 测试认证功能...');
      
      // 测试使用token访问受保护的API
      const protectedApis = [
        'http://localhost:8080/api/v1/dashboard/statistics',
        'http://localhost:8080/api/v1/users/profile',
        'http://localhost:8080/api/v1/clients/list',
        'http://localhost:8080/api/v1/cases/list'
      ];
      
      let authSuccessCount = 0;
      
      for (const api of protectedApis) {
        try {
          const apiResponse = await makeRequest(api, {
            headers: {
              'Authorization': `Bearer ${token}`
            }
          });
          
          if (apiResponse.status === 200) {
            console.log(`✅ ${api.split('/').pop()}: 认证成功`);
            authSuccessCount++;
          } else {
            console.log(`❌ ${api.split('/').pop()}: 认证失败 (${apiResponse.status})`);
          }
          
        } catch (error) {
          console.log(`❌ ${api.split('/').pop()}: 访问失败 - ${error.message}`);
        }
      }
      
      if (authSuccessCount === protectedApis.length) {
        console.log('✅ 所有受保护API认证成功');
        results.authentication = true;
      } else {
        console.log(`⚠️  ${authSuccessCount}/${protectedApis.length} 个API认证成功`);
      }
      
      // 5. 测试前端页面访问（带认证）
      console.log('\n🔍 步骤5: 测试前端页面访问（带认证）...');
      
      const frontendPages = [
        { path: '/dashboard', name: '仪表板' },
        { path: '/clients', name: '客户管理' },
        { path: '/cases', name: '案件管理' },
        { path: '/documents', name: '文档管理' },
        { path: '/users', name: '用户管理' }
      ];
      
      let frontendSuccessCount = 0;
      
      for (const page of frontendPages) {
        try {
          const pageResponse = await makeRequest(`http://localhost:3002${page.path}`, {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Cookie': `token=${token}`
            }
          });
          
          if (pageResponse.status === 200) {
            console.log(`✅ ${page.name}: 页面访问成功`);
            
            // 检查页面是否包含预期内容
            const content = pageResponse.data;
            const hasContent = content.length > 1000; // 简单检查内容长度
            
            if (hasContent) {
              console.log(`   页面内容长度: ${content.length} 字符`);
            }
            
            frontendSuccessCount++;
          } else {
            console.log(`❌ ${page.name}: 页面访问失败 (${pageResponse.status})`);
          }
          
        } catch (error) {
          console.log(`❌ ${page.name}: 访问失败 - ${error.message}`);
        }
      }
      
      if (frontendSuccessCount === frontendPages.length) {
        console.log('✅ 所有前端页面访问成功');
        results.frontendRoutes = true;
      } else {
        console.log(`⚠️  ${frontendSuccessCount}/${frontendPages.length} 个页面访问成功`);
      }
      
    } else {
      console.log('❌ 登录API返回错误');
      results.errors.push('登录API返回错误: ' + JSON.stringify(response.data));
    }
    
  } catch (error) {
    console.error('❌ 登录API测试失败:', error.message);
    results.errors.push('登录API测试失败: ' + error.message);
  }
  
  // 6. 检查前端构建配置
  console.log('\n🔍 步骤6: 检查前端构建配置...');
  
  try {
    // 检查frontend2目录结构
    const frontendFiles = [
      'frontend2/package.json',
      'frontend2/index.html',
      'frontend2/src/main.tsx',
      'frontend2/vite.config.ts'
    ];
    
    let configSuccessCount = 0;
    
    for (const file of frontendFiles) {
      if (fs.existsSync(file)) {
        console.log(`✅ ${file}: 文件存在`);
        configSuccessCount++;
      } else {
        console.log(`❌ ${file}: 文件不存在`);
      }
    }
    
    if (configSuccessCount === frontendFiles.length) {
      console.log('✅ 前端配置文件完整');
    } else {
      console.log(`⚠️  ${configSuccessCount}/${frontendFiles.length} 个配置文件存在`);
    }
    
  } catch (error) {
    console.error('❌ 配置检查失败:', error.message);
  }
  
  // 7. 生成详细测试报告
  console.log('\n📋 详细测试报告');
  console.log('=' .repeat(60));
  
  const testResults = [
    { name: '前端服务运行', value: results.frontendService },
    { name: '登录页面可访问', value: results.loginPage },
    { name: '登录表单存在', value: results.loginForm },
    { name: '登录API工作', value: results.loginApi },
    { name: '认证功能正常', value: results.authentication },
    { name: '前端路由正常', value: results.frontendRoutes },
    { name: '页面内容正常', value: results.pageContent }
  ];
  
  testResults.forEach(test => {
    console.log(`${test.name}: ${test.value ? '✅' : '❌'}`);
  });
  
  const successCount = testResults.filter(t => t.value).length;
  const totalTests = testResults.length;
  
  console.log(`\n📊 测试结果: ${successCount}/${totalTests} 项检查通过`);
  
  if (results.errors.length > 0) {
    console.log('\n❌ 发现的错误:');
    results.errors.forEach((error, index) => {
      console.log(`   ${index + 1}. ${error}`);
    });
  }
  
  console.log('\n📁 生成的测试文件:');
  console.log('   - frontend_index.html: 前端首页内容');
  console.log('   - frontend_login.html: 登录页面内容');
  console.log('   - login_api_response.json: 登录API响应');
  
  if (successCount === totalTests) {
    console.log('\n🎉 所有测试通过！frontend2登录功能完全正常。');
    console.log('💡 建议：可以进行浏览器端UI测试以验证完整的用户登录流程。');
  } else {
    console.log('\n⚠️  部分测试失败，请检查上述错误信息。');
    console.log('💡 建议：修复失败的项目后重新运行测试。');
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
        'User-Agent': 'Law-OA-Detailed-Test/1.0',
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
testFrontendLoginDetailed().catch(console.error);