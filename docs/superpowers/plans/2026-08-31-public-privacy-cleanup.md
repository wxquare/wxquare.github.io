# wxquare.github.io Public Privacy Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 public 中除允许邮箱和已批准技术内容以外的简历、电话、具体工作资料、`to_post` 内容及疑似 API key 迁移到 private，并从当前站点、全部可写 Git 历史和线上页面下线。

**Architecture:** 先在 `wxquare-private` 建立逐文件可校验的归档，再在全新 mirror clone 中使用 `git-filter-repo` 重写全部 refs；离线验证通过后才 force-push。随后从重写后的 `hexo` 新 clone 重建最小 About 页、清理 public 文档和配置、clean build/deploy，并从全新 clone 与线上 URL 双重复验。

**Tech Stack:** Git、git-filter-repo >= 2.47、Hexo、Node.js/npm、GitHub Pages、Bash、Ruby（仅用于不打印敏感原文的临时校验）。

## Global Constraints

- public 唯一允许保留的私人联系内容是 `xianguiwang0316@gmail.com`。
- 保留已发布技术博客、书稿和用户明确批准的 `source/_posts/other/在腾讯的八年，我的职业思考.md`。
- public 不允许出现简历、电话、具体雇主/职位/工作经历、绩效、规划、目标、个人面试记录或内部工作资料。
- 必须迁移并从 public 历史删除 `source/about/`、`source/to_post/`、`scripts/render-resume-pdf.js`、生成站点的 `about/` 和 `to_post/`。
- 任何删除 public 内容的动作都必须晚于 private 60/60 文件校验。
- `git-filter-repo` 版本必须不低于 2.47，并在全新 mirror clone 中运行。
- 不在命令、日志、计划或 commit message 中打印疑似 key 原文；只使用 SHA-256 指纹或 `[REDACTED]`。
- 历史重写不能替代服务端撤销/轮换；force-push 前必须确认 key 已失效。
- 不修改或提交 `wxquare-private/.DS_Store`。
- 远端 refs 在冻结后发生任何变化，必须停止 force-push 并重新核对。
- 不把旧 public 分支整体回滚；技术内容缺失时只从受控副本提取允许公开的文件形成新 commit。

---

### Task 1: 冻结基线并核对线上发布源

**Files:**
- Create: `/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md`

**Interfaces:**
- Consumes: 已批准的设计文档与当前本地 refs。
- Produces: 后续 mirror push 使用的远端 ref、Pages source、旧 URL 和工作区基线。

- [ ] **Step 1: 验证两个工作区只有已知改动**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io status --short --branch
git -C /Users/xianguiwang/Projects/wxquare-private status --short --branch
```

Expected: public 工作区 clean；private 只有既有 `.DS_Store` 改动。当前 README 删除 key 的 commit 是 `347d2f92`，已经位于 `origin/hexo`。发现其他用户改动立即停止。

- [ ] **Step 2: 把已提交的设计与实施计划正常推送到 hexo**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io push origin hexo
```

Expected: `origin/hexo` 包含设计与实施计划；当前 README 已不含 key。此后才开始冻结远端 refs。

- [ ] **Step 3: 获取远端最新 refs 并核对本地分支**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io ls-remote --heads --tags origin
git -C /Users/xianguiwang/Projects/wxquare.github.io branch --all --no-color
git -C /Users/xianguiwang/Projects/wxquare.github.io tag --list
```

Expected: 可写分支为 `hexo`、`master`、`gh-pages`，没有 tag。若结果不同，记录全部新增 refs，并暂停强推范围确认。

- [ ] **Step 4: 核对 GitHub Pages 真实发布源**

Run:

```bash
curl -fsSL https://api.github.com/repos/wxquare/wxquare.github.io/pages
```

Expected: 返回 Pages 配置 JSON。将 `source.branch` 和 `source.path` 与 `_config.yml` 的 `deploy.branch: master` 对照；如果匿名 API 不返回配置，使用已登录 GitHub 页面只读检查 Settings → Pages。

- [ ] **Step 5: 生成旧线上 URL 基线**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io -c core.quotePath=false ls-tree -r --name-only origin/master about to_post
```

