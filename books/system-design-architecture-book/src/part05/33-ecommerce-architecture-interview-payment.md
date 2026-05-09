# 35.3.4 支付系统题库

## 35.3.4 支付系统（10题）

#### 📊 题目1：支付系统的整体架构设计

**问题描述**：
电商平台需要支持多种支付方式（支付宝、微信、银行卡）。如何设计支付系统的整体架构？

**答案**：

**问题分析**：
支付系统的核心要素：
1. 多渠道接入（支付宝、微信、银联）
2. 支付安全性
3. 异步回调处理
4. 对账和资金安全

**架构设计**（Go实现）：

```go
package payment

import (
	"context"
	"time"
)

// PaymentChannel 支付渠道
type PaymentChannel string

const (
	ChannelAlipay PaymentChannel = "ALIPAY"
	ChannelWechat PaymentChannel = "WECHAT"
	ChannelUnion  PaymentChannel = "UNION"
)

// PaymentService 支付服务
type PaymentService struct {
	alipayAdapter  PaymentAdapter
	wechatAdapter  PaymentAdapter
	unionAdapter   PaymentAdapter
	paymentRepo    PaymentRepository
	orderSvc       OrderService
}

// PaymentAdapter 支付适配器接口（适配器模式）
type PaymentAdapter interface {
	// 创建支付
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	// 查询支付状态
	QueryPayment(ctx context.Context, paymentID string) (*PaymentStatus, error)
	// 申请退款
	Refund(ctx context.Context, req *RefundRequest) error
	// 验证回调签名
	VerifyCallback(callback *CallbackData) error
}

// Payment 支付单
type Payment struct {
	PaymentID       string
	OrderID         int64
	UserID          int64
	Channel         PaymentChannel
	Amount          decimal.Decimal
	Status          PaymentStatus
	ThirdPartyID    string  // 第三方支付单号
	CallbackData    string  // 回调原始数据
	CreatedAt       time.Time
	PaidAt          *time.Time
}

// CreatePayment 创建支付
func (s *PaymentService) CreatePayment(ctx context.Context, 
	orderID int64, channel PaymentChannel) (*PaymentResponse, error) {
	
	// 1. 查询订单
	order, err := s.orderSvc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	
	// 2. 校验订单状态
	if order.Status != OrderStatusPending {
		return nil, errors.New("订单状态不正确")
	}
	
	// 3. 创建支付单
	payment := &Payment{
		PaymentID: generatePaymentID(),
		OrderID:   orderID,
		UserID:    order.UserID,
		Channel:   channel,
		Amount:    order.TotalAmount,
		Status:    PaymentStatusPending,
		CreatedAt: time.Now(),
	}
	
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}
	
	// 4. 调用支付渠道
	adapter := s.getAdapter(channel)
	resp, err := adapter.CreatePayment(ctx, &PaymentRequest{
		OutTradeNo:  payment.PaymentID,
		Amount:      payment.Amount,
		Subject:     fmt.Sprintf("订单%d支付", orderID),
		NotifyURL:   "https://api.example.com/payment/callback",
		ReturnURL:   "https://www.example.com/order/success",
	})
	
	if err != nil {
		return nil, err
	}
	
	// 5. 保存第三方支付单号
	payment.ThirdPartyID = resp.TradeNo
	s.paymentRepo.Update(ctx, payment)
	
	return resp, nil
}

// 获取支付适配器
func (s *PaymentService) getAdapter(channel PaymentChannel) PaymentAdapter {
	switch channel {
	case ChannelAlipay:
		return s.alipayAdapter
	case ChannelWechat:
		return s.wechatAdapter
	case ChannelUnion:
		return s.unionAdapter
	default:
		return nil
	}
}

// HandleCallback 处理支付回调
func (s *PaymentService) HandleCallback(ctx context.Context, 
	channel PaymentChannel, callback *CallbackData) error {
	
	// 1. 验证签名
	adapter := s.getAdapter(channel)
	if err := adapter.VerifyCallback(callback); err != nil {
		return fmt.Errorf("签名验证失败: %w", err)
	}
	
	// 2. 查询支付单
	payment, err := s.paymentRepo.FindByID(ctx, callback.OutTradeNo)
	if err != nil {
		return err
	}
	
	// 3. 幂等性检查
	if payment.Status == PaymentStatusSuccess {
		return nil // 已处理，直接返回
	}
	
	// 4. 更新支付单状态
	payment.Status = PaymentStatusSuccess
	payment.PaidAt = &callback.PayTime
	payment.CallbackData = callback.RawData
	
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}
	
	// 5. 更新订单状态
	if err := s.orderSvc.MarkAsPaid(ctx, payment.OrderID); err != nil {
		// 支付成功但订单更新失败，记录补偿任务
		s.createCompensationTask(ctx, payment.PaymentID)
		return err
	}
	
	// 6. 发布支付成功事件
	s.eventBus.Publish(&PaymentSuccessEvent{
		OrderID:   payment.OrderID,
		PaymentID: payment.PaymentID,
		Amount:    payment.Amount,
	})
	
	return nil
}
```

