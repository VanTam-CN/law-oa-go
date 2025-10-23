#!/bin/bash

echo "🚀 Law OA Go 启动中..."

# 停止现有进程
echo "🛑 停止现有服务..."
pkill -f "law-oa-go" 2>/dev/null || true
pkill -f "npm start" 2>/dev/null || true
pkill -f "node.*3003" 2>/dev/null || true
pkill -f "python3.*3003" 2>/dev/null || true
sleep 2

# 启动后端
echo "📦 启动后端服务..."
if [ -f "./main" ]; then
    echo "  使用现有可执行文件"
    nohup ./main > backend.log 2>&1 &
    BACKEND_PID=$!
else
    echo "  编译并启动"
    go build -o main main.go
    nohup ./main > backend.log 2>&1 &
    BACKEND_PID=$!
fi
echo "  后端PID: $BACKEND_PID"
echo $BACKEND_PID > .backend.pid

# 等待后端启动
echo "⏳ 等待后端服务启动..."
sleep 8

# 构建前端（如果需要）
echo "🎨 准备前端服务..."
cd frontend

# 确保dist目录存在
if [ ! -d "./dist" ]; then
    echo "  构建前端项目..."
    npm run build
fi

# 启动前端
echo "  启动静态文件服务器..."
# 使用绝对路径确保dist目录存在
DIST_PATH="$(pwd)/dist"
if [ -d "$DIST_PATH" ]; then
    echo "  使用dist目录: $DIST_PATH"
    nohup python3 -m http.server 3003 --directory "$DIST_PATH" > ../frontend.log 2>&1 &
    FRONTEND_PID=$!
    echo "  前端PID: $FRONTEND_PID"
    echo $FRONTEND_PID > ../.frontend.pid
else
    echo "  ❌ 错误: dist目录不存在"
    echo "  尝试启动开发服务器..."
    nohup npm start > ../frontend.log 2>&1 &
    FRONTEND_PID=$!
    echo "  前端PID: $FRONTEND_PID"
    echo $FRONTEND_PID > ../.frontend.pid
fi

cd ..

echo ""
echo "✅ 服务启动完成！"
echo ""
echo "📱 前端应用: http://localhost:3003"
echo "🔧 后端API: http://localhost:8080"
echo "📊 健康检查: http://localhost:8080/health"
echo ""
echo "📋 查看日志:"
echo "  后端: tail -f backend.log"
echo "  前端: tail -f frontend.log"
echo ""
echo "🛑 停止服务: ./start-fixed-v2.sh stop"
echo ""
