# Hermes Agent Video Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a first-pass 9:16, roughly 5 minute Hermes Agent architecture explainer video from the existing mdBook chapter.

**Architecture:** Reuse the existing open-source video pipeline from `01-llm-boundaries-3min`: write a Markdown video script, generate SVG/PNG pages with Node and `rsvg-convert`, generate a Piper voiceover, create MLT/Kdenlive project files, render MP4 with Kdenlive `melt`, then verify with `ffprobe`, `volumedetect`, and frame extraction.

**Tech Stack:** Markdown, Node.js ESM, SVG, `rsvg-convert`, Piper TTS, Kdenlive/MLT, `ffmpeg`, `ffprobe`.

---

### Task 1: Content Script

**Files:**
- Create: `books/ai-book/video-scripts/16-hermes-agent-architecture-5min.md`

- [ ] **Step 1: Draft the script document**

Create a Markdown document with these sections:

```markdown
# 5 分钟短视频脚本：Hermes Agent 架构解析

## 视频定位

标题建议：Hermes Agent 架构解析：长期运行的 Agent 如何记忆、进化和跨入口工作？

视频类型：知识讲解型 + 技术解读型
目标观众：工程师、技术管理者、AI Agent 学习者
核心观点：Hermes 的价值不是多一个聊天入口，而是把 Agent 做成长期运行、可记忆、可沉淀技能、可跨平台执行、可产生训练轨迹的个人运行时。

## 成片参数

```text
时长：约 5 分钟
比例：9:16 竖屏，素材结构保留 16:9 迁移可能
风格：技术信息图，架构图优先，文字克制
语速：每分钟约 240-280 中文字
配音：Piper 中文 TTS 首版
字幕：全程中文字幕，关键词高亮
```

## 一句话钩子

```text
真正的 Agent，不是多一个聊天入口，而是一个会长期积累记忆和技能的运行时。
```

## 完整口播稿

```text
真正的 Agent，不是多一个聊天入口，而是一个会长期积累记忆和技能的运行时。

很多 AI 助手看起来都很像：你发一句话，它回一句话；换到 Telegram、Slack 或 CLI，也只是多了几个入口。但 Hermes Agent 想解决的问题更深一层：Agent 如何在长期使用中变得更懂你、更会做事？

可以把 OpenClaw 和 Hermes 放在一起理解。OpenClaw 更像个人 Agent Gateway，重点是让 Agent 到达用户所在的地方。Hermes 更像长期运行的 Agent Runtime，重点是记忆、技能、工具、执行后端和轨迹闭环。

一句话概括：OpenClaw 解决“Agent 如何到达用户”，Hermes 解决“Agent 如何长期成长”。

普通聊天机器人围绕单次回复设计。用户输入，模型回答，这条链路很短。但真正的个人 Agent 需要连续性：它要记得用户偏好，记得项目背景，能把复杂任务沉淀成可复用技能，能从 CLI、消息平台、IDE 或定时任务继续工作，也能把执行轨迹保存下来，用于评估和改进。

所以 Hermes 可以抽象成一个长期运行的 Agent Runtime：它包含 Persistent Memory、Skills、Multi-platform Gateway、Tool Registry、Execution Backends 和 Research Trajectory Pipeline。

这背后有一个关键判断：Agent 的能力不只来自模型参数，而来自模型、记忆、技能、工具、入口、执行环境和历史轨迹共同组成的系统。

理解 Hermes，可以先看六个核心组件。

第一是大脑中枢，也就是 LLM、Prompt、Provider 和上下文压缩。它负责理解用户意图、生成推理、选择工具。

第二是记忆系统，包括长期记忆、用户画像、历史会话搜索、Skills 和 Profiles。它解决的是 Agent 如何连续存在。

第三是小脑，负责规划、状态、工作流和反思。真实任务不是单轮问答，而是多步执行、等待、失败重试和跨会话恢复。

第四是工具中心，包括 Tool Registry、Toolsets、Plugins、MCP 和 Skills。它回答 Agent 能做什么，以及哪些能力在当前入口和身份下可用。

