# Claude Code 记忆系统逆向调研

本文整理 Claude Code 2.1.150 的记忆系统实现形态，重点关注文件路径、加载顺序、上下文注入、存储格式、触发时机和可配置项。结论基于三类证据：官方文档、本机可观察数据、安装包静态痕迹。

边界说明：本文不做二进制反编译，不绕过授权，不读取或摘录敏感会话正文。安装包分析仅限 wrapper、package metadata 和 targeted strings 观察；strings 只能证明安装产物中存在相关符号或文案，不能等同于完整源码实现。

## 1. 总体模型

Claude Code 的每个 session 都从新的 context window 开始。跨 session 的知识主要由两套机制承载：

- `CLAUDE.md` 指令栈：由用户、项目或组织维护，表达稳定规则、工作流、架构背景和行为偏好。
- auto memory：由 Claude Code 在工作过程中自动写入，用来沉淀它认为未来会复用的项目经验、调试结论和用户偏好。

除此之外，Claude Code 还会保存 session transcript 和 file history。它们支持恢复会话、回放上下文和撤销文件编辑，但不等价于会在新 session 自动注入的长期记忆。

官方文档对二者的定位很明确：`CLAUDE.md` 用于你希望 Claude 始终遵守的持久指令；auto memory 用于 Claude 自动积累的 learnings。两者都会在每个对话开始时进入上下文，但 `CLAUDE.md` 是规则输入，auto memory 是经验输入。

参考资料：

- https://code.claude.com/docs/en/memory
- https://code.claude.com/docs/en/how-claude-code-works
- https://code.claude.com/docs/en/settings

## 2. 记忆类型

### 2.1 `CLAUDE.md` 指令文件

`CLAUDE.md` 是 Markdown 指令文件。它的作用是把项目约定、构建命令、代码风格、常见陷阱、权限边界等信息注入 Claude Code 的上下文。

常见位置包括：

- 项目级：`<repo>/CLAUDE.md`
- 项目级扩展：`<repo>/.claude/CLAUDE.md`
- 本地私有：`<repo>/CLAUDE.local.md`
- 用户级：`~/.claude/CLAUDE.md`
- 组织级 managed policy：
  - macOS：`/Library/Application Support/ClaudeCode/CLAUDE.md`
  - Linux/WSL：`/etc/claude-code/CLAUDE.md`
  - Windows：`C:\Program Files\ClaudeCode\CLAUDE.md`

本仓库当前存在项目级文件：

```text
/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/CLAUDE.md
```

它包含 Hexo 博客项目概况、目录结构、开发命令、写作规范、常见陷阱和 AI Agent 协作原则。这个文件会成为 Claude Code 每次进入该仓库时的稳定上下文。

### 2.2 `.claude/rules/*.md`

`.claude/rules/` 用于把大型 `CLAUDE.md` 拆成多份规则文件。规则文件可以无条件加载，也可以通过 YAML front matter 的 `paths` 字段做路径级匹配。

示例：

```markdown
---
paths:
  - "source/_posts/**/*.md"
---

# Markdown Writing Rules

- 代码块必须指定语言。
- 中英文之间需要有空格。
```

无 `paths` 的规则在启动时加载；带 `paths` 的规则在 Claude Code 读取匹配文件时触发加载。这个机制适合大型 monorepo：只在处理某类文件时引入对应规则，减少无关上下文。

本仓库当前没有发现 `.claude/rules/`，但已有 `.claude/agents`、`.claude/commands`、`.claude/skills`、`.claude/settings.json` 和 `.claude/templates`。

### 2.3 auto memory

auto memory 是 Claude Code 自动写入的本地 Markdown 记忆目录。默认路径为：

```text
~/.claude/projects/<project>/memory/
```

官方文档说明 `<project>` 由 git repository 派生，因此同一 repo 的不同 worktree 和子目录会共享一套 auto memory。目录内通常包含：

```text
~/.claude/projects/<project>/memory/
├── MEMORY.md
├── debugging.md
├── api-conventions.md
└── ...
```

