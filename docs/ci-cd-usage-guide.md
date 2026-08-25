# Law OA Go 项目 CI/CD 流水线使用指南

## 🚀 快速开始

### 1. 环境准备

确保你的环境满足以下要求：
- Docker 和 Docker Compose
- Go 1.23.0+
- Git
- curl 和 jq
- 必要的权限

### 2. 初始化设置

```bash
# 克隆项目
git clone <repository-url>
cd law-oa-go

# 设置执行权限
chmod +x scripts/*.sh

# 生成配置文件
./scripts/monitoring_integration.sh --generate-configs

# 查看所有可用脚本
ls -la scripts/
```

## 📋 日常开发流程

### 1. 本地开发测试

```bash
# 运行所有测试
./scripts/run_tests.sh -e dev -t all --coverage

# 快速单元测试
./scripts/run_tests.sh -e dev -t unit -p

# 集成测试
./scripts/run_tests.sh -e dev -t integration

# 查看测试报告
cat test-report.md
```

### 2. 代码提交前检查

```bash
# 完整测试套件
./scripts/run_tests.sh -e dev -t all --coverage --bench -q

# 生成质量报告
./scripts/run_tests.sh -e dev -r pre-commit-report.md

# 安全扫描
./scripts/run_tests.sh -e dev -t security --fuzz
```

### 3. 部署到测试环境

```bash
# 运行完整测试
./scripts/run_tests.sh -e staging -t all -q -s --coverage

# 部署应用
./scripts/deploy.sh -e staging -b pgo -t v1.0.0 -d -s

# 启动监控
./scripts/monitoring.sh -e staging -d 60 -i 30 -o staging_monitoring.log
```

### 4. 生产环境部署

```bash
# 严格模式测试
./scripts/run_tests.sh -e production -t all -q -s --coverage --bench --fuzz

# 生产环境部署
./scripts/deploy.sh -e production -b pgo -t v1.0.0 -d -s

# 全面监控
./scripts/monitoring.sh -e production -d 120 -i 15 --check-docker --check-system --check-database
```

## 🔧 监控工具使用

### 1. 设置监控工具

```bash
# 设置所有监控工具
./scripts/monitoring_integration.sh -e production -t all --setup -d

# 检查服务状态
./scripts/monitoring_integration.sh -e production -C

# 测试集成
./scripts/monitoring_integration.sh -e production --test-integration
```

### 2. 访问监控界面

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686

> 说明：Elasticsearch / Kibana 已退出默认 compose 栈，不再作为 CI/CD 指南里的默认可访问组件。若维护历史部署或 legacy 观测环境，请单独按旧方案处理，且不要把 9200/5601 作为默认入口口径。

### 3. 监控命令示例

```bash
# 基本监控
./scripts/monitoring.sh -e production -d 60 -i 30

# 系统资源监控
./scripts/monitoring.sh -e production --check-system --performance-analysis

# 数据库和缓存监控
./scripts/monitoring.sh -e production --check-database --check-redis

# 自动重启监控
./scripts/monitoring.sh -e production --auto-restart --notifications
```

## 🔄 回滚操作

### 1. 快速回滚

```bash
# 回滚到上一个版本
./scripts/rollback.sh -e production -t previous --backup

# 回滚到稳定版本
./scripts/rollback.sh -e production -t stable --health-check

# 模拟回滚
./scripts/rollback.sh -e staging -t previous --dry-run
```

### 2. 数据回滚

```bash
# 完整回滚（包含数据和配置）
./scripts/rollback.sh -e production -t previous --rollback-data --rollback-config

# 仅回滚配置
./scripts/rollback.sh -e production -t previous --rollback-config

# 仅回滚数据
./scripts/rollback.sh -e production -t previous --rollback-data
```

### 3. 高级回滚

```bash
# 蓝绿回滚
./scripts/rollback.sh -e production -t stable --rollback-strategy blue-green

# 优雅回滚
./scripts/rollback.sh -e production -t stable --rollback-strategy graceful --post-check

# 回滚后监控
./scripts/rollback.sh -e production -t previous --monitor --notification
```

## 📊 报告和分析

### 1. 生成测试报告

```bash
# 完整测试报告
./scripts/run_tests.sh -e production -t all -r comprehensive_test_report.md

# 性能测试报告
./scripts/run_tests.sh -e production -t performance -r performance_report.md

# 安全测试报告
./scripts/run_tests.sh -e production -t security -r security_report.md
```

