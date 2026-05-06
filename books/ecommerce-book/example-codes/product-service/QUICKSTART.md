# Product Service - 快速开始指南 🚀

## 📊 项目状态

✅ **完整实现** - 所有层级已完成，可直接运行

```
✅ Domain Layer (领域层) - 5个文件
✅ Infrastructure Layer (基础设施层) - 3个文件  
✅ Application Layer (应用层) - 2个文件
✅ Interface Layer (接口层) - 2个文件（HTTP + gRPC示例）
✅ Main程序 - 1个文件
```

---

## 🚀 快速运行

### 第1步：进入项目目录

```bash
cd /Users/wxquare/go/src/github.com/wxquare.github.io/books/ecommerce-book/example-codes/product-service
```

### 第2步：运行Demo

```bash
go run cmd/main.go
```

### 第3步：观察输出

```
===========================================
🚀 Product Service - DDD 四层架构 Demo
===========================================

📦 Initializing dependencies...
✅ Dependencies initialized

✅ Test data initialized

🌐 HTTP Server starting on :8080...

===========================================
📋 Demo: 查询商品（展示完整数据流转）
===========================================

【数据流转路径】
HTTP Request → Interface Layer → Application Layer → Domain Layer → Infrastructure Layer
             ↓                  ↓                    ↓               ↓
           解析请求          业务编排             业务规则        三级缓存查询
             ↓                  ↓                    ↓               ↓
           DTO转换          调用Repository      聚合根方法      L1→L2→L3

▶️  第一次查询 SKUID=10001 (预期：L1 Miss → L2 Miss → L3 Hit)

🌐 [Interface Layer - HTTP] GET /api/v1/products/10001

🚀 [Application Layer] GetProduct called, SKUID=10001
❌ [L1 Miss] SKUID=10001, checking L2...
❌ [L2 Miss] SKUID=10001, checking L3...
✅ [L3 Hit] SKUID=10001 from MySQL
📝 [Cache Write] product:10001 written to L1+L2
✅ [Application Layer] GetProduct completed
✅ [Interface Layer - HTTP] Response sent
   Response: SKUID=10001, Name=iPhone 17 黑色 128GB, Price=¥7599.00, Status=DRAFT

▶️  第二次查询 SKUID=10001 (预期：L1 Hit)

🌐 [Interface Layer - HTTP] GET /api/v1/products/10001

🚀 [Application Layer] GetProduct called, SKUID=10001
✅ [L1 Hit] SKUID=10001 from Local Cache
✅ [Application Layer] GetProduct completed
✅ [Interface Layer - HTTP] Response sent
   Response: SKUID=10001, Name=iPhone 17 黑色 128GB, Price=¥7599.00, Status=DRAFT

▶️  第三次查询 SKUID=10002 (预期：L1 Miss → L2 Miss → L3 Hit)
...
```

---

## 📊 完整数据流转详解

### HTTP接口调用流程

```
客户端发起HTTP请求
    ↓
┌─────────────────────────────────────────────────────────────┐
│ 1. Interface Layer (接口层)                                  │
│    File: internal/interfaces/http/product_handler.go        │
│                                                              │
│    - 解析HTTP请求（路径参数）                                 │
│    - 参数校验（SKU ID格式）                                   │
│    - 创建DTO请求对象                                          │
│    - 调用应用服务                                             │
└──────────────────────┬──────────────────────────────────────┘
                       │ dto.GetProductRequest{SKUID: 10001}
                       ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Application Layer (应用层)                                │
│    File: internal/application/service/product_service.go    │
│                                                              │
│    - 接收DTO请求                                             │
│    - 转换DTO → Domain对象（domain.SKU_ID）                   │
│    - 调用Repository查询                                      │
│    - 转换Domain Model → DTO响应                              │
└──────────────────────┬──────────────────────────────────────┘
                       │ domain.NewSKU_ID(10001)
                       ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Domain Layer (领域层)                                     │
│    File: internal/domain/repository.go (接口定义)            │
│                                                              │
│    - 定义Repository接口                                      │
│    - FindBySKUID(ctx, skuID) → *Product                     │
│    - 不包含任何技术实现细节                                   │
└──────────────────────┬──────────────────────────────────────┘
                       │ 接口调用
                       ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Infrastructure Layer (基础设施层)                         │
│    File: internal/infrastructure/persistence/               │
│          product_repository.go                              │
│                                                              │
│    三级缓存查询流程：                                         │
│                                                              │
│    Step 1: L1 本地缓存                                       │
│      ├─ 查询 localCache.Get("product:10001")                │
│      ├─ ❌ Miss（第一次查询，缓存为空）                       │
│      └─ 继续查询L2                                           │
│                                                              │
│    Step 2: L2 Redis缓存                                      │
│      ├─ 查询 redisCache.Get(ctx, "product:10001")           │
│      ├─ ❌ Miss（第一次查询，缓存为空）                       │
│      └─ 继续查询L3                                           │
│                                                              │
│    Step 3: L3 MySQL数据库                                    │
│      ├─ 查询 mockDB[10001]                                  │
│      ├─ ✅ Hit（从数据库查到数据）                            │
│      ├─ 转换 ProductDO → domain.Product                     │
│      ├─ 回写L2缓存（TTL 30分钟）                             │
│      └─ 回写L1缓存（TTL 1分钟）                              │
│                                                              │
│    返回：domain.Product{                                     │
│      skuID: SKU_ID{value: 10001},                           │
│      spu: SPU{title: "iPhone 17"},                          │
│      basePrice: Price{amount: 759900},                      │
│      ...                                                     │
│    }                                                         │
└──────────────────────┬──────────────────────────────────────┘
                       │ domain.Product (返回)
                       ↓
        回到 Application Layer
                       │
                       ↓ toDTO(product)
                       │
        回到 Interface Layer
                       │
                       ↓ responseJSON(w, dto)
                       │
                       ↓
              返回HTTP响应给客户端
```

