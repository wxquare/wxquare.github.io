# Code Examples - 配套代码示例

本目录包含电子书《B2B2C电商平台架构设计》的配套代码示例。

## 📁 目录结构

```
examples/
├── product-service/        # 商品中心服务示例（DDD四层架构）
├── order-service/          # 订单服务示例（待添加）
├── inventory-service/      # 库存服务示例（待添加）
├── pricing-service/        # 定价服务示例（待添加）
└── README.md              # 本文件
```

## 🎯 示例说明

### product-service - 商品中心服务

**对应章节**：16.6.1 商品中心设计

**技术栈**：
- Go 1.21+
- DDD四层架构（Domain, Application, Infrastructure, Interface）
- 三级缓存（L1本地 + L2 Redis + L3 MySQL）
- 领域事件模式

**快速开始**：
```bash
cd examples/product-service
go run cmd/main.go
```

**详细文档**：[product-service/README.md](./product-service/README.md)

---

## 🚀 后续计划

### order-service - 订单服务（规划中）
- **对应章节**：16.6.3 订单系统设计
- **核心功能**：订单状态机、Saga模式、价格快照
- **技术亮点**：分布式事务、补偿机制

### inventory-service - 库存服务（规划中）
- **对应章节**：16.6.2 库存系统设计
- **核心功能**：2D库存模型、预占扣减、超卖防护
- **技术亮点**：Redis Lua原子操作、分布式锁

### pricing-service - 定价服务（规划中）
- **对应章节**：16.6.4 定价系统设计
- **核心功能**：四层计价、优惠券、活动规则
- **技术亮点**：规则引擎、价格计算链

---

## 📚 学习路径

1. **第1步**：阅读对应章节理解理论
2. **第2步**：运行Demo观察数据流转
3. **第3步**：阅读源码理解实现细节
4. **第4步**：修改代码验证理解

---

## 🛠️ 环境要求

- Go 1.21+
- （可选）MySQL 8.0+
- （可选）Redis 6.0+
- （可选）Kafka 2.8+

当前Demo为简化实现，使用内存模拟数据库和缓存，可直接运行。

---

## 📖 参考文档

- [电子书源文件](../src/)
- [Chapter 16 - 核心系统设计](../src/part3/chapter16.md)
