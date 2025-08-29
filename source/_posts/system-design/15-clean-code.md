---
title: 代码：Pipeline Pattern + Service Layer 模式写复杂业务代码
date: 2025-06-20
categories:
- 系统设计
---


# Pipeline Pattern + Service Layer 组合架构详解

## 架构概述

**Pipeline Pattern + Service Layer** 是一种将复杂业务流程分解为可组合、可重用组件的设计模式组合。它将传统的面向过程代码转换为面向对象的、高度模块化的架构。

## 核心设计理念

### 1. 分离关注点 (Separation of Concerns)
- **Controller Layer**: 处理HTTP请求/响应
- **Service Layer**: 封装业务逻辑 
- **Pipeline Layer**: 管理处理流程
- **Processor Layer**: 实现具体的处理步骤

### 2. 单一职责原则 (Single Responsibility Principle)
- 每个Processor只负责一个特定的处理步骤
- 每个Service只负责一个业务领域
- Pipeline只负责流程编排

### 3. 开闭原则 (Open/Closed Principle)
- 对扩展开放：可以轻松添加新的Processor
- 对修改封闭：不需要修改现有代码

## 架构层次详解

### Layer 1: Controller Layer (控制层)
```go
// 职责：处理HTTP请求，参数验证，响应格式化
func FlashSaleListV2(ctx *logic.Context) {
    req := ctx.GetRequest().(*aggregateCmd.FlashSaleListReq)
    
    // 委托给Service层处理业务逻辑
    service := NewFlashSaleService()
    resp, errCode := service.GetFlashSaleList(context.Background(), req)
    
    // 设置响应
    ctx.SetResponse(resp)
}
```

**特点**：
- 薄薄的一层，不包含业务逻辑
- 负责请求/响应的格式转换
- 处理框架相关的逻辑

### Layer 2: Service Layer (服务层)
```go
type FlashSaleService interface {
    GetFlashSaleList(ctx context.Context, req *aggregateCmd.FlashSaleListReq) (*aggregateCmd.FlashSaleListResp, errors.ErrorCode)
}

type flashSaleService struct {
    pipeline Pipeline  // 依赖Pipeline来处理具体流程
}

func (s *flashSaleService) GetFlashSaleList(ctx context.Context, req *aggregateCmd.FlashSaleListReq) (*aggregateCmd.FlashSaleListResp, errors.ErrorCode) {
    // 1. 创建处理上下文
    fsCtx := &FlashSaleContext{Request: req}
    
    // 2. 执行处理管道
    if errCode := s.pipeline.Execute(ctx, fsCtx); !errors.Ok.EqualCode(errCode) {
        return nil, errCode
    }
    
    // 3. 构建响应
    return s.buildResponse(fsCtx), errors.Ok
}
```

**特点**：
- 定义业务接口
- 管理事务边界
- 处理业务异常
- 不包含具体的处理逻辑

### Layer 3: Pipeline Layer (管道层)
```go
type Pipeline interface {
    AddProcessor(processor Processor) Pipeline
    Execute(ctx context.Context, fsCtx *FlashSaleContext) errors.ErrorCode
}

type flashSalePipeline struct {
    processors []Processor
}

func (p *flashSalePipeline) Execute(ctx context.Context, fsCtx *FlashSaleContext) errors.ErrorCode {
    for _, processor := range p.processors {
        if errCode := processor.Process(ctx, fsCtx); !errors.Ok.EqualCode(errCode) {
            return errCode
        }
    }
    return errors.Ok
}
```

**特点**：
- 管理处理器的执行顺序
- 统一的错误处理
- 支持流程编排
- 可插拔的处理器架构

### Layer 4: Processor Layer (处理器层)
```go
type Processor interface {
    Process(ctx context.Context, fsCtx *FlashSaleContext) errors.ErrorCode
    Name() string
}

// 具体处理器示例
type PromotionDataProcessor struct{}

func (p *PromotionDataProcessor) Process(ctx context.Context, fsCtx *FlashSaleContext) errors.ErrorCode {
    // 实现具体的处理逻辑
    // 从营销服务获取数据
    // 设置到fsCtx中
    return errors.Ok
}
```

**特点**：
- 实现具体的处理逻辑
- 可独立测试
- 可重用
- 职责单一

## 数据流转模式

