#!/bin/bash

# 启动前端开发服务器的脚本

# 检查是否已安装依赖
if [ ! -d "node_modules" ]; then
  echo "Installing dependencies..."
  npm install
fi

# 启动开发服务器
echo "Starting frontend development server..."
npm start