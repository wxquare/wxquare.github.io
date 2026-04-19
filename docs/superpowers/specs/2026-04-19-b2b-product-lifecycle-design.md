# B端商品生命周期完整链路设计

**设计日期**：2026-04-19  
**目标章节**：`ecommerce-book/src/part3/chapter16.md` 16.6.6  
**状态**：待实施

---

## 1. 设计目标

为Chapter 16添加B端商品生命周期完整链路，与现有的16.6.7 C端交易流形成对照，展示供应商/运营/系统三方协作下商品从录入到归档的完整过程。

### 业务价值

- 补充B端视角，与C端链路形成完整闭环
- 展示供给侧系统设计（对标消费侧）
- 体现B2B2C平台的供应商管理能力
- 提供运营人员的业务流程参考

### 技术目标

- 7个阶段，每个阶段150-200行
- 对标C端链路的内容深度
- 包含业务场景、架构图、Go代码、监控指标
- 总计新增约1150行内容

---

## 2. 章节结构调整

### 调整前

```
16.6.6 商品供给与运营系统
  ├─ 商品上架系统（从无到有）
  ├─ 供应商同步系统（Upsert场景）
  └─ 运营编辑系统（日常维护）

16.6.7 C端交易流完整链路
  └─ 5个阶段
```

### 调整后

```
16.6.6 B端商品生命周期完整链路 ✨新增
  ├─ 阶段1：商品录入
  ├─ 阶段2：审核发布
  ├─ 阶段3：供应商同步
  ├─ 阶段4：库存管理 ✨全新
  ├─ 阶段5：日常维护
  ├─ 阶段6：促销配置 ✨全新
  └─ 阶段7：下架归档

16.6.7 C端交易流完整链路（编号不变）
  └─ 5个阶段
```

**原16.6.6内容去向：**
- 商品上架系统 → 融入阶段1（商品录入）
- 供应商同步系统 → 融入阶段3（供应商同步）
- 运营编辑系统 → 融入阶段5（日常维护）

---

## 3. 七个阶段详细设计

### 阶段1：商品录入

**定位：** 商品从无到有的创建过程

**业务场景：**
1. **手动创建**：运营人员在OMS后台单个录入商品
2. **批量导入**：商家通过Portal上传Excel（100-1000个SKU）
3. **API推送**：供应商首次推送新品

**参与角色：**
- 运营人员：手动录入、数据校验
- 商家：批量导入、信息维护
- 供应商：API推送
- 系统：幂等性控制、状态机管理

**系统架构：**
```
[运营后台/商家Portal/供应商API]
    ↓
ListingService（上架服务）
    ├─ CreateListingTask（幂等性）
    ├─ BatchImport（批量导入）
    ├─ 数据校验（必填项、格式、类目）
    ├─ 状态机（DRAFT → PENDING）
    └─ 事件发布（ListingTaskCreated）
    ↓
存储到 listing_tasks 表
    ↓
通知审核队列
```

**核心数据结构：**
```go
type ListingTask struct {
    TaskCode    string        // 幂等性标识
    ItemInfo    ItemInfo      // 商品信息
    SupplierID  int64         // 供应商ID
    Source      string        // MANUAL/BATCH/API
    Status      ListingStatus // DRAFT/PENDING/APPROVED/REJECTED/PUBLISHED
    ReviewerID  int64         // 审核人
    RejectReason string       // 驳回原因
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type ItemInfo struct {
    Title       string
    CategoryID  int64
    BrandID     int64
    Images      []string
    Description string
    BasePrice   int64
    InitStock   int
}
```

**核心代码实现：**

1. **单品录入**（复用现有CreateListingTask）
2. **批量导入**（新增）：
```go
func (s *ListingService) BatchImport(ctx context.Context, file io.Reader) (*BatchImportTask, error) {
    // 解析Excel → 创建批量任务 → 异步处理
    items, err := s.parseImportFile(file)
    if err != nil {
        return nil, fmt.Errorf("文件解析失败: %w", err)
    }
    
    batchTask := &BatchImportTask{
        TaskID:     generateTaskID(),
        TotalCount: len(items),
        Status:     "PENDING",
    }
    s.taskRepo.Save(ctx, batchTask)
    
    s.taskQueue.Publish(ctx, &BatchImportEvent{
        TaskID: batchTask.TaskID,
        Items:  items,
    })
    
    return batchTask, nil
}
```

**监控指标：**
- 商品录入总数（按渠道：手动/批量/API）
- 录入成功率（> 95%）
- 平均录入耗时（单个 < 500ms，批量 < 5s/100个）
- 数据校验失败率（< 5%）
- 批量导入失败原因分布（TOP 5）

**预计代码量：** ~180行

---

### 阶段2：审核发布

**定位：** 运营审核商品合规性，审核通过后发布到商品中心

**业务场景：**
1. **人工审核**：运营人员审核新商品（合规性、完整性、准确性）
2. **自动审核**：高信用供应商商品自动通过（信用等级 >= 3）
3. **审核驳回**：不合规商品驳回，运营修改后重新提交

**参与角色：**
- 运营审核员：人工审核、驳回理由
- 系统：规则引擎自动审核
- 商家/供应商：修改驳回商品

**系统架构：**
```
审核队列（待审核商品）
    ↓
审核引擎（ReviewEngine）
    ├─ 规则引擎（合规性、完整性、准确性）
    ├─ 人工审核（运营后台）
    └─ 自动审核（高信用供应商）
    ↓
审核通过 → 发布流程
    ├─ ProductCenter.CreateProduct
    ├─ InventoryService.InitStock
    ├─ PricingService.InitPrice
    └─ SearchService.IndexProduct
```