### Context Pattern (上下文模式)
```go
type FlashSaleContext struct {
    // Input - 输入数据
    Request *aggregateCmd.FlashSaleListReq

    // Intermediate - 中间数据，各处理器间传递
    OriginalPromotionItems []*promotionCmd.ActivityItem
    FilteredPromotionItems []*promotionCmd.ActivityItem
    LSItemList             []*lsitemcmd.Item

    // Output - 输出数据
    FlashSaleItems      []*aggregateCmd.FlashSaleItem
    FlashSaleBriefItems []*aggregateCmd.FlashSaleBriefItem
    Session             *aggregateCmd.FlashSaleSession
}
```

**数据流转过程**：
1. Controller创建初始Context
2. 每个Processor读取Context中的数据
3. Processor处理后更新Context
4. 下一个Processor继续处理
5. Service层从Context构建最终响应

## 架构优势分析

### 1. 可测试性 (Testability)
```go
// 单元测试示例
func TestPromotionDataProcessor(t *testing.T) {
    processor := NewPromotionDataProcessor()
    ctx := context.Background()
    fsCtx := &FlashSaleContext{
        Request: mockRequest,
    }
    
    errCode := processor.Process(ctx, fsCtx)
    
    assert.Equal(t, errors.Ok, errCode)
    assert.NotEmpty(t, fsCtx.OriginalPromotionItems)
}
```

### 2. 可扩展性 (Extensibility)
```go
// 添加新功能只需要新增Processor
type CacheProcessor struct{}
type MetricsProcessor struct{}
type ValidationProcessor struct{}

// 在管道中组合
pipeline := NewFlashSalePipeline().
    AddProcessor(NewValidationProcessor()).  // 验证
    AddProcessor(NewCacheProcessor()).       // 缓存
    AddProcessor(NewPromotionDataProcessor()). // 原有逻辑
    AddProcessor(NewMetricsProcessor())      // 监控
```

### 3. 可维护性 (Maintainability)
- **代码职责清晰**：每个组件职责单一
- **错误处理统一**：Pipeline层统一处理
- **日志记录一致**：每个Processor都有统一的日志格式

### 4. 可重用性 (Reusability)
```go
// Processor可以在不同的Pipeline中重用
flashSalePipeline := NewFlashSalePipeline().
    AddProcessor(NewPromotionDataProcessor()).
    AddProcessor(NewItemFilterProcessor())

discountPipeline := NewDiscountPipeline().
    AddProcessor(NewPromotionDataProcessor()). // 重用
    AddProcessor(NewDiscountCalculateProcessor())
```

## 设计模式应用

### 1. Strategy Pattern (策略模式)
```go
// 不同的排序策略
type SortStrategy interface {
    Sort([]*aggregateCmd.FlashSaleItem) []*aggregateCmd.FlashSaleItem
}

type DiscountFirstStrategy struct{}
type StockFirstStrategy struct{}

type FlashSaleSortProcessor struct {
    strategy SortStrategy
}
```

### 2. Chain of Responsibility (责任链模式)
```go
// 每个Processor形成一个责任链
// 请求在链中传递，每个节点都可以处理
Validation → DataFetch → Filter → Assembly → Sort
```

### 3. Template Method Pattern (模板方法模式)
```go
// 基础Processor提供模板
type BaseProcessor struct{}

func (p *BaseProcessor) Process(ctx context.Context, fsCtx *FlashSaleContext) errors.ErrorCode {
    // 模板方法
    if err := p.preProcess(ctx, fsCtx); err != nil {
        return err
    }
    
    if err := p.doProcess(ctx, fsCtx); err != nil {
        return err
    }
    
    return p.postProcess(ctx, fsCtx)
}
```

### 4. Decorator Pattern (装饰器模式)
```go
// 为Processor添加额外功能
type LoggingProcessor struct {
    wrapped Processor
}

type MetricsProcessor struct {
    wrapped Processor
}
```

## 性能优化策略

### 1. 并行处理
```go
type ParallelPipeline struct {
    processors [][]Processor // 二维数组，支持并行执行
}

func (p *ParallelPipeline) Execute(ctx context.Context, fsCtx *FlashSaleContext) errors.ErrorCode {
    for _, parallelGroup := range p.processors {
        // 并行执行同一组的Processor
        errChan := make(chan errors.ErrorCode, len(parallelGroup))
        for _, processor := range parallelGroup {
            go func(proc Processor) {
                errChan <- proc.Process(ctx, fsCtx)
            }(processor)
        }
        
        // 等待所有并行任务完成
        for i := 0; i < len(parallelGroup); i++ {
            if errCode := <-errChan; !errors.Ok.EqualCode(errCode) {
                return errCode
            }
        }
    }
    return errors.Ok
}
```

