# /stats - 博客统计报告

本命令只描述统计流程。开始前读取 `AGENTS.md` 第 3、4、10、11 节；分类目录从 `.agents/config/post-categories.json` 读取。

## 执行步骤

### 1. 统计文章总数

```bash
find source/_posts -name "*.md" -type f | wc -l
```

### 2. 按注册表统计分类

```bash
node -e '
const fs = require("node:fs");
const registry = JSON.parse(fs.readFileSync(".agents/config/post-categories.json", "utf8"));
for (const category of registry.categories) {
  console.log(`${category.slug}\t${category.directory}`);
  for (const subdirectory of category.subdirectories) {
    console.log(`${category.slug}/${subdirectory}\t${category.directory}/${subdirectory}`);
  }
}
' | while IFS=$'\t' read -r category directory; do
  printf '%s: ' "$category"
  find "$directory" -name "*.md" -type f | wc -l
done
```

### 3. 统计标签

解析所有文章的 Front Matter，按实际出现次数统计 `tags`；标签命名和允许的分类映射按 `AGENTS.md` 判断。

### 4. 分析文章长度和更新时间

输出总字数、平均字数、最长/最短文章，以及最近 7、30、90 天的更新数量。

```bash
find source/_posts -name "*.md" -type f -exec wc -w {} + | tail -1
find source/_posts -name "*.md" -type f -mtime -30 -exec ls -lh {} \; | sort -k6,7
```

### 5. 统计图表资源

```bash
find source/diagrams \( -name "*.excalidraw" -o -name "*.mmd" \) -type f | wc -l
```

## 输出格式

```text
📊 博客统计报告

📈 总体统计
- 文章总数：X篇
- 总字数：X字
- 平均文章长度：X字
- 图表文件：X个

📂 分类分布
- [按 post-categories.json 输出]

🏷️ 热门标签
- [按使用次数排序]

📅 更新活跃度
- 最近7天：X篇
- 最近30天：X篇
- 最近90天：X篇

📏 文章长度分析
- 最长文章：[文章名]
- 最短文章：[文章名]

📊 详细报告已保存：[报告路径]
```
