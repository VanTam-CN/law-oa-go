const cases = [
  {
    id: "M-2026-0412",
    name: "衡岳新能源股权回购争议",
    client: "上海衡岳新能源有限公司",
    opponent: "北辰资本管理有限公司",
    owner: "周律师",
    status: "红色命中",
    statusClass: "danger",
    next: "合规负责人复核",
  },
  {
    id: "M-2026-0398",
    name: "云启生物常年法律顾问",
    client: "杭州云启生物科技有限公司",
    opponent: "无直接对方",
    owner: "梁律师",
    status: "待补客户关联方",
    statusClass: "warning",
    next: "客户经理补全",
  },
  {
    id: "M-2026-0387",
    name: "北港建设工程结算仲裁",
    client: "北港城市建设集团",
    opponent: "青澜地产有限公司",
    owner: "陈律师",
    status: "可豁免",
    statusClass: "info",
    next: "生成披露函",
  },
  {
    id: "M-2026-0376",
    name: "宁泰医疗劳动争议批量案",
    client: "宁泰医疗器械有限公司",
    opponent: "离职员工群体",
    owner: "沈律师",
    status: "已通过",
    statusClass: "success",
    next: "生成案号",
  },
];

const clients = [
  {
    name: "上海衡岳新能源有限公司",
    type: "机构客户",
    completeness: "98%",
    risk: "高",
    matters: 7,
    aliases: "衡岳新能源、HY Energy",
    contacts: "法定代表人：顾明；董秘：赵珊",
    relations: ["衡岳储能科技有限公司", "衡岳香港控股", "东辰产业基金"],
  },
  {
    name: "北港城市建设集团",
    type: "机构客户",
    completeness: "91%",
    risk: "中",
    matters: 12,
    aliases: "北港城建、BGCC",
    contacts: "总法：唐敏；项目负责人：陆川",
    relations: ["北港隧道工程", "北港投资控股", "海晟监理"],
  },
  {
    name: "林若安",
    type: "个人客户",
    completeness: "84%",
    risk: "低",
    matters: 2,
    aliases: "Elaine Lin",
    contacts: "本人；配偶：王岑",
    relations: ["岑安咨询工作室", "前任雇主：云舟科技"],
  },
];

const tasks = [
  ["红色冲突复核", "衡岳新能源 vs 北辰资本，命中历史融资文件和对方现有委托。", "今天 16:00", "danger"],
  ["客户关联方补全", "云启生物缺少控股股东、董事及历史名称，需要完成后再检查。", "明天 10:00", "warning"],
  ["披露函确认", "北港建设仲裁事项可豁免，等待双方知情同意模板确认。", "04-26", "info"],
  ["案件启用", "宁泰医疗批量劳动争议已通过冲突检查，可生成案号和文件夹。", "04-26", "success"],
];

const conflicts = [
  {
    title: "直接对立冲突",
    level: "不可豁免",
    cls: "danger",
    summary: "北辰资本是本所现有客户，当前拟代理客户将在同一股权回购争议中直接对立。",
    evidence: ["现有委托：北辰资本常法", "事项相关度：92%", "规则：中国律师执业规范第49条"],
    action: "建议拒绝代理或变更代理安排",
  },
  {
    title: "前任客户信息冲突",
    level: "可豁免",
    cls: "warning",
    summary: "拟办案团队成员 2024 年参与过北辰资本融资尽调，可能掌握相关保密信息。",
    evidence: ["历史事项：B 轮融资", "冷却期：未满 3 年", "缓解：伦理壁垒+知情同意"],
    action: "提交合规审批并生成披露函",
  },
  {
    title: "关联实体命中",
    level: "需人工判断",
    cls: "info",
    summary: "衡岳香港控股与东辰产业基金存在少数股权关系，暂未发现直接对立。",
    evidence: ["股权比例：8.4%", "关系类型：间接持股", "置信度：64%"],
    action: "补充工商关系后复检",
  },
];

const pageTitles = {
  dashboard: "MVP 工作台",
  cases: "案件管理",
  clients: "客户管理",
  conflicts: "利益冲突",
};

const $ = (selector) => document.querySelector(selector);

function statusPill(text, cls) {
  return `<span class="status-pill ${cls}">${text}</span>`;
}

