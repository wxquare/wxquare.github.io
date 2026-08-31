# wxquare.github.io 隐私清理、历史重写与下线设计

## 状态与决策

本文记录 `wxquare.github.io` 的完整隐私清理方案。已选择“方案 1”：先把不允许公开的资料迁移到 `wxquare-private`，再清理 public 当前内容、已发布页面和全部 Git 历史，最后 force-push 重写后的分支。

本设计只定义边界和执行方法，不在设计阶段删除文件、重写历史或推送远端。进入实施前，需要先审核本文并形成逐步执行计划。

## 背景与目标

`wxquare.github.io` 同时保存 Hexo 源码和 GitHub Pages 生成内容。当前 public 仓库中存在不应公开的简历、电话、工作经历与工作资料，也存在一条疑似 API 凭据。仅删除当前分支中的文件不能清除旧 commit、其他发布分支、GitHub 缓存、fork 或其他人的 clone，因此需要同时处理当前内容、发布产物和历史记录。

本次目标是：

- public 中只保留技术博客、书稿和其他明确允许公开的技术内容。
- `source/about/` 只重建一个最小联系页；页面的唯一私人联系内容是现有的单个邮箱。
- public 中不再保留简历、电话、雇主/职位/工作经历、面试工作资料或内部工作资料。
- `source/to_post/` 整体下线，不再从 public 源码、构建产物或历史中出现。
- 被移除的资料在 `wxquare-private` 中保留原路径、校验值和迁移记录。
- 从 public 全部可写 Git refs 中清除已识别的凭据和敏感路径。
- 重新部署后，旧 `to_post` 与旧 About 资料 URL 不再可访问。

## 公开边界

允许继续公开：

- 已发布技术博客与技术转载文章，包括经用户明确确认继续公开的 `source/_posts/other/在腾讯的八年，我的职业思考.md`。
- 系统设计、通用面试方法、AI/Agent 等书稿与章节；它们不包含个人简历或具体面试记录。
- 站点构建源码、主题配置和公开项目说明。
- 一个现有邮箱，作为最小 About/Contact 页的唯一私人联系内容。

禁止继续公开：

- 简历及其 Markdown、HTML、PDF、图片和衍生文件。
- 电话、住址及其他个人联系方式；邮箱是唯一例外。
- 具体雇主、职位、工作经历、绩效、规划、目标、面试记录和内部工作资料。
- `source/to_post/` 下全部内容。
- `source/about/` 的既有内容。
- 用于生成简历的 `scripts/render-resume-pdf.js`。
- 已识别的疑似 API 凭据及其任何历史版本。

本文不对继续公开的技术文章做内容重写；若后续发现某篇技术文章仍包含电话、简历或工作资料，则按同一禁止边界追加清理。

## 当前快照

设计时的本地盘点结果如下：

| 范围 | 已跟踪文件 | 约占用 | 处理方式 |
| --- | ---: | ---: | --- |
| `source/about/` | 28 | 2.5 MB | 完整迁移，public 历史删除，之后重建邮箱页 |
| `source/to_post/` | 31 | 1 MB | 完整迁移并从 public 当前内容与历史删除 |
| `scripts/render-resume-pdf.js` | 1 | 很小 | 迁移并从 public 当前内容与历史删除 |
| 合计 | 60 | 约 3.5 MB | 迁移后清理 |

补充事实：

- `source/about/` 中扫描到一个唯一邮箱，分布在 10 个文件中；重建联系页时只保留这个邮箱。
- 当前已发布的 `master` 中存在 34 个 `to_post` 文件/页面和 14 个 `about` 文件/页面；`gh-pages` 中这两个路径当前均为 0。
- 当前远端分支为 `hexo`、`master`、`gh-pages`；当前没有 Git tag。
- `_config.yml` 当前通过 Hexo Git deploy 把站点发布到 `master`。
- README 中的疑似凭据已由 commit `347d2f92` 从当前 `hexo` 删除并推送到远端；旧 commit 中仍存在，必须继续执行历史重写。
- 本机当前没有 `git-filter-repo`；实施前必须安装并确认版本不低于 GitHub 当前建议的 2.47。

