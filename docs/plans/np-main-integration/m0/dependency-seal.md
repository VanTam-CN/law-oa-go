# M0 依赖封板（dependency-seal）

- 任务 t_7f6ad58e。主工作树 output/np-main-m0 @ 213da4c7c6710c99d660c33a5701a230c674ede9；应用 main d9c92fb；源 np-v04-first-slice @ 2d659d9（对象库分离，git show 须 -C 源仓）。
- 本文件为 M0 设计封板，非运行验收；M1 实现不在本轮授权。
- 33 条逐项展开以 [dependency-seal.json](../../../np-evidence/m0/dependency-seal.json) 为准，本文件只做摘要。

## 一、M 包分工摘要（33 条明细见 JSON）

| M包 | 覆盖条目（JSON n） | 内容 |
|---|---|---|
| M1 | 6、17 | expected_revision 合同净增量 + 000086 冻结迁移（两表，不整组带入82–85）；最小范围，不加完整平台 |
| M2 | 1、27、29–32 | CI与环境脚本；初始化改为main v14+86，保持隔离校验，不执行82–85 |
| M3 | 2–5、7–16、20–26、28、33 | NP挂接、B夹具/脚本适配、SQL/审计日志保护与同阶段测试；不引入材料文件handler |
| M4 | 复用M3表单/Hook；额外main/cache入口 | 仅内存、恢复键与编辑保护、三类响应缓存/SW边界；不是连接池日志功能 |
| M5 | 18–19 | live 端到端 |

## 二、修正要点（相对初稿）

- **audit/common（13/14）**：source-baseline.json 中 slice_equals_main=false，即使 frozen==main，slice 已改，必须按 slice_blob 移植日志保护，不得写"无需改"。
- **"拆fixture后置"区分**：夹具延后到对应里程碑（M3/M4/M5）但在该里程碑**必跑**；日志安全测试（5/15）不得真延期。I09-FILES材料文件流明确DEFERRED/HOLD，完整continuation正向提交判断也不纳入本首切片；既有保护语义不后置，夹具不引入material_files/case_tasks。
- **continuation 排除**：源 continuation_handler_test 的 can_submit/can_reuse_check+facts_version 完整复核不属于首切片；不引入 IntakeContinuationService 或新 continuation 端点。主干原生详情/恢复的角色/版本/缺对象/脱敏场景仍必跑，由主干适用入口适配。

## 三、原生额外改动（必要符号/调用点，详见 JSON 与 backend-closure）

- 冲突链路：ConflictCheckRequest（internal/models/conflict.go:12）无 revision 载体 → M1 增服务端取数 revision，落入既有 search_parameters 审计快照；异步沿用 conflict_check_records（async_conflict_check_service.go:17-22），不新增任务表；Link 与 intake 二次更新（demo_aggregate_handler.go:2378-2387）合并单事务。
- 接案合同：主干 UpdateCaseIntake 无 expected_revision/行锁（demo_aggregate_handler.go:1892-2053）；源合同（源 demo_aggregate_handler.go:2295-2460，必填:2367、行锁:2437-2446、来源键:2390-2404）。
- production schema：case_materials单表无版本列，删除重建在demo_aggregate_handler.go:2017-2035；沿用v14历史bootstrap，86显式迁移并增加两表/索引就绪检查，不重写旧bootstrap；路线B不新增版本栈。
- session/router：复用 AuthTokenSession+TokenSessionStore+000078，不移植源会话代码；NP 路由挂接复用既有中间件。

## 四、迁移 86 冻结与兼容/回滚

- 主干迁移止于 000079+001，000086 编号未占（实测 0 个文件）。
- 源 000086 在全新主干 v14 合成库独立执行成功（docs/np-evidence/m0/np86-dependency-probe.json，np86-verify exit 0：两表、两唯一索引、接案外键均在）；未执行 82–85。不依赖、不整组带入。
- 兼容：无历史旧对象允许原生合同；保护历史/未知不得降级；助手关闭不解除服务端保护。
- 回滚：只至保留 revision/来源保护的候选；持久墓碑不删除。单实例与旧客户端控制是演练门禁要求，尚未验证。

## 五、验证状态

本封板对应的未来集成验证均NOT_RUN；M0原main构建/回归证据不失效，也不自动继承至新集成候选。V-INTAKE原生部分与V-CONFLICT（I14/I15）必须在M1通过后再到M3；NP API/严格写/日志/组件在M3验证，真实main入口在M5。000086独立结构探针PASS仅限结构，正式目标迁移/升级不算通过。M0设计封板不关闭运行待验项。

## 六、边界

只写本文件与 dependency-seal.json；不改看板、不提交推送、不改队友文件。复核与推送由主代理执行。

## 七、主代理封板裁决

- JSON另列8个原生扩展目标的main blob、符号、必要增量与验收；2个计划基础文件是未来M1待实现，不是当前已存在代码。接案revision helper只移植最小符号，不借此加入完整续办/提交判断。
- 新集成候选必须先显式应用86，并在启动时检查两表/两唯一索引，助手关闭也不解除。旧的、确定从未启用NP的未升级环境可继续其旧原生版本，不标作NP候选通过；不增加永久生命周期平台。已升级库缺表/查询错误失败关闭，不能回退旧全量写。
- 无来源历史的对象可用原生全量入口，但仍须满足M1权限/expected_revision/幂等；不承诺外部旧客户端透明兼容。升级/回滚、单实例和客户端切换的实际演练仍是I12/试用前门禁，不是M0纸面通过。
- 33条身份已由主代理逐项从正确Git对象库复核，见[blob复核](../../../np-evidence/m0/seal-blob-verification.json)。本清单封板针对固定SHA；基线变化或编号86被新提交占用时重核，不静默换号。
