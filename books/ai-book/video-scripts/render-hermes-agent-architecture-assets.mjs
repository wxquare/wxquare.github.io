import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(__dirname, "../../..");
const outRoot = resolve(repoRoot, "books/ai-book/video-assets/16-hermes-agent-architecture-5min");
const args = process.argv.slice(2);
const audioIndex = args.indexOf("--audio");
const audioPath = audioIndex >= 0 && args[audioIndex + 1] ? resolve(args[audioIndex + 1]) : "";
const svgDir = resolve(outRoot, "svg");
const pngDir = resolve(outRoot, "png");

mkdirSync(svgDir, { recursive: true });
mkdirSync(pngDir, { recursive: true });

const W = 1080;
const H = 1920;
const font = "-apple-system, BlinkMacSystemFont, 'PingFang SC', 'Noto Sans CJK SC', 'Microsoft YaHei', sans-serif";

const shots = [
  { id: "01", duration: 15, label: "Hook", file: "agent-not-chat-entry", title: ["Agent 不是", "多一个聊天入口"], subtitle: "它应该是长期运行的个人 Runtime", footer: "会记忆、会沉淀技能、会跨入口工作", type: "hook" },
  { id: "02", duration: 22, label: "Position", file: "openclaw-vs-hermes", title: ["OpenClaw 到达用户", "Hermes 长期成长"], subtitle: "Gateway 解决入口，Runtime 解决能力积累", footer: "OpenClaw 解决到达，Hermes 解决成长", type: "compare" },
  { id: "03", duration: 20, label: "Runtime", file: "long-running-runtime", title: ["Hermes =", "Long-running Agent Runtime"], subtitle: "Memory + Skills + Gateway + Tools + Backends", footer: "不是聊天机器人，而是可扩展运行时", type: "runtime" },
  { id: "04", duration: 20, label: "Thesis", file: "ability-system", title: ["Agent 能力", "不只来自模型"], subtitle: "模型、记忆、技能、工具、入口和轨迹共同组成系统", footer: "参数不是全部，Runtime 才让能力持续增长", type: "thesis" },
  { id: "05", duration: 22, label: "Six Parts", file: "six-components", title: ["六大组件", "构成长运行 Agent"], subtitle: "大脑 / 记忆 / 小脑 / 工具 / 执行 / 环境", footer: "先看分层，再看数据流", type: "six" },
  { id: "06", duration: 20, label: "Brain", file: "brain-core", title: ["大脑中枢", "LLM + Prompt + Provider"], subtitle: "理解意图、生成推理、选择工具", footer: "Prompt Builder 决定模型看到什么", type: "brain" },
  { id: "07", duration: 20, label: "Memory", file: "memory-system", title: ["记忆系统", "让 Agent 连续存在"], subtitle: "Memory / User Profile / Session Search / Skills / Profiles", footer: "记忆不是越多越好，而是边界清楚", type: "memory" },
  { id: "08", duration: 20, label: "Tools", file: "tool-runtime", title: ["工具中心", "决定 Agent 能做什么"], subtitle: "Tool Registry / Toolsets / MCP / Plugins / Skills", footer: "Toolsets 是权限治理的基本单位", type: "tools" },
  { id: "09", duration: 20, label: "Execution", file: "action-engine", title: ["执行引擎", "把工具调用变成真实动作"], subtitle: "校验参数、调度后端、回传结果、处理失败", footer: "没有执行引擎，工具调用只是 demo", type: "execution" },
  { id: "10", duration: 22, label: "Flow", file: "task-data-flow", title: ["一次任务", "不是 Prompt -> Answer"], subtitle: "入口 -> 上下文 -> 工具 -> 观察 -> 回流", footer: "Memory 同时参与输入和输出后的沉淀", type: "flow" },
  { id: "11", duration: 22, label: "Context", file: "prompt-system", title: ["Prompt System", "稳定上下文而不是动态拼贴"], subtitle: "稳定前缀 / 按需召回 / 本轮运行态", footer: "连续性和可控性必须同时存在", type: "context" },
  { id: "12", duration: 22, label: "Skills", file: "skills-memory", title: ["Skills", "把经验变成程序性记忆"], subtitle: "Memory = 知道什么，Skill = 下次怎么做", footer: "可验证、可复用、可审查，才值得沉淀", type: "skills" },
  { id: "13", duration: 20, label: "Gateway", file: "gateway-toolsets", title: ["Gateway + Toolsets", "长期在线但不能无限授权"], subtitle: "多入口接入，能力按 profile 和任务分组", footer: "跨平台工作必须配合权限边界", type: "gateway" },
  { id: "14", duration: 20, label: "Risk", file: "learning-risk", title: ["自我进化", "必须被验证约束"], subtitle: "错误经验、隐私泄露、跨 profile 串线、高风险授权", footer: "长期 Agent 需要评测、审核和审计", type: "risk" },
  { id: "15", duration: 22, label: "MVP", file: "minimum-architecture", title: ["自研 Agent", "最小可行架构"], subtitle: "统一入口 / 稳定上下文 / 受控工具 / 可审计执行 / 学习闭环", footer: "不要只接模型 API，要设计 Runtime", type: "mvp" },
  { id: "16", duration: 18, label: "Takeaway", file: "runtime-loop", title: ["真正的 Agent 能力", "来自模型 + Runtime + 轨迹闭环"], subtitle: "长期运行，持续沉淀，受验证约束", footer: "Hermes 的启示：把 Agent 当系统设计", type: "takeaway" },
];