**支付宝适配器示例**：
```go
type AlipayAdapter struct {
	client *alipay.Client
}

func (a *AlipayAdapter) CreatePayment(ctx context.Context, 
	req *PaymentRequest) (*PaymentResponse, error) {
	
	// 调用支付宝SDK
	payReq := alipay.TradeAppPay{
		OutTradeNo:  req.OutTradeNo,
		TotalAmount: req.Amount.String(),
		Subject:     req.Subject,
		NotifyURL:   req.NotifyURL,
	}
	
	orderStr, err := a.client.TradeAppPay(payReq)
	if err != nil {
		return nil, err
	}
	
	return &PaymentResponse{
		PayData: orderStr, // APP端拉起支付宝所需的参数
	}, nil
}

func (a *AlipayAdapter) VerifyCallback(callback *CallbackData) error {
	// 验证支付宝回调签名
	return a.client.VerifySign(callback.RawData)
}
```

**延伸思考**：
1. 支付系统如何实现高可用？
2. 支付渠道故障如何降级？
3. 支付回调丢失如何处理？

---

#### 🔧 题目2：支付回调的幂等性处理

**问题描述**：
支付回调可能重复发送（网络重试、第三方重推）。如何保证支付回调处理的幂等性？

**答案**：

**推荐方案**（Go实现）：

```go
// 幂等性处理
func (s *PaymentService) HandleCallbackIdempotent(ctx context.Context, 
	callback *CallbackData) error {
	
	paymentID := callback.OutTradeNo
	lockKey := fmt.Sprintf("payment:callback:lock:%s", paymentID)
	
	// 1. 获取分布式锁
	lock := redis.NewDistributedLock(s.rdb, lockKey)
	acquired, err := lock.TryLock(ctx, 30*time.Second)
	if err != nil {
		return err
	}
	if !acquired {
		// 其他请求正在处理，直接返回成功
		return nil
	}
	defer lock.Unlock(ctx)
	
	// 2. 查询支付单
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}
	
	// 3. 状态检查（幂等性）
	if payment.Status == PaymentStatusSuccess {
		log.Infof("支付单%s已处理，跳过", paymentID)
		return nil // 已成功，幂等返回
	}
	
	// 4. 使用数据库行锁+版本号
	affected, err := s.paymentRepo.UpdateStatusWithVersion(ctx, 
		paymentID, 
		PaymentStatusSuccess,
		payment.Version,
	)
	
	if err != nil {
		return err
	}
	
	if affected == 0 {
		// 版本号不匹配，说明已被其他请求处理
		log.Warnf("支付单%s已被处理，版本冲突", paymentID)
		return nil
	}
	
	// 5. 执行后续操作
	return s.postPaymentProcess(ctx, payment)
}

// 数据库更新（带版本号）
func (r *PaymentRepository) UpdateStatusWithVersion(ctx context.Context, 
	paymentID string, newStatus PaymentStatus, expectedVersion int) (int64, error) {
	
	query := `UPDATE payments 
	          SET status=?, version=version+1, updated_at=?
	          WHERE payment_id=? AND version=? AND status!=?`
	
	result, err := r.db.ExecContext(ctx, query, 
		newStatus, time.Now(), paymentID, expectedVersion, PaymentStatusSuccess)
	if err != nil {
		return 0, err
	}
	
	return result.RowsAffected()
}
```

**延伸思考**：
1. 如何设计支付回调的重试机制？
2. 回调处理失败如何人工介入？

---

#### 💡 题目3：支付的对账系统设计

**问题描述**：
每天需要与支付宝、微信对账，确保平台账和渠道账一致。如何设计支付对账系统？

**答案**：

**对账流程**（Go实现）：

