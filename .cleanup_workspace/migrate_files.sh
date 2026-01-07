#!/bin/bash

# 文件迁移脚本
# 注意: 此脚本记录了 2025-01-06 已完成的迁移操作

BACKUP_DIR="../backup/2025-01-06_cleanup"

echo "=== 文件迁移报告 ==="
echo "迁移时间: 2025-01-06"
echo "备份目录: $BACKUP_DIR"
echo ""

# 显示已迁移的文件分类
echo "已迁移文件分类:"
echo ""

echo "【备份文件 .bak】"
ls -1 "$BACKUP_DIR/backup_files/" 2>/dev/null | wc -l | xargs echo "  文件数:"
echo ""

echo "【可执行文件】"
ls -1 "$BACKUP_DIR/executables/" 2>/dev/null | wc -l | xargs echo "  文件数:"
du -sh "$BACKUP_DIR/executables/" 2>/dev/null | awk '{print "  大小:", $1}'
echo ""

echo "【日志文件】"
ls -1 "$BACKUP_DIR/logs/" 2>/dev/null | wc -l | xargs echo "  文件数:"
echo ""

echo "【诊断/测试文件】"
ls -1 "$BACKUP_DIR/diagnostics/" 2>/dev/null | wc -l | xargs echo "  文件数:"
echo ""

echo "【脚本文件】"
ls -1 "$BACKUP_DIR/scripts/" 2>/dev/null | wc -l | xargs echo "  文件数:"
echo ""

echo "【临时文件】"
ls -1 "$BACKUP_DIR/temp/" 2>/dev/null | wc -l | xargs echo "  文件数:"
echo ""

echo "✅ 文件迁移已于 2025-01-06 完成"
echo "详细日志请查看: $BACKUP_DIR/MIGRATION_LOG.txt"
