# PR #52 普通 Go Tests 报告

Hermes `t_a397c2ab` 承接[健康检查夹具修复后的两项报告失败](../health-fixture/README.md)。父提交为 `596f848`。原托管作业的单元和集成测试步骤都通过，失败集中在生成覆盖率和 JUnit 报告。[原作业日志](https://github.com/VanTam-CN/law-oa-go/actions/runs/45568119676/job/139047819429)

## 两项失败的证据

- **覆盖率。** 协调者在固定 Go 1.25.0 上重新执行旧工具安装命令，解析到 `golang.org/x/tools v0.53.0`，要求 Go 至少 1.26.0，退出 1，与托管日志一致。
- **JUnit。** 协调者和 6.2 Astra 各自取出原作业两个成功步骤的完整测试输出，用同版 `go-junit-report v2.1.0` 默认解析器回放。两份 XML 均含 0 个断言失败、1 个 SQLite 构建错误，转换器退出 1。仅在回放输入中去掉 SQLite 的五行编译警告，结果均为 0 个错误、退出 0。Astra 另以合成 C 编译警告区分了真实编译失败和警告；协调者也独立复现了警告导致文本解析误报，改用 Go JSON 事件后通过。这里没有在运行时过滤编译警告。
- **原报告的证据边界。** 在 Go 1.25.0 Linux/arm64、独立 PostgreSQL 15 与 Redis 7 中，父提交的原始无标签 `./...` 运行退出 0，文本转换也退出 0，产生 941 个测试条目。旧托管报告把两个程序接在管道中，没有留下各自的退出码；无法据此判定当时是否同时存在真正的失败用例。

## 候选和隔离验证

glm-flash 准备了工作流补丁。协调者在复核时补齐了合并命令，修正了工具构建方式和临时文件管理。最终[候选 71ebc94](https://github.com/VanTam-CN/law-oa-go/commit/71ebc94dbacd5cda18ddd44e8de2df43c23a3ef9) 只改两个既有步骤的命令。对解析后的工作流逐项比较，其他步骤、三个测试选择的范围及标签、触发条件、缓存及上传规则均保留。应用模块的依赖没有修改。

| 验证项 | 协调者复验结果 |
| --- | --- |
| 覆盖率计数 | 独立临时工具模块固定原 gocovmerge 提交及 x/tools v0.35.0。抽取实际步骤运行，重叠区块计数为 3、2、1；真实单元、集成 profile 合并后含 27,179 个区块，和独立参考工具输出逐字节相同。缺少输入时退出 1。 |
| JUnit 清单 | 候选继续执行第三遍无标签全库测试，使用 Go JSON 事件和 v2.1.0 转换器。隔离 Linux 上直接运行候选步骤，Go 与转换器均退出 0。按套件、用例名、结果和次数比对，与原报告的 941 个条目完全相同，含 912 个通过、29 个跳过。 |
| JUnit 错误传播 | 在固定 Go 1.25.0 下直接运行两个候选步骤。带 C 编译警告的正确测试退出 0；真实断言失败、真实编译失败都退出 1。用受控替身分别模拟 Go=7、转换器=0 和 Go=0、转换器=23，两种情况下步骤均退出 1。 |

## 独立审核

6.2 Astra 批准固定候选，在 Go 1.25.0 上独立执行候选脚本，得到相同的 3、2、1 合并计数。C 编译警告随成功用例一起通过；真实断言失败时 XML 记录一个失败；单独模拟 Go 失败和转换器失败时步骤均报红。Astra 同时核对了三组测试选择、报告的 always 条件以及缓存、上传规则。

## 实际 PR 验收

原审核分支 `codex/np-main-i11` 推送后，PR #52 的 GitHub 合成提交为 `989cfdab151cc972fe9b429ebf281a0d05aecdd4`，其父提交是主干 `d9c92fb` 与候选 `71ebc94`，Git 树与候选相同。[结构化记录](verification.json)

| 检查 | 终态及证据 |
| --- | --- |
| [PR 普通 Go Tests](https://github.com/VanTam-CN/law-oa-go/actions/runs/45574098703/job/139067615824) | **成功**。单元、集成、覆盖率合并、试点覆盖率摘要及 JUnit 全部成功；第三遍记录 Go=0、JUnit=0。G1 用例在单元和集成步骤各通过一次，监控回归也通过。缓存动作、Codecov 上传、JUnit 上传均跳过。 |
| [PR 完整 I11](https://github.com/VanTam-CN/law-oa-go/actions/runs/45574098703/job/139067615606) 与 [push 完整 I11](https://github.com/VanTam-CN/law-oa-go/actions/runs/45574092572/job/139067594652) | **均成功**。PR 绑定合成提交，push 绑定候选。各 17 个阶段退出 0、路线 B 四个场景通过，均输出 `I11_FULL_PASS`。 |
| [旧 PGO 支持检查](https://github.com/VanTam-CN/law-oa-go/actions/runs/45574098721/job/139067615772) | **失败**。其依赖的旧性能基准作业跳过。 |
| [全库 benchmark](https://github.com/VanTam-CN/law-oa-go/actions/runs/45574098806/job/139070733228) | **失败**。全库 benchmark 命令退出 1，标准输出重定向到未上传的文件，公开日志没有具名失败断言。 |
| [试点 PGO 采样](https://github.com/VanTam-CN/law-oa-go/actions/runs/45574098721/job/139067616163) | **失败**。公开日志定位到 `law-oa-go/tests/performance`，没有给出具体用例。后续双二进制构建和验证跳过。 |

Astra 独立抽核了目标 PR 的普通 Go Tests、完整 I11、合成提交及缓存和上传状态。协调者另行核实其余作业终态。**报告整改通过，Hermes `t_a397c2ab` 关闭**。基准夹具和失败归因继续由 `t_597f0f6a` 跟踪。合并、部署和真实试用保持 **HOLD**。
