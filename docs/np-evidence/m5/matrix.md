# M5 作者验收逐项索引

本页依据[冻结矩阵](../../plans/np-main-integration/implementation-and-acceptance-plan.md)分列直接观测、继承证据和待核对项。应用候选为 `49e5c15`。证据提交只增加验证材料。M6 总门禁见[作者结论](README.md)。

| 项 | 本轮结果和证据层级 |
| --- | --- |
| I01 | 继承 M0 的原生主干基线，含 Go 两包 138 RUN、132 PASS、6 SKIP 的边界。见 [M0 基线](../../plans/np-main-integration/m0/baseline-report.md)。 |
| I02 | 干净检出的普通构建、带标签夹具构建通过。运行的根二进制校验 VCS 与 SHA-256，生产前端重建七个文件与实测资产一致。[构建记录](environment-and-browser.json) |
| I03 | 获准律师的同一令牌在改角色及停职后，读取已开上下文、新建上下文均 403；恢复后重新可用。原样登录令牌临时失去 PG 登记时，原生详情 200、助手创建 401。Redis 中已核对登录登记与退出失效；另两枚正式登录令牌在撤销或 PG 登记到期后，创建上下文均返回 401。[注册及角色](session-boundaries.json) |
| I04 | 根服务本人详情 200、他人草稿 403、业务缺失 404 已观测。向正式入口写入两个合成身份，核实 PG 各有非空密文和摘要，两份完整详情均不含这六个值。在上下文已生成候选后进行客户隔离墙授权再撤权。别名详情、已有上下文和带来源的严格保存均 403，PG 草稿整行、来源与回执数量不变。[身份脱敏](session-boundaries.json)[缓存读路径](cache-paths.json) |
| I05 | 根服务列表、详情、工作台别名与 PG 版本相同；双页并发保存陈旧方 409 且页面保留输入。原生与来源写入争用另有 PG 用例。[同版本重开](reopened.json) · [四场景](route-b.json) · [并发服务用例](pg-service-checks.json) |
| I06 | 来源、回执和指针由事务产生。PG 对照含大整数与未知字段，原生创建和更新的伪造指针由具名后端回归验证。[数据库对照](route-b.json) · [后端用例](browser-controls.json) |
| I07 | 浏览器实走两轮整理、核对来源、人工修改与晚到生成，组件用例覆盖到期与跨对象响应。[四场景](route-b.json) · [组件结果](browser-controls.json) |
| I08 | 外置代理在转发前、PG 真写者等锁期间及提交后断线，随后逐键核对无记录、放弃或已提交的结果。三项已获 Astra 证据复核。[断线窗口](disconnect.json) |
| I09 | 路线 B 的 I09-MAIN 通过。来源 +3、回执 +4，两条材料与其他受保护行前后散列相同。I09-FILES 两版材料后置。[PG 快照](route-b.json) |
| I10 | 真根服务在同一 PG/Redis 上运行；九条敏感路径有旧值时均返回实时结果及禁止缓存响应头。各请求至 ECHO 标记的窗口内，目标键及三个业务前缀的匹配命令为零；邻路有一次实际 GET。源码在响应捕获前跳过这些路径。隔离墙显式授权后撤权，随后读取别名、上下文及带来源保存均被拒绝；认证 Redis 仍处理登录及退出。[逐路观察](cache-paths.json) |
| I11 | 继承冻结分支的托管正常与负向运行及独立复审，M5 另行运行组件、浏览器和服务集成。[I11 原运行](../m2/full-i11/README.md) |
| I12 | M6 演练待执行。I14 的旧库对照只用于验证启动条件。 |
| I13 | 组件覆盖严格分流、原生入口和非摘要路径；浏览器覆盖混改标题。获准律师变更角色后真实上下文被拒。另一个在职律师可打开本人草稿，助手能力关闭，创建上下文返回 403。[角色变化](session-boundaries.json)[组件与浏览器](browser-controls.json) |
| I14 | 未迁移旧库用旧根服务成功读写，同库启动候选按 86 门禁拒绝。已迁移的专用克隆内，来源指针已清但仍存三条来源、四条回执。开关开和白名单清空时旧协议均 409。来源表临时不可见以及表在但查询列不可用时，旧协议和严格保存均 503，恢复后整行和证据摘要相同。[克隆库故障验证](history-faults.json)锁内竞争有真实 PG 回归。[未迁移对照](unmigrated.json) · [竞争回归](pg-service-checks.json) |
| I15 | 正式异步服务及两类写入后的过时完成已有 PG 回归。正式根入口已确认同步检测等待 PG 写入，再完成原生修改和带来源的严格保存。放行后旧结果 409，关联案件及接案整行与保存后相同。[来源写后的迟到检测](source-conflict.json) 原生单独改事实后迟到回写同样 409。匹配输入生成一条记录并推进版本，重复提交旧版本 409，记录数和两表整行保持不变。[原生](conflict-native.json) · [正常完成](conflict-matching.json) · [来源与回执整行复核](conflict-source.json)。[服务级证据](pg-service-checks.json) |
| I16 | 回查、放弃和提交都按原对象、原键判定；错误身份、缺字段和非法版本有逐名组件断言。同源双页有真实浏览器记录。[服务回执](disconnect.json) · [组件与双页](browser-controls.json) |

