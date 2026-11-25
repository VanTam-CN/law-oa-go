# 完整 I11 隔离验收

工程候选为 `52f9d40a274d04c1b59688d7cfd092c27863168b`。Hermes 执行卡为 `t_7072cca9`。M4 的独立准入记录在 [M4 证据分支](https://github.com/VanTam-CN/law-oa-go/tree/2989b285501f04fde37eb4826cb3f3a108be0915/docs/np-evidence/m4)。[结构证据](verification.json) 分列本机与托管的身份、时间、结果、PG 摘要和负向失败原因。

完整 I11 工程验收通过，可以启动 M5 隔离联调。托管正向运行成功，八个负向 job 均按预期原因失败，两个运行的 artifact 数量均为零。

## 核对口径

| 层次 | 必须取得的证据 |
| --- | --- |
| 后端与数据库 | 根目录构建成功。真实 Redis 7 中预置旧响应，获准、撤权和版本变化均不读取或写入业务缓存。PG 16.14 检查身份、v14 基线和 86 结构。Go 按包与用例名核对 run、pass 和包终态，常规组 32 项，带标签的 PG 组 11 项（含展开子项）。路线 B 的额外 Go 散列检查使用同一 PG 窗口内的浏览器基准。 |
| 前端 | 使用锁文件安装依赖，运行类型检查、12 套共 124 个具名 Jest 用例和生产式构建。历史 33 项逐项映射，committed 与 no_change 两项改为检查仅清除本请求的恢复键。 |
| Chrome 与材料保全 | 七套 M4/路线 B 用例 22 条，另有首次来源选择等 2 条和原生办理按钮 10 条，共 34 条。每条必须取得唯一一次通过事件，逐套拒绝缺项、跳过或重试后才成功。原生按钮使用 API 夹具，正式同步和异步回写另由 PG 用例验证。路线 B 四场景在独立种子窗口完成来源 +3、回执 +4，三条来源关联和一次手工保存。Go 与 Python 对同一受保护记录集的摘要一致。随机种子导致本机与托管摘要可不同。 |
| 正常和预期红灯 | 实际 GitHub 分支 push 运行完整正向。相同 checkout 的 workflow_dispatch 分别验证 PG 不可用、缺 DSN、关键 Skip、部分用例、零用例、未知 fuzz 类型、选中 job 失败及缺凭证。逐个比对预期步骤、日志标记和退出码。最后三项在托管 runner 上把受控输入送入原 fuzz 工作流的校验或汇总脚本。 |
| PR 路由 | 同仓 I11 分支面向 main/develop 的 PR 选择完整矩阵，既有测试 jobs 仍按试点规则隔离。六份工作流按源仓、分支、目标和运行状态进行 2016 次表达式比较。同名 fork、普通分支、后缀变体均有对照。真实 main 目标 PR 的 head/base/test-merge SHA 与事件侧验证另记 `t_ddd82725`。 |

[GitHub 正常运行](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379855010) · [GitHub 负向矩阵](https://github.com/VanTam-CN/law-oa-go/actions/runs/45379862576)

## 首轮故障与独立复核

- 已有合法来源版本时，旧 schema 负向夹具收窄列宽，PG 提前报 22001。修订将 text 临时改为不限制长度的 varchar，保留合法数据，再由真实 schema 门禁拒绝并回滚。
- 托管冷环境首次出现 Go 结构事件解析失败。标准输出与标准错误现分别写入隔离目录，进程退出码单独检查，原始日志和浏览器产物均不上传。
- Astra 的两项阻塞分别归入同一候选。I15-NATIVE 异步迟到检测已加入执行选择器和逐名门禁。独立审核对完整事件、移除该项、改为 Skip 三种输入分别得到成功、拒绝、拒绝。同仓 I11 PR 的副作用约束已纳入原试点条件。相同仓库的分支名大小写变体继续在 P2 `t_fbcffe5f` 跟踪。
- M5 的真实 root main、PG、Redis 和外置断线代理联调另行取证。M6 的独立集成复审、旧终端处置、升级和回滚演练仍待执行。I09-FILES 两版材料能力后置。

## 复跑入口

在本分支干净工作树使用 Node 22.17，按 frontend 锁文件执行 npm ci。提供 Docker、Chrome stable、psql 和 Go 模块依赖，然后从仓库根运行。执行器只对自己启动的回环合成 PG 操作，使用固定端口 18790，浏览器各套串行独占服务端口。

```sh
bash scripts/np-ci/early-pg.sh full
```

五种 PG 负向使用同一入口，参数分别为 missing-pg、missing-dsn、skip、partial、none。三个 fuzz 负向通过 `bash scripts/np-ci/i11-fuzz-probe.sh` 运行。每项都应单独退出 1。复跑 PR 路由比较时，将本目录的 pr-routing.mjs 复制到临时目录，按 M3 的既有方式在临时目录安装 @actions/expressions@0.3.62 与 yaml@2.9.1，从候选仓库根执行。