Expected: `about/` 14 个文件、`to_post/` 34 个文件。把每个路径映射为 `https://wxquare.github.io/<path>`，作为部署后的 404 清单。

- [ ] **Step 6: 写入操作基线**

先运行 `mkdir -p /Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31`，再使用 `apply_patch` 创建 `OPERATION-BASELINE.md`，精确记录：执行日期 `2026-08-31`、public/private 当前 commit、远端每个 heads/tags SHA、Pages source、14+34 个旧发布路径、疑似 key 的 SHA-256 前 12 位 `fc29674c5e20`，以及“key 原文不得写入此文件”。

- [ ] **Step 7: 验证基线文档不含 key 或电话**

Run:

```bash
rg -l -P 'sk-[A-Za-z0-9_-]{20,}|(?<!\d)1[3-9]\d{9}(?!\d)' /Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md
```

Expected: no output。

- [ ] **Step 8: 提交无敏感原文的基线**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare-private add 迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md
git -C /Users/xianguiwang/Projects/wxquare-private diff --cached --check
git -C /Users/xianguiwang/Projects/wxquare-private commit -m "docs: record privacy cleanup baseline"
```

Expected: commit 只包含操作基线；`.DS_Store` 未暂存。

### Task 2: 迁移 60 个敏感文件到 wxquare-private

**Files:**
- Create: `/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/source/about/**`
- Create: `/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/source/to_post/**`
- Create: `/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/scripts/render-resume-pdf.js`
- Create: `/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/MIGRATION-MANIFEST.md`

**Interfaces:**
- Consumes: public `hexo` 工作树中的 60 个已跟踪文件。
- Produces: 保留原相对路径、大小和 SHA-256 的 private 归档。

- [ ] **Step 1: 再次确认源文件清单为 60**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io ls-files source/about source/to_post scripts/render-resume-pdf.js
git -C /Users/xianguiwang/Projects/wxquare.github.io ls-files source/about source/to_post scripts/render-resume-pdf.js | wc -l
```

Expected: 28 个 `source/about`、31 个 `source/to_post`、1 个脚本，合计 60。

- [ ] **Step 2: 按 tracked file 清单复制并保留相对路径**

Run:

```bash
mkdir -p /Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31
git -C /Users/xianguiwang/Projects/wxquare.github.io ls-files -z source/about source/to_post scripts/render-resume-pdf.js | rsync -a --from0 --files-from=- /Users/xianguiwang/Projects/wxquare.github.io/ /Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/
```

Expected: 只复制这 60 个已跟踪文件，不复制 public `.git`、`node_modules` 或其他目录。

- [ ] **Step 3: 逐文件做字节级比较**

Run:

```bash
set -o pipefail
git -C /Users/xianguiwang/Projects/wxquare.github.io ls-files -z source/about source/to_post scripts/render-resume-pdf.js | while IFS= read -r -d '' rel; do cmp -s "/Users/xianguiwang/Projects/wxquare.github.io/$rel" "/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/$rel" || exit 1; done
```

Expected: exit 0, no output。

- [ ] **Step 4: 收集 manifest 数据**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io ls-files -z source/about source/to_post scripts/render-resume-pdf.js | while IFS= read -r -d '' rel; do shasum -a 256 "/Users/xianguiwang/Projects/wxquare.github.io/$rel"; done
```

Expected: 60 行 SHA-256。使用 `apply_patch` 创建 `MIGRATION-MANIFEST.md`，每行记录原路径、目标相对路径、字节数和对应 SHA-256；文件头记录来源分支 `hexo` 与来源 commit。

- [ ] **Step 5: 验证 manifest 与归档**

Run:

```bash
find /Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/source /Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/scripts -type f | wc -l
rg -P -c '^\| `(?:source|scripts)/' /Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/MIGRATION-MANIFEST.md
```

Expected: 两个命令都得到 60。

- [ ] **Step 6: 提交 private 归档**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare-private add 迁移归档/wxquare.github.io/2026-08-31
git -C /Users/xianguiwang/Projects/wxquare-private diff --cached --check
git -C /Users/xianguiwang/Projects/wxquare-private commit -m "archive: migrate private content from public site"
```

Expected: commit 只包含迁移批次目录；`.DS_Store` 仍未暂存。

### Task 3: 完成 key 服务端处置门禁

**Files:**
- Modify: `/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md`

**Interfaces:**
- Consumes: 指纹 `fc29674c5e20` 和仓库上下文中 DeepSeek 默认 API 地址线索。
- Produces: force-push 前的服务端撤销/轮换确认。

- [ ] **Step 1: 确认服务归属**

在 DeepSeek、阿里云 DashScope 及其他实际使用过的服务控制台中按创建时间和 key 尾部识别目标；不把 key 原文粘贴到聊天、Issue 或文档。仓库上下文只提供“较可能是 DeepSeek”的线索，不能代替控制台确认。

- [ ] **Step 2: 撤销或轮换并检查异常使用**

在确认的服务控制台撤销旧 key；如果仍有合法调用，创建新 key 并只存入环境变量或密钥管理服务。检查从首次提交到当前日期的调用量、账单和来源 IP。

- [ ] **Step 3: 记录无敏感原文的确认**

使用 `apply_patch` 在 `OPERATION-BASELINE.md` 记录服务名、撤销/轮换时间、操作者确认和异常使用结论，只引用指纹 `fc29674c5e20`。

- [ ] **Step 4: 提交撤销/轮换确认**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare-private add 迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md
git -C /Users/xianguiwang/Projects/wxquare-private diff --cached --check
git -C /Users/xianguiwang/Projects/wxquare-private commit -m "security: record leaked key revocation"
```

Expected: commit 不含 key 原文；`.DS_Store` 未暂存。

- [ ] **Step 5: 设置强推门禁**

Expected: 没有撤销/轮换确认时，允许继续做本地 mirror 重写和验证，但禁止执行 Task 7 的任何远端写操作。

### Task 4: 安装工具并建立全新 mirror clone

**Files:**
- Create: `/private/tmp/wxquare-history-rewrite-2026-08-31.git`
- Create: `/private/tmp/wxquare-secret-replacements-2026-08-31.txt`

**Interfaces:**
- Consumes: Task 1 已推送到 `origin/hexo` 的设计与实施计划，以及冻结后的远端 refs。
- Produces: 冻结后的全新 bare mirror 与权限为 0600 的临时替换规则。

- [ ] **Step 1: 安装并验证 git-filter-repo**

Run:

```bash
brew install git-filter-repo
git-filter-repo --version
brew list --versions git-filter-repo
```

Expected: 安装成功，版本不低于 2.47。

- [ ] **Step 2: 确认固定临时目标不存在**

Run:

```bash
test ! -e /private/tmp/wxquare-history-rewrite-2026-08-31.git
test ! -e /private/tmp/wxquare-secret-replacements-2026-08-31.txt
```

Expected: 两个命令 exit 0；如果已存在则停止，不能覆盖未知目录或文件。

- [ ] **Step 3: 从 GitHub 创建 mirror**

Run:

```bash
git clone --mirror https://github.com/wxquare/wxquare.github.io.git /private/tmp/wxquare-history-rewrite-2026-08-31.git
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git show-ref
```

Expected: `refs/heads/hexo`、`refs/heads/master`、`refs/heads/gh-pages` 与 Task 1 冻结 SHA 一致。

- [ ] **Step 4: 从旧 README 安全生成精确替换规则**

Run:

```bash
umask 077
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git show 762c15c8a9d9c81efbae0c5d65577e4bc9268cf5:README.md | ruby -e 's=STDIN.read.scan(/sk-[0-9a-f]{32}/).uniq; abort "expected exactly one credential" unless s.length == 1; File.open(ARGV.fetch(0), "wx", 0600) { |f| f.puts("literal:#{s.first}==>REMOVED_SECRET") }' /private/tmp/wxquare-secret-replacements-2026-08-31.txt
stat -f '%Lp %N' /private/tmp/wxquare-secret-replacements-2026-08-31.txt
```

Expected: 不打印 key；规则文件权限为 `600`。

### Task 5: 在 mirror 中重写全部敏感历史

**Files:**
- Modify: `/private/tmp/wxquare-history-rewrite-2026-08-31.git/**`

**Interfaces:**
- Consumes: fresh mirror 和从历史泄露 commit 提取的精确 replacement file。
- Produces: 删除 5 个已知路径并替换疑似 key 的新提交图。

- [ ] **Step 1: 运行 filter-repo**

Run:

```bash
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git filter-repo --sensitive-data-removal --invert-paths --path source/about/ --path source/to_post/ --path scripts/render-resume-pdf.js --path about/ --path to_post/ --replace-text /private/tmp/wxquare-secret-replacements-2026-08-31.txt
```

Expected: exit 0；生成 `filter-repo/changed-refs` 和敏感数据清理报告；`origin` remote 被安全移除是预期行为。

- [ ] **Step 2: 检查 changed refs**

Run:

```bash
sed -n '1,240p' /private/tmp/wxquare-history-rewrite-2026-08-31.git/filter-repo/changed-refs
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git show-ref
```

Expected: `hexo`、`master` 和任何包含敏感对象的 refs 被明确列出；没有无法解释的 ref 消失。

- [ ] **Step 3: 确认禁止路径在全部可达对象中为零**

Run:

```bash
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git rev-list --objects --all | rg ' (source/about/|source/to_post/|scripts/render-resume-pdf\.js$|about/|to_post/)'
```

Expected: no output。

- [ ] **Step 4: 精确验证旧 key 不再可达**

Run:

```bash
OLD_CREDENTIAL=$(ruby -e 'line=File.read(ARGV.fetch(0)); print line.sub(/\Aliteral:/, "").sub(/==>REMOVED_SECRET\s*\z/, "")' /private/tmp/wxquare-secret-replacements-2026-08-31.txt)
test -z "$(git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git log --all -S"$OLD_CREDENTIAL" --format=%H)"
unset OLD_CREDENTIAL
```

Expected: exit 0，不打印 key。

- [ ] **Step 5: 扫描其他密钥模式**

Run:

```bash
for rev in $(git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git rev-list --all); do git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git grep -I -l -E 'sk-[A-Za-z0-9_-]{20,}|AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9]{20,}' "$rev" -- . || true; done
```

Expected: 已识别真实 key 为零；示例或文档命中逐项人工复核，不能把真实凭据当示例放行。

- [ ] **Step 6: 检查路径外个人资料**

Run:

```bash
umask 077
ruby -e 'root=ARGV.fetch(0); out=ARGV.fetch(1); phones=Dir.glob("#{root}/**/*").select { |p| File.file?(p) }.flat_map { |p| data=File.binread(p); next [] if data.include?("\x00"); data.force_encoding("UTF-8").scrub.scan(/(?<!\d)1[3-9]\d{9}(?!\d)/) }.uniq.sort; File.open(out, "wx", 0600) { |f| phones.each { |v| f.puts(v) } }' /Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31 /private/tmp/wxquare-phone-values-2026-08-31.txt
while IFS= read -r phone; do test -z "$(git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git log --all -S"$phone" --format=%H)" || exit 1; done < /private/tmp/wxquare-phone-values-2026-08-31.txt
for rev in $(git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git rev-list --all); do git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git grep -I -l -E '简历|手机号|电话|工作资料|面试材料|绩效|工作规划|工作目标|面试记录' "$rev" -- . ':(exclude)source/_posts/other/在腾讯的八年，我的职业思考.md' || true; done
```

Expected: 电话精确值检查 exit 0；关键词只列文件路径，逐项复核后仅允许规则文档、通用技术语境和明确批准内容。任何路径外真实个人资料都追加到同一次 filter-repo 规则并从 Step 1 重做。

- [ ] **Step 7: 删除临时 key 规则**

Run:

```bash
rm /private/tmp/wxquare-secret-replacements-2026-08-31.txt
test ! -e /private/tmp/wxquare-secret-replacements-2026-08-31.txt
rm /private/tmp/wxquare-phone-values-2026-08-31.txt
test ! -e /private/tmp/wxquare-phone-values-2026-08-31.txt
```

Expected: 临时 key 原文副本被删除，mirror 中只剩重写后的对象。

### Task 6: 离线验证技术内容与分支完整性

**Files:**
- Modify: `/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md`

**Interfaces:**
- Consumes: 重写后的 mirror。
- Produces: 可进入 force-push 的验证证据。

- [ ] **Step 1: 验证三个分支和零 tag**

Run:

```bash
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git for-each-ref --format='%(refname) %(objectname)' refs/heads refs/tags
```

Expected: `hexo`、`master`、`gh-pages` 存在；没有 tag。

- [ ] **Step 2: 验证批准内容仍存在**

Run:

```bash
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git ls-tree -r --name-only refs/heads/hexo | rg '^(source/_posts/other/在腾讯的八年，我的职业思考\.md|books/ai-book/src/|books/system-design-architecture-book/)'
```

Expected: 明确批准文章存在，两个书稿目录均有内容。

- [ ] **Step 3: 验证待重建路径当前为空**

Run:

```bash
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git ls-tree -r --name-only refs/heads/hexo source/about source/to_post scripts/render-resume-pdf.js
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git ls-tree -r --name-only refs/heads/master about to_post
```

Expected: no output。

- [ ] **Step 4: 记录 changed refs 与首次变化 commit**

使用 `apply_patch` 把 `filter-repo/changed-refs` 摘要、首次受影响 commit、重写后 branch SHA、受影响 PR refs 数量写入 private `OPERATION-BASELINE.md`；不写 key 原文。

- [ ] **Step 5: 重新获取远端 refs 做冻结检查**

Run:

```bash
git ls-remote --heads --tags https://github.com/wxquare/wxquare.github.io.git
```

Expected: 与 Task 1 基线完全一致；任何 SHA 或 ref 变化都停止 Task 7。

- [ ] **Step 6: 提交离线重写证据**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare-private add 迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md
git -C /Users/xianguiwang/Projects/wxquare-private diff --cached --check
git -C /Users/xianguiwang/Projects/wxquare-private commit -m "docs: record offline history rewrite verification"
```

