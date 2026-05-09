# 35.4 综合实战案例题库

## 第四部分：综合实战案例（10题）

本部分将前面所学知识点整合，设计完整的端到端场景，涵盖系统设计、技术选型、性能优化、故障处理等多个维度。

---

#### 🚀 案例1：设计一个百万级QPS的商品详情页系统

**问题描述**：
电商平台的商品详情页是流量最大的页面，大促期间QPS可达100万。请设计一套完整的商品详情页系统，保证高性能和高可用。

**答案**：

**系统架构设计**（Go实现）：

```go
package product

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/go-redis/redis/v8"
)

// ProductDetailService 商品详情服务
type ProductDetailService struct {
	productRepo   ProductRepository
	l1Cache       LocalCache       // L1缓存：本地内存
	l2Cache       *redis.Client    // L2缓存：Redis
	cdn           CDNService       // CDN
	mq            MessageQueue     // 消息队列
}

// GetProductDetail 获取商品详情（多级缓存）
func (s *ProductDetailService) GetProductDetail(ctx context.Context, 
	productID int64) (*ProductDetail, error) {
	
	cacheKey := fmt.Sprintf("product:detail:%d", productID)
	
	// L1缓存：本地内存（命中率60%）
	if detail := s.l1Cache.Get(cacheKey); detail != nil {
		return detail.(*ProductDetail), nil
	}
	
	// L2缓存：Redis（命中率95%）
	detailJSON, err := s.l2Cache.Get(ctx, cacheKey).Result()
	if err == nil {
		detail := &ProductDetail{}
		json.Unmarshal([]byte(detailJSON), detail)
		
		// 回填L1
		s.l1Cache.Set(cacheKey, detail, 5*time.Minute)
		return detail, nil
	}
	
	// L3：数据库（命中率5%）
	detail, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	
	// 异步回填缓存
	go func() {
		detailJSON, _ := json.Marshal(detail)
		s.l2Cache.SetEX(context.Background(), cacheKey, detailJSON, time.Hour)
		s.l1Cache.Set(cacheKey, detail, 5*time.Minute)
	}()
	
	return detail, nil
}

// 缓存预热
func (s *ProductDetailService) WarmUpCache(ctx context.Context, productIDs []int64) error {
	for _, pid := range productIDs {
		detail, err := s.productRepo.FindByID(ctx, pid)
		if err != nil {
			continue
		}
		
		// 预热到Redis
		cacheKey := fmt.Sprintf("product:detail:%d", pid)
		detailJSON, _ := json.Marshal(detail)
		s.l2Cache.SetEX(ctx, cacheKey, detailJSON, time.Hour)
	}
	
	return nil
}

// 缓存失效策略（商品更新时）
func (s *ProductDetailService) UpdateProduct(ctx context.Context, 
	product *Product) error {
	
	// 1. 更新数据库
	if err := s.productRepo.Update(ctx, product); err != nil {
		return err
	}
	
	// 2. 发布缓存失效消息
	msg := CacheInvalidateMessage{
		ProductID: product.ProductID,
		Timestamp: time.Now(),
	}
	
	s.mq.Publish("product.cache.invalidate", msg)
	
	return nil
}

// 监听缓存失效消息
func (s *ProductDetailService) StartCacheInvalidateListener() {
	s.mq.Subscribe("product.cache.invalidate", func(msg *CacheInvalidateMessage) {
		cacheKey := fmt.Sprintf("product:detail:%d", msg.ProductID)
		
		// 删除L1和L2缓存
		s.l1Cache.Delete(cacheKey)
		s.l2Cache.Del(context.Background(), cacheKey)
		
		log.Infof("商品%d缓存已失效", msg.ProductID)
	})
}
```

**CDN静态化**：
```go
// 静态化商品详情页（HTML）
func (s *ProductDetailService) GenerateStaticHTML(ctx context.Context, 
	productID int64) (string, error) {
	
	detail, err := s.GetProductDetail(ctx, productID)
	if err != nil {
		return "", err
	}
	
	// 渲染HTML模板
	html := s.renderTemplate(detail)
	
	// 上传到CDN
	cdnURL := fmt.Sprintf("https://cdn.example.com/product/%d.html", productID)
	if err := s.cdn.Upload(cdnURL, html); err != nil {
		return "", err
	}
	
	return cdnURL, nil
}
```

**性能优化点**：
1. **多级缓存**：本地内存（1ms）→ Redis（5ms）→ MySQL（50ms）
2. **CDN静态化**：核心商品预生成HTML，加载时间<100ms
3. **缓存预热**：大促前1小时预热热门商品
4. **异步刷新**：Cache Aside模式，回源不阻塞请求
5. **降级策略**：Redis故障时降级到MySQL+限流

**容量规划**：
```text
QPS：100万
平均响应时间：50ms
并发连接数：100万 * 0.05 = 5万

服务器配置：
- 应用服务器：100台（每台1万QPS）
- Redis集群：50个主节点（每节点2万QPS）
- MySQL：主从+分库分表（32个分片）
```

**延伸思考**：
1. 商品详情页的AB测试如何设计？
2. 图片加载优化（WebP、懒加载）如何实现？
3. 缓存雪崩如何防范？

---

#### 💡 案例2：秒杀系统的完整设计

**问题描述**：
设计一个支持10万QPS的秒杀系统，商品库存100件，要求：防止超卖、保证公平性、抵抗恶意刷单。

**答案**：

**架构设计**（Go实现）：

