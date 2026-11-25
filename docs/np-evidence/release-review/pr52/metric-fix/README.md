# PR #52 数据库等待指标整改

Hermes `t_fc10a051` 对应的代码是 [fd92ef0](https://github.com/VanTam-CN/law-oa-go/commit/fd92ef0b86ede74e9f791f8295daf3015d0e55f0)。父提交为 `89f2eb9`。本轮只修改数据库和监控两个调用点，新增一个共用的指标定义。两者在各自原有的时机取得同一个直方图，首次注册的错误仍会抛出。工作流、数据库配置和路由均沿用父提交。[上一轮红灯归因](../README.md)

## 隔离复验

协调者从两个固定提交导出干净源码，在独立进程中使用默认 Prometheus 注册表。原测试在父提交因重复注册退出 1；修复后使用 `go test -race -count=2 -run '^TestPerformanceMonitoring$' ./tests/performance`，两个具名运行均通过。修复后全库编译通过，数据库包的三条测试在竞态检测下通过。监控包与新增的指标包没有包内测试，单独加载两者的行为由进程探针核验。

| 进程中加载的包 | 监控初始化前 | 修复后的结果 |
| --- | --- | --- |
| 仅数据库 | 已注册 | 两条样本，累计 0.375 秒 |
| 仅监控 | 尚未注册 | 连续初始化两次后写入一条样本，累计 0.25 秒 |
| 数据库与监控 | 已注册，写入 0.125 秒 | 连续初始化两次，仍是原收集器，再写入 0.25 秒后累计 0.375 秒 |

探针用相同描述符向默认注册表查询原收集器；原先缺席时立即撤下探针本身。最后通过 Gather 核对单个指标家族、类型 Histogram、名称 `db_connection_wait_seconds`、帮助文本 `Time spent waiting for database connections`、唯一标签 `database`、11 个默认有限分桶及样本数和总和。父提交的两种单包场景通过，组合场景复现原 panic。样本全部为合成值。运行标识及结构化结果见[验收记录](verification.json)。

## 独立审核

6.2 Astra 从同一提交另行导出干净源码，独立编写并执行三种加载场景及原测试，四项均以竞态检测运行并退出 0。原测试产生两条具名 PASS。在组合进程中，Astra 单独核对监控字段、共享访问器与默认注册表的对象身份；初始化前的 0.125 秒仍可采集，再加 0.375 秒后为 2 条、累计 0.5 秒，各桶的累计样本数也符合预期。Astra 判定这项修复通过。下表的托管结果由协调者另行核实。

## 实际 PR 检查

GitHub 提供的测试合成提交为 `1458ad2f73b0dd2870a324babb77b6ee698f567a`，父提交依次是主干 `d9c92fb` 和候选 `fd92ef0`。合成提交与候选的 Git 树相同。Astra 另行读取了普通 Go、PR 完整 I11 和试点 PGO 的托管日志，确认以下三个作业的判定。

| 作业 | 结论与证据边界 |
| --- | --- |
| [PR 普通 Go Tests](https://github.com/VanTam-CN/law-oa-go/actions/runs/45564822129/job/139036946762) | **失败**。集成步骤明确打印 `TestPerformanceMonitoring` PASS。具名失败是 `TestConflictP0ReadinessCheck_ProductionBadCanonicalHashFailsClosed`，预期 unhealthy，实际 healthy；报告步骤也退出 1，未公开独立的失败断言。 |
| [PR 完整 I11](https://github.com/VanTam-CN/law-oa-go/actions/runs/45564822129/job/139036946425) | **通过**。日志绑定上述 head、base 和实际合成提交；17 个阶段均退出 0，路线 B 四个场景通过，最终输出 `I11_FULL_PASS`。 |
| [push 完整 I11](https://github.com/VanTam-CN/law-oa-go/actions/runs/45564814379/job/139036923767) | **通过**。直接检出候选 `fd92ef0`，17 个阶段成功并输出 `I11_FULL_PASS`。同事件的普通 Go Tests 按规则跳过。 |
| [PR 试点 PGO](https://github.com/VanTam-CN/law-oa-go/actions/runs/45564822447/job/139036948976) | **失败**。本轮采样停在 `law-oa-go/internal/health`，标准与优化二进制对比尚未执行；公开输出没有失败用例名。上一轮停在 `tests/performance`。 |
| [旧 PGO 支持检查](https://github.com/VanTam-CN/law-oa-go/actions/runs/45564822447/job/139036948624) | **失败**。Go 构建帮助输出未通过脚本的 PGO 支持探测。该路径的后续性能基准作业跳过，与上面的试点采样、下面的全库基准测试分别判定。 |
| [PR 全库基准测试](https://github.com/VanTam-CN/law-oa-go/actions/runs/45564822311/job/139040254741) | **失败**。全库 benchmark 命令退出 1，标准输出重定向到未上传的 benchmark.txt。公开 stderr 只有 SQLite 编译警告，未显示首个失败包或断言。 |

## 剩余失败与下一项任务

- **普通 Go Tests。** 目标提交与主干使用相同的健康检查测试文件。夹具把哈希首字符替换为 0，首字符原本为 0 时实际值不会改变。该机制与本次公开失败表现一致，托管日志未保留输入哈希。Hermes `t_861ce93c` 跟踪目标分支验收；既有单文件测试修复 `9005b61` 已在候选上通过 `git apply --check`，尚未应用。生产校验及 unhealthy 断言仍须保留。[旁支修复](https://github.com/VanTam-CN/law-oa-go/commit/9005b61fdac5e25cbdfecf468cbfe437c6cc3466)
- **全库基准。** 协调者分别从 `d9c92fb`、`89f2eb9`、`fd92ef0` 导出源码，编译 `tests/performance` 的竞态测试二进制，禁止测试进程联网。仅执行 `BenchmarkAPIRequest/GetCases`、一次迭代，三个版本均在 HTTP 404 断言处失败并退出 1。Hermes `t_597f0f6a` 跟踪夹具的路由初始化与接口路径。这只确认三个版本有相同的可复现症状。
- **指标契约复跑。** [协调者使用的原始进程探针](contract-probe.py) 接受两个参数，依次为一次性干净源码目录和 `before` 或 `after`。在候选归档上重跑归档探针，输出与首次复验逐字节一致。仅提交脱敏摘要与探针，原始 CI 输出和本机临时路径均未纳入。

**裁决。** PR #52 的指标重复注册已关闭。合并、部署、小范围真实试用继续 HOLD。此次仅对上表的托管作业作出判定；其他检查和既有放行条件仍按各自证据单独处理。
