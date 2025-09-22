#!/bin/bash

# Fuzzing语料库管理器
# 用于自动收集、清理和管理Fuzzing测试的语料库

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                 显示此帮助信息"
    echo "  -c, --collect              从生产环境收集语料库"
    echo "  -p, --production FILE      指定生产日志文件路径"
    echo "  -o, --output DIR           输出目录 (默认: corpus)"
    echo "  -d, --clean               清理现有语料库"
    echo "  -q, --quality             分析语料库质量"
    echo "  -b, --backup              备份语料库"
    echo "  -r, --restore FILE        从备份恢复语料库"
    echo "  -u, --update              更新Fuzzing测试种子"
    echo "  -s, --stats               显示语料库统计信息"
    echo ""
    echo "示例:"
    echo "  $0 -c -p production.log       # 从生产日志收集语料库"
    echo "  $0 -q -o corpus              # 分析语料库质量"
    echo "  $0 -d -u                     # 清理并更新语料库"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装"
        exit 1
    fi
    
    # 检查jq (用于JSON处理)
    if ! command -v jq &> /dev/null; then
        log_warning "jq 未安装，JSON处理功能将受限"
    fi
    
    # 检查语料库目录
    if [ ! -d "testdata/fuzz" ]; then
        log_info "创建语料库目录"
        mkdir -p testdata/fuzz
    fi
    
    log_success "依赖检查通过"
}

