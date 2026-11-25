# PR #52 合并准入处置 · 2026-11-25

**合并 HOLD。决定当前不能合入的一项直接证据是普通 Go 集成测试在 main 通过，在 PR head 因重复注册 Prometheus 指标而失败。** 已有完整 I11 在实际 PR 事件通过。部署与小范围真实试用继续沿用[上轮各自的放行条件](../README.md)。评审任务为 Hermes `t_dec08c52`。

## 核验的代码和事件

| 对象 | 实际取证 |
| --- | --- |
| Git | 基线 main `d9c92fb`，head `89f2eb9`，测试合并提交 `e75945ff`。GitHub 返回该合并提交的两个父提交依次为 base、head。合并树和 head 的树对象均为 `051ee9a97c4bf96eb0e99cca3aea79894de175d1`。 |
| 实际 PR checkout | I11 的 runner 日志显示检出 `e75945f`，随后打印的 I11_REFS 与当时的 base/head 一致。PR 与 push 中的两个完整 I11 作业均成功，PR Test Pipeline 有其他作业失败。 |
| 当前检查 | 60 个 rollup 条目分属 6 次 PR 运行和 1 次 push。17 个成功、24 个跳过、13 个失败、6 个取消。19 个失败或取消均属 PR 事件。GitHub 显示 Draft、UNSTABLE；此次读取的 main 保护标记为 false，适用分支规则接口返回空数组。维护者仍须对拟合入树明确裁决。 |
| 证据范围 | 对照运行 ID、每个 job 的实际步骤、可取的历史日志和固定源码。7 次运行在查询时均有 0 个留存 artifact。单凭这个事后计数不能证明从未上传；PR 副作用隔离另有既有专门验证。 |