| 项 | 本轮结果和证据层级 |
| --- | --- |
| C01 | 浏览器以两种登录持久方式、助手开与关覆盖四种组合，检查存储和 PG。[存储窗口](environment-and-browser.json) |
| C02 | 新登录后从原菜单重开保存版本及来源，最终手工版本可重复读取。[来源重开](route-b.json) · [原生重开](reopened.json) |
| C03 | 浏览器覆盖后退取消、换草稿、步骤切换及令牌过期后清理在页内容。[导航](browser-controls.json) |
| C04 | 转发前断线，真实 GET unknown 时 PG 不变；提交后新登录按原键回查。失去冻结体后的处理另由组件核对。[断线](disconnect.json) · [组件](browser-controls.json) |
| C05 | 状态码白名单、输入保留及未知状态不清键按冻结组件清单核对，PG 校验异参 409。[组件](browser-controls.json) · [异参回放](disconnect.json) |
| C06 | 禁用存储后浏览器阻止写请求，PG 整行与回执计数不变。[存储窗口](environment-and-browser.json) |
| C07 | 同源同一浏览器上下文的两个页面交错保存。一个终态保留另一个未决标识。[双页窗口](browser-controls.json) |
| C08 | 原律师退出、换律师登录、原律师重新登录后逐笔回查。[换账号窗口](browser-controls.json) |
| C09 | 合成旧标签有仅内存文字和预填，实测先核对、再关页、定向清理；真实终端处置归 M6。[旧标签窗口](browser-controls.json) |
| C10 | Redis 九条敏感读路径、实际生产资产以及受控旧 worker 的证据分别记录；旧 worker 接管后显示 HOLD，敏感请求为零。[逐路缓存](cache-paths.json) · [生产资产与旧 worker](environment-and-browser.json) |
| C11 | 原键格式、逐请求去重和仅清自身终态由组件覆盖；旧标签停写顺序另有浏览器证据。[组件及旧标签](browser-controls.json) |
| C12 | 只改摘要的四场景完成后，PG 核对来源、回执和既有材料；助手关闭的存储行为有独立浏览器窗口。[四场景](route-b.json) · [存储](environment-and-browser.json) |
| C13 | 保存并退出及回查等待期间继续输入 B，实际浏览器保留 B，服务端重开只显示 A。[编辑窗口](browser-controls.json) |
| C14 | 重载前取消则不取数；确认后输入 C，迟到结果保留 C。[编辑窗口](browser-controls.json) |

[冻结浏览器用例名](../../../scripts/np-ci/i11-browser-cases.json) · [冻结组件用例名](../../../scripts/np-ci/i11-jest-cases.json)
