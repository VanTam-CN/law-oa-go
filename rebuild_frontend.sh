#!/bin/bash

echo "🔧 重新编译前端代码..."

cd frontend

echo "📦 安装依赖..."
npm install

echo "🧹 清理旧的编译文件..."
rm -rf dist/
rm -rf build/

echo "🔨 重新编译..."
npm run build

echo "✅ 前端重新编译完成！"
echo ""
echo "💡 下一步："
echo "1. 刷新浏览器页面 (Ctrl+F5 或 Cmd+Shift+R)"
echo "2. 清除浏览器缓存"
echo "3. 重新测试冲突检测功能"