**审核维度：**

| 审核维度 | 检查项 | 风险等级 | 处理方式 |
|---------|--------|---------|---------|
| **合规性** | 违禁词、敏感内容 | 高 | 自动拦截 |
| **完整性** | 必填字段、图片数量 | 中 | 人工审核 |
| **准确性** | 价格合理性、类目匹配 | 中 | 人工审核 |
| **一致性** | SPU/SKU关系 | 低 | 自动检查 |

**核心代码实现：**

1. **审核引擎**（新增）：
```go
type ReviewEngine struct {
    rules []ReviewRule
}

type ReviewRule interface {
    Check(ctx context.Context, task *ListingTask) *ReviewResult
}

func (e *ReviewEngine) AutoReview(ctx context.Context, task *ListingTask) (*ReviewResult, error) {
    result := &ReviewResult{Pass: true}
    
    for _, rule := range e.rules {
        ruleResult := rule.Check(ctx, task)
        if !ruleResult.Pass {
            result.Pass = false
            result.Reasons = append(result.Reasons, ruleResult.Reason)
        }
    }
    
    return result, nil
}

// 合规性检查规则
type ComplianceRule struct{}

func (r *ComplianceRule) Check(ctx context.Context, task *ListingTask) *ReviewResult {
    if containsSensitiveWords(task.ItemInfo.Title) {
        return &ReviewResult{Pass: false, Reason: "标题含违禁词"}
    }
    return &ReviewResult{Pass: true}
}
```

2. **审核通过**（复用现有Approve逻辑）

**监控指标：**
- 待审核商品数（实时）
- 审核通过率（目标 > 85%）
- 平均审核时长（人工 < 2小时，自动 < 1秒）
- 审核驳回原因分布（TOP 10）
- 自动审核占比（> 60%）

**预计代码量：** ~150行

---

### 阶段3：供应商同步

**定位：** 供应商数据变更后的增量同步

**业务场景：**
1. **定时同步**：供应商每小时/每天推送数据（全量/增量）
2. **实时推送**：价格/库存变更实时推送
3. **Upsert语义**：存在则更新，不存在则创建

**参与角色：**
- 供应商：推送数据变更
- 系统：接收、转换、同步
- 运营：审核高风险变更

**系统架构：**
```
供应商推送（HTTP Webhook/MQ）
    ↓
Supplier Gateway（防腐层）
    ├─ 协议适配
    ├─ 数据转换
    └─ 熔断保护
    ↓
SyncService（同步服务）
    ├─ Upsert判断（查询mapping表）
    ├─ 差异检测（对比现有数据）
    ├─ 差异化审核（高风险 → 人工）
    └─ 数据更新/创建
    ↓
更新商品中心/库存/价格
```

**差异化审核策略：**

| 变更类型 | 变更范围 | 审核策略 | 理由 |
|---------|---------|---------|------|
| **价格变更** | < 10% | 自动通过 | 正常波动 |
| **价格变更** | 10-50% | 需要审核 | 防止错误 |
| **价格变更** | > 50% | 必须审核+告警 | 高风险 |
| **库存变更** | 任意 | 自动通过 | 实时性要求高 |
| **标题变更** | 轻微修改 | 自动通过 | 低风险 |
| **类目变更** | 任意 | 必须审核 | 影响搜索 |
| **图片变更** | 任意 | 自动通过 | 低风险 |

**核心代码实现：**

1. **Upsert处理**（复用现有UpsertProduct）
2. **差异检测**（新增）：
```go
type ProductDiff struct {
    PriceChange    *PriceChange
    CategoryChange *CategoryChange
    StockChange    *StockChange
    ImageChange    *ImageChange
}

func (s *SyncService) compareDiff(existing *Product, syncData *SyncData) *ProductDiff {
    diff := &ProductDiff{}
    
    if existing.Price != syncData.Price {
        changePercent := math.Abs(float64(syncData.Price-existing.Price)) / float64(existing.Price) * 100
        diff.PriceChange = &PriceChange{
            OldPrice:      existing.Price,
            NewPrice:      syncData.Price,
            ChangePercent: changePercent,
        }
    }
    
    if existing.CategoryID != syncData.CategoryID {
        diff.CategoryChange = &CategoryChange{
            OldCategory: existing.CategoryID,
            NewCategory: syncData.CategoryID,
        }
    }
    
    return diff
}

func (d *ProductDiff) HasHighRiskChange() bool {
    if d.PriceChange != nil && d.PriceChange.ChangePercent > 50 {
        return true
    }
    if d.CategoryChange != nil {
        return true
    }
    return false
}
```

**监控指标：**
- 同步任务总数（按供应商、按类型：全量/增量）
- 同步成功率（> 98%）
- 数据一致性（平台 vs 供应商，> 99.5%）
- 高风险变更拦截率
- 同步延迟（P99 < 5秒）

**预计代码量：** ~160行

---

### 阶段4：库存管理（全新内容）

**定位：** 供应商库存同步、水位监控、补货流程、差异对账

**业务场景：**
1. **实时库存同步**：机票、酒店实时查询供应商库存
2. **定时库存同步**：充值、券码每小时同步一次
3. **库存水位预警**：低于安全库存告警
4. **缺货处理**：自动下架 + 通知补货
5. **库存对账**：每日对账平台库存 vs 供应商库存

**参与角色：**
- 供应商：推送库存变更
- 系统：同步、监控、预警
- 运营：处理缺货、手动补货

