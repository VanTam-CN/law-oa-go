#!/usr/bin/env node

const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080/api/v1';

async function testUserLogin() {
    console.log('=== 测试用户登录功能 ===');

    const loginData = {
        username: 'admin_test',
        password: 'Admin123'
    };

    try {
        console.log('尝试登录...');
        console.log('用户名:', loginData.username);
        console.log('密码:', loginData.password);

        const response = await axios.post(`${API_BASE_URL}/auth/login`, loginData);

        console.log('\n✅ 登录成功！');
        console.log('响应状态:', response.status);
        console.log('用户数据:', JSON.stringify(response.data.data, null, 2));

        if (response.data.data.token) {
            console.log('\n🔑 Token获取成功:', response.data.data.token.substring(0, 50) + '...');
        }

    } catch (error) {
        console.error('\n❌ 登录失败:');
        if (error.response) {
            console.error('状态码:', error.response.status);
            console.error('错误信息:', error.response.data);
        } else {
            console.error('网络错误:', error.message);
        }
    }
}

// 测试用户信息获取
async function testGetUserInfo() {
    console.log('\n=== 测试获取用户信息 ===');

    try {
        // 首先登录获取token
        const loginResponse = await axios.post(`${API_BASE_URL}/auth/login`, {
            username: 'admin_test',
            password: 'Admin123'
        });

        const token = loginResponse.data.data.token;

        // 使用token获取用户信息
        const userResponse = await axios.get(`${API_BASE_URL}/auth/me`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        console.log('✅ 用户信息获取成功！');
        console.log('用户数据:', JSON.stringify(userResponse.data.data, null, 2));

    } catch (error) {
        console.error('\n❌ 获取用户信息失败:');
        if (error.response) {
            console.error('状态码:', error.response.status);
            console.error('错误信息:', error.response.data);
        } else {
            console.error('网络错误:', error.message);
        }
    }
}

// 主函数
async function main() {
    console.log('开始测试用户登录功能...\n');

    await testUserLogin();
    await testGetUserInfo();

    console.log('\n测试完成！');
}

main().catch(console.error);