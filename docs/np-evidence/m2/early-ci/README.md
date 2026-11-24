# M2 早期 CI 验收

开发基线为 `a0acf35`，工程提交为 `9e94e0a` 和 `fa4a790`。本次验收使用开发分支的 push 和手动工作流。完整 I11 依赖 M3/M4 的测试夹具，保持 `NOT_RUN`。

## 判定规则

| 情形 | 判定依据 |
| --- | --- |
| 正常 PG | 在仓库根目录执行 `go build ./...`。PG 16.14 启动后核实库、用户和版本，bootstrap 并校验 000086。并发保存、schema 主测试及三个 schema 子用例必须分别有 run 和 pass，包级退出成功。 |
| 缺 PG | 在独立 runner 省略临时容器，保持相同连接和身份检查。预期在初始化前返回 `M2_PG_UNAVAILABLE`。 |
| 跳过关键测试 | 用临时 Go overlay 让并发保存用例实际执行 `t.Skip`，其余 schema 用例照常执行。即使 Go 返回 0，门禁仍须返回 `M2_REQUIRED_TEST_SKIPPED`。 |
| 没有用例 | 只更换 Go 测试名称筛选器，必需用例清单保持固定。Go 正常退出时，门禁仍须返回 `M2_NO_TESTS`。 |

同一工程 SHA 的 [正常运行](https://github.com/VanTam-CN/law-oa-go/actions/runs/45160569000)、[三个预期失败的 job](https://github.com/VanTam-CN/law-oa-go/actions/runs/45160694603) 和 [validators 手动运行](https://github.com/VanTam-CN/law-oa-go/actions/runs/45160571800) 可分别核查。负向运行保留红色结论。每个失败须在指定步骤产生对应诊断码，安装、编译或调度故障不能算作预期红灯。

## fuzzing 核对

旧配置在同一开发分支产生过 0 个 job。用 actionlint 检出六个 job 对不可用上下文的引用。修复后从参数校验 job 输出选择值，空输入选 all，未知类型退出失败。汇总逐项检查被选中的 job 结果及执行凭证。每个矩阵分片只有在 Go 输出中找到指定目标的 PASS 后才能提供凭证。运行器在步骤摘要记录目标名，原始输出保存在临时目录。

远端选择 validators、时长 1 秒、并行度 1，覆盖三个验证器目标。其余类别的路由、缺省 all、未知类型、选中 job 失败或跳过、凭证缺失均在本地核对。详细退出码、SHA、时间和 job ID 见 [结构证据](verification.json)。

## 复跑边界

在该分支、空闲的本机 18790 端口上逐个运行 `bash scripts/np-ci/early-pg.sh <normal|missing-pg|skip|none>`。三个负向命令均应退出 1。脚本仅管理自己创建的临时容器，每次新建合成数据库。推送开发分支默认执行正常门禁，手动触发同分支的 Test Pipeline 并选择 negative 时三个 job 独立执行。

早期范围覆盖开发分支和上述手动事件。PR 触发的全链副作用隔离、前端与浏览器夹具、完整 I11、合并及部署均未执行。
