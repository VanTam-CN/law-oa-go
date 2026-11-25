# M5 隔离作者集成

当前判定为部分完成。M6 准入为 HOLD。任务记录在 Hermes 的 `t_5b8b79ec`。本轮使用合成数据，所有测试只访问回环端口和专用容器。

## 候选与运行环境

以 I11 归档 `89f2eb9` 为开发基线。M5 应用二进制从干净检出 `49e5c15` 的根包构建，内嵌提交和运行容器内外的文件散列均已核对。PG 16.14 使用 v14 加受控 86 迁移，根服务以生产配置连接 PG 和 Redis。前端复用 M4 的七个生产资产，并从同一 M5 干净检出重建。全部文件的路径与内容散列均与实际提供给浏览器的资产一致。[构建、环境和前端摘要](environment-and-browser.json)

首轮从嵌套工作树构建时，Go 写入了外层仓库的提交号。该轮仅用于诊断。重新构建后，浏览器在真实服务上重新运行。完整 I11 的托管结果继续按原始运行 SHA 引用。[完整 I11](../m2/full-i11/README.md)

## 当前证据与待核对项

| 冻结项 | 作者取证情况 |
| --- | --- |
| I01、I02、I11 | 保留 M0 主干回归和 I11 的独立托管记录。本轮在干净应用检出上另行运行普通 Go 构建和带标签入口构建，业务访问统一指向根服务。 |
| I03、I04、I10 与 C10 服务端 | 已向九条敏感读路径的精确 Redis 键预置旧值。逐条请求均返回 private, no-store。在每个请求至 Redis ECHO 标记的窗口内，目标键和三个业务前缀的操作计数均为零。相邻业务路径实际读键一次。源码确认这三类路径在响应捕获及异步写入调度之前跳过缓存，窗口计数仅支持该时间段的结论。[缓存分支](../../../internal/middleware/cache.go) 来源及已提交回执与 PG 身份一致。角色、在职状态、撤权和兼容令牌另行核对。[逐路缓存](cache-paths.json) · [根服务日志抽样](root-log-sample.json) |
| I05、I06、I07、I09-MAIN | 路线 B 四个具名场景在同一干净窗口依次通过。原菜单可重开来源，陈旧摘要被拒并保留输入，最终手工版本解绑。来源增加 3、回执增加 4，PG 受保护记录集的 Go 与 Python 散列一致。最终手工保存后又用同一律师重新登录，列表、详情、工作台别名与 PG 的 revision 相同，读取前后整行及受保护记录不变。[同版本重开](reopened.json)。原生旧写入和前置权限等子项仍按各自证据判定。[路线 B](route-b.json) |
| I08、I16 | 外置代理分别在转发前和根服务成功响应后断开。另一次先确认真实 PG 写者等待草稿行锁，再断开 TCP。按同一对象、操作者和请求键散列核对回执，验证同键重放、异参拒绝与放弃墓碑。I16 错身份、缺字段和非法版本由具名组件用例覆盖。[断线与回执](disconnect.json) · [组件及同源浏览器](browser-controls.json) |
| I13、I14、I15 | I13 前端分流具名回归已运行。在不含 86 的独立 v14 库中，旧根服务完成原生创建、保存与重开；同库启动 M5 根服务按预期因 86 结构不符退出。[未迁移对照](unmigrated.json) I14 原生并发与 I15 正式检测器已有 PG 服务层回归；实际根入口和权限变化后的写入保护另行取证。[PG 服务及回写](pg-service-checks.json) |
| C01、C02、C03、C06、C07、C10-PROD、C10-OLD | 实际浏览器覆盖两种登录持久模式、禁止存储、重新登录、未保存导航、同源双页、生产资产及受控旧 worker。旧 worker 成功接管时显示 HOLD，敏感请求计数为零。[浏览器窗口](environment-and-browser.json) · [同源与导航](browser-controls.json) |
| C04、C05、C08、C09、C11–C14 | 请求身份及恢复键的 124 个前端具名用例已按冻结名单核对；各项按对应的组件、浏览器和数据库证据判定。断线的两侧另有 I08 真实 TCP 观察。 |
| I09-FILES、I12 | 两版文件保全保持 DEFERRED/HOLD。I12 合成升级、回滚与单实例演练归 M6。 |

[冻结验收矩阵](../../plans/np-main-integration/implementation-and-acceptance-plan.md)

## 复核方法

在隔离开发工作树中使用私有 `np-m5.*` 临时目录。将目录绝对路径导出为 `NP_M5_WORK`，运行 `bash scripts/np-m5/environment.sh init` 建立合成库、Redis 和干净构建。在当前 shell 加载 `"$NP_M5_WORK/env"`，再用既有 `compare-main-browser-db.py seed` 分别生成试点律师与对照律师的清单。将试点律师 ID 设为 `NP_ASSISTANT_ACTORS` 后运行 `bash scripts/np-m5/environment.sh root`。实际运行前校验生产资产摘要与本文引用的 M4 记录一致。

浏览器执行器 `bash scripts/np-m5/browser.sh <suite> <新的私有结果目录>` 读取 `NP_BROWSER_SEED` 和 `NP_BROWSER_OTHER_SEED`。逐名运行保持生产限流，路线 B 四条使用同一批目标和对照。检查器拒绝缺项、重复、失败、意外跳过与重试后才通过的场景。C10 的受控旧 worker 使用专属静态主机。首次来源选择用例中的确认/检测按钮使用前端夹具，其归属为界面行为。

I08 独立窗口使用只服务于本机的 3004 端口代理，设置 `ROOT_URL=http://127.0.0.1:18789`、`NP_M5_PROXY_PORT=3004` 及新的私有 `NP_M5_PROXY_EVENT_FILE`。针对干净合成种子运行 `/usr/bin/python3 scripts/np-m5/disconnect.py`。原始浏览器报告、PG 内容、认证信息及应用日志留在私有窗口。结束时 `environment.sh clean` 按本次标签清理容器和网络。