### 2. 查看监控报告

```bash
# 监控数据备份
./scripts/monitoring_integration.sh -e production --backup

# 恢复监控数据
./scripts/monitoring_integration.sh -e production --restore /path/to/backup

# 检查服务状态
./scripts/monitoring_integration.sh -e production -C
```

### 3. 回滚报告

```bash
# 回滚操作会自动生成报告
# 报告位置：rollback_report_YYYYMMDD_HHMMSS.json

# 查看最新回滚报告
ls -la rollback_report_*.json | tail -1
```

## 🔍 故障排除

### 1. 测试失败

```bash
# 查看详细测试日志
cat test-execution.log

# 运行特定测试
./scripts/run_tests.sh -e dev -t unit --dry-run

# 检查覆盖率
go tool cover -html=test-reports/coverage.html
```

### 2. 部署失败

```bash
# 检查部署日志
cat deployment.log

# 检查容器状态
docker ps -a

# 查看容器日志
docker logs law-oa-app
```

### 3. 监控异常

```bash
# 检查监控服务状态
./scripts/monitoring_integration.sh -e production -C

# 重启监控服务
./scripts/monitoring_integration.sh -e production --restart

# 测试集成
./scripts/monitoring_integration.sh -e production --test-integration
```

### 4. 回滚问题

```bash
# 检查回滚日志
cat /var/log/law-oa-rollback.log

# 验证数据一致性
./scripts/rollback.sh -e production -t stable --data-consistency --dry-run

# 检查备份
ls -la /opt/law-oa-go/backups/
```

## 📈 性能优化

### 1. 测试性能优化

```bash
# 并行测试
./scripts/run_tests.sh -e production -t all -p -j 8

# 跳过耗时测试
./scripts/run_tests.sh -e production -t unit --skip-fuzz --skip-security

# 性能基准测试
./scripts/run_tests.sh -e production -t performance --bench
```

### 2. 监控性能优化

```bash
# 调整监控频率
./scripts/monitoring.sh -e production -d 60 -i 60  # 减少频率

# 选择性监控
./scripts/monitoring.sh -e production --check-docker  # 仅检查Docker

# 禁用自动重启
./scripts/monitoring.sh -e production --no-auto-restart
```

### 3. 回滚性能优化

```bash
# 减少重试次数
./scripts/rollback.sh -e production -t previous --max-retries 2

# 立即回滚
./scripts/rollback.sh -e production -t previous --rollback-strategy immediate

# 跳过健康检查
./scripts/rollback.sh -e production -t previous --no-health-check
```

## 🛠️ 配置自定义

### 1. 自定义测试配置

```bash
# 编辑测试配置文件
vim scripts/run_tests.sh

# 自定义超时时间
./scripts/run_tests.sh -e production -t all --timeout 45m

# 自定义并行任务数
./scripts/run_tests.sh -e production -t all -p -j 12
```

### 2. 自定义监控配置

```bash
# 编辑监控配置文件
vim monitoring/prometheus/config/prometheus.yml
vim monitoring/grafana/config/grafana.ini

# 重新生成配置
./scripts/monitoring_integration.sh --generate-configs
```

### 3. 自定义部署配置

```bash
# 编辑部署配置
vim docker-compose.yml
vim scripts/deploy.sh

# 编辑质量门禁配置
vim scripts/deployment_quality_gate.sh
```

## 📝 最佳实践

### 1. 开发阶段
- 每次提交前运行完整测试
- 保持高测试覆盖率
- 定期检查性能指标
- 及时修复测试失败

### 2. 部署阶段
- 严格遵循质量门禁
- 逐步部署（开发 → 测试 → 生产）
- 监控部署过程
- 准备回滚方案

### 3. 运维阶段
- 持续监控系统状态
- 设置合理的告警阈值
- 定期备份监控数据
- 定期演练回滚流程

### 4. 故障处理
- 建立故障响应流程
- 记录故障处理过程
- 分析故障根本原因
- 完善监控和告警

## 🔗 相关文档

- [CI/CD 流水线完成报告](./ci-cd-pipeline-completion-report.md)
- [CI/CD 测试流水线文档](./ci-cd-testing-pipeline.md)
- [部署质量门禁说明](../scripts/deployment_quality_gate.sh)
- [监控工具集成说明](../scripts/monitoring_integration.sh)

---

**文档维护者：** 开发团队  
**最后更新：** 2024年  
**联系方式：** support@company.com
