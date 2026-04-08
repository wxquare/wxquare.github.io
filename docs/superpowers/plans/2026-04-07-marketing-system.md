# 营销系统文章实施计划

## 总体策略

**执行方式**：分15个任务，每个任务完成后验证并提交git

**文件路径**：`source/_posts/system-design/28-ecommerce-marketing-system.md`

**设计文档**：`docs/superpowers/specs/2026-04-07-marketing-system-design.md`

---

## Task 1: 创建文章骨架（Front Matter + 目录）

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：新建文件

### 内容
```markdown
---
title: 电商系统设计：营销系统深度解析
date: 2026-04-07
categories:
  - system-design
tags:
  - system-design
  - ecommerce
  - marketing
  - coupon
  - points
  - promotion
  - high-concurrency
  - distributed-system
---

# 前言

营销系统是电商平台的增长引擎，通过优惠券、积分、活动等手段实现用户拉新、促活、留存和GMV提升。本文深入解析营销系统的架构设计、核心模块、高并发场景处理和工程实践，适合系统设计面试和电商后端工程师阅读。

<!-- more -->

# 目录

1. 系统概览
2. 营销工具体系
3. 营销计算引擎
4. 高并发场景设计
5. 营销与订单集成
6. 跨系统全链路集成
7. 数据一致性保障
8. 特殊营销场景
9. 工程实践
10. 总结与参考

---

（待补充具体内容）
```

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 创建营销系统文章骨架"
```

---

## Task 2: 第1章 - 系统概览

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：在文件末尾追加内容

### 内容要点
1. **1.1 营销系统的定位**
   - 营销系统在电商平台中的角色
   - 与其他系统的关系
   - 核心价值

2. **1.2 核心业务场景**
   - 用户拉新、促活、留存、GMV提升
   - B2C vs B2B2C对比表格

3. **1.3 核心挑战**
   - 高并发、复杂规则、数据一致性、防刷防薅、成本控制

4. **1.4 系统架构**
   - Mermaid系统架构图
   - 核心常量定义（Go代码）

5. **1.5 核心数据模型概览**
   - Mermaid ER图

6. **1.6 技术选型**
   - 技术选型对比表格

### Go代码块（2个）
- 核心常量定义
- （可选）系统架构说明

### Mermaid图表（3个）
- 系统架构总览图（graph LR）
- 核心数据模型ER图（erDiagram）
- （可选）模块交互图

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第1章系统概览"
```

---

## Task 3: 第2.1章 - 优惠券系统

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点
1. **2.1.1 优惠券类型与数据模型**
   - Go代码：Coupon、CouponUser、CouponLog结构体

2. **2.1.2 优惠券发放策略**
   - 公开领取、定向推送、裂变发券、订单赠送
   - Mermaid流程图：公开领券流程
   - Go代码：ReceiveCoupon核心逻辑

3. **2.1.3 优惠券核销流程**
   - Mermaid状态机图：优惠券状态流转
   - Go代码：UseCoupon核心逻辑

4. **2.1.4 优惠券回退**
   - Go代码：RollbackCoupon核心逻辑

### Go代码块（4个）
- 优惠券数据结构
- ReceiveCoupon
- UseCoupon
- RollbackCoupon

### Mermaid图表（2个）
- 公开领券流程（sequenceDiagram）
- 优惠券状态机（stateDiagram-v2）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第2.1章优惠券系统"
```

---

## Task 4: 第2.2-2.3章 - 积分系统和活动引擎

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点

**2.2 积分系统**

1. **2.2.1 积分账户模型**
   - Go代码：PointsAccount、PointsLog、PointsExpire结构体

2. **2.2.2 积分发放**
   - Go代码：EarnPoints

3. **2.2.3 积分扣减**
   - Go代码：SpendPoints

4. **2.2.4 积分退还**
   - Go代码：RefundPoints

5. **2.2.5 积分过期机制**
   - Go代码：ExpirePointsScanner、processExpiredPoints

**2.3 活动引擎**

1. **2.3.1 活动类型**
   - 活动类型对比表格

2. **2.3.2 活动数据模型**
   - Go代码：Activity、ActivityProduct、规则结构体

3. **2.3.3 活动状态机**
   - Mermaid状态机图

4. **2.3.4 圈品规则**
   - Go代码：IsProductEligible

### Go代码块（9个）
- 积分数据结构（1）
- EarnPoints（1）
- SpendPoints（1）
- RefundPoints（1）
- 积分过期（2）
- 活动数据结构（1）
- IsProductEligible（1）
- （可选补充1）

### Mermaid图表（1个）
- 活动状态机（stateDiagram-v2）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第2.2-2.3章积分系统和活动引擎"
```

---

## Task 5: 第3章 - 营销计算引擎

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点

1. **3.1 优惠计算流程**
   - Mermaid流程图：营销计算总览
   - Go代码：Calculate入口函数