**系统架构：**
```
供应商库存数据源
    ├─ 实时查询（机票、酒店）
    ├─ 定时推送（充值、券码）
    └─ Webhook推送（库存变更）
    ↓
InventoryGateway（库存网关）
    ├─ 协议适配
    ├─ 数据归一化
    └─ 熔断保护
    ↓
InventoryManagementService
    ├─ 库存同步（写入Redis + MySQL）
    ├─ 水位监控（检查安全库存）
    ├─ 预警通知（缺货告警）
    ├─ 补货流程（自动/手动）
    └─ 差异对账（每日任务）
    ↓
库存数据 + 运营通知 + 对账报告
```

**核心设计：**

**1. 库存同步策略**（按品类差异化）
```go
type StockSyncStrategy interface {
    Sync(ctx context.Context, skuID int64) (*SyncResult, error)
}

// 实时查询策略（机票、酒店）
type RealtimeStockSyncStrategy struct {
    supplierClient rpc.SupplierClient
}

func (s *RealtimeStockSyncStrategy) Sync(ctx context.Context, skuID int64) (*SyncResult, error) {
    // 不主动同步，只在用户查询时实时调用
    return &SyncResult{
        SyncType: "REALTIME",
        Message:  "实时库存，无需同步",
    }, nil
}

// 定时推送策略（充值、券码）
type ScheduledStockSyncStrategy struct {
    supplierClient rpc.SupplierClient
    inventoryRepo  *InventoryRepo
}

func (s *ScheduledStockSyncStrategy) Sync(ctx context.Context, skuID int64) (*SyncResult, error) {
    // Step 1: 查询供应商库存
    supplierStock, err := s.supplierClient.QueryStock(ctx, skuID)
    if err != nil {
        return nil, fmt.Errorf("查询供应商库存失败: %w", err)
    }
    
    // Step 2: 获取平台库存
    platformStock, _ := s.inventoryRepo.GetStock(ctx, skuID)
    
    // Step 3: 对比差异
    if supplierStock.Quantity != platformStock.Quantity {
        // Step 4: 更新平台库存（Redis + MySQL双写）
        s.inventoryRepo.UpdateStock(ctx, skuID, supplierStock.Quantity)
        
        // Step 5: 记录同步日志
        s.inventoryRepo.LogSync(ctx, &SyncLog{
            SkuID:       skuID,
            OldQuantity: platformStock.Quantity,
            NewQuantity: supplierStock.Quantity,
            Diff:        supplierStock.Quantity - platformStock.Quantity,
            SyncTime:    time.Now(),
        })
        
        return &SyncResult{
            SyncType: "SCHEDULED",
            Updated:  true,
            Diff:     supplierStock.Quantity - platformStock.Quantity,
        }, nil
    }
    
    return &SyncResult{SyncType: "SCHEDULED", Updated: false}, nil
}
```

**2. 库存水位监控**
```go
type StockWatermarkMonitor struct {
    inventoryRepo *InventoryRepo
    alertService  *AlertService
    productClient rpc.ProductClient
}

func (m *StockWatermarkMonitor) CheckWatermark(ctx context.Context, skuID int64) error {
    // Step 1: 获取库存和配置
    stock, _ := m.inventoryRepo.GetStock(ctx, skuID)
    config, _ := m.inventoryRepo.GetConfig(ctx, skuID)
    
    // Step 2: 安全库存检查
    if stock.Quantity <= config.SafeStock && stock.Quantity > 0 {
        // 发送预警
        m.alertService.SendAlert(ctx, &StockAlert{
            SkuID:      skuID,
            CurrentQty: stock.Quantity,
            SafeStock:  config.SafeStock,
            AlertLevel: "WARNING",
            Message:    fmt.Sprintf("库存低于安全线：当前%d，安全%d", stock.Quantity, config.SafeStock),
        })
        
        // 触发补货流程（如果配置了自动补货）
        if config.AutoReplenish {
            m.triggerReplenishment(ctx, skuID, config.ReplenishQty)
        }
    }
    
    // Step 3: 缺货检查
    if stock.Quantity == 0 {
        // 自动下架
        m.productClient.OffShelf(ctx, &OffShelfRequest{
            SkuID:  skuID,
            Reason: "库存为0，自动下架",
            Type:   "TEMPORARY",
        })
        
        // 紧急告警
        m.alertService.SendAlert(ctx, &StockAlert{
            SkuID:      skuID,
            AlertLevel: "CRITICAL",
            Message:    "商品缺货已自动下架",
        })
    }
    
    return nil
}

// 触发补货流程
func (m *StockWatermarkMonitor) triggerReplenishment(ctx context.Context, skuID int64, qty int) error {
    replenishTask := &ReplenishmentTask{
        TaskID:     generateTaskID(),
        SkuID:      skuID,
        Quantity:   qty,
        Status:     "PENDING",
        CreateTime: time.Now(),
    }
    
    m.inventoryRepo.SaveReplenishTask(ctx, replenishTask)
    
    // 通知供应商（如果是供应商商品）
    m.notifySupplier(ctx, skuID, qty)
    
    return nil
}
```

