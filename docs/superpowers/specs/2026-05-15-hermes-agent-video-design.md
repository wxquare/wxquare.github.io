# Hermes Agent 5 分钟视频设计

## 背景

将 `books/ai-book/src/part3/05-hermes-agent-architecture.md` 改造成一条约 5 分钟的技术讲解视频。视频延续现有开源流水线：Markdown 脚本、SVG/PNG 图文页、Piper 配音、Kdenlive/MLT 工程、MP4 输出。

## 目标

- 生成一条 9:16 竖屏首版视频，适合视频号、抖音、小红书、B 站竖屏和 YouTube Shorts。
- 内容风格为知识讲解型 + 技术解读型，面向工程师、技术管理者和 AI Agent 学习者。
- 保持技术词和架构图准确，避免依赖图片/视频模型直接生成文字。
- 产物结构保留迁移到 16:9 横屏的可能性。

## 非目标

- 不追求复杂动效和商业模板效果。
- 不引入新的视频制作平台。
- 不修改原始章节正文。
- 不修改既有 `01-llm-boundaries-3min` 视频资产。

## 成片设计

成片参数：

```text
时长：约 5 分钟
比例：9:16
画面：14-16 张代码生成图文页
配音：Piper 中文 TTS 首版
项目：Kdenlive / MLT
输出：MP4
```

叙事结构：

```text
0:00-0:15   Hook：真正的 Agent 不是多一个聊天入口
0:15-0:45   定位：OpenClaw 解决入口，Hermes 解决长期能力增长
0:45-1:25   总览：长期运行 Agent Runtime = Memory + Skills + Gateway + Tools + Backends
1:25-2:05   六大组件：大脑、记忆、小脑、工具中心、执行引擎、外部环境
2:05-2:45   数据流：用户输入 -> 工具执行 -> 观察结果 -> 记忆回流
2:45-3:25   Prompt/Memory：稳定快照 + 按需检索，而不是无限塞上下文
3:25-4:05   Skills：把经验变成程序性记忆
4:05-4:35   Gateway/Toolsets：多入口在线，但权限边界必须清楚
4:35-4:55   风险：自我进化必须被验证和人工审核约束
4:55-5:10   总结：Agent 能力来自模型 + Runtime + 轨迹闭环
```

## 产物

新增文件：

```text
books/ai-book/video-scripts/16-hermes-agent-architecture-5min.md
books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs
books/ai-book/video-assets/16-hermes-agent-architecture-5min/
```

生成目录：

```text
books/ai-book/video-assets/16-hermes-agent-architecture-5min/svg/
books/ai-book/video-assets/16-hermes-agent-architecture-5min/png/
books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover.txt
books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover-5min.wav
books/ai-book/video-assets/16-hermes-agent-architecture-5min/shot-list.csv
books/ai-book/video-assets/16-hermes-agent-architecture-5min/hermes-agent-architecture-with-audio.mlt
books/ai-book/video-assets/16-hermes-agent-architecture-5min/hermes-agent-architecture-with-audio.kdenlive
books/ai-book/video-assets/16-hermes-agent-architecture-5min/output/hermes-agent-architecture-5min-final.mp4
```

## Visual Direction

- 使用深色技术信息图风格，延续上一条视频的工程感。
- 页面避免大段文字，优先使用架构分层图、数据流、对比卡片、关键词矩阵。
- 每页保留标题、核心图形、底部总结句，便于后期转成横屏。
- 中英文技术词保持可读，例如 `Memory`、`Skills`、`Tool Runtime`、`Gateway`、`Profiles`。

## 验证

- 运行渲染脚本生成 SVG/PNG、shot list、MLT/Kdenlive 工程。
- 使用 Piper 生成配音，并检查 WAV 时长。
- 使用 Kdenlive 自带 `melt` 渲染 MP4。
- 如遇 MLT 音频静音，使用 `ffmpeg` 将 WAV 重新封装进 MP4。
- 使用 `ffprobe` 检查分辨率、时长、编码和音轨。
- 使用 `volumedetect` 检查音频不是静音。
- 抽帧检查画面尺寸为 1080x1920。

## 风险与处理

- Piper 中文自然度有限：首版以跑通工程链路为目标，后续可替换为 GPT-SoVITS、F5-TTS 或真人录音。
- 5 分钟信息密度高：口播控制在约 1300-1500 中文字，页面控制在 14-16 张。
- 架构词较多：画面用分层图和关键词高亮承载，口播避免堆术语。
- 构建依赖可能缺失：优先复用现有 Node、rsvg-convert、Kdenlive、Piper 路径；缺失时明确报告。
