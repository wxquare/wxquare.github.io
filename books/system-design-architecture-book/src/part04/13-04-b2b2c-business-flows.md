# 32.7-32.8 完整业务链路与 DDD 战术设计

## 32.7 完整业务链路（系统集成与数据流）

**从子系统到完整链路**：前面章节展示了各个子系统的内部设计（商品中心、库存、订单、支付、供应商集成），本章展示这些子系统如何协作，形成端到端的业务链路。

**两条关键链路**：
- **B端链路**（供应商 → 运营 → 平台）：商品生命周期管理，决定"商品如何进入、如何管理、如何退出"
- **C端链路**（用户 → 交易 → 履约）：交易流完整路径，决定"用户如何发现、如何下单、如何完成支付"

**集成视角的关键点**：
- **数据流转**：跨系统的数据传递（事件驱动、同步调用、异步任务）
- **状态同步**：多系统间的状态一致性（商品状态、库存状态、订单状态）
- **异常处理**：跨系统的容错与补偿（Saga、重试、降级）

---

### 32.7.1 B端商品生命周期完整链路

**B端商品生命周期是平台运营的核心能力**，决定了"商品如何进入、如何管理、如何退出"。本节展示从商品录入到下架归档的完整链路，涵盖供应商、运营、系统三方协作。

**完整生命周期（7个阶段）**：

```text
阶段1：商品录入（手动/批量/API）
   ↓ 录入成功率 > 95%
阶段2：审核发布（人工/自动）
   ↓ 审核通过率 > 85%
阶段3：供应商同步（实时/定时）
   ↓ 同步成功率 > 98%
阶段4：库存管理（同步/监控/对账）
   ↓ 库存准确率 > 99.5%
阶段5：日常维护（单品/批量编辑）
   ↓ 编辑成功率 > 99%
阶段6：促销配置（活动关联/价格设置）
   ↓ 生效准时率 > 99.9%
阶段7：下架归档（临时/永久）
   ↓ 归档成功率 > 99%
```

**与C端链路的对比**：

| 维度 | B端链路 | C端链路 |
|-----|--------|--------|
| **阶段数量** | 7个阶段 | 5个阶段 |
| **时间跨度** | 数天到数月（商品生命周期） | 数分钟（单次购物流程） |
| **参与角色** | 供应商、运营、系统 | 用户、系统 |
| **核心关注** | 数据准确性、流程合规性 | 用户体验、转化率 |
| **关键技术** | 幂等性、状态机、异步任务 | 聚合编排、快照、Saga |

---

#### 阶段1：商品录入

**业务场景**：
- **手动录入**：运营人员通过后台表单录入新商品（小批量、高质量）
- **批量导入**：商家/运营通过Excel批量导入（大促前、品类扩展）
- **API推送**：供应商通过OpenAPI实时推送新品（自动化、规模化）

**技术难点**：
- **幂等性保证**：防止重复提交（网络超时、用户重试）
- **异步解耦**：批量导入通过异步任务处理，避免阻塞用户
- **数据校验**：必填字段、格式校验、业务规则校验

**核心设计**：

```go
// 上架任务状态机
type ListingStatus string

const (
    ListingDraft      ListingStatus = "DRAFT"       // 草稿
    ListingPending    ListingStatus = "PENDING"     // 待审核
    ListingApproved   ListingStatus = "APPROVED"    // 审核通过
    ListingRejected   ListingStatus = "REJECTED"    // 审核驳回
    ListingPublished  ListingStatus = "PUBLISHED"   // 已发布
)

// 上架任务
type ListingTask struct {
    TaskCode    string        // 幂等性标识
    ItemInfo    ItemInfo      // 商品信息
    SupplierID  int64         // 供应商ID
    Status      ListingStatus
    ReviewerID  int64         // 审核人
    RejectReason string       // 驳回原因
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// 创建上架任务（幂等性保证）
func (s *ListingService) CreateListingTask(ctx context.Context, req *ListingRequest) (*ListingTask, error) {
    // Step 1: 生成幂等性标识符
    taskCode := s.generateTaskCode(req)
    
    // Step 2: FirstOrCreate（幂等性）
    task := &ListingTask{
        TaskCode:   taskCode,
        ItemInfo:   req.ItemInfo,
        SupplierID: req.SupplierID,
        Status:     ListingDraft,
    }
    
    result := s.db.Where("task_code = ?", taskCode).FirstOrCreate(task)
    if result.RowsAffected > 0 {
        // 首次创建，发布事件
        s.eventPublisher.Publish(ctx, &ListingTaskCreatedEvent{
            TaskCode: taskCode,
            ItemInfo: req.ItemInfo,
        })
    }
    
    return task, nil
}

// 提交审核
func (s *ListingService) SubmitForReview(ctx context.Context, taskCode string) error {
    // 状态转换：DRAFT → PENDING
    return s.updateStatus(ctx, taskCode, ListingDraft, ListingPending)
}

// 审核通过
func (s *ListingService) Approve(ctx context.Context, taskCode string, reviewerID int64) error {
    // Step 1: 状态转换：PENDING → APPROVED
    if err := s.updateStatus(ctx, taskCode, ListingPending, ListingApproved); err != nil {
        return err
    }
    
    // Step 2: 创建商品记录（写入商品中心）
    task, _ := s.getTask(ctx, taskCode)
    itemID, err := s.productCenter.CreateProduct(ctx, &CreateProductRequest{
        ItemInfo:   task.ItemInfo,
        SupplierID: task.SupplierID,
    })
    if err != nil {
        return fmt.Errorf("create product failed: %w", err)
    }
    
    // Step 3: 初始化库存（调用库存服务）
    s.inventoryClient.InitStock(ctx, itemID, task.ItemInfo.InitStock)
    
    // Step 4: 初始化价格（调用计价服务）
    s.pricingClient.InitPrice(ctx, itemID, task.ItemInfo.BasePrice)
    
    // Step 5: 更新搜索索引（异步）
    s.eventPublisher.Publish(ctx, &ProductCreatedEvent{
        ItemID:     itemID,
        ItemInfo:   task.ItemInfo,
        SupplierID: task.SupplierID,
    })
    
    return nil
}
```

**批量导入**：

```go
// 批量导入服务
type BatchImportService struct {
    taskRepo     TaskRepository
    taskQueue    TaskQueue
    validator    ItemValidator
    fileParser   FileParser
}

// 批量导入（异步）
func (s *BatchImportService) BatchImport(ctx context.Context, file io.Reader, operatorID int64) (*BatchImportTask, error) {
    // Step 1: 解析文件（支持Excel/CSV）
    items, parseErr := s.fileParser.Parse(file)
    if parseErr != nil {
        return nil, fmt.Errorf("文件解析失败: %w", parseErr)
    }
    
    // Step 2: 数据校验（预检）
    validItems := make([]*ItemInfo, 0, len(items))
    invalidItems := make([]*ValidationError, 0)
    
    for _, item := range items {
        if err := s.validator.Validate(item); err != nil {
            invalidItems = append(invalidItems, &ValidationError{
                Item:  item,
                Error: err.Error(),
            })
        } else {
            validItems = append(validItems, item)
        }
    }
    
    // Step 3: 创建批量任务
    batchTask := &BatchImportTask{
        TaskID:       generateTaskID(),
        TotalCount:   len(items),
        ValidCount:   len(validItems),
        InvalidCount: len(invalidItems),
        Status:       "PENDING",
        OperatorID:   operatorID,
        CreatedAt:    time.Now(),
    }
    
    if err := s.taskRepo.Save(ctx, batchTask); err != nil {
        return nil, err
    }
    
    // Step 4: 发送到任务队列（异步处理）
    s.taskQueue.Publish(ctx, &BatchImportEvent{
        TaskID:       batchTask.TaskID,
        Items:        validItems,
        InvalidItems: invalidItems,
    })
    
    return batchTask, nil
}

// 批量任务处理（Consumer）
func (s *BatchImportService) ProcessBatchTask(ctx context.Context, event *BatchImportEvent) error {
    successCount := 0
    failedItems := make([]*ImportFailure, 0)
    
    // 逐条处理（控制并发度）
    semaphore := make(chan struct{}, 10) // 限制10并发
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, item := range event.Items {
        wg.Add(1)
        semaphore <- struct{}{}
        
        go func(item *ItemInfo) {
            defer wg.Done()
            defer func() { <-semaphore }()
            
            // 创建上架任务
            if _, err := s.listingService.CreateListingTask(ctx, &ListingRequest{
                ItemInfo:   item,
                SupplierID: item.SupplierID,
            }); err != nil {
                mu.Lock()
                failedItems = append(failedItems, &ImportFailure{
                    Item:  item,
                    Error: err.Error(),
                })
                mu.Unlock()
            } else {
                mu.Lock()
                successCount++
                mu.Unlock()
            }
        }(item)
    }
    
    wg.Wait()
    
    // 更新任务状态
    s.taskRepo.Update(ctx, event.TaskID, &BatchImportResult{
        Status:       "COMPLETED",
        SuccessCount: successCount,
        FailedCount:  len(failedItems),
        FailedItems:  failedItems,
        CompletedAt:  time.Now(),
    })
    
    return nil
}
```