```go
package seckill

import (
	"context"
	"errors"
	"fmt"
	"time"
	
	"github.com/go-redis/redis/v8"
)

// SeckillService 秒杀服务
type SeckillService struct {
	rdb         *redis.Client
	orderSvc    OrderService
	inventorySvc InventoryService
	mq          MessageQueue
}

// CreateSeckill 创建秒杀活动
func (s *SeckillService) CreateSeckill(ctx context.Context, 
	seckill *Seckill) error {
	
	// 1. 创建秒杀活动
	if err := s.seckillRepo.Create(ctx, seckill); err != nil {
		return err
	}
	
	// 2. 库存预热到Redis
	stockKey := fmt.Sprintf("seckill:stock:%d", seckill.SeckillID)
	s.rdb.Set(ctx, stockKey, seckill.Stock, 0)
	
	// 3. 创建商品详情页缓存
	s.preWarmCache(ctx, seckill.ProductID)
	
	return nil
}

// Seckill 秒杀下单
func (s *SeckillService) Seckill(ctx context.Context, 
	userID int64, seckillID int64) error {
	
	// 1. 限流（单用户限流+全局限流）
	if err := s.checkRateLimit(ctx, userID, seckillID); err != nil {
		return err
	}
	
	// 2. 风控检查
	if err := s.riskCheck(ctx, userID); err != nil {
		return err
	}
	
	// 3. 扣减Redis库存（Lua脚本保证原子性）
	stock, err := s.deductStock(ctx, seckillID)
	if err != nil {
		return err
	}
	
	if stock < 0 {
		return errors.New("商品已抢光")
	}
	
	// 4. 发送消息到MQ异步创建订单
	msg := SeckillOrderMessage{
		UserID:    userID,
		SeckillID: seckillID,
		Timestamp: time.Now(),
	}
	
	if err := s.mq.Publish("seckill.order.create", msg); err != nil {
		// MQ发送失败，回补库存
		s.increaseStock(ctx, seckillID)
		return err
	}
	
	return nil
}

// 扣减库存（Lua脚本）
func (s *SeckillService) deductStock(ctx context.Context, 
	seckillID int64) (int64, error) {
	
	stockKey := fmt.Sprintf("seckill:stock:%d", seckillID)
	
	// Lua脚本保证原子性
	script := `
		local stock = redis.call('GET', KEYS[1])
		if tonumber(stock) <= 0 then
			return -1
		end
		redis.call('DECR', KEYS[1])
		return stock - 1
	`
	
	result, err := s.rdb.Eval(ctx, script, []string{stockKey}).Result()
	if err != nil {
		return 0, err
	}
	
	return result.(int64), nil
}

// 限流（令牌桶）
func (s *SeckillService) checkRateLimit(ctx context.Context, 
	userID int64, seckillID int64) error {
	
	// 单用户限流：1秒内最多1次
	userKey := fmt.Sprintf("seckill:ratelimit:user:%d:%d", userID, seckillID)
	exists, _ := s.rdb.Exists(ctx, userKey).Result()
	if exists > 0 {
		return errors.New("操作太频繁")
	}
	
	s.rdb.SetEX(ctx, userKey, 1, time.Second)
	
	// 全局限流：令牌桶（10万QPS）
	globalKey := fmt.Sprintf("seckill:ratelimit:global:%d", seckillID)
	token, _ := s.rdb.Incr(ctx, globalKey).Result()
	if token == 1 {
		s.rdb.Expire(ctx, globalKey, time.Second)
	}
	
	if token > 100000 {
		return errors.New("系统繁忙，请稍后再试")
	}
	
	return nil
}

// 异步创建订单（消费MQ）
func (s *SeckillService) CreateOrderAsync() {
	s.mq.Subscribe("seckill.order.create", func(msg *SeckillOrderMessage) {
		ctx := context.Background()
		
		// 1. 防重（幂等性）
		orderKey := fmt.Sprintf("seckill:order:%d:%d", msg.UserID, msg.SeckillID)
		exists, _ := s.rdb.Exists(ctx, orderKey).Result()
		if exists > 0 {
			log.Warnf("用户%d已抢购过", msg.UserID)
			return
		}
		
		// 2. 创建订单
		order, err := s.orderSvc.CreateSeckillOrder(ctx, msg.UserID, msg.SeckillID)
		if err != nil {
			log.Errorf("创建订单失败: %v", err)
			// 回补库存
			s.increaseStock(ctx, msg.SeckillID)
			return
		}
		
		// 3. 标记已抢购
		s.rdb.SetEX(ctx, orderKey, order.OrderID, 24*time.Hour)
		
		// 4. 通知用户
		s.notifySvc.Send(ctx, msg.UserID, "秒杀成功，请尽快支付")
	})
}

// 定时任务：取消未支付订单
func (s *SeckillService) CancelUnpaidOrders() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		// 查询15分钟前创建的未支付秒杀订单
		ctx := context.Background()
		cutoffTime := time.Now().Add(-15 * time.Minute)
		
		orders, _ := s.orderRepo.FindUnpaidSeckillOrders(ctx, cutoffTime)
		
		for _, order := range orders {
			// 取消订单
			s.orderSvc.Cancel(ctx, order.OrderID)
			
			// 回补库存
			s.increaseStock(ctx, order.SeckillID)
		}
	}
}
```

**架构要点**：
1. **前端限流**：按钮置灰、验证码、排队页
2. **网关限流**：Nginx限流（10万QPS）
3. **Redis预扣库存**：Lua脚本原子操作
4. **异步下单**：MQ削峰，提升吞吐
5. **超时取消**：15分钟未支付自动取消+回补库存

