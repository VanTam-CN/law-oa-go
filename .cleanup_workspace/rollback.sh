#!/bin/bash

# 文件回滚脚本
# 用于从备份目录恢复文件到项目

BACKUP_DIR="../backup/2025-01-06_cleanup"

echo "=== 文件回滚脚本 ==="
echo "备份目录: $BACKUP_DIR"
echo ""
echo "⚠️  警告: 此操作将把备份文件复制回项目目录"
echo ""
echo "使用方法:"
echo "  1. 恢复所有备份文件:"
echo "     cp -r $BACKUP_DIR/backup_files/* ../"
echo "     cp -r $BACKUP_DIR/executables/* ../"
echo "     cp -r $BACKUP_DIR/logs/* ../"
echo "     cp -r $BACKUP_DIR/diagnostics/* ../"
echo "     cp -r $BACKUP_DIR/scripts/* ../"
echo "     cp -r $BACKUP_DIR/temp/* ../"
echo ""
echo "  2. 恢复单个文件:"
echo "     cp $BACKUP_DIR/<子目录>/<文件名> ../<原位置>/"
echo ""
echo "  3. 查看备份文件列表:"
echo "     ls -la $BACKUP_DIR/"
echo ""
echo "备份分类:"
echo "  - backup_files/  .bak 备份文件"
echo "  - executables/   可执行文件"
echo "  - logs/          日志文件"
echo "  - diagnostics/   诊断/测试文件"
echo "  - scripts/       测试脚本"
echo "  - temp/          临时文件"
echo ""

# 交互式恢复菜单
read -p "是否查看备份文件列表? (y/n): " choice
if [ "$choice" = "y" ]; then
    echo ""
    echo "=== 备份文件列表 ==="
    echo ""
    echo "【backup_files/】"
    ls -1 "$BACKUP_DIR/backup_files/" 2>/dev/null
    echo ""
    echo "【executables/】"
    ls -1 "$BACKUP_DIR/executables/" 2>/dev/null
    echo ""
    echo "【logs/】"
    ls -1 "$BACKUP_DIR/logs/" 2>/dev/null
    echo ""
    echo "【diagnostics/】"
    ls -1 "$BACKUP_DIR/diagnostics/" 2>/dev/null | head -10
    echo "  ... (共 $(ls -1 "$BACKUP_DIR/diagnostics/" 2>/dev/null | wc -l) 个文件)"
fi