**API推送**：

```go
// OpenAPI Service（供应商接口）
type OpenAPIService struct {
    listingService *ListingService
    rateLimiter    RateLimiter      // 限流器
    authService    AuthService      // 鉴权
}

// API推送商品
func (s *OpenAPIService) PushProduct(ctx context.Context, req *PushProductRequest) (*PushProductResponse, error) {
    // Step 1: 鉴权（API Key + Signature）
    supplierID, err := s.authService.Authenticate(ctx, req.APIKey, req.Signature)
    if err != nil {
        return nil, fmt.Errorf("鉴权失败: %w", err)
    }
    
    // Step 2: 限流（防止刷接口）
    if !s.rateLimiter.Allow(supplierID) {
        return nil, fmt.Errorf("请求过于频繁，请稍后重试")
    }
    
    // Step 3: 参数校验
    if err := s.validatePushRequest(req); err != nil {
        return nil, fmt.Errorf("参数校验失败: %w", err)
    }
    
    // Step 4: 创建上架任务（幂等性）
    task, err := s.listingService.CreateListingTask(ctx, &ListingRequest{
        ItemInfo:   req.ItemInfo,
        SupplierID: supplierID,
    })
    if err != nil {
        return nil, fmt.Errorf("创建任务失败: %w", err)
    }
    
    // Step 5: 自动提交审核（API推送默认进入审核）
    if err := s.listingService.SubmitForReview(ctx, task.TaskCode); err != nil {
        return nil, err
    }
    
    return &PushProductResponse{
        TaskCode: task.TaskCode,
        Status:   string(task.Status),
        Message:  "商品已提交审核，预计1-3小时内完成",
    }, nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **录入成功率** | > 95% | 按录入方式（手动/批量/API） |
| **批量导入耗时** | < 10s/千条 | 按文件大小 |
| **API响应时间** | P99 < 500ms | 按供应商 |
| **幂等性命中率** | < 5% | 按操作类型 |

---

#### 阶段2：审核发布

**业务场景**：
- **人工审核**：高风险商品（特定类目、新供应商）需人工审核
- **自动审核**：低风险商品通过规则引擎自动审核通过
- **审核驳回**：不合规商品驳回并通知修改

**技术难点**：
- **规则引擎**：多维度规则组合（合规性、完整性、准确性）
- **审核SLA**：自动审核秒级响应，人工审核小时级
- **审核日志**：完整记录审核决策，支持溯源

**自动审核规则引擎**：

```go
// 审核引擎
type ReviewEngine struct {
    rules []ReviewRule
}

// 审核规则接口
type ReviewRule interface {
    Check(ctx context.Context, task *ListingTask) *ReviewResult
}

// 自动审核
func (e *ReviewEngine) AutoReview(ctx context.Context, task *ListingTask) (*ReviewResult, error) {
    results := make([]*ReviewResult, 0, len(e.rules))
    
    // 执行所有规则
    for _, rule := range e.rules {
        result := rule.Check(ctx, task)
        results = append(results, result)
        
        // 任何规则拒绝则直接返回
        if result.Decision == ReviewReject {
            return result, nil
        }
    }
    
    // 所有规则通过
    return &ReviewResult{
        Decision: ReviewApprove,
        Score:    e.calculateScore(results),
    }, nil
}

// 规则1: 合规性检查
type ComplianceRule struct {
    sensitiveWords []string
    bannedCategories []int64
}

func (r *ComplianceRule) Check(ctx context.Context, task *ListingTask) *ReviewResult {
    // 违禁词检测
    for _, word := range r.sensitiveWords {
        if strings.Contains(task.ItemInfo.Title, word) || 
           strings.Contains(task.ItemInfo.Description, word) {
            return &ReviewResult{
                Decision: ReviewReject,
                Reason:   fmt.Sprintf("包含违禁词: %s", word),
            }
        }
    }
    
    // 禁售类目检测
    for _, category := range r.bannedCategories {
        if task.ItemInfo.CategoryID == category {
            return &ReviewResult{
                Decision: ReviewReject,
                Reason:   "该类目禁止销售",
            }
        }
    }
    
    return &ReviewResult{Decision: ReviewApprove}
}

// 规则2: 完整性检查
type CompletenessRule struct{}

func (r *CompletenessRule) Check(ctx context.Context, task *ListingTask) *ReviewResult {
    item := task.ItemInfo
    
    // 必填字段检查
    if item.Title == "" || item.Description == "" || item.BasePrice <= 0 {
        return &ReviewResult{
            Decision: ReviewReject,
            Reason:   "缺少必填字段",
        }
    }
    
    // 图片数量检查
    if len(item.Images) < 3 {
        return &ReviewResult{
            Decision: ReviewReject,
            Reason:   "商品图片不足3张",
        }
    }
    
    return &ReviewResult{Decision: ReviewApprove}
}

// 规则3: 准确性检查
type AccuracyRule struct{}