```go
package reconciliation

import (
	"context"
	"time"
)

// ReconciliationService 对账服务
type ReconciliationService struct {
	paymentRepo PaymentRepository
	alipayClient *alipay.Client
	wechatClient *wechat.Client
}

// DailyReconciliation 每日对账
func (s *ReconciliationService) DailyReconciliation(ctx context.Context, date time.Time) error {
	// 1. 下载渠道对账单
	alipayBill, err := s.downloadAlipayBill(ctx, date)
	if err != nil {
		return err
	}
	
	wechatBill, err := s.downloadWechatBill(ctx, date)
	if err != nil {
		return err
	}
	
	// 2. 查询平台当日支付记录
	platformRecords, err := s.paymentRepo.FindByDate(ctx, date)
	if err != nil {
		return err
	}
	
	// 3. 三方对账
	diff := s.compare(platformRecords, alipayBill, wechatBill)
	
	// 4. 处理差异
	if err := s.handleDifferences(ctx, diff); err != nil {
		return err
	}
	
	// 5. 生成对账报告
	report := s.generateReport(diff)
	s.saveReport(ctx, report)
	
	return nil
}

// ReconciliationDiff 对账差异
type ReconciliationDiff struct {
	OnlyInPlatform   []*Payment  // 只在平台有
	OnlyInChannel    []*ChannelRecord  // 只在渠道有
	AmountMismatch   []*Mismatch  // 金额不一致
	StatusMismatch   []*Mismatch  // 状态不一致
}

// compare 比对数据
func (s *ReconciliationService) compare(platform []*Payment, 
	alipay, wechat []*ChannelRecord) *ReconciliationDiff {
	
	diff := &ReconciliationDiff{}
	
	// 构建平台数据map
	platformMap := make(map[string]*Payment)
	for _, p := range platform {
		platformMap[p.ThirdPartyID] = p
	}
	
	// 构建渠道数据map
	channelMap := make(map[string]*ChannelRecord)
	for _, c := range alipay {
		channelMap[c.TradeNo] = c
	}
	for _, c := range wechat {
		channelMap[c.TransactionID] = c
	}
	
	// 比对
	for tradeNo, channelRecord := range channelMap {
		platformRecord, exists := platformMap[tradeNo]
		
		if !exists {
			// 只在渠道有，平台无
			diff.OnlyInChannel = append(diff.OnlyInChannel, channelRecord)
		} else {
			// 金额比对
			if !platformRecord.Amount.Equal(channelRecord.Amount) {
				diff.AmountMismatch = append(diff.AmountMismatch, &Mismatch{
					TradeNo:        tradeNo,
					PlatformAmount: platformRecord.Amount,
					ChannelAmount:  channelRecord.Amount,
				})
			}
			
			// 状态比对
			if platformRecord.Status != channelRecord.Status {
				diff.StatusMismatch = append(diff.StatusMismatch, &Mismatch{
					TradeNo:       tradeNo,
					PlatformStatus: platformRecord.Status,
					ChannelStatus:  channelRecord.Status,
				})
			}
			
			delete(platformMap, tradeNo)
		}
	}
	
	// 只在平台有的
	for _, p := range platformMap {
		diff.OnlyInPlatform = append(diff.OnlyInPlatform, p)
	}
	
	return diff
}

// handleDifferences 处理差异
func (s *ReconciliationService) handleDifferences(ctx context.Context, 
	diff *ReconciliationDiff) error {
	
	// 1. 只在渠道有的（平台漏单）
	for _, record := range diff.OnlyInChannel {
		log.Warnf("平台漏单: %s", record.TradeNo)
		// 补单：创建支付记录
		s.createMissingPayment(ctx, record)
	}
	
	// 2. 只在平台有的（渠道无记录，可能未支付成功）
	for _, payment := range diff.OnlyInPlatform {
		log.Warnf("渠道无记录: %s", payment.PaymentID)
		// 主动查询第三方状态
		s.queryThirdPartyStatus(ctx, payment)
	}
	
	// 3. 金额不一致
	for _, mismatch := range diff.AmountMismatch {
		log.Errorf("金额不一致: %s, 平台=%v, 渠道=%v", 
			mismatch.TradeNo, mismatch.PlatformAmount, mismatch.ChannelAmount)
		// 转人工处理
		s.createManualTask(ctx, "AMOUNT_MISMATCH", mismatch)
	}
	
	// 4. 状态不一致
	for _, mismatch := range diff.StatusMismatch {
		log.Warnf("状态不一致: %s", mismatch.TradeNo)
		// 以渠道状态为准，更新平台状态
		s.syncStatus(ctx, mismatch)
	}
	
	return nil
}
```

**对账报告**：
```go
type ReconciliationReport struct {
	Date              time.Time
	TotalCount        int
	MatchCount        int
	MismatchCount     int
	OnlyInPlatform    int
	OnlyInChannel     int
	AmountMismatch    int
	TotalAmount       decimal.Decimal
	ChannelTotalAmount decimal.Decimal
}
```

**延伸思考**：
1. 对账差异如何自动修复？
2. 对账失败如何告警和处理？
3. 实时对账和T+1对账如何结合？

---

#### 📊 题目4：支付的异步回调处理

**问题描述**：
支付成功后，第三方通过回调通知平台。回调可能延迟、丢失、重复。如何设计健壮的回调处理机制？

**答案**：

**推荐方案**（Go实现）：

