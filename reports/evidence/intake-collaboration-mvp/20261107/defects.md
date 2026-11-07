# 缺陷关闭与首次失败索引

最终运行源码 `3f970040f5e40b8016c18a070b5aa1345d2b2eec`。下列本轮工程阻断已关闭；原始失败和审查保留各自时点，不代表生产通过。

| 项目 | 首次证据 | 修复与关闭 |
|---|---|---|
| claim状态、通用Submit唯一槽、legacy读写/重放 | [原始记录](review/core-review.md) | 统一持久来源与冻结消费；transactions/53-final-handoff.md及bootstrap最终回归 |
| 6位时间编号并发碰撞 | [原始记录](transactions/09-full-real-pg-race.log) | 完整UUID请求编号；transactions/52-final-transition-results-summary.json |
| B1 旧接案审批绕过唯一成案 | [原始记录](review/backend-final-review.md) | 两类入口与成案/重试统一守卫，review/backend-final-recheck.md |
| B2 换检测后待办不可见 | [原始记录](review/backend-final-review.md) | 权威快照批量投影；inbox/b2/REPORT.md及独立复核 |
| B3 历史快照对象授权 | [原始记录](review/backend-final-review.md) | 所选历史与当前上下文分别校验；撤权隐藏ID、恢复只读，独立复核 |
| R1 换客户未重提时补充被拒 | [原始记录](review/backend-final-recheck.md) | 可编辑与冻结上下文分别授权；inbox/r1/REPORT.md及7c独立40项 |
| B4 新库缺claim/唯一索引 | [原始记录](browser/r3-initial-submit-first-red.md) | v16 bootstrap parity、真实新库/v15升级/首提HTTP；bootstrap/HANDOFF.md |
| B5 中间节点提前成案409 | [原始记录](browser/final-a-stage-first-red.md) | 3f仅最终approved进入成案；作者最终6PG、独立原4PG，R5两条新浏览器主链 |
| F1/F2 当前/历史快照与旧轮次token | [原始记录](review/frontend-review.md) | 明确当前版本/提交token；独立回归及R5两轮审批 |
| F3/F4 迟到草稿与空白入口残留 | [原始记录](review/frontend-review.md) | 请求上下文隔离、不覆盖新输入、清空对象；独立DOM及换账号浏览器 |
| F5 普通非接案审批被禁用 | [原始记录](review/frontend-review.md) | 服务端submission_snapshot_required，确认无关联才允许普通流程；失败不降级 |
| F6 检测/复核/覆盖摘要误导 | [原始记录](review/frontend-review.md) | 区分事实与当前审核；R4作者32及独立29DOM通过 |
| 草稿丢复核、弹窗/抽屉遮挡 | [原始记录](iteration.md) | 绑定恢复、camelCase及浮层修复；mobile-drawer-retest与R5窄视口真实操作 |

客户写成功后丢响应/网络中断属于主动故障注入，见[browser/client-response-loss.json](browser/client-response-loss.json)。测试初次envelope预期、C06字符串ID fixture与socket权限失败保留，不冒充产品缺陷，也不将部分叶子成功改判整个命令通过。

AntD虚拟滚动、离场图标仍参与可访问名称、异步焦点及路由等定位问题按实际DOM处理，没有强制点击绕过遮挡。[最终报告](final-report.md)列明未验证边界及剩余产品限制。
