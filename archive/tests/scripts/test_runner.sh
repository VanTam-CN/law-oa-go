#!/bin/bash

echo "运行Go单元测试..."
echo "=================="

# 运行不需要数据库的单元测试
echo "1. 运行单元测试 (无数据库依赖)..."
go test -run="TestModel.*" ./internal/models/models_unit_test.go ./internal/models/models.go -v

if [ $? -eq 0 ]; then
    echo "✅ 单元测试通过"
else
    echo "❌ 单元测试失败"
    exit 1
fi

echo ""

# 检查是否安装了SQLite，用于测试
if command -v sqlite3 &> /dev/null; then
    echo "2. 检查SQLite是否可用..."
    echo "✅ SQLite已安装，可用于数据库测试"
else
    echo "⚠️  SQLite未安装，跳过数据库测试"
    echo "   如需运行数据库测试，请安装SQLite或配置MySQL测试数据库"
fi

echo ""
echo "测试完成！"
echo ""
echo "📊 测试覆盖率报告:"
echo "   - 单元测试: ✅ 通过 (模型结构、验证、关系、方法、默认值、时间戳)"
echo "   - 数据库测试: ⚠️  跳过 (需要数据库配置)"
echo "   - 集成测试: ⏳ 待完成"
echo "   - 基准测试: ⏳ 待完成"