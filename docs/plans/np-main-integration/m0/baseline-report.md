# Law OA 主干集成 M0：基线核验与审核修订

日期：2026-11-20。方案基线2481a47，应用基线d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7；执行分支codex/np-main-m0。Go/前端源码相对应用基线无差异。本轮不进入M1、不修改业务/迁移/workflow，不合并、不部署、不使用真实案件或服务器。Hindsight未调用，不作为阻塞。

## 1. 看板与执行责任

Hermes看板`law-oa-np-main`是任务状态主记录，默认看板未切换；本报告保存执行快照和证据，不复制实时台账。查询：`hermes kanban --board law-oa-np-main list --json`。

| 工作组 | 卡片 | 分工 |
|---|---|---|
| 接入探针 | t_562fd924 | Codex操作员创建/幂等回读/评论/阻塞恢复/提交复审/完成回读；无自动worker |
| 基线与隔离回归 | t_9fe3809b | Codex运行并检查事实 |
| 后端最小闭包 | t_edb90c67 | GLM Flash A核查，Codex复核代记 |
| 前端/CI闭包 | t_a56dfe09 | GLM Flash B核查，Codex复核代记 |
| 出口与取舍 | t_cbf2406f | Codex综合裁决，保留未决与边界 |

接入范围为操作员代记，不证明原生worker自治、任务权限隔离、并发抢占或常驻调度。探针原失败保留：第一次unblock误将说明文字当成第二任务；重复同类阻塞触发triage防循环；show.task没有current_run_id，应从runs读取；短命CLI认领不能当作活跃worker所有权拒绝测试。未修改Hermes规则或使用force绕过。旧探针已加更正说明并归档。

## 2. 实际运行结果

| 检查 | 结果 | 适用边界 |
|---|---|---|
| `go build ./...` | PASS，exit 0 | 本机Go1.25.6，不是远端CI Go1.25.0 |
| handlers/auth `go test -json -race -count=1` | 138 RUN，132 PASS，6 SKIP，0 FAIL；exit 0 | 只覆盖两个包，不是全库；6项Skip保留为未覆盖，不能当作已通过 |
| `npm ci --no-audit --no-fund` | PASS | 锁文件安装；HUSKY=0保护共享Git hooks，PUPPETEER_SKIP_DOWNLOAD=true，不声称浏览器已测 |
| `npm run type-check` | PASS | Node22.17.0/npm11.13.0，与CI Node20不等价 |
| 两套登录Jest | 2 suites/4 tests PASS | 未冒充源首切片5套33项或NP组件已移植 |
| `npm run build` | PASS | 构建成功不证明原生保存/来源链路 |
| PG16.14显式bootstrap | PASS，v14；已采集实际schema | 全新合成容器、回环端口18894；migration使用development配置，执行BootstrapProductionSchema，不靠应用AutoMigrate；生产TLS/main+Redis未测 |

PG原始失败全部保留：首轮使用了不绑定的APP_ENV并遇连接拒绝；第二轮production+sslmode=disable被配置校验拒绝。随后按M0合成schema探针范围使用development迁移配置、显式bootstrap取得v14结构；这不消除生产TLS联调门禁，也不把前两次失败裁为应用缺陷。任务创建的PG容器、匿名卷及网络均已定向清理，没有触及其他环境。

六项跳过：TestAuthHandler_GetProfile、TestAuthHandler_UpdateProfile、TestAuthHandler_ChangePassword、TestAuthHandler_RefreshToken/Refresh_Token_Success、TestAuthHandler_RefreshToken/Refresh_Token_Invalid_Token、TestCaseHandler_CreateCase。新NP安全门禁不得沿用这些Skip放行。

## 3. I09业务决定与最小范围

用户明确选择B：主干首轮只验case_materials及实际已有相关记录/主体/对照草稿保全；I09-MAIN必测，I09-FILES两版文件流DEFERRED/HOLD。实际PG有case_materials，没有material_files；不能以空表、schema存在或构建通过证明材料保全。两版文件源用例保留为后置清单，不删断言或假造PASS。该决定不授权M1。

## 4. 证据

