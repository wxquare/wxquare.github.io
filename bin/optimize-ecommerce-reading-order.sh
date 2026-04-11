#!/bin/bash

# 按系统依赖层级优化电商系列文章的阅读顺序
# 原则：基础数据层 -> 业务规则层 -> 核心交易流程 -> 运营管理层

cd "$(dirname "$0")/.." || exit 1

# 定义新的阅读顺序映射
# 格式：文件名|新序号|新标题|新日期
declare -A new_order=(
    ["20-ecommerce-overview.md"]="一|电商系统设计（一）：全景概览与领域划分|2025-05-01"
    ["27-ecommerce-product-center.md"]="二|电商系统设计（二）：商品中心系统|2025-05-15"
    ["22-ecommerce-inventory.md"]="三|电商系统设计（三）：库存系统|2025-05-29"
    ["28-ecommerce-marketing-system.md"]="四|电商系统设计（四）：营销系统深度解析|2025-06-12"
    ["23-ecommerce-pricing-engine.md"]="五|电商系统设计（五）：计价引擎|2025-06-26"
    ["24-ecommerce-pricing-ddd.md"]="六|电商系统设计（六）：计价系统 DDD 实践|2025-07-10"
    ["26-ecommerce-order-system.md"]="七|电商系统设计（七）：订单系统|2025-07-24"
    ["29-ecommerce-payment-system.md"]="八|电商系统设计（八）：支付系统深度解析|2025-08-07"
    ["21-ecommerce-listing.md"]="九|电商系统设计（九）：商品上架系统|2025-08-21"
    ["25-ecommerce-b-side-ops.md"]="十|电商系统设计（十）：B 端运营系统|2025-09-04"
)

echo "=========================================="
echo "优化电商系列文章阅读顺序"
echo "=========================================="
echo ""
echo "优化原则：按系统依赖层级"
echo "  第一层：基础数据层（全景、商品中心）"
echo "  第二层：业务规则层（库存、营销、计价）"
echo "  第三层：核心交易流程（订单、支付）"
echo "  第四层：运营管理层（商品上架、B端运营）"
echo ""

# 遍历文章并更新
for file in "${!new_order[@]}"; do
    filepath="source/_posts/system-design/$file"
    
    if [ ! -f "$filepath" ]; then
        echo "❌ 文件不存在: $filepath"
        continue
    fi
    
    # 分离序号、标题和日期
    IFS='|' read -r new_num new_title new_date <<< "${new_order[$file]}"
    
    echo "📝 更新: $file"
    echo "   序号: ($new_num)"
    echo "   标题: $new_title"
    echo "   日期: $new_date"
    
    # 使用 sed 更新 Front Matter 中的 title 和 date
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/^title: .*/title: $new_title/" "$filepath"
        sed -i '' "s/^date: .*/date: $new_date/" "$filepath"
    else
        sed -i "s/^title: .*/title: $new_title/" "$filepath"
        sed -i "s/^date: .*/date: $new_date/" "$filepath"
    fi
    
    echo "   ✓ 更新完成"
    echo ""
done

echo "=========================================="
echo "✅ 所有文章已按新顺序更新完成！"
echo "=========================================="
echo ""
echo "新的阅读顺序："
echo "  (一) 全景概览与领域划分 - 建立整体认知"
echo "  (二) 商品中心系统 - 商品数据基础"
echo "  (三) 库存系统 - 库存管理"
echo "  (四) 营销系统深度解析 - 营销规则"
echo "  (五) 计价引擎 - 价格计算"
echo "  (六) 计价系统 DDD 实践 - 深入计价设计"
echo "  (七) 订单系统 - 核心交易"
echo "  (八) 支付系统深度解析 - 支付流程"
echo "  (九) 商品上架系统 - 商品上架流程"
echo "  (十) B 端运营系统 - 运营管理"
echo ""
