# LLM 能力边界 3 分钟视频素材

生成时间：2026-05-08T09:33:29.565Z

## 文件结构

```text
svg-ai/   10 张可编辑 SVG 图文页
png-ai/   10 张 1080x1920 PNG 图文页
voiceover/   配音文本和 WAV 音频
llm-boundaries-ai-bg-with-audio.mlt   MLT 项目文件
llm-boundaries-ai-bg-with-audio.kdenlive   Kdenlive 项目文件
shot-list.csv    分镜时长表
```

## 建议用法

1. 用 Kdenlive 打开 `llm-boundaries-ai-bg-with-audio.kdenlive`，打不开时再打开 `llm-boundaries-ai-bg-with-audio.mlt`。
2. 项目已绑定音频轨：`/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav`。
3. 每张图已经按分镜时长排列，总时长约 180 秒。
4. 如需改字，编辑 `svg/*.svg` 后重新运行：

```bash
node books/ai-book/video-scripts/render-llm-boundaries-assets.mjs
```

如果要把配音一起写入项目：

```bash
node books/ai-book/video-scripts/render-llm-boundaries-assets.mjs --audio /absolute/path/to/voiceover.mp3
```
