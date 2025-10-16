#!/bin/bash

# 构建前端应用的脚本

echo "Building frontend application..."

# 检查是否已安装依赖
if [ ! -d "node_modules" ]; then
  echo "Installing dependencies..."
  npm install
fi

# 构建应用
echo "Running build..."
npm run build

# 检查构建是否成功
if [ $? -eq 0 ]; then
  echo "Build completed successfully!"
  echo "Output files are in the 'build' directory"
else
  echo "Build failed!"
  exit 1
fi