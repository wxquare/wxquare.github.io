# `source/images` 与 `source/photos` 合并设计

## 目标

将博客图片资源统一收敛到 `source/images/`，消除 `source/photos/` 与 `source/images/` 的职责重叠，并删除当前仓库中没有静态引用的图片资源。

## 范围与约束

- 只处理 `source/images/` 和 `source/photos/` 及引用这两个目录的文本文件。
- 保留当前仍被源码引用的资源文件名，不做语义化重命名。
- 不修改文章标题、分类、永久链接、图表目录、资料库、演示资料或其他已有未提交改动。
- `source/booklist/` 等项目规则明确保留的遗留目录不在本次范围内。
- 引用检查针对当前仓库 `source/` 中的 Markdown、Nunjucks、Stylus、JavaScript、JSON 和 YAML 文件；匹配资源路径和资源文件名。

## 现状盘点

| 目录 | 文件数 | 有当前源码引用 | 无当前源码引用 |
| --- | ---: | ---: | ---: |
| `source/images/` | 151 | 84 | 67 |
| `source/photos/` | 12 | 4 | 8 |
| 合计 | 163 | 88 | 75 |

两份 `kafka_architecture.png` 的 SHA-256 相同，但两份均无当前源码引用，因此一并删除，不产生合并冲突。

## 变更设计

### 1. 迁移仍在使用的 `photos` 资源

将以下文件移动到 `source/images/`，保留原文件名：

- `source/photos/gc_setGCPercent.jpg`
- `source/photos/pageheap.gif`
- `source/photos/threadheap.gif`
- `source/photos/threelayer.jpg`

修改 `source/_posts/fundamentals/6-golang-practice.md` 中的 4 条引用，将 `/photos/` 改为 `/images/`。这样文章继续使用统一的站内资源 URL。

### 2. 删除无当前源码引用的资源

删除以下 `source/images/` 文件：

- `source/images/0vBc0hN.png`
- `source/images/1ae6b34b-78d9-4e27-b7d8-cb3c89f13d7b.png`
- `source/images/4edXG0T.png`
- `source/images/4fff8fa9805f408dcf9ae012359e6a7a.png`
- `source/images/4j99mhe.png`
- `source/images/54GYsSx.png`
- `source/images/E-commerce-category-brand-product.webp`
- `source/images/E-commerce-product-management-ER.jpg`
- `source/images/E-commerce-product-management.webp`
- `source/images/MzExP06.png`
- `source/images/ONjORqk.png`
- `source/images/OfVllex.png`
- `source/images/Q6z24La.png`
- `source/images/TcUo2fw.png`
- `source/images/V5q57vU.png`
- `source/images/Xkm5CXz.png`
- `source/images/arch-blu.png`
- `source/images/arch-comm-02.png`
- `source/images/arch-qua-01.jpeg`
- `source/images/avatar.jpeg`
- `source/images/b4YtAEN.png`
- `source/images/bWxPtQA.png`
- `source/images/bgLMI2u.png`
- `source/images/big-promotion-system-stability.png`
- `source/images/cdCv5g7.png`
- `source/images/codis.png`
- `source/images/e-commerce-mindmap.png`
- `source/images/e-commerce-mindmap.puml`
- `source/images/e-commerce-system.png`
- `source/images/e-commerce-第 19 页.drawio.png`
- `source/images/fNcl65g.png`
- `source/images/h9TAuGI.jpg`
- `source/images/how-to-evaluate-tech-design.png`
- `source/images/item-example.png`
- `source/images/item-sku-er.png`
- `source/images/item-sku-example.png`
- `source/images/item-sku.png`
- `source/images/jj3A5N8.png`
- `source/images/jrUBAF7.png`
- `source/images/k8s_ingress.png`
- `source/images/k8s_ingress_background.png`
- `source/images/k8s_services.png`
- `source/images/k8s_services_background.png`
- `source/images/k8s_services_name_space_load_balacing.png`
- `source/images/kafka_architecture.png`
- `source/images/kxtjqgE.png`
- `source/images/linux_namespace_lab.png`
- `source/images/lmstfy-internal.png`
- `source/images/load_balancer_architecture.jpeg`
- `source/images/lvs-ngnix-load-balancer.jpg`
- `source/images/n16iOGk.png`
- `source/images/order_state_machine.png`
- `source/images/quantization_result.jpg`
- `source/images/redis-data-type.jpg`
- `source/images/return&refund.png`
- `source/images/rgSrvjG.png`
- `source/images/system-monitor.png`
- `source/images/tech-principles.png`
- `source/images/tech-principles.webp`
- `source/images/tensorflow-model-quantization.jpg`
- `source/images/tf_model_pruning1.png`
- `source/images/wXGqG5f.png`
- `source/images/xxljob-architecture.png`
- `source/images/yB5SYwm.png`
- `source/images/zdCAkB3.png`
- `source/images/截屏2024-06-03 22.10.45.png`
- `source/images/订单状态一致性.png`

删除以下 `source/photos/` 文件：

- `source/photos/golang哈希一致性算法实践 _ wxquare's Blogs.pdf`
- `source/photos/goroutine-scheduler-model.png`
- `source/photos/hash_consistent.jpg`
- `source/photos/hexo_deploy.jpg`
- `source/photos/kafka_architecture.png`
- `source/photos/perf_kcf2.0.jpg`
- `source/photos/vscode_go_ext.jpg`
- `source/photos/截屏2023-02-17 下午4.07.23.png`

迁移后 `source/photos/` 为空并删除该目录。

## 兼容性与风险控制

- 所有被保留并迁移的文件保持原文件名，新的规范 URL 为 `/images/<filename>`。
- 现有文章中的 4 条 `/photos/` 引用会同步更新，避免当前站点生成缺图。
- 本次不保证历史外部 `raw/hexo/source/photos/...` URL 继续可用；仓库当前源码中的引用会全部切换到 `/images/`。
- 删除动作只针对当前静态扫描确认无引用的文件，Git 变更可审查、可恢复。
- 不删除 `source/images/` 目录本身，也不重组 `source/diagrams/`。

## 验证方案

1. 检查 `source/photos/` 不存在。
2. 检查仓库源码中不再存在 `/photos/` 或 `source/photos/` 引用。
3. 检查所有仍存在的 `/images/<filename>` 引用都能解析到 `source/images/<filename>`。
4. 运行 `npm test`。
5. 运行 `npm run clean && npm run build`，确认 Hexo 构建成功。
6. 运行 `npm run check:links`，确认生成站点的资源链接没有新增断链。
7. 查看最终 `git status --short`，确认仅包含本次资源合并相关变更以及执行前已经存在的用户改动。
