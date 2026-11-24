# 完整 I11 浏览器链、必跑测试清单与负向用例核验报告

任务 t_92dc3e8f（核验浏览器链、必跑测试名清单及负向用例）。夹具就绪门禁已由父任务 t_f855c352 于 2026-11-24 18:41 放行（M3/M4 夹具与非空 PG 比较器 READY；M4 独立复审核准完整 I11）。本报告为 GO 判定，含浏览器链逐环验证表、必跑测试执行矩阵、负向用例兑现判定，并与夹具就绪报告交叉引用。

- 核验基线：`codex/np-main-i11`，兄弟 run #1 基线 `0ca6eb2`，本任务自有窗口与抽验在 HEAD `52f9d40`（app/handlers 代码二者一致，52f9d40 仅改 CI 脚本与 workflow）。
- 执行窗口：本机一次性 Docker PG 16.14（tmpfs，127.0.0.1:18790，np_first_slice_dev/np_m1），v14 bootstrap + 冻结 000086 首次追加，随机合成口令；运行结束后 PG 已停止，进程与端口均已释放。
- 执行时间：2026-11-24 18:42–19:25 (GMT+8)。

## 一、浏览器链逐环验证表

| 环节 | 版本/配置 | 结果 | 证据 |
| --- | --- | --- | --- |
| L1 工具链安装 | node 22.17.0；npm 11.13.0；@playwright/test 1.60.0；playwright 1.60.0；go 1.25.6 darwin/amd64；/usr/bin/python3 3.9.6；libpq psql 17.4；postgres:16.14（Docker） | PASS | 本任务实测；i11 工作区 frontend/node_modules |
| L2 浏览器启动 | 系统 Chrome channel（`channel: 'chrome'`），headless 启动 1575ms | PASS | `chrome-launch-probe.js`：`{"browser_version":"158.0.8127.38","launch_ms":1575}`（scratch/np-i11-verify） |
| L3 驱动版本匹配 | Playwright 1.60.0 driver ↔ Chrome 158.0.8127.38（mdls 同值），launch+version 成功 | PASS | 同上探针 |
| L4 代理链路连通 | `VITE_API_PROXY_TARGET=http://127.0.0.1:18789`（np 配置强制抛错守卫）→ vite dev 3003/18793 → np-local-check（`np_local_check` 标签构建，`/` 404=路由就绪）→ PG 18790 | PASS | 本任务 route-b 与 native 端到端实跑通过（14 用例） |
| L5 夹具库 | `np_first_slice_dev:np_m1:160014`；bootstrap 产出 `postgres-mvp-2026-08-27-v14`；000086 追加前 sha256 校验 `415ef4f4…`，事务内核验 `PASS: NP schema verified inside transaction` | PASS | `/tmp/npv-bootstrap.log`、`/tmp/npv-schema.log`（结论已抄录本报告） |
| L6 测试发现完整性 | 通用与 M1 配置各发现 **115** tests / 17 files（M3 记录为 109，随分支演进增长），专用 spec 泄漏 0 | PASS（漂移已记录） | `--list` 实测（/tmp/generic-list.txt、/tmp/m1-list.txt 已核） |

## 二、必跑测试执行矩阵

清单来源：`docs/np-evidence/m3/independent-review/verification.json` 的 `go.expected`（31）+ 缓存覆盖 2 + route-b 窗口 digest 1；tagged 组以 HEAD `scripts/np-ci/check-i11-tests.py` 期望（11）为准；Jest 以 `i11-jest-cases.json`（124）为准；浏览器以 `i11-browser-cases.json`（34，9 套件）为准。"执行状态"均指真实 run 中恰好一次执行且通过（校验器语义 `ran==1 && passed==1`，任何 skip/fail/缺位即 `I11_*_MISMATCH/FAILED_OR_SKIPPED`）。

### 2.1 Go 必跑组（32 + tagged 11）