[固定 SHA、运行与 19 个失败/取消作业](checks.json) · [逐项根据与处置](causes.md) · [PR 的实际 I11 作业](https://github.com/VanTam-CN/law-oa-go/actions/runs/45398131150/job/138459274871)

## 第一项整改的依据和边界

1. Linux 托管 job `138459274338` 的集成测试打印了 `TestPerformanceMonitoring` 和 `duplicate metrics collector registration attempted`。堆栈指向 `internal/monitoring/performance.go:102` 创建等待时长直方图。该作业的报告步骤也运行整套测试并退出 1，未公开报告内第一条断言。
2. 我在同一宿主用 Go 1.25.6，以各自的全新 Git 归档执行同名测试。main 退出 0，head 退出 1，出现与托管一致的 panic。旧归档曾比较两个中间 PR head；这能排除当次格式修改引发故障，不能把此失败归入 main 基线。
3. 两套源码都声明了同名直方图。与 main 相比，head 性能测试的依赖闭包新增 `internal/database`；路由与处理器分别从 `c9084ae`、`2464646` 引入它。database 包初始化先注册 `db_connection_wait_seconds`，随后性能监控初始化又向默认注册表注册同名、同类型、同标签与同分桶的指标。此前还有其他同名候选指标，实际堆栈落在这个 HistogramVec。
4. 受控反事实核验仅改变一个环节。在一次性源码副本内移除 monitoring 中未被读取的私有字段和它的重复初始化，保留 database 持有的同名指标，其余代码保持相同。`-race -count=2` 的两个具名运行都通过。较早的临时 overlay 探针仍观测到原始 panic，已排除在正向证据外。仓库里的业务代码、工作流与 PR head 均未修改。[双 SHA 与反事实实测](monitoring-ab.json)

**建议下一项仅处理该指标的注册归属**，交接 Hermes `t_fc10a051`。实现时保留指标名、类型、标签和分桶，复核仅启动监控模块的调用方及已有仪表盘契约；若需要单独初始化也能注册该指标，用同一注册所有者承接。保留 router 与严格写入的 schema 检查。验证包括重复初始化、同名测试 race 回归、实际 PR 的普通 Go 测试和 I11。现有低成本证据已能选定整改对象；正式修改及推送实现按独立任务授权。

PGO 本地采样实际报出性能包失败，其采样脚本把详细输出留在临时文件中。benchmark 作业只公开命令失败，标准输出另写文件。两者共用的用例值得优先复跑，当前证据不足以宣称修这一处能关闭这两个作业。复跑若仍红，只提取首个具名失败和脱敏原因，保留原退出码。

## 旁支如何使用

| 旁支或历史处置 | 这次能采用的证据 | 仍须满足的条件 |
| --- | --- | --- |
| [#21](https://github.com/VanTam-CN/law-oa-go/pull/21) `e5c9704` | 已有 Go flag 探测和 Redis 准备候选。旧脚本的 `go build --help` 与 `go help` 都不能据以否定当前工具链支持 PGO/fuzz。本地 Go 1.25.6 用 `go help build` 和 `go help testflag` 分别读到了对应 flag。 | 只评估命中当前失败步骤的改动。两处探测改正后仍须见到后续真实作业结果。 |
| [#23](https://github.com/VanTam-CN/law-oa-go/pull/23) `9160139` | 指出旧 PGO 工作负载缺失，方案涉及改为手动触发。 | 当前 `check-pgo` 的本次红灯发生在支持检查，实际尚未到旧 workload。整份旁支工作流不能直接替换 #52 的专用 PGO 作业；历史卡 `t_62697ded` 另行处置。 |
| [#54](https://github.com/VanTam-CN/law-oa-go/pull/54) 及 [fuzz 目录证据](https://github.com/VanTam-CN/law-oa-go/tree/4824c714b71e9022cdec40e27880d6835af6817a/docs/np-evidence/m2/fuzzing-observability) | `69258f7` 的运行 `45412813406` 经 API 核对属于 **pull_request**、成功，留存 16 份 artifacts。可借鉴 runner 服务准备与按目标归因。 | #52 的 PR 隔离约束要求失败和取消时也不上传原始产物。禁直接携入无条件上传、不能把 #54 的 WorkerPool 缺陷算成 #52 本次失败；#52 该目标已过。 |
| [#54–58 lint 治理入口](https://github.com/VanTam-CN/law-oa-go/pulls?q=is%3Apr+is%3Aopen+base%3Amain+lint) | #54 的静态分析治理与 #55–58 的专项取舍均有各自工作树、检查和负责人。#52 的格式步本轮成功，静态分析另红。 | 这些开放旁支的 head 均不在 #52；不能只靠删除规则、修改门禁或旁支的绿灯关闭本次静态分析。选定治理方案后需在实际集成树上复验。 |

取消的作业不具备通过证据。部分作业未进入目标步骤，另有一个最终标为取消的缓存作业已经在设置 Redis 时出错。保持原平台结论，逐项验收按实际执行阶段判断。对 runner、基线或旁支问题的归因均不构成维护者豁免。若 base、head、test-merge 任一变化，重新取证并按差异选择回归；发布与真实资料使用继续分别审批。

## 独立终审

6.2 Astra 对固定评审提交 `720f8e5c682720527f289d68ef065df0df8aa4d6` 独立只读审核，同意三项 HOLD 及首项整改取舍。审核逐项比对 7 次运行及 19 个作业 ID，抽核两个在取消前已出现服务错误的作业，复算临时替代文件的 SHA-256，并核实 #54 运行及其产物边界。Astra 读取协调者的运行记录，未宣称独立重跑 Go 测试。推送前再次读取 #52，三个 SHA、Draft/UNSTABLE 和四类作业计数均与冻结快照一致。

## 后续整改

本页冻结在 `89f2eb9`。指标修复及其在后续候选上的独立测试、实际 PR 检查和剩余失败，见[数据库等待指标整改记录](metric-fix/README.md)。

`596f848` 的后续回归已归入[健康检查测试夹具与托管验收](health-fixture/README.md)。

`71ebc94` 的接续证据见[普通 Go Tests 覆盖率及 JUnit 整改](go-reports/README.md)。

性能包的 19 项基准、PGO 及真实 PR 作业结果见[性能夹具验收](performance-fixture/README.md)。