**3. 库存对账**（每日任务）
```go
type StockReconciliationJob struct {
    inventoryRepo  *InventoryRepo
    supplierClient rpc.SupplierClient
    alertService   *AlertService
}

func (j *StockReconciliationJob) Run(ctx context.Context, date time.Time) error {
    // Step 1: 查询所有需要对账的商品
    skuIDs, _ := j.inventoryRepo.GetAllActiveSkus(ctx)
    
    var totalChecked, totalMismatch int
    var mismatchDetails []*StockMismatch
    
    for _, skuID := range skuIDs {
        // Step 2: 查询供应商库存
        supplierStock, err := j.supplierClient.QueryStock(ctx, skuID)
        if err != nil {
            log.Warnf("查询供应商库存失败: sku=%d, err=%v", skuID, err)
            continue
        }
        
        // Step 3: 查询平台库存
        platformStock, _ := j.inventoryRepo.GetStock(ctx, skuID)
        
        totalChecked++
        
        // Step 4: 对比差异
        if supplierStock.Quantity != platformStock.Quantity {
            totalMismatch++
            
            mismatch := &StockMismatch{
                SkuID:         skuID,
                SupplierQty:   supplierStock.Quantity,
                PlatformQty:   platformStock.Quantity,
                Diff:          supplierStock.Quantity - platformStock.Quantity,
                ReconcileDate: date,
            }
            mismatchDetails = append(mismatchDetails, mismatch)
            
            // 记录差异
            j.inventoryRepo.LogMismatch(ctx, mismatch)
            
            // 如果差异 > 阈值，告警
            if math.Abs(float64(mismatch.Diff)) > 10 {
                j.alertService.SendAlert(ctx, &StockMismatchAlert{
                    SkuID: skuID,
                    Diff:  mismatch.Diff,
                })
            }
        }
    }
    
    // Step 5: 生成对账报告
    report := &ReconciliationReport{
        Date:            date,
        TotalChecked:    totalChecked,
        TotalMismatch:   totalMismatch,
        MismatchRate:    float64(totalMismatch) / float64(totalChecked) * 100,
        MismatchDetails: mismatchDetails,
    }
    j.inventoryRepo.SaveReport(ctx, report)
    
    log.Infof("库存对账完成: 检查%d个, 差异%d个, 差异率%.2f%%", 
        totalChecked, totalMismatch, report.MismatchRate)
    
    return nil
}
```

**监控指标：**
- 库存同步频率（实时/小时/天）
- 库存准确率（平台 vs 供应商，> 99.5%）
- 缺货商品数（实时）
- 预警触发次数（安全库存预警、缺货告警）
- 对账差异率（< 0.5%）
- 补货任务完成率（> 90%）

**预计代码量：** ~200行

---

### 阶段5：日常维护

**定位：** 已上线商品的日常编辑和批量操作

**业务场景：**
1. **单品编辑**：修改标题、描述、图片（实时生效）
2. **批量编辑**：批量调价、批量修改属性（异步任务）
3. **批量导入更新**：Excel批量操作

**参与角色：**
- 运营人员：单品编辑
- 商家：批量编辑、Excel导入
- 系统：异步任务执行、进度追踪

**系统架构：**
```
运营后台/商家Portal
    ↓
EditService（编辑服务）
    ├─ 单品编辑（同步）
    └─ 批量编辑（异步任务）
    ↓
BatchEditWorker（异步Worker）
    ├─ 逐个处理
    ├─ 错误重试
    └─ 进度更新
    ↓
更新商品中心 + 搜索索引 + 缓存失效
```

**核心代码实现：**（复用现有BatchEdit，补充监控）
```go
func (s *EditService) BatchEdit(ctx context.Context, req *BatchEditRequest) (*EditTask, error) {
    startTime := time.Now()
    
    task := &EditTask{
        TaskID:     generateTaskID(),
        ItemIDs:    req.ItemIDs,
        EditType:   "BATCH",
        Changes:    req.Changes,
        Status:     "PENDING",
        TotalCount: len(req.ItemIDs),
    }
    s.taskRepo.Save(ctx, task)
    
    s.taskQueue.Publish(ctx, &BatchEditTaskEvent{
        TaskID: task.TaskID,
    })
    
    // 记录指标
    metrics.BatchEditTaskTotal.WithLabelValues("created").Inc()
    metrics.BatchEditTaskSize.Observe(float64(len(req.ItemIDs)))
    metrics.BatchEditTaskDuration.WithLabelValues("create").Observe(time.Since(startTime).Seconds())
    
    return task, nil
}
```

**监控指标：**
- 编辑操作总数（单品/批量）
- 编辑成功率（> 99%）
- 批量任务平均完成时间（100个 < 30秒）
- 编辑失败原因分布（TOP 5）
- 进度追踪查询QPS

**预计代码量：** ~140行

---

### 阶段6：促销配置（全新内容）

**定位：** 关联促销活动、设置促销价格、活动排期

**业务场景：**
1. **关联促销活动**：选择商品参与满减/折扣/买赠活动
2. **设置促销价格**：限时秒杀价、预售价
3. **活动排期**：设置开始/结束时间
4. **促销规则校验**：价格合理性、库存充足性

**参与角色：**
- 运营人员：配置促销、设置价格
- 系统：规则校验、定时生效/失效

**系统架构：**
```
运营后台促销配置
    ↓
PromotionConfigService
    ├─ 关联促销活动
    ├─ 价格合理性校验（不高于原价、不低于成本）
    ├─ 库存充足性检查（防止超卖）
    └─ 时间调度（生效/失效任务）
    ↓
写入promotion_items表
    ↓
定时任务Scheduler
    ├─ 活动生效（更新促销标签、刷新缓存）
    ├─ 活动失效（清理促销标签）
    └─ 刷新搜索索引
```