```go
// 回调处理器
type CallbackHandler struct {
	paymentSvc  *PaymentService
	orderSvc    *OrderService
	lockSvc     *DistributedLockService
}

// HandleCallback 处理回调
func (h *CallbackHandler) HandleCallback(ctx context.Context, 
	channel PaymentChannel, rawData []byte) error {
	
	// 1. 解析回调数据
	callback, err := parseCallback(channel, rawData)
	if err != nil {
		return fmt.Errorf("解析回调失败: %w", err)
	}
	
	// 2. 记录回调日志（用于排查问题）
	h.logCallback(ctx, callback)
	
	// 3. 验证签名
	adapter := h.paymentSvc.getAdapter(channel)
	if err := adapter.VerifyCallback(callback); err != nil {
		log.Errorf("回调签名验证失败: %v", err)
		return err
	}
	
	// 4. 幂等性处理（分布式锁）
	lockKey := fmt.Sprintf("payment:callback:%s", callback.OutTradeNo)
	acquired, err := h.lockSvc.TryLock(ctx, lockKey, 30*time.Second)
	if err != nil {
		return err
	}
	if !acquired {
		log.Infof("回调%s正在处理中，跳过", callback.OutTradeNo)
		return nil
	}
	defer h.lockSvc.Unlock(ctx, lockKey)
	
	// 5. 处理支付结果
	return h.paymentSvc.HandleCallback(ctx, channel, callback)
}

// 主动查询（回调超时补偿）
func (h *CallbackHandler) QueryPaymentStatus(ctx context.Context) {
	// 定时任务：查询10分钟前创建但未回调的支付单
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		cutoffTime := time.Now().Add(-10 * time.Minute)
		
		// 查询超时支付单
		payments, err := h.paymentSvc.FindPendingPayments(ctx, cutoffTime)
		if err != nil {
			log.Errorf("查询超时支付单失败: %v", err)
			continue
		}
		
		for _, payment := range payments {
			// 主动查询第三方状态
			go func(p *Payment) {
				adapter := h.paymentSvc.getAdapter(p.Channel)
				status, err := adapter.QueryPayment(ctx, p.ThirdPartyID)
				if err != nil {
					log.Errorf("查询支付状态失败: %v", err)
					return
				}
				
				// 如果已支付，补偿处理
				if status.Status == "SUCCESS" {
					log.Warnf("支付单%s回调丢失，主动补偿", p.PaymentID)
					h.paymentSvc.MarkAsPaid(ctx, p.PaymentID)
				}
			}(payment)
		}
	}
}
```

**回调重试策略**：
```go
// 回调处理失败时的重试
func (h *CallbackHandler) retryCallback(ctx context.Context, 
	callback *CallbackData) error {
	
	maxRetries := 5
	backoff := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		30 * time.Second,
		2 * time.Minute,
		10 * time.Minute,
	}
	
	for i := 0; i < maxRetries; i++ {
		err := h.HandleCallback(ctx, callback.Channel, callback.RawData)
		if err == nil {
			return nil // 成功
		}
		
		log.Warnf("回调处理失败，第%d次重试: %v", i+1, err)
		
		if i < maxRetries-1 {
			time.Sleep(backoff[i])
		}
	}
	
	// 所有重试失败，记录人工任务
	return h.createManualTask(ctx, "CALLBACK_FAILED", callback)
}
```

**延伸思考**：
1. 回调接口如何防止伪造（恶意请求）？
2. 回调处理超时如何设置？

---

#### 🔧 题目5：支付的分账系统设计（平台+商家）

**问题描述**：
B2B2C平台，用户支付100元，平台抽佣10%，商家获得90元。如何设计支付分账系统？

**答案**：

**推荐方案**（Go实现）：

```go
// Settlement 结算单
type Settlement struct {
	SettlementID   string
	OrderID        int64
	MerchantID     int64
	TotalAmount    decimal.Decimal  // 订单总额
	PlatformAmount decimal.Decimal  // 平台佣金
	MerchantAmount decimal.Decimal  // 商家收入
	Status         SettlementStatus
	SettledAt      *time.Time
}

// SettlementService 结算服务
type SettlementService struct {
	settlementRepo SettlementRepository
	paymentSvc     PaymentService
}

// CreateSettlement 创建结算单
func (s *SettlementService) CreateSettlement(ctx context.Context, 
	orderID int64) error {
	
	// 1. 查询订单
	order := s.orderSvc.GetOrder(ctx, orderID)
	
	// 2. 计算佣金
	commissionRate := s.getCommissionRate(ctx, order.MerchantID)
	platformAmount := order.TotalAmount.Mul(commissionRate)
	merchantAmount := order.TotalAmount.Sub(platformAmount)
	
	// 3. 创建结算单
	settlement := &Settlement{
		SettlementID:   generateSettlementID(),
		OrderID:        orderID,
		MerchantID:     order.MerchantID,
		TotalAmount:    order.TotalAmount,
		PlatformAmount: platformAmount,
		MerchantAmount: merchantAmount,
		Status:         SettlementPending,
	}
	
	return s.settlementRepo.Create(ctx, settlement)
}

// Settle 执行结算（T+N结算）
func (s *SettlementService) Settle(ctx context.Context, merchantID int64, date time.Time) error {
	// 1. 查询该商家待结算的订单
	settlements, err := s.settlementRepo.FindPendingByMerchant(ctx, merchantID, date)
	if err != nil {
		return err
	}
	
	// 2. 汇总金额
	totalAmount := decimal.Zero
	for _, s := range settlements {
		totalAmount = totalAmount.Add(s.MerchantAmount)
	}
	
	// 3. 调用支付渠道分账/转账
	if err := s.paymentSvc.Transfer(ctx, &TransferRequest{
		ToAccount: s.getMerchantAccount(ctx, merchantID),
		Amount:    totalAmount,
		Remark:    fmt.Sprintf("商家%d的%s结算", merchantID, date.Format("2006-01-02")),
	}); err != nil {
		return err
	}
	
	// 4. 更新结算单状态
	for _, settlement := range settlements {
		settlement.Status = SettlementCompleted
		settlement.SettledAt = timePtr(time.Now())
		s.settlementRepo.Update(ctx, settlement)
	}
	
	return nil
}

// 佣金率配置
func (s *SettlementService) getCommissionRate(ctx context.Context, merchantID int64) decimal.Decimal {
	// 根据商家等级、类目等确定佣金率
	merchant := s.merchantSvc.GetMerchant(ctx, merchantID)
	
	switch merchant.Level {
	case "VIP":
		return decimal.NewFromFloat(0.05) // 5%
	case "GOLD":
		return decimal.NewFromFloat(0.08) // 8%
	default:
		return decimal.NewFromFloat(0.10) // 10%
	}
}
```

