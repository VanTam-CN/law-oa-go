# M5 隔离作者集成

M5 作者已完成冻结矩阵的适用隔离检查，M6 准入待固定提交的独立证据复审。逐项列出的检查层级与继承范围见[验收矩阵](matrix.md)。任务记录在 Hermes 的 `t_5b8b79ec`。本轮使用合成数据，所有测试只访问回环端口和专用容器。

## 候选与运行环境

以 I11 归档 `89f2eb9` 为开发基线。M5 应用二进制从干净检出 `49e5c15` 的根包构建，内嵌提交和运行容器内外的文件散列均已核对。PG 16.14 使用 v14 加受控 86 迁移，根服务以生产配置连接 PG 和 Redis。前端复用 M4 的七个生产资产，并从同一 M5 干净检出重建。全部文件的路径与内容散列均与实际提供给浏览器的资产一致。[构建、环境和前端摘要](environment-and-browser.json)

首轮从嵌套工作树构建时，Go 写入了外层仓库的提交号。该轮仅用于诊断。重新构建后，浏览器在真实服务上重新运行。完整 I11 的托管结果继续按原始运行 SHA 引用。[完整 I11](../m2/full-i11/README.md)

## 当前证据与待核对项

| 冻结项 | 作者取证情况 |
| --- | --- |
| I01、I02、I11 | 保留 M0 主干回归和 I11 的独立托管记录。本轮在干净应用检出上另行运行普通 Go 构建和带标签入口构建，业务访问统一指向根服务。 |
| I03、I04、I10 与 C10 服务端 | 已向九条敏感读路径的精确 Redis 键预置旧值。逐条请求均返回 private, no-store。在每个请求至 Redis ECHO 标记的窗口内，目标键和三个业务前缀的操作计数均为零。相邻业务路径实际读键一次。源码确认这三类路径在响应捕获及异步写入调度之前跳过缓存，窗口计数仅支持该时间段的结论。[缓存分支](../../../internal/middleware/cache.go) 来源及已提交回执与 PG 身份一致。M5 实测角色变化、停职、退出、撤销、到期和合法登录令牌失去登记，助手入口均拒绝。隔离墙撤权后，原生详情、已打开的上下文和已核对的来源保存均重新鉴权。[会话与客户授权](session-boundaries.json)[逐路缓存](cache-paths.json) · [根服务日志抽样](root-log-sample.json) |
| I05、I06、I07、I09-MAIN | 路线 B 四个具名场景在同一干净窗口依次通过。原菜单可重开来源，陈旧摘要被拒并保留输入，最终手工版本解绑。来源增加 3、回执增加 4，PG 受保护记录集的 Go 与 Python 散列一致。最终手工保存后又用同一律师重新登录，列表、详情、工作台别名与 PG 的 revision 相同，读取前后整行及受保护记录不变。[同版本重开](reopened.json)。原生旧写入和前置权限等子项仍按各自证据判定。[路线 B](route-b.json) |
| I08、I16 | 外置代理分别在转发前和根服务成功响应后断开。另一次先确认真实 PG 写者等待草稿行锁，再断开 TCP。按同一对象、操作者和请求键散列核对回执，验证同键重放、异参拒绝与放弃墓碑。I16 错身份、缺字段和非法版本由具名组件用例覆盖。[断线与回执](disconnect.json) · [组件及同源浏览器](browser-controls.json) |
| I13、I14、I15 | I13 前端分流具名回归已运行。在不含 86 的独立 v14 库中，旧根服务完成原生创建、保存与重开；同库启动 M5 根服务按预期因 86 结构不符退出。[未迁移对照](unmigrated.json) 在已迁移克隆库中实测关闭助手、临时隐藏来源表、触发历史查询失败。旧全量写始终受保护，PG 接案、来源、回执、当事人和材料整行摘要不变。I15-NP 的正式根入口返回迟到 409，关联案件及草稿整行未变。[历史保护](history-faults.json) · [同步迟到](source-conflict.json)。原生单独改事实后，正式迟到检测也返回 409。匹配输入正常完成一次，关联案件与草稿引用同一结果，原版本再提交时返回 409 且无新增记录。[原生迟到](conflict-native.json) · [正常完成](conflict-matching.json) · [来源和回执整行核对](conflict-source.json)。[PG 服务及回写](pg-service-checks.json) |
| C01、C02、C03、C06、C07、C10-PROD、C10-OLD | 实际浏览器覆盖两种登录持久模式、禁止存储、重新登录、未保存导航、同源双页、生产资产及受控旧 worker。旧 worker 成功接管时显示 HOLD，敏感请求计数为零。[浏览器窗口](environment-and-browser.json) · [同源与导航](browser-controls.json) |
| C04、C05、C08、C09、C11–C14 | 请求身份及恢复键的 124 个前端具名用例已按冻结名单核对；各项按对应的组件、浏览器和数据库证据判定。断线的两侧另有 I08 真实 TCP 观察。 |
| I09-FILES、I12 | 两版文件保全保持 DEFERRED/HOLD。I12 合成升级、回滚与单实例演练归 M6。 |