func (r *AccuracyRule) Check(ctx context.Context, task *ListingTask) *ReviewResult {
    item := task.ItemInfo
    
    // 价格合理性检查
    if item.BasePrice < 1 || item.BasePrice > 1000000 {
        return &ReviewResult{
            Decision: ReviewManual, // 转人工审核
            Reason:   "价格异常，需人工确认",
        }
    }
    
    // 类目匹配检查（示例：通过标题关键词）
    if !r.isCategoryMatched(item.Title, item.CategoryID) {
        return &ReviewResult{
            Decision: ReviewManual,
            Reason:   "类目与标题不匹配，需人工确认",
        }
    }
    
    return &ReviewResult{Decision: ReviewApprove}
}
```

**审核策略**：

| 审核维度 | 检查项 | 风险等级 | 处理策略 |
|---------|--------|---------|---------|
| **合规性** | 违禁词检测、敏感内容 | 高 | 自动拒绝 |
| **完整性** | 必填字段、图片数量 | 中 | 自动拒绝 |
| **准确性** | 价格合理性、类目匹配 | 中 | 转人工审核 |
| **一致性** | SPU/SKU关系、属性匹配 | 低 | 自动通过 |

**审核流程**：

```go
// 审核服务
func (s *ListingService) ProcessReview(ctx context.Context, taskCode string) error {
    task, _ := s.getTask(ctx, taskCode)
    
    // Step 1: 自动审核
    autoResult, err := s.reviewEngine.AutoReview(ctx, task)
    if err != nil {
        return err
    }
    
    switch autoResult.Decision {
    case ReviewApprove:
        // 自动通过 → 直接发布
        return s.Approve(ctx, taskCode, SystemReviewerID)
        
    case ReviewReject:
        // 自动拒绝 → 驳回
        return s.Reject(ctx, taskCode, autoResult.Reason)
        
    case ReviewManual:
        // 转人工审核 → 进入审核队列
        return s.assignToReviewer(ctx, taskCode)
    }
    
    return nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **审核通过率** | > 85% | 按供应商、按类目 |
| **自动审核占比** | > 70% | 按审核决策 |
| **人工审核SLA** | < 4h | P99耗时 |
| **审核驳回率** | < 15% | 按驳回原因 |

---

#### 阶段3：供应商同步

**业务场景**：
- 供应商定时推送商品数据（每小时/每天）
- 供应商实时推送价格/库存变更
- 供应商商品可能已存在，也可能不存在

**核心挑战**：Upsert语义

```text
如果商品存在 → 更新
如果商品不存在 → 创建
```

**实现方案**：

```go
// 供应商同步任务
type SyncTask struct {
    SyncID       string    // 同步批次ID
    SupplierID   int64     // 供应商ID
    SupplierSkuID string   // 供应商SKU ID
    SyncData     SyncData  // 同步数据
    SyncType     string    // FULL/INCREMENTAL
    Status       string    // PENDING/SUCCESS/FAILED
}

// Upsert处理（幂等性保证）
func (s *SyncService) UpsertProduct(ctx context.Context, req *SyncRequest) error {
    // Step 1: 根据供应商SKU ID查询平台商品ID
    mapping, err := s.repo.GetMapping(ctx, req.SupplierID, req.SupplierSkuID)
    
    if err == ErrNotFound {
        // 场景1：商品不存在 → 创建（走上架流程）
        return s.createNewProduct(ctx, req)
    } else {
        // 场景2：商品存在 → 更新（走同步流程）
        return s.updateExistingProduct(ctx, mapping.ItemID, req)
    }
}

// 创建新商品（供应商同步触发的上架）
func (s *SyncService) createNewProduct(ctx context.Context, req *SyncRequest) error {
    // Step 1: 创建上架任务
    task, err := s.listingService.CreateListingTask(ctx, &ListingRequest{
        ItemInfo:   transformToItemInfo(req.SyncData),
        SupplierID: req.SupplierID,
        Source:     "SUPPLIER_SYNC",  // 标记来源
    })
    
    // Step 2: 根据供应商信用等级，决定是否需要审核
    if s.needReview(req.SupplierID) {
        // 低信用供应商：需要人工审核
        task.Status = ListingPending
    } else {
        // 高信用供应商：自动通过
        task.Status = ListingApproved
        s.listingService.Approve(ctx, task.TaskCode, SYSTEM_REVIEWER_ID)
    }
    
    return nil
}

// 更新现有商品（供应商同步）
func (s *SyncService) updateExistingProduct(ctx context.Context, itemID int64, req *SyncRequest) error {
    // Step 1: 对比差异
    existing, _ := s.productCenter.GetProduct(ctx, itemID)
    diff := s.compareDiff(existing, req.SyncData)
    
    // Step 2: 根据差异类型决定是否需要审核
    if diff.HasHighRiskChange() {
        // 高风险变更（价格变化>50%、类目变更）→ 需要审核
        return s.createReviewTask(ctx, itemID, diff)
    } else {
        // 低风险变更（库存、图片）→ 直接更新
        return s.productCenter.UpdateProduct(ctx, itemID, diff)
    }
}

// 判断供应商是否需要审核
func (s *SyncService) needReview(supplierID int64) bool {
    supplier, _ := s.supplierRepo.Get(ctx, supplierID)
    
    // 根据供应商信用等级和历史表现决定
    return supplier.CreditLevel < 3 || supplier.RejectRate > 0.1
}
```

**差异化审核策略**：

| 变更类型 | 变更范围 | 审核策略 | 理由 |
|---------|---------|---------|------|
| **价格变更** | < 10% | 自动通过 | 正常波动 |
| **价格变更** | 10-50% | 需要审核 | 防止错误 |
| **价格变更** | > 50% | 必须审核 + 告警 | 高风险 |
| **库存变更** | 任意 | 自动通过 | 实时性要求高 |
| **标题变更** | 轻微修改 | 自动通过 | 低风险 |
| **类目变更** | 任意 | 必须审核 | 影响搜索 |
| **图片变更** | 任意 | 自动通过 | 低风险 |

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **同步成功率** | > 98% | 按供应商、按同步类型 |
| **同步耗时** | P99 < 5s | 按数据大小 |
| **差异化审核命中率** | 10-20% | 按变更类型 |
| **供应商数据质量** | 错误率 < 5% | 按供应商 |

---

#### 阶段4：库存管理

**业务场景**：
- **实时同步**：热卖商品库存通过WebHook实时推送（减库存事件）
- **定时同步**：长尾商品库存通过定时任务批量拉取（每小时/每天）
- **水位监控**：库存低于阈值时触发告警，通知供应商补货
- **日终对账**：每日对账供应商库存与平台库存，发现差异自动修正

**技术难点**：
- **同步策略**：实时 vs 定时的平衡（成本 vs 准确性）
- **库存准确性**：多方数据源（供应商、订单系统、售后退款）一致性
- **并发控制**：高并发扣减库存时的原子性保证
- **对账修正**：发现差异后的自动修正 vs 人工介入

**核心设计**：

```go
// 库存同步策略（策略模式）
type StockSyncStrategy interface {
    Sync(ctx context.Context, skuID int64) (*SyncResult, error)
}

// 实时同步策略（高价值商品）
type RealtimeStockSyncStrategy struct {
    supplierClient SupplierClient
    inventoryRepo  InventoryRepository
    cache          *redis.Client
}

func (s *RealtimeStockSyncStrategy) Sync(ctx context.Context, skuID int64) (*SyncResult, error) {
    // Step 1: 调用供应商API实时查询库存
    supplierStock, err := s.supplierClient.GetStock(ctx, skuID)
    if err != nil {
        return nil, fmt.Errorf("供应商库存查询失败: %w", err)
    }
    
    // Step 2: 更新库存（数据库 + 缓存）
    if err := s.inventoryRepo.Update(ctx, skuID, supplierStock); err != nil {
        return nil, err
    }
    
    // Step 3: 更新缓存（防止穿透）
    s.cache.Set(ctx, fmt.Sprintf("stock:%d", skuID), supplierStock, 5*time.Minute)
    
    return &SyncResult{
        SKUID:         skuID,
        OldStock:      0, // 旧库存
        NewStock:      supplierStock,
        SyncTime:      time.Now(),
        SyncType:      "REALTIME",
    }, nil
}

// 定时同步策略（长尾商品）
type ScheduledStockSyncStrategy struct {
    supplierClient SupplierClient
    inventoryRepo  InventoryRepository
}

func (s *ScheduledStockSyncStrategy) Sync(ctx context.Context, skuID int64) (*SyncResult, error) {
    // Step 1: 批量拉取供应商库存（减少API调用）
    supplierStocks, err := s.supplierClient.BatchGetStock(ctx, []int64{skuID})
    if err != nil {
        return nil, err
    }
    
    // Step 2: 批量更新数据库
    updates := make(map[int64]int32)
    for _, stock := range supplierStocks {
        updates[stock.SKUID] = stock.Quantity
    }
    
    if err := s.inventoryRepo.BatchUpdate(ctx, updates); err != nil {
        return nil, err
    }
    
    return &SyncResult{
        SKUID:    skuID,
        NewStock: supplierStocks[skuID].Quantity,
        SyncTime: time.Now(),
        SyncType: "SCHEDULED",
    }, nil
}

// 库存同步服务（根据商品分级选择策略）
type StockSyncService struct {
    realtimeStrategy  *RealtimeStockSyncStrategy
    scheduledStrategy *ScheduledStockSyncStrategy
    productRepo       ProductRepository
}

func (s *StockSyncService) SyncStock(ctx context.Context, skuID int64) error {
    // 根据商品热度选择同步策略
    product, _ := s.productRepo.Get(ctx, skuID)
    
    var strategy StockSyncStrategy
    if product.Hotness > 80 { // 热卖商品
        strategy = s.realtimeStrategy
    } else {
        strategy = s.scheduledStrategy
    }
    
    _, err := strategy.Sync(ctx, skuID)
    return err
}
```

**库存水位监控**：

```go
// 库存水位监控器
type StockWatermarkMonitor struct {
    inventoryRepo InventoryRepository
    alertService  AlertService
    supplierClient SupplierClient
}

// 检查库存水位（定时任务，每5分钟执行）
func (m *StockWatermarkMonitor) CheckWatermark(ctx context.Context, skuID int64) error {
    // Step 1: 查询当前库存
    stock, err := m.inventoryRepo.GetStock(ctx, skuID)
    if err != nil {
        return err
    }
    
    // Step 2: 计算水位线（根据历史销量）
    watermark := m.calculateWatermark(ctx, skuID)
    
    // Step 3: 库存低于水位线 → 告警
    if stock.Available < watermark {
        m.alertService.Send(ctx, &Alert{
            Level:   "WARNING",
            Type:    "LOW_STOCK",
            SKUID:   skuID,
            Message: fmt.Sprintf("库存低于水位线（当前: %d, 水位: %d）", stock.Available, watermark),
        })
        
        // Step 4: 通知供应商补货
        m.supplierClient.RequestReplenishment(ctx, &ReplenishmentRequest{
            SKUID:        skuID,
            RequestQty:   watermark * 2, // 建议补货量
            UrgencyLevel: "NORMAL",
        })
    }
    
    return nil
}

// 计算水位线（基于历史销量）
func (m *StockWatermarkMonitor) calculateWatermark(ctx context.Context, skuID int64) int32 {
    // 查询最近7天日均销量
    avgDailySales := m.inventoryRepo.GetAvgDailySales(ctx, skuID, 7)
    
    // 水位线 = 3天销量（安全库存）
    return avgDailySales * 3
}
```

**库存对账**：

```go
// 库存对账任务（每日凌晨执行）
type StockReconciliationJob struct {
    inventoryRepo  InventoryRepository
    supplierClient SupplierClient
    diffRepo       DiffRepository
}

func (j *StockReconciliationJob) Run(ctx context.Context, date time.Time) error {
    // Step 1: 批量拉取所有SKU的供应商库存
    supplierStocks, err := j.supplierClient.GetAllStocks(ctx)
    if err != nil {
        return err
    }
    
    // Step 2: 批量查询平台库存
    platformStocks, err := j.inventoryRepo.GetAllStocks(ctx)
    if err != nil {
        return err
    }
    
    // Step 3: 对比差异
    diffs := make([]*StockDiff, 0)
    for skuID, supplierQty := range supplierStocks {
        platformQty := platformStocks[skuID]
        
        if supplierQty != platformQty {
            diff := &StockDiff{
                SKUID:        skuID,
                SupplierQty:  supplierQty,
                PlatformQty:  platformQty,
                Difference:   supplierQty - platformQty,
                ReconcileDate: date,
            }
            diffs = append(diffs, diff)
        }
    }
    
    // Step 4: 记录差异
    if err := j.diffRepo.BatchSave(ctx, diffs); err != nil {
        return err
    }
    
    // Step 5: 自动修正（差异 < 10% 自动修正，> 10% 人工介入）
    for _, diff := range diffs {
        if math.Abs(float64(diff.Difference)/float64(diff.PlatformQty)) < 0.1 {
            // 小差异：自动修正
            j.inventoryRepo.Update(ctx, diff.SKUID, diff.SupplierQty)
        } else {
            // 大差异：告警 + 人工介入
            j.alertService.Send(ctx, &Alert{
                Level:   "ERROR",
                Type:    "STOCK_MISMATCH",
                SKUID:   diff.SKUID,
                Message: fmt.Sprintf("库存差异过大（供应商: %d, 平台: %d）", diff.SupplierQty, diff.PlatformQty),
            })
        }
    }
    
    return nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **库存准确率** | > 99.5% | 按SKU、按供应商 |
| **实时同步成功率** | > 98% | 按供应商API可用性 |
| **定时同步耗时** | < 10min/批次 | 按SKU数量 |
| **水位告警响应时间** | < 5min | P99 |
| **日终对账差异率** | < 2% | 按供应商 |
| **自动修正覆盖率** | > 80% | 按差异范围 |

---

#### 阶段5：日常维护

**业务场景**：
- 单品编辑（修改标题、描述、图片）
- 批量编辑（批量调价、批量上下架）
- 批量导入导出（Excel操作）

**核心设计**：

```go
// 运营编辑任务
type EditTask struct {
    TaskID      string       // 任务ID
    ItemIDs     []int64      // 商品ID列表（支持批量）
    EditType    string       // SINGLE/BATCH
    Changes     []Change     // 变更内容
    Status      string       // PENDING/EXECUTING/SUCCESS/FAILED
    Progress    int          // 进度（0-100）
    TotalCount  int          // 总数
    SuccessCount int         // 成功数
    FailedCount int          // 失败数
}

// 批量编辑（异步任务）
func (s *EditService) BatchEdit(ctx context.Context, req *BatchEditRequest) (*EditTask, error) {
    // Step 1: 创建批量编辑任务
    task := &EditTask{
        TaskID:     generateTaskID(),
        ItemIDs:    req.ItemIDs,
        EditType:   "BATCH",
        Changes:    req.Changes,
        Status:     "PENDING",
        TotalCount: len(req.ItemIDs),
    }
    s.taskRepo.Save(ctx, task)
    
    // Step 2: 发布异步任务
    s.taskQueue.Publish(ctx, &BatchEditTaskEvent{
        TaskID: task.TaskID,
    })
    
    return task, nil
}

// 批量编辑执行器（异步）
func (w *BatchEditWorker) Execute(ctx context.Context, taskID string) error {
    task, _ := w.taskRepo.Get(ctx, taskID)
    
    // 逐个处理商品
    for i, itemID := range task.ItemIDs {
        err := w.editSingleItem(ctx, itemID, task.Changes)
        
        if err == nil {
            task.SuccessCount++
        } else {
            task.FailedCount++
            log.Errorf("edit item %d failed: %v", itemID, err)
        }
        
        // 更新进度
        task.Progress = (i + 1) * 100 / task.TotalCount
        w.taskRepo.Update(ctx, task)
    }
    
    // 更新任务状态
    if task.FailedCount == 0 {
        task.Status = "SUCCESS"
    } else if task.SuccessCount == 0 {
        task.Status = "FAILED"
    } else {
        task.Status = "PARTIAL_SUCCESS"
    }
    
    return nil
}
```

**进度追踪**：

```go
// 查询任务进度
func (s *EditService) GetTaskProgress(ctx context.Context, taskID string) (*TaskProgress, error) {
    task, _ := s.taskRepo.Get(ctx, taskID)
    
    return &TaskProgress{
        TaskID:       task.TaskID,
        Status:       task.Status,
        Progress:     task.Progress,
        TotalCount:   task.TotalCount,
        SuccessCount: task.SuccessCount,
        FailedCount:  task.FailedCount,
        EstimateLeft: s.estimateTimeLeft(task),
    }, nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **编辑成功率** | > 99% | 按编辑类型（单品/批量） |
| **批量编辑耗时** | < 5s/千条 | 按数据大小 |
| **进度更新频率** | 每秒 | 任务执行期间 |
| **部分成功占比** | < 10% | 按失败原因 |

---

#### 阶段6：促销配置

**业务场景**：
- **活动关联**：将商品关联到大促活动（618、双11）
- **价格设置**：配置活动价、满减、折扣券
- **定时生效**：活动开始时自动生效，结束时自动失效
- **价格验证**：确保活动价 < 原价，防止虚假促销

**技术难点**：
- **定时生效**：活动开始/结束时间精确到秒，需要定时任务支持
- **价格一致性**：促销价变更后需同步到商品中心、搜索、缓存
- **并发控制**：大促开始时大量商品同时生效，避免雪崩

**核心设计**：

```go
// 促销配置服务
type PromotionConfigService struct {
    productRepo    ProductRepository
    pricingClient  PricingClient
    promotionRepo  PromotionRepository
    cache          *redis.Client
}

// 配置促销（运营人员）
func (s *PromotionConfigService) ConfigPromotion(ctx context.Context, req *ConfigPromotionRequest) error {
    // Step 1: 参数校验
    if err := s.validatePromotionConfig(req); err != nil {
        return fmt.Errorf("配置校验失败: %w", err)
    }
    
    // Step 2: 价格验证（活动价 < 原价）
    product, _ := s.productRepo.Get(ctx, req.SKUID)
    if req.PromotionPrice >= product.BasePrice {
        return fmt.Errorf("促销价必须低于原价")
    }
    
    // Step 3: 创建促销配置
    promotionConfig := &PromotionConfig{
        ConfigID:       generateConfigID(),
        SKUID:          req.SKUID,
        ActivityID:     req.ActivityID,
        PromotionPrice: req.PromotionPrice,
        PromotionType:  req.PromotionType, // DISCOUNT/COUPON/FULL_REDUCTION
        StartTime:      req.StartTime,
        EndTime:        req.EndTime,
        Status:         "PENDING", // 待生效
        CreatedBy:      req.OperatorID,
    }
    
    if err := s.promotionRepo.Save(ctx, promotionConfig); err != nil {
        return err
    }
    
    // Step 4: 注册定时任务（生效/失效）
    s.scheduleActivation(ctx, promotionConfig)
    s.scheduleDeactivation(ctx, promotionConfig)
    
    return nil
}

// 批量配置促销（大促场景）
func (s *PromotionConfigService) BatchConfigPromotion(ctx context.Context, req *BatchConfigRequest) (*BatchConfigTask, error) {
    // Step 1: 创建批量任务
    task := &BatchConfigTask{
        TaskID:     generateTaskID(),
        ActivityID: req.ActivityID,
        TotalCount: len(req.Configs),
        Status:     "PENDING",
    }
    
    s.taskRepo.Save(ctx, task)
    
    // Step 2: 异步处理
    s.taskQueue.Publish(ctx, &BatchConfigEvent{
        TaskID:  task.TaskID,
        Configs: req.Configs,
    })
    
    return task, nil
}

// 定时任务：促销生效
type PromotionActivationJob struct {
    promotionRepo  PromotionRepository
    pricingClient  PricingClient
    searchClient   SearchClient
    cache          *redis.Client
}

func (j *PromotionActivationJob) Run(ctx context.Context) {
    // Step 1: 查询即将生效的促销（未来5分钟）
    now := time.Now()
    upcoming := j.promotionRepo.FindUpcoming(ctx, now, now.Add(5*time.Minute))
    
    // Step 2: 逐个生效
    for _, config := range upcoming {
        if time.Now().After(config.StartTime) {
            j.activatePromotion(ctx, config)
        }
    }
}

func (j *PromotionActivationJob) activatePromotion(ctx context.Context, config *PromotionConfig) error {
    // Step 1: 更新价格服务（促销价生效）
    if err := j.pricingClient.UpdatePromotionPrice(ctx, &UpdatePriceRequest{
        SKUID:          config.SKUID,
        PromotionPrice: config.PromotionPrice,
        ValidUntil:     config.EndTime,
    }); err != nil {
        return err
    }
    
    // Step 2: 更新搜索索引（展示促销标签）
    j.searchClient.UpdatePromotionTag(ctx, config.SKUID, config.ActivityID)
    
    // Step 3: 清理缓存（强制刷新）
    j.cache.Del(ctx, fmt.Sprintf("product:%d", config.SKUID))
    j.cache.Del(ctx, fmt.Sprintf("price:%d", config.SKUID))
    
    // Step 4: 更新促销配置状态
    config.Status = "ACTIVE"
    j.promotionRepo.Update(ctx, config)
    
    // Step 5: 发布促销生效事件
    j.eventPublisher.Publish(ctx, &PromotionActivatedEvent{
        SKUID:      config.SKUID,
        ActivityID: config.ActivityID,
        ActiveTime: time.Now(),
    })
    
    return nil
}

// 定时任务：促销失效
type PromotionDeactivationJob struct {
    promotionRepo  PromotionRepository
    pricingClient  PricingClient
    searchClient   SearchClient
    cache          *redis.Client
}

func (j *PromotionDeactivationJob) Run(ctx context.Context) {
    // 查询已过期的促销
    expired := j.promotionRepo.FindExpired(ctx, time.Now())
    
    for _, config := range expired {
        j.deactivatePromotion(ctx, config)
    }
}

func (j *PromotionDeactivationJob) deactivatePromotion(ctx context.Context, config *PromotionConfig) error {
    // Step 1: 恢复原价
    j.pricingClient.RestoreOriginalPrice(ctx, config.SKUID)
    
    // Step 2: 移除促销标签
    j.searchClient.RemovePromotionTag(ctx, config.SKUID)
    
    // Step 3: 清理缓存
    j.cache.Del(ctx, fmt.Sprintf("product:%d", config.SKUID))
    j.cache.Del(ctx, fmt.Sprintf("price:%d", config.SKUID))
    
    // Step 4: 更新状态
    config.Status = "EXPIRED"
    j.promotionRepo.Update(ctx, config)
    
    return nil
}
```

**价格验证策略**：

```go
// 价格验证器
type PriceValidator struct {
    productRepo ProductRepository
}

func (v *PriceValidator) Validate(ctx context.Context, req *ConfigPromotionRequest) error {
    product, _ := v.productRepo.Get(ctx, req.SKUID)
    
    // 规则1: 促销价 < 原价
    if req.PromotionPrice >= product.BasePrice {
        return fmt.Errorf("促销价必须低于原价")
    }
    
    // 规则2: 折扣不能过低（防止价格战）
    discount := float64(product.BasePrice-req.PromotionPrice) / float64(product.BasePrice)
    if discount > 0.7 {
        return fmt.Errorf("折扣过大（> 70%%），需审批")
    }
    
    // 规则3: 价格必须为整数（避免定价错误）
    if req.PromotionPrice%100 != 0 {
        return fmt.Errorf("价格必须为整数（单位：分）")
    }
    
    return nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **生效准时率** | > 99.9% | 按活动、按SKU |
| **失效准时率** | > 99.9% | 按活动、按SKU |
| **价格一致性** | 100% | 商品中心、搜索、缓存 |
| **配置错误率** | < 1% | 按配置类型 |

---

#### 阶段7：下架归档

**业务场景**：
- **临时下架**：商品缺货、质量问题临时下架，后续可恢复
- **永久下架**：商品停产、违规下架，不可恢复
- **订单检查**：下架前检查是否有进行中的订单，避免影响用户
- **历史归档**：永久下架商品归档到历史库，释放主库空间

**技术难点**：
- **订单安全**：下架前必须检查订单状态，防止影响履约
- **数据一致性**：下架需同步到商品中心、搜索、库存、价格
- **归档策略**：历史数据归档到冷存储，降低成本

**核心设计**：

```go
// 下架服务
type OffShelfService struct {
    productRepo   ProductRepository
    orderRepo     OrderRepository
    searchClient  SearchClient
    inventoryClient InventoryClient
    pricingClient PricingClient
}

// 下架商品
func (s *OffShelfService) OffShelf(ctx context.Context, req *OffShelfRequest) error {
    // Step 1: 检查进行中的订单
    activeOrders, err := s.orderRepo.FindActiveOrdersBySKU(ctx, req.SKUID)
    if err != nil {
        return err
    }
    
    if len(activeOrders) > 0 && req.OffShelfType == "PERMANENT" {
        return fmt.Errorf("存在进行中的订单（%d个），暂不能永久下架", len(activeOrders))
    }
    
    // Step 2: 更新商品状态
    product, _ := s.productRepo.Get(ctx, req.SKUID)
    
    if req.OffShelfType == "TEMPORARY" {
        // 临时下架 → OFF_SHELF
        product.Status = "OFF_SHELF"
        product.OffShelfReason = req.Reason
        product.OffShelfTime = time.Now()
    } else {
        // 永久下架 → ARCHIVED
        product.Status = "ARCHIVED"
        product.ArchivedReason = req.Reason
        product.ArchivedTime = time.Now()
    }
    
    s.productRepo.Update(ctx, product)
    
    // Step 3: 从搜索索引中移除
    s.searchClient.RemoveProduct(ctx, req.SKUID)
    
    // Step 4: 冻结库存（防止误售）
    s.inventoryClient.FreezeStock(ctx, req.SKUID)
    
    // Step 5: 移除促销配置
    s.pricingClient.RemovePromotions(ctx, req.SKUID)
    
    // Step 6: 发布下架事件
    s.eventPublisher.Publish(ctx, &ProductOffShelfEvent{
        SKUID:         req.SKUID,
        OffShelfType:  req.OffShelfType,
        Reason:        req.Reason,
        OffShelfTime:  time.Now(),
    })
    
    return nil
}

// 恢复上架（仅临时下架可恢复）
func (s *OffShelfService) RestoreOnShelf(ctx context.Context, skuID int64) error {
    product, _ := s.productRepo.Get(ctx, skuID)
    
    // 只有临时下架可恢复
    if product.Status != "OFF_SHELF" {
        return fmt.Errorf("商品状态不支持恢复（当前状态: %s）", product.Status)
    }
    
    // Step 1: 恢复商品状态
    product.Status = "ON_SHELF"
    s.productRepo.Update(ctx, product)
    
    // Step 2: 恢复搜索索引
    s.searchClient.AddProduct(ctx, product)
    
    // Step 3: 解冻库存
    s.inventoryClient.UnfreezeStock(ctx, skuID)
    
    return nil
}

// 归档服务（永久下架商品）
type ArchiveService struct {
    productRepo     ProductRepository
    archiveRepo     ArchiveRepository
    orderRepo       OrderRepository
    inventoryClient InventoryClient
}

// 归档商品（异步任务，每日凌晨执行）
func (s *ArchiveService) ArchiveProduct(ctx context.Context, skuID int64) error {
    product, _ := s.productRepo.Get(ctx, skuID)
    
    // 只归档永久下架的商品
    if product.Status != "ARCHIVED" {
        return nil
    }
    
    // Step 1: 检查是否有未完成的订单（防御性检查）
    activeOrders, _ := s.orderRepo.FindActiveOrdersBySKU(ctx, skuID)
    if len(activeOrders) > 0 {
        return fmt.Errorf("仍有进行中的订单，暂不归档")
    }
    
    // Step 2: 归档商品数据（写入历史库）
    archiveData := &ArchivedProduct{
        SKUID:        skuID,
        ProductData:  product.ToJSON(),
        ArchivedTime: time.Now(),
    }
    s.archiveRepo.Save(ctx, archiveData)
    
    // Step 3: 归档订单数据
    historicalOrders, _ := s.orderRepo.FindAllOrdersBySKU(ctx, skuID)
    for _, order := range historicalOrders {
        s.archiveRepo.SaveOrder(ctx, &ArchivedOrder{
            OrderID:      order.OrderID,
            SKUID:        skuID,
            OrderData:    order.ToJSON(),
            ArchivedTime: time.Now(),
        })
    }
    
    // Step 4: 删除主库数据（释放空间）
    s.productRepo.Delete(ctx, skuID)
    s.inventoryClient.DeleteStock(ctx, skuID)
    
    return nil
}
```

**监控指标**：

| 指标 | 目标值 | 监控维度 |
|-----|-------|---------|
| **下架成功率** | > 99% | 按下架类型（临时/永久） |
| **订单冲突率** | < 1% | 永久下架前的订单检查 |
| **恢复成功率** | 100% | 临时下架恢复 |
| **归档耗时** | < 5s/商品 | 按数据大小 |
| **归档完整性** | 100% | 商品数据、订单数据 |

---

### 32.7.2 C端交易流完整链路

**交易流是电商的核心价值链**，从用户搜索到完成支付的完整路径。本节展示五个阶段的设计与集成。

**与B端链路的对比**：

| 维度 | B端链路（32.7.1） | C端链路（32.7.2） |
|-----|-----------------|-----------------|
| **参与方** | 供应商、运营、系统 | 用户、系统 |
| **时间跨度** | 数天到数月（商品生命周期） | 数分钟（单次购物流程） |
| **关键技术** | 幂等性、状态机、异步任务 | 聚合编排、快照、Saga |
| **核心关注** | 数据准确性、流程合规性 | 用户体验、转化率 |

---

#### 阶段1：搜索与导购

**业务场景**：用户搜索"iPhone 15"

**系统架构**：

```text
用户输入关键词
    ↓
API Gateway → Search Aggregation
    ↓
Query理解（分词、纠错、意图识别）
    ↓
Elasticsearch召回（相关性排序）
    ↓
Hydrate编排（并发调用多个服务）
    ├─ Product Service（商品信息）
    ├─ Inventory Service（库存状态）
    ├─ Pricing Service（价格计算）
    └─ Marketing Service（活动标签）
    ↓
返回搜索结果
```

**核心代码**：

```go
// 搜索聚合服务
type SearchAggregation struct {
    esClient        *elasticsearch.Client
    productClient   rpc.ProductClient
    inventoryClient rpc.InventoryClient
    pricingClient   rpc.PricingClient
    marketingClient rpc.MarketingClient
}

func (a *SearchAggregation) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
    // Step 1: Query理解（分词、意图识别）
    query := a.parseQuery(req.Keyword)
    
    // Step 2: ES召回（按相关性排序）
    hits, err := a.esClient.Search(ctx, query)
    if err != nil {
        return nil, err
    }
    
    skuIDs := extractSkuIDs(hits)
    
    // Step 3: Hydrate编排（并发调用）
    var products map[int64]*Product
    var stocks map[int64]*Stock
    var prices map[int64]*Price
    var promos map[int64]*PromoInfo
    
    g, ctx := errgroup.WithContext(ctx)
    
    // 并发调用4个服务
    g.Go(func() error {
        products, _ = a.productClient.BatchGet(ctx, skuIDs)
        return nil
    })
    g.Go(func() error {
        stocks, _ = a.inventoryClient.BatchCheck(ctx, skuIDs)
        return nil
    })
    g.Go(func() error {
        priceItems := buildPriceItems(skuIDs)
        prices, _ = a.pricingClient.BatchCalculate(ctx, priceItems)
        return nil
    })
    g.Go(func() error {
        promos, _ = a.marketingClient.BatchGet(ctx, skuIDs, req.UserID)
        // 降级：Marketing故障时使用空促销
        if promos == nil {
            promos = make(map[int64]*PromoInfo)
        }
        return nil
    })
    
    g.Wait()
    
    // Step 4: 聚合结果
    return a.buildSearchResponse(hits, products, stocks, prices, promos), nil
}
```

**性能优化**：
- ES查询：P99 < 50ms
- Hydrate并发：4个服务并发调用，总耗时 < 200ms
- 缓存策略：热门搜索词缓存5分钟

---

#### 阶段2：商品详情页（PDP）

**业务场景**：用户点击商品进入详情页

**核心设计**：

```go
// 详情页聚合服务
func (a *DetailAggregation) GetDetail(ctx context.Context, skuID int64, userID int64) (*DetailResponse, error) {
    // 并发调用5个服务
    var product *Product
    var stock *Stock
    var price *Price
    var promos []*Promotion
    var reviews []*Review
    
    g, ctx := errgroup.WithContext(ctx)
    
    g.Go(func() error {
        product, _ = a.productClient.Get(ctx, skuID)
        return nil
    })
    g.Go(func() error {
        stock, _ = a.inventoryClient.Check(ctx, skuID)
        return nil
    })
    g.Go(func() error {
        price, _ = a.pricingClient.Calculate(ctx, skuID, userID)
        return nil
    })
    g.Go(func() error {
        promos, _ = a.marketingClient.GetPromotions(ctx, skuID, userID)
        return nil
    })
    g.Go(func() error {
        reviews, _ = a.reviewClient.GetTopReviews(ctx, skuID, 5)
        return nil
    })
    
    g.Wait()
    
    // 生成快照（用于后续试算）
    snapshot := a.generateSnapshot(product, price, promos)
    
    return &DetailResponse{
        Product:   product,
        Stock:     stock,
        Price:     price,
        Promos:    promos,
        Reviews:   reviews,
        Snapshot:  snapshot,  // 快照ID，5分钟有效
    }, nil
}
```

---

#### 阶段3：购物车

**业务场景**：用户加购商品

**未登录加购**：

```go
// 未登录用户（Cookie存储）
func (c *CartService) AddToCartAnonymous(ctx context.Context, req *AddCartRequest) error {
    // Step 1: 获取匿名cartID（存储在Cookie）
    cartID := req.AnonymousCartID
    if cartID == "" {
        cartID = generateCartID()
    }
    
    // Step 2: 存储到Redis（TTL=7天）
    cartKey := fmt.Sprintf("cart:anon:%s", cartID)
    cartData, _ := c.redis.Get(ctx, cartKey).Result()
    
    cart := parseCart(cartData)
    cart.AddItem(req.SkuID, req.Quantity)
    
    c.redis.Set(ctx, cartKey, marshal(cart), 7*24*time.Hour)
    
    return nil
}
```

**登录后合并**：

```go
// 用户登录后合并购物车
func (c *CartService) MergeCartOnLogin(ctx context.Context, userID int64, anonymousCartID string) error {
    // Step 1: 获取匿名购物车
    anonCartKey := fmt.Sprintf("cart:anon:%s", anonymousCartID)
    anonCart, _ := c.redis.Get(ctx, anonCartKey).Result()
    
    // Step 2: 获取用户购物车
    userCartKey := fmt.Sprintf("cart:user:%d", userID)
    userCart, _ := c.redis.Get(ctx, userCartKey).Result()
    
    // Step 3: 合并（相同商品累加数量）
    merged := mergeCarts(parseCart(anonCart), parseCart(userCart))
    
    // Step 4: 保存到用户购物车
    c.redis.Set(ctx, userCartKey, marshal(merged), 0)  // 永久存储
    
    // Step 5: 删除匿名购物车
    c.redis.Del(ctx, anonCartKey)
    
    // Step 6: 异步持久化到MySQL（防止Redis丢失）
    c.eventPublisher.Publish(ctx, &CartMergedEvent{
        UserID: userID,
        Items:  merged.Items,
    })
    
    return nil
}
```

---

#### 阶段4：结算页试算

**业务场景**：用户点击"去结算"

**核心设计**：

```go
// 结算页聚合服务
func (a *CheckoutAggregation) Calculate(ctx context.Context, req *CalculateRequest) (*CalculateResponse, error) {
    // Step 1: 判断是否使用快照（ADR-008）
    var products []*Product
    var promos []*Promotion
    
    if req.Snapshot != nil && !req.Snapshot.IsExpired() {
        // 快照未过期，使用快照数据（性能优先）
        products = req.Snapshot.Products
        promos = req.Snapshot.Promos
    } else {
        // 快照过期，实时查询
        products, _ = a.productClient.BatchGet(ctx, req.SkuIDs)
        promos, _ = a.marketingClient.GetPromotions(ctx, req.SkuIDs, req.UserID)
    }
    
    // Step 2: 实时查询库存（不能用快照）
    stocks, _ := a.inventoryClient.BatchCheck(ctx, req.SkuIDs)
    
    // Step 3: 计算价格
    prices, _ := a.pricingClient.BatchCalculate(ctx, products, promos)
    
    // Step 4: 检查可下单性
    canCheckout := a.checkCanCheckout(stocks, req.Items)
    
    return &CalculateResponse{
        Items:       buildItems(products, stocks, prices),
        TotalPrice:  calculateTotal(prices),
        CanCheckout: canCheckout,
        Warnings:    a.generateWarnings(stocks, promos),
    }, nil
}
```

---

#### 阶段5：下单与支付

**完整下单流程**（Saga模式）：

```go
// 订单创建Saga（编排多个服务调用）
type CreateOrderSaga struct {
    productClient   rpc.ProductClient
    inventoryClient rpc.InventoryClient
    pricingClient   rpc.PricingClient
    marketingClient rpc.MarketingClient
    orderRepo       *OrderRepo
}

func (s *CreateOrderSaga) Execute(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    var err error
    var reserved *ReserveResult
    var couponLock *CouponLock
    
    // Step 1: 实时查询商品信息（ADR-009：不使用快照）
    products, err := s.productClient.BatchGet(ctx, req.SkuIDs)
    if err != nil {
        return nil, fmt.Errorf("query products failed: %w", err)
    }
    
    // Step 2: 实时查询营销信息
    promos, err := s.marketingClient.GetPromotions(ctx, req.SkuIDs, req.UserID)
    if err != nil {
        return nil, fmt.Errorf("query promotions failed: %w", err)
    }
    
    // Step 3: 校验营销活动有效性
    for _, promo := range promos {
        if !s.validatePromotion(promo) {
            return nil, fmt.Errorf("promotion %s expired", promo.ID)
        }
    }
    
    // Step 4: 库存预占（CAS操作）
    reserved, err = s.inventoryClient.Reserve(ctx, req.Items)
    if err != nil {
        return nil, fmt.Errorf("库存不足: %w", err)
    }
    defer func() {
        if err != nil {
            // 补偿：释放库存
            s.inventoryClient.Release(ctx, reserved.ReserveID)
        }
    }()
    
    // Step 5: 优惠券锁定
    if req.CouponCode != "" {
        couponLock, err = s.marketingClient.LockCoupon(ctx, req.CouponCode, req.UserID)
        if err != nil {
            return nil, fmt.Errorf("优惠券锁定失败: %w", err)
        }
        defer func() {
            if err != nil {
                // 补偿：释放优惠券
                s.marketingClient.UnlockCoupon(ctx, couponLock.LockID)
            }
        }()
    }
    
    // Step 6: 实时计算价格
    price, err := s.pricingClient.Calculate(ctx, products, promos)
    if err != nil {
        return nil, fmt.Errorf("价格计算失败: %w", err)
    }
    
    // Step 7: 价格校验（ADR-011）
    if req.ExpectedPrice > 0 {
        if err := s.validatePriceChange(req.ExpectedPrice, price.FinalPrice); err != nil {
            return nil, err
        }
    }
    
    // Step 8: 生成商品快照
    snapshot := s.generateProductSnapshot(products, promos, price)
    
    // Step 9: 创建订单
    order := &Order{
        OrderID:         s.generateOrderID(),
        UserID:          req.UserID,
        Items:           req.Items,
        TotalPrice:      price.FinalPrice,
        ProductSnapshot: marshal(snapshot),
        ReserveID:       reserved.ReserveID,
        CouponLockID:    couponLock.LockID,
        Status:          StatusPendingPayment,
        ExpireTime:      time.Now().Add(15 * time.Minute),
    }
    
    err = s.orderRepo.Create(ctx, order)
    if err != nil {
        return nil, fmt.Errorf("订单创建失败: %w", err)
    }
    
    // Step 10: 发布订单创建事件（异步）
    s.eventPublisher.Publish(ctx, &OrderCreatedEvent{
        OrderID: order.OrderID,
        UserID:  order.UserID,
        Items:   order.Items,
    })
    
    return order, nil
}
```

**交易流监控**：

| 阶段 | 关键指标 | 目标值 |
|------|---------|--------|
| **搜索** | 搜索→点击转化率 | > 15% |
| **详情页** | 详情→加购转化率 | > 8% |
| **购物车** | 加购→结算转化率 | > 30% |
| **结算页** | 结算→下单转化率 | > 60% |
| **支付** | 下单→支付转化率 | > 85% |
| **整体** | 搜索→支付转化率 | > 2% |

---

## 32.8 DDD战术设计实践

**领域模型是系统设计的核心**。本节展示如何在订单域应用DDD战术模式。

### 聚合设计：Order聚合根

```go
// Order聚合根
type Order struct {
    // 聚合根ID
    orderID OrderID  // 值对象
    
    // 基本信息
    userID    int64
    shopID    int64
    
    // 订单明细（实体集合）
    items []*OrderItem
    
    // 价格信息（值对象）
    pricing *OrderPricing
    
    // 状态（值对象）
    status OrderStatus
    
    // 时间戳
    createdAt time.Time
    updatedAt time.Time
    
    // 领域事件（未提交）
    domainEvents []DomainEvent
}

// 值对象：OrderID
type OrderID struct {
    value string
}

func NewOrderID() OrderID {
    return OrderID{value: generateSnowflakeID()}
}

func (id OrderID) String() string {
    return id.value
}

// 值对象：OrderPricing
type OrderPricing struct {
    subtotal       int64  // 商品总价
    discount       int64  // 折扣金额
    couponDiscount int64  // 优惠券
    payableAmount  int64  // 应付金额
}

func (p *OrderPricing) Calculate() int64 {
    return p.subtotal - p.discount - p.couponDiscount
}

// 实体：OrderItem
type OrderItem struct {
    itemID    int64
    skuID     int64
    quantity  int
    unitPrice int64
    
    // 快照
    snapshot *ItemSnapshot
}

// 聚合根方法：状态转换
func (o *Order) TransitionTo(newStatus OrderStatus) error {
    // 检查状态转换是否合法
    if !o.status.CanTransitionTo(newStatus) {
        return fmt.Errorf("不允许从 %s 转换到 %s", o.status, newStatus)
    }
    
    oldStatus := o.status
    o.status = newStatus
    o.updatedAt = time.Now()
    
    // 发布领域事件
    o.addDomainEvent(&OrderStatusChangedEvent{
        OrderID:   o.orderID,
        OldStatus: oldStatus,
        NewStatus: newStatus,
        ChangedAt: o.updatedAt,
    })
    
    return nil
}

// 聚合根方法：添加商品项
func (o *Order) AddItem(item *OrderItem) error {
    // 不变量检查：订单金额不能超过限额
    if o.calculateTotal()+item.Total() > MAX_ORDER_AMOUNT {
        return errors.New("订单金额超过限额")
    }
    
    o.items = append(o.items, item)
    
    // 发布领域事件
    o.addDomainEvent(&OrderItemAddedEvent{
        OrderID: o.orderID,
        Item:    item,
    })
    
    return nil
}

// 不变量：订单金额 = 所有商品项之和
func (o *Order) calculateTotal() int64 {
    total := int64(0)
    for _, item := range o.items {
        total += item.Total()
    }
    return total
}

// 领域事件管理
func (o *Order) addDomainEvent(event DomainEvent) {
    o.domainEvents = append(o.domainEvents, event)
}

func (o *Order) DomainEvents() []DomainEvent {
    return o.domainEvents
}

func (o *Order) ClearDomainEvents() {
    o.domainEvents = nil
}
```

### Repository模式

```go
// OrderRepository接口（领域层定义）
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, orderID OrderID) (*Order, error)
    FindByUserID(ctx context.Context, userID int64, limit int) ([]*Order, error)
}

