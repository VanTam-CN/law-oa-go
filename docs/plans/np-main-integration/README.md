# NP 下一阶段：主干集成准备 v0.2

2026-11-20修订；v0.1形成于2026-11-19。**只读核验与方案阶段；未实施、未合并、未部署。** 本版吸收9080b25的独立只读复核：定位job级if上下文错误，细化I13前后端分流，新增M0材料边界A/B出口。没有把缓存方案写成风险已消除，也没有默认降低I09验收。

## 固定基线

- main：`d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7`。
- 首切片：`2d659d9dd750f49ddb1f517746e22851392231a6`；应用候选 `4c184d5d46f1e101d3ff0ced04f62ebd0b16d6f3`。
- 冻结父：`bcd1b8992b3f3450d5ab406afd74336cc282eece`。
- 准备分支：`codex/np-main-integration-prep`，从上述 main 新建，原脏工作树和首切片均不纳入修改。

首切片工程 PASS 只适用于其固定候选；main 集成、缓存改造、真实 main+Redis、CI 新门禁及升级回滚均尚未验收。

## 阅读顺序

1. [实施与验收计划](implementation-and-acceptance-plan.md)：主代理复核裁决、阶段依赖、范围、工作包、门禁及估算。
2. [缓存策略提案](cache-policy-proposal.md)：推荐仅内存方案、历史正文处置和最小恢复标识。
3. [目录映射复核](directory-mapping-review.md)、[前置能力复核](prerequisites-review.md)、[CI 复核](ci-review.md)：GLM Flash 并行检查，正文以固定路径和 SHA 为证据；主代理复核裁决见第 1 项。
4. [源文件映射事实](source-baseline.json)、[远端只读观察](remote-observation.json)：精确映射、共同祖先集合核对，以及现有工作流状态。
5. [独立集成 PR 正文准备稿](draft-pr-body.md)：只准备正文，本轮不创建 PR。

## Git 与本轮边界

准备分支仅提交本目录的文档与只读快照。PR #47 保持原固定候选和 Draft 状态，不追加本轮主干准备材料，也不把它直接改投 main。

本轮没有创建 main 目标 PR：现有 PR 工作流配置了评论、第三方上传，以及按 secrets 条件执行的 Slack/邮件通知；实际成功与否未验证。先完成 CI 事件/副作用隔离方案，后续在获准实现的分支收口后再创建独立集成 PR。没有为绕开这些副作用修改仓库权限、停用 Actions 或改工作流。

本次未运行业务构建、数据库、浏览器、真实环境、模型接口或部署，也未手动触发Actions。现有定时CI的成功/失败只是历史观察。另须区别记录：准备分支发布后，既有fuzzing.yml自动形成[失败run 44508335288](https://github.com/VanTam-CN/law-oa-go/actions/runs/44508335288)，jobs=0、check-runs=0；CLI提示可能是工作流文件问题，具体校验原因未取到。此项已登记为CI前置问题，不是业务测试失败或新门禁通过。