## 目标目录状态

private 迁移归档使用固定批次目录，并保留 public 原始相对路径：

```text
wxquare-private/
└── 迁移归档/
    └── wxquare.github.io/
        └── 2026-08-31/
            ├── source/
            │   ├── about/
            │   └── to_post/
            ├── scripts/
            │   └── render-resume-pdf.js
            └── MIGRATION-MANIFEST.md
```

public 清理后的相关结构为：

```text
wxquare.github.io/
├── source/
│   └── about/
│       └── index.md          # 页面正文只含允许公开的邮箱
├── README.md                 # 不再链接简历/工作资料，不含凭据
├── AGENTS.md                 # 不再把 about 定义为简历目录
├── CONTRIBUTING.md           # 不再引导写入简历/工作资料
└── _config.yml               # 允许最小 About 页由 Hexo 正常渲染
```

`source/to_post/` 和 `scripts/render-resume-pdf.js` 在 public 中不再存在。最小 About 页可保留构建必需的 front matter 和站点已有公共导航，但其页面正文不得包含邮箱之外的私人资料。

## private 迁移设计

迁移先于任何 public 删除或历史重写，避免唯一副本丢失。执行时按以下规则处理：

1. 创建批次目录 `迁移归档/wxquare.github.io/2026-08-31/`。
2. 按原始相对路径复制 60 个已跟踪文件，不改文件名、不转换格式、不重写正文。
3. 生成 `MIGRATION-MANIFEST.md`，记录来源仓库、来源分支与 commit、迁移日期、原始路径、目标路径、文件大小和 SHA-256。
4. 对源文件与归档文件逐个比较 SHA-256，并核对文件数。
5. 只有在 60/60 文件校验通过后，才允许进入 public 清理。

归档是 private 仓库中的普通受控内容。它不会依赖 public 被重写后的 commit 才能恢复；如需回滚，可按 manifest 找回原文件，但不得直接重新发布到 public。

## public 当前内容清理

历史重写会删除这些路径在所有旧 commit 中的版本，因此最小 About 页应在重写完成后以一个新的正常 commit 重建。该 commit 同时完成：

- 新建只含允许邮箱的 `source/about/index.md`。
- 删除 `_config.yml` 中对 `about/**` 的 `skip_render`，使最小联系页由 Hexo 正常渲染。
- 更新 README，删除简历/工作资料链接、敏感目录说明和疑似凭据。
- 更新 `AGENTS.md` 与 `CONTRIBUTING.md`，禁止把 public About 当作简历或工作资料目录。
- 全仓扫描邮箱、电话、简历关键词、公司/职位/工作资料关键词和常见密钥格式，再人工复核命中。

扫描结果必须区分两类：技术文章中允许继续公开的普通讨论，以及能识别个人或工作信息的禁止内容。不能仅靠关键词自动删除文章。

## Git 历史重写原理

Git commit 包含父 commit、目录树和元数据的哈希。历史中任意文件被删除或替换后，该 commit 的哈希会变化；其所有后代因为父哈希变化，也会得到新的 commit ID。首个受影响提交之前、内容与父关系都未变化的 commit 可以保留原 ID。因此历史重写不是“修改旧 commit”，而是从受影响点开始生成一张内容相似但 ID 不同的新提交图。

普通 `git push` 会拒绝把远端分支从旧提交图移动到不相容的新提交图。force-push 的作用是明确允许远端分支引用改指向重写后的 commit。它不会神奇地删除其他人的 clone、fork、GitHub 缓存或只读 PR refs；它只替换我们有权限更新的远端 refs。

这也是本方案要求协作者重新 clone 的原因：旧 clone 仍持有旧提交图，如果有人把旧分支 merge 回新分支，已经清除的敏感历史会被重新引入。

## 历史重写范围与方法

历史重写在全新临时 mirror clone 中进行，不在当前有未提交 README 改动的工作区直接执行。执行前冻结 public 写入并保存所有远端 ref 的原始 SHA，作为审计与应急恢复记录。

使用 `git-filter-repo >= 2.47`，从所有被重写 refs 删除以下路径：

