# 缓存候选的 I11 八项负向核验

受测候选为 `94d531faeb6fa65f76599e3d5c04d83b72fda540`。复核采用冻结版本的 [PG 入口](https://github.com/VanTam-CN/law-oa-go/blob/94d531faeb6fa65f76599e3d5c04d83b72fda540/scripts/np-ci/early-pg.sh)、[Go 事件检查器](https://github.com/VanTam-CN/law-oa-go/blob/94d531faeb6fa65f76599e3d5c04d83b72fda540/scripts/np-ci/check-go-events.py) 和 [fuzz 探针](https://github.com/VanTam-CN/law-oa-go/blob/94d531faeb6fa65f76599e3d5c04d83b72fda540/scripts/np-ci/i11-fuzz-probe.sh)。

| 探针 | 触发的条件 | 冻结检查器应返回的原因 |
| --- | --- | --- |
| missing-pg | 提供正确的连接配置，隔离端口没有 PostgreSQL | M2_PG_UNAVAILABLE |
| missing-dsn | 清除 PG 连接配置 | M2_DSN_MISSING |
| skip | 临时编译覆盖让关键原生草稿用例 Skip | M2_REQUIRED_TEST_SKIPPED |
| partial | 只运行 schema 测试，遗漏关键草稿用例 | M2_REQUIRED_TEST_MISSING |
| none | 选择器没有匹配任何用例 | M2_NO_TESTS |
| fuzz-unknown | 给已注册的预检步骤传入未知测试类型 | I11_FUZZ_PREFLIGHT_REJECTED |
| fuzz-selected-failed | 选中组返回 failure，即使其验证标记为 true | I11_FUZZ_SUMMARY_REJECTED fuzz-selected-failed |
| fuzz-missing-proof | 选中组返回 success，但缺少 verified 标记 | I11_FUZZ_SUMMARY_REJECTED fuzz-missing-proof |

五项 PG 探针检查真实入口及其失败顺序。三项 fuzz 探针用受控输入调用候选中注册的预检、汇总步骤。首稿 `2ae4276` 将零用例与缺部分用例混同，并在一个作业中汇总退出状态。协调者退回首稿，以冻结检查器的两个独立错误标记和八个原生作业终态为验收依据。

[同候选的 17 阶段正向运行](../cache-full-i11/README.md)和 [PR #52 测试合并提交核验](../test-merge-i11/README.md)分别记录。

## 协调者核验

**八项全部命中预期红因。** GLM 在独立分支 `codex/np-cache-i11-neg8-hosted` 推送工作流 `de33450`，GitHub 第一次运行即取得八个失败作业。协调者直接读取八份作业日志和步骤终态，逐项核对了上表标记、唯一失败步骤及退出码 1。各作业均先验证工作流提交的唯一父提交与文件差异，再切到干净的 `94d531f` 执行。矩阵关闭 fail-fast。最终运行留存产物为 0。[托管矩阵](https://github.com/VanTam-CN/law-oa-go/actions/runs/45768562580) · [每项作业链接和实际错误标记](verification.json)

两项 fuzz 汇总各自传入固定的 validators 组，前者提供验证标记但组失败，后者组成功但缺标记；二者由同一冻结汇总条件拒绝。上述负向运行的 GitHub 终态为 failure，八项负向验收为 PASS。

**6.2 Astra 独审通过。** 独审读取同一次托管运行的八份原始日志，核对了唯一直接父提交、八次 RUN_SHA、失败步骤、退出码和每个冻结输入的拒绝原因。独审亦确认 PR 事件的 I11 通过仅归属 `9883cb1`。核验时负向工作流提交触发 1 次运行，该分支的 Actions 缓存条目与该提交的部署记录均为 0。Hermes 任务为 `t_90117f63`。合并、部署、真实试用均为 HOLD。