**延伸思考**：
1. 秒杀如何防止黄牛？
2. 分布式锁如何选型（Redis vs Etcd）？

---

#### 🔧 案例3：订单履约的全链路监控

**问题描述**：
订单从创建到签收，涉及多个服务（订单、库存、物流、支付）。如何设计全链路监控，快速定位问题？

**答案**：

**推荐方案**：分布式追踪（OpenTelemetry + Jaeger）

```go
package tracing

import (
	"context"
	
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OrderFulfillmentService 订单履约服务
type OrderFulfillmentService struct {
	tracer trace.Tracer
}

// FulfillOrder 履约订单（带追踪）
func (s *OrderFulfillmentService) FulfillOrder(ctx context.Context, 
	orderID int64) error {
	
	// 创建根Span
	ctx, span := s.tracer.Start(ctx, "FulfillOrder",
		trace.WithAttributes(
			attribute.Int64("order.id", orderID),
		),
	)
	defer span.End()
	
	// 步骤1：扣减库存
	if err := s.deductInventory(ctx, orderID); err != nil {
		span.RecordError(err)
		return err
	}
	
	// 步骤2：创建拣货单
	if err := s.createPickingOrder(ctx, orderID); err != nil {
		span.RecordError(err)
		return err
	}
	
	// 步骤3：创建物流运单
	if err := s.createShipment(ctx, orderID); err != nil {
		span.RecordError(err)
		return err
	}
	
	span.SetAttributes(attribute.String("status", "success"))
	return nil
}

// deductInventory 扣减库存（子Span）
func (s *OrderFulfillmentService) deductInventory(ctx context.Context, 
	orderID int64) error {
	
	ctx, span := s.tracer.Start(ctx, "DeductInventory")
	defer span.End()
	
	// 调用库存服务
	start := time.Now()
	err := s.inventorySvc.Deduct(ctx, orderID)
	duration := time.Since(start)
	
	span.SetAttributes(
		attribute.String("service", "inventory"),
		attribute.Int64("duration_ms", duration.Milliseconds()),
	)
	
	if err != nil {
		span.RecordError(err)
		return err
	}
	
	return nil
}
```

**监控指标**：
```go
// RED指标（请求速率、错误率、耗时）
type Metrics struct {
	// Rate: 请求速率
	OrderCreateRate   *prometheus.CounterVec
	
	// Error: 错误率
	OrderCreateErrors *prometheus.CounterVec
	
	// Duration: 耗时分布
	OrderCreateDuration *prometheus.HistogramVec
}

// 记录指标
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	start := time.Now()
	
	// 请求计数
	s.metrics.OrderCreateRate.WithLabelValues("order_service").Inc()
	
	// 执行业务逻辑
	order, err := s.doCreateOrder(ctx, req)
	
	// 记录耗时
	duration := time.Since(start).Seconds()
	s.metrics.OrderCreateDuration.WithLabelValues("order_service").Observe(duration)
	
	// 记录错误
	if err != nil {
		s.metrics.OrderCreateErrors.WithLabelValues("order_service", "error").Inc()
	} else {
		s.metrics.OrderCreateErrors.WithLabelValues("order_service", "success").Inc()
	}
	
	return order, err
}
```

**延伸思考**：
1. 如何设计告警规则（P95延迟>500ms告警）？
2. 如何追踪跨语言调用链（Go → Java → Python）？

---

#### 📊 案例4：大促准备的全链路压测

**问题描述**：
618大促前，需要对整个系统进行压测，验证系统能否支撑预期流量。如何设计全链路压测方案？

**答案**：

**压测方案**（Go实现）：