**结算周期**：
```text
T+0：实时结算（高成本，高信用商家）
T+1：次日结算（平衡）
T+7：周结算（标准）
T+30：月结算（新商家）
```

**延伸思考**：
1. 如何设计结算的对账机制？
2. 商家提现如何设计？
3. 结算失败如何处理？

---

#### 📊 题目6：支付密码和安全设计

**问题描述**：
支付环节涉及资金安全，如何设计支付密码、短信验证码等安全机制？

**答案**：

**推荐方案**（Go实现）：

```go
// PaymentSecurityService 支付安全服务
type PaymentSecurityService struct {
	rdb        *redis.Client
	smsSvc     SMSService
	encryptSvc EncryptService
}

// VerifyPaymentPassword 验证支付密码
func (s *PaymentSecurityService) VerifyPaymentPassword(ctx context.Context, 
	userID int64, password string) error {
	
	// 1. 获取用户存储的支付密码（加密）
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	
	if user.PaymentPassword == "" {
		return errors.New("请先设置支付密码")
	}
	
	// 2. 验证密码
	if !s.encryptSvc.VerifyPassword(password, user.PaymentPassword) {
		// 记录失败次数
		failCount := s.incrFailCount(ctx, userID)
		
		// 超过5次锁定账户
		if failCount >= 5 {
			s.lockAccount(ctx, userID, 30*time.Minute)
			return errors.New("密码错误次数过多，账户已锁定30分钟")
		}
		
		return fmt.Errorf("密码错误，还可尝试%d次", 5-failCount)
	}
	
	// 3. 清除失败计数
	s.clearFailCount(ctx, userID)
	
	return nil
}

// SendPaymentSMS 发送支付验证码
func (s *PaymentSecurityService) SendPaymentSMS(ctx context.Context, 
	userID int64, phone string) error {
	
	// 1. 限流检查（防止短信轰炸）
	key := fmt.Sprintf("sms:limit:%s", phone)
	count, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		s.rdb.Expire(ctx, key, time.Hour)
	}
	if count > 5 {
		return errors.New("发送次数过多，请1小时后再试")
	}
	
	// 2. 生成6位验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	
	// 3. 存储验证码（5分钟有效）
	codeKey := fmt.Sprintf("sms:code:%s", phone)
	s.rdb.SetEX(ctx, codeKey, code, 5*time.Minute)
	
	// 4. 发送短信
	return s.smsSvc.Send(ctx, phone, fmt.Sprintf("您的支付验证码是%s，5分钟内有效", code))
}

// VerifySMSCode 验证短信验证码
func (s *PaymentSecurityService) VerifySMSCode(ctx context.Context, 
	phone, code string) error {
	
	codeKey := fmt.Sprintf("sms:code:%s", phone)
	
	// 查询验证码
	storedCode, err := s.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return errors.New("验证码已过期")
	}
	if err != nil {
		return err
	}
	
	// 验证
	if storedCode != code {
		return errors.New("验证码错误")
	}
	
	// 验证成功，删除验证码（防止重复使用）
	s.rdb.Del(ctx, codeKey)
	
	return nil
}

// 风控检查
func (s *PaymentSecurityService) RiskCheck(ctx context.Context, 
	userID int64, amount decimal.Decimal) error {
	
	// 规则1：大额支付需要额外验证
	if amount.GreaterThan(decimal.NewFromInt(5000)) {
		// 需要短信验证码或支付密码
		return errors.New("REQUIRE_SMS_OR_PASSWORD")
	}
	
	// 规则2：新用户限额
	user := s.userSvc.GetUser(ctx, userID)
	if user.RegisterDays() < 7 && amount.GreaterThan(decimal.NewFromInt(1000)) {
		return errors.New("新用户单笔限额1000元")
	}
	
	// 规则3：异常IP检测
	ip := s.getRequestIP(ctx)
	if s.isBlacklistIP(ctx, ip) {
		return errors.New("异常IP，禁止支付")
	}
	
	// 规则4：高频支付检测
	recentPayments := s.getRecentPaymentCount(ctx, userID, 10*time.Minute)
	if recentPayments > 10 {
		return errors.New("支付频率异常")
	}
	
	return nil
}
```

