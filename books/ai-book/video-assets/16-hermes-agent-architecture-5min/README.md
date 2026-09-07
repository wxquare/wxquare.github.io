# Hermes Agent 架构解析 5 分钟视频素材

生成时间：2026-05-15T15:57:45.444Z

## 文件结构

```text
svg/           16 张可编辑 SVG 图文页
png/           16 张 1080x1920 PNG 图文页
voiceover/     配音文本和 WAV 音频
hermes-agent-architecture-with-audio.mlt   MLT 项目文件
hermes-agent-architecture-with-audio.kdenlive  Kdenlive 项目文件
shot-list.csv 分镜时长表
```

## 建议用法

1. 用 Kdenlive 打开 `hermes-agent-architecture-with-audio.kdenlive`，打不开时再打开 `hermes-agent-architecture-with-audio.mlt`。
2. 每张图已经按分镜时长排列，总时长约 325 秒。
3. 如需改字，修改 `books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs` 后重新运行。

```bash
node books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs
node books/ai-book/video-scripts/render-hermes-agent-architecture-assets.mjs --audio books/ai-book/video-assets/16-hermes-agent-architecture-5min/voiceover/voiceover-5min.wav
```
