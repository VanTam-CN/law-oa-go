#!/bin/bash

echo "🚀 Law OA Go 启动中..."

# 停止现有进程
echo "🛑 停止现有服务..."
pkill -f "law-oa-go" 2>/dev/null || true
pkill -f "npm start" 2>/dev/null || true
pkill -f "node.*3003" 2>/dev/null || true
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

# 启动前端
echo "🎨 启动前端服务..."
cd frontend

if [ -d "./dist" ]; then
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
echo $FRONTEND_PID > ../.frontend.pid

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
echo "🛑 停止服务: ./start-fixed.sh stop"
echo ""

# 检查参数
if [ "$1" = "stop" ]; then
    echo "🛑 停止所有服务..."
    if [ -f ".backend.pid" ]; then
        kill $(cat .backend.pid) 2>/dev/null || true
        rm -f .backend.pid
    fi
    if [ -f ".frontend.pid" ]; then
        kill $(cat .frontend.pid) 2>/dev/null || true
        rm -f .frontend.pid
    fi
    echo "✅ 服务已停止"
fi

if [ "$1" = "status" ]; then
    echo "📊 服务状态:"
    if [ -f ".backend.pid" ] && ps -p $(cat .backend.pid) > /dev/null 2>&1; then
        echo "  ✅ 后端服务运行中 (PID: $(cat .backend.pid))"
    else
        echo "  ❌ 后端服务未运行"
    fi

    if [ -f ".frontend.pid" ] && ps -p $(cat .frontend.pid) > /dev/null 2>&1; then
        echo "  ✅ 前端服务运行中 (PID: $(cat .frontend.pid))"
    else
        echo "  ❌ 前端服务未运行"
    fi

    echo ""
    echo "🔌 端口占用:"
    lsof -i :8080 2>/dev/null || echo "  端口8080: 未占用"
    lsof -i :3003 2>/dev/null || echo "  端口3003: 未占用"
fi

if [ "$1" = "logs" ]; then
    echo "📋 最新日志:"
    echo ""
    echo "🔧 后端日志 (最后20行):"
    echo "----------------------------------------"
    tail -20 backend.log 2>/dev/null || echo "无后端日志"
    echo ""
    echo "🎨 前端日志 (最后20行):"
    echo "----------------------------------------"
    tail -20 frontend.log 2>/dev/null || echo "无前端日志"
fi

if [ "$1" = "test" ]; then
    echo "🧪 测试服务连接..."

    # 测试后端
    echo "测试后端健康检查..."
    if curl -s http://localhost:8080/health >/dev/null; then
        echo "  ✅ 后端服务正常"
    else
        echo "  ❌ 后端服务异常"
    fi

    # 测试前端
    echo "测试前端服务..."
    if curl -s http://localhost:3003 >/dev/null; then
        echo "  ✅ 前端服务正常"
    else
        echo "  ❌ 前端服务异常"
    fi

    # 测试API
    echo "测试API接口..."
    if curl -s http://localhost:8080/api/v1/health >/dev/null; then
        echo "  ✅ API接口正常"
    else
        echo "  ❌ API接口异常"
    fi
fi