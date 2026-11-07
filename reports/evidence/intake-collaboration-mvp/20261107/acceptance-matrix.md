# C01–C18 最终工程验收矩阵

运行源码 `3f970040f5e40b8016c18a070b5aa1345d2b2eec`，固定R4前端，本地隔离PostgreSQL和全虚构资料。以下PASS仅为主方案定义的工程检查；作者、独立审查者与浏览器证据分别保留，真人独立使用及生产均NOT_RUN。

[115个子条件](final-subconditions.csv)保留原稿状态及最终补证；原稿中7c的B5首红不改写为通过。[源码继承范围](candidate-binding.json)、[总报告](final-report.md)及[缺陷关闭索引](defects.md)共同限定结论。

| 编号 | 场景 | 结果 | 观察与边界 | 主证据 |
|---|---|---|---|---|
| C01 | 原地创建/取消/故障恢复 | PASS | 30输入及13选择保持；实际恢复POST200 | [证据](browser/r5-all-fields-preserved.json) |
| C02 | 双击、丢响应、幂等冲突 | PASS | 一个客户；原请求同ID，改请求409；PG日志无被测身份原文 | [证据](browser/client-response-loss.json) |
| C03 | 后页客户与同名身份 | PASS | 65个客户、第二页选择；同名不同身份独立对象 | [证据](browser/client-paging-mobile.json) |
| C04 | 重登/换账号/旧版本 | PASS | 恢复指定草稿；他账号不可见；双窗口旧revision409且输入保留 | [证据](browser/final-a-events.json) |
| C05 | 三类角色工作交接 | PASS | 律师补充、业务审批、核查岗实际直达办理；列表/计数/详情同源 | [证据](browser/r5-officer-workflow.json) |
| C06 | 对象授权与伪造入口 | PASS | 代表性六角色；撤权及恢复、转交、来源/模式伪造直接API拒绝；C06额外2PG独立复跑 | [证据](inbox/c06/HANDOFF.md) |
| C07 | 工作流任务完成与延后 | PASS | 手工完成/删除拒绝；延后不改业务；普通提醒保留原行为 | [证据](final-subconditions.csv) |
| C08 | 结构化退回与稳定回复 | PASS | 两项ID/退回记录准确配对；保存刷新不丢、键盘定位正确 | [证据](browser/r5-final-db.json) |
| C09 | 事实变化重新核查 | PASS | 旧检测重提409；A重新检测及独立复核后成功，旧快照不覆盖 | [证据](browser/r5-chain-a.json) |
| C10 | 仅说明复用及伪造清单 | PASS | B不变事实复用原检测/review7；未知/伪造项PG拒绝 | [证据](browser/r5-chain-b.json) |
| C11 | 迟到决定/重放/备用入口 | PASS | 旧版本409；幂等同案；冻结与绑定不能被通用入口覆盖 | [证据](bootstrap/HANDOFF.md) |
| C12 | 并发唯一占槽 | PASS | 真实PG同起跑三轮及双向有序交错；最多一个有效占槽/案件 | [证据](final-subconditions.csv) |
| C13 | 待办事务故障恢复 | PASS | 真实PG故障回滚业务及交接；去故障原请求重试且不重复 | [证据](bootstrap/HANDOFF.md) |
| C14 | 冻结、成案、角色完整性 | PASS | 真实PG写入故障及恢复；中间不成案；A复杂主体逐项匹配，案件唯一 | [证据](browser/r5-final-consistency.json) |
| C15 | 历史兼容与不可改写 | PASS | 旧文字和来源明确；冲突阻断；历史字节和撤权访问保护 | [证据](final-subconditions.csv) |
| C16 | 两套独立完整页面主链 | PASS | 3f同源前后端，A/B均新客户到案件；B独立重登，没有拼接旧失败样本 | [证据](browser/README.md) |
| C17 | 390×844及键盘长清单 | PASS | 导航/弹窗可达；B两项243/228字、焦点/保存刷新/无横溢；仅窄视口 | [证据](browser/r5-chain-b.json) |
| C18 | 手册对照 | PASS | 14章工程代理逐步对照，6张实际截图；真人无协助使用尚未执行 | [证据](manual-check.md) |

C12–C14以真实PG事务运行支撑；SQLite/DOM不会替代它。C01–C15的具体用例、叶子日志行号和各自证据类型见子条件CSV。C18是工程代理对照，不代表真人首次使用、满意度或耗时达标。