---

## 🎯 gRPC接口示例

虽然gRPC代码不要求编译通过，但已提供完整示例，展示如何集成：

### 文件位置

```
internal/interfaces/grpc/
├── proto/
│   └── product.proto          # Protobuf定义
└── product_handler.go         # gRPC Handler实现
```

### gRPC调用流程（与HTTP类似）

```
gRPC Client
    ↓
grpc.GetProduct(req) → product_handler.GetProduct()
    ↓
转换 gRPC Request → DTO
    ↓
调用 productService.GetProduct()
    ↓
... (后续流程与HTTP相同)
    ↓
转换 DTO → gRPC Response
    ↓
返回给gRPC Client
```

### 如何启用gRPC（实际项目）

1. **生成Go代码**：
```bash
protoc --go_out=. --go-grpc_out=. \
  internal/interfaces/grpc/proto/product.proto
```

2. **取消main.go中的注释**：
```go
// 启动gRPC服务器
go startGRPCServer(dependencies.grpcHandler)
```

3. **安装依赖**：
```bash
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

---

## 📁 完整文件清单

### Domain Layer（领域层）- 5个文件 ✅
```
internal/domain/
├── value_objects.go       # SKU_ID, Price, Specifications, ProductStatus
├── spu.go                 # SPU实体
├── product.go             # Product聚合根（核心业务逻辑）
├── events.go              # 4种领域事件
└── repository.go          # Repository接口
```

### Infrastructure Layer（基础设施层）- 3个文件 ✅
```
internal/infrastructure/
├── cache/
│   └── cache.go           # 本地缓存 + Redis缓存
└── persistence/
    ├── data_object.go     # ProductDO, SPUDO
    └── product_repository.go  # Repository实现（三级缓存）
```

### Application Layer（应用层）- 2个文件 ✅
```
internal/application/
├── dto/
│   └── product_dto.go     # GetProductRequest, GetProductResponse
└── service/
    └── product_service.go # 应用服务（业务编排）
```

### Interface Layer（接口层）- 2个文件 ✅
```
internal/interfaces/
├── http/
│   └── product_handler.go     # HTTP接口实现
└── grpc/
    ├── proto/
    │   └── product.proto      # Protobuf定义
    └── product_handler.go     # gRPC接口实现（示例）
```

### Main程序 - 1个文件 ✅
```
cmd/
└── main.go                # 服务启动 + Demo演示
```

### 配置和文档 - 4个文件 ✅
```
├── go.mod                 # Go模块定义
├── README.md              # 项目说明
├── RESTORE_GUIDE.md       # 恢复指南
└── QUICKSTART.md          # 本文件（快速开始）
```

---

## 🔍 关键设计模式

### 1. 依赖倒置（DIP）

```
Application Layer 依赖 → domain.ProductRepository (接口)
                              ↑
                              │ 实现
Infrastructure Layer ─────────┘
                    (persistence.ProductRepositoryImpl)
```

### 2. 三级缓存

```
L1 本地缓存（1分钟TTL）  ← 进程内，最快
   ↓ Miss
L2 Redis缓存（30分钟TTL） ← 跨服务，共享
   ↓ Miss
L3 MySQL数据库           ← 持久化存储
```

### 3. Cache-Aside模式

```
读流程：
1. 查询L1缓存 → Miss
2. 查询L2缓存 → Miss
3. 查询数据库 → Hit
4. 回写L2和L1缓存

写流程：
1. 更新数据库 ✅
2. 删除缓存 ✅（而非更新缓存）
3. 下次查询时重新加载最新数据
```

---

## 📚 学习路径

1. **运行Demo** - 观察完整的数据流转输出
2. **理解分层** - 每层的职责和边界
3. **阅读代码** - 按数据流转顺序阅读
4. **修改尝试** - 添加新的业务方法

---

## 🎉 总结

这个示例完整展示了：

✅ **DDD四层架构** - 清晰的职责划分  
✅ **完整数据流转** - HTTP → Interface → Application → Domain → Infrastructure  
✅ **三级缓存** - 性能优化最佳实践  
✅ **双接口支持** - HTTP + gRPC（示例）  
✅ **可运行代码** - 立即执行查看效果  

对应章节：**16.6.1 商品中心设计**