**核心代码实现：**
```go
type PromotionConfigService struct {
    promotionRepo   *PromotionRepo
    productClient   rpc.ProductClient
    inventoryClient rpc.InventoryClient
}

// 配置商品促销
func (s *PromotionConfigService) ConfigPromotion(ctx context.Context, req *ConfigPromotionRequest) error {
    // Step 1: 校验促销活动有效性
    promotion, err := s.promotionRepo.GetByID(ctx, req.PromotionID)
    if err != nil || promotion.Status != "ACTIVE" {
        return fmt.Errorf("促销活动无效")
    }
    
    // Step 2: 批量获取商品信息
    products, _ := s.productClient.BatchGet(ctx, req.SkuIDs)
    
    // Step 3: 价格合理性校验
    for _, skuID := range req.SkuIDs {
        product := products[skuID]
        
        // 促销价不能高于原价
        if req.PromoPrice >= product.BasePrice {
            return fmt.Errorf("促销价不能高于原价: sku=%d", skuID)
        }
        
        // 折扣不能低于成本价
        if req.PromoPrice < product.CostPrice {
            return fmt.Errorf("促销价低于成本价: sku=%d", skuID)
        }
    }
    
    // Step 4: 库存充足性检查
    stocks, _ := s.inventoryClient.BatchCheck(ctx, req.SkuIDs)
    for _, skuID := range req.SkuIDs {
        stock := stocks[skuID]
        if stock.Quantity < req.MinStock {
            return fmt.Errorf("库存不足，无法参与活动: sku=%d, 需要%d, 实际%d", 
                skuID, req.MinStock, stock.Quantity)
        }
    }
    
    // Step 5: 写入促销配置
    for _, skuID := range req.SkuIDs {
        promotionItem := &PromotionItem{
            PromotionID: req.PromotionID,
            SkuID:       skuID,
            PromoPrice:  req.PromoPrice,
            StartTime:   promotion.StartTime,
            EndTime:     promotion.EndTime,
            Status:      "SCHEDULED",
            CreatedAt:   time.Now(),
        }
        s.promotionRepo.SaveItem(ctx, promotionItem)
    }
    
    // Step 6: 调度生效任务
    s.scheduleActivation(ctx, promotion.StartTime, req.PromotionID, req.SkuIDs)
    
    return nil
}

// 促销生效任务（定时任务，每分钟扫描）
type PromotionActivationJob struct {
    promotionRepo *PromotionRepo
    cacheService  *CacheService
    searchClient  rpc.SearchClient
}

func (j *PromotionActivationJob) Run(ctx context.Context) {
    now := time.Now()
    
    // Step 1: 查询待生效的促销（start_time <= now AND status = SCHEDULED）
    items, _ := j.promotionRepo.FindScheduledItems(ctx, now)
    
    for _, item := range items {
        // Step 2: 更新状态：SCHEDULED → ACTIVE
        j.promotionRepo.UpdateStatus(ctx, item.ID, "SCHEDULED", "ACTIVE")
        
        // Step 3: 刷新商品缓存（让促销价生效）
        j.cacheService.InvalidateProduct(ctx, item.SkuID)
        
        // Step 4: 更新搜索索引（展示促销标签）
        j.searchClient.UpdatePromotionTag(ctx, item.SkuID, true)
        
        log.Infof("促销已生效: promotion=%d, sku=%d", item.PromotionID, item.SkuID)
    }
    
    // Step 5: 查询已失效的促销（end_time <= now AND status = ACTIVE）
    expiredItems, _ := j.promotionRepo.FindExpiredItems(ctx, now)
    
    for _, item := range expiredItems {
        // 更新状态：ACTIVE → EXPIRED
        j.promotionRepo.UpdateStatus(ctx, item.ID, "ACTIVE", "EXPIRED")
        j.cacheService.InvalidateProduct(ctx, item.SkuID)
        j.searchClient.UpdatePromotionTag(ctx, item.SkuID, false)
        
        log.Infof("促销已失效: promotion=%d, sku=%d", item.PromotionID, item.SkuID)
    }
}
```

**监控指标：**
- 促销配置商品数（按活动类型）
- 价格校验失败率（< 2%）
- 库存校验失败率（< 1%）
- 促销生效准时率（> 99.9%）
- 促销失效准时率（> 99.9%）
- 促销商品转化率（对比非促销）

**预计代码量：** ~180行

---

### 阶段7：下架归档

**定位：** 商品停售和历史数据归档

**业务场景：**
1. **临时下架**：库存补货中、价格调整中、供应商暂停
2. **永久下架**：停售、违规、供应商终止合作
3. **数据归档**：历史销售数据迁移到归档库

**参与角色：**
- 运营人员：下架操作、归档确认
- 系统：自动下架（缺货）、数据迁移

**系统架构：**
```
运营后台下架操作 / 系统自动触发
    ↓
OffShelfService
    ├─ 下架类型判断（临时/永久）
    ├─ 处理中订单检查
    ├─ 搜索索引更新（不可搜）
    └─ 缓存清理
    ↓
永久下架 → 归档流程
    ├─ 数据迁移（MySQL → 归档库）
    ├─ 销售统计（订单数、GMV）
    ├─ 清理热数据（可选）
    └─ 归档报告
```