### 2. 缓存策略
```go
type CacheProcessor struct {
    cache Cache
    ttl   time.Duration
}

func (p *CacheProcessor) Process(ctx context.Context, fsCtx *FlashSaleContext) errors.ErrorCode {
    key := p.buildCacheKey(fsCtx.Request)
    
    // 尝试从缓存获取
    if cached := p.cache.Get(key); cached != nil {
        fsCtx.FlashSaleItems = cached.Items
        return errors.Ok
    }
    
    // 缓存未命中，继续处理
    return errors.Ok
}
```

### 3. 熔断器模式
```go
type CircuitBreakerProcessor struct {
    wrapped        Processor
    circuitBreaker CircuitBreaker
}

func (p *CircuitBreakerProcessor) Process(ctx context.Context, fsCtx *FlashSaleContext) errors.ErrorCode {
    if p.circuitBreaker.IsOpen() {
        return errors.ErrorServiceUnavailable
    }
    
    errCode := p.wrapped.Process(ctx, fsCtx)
    
    if !errors.Ok.EqualCode(errCode) {
        p.circuitBreaker.RecordFailure()
    } else {
        p.circuitBreaker.RecordSuccess()
    }
    
    return errCode
}
```

## 监控和可观测性

### 1. 指标收集
```go
type MetricsCollector interface {
    RecordProcessorLatency(processorName string, duration time.Duration)
    RecordProcessorError(processorName string, errorCode string)
    RecordPipelineExecution(pipelineName string, itemCount int)
}

type PrometheusMetricsCollector struct{}

func (p *PrometheusMetricsCollector) RecordProcessorLatency(processorName string, duration time.Duration) {
    processorLatencyHistogram.WithLabelValues(processorName).Observe(duration.Seconds())
}
```

### 2. 链路追踪
```go
type TracingProcessor struct {
    wrapped Processor
    tracer  opentracing.Tracer
}

func (p *TracingProcessor) Process(ctx context.Context, fsCtx *FlashSaleContext) errors.ErrorCode {
    span, ctx := opentracing.StartSpanFromContext(ctx, p.wrapped.Name())
    defer span.Finish()
    
    span.SetTag("processor.name", p.wrapped.Name())
    span.SetTag("request.platform", fsCtx.Request.GetPlatform())
    
    errCode := p.wrapped.Process(ctx, fsCtx)
    
    if !errors.Ok.EqualCode(errCode) {
        span.SetTag("error", true)
        span.LogFields(log.String("error.code", fmt.Sprintf("%d", errCode.Code)))
    }
    
    return errCode
}
```

## 适用场景

### 适合使用的场景：
1. **复杂的数据处理流程**：需要多个步骤的数据转换
2. **需要灵活配置的业务流程**：不同场景需要不同的处理步骤
3. **高度可测试的代码**：需要单元测试覆盖率
4. **团队协作开发**：不同开发者可以并行开发不同的Processor
5. **需要监控和调试**：需要了解每个步骤的执行情况

### 不适合的场景：
1. **简单的CRUD操作**：过度设计
2. **性能要求极高的场景**：可能引入额外开销
3. **变化很少的稳定流程**：增加复杂性

## 最佳实践

### 1. Processor设计原则
- **无状态**：Processor应该是无状态的
- **幂等性**：相同输入应该产生相同输出
- **快速失败**：尽早发现并报告错误

### 2. Context设计原则
- **不可变性**：尽量避免修改已设置的数据
- **清晰命名**：字段名要清楚表达含义
- **类型安全**：使用强类型而不是interface{}

### 3. Pipeline设计原则
- **顺序重要**：Processor的顺序要有逻辑意义
- **错误传播**：错误要能正确向上传播
- **资源管理**：确保资源得到正确释放



## 业务引擎
- 对于简单的接口逻辑可以直接写过程代码
- 复杂接口可以考虑使用责任链的方式
- 复杂度更高的代码流程控制的方式

## 工作流引擎与任务编排
- https://github.com/s8sg/goflow
- https://github.com/go-workflow/go-workflow
## 规则引擎与风控、资损、校验

- https://github.com/bilibili/gengine

## 脚本执行引擎与低代码平台

- https://github.com/d5/tengo
- https://github.com/mattn/anko

## 总结

**Pipeline Pattern + Service Layer** 组合架构通过将复杂的业务流程分解为独立的、可组合的组件，实现了：

- **高内聚低耦合**的模块化设计
- **易于测试和维护**的代码结构  
- **灵活可配置**的业务流程
- **强大的扩展能力**和**重用性**

这种架构特别适合处理电商、金融等复杂业务场景，能够显著提升代码质量和开发效率。