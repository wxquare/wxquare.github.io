# 从内容到短视频：开源工具生成 3 分钟 AI 教程视频

本文记录一次完整流程：把《LLM 能力边界与架构约束》这类书稿内容，制作成一条 9:16、约 3 分钟、有图文页、有 AI 配音、可用 Kdenlive 编辑并导出 MP4 的短视频。

这套流程尽量使用开源或本地工具：

```text
内容脚本：Markdown
图文页：SVG + PNG
图像渲染：Node.js + rsvg-convert
AI 配音：Piper TTS
视频项目：Kdenlive / MLT
最终合成：Kdenlive 自带 melt + ffmpeg
验证：ffprobe + volumedetect
```

## 目标产物

以本次视频为例，最终产物在：

```text
books/ai-book/video-assets/01-llm-boundaries-3min/
```

核心文件：

```text
video-scripts/01-llm-boundaries-3min.md
  3 分钟口播稿、分镜表、字幕文案、画面提示词

video-scripts/render-llm-boundaries-assets.mjs
  生成 10 张 SVG / PNG 和 Kdenlive 项目的脚本

video-assets/01-llm-boundaries-3min/svg/
  10 张可编辑 SVG 图文页

video-assets/01-llm-boundaries-3min/png/
  10 张 1080x1920 PNG 图文页

video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
  Piper 生成的 AI 配音

video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.kdenlive
  Kdenlive 项目文件

video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4
  最终成片
```

## Step 1：把书稿改成 3 分钟视频脚本

不要直接把书稿喂给视频工具。短视频需要重新组织成：

```text
标题
视频定位
一句话钩子
完整口播稿
分镜表
字幕版文案
画面素材提示词
发布标题
封面文案
```

本次脚本放在：

```text
books/ai-book/video-scripts/01-llm-boundaries-3min.md
```

三分钟视频的结构可以按这个节奏：

```text
0:00-0:10  Hook：别把 LLM 当确定性系统
0:10-0:25  解释误区：不是模型弱，而是边界错
0:25-0:42  LLM 本质：上下文中的概率生成器
0:42-1:05  LLM 擅长：总结、分类、抽取、规划
1:05-1:25  例子：客户反馈分类
1:25-1:45  LLM 不擅长：计算、实时事实、长期记忆、高风险执行
1:45-2:05  例子：复利计算应该交给工具
2:05-2:28  Agent 架构：LLM + 工程系统
2:28-2:45  RAG / Tool / Memory / Policy / Eval / Trace
2:45-3:00  总结：知道什么时候不让模型单独决定
```

口播稿建议控制在 800-900 个中文字。按 240-280 字/分钟的中文口播速度，比较接近 3 分钟。

## Step 2：设计 10 张 9:16 图文页

为了保证中文字、技术词和公式准确，不建议直接用图片模型生成带文字的整页图。更稳定的做法是：

```text
用代码生成 SVG
-> 用 rsvg-convert 转成 PNG
-> 导入 Kdenlive
```

本次用一个 Node.js 脚本生成 10 张图文页：

```bash
node books/ai-book/video-scripts/render-llm-boundaries-assets.mjs
```

脚本会输出：

```text
books/ai-book/video-assets/01-llm-boundaries-3min/svg/
books/ai-book/video-assets/01-llm-boundaries-3min/png/
books/ai-book/video-assets/01-llm-boundaries-3min/shot-list.csv
books/ai-book/video-assets/01-llm-boundaries-3min/llm-boundaries-image-only.mlt
books/ai-book/video-assets/01-llm-boundaries-3min/llm-boundaries-image-only.kdenlive
```

每张 PNG 都是：

```text
分辨率：1080x1920
比例：9:16
用途：短视频图文页
```

可以用下面命令检查：

```bash
file books/ai-book/video-assets/01-llm-boundaries-3min/png/*.png
```

## Step 3：准备开源 AI 配音环境

本次用 Piper TTS 先跑通第一版。Piper 优点是轻量、本地、开源，缺点是中文自然度不如 GPT-SoVITS、F5-TTS 这类方案。

先创建临时 Python 环境：

```bash
/usr/local/bin/python3.11 -m venv /private/tmp/piper-tts-venv
```

安装 Piper：

```bash
/private/tmp/piper-tts-venv/bin/python -m pip install piper-tts
```

