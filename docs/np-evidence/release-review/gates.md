# NP M1–M6 发布评审证据台账

本页汇总 NP 主干集成 M1–M6 的验收归属、固定提交和证据层级，供发布评审复核。复核基线为 `250bccd`（分支 `codex/np-release-review`）；本页为该基线新增的证据台账，不包含代码或流程变更。逐项定义以[冻结验收矩阵](../../plans/np-main-integration/implementation-and-acceptance-plan.md)为准；本页只登记证据归属与边界，不新增验收结论。

当前应用候选为 `49e5c15`，远端 `main` 为 `d9c92fb`，从应用候选到复核基线新增 33 个提交、46 个路径；载体 PR [#52](https://github.com/VanTam-CN/law-oa-go/pull/52) 为 Open Draft、合并状态 UNSTABLE，检查汇总含 13 个 FAILURE（含与 NP 无关的历史红灯）。合并、部署与真实试用未取得放行证据，状态见文末清单。基线观测数据存于同目录 [refs.json](refs.json)。

## 证据层级口径

| 层级 | 含义 | 本轮示例 |
| --- | --- | --- |
| 真实托管 | GitHub Actions 真实 runner 上的运行，按固定 SHA 可回查 | 完整 I11 正常与负向运行；早期 CI 三次运行 |
| 独立复核 | 复核者在固定输入上独立执行并得出结论，非作者自证 | M3 Astra 只读复审；M6 Astra 独立重跑 |
| 隔离合成 | 本机回环端口、专用容器、合成数据的作者验证；浏览器为真实 Chrome，C12 故障为受控注入 | M1/M3/M5/M6 各阶段证据 |
| 继承 | 沿用前阶段已验收证据，不在本阶段重跑，边界随引用保留 | M0 基线、M4 夹具与生产资产、完整 I11 在 M5/M6 的引用 |

除外部 GitHub 运行外，引用证据均为脱敏汇总，不含正文、令牌、真实键或真实 ID。

## 里程碑总览

| 里程碑 | 内容与归属 | 固定提交 | 状态与层级 | 主要证据 |
| --- | --- | --- | --- | --- |
| M0 | 基线核验与审核修订；Codex 分组执行，GLM Flash A/B 辅助核验，主代理裁决 | 方案 `2481a47`；应用基线 `d9c92fb`；分支 `codex/np-main-m0` | 继承基线，不进入后续矩阵 | [M0 基线报告](../../plans/np-main-integration/m0/baseline-report.md)、[Go 基线](../m0/go-baseline.json)、[PG 结构基线](../m0/pg-schema-baseline.json)、[路线 B 清单](../m0/route-b-test-inventory.json) |
| M1 | 原生接案基础：作者隔离验证，GLM Flash 分担实现，主调度整合复核 | 基线 `292805c`（M0 v0.4）；Hermes `t_bee02828` | 隔离合成 + 作者复核；历次复审 r1–r10 | [M1 README](../m1/README.md)、[verification](../m1/verification.json)、[review-r1](../m1/review-r1/README.md)–[review-r10](../m1/review-r10/README.md) |
| M2 早期 CI | 开发分支 push/manual 门禁与 PR 副作用隔离 | 基线 `a0acf35`；工程 `9e94e0a`、`fa4a790`；PR 隔离工程 `b070323` | 真实托管 + 真实 PR 事件 | [早期 CI](../m2/early-ci/README.md)、[PR 隔离](../m2/pr-isolation/README.md) |
| M2 完整 I11 | 完整隔离验收与托管正负向；Hermes `t_7072cca9` | 工程候选 `52f9d40` | 真实托管 + 独立复审；M4 准入见外部证据分支 `2989b28` | [完整 I11](../m2/full-i11/README.md)、[正常运行 run 45379855010](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379855010)、[负向矩阵 run 45379862576](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379862576) |
| M3 | 独立复审、目标分支隔离验收、路线 B 夹具；Astra 复审，GLM Flash 参与移植 | 复审对象 `4e4fb84`；隔离验收 `c4b644c`；夹具基于 PR #48 `448047e`、代码 `c8b3162` | 只读独立复审 + 隔离合成 | [独立复审](../m3/independent-review/README.md)、[隔离验收](../m3/isolated-acceptance/README.md)、[夹具准备](../m3/route-b-fixtures/README.md)、[M2/M3 PR 隔离](../m3/pr-isolation/README.md) |
| M4 | 浏览器仅内存、同源双页、编辑保全、旧终端与缓存边界 | 应用 `a9da0d2`；Astra 准入记录 `2989b28` | 合成浏览器与 PG，独立复审后进入完整 I11 | [固定 M4 报告及结构证据](https://github.com/VanTam-CN/law-oa-go/tree/2989b285501f04fde37eb4826cb3f3a108be0915/docs/np-evidence/m4) |
| I11 gofumpt 清理 | i11/m3/m4 三线格式不动点清理；Codex 执行，CI 判定复跑 | 三线修复 `e2034e9`/`3cb6737`/`c4dbc3d`；Hermes `t_b453c53f` | 真实托管质量步转绿；伴随历史红灯未闭合 | [gofumpt 清理](../i11/gofumpt-cleanup/README.md)、[质量步 run 45395793290](https://github.com/VanTam-CN/law-oa-go/actions/runs/45395793290)、[np-i11 push run 45395793181](https://github.com/VanTam-CN/law-oa-go/actions/runs/45395793181) |
| M5 | 隔离作者集成，冻结矩阵逐项；GLM Flash 执行，Astra 对证据提交只读复审 | 应用候选 `49e5c15`；I11 归档基线 `89f2eb9`；证据 `753b5b0`；首轮复审 `60ae91b`；Hermes `t_5b8b79ec` | 隔离合成；Astra 只读复审通过，两缺口补跑关闭 | [M5 README](../m5/README.md)、[逐项矩阵](../m5/matrix.md)、[最终复审](../m5/final-review.json) |
| M6 | I12 合成切换演练：作者 glm-flash，Astra 独立重跑，协调者比对导出 | 开发基线 `04bcce9`；作者窗口 `e26692c`；导出规则固化 `4b6fb43`；独立复核固定 detached `4b6fb43`；Hermes `t_1fb1dac9` | 隔离合成双跑通过；2026-11-25 独立裁决 I12 关闭 | [M6 README](../m6/README.md)、[作者运行](../m6/author.json)、[Astra 独立运行](../m6/independent-astra.json)、[检查器反例](../m6/switch-controls.json) |

## 路线 B 验收矩阵 I01–I16

逐项结果取自 M5 逐项矩阵与各阶段证据；「继承」指沿用前阶段证据并保留原边界。

| 项 | 结果与层级 | 证据 |
| --- | --- | --- |
| I01 | 继承 M0 主干基线（含 138 RUN/132 PASS/6 SKIP 边界）；M1 真实路由回归、M5 复用 | [M0 基线报告](../../plans/np-main-integration/m0/baseline-report.md)、[M1 README](../m1/README.md) |
| I02 | 干净检出普通与带标签构建、根二进制 VCS/SHA 校验、生产前端重建一致 | [环境与浏览器](../m5/environment-and-browser.json) |
| I03 | 真实登录、退出/撤销/过期/非在职拒绝、Redis 登记失效实测 | [会话边界](../m5/session-boundaries.json) |
| I04 | 本人/他人/缺失 200/403/404；密文与摘要哨兵不泄漏；客户撤权后各入口 403、PG 不变 | [会话边界](../m5/session-boundaries.json)、[缓存读路径](../m5/cache-paths.json) |
| I05 | 列表/详情/别名同 revision；双页并发 409 保留输入；原生与来源写入 PG 争用 | [同版本重开](../m5/reopened.json)、[路线 B](../m5/route-b.json)、[服务用例](../m5/pg-service-checks.json) |
| I06 | 来源/回执/指针同事务，大整数与未知字段保留，伪造指针具名回归拒绝 | [路线 B](../m5/route-b.json)、[组件结果](../m5/browser-controls.json) |
| I07 | 多轮整理、人工修改、到期与晚到回包不覆盖（真实页面 + 组件） | [路线 B](../m5/route-b.json)、[组件结果](../m5/browser-controls.json) |
| I08 | 转发前、锁等待、提交后三种真实断线时序，逐键核对无记录/放弃/已提交 | [断线窗口](../m5/disconnect.json) |
| I09-MAIN | 路线 B 来源 +3、回执 +4，两条材料与受保护行散列不变 | [路线 B](../m5/route-b.json) |
| I09-FILES | 未闭合，DEFERRED/HOLD，不以 Skip 或替代断言变绿 | 见文末未证明清单 |
| I10 | 真根服务 + 同一 PG/Redis：九条敏感路径禁缓存、ECHO 窗口零操作；撤权后重读 403 带 no-store | [逐路缓存](../m5/cache-paths.json)、[撤权重读](../m5/revoked-history.json) |
| I11 | 完整 I11 托管正负向通过（真实托管）；M5/M6 按原始运行 SHA 继承，无新托管 I11 | [完整 I11](../m2/full-i11/README.md)、[M5 README](../m5/README.md) |
| I12 | 合成升级/重启/回滚演练通过，作者与 Astra 独立双跑 | [M6 README](../m6/README.md)、[作者运行](../m6/author.json)、[独立运行](../m6/independent-astra.json) |
| I13 | 严格分流组件回归、混改标题浏览器用例、角色变更后真实上下文拒绝 | [会话边界](../m5/session-boundaries.json)、[组件与浏览器](../m5/browser-controls.json) |
| I14 | 未迁移旧库兼容对照（不报 NP PASS）；已迁移克隆：开关开 + 空白名单旧协议 409，来源表不可用均 503，恢复后整行与摘要不变；锁内竞争 PG 回归 | [历史故障](../m5/history-faults.json)、[未迁移对照](../m5/unmigrated.json)、[服务用例](../m5/pg-service-checks.json) |
| I15 | I15-NATIVE/I15-NP/M5 真链分层：迟到同步与异步均 409，匹配仅推进一次，PG 整行对照 | [来源冲突](../m5/source-conflict.json)、[原生迟到](../m5/conflict-native.json)、[正常完成](../m5/conflict-matching.json)、[整行复核](../m5/conflict-source.json) |
| I16 | 回查/放弃/提交按原对象原键判定；错身份、缺字段、非法版本逐名覆盖；同源双页真实浏览器 | [断线窗口](../m5/disconnect.json)、[组件与双页](../m5/browser-controls.json) |

## 缓存矩阵 C01–C14

| 项 | 结果与层级 | 证据 |
| --- | --- | --- |
| C01 | 两种登录持久 × 助手开关四种组合，存储与 PG 核对 | [环境与浏览器](../m5/environment-and-browser.json) |
| C02 | 新登录原菜单重开，版本与来源恢复，最终手工版本可重复读取 | [路线 B](../m5/route-b.json)、[同版本重开](../m5/reopened.json) |
| C03 | 后退取消、换草稿、步骤切换、令牌过期清理 | [导航与控制](../m5/browser-controls.json) |
| C04 | 转发前断线 GET unknown 不写；提交后新登录按原键回查 | [断线](../m5/disconnect.json)、[组件](../m5/browser-controls.json) |
| C05 | 状态码白名单、输入保留、未知状态不清键；PG 异参 409 | [组件](../m5/browser-controls.json)、[异参回放](../m5/disconnect.json) |
| C06 | 禁用存储阻止写请求，PG 整行与回执计数不变 | [存储窗口](../m5/environment-and-browser.json) |
| C07 | 同源双页交错保存，A 终态保留 B 未决标识 | [双页窗口](../m5/browser-controls.json) |
| C08 | 换律师登出登入，原律师重登逐笔回查 | [换账号窗口](../m5/browser-controls.json) |
| C09 | 旧标签内存内容先核对再定向清理；真实终端处置归 M6 边界 | [旧标签窗口](../m5/browser-controls.json) |
| C10 | 生产资产与受控旧 worker 分层：旧 worker 接管显示 HOLD、敏感请求为零；撤权前后 200/403/200 | [逐路缓存](../m5/cache-paths.json)、[撤权重读](../m5/revoked-history.json)、[生产资产](../m5/environment-and-browser.json) |
| C11 | 原键格式、逐请求去重、仅清自身终态；旧标签停写顺序 | [组件及旧标签](../m5/browser-controls.json) |
| C12 | 四场景后 PG 来源/回执/既有材料核对；助手关闭存储行为；生成 503 后人工编辑不落业务存储 | [生成失败](../m5/generation-failure.json)、[四场景](../m5/route-b.json)、[存储](../m5/environment-and-browser.json) |
| C13 | 保存 A 期间输入 B：浏览器保留 B，服务端重开只见 A | [编辑窗口](../m5/browser-controls.json) |
| C14 | 重载确认后输入 C，晚到响应不覆盖 | [编辑窗口](../m5/browser-controls.json) |

冻结用例名对照：[浏览器用例清单](../../../scripts/np-ci/i11-browser-cases.json)、[组件用例清单](../../../scripts/np-ci/i11-jest-cases.json)。

## I09 拆分

- **I09-MAIN（主干记录保全）**：路线 B 四场景干净窗口来源 0→3、回执 0→4，两条独立材料、客户/主体/对照整行散列前后一致（含大于 2^53 的未知整数）。M3 首验、M5 真入口复验；M6 用独立种子验证 I12 持久性。证据：[路线 B](../m5/route-b.json)、[M3 独立复审](../m3/independent-review/README.md)、[M6 作者运行](../m6/author.json)。
- **I09-FILES（两版文件保全）**：用户在 M0 出口选定后置。DEFERRED/HOLD，源用例清单保留，未以 Skip、假文件或删除断言宣称通过。未证明，见文末清单。

## I11 托管 SHA 边界

- 完整 I11 的权威托管结果绑定原始运行 SHA：工程候选 `52f9d40`，[正常运行 run 45379855010](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379855010) 与[负向矩阵 run 45379862576](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379862576)。M5 与 M6 不产生新的完整 I11 托管运行，按 [M5 README](../m5/README.md) 明示继续引用原始 SHA。
- 早期 CI 托管结果绑定开发分支：基线 `a0acf35`、工程 `9e94e0a`/`fa4a790`，runs [45160569000](https://github.com/VanTam-CN/law-oa-go/actions/runs/45160569000)、[45160694603](https://github.com/VanTam-CN/law-oa-go/actions/runs/45160694603)、[45160571800](https://github.com/VanTam-CN/law-oa-go/actions/runs/45160571800)，见[早期 CI](../m2/early-ci/README.md)。
- gofumpt 清理使 i11/m3/m4 三线 head 前进（`e2034e9`/`3cb6737`/`c4dbc3d`），其新增托管运行仅证明格式与 np-i11 门禁不受影响；完整 I11 权威结论仍以原始 SHA 为准，见[清理台账](../i11/gofumpt-cleanup/README.md)。
- M5/M6 的应用候选运行（浏览器、PG、切换演练）均为本机隔离合成执行，非托管 CI；托管验收只在上述 I11 专用运行内主张。

## 未证明与未闭合事项

| 事项 | 状态 | 依据 |
| --- | --- | --- |
| I09-FILES 两版材料保全 | DEFERRED/HOLD，未闭合 | [M0 基线报告](../../plans/np-main-integration/m0/baseline-report.md)、[M5 README](../m5/README.md) |
| 真实外部模型生成 | 未证明；C12 仅在浏览器边界注入合成 503，未验证真实模型链路 | [生成失败](../m5/generation-failure.json) |
| 合并、部署、真人试用 | 全部 HOLD，未取得生产或试用证据 | [冻结矩阵放行门禁](../../plans/np-main-integration/implementation-and-acceptance-plan.md)、[refs.json](refs.json) |
| 载体 PR #52 合并 | Open Draft，UNSTABLE；检查汇总 13 FAILURE，质量与测试等失败或取消的归属须对最终合入树逐条交代 | [refs.json](refs.json)、[gofumpt 清理](../i11/gofumpt-cleanup/README.md) |
| 原生新建接案同秒碰撞 500（`intake_code` 秒级，23505） | Hermes `t_a4646b57` 分开跟踪；修复 `a9a5116` 与回归 `a7ded85` 未进入候选 `49e5c15` | [M5 README](../m5/README.md)、[refs.json](refs.json) |
| CI 分流大小写变体 P2 | Hermes `t_fbcffe5f` 未闭合 | [M3 隔离验收](../m3/isolated-acceptance/README.md) |
| #52 普通质量与测试检查 | 格式步通过。静态分析、E2E 工具链和 Go 性能测试分别处置。后续同机 base/head 对照已将 TestPerformanceMonitoring 确认为候选新增可达回归；此前仅有两个较早 PR head 的对照。 | [冻结格式检查](../i11/gofumpt-cleanup/README.md) · [逐项处置与首项整改](pr52/README.md) |
| M5 独立复审层级 | Astra 对固定证据提交 `753b5b0` 只读复审，未独立重跑运行时；I12 在 M6 另建空库独立冷跑 | [M5 README](../m5/README.md)、[M6 README](../m6/README.md) |
| I12 回滚范围 | 仅覆盖 `89f2eb9` ↔ `49e5c15` 构建对；更早无保护版本只允许升级前 v14 阶段运行 | [M6 README](../m6/README.md) |
| M0 遗留边界 | 六项 Skip 与双包范围不继承为通过；生产 TLS/main+Redis 属 M5 实测，M0 基线不覆盖 | [M0 基线报告](../../plans/np-main-integration/m0/baseline-report.md) |
| M4 证据位置 | M4 独立准入在外部证据分支 `2989b28`，本地树仅继承其夹具与生产资产结论 | [完整 I11](../m2/full-i11/README.md) |

## 归属与边界声明

- M1–M6 的作者验证在隔离合成环境完成（回环端口、专用容器、合成数据），真实浏览器与真实 PG/Redis 参与运行；真实外部模型、真实外部用户、真实服务器未参与。
- Astra 的 M3 复审为只读复审，M5 复审为只读证据复审；M6 为固定提交上的独立重跑。作者通过不等于独立复核，独立复核不等于真实试用，合并/部署未授权。
- 本页仅为证据台账，登记各阶段已归档证据的归属与边界；复核结论以逐条点击的证据文件为准。