**延伸思考**：
1. 支付密码如何加密存储？
2. 如何设计支付的二次确认（大额支付）？
3. 支付安全如何平衡用户体验？

---

#### 🔧 题目7：支付渠道的路由和降级

**问题描述**：
支付宝渠道故障时，如何自动切换到微信支付？如何设计支付渠道的路由和降级策略？

**答案**：

**推荐方案**（Go实现）：

```go
// ChannelRouter 支付渠道路由器
type ChannelRouter struct {
	healthChecker *ChannelHealthChecker
	config        *RoutingConfig
}

// SelectChannel 选择支付渠道
func (r *ChannelRouter) SelectChannel(ctx context.Context, 
	preferredChannel PaymentChannel) (PaymentChannel, error) {
	
	// 1. 检查首选渠道健康状态
	if r.healthChecker.IsHealthy(preferredChannel) {
		return preferredChannel, nil
	}
	
	log.Warnf("渠道%s不可用，尝试降级", preferredChannel)
	
	// 2. 降级到备用渠道
	fallbackChannels := r.config.GetFallback(preferredChannel)
	for _, channel := range fallbackChannels {
		if r.healthChecker.IsHealthy(channel) {
			log.Infof("降级到渠道%s", channel)
			return channel, nil
		}
	}
	
	// 3. 所有渠道都不可用
	return "", errors.New("支付渠道暂时不可用，请稍后再试")
}

// ChannelHealthChecker 渠道健康检查
type ChannelHealthChecker struct {
	rdb *redis.Client
}

func (c *ChannelHealthChecker) IsHealthy(channel PaymentChannel) bool {
	key := fmt.Sprintf("payment:channel:health:%s", channel)
	
	// 从Redis读取健康状态
	status, err := c.rdb.Get(context.Background(), key).Result()
	if err != nil || status != "UP" {
		return false
	}
	
	return true
}

// 健康检查任务（心跳）
func (c *ChannelHealthChecker) StartHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		// 对每个渠道执行健康检查
		for _, channel := range AllChannels {
			go c.checkChannel(ctx, channel)
		}
	}
}

func (c *ChannelHealthChecker) checkChannel(ctx context.Context, 
	channel PaymentChannel) {
	
	adapter := getAdapter(channel)
	
	// 调用渠道健康检查接口（或创建1分钱订单测试）
	err := adapter.HealthCheck(ctx)
	
	key := fmt.Sprintf("payment:channel:health:%s", channel)
	if err != nil {
		// 不健康
		c.rdb.SetEX(ctx, key, "DOWN", 5*time.Minute)
		log.Errorf("渠道%s健康检查失败: %v", channel, err)
		
		// 告警
		c.alertSvc.Send(fmt.Sprintf("支付渠道%s故障", channel))
	} else {
		// 健康
		c.rdb.SetEX(ctx, key, "UP", 5*time.Minute)
	}
}

// 路由配置
type RoutingConfig struct {
	fallbacks map[PaymentChannel][]PaymentChannel
}

func NewRoutingConfig() *RoutingConfig {
	return &RoutingConfig{
		fallbacks: map[PaymentChannel][]PaymentChannel{
			ChannelAlipay: {ChannelWechat, ChannelUnion},  // 支付宝 → 微信 → 银联
			ChannelWechat: {ChannelAlipay, ChannelUnion},
			ChannelUnion:  {ChannelAlipay, ChannelWechat},
		},
	}
}
```

**延伸思考**：
1. 如何设计支付渠道的成本优化（选择手续费低的）？
2. 支付渠道限额如何处理？

---

#### 💡 题目8：支付的退款处理

**问题描述**：
用户申请退款，需要原路退回。如何设计退款流程，处理退款失败、部分退款等场景？

**答案**：

**推荐方案**（Go实现）：

