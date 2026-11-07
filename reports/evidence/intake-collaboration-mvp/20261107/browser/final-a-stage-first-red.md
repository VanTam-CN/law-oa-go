# B5：第一审批节点错误触发成案

2026-11-07 17:35:23 CST，候选 `7c5618f26ed5600de76e4ed94a682252d10a8aef`，真实浏览器、全虚构隔离 PostgreSQL。

客户71；接案eba99230-1fa1-48ee-b9a0-b0b532879462；审批18e5b8f062cb0200。新客户、重登恢复、双窗口409、正式检测/独立复核、两项退回、390px逐项回复保存刷新、修改身份后重提409、新检测复核及原审批第二次提交均已完成。

管理合伙人点击“确认同意并通过本阶段”，integration decision 返回409。页面确认框保留；同一审批已进入under_review，下一审批人为合规负责人。日志显示中间阶段仍进入自动成案，报“审批未处于通过状态，已阻止成案”。因此不将该样本计为完整C16通过；未继续最终审批，未手改业务数据。

根因位于ProcessApprovalWithConflict：ProcessApproval提交节点后，外层仍仅以decision==approve和requiresCaseCreation触发ensureApprovedApprovalHasCase，没有检查updatedApproval.Status是否最终approved。

首红现场和此前页面事件保留。修复后需要真实PG两级HTTP回归、独立专项复查，并在新候选重新完成两条新样本主链。