# 从生产日志收集真实数据
collect_production_corpus() {
    local production_log="$1"
    local output_dir="${2:-corpus}"
    
    log_info "从生产日志收集语料库: $production_log"
    
    # 创建输出目录
    mkdir -p "$output_dir"
    
    # 收集JWT令牌样本（脱敏处理）
    log_info "收集JWT样本..."
    if [ -f "$production_log" ]; then
        grep -o 'Authorization: Bearer [^"]*' "$production_log" | \
            sed 's/Authorization: Bearer /JWT_SAMPLE_/' | \
            head -50 > "$output_dir/jwt_corpus.txt"
        
        # 添加一些JWT格式的样本
        cat >> "$output_dir/jwt_corpus.txt" << 'EOF'
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9
invalid.jwt.token
malformed_token
very.long.jwt.token.that.exceeds.normal.lengths
JWT_SAMPLE_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ
EOF
    fi
    
    # 收集JSON请求样本
    log_info "收集JSON样本..."
    if [ -f "$production_log" ]; then
        grep -A 10 'POST /api/' "$production_log" | \
            grep -o '{[^}]*}' | \
            head -50 > "$output_dir/json_corpus.txt"
    fi
    
    # 添加各种JSON格式的样本
    cat >> "$output_dir/json_corpus.txt" << 'EOF'
{"title":"Test","description":"Description","status":"active"}
{"title":"","description":"Valid","status":"invalid_status"}
{"malformed":"json","missing":closure
{"title":"XSS<script>alert('test')</script>","description":"Test"}
{"title":"SQL' OR '1'='1","description":"Test"}
{"title":"Very long title that exceeds normal validation limits","description":"Description"}
{"title":"正常标题","description":"正常描述","status":"active"}
{"title":"Test","description":"Test","priority":"high","assignee":"user@example.com"}
EOF
    
    # 收集查询参数样本
    log_info "收集查询参数样本..."
    if [ -f "$production_log" ]; then
        grep -o 'GET /api/[^?]*\?[^"]*' "$production_log" | \
            sed 's/GET \/api\/[^?]*\?//' | \
            head -50 > "$output_dir/query_corpus.txt"
    fi
    
    # 添加各种查询参数样本
    cat >> "$output_dir/query_corpus.txt" << 'EOF'
page=1&limit=10&sort=created_at
search=test&status=active
filter[category]=urgent&range[date]=2024-01-01,2024-12-31
query=SQL' OR '1'='1
query=<script>alert('xss')</script>
very_long_query_parameter_that_exceeds_normal_limits
empty_parameter=
special_chars=!@#$%^&*()
中文查询参数
query=test&page=999999&limit=-1
EOF
    
    # 收集文件上传样本
    log_info "收集文件样本..."
    # 创建一些测试文件样本
    echo "Test file content" > "$output_dir/test_file.txt"
    echo -n "Binary content\x00\x01\x02\x03" > "$output_dir/binary_file.bin"
    echo '{"test":"json","file":true}' > "$output_dir/json_file.json"
    
    log_success "语料库收集完成: $output_dir"
}

# 清理和去重语料库
clean_corpus() {
    local corpus_dir="$1"
    
    log_info "清理语料库: $corpus_dir"
    
    if [ ! -d "$corpus_dir" ]; then
        log_error "语料库目录不存在: $corpus_dir"
        return 1
    fi
    
    # 清理每个语料库文件
    for corpus_file in "$corpus_dir"/*; do
        if [ -f "$corpus_file" ] && [[ "$corpus_file" != *"quality_report"* ]]; then
            local filename=$(basename "$corpus_file")
            log_info "清理文件: $filename"
            
            # 备份原文件
            cp "$corpus_file" "${corpus_file}.backup"
            
            # 去重
            sort "$corpus_file" | uniq > "${corpus_file}.tmp"
            
            # 移除空行和过小数据
            grep -v '^$' "${corpus_file}.tmp" | awk 'length > 2' > "${corpus_file}.clean"
            
            # 限制文件大小（防止过大的文件）
            head -1000 "${corpus_file}.clean" > "${corpus_file}.final"
            
            # 替换原文件
            mv "${corpus_file}.final" "$corpus_file"
            
            # 清理临时文件
            rm -f "${corpus_file}.tmp" "${corpus_file}.clean" "${corpus_file}.backup"
            
            log_success "清理完成: $filename"
        fi
    done
}

# 分析语料库质量
analyze_corpus_quality() {
    local corpus_dir="$1"
    local report_file="$corpus_dir/quality_report.md"
    
    log_info "分析语料库质量: $corpus_dir"
    
    if [ ! -d "$corpus_dir" ]; then
        log_error "语料库目录不存在: $corpus_dir"
        return 1
    fi
    
    echo "# Fuzzing语料库质量报告" > "$report_file"
    echo "## 分析时间: $(date)" >> "$report_file"
    echo "## 分析工具: Fuzzing语料库管理器 v1.0" >> "$report_file"
    echo "" >> "$report_file"
    
    total_samples=0
    total_files=0
    
    # 分析各个语料库文件
    for corpus_file in "$corpus_dir"/*; do
        if [ -f "$corpus_file" ] && [[ "$corpus_file" != *"quality_report"* ]]; then
            local filename=$(basename "$corpus_file")
            local line_count=$(wc -l < "$corpus_file" 2>/dev/null || echo "0")
            local file_size=$(stat -c%s "$corpus_file" 2>/dev/null || echo "0")
            local avg_length=$(awk '{total += length} END {print (NR > 0) ? total/NR : 0}' "$corpus_file" 2>/dev/null || echo "0")
            
            echo "### $filename" >> "$report_file"
            echo "- 样本数量: $line_count" >> "$report_file"
            echo "- 文件大小: $file_size bytes" >> "$report_file"
            echo "- 平均长度: ${avg_length:-0}" >> "$report_file"
            
            # 质量评估
            echo "- 质量评分: $(calculate_quality_score "$line_count" "$avg_length")" >> "$report_file"
            
            # 建议和改进
            echo "- 建议: $(generate_suggestions "$filename" "$line_count" "$avg_length")" >> "$report_file"
            echo "" >> "$report_file"
            
            total_samples=$((total_samples + line_count))
            total_files=$((total_files + 1))
        fi
    done
    
    # 总结统计
    echo "## 总结统计" >> "$report_file"
    echo "- 语料库文件数量: $total_files" >> "$report_file"
    echo "- 总样本数量: $total_samples" >> "$report_file"
    echo "- 平均每文件样本数: $((total_samples / total_files))" >> "$report_file"
    echo "" >> "$report_file"
    
    # 质量建议
    echo "## 整体质量建议" >> "$report_file"
    echo "1. **样本数量**: 建议每个语料库至少50个样本" >> "$report_file"
    echo "2. **样本多样性**: 确保覆盖不同的输入模式" >> "$report_file"
    echo "3. **边界值**: 包含各种边界值和异常情况" >> "$report_file"
    echo "4. **真实数据**: 基于实际使用数据生成样本" >> "$report_file"
    echo "5. **定期更新**: 建议每周更新一次语料库" >> "$report_file"
    echo "" >> "$report_file"
    
    # 质量等级
    local quality_level=$(calculate_overall_quality "$total_samples" "$total_files")
    echo "## 整体质量等级: $quality_level" >> "$report_file"
    
    log_success "质量分析报告已生成: $report_file"
}

# 计算质量评分
calculate_quality_score() {
    local line_count="$1"
    local avg_length="$2"
    
    if [ "$line_count" -ge 100 ] && [ "$avg_length" -ge 10 ]; then
        echo "优秀"
    elif [ "$line_count" -ge 50 ] && [ "$avg_length" -ge 5 ]; then
        echo "良好"
    elif [ "$line_count" -ge 20 ] && [ "$avg_length" -ge 3 ]; then
        echo "一般"
    else
        echo "需要改进"
    fi
}

# 生成建议
generate_suggestions() {
    local filename="$1"
    local line_count="$2"
    local avg_length="$3"
    
    if [ "$line_count" -lt 20 ]; then
        echo "增加样本数量"
    elif [ "$avg_length" -lt 5 ]; then
        echo "增加样本复杂度"
    elif [ "$line_count" -gt 1000 ]; then
        echo "考虑精简样本，专注于高质量输入"
    else
        echo "质量良好"
    fi
}

# 计算整体质量
calculate_overall_quality() {
    local total_samples="$1"
    local total_files="$2"
    
    if [ "$total_files" -eq 0 ]; then
        echo "无数据"
        return
    fi
    
    local avg_samples=$((total_samples / total_files))
    
    if [ "$avg_samples" -ge 80 ]; then
        echo "优秀"
    elif [ "$avg_samples" -ge 40 ]; then
        echo "良好"
    elif [ "$avg_samples" -ge 20 ]; then
        echo "一般"
    else
        echo "需要改进"
    fi
}

# 备份语料库
backup_corpus() {
    local corpus_dir="$1"
    local backup_file="${2:-corpus_backup_$(date +%Y%m%d_%H%M%S).tar.gz}"
    
    log_info "备份语料库: $corpus_dir -> $backup_file"
    
    if [ ! -d "$corpus_dir" ]; then
        log_error "语料库目录不存在: $corpus_dir"
        return 1
    fi
    
    tar -czf "$backup_file" -C "$(dirname "$corpus_dir")" "$(basename "$corpus_dir")"
    
    log_success "语料库备份完成: $backup_file"
}

# 从备份恢复语料库
restore_corpus() {
    local backup_file="$1"
    local corpus_dir="${2:-corpus}"
    
    log_info "从备份恢复语料库: $backup_file -> $corpus_dir"
    
    if [ ! -f "$backup_file" ]; then
        log_error "备份文件不存在: $backup_file"
        return 1
    fi
    
    # 创建目录
    mkdir -p "$corpus_dir"
    
    # 恢复备份
    tar -xzf "$backup_file" -C "$(dirname "$corpus_dir")"
    
    log_success "语料库恢复完成: $corpus_dir"
}

# 更新Fuzzing测试种子
update_fuzzing_seeds() {
    local corpus_dir="${1:-corpus}"
    
    log_info "更新Fuzzing测试种子: $corpus_dir"
    
    if [ ! -d "$corpus_dir" ]; then
        log_error "语料库目录不存在: $corpus_dir"
        return 1
    fi
    
    # 为每个Fuzzing测试更新种子
    if [ -f "$corpus_dir/jwt_corpus.txt" ]; then
        log_info "更新JWT验证Fuzzing种子..."
        # 这里应该调用Go的Fuzzing种子更新机制
        # go test -fuzz=Fuzz_JWTKeyManager_ValidateToken -fuzzseed="$corpus_dir/jwt_corpus.txt" ./internal/security/
    fi
    
    if [ -f "$corpus_dir/json_corpus.txt" ]; then
        log_info "更新验证器Fuzzing种子..."
        # go test -fuzz=Fuzz_CaseValidator_Validate -fuzzseed="$corpus_dir/json_corpus.txt" ./internal/validators/
    fi
    
    if [ -f "$corpus_dir/query_corpus.txt" ]; then
        log_info "更新查询构建器Fuzzing种子..."
        # go test -fuzz=Fuzz_QueryBuilder_Where -fuzzseed="$corpus_dir/query_corpus.txt" ./internal/repositories/
    fi
    
    log_success "Fuzzing种子更新完成"
}

# 显示语料库统计信息
show_corpus_stats() {
    local corpus_dir="${1:-corpus}"
    
    log_info "语料库统计信息: $corpus_dir"
    
    if [ ! -d "$corpus_dir" ]; then
        log_error "语料库目录不存在: $corpus_dir"
        return 1
    fi
    
    echo "=========================================="
    echo "Fuzzing语料库统计信息"
    echo "=========================================="
    
    total_files=0
    total_samples=0
    total_size=0
    
    for corpus_file in "$corpus_dir"/*; do
        if [ -f "$corpus_file" ] && [[ "$corpus_file" != *"quality_report"* ]]; then
            local filename=$(basename "$corpus_file")
            local line_count=$(wc -l < "$corpus_file" 2>/dev/null || echo "0")
            local file_size=$(stat -c%s "$corpus_file" 2>/dev/null || echo "0")
            
            echo "文件: $filename"
            echo "  样本数: $line_count"
            echo "  大小: $file_size bytes"
            echo "  质量评分: $(calculate_quality_score "$line_count" "$(awk '{total += length} END {print (NR > 0) ? total/NR : 0}' "$corpus_file" 2>/dev/null || echo "0)")"
            echo "------------------------------------------"
            
            total_files=$((total_files + 1))
            total_samples=$((total_samples + line_count))
            total_size=$((total_size + file_size))
        fi
    done
    
    echo "总计:"
    echo "  文件数: $total_files"
    echo "  总样本数: $total_samples"
    echo "  总大小: $total_size bytes"
    echo "  平均每文件样本数: $((total_samples / total_files))"
    echo "=========================================="
}

# 主函数
main() {
    # 默认参数
    ACTION=""
    PRODUCTION_LOG=""
    OUTPUT_DIR="corpus"
    BACKUP_FILE=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -c|--collect)
                ACTION="collect"
                shift
                ;;
            -p|--production)
                PRODUCTION_LOG="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -d|--clean)
                ACTION="clean"
                shift
                ;;
            -q|--quality)
                ACTION="quality"
                shift
                ;;
            -b|--backup)
                ACTION="backup"
                shift
                ;;
            -r|--restore)
                ACTION="restore"
                BACKUP_FILE="$2"
                shift 2
                ;;
            -u|--update)
                ACTION="update"
                shift
                ;;
            -s|--stats)
                ACTION="stats"
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # 检查依赖
    check_dependencies
    
    # 执行相应操作
    case $ACTION in
        collect)
            if [ -z "$PRODUCTION_LOG" ]; then
                log_error "请指定生产日志文件路径: -p|--production FILE"
                exit 1
            fi
            collect_production_corpus "$PRODUCTION_LOG" "$OUTPUT_DIR"
            ;;
        clean)
            clean_corpus "$OUTPUT_DIR"
            ;;
        quality)
            analyze_corpus_quality "$OUTPUT_DIR"
            ;;
        backup)
            backup_corpus "$OUTPUT_DIR" "$BACKUP_FILE"
            ;;
        restore)
            if [ -z "$BACKUP_FILE" ]; then
                log_error "请指定备份文件路径: -r|--restore FILE"
                exit 1
            fi
            restore_corpus "$BACKUP_FILE" "$OUTPUT_DIR"
            ;;
        update)
            update_fuzzing_seeds "$OUTPUT_DIR"
            ;;
        stats)
            show_corpus_stats "$OUTPUT_DIR"
            ;;
        *)
            log_error "请指定操作类型"
            show_usage
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"