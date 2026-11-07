# C01–C18 acceptance matrix（进行中）

本文件为进行中的证据索引，尚未冻结。初版只整理已有证据；本次按授权增补 B2/R1 实际作者与隔离 PG 结果，没有更新 tasks。最终 C01–C18 判定全部留给主线程填写。

初版索引时主线程报告：受影响测试 178 项通过、transactions 34 项通过；当时 B1/B2/B3、前端 F1–F4 仍在修复，C16/C17 未完成。这些计数是各自运行记录，不是 C01–C18 的最终结论。178 项原始日志尚未在本文件逐条核验；34 项链接到现存事务结果，修复后仍需绑定新候选。

B2/R1增量证据：[B2真实重提与待办报告](inbox/b2/REPORT.md)、[R1换客户未重提报告](inbox/r1/REPORT.md)。二者是不同状态的回归；本次没有替主线程勾选任何条件。其他行保留初版证据时点，不据其旧修复状态覆盖后续交接；bootstrap parity 与最终共同哈希由主线程/Kuhn 汇合。

author=实施者自测；independent=独立审查/独立执行（DOM mock不算browser）；browser=真实浏览器操作；realPG=真实隔离PostgreSQL（不等于生产）；notrun=缺少对应运行证据。各维度可同时成立，不能互相替代。

主方案条件逐字摘录自 /private/tmp/lawoa-intake-collaboration-20261107/docs/plans/2026-09-22-intake-collaboration-mvp-glm.md 第11节：