**核心代码实现：**
```go
// 下架服务
type OffShelfService struct {
    productRepo    *ProductRepo
    orderClient    rpc.OrderClient
    searchClient   rpc.SearchClient
    cacheService   *CacheService
    archiveService *ArchiveService
}

func (s *OffShelfService) OffShelf(ctx context.Context, req *OffShelfRequest) error {
    // Step 1: 检查是否有处理中的订单
    activeOrders, _ := s.orderClient.CheckActiveOrders(ctx, req.SkuID)
    if len(activeOrders) > 0 && req.OffShelfType == "PERMANENT" {
        return fmt.Errorf("存在%d个处理中的订单，无法永久下架", len(activeOrders))
    }
    
    // Step 2: 更新商品状态
    product, _ := s.productRepo.GetByID(ctx, req.SkuID)
    oldStatus := product.Status
    
    if req.OffShelfType == "TEMPORARY" {
        product.Status = "OFF_SHELF_TEMP"
        product.OffShelfReason = req.Reason
        product.ExpectedOnShelfTime = req.ExpectedOnShelfTime
    } else {
        product.Status = "OFF_SHELF_PERMANENT"
        product.OffShelfReason = req.Reason
        product.OffShelfTime = time.Now()
    }
    
    s.productRepo.Update(ctx, product)
    
    // Step 3: 更新搜索索引（不可搜）
    s.searchClient.RemoveFromIndex(ctx, req.SkuID)
    
    // Step 4: 清理缓存
    s.cacheService.InvalidateProduct(ctx, req.SkuID)
    
    // Step 5: 永久下架 → 触发归档流程
    if req.OffShelfType == "PERMANENT" {
        go s.archiveService.ArchiveProduct(ctx, req.SkuID)
    }
    
    // Step 6: 记录操作日志
    s.productRepo.LogOperation(ctx, &OperationLog{
        SkuID:      req.SkuID,
        Operation:  "OFF_SHELF",
        OldStatus:  oldStatus,
        NewStatus:  product.Status,
        Reason:     req.Reason,
        OperatorID: req.OperatorID,
        CreatedAt:  time.Now(),
    })
    
    return nil
}

// 归档服务
type ArchiveService struct {
    productRepo *ProductRepo
    orderClient rpc.OrderClient
    archiveDB   *gorm.DB
}

func (s *ArchiveService) ArchiveProduct(ctx context.Context, skuID int64) error {
    // Step 1: 查询商品完整信息
    product, _ := s.productRepo.GetByID(ctx, skuID)
    
    // Step 2: 查询销售统计
    salesStats, _ := s.orderClient.GetSalesStats(ctx, skuID)
    
    // Step 3: 构建归档记录
    archive := &ProductArchive{
        SkuID:         skuID,
        ProductData:   marshal(product),
        TotalOrders:   salesStats.TotalOrders,
        TotalSales:    salesStats.TotalSales,
        TotalGMV:      salesStats.TotalGMV,
        FirstSaleTime: salesStats.FirstSaleTime,
        LastSaleTime:  salesStats.LastSaleTime,
        ArchiveTime:   time.Now(),
        ArchiveReason: product.OffShelfReason,
    }
    
    // Step 4: 写入归档库
    err := s.archiveDB.Create(archive).Error
    if err != nil {
        return fmt.Errorf("归档失败: %w", err)
    }
    
    // Step 5: 清理热数据（可选，根据数据保留策略）
    // 保留6个月内的数据在主库，超过6个月的迁移到归档库
    if time.Since(product.OffShelfTime) > 180*24*time.Hour {
        // s.productRepo.Delete(ctx, skuID)
    }
    
    log.Infof("商品归档完成: sku=%d, orders=%d, gmv=%d", 
        skuID, salesStats.TotalOrders, salesStats.TotalGMV)
    
    return nil
}
```

**监控指标：**
- 下架商品数（临时/永久，按原因分类）
- 临时下架商品重新上架率
- 归档任务成功率（> 99%）
- 归档数据量（按月统计）
- 归档延迟（下架 → 归档完成，< 1小时）

**预计代码量：** ~140行

---

## 4. 整体架构视图

### B端完整链路流程图

