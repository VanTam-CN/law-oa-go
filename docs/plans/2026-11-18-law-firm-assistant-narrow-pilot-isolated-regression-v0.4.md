# NP v0.4 首切片隔离回归准备

2026-11-18（Asia/Shanghai）。本轮完成候选冻结、执行清单和虚构夹具定义。编译、建库、服务启动、用例执行和切换演练均为 NOT_RUN。

配套 [设计第 13 节](2026-11-18-law-firm-assistant-narrow-pilot-design-v0.4.md#13-实施顺序与首个切片待另行授权)、[验收矩阵](2026-11-18-law-firm-assistant-narrow-pilot-acceptance-v0.4.md)及 [I0](2026-11-18-law-firm-assistant-i0-readonly-baseline-v0.4.md)。

## 1. 候选与隔离环境

| 项目 | 准备情况 |
|---|---|
| 源码快照 | 本地 `output/np-v04-final5-dev-baseline` 是独立 Git 仓库。提交 `bcd1b8992b3f3450d5ab406afd74336cc282eece`，树 `e67f1564882109e2c149bba93e4b0a56558c5ab7`。提交内的 1,112 个源码文件逐一匹配交付清单，另保留原 `SHA256SUMS-SOURCES.txt`，清单 SHA-256 为 `aca2483dffd957ca857c44e3596773a7f1b06c3826a3701029d27209e483f61d` |
| 文件范围 | 655 个后端、457 个前端源码文件，含 Go / npm 依赖清单、原测试和 SQL。取自 `output/trial-server-delivery-20261008-final5`。旧工作树差异纳入 0 项。环境和构建准备参考原交付包中的部署模板，另外生成隔离配置 |
| 本机工具 | 已核对 Go 1.25.6、Node 22.17.0、npm 11.13.0、Docker CLI 29.6.2，当前 context 为 `desktop-linux`。未核对 daemon、浏览器二进制、镜像和依赖安装，可构建性待后续证据 |
| 执行环境 | 后续用新工作副本及独立项目名、PG 16.14 空库、Redis、临时文件卷、全新的虚构账号和仅回环入口。每轮记录候选 SHA、镜像标识、数据库名、宿主端口及数据卷标识。PG/Redis 无宿主映射，服务端 API 仅指向本轮隔离实例 |
| PG 初始化 | 使用候选的 bootstrap，再用归档内四个明确的追加迁移 `000082_intake_collaboration_mvp`、`000083_approval_supplement_replies`、`000084_matter_collaboration`、`000085_case_task_change_audit` 的 up 文件；先核对空库，停止于首个失败。准备参考交付包 `scripts/fresh-db-checklist.sh`，跳过全量历史迁移 |
| 依赖和浏览器 | 后端在快照的 `backend` 下使用 go.mod/go.sum。前端从 `package-lock.json` 安装。现有 Playwright 配置默认用回环 3003 且非 CI 时允许复用服务，Vite API 默认指向 8080。隔离跑时明确锁定 API 代理、使用新浏览器上下文并禁止复用任意已运行前端，确认无意外外部调用 |

复核脚本入口可用 `git rev-parse HEAD` 和 `shasum -a 256 -c SHA256SUMS-SOURCES.txt` 检查。在快照之外创建工作副本进行构建和试验，保留冻结提交作为基线。后续批准导入的本地差异需逐笔登记，重新生成候选 SHA。仅工具可见和哈希相符不代表构建已完成。

## 2. 现有基线回归准备

以下是已查到的原生测试入口和相应证据。执行前先确认工作副本及数据库连接边界。原生测试和后续首切片断言使用不同候选标识。

| 目的 | 现有用例或待执行入口 | 当前状态 |
|---|---|---|
| 已有草稿读取及版本响应 | 后端 `internal/handlers/intake_draft_revision_test.go` 的 `TestDraftRestoreReturnsRevisionAndProtectedIdentityStatus`、`TestDraftListCarriesRevisionTokens` | NOT_RUN |
| 缺版本、旧版本和非目标 metadata | 同文件 `TestUpdateIntakeRejectsMissingExpectedRevisionWithoutMutation`、`TestUpdateIntakeRejectsStaleRevisionWithoutOverwrite`。另见 `intake_draft_revision_contract_test.go` 的 `TestUpdateIntakeWithoutMetadataKeepsStoredBusinessFields` | NOT_RUN |
| 原生受保护主体和材料 | 后端 `intake_party_identity_guard_test.go` 的 `TestUpdateIntakeDraftInheritsProtectedIdentityWhenInputEmpty` 和 `mvp_material_independent_test.go` 的 `TestIndependentMaterialVersionsDownloadOriginalBytes` | NOT_RUN |
| 原生浏览器与真实 PG | 新库内以虚构记录从原菜单打开已有草稿。核对旧版本冲突时保留未保存输入，单独检查原生主体、材料操作。在隔离 PG 对照相关行 ID、引用和字段。SQLite/Mock 证据单列 | NOT_RUN |

首批可用命令入口，待执行阶段从工作副本的 `backend` 目录运行：

~~~sh
go test ./internal/handlers -count=1 -run '^(TestDraftRestoreReturnsRevisionAndProtectedIdentityStatus|TestDraftListCarriesRevisionTokens|TestUpdateIntakeRejectsMissingExpectedRevisionWithoutMutation|TestUpdateIntakeRejectsStaleRevisionWithoutOverwrite|TestUpdateIntakeWithoutMetadataKeepsStoredBusinessFields|TestUpdateIntakeDraftInheritsProtectedIdentityWhenInputEmpty|TestIndependentMaterialVersionsDownloadOriginalBytes)$'
~~~

## 3. 首切片的虚构夹具和验收准备

只选一份**已有草稿**。使用模拟模型，标题照旧，律师核对来源后只应用摘要候选，再走原生保存。重登或打开新浏览器会话，由原菜单重开同一草稿并核对已保存来源。新建、标题候选和真实模型评测在后续切片覆盖。

夹具为一名有权虚构律师、一名无权虚构律师、一个虚构客户、目标草稿 D1 和用于防错对象的对照草稿 D2。D1 原标题「虚构接案·甲」，原摘要含「原摘要标记」。至少带一个稳定主体行、一个材料及两版纯文本假附件、一个无关 metadata 键。记录原版本 R0、原生表按 ID 排序后的比对摘要和材料引用。提交时使用实际 ID 和实际服务端版本，文档中的 D1/R0 仅为阅读代号。

允许的模拟模型输入为「这是虚构资料。因对方未按时交货，我方要求确认交付时间。」模拟模型的拟用摘要保留待核实口径，来源应回查到实际提交的虚构原句及人工选择。输入出现新一轮时沿用 v0.4 的全部旧候选失效规则。原生工作台不因为助手应用而自动改标题、主体、材料、办理状态或未知 metadata。

| 首切片场景 | 准备操作及可判定结果 | 归属 / 当前状态 |
|---|---|---|
| 原生保存后重开 | 从原菜单恢复 D1。查看并核对虚构来源，只应用摘要并点原生保存。保存回执给出同一 ID、实际新版本和来源引用。重登后由原菜单打开 D1，摘要和来源可见且对应该版本，对照 D2 与非目标字段/材料/主体 ID 保持 | 本切片取 I02/A28/A31 的适用层，来源一致性仍按 A15–A17。NOT_RUN |
| 并发防覆盖 | A、B 两个有权标签均读 R0。A 使用当前来源保存，B 在原表单手工改另一段摘要，B 请求不含来源选择。先让 A 保存成功并记录新版本，再提交 B 的旧版本编辑。B 必须因业务版本冲突收到提示并保留未保存输入；库中摘要、来源、版本仍对应 A，B 没有额外快照或成功回执 | I03/A24 的已有草稿分支。NOT_RUN |
| 断线后回查和同键重试 | 对 D1 的 PUT 在已提交、客户端未收响应时实际切断连接。先记录 PG 已提交，再检查浏览器显示待核实。重登后用原键 GET 查原版本和来源；在能够重建完整同参时，同键重试返回原结果。库中仅一次版本推进、一份对应来源快照、一份持久化结果。另核对同键异参拒绝 | I04/A25/A27 的更新分支。NOT_RUN |
| 未提交、未知与明确放弃 | 真实断线覆盖提交前窗口，记录实际回滚或提交结果。无回执时只读查询仍返回未知，不给新键。明确放弃先获得同键并提交时，延迟原请求不得写；保存先提交时返回真实已存结果。按 A27 分别留证，不用只模拟 UI 报错替代 | I04/A27/A40 的适用更新分支。NOT_RUN |

验收记录为候选 SHA、数据集代号、请求键和版本的脱敏关联、连接断开与提交的时序、浏览器结果及 PG 前后取证。每条保留缺失证据和首次失败。首切片只更新所覆盖的子断言，所有 A 项总状态仍按矩阵完整要求裁定。升级和回滚各一次的同候选隔离演练另记在开放试用前，包含单实例、前后端版本匹配、临时会话失效、可回查回执及来源保全。

## 4. 开始执行前仍需核对

- 冻结拟执行候选的工具链、镜像版本、端口和数据边界；明确授权人员、来源可读角色以及旧全量调用方的迁移或兼容安排。
- 访问密钥与试验数据使用独立配置。正式试用入口的运维恢复单列，不影响准备隔离夹具。
- 首切片只用 Stub。待 G01/G04 取得实际证据、G05 技术评测通过并完成开放试用前切换演练，再按 v0.4 评估开放模拟试用。
