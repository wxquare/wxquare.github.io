# Coding Agent MVP

这是《AI Agent 工程实践》第 12 章的可运行原型。它实现了一个最小 Coding Agent 闭环：

- 读取项目规则和 repo map；
- 让模型输出 JSON action；
- 通过注册工具执行 read/search/edit/shell/diff；
- 用 policy 控制写文件和 shell；
- 保存 JSONL trace；
- 输出执行步骤、变更文件和 git diff。

## 目录

```text
coding-agent-mvp/
├── agent.py
├── config.py
├── context.py
├── llm.py
├── models.py
├── policy.py
├── run.py
├── tools.py
├── trace_writer.py
├── verifier.py
├── agent.config.example.toml
├── demo/
└── tests/
```

## 配置 DeepSeek

复制示例配置：

```bash
cp agent.config.example.toml agent.config.toml
```

编辑 `agent.config.toml`：

```toml
[llm]
provider = "deepseek"
base_url = "https://api.deepseek.com"
api_key = "sk-your-deepseek-api-key"
model = "deepseek-v4-flash"
temperature = 0
max_tokens = 4096
thinking = "disabled"
```

`agent.config.toml` 已经被 `.gitignore` 忽略，不要把真实 API key 提交到仓库。

## 运行 Demo

在当前目录执行：

```bash
python3 run.py "给 calculator.py 的 divide 函数补充除零错误处理，并添加 pytest 测试。要求：b 为 0 时抛出 ValueError；保留原有正常除法行为；运行 pytest 验证。" --repo demo --config agent.config.toml
```

如果你的系统有 `python` 命令，也可以使用 `python run.py ...`。

运行结束后会看到：

- `status`
- `changed_files`
- 每一步 thought、tool call、tool result
- `traces/*.jsonl`
- `git diff`

## 本地测试

运行原型自己的测试：

```bash
python3 -m unittest discover -s tests -v
```

可选安装 demo 测试依赖：

```bash
python3 -m pip install -r requirements.txt
```

## 安全边界

这个原型刻意保持小而可审查：

- 文件路径必须在 `--repo` 指定的 workspace 内；
- 禁止访问 `.git`、`node_modules`、`.venv`、`agent.config.toml` 等路径；
- 编辑工具只支持 `replace_in_file` 和 `create_file`；
- shell 只允许配置文件中列出的命令；
- `git` 写操作不开放，只提供 `git_diff` 读操作。