其中 `MEMORY.md` 是入口索引。Claude Code 在 session 启动时只加载 `MEMORY.md` 的前 200 行或前 25KB，二者取更小值。其他 topic 文件不会在启动时全部加载，而是在需要时通过标准文件读取能力按需读取。

本机当前项目的 auto memory 目录为：

```text
/Users/xianguiwang/.claude/projects/-Users-xianguiwang-gopath-src-github-com-wxquare-wxquare-github-io/memory/
```

观察结果：该目录存在，但当前为空。

### 2.4 session transcript

Claude Code 会把会话记录写入本地 JSONL 文件：

```text
~/.claude/projects/<project>/*.jsonl
```

本机当前项目存在：

```text
/Users/xianguiwang/.claude/projects/-Users-xianguiwang-gopath-src-github-com-wxquare-wxquare-github-io/a480445b-a052-476a-8683-3679211745ab.jsonl
```

只读结构观察显示该文件首行包含键：

```text
permissionMode,sessionId,type
```

这类 transcript 支持 resume、continue、fork session 等功能。它保存了历史会话材料，但新 session 不会天然把所有历史 transcript 当作长期记忆加载。跨 session 的稳定注入仍主要依赖 `CLAUDE.md` 和 auto memory。

### 2.5 checkpoint / file history

Claude Code 在编辑文件前会做本地快照，用于撤销或 rewind。默认位置形态类似：

```text
~/.claude/file-history/<session>/<file>@v1
```

本机观察到：

```text
/Users/xianguiwang/.claude/file-history/a480445b-a052-476a-8683-3679211745ab/5256a36ba60033e0@v1
```

这部分更像安全机制，不是记忆系统的主要知识来源。它帮助恢复文件状态，但不会作为项目知识主动进入新会话。

## 3. 加载机制

### 3.1 启动时加载

Claude Code 启动一个新 session 时，会构建初始上下文。与记忆相关的输入主要包括：

- 当前工作目录及祖先/项目范围内的 `CLAUDE.md`
- `.claude/CLAUDE.md`
- 用户级或 managed policy 的 `CLAUDE.md`
- 无路径限制的 `.claude/rules/*.md`
- auto memory 的 `MEMORY.md` 前 200 行或 25KB
- 系统指令、工具定义摘要、已加载 skill 的说明等其他上下文

官方文档强调：这些记忆内容是 context，不是强制配置。它们会影响模型行为，但不是硬约束。真正的硬约束应放在 settings 的 permissions、sandbox、managed policy 等配置中。

### 3.2 子目录 `CLAUDE.md` 按需加载

Claude Code 会发现当前工作目录下子目录中的 `CLAUDE.md` 和 `CLAUDE.local.md`。这些文件不是全部在启动时加载，而是在 Claude Code 读取对应子目录文件时注入上下文。

这说明 `CLAUDE.md` 不是单个 flat file，而是一组按 scope 和路径逐步进入上下文的指令栈。

### 3.3 `.claude/rules/` 路径匹配

`.claude/rules/*.md` 支持 YAML front matter：

```yaml
---
paths:
  - "src/api/**/*.ts"
---
```

没有 `paths` 字段的规则无条件加载。带 `paths` 的规则只有当 Claude Code 读取匹配文件时才触发。触发条件是文件读取，而不是每次 tool use。

### 3.4 auto memory 按需读取和写入

auto memory 默认开启。Claude Code 会在工作过程中判断哪些信息值得保存，比如：

- 构建命令和测试命令
- 项目架构事实
- 调试过程中的关键发现
- 用户反复纠正过的偏好
- 某类任务的稳定处理方式

当 UI 中出现 `Writing memory` 或 `Recalled memory` 时，表示 Claude Code 正在写入或读取：

```text
~/.claude/projects/<project>/memory/
```

`MEMORY.md` 用作索引和启动注入入口，详细内容会被移动到其他 topic 文件中。这样做可以限制启动上下文膨胀，同时保留按需召回能力。

## 4. 配置项

### 4.1 `autoMemoryEnabled`

项目 settings 可关闭 auto memory：

