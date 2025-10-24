#!/bin/bash
# 快速启动脚本 - Law OA Go
# 简化版启动脚本，适用于快速开发测试

echo "🚀 Law OA Go 快速启动"
echo "======================"

# 检查是否已有进程在运行
if lsof -i :8080 >/dev/null 2>&1; then
    echo "⚠️  端口8080已被占用，正在停止现有进程..."
    lsof -ti:8080 | xargs kill -9 2>/dev/null || true
    sleep 2
fi

if lsof -i :3003 >/dev/null 2>&1; then
    echo "⚠️  端口3003已被占用，正在停止现有进程..."
    lsof -ti:3003 | xargs kill -9 2>/dev/null || true
    sleep 2
fi

# 启动后端
echo "📦 启动后端服务..."
if [[ -f "./main" ]]; then
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

# 等待后端启动
echo "⏳ 等待后端服务启动..."
sleep 5

# 启动前端
echo "🎨 启动前端服务..."
cd frontend

if [[ -d "./dist" ]]; then
    echo "  使用静态文件服务器"
    if command -v python3 >/dev/null 2>&1; then
        nohup python3 -m http.server 3003 --directory dist > ../frontend.log 2>&1 &
        FRONTEND_PID=$!
    else
        echo "  启动开发服务器"
        nohup npm start > ../frontend.log 2>&1 &
        FRONTEND_PID=$!
    fi
else
    echo "  启动开发服务器"
    nohup npm start > ../frontend.log 2>&1 &
    FRONTEND_PID=$!
fi
echo "  前端PID: $FRONTEND_PID"

cd ..

# 保存PID
echo $BACKEND_PID > .backend.pid
echo $FRONTEND_PID > .frontend.pid

echo ""
echo "✅ 服务启动完成！"
echo ""
echo "📱 前端: http://localhost:3003"
echo "🔧 后端: http://localhost:8080"
echo "📊 健康检查: http://localhost:8080/health"
echo ""
echo "📋 查看日志:"
echo "  后端: tail -f backend.log"
echo "  前端: tail -f frontend.log"
echo ""
echo "🛑 停止服务:"
echo "  kill $BACKEND_PID $FRONTEND_PID"
echo "  或者运行: ./quick-start.sh stop"
echo ""

# 如果传入stop参数，则停止服务
if [[ "$1" == "stop" ]]; then
    echo "🛑 停止所有服务..."
    if [[ -f ".backend.pid" ]]; then
        kill $(cat .backend.pid) 2>/dev/null || true
        rm -f .backend.pid
    fi
    if [[ -f ".frontend.pid" ]]; then
        kill $(cat .frontend.pid) 2>/dev/null || true
        rm -f .frontend.pid
    fi
    echo "✅ 服务已停止"
fi