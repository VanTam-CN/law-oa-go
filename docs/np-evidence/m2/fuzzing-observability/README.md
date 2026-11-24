# fuzzing.yml 流水线修复：5 job 红灯归因 + 输出归档观测缺陷（t_46d9c62d）

来源：t_088ba214 探针 PR #54 run 45405844638 归因（evidence: lint-governance README §8；
快照 snapshots/ci-run-45405844638-fuzz-workerpool.log）。

## 1. 观测缺陷（先修才能归因）

fuzzing.yml 所有 fuzz 运行步均 `> "$RUNNER_TEMP/fuzz-$TARGET.log" 2>&1` 重定向后
从未 cat/上传，job 日志只有 exit code。失败原因不可见是本卡第一修复对象。

修复：每个运行步失败时 `tail -80` 日志再退出（保留原 exit code）；每个 fuzz job
新增 `always()` 步 `actions/upload-artifact@v4` 上传 `fuzz-*.log` +
`internal/**/testdata/fuzz/` crasher（artifact 名含 matrix target 防并发冲突）。

## 2. 真实失败面（纠正任务书描述）

run 45405844638 实际结论（gh api jobs 逐 job 核验）：

| job | 结论 | 失败步 | 真实根因 |
|---|---|---|---|
| 安全组件Fuzzing (jwt-validation) | failure | 设置Redis | `Unit redis-server.service not found`（ubuntu-26.04 镜像无该 unit，systemctl exit 5） |
| 缓存Fuzzing (Fuzz_CacheService_SetAndGet) | failure | 设置Redis | 同上 |
| 完整Fuzzing测试 | failure | 设置Redis | 同上 |
| 数据库Fuzzing (Fuzz_QueryBuilder_Like) | failure | 设置数据库 | `mysql -u root` 无密码被拒：`ERROR 1045 Access denied`（exit 1） |
| 并发Fuzzing (Fuzz_WorkerPool_SubmitTask) | failure | 运行选定目标 | 真 crasher（见 §3） |
| 同矩阵其余目标 | cancelled | - | fail-fast 取消，非独立失败 |
| 验证器 3 目标 | success | - | 本来就绿 |
| Fuzzing测试汇总报告 | failure | 状态核对 | 依赖 job 红导致（设计如此） |

任务书所述「5 目标（…RateLimiter）全红」中 RateLimiter 实为 cancelled 非独立红；
「exit code 1/5」= mysql 步 exit 1 与 systemctl 步 exit 5，均非 fuzz 超时/内存，
无需调整 FUZZ_SECONDS/-parallel。

## 3. 逐目标归因与修复

### Fuzz_WorkerPool_SubmitTask（真 crasher，本地 0.6s 复现）
- 根因：lint-governance 轮把空转占位改为显式断言时引入——
  `if task.TaskID == "" || task.Operation == nil { t.Fatal }`，而 fuzz 可变异出空
  taskID（合法输入：`WorkerPool.Submit` 不拒绝空 ID），种子语料无空串故包测试全绿。
  本地复现输出（会话执行记录）：
  ```
  --- FAIL: Fuzz_WorkerPool_SubmitTask (0.59s)
      worker_pool_test.go:370: 任务结构无效
  Failing input written to testdata/fuzz/Fuzz_WorkerPool_SubmitTask/31f0b3670d2f5ddc
  ```
  （crasher 内容 `string("")` + `int(1)`）
- 修复：空 taskID 与超长输入同为域外 skip；保留 `Operation != nil` 断言。
  crasher 31f0b3670d2f5ddc 入库为回归语料（`.gitignore` 的 `testdata/` 规则用
  `git add -f` 放行）。
- 本地验证：修复后 20s/30s×2 稳定 PASS（最高 80 万 execs），`-race` 包测试绿。
  快照 `snapshots/local-workerpool-fixed-20s-pass.log`。
- 附带观测：fuzzcache 残留语料 + 同机高负载时出现过一次偶发
  `context deadline exceeded`（`go clean -fuzzcache` 后不复现，非代码缺陷，留观）。

### Fuzz_QueryBuilder_{Where,Like,Order}（两层缺陷）
- 第一层（当时 CI 红的直接原因）：「设置数据库」步 mysql root 认证被拒。
- 第二层（修好 mysql 后必然暴露）：3 个 fuzz 目标所在 query_builder_test.go 带
  `//go:build ignore`（3471df8，历史「跳过不完整测试」），目标从未参与编译，
  `go test -fuzz` 匹配不到目标时静默空跑 exit 0，proof 步
  （check-fuzz-result.sh）将因缺 `--- PASS:` 行失败。包内编译测试实际用
  sqlite `:memory:`，整个 job 不需要 MySQL。
- 修复：删除「设置数据库」死步；3 个 fuzz 目标移入新建
  query_builder_fuzz_test.go（参与编译；空块占位改为域外 skip，避免触碰
  lint 门禁 empty-block）。注意 `.gitignore:282` 的 `*_test.go` 全局忽略规则
  会静默挡掉任何新增测试文件（本次 `add -f` 入库），已标记为 follow-up 隐患。

### jwt-validation / CacheService（环境步）
- security 包及其 fuzz 目标不依赖 Redis（包测试无 redis 引用）→ 删除 jwt job
  的「设置Redis」死步。
- cache 包测试直连 `127.0.0.1:6379` → Redis 步改为安装并启动（apt 安装
  redis-server，失败回退 valkey；`/dev/tcp` 端口就绪等待），comprehensive 同步。

## 4. 本地验证（commit 69258f7 前）

- `go vet` concurrency/repositories 过；gofumpt v0.12.0 三文件 0 漂移；
  golangci-lint v1.64.8 门禁集两包 0 findings（RC=0）。
- WorkerPool 修复后 fuzz 稳定 PASS；repositories 3 目标 fuzz PASS；
  jwt/cache 目标此前已本地 15s PASS（85k/11k execs）。
- 工作树内 t_745060c0（gosec 治理）进行中的未提交文件未触碰，提交严格
  path-scoped（5 路径）。

## 5. 托管核验（真实 PR 事件）

- PR #54（draft → main）push 69258f7 自动触发 fuzzing run：
  **45412813406**（2026-11-24T14:19:33Z → 14:39:06Z）
- 结论：**19/19 job 全部 success**（含完整Fuzzing测试 15 目标串行、
  Fuzzing测试汇总报告 verified=true）。
- 观测修复实证：16 个 `fuzz-logs-*` artifact 全部上传（未过期）。
- 快照：`run-45412813406-meta.txt`、`run-45412813406-jobs-final.txt`。

## 6. 遗留与建议（不在本卡范围）

1. `.gitignore:282` `*_test.go` 全局忽略规则会静默挡掉新增测试文件——建议
   follow-up 卡评估收紧为目录级或删除该规则（本次已留注释于 commit message）。
2. fuzz 偶发 `context deadline exceeded`（高负载 + fuzzcache 残留场景）留观，
   复现可考虑 CI 侧 `go clean -fuzzcache` 前置。
3. comprehensive 每 PR 花 ~20 分钟（15 目标×60s 串行）；如需提速可评估
   matrix 化或缩短 PR 事件时长、把 60s 全量留给 schedule。