| 编号 | 场景（原文） | 必须观察到的结果（原文） | 当前证据及类型 | 剩余项 | 冻结判定 |
|---|---|---|---|---|---|
| C01 | 新客户原地创建、取消、失败 | 成功回填；当前立案全部字段保留；取消无写入；错误可恢复 | author: [前端作者记录](frontend/README.md)；browser: 未完成 | F1–F4及弹窗修复后重验 | 待主线程填写 |
| C02 | 双击、响应丢失重试、同 key 改请求 | 同请求一个客户；结果可回读；冲突请求 409；无身份明文日志 | author + realPG: [客户幂等原始日志](client/03-behavior-isolated-postgres.log)；browser: NOT_RUN/未闭环 | 响应丢失/同key改请求的真实页面链仍待补 | 待主线程填写 |
| C03 | 超过 50 个客户；同名不同身份 | 后页客户能选；新客户不用等列表刷新；无按名称自动合并 | author + realPG: [客户分页与身份日志](client/03-behavior-isolated-postgres.log)；browser: 未闭环 | 后页选择与新客户回填页面证据待补 | 待主线程填写 |
| C04 | 退出重登、换账号、两个窗口保存 | 正确草稿恢复；不串账号；旧 revision 拒绝且本地输入保留 | author: [前端作者记录](frontend/README.md)；independent: [F3/F4独立审查](review/frontend-review.md) | F3/F4修复中，恢复迟到响应与新入口需重验 | 待主线程填写 |
| C05 | 律师/核查/审批三类待办直达 | 正确对象与动作；不能以 uint 猜 UUID；状态/计数一致 | author + realPG: [B2真实重提与待办报告](inbox/b2/REPORT.md)、[R1换客户未重提报告](inbox/r1/REPORT.md)；independent: [独立复审：B2已核查、发现R1](review/backend-final-recheck.md) | B2与R1作者限定验证已补；R1独立复核、三类角色最终候选及browser仍由主线程冻结 | 待主线程填写 |
| C06 | 无权限、撤销权限、转交后旧待办、伪造完成模式/关联来源 | 服务端拒绝；不泄漏受限标题/身份/意见；新责任人任务正确；不能自建待办取得对象权限 | author + realPG: [R1换客户未重提报告](inbox/r1/REPORT.md)（当前/冻结客户分别撤权与恢复，List/Stats/GET/home同源）；author: [既有Inbox负向](inbox/01-inbox-local-tests.jsonl)；independent: [独立复审：B2已核查、发现R1](review/backend-final-recheck.md) | R1已补双侧授权证据，不能代替全部角色/历史读取/浏览器冻结；B3以Kuhn交接及独立复审为准 | 待主线程填写 |
| C07 | 手工完成/删除 workflow 项、延后 | 不能让业务任务消失；延后不改变办理期限；普通提醒可正常完成 | author: [Inbox日志](inbox/01-inbox-local-tests.jsonl) | 本地用例证据存在，最终候选及browser未确认 | 待主线程填写 |
| C08 | 退回两项、逐项保存、刷新 | 同一退回记录、稳定 item_id、完整意见、正确定位与持久化回复 | author + realPG: [快照作者记录](snapshot/RESULT.md)；author DOM: [前端作者记录](frontend/README.md)；author + realPG: [R1换客户未重提报告](inbox/r1/REPORT.md)（补充入口中间状态） | 逐项保存刷新真实页面链未闭合 | 待主线程填写 |
| C09 | 修改对方身份/角色后直接重提 | 旧检测被拒；新检测与独立复核完成后才可推进；旧快照不改 | author + realPG: [事务逐项结果](transactions/52-final-transition-results-summary.json)、[B2真实重提与待办报告](inbox/b2/REPORT.md)、[R1换客户未重提报告](inbox/r1/REPORT.md)；independent: [独立复审：B2已核查、发现R1](review/backend-final-recheck.md) | B2重提后及R1尚未重提的可见性分别验证；检测/独立复核为fixture输入，真实页面主链及冻结待主线程 | 待主线程填写 |
| C10 | 仅回复说明、未知字段、伪造清单 ID | 有效不变事实按服务端判断复用；未知/伪造项拒绝；不允许自报已核验 | author + realPG: [快照作者记录](snapshot/RESULT.md) | 最终候选与真实页面复验待填 | 待主线程填写 |
| C11 | 旧轮次页面迟到决定、重复重提、通用接口覆盖快照/检测绑定 | 旧版本 409；相同请求幂等；不同请求不能复活旧轮次；备用入口不能绕过统一版本门禁 | author + realPG: [事务逐项结果](transactions/52-final-transition-results-summary.json)；independent: [B1审查](review/backend-final-review.md) | B1备用入口修复中，已有测试不能覆盖全部入口 | 待主线程填写 |
| C12 | 原审批重提与新审批争同接案 | 同起跑并发至少 3 轮及两种有序交错各 1 次；至多一个有效持槽审批、一个案件 | author + realPG: [事务逐项结果](transactions/52-final-transition-results-summary.json)；independent: [B1审查](review/backend-final-review.md) | B1修复后补备用入口竞争；不可写最终通过 | 待主线程填写 |
| C13 | 退回/重提中待办写入故障 | 业务与交接一起回滚；清除故障后原请求可重试，待办不重复 | author + realPG: [事务原始日志](transactions/49-final-transition-race.jsonl) | 原事务故障证据仍按候选绑定复核；B2/R1后读取回归见新报告，不替代故障注入验收 | 待主线程填写 |
| C14 | 快照写入/成案主体写入故障及重试 | 无假 completed、无部分案件；恢复后唯一完整案件且主体与本轮批准快照一致 | author + realPG: [事务原始日志](transactions/49-final-transition-race.jsonl)；static: [contract-map缺口](contract-map.md) | 成案失败状态来源及多级触发风险待修/复验 | 待主线程填写 |
| C15 | 历史待办、旧文字退回、多版本快照 | 不猜对象；兼容来源明确；原历史字节内容不覆盖；来源冲突单列阻断 | author: [快照记录](snapshot/RESULT.md)；independent: [前端F1/F2](review/frontend-review.md)、[后端B3](review/backend-final-review.md) | F1/F2/B3修复中，历史兼容/授权未闭合 | 待主线程填写 |
| C16 | 桌面完整主链与第二次独立样本 | 同一对象链从新客户到成案；第二样本独立重登完成；不能拼接不同成功片段 | browser: 主线程进行中，完整两样本未完成；independent: [前端审查边界](review/frontend-review.md) | 不得拼接不同对象成功片段；等待同一候选两条完整链 | 待主线程填写 |
| C17 | 390×844、键盘和长补充清单 | 菜单可达、弹窗按钮不被遮挡、字段定位与焦点可用；仅窄视口证据 | browser: NOT_RUN/未完成；author DOM: [前端记录](frontend/README.md) | 390×844、键盘、长清单真实视口证据待补，DOM不能替代 | 待主线程填写 |
| C18 | 手册对照 | 按手册无需开发者解释可找到入口；截图标题/候选一致，未开放能力如实标明 | notrun: 未确认最终候选对应手册/截图验收记录 | 手册/截图与冻结候选一致性由主线程最终填写 | 待主线程填写 |

C12–C14必须有真实隔离PG事务证据；SQLite/mock不能代替。完整角色负向、共同委托/多对方/同名异身份/同主体多角色的同对象成案链仍按主方案验收。既有报告可能保留旧状态，读取时以运行时间/源码哈希及后续修复为准，不修改原证据。