2. **3.2 优惠叠加与互斥规则**
   - 规则说明
   - Go代码：RuleEngine（FindBestCombination、calculatePlan、calculateCouponDiscount、calculateActivityDiscount）

3. **3.3 优惠分摊到商品**
   - Go代码：allocateDiscountToItems

### Go代码块（6个）
- Calculate入口
- FindBestCombination
- calculatePlan
- calculateCouponDiscount
- calculateActivityDiscount
- allocateDiscountToItems

### Mermaid图表（1个）
- 营销计算流程图（graph TD）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第3章营销计算引擎"
```

---

## Task 6: 第4章 - 高并发场景设计

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点

**4.1 秒杀/抢券设计**

1. **4.1.1 秒杀系统架构**
   - Mermaid架构图

2. **4.1.2 流量削峰**
   - Go代码：VerifyCaptcha

3. **4.1.3 分布式锁**
   - Go代码：SecKillProduct

4. **4.1.4 库存预扣与异步确认**
   - Mermaid序列图

**4.2 防刷防薅**

1. **4.2.1 用户行为风控**
   - Go代码：CheckRateLimit（滑动窗口限流）

2. **4.2.2 营销预算控制**
   - Go代码：CheckBudget、DeductBudget

### Go代码块（5个）
- VerifyCaptcha
- SecKillProduct
- CheckRateLimit
- CheckBudget
- DeductBudget

### Mermaid图表（2个）
- 秒杀系统架构图（graph TB）
- 秒杀流程序列图（sequenceDiagram）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第4章高并发场景设计"
```

---

## Task 7: 第5章 - 营销与订单集成

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点

1. **5.1 下单时的营销扣减（Saga模式）**
   - Mermaid序列图：订单创建流程
   - Go代码：CreateOrder（Saga模式）

2. **5.2 取消订单时的营销回退**
   - Go代码：CancelOrder（Saga回滚）

### Go代码块（2个）
- CreateOrder（Saga模式）
- CancelOrder

### Mermaid图表（1个）
- 订单创建流程（sequenceDiagram）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第5章营销与订单集成"
```

---

## Task 8: 第6章 - 跨系统全链路集成

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点

1. **6.1 与商品系统集成（圈品规则）**
   - Go代码：GetProductMarketingInfo

2. **6.2 与计价中心集成（价格计算）**
   - Mermaid流程图

3. **6.3 与用户系统集成（用户画像）**
   - Go代码：SendTargetedCoupons

4. **6.4 与支付系统集成（补贴核算）**
   - Go代码：SettleMarketingCost

### Go代码块（3个）
- GetProductMarketingInfo
- SendTargetedCoupons
- SettleMarketingCost

### Mermaid图表（1个）
- 计价中心集成流程（sequenceDiagram）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第6章跨系统全链路集成"
```

---

## Task 9: 第7章 - 数据一致性保障

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点

1. **7.1 分布式事务（Saga模式）**
   - 简要说明（引用第5章）

2. **7.2 补偿任务与重试**
   - Go代码：CompensationWorker、processCompensationTasks、executeCompensation

3. **7.3 最终一致性方案**
   - Go代码：publishMarketingEvent、consumeMarketingEvents

4. **7.4 数据对账**
   - Go代码：ReconcileMarketingData

### Go代码块（5个）
- CompensationTask结构体
- CompensationWorker
- processCompensationTasks
- executeCompensation
- publishMarketingEvent
- consumeMarketingEvents
- ReconcileMarketingData

（实际为6-7个，根据拆分情况）

### Mermaid图表（0个）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第7章数据一致性保障"
```

---

## Task 10: 第8章 - 特殊营销场景

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点

1. **8.1 跨店铺满减**
   - Go代码：CalculateCrossShopFullReduction

2. **8.2 阶梯优惠**
   - Go代码：CalculateTieredDiscount

3. **8.3 组合优惠（买A送B）**
   - Go代码：ApplyBundleGiftActivity

4. **8.4 新人专享**
   - Go代码：IsNewUserEligible、ApplyNewUserDiscount

### Go代码块（6个）
- CalculateCrossShopFullReduction
- TieredDiscountRule结构体
- CalculateTieredDiscount
- BundleGiftRule结构体
- ApplyBundleGiftActivity
- IsNewUserEligible
- ApplyNewUserDiscount

（实际为7个）

### Mermaid图表（0个）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第8章特殊营销场景"
```

---

## Task 11: 第9章 - 工程实践

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点

1. **9.1 营销活动ID生成**
   - Go代码：SnowflakeIDGenerator

2. **9.2 监控告警**
   - 监控指标说明
   - Go代码：RecordMetrics

3. **9.3 性能优化**
   - 优化方案对比表格
   - Go代码：多级缓存（CouponCache.GetCoupon）

4. **9.4 容量规划与压测**
   - 压测指标说明

5. **9.5 故障处理与降级**
   - 故障处理对比表格
   - Go代码：熔断降级（CircuitBreaker）