- [看板读写探针](../../../np-evidence/m0/kanban-probe.json)、[首失败](../../../np-evidence/m0/kanban-probe-first-red.json)、[v2部分探针](../../../np-evidence/m0/kanban-probe-v2-partial.json)。
- [建组回执](../../../np-evidence/m0/kanban-setup.json)、[Go运行元数据](../../../np-evidence/m0/go-baseline.json)、[结构化Go事件](../../../np-evidence/m0/go-test-events.json)。
- [前端运行元数据](../../../np-evidence/m0/frontend-baseline.json)、[PG初次失败](../../../np-evidence/m0/pg-first-attempt.json)、[PG第二轮](../../../np-evidence/m0/pg-rerun.json)、[最终schema运行](../../../np-evidence/m0/pg-schema-baseline.json)、[实际结构](../../../np-evidence/m0/pg-schema.txt)。
- [原始日志位置与散列](../../../np-evidence/m0/raw-log-index.json)：原始正文/SQL/测试日志仅留本机隔离输出目录，不提交到Git；提交结构摘要与命令、时间、退出码、散列。

当前运行结果不继承至主干新I/C验收矩阵。

## 5. GLM辅助核验与主代理裁决

[后端报告](backend-closure.md)由GLM Flash A落盘，主代理复核并修正引用；[前端/CI报告](frontend-ci-closure.md)来自GLM Flash B的已读取源码及草稿，主代理接手整理，修正请求/回执混淆、旧迁移编号及PR job可达性。两项均不是独立运行验收，详情见[代理复核记录](../../../np-evidence/m0/agent-review.json)。

| 取舍 | 核验结论 | 工作量归属 |
|---|---|---|
| 认证/会话 | main已有AuthTokenSession与持久会话存储，原有回归可执行 | M1复用，不重复移植源会话 |
| 接案版本/旧写入者 | main缺draft_revision/expected_revision，已有事务不等于锁内版本保护 | M1原生读写/恢复与相关写入者统一整改；不能让助手包再计一次 |
| 异步检测载体 | 现有conflict_check_records.search_parameters可承载服务端输入revision；请求/关联结构目前没有此字段 | M1在现有结构传递，不新增任务表/队列平台；同步二次回写一并收口 |
| NP核心两表 | 新合成v14库直接应用源86迁移，两个唯一索引和外键均成功，未执行82–85；无material_files | 结构依赖已证，不代表正式迁移编号/整套fixture/业务写入通过；归M1/M3各自实现，验证只计一次 |
| 回执action | 请求侧action='update'必填，回执目前不返回；不能强制回执必填 | M3/M4核对request_id/id/revision；action如存在才比较，减少不必要接口修改 |
| 正文/恢复/退出 | 原生localStorage、单槽恢复、保存中继续输入及晚到保护仍有改造点 | M4一个前端包，C13/C14与已有回查一并验收 |
| SW注册尝试 | main尝试注册/sw.js；本轮构建未找到该文件，旧浏览器控制状态未测 | 归M4/M5的C10，不能静态推定无worker或已泄漏 |
| CI产物流 | 六套workflow须按实际job条件收口；ci-cd评论job不在PR路径 | M2局部改造，不整体重建流水线；原始产物与汇总一起处理 |
| 材料版本栈 | I09用户已选B | 本轮不计完整material_files平台；I09-MAIN保护仍必须实现/验证 |

[最小迁移依赖探针](../../../np-evidence/m0/np86-dependency-probe.json)与[实际索引/外键](../../../np-evidence/m0/np86-structure.txt)只证明源SQL可应用于主干v14。探针先在错误对象库查source SHA失败，已改为明确git -C源仓库取固定对象；未修改源文件或主干代码。临时PG已清理。

## 6. 第一轮出口快照（213da4c历史记录）

**以下是213da4c第一轮状态；其三项剩余文档工作已在后续§7收口。** 当时操作员看板接入PASS，完整M0出口HOLD；M1未启动。 接案版本、原生恢复、相关写入者与缓存/CI整改均属后续实现，不能以本轮构建成功替代。

继续M0的收口项合并登记在t_cbf2406f：

1. 路线B的逐用例测试清单：把源fixture、文件比较脚本和I09-FILES依赖明确拆出，保留原断言；确认新主干必跑集合及目录/依赖清单。当前没有修改测试代码。
2. 主干最小依赖最终封板：区分NP两表SQL已证兼容与读写服务、工厂、脚本/路由尚需移植，冻结迁移编号及升级读写兼容条件；正式迁移/回滚仍未执行。
3. 原W与M1–M6工作包净工时对照：按上表扣除会话、材料平台和共用验证的重复量；本轮已读回历史W01–W11范围并作下表归属核对，但该表基于final5，不是主干净工作量，不能直接加减成新总数。50–80/含缓冲60–96仅保留final5历史参考，不是main承诺。