Expected: commit 不含任何敏感原文；`.DS_Store` 未暂存。

### Task 7: force-push 重写后的全部 refs

**Files:**
- Modify: GitHub remote `wxquare/wxquare.github.io` refs
- Modify: `/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md`

**Interfaces:**
- Consumes: Task 3 的 key 撤销确认、Task 6 的全绿验证和未变化远端基线。
- Produces: GitHub 上指向新提交图的 `hexo`、`master`、`gh-pages`。

- [ ] **Step 1: 核验全部门禁**

Expected: private 60/60 归档通过；key 已撤销/轮换；mirror 禁止路径和旧 key 为零；技术内容存在；远端 refs 未变化。任一条件不满足都不得继续。

- [ ] **Step 2: 重新添加 origin 并核对 push 预览**

Run:

```bash
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git remote add origin https://github.com/wxquare/wxquare.github.io.git
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git push --mirror --force --dry-run origin
```

Expected: 只重写已审阅 refs。若出现新增/删除未知分支或 tag，停止。

- [ ] **Step 3: 执行官方镜像强推**

Run:

```bash
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git push --force --mirror origin
```

Expected: `hexo`、`master`、`gh-pages` 更新成功。只有只读 `refs/pull/*` 拒绝可按 GitHub 官方说明视为预期；任何其他拒绝都停止部署并调查。

