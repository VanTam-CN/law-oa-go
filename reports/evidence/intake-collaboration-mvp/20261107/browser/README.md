# 实际浏览器与数据库证据

最终运行候选 `3f970040f5e40b8016c18a070b5aa1345d2b2eec`；前端固定 `index-fDMF1BW8.js`，后端 `qa-browser-r5`。哈希见 [candidate-r5.json](candidate-r5.json)。使用 Playwright MCP、真实本机 router、独立 PostgreSQL 和全虚构账号资料；没有 mock 审批、冲突复核、成案或数据库结果。

## 两条独立完整主链

| 对象 | A：修改身份后重新复核 | B：仅回复文字说明 |
|---|---|---|
| 客户 | 72 | 73 |
| 接案 | 5eeb7bd3-ec6a-48e9-834b-4b5edff6c055 | 839f3fd3-81e8-4329-83da-46a486ac36ef |
| 原审批 | 18e5ba60a968c0a0 | 18e5bad3447ea150 |
| 正式案件 | 3 | 4 |
| 主体角色 / 实体 | 5 / 4 | 2 / 2 |
| 提交版本 | 1→2 | 1→2 |
| 独立复核 | review 5→6，检测编号更新 | review 7 保持，同检测编号 |
| 第一审批节点 | HTTP200，under_review，下一人5，未成案 | 相同 |
| 最终节点 | HTTP200，approved，唯一案件3 | HTTP200，approved，唯一案件4 |

两条都从新客户开始，经保存退出、重新登录恢复、正式检测与独立复核、首提、审批人两项退回、律师从工作台补充入口进入同一申请、回复保存刷新、原审批重提、两个节点批准，到实际案件详情。A的五条角色覆盖共同委托、多对方、同名异身份和同主体多角色；B验证不变事实下沿用有效独立复核。事件记录分别为 [r5-chain-a.json](r5-chain-a.json)、[r5-chain-b.json](r5-chain-b.json)。它们不是由旧失败样本拼接。

[r5-final-db.json](r5-final-db.json) 将退回记录ID、旧快照、稳定补充项ID、第二轮请求/回复ID和案件关联配对；两条开放流程待办均为0。A第一次事件读取标题早于异步恢复，后续首提记录的restored_title已确认正确；不把早读空字符串当作恢复通过。

[r5-final-consistency.sql](r5-final-consistency.sql) 与 [结果](r5-final-consistency.json) 在不输出身份原文/摘要的前提下，比较批准快照与案件主体的名称、身份摘要、业务角色多重集合。两条均无缺失、无额外角色；每个虚构标题一个案件，每个接案一个完成的占槽。旧NULL-intake兼容tracking与接案占槽分开，不能把tracking总行数解释为案件数。

## 补充场景

- 客户响应丢失、双击及同key冲突：[client-response-loss.json](client-response-loss.json)、[client-db-counts.json](client-db-counts.json)。先真实创建再中断响应，重试同ID；相关5个文件与最终源码同哈希，见最终子条件索引。
- 最终R5全字段比较：[r5-all-fields-preserved.json](r5-all-fields-preserved.json)。30输入、13选择，取消/网络失败/恢复后真实200均保持一致。网络失败只注入本机客户POST，恢复时显式解除拦截；不将该故障注入称为真实服务器故障。身份值未写入该比较证据。
- 客户后页选择与弹窗键盘：[client-paging-mobile.json](client-paging-mobile.json)；抽屉修复：[mobile-drawer-retest.md](mobile-drawer-retest.md)。
- 同源前端的真实双窗口409、换账号不串草稿、修改身份后拒绝旧检测：[final-a-events.json](final-a-events.json)。相关功能代码在B5修复中未变。
- 最终R5核查岗从工作台“独立复核”到指定check并实际提交：[r5-officer-workflow.json](r5-officer-workflow.json)。完成前需要处理2项、完成后1项；剩余是旧审批，目标复核动作已消失。
- 最终R5 B在390×844显示两条243/228字要求，键盘定位摘要、保存回复、刷新保持稳定ID，页面无横向溢出；仅证明窄视口，不称原生手机或真机测试。

## 原始失败与边界

R3初始化缺字段、7c第一节点提前尝试成案等首红原样保留，见 [final-a-stage-first-red.md](final-a-stage-first-red.md) 与 [r3-initial-submit-first-red.md](r3-initial-submit-first-red.md)。旧样本没有通过改数据库修绿。

工具中出现的延迟路由、AntD虚拟列表滚动、loading图标已离场但仍参与可访问名称、0毫秒异步焦点、未保存修改对话框等均按实际DOM处理，没有强制点击遮挡元素。缺陷修复和工具定位修正分开记录。

工程代理操作不等于真人律师试用；未执行生产部署、真实数据迁移、备份恢复或异机灾备。`production_ready=false`。