上述任务属于继续M0的核验与计划收口，不授权M1。真实服务器运行版本、外部旧全量调用方、升级回滚单实例、生产TLS/main+Redis、浏览器保存/来源/真实断线、CI正常/负向及独立集成验收均未由本轮证明。

### 历史W包对照（范围去重，不作为主干报价）

来源是原工作区`docs/plans/2026-11-18-law-firm-assistant-i0-readonly-baseline-v0.4.md` §5，属历史计划；本轮只读，未修改或把脏工作区内容并入候选。

| 当前M包 | 原W归属 | 去重处理 |
|---|---|---|
| M0 | W01基线与协议 | 本轮执行证据计入基线；最终依赖封板仍要完成，不把文档审核当运行 |
| M1 | W02原生保存后端、W04来源事务 | main缺少final5原有版本锁，不能沿用W02“已抵减3人日”的前提；源首切片新增写协议已有代码，只计主干适配与缺失基础，不重复从零实现 |
| M2 | W10验证有交叉，CI配置实现未单列 | 实际测试/证据复跑与M5/M6只计一次；工作流实现净增量单列后估，不把六套整体重建计入 |
| M3 | W04/W05/W06 Stub/W08及W07挂接 | 源首切片已有上下文/候选/来源/保存实现，按净增量移植；当前不新增真实提供方或完整模型评测 |
| M4 | W03内容保护、W07原生前端、W08部分交互 | 缓存、恢复键、退出/重载与回执校验合并计，不能在M3再计一次实现；SW边界并入既有C10 |
| M5/M6 | W10全链与独立复核 | 同候选作者集成与独立执行分别记录实际工作，但每次验证不同时算进多个包；源PASS不能省略主干验证 |
| 当前不纳入 | W09完整模型技术评测、W11律师模拟准备及真实部署 | 不把全窄版历史预算直接当首切片主干集成预算；后置不等于这些工作已完成 |

路线B排除了当前新建材料版本平台，但旧估算本来假设该平台可复用，不能再从50–80中虚减一份原本没有单列的材料平台人日。新主干总量待最小依赖/夹具封板后按净工作项重估。

### 看板回报实际复核

本轮三个执行卡的交接实际经过request-review→readback→complete→readback，见[交接回执](../../../np-evidence/m0/kanban-handoff.json)。期间原15分钟认领到期，旧run_id提交复审被拒绝；主调度以看板实时状态确认ready且无现任执行者，重新认领后按新run提交，没有force。创建回执中的run_id不是长期有效凭据，实时状态必须回读；该观察不等于原生自治worker权限隔离验收。

## 7. M0三项封板快照（c8716b7历史记录，已被独立审核要求修订）

| 交付 | 结果 | 固定证据 |
|---|---|---|
| 路线B逐用例拆分 | 文档封板PASS：26个Go/33个展开前端/4条浏览器既有来源项；另列新增安全/缓存及2个详情适配用例，均未在新main运行 | [逐项清单](route-b-test-split.md)、[结构化索引](../../../np-evidence/m0/route-b-test-inventory.json) |
| 最小依赖 | 文档封板PASS：33条路径/blob逐项复核、8个原生扩展目标、2个计划基础文件；main B=v14+86，不引入完整材料/续办平台 | [依赖封板](dependency-seal.md)、[JSON](../../../np-evidence/m0/dependency-seal.json)、[对象复核](../../../np-evidence/m0/seal-blob-verification.json) |
| 净工时 | 已去重复审：剩余M1–M6 25–41人日，20%缓冲后30–50人日；仅本首切片条件区间 | [估算与成立条件](net-effort-estimate.md)、[可计算账](../../../np-evidence/m0/net-effort-estimate.json) |

以下保留c8716b7当轮三项文档交付结论；其封板已被后续§8独立审核置为HOLD，不能据本段启动M1。原记录为：本轮三项M0文档/证据收口通过，可以作为下一阶段评审输入；**不是M1已实施或集成/试用门禁通过**。实时结果/阻塞记录在Hermes t_cbc0ab2c、t_7f6ad58e、t_ab617f55及出口t_cbf2406f；[复审记录](../../../np-evidence/m0/m0-close-review.json)保存固定评审对象和裁决，不承担实时状态。

