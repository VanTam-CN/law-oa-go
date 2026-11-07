# 接案协作 MVP 工程交付

2026-11-07，Codex 作者实施、独立代理审查及主线程本地浏览器验证。N0–N6 范围已完成，C01–C18 在本轮规定的工程环境下通过；真人律师试用、正式环境及生产验收未执行，`production_ready=false`。历史方案的“GLM 执行版”是来源名称，本轮没有冒称 GLM 执行。

## 交付内容

律师可在接案页原地建客户、保存并恢复草稿，从工作台进入指定补充申请；审批人按结构化清单退回，律师逐项回复并在原审批重提。事实变化必须重新检测及独立复核；仅说明变化由服务端判断能否复用。两级审批仅在最终批准后生成一个完整案件，依据是本轮冻结快照。

服务端同时约束对象权限、旧版本决定、唯一占槽、客户创建幂等、待办交接与事务回滚。历史快照不可覆盖，workflow 待办不可手工完成或删除。迁移 82–84 与 PostgreSQL bootstrap v16 已在隔离库验证。

## 候选与 PR

- 范围基线：`a806831a7cef839c2b971cd4ab260cce77f42b8b`。
- 初始源码提交：`7c5618f26ed5600de76e4ed94a682252d10a8aef`；最终运行源码：`3f970040f5e40b8016c18a070b5aa1345d2b2eec`，修复中间审批节点提前成案。
- 浏览器后端 SHA-256：`fa46d5933e3d0f3c8b5b7c9780325418a1e414103f04d42ecc1e7346c05f29ae`；前端固定 `index-fDMF1BW8.js`。完整产物见 [运行绑定](browser/candidate-r5.json)。
- 最终交付提交在上述源码之后只增加测试、文档和证据；其 SHA 以 PR head 为准。[候选内容核对](candidate-binding.json) 记录测试后的源码一致性。
- [Draft PR #46](https://github.com/VanTam-CN/law-oa-go/pull/46)，目标为精确基线分支 `codex/intake-collaboration-mvp-base`。当前 main 与基线分叉，前置修复尚未按此链合入 main；本 PR 不自动合并、部署或改变 main。

## 实际验证

不同阶段、同一用例的重复执行不累加成总数；作者执行不等于独立执行。

| 层次 | 实际结果 | 证据与适用范围 |
|---|---|---|
| 作者后端基线回归 | 300 叶子 PASS，0 FAIL/SKIP，其中 111 使用真实 PG；race、build、受影响包 vet 通过 | [bootstrap 交接](bootstrap/HANDOFF.md)。运行时尚未提交，按 601 文件哈希绑定到 7c；不是在 3f 上重新运行了全部 300 项 |
| B5 修复后的专项 | 原 4 个真实 PG/HTTP 叶子通过；扩展实际建案配置后最终 6 PASS，0 FAIL/SKIP | [多级审批](multistage/HANDOFF.md)、[配置双 tracking](multistage-config/HANDOFF.md)。最终 6 包含原 4；build/vet 通过，生产源码不再变化 |
| C06 权限补证 | 作者 2 个真实 PG/HTTP 叶子 PASS | [C06 交接](inbox/c06/HANDOFF.md)。有效助理请求拒绝且无写入，律师阳性对照成功；核查岗撤权和恢复四入口同步 |
| 独立后端 | 7c 阶段 40 PASS，其中 20 PG；3f 的 B5 另 4 PG PASS；C06 另 2 PG PASS | [独立复核](review/backend-final-recheck.md)。新增配置场景另有测试审读，不能当作独立运行 |
| 前端作者 | R4 32 项 DOM/mock 回归 PASS，type-check、修改范围 lint、build 通过 | [R4 交接](frontend/approval-review-fix/README.md)。不将前轮 75 项重复计入；不声称全库 lint 无历史问题 |
| 独立前端 | 29 项 DOM/mock 回归 PASS，F1–F6 关闭 | [独立前端审查](review/frontend-review.md)。不是独立浏览器或真人体验证据 |
| 浏览器 + SQL | 同一最终候选 A/B 两条独立完整链通过；案件 3/4 各唯一、主体与批准快照一致、开放流程待办均为 0 | [浏览器索引](browser/README.md)。A 覆盖 5 个角色关系/4 个实体；B 仅说明变化复用有效复核 |
| 迁移、手册、规格 | 隔离 up/down/up、真实 v16 初始化及 v15 升级通过；14 章手册与 6 张截图对照；OpenSpec strict 记录在命令索引 | [迁移回退](migration-rollback.md)、[手册核对](manual-check.md)、[命令记录](commands.jsonl) |

[C01–C18 最终矩阵](acceptance-matrix.md) 与 [115 子条件逐项映射](final-subconditions.csv) 给出证据类型、历史状态和最终补证。首次红灯保留在原始日志及 [缺陷索引](defects.md)，没有通过修改失败样本数据库制造成功。`SHA256SUMS` 覆盖本目录交付文件，原始日志和清单中的旧时间/旧 SHA 不改写。

## 使用与限制

[试用者操作手册](../../../../docs/试用者操作手册.md) 可用于下一次受控真人试用。当前待办标题有通用事项名，需进入后核对对象；案件详情未逐条列出全部主体角色，工程核对使用批准快照与案件表。窄视口 390×844 通过不代表手机真机。

兼容建案配置可能产生 NULL-intake tracking 与非 NULL 唯一占槽两行。最终状态读取和同案重放已覆盖“退回重提”及“无退回直接两级批准”；不能把 tracking 总行数当作案件数量，也不能声称 getter 显式优先 canonical 行。

GitHub 检查列表当前为空，CI 为 NOT_RUN。真实律师独立使用、满意度/耗时测量、生产部署、真实数据迁移、备份恢复和异机灾备均 NOT_RUN。下一阶段是核对前置基线的合并路径，并按现有试用边界安排真人操作；本次不因此宣称产品或生产验收完成。
