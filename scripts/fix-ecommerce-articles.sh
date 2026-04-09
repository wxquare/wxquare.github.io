#!/bin/bash

# 优化电商系列文章的标题和日期
# 使其按照逻辑顺序排列

cd "$(dirname "$0")/.." || exit 1

# 定义文章标题和日期映射
declare -A articles=(
    ["20-ecommerce-overview.md"]="电商系统设计（一）：全景概览与领域划分|2025-05-01"
    ["21-ecommerce-listing.md"]="电商系统设计（二）：商品上架系统|2025-05-15"
    ["22-ecommerce-inventory.md"]="电商系统设计（三）：库存系统|2025-05-29"
    ["23-ecommerce-pricing-engine.md"]="电商系统设计（四）：计价引擎|2025-06-12"
    ["24-ecommerce-pricing-ddd.md"]="电商系统设计（五）：计价系统 DDD 实践|2025-06-26"
    ["25-ecommerce-b-side-ops.md"]="电商系统设计（六）：B 端运营系统|2025-07-10"
    ["26-ecommerce-order-system.md"]="电商系统设计（七）：订单系统|2025-07-24"
    ["27-ecommerce-product-center.md"]="电商系统设计（八）：商品中心系统|2025-08-07"
    ["28-ecommerce-marketing-system.md"]="电商系统设计（九）：营销系统深度解析|2025-08-21"
    ["29-ecommerce-payment-system.md"]="电商系统设计（十）：支付系统深度解析|2025-09-04"
)

# 遍历文章并更新
for file in "${!articles[@]}"; do
    filepath="source/_posts/system-design/$file"
    
    if [ ! -f "$filepath" ]; then
        echo "文件不存在: $filepath"
        continue
    fi
    
    # 分离标题和日期
    IFS='|' read -r new_title new_date <<< "${articles[$file]}"
    
    echo "正在更新: $file"
    echo "  新标题: $new_title"
    echo "  新日期: $new_date"
    
    # 使用 sed 更新 Front Matter 中的 title 和 date
    # macOS 的 sed 需要 -i '' 而不是 -i
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/^title: .*/title: $new_title/" "$filepath"
        sed -i '' "s/^date: .*/date: $new_date/" "$filepath"
    else
        sed -i "s/^title: .*/title: $new_title/" "$filepath"
        sed -i "s/^date: .*/date: $new_date/" "$filepath"
    fi
    
    echo "  ✓ 更新完成"
    echo ""
done

echo "所有文章已更新完成！"
echo ""
echo "建议执行以下命令查看更改："
echo "  git diff source/_posts/system-design/*-ecommerce*.md"