```go
package loadtest

import (
	"context"
	"sync"
	"time"
)

// LoadTester 压测工具
type LoadTester struct {
	targetQPS int
	duration  time.Duration
	workers   int
}

// Run 执行压测
func (lt *LoadTester) Run(ctx context.Context, 
	testFunc func(context.Context) error) *LoadTestResult {
	
	result := &LoadTestResult{
		StartTime: time.Now(),
	}
	
	// 计算每个worker的QPS
	qpsPerWorker := lt.targetQPS / lt.workers
	interval := time.Second / time.Duration(qpsPerWorker)
	
	var wg sync.WaitGroup
	resultChan := make(chan *RequestResult, lt.targetQPS*int(lt.duration.Seconds()))
	
	// 启动workers
	for i := 0; i < lt.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			
			timeout := time.After(lt.duration)
			
			for {
				select {
				case <-ticker.C:
					// 执行请求
					reqResult := lt.executeRequest(ctx, testFunc)
					resultChan <- reqResult
					
				case <-timeout:
					return
				}
			}
		}()
	}
	
	// 等待所有workers完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// 收集结果
	for reqResult := range resultChan {
		result.TotalRequests++
		
		if reqResult.Success {
			result.SuccessCount++
			result.TotalLatency += reqResult.Latency
		} else {
			result.FailureCount++
		}
		
		// 记录延迟分布
		result.LatencyDistribution = append(result.LatencyDistribution, reqResult.Latency)
	}
	
	result.EndTime = time.Now()
	result.Calculate()
	
	return result
}

// executeRequest 执行单次请求
func (lt *LoadTester) executeRequest(ctx context.Context, 
	testFunc func(context.Context) error) *RequestResult {
	
	result := &RequestResult{
		StartTime: time.Now(),
	}
	
	err := testFunc(ctx)
	
	result.EndTime = time.Now()
	result.Latency = result.EndTime.Sub(result.StartTime)
	result.Success = (err == nil)
	
	return result
}

// LoadTestResult 压测结果
type LoadTestResult struct {
	StartTime           time.Time
	EndTime             time.Time
	TotalRequests       int
	SuccessCount        int
	FailureCount        int
	TotalLatency        time.Duration
	LatencyDistribution []time.Duration
	
	// 计算指标
	QPS         float64
	AvgLatency  time.Duration
	P50Latency  time.Duration
	P95Latency  time.Duration
	P99Latency  time.Duration
	SuccessRate float64
}

// Calculate 计算统计指标
func (r *LoadTestResult) Calculate() {
	duration := r.EndTime.Sub(r.StartTime).Seconds()
	r.QPS = float64(r.TotalRequests) / duration
	
	if r.SuccessCount > 0 {
		r.AvgLatency = r.TotalLatency / time.Duration(r.SuccessCount)
	}
	
	r.SuccessRate = float64(r.SuccessCount) / float64(r.TotalRequests) * 100
	
	// 计算P50/P95/P99
	sort.Slice(r.LatencyDistribution, func(i, j int) bool {
		return r.LatencyDistribution[i] < r.LatencyDistribution[j]
	})
	
	if len(r.LatencyDistribution) > 0 {
		r.P50Latency = r.LatencyDistribution[len(r.LatencyDistribution)*50/100]
		r.P95Latency = r.LatencyDistribution[len(r.LatencyDistribution)*95/100]
		r.P99Latency = r.LatencyDistribution[len(r.LatencyDistribution)*99/100]
	}
}

// 使用示例
func TestOrderCreate() {
	tester := &LoadTester{
		targetQPS: 10000,  // 目标1万QPS
		duration:  5 * time.Minute,
		workers:   100,
	}
	
	result := tester.Run(context.Background(), func(ctx context.Context) error {
		// 模拟下单
		return orderSvc.CreateOrder(ctx, &CreateOrderRequest{
			UserID: randomUserID(),
			Items:  randomItems(),
		})
	})
	
	// 输出结果
	fmt.Printf("QPS: %.2f\n", result.QPS)
	fmt.Printf("成功率: %.2f%%\n", result.SuccessRate)
	fmt.Printf("平均延迟: %v\n", result.AvgLatency)
	fmt.Printf("P95延迟: %v\n", result.P95Latency)
	fmt.Printf("P99延迟: %v\n", result.P99Latency)
}
```

**压测环境隔离**：
```text
生产环境：不能压测
预发环境：配置与生产一致，数据隔离
压测流量标记：HTTP Header: X-Load-Test: true
```

**延伸思考**：
1. 如何设计压测数据构造？
2. 压测导致的脏数据如何清理？

---

#### 🔧 案例5：跨境电商的多币种结算

**问题描述**：
跨境电商平台支持美元、欧元、人民币等多币种。如何设计多币种的定价、支付、结算系统？

**答案**：

**推荐方案**（Go实现）：

```go
package currency

import (
	"context"
	"time"
	
	"github.com/shopspring/decimal"
)

// Currency 币种
type Currency string

const (
	CNY Currency = "CNY"  // 人民币
	USD Currency = "USD"  // 美元
	EUR Currency = "EUR"  // 欧元
)

// ExchangeRateService 汇率服务
type ExchangeRateService struct {
	rdb   *redis.Client
	repo  ExchangeRateRepository
}

// GetExchangeRate 获取汇率
func (s *ExchangeRateService) GetExchangeRate(ctx context.Context, 
	from, to Currency) (decimal.Decimal, error) {
	
	if from == to {
		return decimal.NewFromInt(1), nil
	}
	
	// 从缓存读取
	cacheKey := fmt.Sprintf("exchange_rate:%s:%s", from, to)
	rateStr, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		rate, _ := decimal.NewFromString(rateStr)
		return rate, nil
	}
	
	// 从数据库读取
	rate, err := s.repo.FindLatest(ctx, from, to)
	if err != nil {
		return decimal.Zero, err
	}
	
	// 缓存1小时
	s.rdb.SetEX(ctx, cacheKey, rate.Rate.String(), time.Hour)
	
	return rate.Rate, nil
}

// Convert 货币转换
func (s *ExchangeRateService) Convert(ctx context.Context, 
	amount decimal.Decimal, from, to Currency) (decimal.Decimal, error) {
	
	rate, err := s.GetExchangeRate(ctx, from, to)
	if err != nil {
		return decimal.Zero, err
	}
	
	return amount.Mul(rate), nil
}

// ProductPricing 商品多币种定价
type ProductPricing struct {
	ProductID int64
	Prices    map[Currency]decimal.Decimal
}

// GetPrice 获取指定币种的价格
func (p *ProductPricing) GetPrice(currency Currency) (decimal.Decimal, error) {
	if price, exists := p.Prices[currency]; exists {
		return price, nil
	}
	
	return decimal.Zero, errors.New("该币种暂不支持")
}

// OrderService 订单服务（多币种）
type OrderService struct {
	exchangeRateSvc *ExchangeRateService
}

// CreateOrder 创建订单（多币种）
func (s *OrderService) CreateOrder(ctx context.Context, 
	req *CreateOrderRequest) (*Order, error) {
	
	// 1. 计算订单金额（用户选择的币种）
	var totalAmount decimal.Decimal
	for _, item := range req.Items {
		price := item.Product.GetPrice(req.Currency)
		totalAmount = totalAmount.Add(price.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}
	
	// 2. 转换为平台基准币种（CNY）
	baseCurrencyAmount, err := s.exchangeRateSvc.Convert(ctx, 
		totalAmount, req.Currency, CNY)
	if err != nil {
		return nil, err
	}
	
	// 3. 创建订单
	order := &Order{
		OrderID:            generateOrderID(),
		UserID:             req.UserID,
		Currency:           req.Currency,        // 显示币种
		TotalAmount:        totalAmount,         // 显示金额
		BaseCurrency:       CNY,                 // 基准币种
		BaseCurrencyAmount: baseCurrencyAmount,  // 基准金额
		ExchangeRate:       s.getExchangeRate(ctx, req.Currency, CNY),
		CreatedAt:          time.Now(),
	}
	
	return order, s.orderRepo.Create(ctx, order)
}

// 汇率快照（订单创建时记录汇率）
func (s *OrderService) getExchangeRate(ctx context.Context, 
	from, to Currency) decimal.Decimal {
	
	rate, _ := s.exchangeRateSvc.GetExchangeRate(ctx, from, to)
	return rate
}
```

