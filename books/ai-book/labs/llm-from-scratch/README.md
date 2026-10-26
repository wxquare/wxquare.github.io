# LLM from Scratch

一个从随机初始化开始训练 Decoder-only GPT 的最小实验。它使用字符级 Tokenizer 和本地文本，目标是理解训练闭环，而不是训练通用对话模型。

配套正文：[第 5 章：从零训练一个 LLM——一条可复现的工程路径](https://wxquare.github.io/ai-book/part1/05-train-llm-from-scratch.html)。正文负责预算、数据、训练、恢复与评估方法；本目录提供可在单机运行的最小实现。

## 实现边界

当前实验实现字符级 Tokenizer、学习式绝对位置编码、多头注意力（MHA）、LayerNorm、GELU、单进程训练、采样窗口验证，以及用于评估和生成的模型权重 checkpoint。它没有实现 BPE、RoPE、GQA、RMSNorm、SwiGLU、BPB、分布式训练、数据 manifest、SFT 或偏好训练。

现有 checkpoint 保存模型权重、模型配置和字符词表，可用于评估与生成；它不保存优化器、学习率调度器、随机数状态或数据读取位置，因此不能保证中断后的逐步精确续训。正文中的 1.15B 参数训练计划是教学设计，不是本实验已经执行过的规模或性能结果。

## 环境

需要 Python 3.10+ 和 PyTorch。建议在本目录所在工作区创建虚拟环境：

```bash
python3.12 -m venv .venv
source .venv/bin/activate
python -m pip install -r requirements.txt
```

设备自动按 CUDA、Apple MPS、CPU 的顺序选择；也可通过 `--device cpu`、`--device mps` 或 `--device cuda` 强制指定。

## 训练语料与再分发边界

仓库内的 `data/tiny-shakespeare.txt` 是一个 769 字节的教学片段，截取自 Andrej Karpathy 的 [Tiny Shakespeare 语料](https://github.com/karpathy/char-rnn/blob/master/data/tinyshakespeare/input.txt)，内容对应 William Shakespeare 的《Romeo and Juliet》。上游 `char-rnn` README 将 Tiny Shakespeare 说明为莎士比亚作品子集；Project Gutenberg 的 [Romeo and Juliet eBook #1513](https://www.gutenberg.org/ebooks/1513) 将该作品标记为在美国属于公有领域。核验日期：2026-10-25。

本地文件只保留运行 smoke test 所需的短片段，SHA-256 为 `99bbc72ca884210f9ce1b6dbfc6defa3f346315715bf7fd0b14431e943035d51`。这里的再分发依据是原文学作品的公有领域状态，不把 `char-rnn` 的 MIT 代码许可证泛化为数据集许可证；在其他司法辖区发布或复用前，使用者仍需自行确认当地版权状态。若希望完全避开第三方语料，直接用有明确授权的 UTF-8 文本替换该文件即可。

## 快速运行

```bash
python train.py --input data/tiny-shakespeare.txt --preset smoke --max-steps 100 --out-dir out/smoke
python evaluate.py --checkpoint out/smoke/checkpoint.pt --input data/tiny-shakespeare.txt
python generate.py --checkpoint out/smoke/checkpoint.pt --prompt "ROMEO" --max-new-tokens 80 --temperature 0.8 --top-k 10
```

Checkpoint 只能写在本实验目录的 `out/` 下；`--out-dir` 传入其他路径会报错。训练会固定随机种子、启用 PyTorch 确定性算法并在 CUDA 上关闭 cuDNN benchmark。不同 PyTorch 版本和设备后端（尤其是 MPS）仍可能报告不支持确定性实现的警告，此时请使用 CPU 获得最可复现的教学实验结果。

`smoke` 是 2 层、48 维的小模型，适合验证流程；`default` 是 6 层、384 维、6 头、128 上下文的千万级模型配置：

```bash
python train.py --input path/to/your-corpus.txt --preset default --max-steps 1000 --out-dir out/default
```

## 数据流

```text
UTF-8 文本 → 字符词表 → token id → 固定窗口 batch
→ Causal Self-Attention → next-token cross entropy
→ AdamW 反向传播 → checkpoint → 评估与自回归生成
```

每个 batch 的输入是长度为 `block_size` 的 token 序列，目标是同一序列向右移动一位。Causal Mask 保证第 `t` 个位置只能访问不晚于 `t` 的 token，因此训练目标不会泄漏未来文本。

## 替换语料

训练输入必须是 UTF-8 纯文本。先用小型、授权清晰的语料验证流程；字符级 Tokenizer 会拒绝生成提示词中未出现过的字符。若要训练更实用的模型，建议依次演进：

1. 用 BPE 或 SentencePiece 替换字符词表。
2. 使用 TinyStories 等经过清洗的数据集。
3. 从随机初始化训练转向现有 Base Model 的继续预训练。
4. 在评测集验证后，再进行 SFT 与偏好优化。

## 测试

```bash
python -m pytest -v
```

测试覆盖 Tokenizer、next-token batch、因果遮罩、反向传播、训练 loss 下降、checkpoint 恢复、评估和生成。
