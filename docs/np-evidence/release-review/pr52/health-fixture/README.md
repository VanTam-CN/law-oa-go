# PR #52 健康检查测试夹具

Hermes `t_861ce93c` 承接[上一轮普通 Go Tests 的具名失败](../metric-fix/README.md)。本轮从 `fd92ef0` 仅复用[旁支的测试补丁](https://github.com/VanTam-CN/law-oa-go/commit/9005b61fdac5e25cbdfecf468cbfe437c6cc3466)，修改 `internal/health/conflict_p0_check_test.go`，增加 9 行、删除 1 行。两份补丁的 Git patch-id 相同。[候选提交](https://github.com/VanTam-CN/law-oa-go/commit/596f8480f619e730411d6fc9af0134e74b06ef46) · [结构化证据](verification.json)

旧夹具将原哈希首字符替换成 0。如果首字符本来就是 0，数据库仍保存正确哈希，导致健康检查返回 healthy，与用例期望相反。新夹具先从合成 SQLite 库读取 G1 的哈希，首字符为 f 时改成 0，其余情况改成 f。它保留 64 位十六进制格式，明确断言值已经变化，再保存并校验系统返回 unhealthy 且提示 G1 完整性哈希不匹配。生产健康检查及工作流均沿用父提交。

## 隔离复验

| 执行方与源码 | 结果 |
| --- | --- |
| glm-flash，候选工作树 | 具名用例执行一次，退出 0 |
| 协调者，父提交干净归档，具名测试开启竞态检测 | 100 次中 93 次通过、7 次失败，包退出 1 |
| 协调者，候选干净归档，相同命令 | 100 次全部通过，包退出 0 |
| 协调者，候选干净归档，整个 health 包开启竞态检测 | 65 个具名测试全部通过，0 跳过，包退出 0 |

PR 测试合成提交 `d5594f3e49cb67c386b9673253aa7b5651f6d870` 的父提交是主干 `d9c92fb` 与候选 `596f848`，其 Git 树与候选相同。下文分别记录合成提交和直接推送候选的作业。

## 独立审核

6.2 Astra 从候选提交另行导出源码。在临时 SQLite 中以竞态检测运行“完整证据通过、坏哈希拒绝、规范 SHA-256 校验”三个原有用例，每个执行两次并通过。它还逐一覆盖 16 种合法首字符，每组先验证原始记录为 healthy，再执行候选的篡改步骤并从数据库读回。16 组都只改变首字符，保留后 63 个字符和合法格式，随后均返回 unhealthy 并指出 G1 哈希不匹配。两项验证各自退出 0。Astra 据此判定夹具修复通过。

## 托管验收

| 检查 | 观察结果 |
| --- | --- |
| [PR 普通 Go Tests](https://github.com/VanTam-CN/law-oa-go/actions/runs/45568119676/job/139047819429) | **作业失败**。单元测试、集成测试步骤均成功，G1 坏哈希用例分别具名 PASS，上轮修复的 `TestPerformanceMonitoring` 也通过。 |
| 同作业的报告步骤 | 合并覆盖率前安装 `gocovmerge@latest` 时，解析到的 `golang.org/x/tools v0.53.0` 要求 Go ≥1.26.0，runner 为 1.25.0。随后生成 JUnit XML 的管道也退出 1，日志没有公开其失败断言。两项跟踪于 Hermes `t_a397c2ab`。 |
| [PR 完整 I11](https://github.com/VanTam-CN/law-oa-go/actions/runs/45568119676/job/139047819155) 与 [push 完整 I11](https://github.com/VanTam-CN/law-oa-go/actions/runs/45568110873/job/139047794451) | **均通过**。两份日志各核实 17 个阶段退出 0、路线 B 四个场景通过、最终输出 `I11_FULL_PASS`。PR 实际检出 `d5594f3`，push 实际检出 `596f848`。 |
| [旧 PGO 支持检查](https://github.com/VanTam-CN/law-oa-go/actions/runs/45568119762/job/139047819470) | **失败**。构建帮助输出仍未通过脚本的支持检测，其依赖的旧“性能基准测试”作业跳过。 |
| [PR 全库 benchmark](https://github.com/VanTam-CN/law-oa-go/actions/runs/45568119726/job/139050589495) | **失败**。基准命令将输出重定向到 `benchmark.txt`，退出 1。公开日志显示 SQLite 编译警告，未指出首个失败包或断言。沿用 Hermes `t_597f0f6a` 跟踪。 |
| [试点 PGO](https://github.com/VanTam-CN/law-oa-go/actions/runs/45568119762/job/139047820037) | **失败**。本轮采样到达 `law-oa-go/tests/performance`，与上一轮停在 `internal/health` 的位置不同。日志仅给出包名，无法断定哪条断言失败。后续两个应用二进制的构建和验证步骤跳过。 |

Astra 另行抽核实际 PR 的普通 Go Tests 和完整 I11，确认检出提交、具名结果和两个报告环节的失败边界。协调者跟踪 PGO 和全库 benchmark 至终态。夹具整改关闭，剩余报告与基准问题分别由 `t_a397c2ab`、`t_597f0f6a` 跟踪。合并、部署、真实试用继续 **HOLD**。
