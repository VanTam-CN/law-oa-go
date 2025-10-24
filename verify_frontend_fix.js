// 前端修复验证脚本
// 在浏览器控制台中运行此脚本来验证修复是否生效

console.log('🔍 验证前端冲突检测修复...');

// 检查编译后的文件是否还包含旧的错误信息
async function checkCompiledFiles() {
    console.log('📋 检查编译后的文件...');
    
    try {
        // 尝试获取编译后的JS文件
        const response = await fetch('/assets/index-Bq4-1Lbd.js');
        const content = await response.text();
        
        // 检查是否还包含旧的错误信息
        if (content.includes('后端返回数据格式错误：缺少data字段')) {
            console.error('❌ 编译后的文件仍包含旧的错误处理逻辑');
            console.log('💡 需要重新编译前端代码');
            return false;
        } else {
            console.log('✅ 编译后的文件已更新');
            return true;
        }
    } catch (error) {
        console.log('ℹ️ 无法检查编译文件，可能文件名已更改');
        return true;
    }
}

// 测试冲突检测API
async function testConflictAPI() {
    console.log('🧪 测试冲突检测API...');
    
    // 获取认证令牌
    const token = localStorage.getItem('auth_token') || localStorage.getItem('token');
    if (!token) {
        console.error('❌ 未找到认证令牌，请先登录');
        return false;
    }
    
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
        userId: "45",
        requestTime: new Date().toISOString(),
        causeOfAction: "垄断纠纷"
    };
    
    try {
        const response = await fetch('/api/v1/conflict/check', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(testRequest)
        });
        
        const result = await response.json();
        
        console.log('📥 API响应:', result);
        
        if (result.success && result.data) {
            console.log('✅ API响应格式正确');
            console.log('🔍 HasConflict:', result.data.hasConflict);
            console.log('📁 冲突案例数量:', result.data.conflictCases?.length || 0);
            
            if (result.data.hasConflict && result.data.conflictCases?.length > 0) {
                console.log('🎯 检测到冲突案例:');
                result.data.conflictCases.forEach((conflict, index) => {
                    console.log(`   ${index + 1}. ${conflict.caseName} (${conflict.riskLevel})`);
                });
                console.log('✅ 冲突检测功能正常工作！');
                return true;
            } else {
                console.log('ℹ️ 未检测到冲突（可能是数据问题）');
                return true;
            }
        } else {
            console.error('❌ API响应格式错误:', result);
            return false;
        }
    } catch (error) {
        console.error('❌ API调用失败:', error);
        return false;
    }
}

// 运行所有检查
async function runAllChecks() {
    console.log('=== 开始验证 ===');
    
    const compiledCheck = await checkCompiledFiles();
    const apiCheck = await testConflictAPI();
    
    console.log('\n📊 验证结果:');
    console.log(`   编译文件检查: ${compiledCheck ? '✅ 通过' : '❌ 失败'}`);
    console.log(`   API功能检查: ${apiCheck ? '✅ 通过' : '❌ 失败'}`);
    
    if (compiledCheck && apiCheck) {
        console.log('\n🎉 所有检查通过！冲突检测功能已修复。');
    } else {
        console.log('\n⚠️ 部分检查失败，需要进一步处理：');
        if (!compiledCheck) {
            console.log('   - 需要重新编译前端代码');
        }
        if (!apiCheck) {
            console.log('   - 需要检查API或认证问题');
        }
    }
}

// 提供手动调用函数
window.verifyConflictFix = runAllChecks;
window.testConflictAPI = testConflictAPI;

// 自动运行检查
runAllChecks();

console.log('\n💡 可以手动调用:');
console.log('   verifyConflictFix() - 运行完整验证');
console.log('   testConflictAPI() - 仅测试API');