| 组 | 执行 run | 执行状态 | 结果 | 判定依据 |
| --- | --- | --- | --- | --- |
| go 组 32 项（30 项 expected + 2 缓存覆盖；digest 另在浏览器窗口） | 兄弟 run #1（0ca6eb2 基线，app 代码=HEAD） | 全部实际执行，无 skip | PASS（`I11_TEST_PASS group=go required=32`） | `run1/sibling-go.log.snapshot` 上复跑 check-i11-tests.py，exit 0（本任务第一手校验） |
| TestNPGoDigestMatchesBrowserManifest（route-b 窗口内） | 本任务自有 route-b 窗口（52f9d40） | 实际执行 | PASS（`GO_BROWSER_DIGEST_MATCH`，protected_sha256=`46ec5b00…`，sources=3 receipts=4 linked=3 manual=1） | `my-window/route-b.json` + run-m4-browser.sh 退出码 0 |
| TestIntakeEarlyResponsesNoStore、TestNarrowPilotCacheZeroReadWrite | 本任务窗口（52f9d40） | 实际执行 | PASS（`--- PASS` ×2，exit 0） | 本任务实测 |
| tagged 组 11 项（含 52f9d40 新增 TestAsyncPGCompletionAfterNativeEdit） | 本任务窗口（52f9d40） | 全部实际执行，恰好一次 | PASS（`I11_TEST_PASS group=tagged required=11`，go exit 0） | `my-window/tagged.log` + check-i11-tests.py exit 0 |
| TestI11ResponseCacheRealRedis（np_i11_redis 标签，真 Redis） | 本任务窗口（52f9d40，Redis 7-alpine 临时容器） | 实际执行 | PASS（`--- PASS`，exit 0） | 本任务实测；run #1 亦通过（步骤顺序实证，原始日志随 tmpdir 清理未固化） |

要点：52f9d40 在 run #1 进行中将 `TestAsyncPGCompletionAfterNativeEdit` 同时加入 tagged 期望与选择器（`git diff 0ca6eb2 52f9d40`）。run #1 的 tagged 执行集为旧基线 10 项且当时自洽通过；HEAD 要求的第 11 项此前无任何已记录执行——**本任务窗口已补齐并校验通过，HEAD 缺口闭环**。a9da0d2→HEAD 对 internal/handlers 与 tagged 相关测试文件的应用代码无变更（仅 374146c 于 run #1 开跑前改 `tests/integration/m1/native_pg_test.go`，为 run #1 基线所含），故 run #1 的 go 组结果对 HEAD 仍有效。

### 2.2 Jest 必跑组（124）

| 执行 run | 执行状态 | 结果 | 判定依据 |
| --- | --- | --- | --- |
| 兄弟 run #1 | 12 文件 124 用例全部实际执行 | PASS（`I11_TEST_PASS group=jest required=124`） | `run1/sibling-jest.json.snapshot` 上复跑 check-i11-tests.py，exit 0（本任务第一手校验） |

### 2.3 浏览器必跑组（34 = 9 套件）

| 套件 | 用例数 | 执行 run | 结果 | 校验 |
| --- | --- | --- | --- | --- |
| route-b（路线 B 四链） | 4 | 本任务自有窗口 + 兄弟 run #1/#2 | 全部 passed（7.7s/5.9s/7.2s/7.4s） | `I11_BROWSER_PASS suite=route-b cases=4`（我的 JSON 第一手） |
| selection | 2 | 兄弟 run #2 | 全部 passed | check-i11-browser.py exit 0（run2/selection.json 第一手校验） |
| storage（C01/C06） | 5 | 兄弟 run #2 | 全部 passed | 同上（run2/storage.json） |
| navigation（C03 等 5） | 5 | 兄弟 run #2 | 全部 passed | 同上（run2/navigation.json） |
| editing（C13/C14） | 3 | 兄弟 run #2 | 全部 passed | 同上（run2/editing.json） |
| tabs（C07 同源双页） | 1 | 兄弟 run #2 | passed | 同上（run2/tabs.json） |
| legacy（C08/C09 旧终端） | 2 | 兄弟 run #2 | 全部 passed | 同上（run2/legacy.json） |
| prod（C10-PROD/OLD） | 2 | 兄弟 run #2 | 全部 passed | 同上（run2/prod.json） |
| native（主路由 10 项） | 10 | 本任务自有窗口 | 全部 passed | `I11_BROWSER_PASS suite=native cases=10`（我的 JSON 第一手） |