第五是执行引擎，把模型选择的工具调用变成真实动作，并处理参数校验、调度、结果回传、超时、失败和重试。

第六是外部环境，包括 Gateway、Cron、ACP、CLI、本地、Docker、SSH、Modal 等执行后端。它让 Agent 不只停留在聊天窗口，而是进入真实工作流。

一次 Hermes 任务的数据流也不是 Prompt 到 Answer 这么简单。用户从 Telegram、CLI 或 Cron 发起请求，入口层先标准化消息，Session Router 找到正确的用户、线程和 profile。Prompt Builder 再组装人格、长期记忆、用户画像、相关 Skills、项目上下文和工具边界。

模型判断是否需要工具。Tool Runtime 在对应 toolset 和执行后端里调用 web、terminal、file、browser、memory 或 MCP 工具。工具结果回到 Agent Loop，模型继续推理，直到给出答案。最后，会话写入 Session Store，稳定事实进入 Memory，稳定流程进入 Skill 候选，执行轨迹进入研究数据。

这就是 Hermes 的重点：Memory 不只是输入，也参与输出后的回流。没有回流，Agent 每次都从头开始；没有约束，Agent 又会把错误、噪声和越权信息永久化。

所以 Prompt System 的设计很关键。Hermes 不是每次都把所有历史塞进上下文，而是分三层治理：第一层是稳定前缀，包括 Persona、User Profile 和 Persistent Memory；第二层是按需召回，包括 Skills、上下文文件和 Session Search；第三层是本轮运行态，包括当前任务和工具观察结果。

这让上下文既有连续性，又不会失控。

Memory 保存“我知道什么”，Skills 保存“我下次怎么做”，Tool 保存“我实际能执行什么”。这三者不要混在一起。比如一次发布检查，如果反复成功，Hermes-style Skill 应该沉淀触发条件、步骤、依赖工具、验证命令和适用边界，而不是只记一句“发布前要检查”。

这也是 Hermes 最值得借鉴的地方：它把经验变成程序性记忆。一次成功任务本身价值有限，能被压缩成可验证、可复用、可审查的流程，才会真正提升长期能力。

但自我进化也有风险。错误步骤可能被固化，隐私可能被写入长期记忆，跨 profile 的信息可能串线，高风险工具可能被错误授权。所以长期 Agent 必须有验证、评测、人工审核、权限边界和可观测轨迹。

Gateway 和 Toolsets 正是边界的一部分。多入口让 Agent 活在用户所在的平台里，Toolsets 则决定它在每个平台、每个 profile、每种任务下能调用哪些能力。长期在线不等于无限授权。

最后总结一下：Hermes 的核心价值，不是多一个聊天入口，而是把 Agent 做成一个长期运行的个人运行时。它会记忆，会沉淀技能，会跨入口工作，会调用工具，会在不同后端执行，也会把轨迹变成未来改进的数据资产。

如果你要自研个人或团队 Agent，最小可行架构不是“接一个模型 API”，而是至少包含五件事：统一入口，稳定上下文，受控工具集，可审计执行，以及被验证约束的学习闭环。

