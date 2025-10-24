// 前端参数调试脚本
// 在浏览器控制台中运行此脚本来调试冲突检测参数

console.log('🔍 前端冲突检测参数调试...');

// 模拟前端冲突检测调用
async function debugConflictCheck() {
    // 检查当前用户信息
    const userInfo = localStorage.getItem('user') || sessionStorage.getItem('user');
    console.log('👤 当前用户信息:', userInfo);
    
    // 检查认证令牌
    const token = localStorage.getItem('auth_token') || localStorage.getItem('token');
    console.log('🔑 认证令牌:', token ? token.substring(0, 50) + '...' : '未找到');
    
    // 模拟冲突检测请求
    const testRequest = {
        clientId: "57",
        clientName: "字节跳动科技有限公司", 
        caseName: "字节跳动诉腾讯垄断纠纷案",
        caseType: "commercial",
        clientType: "COMPANY",
        otherParties: ["腾讯"],
        searchYears: 5,
        includeCorporateRelations: true,
        searchDepth: "DEEP",
        userId: "45", // 张伟律师
        requestTime: new Date().toISOString(),
        causeOfAction: "垄断纠纷"
    };
    
    console.log('📤 测试请求参数:', testRequest);
    
    try {
        // 发送测试请求
        const response = await fetch('/api/v1/conflict/check', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(testRequest)
        });
        
        console.log('📥 响应状态:', response.status);
        
        const result = await response.json();
        console.log('📋 响应数据:', result);
        
        // 分析响应
        if (result.success) {
            if (result.data) {
                console.log('✅ 包含data字段');
                console.log('🔍 HasConflict:', result.data.hasConflict);
                console.log('📁 冲突案例数量:', result.data.conflictCases?.length || 0);
                
                if (result.data.hasConflict && result.data.conflictCases?.length > 0) {
                    console.log('⚠️ 检测到冲突案例:');
                    result.data.conflictCases.forEach((conflict, index) => {
                        console.log(`   ${index + 1}. ${conflict.caseName} (${conflict.riskLevel})`);
                    });
                } else {
                    console.log('ℹ️ 未检测到冲突');
                }
            } else {
                console.error('❌ 缺少data字段 - 这是问题所在！');
            }
        } else {
            console.error('❌ API调用失败:', result.error);
        }
        
    } catch (error) {
        console.error('❌ 请求失败:', error);
    }
}

// 检查前端冲突检测服务
function checkConflictService() {
    console.log('🔍 检查前端冲突检测服务...');
    
    // 检查是否有ConflictCheckService
    if (typeof window.ConflictCheckService !== 'undefined') {
        console.log('✅ ConflictCheckService 已加载');
    } else {
        console.log('❌ ConflictCheckService 未找到');
    }
    
    // 检查API配置
    console.log('🔧 API配置检查:');
    console.log('   Base URL:', window.location.origin);
    console.log('   Current Path:', window.location.pathname);
}

// 运行调试
console.log('=== 开始前端调试 ===');
checkConflictService();
debugConflictCheck();

// 提供手动测试函数
window.testConflictCheck = debugConflictCheck;
console.log('💡 可以手动调用: testConflictCheck()');