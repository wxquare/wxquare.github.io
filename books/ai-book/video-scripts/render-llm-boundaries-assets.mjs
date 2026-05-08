import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(__dirname, "../../..");
const outRoot = resolve(repoRoot, "books/ai-book/video-assets/01-llm-boundaries-3min");
const aiBgDir = resolve(outRoot, "ai-generated/backgrounds-1080p");

const args = process.argv.slice(2);
const audioIndex = args.indexOf("--audio");
const audioPath = audioIndex >= 0 && args[audioIndex + 1] ? resolve(args[audioIndex + 1]) : "";
const useAiBackgrounds = args.includes("--ai-bg");
const svgDir = resolve(outRoot, useAiBackgrounds ? "svg-ai" : "svg");
const pngDir = resolve(outRoot, useAiBackgrounds ? "png-ai" : "png");

mkdirSync(svgDir, { recursive: true });
mkdirSync(pngDir, { recursive: true });

const W = 1080;
const H = 1920;
const font = "-apple-system, BlinkMacSystemFont, 'PingFang SC', 'Noto Sans CJK SC', 'Microsoft YaHei', sans-serif";

const shots = [
  {
    id: "01",
    file: "01-deterministic-boundary",
    duration: 10,
    label: "Hook",
    title: ["别把 LLM", "当确定性系统"],
    subtitle: "算钱 / 查实时数据 / 执行线上命令",
    footer: "模型很强，但边界错了就会出问题",
    type: "warning",
  },
  {
    id: "02",
    file: "02-not-database-calculator-policy",
    duration: 15,
    label: "Boundary",
    title: ["LLM 不是", "数据库 / 计算器 / 权限系统"],
    subtitle: "它不是万能后端，而是智能组件",
    footer: "不要让概率模型承担确定性职责",
    type: "notSystem",
  },
  {
    id: "03",
    file: "03-probability-generator",
    duration: 17,
    label: "Model Nature",
    title: ["LLM =", "上下文中的概率生成器"],
    subtitle: "根据当前上下文生成最可能的下一个 token",
    footer: "生成看起来合理，不代表一定真实",
    type: "tokens",
  },
  {
    id: "04",
    file: "04-good-at-semantic-tasks",
    duration: 23,
    label: "Strengths",
    title: ["它擅长", "语义清晰的任务"],
    subtitle: "总结、改写、分类、抽取、代码片段、规划",
    footer: "答案空间越可约束，越容易工程化",
    type: "strengths",
  },
  {
    id: "05",
    file: "05-feedback-json-example",
    duration: 20,
    label: "Example",
    title: ["客户反馈分类", "适合 LLM"],
    subtitle: "标签有限、输入短、输出可用 Schema 约束",
    footer: "分类不是靠感觉，而是靠标签、样本和评测",
    type: "json",
  },
  {
    id: "06",
    file: "06-weaknesses",
    duration: 20,
    label: "Limits",
    title: ["它不擅长", "确定性与高风险任务"],
    subtitle: "精确计算、实时事实、长期记忆、高风险执行",
    footer: "这些能力应该交给系统，而不是模型记忆",
    type: "weaknesses",
  },
  {
    id: "07",
    file: "07-compound-interest-tool",
    duration: 20,
    label: "Tool Use",
    title: ["复利计算", "模型选公式，工具算结果"],
    subtitle: "FV = P × (1 + r)^n",
    footer: "LLM 负责理解，Calculator 负责计算",
    type: "compound",
  },
  {
    id: "08",
    file: "08-agent-engineering-system",
    duration: 23,
    label: "Architecture",
    title: ["Agent =", "LLM + 工程系统"],
    subtitle: "模型不是系统，Runtime 才是系统",
    footer: "把模型放进可控、可验证、可审计的架构里",
    type: "agent",
  },
  {
    id: "09",
    file: "09-six-modules",
    duration: 17,
    label: "Runtime",
    title: ["RAG / Tool / Memory", "Policy / Eval / Trace"],
    subtitle: "六个模块把模型变成可工作的 Agent",
    footer: "检索、执行、记忆、权限、验证、审计",
    type: "sixModules",
  },
  {
    id: "10",
    file: "10-first-principle",
    duration: 15,
    label: "Takeaway",
    title: ["成熟 AI 架构", "知道什么时候不让模型单独决定"],
    subtitle: "让 LLM 做理解、生成、归纳和规划",
    footer: "让工程系统做计算、检索、执行、权限和验证",
    type: "takeaway",
  },
];