function esc(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function text(x, y, content, size, weight = 500, color = "#f7fbff", anchor = "middle") {
  return `<text x="${x}" y="${y}" text-anchor="${anchor}" font-family="${font}" font-size="${size}" font-weight="${weight}" fill="${color}">${esc(content)}</text>`;
}

function rect(x, y, w, h, rx, fill, stroke = "none", sw = 0, opacity = 1) {
  return `<rect x="${x}" y="${y}" width="${w}" height="${h}" rx="${rx}" fill="${fill}" stroke="${stroke}" stroke-width="${sw}" opacity="${opacity}"/>`;
}

function line(x1, y1, x2, y2, color = "#60a5fa", width = 4, opacity = 1) {
  return `<line x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" stroke="${color}" stroke-width="${width}" stroke-linecap="round" opacity="${opacity}"/>`;
}

function circle(cx, cy, r, fill, stroke = "none", sw = 0, opacity = 1) {
  return `<circle cx="${cx}" cy="${cy}" r="${r}" fill="${fill}" stroke="${stroke}" stroke-width="${sw}" opacity="${opacity}"/>`;
}

function header(shot) {
  return `
    ${rect(72, 72, 290, 52, 26, "#0b1220", "#14b8a6", 2, 0.95)}
    ${text(217, 108, `${shot.id} / ${shot.label}`, 24, 700, "#99f6e4")}
  `;
}

function titleBlock(shot, y = 250) {
  const [a, b] = shot.title;
  return `
    ${text(W / 2, y, a, a.length > 16 ? 56 : 70, 850)}
    ${text(W / 2, y + 92, b, b.length > 18 ? 48 : 62, 850, "#a7f3d0")}
    ${text(W / 2, y + 164, shot.subtitle, shot.subtitle.length > 28 ? 28 : 32, 600, "#dbeafe")}
  `;
}

function footer(shot) {
  return `
    ${rect(72, 1672, 936, 128, 28, "#0b1220", "#0f766e", 2, 0.90)}
    ${text(W / 2, 1748, shot.footer, shot.footer.length > 25 ? 28 : 34, 750, "#f8fafc")}
  `;
}

function standardBackground() {
  const grid = [];
  for (let x = 80; x < W; x += 100) grid.push(line(x, 0, x, H, "#0f766e", 1, 0.08));
  for (let y = 140; y < H; y += 100) grid.push(line(0, y, W, y, "#0f766e", 1, 0.08));
  return `
    <defs>
      <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#030712"/>
        <stop offset="45%" stop-color="#0f172a"/>
        <stop offset="100%" stop-color="#042f2e"/>
      </linearGradient>
      <radialGradient id="glow" cx="50%" cy="35%" r="62%">
        <stop offset="0%" stop-color="#14b8a6" stop-opacity="0.32"/>
        <stop offset="52%" stop-color="#0f766e" stop-opacity="0.10"/>
        <stop offset="100%" stop-color="#020617" stop-opacity="0"/>
      </radialGradient>
    </defs>
    ${rect(0, 0, W, H, 0, "url(#bg)")}
    ${rect(0, 0, W, H, 0, "url(#glow)")}
    ${grid.join("\n")}
  `;
}

function chip(x, y, label, accent = "#14b8a6", w = 250) {
  return `
    ${rect(x, y, w, 76, 22, "#020617", accent, 2, 0.82)}
    ${text(x + w / 2, y + 49, label, label.length > 12 ? 24 : 30, 800, "#f8fafc")}
  `;
}

function moduleCard(x, y, title, body, accent = "#14b8a6", w = 292, h = 164) {
  const bodyY = h < 130 ? y + 76 : y + 118;
  return `
    ${rect(x, y, w, h, 26, "#0b1220", accent, 2, 0.94)}
    ${circle(x + 50, y + 54, 21, accent, "none", 0, 0.95)}
    ${text(x + 88, y + 62, title, title.length > 10 ? 25 : 30, 850, "#f8fafc", "start")}
    ${text(x + w / 2, bodyY, body, body.length > 12 ? 22 : 24, 600, "#cbd5e1")}
  `;
}

function arrow(x1, y1, x2, y2, color = "#60a5fa") {
  const head = x2 >= x1 ? -1 : 1;
  return `
    ${line(x1, y1, x2, y2, color, 4, 0.92)}
    ${line(x2, y2, x2 + head * 18, y2 - 12, color, 4, 0.92)}
    ${line(x2, y2, x2 + head * 18, y2 + 12, color, 4, 0.92)}
  `;
}

function drawHook() {
  return `
    ${circle(W / 2, 1010, 150, "#0f766e", "#99f6e4", 4, 0.95)}
    ${text(W / 2, 1000, "Agent", 58, 900)}
    ${text(W / 2, 1064, "Runtime", 36, 850, "#ccfbf1")}
    ${["CLI", "Slack", "Telegram", "IDE", "Cron"].map((m, i) => {
      const x = 110 + i * 172;
      return `${chip(x, 780 + (i % 2) * 440, m, "#60a5fa", 138)}${line(x + 69, 855 + (i % 2) * 440, W / 2, 1010, "#60a5fa", 3, 0.55)}`;
    }).join("\n")}
  `;
}

function drawCompare() {
  return `
    ${rect(96, 760, 408, 520, 34, "#111827", "#60a5fa", 3, 0.94)}
    ${text(300, 850, "OpenClaw", 46, 900, "#bfdbfe")}
    ${text(300, 930, "Agent Gateway", 30, 800, "#dbeafe")}
    ${chip(155, 1030, "多入口接入", "#60a5fa", 290)}
    ${chip(155, 1140, "到达用户", "#60a5fa", 290)}
    ${rect(576, 760, 408, 520, 34, "#052e2b", "#14b8a6", 3, 0.96)}
    ${text(780, 850, "Hermes", 46, 900, "#99f6e4")}
    ${text(780, 930, "Agent Runtime", 30, 800, "#ccfbf1")}
    ${chip(635, 1030, "长期记忆", "#14b8a6", 290)}
    ${chip(635, 1140, "能力成长", "#14b8a6", 290)}
  `;
}

function drawRuntime() {
  const modules = ["Memory", "Skills", "Gateway", "Tools", "Backends", "Trajectories"];
  return `
    ${circle(W / 2, 1010, 145, "#0f766e", "#99f6e4", 4, 0.96)}
    ${text(W / 2, 990, "Hermes", 48, 900)}
    ${text(W / 2, 1050, "Runtime", 36, 850, "#ccfbf1")}
    ${modules.map((m, i) => {
      const angle = (-90 + i * 60) * Math.PI / 180;
      const x = W / 2 + Math.cos(angle) * 335;
      const y = 1010 + Math.sin(angle) * 335;
      return `${line(W / 2, 1010, x, y, "#14b8a6", 3, 0.55)}${circle(x, y, 76, "#0b1220", "#14b8a6", 3, 0.95)}${text(x, y + 9, m, m.length > 8 ? 23 : 28, 850)}`;
    }).join("\n")}
  `;
}

function drawThesis() {
  return `
    ${moduleCard(132, 840, "Model", "理解与生成", "#60a5fa", 300)}
    ${moduleCard(648, 840, "Runtime", "记忆与执行", "#14b8a6", 300)}
    ${arrow(440, 922, 640, 922)}
    ${rect(156, 1160, 768, 150, 32, "#0b1220", "#facc15", 2, 0.92)}
    ${text(W / 2, 1222, "能力 = 模型 + 上下文 + 工具 + 轨迹", 38, 900, "#fef3c7")}
    ${text(W / 2, 1284, "不是只换更大的模型", 30, 750, "#fef9c3")}
  `;
}

function drawSix() {
  const modules = [
    ["大脑", "推理核心"], ["记忆", "长期上下文"], ["小脑", "规划状态"],
    ["工具", "能力注册"], ["执行", "调度回传"], ["环境", "真实世界"],
  ];
  return modules.map(([a, b], i) => moduleCard(86 + (i % 3) * 306, 762 + Math.floor(i / 3) * 220, a, b)).join("\n");
}

function drawBrain() {
  return `
    ${rect(120, 790, 840, 420, 36, "#0b1220", "#60a5fa", 3, 0.94)}
    ${text(W / 2, 880, "Prompt Builder", 46, 900, "#bfdbfe")}
    ${["Persona", "Memory", "Skills", "Tools", "Session"].map((m, i) => chip(168 + (i % 2) * 378, 960 + Math.floor(i / 2) * 96, m, "#60a5fa", i === 4 ? 742 : 330)).join("\n")}
    ${arrow(W / 2, 1218, W / 2, 1348, "#14b8a6")}
    ${circle(W / 2, 1435, 82, "#0f766e", "#99f6e4", 3, 0.95)}
    ${text(W / 2, 1448, "LLM", 42, 900)}
  `;
}

function drawMemory() {
  return `
    ${moduleCard(120, 760, "Persistent Memory", "长期事实", "#14b8a6", 840, 138)}
    ${moduleCard(120, 930, "User Profile", "偏好、角色、习惯", "#14b8a6", 840, 138)}
    ${moduleCard(120, 1100, "Session Search", "历史细节按需检索", "#60a5fa", 840, 138)}
    ${moduleCard(120, 1270, "Skills / Profiles", "做法沉淀与身份隔离", "#60a5fa", 840, 138)}
  `;
}

function drawTools() {
  return `
    ${moduleCard(108, 790, "Tool Registry", "schema / 可用性", "#14b8a6", 864, 150)}
    ${arrow(W / 2, 955, W / 2, 1044)}
    ${rect(108, 1060, 864, 300, 32, "#0b1220", "#60a5fa", 2, 0.94)}
    ${text(W / 2, 1130, "Toolsets", 42, 900, "#bfdbfe")}
    ${["web", "terminal", "file", "browser", "memory", "MCP"].map((m, i) => chip(150 + (i % 3) * 260, 1190 + Math.floor(i / 3) * 92, m, "#60a5fa", 210)).join("\n")}
  `;
}

function drawExecution() {
  return `
    ${moduleCard(150, 800, "Tool Call", "模型输出", "#14b8a6", 780, 118)}
    ${arrow(W / 2, 930, W / 2, 990, "#14b8a6")}
    ${moduleCard(150, 1010, "Validate", "参数校验", "#14b8a6", 780, 118)}
    ${arrow(W / 2, 1140, W / 2, 1200, "#14b8a6")}
    ${moduleCard(150, 1220, "Backend", "执行环境", "#facc15", 780, 118)}
    ${arrow(W / 2, 1350, W / 2, 1410, "#14b8a6")}
    ${moduleCard(150, 1430, "Observe", "结果回传", "#14b8a6", 780, 118)}
  `;
}

function drawFlow() {
  const items = [
    ["入口", "平台事件"],
    ["上下文", "Prompt Builder"],
    ["工具", "Tool Runtime"],
    ["观察", "结构化结果"],
    ["回流", "Memory / Skill"],
  ];
  return `
    ${items.map(([a, b], i) => {
      const y = 745 + i * 158;
      return `${moduleCard(170, y, a, b, i === 4 ? "#facc15" : "#14b8a6", 740, 104)}${i < 4 ? line(W / 2, y + 112, W / 2, y + 148, "#60a5fa", 4, 0.75) : ""}`;
    }).join("\n")}
  `;
}

function drawContext() {
  return `
    ${moduleCard(130, 780, "稳定前缀", "Persona / Profile / Memory", "#14b8a6", 820, 150)}
    ${moduleCard(130, 990, "按需召回", "Skills / Context Files / Sessions", "#60a5fa", 820, 150)}
    ${moduleCard(130, 1200, "本轮运行态", "当前任务 / 工具观察结果", "#facc15", 820, 150)}
  `;
}

function drawSkills() {
  return `
    ${chip(110, 800, "Task Trace", "#60a5fa", 260)}
    ${arrow(380, 838, 505, 838)}
    ${chip(512, 800, "Reflect", "#14b8a6", 220)}
    ${arrow(742, 838, 870, 838)}
    ${chip(782, 940, "SKILL.md", "#facc15", 220)}
    ${rect(122, 1080, 836, 330, 34, "#0b1220", "#14b8a6", 2, 0.94)}
    ${text(174, 1160, "trigger: 发布前检查", 30, 800, "#ccfbf1", "start")}
    ${text(174, 1230, "steps: 搜索 -> 测试 -> diff -> 风险摘要", 30, 800, "#ccfbf1", "start")}
    ${text(174, 1300, "verification: run tests + review", 30, 800, "#ccfbf1", "start")}
  `;
}

function drawGateway() {
  return `
    ${["CLI", "Telegram", "Slack", "IDE", "Cron"].map((m, i) => chip(92 + (i % 3) * 306, 770 + Math.floor(i / 3) * 110, m, "#60a5fa", 250)).join("\n")}
    ${arrow(W / 2, 1020, W / 2, 1140, "#14b8a6")}
    ${moduleCard(150, 1160, "Toolsets", "按入口、profile、任务分组授权", "#14b8a6", 780, 170)}
    ${text(W / 2, 1430, "长期在线 ≠ 无限授权", 46, 900, "#fef3c7")}
  `;
}

function drawRisk() {
  const risks = [["错误经验", "自动固化"], ["隐私数据", "写入记忆"], ["Profile", "信息串线"], ["危险工具", "错误授权"]];
  return risks.map(([a, b], i) => moduleCard(150 + (i % 2) * 390, 780 + Math.floor(i / 2) * 230, a, b, "#ef4444", 300, 164)).join("\n") +
    text(W / 2, 1390, "验证 / 评测 / 审核 / 审计", 42, 900, "#fecaca");
}

function drawMvp() {
  const items = ["统一入口", "稳定上下文", "受控工具集", "可审计执行", "学习闭环"];
  return items.map((m, i) => moduleCard(170, 760 + i * 132, `${i + 1}. ${m}`, ["Intake", "Context", "Toolsets", "Trace", "Eval + Review"][i], i === 4 ? "#facc15" : "#14b8a6", 740, 104)).join("\n");
}

function drawTakeaway() {
  return `
    ${moduleCard(130, 820, "Model", "理解、生成、规划", "#60a5fa", 820, 130)}
    ${arrow(W / 2, 970, W / 2, 1040, "#14b8a6")}
    ${moduleCard(130, 1060, "Runtime", "记忆、工具、执行、权限", "#14b8a6", 820, 130)}
    ${arrow(W / 2, 1210, W / 2, 1280, "#facc15")}
    ${moduleCard(130, 1300, "Trajectory Loop", "评估、沉淀、改进", "#facc15", 820, 130)}
  `;
}

function visual(shot) {
  switch (shot.type) {
    case "hook": return drawHook();
    case "compare": return drawCompare();
    case "runtime": return drawRuntime();
    case "thesis": return drawThesis();
    case "six": return drawSix();
    case "brain": return drawBrain();
    case "memory": return drawMemory();
    case "tools": return drawTools();
    case "execution": return drawExecution();
    case "flow": return drawFlow();
    case "context": return drawContext();
    case "skills": return drawSkills();
    case "gateway": return drawGateway();
    case "risk": return drawRisk();
    case "mvp": return drawMvp();
    case "takeaway": return drawTakeaway();
    default: return "";
  }
}

function svg(shot) {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">
  ${standardBackground()}
  ${header(shot)}
  ${titleBlock(shot)}
  ${visual(shot)}
  ${footer(shot)}
</svg>
`;
}

function frames(seconds) {
  return Math.round(seconds * 30);
}

function timecode(frame) {
  const sec = Math.floor(frame / 30);
  const f = frame % 30;
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}.${String(f).padStart(3, "0")}`;
}

function buildMltProject() {
  let current = 0;
  const producers = [];
  const entries = [];

  shots.forEach((shot, i) => {
    const id = `producer${i}`;
    const len = frames(shot.duration);
    const out = len - 1;
    const pngPath = resolve(pngDir, `${shot.id}-${shot.file}.png`);
    producers.push(`
  <producer id="${id}" in="0" out="${out}">
    <property name="length">${timecode(len)}</property>
    <property name="eof">pause</property>
    <property name="resource">${esc(pngPath)}</property>
    <property name="ttl">1</property>
    <property name="aspect_ratio">1</property>
    <property name="mlt_service">qimage</property>
  </producer>`);
    entries.push(`    <entry producer="${id}" in="0" out="${out}"/>`);
    current += len;
  });

  const audioProducer = audioPath ? `
  <producer id="voiceover" in="0" out="${Math.max(current - 1, 1)}">
    <property name="resource">${esc(audioPath)}</property>
    <property name="mlt_service">avformat-novalidate</property>
  </producer>` : "";
  const audioPlaylist = audioPath ? `
  <playlist id="playlist_audio">
    <entry producer="voiceover" in="0" out="${Math.max(current - 1, 1)}"/>
  </playlist>` : "";
  const audioTrack = audioPath ? `    <track producer="playlist_audio"/>` : "";

  return `<?xml version="1.0" standalone="no"?>
<mlt LC_NUMERIC="C" version="7.0.0" title="Hermes Agent Architecture 5min" producer="tractor0">
  <profile description="vertical_1080x1920_30fps" width="1080" height="1920" progressive="1" sample_aspect_num="1" sample_aspect_den="1" display_aspect_num="9" display_aspect_den="16" frame_rate_num="30" frame_rate_den="1" colorspace="709"/>
${producers.join("\n")}
${audioProducer}
  <playlist id="playlist_video">
${entries.join("\n")}
  </playlist>
${audioPlaylist}
  <tractor id="tractor0" in="0" out="${Math.max(current - 1, 1)}">
    <track producer="playlist_video"/>
${audioTrack}
  </tractor>
</mlt>
`;
}

function writeShotList() {
  let start = 0;
  const rows = ["id,file,duration_seconds,start_seconds,end_seconds,label,title,subtitle"];
  for (const shot of shots) {
    rows.push([
      shot.id,
      `${shot.id}-${shot.file}.png`,
      shot.duration,
      start,
      start + shot.duration,
      shot.label,
      shot.title.join(" "),
      shot.subtitle,
    ].map((v) => `"${String(v).replaceAll('"', '""')}"`).join(","));
    start += shot.duration;
  }
  writeFileSync(resolve(outRoot, "shot-list.csv"), `${rows.join("\n")}\n`, "utf8");
}

function writeReadme() {
  const total = shots.reduce((sum, shot) => sum + shot.duration, 0);
  const projectFile = audioPath ? "hermes-agent-architecture-with-audio.mlt" : "hermes-agent-architecture-image-only.mlt";
  const kdenliveFile = projectFile.replace(/\.mlt$/, ".kdenlive");
  writeFileSync(resolve(outRoot, "README.md"), `# Hermes Agent 架构解析 5 分钟视频素材

生成时间：${new Date().toISOString()}

## 文件结构

\`\`\`text
svg/           16 张可编辑 SVG 图文页
png/           16 张 1080x1920 PNG 图文页
voiceover/     配音文本和 WAV 音频
${projectFile}   MLT 项目文件
${kdenliveFile}  Kdenlive 项目文件
shot-list.csv 分镜时长表
\`\`\`

## 建议用法

1. 用 Kdenlive 打开 \`${kdenliveFile}\`，打不开时再打开 \`${projectFile}\`。
2. 每张图已经按分镜时长排列，总时长约 ${total} 秒。
3. 如需改字，修改 \`books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs\` 后重新运行。

\`\`\`bash
node books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs
node books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs --audio books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover-5min.wav
\`\`\`
`, "utf8");
}

for (const shot of shots) {
  const svgPath = resolve(svgDir, `${shot.id}-${shot.file}.svg`);
  const pngPath = resolve(pngDir, `${shot.id}-${shot.file}.png`);
  writeFileSync(svgPath, svg(shot), "utf8");
  execFileSync("rsvg-convert", ["-w", String(W), "-h", String(H), svgPath, "-o", pngPath], { stdio: "inherit" });
}

writeShotList();
writeReadme();

const projectFile = audioPath ? "hermes-agent-architecture-with-audio.mlt" : "hermes-agent-architecture-image-only.mlt";
const projectXml = buildMltProject();
writeFileSync(resolve(outRoot, projectFile), projectXml, "utf8");
writeFileSync(resolve(outRoot, projectFile.replace(/\.mlt$/, ".kdenlive")), projectXml, "utf8");

if (audioPath && !existsSync(audioPath)) {
  console.warn(`Audio path was written into the project but does not exist yet: ${audioPath}`);
}

console.log(`Generated ${shots.length} SVG files and ${shots.length} PNG files in ${outRoot}`);
console.log(`Kdenlive/MLT project: ${resolve(outRoot, projectFile)}`);
