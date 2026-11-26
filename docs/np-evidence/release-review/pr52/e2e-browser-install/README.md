# PR #52 普通 E2E 的浏览器安装故障

PR 的头提交为 `b22736b`，测试合并提交为 `9883cb1`。两者具有同一棵源码树。2026-11-25 的普通 E2E 作业中，GitHub 将 `ubuntu-latest` 解析为 Ubuntu 26.04。使用项目锁文件执行 `npm ci` 成功，随后 `npx playwright install --with-deps chromium` 退出 1。原始日志给出的原因是 `Playwright does not support chromium on ubuntu26.04-x64`。锁文件中的 Playwright、Playwright Core 与测试运行器均为 1.60.0。[原失败作业](https://github.com/VanTam-CN/law-oa-go/actions/runs/45584687934/job/139107230670)

Playwright 安装是该作业唯一失败的步骤。构建前端、启动前后端、执行 Chromium 测试均显示 skipped。PG 服务初始化成功，结束时打印的 PostgreSQL 日志无法解释此前的浏览器安装错误。

## 候选与证据边界

GLM 从 PR 头建立独立分支 `codex/np-pr52-e2e-runner`。可单独移植的提交为 `8a8fb8a`。协调者逐字节比较了工作流，唯一变化为普通 `e2e-tests` 作业将 `runs-on` 从 `ubuntu-latest` 固定到 `ubuntu-24.04`。锁文件、业务代码、Playwright 测试和其他作业均沿用 PR 版本。[单行修订](https://github.com/VanTam-CN/law-oa-go/commit/8a8fb8a9b8857fc87634561def33a6cc0783c9dc)

首次隔离运行已成功安装并启动浏览器。其末尾版本摘要在仓库根目录执行 npx，额外下载了 Playwright 1.65.1，该摘要不能证明前端使用的版本。修订后的工作流删除这一步，在前端目录将实际安装的 playwright 和 playwright-core 分别与同一锁文件核对，再启动该安装的浏览器断言合成页面标题。[首次运行](https://github.com/VanTam-CN/law-oa-go/actions/runs/45770085878/job/139740362860)

本卡只验证浏览器安装与启动。普通 E2E 的业务用例、后端启动和整套作业的结果需在后续验证。PR #52 当前的测试合并提交仍为 `9883cb1`，缓存候选 `94d531f` 的 I11 验收另行记录。

## 最终结果

**独立候选的浏览器安装首因已关闭。** 最终托管入口 `35fadeb` 检出 `8a8fb8a`。在 Ubuntu 24.04 上，以原命令安装依赖和浏览器成功。启动步骤确认本地 playwright、playwright-core 均为锁定的 1.60.0，实际启动 Chromium 148.0.7778.96（revision 1223），成功读取合成页面的预期标题。协调者复查作业步骤和原始日志，6.2 Astra 独审通过。Hermes 任务为 `t_0df7039f`。合并、部署、真实试用均为 HOLD。[最终托管作业](https://github.com/VanTam-CN/law-oa-go/actions/runs/45770597811/job/139742025811) · [结构化证据](verification.json)