function esc(s) {
  return String(s)
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
    ${rect(72, 72, 268, 52, 26, "#0b1220", "#2563eb", 2, 0.95)}
    ${text(206, 108, `${shot.id} / ${shot.label}`, 24, 700, "#93c5fd")}
  `;
}

function titleBlock(shot, y = 260) {
  const [a, b] = shot.title;
  return `
    ${text(W / 2, y, a, 72, 800)}
    ${text(W / 2, y + 92, b, b.length > 18 ? 54 : 66, 800, "#bfdbfe")}
    ${text(W / 2, y + 164, shot.subtitle, 32, 500, "#dbeafe")}
  `;
}

function footer(shot) {
  return `
    ${rect(72, 1672, 936, 128, 34, "#0b1220", "#1d4ed8", 2, 0.88)}
    ${text(W / 2, 1748, shot.footer, shot.footer.length > 23 ? 30 : 34, 700, "#f8fafc")}
  `;
}

function standardBackground() {
  const grid = [];
  for (let x = 80; x < W; x += 100) grid.push(line(x, 0, x, H, "#1d4ed8", 1, 0.08));
  for (let y = 140; y < H; y += 100) grid.push(line(0, y, W, y, "#1d4ed8", 1, 0.08));
  return `
    <defs>
      <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#050816"/>
        <stop offset="48%" stop-color="#0f172a"/>
        <stop offset="100%" stop-color="#111827"/>
      </linearGradient>
      <radialGradient id="glow" cx="50%" cy="34%" r="60%">
        <stop offset="0%" stop-color="#2563eb" stop-opacity="0.38"/>
        <stop offset="45%" stop-color="#1d4ed8" stop-opacity="0.10"/>
        <stop offset="100%" stop-color="#020617" stop-opacity="0"/>
      </radialGradient>
      <filter id="soft">
        <feGaussianBlur stdDeviation="14"/>
      </filter>
    </defs>
    ${rect(0, 0, W, H, 0, "url(#bg)")}
    ${rect(0, 0, W, H, 0, "url(#glow)")}
    ${grid.join("\n")}
  `;
}

function aiBackground(shot) {
  const bgPath = resolve(aiBgDir, `${shot.file}-bg.png`);
  if (!existsSync(bgPath)) return standardBackground();
  const bgData = readFileSync(bgPath).toString("base64");
  return `
    <defs>
      <linearGradient id="aiTopShade" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#020617" stop-opacity="0.92"/>
        <stop offset="42%" stop-color="#020617" stop-opacity="0.62"/>
        <stop offset="100%" stop-color="#020617" stop-opacity="0"/>
      </linearGradient>
      <linearGradient id="aiBottomShade" x1="0" y1="1" x2="0" y2="0">
        <stop offset="0%" stop-color="#020617" stop-opacity="0.94"/>
        <stop offset="55%" stop-color="#020617" stop-opacity="0.70"/>
        <stop offset="100%" stop-color="#020617" stop-opacity="0"/>
      </linearGradient>
      <radialGradient id="aiCenterVignette" cx="50%" cy="50%" r="75%">
        <stop offset="0%" stop-color="#020617" stop-opacity="0.06"/>
        <stop offset="72%" stop-color="#020617" stop-opacity="0.04"/>
        <stop offset="100%" stop-color="#020617" stop-opacity="0.36"/>
      </radialGradient>
    </defs>
    <image href="data:image/png;base64,${bgData}" x="0" y="0" width="${W}" height="${H}" preserveAspectRatio="xMidYMid slice"/>
    ${rect(0, 0, W, H, 0, "#020617", "none", 0, 0.08)}
    ${rect(0, 0, W, 760, 0, "url(#aiTopShade)")}
    ${rect(0, 1280, W, 640, 0, "url(#aiBottomShade)")}
    ${rect(0, 0, W, H, 0, "url(#aiCenterVignette)")}
  `;
}

function background(shot) {
  return useAiBackgrounds ? aiBackground(shot) : standardBackground();
}

function iconCard(x, y, title, body, accent = "#60a5fa") {
  return `
    ${rect(x, y, 286, 168, 28, "#0b1220", accent, 2, 0.94)}
    ${circle(x + 52, y + 56, 22, accent, "none", 0, 0.95)}
    ${text(x + 92, y + 62, title, 30, 800, "#f8fafc", "start")}
    ${text(x + 143, y + 120, body, 24, 500, "#cbd5e1")}
  `;
}

function drawWarning() {
  return `
    ${rect(160, 760, 760, 480, 46, "#111827", "#ef4444", 4, 0.92)}
    ${text(W / 2, 870, "calculate money", 42, 700, "#fecaca")}
    ${text(W / 2, 990, "query live price", 42, 700, "#fecaca")}
    ${text(W / 2, 1110, "restart service", 42, 700, "#fecaca")}
    ${line(274, 835, 806, 1145, "#ef4444", 12, 0.86)}
    ${line(806, 835, 274, 1145, "#ef4444", 12, 0.86)}
  `;
}

function drawNotSystem() {
  return `
    ${circle(W / 2, 928, 150, "#1d4ed8", "#93c5fd", 4, 0.96)}
    ${text(W / 2, 948, "LLM", 72, 900)}
    ${iconCard(90, 1200, "数据库", "实时事实", "#f97316")}
    ${iconCard(397, 1200, "计算器", "精确计算", "#f97316")}
    ${iconCard(704, 1200, "权限系统", "高风险执行", "#f97316")}
  `;
}

function drawTokens() {
  const tokens = ["LLM", "生成", "下一个", "最可能", "Token"];
  return `
    ${rect(94, 760, 892, 420, 42, "#0b1220", "#60a5fa", 3, 0.92)}
    ${tokens.map((t, i) => rect(142 + i * 166, 872, 136, 72, 18, "#1e3a8a", "#93c5fd", 2, 0.95)).join("\n")}
    ${tokens.map((t, i) => text(210 + i * 166, 920, t, 26, 800)).join("\n")}
    ${tokens.map((_, i) => rect(142 + i * 166, 1012, 136, 18 + i * 10, 9, "#60a5fa", "none", 0, 0.8)).join("\n")}
    ${text(W / 2, 1120, "概率高 ≠ 事实真", 42, 900, "#fef3c7")}
  `;
}

function drawStrengths() {
  const cards = [
    ["总结", "文档压缩"], ["改写", "表达优化"], ["分类", "标签判断"],
    ["抽取", "结构化"], ["代码", "片段生成"], ["规划", "任务拆解"],
  ];
  return cards.map(([a, b], i) => iconCard(86 + (i % 3) * 306, 762 + Math.floor(i / 3) * 210, a, b)).join("\n");
}

function drawJson() {
  return `
    ${rect(92, 748, 896, 218, 34, "#0b1220", "#60a5fa", 2, 0.94)}
    ${text(142, 824, "用户反馈", 30, 800, "#93c5fd", "start")}
    ${text(142, 890, "订单还没发货，等了 3 天。", 38, 700, "#f8fafc", "start")}
    ${rect(92, 1018, 896, 360, 34, "#020617", "#22c55e", 2, 0.96)}
    ${text(142, 1096, "{", 34, 700, "#bbf7d0", "start")}
    ${text(178, 1162, '"category": "物流",', 34, 700, "#bbf7d0", "start")}
    ${text(178, 1228, '"sentiment": "不满",', 34, 700, "#bbf7d0", "start")}
    ${text(178, 1294, '"urgency": "中"', 34, 700, "#bbf7d0", "start")}
    ${text(142, 1360, "}", 34, 700, "#bbf7d0", "start")}
  `;
}

function drawWeaknesses() {
  const cards = [
    ["精确计算", "交给计算器"], ["实时事实", "交给 API"],
    ["长期记忆", "交给存储"], ["高风险执行", "交给权限"],
  ];
  return cards.map(([a, b], i) => iconCard(170 + (i % 2) * 380, 762 + Math.floor(i / 2) * 230, a, b, "#ef4444")).join("\n");
}

function drawCompound() {
  return `
    ${rect(88, 800, 260, 230, 34, "#0b1220", "#60a5fa", 2, 0.95)}
    ${text(218, 900, "用户问题", 34, 800)}
    ${text(218, 970, "复利 20 年", 30, 700, "#bfdbfe")}
    ${line(360, 916, 470, 916)}
    ${rect(470, 800, 300, 230, 34, "#172554", "#93c5fd", 2, 0.95)}
    ${text(620, 890, "LLM", 50, 900)}
    ${text(620, 970, "识别公式", 30, 700, "#bfdbfe")}
    ${line(782, 916, 894, 916)}
    ${rect(730, 1130, 260, 180, 30, "#052e16", "#22c55e", 2, 0.95)}
    ${text(860, 1210, "Calculator", 34, 800, "#bbf7d0")}
    ${text(860, 1268, "确定性计算", 26, 600, "#bbf7d0")}
    ${line(620, 1035, 810, 1130, "#22c55e")}
    ${text(W / 2, 1410, "FV = P × (1 + r)^n", 48, 900, "#fef3c7")}
  `;
}

function drawAgent() {
  return `
    ${circle(W / 2, 980, 150, "#1d4ed8", "#93c5fd", 4, 0.96)}
    ${text(W / 2, 1000, "LLM", 72, 900)}
    ${["RAG", "Tool", "Memory", "Policy", "Eval", "Trace"].map((m, i) => {
      const angle = (-90 + i * 60) * Math.PI / 180;
      const x = W / 2 + Math.cos(angle) * 330;
      const y = 980 + Math.sin(angle) * 330;
      return `${line(W / 2, 980, x, y, "#60a5fa", 3, 0.72)}${circle(x, y, 78, "#0b1220", "#60a5fa", 3, 0.96)}${text(x, y + 10, m, 30, 850)}`;
    }).join("\n")}
  `;
}

function drawSixModules() {
  const modules = [
    ["RAG", "查外部知识"], ["Tool", "确定性操作"], ["Memory", "保存状态"],
    ["Policy", "权限控制"], ["Eval", "质量验证"], ["Trace", "过程审计"],
  ];
  return modules.map(([a, b], i) => iconCard(86 + (i % 3) * 306, 762 + Math.floor(i / 3) * 210, a, b, "#22c55e")).join("\n");
}

function drawTakeaway() {
  return `
    ${rect(96, 760, 888, 470, 52, "#0b1220", "#60a5fa", 3, 0.94)}
    ${text(W / 2, 880, "模型负责", 48, 900, "#bfdbfe")}
    ${text(W / 2, 950, "理解 / 生成 / 归纳 / 规划", 42, 900)}
    ${line(250, 1030, 830, 1030, "#334155", 4, 1)}
    ${text(W / 2, 1120, "系统负责", 48, 900, "#bbf7d0")}
    ${text(W / 2, 1190, "计算 / 检索 / 执行 / 验证", 42, 900)}
    ${text(W / 2, 1370, "不要迷信模型，要设计边界", 48, 900, "#fef3c7")}
  `;
}

function aiChip(x, y, label, accent = "#60a5fa", w = 250) {
  return `
    ${rect(x, y, w, 74, 24, "#020617", accent, 2, 0.70)}
    ${text(x + w / 2, y + 48, label, label.length > 9 ? 26 : 30, 800, "#f8fafc")}
  `;
}

function aiVisual(shot) {
  switch (shot.type) {
    case "warning":
      return `
        ${aiChip(150, 1180, "calculate money", "#ef4444", 330)}
        ${aiChip(500, 1180, "query live price", "#ef4444", 360)}
        ${aiChip(326, 1280, "restart service", "#ef4444", 430)}
        ${line(250, 1160, 830, 1370, "#ef4444", 8, 0.88)}
        ${line(830, 1160, 250, 1370, "#ef4444", 8, 0.88)}
      `;
    case "notSystem":
      return `
        ${aiChip(120, 1220, "数据库", "#ef4444")}
        ${aiChip(415, 1220, "计算器", "#ef4444")}
        ${aiChip(710, 1220, "权限系统", "#ef4444")}
      `;
    case "tokens":
      return `
        ${["上下文", "Token", "概率", "生成"].map((m, i) => aiChip(110 + i * 220, 1210, m, "#60a5fa", 188)).join("\n")}
        ${text(W / 2, 1348, "看起来合理 ≠ 一定真实", 38, 900, "#fef3c7")}
      `;
    case "strengths":
      return `
        ${["总结", "改写", "分类", "抽取", "代码", "规划"].map((m, i) => aiChip(90 + (i % 3) * 306, 1130 + Math.floor(i / 3) * 100, m, "#22c55e", 250)).join("\n")}
      `;
    case "json":
      return `
        ${rect(118, 1060, 844, 250, 32, "#020617", "#22c55e", 2, 0.72)}
        ${text(170, 1130, '{ "category": "物流",', 34, 800, "#bbf7d0", "start")}
        ${text(170, 1196, '"sentiment": "不满",', 34, 800, "#bbf7d0", "start")}
        ${text(170, 1262, '"urgency": "中" }', 34, 800, "#bbf7d0", "start")}
      `;
    case "weaknesses":
      return `
        ${["精确计算", "实时事实", "长期记忆", "高风险执行"].map((m, i) => aiChip(145 + (i % 2) * 420, 1140 + Math.floor(i / 2) * 110, m, "#ef4444", 370)).join("\n")}
      `;
    case "compound":
      return `
        ${rect(132, 1140, 816, 230, 34, "#020617", "#22c55e", 2, 0.72)}
        ${text(258, 1240, "LLM", 42, 900, "#bfdbfe")}
        ${line(350, 1228, 730, 1228, "#22c55e", 4, 0.9)}
        ${text(820, 1240, "Calculator", 34, 900, "#bbf7d0")}
        ${text(W / 2, 1330, "FV = P × (1 + r)^n", 42, 900, "#fef3c7")}
      `;
    case "agent":
      return `
        ${text(W / 2, 1228, "RAG / Tool / Memory / Policy / Eval / Trace", 34, 900, "#bfdbfe")}
      `;
    case "sixModules":
      return `
        ${["RAG", "Tool", "Memory", "Policy", "Eval", "Trace"].map((m, i) => aiChip(90 + (i % 3) * 306, 1130 + Math.floor(i / 3) * 100, m, "#22c55e", 250)).join("\n")}
      `;
    case "takeaway":
      return `
        ${rect(118, 1110, 844, 250, 34, "#020617", "#60a5fa", 2, 0.72)}
        ${text(W / 2, 1190, "模型负责理解、生成、归纳、规划", 36, 900, "#bfdbfe")}
        ${text(W / 2, 1274, "系统负责计算、检索、执行、验证", 36, 900, "#bbf7d0")}
      `;
    default:
      return "";
  }
}

function visual(shot) {
  if (useAiBackgrounds) return aiVisual(shot);
  switch (shot.type) {
    case "warning": return drawWarning();
    case "notSystem": return drawNotSystem();
    case "tokens": return drawTokens();
    case "strengths": return drawStrengths();
    case "json": return drawJson();
    case "weaknesses": return drawWeaknesses();
    case "compound": return drawCompound();
    case "agent": return drawAgent();
    case "sixModules": return drawSixModules();
    case "takeaway": return drawTakeaway();
    default: return "";
  }
}

function svg(shot) {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">
  ${background(shot)}
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
  const absPngs = [];

  shots.forEach((shot, i) => {
    const id = `producer${i}`;
    const len = frames(shot.duration);
    const out = len - 1;
    const pngPath = resolve(pngDir, `${shot.id}-${shot.file}.png`);
    absPngs.push(pngPath);
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
<mlt LC_NUMERIC="C" version="7.0.0" title="LLM Boundaries 3min" producer="tractor0">
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
  const total = shots.reduce((sum, s) => sum + s.duration, 0);
  const variantPrefix = useAiBackgrounds ? "llm-boundaries-ai-bg" : "llm-boundaries";
  const projectName = audioPath ? `${variantPrefix}-with-audio.mlt` : `${variantPrefix}-image-only.mlt`;
  const kdenliveName = audioPath ? `${variantPrefix}-with-audio.kdenlive` : `${variantPrefix}-image-only.kdenlive`;
  const audioLine = audioPath
    ? `voiceover/   配音文本和 WAV 音频\n${projectName}   MLT 项目文件\n${kdenliveName}   Kdenlive 项目文件`
    : `${projectName}   MLT 项目文件\n${kdenliveName}   Kdenlive 项目文件`;
  const audioUsage = audioPath
    ? `2. 项目已绑定音频轨：\`${audioPath}\`。`
    : "2. 当前项目没有音频轨，请把配音文件拖入时间线。";
  writeFileSync(resolve(outRoot, "README.md"), `# LLM 能力边界 3 分钟视频素材

生成时间：${new Date().toISOString()}

## 文件结构

\`\`\`text
${useAiBackgrounds ? "svg-ai/" : "svg/"}   10 张可编辑 SVG 图文页
${useAiBackgrounds ? "png-ai/" : "png/"}   10 张 1080x1920 PNG 图文页
${audioLine}
shot-list.csv    分镜时长表
\`\`\`

## 建议用法

1. 用 Kdenlive 打开 \`${kdenliveName}\`，打不开时再打开 \`${projectName}\`。
${audioUsage}
3. 每张图已经按分镜时长排列，总时长约 ${total} 秒。
4. 如需改字，编辑 \`svg/*.svg\` 后重新运行：

\`\`\`bash
node books/ai-book/video-scripts/render-llm-boundaries-assets.mjs
\`\`\`

如果要把配音一起写入项目：

\`\`\`bash
node books/ai-book/video-scripts/render-llm-boundaries-assets.mjs --audio /absolute/path/to/voiceover.mp3
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
const projectPrefix = useAiBackgrounds ? "llm-boundaries-ai-bg" : "llm-boundaries";
const projectFile = audioPath ? `${projectPrefix}-with-audio.mlt` : `${projectPrefix}-image-only.mlt`;
const projectXml = buildMltProject();
writeFileSync(resolve(outRoot, projectFile), projectXml, "utf8");
writeFileSync(resolve(outRoot, projectFile.replace(/\.mlt$/, ".kdenlive")), projectXml, "utf8");

if (audioPath && !existsSync(audioPath)) {
  console.warn(`Audio path was written into the project but does not exist yet: ${audioPath}`);
}

console.log(`Generated ${shots.length} SVG files and ${shots.length} PNG files in ${outRoot}`);
console.log(`Kdenlive/MLT project: ${resolve(outRoot, projectFile)}`);