**结算处理**：
```go
// SettlementService 结算服务（多币种）
type SettlementService struct {
	exchangeRateSvc *ExchangeRateService
}

// Settle 结算
func (s *SettlementService) Settle(ctx context.Context, 
	merchantID int64, date time.Time) error {
	
	// 1. 查询待结算订单
	orders, _ := s.orderRepo.FindPendingSettlement(ctx, merchantID, date)
	
	// 2. 按币种分组汇总
	settlementByCurrency := make(map[Currency]decimal.Decimal)
	for _, order := range orders {
		current := settlementByCurrency[order.Currency]
		settlementByCurrency[order.Currency] = current.Add(order.MerchantAmount)
	}
	
	// 3. 转换为商家收款币种并结算
	merchantCurrency := s.getMerchantCurrency(ctx, merchantID)
	
	for currency, amount := range settlementByCurrency {
		// 转换币种
		settleAmount, _ := s.exchangeRateSvc.Convert(ctx, 
			amount, currency, merchantCurrency)
		
		// 调用支付渠道转账
		s.paymentSvc.Transfer(ctx, merchantID, settleAmount, merchantCurrency)
	}
	
	return nil
}
```

**延伸思考**：
1. 汇率波动如何处理（订单创建时汇率 vs 支付时汇率）？
2. 跨境支付的关税如何计算？

---

#### 💡 案例6：电商搜索的智能排序

**问题描述**：
用户搜索"手机"，返回1000个结果。如何设计排序算法，让用户最可能购买的商品排在前面？

**答案**：

**推荐方案**：多因子排序模型

```go
package search

import (
	"context"
	"math"
	
	"github.com/shopspring/decimal"
)

// SearchRankingService 搜索排序服务
type SearchRankingService struct {
	userProfileSvc UserProfileService
}

// RankProducts 对商品排序
func (s *SearchRankingService) RankProducts(ctx context.Context, 
	userID int64, products []*Product) []*ScoredProduct {
	
	scoredProducts := make([]*ScoredProduct, 0, len(products))
	
	for _, product := range products {
		score := s.calculateScore(ctx, userID, product)
		scoredProducts = append(scoredProducts, &ScoredProduct{
			Product: product,
			Score:   score,
		})
	}
	
	// 按分数降序排序
	sort.Slice(scoredProducts, func(i, j int) bool {
		return scoredProducts[i].Score > scoredProducts[j].Score
	})
	
	return scoredProducts
}

// calculateScore 计算商品综合分数
func (s *SearchRankingService) calculateScore(ctx context.Context, 
	userID int64, product *Product) float64 {
	
	// 多因子加权求和
	score := 0.0
	
	// 1. 文本相关性（权重20%）
	textRelevance := s.calculateTextRelevance(product)
	score += textRelevance * 0.2
	
	// 2. 销量（权重15%）
	salesScore := s.normalizeSales(product.SalesCount)
	score += salesScore * 0.15
	
	// 3. 好评率（权重10%）
	ratingScore := product.Rating / 5.0
	score += ratingScore * 0.1
	
	// 4. 价格（权重10%）
	priceScore := s.calculatePriceScore(product.Price)
	score += priceScore * 0.1
	
	// 5. 个性化（权重30%）
	personalScore := s.calculatePersonalScore(ctx, userID, product)
	score += personalScore * 0.3
	
	// 6. 时效性（权重5%）
	timeScore := s.calculateTimeScore(product.CreatedAt)
	score += timeScore * 0.05
	
	// 7. 商家质量（权重10%）
	merchantScore := s.calculateMerchantScore(product.MerchantID)
	score += merchantScore * 0.1
	
	return score
}

// 个性化分数（基于用户画像）
func (s *SearchRankingService) calculatePersonalScore(ctx context.Context, 
	userID int64, product *Product) float64 {
	
	profile := s.userProfileSvc.GetProfile(ctx, userID)
	
	score := 0.0
	
	// 1. 品牌偏好
	if contains(profile.FavoriteBrands, product.Brand) {
		score += 0.3
	}
	
	// 2. 类目偏好
	if contains(profile.FavoriteCategories, product.CategoryID) {
		score += 0.3
	}
	
	// 3. 价格区间偏好
	if product.Price.GreaterThanOrEqual(profile.MinPrice) &&
		product.Price.LessThanOrEqual(profile.MaxPrice) {
		score += 0.2
	}
	
	// 4. 历史浏览相似度
	similarity := s.calculateSimilarity(product, profile.ViewedProducts)
	score += similarity * 0.2
	
	return score
}

// 销量归一化（对数变换）
func (s *SearchRankingService) normalizeSales(salesCount int64) float64 {
	if salesCount == 0 {
		return 0
	}
	
	// 对数变换平滑销量差异
	return math.Log10(float64(salesCount)+1) / math.Log10(1000000)
}
```

