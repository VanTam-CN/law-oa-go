# 七路径集成候选的普通 Chromium 验收

Hermes 卡 `t_91c609d0`。在 `f3cd764` 上引入已审的缓存提交 `0b58108`，形成独立候选 [ee1db31](https://github.com/VanTam-CN/law-oa-go/commit/ee1db311682410653e6129c985c52efa3ef91a97)。本地与 GitHub 的完整树均为 `82a38945b80adf59aee952bfd13e1b965bcfc8f2`，与[七路径集成预演](../integration-delta/README.md)相符。

GLM-Flash 从候选建立验证分支，使用此前托管成功的工作流。协调者逐字节比对工作流，只有触发分支和候选 SHA 两处替换。GitHub 实际触发提交为 `f1e5c87`，其唯一父提交为 `ee1db31`；二者仅差一个工作流文件。运行以只读权限检出，执行前切回无改动的候选工作树。[本次 GitHub 作业](https://github.com/VanTam-CN/law-oa-go/actions/runs/46160296336/job/141025887805)

**普通 Chromium 验收通过。** 托管作业首次运行成功。115 个用例全部首次通过，0 重试、0 跳过；与冻结清单逐项比对，缺失、新增和重复均为 0。6.2 Astra 独立提取本次作业的 115 条 CASE，复核名称、ID、实际尝试、就绪标记和运行版本，确认此项 PASS。[逐用例对照](cases.tsv) · [结构化结果](verification.json)

## 核对口径

以[上轮冻结的用例清单](../e2e-full/cases.tsv)为准，按 Chromium 项目、测试 ID 和标题一一匹配运行报告，分别核对预期、终态和每次尝试。普通配置含 17 个规格文件，排除专用的 `np-first-slice.live.spec.ts`。服务使用本次作业的合成 PG15 和 `cmd/server`。普通浏览器导航至 3003 开发服务，4173 仅检查构建预览首页返回 HTTP 200。3003 的判定依据是本次执行命令和固定候选的 Playwright 配置；日志没有逐请求网络轨迹。

## 其他门禁

| 门禁 | 在 ee1db31 上的状态 | 可参考的历史基线 |
| --- | --- | --- |
| 缓存定向回归 | NOT_RUN | [94d531f 定向验证](../cache-hit-hosted/README.md) |
| 完整 I11 正向 | NOT_RUN | [94d531f 正向](../cache-full-i11/README.md)、[PR 测试合并 9883cb1 正向](../test-merge-i11/README.md) |
| I11 八项负向 | NOT_RUN | [94d531f 八项预期红灯](../cache-i11-negatives/README.md) |
| 4173 登录并读取本人草稿 | NOT_RUN | [f3cd764 构建预览专项](../preview-auth-read/README.md) |
| PR #52 合并、部署、小范围真实试用 | 分别 HOLD | [放行条件](../../README.md) |