关键范围裁决：初始化脚本拆为v14+86；continuation正向提交/复核能力后置，陈旧保护复用现有PUT拒绝，GET重载能取最新；保护比较不依赖material_files，按整行规范文本保护非空材料、客户/主体及对照草稿。原始源测试与文件断言不改、不Skip，不用删断言制造绿色。

本轮仅做静态封板/结构核对/估算复审，没有重复上一轮Go/npm/PG运行；其历史证据保留，不自动覆盖未来集成候选。实际旧客户端清单、单实例切换、生产TLS/main+Redis、真实断线/浏览器/CI正常负向与升级回滚仍按M1–M6各关口验证；外部旧请求不承诺透明兼容，若必须长期兼容或多实例，则触发重新估算。未进入M1、未合并或部署。


## 8. c8716b7独立审核与本轮文档修订

实际模型`gpt-6.2-astra`对固定c8716b7进行独立只读审核，主调度逐项查证后采纳0项P0、5项P1、3项P2。该轮Hermes t_e3878bd1已完成的是审核交付，候选纸面封板结论为HOLD；上文历史Done/PASS均不表示M1准入。本轮用户授权glm-flash配合修订文档/证据和推送，未授权任何M1实现。

处理规则见[实施方案§3](../implementation-and-acceptance-plan.md)、[依赖清单](dependency-seal.md)、[B逐用例](route-b-test-split.md)及[本轮处置记录](astra-review-resolution.md)：

| 本轮编号 | 审核问题 | 规则责任与未验事项 |
|---|---|---|
| M0-R01 / P1-1 | 原生恢复路由和真实router验证漏列 | M1明确list/detail/alias/dashboard读契约及200/403/业务404、409后重读最新；M5原菜单重开；NOT_RUN |
| M0-R02 / P1-2 | 主体身份保全及授权db工厂依赖漏列 | M1先准备身份后替换并验证回滚，M3复用现有四repo/tx授权，客户切换和对方继承分开；NOT_RUN |
| M0-R03 / P1-3 | I15测试状态不可达且跨阶段绑定 | M1原生事实修改与旧完成、服务级async完成；M3 NP来源联动；M5真入口；公共async拒绝不是完成回写证据；NOT_RUN |
| M0-R04 / P1-4 | 旧成功回执夹具缺身份/旧单槽断言 | M3/M4 ADAPT正向、保留缺项负向及v2逐请求断言；源33项为历史来源数；NOT_RUN |
| M0-R05 / P1-5 | M2完整I11依赖M5才交付的fixture | M3准备/M4验证B fixture，M2早期配置可先做，完整正常负向CI此后完成再到M5；NOT_RUN |
| M0-R06 / P2-6 | B非空材料及大整数seed缺明确责任 | M3至少两材料、非空关联与精确IDs；原四场景和新负向独立窗口；NOT_RUN |
| M0-R07 / P2-7 | 86追加迁移和备用entry边界不完整 | M1受控事务/校验和/已应用schema核验，共享router保护覆盖备用装配；支持入口仅root；M6合成I12演练；NOT_RUN |
| M0-R08 / P2-8 | dev前端无法验证PROD SW分支 | M4同候选PROD产物及受控旧worker，M5真root API；实际缺资产如实记录，旧终端不明HOLD；NOT_RUN |

路线B选择继续有效，保全门槛不削弱；I09-FILES维持DEFERRED/HOLD。33条原映射/blob是历史固定事实，新增最小helper/调用点在修订依赖账明示，不篡改旧blob复核记录；结构兼容探针也不升级为迁移演练PASS。

当前工时以[本轮净工时账](net-effort-estimate.md)为准，历史25–41/缓冲30–50暂估保留作对照；八条意见按原包已计、遗漏净增量和阶段转移分别计量，不能按问题条数叠加，也不能拿缓冲吞范围变化。M0已发生核验/文档/代理耗时不混入剩余实现人日。

本轮Hermes分工：t_5a4a0afa依赖修订、t_23c34a57验收/夹具修订、t_2791e279主调度矩阵/估算/复审交付；固定文件留证据，实时进展只查law-oa-np-main看板。本轮仅静态文档检查，不重跑Go/npm/PG/Redis/浏览器或远端CI，不接真实服务器。纸面修订完成仍不代表M1已获授权，不合并、不部署。

本轮去重重估：**26–45人日，含20%缓冲32–54人日**；M1主体保全净增1–2、M3回执/B夹具共同上界+1、M6合成升级/重复状态/回滚演练上界+1，其余原已计或只移阶段。上界并非固定交付保证。
