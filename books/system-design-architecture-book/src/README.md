# 系统设计与架构实战：原理、工程与电商案例

这是一本把《程序员系统设计与面试指南》和《电商系统架构设计与实现》合并后的系统设计实战书。新书以系统设计方法论为主线，尽量保留电商书的完整内容，把商品、库存、营销、计价、搜索、购物车、订单、支付、供应商同步和 B2B2C 平台作为一条连续的业务案例链路。

## 定位

本书不是单纯的面试题库，也不是只讲电商业务的行业手册，而是一本面向中高级工程师的架构实践指南：

- 用方法论回答系统怎么拆、边界怎么定、系统之间怎么协作。
- 用基础设施章节解释 MySQL、Redis、Kafka、Elasticsearch、Kubernetes 等组件的能力边界。
- 用可靠性章节补齐对账、补偿、DLQ、容量规划、限流、熔断和降级。
- 用电商实战篇把前面的原则放进完整业务链路中验证。
- 用面试篇把工程判断转化为可表达、可答辩的结构化语言。

## 阅读路线

1. **系统设计主线**：第 1-18 章，先建立方法论、基础设施和可靠性框架。
2. **电商实战主线**：第 19-31 章，按商品、库存、营销、计价、搜索、交易、支付和综合平台顺序阅读。
3. **面试准备主线**：第 32-37 章，结合前文内容整理系统设计表达、追问和白板答辩。
4. **基础补齐主线**：第 38-49 章，补足操作系统、网络、语言实践和算法题型。

## 合并原则

本书采用“高保真并入”的方式生成：

- 原 `system-design-book` 作为基础骨架保留。
- 原 `ecommerce-book` 的方法论章节上提到第一部分和第三部分。
- 原 `ecommerce-book` 的核心业务章节完整并入第四部分。
- 原 `ecommerce-book` 的供应商同步、商品供给治理、全局 ID 和面试题附录升级为正式章节。
- 新增章节只承担桥接、总结和补齐职责，不替代原有长文内容。

## 本地构建

需安装 [mdBook](https://github.com/rust-lang/mdBook)。本书复用 `books/scripts/mermaid-preprocessor.py` 处理 Mermaid 图表。

```bash
cd books/system-design-architecture-book
mdbook build
mdbook serve
```

也可以在仓库根目录生成到 Hexo 本地预览目录并启动服务：

```bash
npm run server:system-design-architecture-book
```

启动后访问：

```text
http://localhost:3000/system-design-architecture-book/
```

## 源稿关系

本书生成后，原有两本书仍保留在仓库中：

- `books/system-design-book/`
- `books/ecommerce-book/`

后续可以逐步将新书作为主线版本维护，原书作为历史版本或专题版本保留。
