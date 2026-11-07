# 390px 导航关闭复测

2026-11-07 16:27 CST，固定构建 `index-p6KdlFyE.js`；真实 Chromium 页面，视口 390×844。

1. 新建接案页输入 `QA 移动导航修复验证`，保持未保存。
2. 打开导航菜单，点击工作台。
3. 出现“尚有未保存修改”，点击“继续编辑”。
4. 点击仍打开的导航抽屉“关闭”，实际点击成功，抽屉隐藏。
5. URL 保持 `/case/create`，标题仍为上述输入。

结果：**PASS（仅本项修复）**。没有强制点击、删除遮挡元素或程序化伪造关闭状态。首红保留在 `screenshots/05-mobile-drawer-close-blocked.png`；原来浏览器提示 `app-header intercepts pointer events`。

重载测试页面时触发真实 beforeunload 提示，已接受丢弃本轮虚构测试输入后重建同一条件；正式用户数据未受影响。本项不代替长补正清单、完整链或整体 C17。
