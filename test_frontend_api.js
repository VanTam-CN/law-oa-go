#!/usr/bin/env node

const http = require('http');

// 模拟前端API调用，检查token传递
async function testFrontendAPICalls() {
  console.log('🔍 测试前端API调用和token传递...\n');
  
  // 1. 获取token
  const loginData = {
    email: 'lawyer1@lawoa.com',
    password: 'password'
  };
  
  const postData = JSON.stringify(loginData);
  
  const loginOptions = {
    hostname: 'localhost',
    port: 8080,
    path: '/api/v1/auth/login',
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Content-Length': Buffer.byteLength(postData)
    }
  };
  
  try {
    const loginResponse = await new Promise((resolve, reject) => {
      const req = http.request(loginOptions, (res) => {
        let data = '';
        res.on('data', (chunk) => data += chunk);
        res.on('end', () => resolve({ status: res.statusCode, data: JSON.parse(data) }));
      });
      req.on('error', reject);
      req.write(postData);
      req.end();
    });
    
    if (loginResponse.status === 200 && loginResponse.data.success) {
      const token = loginResponse.data.data.token;
      console.log('✅ 登录成功，获取到token');
      
      // 2. 模拟前端通过Vite代理访问API
      const frontendAPITests = [
        {
          name: '通过前端代理访问仪表盘统计',
          url: '/api/v1/dashboard/statistics',
          useProxy: true
        },
        {
          name: '直接访问后端仪表盘统计',
          url: '/api/v1/dashboard/statistics',
          useProxy: false
        },
        {
          name: '通过前端代理访问用户资料',
          url: '/api/v1/users/profile',
          useProxy: true
        },
        {
          name: '直接访问后端用户资料',
          url: '/api/v1/users/profile',
          useProxy: false
        }
      ];
      
      console.log('\n📋 测试前端API调用方式:');
      
      for (const test of frontendAPITests) {
        console.log(`\n测试: ${test.name}`);
        
        const options = {
          hostname: test.useProxy ? 'localhost' : 'localhost',
          port: test.useProxy ? 3002 : 8080,
          path: test.url,
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        };
        
        try {
          const response = await new Promise((resolve, reject) => {
            const req = http.request(options, (res) => {
              let data = '';
              res.on('data', (chunk) => data += chunk);
              res.on('end', () => {
                try {
                  const parsedData = data ? JSON.parse(data) : null;
                  resolve({ status: res.statusCode, data: parsedData });
                } catch (e) {
                  resolve({ status: res.statusCode, data: data });
                }
              });
            });
            req.on('error', reject);
            req.end();
          });
          
          console.log(`   状态码: ${response.status}`);
          if (response.status === 200) {
            console.log(`✅ ${test.name}: 成功`);
            if (response.data && response.data.data) {
              console.log(`   📊 有数据返回`);
            } else {
              console.log(`   ⚠️  无数据返回`);
            }
          } else if (response.status === 401) {
            console.log(`❌ ${test.name}: 认证失败 (401)`);
          } else if (response.status === 404) {
            console.log(`❌ ${test.name}: 接口不存在 (404)`);
          } else {
            console.log(`❌ ${test.name}: 其他错误 (${response.status})`);
            if (response.data) {
              console.log(`   错误: ${response.data.error?.message || response.data.message || '未知错误'}`);
            }
          }
        } catch (error) {
          console.log(`❌ ${test.name}: 连接失败 - ${error.message}`);
        }
      }
      
      // 3. 测试不带token的情况
      console.log('\n📋 测试不带token的访问:');
      
      const noTokenTests = [
        {
          name: '不带token访问仪表盘统计',
          url: '/api/v1/dashboard/statistics'
        }
      ];
      
      for (const test of noTokenTests) {
        console.log(`\n测试: ${test.name}`);
        
        const options = {
          hostname: 'localhost',
          port: 8080,
          path: test.url,
          method: 'GET',
          headers: {
            'Content-Type': 'application/json'
            // 不带Authorization头
          }
        };
        
        try {
          const response = await new Promise((resolve, reject) => {
            const req = http.request(options, (res) => {
              let data = '';
              res.on('data', (chunk) => data += chunk);
              res.on('end', () => {
                try {
                  const parsedData = data ? JSON.parse(data) : null;
                  resolve({ status: res.statusCode, data: parsedData });
                } catch (e) {
                  resolve({ status: res.statusCode, data: data });
                }
              });
            });
            req.on('error', reject);
            req.end();
          });
          
          if (response.status === 401) {
            console.log(`✅ ${test.name}: 正确拒绝访问 (401)`);
          } else {
            console.log(`❌ ${test.name}: 异常状态码 (${response.status})`);
          }
        } catch (error) {
          console.log(`❌ ${test.name}: 连接失败 - ${error.message}`);
        }
      }
      
      console.log('\n🎉 前端API调用测试完成！');
      
    } else {
      console.log('❌ 登录失败:', loginResponse.data);
    }
  } catch (error) {
    console.log('❌ 测试失败:', error.message);
  }
}

testFrontendAPICalls();