```go
// RefundService 退款服务
type RefundService struct {
	paymentRepo PaymentRepository
}

// Refund 申请退款
func (s *RefundService) Refund(ctx context.Context, req *RefundRequest) error {
	// 1. 查询原支付记录
	payment, err := s.paymentRepo.FindByOrderID(ctx, req.OrderID)
	if err != nil {
		return err
	}
	
	// 2. 校验退款金额
	if req.RefundAmount.GreaterThan(payment.Amount) {
		return errors.New("退款金额超过支付金额")
	}
	
	// 3. 检查是否已退款
	totalRefunded, err := s.paymentRepo.GetTotalRefundedAmount(ctx, payment.PaymentID)
	if err != nil {
		return err
	}
	
	if totalRefunded.Add(req.RefundAmount).GreaterThan(payment.Amount) {
		return errors.New("累计退款金额超过支付金额")
	}
	
	// 4. 创建退款记录
	refund := &PaymentRefund{
		RefundID:    generateRefundID(),
		PaymentID:   payment.PaymentID,
		OrderID:     req.OrderID,
		Amount:      req.RefundAmount,
		Reason:      req.Reason,
		Status:      RefundStatusPending,
		CreatedAt:   time.Now(),
	}
	
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return err
	}
	
	// 5. 调用第三方退款接口
	adapter := s.getAdapter(payment.Channel)
	err = adapter.Refund(ctx, &ThirdPartyRefundRequest{
		OutRefundNo:   refund.RefundID,
		OutTradeNo:    payment.ThirdPartyID,
		RefundAmount:  req.RefundAmount,
		TotalAmount:   payment.Amount,
		RefundReason:  req.Reason,
	})
	
	if err != nil {
		refund.Status = RefundStatusFailed
		refund.FailReason = err.Error()
		s.refundRepo.Update(ctx, refund)
		return err
	}
	
	// 6. 更新退款状态
	refund.Status = RefundStatusSuccess
	refund.RefundedAt = timePtr(time.Now())
	s.refundRepo.Update(ctx, refund)
	
	return nil
}

// 退款重试（定时任务）
func (s *RefundService) RetryFailedRefunds(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		// 查询失败的退款（创建时间<30分钟前）
		refunds, err := s.refundRepo.FindFailed(ctx, time.Now().Add(-30*time.Minute))
		if err != nil {
			log.Errorf("查询失败退款失败: %v", err)
			continue
		}
		
		for _, refund := range refunds {
			// 重试退款
			go func(r *PaymentRefund) {
				if r.RetryCount >= 5 {
					log.Errorf("退款%s重试次数过多，转人工处理", r.RefundID)
					s.createManualTask(ctx, r.RefundID)
					return
				}
				
				payment, _ := s.paymentRepo.FindByID(ctx, r.PaymentID)
				adapter := s.getAdapter(payment.Channel)
				
				err := adapter.Refund(ctx, &ThirdPartyRefundRequest{
					OutRefundNo:  r.RefundID,
					OutTradeNo:   payment.ThirdPartyID,
					RefundAmount: r.Amount,
					TotalAmount:  payment.Amount,
				})
				
				if err == nil {
					r.Status = RefundStatusSuccess
					r.RefundedAt = timePtr(time.Now())
				} else {
					r.RetryCount++
					r.FailReason = err.Error()
				}
				
				s.refundRepo.Update(ctx, r)
			}(refund)
		}
	}
}
```

**部分退款处理**：
```go
// 部分退款（一单多件商品，退部分）
func (s *RefundService) PartialRefund(ctx context.Context, 
	orderID int64, items []RefundItem) error {
	
	// 1. 计算退款金额
	var refundAmount decimal.Decimal
	for _, item := range items {
		itemAmount := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
		refundAmount = refundAmount.Add(itemAmount)
	}
	
	// 2. 分摊运费
	order := s.orderSvc.GetOrder(ctx, orderID)
	refundItemCount := len(items)
	totalItemCount := len(order.Items)
	
	shippingRefund := order.ShippingFee.
		Mul(decimal.NewFromInt(int64(refundItemCount))).
		Div(decimal.NewFromInt(int64(totalItemCount)))
	
	refundAmount = refundAmount.Add(shippingRefund)
	
	// 3. 执行退款
	return s.Refund(ctx, &RefundRequest{
		OrderID:      orderID,
		RefundAmount: refundAmount,
		RefundItems:  items,
		Reason:       "部分退货",
	})
}
```

**延伸思考**：
1. 退款失败如何通知用户？
2. 如何设计退款的限额控制（防止洗钱）？

---

#### 🔧 题目9：支付的容灾和降级

**问题描述**：
支付是核心链路，不能中断。如何设计支付系统的容灾和降级方案？

**答案**：

**推荐方案**（Go实现）：

```go
// PaymentFallbackService 支付降级服务
type PaymentFallbackService struct {
	primarySvc   *PaymentService
	fallbackMode bool
}

// Pay 支付（带降级）
func (s *PaymentFallbackService) Pay(ctx context.Context, 
	req *PaymentRequest) (*PaymentResponse, error) {
	
	// 1. 尝试正常支付
	resp, err := s.primarySvc.CreatePayment(ctx, req)
	if err == nil {
		return resp, nil
	}
	
	log.Warnf("支付失败: %v，尝试降级", err)
	
	// 2. 降级方案
	if s.shouldFallback(err) {
		return s.fallbackPay(ctx, req)
	}
	
	return nil, err
}

// 降级支付
func (s *PaymentFallbackService) fallbackPay(ctx context.Context, 
	req *PaymentRequest) (*PaymentResponse, error) {
	
	// 降级策略1：切换支付渠道
	if req.Channel == ChannelAlipay {
		req.Channel = ChannelWechat
		return s.primarySvc.CreatePayment(ctx, req)
	}
	
	// 降级策略2：使用货到付款
	if s.isCODAvailable(req) {
		return s.createCODOrder(ctx, req)
	}
	
	// 降级策略3：延迟支付（订单保留，稍后支付）
	return s.createDelayedPayment(ctx, req)
}

// 熔断器
type CircuitBreaker struct {
	failureThreshold int
	timeout          time.Duration
	state            CircuitState
	failureCount     int
	lastFailTime     time.Time
}

type CircuitState int

const (
	StateClosed CircuitState = 0  // 闭合（正常）
	StateOpen   CircuitState = 1  // 开启（熔断）
	StateHalfOpen CircuitState = 2  // 半开（尝试恢复）
)

func (cb *CircuitBreaker) Execute(ctx context.Context, 
	fn func() error) error {
	
	// 检查熔断器状态
	if cb.state == StateOpen {
		// 检查是否可以尝试恢复
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = StateHalfOpen
		} else {
			return errors.New("熔断器开启，拒绝请求")
		}
	}
	
	// 执行函数
	err := fn()
	
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
	
	return err
}

func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailTime = time.Now()
	
	if cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
		log.Warn("熔断器开启")
	}
}

func (cb *CircuitBreaker) onSuccess() {
	if cb.state == StateHalfOpen {
		// 半开状态成功，恢复到闭合
		cb.state = StateClosed
		cb.failureCount = 0
		log.Info("熔断器关闭，恢复正常")
	}
}
```