真正的 Agent 能力，来自模型加 Runtime，加长期轨迹闭环。
```

## 分镜表

| 时间 | 口播重点 | 画面设计 | 屏幕文字 |
|:---|:---|:---|:---|
| 0:00-0:15 | Hook | 大标题 + 多入口图标汇入 Runtime | Agent 不是聊天入口 |
| 0:15-0:37 | Hermes 定位 | OpenClaw 与 Hermes 对比 | Gateway 解决入口，Runtime 解决成长 |
| 0:37-0:57 | Runtime 总览 | Memory、Skills、Gateway、Tools、Backends 环绕 Runtime | Long-running Agent Runtime |
| 0:57-1:17 | 核心判断 | 模型与 Runtime 关系图 | Agent 能力不只来自模型 |
| 1:17-1:39 | 六大组件 | 六组件架构图 | 大脑、记忆、小脑、工具、执行、环境 |
| 1:39-1:59 | 大脑中枢 | Prompt Builder 组装上下文 | LLM + Prompt + Provider |
| 1:59-2:19 | 记忆系统 | Memory、Profile、Session Search、Skills 分层 | 让 Agent 连续存在 |
| 2:19-2:39 | 工具中心 | Tool Registry 到 Toolsets | 决定 Agent 能做什么 |
| 2:39-2:59 | 执行引擎 | Tool Call 到 Backend | 把调用变成真实动作 |
| 2:59-3:21 | 数据流 | 入口、上下文、工具、观察、回流 | 不只是 Prompt 到 Answer |
| 3:21-3:43 | Prompt System | 稳定前缀、按需召回、本轮运行态 | 稳定上下文，而不是动态拼贴 |
| 3:43-4:05 | Skills | 任务轨迹到 SKILL.md | 把经验变成程序性记忆 |
| 4:05-4:25 | Gateway 与 Toolsets | 多入口加权限边界 | 长期在线不等于无限授权 |
| 4:25-4:45 | 风险 | 错误经验、隐私、串线、危险授权 | 自我进化必须被验证约束 |
| 4:45-5:07 | 自研启示 | 最小可行架构清单 | 不要只接模型 API |
| 5:07-5:25 | 总结 | 模型、Runtime、轨迹闭环合并 | 真正能力来自系统闭环 |

## 字幕版文案

```text
真正的 Agent，不是多一个聊天入口，
而是一个会长期积累记忆和技能的运行时。

Hermes Agent 解决的问题是：
Agent 如何在长期使用中变得更懂你、更会做事。

OpenClaw 更像个人 Agent Gateway，
重点是让 Agent 到达用户所在的地方。

Hermes 更像长期运行的 Agent Runtime，
重点是记忆、技能、工具、执行后端和轨迹闭环。
```

## 发布标题备选

1. Hermes Agent 架构解析：长期运行的 Agent 如何自我进化？
2. 真正的 AI Agent，不只是聊天入口
3. Memory、Skills、Gateway：Hermes Agent 的架构启示

## 封面文案

```text
Hermes Agent
长期运行的 Agent Runtime
```
```

- [ ] **Step 2: Validate script length**

Run:

```bash
node -e "const fs=require('fs'); const s=fs.readFileSync('books/ai-book/video-scripts/16-hermes-agent-architecture-5min.md','utf8'); const m=s.match(/## 完整口播稿\\n\\n```text\\n([\\s\\S]*?)\\n```/); const body=m?m[1]:''; console.log([...body.replace(/\\s+/g,'')].length)"
```

Expected: a character count between `1300` and `1500`.

### Task 2: Asset Renderer

