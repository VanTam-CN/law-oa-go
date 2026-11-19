# PR #47 独立审查处置 R1

2026-11-19，Asia/Shanghai。审查输入：b3809f604dfa2b0b43e8108bc5644d0ba6ed8c29；冻结父 bcd1b8992b3f3450d5ab406afd74336cc282eece。保持原 Draft PR，不合并、不部署。

## 逐项裁决

| 审查项 | 裁决与处理 | 对应验证 |
|---|---|---|
| P1-1 主程序数据库日志 | 成立。main 实际走 InitOptimizedComponents → NewOptimizedDatabase；只改 InitWithConfig 不足。实际工厂补参数化，敏感会话仍禁 SQL | 工厂真实日志金丝雀 first red → PASS；独立 -race 复跑 PASS |
| P1-2 4xx 清恢复键 | 成立。408/429/499、未知 4xx 不证明原请求未提交。采用更窄的 409 清键白名单：INTAKE_DRAFT_REVISION_CONFLICT、SOURCE_SELECTION_STALE、SAVE_REQUEST_ABANDONED。401/403/404 和 IDEMPOTENCY_PAYLOAD_MISMATCH 同样不能证明先前请求未提交，继续保留 | 主代理实现并补 22 个 hook 用例：同键同体重试、改文拒绝新键、重挂载先回查、显式放弃和终态清理 |
| P1-3a 旧原生伪造来源路径 | 成立。创建过滤该客户端路径；更新的普通合并和检测失效重建两条分支均忽略客户端同名字段，不删除合法服务端已有字段 | 新 PG 测试覆盖创建、draft、lawyer_facts_confirmed、conflict_ready |
| P1-3b SQL NULL metadata | 条件成立，不能称当前冻结结构已触发：productionCaseIntake 的 metadata 为 NOT NULL，默认 {}。兼容防御对手工/来源两条 jsonb_set 使用 COALESCE | 在专用合成库临时放宽 NOT NULL 后实际保存，检查版本/来源入库及旧版本拒绝，清理恢复非空约束 |
| P2-1 组件晚到响应 | 当前唯一挂载点有 actor+对象+revision key，尚无现用入口跨草稿污染的复现证据；新增内部单调 epoch，切对象/版本及新请求使旧响应失效。每个 await、catch、finally 均校验，切换时清旧来源和 busy；不以同步包装代替异步隔离 | GLM 提供组件/夹具初稿，主代理收口；4 个 unkeyed 用例覆盖成功/拒绝的旧 GET、迟到生成、切走再返回。旧组件四项失败，修后全部通过 |
| P2-2 非 UUID ID | 成立，新增守卫查 PG 前校验并返回 400 INVALID_REQUEST | PG 分别走旧全量和严格局部请求 |
| P2-3 原生浏览器草稿缓存 | 接受边界澄清，不删除原生恢复数据。对话、结构化源包、候选及授权不落浏览器存储；已应用摘要仍会被原生草稿缓存，Stub 又复制原句，因此其中可能包含原始录入文字 | 文档校正；共享终端/真实资料试用须决定该缓存的启停与清理，不以本轮放行 |
| P2-4 执行证据缺参数 | 成立。保留旧 stdout 的局限；本轮逐命令新增 argv、来源 SHA、开始/结束时间、退出码及日志散列；竞态测试显式记录 -race，另安排未参与修改的执行者 | 后端、前端、类型和浏览器独立元信息已落盘，均退出 0 |

## 证据解释

- review-r1/pg-first-red 是首探针，包含测试自身误判（创建成功返回 201，以及初版 SQL 占位表达）；不能把这些当成产品缺陷复现。修正夹具后在旧业务代码快照另行复现，证据为 pg-reproduction-old-code，不抹掉首失败。
- 旧代码复现目录由 git archive b3809f6 backend 生成，只额外叠加新回归测试和夹具 POST 创建路由。其 metadata NULL 测试是模拟遗留结构；旧表结构自身有非空约束。
- NULL 测试的 ALTER 只发生在 np_first_slice_dev 合成库，测试 cleanup 恢复 NOT NULL；未操作试用数据库或执行真实环境迁移。
- 原生全量路径的未知 metadata 整包兼容并未升级为完整增量协议；本轮只保护服务端来源保留路径。严格摘要保存继续使用 PG 路径更新保护大整数与高精度数。
- 原外部审核者确实复跑了前端三套七用例和类型检查；其 Go/PG/浏览器未执行，不能标作独立 PASS。此次另行执行与作者复测分层列出。

## 本轮结果

应用候选：`4c184d5d46f1e101d3ff0ced04f62ebd0b16d6f3`。三个增量提交：

- `58da474bf917c60ff3107ce4aee56eb41c7441da`：后端来源保留路径、NULL 防御、非法 ID、主连接池参数化及对应回归。
- `62ea83a1be0e02a3e693389c13efdaccd499abdc`：不明保存错误保留原键和冻结体，22 个 hook 用例。
- `4c184d5d46f1e101d3ff0ced04f62ebd0b16d6f3`：组件跨作用域晚到响应隔离，4 个组件复用用例。