[冻结验收矩阵](../../plans/np-main-integration/implementation-and-acceptance-plan.md)

## 复核方法

在隔离开发工作树中使用私有 `np-m5.*` 临时目录。将目录绝对路径导出为 `NP_M5_WORK`，运行 `bash scripts/np-m5/environment.sh init` 建立合成库、Redis 和干净构建。在当前 shell 加载 `"$NP_M5_WORK/env"`，再用既有 `compare-main-browser-db.py seed` 分别生成试点律师与对照律师的清单。将试点律师 ID 设为 `NP_ASSISTANT_ACTORS` 后运行 `bash scripts/np-m5/environment.sh root`。实际运行前校验生产资产摘要与本文引用的 M4 记录一致。

浏览器执行器 `bash scripts/np-m5/browser.sh <suite> <新的私有结果目录>` 读取 `NP_BROWSER_SEED` 和 `NP_BROWSER_OTHER_SEED`。逐名运行保持生产限流，路线 B 四条使用同一批目标和对照。检查器拒绝缺项、重复、失败、意外跳过与重试后才通过的场景。C10 的受控旧 worker 使用专属静态主机。首次来源选择用例中的确认/检测按钮使用前端夹具，其归属为界面行为。

I08 独立窗口使用只服务于本机的 3004 端口代理，设置 `ROOT_URL=http://127.0.0.1:18789`、`NP_M5_PROXY_PORT=3004` 及新的私有 `NP_M5_PROXY_EVENT_FILE`。针对干净合成种子运行 `/usr/bin/python3 scripts/np-m5/disconnect.py`。原始浏览器报告、PG 内容、认证信息及应用日志留在私有窗口。结束时 `environment.sh clean` 按本次标签清理容器和网络。


## 新建接案的既有问题

隔离准备期间，两次独立新建落在同一秒，其中一次返回 500，根服务记录 PG 唯一键错误 23505。旧主干 `d9c92fb` 和当前 `49e5c15` 均使用精确到秒的 `intake_code`。本阶段操作对象是已有草稿。三种 I15 验证独立创建合成对象并错开准备时间，完整保留这次失败。原生新建风险已登记 Hermes `t_a4646b57`，由后续试用决策处置。

## 补充窗口的复核入口

用独立种子窗口完成四场景后，可用该窗口的两个清单作为 `NP_BROWSER_SEED`、`NP_BROWSER_OTHER_SEED`。运行 `reopened-draft.py` 与 `cache-paths.py` 检查原对象。以显式试点律师名单启动 `history-window.sh init` 和 `history-window.sh root`，随后运行 `history-faults.py`、`session-boundaries.py`。它们分别核对专用克隆中的历史故障和客户撤权，结束后用 `history-window.sh clean` 清理专用实例。I14 旧库对照单独运行 `legacy-baseline.sh`。

I15 的 `source-conflict.py` 读取已授权的合成律师清单，每个模式创建自己的合成对象。分别以 `native`、`source`、`matching` 运行，准备新对象的时间至少错开两秒。同步检测来自真正的根服务；正式异步完成使用同候选的服务级 PG 回归。此宿主的复跑命令使用 `/usr/bin/python3`。第二组访问探针的运行次数和归档边界见 [glm-flash 对照](worker-checks.json)。


## 独立复审补跑

在首次总复审 `60ae91b` 之后，补跑了两项。第一项用同一令牌读取已保存的历史来源与已提交回执，撤权前两次 200；将先前的授权投影预置在合成 Redis 中，再次访问原地址，两次均 403、private/no-store，整份拒绝响应无此前摘要及来源原句，恢复授权后读取结果与原投影相同。[撤权后重读](revoked-history.json)

第二项沿用生产前端与真实登录、草稿、上下文入口，仅在浏览器边界为生成请求返回一次合成 503。页面显示生成失败后仍能人工编辑标题与摘要。存储写入观察与两个存储区的最终读取均未发现业务正文，PG 快照与原存储窗口一致。[生成失败](generation-failure.json) · [逐项矩阵](matrix.md)