**机器学习排序**：
```go
// LearningToRank 学习排序模型
type LearningToRank struct {
	model MLModel
}

// Rank 使用模型排序
func (ltr *LearningToRank) Rank(ctx context.Context, 
	userID int64, products []*Product) []*ScoredProduct {
	
	scoredProducts := make([]*ScoredProduct, 0)
	
	for _, product := range products {
		// 提取特征
		features := ltr.extractFeatures(ctx, userID, product)
		
		// 模型预测分数
		score := ltr.model.Predict(features)
		
		scoredProducts = append(scoredProducts, &ScoredProduct{
			Product: product,
			Score:   score,
		})
	}
	
	// 排序
	sort.Slice(scoredProducts, func(i, j int) bool {
		return scoredProducts[i].Score > scoredProducts[j].Score
	})
	
	return scoredProducts
}

// 特征提取
func (ltr *LearningToRank) extractFeatures(ctx context.Context, 
	userID int64, product *Product) []float64 {
	
	return []float64{
		float64(product.SalesCount),
		product.Rating,
		product.Price.InexactFloat64(),
		float64(product.ReviewCount),
		// ... 更多特征
	}
}
```

**延伸思考**：
1. 如何设计AB测试验证排序效果？
2. 如何平衡新品曝光和热销商品？

---

#### 🚀 案例7：异常流量的应急处理

**问题描述**：
凌晨2点，监控告警：订单服务QPS突增10倍，响应时间飙升至5秒，疑似遭受攻击。如何快速定位和处理？

**答案**：

**应急响应流程**（Go实现）：

```go
package emergency

import (
	"context"
	"time"
)

// EmergencyHandler 应急处理器
type EmergencyHandler struct {
	rateLimiter *RateLimiter
	ipBlacklist *IPBlacklist
	alertSvc    AlertService
}

// HandleAbnormalTraffic 处理异常流量
func (h *EmergencyHandler) HandleAbnormalTraffic(ctx context.Context) error {
	// 第1步：分析流量特征
	analysis := h.analyzeTraffic(ctx)
	
	// 第2步：判断攻击类型
	attackType := h.identifyAttackType(analysis)
	
	// 第3步：执行防御措施
	switch attackType {
	case AttackTypeDDoS:
		return h.handleDDoS(ctx, analysis)
	case AttackTypeCrawler:
		return h.handleCrawler(ctx, analysis)
	case AttackTypeBrushOrder:
		return h.handleBrushOrder(ctx, analysis)
	default:
		return h.handleUnknown(ctx, analysis)
	}
}

// analyzeTraffic 分析流量
func (h *EmergencyHandler) analyzeTraffic(ctx context.Context) *TrafficAnalysis {
	now := time.Now()
	last5Min := now.Add(-5 * time.Minute)
	
	// 查询最近5分钟的请求日志
	logs := h.logSvc.Query(ctx, last5Min, now)
	
	analysis := &TrafficAnalysis{
		TotalRequests: len(logs),
		IPDistribution: make(map[string]int),
		UADistribution: make(map[string]int),
		URLDistribution: make(map[string]int),
	}
	
	for _, log := range logs {
		// IP分布
		analysis.IPDistribution[log.IP]++
		
		// User-Agent分布
		analysis.UADistribution[log.UserAgent]++
		
		// URL分布
		analysis.URLDistribution[log.URL]++
	}
	
	// 识别异常IP（单IP请求占比>10%）
	for ip, count := range analysis.IPDistribution {
		ratio := float64(count) / float64(analysis.TotalRequests)
		if ratio > 0.1 {
			analysis.AbnormalIPs = append(analysis.AbnormalIPs, ip)
		}
	}
	
	return analysis
}

// handleDDoS 处理DDoS攻击
func (h *EmergencyHandler) handleDDoS(ctx context.Context, 
	analysis *TrafficAnalysis) error {
	
	// 1. 立即限流（全局QPS降低到正常值的50%）
	h.rateLimiter.SetGlobalLimit(10000)
	
	// 2. 封禁异常IP
	for _, ip := range analysis.AbnormalIPs {
		h.ipBlacklist.Add(ip, 1*time.Hour)
		log.Warnf("封禁IP: %s", ip)
	}
	
	// 3. 启用验证码
	h.enableCaptcha()
	
	// 4. 通知运维
	h.alertSvc.Send("紧急：疑似DDoS攻击，已自动防御")
	
	return nil
}

// handleBrushOrder 处理刷单攻击
func (h *EmergencyHandler) handleBrushOrder(ctx context.Context, 
	analysis *TrafficAnalysis) error {
	
	// 1. 识别刷单用户
	suspiciousUsers := h.identifySuspiciousUsers(ctx)
	
	// 2. 限制下单频率
	for _, userID := range suspiciousUsers {
		h.rateLimiter.SetUserLimit(userID, 1) // 1分钟1单
		log.Warnf("限制用户%d下单频率", userID)
	}
	
	// 3. 启用风控策略（大额订单人工审核）
	h.enableManualReview()
	
	return nil
}

// 降级策略
func (h *EmergencyHandler) Degrade(ctx context.Context) error {
	// Level 1：关闭非核心功能
	h.disableRecommendation()  // 关闭推荐
	h.disableSearch()          // 关闭搜索
	
	// Level 2：只读模式（禁止下单）
	h.enableReadOnlyMode()
	
	// Level 3：返回静态页面
	h.enableStaticMode()
	
	return nil
}
```

