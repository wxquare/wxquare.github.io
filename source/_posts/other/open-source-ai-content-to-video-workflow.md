---
title: '从内容到短视频：用开源工具生成 3 分钟 AI 教程视频'
date: '2026-05-08'
description: 记录如何使用开源工具把技术内容整理为脚本、配音、画面和成片，形成可复用的 AI 教程视频工作流。
categories:
  - AI
tags:
  - ai-video
  - open-source
  - kdenlive
  - piper-tts
  - content-creation
---

最近我把《LLM 能力边界与架构约束》这一章，做成了一条 9:16、约 3 分钟、有图文页、有 AI 配音、可用 Kdenlive 打开编辑的短视频。这个过程没有使用剪映，也没有依赖商业视频模板，而是尽量用开源和本地工具跑通了一条可复用的流水线。

这篇文章记录完整过程：如何从一段书稿内容出发，生成视频脚本、10 张竖屏图文页、AI 配音、Kdenlive 工程文件，并最终导出 MP4。

本文是一次 macOS 环境下的实践复盘。脚本和素材已迁移到独立的视频制作项目；由于该项目不在本文仓库中，下面的路径和命令需要按读者自己的项目目录替换，不能直接假设 `/Users/xianguiwang/Projects/video-production/` 存在。

## 环境与验收标准

- macOS；Node.js 20+；Python 3.11；Piper TTS；Kdenlive；`rsvg-convert`；`ffmpeg` 和 `ffprobe`。
- 输入：一章 Markdown 书稿、一份视频脚本和一段可公开使用的配音模型。
- 输出：1080×1920 的 PNG 页面、可编辑的 Kdenlive 工程、带音频的 MP4。
- 验收：画面比例正确、音频轨存在且不是静音、画面和口播时长对齐、抽帧无文字溢出。

如果没有独立脚本仓库，这篇文章只能作为流程记录；要做成可复现教程，需要同时公开生成脚本、最小输入样例和固定版本。

## 一、目标：把文章变成可发布的视频资产

这次的目标不是做一个炫技 demo，而是跑通一个可以重复使用的内容生产流程：

```text
书稿内容
-> 三分钟视频脚本
-> 10 张 9:16 图文页
-> AI 配音音频
-> Kdenlive 工程
-> MP4 成片
```

最终产物包括：

```text
video-scripts/01-llm-boundaries-3min.md
video-scripts/render-llm-boundaries-assets.mjs
video-assets/01-llm-boundaries-3min/png/
video-assets/01-llm-boundaries-3min/svg/
video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.kdenlive
video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4
```

整套方案用到的工具如下：

```text
内容脚本：Markdown
图文页：SVG + PNG
图像渲染：Node.js + rsvg-convert
AI 配音：Piper TTS
视频项目：Kdenlive / MLT
最终合成：Kdenlive 自带 melt + ffmpeg
质量验证：ffprobe + volumedetect + 抽帧检查
```

## 二、为什么不用直接 AI 生成整条视频

直接让 AI 生成视频当然很方便，但技术教程类内容有一个核心问题：文字必须准确。

如果画面中出现“LLM”“RAG”“Policy”“Eval”“Trace”“概率生成器”等技术词，一旦交给图片或视频模型直接生成，常见问题是：

- 中文字形变形；
- 英文缩写写错；
- 公式、架构词、标题不稳定；
- 同一套页面的视觉风格不一致；
- 后期改一个词需要重新生成整张图。

所以我选择了一条更工程化的路线：

```text
技术文字和版式用代码生成，保证准确；
配音用开源 TTS 先跑通；
视频剪辑用 Kdenlive / MLT 自动化；
最终质量用命令行工具验证。
```

这条路线不一定是最快的，但它非常适合做系列化技术内容。

## 三、第一步：把书稿改成三分钟视频脚本

不要直接把书稿塞进视频工具。书稿适合阅读，短视频适合节奏、钩子和分镜。

我把原始章节重新整理成一个视频脚本，结构包括：

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

三分钟内容可以按下面的节奏拆：

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

本次 3 分钟成片的口播字数应由实际音频时长倒推。以 240—280 字/分钟估算，纯口播约为 720—840 字；停顿、片头和片尾会进一步减少可用字数，因此不要把 800—900 字当作固定标准。

## 四、第二步：用 SVG 生成 10 张竖屏图文页

图文页我没有交给图片模型直接生成，而是用 Node.js 生成 SVG，再转换成 PNG。

这样做的好处是：

- 文字 100% 可控；
- 字号、间距、颜色可以统一；
- 一次生成 10 张页面；
- 后续修改文案只需要改脚本；
- SVG 可以继续编辑，PNG 可以直接导入视频软件。

生成命令：

```bash
node video-scripts/render-llm-boundaries-assets.mjs
```

脚本输出：

