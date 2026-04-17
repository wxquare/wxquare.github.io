#!/bin/bash

# 从Hexo的source/book迁移到mdBook的src

SOURCE_DIR="../source/book"
TARGET_DIR="./src"

echo "开始迁移内容..."

# 函数：处理单个文件
process_file() {
    local src_file=$1
    local dest_file=$2
    
    if [ ! -f "$src_file" ]; then
        echo "⚠️  文件不存在: $src_file"
        return
    fi
    
    echo "✓ 处理: $(basename $src_file)"
    
    # 使用awk处理，更简单可靠
    awk '
    BEGIN { in_frontmatter=0; frontmatter_count=0 }
    /^---$/ { 
        frontmatter_count++
        if (frontmatter_count <= 2) {
            in_frontmatter = (frontmatter_count == 1)
            next
        }
    }
    !in_frontmatter && !/^\*\*导航\*\*:/ {
        gsub(/\.html\)/, ".md)")
        print
    }
    ' "$src_file" > "$dest_file"
}

# 迁移首页
echo "📄 迁移首页..."
process_file "$SOURCE_DIR/index.md" "$TARGET_DIR/README.md"

# 迁移第一部分（第1-4章）
echo "📚 迁移第一部分（第1-4章）..."
for i in {1..4}; do
    process_file "$SOURCE_DIR/chapter$i.md" "$TARGET_DIR/part1/chapter$i.md"
done

# 迁移第二部分 - 全局架构（第5-6章）
echo "📚 迁移第二部分 - Part A（第5-6章）..."
for i in 5 6; do
    process_file "$SOURCE_DIR/chapter$i.md" "$TARGET_DIR/part2/overview/chapter$i.md"
done

# 迁移第二部分 - 商品供给（第7-10章）
echo "📚 迁移第二部分 - Part B（第7-10章）..."
for i in {7..10}; do
    process_file "$SOURCE_DIR/chapter$i.md" "$TARGET_DIR/part2/supply/chapter$i.md"
done

# 迁移第二部分 - 交易链路（第11-15章）
echo "📚 迁移第二部分 - Part C（第11-15章）..."
for i in {11..15}; do
    process_file "$SOURCE_DIR/chapter$i.md" "$TARGET_DIR/part2/transaction/chapter$i.md"
done

# 迁移第三部分（第16章）
echo "📚 迁移第三部分（第16章）..."
process_file "$SOURCE_DIR/chapter16.md" "$TARGET_DIR/part3/chapter16.md"

# 创建占位文件
echo "📝 创建占位文件..."
echo "# 附录A 技术栈选型指南" > "$TARGET_DIR/appendix/tech-stack.md"
echo "# 附录B 面试题精选" > "$TARGET_DIR/appendix/interview.md"
echo "# 附录C 系统集成模式速查表" > "$TARGET_DIR/appendix/integration.md"
echo "# 附录D 术语表" > "$TARGET_DIR/appendix/glossary.md"
echo "# 附录E 参考资料" > "$TARGET_DIR/appendix/references.md"

echo ""
echo "✅ 迁移完成！"
echo "📊 统计信息："
echo "   - 第一部分：4章"
echo "   - 第二部分：11章"
echo "   - 第三部分：1章"
echo ""
echo "下一步："
echo "1. 运行: mdbook serve --open"
echo "2. 检查内容是否正确显示"