**延伸思考**：
1. 如何设计支付系统的多机房容灾？
2. 支付降级后如何通知用户？

---

#### 📊 题目10：预授权支付的设计（酒店、租车场景）

**问题描述**：
酒店预订需要预授权（冻结资金但不扣款），退房时根据实际消费扣款。如何设计预授权支付？

**答案**：

**推荐方案**（Go实现）：

```go
// PreAuthService 预授权服务
type PreAuthService struct {
	paymentAdapter PaymentAdapter
	preAuthRepo    PreAuthRepository
}

// PreAuthorize 预授权
func (s *PreAuthService) PreAuthorize(ctx context.Context, 
	req *PreAuthRequest) (*PreAuthResponse, error) {
	
	// 1. 创建预授权记录
	preAuth := &PreAuthorization{
		PreAuthID:  generatePreAuthID(),
		OrderID:    req.OrderID,
		UserID:     req.UserID,
		Amount:     req.Amount,  // 冻结金额
		Status:     PreAuthStatusFrozen,
		CreatedAt:  time.Now(),
		ExpireAt:   time.Now().Add(30 * 24 * time.Hour), // 30天有效期
	}
	
	if err := s.preAuthRepo.Create(ctx, preAuth); err != nil {
		return nil, err
	}
	
	// 2. 调用支付渠道预授权接口
	resp, err := s.paymentAdapter.PreAuthorize(ctx, &ThirdPartyPreAuthRequest{
		OutRequestNo: preAuth.PreAuthID,
		Amount:       req.Amount,
		ExpireTime:   preAuth.ExpireAt,
	})
	
	if err != nil {
		preAuth.Status = PreAuthStatusFailed
		s.preAuthRepo.Update(ctx, preAuth)
		return nil, err
	}
	
	// 3. 保存第三方预授权号
	preAuth.ThirdPartyID = resp.AuthNo
	s.preAuthRepo.Update(ctx, preAuth)
	
	return &PreAuthResponse{
		PreAuthID: preAuth.PreAuthID,
		AuthNo:    resp.AuthNo,
	}, nil
}

// Complete 完成预授权（实际扣款）
func (s *PreAuthService) Complete(ctx context.Context, 
	preAuthID string, actualAmount decimal.Decimal) error {
	
	// 1. 查询预授权
	preAuth, err := s.preAuthRepo.FindByID(ctx, preAuthID)
	if err != nil {
		return err
	}
	
	// 2. 校验金额
	if actualAmount.GreaterThan(preAuth.Amount) {
		return errors.New("实际金额超过预授权金额")
	}
	
	// 3. 调用支付渠道完成预授权
	err = s.paymentAdapter.CompletePreAuth(ctx, &CompletePreAuthRequest{
		AuthNo: preAuth.ThirdPartyID,
		Amount: actualAmount,
	})
	
	if err != nil {
		return err
	}
	
	// 4. 更新状态
	preAuth.Status = PreAuthStatusCompleted
	preAuth.ActualAmount = actualAmount
	preAuth.CompletedAt = timePtr(time.Now())
	s.preAuthRepo.Update(ctx, preAuth)
	
	// 5. 多余金额解冻
	if actualAmount.LessThan(preAuth.Amount) {
		unfreezeAmount := preAuth.Amount.Sub(actualAmount)
		log.Infof("解冻多余金额: %v", unfreezeAmount)
	}
	
	return nil
}

// Cancel 取消预授权
func (s *PreAuthService) Cancel(ctx context.Context, preAuthID string) error {
	preAuth, err := s.preAuthRepo.FindByID(ctx, preAuthID)
	if err != nil {
		return err
	}
	
	// 调用支付渠道取消预授权
	err = s.paymentAdapter.CancelPreAuth(ctx, preAuth.ThirdPartyID)
	if err != nil {
		return err
	}
	
	preAuth.Status = PreAuthStatusCancelled
	s.preAuthRepo.Update(ctx, preAuth)
	
	return nil
}
```

**延伸思考**：
1. 预授权过期如何自动解冻？
2. 预授权场景下的对账如何设计？

---