```text
video-assets/01-llm-boundaries-3min/svg/
video-assets/01-llm-boundaries-3min/png/
video-assets/01-llm-boundaries-3min/shot-list.csv
video-assets/01-llm-boundaries-3min/llm-boundaries-image-only.mlt
video-assets/01-llm-boundaries-3min/llm-boundaries-image-only.kdenlive
```

每张 PNG 都是 1080×1920，适合常见竖屏发布场景；不同平台的封面安全区、码率和时长限制仍需单独核对。

可以用下面命令检查图片尺寸：

```bash
file video-assets/01-llm-boundaries-3min/png/*.png
```

## 五、第三步：用 Piper 生成开源 AI 配音

配音这一步，我先选了 Piper TTS。它的优点是轻量、本地、开源，适合快速跑通流程；缺点是中文自然度不如 GPT-SoVITS、F5-TTS 这类更复杂的方案。

先设置项目目录和 Python 虚拟环境路径（下面命令使用 macOS/zsh 语法）：

```bash
export VIDEO_PROJECT="$PWD"
export PIPER_VENV="$VIDEO_PROJECT/.venv-piper"
python3.11 -m venv "$PIPER_VENV"
```

安装 Piper：

```bash
"$PIPER_VENV/bin/python" -m pip install piper-tts
```

准备模型目录：

```bash
export PIPER_MODEL_DIR="$VIDEO_PROJECT/models/zh_CN-huayan-medium"
mkdir -p "$PIPER_MODEL_DIR"
```

下载中文 voice model 和配置文件：

```bash
curl -L "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/zh/zh_CN/huayan/medium/zh_CN-huayan-medium.onnx?download=true" \
  -o "$PIPER_MODEL_DIR/zh_CN-huayan-medium.onnx"

curl -L "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/zh/zh_CN/huayan/medium/zh_CN-huayan-medium.onnx.json?download=true" \
  -o "$PIPER_MODEL_DIR/zh_CN-huayan-medium.onnx.json"
```

检查 Piper 是否可用：

```bash
"$PIPER_VENV/bin/piper" --help
```

## 六、第四步：从 Markdown 抽出口播稿

我把视频脚本中的“完整口播稿”单独保存为：

```text
video-assets/01-llm-boundaries-3min/voiceover/voiceover.txt
```

这里要注意几件事：

- 只保留口播稿，不要复制分镜表；
- 不要复制 Markdown 标题；
- 保留自然段换行，让 TTS 有停顿；
- 尽量避免连续英文缩写，否则中文 TTS 容易读得生硬；
- 技术词可以适当改写成口语，比如把“LLM”读作“大语言模型”。

## 七、第五步：生成第一版配音

基础版命令如下：

```bash
"$PIPER_VENV/bin/piper" \
  -m "$PIPER_MODEL_DIR/zh_CN-huayan-medium.onnx" \
  -c "$PIPER_MODEL_DIR/zh_CN-huayan-medium.onnx.json" \
  -i video-assets/01-llm-boundaries-3min/voiceover/voiceover.txt \
  -f video-assets/01-llm-boundaries-3min/voiceover/voiceover.wav \
  --sentence-silence 0.25
```

这版能生成成功，但语速偏快，时长约 144 秒。为了让它接近 3 分钟，我加大了 `--length-scale` 和 `--sentence-silence`：

```bash
"$PIPER_VENV/bin/piper" \
  -m "$PIPER_MODEL_DIR/zh_CN-huayan-medium.onnx" \
  -c "$PIPER_MODEL_DIR/zh_CN-huayan-medium.onnx.json" \
  -i video-assets/01-llm-boundaries-3min/voiceover/voiceover.txt \
  -f video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav \
  --length-scale 1.55 \
  --sentence-silence 0.6
```

检查音频时长：

```bash
python3 -c "import wave; p='video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav'; w=wave.open(p); print(w.getnframes()/w.getframerate(), w.getframerate(), w.getnchannels())"
```

本次结果大约是：

```text
时长：176.7 秒
采样率：22050 Hz
声道：mono
```

## 八、第六步：生成 Kdenlive 工程

有了图文页和配音之后，就可以生成带音频轨的 MLT / Kdenlive 工程：

```bash
node video-scripts/render-llm-boundaries-assets.mjs \
  --audio video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
```

生成结果：

```text
video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.mlt
video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.kdenlive
```

如果要手工微调，可以用 Kdenlive 打开：

```bash
open -a /Applications/kdenlive.app \
  video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.kdenlive
```

打开后可以继续调整：

- 每一页的停留时长；
- 转场；
- 字幕；
- 背景音乐；
- 音量；
- 封面页；
- 片尾页。

## 九、第七步：用 melt 渲染视频

Kdenlive 的 macOS 应用通常自带 `melt`、`ffmpeg` 和 `ffprobe`。不同安装方式路径不同，先设置三个变量，再执行后续命令：

```text
export MELT_BIN="/Applications/kdenlive.app/Contents/MacOS/melt"
export FFMPEG_BIN="/Applications/kdenlive.app/Contents/MacOS/ffmpeg"
export FFPROBE_BIN="/Applications/kdenlive.app/Contents/MacOS/ffprobe"
```