```json
{
  "autoMemoryEnabled": false
}
```

也可以在 Claude Code 中通过 `/memory` 切换。

### 4.2 `autoMemoryDirectory`

用户 settings 可改变 auto memory 存储位置：

```json
{
  "autoMemoryDirectory": "~/my-custom-memory-dir"
}
```

官方文档说明该值必须是绝对路径或以 `~/` 开头。它只接受 policy、user settings 或 `--settings`，不接受 project/local settings，原因是项目仓库不应能把记忆写入重定向到敏感路径。

### 4.3 `claudeMd`

`claudeMd` 是 managed settings 中的组织级注入内容，作用类似组织级 `CLAUDE.md`：

```json
{
  "claudeMd": "Always run `make lint` before committing."
}
```

它只在 managed/policy settings 中生效。写在 user、project 或 local settings 中无效。

### 4.4 `claudeMdExcludes`

`claudeMdExcludes` 用 glob 或绝对路径排除某些 `CLAUDE.md` 或 rules 文件：

```json
{
  "claudeMdExcludes": [
    "**/vendor/**/CLAUDE.md",
    "/home/user/monorepo/other-team/.claude/rules/**"
  ]
}
```

它适合大型 monorepo，避免不相关团队的规则污染当前任务上下文。managed policy 的 `CLAUDE.md` 不能被排除。

### 4.5 环境变量

与记忆相关的关键环境变量：

```text
CLAUDE_CODE_DISABLE_AUTO_MEMORY=1
CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1
```

`CLAUDE_CODE_DISABLE_AUTO_MEMORY=1` 用于禁用 auto memory。

`CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1` 配合 `claude --add-dir` 使用。默认情况下，额外目录中的 `CLAUDE.md` 不会加载；设置该变量后，额外目录中的 `CLAUDE.md`、`.claude/CLAUDE.md`、`.claude/rules/*.md` 和 `CLAUDE.local.md` 才会进入加载范围。

## 5. 本机安装和静态痕迹

### 5.1 版本和入口

本机观察：

```text
Claude Code version: 2.1.150
which claude: /usr/local/bin/claude
/usr/local/bin/claude -> ../lib/node_modules/@anthropic-ai/claude-code/bin/claude.exe
```

`file /usr/local/bin/claude` 显示它是 macOS arm64 Mach-O executable。

### 5.2 npm 包形态

安装包目录：

```text
/usr/local/lib/node_modules/@anthropic-ai/claude-code
```

关键文件：

```text
LICENSE.md
README.md
bin/claude.exe
cli-wrapper.cjs
install.cjs
package.json
sdk-tools.d.ts
```

`package.json` 显示：

```json
{
  "name": "@anthropic-ai/claude-code",
  "version": "2.1.150",
  "bin": {
    "claude": "bin/claude.exe"
  },
  "optionalDependencies": {
    "@anthropic-ai/claude-code-darwin-arm64": "2.1.150",
    "@anthropic-ai/claude-code-darwin-x64": "2.1.150",
    "@anthropic-ai/claude-code-linux-x64": "2.1.150",
    "@anthropic-ai/claude-code-linux-arm64": "2.1.150"
  }
}
```

`cli-wrapper.cjs` 的职责是检测平台，定位对应 optional dependency 的原生 binary，然后启动它。正常安装时，postinstall 会把原生 binary 放到 `bin/claude.exe`，因此 wrapper 通常只是 fallback launcher。

### 5.3 binary strings 观察

对 `/usr/local/lib/node_modules/@anthropic-ai/claude-code/bin/claude.exe` 做 targeted strings 观察，可以看到以下记忆相关符号或文案：

```text
CLAUDE.md
MEMORY.md
autoMemoryEnabled
autoMemoryDirectory
claudeMd
claudeMdExcludes
CLAUDE_CODE_DISABLE_AUTO_MEMORY
CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD
getClaudeMds
setCachedClaudeMdContent
getCachedClaudeMdContent
setAdditionalDirectoriesForClaudeMd
getAdditionalDirectoriesForClaudeMd
```

还可观察到相关文案片段，例如：

