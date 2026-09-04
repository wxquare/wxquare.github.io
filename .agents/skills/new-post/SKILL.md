---
disable-model-invocation: true
---

# /new-post - 创建新博客文章

本技能只描述创建流程。开始前读取 [`AGENTS.md`](../../../AGENTS.md) 第 3、4、7.1、10、11 节；分类目录和 Front Matter 映射读取 `.agents/config/post-categories.json`。

## 执行步骤

1. **收集基本信息**
   - 标题
   - 主分类 slug
   - 标签
   - 是否需要系列编号或其他文件名前缀

2. **解析文章路径**
   - 在注册表中按主分类 slug 查找 `directory`、`label`、`frontMatterLabels` 和 `subdirectories`。
   - 如果用户要求子目录，确认它存在于该分类的 `subdirectories`，再拼接目标路径。
   - 如果 slug 或子目录不在注册表中，先请用户选择有效值，不要自行创建新的分类。

3. **生成文件名**
   - 时效性文章：`YYYY-MM-DD-标题.md`
   - 系列文章：`数字-标题.md`
   - 技术笔记：描述性文件名
   - 创建前检查目标路径是否已有同名文件。

4. **创建文章骨架**

```yaml
---
title: [文章标题]
date: [YYYY-MM-DD]
categories:
  - [注册表中的 Front Matter 展示名]
tags:
  - [标签1]
  - [标签2]
---

## 引言

[文章背景和目的]

## 核心内容

### 小节

[内容]

## 总结

[总结要点]

## 参考资料

- [参考链接]
```

5. **按规范补全内容**
   - 使用 `AGENTS.md` 指定的 Front Matter、命名、标签、图片、代码块和内容结构规则。
   - 技术教程可参考 `.agents/templates/tech-tutorial.md`；其他主题选择对应模板。

6. **完成提示**
   - 输出新文件的绝对路径。
   - 提示运行 `/review-post`、`npm run clean` 和 `npm run build`。
   - 用户要求时再创建图表或代码示例目录，并遵守 `AGENTS.md` 的源码边界。
