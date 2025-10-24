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

# 确保dist目录存在且是最新的
if [ ! -d "./dist" ] || [ "./src" -nt "./dist" ]; then
    echo "  构建前端项目..."
    npm run build
fi

# 启动前端
echo "  启动静态文件服务器..."
# 使用绝对路径确保dist目录存在
DIST_PATH="$(pwd)/dist"
if [ -d "$DIST_PATH" ]; then
    echo "  使用dist目录: $DIST_PATH"
    # 使用Node.js服务器（更稳定）
    node -e "
const http = require('http');
const fs = require('fs');
const path = require('path');

const server = http.createServer((req, res) => {
  let filePath = path.join(__dirname, 'dist', req.url === '/' ? 'index.html' : req.url);

  const extname = path.extname(filePath);
  let contentType = 'text/html';
  switch (extname) {
    case '.js': contentType = 'text/javascript'; break;
    case '.css': contentType = 'text/css'; break;
    case '.json': contentType = 'application/json'; break;
    case '.png': contentType = 'image/png'; break;
    case '.jpg': contentType = 'image/jpg'; break;
    case '.svg': contentType = 'image/svg+xml'; break;
  }

  fs.readFile(filePath, (error, content) => {
    if (error) {
      if(error.code == 'ENOENT') {
        res.writeHead(404);
        res.end('File not found');
      } else {
        res.writeHead(500);
        res.end('Server error');
      }
    } else {
      res.writeHead(200, { 'Content-Type': contentType });
      res.end(content, 'utf-8');
    }
  });
});

server.listen(3003, () => {
  console.log('Frontend server running on http://localhost:3003');
});
" > ../frontend.log 2>&1 &
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
echo "🛑 停止服务: ./start-final.sh stop"
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
    if curl -s http://localhost:8080/health >/dev/null 2>&1 || curl -s http://localhost:8080/api/v1/health >/dev/null 2>&1; then
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

    # 测试API路径匹配
    echo "测试API路径..."
    if curl -s http://localhost:8080/api/v1 >/dev/null 2>&1; then
        echo "  ✅ API路径 /api/v1 可访问"
    else
        echo "  ❌ API路径 /api/v1 不可访问"
    fi
fi