准备模型目录：

```bash
mkdir -p /private/tmp/piper-models/zh_CN-huayan-medium
```

下载中文 voice model 和配置文件：

```bash
curl -L "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/zh/zh_CN/huayan/medium/zh_CN-huayan-medium.onnx?download=true" \
  -o /private/tmp/piper-models/zh_CN-huayan-medium/zh_CN-huayan-medium.onnx

curl -L "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/zh/zh_CN/huayan/medium/zh_CN-huayan-medium.onnx.json?download=true" \
  -o /private/tmp/piper-models/zh_CN-huayan-medium/zh_CN-huayan-medium.onnx.json
```

检查 Piper 是否可用：

```bash
/private/tmp/piper-tts-venv/bin/piper --help
```

## Step 4：从 Markdown 抽出口播稿

本次先把 `01-llm-boundaries-3min.md` 中的“完整口播稿”保存为：

```text
books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover.txt
```

如果手工操作，直接复制“完整口播稿”代码块内容即可。注意：

- 不要复制分镜表；
- 不要复制 Markdown 标题；
- 保留自然段换行，Piper 会按句子生成停顿；
- 避免太多英文缩写连续出现，否则 TTS 容易读得奇怪。

## Step 5：生成第一版 AI 配音

基础版：

```bash
/private/tmp/piper-tts-venv/bin/piper \
  -m /private/tmp/piper-models/zh_CN-huayan-medium/zh_CN-huayan-medium.onnx \
  -c /private/tmp/piper-models/zh_CN-huayan-medium/zh_CN-huayan-medium.onnx.json \
  -i books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover.txt \
  -f books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover.wav \
  --sentence-silence 0.25
```

这版生成成功，但语速偏快，时长约 144 秒。

为了接近 3 分钟，可以加大 `--length-scale` 和 `--sentence-silence`：

```bash
/private/tmp/piper-tts-venv/bin/piper \
  -m /private/tmp/piper-models/zh_CN-huayan-medium/zh_CN-huayan-medium.onnx \
  -c /private/tmp/piper-models/zh_CN-huayan-medium/zh_CN-huayan-medium.onnx.json \
  -i books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover.txt \
  -f books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav \
  --length-scale 1.55 \
  --sentence-silence 0.6
```

检查音频时长：

```bash
python3 -c "import wave; p='books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav'; w=wave.open(p); print(w.getnframes()/w.getframerate(), w.getframerate(), w.getnchannels())"
```

本次结果：

```text
时长：约 176.7 秒
采样率：22050 Hz
声道：mono
```

## Step 6：生成带音频轨的 Kdenlive 项目

把 `voiceover-180s.wav` 写进 MLT / Kdenlive 项目：

```bash
node books/ai-book/video-scripts/render-llm-boundaries-assets.mjs \
  --audio books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
```

生成：

```text
books/ai-book/video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.mlt
books/ai-book/video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.kdenlive
```

如果要手工微调，可以用 Kdenlive 打开：

```bash
open -a /Applications/kdenlive.app \
  books/ai-book/video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.kdenlive
```

## Step 7：用 Kdenlive / melt 渲染视频

Kdenlive 的 macOS App 内部自带：

```text
/Applications/kdenlive.app/Contents/MacOS/melt
/Applications/kdenlive.app/Contents/MacOS/ffmpeg
/Applications/kdenlive.app/Contents/MacOS/ffprobe
```

先创建输出目录：

```bash
mkdir -p books/ai-book/video-assets/01-llm-boundaries-3min/output
```

直接用 Kdenlive 的 `--render` 可能卡在前置阶段。更稳定的方式是直接调用 `melt` 渲染 MLT：

```bash
/Applications/kdenlive.app/Contents/MacOS/melt \
  /Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/books/ai-book/video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.mlt \
  -consumer avformat:/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/books/ai-book/video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min.mp4 \
  vcodec=libx264 \
  acodec=aac \
  video_off=0 \
  real_time=-1
```

渲染过程中会看到类似：

```text
Current Frame:       2704, percentage:         50
Current Frame:       5399, percentage:         99
```

## Step 8：修复 MLT 混音静音问题

本次 `melt` 成功生成了 MP4，但检查发现音频几乎是静音：

```text
mean_volume: -91.0 dB
max_volume: -91.0 dB
```