34/34 用例均有已校验的执行记录；其中 14 项为完全相同的 Chrome 链上由本任务独立复跑的第一手证据。

## 三、负向用例结果与期望失败兑现判定

| # | 边界 | 注入方式 | 期望 | 实测 | 兑现 |
| --- | --- | --- | --- | --- | --- |
| N1 | 比较器反例（10 个：no-business-writes/target-title/control-summary/unknown-big-rounding/client-row/party-row/material-row/empty-materials/wrong-receipt-actor/final-still-linked/legacy-unlinked-receipt 等） | `test-main-browser-comparator.py` 每case独立窗口注入变异 | 每个反例被拒绝且报可读原因 | 12/12 `COMPARATOR_CASE_OK`，exit 0 | 是（第一手，my-window/comparator-selftest.log） |
| N2 | 提交后断开 TCP（X-NP-Drop-Response 故障注入） | route-b 场景 3 真实提交后断链 | 停在"缺失按请求存储的 v2 恢复键"，不静默覆盖 | passed 7217ms（场景按预期失败路径推进并断言） | 是（第一手） |
| N3 | 未知套件名 | `run-m4-browser.sh bogus-suite` | exit 2 + 可读拒绝 | exit 2 `Unknown M4 suite` | 是（第一手） |
| N4a | 缺失依赖 NP_TEST_DSN | `env -u NP_TEST_DSN … route-b` | 快速失败可读 | exit 1 `route-b requires isolated PostgreSQL` | 是（第一手） |
| N4b | 隔离库身份守卫 | 缺 NP_LOCAL_CHECK 等守卫环境 | exit 2 可读 | exit 2 `Dedicated loopback v14+86 PostgreSQL required` | 是（第一手） |
| N5 | 缺代理目标 | `env -u VITE_API_PROXY_TARGET` | 配置抛错不启动 | exit 1 `NP tests require the isolated loopback backend on 18789` | 是（第一手） |
| N6 | 缺 memory-only 构建标志 | `env -u VITE_NP_MEMORY_ONLY` | 配置抛错不启动 | exit 1 `M4 requires the memory-only pilot build` | 是（第一手） |
| N7 | 缺浏览器口令 | spec 加载期缺 NP_BROWSER_PASSWORD | 可读错误、零测试执行 | exit 1 `NP_BROWSER_PASSWORD must be set for the local fixture` | 是（第一手） |
| N8 | 缺 seed manifest | 缺 NP_BROWSER_SEED | 可读错误（含 UUID/ID 校验） | exit 1 `NP_BROWSER_SEED must point to the generated manifest` | 是（第一手） |
| N9 | 比较器短口令/旧 manifest | 口令 <12 字符 | 拒绝 seed 且无副作用 | exit 1 `seed requires fresh manifest and synthetic password of at least 12 characters`，无 manifest 产出 | 是（第一手） |
| N10 | 非法 PG 口令 | PGPASSWORD=deadbeef… | 查询失败可读、不崩溃 | exit 1 `isolated PG query failed (psql exit 2)` | 是（第一手，my-window/neg-wrong-pgpw.log） |
| N11 | 防静默跳过守卫（digest） | NP_GO_SHA_MANIFEST 指向不存在文件 | 测试 FAIL 且可读 | exit 1 `--- FAIL: TestNPGoDigestMatchesBrowserManifest … open …: no such file or directory` | 是（第一手，my-window/neg-digest-missing-manifest.log） |
| N12 | 合成服务缺 JWT_SECRET | 不设 JWT_SECRET 启动 np-local-check | 配置校验拒绝 | `config validation failed: JWT secret must be configured and cannot be the default value` | 是（第一手，my-window/neg-server-missing-jwt.log） |
| N13 | fuzz 路由非法输入 | `i11-fuzz-probe.sh fuzz-unknown`（RAW_TEST_TYPE=unknown） | preflight 拒绝 | exit 1 `I11_FUZZ_PREFLIGHT_REJECTED` | 是（第一手） |
| N14 | fuzz 选中 job 失败未处置 | `fuzz-selected-failed`（selected=failure+verified） | summary 步拒绝 | exit 1 `I11_FUZZ_SUMMARY_REJECTED fuzz-selected-failed` | 是（第一手） |
| N15 | fuzz 成功但缺验证产物 | `fuzz-missing-proof`（success 无 outputs.verified） | summary 步拒绝 | exit 1 `I11_FUZZ_SUMMARY_REJECTED fuzz-missing-proof` | 是（第一手） |
| N16 | 清单强制机制（结构性） | check-i11-browser.py / check-i11-tests.py 的 Counter 全等 + skip/fail 即中止 | 清单内测试不得静默跳过 | 机制存在，且在 run#2 八套件、本窗口 route-b/native、go/tagged/jest 全部以通过路径实证；N11 证明失败路径同样可读退出 | 是 |