- [ ] **Step 4: 从远端复验新 SHA**

Run:

```bash
git ls-remote --heads --tags https://github.com/wxquare/wxquare.github.io.git
```

Expected: 三个分支指向 Task 6 记录的新提交；没有意外 refs。

- [ ] **Step 5: 记录强推结果**

使用 `apply_patch` 在 private `OPERATION-BASELINE.md` 记录时间、推送结果、新 SHA、预期/非预期拒绝和后续动作。

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare-private add 迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md
git -C /Users/xianguiwang/Projects/wxquare-private diff --cached --check
git -C /Users/xianguiwang/Projects/wxquare-private commit -m "docs: record public history force-push"
```

Expected: commit 不含敏感原文；`.DS_Store` 未暂存。

### Task 8: 从新历史重建最小 public 源码

**Files:**
- Create: `/private/tmp/wxquare-sanitized-source-2026-08-31/source/about/index.md`
- Modify: `/private/tmp/wxquare-sanitized-source-2026-08-31/README.md`
- Modify: `/private/tmp/wxquare-sanitized-source-2026-08-31/AGENTS.md`
- Modify: `/private/tmp/wxquare-sanitized-source-2026-08-31/CONTRIBUTING.md`
- Modify: `/private/tmp/wxquare-sanitized-source-2026-08-31/_config.yml`

**Interfaces:**
- Consumes: GitHub 重写后的 `hexo`。
- Produces: 只保留邮箱联系页、无敏感目录指引的 public 源码 commit。

- [ ] **Step 1: 新 clone 重写后的 hexo**

Run:

```bash
test ! -e /private/tmp/wxquare-sanitized-source-2026-08-31
git clone --branch hexo --single-branch https://github.com/wxquare/wxquare.github.io.git /private/tmp/wxquare-sanitized-source-2026-08-31
```

Expected: fresh clone 成功；禁止路径不存在。

- [ ] **Step 2: 创建最小 About 页**

使用 `apply_patch` 创建以下精确内容：

```markdown
---
title: Contact
date: 2026-08-31
---