解决方案：保留视频轨，重新把 Piper 生成的 WAV 封装进去：

```bash
/Applications/kdenlive.app/Contents/MacOS/ffmpeg \
  -y \
  -i books/ai-book/video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min.mp4 \
  -i books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav \
  -map 0:v:0 \
  -map 1:a:0 \
  -c:v copy \
  -c:a aac \
  -b:a 192k \
  -af apad \
  -shortest \
  -movflags +faststart \
  books/ai-book/video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4
```

最终使用：

```text
books/ai-book/video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4
```

不要使用中间文件：

```text
books/ai-book/video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min.mp4
```

它的视频正常，但音频有问题。

## Step 9：验证最终成片

查看视频和音频轨：

```bash
/Applications/kdenlive.app/Contents/MacOS/ffprobe \
  -v error \
  -show_entries format=duration,size \
  -show_streams \
  books/ai-book/video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4
```

本次结果：

```text
视频：H.264
分辨率：1080x1920
比例：9:16
帧率：30 fps
时长：180 秒
音频：AAC mono
```

检查音量：

```bash
/Applications/kdenlive.app/Contents/MacOS/ffmpeg \
  -i books/ai-book/video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4 \
  -af volumedetect \
  -vn -sn -dn \
  -f null /dev/null
```

本次结果：

```text
mean_volume: -16.3 dB
max_volume: -0.1 dB
```

这说明最终 MP4 不是静音。

抽一帧检查画面：

```bash
/Applications/kdenlive.app/Contents/MacOS/ffmpeg \
  -y \
  -ss 00:02:10 \
  -i books/ai-book/video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4 \
  -frames:v 1 \
  books/ai-book/video-assets/01-llm-boundaries-3min/output/preview-02m10s.png
```

检查图片：

```bash
file books/ai-book/video-assets/01-llm-boundaries-3min/output/preview-02m10s.png
```

应输出：

```text
PNG image data, 1080 x 1920
```

## 常见问题

### 1. Kdenlive 打开项目但没有声音

先确认 `voiceover-180s.wav` 文件存在：

```bash
ls -lh books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
```

如果 Kdenlive 时间线里没有音频轨，重新生成项目：

```bash
node books/ai-book/video-scripts/render-llm-boundaries-assets.mjs \
  --audio books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
```

### 2. `kdenlive --render` 卡住

改用 `melt`：

```bash
/Applications/kdenlive.app/Contents/MacOS/melt path/to/project.mlt \
  -consumer avformat:path/to/output.mp4 \
  vcodec=libx264 \
  acodec=aac \
  real_time=-1
```

### 3. MP4 有音轨但几乎静音

用 `volumedetect` 检查。如果是 `-91 dB` 这类结果，重新用 ffmpeg 把 WAV 封装进去。

### 4. 中文读音不自然

Piper 适合先跑通链路。后续可以换：

- GPT-SoVITS：中文效果更好，适合克隆自己的声音；
- F5-TTS：自然度更好，但预训练模型许可需要单独确认；
- ChatTTS：对话感强，但也需要确认许可和用途。

### 5. 图文页文字需要修改

编辑 SVG：

```text
books/ai-book/video-assets/01-llm-boundaries-3min/svg/
```

或者改生成脚本：

```text
books/ai-book/video-scripts/render-llm-boundaries-assets.mjs
```

然后重新生成：

```bash
node books/ai-book/video-scripts/render-llm-boundaries-assets.mjs \
  --audio books/ai-book/video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
```

## 下一条视频的复用方式

下一条视频可以复用这套目录结构：

```text
books/ai-book/video-scripts/02-xxx-3min.md
books/ai-book/video-scripts/render-xxx-assets.mjs
books/ai-book/video-assets/02-xxx-3min/
```

流程保持不变：

```text
书稿内容
-> 三分钟口播稿
-> 分镜表
-> SVG / PNG 图文页
-> Piper 配音
-> Kdenlive / MLT 项目
-> melt 渲染视频
-> ffmpeg 修正音频
-> ffprobe / volumedetect 验证
```

这条流水线的关键思想是：**文字和技术图用代码生成，保证准确；声音先用开源 TTS 跑通；视频合成用 Kdenlive / MLT 自动化；最终质量用 ffprobe 和抽帧验证。**