最终证据提交只补文档/证据、测试末尾空行和“新草稿可继续生成”按钮断言；不改应用行为。新增按钮断言由主代理另做定向复测，不冒充此前固定提交的独立执行。后端独立执行固定 58da474，其 backend 树与 4c184d5 一致，树为 `b810d1390ff72c997b5e163d7e4e81328cc5ff02`。

| 层次 | 检查 | 结果/证据 |
|---|---|---|
| 作者/主代理 | 全 Go 构建、前端类型及构建 | PASS，go-build-fixed、typecheck-integration、frontend-build 的 JSON 含 argv/显式退出码；保留原有大包警告 |
| 作者/主代理 | 定向 backend -race | PASS，backend-race.json；实际 PG 十项顶层测试，含本轮新增三项；另含原生七项、上下文三项、日志工厂和审计 |
| 作者/主代理 | 前端集成 | PASS，五套 33 用例；client-integration.json |
| GLM Maxwell 独立执行 | 固定干净 58da474 的 Go -race/PG | PASS，independent-backend.json/log，27 RUN 含五个子用例，22 个顶层测试；不是另一台机器 |
| GLM Maxwell 独立执行 | 固定 4c184d5 的前端/浏览器 | PASS，五套 33 用例、类型检查、四条真实浏览器；各 JSON 完整命令和退出码 0，工作树源码 clean |
| GLM Maxwell 独立取证 | 同轮浏览器 PG before/after | PASS，来源 15→18、回执 18→22，保护字段/子记录/两版材料/D2 散列不变，最后手工保存解绑；历史证据原字节保留 |

主代理复审包括实际入口、三个允许清键返回码与事务语义、两条 metadata 重建分支、两种 NULL 写入、异步成功/失败/切走再返回，以及本轮证据层次。前端和 PG 行为不能用静态源码替代；原工程报告的旧 PASS 也不覆盖本轮新候选。

GLM Hooke 另对未参与实现的 hook 白名单与后端抛错点作只读复核，限定范围无新问题，未重跑测试；不将其连接池自查当作独立后端验收。Maxwell 未参与本轮实现，在两个固定干净工作树上实际执行并复核结果。22 个 hook 用例的准确分解是 15 个不明错误参数组合、3 个确定性 409、重挂载回查/放弃 1 个、成功回执 2 个、带来源的最小缓存 1 个；独立回报中的“14+8”是计数表述误差，实际 stdout 为 22 PASS。

独立浏览器的实际 argv 为 `npx playwright test e2e/np-first-slice.live.spec.ts`，使用默认 `playwright.config.ts`，没有传 `--config`，不补造参数。默认配置允许复用前端，但本轮日志确实新启 WebServer，BABEL 编译路径指向独立工作树，代理环境固定 18789；本轮没有复用旧前端。以后按环境文档的 `--config playwright.np.config.ts` 标准命令复现，以配置层强制禁止复用；本次不声称已经使用该配置。

本轮三个 P1 在首切片适用边界内已修复、独立复测通过，复审未发现新的未关闭 P0/P1。**首切片工程验收 PASS；扩大试用 HOLD；合并部署 HOLD。** 这不是完整窄版验收或 GitHub 正式批准。

## 门禁与剩余边界

扩大试用、合并部署继续 HOLD；PR 保持 Draft，CI 为 NOT_RUN。以下未因修复自动通过：

- 真实 main.go + Redis 的完整联调；本轮工厂行为回归不能代替该联调。
- 完整非摘要字段/主体/材料和外部旧调用迁移；已有来源历史的草稿仍拒绝旧全量写入。
- 升级回滚全过程单实例演练、真实模型、跨角色来源权限、真实律师和真实资料试用。
- 与 main 的目录映射和独立集成 PR；不得将冻结比较 PR 直接合入 main 或替换试用制品。
- 真实资料/共享终端试用前，明确原生摘要浏览器缓存的保留、禁用或清理策略；本轮保留已有恢复数据，未改浏览器凭据与缓存策略。

全窄版估算仍为 50–80 人日、含缓冲 60–96 人日；本次是首切片审查修复，没有扩大产品能力或承诺完整门禁已通过。

## 收尾与 Git

最终复审核对 `apiRequest` 实际透传 HTTP status 和 `body.error.code`，白名单不依赖错误消息解析。只对本轮差异及相关调用链复核，没有声称完成全库安全审计。

本机专用 18789/3003 已关闭，np-first-slice 的 PG/Redis 容器已停止，合成卷保留；NULL 夹具恢复 NOT NULL 已只读确认。冻结 worktree 仍 clean，原脏工作树未纳入。新增证据目录公开上传检查零命中；整条 Git 历史仍有原冻结的六个已识别测试/存储键命中，未改写为零命中。Git 保持原提交历史，推送到原开发分支和 Draft PR #47；最终确切远端 SHA 以 PR head 读回为准。

原始 stdout 的 CR/结尾空行按原字节保留，未为消除 whitespace 提示改写证据；代码和说明文档的差异检查通过，日志散列由 evidence-index.json 绑定。