**监控告警**：
```go
// AlertRule 告警规则
type AlertRule struct {
	Name      string
	Metric    string
	Threshold float64
	Duration  time.Duration
	Severity  string  // P0/P1/P2/P3
}

// 告警规则示例
var alertRules = []AlertRule{
	{
		Name:      "订单QPS异常",
		Metric:    "order_create_qps",
		Threshold: 10000,  // QPS超过1万
		Duration:  1 * time.Minute,
		Severity:  "P0",
	},
	{
		Name:      "订单延迟异常",
		Metric:    "order_create_p99_latency",
		Threshold: 1000,  // P99超过1秒
		Duration:  5 * time.Minute,
		Severity:  "P1",
	},
}
```

**延伸思考**：
1. 如何设计自动化的应急响应系统？
2. 如何平衡防御和用户体验（误封正常用户）？

---

#### 🔧 案例8：订单的柔性事务设计

**问题描述**：
订单创建涉及多个服务（扣库存、扣优惠券、扣积分）。如何设计柔性事务，保证最终一致性？

**答案**：

**推荐方案**：本地消息表 + 定时补偿

```go
package transaction

import (
	"context"
	"time"
)

// LocalMessageTable 本地消息表
type LocalMessage struct {
	MessageID   string
	BizType     string          // 业务类型
	BizID       string          // 业务ID
	Content     string          // 消息内容
	Status      MessageStatus   // 待发送/已发送/发送失败
	RetryCount  int
	NextRetryAt time.Time
	CreatedAt   time.Time
}

// OrderService 订单服务（柔性事务）
type OrderService struct {
	orderRepo   OrderRepository
	messageSvc  LocalMessageService
	eventBus    EventBus
}

// CreateOrder 创建订单（本地消息表）
func (s *OrderService) CreateOrder(ctx context.Context, 
	req *CreateOrderRequest) (*Order, error) {
	
	// 开启数据库事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	
	// 1. 创建订单
	order := &Order{
		OrderID:   generateOrderID(),
		UserID:    req.UserID,
		Items:     req.Items,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	
	if err := s.orderRepo.CreateWithTx(ctx, tx, order); err != nil {
		return nil, err
	}
	
	// 2. 写入本地消息表（同一个事务）
	messages := []LocalMessage{
		{
			MessageID: generateMessageID(),
			BizType:   "ORDER_CREATED",
			BizID:     fmt.Sprintf("%d", order.OrderID),
			Content:   s.serializeOrderEvent(order),
			Status:    MessageStatusPending,
			CreatedAt: time.Now(),
		},
	}
	
	for _, msg := range messages {
		if err := s.messageSvc.CreateWithTx(ctx, tx, &msg); err != nil {
			return nil, err
		}
	}
	
	// 3. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	
	// 4. 异步发送消息（事务外）
	go s.publishPendingMessages(context.Background())
	
	return order, nil
}

// publishPendingMessages 发布待发送消息
func (s *OrderService) publishPendingMessages(ctx context.Context) {
	// 查询待发送消息
	messages, err := s.messageSvc.FindPending(ctx, 100)
	if err != nil {
		return
	}
	
	for _, msg := range messages {
		// 发送到消息队列
		err := s.eventBus.Publish(msg.BizType, msg.Content)
		
		if err == nil {
			// 发送成功，更新状态
			msg.Status = MessageStatusSent
			s.messageSvc.Update(ctx, &msg)
		} else {
			// 发送失败，记录重试
			msg.RetryCount++
			msg.NextRetryAt = time.Now().Add(time.Duration(msg.RetryCount) * time.Minute)
			s.messageSvc.Update(ctx, &msg)
		}
	}
}

// RetryFailedMessages 定时重试失败消息
func (s *OrderService) RetryFailedMessages() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		ctx := context.Background()
		
		// 查询需要重试的消息
		messages, _ := s.messageSvc.FindRetryable(ctx, time.Now())
		
		for _, msg := range messages {
			if msg.RetryCount >= 5 {
				// 超过最大重试次数，转人工处理
				msg.Status = MessageStatusFailed
				s.messageSvc.Update(ctx, &msg)
				s.createManualTask(ctx, &msg)
				continue
			}
			
			// 重试发送
			s.publishMessage(ctx, &msg)
		}
	}
}

// 下游服务消费消息
type InventoryConsumer struct {
	inventorySvc InventoryService
}

func (c *InventoryConsumer) Consume(ctx context.Context, msg *OrderCreatedEvent) error {
	// 幂等性检查
	if c.isProcessed(ctx, msg.OrderID) {
		log.Infof("订单%d已处理，跳过", msg.OrderID)
		return nil
	}
	
	// 扣减库存
	err := c.inventorySvc.Deduct(ctx, msg.Items)
	if err != nil {
		// 扣减失败，发送补偿消息
		c.publishCompensation(ctx, msg.OrderID)
		return err
	}
	
	// 标记已处理
	c.markAsProcessed(ctx, msg.OrderID)
	
	return nil
}
```

**延伸思考**：
1. 本地消息表 vs Saga vs TCC如何选择？
2. 消息发送失败如何保证最终一致性？

---

#### 📊 案例9：用户画像系统的设计

**问题描述**：
为了实现个性化推荐，需要构建用户画像（年龄、性别、消费能力、兴趣偏好）。如何设计用户画像系统？

**答案**：

**推荐方案**：实时+离线双层架构