```text
source/about/
source/to_post/
scripts/render-resume-pdf.js
about/
to_post/
```

前 3 条覆盖 Hexo 源码历史，后 2 条覆盖 `master`/`gh-pages` 中的生成站点历史。同时通过临时 `--replace-text` 规则精确替换已识别凭据。替换文件只在受限临时目录中存在，不写入仓库、不在日志中输出凭据原文。

这 5 个路径是已知的最小范围，不是无条件封闭列表。若全仓或全历史扫描在其他路径发现个人电话、简历、具体工作资料或凭据，必须在推送前把对应路径或精确文本加入同一次重写，再重新验证。经用户明确批准继续公开的文章不因普通职业关键词自动删除。

待执行命令的结构如下；这是说明性示例，不是本设计阶段要运行的命令：

```bash
git filter-repo \
  --sensitive-data-removal \
  --invert-paths \
  --path source/about/ \
  --path source/to_post/ \
  --path scripts/render-resume-pdf.js \
  --path about/ \
  --path to_post/ \
  --replace-text /private/tmp/wxquare-secret-replacements.txt
```

重写后先在 mirror 中验证，不立即推送：

- 检查 `.git/filter-repo/changed-refs` 和重写统计。
- 全历史搜索 5 个禁止路径，结果必须为零。
- 全历史搜索已识别凭据的指纹与原文，结果必须为零；日志不打印原文。
- 扫描重写后的 blobs 中是否仍含电话、简历和工作资料。
- 确认技术博客、书稿及 `hexo`、`master`、`gh-pages` 的预期历史仍可达。
- 对比冻结时的远端 ref 列表，解释每个会改变或删除的 ref。

## force-push 设计

验证通过后，重新添加被 `git-filter-repo` 移除的 remote，并在解除相关分支保护后，从 mirror 执行 GitHub 官方敏感数据清理流程中的镜像强推：

```bash
git push --force --mirror origin
```

`--mirror` 会把本地 mirror 的所有 refs 映射到远端，`--force` 允许 refs 指向重写后的提交图。它比逐分支强推覆盖更完整，也更危险：本地不存在的远端可写 ref 可能被删除，因此必须使用新鲜 mirror clone，并在推送前逐项对比 refs。GitHub 的 `refs/pull/*` 是只读的，相关拒绝是预期现象；其他拒绝均视为失败并停止后续部署。

当前没有 tag，所以本批次没有实际 tag 需要更新；`--mirror` 仍会覆盖未来检查时发现的所有 refs。若 mirror 前检查发现新增 tag、分支或未合并 PR，必须暂停、更新设计输入并重新确认推送范围。

force-push 完成后：

1. 从远端重新 clone `hexo`，不继续使用旧工作区。
2. 创建 public 当前内容清理 commit 和最小 About 页。
3. 正常 push `hexo`。
4. 从干净源码执行 Hexo clean/build/deploy，覆盖 `master` 的生成内容；若 `gh-pages` 仍承担任何发布用途，则同步生成或明确删除该远端分支，不能保留旧页面。
5. 恢复分支保护，并禁止旧提交图再次进入。

## 密钥处置边界

历史重写不能让已经泄露的凭据失效。执行 Git 清理前必须先到对应服务撤销或轮换该凭据，并检查该服务的用量、账单和访问日志。当前可把它判断为一个 `sk-` 前缀 API key。仓库附近代码的默认 API 地址指向 DeepSeek，因此它较可能是 DeepSeek key；但多家服务使用相同前缀，格式和上下文都不足以最终定性，必须在实际服务控制台确认。

撤销/轮换与 Git 清理是两个独立动作：

- 撤销/轮换阻止凭据继续被使用。
- 历史重写降低凭据及其他隐私资料继续从 public Git 历史被发现的风险。

完成强推后仍需根据 [GitHub 官方敏感数据清理说明](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository) 检查 PR refs、cached views、fork 和 GitHub 服务端残留；必要时向 GitHub Support 提供仓库、首个受影响 commit 和受影响 PR 信息，请求清除不可由仓库所有者直接更新的引用与缓存。

## 部署与 URL 下线

