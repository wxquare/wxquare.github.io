# Task 3 报告：归档原文并更新旧 URL alias

## 状态

已完成并提交。

## 变更

- 使用 Git rename 将两篇原文移至 `source/archive/system-design/`。
- 两篇归档原文增加 `published: false` 和指向新文章的归档说明。
- 两个旧路径改为固定重定向目标：
  - `/system-design/25-ecommerce-pricing-ddd/`
  - `/system-design/43-acc-ddd-notes/`
- alias 测试新增两条固定目标断言，并从 slug alias 集合移除旧条目。

## Commit

`f1cb54c7 fix: archive merged DDD posts and preserve aliases`

## 测试摘要

- `node --test test/legacy-post-alias.test.js`：通过，2/2。
- `npm run clean`：通过。
- `npm run build`：仍失败。Hexo 在渲染既有站内入口时发现归档文章仍被其他文章通过 `post_link system-design/25-ecommerce-pricing-ddd` 和 `post_link system-design/43-acc-ddd-notes` 引用；由于两篇文章均为 `published: false`，这些 slug 不再属于 posts 集合，渲染报 `Post not found`。按 Task 3 要求未修改其他站内入口。
- 因构建失败，未执行成功的 public HTML 目标检查。

## Concerns

现有站内入口仍引用两个已归档 slug。若要恢复整站构建，需要后续任务更新这些入口指向新 slug；本 Task 按要求未触碰它们。
