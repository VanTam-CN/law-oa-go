# M0独立审核处置与文档复审（v0.4）

- 修订输入：`c8716b7a964e4035b938761b8d4201b28b501cba`，分支`codex/np-main-m0`。
- 应用主干：`d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7`；源首切片：`2d659d9dd750f49ddb1f517746e22851392231a6`，独立Git对象库。
- 独立审核：实际`gpt-6.2-astra`，代理`01a1f738-9fd8-7660-ae93-4f5da77925a6`；主调度逐项查证，采纳0P0/5P1/3P2。审核任务`t_e3878bd1`完成不等于M0准入。
- 当前授权：仅文档和结构证据修订、Hermes记录及推送原M0分支。没有M1、业务/迁移/测试脚本/workflow改动、运行回归、真实环境连接、PR创建、合并或部署。

## 1. 审核发现与最小处理

下表的源码行号绑定固定SHA，是静态证据；所有未来运行验证NOT_RUN。本轮M0-R编号不替换原准备轮MI-R编号。

| ID | 静态证据与问题 | 处理与验收归属 |
|---|---|---|
| M0-R01 / P1-1 | main router:378–384没有GET列表/详情，:399有工作台别名；源intake_draft_revision_test.go:109–112自行注册GET不能证明真实入口 | M1明确原生list/detail/alias与dashboard读取，不引入完整continuation。真实router本人200/越权403/业务缺失404，旧PUT409后GET最新；客户/对方全响应明文密文digest哨兵不泄漏。I04/I05，M5原菜单恢复 |
| M0-R02 / P1-2 | main demo_aggregate_handler:1996–2013先删除主体再generic prepare；源:2450–2481先准备新客户主档身份及旧对方身份后删除。源context:109依赖NewAuthorizationServiceWithDB，main授权仅四repo构造 | M1最小主体helper/事务顺序、客户切换及失败回滚；M3复用当前db/tx的现有repo构造或最小适配，客户撤权仍拒绝；不带入整套授权平台。I04/I06/I09-MAIN/I13 |
| M0-R03 / P1-3 | main检测入口:2188要求lawyer_facts_confirmed，源严格保存context:113要求draft；main通用async入口拒绝intake_id | I15-NATIVE先用合法原生事实修改使快照失效/回draft/revision推进，再恢复旧同步和服务级异步完成，intake+cases原子不变；公共入口拒绝不算晚到回写证据。I15-NP在M3追加来源联动，M5真入口，不阻塞M1首次执行 |
| M0-R04 / P1-4 | 源Hook测试:60 abandoned缺id、:69 committed/no_change缺request_id却期待清键，断言仍旧单槽 | 正常fixture回显捕获请求的request_id、id和合法revision；原缺字段改负向保键保输入。M3/M4按v2逐请求键断言并保留旧格式/双标签场景，I16/C04–C07；历史33只作来源数 |
| M0-R05 / P1-5 | 原依赖JSON n18/n19把spec/config放M5，而主计划要求M2完整I11四链必跑，形成前置循环 | M3准备B spec/config/seed/比较器，M4缓存fixture完成；M2先配置/独立负向，完整正常/负向待fixture完成再执行。M5只真root+PG/Redis/外TCP代理，不重复计实现 |
| M0-R06 / P2-6 | 源np-local-check:105–112只有目标/对照/客户/主体，seed-browser.py:28–43只一个材料两版，不满足B非空两材料及精度合同 | M3主干B seed至少2条独立材料、非空关联、精确ID和>2^53未知metadata；Go/Python直接PG规范文本散列。原四场景单独干净窗口来源+3/回执+4，新负向独立窗口；M2消费，不建材料版本平台 |
| M0-R07 / P2-7 | main cmd/migrate:32起production仅bootstrap；cmd/server:82也调用router.Init，只写root startup检查会遗漏装配入口 | M1受控86追加事务/SQL校验和/实际schema及已应用状态核验，不能盲重跑；共享router检查防备用入口旁路，唯一支持/发布目标为root。M6合成I12升级/重复状态/回滚演练，不执行down删除历史 |
| M0-R08 / P2-8 | 源Playwright config:17用npm run dev；main src/main.tsx:159只PROD注册SW | M4同候选PROD产物隔离host观测资产/注册/controller/CacheStorage，受控旧worker必须实际接管；缺资产如实记录，不凭404认为无旧worker。M5同产物真root链，旧终端不能证明处置则HOLD；不新建PWA |

## 2. 路线B与最小范围裁决

B仍是用户确认的范围，不新增material_files或两版文件平台，I09-FILES继续DEFERRED/HOLD。I09-MAIN必须非空，保护现有材料、客户、主体和对照草稿完整行；只排除目标接案description、updated_at、metadata.draft_revision、metadata.np_description_source_revision。不能把两版文件后置变成其他保护场景一并后置。

26个Go、33个展开前端和4条浏览器共63个既有来源项的历史计数保留。主干KEEP/ADAPT/SPLIT及新增组另列，不能宣称只运行63项即可通过新矩阵。新helper、B seed、路由与测试只是计划条目，本轮没有实现。

## 3. 工时与复审入口

[净工时账](net-effort-estimate.md)及[结构化估算](../../../np-evidence/m0/net-effort-estimate.json)是本轮工时权威记录；逐条标注原已计、净增量和阶段移转，历史25–41/缓冲30–50保留对照。原W包、M包、作者验证、完整CI、独立执行不重复计量；M0文档/代理墙钟不计入剩余实现人日。

最终静态检查与主调度复审见[本轮复审证据](../../../np-evidence/m0/m0-review-revision.json)。历史m0-close-review.json/m0-close-validation.json是c8716b7原轮记录，未改写，不可用其PASS覆盖本轮或未来运行门禁。

## 4. 协作与结论边界

GLM Flash实际模型为`zhipu-coding/glm-5.3-flash`：Zeno提供后端初稿（t_5a4a0afa），Hegel提供B用例/fixture初稿（t_23c34a57）；主调度纠正源码与编号引用、接手依赖封板/JSON并完成矩阵/缓存/估算/复审/推送（t_2791e279）。不声称执行者完成了新一轮独立工时审查。任务状态只在Hermes的law-oa-np-main看板更新；此文件是固定交付说明，不是实时进度或长期记忆。

本轮文档修订复审结论以结构证据为准；即使纸面缺项处理完成，所有集成I/C运行门禁仍NOT_RUN、M1仍未授权启动。实际旧外部全量客户端、现网单实例切换、生产TLS及真实律师试用仍待后续独立验证，不把范围收缩或模型意见当通过证据。

最终主调度复审：**文档规则修订通过**。5项P1/3项P2均有依赖、验收、阶段与计量落点；剩余26–45人日、含缓冲32–54人日，算式与新旧去重账一致。此处没有新的Astra批准或集成运行PASS；M1实施/合并/部署继续HOLD。