构建前清空 Hexo 生成目录，避免旧页面因增量构建残留。部署完成后验证：

- `/about/` 可访问，页面正文只出现允许公开的邮箱，不出现简历、电话、公司、职位或工作经历。
- 所有已记录的 `/to_post/...` URL 返回 404，而不是从缓存或旧发布分支继续返回 200。
- 旧 About 简历、PDF、HTML 和附件 URL 返回 404。
- `master` 与实际 GitHub Pages 配置一致，不从 `gh-pages` 或其他旧 ref 继续提供内容。
- GitHub Pages/CDN 刷新等待期后再次从未登录会话验证。

如果 GitHub Pages 设置与 `_config.yml` 显示的 `master` 不一致，实施时必须先停下，记录实际 Pages source，再调整部署步骤，不能假定配置文件就是线上真实状态。

## 验证矩阵

| 层面 | 验证项 | 通过条件 |
| --- | --- | --- |
| private 归档 | 文件数、大小、SHA-256、manifest | 60/60 文件一致 |
| public 当前树 | 禁止路径与敏感信息扫描 | 除最小邮箱页外无禁止内容 |
| public 全历史 | 路径、凭据、电话、简历/工作资料扫描 | 被禁路径和凭据为零；其他命中人工确认 |
| Git refs | 分支/tag 对比 | 所有可写 refs 指向新历史，无意外 ref 删除 |
| 构建 | Hexo clean/build | 成功，无失效内部链接或敏感产物 |
| 线上站点 | About 与旧 URL | About 仅邮箱；旧资料 URL 为 404 |
| 外部残留 | PR refs、cached views、fork | 已核查并记录；需要时提交 GitHub Support 请求 |

最终验证还包括 `git diff --check`、构建产物二次扫描，以及从全新 clone 复验 Git 历史。任何一层失败都不得宣告清理完成。

## 风险、回滚与协作规则

主要风险：

- 首个受影响 commit 及其后代 ID 变化，相关旧链接、签名、PR diff 和基于 SHA 的引用可能失效。
- force-push 若基于过期 mirror，可能覆盖他人在冻结后新增的提交。
- fork、clone、PR refs 和缓存可能继续保存旧对象。
- 旧协作者误 merge 会重新污染新历史。
- 仅重新部署而未 clean，可能把旧静态文件继续留在线上。

控制措施：

- force-push 前冻结写入并再次 `fetch`、对比远端 refs；发现变化立即停止。
- 保存原 ref SHA 清单和操作日志，但不在 public 保存敏感内容或凭据原文。
- private 归档校验完成后才清理 public。
- 推送后要求所有协作者删除旧 clone 并重新 clone；不得 merge 旧分支，确需保留的独立工作只能在人工检查后 cherry-pick 到新历史。
- 启用 GitHub secret scanning/push protection，并在本地提交前增加密钥扫描。

回滚只适用于操作故障，不用于重新公开敏感历史。若强推后发现技术内容缺失，应从受控的临时备份中只提取允许公开的文件并形成新 commit；不得把原分支整体 force-push 回远端。若部署失败，保留已清理的 Git 历史，修复构建后重新部署。

## 验收标准

- 60 个目标文件完整迁移到 private 批次目录，manifest 与 SHA-256 全部一致。
- public 当前源码不存在 `source/to_post/`、旧 `source/about/` 内容和简历生成脚本。
- `/about/` 页面正文只包含允许公开的现有邮箱。
- README、`AGENTS.md`、`CONTRIBUTING.md` 和 `_config.yml` 与新的公开边界一致。
- public 所有可写 Git refs 中不存在 5 个已知禁止路径、已识别凭据，以及扫描确认的其他个人电话、简历或具体工作资料。
- 技术博客、技术转载和书稿仍然存在且可构建。
- 线上旧 `to_post`、简历、PDF、HTML 和附件 URL 全部下线。
- 密钥已在服务端撤销或轮换，并完成异常使用检查。
- GitHub PR refs、缓存、fork 与协作者 clone 的后续处置已记录。
- 所有协作者收到“必须重新 clone、不得 merge 旧历史”的说明。
