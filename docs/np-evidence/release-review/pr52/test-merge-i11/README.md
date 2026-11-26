# PR #52 测试合并提交的 I11 核验

GitHub 当前给出的测试合并提交为 `9883cb11b68aa80a9291e55719aef56d29e13aab`。两个父提交依次是主干 `d9c92fb` 和 PR 头 `b22736b`，合并树与 PR 头的树对象相同。与缓存候选 `94d531f` 相比，PR 的树缺少缓存中间件、入口及相应测试的改动，还有两份证据文件的差异。[身份、作业与逐阶段记录](verification.json)

**该提交的 I11 作业通过，所属 Test Pipeline 整体失败。** 本轮重新读取了已有 pull_request 作业的原始日志。事件标记、实际 checkout 与 I11_SOURCE 均绑定当前测试合并提交，17 个阶段有配对起止记录且均返回 0。34 个浏览器场景通过，路线 B 占 4 个。PG 来源 +3、回执 +4，受保护行摘要一致，终态出现 I11_FULL_PASS。普通 E2E 作业失败，PR 整体放行仍为 HOLD。[I11 作业](https://github.com/VanTam-CN/law-oa-go/actions/runs/45584687934/job/139103174713) · [同一次 PR 流水线](https://github.com/VanTam-CN/law-oa-go/actions/runs/45584687934)

本轮复用了现有 GitHub 运行。head、base 或测试合并提交发生变化时，按新 SHA 核验。