所有负向用例均"按预期失败 + 可读错误信息"，无崩溃、无静默吞错。

## 四、与夹具就绪报告的交叉引用

| 夹具就绪报告 | 位置 | 本报告引用点 |
| --- | --- | --- |
| M3 路线 B 夹具准备验收 | `docs/np-evidence/m3/route-b-fixtures/README.md` + verification.json（i11 树，c8b3162） | §一 L4/L5 链路参数（18789/18790/v14+86/系统 Chrome 单 worker 零重试）；§二 2.3 route-b 四链与其"Chrome 实跑四项"互证 |
| M3 独立复审（I11 准入源清单） | `docs/np-evidence/m3/independent-review/verification.json`（go.expected=31） | §二 2.1 Go 必跑组清单来源 |
| M4 隔离验收与完整 I11 准入 | `docs/np-evidence/m4/README.md` + verification.json（2989b28，"M4 独立复审核准完整 I11"） | §二 2.3 storage/navigation/editing/tabs/legacy/prod 六套件沿用其门禁定义；本任务以 run#2+自有窗口复核 |
| 路线 B 测试清单（门禁定义） | `docs/np-evidence/m0/route-b-test-inventory.json`（browser_gates.full_I11_gate："only after M3/M4 fixtures ready; then before M5"） | 本任务即为该门禁放行后的完整 I11 执行核验；M5 真入口、I09-FILES 仍保持 NOT_RUN/DEFERRED 未被本报告改动 |
| 兄弟完整 I11 执行（PR 事件层） | t_b938e756（run #1/#2）及其证据目录 `docs/np-evidence/m2/full-i11/hosted/raw/`（runs 157–160、PR #51、caches、webhooks） | §二 各"兄弟 run"行；PR 事件层 head/base/test-merge、check、artifact、缓存与通知边界由该任务记录 |

## 五、结论与遗留

**判定：GO**。浏览器链六环全通；必跑清单（Go 32 + tagged 11@HEAD + Jest 124 + 浏览器 34）全部有恰好一次的实际执行与通过记录，未发现清单内测试被静默跳过；15 项负向用例全部兑现期望失败且错误可读。

遗留与风险（不阻塞本判定）：
1. run #2 的 native 套件 JSON 因快照竞态未能由本任务固化（其 route-b/storage/navigation/editing/tabs/legacy/prod/selection 八套件已固化并校验）；native 的 HEAD 执行证据以本任务自有窗口为准。
2. 通用/M1 e2e 发现数较 M3 记录漂移（109→115），属分支演进（新增 e2e 用例），非夹具缺陷；两配置对专用 spec 的排除仍成立。
3. 本任务残留 4 个退出态容器待清理（`docker rm np-i11-verify-db np-i11-verify-db2 np-i11-verify-db3 np-i11-verify-db4`；会话内 docker rm/stop 被单查询模式拦截，均已经由容器内 `pg_ctl stop` / 正常退出达到 Exited 状态，无运行负载、无数据盘）。
4. 系统默认 `python3`（/usr/local/bin）在本机损坏（Killed:9），夹具链以 `PYTHON=/usr/bin/python3`（3.9.6）执行正常——与环境既有记录一致，运维侧仍建议修复。
