// 测试律师API的前端调用
const fetch = require('node-fetch');

async function testLawyerAPI() {
    try {
        // 1. 先登录获取token
        console.log('1. 登录获取token...');
        const loginResponse = await fetch('http://localhost:8080/api/v1/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                email: 'admin@lawoa.com',
                password: 'admin123'
            })
        });

        const loginData = await loginResponse.json();
        console.log('登录响应:', JSON.stringify(loginData, null, 2));

        if (!loginData.success || !loginData.data.token) {
            throw new Error('登录失败');
        }

        const token = loginData.data.token;
        console.log('✅ 登录成功，获取到token');

        // 2. 调用律师列表API
        console.log('\n2. 调用律师列表API...');
        const lawyersResponse = await fetch('http://localhost:8080/api/v1/lawfirm/lawyers?page=1&page_size=9999', {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });

        const lawyersData = await lawyersResponse.json();
        console.log('律师API响应:', JSON.stringify(lawyersData, null, 2));

        // 3. 模拟前端数据处理逻辑
        console.log('\n3. 模拟前端数据处理...');
        let lawyerData = [];
        if (lawyersData?.data) {
            lawyerData = lawyersData.data;
        } else if (Array.isArray(lawyersData)) {
            lawyerData = lawyersData;
        }

        console.log(`提取到律师数据数量: ${lawyerData.length}`);

        if (lawyerData.length > 0) {
            const formattedLawyers = lawyerData.map((lawyer) => ({
                id: lawyer.id.toString(),
                name: lawyer.name,
                level: 'SENIOR',
                specialties: ['法律咨询'],
                email: lawyer.email || '',
                phone: lawyer.phone || ''
            }));

            console.log('✅ 格式化后的律师数据:');
            console.log(formattedLawyers.map(l => `${l.name} (${l.id})`).join(', '));

            // 4. 模拟下拉框显示数据
            console.log('\n4. 模拟下拉框选项:');
            formattedLawyers.forEach(lawyer => {
                console.log(`  <Option key="${lawyer.id}" value="${lawyer.id}">`);
                console.log(`    ${lawyer.name} - ${lawyer.email}`);
                console.log(`  </Option>`);
            });

        } else {
            console.log('❌ 没有找到律师数据');
        }

    } catch (error) {
        console.error('❌ 测试失败:', error);
    }
}

testLawyerAPI();