如果路径不同，用 `which melt`、`which ffmpeg` 或 Kdenlive 的安装目录替换变量值。

先创建输出目录：

```bash
mkdir -p video-assets/01-llm-boundaries-3min/output
```

直接调用 `melt` 渲染 MLT 文件：

```bash
"$MELT_BIN" \
  video-assets/01-llm-boundaries-3min/llm-boundaries-with-audio.mlt \
  -consumer avformat:video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min.mp4 \
  vcodec=libx264 \
  acodec=aac \
  video_off=0 \
  real_time=-1
```

渲染过程中会看到类似进度：

```text
Current Frame:       2704, percentage:         50
Current Frame:       5399, percentage:         99
```

## 十、第八步：修复 MLT 混音静音问题

这次遇到一个实际问题：`melt` 成功生成了 MP4，但音频几乎是静音。

用 `volumedetect` 检查后发现：

```text
mean_volume: -91.0 dB
max_volume: -91.0 dB
```

解决方案是保留视频轨，重新把 Piper 生成的 WAV 封装进去：

```bash
"$FFMPEG_BIN" \
  -y \
  -i video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min.mp4 \
  -i video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav \
  -map 0:v:0 \
  -map 1:a:0 \
  -c:v copy \
  -c:a aac \
  -b:a 192k \
  -af apad \
  -shortest \
  -movflags +faststart \
  video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4
```

最终使用：

```text
video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4
```

不要使用中间文件：

```text
video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min.mp4
```

它的视频正常，但音频有问题。

## 十一、第九步：验证最终成片

视频生成之后，不要只靠播放器看一眼。最好用命令行做几项验证。

查看视频和音频轨：

```bash
"$FFPROBE_BIN" \
  -v error \
  -show_entries format=duration,size \
  -show_streams \
  video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4
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
"$FFMPEG_BIN" \
  -i video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4 \
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

也可以抽一帧检查画面：

```bash
"$FFMPEG_BIN" \
  -y \
  -ss 00:02:10 \
  -i video-assets/01-llm-boundaries-3min/output/llm-boundaries-3min-final.mp4 \
  -frames:v 1 \
  video-assets/01-llm-boundaries-3min/output/preview-02m10s.png
```

检查图片：

```bash
file video-assets/01-llm-boundaries-3min/output/preview-02m10s.png
```

应输出：

```text
PNG image data, 1080 x 1920
```

除了检查总时长，还应抽查片头、转场、重点术语和片尾：口播进入下一段时，画面是否已经切换；字幕是否超出安全区；最后一页是否被音频截断。总时长一致不能证明每个分镜都同步。

## 十二、常见问题

### 1. Kdenlive 打开项目但没有声音

先确认 `voiceover-180s.wav` 文件存在：

```bash
ls -lh video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
```

如果 Kdenlive 时间线里没有音频轨，重新生成项目：

```bash
node video-scripts/render-llm-boundaries-assets.mjs \
  --audio video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
```

### 2. `kdenlive --render` 卡住

可以改用 `melt`：

```bash
/Applications/kdenlive.app/Contents/MacOS/melt path/to/project.mlt \
  -consumer avformat:path/to/output.mp4 \
  vcodec=libx264 \
  acodec=aac \
  real_time=-1
```

### 3. MP4 有音轨但几乎静音

用 `volumedetect` 检查。如果是 `-91 dB` 这类结果，可以重新用 `ffmpeg` 把 WAV 封装进去。

### 4. 中文读音不自然

Piper 适合先跑通链路。如果后续要提高中文自然度，可以考虑：

- GPT-SoVITS：中文效果更好，适合克隆自己的声音；
- F5-TTS：自然度更好，但预训练模型许可需要单独确认；
- ChatTTS：对话感强，但也需要确认许可和用途。

### 5. 图文页文字需要修改

可以直接编辑 SVG：

```text
video-assets/01-llm-boundaries-3min/svg/
```

也可以修改生成脚本：

```text
video-scripts/render-llm-boundaries-assets.mjs
```

然后重新生成：

```bash
node video-scripts/render-llm-boundaries-assets.mjs \
  --audio video-assets/01-llm-boundaries-3min/voiceover/voiceover-180s.wav
```

## 十三、下一条视频如何复用

下一条视频可以复用这套结构：

```text
video-scripts/02-xxx-3min.md
video-scripts/render-xxx-assets.mjs
video-assets/02-xxx-3min/
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

这条流水线的核心思想是：**文字和技术图用代码生成，保证准确；声音先用开源 TTS 跑通；视频合成用 Kdenlive / MLT 自动化；最终质量用 ffprobe 和抽帧验证。**

对技术内容创作者来说，这比单纯追求“一键生成视频”更稳定。因为技术教程最重要的不是画面有多花，而是概念准确、文字可信、流程可复现。