```
┌─────────────────────────────────────────────────────────┐
│                   B端商品生命周期                         │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  阶段1：商品录入                                          │
│  ┌────────┐  ┌────────┐  ┌────────┐                    │
│  │手动创建│  │批量导入│  │API推送│                       │
│  └───┬────┘  └───┬────┘  └───┬────┘                    │
│      └───────────┴───────────┘                          │
│               ↓                                          │
│       ListingService（幂等性、状态机）                    │
│               ↓                                          │
│  ─────────────────────────────────────────────          │
│                                                          │
│  阶段2：审核发布                                          │
│  ┌────────┐  ┌────────┐                                │
│  │人工审核│  │自动审核│                                  │
│  └───┬────┘  └───┬────┘                                │
│      └───────┬───┘                                      │
│              ↓                                           │
│      审核引擎（规则引擎）                                  │
│              ↓                                           │
│  写入商品中心 + 初始化库存/价格                            │
│              ↓                                           │
│  ─────────────────────────────────────────────          │
│                                                          │
│  阶段3：供应商同步                                        │
│  ┌────────┐  ┌────────┐  ┌────────┐                    │
│  │定时同步│  │实时推送│  │Webhook│                       │
│  └───┬────┘  └───┬────┘  └───┬────┘                    │
│      └───────────┴───────────┘                          │
│               ↓                                          │
│  Supplier Gateway（防腐层）+ SyncService                 │
│               ↓                                          │
│  Upsert + 差异检测 + 差异化审核                           │
│               ↓                                          │
│  ─────────────────────────────────────────────          │
│                                                          │
│  阶段4：库存管理 ✨                                       │
│  ┌────────┐  ┌────────┐  ┌────────┐                    │
│  │库存同步│  │水位监控│  │对账任务│                       │
│  └───┬────┘  └───┬────┘  └───┬────┘                    │
│      └───────────┴───────────┘                          │
│               ↓                                          │
│  InventoryManagementService                             │
│      ├─ 同步策略（实时/定时）                             │
│      ├─ 预警规则（安全库存/缺货）                          │
│      └─ 对账机制（每日）                                  │
│               ↓                                          │
│  ─────────────────────────────────────────────          │
│                                                          │
│  阶段5：日常维护                                          │
│  ┌────────┐  ┌────────┐                                │
│  │单品编辑│  │批量编辑│                                  │
│  └───┬────┘  └───┬────┘                                │
│      └───────┬───┘                                      │
│              ↓                                           │
│      EditService + BatchEditWorker                      │
│              ↓                                           │
│  更新商品 + 刷新索引 + 失效缓存                            │
│              ↓                                           │
│  ─────────────────────────────────────────────          │
│                                                          │
│  阶段6：促销配置 ✨                                       │
│  ┌────────┐  ┌────────┐  ┌────────┐                    │
│  │活动关联│  │价格设置│  │时间调度│                       │
│  └───┬────┘  └───┬────┘  └───┬────┘                    │
│      └───────────┴───────────┘                          │
│               ↓                                          │
│  PromotionConfigService（校验+调度）                     │
│               ↓                                          │
│  生效/失效任务（刷新缓存+索引）                            │
│               ↓                                          │
│  ─────────────────────────────────────────────          │
│                                                          │
│  阶段7：下架归档                                          │
│  ┌────────┐  ┌────────┐                                │
│  │临时下架│  │永久下架│                                  │
│  └───┬────┘  └───┬────┘                                │
│      └───────┬───┘                                      │
│              ↓                                           │
│      OffShelfService                                    │
│              ↓                                           │
│  永久下架 → ArchiveService（数据迁移）                    │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### 与C端链路对比

| 维度 | B端链路 | C端链路 |
|-----|--------|--------|
| **阶段数量** | 7个阶段 | 5个阶段 |
| **时间跨度** | 数天到数月（商品生命周期） | 数分钟（单次购物流程） |
| **参与角色** | 供应商、运营、系统 | 用户、系统 |
| **核心关注** | 数据准确性、流程合规性 | 用户体验、转化率 |
| **关键技术** | 幂等性、状态机、异步任务 | 聚合编排、快照、Saga |

---

## 5. 实施方案

### 内容复用策略

| 新阶段 | 复用内容来源 | 复用比例 | 新增内容 |
|-------|------------|---------|---------|
| 阶段1 | 原16.6.6商品上架系统 | 80% | 批量导入流程、监控指标 |
| 阶段2 | 原16.6.6审核逻辑 | 50% | 审核引擎、规则配置、监控 |
| 阶段3 | 原16.6.6供应商同步系统 | 80% | 差异检测详细代码、监控 |
| 阶段4 | **全新内容** | 0% | 库存同步策略、水位监控、对账 |
| 阶段5 | 原16.6.6运营编辑系统 | 80% | 监控指标采集 |
| 阶段6 | **全新内容** | 0% | 促销配置、价格校验、调度 |
| 阶段7 | 部分下架逻辑 | 30% | 归档流程、数据迁移 |

### 实施步骤

1. **重命名现有16.6.6**
   - 标题改为"B端商品生命周期完整链路"
   - 添加引导语（说明7个阶段）

2. **重构阶段1**
   - 保留现有"商品上架系统"代码
   - 补充批量导入代码
   - 添加监控指标

3. **新增阶段2**
   - 提取现有审核逻辑
   - 补充审核引擎设计
   - 添加监控指标

4. **重构阶段3**
   - 保留现有"供应商同步系统"代码
   - 补充差异检测详细实现
   - 添加监控指标

5. **新增阶段4**（核心新增）
   - 库存同步策略（~70行）
   - 水位监控机制（~60行）
   - 对账任务（~70行）

6. **重构阶段5**
   - 保留现有"运营编辑系统"代码
   - 补充监控指标采集

7. **新增阶段6**（核心新增）
   - 促销配置服务（~100行）
   - 生效/失效任务（~80行）

8. **新增阶段7**
   - 下架服务（~70行）
   - 归档服务（~70行）

9. **调整16.6.7编号**
   - C端链路保持不变
   - 后续章节无需调整（16.6.8 DDD战术设计等）

### 预计工作量

- **代码编写**：1150行（复用600行，新增550行）
- **章节重构**：调整现有内容布局
- **监控指标**：补充所有阶段的监控指标
- **验证构建**：确保构建成功

---

## 6. 关键设计决策

### 决策1：为什么是7个阶段而不是5个？

**理由：**
- B端流程更复杂（涉及审核、同步、库存管理）
- 促销配置是独立业务（不应合并到其他阶段）
- 下架归档是完整生命周期的必要环节

### 决策2：为什么库存管理独立为一个阶段？

**理由：**
- 库存管理是供应商管理的核心能力
- 涉及实时同步、定时同步、预警、对账等多个子系统
- 内容量达200行，足以独立成阶段

### 决策3：为什么促销配置独立为一个阶段？

**理由：**
- 促销是电商的核心运营手段
- 涉及价格校验、库存检查、时间调度等复杂逻辑
- 与日常维护（编辑商品信息）是不同的业务场景

### 决策4：内容复用 vs 全新编写的权衡

**复用策略：**
- 阶段1/3/5：保留现有代码框架（80%），补充监控和细节
- 阶段2：提取现有审核逻辑（50%），补充审核引擎
- 阶段4/6：全新编写（0%复用）
- 阶段7：部分复用下架逻辑（30%），新增归档流程

**理由：**
- 最大化复用现有高质量代码
- 重点投入到核心新增内容（库存管理、促销配置）
- 避免重复劳动，提升实施效率

---

## 7. 与现有章节的关联

### 与16.2品类分析的关联

- 阶段3供应商同步：引用16.2.5的差异化设计策略
- 阶段4库存管理：引用16.2的库存模型（实时/池化/无限）

### 与16.6.2库存系统设计的关联

- 阶段4库存管理：引用16.6.2的二维库存模型
- 阶段4对账机制：补充16.6.2未覆盖的供应商对账

### 与16.6.8 DDD战术设计的关联

- 各阶段状态机设计遵循DDD原则
- 领域事件发布机制（ListingTaskCreated、ProductSynced等）

### 与16.6.9 ADR的关联

- 阶段2差异化审核策略（高信用供应商自动通过）
- 阶段4库存同步策略（按品类差异化）

---

## 8. 监控指标体系

### 全局指标

| 指标 | 目标值 | 告警阈值 |
|-----|-------|---------|
| B端链路完整性 | 100% | < 95% |
| 商品数据准确率 | > 99% | < 98% |
| 平均上架时长 | < 24小时 | > 48小时 |
| 供应商同步延迟 | < 1小时 | > 2小时 |

### 各阶段指标汇总

| 阶段 | 核心指标 | 目标值 |
|-----|---------|--------|
| 阶段1 | 录入成功率 | > 95% |
| 阶段2 | 审核通过率 | > 85% |
| 阶段3 | 同步成功率 | > 98% |
| 阶段4 | 库存准确率 | > 99.5% |
| 阶段5 | 编辑成功率 | > 99% |
| 阶段6 | 生效准时率 | > 99.9% |
| 阶段7 | 归档成功率 | > 99% |

---

## 9. 风险与应对

### 风险1：内容重复

**风险**：原16.6.6的上架/同步/编辑代码可能与新阶段重复

**应对**：
- 采用引用方式，避免完全复制代码
- 新阶段补充原有内容未覆盖的部分（监控、异常处理）
- 保持代码示例的差异化（不同场景、不同视角）

### 风险2：新增内容质量

**风险**：阶段4库存管理、阶段6促销配置是全新内容，质量需要保证

**应对**：
- 参考16.6.2库存系统设计的代码风格
- 参考16.6.9 ADR的决策逻辑
- 保持与现有代码的一致性（命名、结构、注释风格）

### 风险3：篇幅过长

**风险**：新增1150行后，16.6章节可能过长（接近6000行）

**应对**：
- 每个阶段严格控制在150-200行
- 代码示例精简，只展示核心逻辑
- 避免冗余说明

---

## 10. 验收标准

### 内容完整性

- [ ] 7个阶段全部完成
- [ ] 每个阶段包含：业务场景、系统架构、核心代码、监控指标
- [ ] 阶段4和阶段6的全新内容达到与其他阶段相同的质量
- [ ] 所有代码示例可编译（Go语法正确）

### 结构一致性

- [ ] 与16.6.7 C端链路的结构风格一致
- [ ] 阶段标题统一格式：`### 阶段X：XXX`
- [ ] 代码块有语言标记（```go）
- [ ] 表格格式正确

### 集成验证

- [ ] 构建成功：`npm run build`
- [ ] 无markdown语法错误
- [ ] 章节编号正确（16.6.7不变）
- [ ] 内部引用链接正确

### 技术质量

- [ ] Go代码符合项目规范
- [ ] 监控指标设计合理
- [ ] 系统架构图清晰
- [ ] 与现有章节无冲突

---

## 11. 后续优化方向

### 短期

- 补充流程图（Mermaid时序图）
- 补充表格（各阶段对比表）

### 中期

- 补充实际案例（某次大促前的批量上架）
- 补充异常处理场景（供应商故障、数据冲突）

### 长期

- 补充与第10章（供给运营系统）的关联
- 补充国际化场景（跨境供应商）

---

## 附录：章节内容分配

### 阶段1：商品录入（180行）

- 业务场景（3个场景）：30行
- 系统架构图：20行
- 核心代码（单品+批量）：80行
- 监控指标：20行
- 小结：30行

### 阶段2：审核发布（150行）

- 业务场景（3个场景）：30行
- 系统架构图：20行
- 核心代码（审核引擎+审核通过）：60行
- 监控指标：20行
- 小结：20行

### 阶段3：供应商同步（160行）

- 业务场景（3个场景）：30行
- 系统架构图：20行
- 核心代码（Upsert+差异检测）：70行
- 监控指标：20行
- 小结：20行

### 阶段4：库存管理（200行）✨全新

- 业务场景（5个场景）：40行
- 系统架构图：25行
- 核心代码（同步策略+监控+对账）：100行
- 监控指标：20行
- 小结：15行

### 阶段5：日常维护（140行）

- 业务场景（3个场景）：30行
- 系统架构图：20行
- 核心代码（单品+批量编辑）：50行
- 监控指标：20行
- 小结：20行

### 阶段6：促销配置（180行）✨全新

- 业务场景（4个场景）：35行
- 系统架构图：25行
- 核心代码（配置+校验+调度）：90行
- 监控指标：20行
- 小结：10行

### 阶段7：下架归档（140行）

- 业务场景（3个场景）：30行
- 系统架构图：20行
- 核心代码（下架+归档）：60行
- 监控指标：20行
- 小结：10行

**总计：1150行**

---

## 设计确认

该设计已与用户确认，包括：
- ✅ 7个阶段的完整生命周期
- ✅ 对标C端链路的内容深度（每阶段150-200行）
- ✅ 综合库存管理（技术+业务）
- ✅ 整合现有内容，聚焦核心新增

**状态**：设计已批准，待实施