```text
Custom directory path for auto-memory storage...
CLAUDE.md-style instructions injected as organization-managed memory...
Glob patterns or absolute paths of CLAUDE.md files to exclude from loading...
Cannot write to memory while it is toggled off...
Cannot read memory while it is toggled off...
```

这些静态痕迹与官方文档描述相互印证：Claude Code 内部确实有 `CLAUDE.md` discovery/cache、additional directories、auto memory toggle、memory directory、excludes 等概念。

需要注意：strings 输出不是源码，不能证明完整控制流、优先级实现细节或所有边界行为。准确行为仍应以官方文档和可重复 CLI 实验为准。

## 6. 对当前仓库的含义

当前仓库：

```text
/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io
```

已具备较完整的 Claude Code 项目上下文：

```text
/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/CLAUDE.md
/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/.claude/settings.json
/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/.claude/agents/
/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/.claude/commands/
/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/.claude/skills/
/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/.claude/templates/
```

当前 auto memory 目录存在但为空：

```text
/Users/xianguiwang/.claude/projects/-Users-xianguiwang-gopath-src-github-com-wxquare-wxquare-github-io/memory/
```

这意味着现在真正稳定影响 Claude Code 行为的主要是项目 `CLAUDE.md`、`.claude/settings.json`、agents、commands 和 skills，而不是 auto memory。

## 7. 实践建议

### 7.1 把稳定规则写进 `CLAUDE.md`

适合放入 `CLAUDE.md` 的内容：

- 构建、测试、部署命令
- 项目目录结构
- 写作或代码规范
- 不允许修改的目录
- 每次提交前必须执行的验证
- Agent 反复犯错后的修正规则

`CLAUDE.md` 应保持短而具体。过长的规则会增加 context 成本，也会降低模型遵循稳定性。

### 7.2 大型规则拆入 `.claude/rules/`

如果项目规则继续增长，建议新增：

```text
.claude/rules/
```

可以按主题拆分：

```text
.claude/rules/hexo.md
.claude/rules/writing.md
.claude/rules/mdbook.md
.claude/rules/security.md
```

对只适用于某些路径的规则，使用 `paths` front matter，避免每个 session 都加载所有规则。

### 7.3 把偏好和踩坑交给 auto memory

auto memory 适合保存：

- 用户偏好的执行方式
- 某个测试失败的排查经验
- 本地环境特殊性
- 项目中不容易从代码直接看出的约定

可以通过自然语言触发，例如：

```text
remember that this repo requires npm run clean before npm run build after article changes
```

也可以用 `/memory` 查看和编辑。

### 7.4 不要把敏感信息写入记忆

不要把以下内容写入 `CLAUDE.md` 或 auto memory：

- API key、token、cookie
- 私有证书路径或密码
- 内部系统凭证
- 可用于越权访问的具体命令和参数

如果发现 auto memory 写入了不该保存的信息，应立即用 `/memory` 打开对应文件并删除。

### 7.5 用 `claudeMdExcludes` 控制 monorepo 噪声

大型仓库中，多个团队可能各自维护 `CLAUDE.md` 或 `.claude/rules/`。如果 Claude Code 误读了不相关规则，可以在 local settings 中加入：

```json
{
  "claudeMdExcludes": [
    "**/unrelated-team/**/CLAUDE.md",
    "**/vendor/**/.claude/rules/**"
  ]
}
```

建议放在 `.claude/settings.local.json`，避免把个人排除规则提交给所有协作者。

## 8. 一句话总结

Claude Code 的记忆系统不是单一数据库，而是一组上下文来源：

- `CLAUDE.md` 和 `.claude/rules/` 提供显式、稳定、可审查的指令。
- auto memory 提供本地、按项目、可编辑的经验沉淀。
- JSONL transcript 和 file history 支持 session 恢复和撤销，但不是新 session 的主要长期记忆入口。

工程上应把可审计的规则放在 `CLAUDE.md` / `.claude/rules/`，把动态偏好和经验交给 auto memory，并定期用 `/memory` 审计本地记忆内容。
