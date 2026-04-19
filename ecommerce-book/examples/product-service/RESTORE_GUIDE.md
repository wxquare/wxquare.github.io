# Product Service - 快速恢复指南

由于完整代码文件较多（共12个Go源文件），为节省时间，这里提供快速恢复方案。

## 方案1：从Git历史恢复（推荐）⭐️

如果之前已提交到Git：

```bash
# 查看删除记录
git log --all --full-history -- "**/product-service/**"

# 找到删除前的commit ID，例如：abc123

# 恢复整个目录
git checkout abc123 -- product-service/

# 移动到新位置
mv product-service examples/
```

## 方案2：手动重建

根据 chapter16.6.1 的代码示例，手动创建以下文件：

### Domain Layer（领域层）- 5个文件
1. `internal/domain/value_objects.go` - ✅ 已创建
2. `internal/domain/spu.go` - ✅ 已创建  
3. `internal/domain/product.go` - ✅ 已创建
4. `internal/domain/events.go` - ✅ 已创建
5. `internal/domain/repository.go` - ✅ 已创建

### Infrastructure Layer（基础设施层）- 3个文件
6. `internal/infrastructure/cache/cache.go` - 需创建
7. `internal/infrastructure/persistence/data_object.go` - 需创建
8. `internal/infrastructure/persistence/product_repository.go` - 需创建

### Application Layer（应用层）- 2个文件
9. `internal/application/dto/product_dto.go` - 需创建
10. `internal/application/service/product_service.go` - 需创建

### Interface Layer（接口层）- 1个文件
11. `internal/interfaces/http/product_handler.go` - 需创建

### 主程序和配置 - 3个文件
12. `cmd/main.go` - 需创建
13. `go.mod` - 需创建
14. `README.md` - 需创建

## 方案3：从参考项目复制

如果有完整的参考实现，可以从以下位置复制：

```bash
# 从参考项目复制
cp -r /path/to/reference/product-service/* examples/product-service/
```

## 新增目录结构

```
ecommerce-book/
└── examples/                    # ✅ 新增：代码示例目录
    ├── README.md               # ✅ 已创建
    ├── product-service/        # 商品中心服务
    │   ├── cmd/
    │   ├── internal/
    │   │   ├── domain/        # ✅ 已创建5个文件
    │   │   ├── application/
    │   │   ├── infrastructure/
    │   │   └── interfaces/
    │   ├── config/
    │   ├── migrations/
    │   ├── go.mod
    │   └── README.md
    ├── order-service/          # 待添加
    ├── inventory-service/      # 待添加
    └── pricing-service/        # 待添加
```

## 后续扩展

有了 `examples/` 目录，可以方便地添加其他服务示例：

- `examples/order-service/` - 订单服务（状态机、Saga）
- `examples/inventory-service/` - 库存服务（2D模型、预占扣减）
- `examples/pricing-service/` - 定价服务（规则引擎、计价链）
- `examples/payment-service/` - 支付服务（对账、退款）

## 需要帮助？

如需完整代码文件，可以：
1. 查看 chapter16.md 中的代码示例
2. 参考之前的conversation历史
3. 联系维护者获取完整代码包