**Files:**
- Create: `books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs`
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/svg/`
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/png/`
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/shot-list.csv`

- [ ] **Step 1: Implement renderer by adapting the existing LLM boundary renderer**

Create a focused Node ESM script with:

```javascript
import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
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
  { id: "01", duration: 15, label: "Hook", title: ["Agent 不是", "多一个聊天入口"], subtitle: "它应该是长期运行的个人 Runtime", footer: "会记忆、会沉淀技能、会跨入口工作", type: "hook" },
  { id: "02", duration: 22, label: "Position", title: ["OpenClaw 到达用户", "Hermes 长期成长"], subtitle: "Gateway 解决入口，Runtime 解决能力积累", footer: "Hermes 的重点是长期能力增长", type: "compare" },
  { id: "03", duration: 20, label: "Runtime", title: ["Hermes =", "Long-running Agent Runtime"], subtitle: "Memory + Skills + Gateway + Tools + Backends", footer: "不是聊天机器人，而是可扩展运行时", type: "runtime" },
  { id: "04", duration: 20, label: "Thesis", title: ["Agent 能力", "不只来自模型"], subtitle: "模型、记忆、技能、工具、入口和轨迹共同组成系统", footer: "参数不是全部，Runtime 才让能力持续增长", type: "thesis" },
  { id: "05", duration: 22, label: "Six Parts", title: ["六大组件", "构成长运行 Agent"], subtitle: "大脑 / 记忆 / 小脑 / 工具 / 执行 / 环境", footer: "先看分层，再看数据流", type: "six" },
  { id: "06", duration: 20, label: "Brain", title: ["大脑中枢", "LLM + Prompt + Provider"], subtitle: "理解意图、生成推理、选择工具", footer: "Prompt Builder 决定模型看到什么", type: "brain" },
  { id: "07", duration: 20, label: "Memory", title: ["记忆系统", "让 Agent 连续存在"], subtitle: "Memory / User Profile / Session Search / Skills / Profiles", footer: "记忆不是越多越好，而是边界清楚", type: "memory" },
  { id: "08", duration: 20, label: "Tools", title: ["工具中心", "决定 Agent 能做什么"], subtitle: "Tool Registry / Toolsets / MCP / Plugins / Skills", footer: "Toolsets 是权限治理的基本单位", type: "tools" },
  { id: "09", duration: 20, label: "Execution", title: ["执行引擎", "把工具调用变成真实动作"], subtitle: "校验参数、调度后端、回传结果、处理失败", footer: "没有执行引擎，工具调用只是 demo", type: "execution" },
  { id: "10", duration: 22, label: "Flow", title: ["一次任务", "不是 Prompt -> Answer"], subtitle: "入口 -> 上下文 -> 工具 -> 观察 -> 回流", footer: "Memory 同时参与输入和输出后的沉淀", type: "flow" },
  { id: "11", duration: 22, label: "Context", title: ["Prompt System", "稳定上下文而不是动态拼贴"], subtitle: "稳定前缀 / 按需召回 / 本轮运行态", footer: "连续性和可控性必须同时存在", type: "context" },
  { id: "12", duration: 22, label: "Skills", title: ["Skills", "把经验变成程序性记忆"], subtitle: "Memory = 知道什么，Skill = 下次怎么做", footer: "可验证、可复用、可审查，才值得沉淀", type: "skills" },
  { id: "13", duration: 20, label: "Gateway", title: ["Gateway + Toolsets", "长期在线但不能无限授权"], subtitle: "多入口接入，能力按 profile 和任务分组", footer: "跨平台工作必须配合权限边界", type: "gateway" },
  { id: "14", duration: 20, label: "Risk", title: ["自我进化", "必须被验证约束"], subtitle: "错误经验、隐私泄露、跨 profile 串线、高风险授权", footer: "长期 Agent 需要评测、审核和审计", type: "risk" },
  { id: "15", duration: 22, label: "MVP", title: ["自研 Agent", "最小可行架构"], subtitle: "统一入口 / 稳定上下文 / 受控工具 / 可审计执行 / 学习闭环", footer: "不要只接模型 API，要设计 Runtime", type: "mvp" },
  { id: "16", duration: 18, label: "Takeaway", title: ["真正的 Agent 能力", "来自模型 + Runtime + 轨迹闭环"], subtitle: "长期运行，持续沉淀，受验证约束", footer: "Hermes 的启示：把 Agent 当系统设计", type: "takeaway" },
];
```

Use small drawing helpers: `esc`, `text`, `rect`, `line`, `circle`, `header`, `titleBlock`, `footer`, `standardBackground`, `chip`, `moduleCard`, `flowArrow`, `visual`, `svg`, `frames`, `timecode`, `buildMltProject`, `writeShotList`, `writeReadme`.

- [ ] **Step 2: Generate first image-only assets**

Run:

```bash
node books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs
```

Expected:

```text
Generated 16 SVG files and 16 PNG files
Kdenlive/MLT project: /Users/wxquare/go/src/github.com/wxquare.github.io/books/ai-book/video-assets/16-hermes-agent-architecture-5min/hermes-agent-architecture-image-only.mlt
```

- [ ] **Step 3: Check image dimensions**

Run:

```bash
file books/ai-book/video-assets/16-hermes-agent-architecture-5min/png/*.png
```

Expected: every line contains `PNG image data, 1080 x 1920`.

### Task 3: Voiceover

**Files:**
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover.txt`
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover-5min.wav`

- [ ] **Step 1: Extract voiceover text**

Create `voiceover.txt` from the `完整口播稿` section only.

- [ ] **Step 2: Generate Piper voiceover**

Run:

```bash
/private/tmp/piper-tts-venv/bin/piper \
  -m /private/tmp/piper-models/zh_CN-huayan-medium/zh_CN-huayan-medium.onnx \
  -c /private/tmp/piper-models/zh_CN-huayan-medium/zh_CN-huayan-medium.onnx.json \
  -i books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover.txt \
  -f books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover-5min.wav \
  --length-scale 1.45 \
  --sentence-silence 0.55
```

Expected: WAV file is created.

- [ ] **Step 3: Check WAV duration**

Run:

```bash
python3 -c "import wave; p='books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover-5min.wav'; w=wave.open(p); print(round(w.getnframes()/w.getframerate(), 1), w.getframerate(), w.getnchannels())"
```

Expected: duration is roughly `285` to `325` seconds.

### Task 4: Kdenlive Project and MP4

**Files:**
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/hermes-agent-architecture-with-audio.mlt`
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/hermes-agent-architecture-with-audio.kdenlive`
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/hermes-agent-architecture-5min.mp4`
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/hermes-agent-architecture-5min-final.mp4`

- [ ] **Step 1: Generate project with audio**

Run:

```bash
node books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs \
  --audio books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover-5min.wav
```

Expected: MLT and Kdenlive files are created with audio.

- [ ] **Step 2: Render image project to MP4**

Run:

```bash
mkdir -p books/ai-book/video-assets/16-hermes-agent-architecture-5min/output
/Applications/kdenlive.app/Contents/MacOS/melt \
  books/ai-book/video-assets/16-hermes-agent-architecture-5min/hermes-agent-architecture-with-audio.mlt \
  -consumer avformat:books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/hermes-agent-architecture-5min.mp4 \
  vcodec=libx264 \
  acodec=aac \
  video_off=0 \
  real_time=-1
```

Expected: MP4 file is created.

- [ ] **Step 3: Rewrap WAV audio into final MP4**

Run:

```bash
/Applications/kdenlive.app/Contents/MacOS/ffmpeg \
  -y \
  -i books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/hermes-agent-architecture-5min.mp4 \
  -i books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover-5min.wav \
  -map 0:v:0 \
  -map 1:a:0 \
  -c:v copy \
  -c:a aac \
  -b:a 192k \
  -af apad \
  -shortest \
  -movflags +faststart \
  books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/hermes-agent-architecture-5min-final.mp4
```

Expected: final MP4 file is created with audible AAC audio.

### Task 5: Verification

**Files:**
- Generate: `books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/preview-02m30s.png`

- [ ] **Step 1: Inspect final media streams**

Run:

```bash
/Applications/kdenlive.app/Contents/MacOS/ffprobe \
  -v error \
  -show_entries format=duration,size \
  -show_streams \
  books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/hermes-agent-architecture-5min-final.mp4
```

Expected: video is H.264, `1080x1920`, roughly 5 minutes, audio is AAC mono.

- [ ] **Step 2: Check volume**

Run:

```bash
/Applications/kdenlive.app/Contents/MacOS/ffmpeg \
  -i books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/hermes-agent-architecture-5min-final.mp4 \
  -af volumedetect \
  -vn -sn -dn \
  -f null /dev/null
```

Expected: `mean_volume` is far above `-91.0 dB`; target range is around `-24 dB` to `-10 dB`.

- [ ] **Step 3: Extract preview frame**

Run:

```bash
/Applications/kdenlive.app/Contents/MacOS/ffmpeg \
  -y \
  -ss 00:02:30 \
  -i books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/hermes-agent-architecture-5min-final.mp4 \
  -frames:v 1 \
  books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/preview-02m30s.png
file books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/preview-02m30s.png
```

Expected: `PNG image data, 1080 x 1920`.

- [ ] **Step 4: Build blog/mdBook if article files changed**

Run only if source posts or mdBook source files changed:

```bash
npm run clean && npm run build
```

Expected: build completes successfully.