function renderTasks() {
  $("#taskList").innerHTML = tasks
    .map(
      ([title, desc, due, cls]) => `
        <article class="task-item">
          <div>
            ${statusPill(title, cls)}
            <p>${desc}</p>
          </div>
          <strong>${due}</strong>
        </article>
      `,
    )
    .join("");
}

function renderCases() {
  $("#caseRows").innerHTML = cases
    .map(
      (item) => `
        <tr>
          <td><strong>${item.id}</strong><br />${item.name}</td>
          <td>${item.client}</td>
          <td>${item.opponent}</td>
          <td>${item.owner}</td>
          <td>${statusPill(item.status, item.statusClass)}</td>
          <td>${item.next}</td>
        </tr>
      `,
    )
    .join("");
}

function renderClients(selected = clients[0]) {
  $("#clientList").innerHTML = clients
    .map(
      (client, index) => `
        <article class="client-card" data-client-index="${index}">
          <div>
            <strong>${client.name}</strong>
            <p>${client.type} · 别名：${client.aliases}</p>
          </div>
          ${statusPill(`完整率 ${client.completeness}`, client.risk === "高" ? "danger" : client.risk === "中" ? "warning" : "success")}
        </article>
      `,
    )
    .join("");

  $("#clientProfile").innerHTML = `
    <p class="eyebrow">客户档案</p>
    <h3>${selected.name}</h3>
    <p class="muted">${selected.contacts}</p>
    <div class="profile-metrics">
      <div><strong>${selected.matters}</strong><span>历史事项</span></div>
      <div><strong>${selected.completeness}</strong><span>资料完整</span></div>
      <div><strong>${selected.risk}</strong><span>风险等级</span></div>
    </div>
    <strong>关联方</strong>
    <div class="relationship-list">
      ${selected.relations.map((relation) => `<span>${relation}<em>已纳入冲突检索</em></span>`).join("")}
    </div>
  `;

  document.querySelectorAll("[data-client-index]").forEach((node) => {
    node.addEventListener("click", () => renderClients(clients[Number(node.dataset.clientIndex)]));
  });
}

function renderConflicts() {
  $("#conflictResults").innerHTML = conflicts
    .map(
      (item) => `
        <article class="conflict-card">
          <div class="conflict-card-head">
            <div>
              ${statusPill(item.level, item.cls)}
              <h3>${item.title}</h3>
              <p>${item.summary}</p>
            </div>
            <strong>${item.action}</strong>
          </div>
          <div class="conflict-evidence">
            ${item.evidence.map((text) => `<div>${text}</div>`).join("")}
          </div>
          <div class="conflict-actions">
            <button class="ghost-button">查看证据链</button>
            <button class="ghost-button">分配复核人</button>
            <button class="primary-button">生成处理意见</button>
          </div>
        </article>
      `,
    )
    .join("");
}

function switchView(view) {
  document.querySelectorAll(".view").forEach((node) => node.classList.remove("active"));
  document.querySelectorAll(".nav-item").forEach((node) => node.classList.remove("active"));
  $(`#${view}View`).classList.add("active");
  $(`[data-view="${view}"]`).classList.add("active");
  $("#pageTitle").textContent = pageTitles[view];
}

function openDrawer(title = "新建案件") {
  $("#drawerTitle").textContent = title;
  $("#drawer").classList.add("open");
  $("#drawer").setAttribute("aria-hidden", "false");
}

function closeDrawer() {
  $("#drawer").classList.remove("open");
  $("#drawer").setAttribute("aria-hidden", "true");
}

document.querySelectorAll(".nav-item").forEach((button) => {
  button.addEventListener("click", () => switchView(button.dataset.view));
});

document.querySelectorAll("#quickCaseBtn, #caseCreateBtn").forEach((button) => {
  button.addEventListener("click", () => openDrawer("新建案件"));
});

$("#clientCreateBtn").addEventListener("click", () => openDrawer("新增客户"));

document.querySelectorAll("[data-close='drawer']").forEach((button) => {
  button.addEventListener("click", closeDrawer);
});

$("#conflictForm").addEventListener("submit", (event) => {
  event.preventDefault();
  renderConflicts();
});

renderTasks();
renderCases();
renderClients();
renderConflicts();
