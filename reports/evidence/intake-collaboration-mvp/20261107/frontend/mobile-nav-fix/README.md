# R3：移动导航与冲突展示收尾交接

已冻结，可交 Planck 复核。时间 2026-11-07T08:25:55.807Z，基线 a806831a7cef839c2b971cd4ab260cce77f42b8b，工作目录 /private/tmp/lawoa-intake-collaboration-20261107。相对 R2 只有五个文件发生变化（3 个实现文件、2 个测试文件），详见 source-manifest.json 的 delta_from_r2。F1—F4 的实现与客户流程未再修改；Batch01Prototype.tsx 的新增差异仅在 ConflictCheckResults。

## 固定产物

- 当前入口 **index-p6KdlFyE.js**；静态目录 /private/tmp/lawoa-intake-collaboration-20261107/frontend/dist。
- [源码清单](source-manifest.json)（27 个累计前端文件、相对 R2 五文件差异）；[构建清单](build-manifest.json)（7 个静态文件）。
- [构建归档](frontend-dist.tar.gz)，SHA256 ef0dcd9b5cee75717585aeddd487edadee337378609a83722c4e80c802c0254b。
- [源码归档](frontend-source.tar.gz)，SHA256 31f28869cd5706c3d733038abf5bcb5655b1a7fb0e7a6ba54334f585da6c6860。
- 两归档内每个文件均与源码/dist 字节哈希一致：integrity-check.json。旧 R2 归档保留，不覆盖。

## 修复与复查点

1. [Sidebar](../../../../../../frontend/src/components/layout/Sidebar.tsx#L412)：仅移动导航 Drawer 显式 zIndex=1040，高于固定 header/sidebar 的 1030。[DraftNavigationGuard](../../../../../../frontend/src/pages/batch01/intake/DraftNavigationGuard.tsx#L16) 的未保存 Modal 为 1050，继续高于 Drawer。没有修改全局层级、header、客户弹窗或依赖。
2. [ConflictCheckResults](../../../../../../frontend/src/pages/batch01/Batch01Prototype.tsx#L6371)：把已打开详情 GET /review 或本页 POST /review 的有效 checkId 对应复核留在组件内，更新同一队列行、待复核统计及筛选；不为未打开的行请求 review。迟到的旧详情复核响应不应用到新选择。机器风险列改名“自动检索风险”，与使用已读复核的“处置状态”区分。
3. 空证据只写“未发现匹配证据”；未复核提示只说明独立复核仍必需，不再由空证据推断覆盖受限或覆盖完整性未知。

## 实际验证

- typecheck / lint / build / diff-check 均 exit 0，分别见 *-final.log 和 diff-check.log。
- 导航相关：3 文件 9/9（MobileNavigationLayer、Sidebar.mvp、IntakeCollaborationPage），jest-final.log。
- 冲突详情相关：1 文件 6/6（ConflictDetailModal），accuracy-tests-final.log；其中两项新增回归验证已读 review 更新对应“处置状态”、未打开行不额外读取，以及不猜测覆盖受限。
- 未重跑 R2 的完整 75 项；该批证据仍绑定 R2，单列在 build-manifest.json.prior_evidence，不能称本次重跑。
- 导航首红 layer-first-red.log 明确为 Drawer 1000 不高于 header 1030；准确展示首红 accuracy-first-red.log 复现列表旧结论和空证据误文案。首次测试定位修正和非空断言 lint 失败亦保留，未屏蔽断言、调低检查或安装依赖。
- JSDOM 使用真实 Drawer/Modal、真实路由阻拦及编译后的实际 header/sidebar 样式，验证继续编辑后关闭 Drawer 保留输入/路由、直接离开后才关闭。它不能证明真实 pointer 命中或 390px 键盘可达性。

## 主线程证据与下一步

用户已报告 R2 的真实 A 草稿恢复通过、1200×1041 详情已复核且隐藏旧冲突审批通过；C01/C02 真实客户69、C03 第二页选14、390×844 新增客户弹窗键盘与按钮位置通过，导航“继续编辑”保留输入通过。这些是主线程已记录的实测结果，本作者未重复浏览器/真实 API 操作。

当前构建只需优先复验：未保存输入→导航工作台→继续编辑→Drawer 关闭按钮可点且输入保留；A 详情关闭后对应列表处置状态已更新、空证据不再误称覆盖受限。随后 Planck 可基于相对 R2 五文件差异增量复审。C16、整体 C17、两条新完整链不由本补丁宣告通过；不批准旧 d08a805d。

已停止编辑。未改后端、安装依赖、启动/停止服务、操作浏览器、调用外网/代理、commit 或 push。