### Go代码块（4个）
- SnowflakeIDGenerator
- RecordMetrics
- CouponCache.GetCoupon
- CircuitBreaker降级

### Mermaid图表（0个）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第9章工程实践"
```

---

## Task 12: 第10章 - 总结与参考

### 文件操作
- **文件**：`source/_posts/system-design/28-ecommerce-marketing-system.md`
- **操作**：追加内容

### 内容要点

1. **核心设计要点**
   - 6点总结

2. **面试高频问题**
   - 5个常见面试题及答案要点

3. **扩展阅读**
   - 推荐书籍和资料

4. **相关系列文章**
   - 链接到订单系统、商品系统文章

### Go代码块（0个）

### Mermaid图表（0个）

### 验证
```bash
npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "feat(marketing): 完成第10章总结与参考"
```

---

## Task 13: 内容完善和自检

### 操作
1. **通读全文**，检查：
   - Front Matter格式正确
   - 章节标题层级正确（一级标题用#，二级用##）
   - 代码块语言标注正确（```go）
   - Mermaid图表渲染正确
   - 中英文之间有空格
   - 专业术语统一（优惠券、积分、活动、营销系统）

2. **统计验证**
   - Go代码块数量：预期46个左右
   - Mermaid图表数量：预期16个左右
   - 总行数：预期2300行左右

3. **交叉引用检查**
   - 与订单系统文章的术语一致性
   - 与商品系统文章的术语一致性

### 验证命令
```bash
# 统计Go代码块
grep -c '^```go' source/_posts/system-design/28-ecommerce-marketing-system.md

# 统计Mermaid图表
grep -c '^```mermaid' source/_posts/system-design/28-ecommerce-marketing-system.md

# 统计行数
wc -l source/_posts/system-design/28-ecommerce-marketing-system.md

# 构建验证
npm run clean && npm run build
```

### Git提交
```bash
git add source/_posts/system-design/28-ecommerce-marketing-system.md
git commit -m "refactor(marketing): 内容完善和自检"
```

---

## Task 14: 构建验证

### 操作
```bash
# 清理缓存
npm run clean

# 重新构建
npm run build
```

### 预期结果
- 构建成功，无报错
- 无Hexo渲染错误
- 所有Mermaid图表正确渲染

### 如果失败
- 检查Markdown语法
- 检查Front Matter格式
- 检查代码块闭合

---

## Task 15: 最终检查和提交

### 操作
1. **检查git status**
   ```bash
   git status
   ```

2. **提交设计文档和实施计划**
   ```bash
   git add docs/superpowers/specs/2026-04-07-marketing-system-design.md
   git add docs/superpowers/plans/2026-04-07-marketing-system.md
   git commit -m "docs(marketing): 添加营销系统设计文档和实施计划"
   ```

3. **最终验证**
   ```bash
   npm run build
   npm run server
   ```
   访问 http://localhost:4000，查看文章渲染效果

4. **生成项目总结**
   总结营销系统文章的完成情况，包括：
   - 文章结构
   - 技术亮点
   - Git提交记录
   - 质量保证措施

---

## 执行检查清单

- [ ] Task 1: 创建文章骨架
- [ ] Task 2: 第1章系统概览
- [ ] Task 3: 第2.1章优惠券系统
- [ ] Task 4: 第2.2-2.3章积分系统和活动引擎
- [ ] Task 5: 第3章营销计算引擎
- [ ] Task 6: 第4章高并发场景设计
- [ ] Task 7: 第5章营销与订单集成
- [ ] Task 8: 第6章跨系统全链路集成
- [ ] Task 9: 第7章数据一致性保障
- [ ] Task 10: 第8章特殊营销场景
- [ ] Task 11: 第9章工程实践
- [ ] Task 12: 第10章总结与参考
- [ ] Task 13: 内容完善和自检
- [ ] Task 14: 构建验证
- [ ] Task 15: 最终检查和提交

---

## 预估时间

- Task 1: 5分钟
- Task 2: 15分钟
- Task 3: 15分钟
- Task 4: 20分钟
- Task 5: 15分钟
- Task 6: 15分钟
- Task 7: 10分钟
- Task 8: 10分钟
- Task 9: 15分钟
- Task 10: 15分钟
- Task 11: 15分钟
- Task 12: 5分钟
- Task 13: 10分钟
- Task 14: 5分钟
- Task 15: 10分钟

**总计**：约 180分钟（3小时）

---

## 质量保证

1. **代码质量**
   - 所有Go代码语义正确
   - 关键逻辑有中文注释

2. **图表质量**
   - 所有Mermaid图表渲染正确
   - 节点命名清晰

3. **内容质量**
   - 术语统一
   - 逻辑连贯
   - 与订单、商品系统文章一致

4. **构建验证**
   - 每个Task完成后都执行`npm run build`
   - 最终执行`npm run clean && npm run build`全量验证