```go
package userprofile

import (
	"context"
	"time"
)

// UserProfile 用户画像
type UserProfile struct {
	UserID int64
	
	// 基础信息
	Age    int
	Gender string
	City   string
	
	// 消费画像
	AvgOrderAmount    decimal.Decimal  // 客单价
	TotalOrderCount   int              // 订单数
	ConsumptionLevel  string           // 消费能力：高/中/低
	
	// 兴趣画像
	FavoriteCategories []int64         // 偏好类目
	FavoriteBrands     []string        // 偏好品牌
	PriceRange         PriceRange      // 价格区间
	
	// 行为特征
	ActiveTime         []int           // 活跃时段
	ShoppingFrequency  string          // 购物频次
	LastPurchaseTime   time.Time
	
	UpdatedAt time.Time
}

// UserProfileService 用户画像服务
type UserProfileService struct {
	repo        UserProfileRepository
	rdb         *redis.Client
	kafkaWriter *kafka.Writer
}

// GetProfile 获取用户画像
func (s *UserProfileService) GetProfile(ctx context.Context, 
	userID int64) (*UserProfile, error) {
	
	// 从缓存读取
	cacheKey := fmt.Sprintf("user:profile:%d", userID)
	profileJSON, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		profile := &UserProfile{}
		json.Unmarshal([]byte(profileJSON), profile)
		return profile, nil
	}
	
	// 从数据库读取
	profile, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// 缓存
	profileJSON, _ = json.Marshal(profile)
	s.rdb.SetEX(ctx, cacheKey, profileJSON, 6*time.Hour)
	
	return profile, nil
}

// UpdateProfileRealtime 实时更新画像
func (s *UserProfileService) UpdateProfileRealtime(ctx context.Context, 
	event *UserBehaviorEvent) error {
	
	// 将行为事件写入Kafka
	return s.kafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", event.UserID)),
		Value: s.serializeEvent(event),
	})
}

// Flink实时计算（伪代码）
/*
用户行为流 → Flink → 实时画像

Flink Job:
1. 消费Kafka用户行为流
2. 计算实时指标（浏览、加购、下单）
3. 更新Redis画像缓存
4. 每小时写入HBase
*/

// 离线计算（每日凌晨执行）
func (s *UserProfileService) BatchUpdateProfiles() error {
	ctx := context.Background()
	yesterday := time.Now().AddDate(0, 0, -1)
	
	// 1. 查询昨天的用户行为数据
	behaviors, _ := s.behaviorRepo.FindByDate(ctx, yesterday)
	
	// 2. 聚合计算
	profileUpdates := s.aggregateBehaviors(behaviors)
	
	// 3. 批量更新画像
	for _, update := range profileUpdates {
		s.repo.Update(ctx, update)
		
		// 清除缓存
		cacheKey := fmt.Sprintf("user:profile:%d", update.UserID)
		s.rdb.Del(ctx, cacheKey)
	}
	
	return nil
}

// 消费能力分层
func (s *UserProfileService) calculateConsumptionLevel(
	avgOrderAmount decimal.Decimal, totalOrderCount int) string {
	
	if avgOrderAmount.GreaterThanOrEqual(decimal.NewFromInt(500)) && 
		totalOrderCount >= 10 {
		return "高"
	} else if avgOrderAmount.GreaterThanOrEqual(decimal.NewFromInt(200)) && 
		totalOrderCount >= 3 {
		return "中"
	} else {
		return "低"
	}
}
```

**延伸思考**：
1. 如何保护用户隐私（GDPR合规）？
2. 画像准确性如何评估？

---

#### 💡 案例10：大促后的系统复盘

**问题描述**：
618大促结束后，需要对系统表现进行复盘。如何设计复盘报告，总结经验和改进点？

**答案**：

**复盘维度**：

1. **业务指标**：
   - GMV：50亿
   - 订单量：1000万
   - 转化率：3.5%
   - 客单价：500元

2. **技术指标**：
   - 峰值QPS：50万
   - 平均响应时间：200ms
   - P99响应时间：800ms
   - 可用性：99.95%

3. **故障复盘**：
   - 23:00-23:15 订单服务QPS突增导致响应变慢
   - 根因：数据库连接池不足
   - 影响：15分钟内订单延迟，影响1000笔订单
   - 改进：增加连接池大小，增加熔断降级

4. **优化建议**：
   - 缓存命中率从90%提升到95%
   - 数据库慢查询优化（TOP 10）
   - 增加自动扩容策略

```go
// PromotionReview 大促复盘
type PromotionReview struct {
	PromotionName string
	StartTime     time.Time
	EndTime       time.Time
	
	// 业务指标
	GMV           decimal.Decimal
	OrderCount    int64
	ConversionRate float64
	
	// 技术指标
	PeakQPS       int64
	AvgLatency    time.Duration
	P99Latency    time.Duration
	Availability  float64
	
	// 故障列表
	Incidents     []*Incident
	
	// 改进建议
	Improvements  []string
}

// GenerateReviewReport 生成复盘报告
func GenerateReviewReport(ctx context.Context, 
	promotionID int64) (*PromotionReview, error) {
	
	// 1. 查询业务数据
	orders := queryOrders(ctx, promotionID)
	
	// 2. 查询监控数据
	metrics := queryMetrics(ctx, promotionID)
	
	// 3. 查询故障记录
	incidents := queryIncidents(ctx, promotionID)
	
	// 4. 生成报告
	review := &PromotionReview{
		PromotionName: "618大促",
		GMV:           calculateGMV(orders),
		OrderCount:    int64(len(orders)),
		PeakQPS:       metrics.PeakQPS,
		Incidents:     incidents,
	}
	
	return review, nil
}
```

**延伸思考**：
1. 如何设计大促演练（压测、故障演练）？
2. 如何量化技术优化的ROI？

---