// OrderRepositoryImpl实现（基础设施层）
type OrderRepositoryImpl struct {
    db            *gorm.DB
    eventPublisher EventPublisher
}

func (r *OrderRepositoryImpl) Save(ctx context.Context, order *Order) error {
    // Step 1: 转换聚合根 → 数据模型
    orderDO := r.toDataObject(order)
    
    // Step 2: 保存到数据库
    err := r.db.Transaction(func(tx *gorm.DB) error {
        // 保存订单主表
        if err := tx.Create(orderDO).Error; err != nil {
            return err
        }
        
        // 保存订单明细表
        for _, item := range order.Items() {
            itemDO := r.toItemDataObject(item, orderDO.ID)
            if err := tx.Create(itemDO).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
    
    if err != nil {
        return err
    }
    
    // Step 3: 发布领域事件（事务提交后）
    for _, event := range order.DomainEvents() {
        r.eventPublisher.Publish(ctx, event)
    }
    order.ClearDomainEvents()
    
    return nil
}
```

### 领域事件与Outbox模式

```go
// Outbox表（确保事件必达）
type Outbox struct {
    ID          int64
    EventType   string
    EventData   string  // JSON
    Status      string  // PENDING/PUBLISHED/FAILED
    RetryCount  int
    CreatedAt   time.Time
}

// 发布领域事件（Outbox模式）
func (p *EventPublisher) Publish(ctx context.Context, event DomainEvent) error {
    // Step 1: 序列化事件
    eventData, _ := json.Marshal(event)
    
    // Step 2: 写入Outbox表（与业务在同一事务）
    outbox := &Outbox{
        EventType: event.Type(),
        EventData: string(eventData),
        Status:    "PENDING",
        CreatedAt: time.Now(),
    }
    
    return p.db.Create(outbox).Error
}

// Outbox轮询器（定时扫描未发布的事件）
func (w *OutboxWorker) Run() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // Step 1: 查询待发布事件（PENDING状态）
        var outboxes []*Outbox
        w.db.Where("status = ? AND retry_count < ?", "PENDING", 3).
            Limit(100).
            Find(&outboxes)
        
        // Step 2: 发布到Kafka
        for _, outbox := range outboxes {
            err := w.kafkaProducer.Send(outbox.EventType, outbox.EventData)
            
            if err == nil {
                // 发布成功，标记为PUBLISHED
                w.db.Model(outbox).Update("status", "PUBLISHED")
            } else {
                // 发布失败，重试计数+1
                w.db.Model(outbox).Updates(map[string]interface{}{
                    "retry_count": gorm.Expr("retry_count + 1"),
                    "status":      "FAILED",
                })
            }
        }
    }
}
```

---