xianguiwang0316@gmail.com
```

- [ ] **Step 3: 清理 Hexo 渲染配置**

使用 `apply_patch` 从 `_config.yml` 删除：

```yaml
skip_render:
  - "about/**"
```

Expected: About Markdown 由 Hexo 正常渲染。

- [ ] **Step 4: 更新 public 文档边界**

使用 `apply_patch`：

- 从 README 精选内容删除两个 `source/about/material/` 链接。
- 把 README 目录树中的 About 说明改为 `联系页（仅邮箱）`。
- 把 README“电商架构与性能优化”改为只描述公开技术案例，不称为个人面试材料。
- 把 `AGENTS.md` 的 `source/about/` 说明改为“最小联系页，仅允许邮箱；不得存放简历、电话或工作资料”。
- 把 `CONTRIBUTING.md` 的 `source/about/` 说明改为“minimal contact page; email only”。

- [ ] **Step 5: 全仓检查敏感目录引用**

Run:

```bash
rg -n 'source/to_post|source/about/.+|render-resume-pdf|简历|手机号|电话|工作资料|面试材料' README.md AGENTS.md CONTRIBUTING.md _config.yml docs source --glob '!source/_posts/other/在腾讯的八年，我的职业思考.md'
```

Expected: 只剩明确的禁止规则或通用技术文章语境；不存在到已删除个人资料的链接。

- [ ] **Step 6: 提交并推送 sanitized source**

Run:

```bash
git -C /private/tmp/wxquare-sanitized-source-2026-08-31 add source/about/index.md README.md AGENTS.md CONTRIBUTING.md _config.yml
git -C /private/tmp/wxquare-sanitized-source-2026-08-31 diff --cached --check
git -C /private/tmp/wxquare-sanitized-source-2026-08-31 commit -m "privacy: remove personal and work materials from public site"
git -C /private/tmp/wxquare-sanitized-source-2026-08-31 push origin hexo
```

Expected: commit 只包含 5 个目标文件；`hexo` 正常快进 push。

### Task 9: clean build、敏感产物扫描与部署

**Files:**
- Generate: `/private/tmp/wxquare-sanitized-source-2026-08-31/public/**`
- Generate: `/private/tmp/wxquare-sanitized-source-2026-08-31/.deploy_git/**`
- Modify: GitHub remote `master`

**Interfaces:**
- Consumes: sanitized `hexo` source。
- Produces: 不含旧 About/to_post 文件的 GitHub Pages 部署。

- [ ] **Step 1: 安装依赖并 clean build**

Run:

```bash
npm ci
npm run clean
npm run build
```

Working directory: `/private/tmp/wxquare-sanitized-source-2026-08-31`

Expected: 三个命令 exit 0。

- [ ] **Step 2: 验证生成目录**

Run:

```bash
test -f public/about/index.html
test ! -e public/to_post
find public/about -type f
```

Expected: About 只有 `public/about/index.html`，没有 `public/to_post`。

- [ ] **Step 3: 扫描生成内容**

Run:

```bash
rg -l -P 'sk-[A-Za-z0-9_-]{20,}|(?<!\d)1[3-9]\d{9}(?!\d)|resume|简历|工作资料|面试材料' public/about public README.md
```

Expected: About 只含允许邮箱和站点公共 chrome；其他命中逐项复核，不得包含真实 key、电话、简历或具体工作资料。

- [ ] **Step 4: 运行仓库提交前检查**

Run:

```bash
bash bin/pre-commit-check.sh
```

Expected: clean/build 验证通过。

- [ ] **Step 5: 部署 master**

Run:

```bash
npm run deploy
```

Expected: Hexo 从干净构建覆盖 `master`，部署日志无认证或 push 错误。

- [ ] **Step 6: 检查远端 master 不含旧发布路径**

Run:

```bash
git ls-remote --heads https://github.com/wxquare/wxquare.github.io.git master
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git fetch origin master
git -C /private/tmp/wxquare-history-rewrite-2026-08-31.git ls-tree -r --name-only FETCH_HEAD about to_post
```

Expected: 只有 `about/index.html`；没有 `to_post/` 或旧 About 附件。

### Task 10: 线上下线验证与 GitHub 残留处置

**Files:**
- Modify: `/Users/xianguiwang/Projects/wxquare-private/迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md`

**Interfaces:**
- Consumes: Task 1 的 48 个旧 URL 与部署后的 Pages。
- Produces: About 内容证明、旧 URL 404 结果、PR/cache/fork 处置记录。

- [ ] **Step 1: 验证 About 页面**

Run:

```bash
curl -fsSL https://wxquare.github.io/about/
```

Expected: HTTP 200；私人联系内容只有允许邮箱，不含电话、简历、雇主、职位或具体工作经历。

- [ ] **Step 2: 验证全部旧 URL**

对 `OPERATION-BASELINE.md` 中 14 个旧 About 文件和 34 个旧 to_post 文件逐个执行 `curl -sS -o /dev/null -w '%{http_code}'`。对 Markdown 源文件对应的 `.html`/目录 URL 同时检查。

Expected: 全部返回 404；若 GitHub Pages/CDN 尚未刷新，间隔不超过 60 秒复检并记录，不把暂时缓存视为完成。

- [ ] **Step 3: 检查 GitHub refs、PR 与 fork**

Run:

```bash
git ls-remote https://github.com/wxquare/wxquare.github.io.git
```

Expected: 记录 `refs/pull/*` 是否引用 changed refs，并检查仓库 fork 数量。受影响 PR refs、cached commit views 或 fork 无法自行删除时，按 GitHub 官方敏感数据清理文档整理 Support 请求所需信息。

- [ ] **Step 4: 记录线上验证并提交 private 操作日志**

使用 `apply_patch` 写入每个 URL 的状态码、验证时间、Pages source、PR/fork/cache 结论和仍需人工完成的 GitHub Support 动作。

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare-private add 迁移归档/wxquare.github.io/2026-08-31/OPERATION-BASELINE.md
git -C /Users/xianguiwang/Projects/wxquare-private diff --cached --check
git -C /Users/xianguiwang/Projects/wxquare-private commit -m "docs: record public privacy cleanup verification"
```

Expected: `.DS_Store` 未暂存。

### Task 11: 同步本地 public checkout 并做最终复验

**Files:**
- Modify: `/Users/xianguiwang/Projects/wxquare.github.io/.git/**`
- Remove after success: `/private/tmp/wxquare-history-rewrite-2026-08-31.git`
- Remove after success: `/private/tmp/wxquare-sanitized-source-2026-08-31`

**Interfaces:**
- Consumes: 已验证的远端新历史和 private 归档。
- Produces: 与新 `origin/hexo` 一致、无旧可达对象的本地 public checkout。

- [ ] **Step 1: 最后核对旧 checkout 的未提交内容**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io status --short
git -C /Users/xianguiwang/Projects/wxquare.github.io diff -- README.md
```

Expected: 工作区 clean；README key 删除已经包含在远端新历史中。若有任何未提交改动，停止本地同步。

- [ ] **Step 2: fetch 新历史并移动本地分支**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io fetch --prune origin
git -C /Users/xianguiwang/Projects/wxquare.github.io branch -f master origin/master
git -C /Users/xianguiwang/Projects/wxquare.github.io reset --hard origin/hexo
```

Expected: `hexo` 与 `origin/hexo` 一致，`master` 与 `origin/master` 一致。该 reset 只在 Step 1 精确满足时执行。

- [ ] **Step 3: 清除旧 checkout 的 reflog 与不可达旧对象**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io reflog expire --expire=now --all
git -C /Users/xianguiwang/Projects/wxquare.github.io gc --prune=now
git -C /Users/xianguiwang/Projects/wxquare.github.io fsck --full --unreachable --no-reflogs
```

Expected: 不再报告保存旧敏感历史的不可达 commit/blob。

- [ ] **Step 4: 从当前树和全历史最终复验**

Run:

```bash
git -C /Users/xianguiwang/Projects/wxquare.github.io status --short --branch
git -C /Users/xianguiwang/Projects/wxquare.github.io rev-list --objects --all | rg -P ' (source/about/(?!index\.md$)|source/to_post/|scripts/render-resume-pdf\.js$|about/(?!index\.html$)|to_post/)'
rg -l -P 'sk-[A-Za-z0-9_-]{20,}|(?<!\d)1[3-9]\d{9}(?!\d)|resume|简历|工作资料|面试材料' /Users/xianguiwang/Projects/wxquare.github.io/source/about /Users/xianguiwang/Projects/wxquare.github.io/README.md /Users/xianguiwang/Projects/wxquare.github.io/AGENTS.md /Users/xianguiwang/Projects/wxquare.github.io/CONTRIBUTING.md
```

Expected: 工作区 clean；历史中没有禁止路径；只出现允许邮箱和明确的禁止规则文字。

- [ ] **Step 5: 删除受控临时副本**

确认 private 归档、远端历史、clean build 和线上 404 全部通过后，删除两个精确临时路径，不使用通配符：

```bash
rm -rf /private/tmp/wxquare-history-rewrite-2026-08-31.git
rm -rf /private/tmp/wxquare-sanitized-source-2026-08-31
```

Expected: 两个临时目录不存在；可恢复资料只保留在 `wxquare-private/迁移归档/wxquare.github.io/2026-08-31/`。

- [ ] **Step 6: 协作者通知**

通知所有协作者：旧 clone 必须删除并重新 clone；不得 merge 旧分支；确需保留的独立修改只能在人工检查无敏感内容后 cherry-pick。启用 GitHub secret scanning/push protection，并记录 GitHub Support 是否仍需处理 PR refs 或 cached views。
