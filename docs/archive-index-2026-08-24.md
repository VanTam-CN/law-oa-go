# 归档索引 - 2026-08-24

本文档记录 2026-08-24 归档审查后恢复回仓库的内容与校验方式，便于后续追踪。

## 归档路径

- 归档根目录：`/Users/mac/Desktop/FT/law-oa-go-archive-20260824`
- 恢复来源：`/Users/mac/Desktop/FT/law-oa-go-archive-20260824/reports/image2-feature-gallery`
- 恢复来源：`/Users/mac/Desktop/FT/law-oa-go-archive-20260824/.claude/steering`
- 归档来源：`/Users/mac/Desktop/FT/law-oa-go-archive-20260824/deployments/docker-compose.prod.yml`

## 归档类别

- 已跟踪的大型制品目录：`reports/image2-feature-gallery`
- 已跟踪的 AI steering 目录：`.claude/steering`
- Playwright 报告保留策略：仅保留最终归档版与可复核证据，不保留临时草稿或一次性调试输出
- `SHA256SUMS.txt` 是归档锚点文件；其自身的 SHA-256 值为 `373c87009a9cc8b374091fac2e0505255ff8295208e8b79246287c974ac43491`
- `frontend/playwright-report/index.html` 保留为当前工作树的生成验收证据；它属于本次工作树的可复核产物，重建只会产生噪声，不代表生产验收结论

## 刻意保留和恢复内容

- 已恢复：`reports/image2-feature-gallery/`
- 已恢复：`.claude/steering/product.md`
- 已恢复：`.claude/steering/structure.md`
- 已恢复：`.claude/steering/tech.md`
- 刻意未恢复：归档里的其余 `.claude`、`.spec-workflow`、`.bmad-core`、`services/document-service`、`migrate` 等目录
- 刻意移除并归档：旧 `deployments/docker-compose.prod.yml`。该入口仍强制 MySQL + Redis，并引用不存在的根 `Dockerfile.prod`，与当前 PostgreSQL 默认栈和可选 cache/observability profile 冲突；生产启动脚本已改为使用根 `docker-compose.yml`

## 恢复方式

- 仅做原路径复制，不移动归档副本
- 恢复命令使用 `cp -R` 将归档目录复制回仓库对应路径
- `.gitignore` 仅保留对 `/.claude/reports/` 的忽略，不再屏蔽整个 `/.claude/`

## 最小恢复示例

- 先核对归档清单中的 `manifest` 路径、目标文件是否仍缺失，以及工作区里是否已有同名改动
- 逐项恢复，不使用一键覆盖脚本；恢复时按同级归档路径复制回仓库对应位置

```bash
cp -R "/Users/mac/Desktop/FT/law-oa-go-archive-20260824/<manifest-path>" "./<manifest-path>"
```

- 恢复后再复核 `git status`、`rg` 和必要的校验和，确认只回流了被批准的条目

## 校验命令

- `rg -n "reports/image2-feature-gallery|\\.claude/steering|\\.claude/reports" CLAUDE.md docs/batch-01-core-intake-iteration-prd.md .gitignore`
- `rg -n "law-oa-go|migrate\\\\main\\.go|go run \\.|bin/law-oa-go" run_backend.sh scripts/performance-test.sh scripts/verify_postgres_data scripts/start.bat scripts/setup_legal_statutes.sh Makefile`
- `rg -n "deployments/docker-compose.prod.yml" scripts README.md docs Makefile .github`
- `sha256sum /Users/mac/Desktop/FT/law-oa-go-archive-20260824/README.md`

## 基线结果

- `reports/image2-feature-gallery/` 已回仓库原路径
- `.claude/steering/` 已回仓库原路径且仅恢复三份 steering 文件
- `scripts/setup_legal_statutes.sh` 中的 `migrate_legal_statutes` 为脚本内生成，不是外部死引用
- `scripts/start.bat` 的迁移入口已切换为 `./cmd/migrate -command bootstrap`
- 归档 `README.md` 已追加恢复说明，且归档校验文件同步更新
- 这些结论仅表示归档审查与引用修复完成，不表示生产环境已验证

## 范围说明

- 本次恢复未包含当前工作区里其他用户正在进行的修改
- 本文档只记录归档审查、恢复和引用消除结果，不作为生产放行证明
